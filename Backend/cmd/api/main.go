package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/config"
	"github.com/oms-saas/oms-saas-go/internal/database"
	"github.com/oms-saas/oms-saas-go/internal/email"
	"github.com/oms-saas/oms-saas-go/internal/handler"
	"github.com/oms-saas/oms-saas-go/internal/middleware"
	"github.com/oms-saas/oms-saas-go/internal/repository"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()
	migrationsDir := filepath.Join("db", "migrations")

	migratePool, err := database.NewPool(ctx, cfg.MigrationDatabaseURL)
	if err != nil {
		slog.Error("migration database connection failed", "error", err)
		os.Exit(1)
	}
	if err := database.RunMigrations(ctx, migratePool, migrationsDir); err != nil {
		migratePool.Close()
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	migratePool.Close()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	jwt := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	mailer := email.NewFromEnv()
	h := handler.New(repo, jwt, pool, mailer)
	authMW := middleware.NewJWTMiddleware(jwt)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.Health)
	r.Get("/api/v1/health", h.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(10, time.Minute))
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)
			r.Post("/logout", h.Logout)
			r.Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
				h.ForgotPassword(w, r, mailer)
			})
			r.Post("/reset-password", h.ResetPassword)
		})

		r.Post("/billing/webhook/stripe", h.StripeWebhook(cfg))

		r.Group(func(r chi.Router) {
			r.Use(authMW.Authenticate)

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", h.GetUserMe)
				r.Patch("/me", h.UpdateUserMe)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("owner", "admin"))
					r.Get("/", h.ListUsers)
					r.Post("/", h.CreateUser)
					r.Patch("/{id}", h.UpdateUser)
					r.Delete("/{id}", h.DeleteUser)
				})
			})

			r.Route("/tenant", func(r chi.Router) {
				r.Get("/me", h.GetTenantMe)
				r.Patch("/me", middleware.RequireRole("owner", "admin")(http.HandlerFunc(h.UpdateTenantMe)).ServeHTTP)
			})

			r.Route("/categories", func(r chi.Router) {
				r.Get("/", h.ListCategories)
				r.Post("/", h.CreateCategory)
				r.Put("/{id}", h.UpdateCategory)
				r.Delete("/{id}", h.DeleteCategory)
			})

			r.Route("/products", func(r chi.Router) {
				r.Get("/", h.ListProducts)
				r.Get("/export", h.ExportProductsCSV)
				r.Post("/", h.CreateProduct)
				r.Get("/{id}", h.GetProduct)
				r.Put("/{id}", h.UpdateProduct)
				r.With(middleware.RequireRole("owner", "admin")).Delete("/{id}", h.DeleteProduct)
			})

			r.Route("/customers", func(r chi.Router) {
				r.Get("/", h.ListCustomers)
				r.Post("/", h.CreateCustomer)
				r.Get("/{id}", h.GetCustomer)
				r.Put("/{id}", h.UpdateCustomer)
				r.Delete("/{id}", h.DeleteCustomer)
			})

			r.Route("/warehouses", func(r chi.Router) {
				r.Get("/", h.ListWarehouses)
				r.Post("/", h.CreateWarehouse)
				r.Put("/{id}", h.UpdateWarehouse)
				r.Delete("/{id}", h.DeleteWarehouse)
			})

			r.Route("/inventory", func(r chi.Router) {
				r.Get("/", h.ListInventory)
				r.Get("/warehouses", h.ListWarehouses)
				r.Post("/adjust", h.AdjustInventory)
				r.Post("/transfer", h.TransferInventory)
			})

			r.Route("/orders", func(r chi.Router) {
				r.Get("/", h.ListOrders)
				r.Post("/", h.CreateOrder)
				r.Get("/{id}", h.GetOrder)
				r.Patch("/{id}/status", h.UpdateOrderStatus)
			})

			r.Route("/reports", func(r chi.Router) {
				r.Get("/dashboard", h.GetDashboard)
				r.Get("/sales", h.GetSalesReport)
				r.Get("/sales/export", h.ExportSalesCSV)
				r.Get("/inventory", h.GetInventoryReport)
				r.Get("/inventory/export", h.ExportInventoryCSV)
			})

			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", h.ListNotifications)
				r.Patch("/{id}/read", h.MarkNotificationRead)
			})

			r.Route("/billing", func(r chi.Router) {
				r.Get("/plan", h.GetBillingPlan)
				r.With(middleware.RequireRole("owner", "admin")).Post("/checkout", h.CreateCheckout(cfg))
				r.With(middleware.RequireRole("owner", "admin")).Post("/portal", h.CreateBillingPortal(cfg))
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

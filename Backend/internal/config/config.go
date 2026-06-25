package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort              string
	DatabaseURL          string
	MigrationDatabaseURL string
	JWTSecret          string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	CORSOrigins        []string
	StripeSecretKey    string
	StripeWebhookSecret string
	StripePriceStarter string
	StripePricePro     string
}

func Load() Config {
	accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))

	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return Config{
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://oms_app:oms@localhost:5433/oms?sslmode=disable"),
		MigrationDatabaseURL: getEnv("MIGRATION_DATABASE_URL", "postgres://oms:oms@localhost:5433/oms?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTAccessTTL:        accessTTL,
		JWTRefreshTTL:       refreshTTL,
		CORSOrigins:         origins,
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceStarter:  getEnv("STRIPE_PRICE_STARTER", ""),
		StripePricePro:      getEnv("STRIPE_PRICE_PRO", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

type JWTMiddleware struct {
	jwt *auth.JWTManager
}

func NewJWTMiddleware(jwt *auth.JWTManager) *JWTMiddleware {
	return &JWTMiddleware{jwt: jwt}
}

func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := m.jwt.ParseAccessToken(token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
		if tenantID, err := uuid.Parse(claims.TenantID); err == nil {
			ctx = auth.WithTenantID(ctx, tenantID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}
			if _, ok := allowed[claims.Role]; !ok {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}

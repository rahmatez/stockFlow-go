package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/email"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

type Handler struct {
	repo   *repository.Repository
	jwt    *auth.JWTManager
	pool   *pgxpool.Pool
	mailer email.Sender
}

func New(repo *repository.Repository, jwt *auth.JWTManager, pool *pgxpool.Pool, mailer email.Sender) *Handler {
	return &Handler{repo: repo, jwt: jwt, pool: pool, mailer: mailer}
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		slog.Error("request error", "tenant_id", claims.TenantID, "error", err)
	} else {
		slog.Error("request error", "error", err)
	}
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, apperror.ErrUnauthorized):
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, apperror.ErrForbidden):
		response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, apperror.ErrConflict):
		response.Error(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, apperror.ErrValidation):
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, apperror.ErrLimitExceeded):
		response.Error(w, http.StatusPaymentRequired, "LIMIT_EXCEEDED", "plan limit exceeded, please upgrade")
	case errors.Is(err, apperror.ErrInvalidStatus):
		response.Error(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
	case errors.Is(err, apperror.ErrInsufficientStock):
		response.Error(w, http.StatusBadRequest, "INSUFFICIENT_STOCK", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func (h *Handler) tenantID(r *http.Request) (uuid.UUID, error) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.ErrUnauthorized
	}
	return uuid.Parse(claims.TenantID)
}

func (h *Handler) userID(r *http.Request) (*uuid.UUID, error) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return nil, apperror.ErrUnauthorized
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func listParams(r *http.Request) repository.ListParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	return repository.ListParams{
		Page:   page,
		Limit:  limit,
		Search: r.URL.Query().Get("search"),
	}.Normalize()
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if h.pool != nil {
		if err := h.pool.Ping(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database unreachable")
			return
		}
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

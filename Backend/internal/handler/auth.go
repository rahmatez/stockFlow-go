package handler

import (
	"net/http"
	"strings"

	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/models"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

type registerRequest struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
}

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	TenantSlug string `json:"tenant_slug"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.TenantName == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "all fields are required")
		return
	}
	if len(req.Password) < 8 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}

	user, tenant, err := h.repo.RegisterTenant(r.Context(), req.TenantName, req.Email, req.Password, req.FullName)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	tokens, hash, expires, err := h.jwt.GenerateTokenPair(user.ID.String(), user.TenantID.String(), user.Email, user.Role)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	if err := h.repo.SaveRefreshToken(r.Context(), user.ID, hash, expires); err != nil {
		h.handleError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id": user.ID, "email": user.Email, "full_name": user.FullName, "role": user.Role,
		},
		"tenant": tenant,
		"tokens": tokens,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	var user *models.User
	var err error
	req.TenantSlug = strings.TrimSpace(req.TenantSlug)

	if req.TenantSlug != "" {
		user, err = h.repo.GetUserByEmailAndTenant(r.Context(), req.Email, req.TenantSlug)
	} else {
		count, countErr := h.repo.CountUsersByEmail(r.Context(), req.Email)
		if countErr != nil {
			h.handleError(w, r, countErr)
			return
		}
		if count > 1 {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "multiple workspaces found, please provide tenant_slug")
			return
		}
		user, err = h.repo.GetUserByEmail(r.Context(), req.Email)
	}
	if err != nil {
		if err == apperror.ErrNotFound {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			return
		}
		h.handleError(w, r, err)
		return
	}

	if err := repository.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	tokens, hash, expires, err := h.jwt.GenerateTokenPair(user.ID.String(), user.TenantID.String(), user.Email, user.Role)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	if err := h.repo.SaveRefreshToken(r.Context(), user.ID, hash, expires); err != nil {
		h.handleError(w, r, err)
		return
	}

	tenant, _ := h.repo.GetTenant(r.Context(), user.TenantID)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id": user.ID, "email": user.Email, "full_name": user.FullName, "role": user.Role,
		},
		"tenant": tenant,
		"tokens": tokens,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	hash := auth.HashToken(req.RefreshToken)
	user, err := h.repo.GetRefreshToken(r.Context(), hash)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid refresh token")
		return
	}

	_ = h.repo.DeleteRefreshToken(r.Context(), hash)

	tokens, newHash, expires, err := h.jwt.GenerateTokenPair(user.ID.String(), user.TenantID.String(), user.Email, user.Role)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	if err := h.repo.SaveRefreshToken(r.Context(), user.ID, newHash, expires); err != nil {
		h.handleError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = decodeJSON(r, &req)
	if req.RefreshToken != "" {
		_ = h.repo.DeleteRefreshToken(r.Context(), auth.HashToken(req.RefreshToken))
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) GetTenantMe(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	tenant, err := h.repo.GetTenant(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	sub, _ := h.repo.GetTenantSubscription(r.Context(), tenantID)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"tenant":       tenant,
		"subscription": sub,
	})
}

func (h *Handler) UpdateTenantMe(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	tenant, err := h.repo.UpdateTenant(r.Context(), tenantID, req.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, tenant)
}

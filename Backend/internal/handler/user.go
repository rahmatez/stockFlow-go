package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/email"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) GetUserMe(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, err := h.userID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	u, err := h.repo.GetUserByID(r.Context(), tenantID, *userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"id": u.ID, "email": u.Email, "full_name": u.FullName, "role": u.Role,
	})
}

func (h *Handler) UpdateUserMe(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, err := h.userID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		FullName    *string `json:"full_name"`
		OldPassword *string `json:"old_password"`
		NewPassword *string `json:"new_password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if req.FullName != nil && *req.FullName != "" {
		u, err := h.repo.UpdateUserProfile(r.Context(), tenantID, *userID, *req.FullName)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"id": u.ID, "email": u.Email, "full_name": u.FullName, "role": u.Role,
		})
		return
	}
	if req.OldPassword != nil && req.NewPassword != nil {
		if err := h.repo.UpdateUserPassword(r.Context(), tenantID, *userID, *req.OldPassword, *req.NewPassword); err != nil {
			h.handleError(w, r, err)
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"message": "password updated"})
		return
	}
	response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "full_name or password fields required")
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	users, err := h.repo.ListUsers(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	items := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		items = append(items, map[string]interface{}{
			"id": u.ID, "email": u.Email, "full_name": u.FullName, "role": u.Role, "created_at": u.CreatedAt,
		})
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "email, password, and full_name are required")
		return
	}
	if len(req.Password) < 8 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}
	u, err := h.repo.CreateUser(r.Context(), tenantID, repository.CreateUserInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName, Role: req.Role,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	emailSent := false
	if _, noop := h.mailer.(email.NoopSender); !noop {
		tenant, _ := h.repo.GetTenant(r.Context(), tenantID)
		workspace := "your workspace"
		if tenant != nil {
			workspace = tenant.Name
		}
		body := fmt.Sprintf(
			"You've been invited to StockFlow\n\nWorkspace: %s\nEmail: %s\nTemporary password: %s\n\nSign in at: http://localhost:3000/login\n\nPlease change your password after signing in.",
			workspace, req.Email, req.Password,
		)
		if err := h.mailer.Send(req.Email, "You've been invited to StockFlow", body); err == nil {
			emailSent = true
		}
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"id": u.ID, "email": u.Email, "full_name": u.FullName, "role": u.Role, "email_sent": emailSent,
	})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid id")
		return
	}
	var req struct {
		FullName *string `json:"full_name"`
		Role     *string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	u, err := h.repo.UpdateUser(r.Context(), tenantID, id, repository.UpdateUserInput{
		FullName: req.FullName, Role: req.Role,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"id": u.ID, "email": u.Email, "full_name": u.FullName, "role": u.Role,
	})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	currentUserID, err := h.userID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid id")
		return
	}
	if id == *currentUserID {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "cannot delete yourself")
		return
	}
	if err := h.repo.DeleteUser(r.Context(), tenantID, id); err != nil {
		if err == apperror.ErrNotFound {
			h.handleError(w, r, err)
			return
		}
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

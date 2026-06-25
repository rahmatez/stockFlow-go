package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/email"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request, mailer email.Sender) {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Email == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "email is required")
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), strings.ToLower(req.Email))
	if err != nil {
		// Don't reveal whether email exists
		response.JSON(w, http.StatusOK, map[string]string{"message": "if the email exists, a reset link has been sent"})
		return
	}

	token := uuid.NewString()
	hash := auth.HashToken(token)
	expires := time.Now().Add(time.Hour)
	if err := h.repo.SavePasswordResetToken(r.Context(), user.ID, hash, expires); err != nil {
		h.handleError(w, r, err)
		return
	}

	_ = mailer.Send(user.Email, "Reset your StockFlow password",
		"Use this token to reset your password: "+token+"\n\nThis token expires in 1 hour.")

	response.JSON(w, http.StatusOK, map[string]string{"message": "if the email exists, a reset link has been sent"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Token == "" || req.NewPassword == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "token and new_password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}

	user, err := h.repo.GetUserByResetToken(r.Context(), auth.HashToken(req.Token))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid or expired token")
		return
	}

	if err := h.repo.SetUserPassword(r.Context(), user.TenantID, user.ID, req.NewPassword); err != nil {
		h.handleError(w, r, err)
		return
	}

	_ = h.repo.DeletePasswordResetToken(r.Context(), auth.HashToken(req.Token))
	response.JSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}

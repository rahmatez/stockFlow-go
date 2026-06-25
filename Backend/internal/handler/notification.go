package handler

import (
	"net/http"

	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	_ = h.repo.CheckAndCreateLowStockNotifications(r.Context(), tenantID)
	_ = h.repo.CheckAndCreateLimitWarningNotifications(r.Context(), tenantID)

	items, unread, err := h.repo.ListNotifications(r.Context(), tenantID, 20)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"notifications": items,
		"unread_count":  unread,
	})
}

func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.MarkNotificationRead(r.Context(), tenantID, id); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

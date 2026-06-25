package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	p := listParams(r)
	items, total, err := h.repo.ListOrders(r.Context(), tenantID, p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, items, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
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
	item, err := h.repo.GetOrder(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, _ := h.userID(r)

	var req struct {
		CustomerID *uuid.UUID `json:"customer_id"`
		Notes      *string    `json:"notes"`
		Items      []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
		} `json:"items"`
	}
	if err := decodeJSON(r, &req); err != nil || len(req.Items) == 0 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "items are required")
		return
	}

	var items []repository.OrderItemInput
	for _, it := range req.Items {
		items = append(items, repository.OrderItemInput{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	order, err := h.repo.CreateOrder(r.Context(), tenantID, repository.CreateOrderInput{
		CustomerID: req.CustomerID, Notes: req.Notes, Items: items, UserID: userID,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, order)
}

func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, _ := h.userID(r)
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid id")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Status == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "status is required")
		return
	}

	order, err := h.repo.UpdateOrderStatus(r.Context(), tenantID, id, req.Status, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, order)
}

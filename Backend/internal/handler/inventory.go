package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	p := listParams(r)
	items, total, err := h.repo.ListInventoryMovements(r.Context(), tenantID, p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, items, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

func (h *Handler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	items, err := h.repo.ListWarehouses(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, _ := h.userID(r)

	var req struct {
		WarehouseID  *uuid.UUID `json:"warehouse_id"`
		ProductID    uuid.UUID  `json:"product_id"`
		MovementType string     `json:"movement_type"`
		Quantity     int        `json:"quantity"`
		Notes        *string    `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil || req.ProductID == uuid.Nil || req.Quantity <= 0 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "product_id and quantity are required")
		return
	}
	if req.MovementType == "" {
		req.MovementType = "IN"
	}

	whID := req.WarehouseID
	if whID == nil {
		wh, err := h.repo.GetDefaultWarehouse(r.Context(), tenantID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		whID = &wh.ID
	}

	movement, err := h.repo.AdjustInventory(r.Context(), tenantID, repository.InventoryAdjustInput{
		WarehouseID: *whID, ProductID: req.ProductID, MovementType: req.MovementType,
		Quantity: req.Quantity, Notes: req.Notes, UserID: userID,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, movement)
}

func (h *Handler) TransferInventory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	userID, _ := h.userID(r)

	var req struct {
		FromWarehouseID uuid.UUID `json:"from_warehouse_id"`
		ToWarehouseID   uuid.UUID `json:"to_warehouse_id"`
		ProductID       uuid.UUID `json:"product_id"`
		Quantity        int       `json:"quantity"`
		Notes           *string   `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if err := h.repo.TransferInventory(r.Context(), tenantID, repository.InventoryTransferInput{
		FromWarehouseID: req.FromWarehouseID, ToWarehouseID: req.ToWarehouseID,
		ProductID: req.ProductID, Quantity: req.Quantity, Notes: req.Notes, UserID: userID,
	}); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "transfer completed"})
}

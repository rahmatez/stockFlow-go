package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	items, err := h.repo.ListCategories(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	item, err := h.repo.CreateCategory(r.Context(), tenantID, req.Name, req.Description)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
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
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	item, err := h.repo.UpdateCategory(r.Context(), tenantID, id, req.Name, req.Description)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.DeleteCategory(r.Context(), tenantID, id); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	p := listParams(r)
	items, total, err := h.repo.ListProducts(r.Context(), tenantID, p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, items, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
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
	item, err := h.repo.GetProduct(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		CategoryID        *uuid.UUID `json:"category_id"`
		SKU               string     `json:"sku"`
		Name              string     `json:"name"`
		Description       *string    `json:"description"`
		CostPrice         float64    `json:"cost_price"`
		SellPrice         float64    `json:"sell_price"`
		LowStockThreshold int        `json:"low_stock_threshold"`
		IsActive          *bool      `json:"is_active"`
	}
	if err := decodeJSON(r, &req); err != nil || req.SKU == "" || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "sku and name are required")
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	if req.LowStockThreshold == 0 {
		req.LowStockThreshold = 10
	}
	item, err := h.repo.CreateProduct(r.Context(), tenantID, repository.ProductInput{
		CategoryID: req.CategoryID, SKU: req.SKU, Name: req.Name, Description: req.Description,
		CostPrice: req.CostPrice, SellPrice: req.SellPrice, LowStockThreshold: req.LowStockThreshold, IsActive: active,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
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
		CategoryID        *uuid.UUID `json:"category_id"`
		SKU               string     `json:"sku"`
		Name              string     `json:"name"`
		Description       *string    `json:"description"`
		CostPrice         float64    `json:"cost_price"`
		SellPrice         float64    `json:"sell_price"`
		LowStockThreshold int        `json:"low_stock_threshold"`
		IsActive          bool       `json:"is_active"`
	}
	if err := decodeJSON(r, &req); err != nil || req.SKU == "" || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "sku and name are required")
		return
	}
	item, err := h.repo.UpdateProduct(r.Context(), tenantID, id, repository.ProductInput{
		CategoryID: req.CategoryID, SKU: req.SKU, Name: req.Name, Description: req.Description,
		CostPrice: req.CostPrice, SellPrice: req.SellPrice, LowStockThreshold: req.LowStockThreshold, IsActive: req.IsActive,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.DeleteProduct(r.Context(), tenantID, id); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

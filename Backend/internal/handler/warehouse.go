package handler

import (
	"net/http"

	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
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
	wh, err := h.repo.CreateWarehouse(r.Context(), tenantID, repository.WarehouseInput{Name: req.Name})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, wh)
}

func (h *Handler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	wh, err := h.repo.UpdateWarehouse(r.Context(), tenantID, id, repository.WarehouseInput{Name: req.Name})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, wh)
}

func (h *Handler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.DeleteWarehouse(r.Context(), tenantID, id); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "warehouse deleted"})
}

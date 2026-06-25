package handler

import (
	"net/http"

	"github.com/oms-saas/oms-saas-go/internal/repository"
	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	p := listParams(r)
	items, total, err := h.repo.ListCustomers(r.Context(), tenantID, p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, items, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
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
	item, err := h.repo.GetCustomer(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	var req struct {
		Name  string  `json:"name"`
		Email *string `json:"email"`
		Phone *string `json:"phone"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	item, err := h.repo.CreateCustomer(r.Context(), tenantID, repository.CustomerInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
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
		Name  string  `json:"name"`
		Email *string `json:"email"`
		Phone *string `json:"phone"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	item, err := h.repo.UpdateCustomer(r.Context(), tenantID, id, repository.CustomerInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
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
	if err := h.repo.DeleteCustomer(r.Context(), tenantID, id); err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

package handler

import (
	"net/http"
	"strconv"

	"github.com/oms-saas/oms-saas-go/internal/response"
)

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	stats, err := h.repo.GetDashboardStats(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	sales, err := h.repo.GetSalesReport(r.Context(), tenantID, 7)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	topProducts, err := h.repo.GetTopProducts(r.Context(), tenantID, 5)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	lowStock, err := h.repo.GetLowStockProducts(r.Context(), tenantID, 5)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	recentOrders, err := h.repo.GetRecentOrders(r.Context(), tenantID, 5)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	trends, _ := h.repo.GetDashboardTrends(r.Context(), tenantID)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"stats":         stats,
		"trends":        trends,
		"sales_chart":   sales,
		"top_products":  topProducts,
		"low_stock":     lowStock,
		"recent_orders": recentOrders,
	})
}

func (h *Handler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	sales, err := h.repo.GetSalesReport(r.Context(), tenantID, days)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, sales)
}

func (h *Handler) GetInventoryReport(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	lowStock, err := h.repo.GetLowStockProducts(r.Context(), tenantID, 20)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, lowStock)
}

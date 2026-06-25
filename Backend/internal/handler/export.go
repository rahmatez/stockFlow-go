package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
)

func (h *Handler) ExportProductsCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	p := listParams(r)
	p.Limit = 10000
	products, _, err := h.repo.ListProducts(r.Context(), tenantID, p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="products.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"SKU", "Name", "Sell Price", "Cost Price", "Stock", "Active"})
	for _, p := range products {
		_ = cw.Write([]string{
			p.SKU, p.Name,
			fmt.Sprintf("%.2f", p.SellPrice),
			fmt.Sprintf("%.2f", p.CostPrice),
			strconv.Itoa(p.StockOnHand),
			strconv.FormatBool(p.IsActive),
		})
	}
	cw.Flush()
}

func (h *Handler) ExportSalesCSV(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="sales-%d-days.csv"`, days))
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Date", "Revenue", "Orders"})
	for _, s := range sales {
		_ = cw.Write([]string{s.Date, fmt.Sprintf("%.2f", s.Revenue), strconv.FormatInt(s.Orders, 10)})
	}
	cw.Flush()
}

func (h *Handler) ExportInventoryCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	lowStock, err := h.repo.GetLowStockProducts(r.Context(), tenantID, 1000)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="inventory-low-stock.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"SKU", "Name", "Stock", "Threshold"})
	for _, p := range lowStock {
		_ = cw.Write([]string{p.SKU, p.Name, strconv.Itoa(p.StockOnHand), strconv.Itoa(p.LowStockThreshold)})
	}
	cw.Flush()
}

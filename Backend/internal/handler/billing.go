package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/oms-saas/oms-saas-go/internal/config"
	"github.com/oms-saas/oms-saas-go/internal/response"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
)

func (h *Handler) GetBillingPlan(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.tenantID(r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	sub, err := h.repo.GetTenantSubscription(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	plans, err := h.repo.ListPlans(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	productCount, _ := h.repo.CountProducts(r.Context(), tenantID)
	orderCount, _ := h.repo.CountOrdersThisMonth(r.Context(), tenantID)
	userCount, _ := h.repo.CountUsers(r.Context(), tenantID)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"subscription": sub,
		"plans":        plans,
		"usage": map[string]int64{
			"products":     productCount,
			"orders_month": orderCount,
			"users":        userCount,
		},
	})
}

func (h *Handler) CreateCheckout(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, err := h.tenantID(r)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		var req struct {
			PlanSlug   string `json:"plan_slug"`
			SuccessURL string `json:"success_url"`
			CancelURL  string `json:"cancel_url"`
		}
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
			return
		}

		if cfg.StripeSecretKey == "" {
			response.Error(w, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "stripe is not configured")
			return
		}

		stripe.Key = cfg.StripeSecretKey

		var priceID string
		switch req.PlanSlug {
		case "starter":
			priceID = cfg.StripePriceStarter
		case "pro":
			priceID = cfg.StripePricePro
		default:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid plan")
			return
		}
		if priceID == "" {
			response.Error(w, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "stripe price not configured")
			return
		}

		sub, _ := h.repo.GetTenantSubscription(r.Context(), tenantID)
		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
			},
			SuccessURL: stripe.String(req.SuccessURL),
			CancelURL:  stripe.String(req.CancelURL),
			Metadata: map[string]string{
				"tenant_id": tenantID.String(),
				"plan_slug": req.PlanSlug,
			},
		}
		if sub != nil && sub.StripeCustomerID != nil {
			params.Customer = sub.StripeCustomerID
		}

		sess, err := checkoutsession.New(params)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{
			"checkout_url": sess.URL,
			"session_id":   sess.ID,
		})
	}
}

func (h *Handler) CreateBillingPortal(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, err := h.tenantID(r)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		if cfg.StripeSecretKey == "" {
			response.Error(w, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "stripe is not configured")
			return
		}

		var req struct {
			ReturnURL string `json:"return_url"`
		}
		_ = decodeJSON(r, &req)

		sub, err := h.repo.GetTenantSubscription(r.Context(), tenantID)
		if err != nil || sub.StripeCustomerID == nil {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "no stripe customer found")
			return
		}

		stripe.Key = cfg.StripeSecretKey
		returnURL := req.ReturnURL
		if returnURL == "" {
			returnURL = "http://localhost:3000/settings/billing"
		}

		sess, err := session.New(&stripe.BillingPortalSessionParams{
			Customer:  sub.StripeCustomerID,
			ReturnURL: stripe.String(returnURL),
		})
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{"portal_url": sess.URL})
	}
}

func (h *Handler) StripeWebhook(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.StripeWebhookSecret == "" {
			response.Error(w, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "webhook not configured")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "cannot read body")
			return
		}

		event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), cfg.StripeWebhookSecret)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid signature")
			return
		}

		switch event.Type {
		case "checkout.session.completed":
			var sess stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &sess); err == nil {
				tenantID := sess.Metadata["tenant_id"]
				planSlug := sess.Metadata["plan_slug"]
				if tenantID != "" && planSlug != "" {
					tid, _ := parseUUID(tenantID)
					customerID, subscriptionID := "", ""
					if sess.Customer != nil {
						customerID = sess.Customer.ID
					}
					if sess.Subscription != nil {
						subscriptionID = sess.Subscription.ID
					}
					_ = h.repo.UpdateTenantSubscriptionPlan(r.Context(), tid, planSlug, customerID, subscriptionID)
				}
			}
		case "customer.subscription.updated":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
				_ = h.repo.UpdateTenantSubscriptionStatus(r.Context(), sub.ID, string(sub.Status))
			}
		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
				_ = h.repo.DowngradeTenantToFree(r.Context(), sub.ID)
			}
		case "invoice.payment_failed":
			var inv stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &inv); err == nil && inv.Subscription != nil {
				_ = h.repo.UpdateTenantSubscriptionStatus(r.Context(), inv.Subscription.ID, "past_due")
			}
		}

		response.JSON(w, http.StatusOK, map[string]string{"received": "true"})
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

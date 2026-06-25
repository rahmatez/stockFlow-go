package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/config"
	"github.com/oms-saas/oms-saas-go/internal/email"
)

func TestStripeWebhookRejectsInvalidSignature(t *testing.T) {
	h := New(nil, nil, nil, email.NoopSender{})
	cfg := config.Config{StripeWebhookSecret: "whsec_test"}

	body := []byte(`{"type":"checkout.session.completed","data":{"object":{}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "invalid")
	w := httptest.NewRecorder()

	h.StripeWebhook(cfg)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestStripeWebhookNotConfigured(t *testing.T) {
	h := New(nil, nil, nil, email.NoopSender{})
	cfg := config.Config{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook/stripe", nil)
	w := httptest.NewRecorder()

	h.StripeWebhook(cfg)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/auth"
	"github.com/oms-saas/oms-saas-go/internal/email"
)

func TestHealth(t *testing.T) {
	h := New(nil, auth.NewJWTManager("test", 0, 0), nil, email.NoopSender{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

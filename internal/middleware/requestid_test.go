package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/R055LE/go-deploy-lab/internal/middleware"
)

func TestRequestIDGenerated(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r.Context())
		if id == "" {
			t.Error("expected request_id in context, got empty")
		}
	})

	h := middleware.RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID in response header")
	}
}

func TestRequestIDHonorsExisting(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r.Context())
		if id != "existing-id" {
			t.Errorf("expected 'existing-id', got %q", id)
		}
	})

	h := middleware.RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "existing-id" {
		t.Errorf("expected 'existing-id' in response, got %q", got)
	}
}

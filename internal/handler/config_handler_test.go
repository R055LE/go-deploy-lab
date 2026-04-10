package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/R055LE/go-deploy-lab/internal/handler"
	"github.com/R055LE/go-deploy-lab/internal/model"
	"github.com/R055LE/go-deploy-lab/internal/store"
)

// fakeStore is an in-memory Store implementation for testing.
type fakeStore struct {
	entries map[string]model.ConfigEntry
}

func newFakeStore() *fakeStore {
	return &fakeStore{entries: make(map[string]model.ConfigEntry)}
}

func (f *fakeStore) key(ns, k string) string { return ns + "/" + k }

func (f *fakeStore) List(_ context.Context, namespace string) ([]model.ConfigEntry, error) {
	var result []model.ConfigEntry
	for _, e := range f.entries {
		if e.Namespace == namespace {
			result = append(result, e)
		}
	}
	return result, nil
}

func (f *fakeStore) Get(_ context.Context, namespace, key string) (*model.ConfigEntry, error) {
	e, ok := f.entries[f.key(namespace, key)]
	if !ok {
		return nil, store.ErrNotFound
	}
	return &e, nil
}

func (f *fakeStore) Put(_ context.Context, namespace, key, value string) (*model.ConfigEntry, error) {
	e := model.ConfigEntry{
		ID:        1,
		Namespace: namespace,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.entries[f.key(namespace, key)] = e
	return &e, nil
}

func (f *fakeStore) Delete(_ context.Context, namespace, key string) error {
	k := f.key(namespace, key)
	if _, ok := f.entries[k]; !ok {
		return store.ErrNotFound
	}
	delete(f.entries, k)
	return nil
}

func (f *fakeStore) Ping(_ context.Context) error { return nil }
func (f *fakeStore) Close()                        {}

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestPutAndGet(t *testing.T) {
	s := newFakeStore()
	ch := handler.NewConfigHandler(s, testLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/configs/{namespace}/{key}", ch.Put)
	mux.HandleFunc("GET /api/v1/configs/{namespace}/{key}", ch.Get)

	body := bytes.NewBufferString(`{"value":"bar"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/configs/test-ns/foo", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/configs/test-ns/foo", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d", rec.Code)
	}

	var entry model.ConfigEntry
	if err := json.NewDecoder(rec.Body).Decode(&entry); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if entry.Value != "bar" {
		t.Errorf("expected value 'bar', got %q", entry.Value)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newFakeStore()
	ch := handler.NewConfigHandler(s, testLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/configs/{namespace}/{key}", ch.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/configs/test-ns/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := newFakeStore()
	ch := handler.NewConfigHandler(s, testLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/configs/{namespace}/{key}", ch.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/configs/test-ns/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteSuccess(t *testing.T) {
	s := newFakeStore()
	if _, err := s.Put(context.Background(), "ns", "k", "v"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ch := handler.NewConfigHandler(s, testLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/configs/{namespace}/{key}", ch.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/configs/ns/k", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestListEmpty(t *testing.T) {
	s := newFakeStore()
	ch := handler.NewConfigHandler(s, testLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/configs/{namespace}", ch.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/configs/empty-ns", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []model.ConfigEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.Health()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReady(t *testing.T) {
	s := newFakeStore()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	handler.Ready(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/R055LE/go-deploy-lab/internal/middleware"
	"github.com/R055LE/go-deploy-lab/internal/model"
	"github.com/R055LE/go-deploy-lab/internal/store"
)

type ConfigHandler struct {
	store  store.Store
	logger *slog.Logger
}

func NewConfigHandler(s store.Store, logger *slog.Logger) *ConfigHandler {
	return &ConfigHandler{store: s, logger: logger}
}

type putRequest struct {
	Value string `json:"value"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *ConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	if namespace == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "namespace is required"})
		return
	}

	entries, err := h.store.List(r.Context(), namespace)
	if err != nil {
		h.logger.Error("list configs", slog.String("error", err.Error()), slog.String("request_id", middleware.GetRequestID(r.Context())))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	if entries == nil {
		entries = []model.ConfigEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	entry, err := h.store.Get(r.Context(), namespace, key)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	if err != nil {
		h.logger.Error("get config", slog.String("error", err.Error()), slog.String("request_id", middleware.GetRequestID(r.Context())))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (h *ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	var req putRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	entry, err := h.store.Put(r.Context(), namespace, key, req.Value)
	if err != nil {
		h.logger.Error("put config", slog.String("error", err.Error()), slog.String("request_id", middleware.GetRequestID(r.Context())))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (h *ConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	key := r.PathValue("key")

	err := h.store.Delete(r.Context(), namespace, key)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	if err != nil {
		h.logger.Error("delete config", slog.String("error", err.Error()), slog.String("request_id", middleware.GetRequestID(r.Context())))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

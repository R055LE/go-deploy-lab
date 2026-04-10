package handler

import (
	"encoding/json"
	"net/http"

	"github.com/R055LE/go-deploy-lab/internal/store"
)

type healthResponse struct {
	Status string `json:"status"`
}

// Health returns a liveness handler — always 200 if the process is alive.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
	}
}

// Ready returns a readiness handler — 200 only if the database is reachable.
func Ready(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := s.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(healthResponse{Status: "unavailable"})
			return
		}
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
	}
}

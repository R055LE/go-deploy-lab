package model

import "time"

// ConfigEntry represents a single configuration key-value pair within a namespace.
type ConfigEntry struct {
	ID        int64     `json:"id"`
	Namespace string    `json:"namespace"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

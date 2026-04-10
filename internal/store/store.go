package store

import (
	"context"

	"github.com/R055LE/go-deploy-lab/internal/model"
)

// Store defines the persistence interface for config entries.
type Store interface {
	List(ctx context.Context, namespace string) ([]model.ConfigEntry, error)
	Get(ctx context.Context, namespace, key string) (*model.ConfigEntry, error)
	Put(ctx context.Context, namespace, key, value string) (*model.ConfigEntry, error)
	Delete(ctx context.Context, namespace, key string) error
	Ping(ctx context.Context) error
	Close()
}

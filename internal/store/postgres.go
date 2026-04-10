package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/R055LE/go-deploy-lab/internal/metrics"
	"github.com/R055LE/go-deploy-lab/internal/model"
)

var ErrNotFound = errors.New("not found")

// Postgres implements Store using a pgx connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a connection pool and registers the active-connection gauge.
func NewPostgres(ctx context.Context, databaseURL string, reg prometheus.Registerer) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_active_connections",
			Help: "Number of active database connections.",
		},
		func() float64 {
			return float64(pool.Stat().AcquiredConns())
		},
	))

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) List(ctx context.Context, namespace string) ([]model.ConfigEntry, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("list").Observe(time.Since(start).Seconds()) }()

	rows, err := p.pool.Query(ctx,
		"SELECT id, namespace, key, value, created_at, updated_at FROM config_entries WHERE namespace = $1 ORDER BY key",
		namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("list configs: %w", err)
	}
	defer rows.Close()

	var entries []model.ConfigEntry
	for rows.Next() {
		var e model.ConfigEntry
		if err := rows.Scan(&e.ID, &e.Namespace, &e.Key, &e.Value, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (p *Postgres) Get(ctx context.Context, namespace, key string) (*model.ConfigEntry, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("get").Observe(time.Since(start).Seconds()) }()

	var e model.ConfigEntry
	err := p.pool.QueryRow(ctx,
		"SELECT id, namespace, key, value, created_at, updated_at FROM config_entries WHERE namespace = $1 AND key = $2",
		namespace, key,
	).Scan(&e.ID, &e.Namespace, &e.Key, &e.Value, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	return &e, nil
}

func (p *Postgres) Put(ctx context.Context, namespace, key, value string) (*model.ConfigEntry, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("put").Observe(time.Since(start).Seconds()) }()

	var e model.ConfigEntry
	err := p.pool.QueryRow(ctx, `
		INSERT INTO config_entries (namespace, key, value, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (namespace, key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		RETURNING id, namespace, key, value, created_at, updated_at
	`, namespace, key, value).Scan(&e.ID, &e.Namespace, &e.Key, &e.Value, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("put config: %w", err)
	}
	return &e, nil
}

func (p *Postgres) Delete(ctx context.Context, namespace, key string) error {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("delete").Observe(time.Since(start).Seconds()) }()

	tag, err := p.pool.Exec(ctx,
		"DELETE FROM config_entries WHERE namespace = $1 AND key = $2",
		namespace, key,
	)
	if err != nil {
		return fmt.Errorf("delete config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) Close() {
	p.pool.Close()
}

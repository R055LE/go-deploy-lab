package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/R055LE/go-deploy-lab/internal/config"
	"github.com/R055LE/go-deploy-lab/internal/handler"
	"github.com/R055LE/go-deploy-lab/internal/metrics"
	"github.com/R055LE/go-deploy-lab/internal/middleware"
	"github.com/R055LE/go-deploy-lab/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := newLogger(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	reg := prometheus.DefaultRegisterer
	metrics.Register(reg)

	db, err := store.NewPostgres(ctx, cfg.DatabaseURL, reg)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	mux := newRouter(db, logger)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", slog.Int("port", cfg.Port))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		logger.Info("server stopped gracefully")
	}

	return nil
}

func newRouter(s store.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	ch := handler.NewConfigHandler(s, logger)

	mux.HandleFunc("GET /health", handler.Health())
	mux.HandleFunc("GET /ready", handler.Ready(s))
	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /api/v1/configs/{namespace}", ch.List)
	mux.HandleFunc("GET /api/v1/configs/{namespace}/{key}", ch.Get)
	mux.HandleFunc("PUT /api/v1/configs/{namespace}/{key}", ch.Put)
	mux.HandleFunc("DELETE /api/v1/configs/{namespace}/{key}", ch.Delete)

	var h http.Handler = mux
	h = middleware.Metrics(h)
	h = middleware.Logging(logger)(h)
	h = middleware.RequestID(h)

	return h
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

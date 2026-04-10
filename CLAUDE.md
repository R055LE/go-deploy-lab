# go-deploy-lab

## Project Overview

A deliberately simple Go application with a deliberately thorough deployment lifecycle. Config registry (namespaced key-value store) backed by PostgreSQL. The app is boring on purpose — the infrastructure discipline is the point.

## Git Conventions

- Do NOT add Co-Authored-By trailers to commits.
- Use conventional commit format: `feat:`, `fix:`, `docs:`, `ci:`, `chore:`, etc.
- Keep commits atomic and descriptive.

## Architecture

- `cmd/server/` — Entrypoint. Config loading, signal handling, graceful shutdown.
- `internal/config/` — 12-factor env-based config with validation.
- `internal/handler/` — HTTP handlers for CRUD + health/readiness probes.
- `internal/middleware/` — Request ID, structured logging, Prometheus metrics.
- `internal/store/` — PostgreSQL repository using pgx/v5 connection pool.
- `internal/metrics/` — Prometheus collectors.
- `migrations/` — SQL migrations run via golang-migrate init container.
- `docker/` — Multi-stage Dockerfile targeting distroless.
- `k8s/` — Kubernetes manifests (deployment, service, ingress, HPA, PostgreSQL StatefulSet).
- `policies/` — Kyverno ClusterPolicies (non-root, resource limits, drop capabilities).
- `monitoring/` — Prometheus scrape config, Grafana dashboard as code.

## Build Order

```bash
task test       # Run tests
task build      # Build binary
task docker:build  # Build container image
```

## Key Constraints

- Go 1.26.2, pgx/v5 for database (not database/sql + lib/pq)
- Structured logging via stdlib slog (no external logging deps)
- Distroless final image — no shell, no package manager
- Migrations run as init container, not on application boot
- All containers run as non-root with dropped capabilities

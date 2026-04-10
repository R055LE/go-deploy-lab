# Architecture

## Overview

```
                    ┌──────────────┐
                    │   Ingress    │
                    │  (nginx)     │
                    └──────┬───────┘
                           │
              ┌────────────▼────────────┐
              │     Go Deploy Lab       │
              │                         │
              │  /health  /ready        │
              │  /metrics               │
              │  /api/v1/configs/...    │
              └────────────┬────────────┘
                           │
              ┌────────────▼────────────┐
              │     PostgreSQL          │
              │   (config_entries)      │
              └─────────────────────────┘
```

## Request Flow

1. Ingress terminates external traffic at `go-deploy-lab.localhost`
2. Request hits the middleware chain: Request ID → Logging → Metrics → Router
3. Handler executes the operation against PostgreSQL via the store interface
4. Response includes `X-Request-ID` header for tracing

## Deployment Flow

1. CI builds and scans the container image
2. Image is pushed to GHCR (main branch only)
3. ArgoCD detects the manifest change and syncs to the cluster
4. Init container runs `golang-migrate` against PostgreSQL
5. Application container starts after migrations complete
6. Rolling update ensures zero downtime (`maxUnavailable: 0, maxSurge: 1`)
7. Readiness probe gates traffic until the DB connection is verified

## Observability

- **Metrics**: Prometheus scrapes `/metrics` — request rate, latency histograms, error count, DB query duration, active DB connections
- **Logs**: Structured JSON via `slog` — every log line includes `request_id`, `method`, `path`, `status`, `duration`
- **Dashboard**: Grafana dashboard provisioned as code from `monitoring/grafana/dashboards/config-registry.json`

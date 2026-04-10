# Go Deploy Lab

[![CI](https://github.com/R055LE/go-deploy-lab/actions/workflows/ci.yml/badge.svg)](https://github.com/R055LE/go-deploy-lab/actions/workflows/ci.yml)

A deliberately simple Go application with a deliberately thorough deployment lifecycle — database migrations, container hardening, GitOps, observability, and graceful operational behavior. The app is boring on purpose; the infrastructure discipline is the point.

## Why This Exists

Portfolio projects that demonstrate container hardening, IaC, and platform bootstrap are common. What's less common is deploying a real application through those patterns end-to-end: database migrations via init container, secret injection, rolling updates with zero downtime, structured logging, Prometheus metrics, and a Grafana dashboard — all running on hardened, non-root, distroless containers.

## What This Demonstrates

- **Go application with production patterns**: graceful shutdown, structured JSON logging (`slog`), request ID propagation, health/readiness probes
- **PostgreSQL with pgx/v5**: connection pooling, query timeouts, instrumented with Prometheus histograms
- **Database migrations as init container**: `golang-migrate` runs before the application starts — not auto-migrate on boot
- **Container hardening**: multi-stage build, distroless final image (no shell, no package manager), non-root user (UID 65532), all capabilities dropped
- **Kubernetes deployment**: rolling updates (`maxUnavailable: 0`), HPA, resource limits, `automountServiceAccountToken: false`
- **Policy enforcement**: Kyverno ClusterPolicies requiring non-root, resource limits, and capability drops
- **Observability**: Prometheus metrics (request rate, latency histograms, DB query duration, active connections), Grafana dashboard as code
- **CI pipeline**: lint → test → build → Trivy scan → SBOM generation → GHCR push, plus kubeconform manifest validation

## The Application

A config registry — namespaced key-value store backed by PostgreSQL.

| Endpoint | Method | Description |
|---|---|---|
| `/api/v1/configs/{namespace}` | GET | List all keys in a namespace |
| `/api/v1/configs/{namespace}/{key}` | GET | Get a config value |
| `/api/v1/configs/{namespace}/{key}` | PUT | Create or update a config value |
| `/api/v1/configs/{namespace}/{key}` | DELETE | Delete a config value |
| `/health` | GET | Liveness probe |
| `/ready` | GET | Readiness probe (DB connected) |
| `/metrics` | GET | Prometheus metrics |

## Quick Start

### Prerequisites

```bash
./scripts/prerequisites.sh
```

### Run Locally (with Docker Compose or a local PostgreSQL)

```bash
export DATABASE_URL="postgres://app:changeme@localhost:5432/configstore?sslmode=disable"
task test
task build
task run
```

### Run on Kind

```bash
./scripts/setup-local.sh
kubectl -n go-deploy-lab port-forward svc/go-deploy-lab 8080:80
curl http://localhost:8080/health
```

### Test the API

```bash
# Create a config entry
curl -X PUT http://localhost:8080/api/v1/configs/myapp/log-level \
  -H "Content-Type: application/json" \
  -d '{"value":"debug"}'

# Read it back
curl http://localhost:8080/api/v1/configs/myapp/log-level

# List all configs in a namespace
curl http://localhost:8080/api/v1/configs/myapp

# Delete it
curl -X DELETE http://localhost:8080/api/v1/configs/myapp/log-level
```

## Project Structure

```
go-deploy-lab/
├── cmd/server/             # Application entrypoint
├── internal/
│   ├── config/             # 12-factor env-based configuration
│   ├── handler/            # HTTP handlers + tests
│   ├── middleware/          # Request ID, logging, Prometheus metrics
│   ├── model/              # Domain types
│   ├── store/              # PostgreSQL repository (pgx/v5)
│   └── metrics/            # Prometheus collectors
├── migrations/             # SQL migrations (golang-migrate)
├── docker/
│   ├── Dockerfile          # Multi-stage → distroless
│   └── Dockerfile.migrations
├── k8s/
│   ├── deployment.yml      # Hardened deployment with init container
│   ├── postgres/           # Local dev StatefulSet
│   └── ...                 # service, ingress, hpa, configmap, secret
├── argocd/                 # GitOps Application manifest
├── policies/               # Kyverno admission policies
├── monitoring/
│   ├── prometheus/         # Scrape config
│   └── grafana/            # Dashboard as code (JSON)
├── scripts/                # Prerequisites, setup, teardown
└── Taskfile.yml            # Build, test, lint, run tasks
```

## Design Decisions

| Decision | Rationale |
|---|---|
| `pgx/v5` over `database/sql` + `lib/pq` | Direct driver with connection pooling, prepared statements, and pgx-native types. No ORM. |
| `slog` (stdlib) over zerolog/zap | Zero external dependencies for logging. Structured JSON by default. |
| Distroless over Alpine | No shell = smaller attack surface. If you can't exec into it, attackers can't either. |
| Init container for migrations | Separates migration lifecycle from application lifecycle. Rollback doesn't re-run migrations. |
| No Helm chart | Raw manifests are more readable for a single-app deployment. Helm adds complexity without proportional value here. |
| Kyverno over OPA/Gatekeeper | Kubernetes-native CRDs, simpler policy language for admission control. OPA used in other labs for Rego-based static analysis. |

## Related Projects

- [Container Hardening Lab](https://github.com/R055LE/container-hardening-lab) — CIS/Iron Bank container hardening patterns used in this project's Dockerfile
- [K8s Bootstrap Lab](https://github.com/R055LE/k8s-bootstrap-lab) — The platform this app deploys onto (ArgoCD, Prometheus, Grafana, Kyverno)
- [IaC Security Lab](https://github.com/R055LE/iac-security-lab) — Terraform policy-as-code for the infrastructure layer beneath Kubernetes

# Deployment Lifecycle

## Database Migrations

Migrations are managed by [golang-migrate](https://github.com/golang-migrate/migrate) and run as a Kubernetes **init container** before the application starts.

### Why init container, not auto-migrate on boot?

- **Separation of concerns**: Migration failures don't crash the application. They block startup with a clear error.
- **Rollback safety**: Rolling back a deployment doesn't re-run migrations. The database schema and application version can be managed independently.
- **Concurrency**: Only one init container runs per pod rollout. Auto-migrate with multiple replicas risks race conditions.

### Migration file convention

```
migrations/
├── 001_create_config_entries.up.sql
├── 001_create_config_entries.down.sql
├── 002_add_ttl.up.sql          # future example
└── 002_add_ttl.down.sql
```

### Running locally

```bash
export DATABASE_URL="postgres://app:changeme@localhost:5432/configstore?sslmode=disable"
task migrate:up
task migrate:down   # rolls back one step
```

## Secret Management

Three approaches, from simplest to production-grade:

| Approach | When to use |
|---|---|
| **Kubernetes Secret** (current) | Local dev, demos. Credentials in plaintext YAML — never commit real values. |
| **SealedSecrets** | GitOps-compatible. Encrypt secrets client-side, decrypt in-cluster. Safe to commit. |
| **External Secrets Operator** | Production. Pulls secrets from AWS Secrets Manager, Vault, etc. at runtime. |

The `k8s/secret.yml` file documents the required secret shape. Replace it with your preferred approach.

## Rolling Updates

The deployment uses `maxUnavailable: 0, maxSurge: 1`:

1. A new pod is created with the updated image
2. The init container runs migrations (if any new migration files exist)
3. The application container starts and begins health checks
4. Once the readiness probe passes, traffic shifts to the new pod
5. The old pod is terminated

This guarantees zero downtime for any deployment that doesn't include a breaking migration.

## Configuration

All configuration is via environment variables (12-factor):

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | (required) | PostgreSQL connection string |
| `PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown deadline |

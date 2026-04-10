#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="go-deploy-lab"
NAMESPACE="go-deploy-lab"

echo "=== go-deploy-lab local setup ==="

# Idempotent: skip if cluster exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "Kind cluster '${CLUSTER_NAME}' already exists, skipping creation."
else
  echo "Creating Kind cluster..."
  kind create cluster --name "${CLUSTER_NAME}" --wait 60s
fi

kubectl config use-context "kind-${CLUSTER_NAME}"

# Deploy PostgreSQL
echo "Deploying PostgreSQL..."
kubectl apply -f k8s/namespace.yml
kubectl apply -f k8s/postgres/
kubectl -n "${NAMESPACE}" rollout status statefulset/postgres --timeout=120s

# Run migrations
echo "Running migrations..."
kubectl apply -f k8s/secret.yml
kubectl -n "${NAMESPACE}" run migrate --rm -i --restart=Never \
  --image=migrate/migrate:v4.18.2 \
  --overrides="$(cat <<EOF
{
  "spec": {
    "containers": [{
      "name": "migrate",
      "image": "migrate/migrate:v4.18.2",
      "command": ["migrate", "-path", "/migrations", "-database", "postgres://app:changeme@postgres:5432/configstore?sslmode=disable", "up"],
      "volumeMounts": [{"name": "migrations", "mountPath": "/migrations"}]
    }],
    "volumes": [{
      "name": "migrations",
      "configMap": {"name": "migrations"}
    }]
  }
}
EOF
)" 2>/dev/null || true

# Build and load app image
echo "Building application image..."
docker build -f docker/Dockerfile -t "${CLUSTER_NAME}:local" .
kind load docker-image "${CLUSTER_NAME}:local" --name "${CLUSTER_NAME}"

# Deploy application
echo "Deploying application..."
kubectl apply -f k8s/configmap.yml
kubectl apply -f k8s/deployment.yml
kubectl apply -f k8s/service.yml
kubectl -n "${NAMESPACE}" rollout status deployment/go-deploy-lab --timeout=120s

echo
echo "=== Setup complete ==="
echo "Port-forward: kubectl -n ${NAMESPACE} port-forward svc/go-deploy-lab 8080:80"
echo "Health check: curl http://localhost:8080/health"

#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
MISSING=0

check_tool() {
  local name=$1
  local install_hint=$2
  if command -v "$name" &>/dev/null; then
    printf "${GREEN}✓${NC} %s\n" "$name"
  else
    printf "${RED}✗${NC} %s — %s\n" "$name" "$install_hint"
    MISSING=1
  fi
}

echo "Checking prerequisites..."
echo

check_tool go          "https://go.dev/dl/"
check_tool docker      "https://docs.docker.com/get-docker/"
check_tool kubectl     "https://kubernetes.io/docs/tasks/tools/"
check_tool kind        "https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
check_tool task        "https://taskfile.dev/installation/"
check_tool migrate     "https://github.com/golang-migrate/migrate (optional — only needed for local migrations)"

echo
if [[ "$MISSING" -eq 1 ]]; then
  echo "Some tools are missing. Install them and re-run."
  exit 1
else
  echo "All prerequisites satisfied."
fi

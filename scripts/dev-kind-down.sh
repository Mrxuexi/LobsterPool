#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${1:-lobsterpool-dev}"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required but not installed" >&2
  exit 1
fi

kind delete cluster --name "${CLUSTER_NAME}"

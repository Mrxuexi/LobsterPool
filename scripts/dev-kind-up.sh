#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${1:-lobsterpool-dev}"
NAMESPACE="${2:-lobsterpool-local}"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required but not installed" >&2
  exit 1
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required but not installed" >&2
  exit 1
fi

if ! kind get clusters | grep -Fxq "${CLUSTER_NAME}"; then
  kind create cluster --name "${CLUSTER_NAME}"
fi

kubectl --context "kind-${CLUSTER_NAME}" create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl --context "kind-${CLUSTER_NAME}" apply -f -

cat <<EOF
kind cluster is ready.
cluster: ${CLUSTER_NAME}
context: kind-${CLUSTER_NAME}
namespace: ${NAMESPACE}

Example backend env:
LP_DEFAULT_CLUSTER=kind-dev
LP_KUBE_CLUSTERS=[{"name":"kind-dev","display_name":"Kind Dev","namespace":"${NAMESPACE}","kubeconfig":"${HOME}/.kube/config","context":"kind-${CLUSTER_NAME}"}]
EOF

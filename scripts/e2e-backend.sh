#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DB_PATH="${TMPDIR:-/tmp}/lobsterpool-e2e.db"
GOFLAGS_EXTRA=()

rm -f "${DB_PATH}" "${DB_PATH}-shm" "${DB_PATH}-wal"

if [[ "$(uname -s)" == "Darwin" ]]; then
  GOFLAGS_EXTRA=(-ldflags=-linkmode=external)
fi

cd "${ROOT_DIR}/backend"
LP_PORT=18080 \
LP_DB_PATH="${DB_PATH}" \
LP_DEV_MODE=true \
go run "${GOFLAGS_EXTRA[@]}" ./cmd/server

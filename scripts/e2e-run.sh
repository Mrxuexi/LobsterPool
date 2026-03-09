#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_LOG="${TMPDIR:-/tmp}/lobsterpool-e2e-backend.log"
FRONTEND_LOG="${TMPDIR:-/tmp}/lobsterpool-e2e-frontend.log"
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  if [[ -n "${FRONTEND_PID}" ]]; then
    kill "${FRONTEND_PID}" >/dev/null 2>&1 || true
    wait "${FRONTEND_PID}" >/dev/null 2>&1 || true
  fi

  if [[ -n "${BACKEND_PID}" ]]; then
    kill "${BACKEND_PID}" >/dev/null 2>&1 || true
    wait "${BACKEND_PID}" >/dev/null 2>&1 || true
  fi
}

wait_for_url() {
  local url="$1"

  for _ in $(seq 1 120); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "Timed out waiting for ${url}" >&2
  echo "Backend log:" >&2
  cat "${BACKEND_LOG}" >&2 || true
  echo "Frontend log:" >&2
  cat "${FRONTEND_LOG}" >&2 || true
  return 1
}

trap cleanup EXIT INT TERM

"${ROOT_DIR}/scripts/e2e-backend.sh" >"${BACKEND_LOG}" 2>&1 &
BACKEND_PID=$!

(
  cd "${ROOT_DIR}/frontend"
  VITE_API_PROXY_TARGET=http://127.0.0.1:18080 npm run dev -- --host 127.0.0.1 --port 4173
) >"${FRONTEND_LOG}" 2>&1 &
FRONTEND_PID=$!

wait_for_url "http://127.0.0.1:18080/api/v1/health"
wait_for_url "http://127.0.0.1:4173"

cd "${ROOT_DIR}/frontend"
npx playwright test "$@"

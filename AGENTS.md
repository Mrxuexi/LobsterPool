# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is LobsterPool

LobsterPool (龙虾池) is a Kubernetes-based platform for provisioning OpenClaw instances. Users fill in API keys and a Mattermost Bot Token, and the platform automatically creates K8s Secrets, Deployments, and Services. It is **not** an OpenClaw runtime — it is an instance provisioner.

## Commands

### Development (runs both backend + frontend)
```
make dev
```
Backend: `cd backend && LP_DEV_MODE=true go run ./cmd/server` (port 8080)
Frontend: `cd frontend && npm run dev` (Vite dev server, proxies `/api` to :8080)

### Build
```
make build              # both
make build-backend      # Go binary → bin/lobsterpool
make build-frontend     # Vite build → frontend/dist
```

### Lint (frontend only)
```
cd frontend && npm run lint
```

### Docker
```
make docker-build       # uses deploy/Dockerfile
```

### Clean
```
make clean              # removes bin/, frontend/dist, backend/lobsterpool.db
```

## Architecture

**Monorepo with two main directories:**

- `backend/` — Go (Gin framework), SQLite (pure-Go via modernc.org/sqlite, WAL mode), Kubernetes client-go
- `frontend/` — React 19, TypeScript, Vite, Ant Design 6, react-router-dom 7, i18next (zh/en)

### Backend structure (`backend/`)

```
cmd/server/main.go          — Entry point, wires config → db → stores → provider → handlers → router
internal/
  config/config.go          — Env-based config (LP_PORT, LP_DB_PATH, LP_NAMESPACE, LP_KUBECONFIG, LP_STATIC_DIR, LP_DEV_MODE)
  database/                 — SQLite open + migrations
  models/                   — Instance and ClawTemplate stores (data layer)
  provider/provider.go      — Provider interface: CreateInstance, DeleteInstance, GetInstanceStatus
  provider/kubernetes.go    — Real K8s provider (creates Secret, Deployment, Service)
  provider/mock.go          — Mock provider used when LP_DEV_MODE=true or LP_MOCK_PROVIDER=true
  handler/                  — HTTP handlers (instance, template, auth, health)
  middleware/               — CORS, request logger
  router/router.go          — All routes under /api/v1
```

### API routes

All under `/api/v1`:
- `GET /health`
- `GET|POST /templates`, `GET /templates/:id`
- `GET|POST /instances`, `GET|DELETE /instances/:id`
- Auth endpoints (placeholder): `/auth/login`, `/auth/logout`, `/auth/me`, `/auth/oauth/github`

### Frontend structure (`frontend/src/`)

```
App.tsx             — Router setup (/, /instances, /instances/create, /instances/:id, /templates)
api/client.ts       — Axios API client
pages/              — InstanceList, InstanceCreate, InstanceDetail, TemplateList
components/         — Layout, StatusBadge, SecretField
i18n/               — en.json, zh.json (Chinese/English)
types/index.ts      — Shared TypeScript types
```

### Provider abstraction

The `Provider` interface (`backend/internal/provider/provider.go`) abstracts infrastructure. Currently two implementations:
- **KubernetesProvider** — creates real K8s resources (Secret + Deployment + Service) in configurable namespace
- **MockProvider** — used in dev mode, no K8s cluster needed

### Key environment variables

| Variable | Default | Purpose |
|---|---|---|
| LP_PORT | 8080 | Server port |
| LP_DB_PATH | lobsterpool.db | SQLite database path |
| LP_NAMESPACE | lobsterpool | K8s namespace for instances |
| LP_KUBECONFIG | ~/.kube/config | Kubeconfig path |
| LP_STATIC_DIR | ./static | Frontend build directory |
| LP_DEV_MODE | false | Enables mock provider, skips K8s |

## Conventions

- Backend uses standard Go project layout (`cmd/` + `internal/`)
- Frontend uses i18next for all user-facing strings — add translations to both `zh.json` and `en.json`
- K8s resources for instances follow naming pattern: `claw-{instance_id}`
- The Vite dev server proxies `/api` requests to the Go backend at localhost:8080
- In production, the Go server serves the frontend build as static files

## 设计原则

- 总是遵循第一性原理
.PHONY: dev dev-backend dev-frontend dev-backend-k8s dev-kind-up dev-kind-down build build-backend build-frontend lint lint-frontend test test-backend test-e2e docker-build clean

UNAME_S := $(shell uname -s)
BACKEND_DEV_GOFLAGS :=
BACKEND_BUILD_ENV := CGO_ENABLED=0
BACKEND_BUILD_GOFLAGS :=
BACKEND_TEST_GOFLAGS :=

ifeq ($(UNAME_S),Darwin)
# Go 1.21.x can emit Mach-O binaries without LC_UUID on some macOS setups.
# For local macOS development, force the external linker so dyld accepts the binary.
BACKEND_DEV_GOFLAGS := -ldflags='-linkmode=external'
BACKEND_BUILD_ENV :=
BACKEND_BUILD_GOFLAGS := -ldflags='-linkmode=external'
BACKEND_TEST_GOFLAGS := -ldflags='-linkmode=external'
endif

# Development
dev:
	@echo "Starting backend and frontend dev servers..."
	$(MAKE) dev-backend &
	$(MAKE) dev-frontend
	@wait

dev-backend:
	cd backend && LP_DEV_MODE=true go run $(BACKEND_DEV_GOFLAGS) ./cmd/server

dev-backend-k8s:
	cd backend && LP_DEV_MODE=false LP_DEFAULT_CLUSTER=kind-dev LP_KUBE_CLUSTERS='[{"name":"kind-dev","display_name":"Kind Dev","namespace":"lobsterpool-local","kubeconfig":"'"$$HOME"'/.kube/config","context":"kind-lobsterpool-dev"}]' go run $(BACKEND_DEV_GOFLAGS) ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev-kind-up:
	./scripts/dev-kind-up.sh lobsterpool-dev lobsterpool-local

dev-kind-down:
	./scripts/dev-kind-down.sh lobsterpool-dev

# Build
build: build-frontend build-backend

build-backend:
	cd backend && $(BACKEND_BUILD_ENV) go build $(BACKEND_BUILD_GOFLAGS) -o ../bin/lobsterpool ./cmd/server

build-frontend:
	cd frontend && npm run build

# Checks
lint: lint-frontend

lint-frontend:
	cd frontend && npm run lint

test: test-backend

test-backend:
	cd backend && go test $(BACKEND_TEST_GOFLAGS) ./...

test-e2e:
	cd frontend && npm run test:e2e

# Docker
docker-build:
	docker build -f deploy/Dockerfile -t lobsterpool:latest .

# Clean
clean:
	rm -rf bin/
	rm -rf frontend/dist
	rm -f backend/lobsterpool.db
	rm -f backend/lobsterpool.db-shm
	rm -f backend/lobsterpool.db-wal
	rm -f lobsterpool.db
	rm -f lobsterpool.db-shm
	rm -f lobsterpool.db-wal

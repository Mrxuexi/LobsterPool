# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o lobsterpool ./cmd/server

# Stage 3: Final image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend-builder /app/lobsterpool /lobsterpool
COPY --from=frontend-builder /app/frontend/dist /static
ENV LP_STATIC_DIR=/static
ENV LP_DB_PATH=/data/lobsterpool.db
EXPOSE 8080
ENTRYPOINT ["/lobsterpool"]

# syntax=docker/dockerfile:1.7

########################################
# Stage 1 - Frontend build (Vue3 via Vite)
########################################
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web

# Install dependencies first to leverage Docker layer caching
# Note: devDependencies (including vite) are required for build
COPY web/package*.json ./
RUN npm ci --include=dev

# Copy the rest of the frontend source (node_modules excluded via .dockerignore)
COPY web/ .
RUN npm run build

########################################
# Stage 2 - Backend build (Go 1.25.1)
########################################
FROM golang:1.25.1-alpine AS backend-builder
WORKDIR /app

# Install build tooling and download Go modules
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code (large artifacts excluded via .dockerignore)
COPY . .

# Build-time metadata (override via --build-arg)
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

ENV CGO_ENABLED=0

RUN go build -trimpath -ldflags "-s -w \
    -X 'github.com/cuihe500/vaulthub/pkg/version.Version=${VERSION}' \
    -X 'github.com/cuihe500/vaulthub/pkg/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/cuihe500/vaulthub/pkg/version.BuildTime=${BUILD_TIME}'" \
    -o /tmp/vaulthub ./cmd/vaulthub

########################################
# Stage 3 - Runtime image (Alpine)
########################################
FROM alpine:3
WORKDIR /app

RUN addgroup -S vaulthub && adduser -S vaulthub -G vaulthub \
    && apk add --no-cache ca-certificates tzdata \
    && mkdir -p /app/configs /app/web /app/internal/database/migrations

COPY --from=backend-builder /tmp/vaulthub ./vaulthub
COPY --from=frontend-builder /app/web/dist ./web/dist
COPY configs/config.toml.example ./configs/config.toml
COPY internal/database/migrations ./internal/database/migrations

ENV GIN_MODE=release \
    TZ=Asia/Shanghai \
    VAULTHUB_STATIC_DIR=/app/web/dist \
    VAULTHUB_CONFIG=/app/configs/config.toml

VOLUME ["/app/configs"]
EXPOSE 8080

USER vaulthub

ENTRYPOINT ["./vaulthub"]
CMD ["serve", "--config", "/app/configs/config.toml"]

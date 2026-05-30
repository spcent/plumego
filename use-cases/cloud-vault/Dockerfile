# Build stage: frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Build stage: backend
FROM golang:1.26-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cloud-vault ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/cloud-vault .
COPY config.example.toml .
COPY env.example .

# Create non-root user
RUN addgroup -g 1000 cloudvault && \
    adduser -D -u 1000 -G cloudvault cloudvault && \
    mkdir -p /app/data /app/backups && \
    chown -R cloudvault:cloudvault /app

USER cloudvault

VOLUME ["/app/data", "/app/backups"]
EXPOSE 8080

ENTRYPOINT ["./cloud-vault"]
CMD ["--config", "/app/config.toml"]

# syntax=docker/dockerfile:1

FROM node:24-bookworm AS frontend
WORKDIR /src/use-cases/dbadmin/web

COPY use-cases/dbadmin/web/package.json use-cases/dbadmin/web/package-lock.json ./
RUN npm ci

COPY use-cases/dbadmin/web ./
RUN npm run build

FROM golang:1.26-bookworm AS backend
WORKDIR /src

COPY go.mod go.sum ./
COPY contract ./contract
COPY core ./core
COPY health ./health
COPY internal/httputil ./internal/httputil
COPY log ./log
COPY metrics ./metrics
COPY middleware ./middleware
COPY router ./router
COPY security ./security
COPY store ./store
COPY x/frontend ./x/frontend
COPY use-cases/dbadmin ./use-cases/dbadmin
COPY --from=frontend /src/use-cases/dbadmin/web/dist /src/use-cases/dbadmin/web/dist

WORKDIR /src/use-cases/dbadmin
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/dbadmin .

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --system --create-home --home-dir /app --shell /usr/sbin/nologin dbadmin

WORKDIR /app
COPY --from=backend /out/dbadmin /usr/local/bin/dbadmin
COPY use-cases/dbadmin/env.example /app/.env.example

USER dbadmin
ENV APP_ADDR=0.0.0.0:8080
ENV DBADMIN_DATA_DIR=/app/data
VOLUME ["/app/data"]
EXPOSE 8080

ENTRYPOINT ["dbadmin"]

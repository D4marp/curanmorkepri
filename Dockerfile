# syntax=docker/dockerfile:1
# =============================================================================
# CURANMOR AI — Ditreskrimum Polda Kepulauan Riau
# Multi-stage build: kompilasi binari Go statis, lalu image runtime minimal
# (distroless-style, non-root) berisi binari + migrations + scripts + docs + web.
# =============================================================================

FROM golang:1.24-alpine AS build
WORKDIR /src

# Dependensi sudah di-vendor (folder vendor/) sehingga build ini TIDAK
# memerlukan akses jaringan sama sekali — cocok untuk lingkungan internal
# Polri yang terbatas aksesnya ke internet publik.
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor go build -trimpath -ldflags="-s -w" -o /out/curanmor-api ./cmd/api

# -----------------------------------------------------------------------------
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S curanmor && adduser -S curanmor -G curanmor

WORKDIR /app
COPY --from=build /out/curanmor-api /app/curanmor-api
COPY migrations/ /app/migrations/
COPY scripts/ /app/scripts/
COPY docs/ /app/docs/
COPY web/ /app/web/

RUN mkdir -p /app/uploads && chown -R curanmor:curanmor /app

USER curanmor
EXPOSE 8080

ENV MIGRATIONS_DIR=/app/migrations \
    SEED_DIR=/app/scripts \
    DOCS_DIR=/app/docs \
    WEB_DIR=/app/web \
    UPLOAD_DIR=/app/uploads \
    HTTP_PORT=8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/index.html >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/curanmor-api"]

# Stage 1 — Builder
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Copy dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy source and build static binaries
COPY . .
RUN go build -ldflags="-s -w" -o /tabby ./cmd/tabby && \
    go build -ldflags="-s -w" -o /healthcheck ./cmd/healthcheck

# Stage 2 — Runtime (scratch: empty filesystem)
FROM scratch

# trunk-ignore-all(checkov/CKV_DOCKER_3): We run as root because the filesystem is empty otherwise

# CA certificates for future HTTPS API calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binaries
COPY --from=builder /tabby /tabby
COPY --from=builder /healthcheck /healthcheck

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["/healthcheck"]

ENTRYPOINT ["/tabby", "-config", "/etc/tabby/tabby.toml"]

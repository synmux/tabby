# Stage 1 — Builder
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Copy dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tabby ./cmd/tabby

# Stage 2 — Runtime (scratch: empty filesystem)
FROM scratch

# CA certificates for future HTTPS API calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary
COPY --from=builder /tabby /tabby

EXPOSE 8080

ENTRYPOINT ["/tabby", "-config", "/etc/tabby/tabby.toml"]

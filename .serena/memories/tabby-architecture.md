# Tabby Architecture

## Overview
Go web server serving a browser new tab page dashboard with bento-grid layout and live SSE updates.

## Key Dependencies
- `github.com/BurntSushi/toml` — config parsing
- `modernc.org/sqlite` — CGo-free SQLite cache

## Data Pipeline
Scheduler (per-widget ticker) → Widget.Fetch() → Render HTML partial → Cache in SQLite → SSE Broker broadcasts → htmx sse-swap updates DOM

## Package Map
- `cmd/tabby/main.go` — Entry point, wiring all components
- `internal/config/` — TOML config parsing with Duration support
- `internal/server/` — HTTP server, routing (Go 1.22+ ServeMux), middleware (logging, recovery)
- `internal/sse/` — SSE broker with channel-per-client fan-out
- `internal/cache/` — SQLite cache with TTL-based expiry
- `internal/widget/` — Widget interface, Registry, Scheduler, FetchAll
- `internal/widgets/*/` — 8 widget implementations (all mock data currently)
- `web/` — Embedded static assets + HTML templates via //go:embed

## Widget Interface
```go
type Widget interface {
    Name() string
    Fetch(ctx context.Context) (*Data, error)
    RefreshInterval() time.Duration
    Enabled() bool
}
```

## Current Status
Phase 1-3 complete: runnable server with 8 mock widgets, SSE pipeline, SQLite cache, all tests passing.
Phase 4 (real API calls) and Phase 5 (TLS, polish) are future work.

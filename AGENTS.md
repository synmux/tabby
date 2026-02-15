# Tabby — Agent Context

## Overview

Tabby is a Go web server that serves an HTML dashboard as a browser new tab page. It aggregates data from multiple sources into a bento-grid layout with live SSE updates.

## Architecture

- **Server:** `net/http` stdlib, no framework
- **Frontend:** Vanilla JS + htmx (SSE extension), no build step
- **Config:** TOML via `github.com/BurntSushi/toml`
- **Cache:** SQLite via `modernc.org/sqlite` (CGo-free)
- **Live updates:** SSE → htmx `sse-swap` for DOM updates
- **Assets:** Embedded in binary via `//go:embed`

## Project Structure

```
cmd/tabby/main.go           — Entry point, wiring
internal/config/             — TOML config parsing
internal/server/             — HTTP server, routing, middleware
internal/sse/                — SSE broker (client registry + broadcast)
internal/cache/              — SQLite cache (get/set/purge)
internal/widget/             — Widget interface, Registry, Scheduler, Fetcher
internal/widgets/*/          — Individual widget implementations
web/                         — Embedded static assets + HTML templates
```

## Key Patterns

- **Widget interface:** Every data source implements `Widget` (Name, Fetch, RefreshInterval, Enabled)
- **Data pipeline:** Scheduler → Widget.Fetch() → Cache in SQLite → SSE Broadcast → htmx swap
- **Registry:** Central store for widgets and their latest data
- **Mock data:** All widgets currently use mock data with randomised values per fetch

## Commands

```bash
go build ./cmd/tabby    # Build
go test ./...           # Run all tests
go vet ./...            # Static analysis
go run ./cmd/tabby      # Run server (default localhost:8080)
```

## Configuration

Copy `tabby.example.toml` to `tabby.toml`. Server runs with defaults if no config file exists.

## Conventions

- British English spelling
- Descriptive variable names (no single-letter names)
- Table-driven tests with stdlib `testing` only
- No external test frameworks
- `html/template` for all HTML rendering (XSS-safe)

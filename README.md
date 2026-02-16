# Tabby

A Go web server that serves an attractive HTML dashboard as a browser new tab page. Aggregates data from multiple sources into a bento-grid layout with live SSE updates.

## Features

- **Bento grid layout** — 4-column responsive grid with widget cards
- **Live updates** — Server-Sent Events push HTML partials, htmx swaps them into the DOM
- **8 widget types** — Weather, RSS, Uptime, GitHub, Linear, Notion, OMI, Calendar
- **SQLite cache** — Persists widget data across restarts
- **Single binary** — All assets embedded via `//go:embed`
- **Zero build step** — No Node.js, no bundler, no framework

## Quick Start

```bash
# Build
go build ./cmd/tabby

# Run (uses defaults if no config file)
./tabby

# Or run directly
go run ./cmd/tabby
```

Open `http://localhost:8080` in your browser.

## Configuration

Copy the example configuration and customise:

```bash
cp tabby.example.toml tabby.toml
```

See `tabby.example.toml` for all available options. The server runs with sensible defaults if no configuration file is present.

## Architecture

```plaintext
Scheduler (per-widget ticker)
  → Widget.Fetch()
  → Render HTML partial
  → Store in SQLite cache
  → SSE Broker broadcasts
  → htmx sse-swap updates DOM
```

### API Endpoints

| Method | Path                          | Purpose            |
| ------ | ----------------------------- | ------------------ |
| GET    | `/`                           | Main dashboard     |
| GET    | `/static/*`                   | Embedded CSS/JS    |
| GET    | `/api/events`                 | SSE stream         |
| GET    | `/api/widgets`                | List all widgets   |
| GET    | `/api/widgets/{name}`         | Single widget data |
| POST   | `/api/widgets/{name}/refresh` | Force refresh      |
| GET    | `/api/health`                 | Health check       |

## Docker

Build the image:

```bash
docker build -t tabby .
```

Run with defaults:

```bash
docker run -p 8080:8080 tabby
```

Run with a custom configuration:

```bash
docker run -p 8080:8080 \
  -v ./tabby.toml:/etc/tabby/tabby.toml:ro \
  tabby
```

Persist the cache across restarts:

```bash
docker run -p 8080:8080 \
  -v ./tabby.toml:/etc/tabby/tabby.toml:ro \
  -v tabby-cache:/var/lib/tabby/ \
  tabby
```

> **Important:** Your `tabby.toml` must set `host = "0.0.0.0"` under `[server]` when running inside Docker. The default `localhost` binds only to the container's loopback interface and will not be reachable from the host.

When using the cache volume, set `database_path` in your config:

```toml
[cache]
database_path = "/var/lib/tabby/tabby-cache.db"
```

## Development

```bash
# Lint
trunk check -a --show-existing

# Run tests
go test ./...

# Static analysis
go vet ./...

# Run with live reloading (using air or similar)
go run ./cmd/tabby
```

## Current Status

All widgets use **mock data** — dynamic values that change on each fetch for visual testing of the SSE pipeline. Real API integrations are planned for a future phase.

## Requirements

- Go 1.22+ (uses enhanced `ServeMux` patterns)

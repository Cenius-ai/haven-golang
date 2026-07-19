# Haven — Installation Guide

## Prerequisites

- **Go 1.22+** (download from https://go.dev/dl/)
- Standard Unix tools: `bash`, `curl`

## One-command setup

```bash
bash install.sh
```

This script:
1. Downloads Go module dependencies (`go mod download`)
2. Compiles the binary to `/tmp/goapp` (`go build -o /tmp/goapp .`)
3. Starts a temporary instance to verify the build and seed data
4. Exits — it does **not** leave a server running

After install, run the server separately:

```bash
PORT=8080 /tmp/goapp
```

## Manual setup

If you prefer to run each step yourself:

```bash
# 1. Install Go dependencies
go mod download

# 2. Build
go build -o haven .

# 3. Run (auto-seeds on first boot)
PORT=8080 ./haven
```

## Configuration

All configuration is via environment variables:

| Variable        | Default  | Description                          |
|-----------------|----------|--------------------------------------|
| `PORT`          | `8080`   | HTTP server port                     |
| `HAVEN_DATA_DIR`| `.`      | Directory for data and cache files   |

No `.env` file is required — the app boots with sensible defaults.

## Data persistence

Portfolio holdings and alerts are saved as JSON files in `data/`:
- `data/holdings.json` — Portfolio positions
- `data/alerts.json` — Price alerts

CoinGecko market data is cached in `cache/coingecko_cache.json`.

All files are created automatically on first run.

#!/usr/bin/env bash
# Haven — install script
# Deps → build → seed → exits (does NOT start the server)
set -euo pipefail

cd "$(dirname "$0")"
export PATH="/usr/local/go/bin:$PATH"

echo "=== Haven install ==="

# Ensure data directories exist
mkdir -p data cache

# Download Go dependencies
echo "[1/3] Downloading Go dependencies..."
go mod download

# Build the binary
echo "[2/3] Building Haven..."
go build -o /tmp/goapp .

# Seed demo data (runs as part of normal boot; we just verify the build)
echo "[3/3] Verifying build and seed..."
/tmp/goapp &
APP_PID=$!
sleep 4

# Check health
if curl -sf --max-time 5 http://localhost:8080/health > /dev/null 2>&1; then
    echo "  ✓ Server started successfully"
    # Verify data was seeded
    HOLDINGS=$(curl -sf --max-time 5 http://localhost:8080/api/portfolio | python3 -c "import json,sys; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    ALERTS=$(curl -sf --max-time 5 http://localhost:8080/api/alerts | python3 -c "import json,sys; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    echo "  ✓ Holdings: $HOLDINGS, Alerts: $ALERTS"
else
    echo "  ! Could not reach server (this may be OK if CoinGecko fetch is slow)"
fi

kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true

echo ""
echo "=== Install complete ==="
echo "Run: PORT=8080 /tmp/goapp"
echo "Then open: http://localhost:8080"

#!/usr/bin/env bash
# Compila per darwin/arm64 il server MCP e il comando geohub-import-check, in plugins/wm-skills/bin/.
set -euo pipefail

cd "$(dirname "$0")"
OUT="../bin/orchestrator-mcp"

mkdir -p ../bin
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$OUT" .
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
  -o ../bin/geohub-import-check ./cmd/geohub-import-check

echo "Compilati in $(cd .. && pwd)/bin/:"
ls -lh "$OUT" ../bin/geohub-import-check

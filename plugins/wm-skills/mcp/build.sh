#!/usr/bin/env bash
# Compila il server MCP per darwin/arm64 e lo colloca in plugins/wm-skills/bin/.
set -euo pipefail

cd "$(dirname "$0")"
OUT="../bin/orchestrator-mcp"

mkdir -p ../bin
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$OUT" .

echo "Compilato: $(cd .. && pwd)/bin/orchestrator-mcp"
ls -lh "$OUT"

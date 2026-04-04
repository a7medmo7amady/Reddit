#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$SCRIPT_DIR/Architecture"
PORT=9090
IMAGE="structurizr/structurizr"

if [ ! -d "$WORKSPACE_DIR" ]; then
  echo "Error: workspace directory not found: $WORKSPACE_DIR"
  exit 1
fi

if [ ! -f "$WORKSPACE_DIR/workspace.dsl" ] && [ ! -f "$WORKSPACE_DIR/workspace.json" ]; then
  echo "Warning: no workspace.dsl or workspace.json found in $WORKSPACE_DIR"
fi

docker run --rm -it \
  -p "${PORT}:8080" \
  -v "$WORKSPACE_DIR:/usr/local/structurizr" \
  "$IMAGE" local
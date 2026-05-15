#!/usr/bin/env bash
set -euo pipefail

GO_VERSION="1.26"

if [ $# -eq 0 ]; then
  echo "Usage: $0 <service-dir>"
  echo "Example: $0 user-service"
  exit 1
fi

SERVICE_DIR="$1"

docker run --rm \
  -v "$PWD/$SERVICE_DIR":/app \
  -w /app \
  "golang:${GO_VERSION}-alpine" \
  sh -c "go mod tidy && go mod edit -go=${GO_VERSION} && go mod edit -toolchain=none"

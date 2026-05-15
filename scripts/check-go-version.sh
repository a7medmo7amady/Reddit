#!/usr/bin/env bash
set -euo pipefail

if grep -R "go 1.2[0-5]" . --include="go.mod"; then
  echo "Invalid Go version found. Project requires Go 1.26."
  exit 1
fi

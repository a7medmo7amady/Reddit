#!/usr/bin/env bash
# Generates a self-signed cert for local development.
# Run once from the repo root: bash nginx/gen-dev-certs.sh
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)/certs"
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "$DIR/server.key" \
  -out    "$DIR/server.crt" \
  -subj   "/CN=localhost"
echo "Certs written to $DIR"

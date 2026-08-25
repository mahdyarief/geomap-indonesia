#!/bin/bash
set -e

# Helper script to run Go import scripts inside Docker
# Usage: ./scripts/run-import.sh <script-name>
# Example: ./scripts/run-import.sh import_master

SCRIPT_NAME="${1:-import_master}"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
WSL_PROJECT_DIR="/mnt${PROJECT_DIR//\/\//\/}"

echo "Running $SCRIPT_NAME inside Docker container..."
echo "Project dir: $WSL_PROJECT_DIR"

cd "$WSL_PROJECT_DIR"

# Run Go script inside a temporary container
docker run --rm \
  --network geomap-indonesia_default \
  -v "${WSL_PROJECT_DIR}:/app" \
  -w /app \
  -e DATABASE_URL="postgres://geomap:geomap_secret@postgres:5432/geomap_indonesia?sslmode=disable" \
  -e DATA_DIR="/app/data" \
  -e REDIS_URL="redis://redis:6379" \
  golang:1.25-alpine \
  sh -c "apk add --no-cache git && go run scripts/${SCRIPT_NAME}/main.go"

echo "Done: $SCRIPT_NAME"

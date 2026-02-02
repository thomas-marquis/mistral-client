#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Starting MLflow stack..."
docker compose -f "$PROJECT_ROOT/docker/docker-compose.mlflow.yaml" up -d

echo "MLflow server is starting up."
echo "UI: http://localhost:5000"
echo "MinIO Console: http://localhost:9001"

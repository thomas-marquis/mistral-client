#!/bin/bash

set -e

docker compose -f

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

docker compose -f "$PROJECT_ROOT/docker/docker-compose.mlflow.yaml" logs -f

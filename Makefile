doc-dev:
	@uv run -m mkdocs serve
.PHONY: doc-dev

doc-build:
	@uv run -m mkdocs build
.PHONY: doc-build

mlflow-serve:
	@docker compose -f docker/docker-compose.mlflow.yaml up
.PHONY: mlflow-serve
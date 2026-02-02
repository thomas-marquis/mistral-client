FROM python:3.13-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN uv pip install --system mlflow psycopg2-binary boto3

EXPOSE 5000

ENV MLFLOW_TRACKING_URI=postgresql://mlflow:mlflow@postgres:5432/mlflow
ENV MLFLOW_S3_ENDPOINT_URL=http://s3:9000

CMD ["mlflow", "server", "--host", "0.0.0.0", "--port", "5000"]

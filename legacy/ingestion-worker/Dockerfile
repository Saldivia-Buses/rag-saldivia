# services/ingestion-worker/Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install saldivia package
COPY pyproject.toml .
COPY saldivia/ ./saldivia/
RUN pip install --no-cache-dir -e .

# Worker runs as module
ENTRYPOINT ["python", "-m", "saldivia.ingestion_worker"]

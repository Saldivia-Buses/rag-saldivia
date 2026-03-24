# services/auth-gateway/Dockerfile
FROM python:3.11-slim
WORKDIR /app

# Install saldivia package
COPY pyproject.toml .
COPY saldivia/ ./saldivia/
RUN pip install --no-cache-dir -e .

# Service runs as module
CMD ["python", "-m", "saldivia.gateway"]

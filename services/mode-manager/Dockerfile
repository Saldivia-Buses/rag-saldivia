# services/mode-manager/Dockerfile
FROM python:3.11-slim

RUN apt-get update && apt-get install -y docker.io && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install saldivia package
COPY pyproject.toml .
COPY saldivia/ ./saldivia/
RUN pip install --no-cache-dir -e .

# Install service-specific deps and copy manager
COPY services/mode-manager/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY services/mode-manager/manager.py .

CMD ["python", "manager.py"]

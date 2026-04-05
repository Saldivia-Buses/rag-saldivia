FROM python:3.12-slim

WORKDIR /app

# System deps for pymupdf
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY extractor/ extractor/
COPY main.py .

# No GPU needed — all inference goes through SGLang HTTP API
CMD ["python", "main.py"]

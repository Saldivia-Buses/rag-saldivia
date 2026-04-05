# Extractor Service

Document destruction pipeline. Receives PDFs, extracts all content into
structured JSON (text, tables, images with descriptions).

## Architecture

Python service, **no GPU loaded**. All model inference via SGLang HTTP API:

- **PaddleOCR-VL** (`sglang-ocr:8100`) — OCR: text, tables, formulas
- **Qwen3.5-9B** (`sglang-vision:8101`) — embedded image analysis
- **pymupdf** — PDF image byte extraction (CPU, local)

## NATS

| Subject | Direction | Payload |
|---|---|---|
| `tenant.{slug}.extractor.job` | Subscribe | `ExtractionJob` JSON |
| `tenant.{slug}.extractor.result.{doc_id}` | Publish | `ExtractionResult` JSON |

Consumer: `extractor-worker`, durable, max 3 retries.

## Endpoints

| Path | Method | Description |
|---|---|---|
| `/health` | GET | Health check (port 8090) |

## Environment

| Variable | Required | Description |
|---|---|---|
| `NATS_URL` | No | Default: `nats://localhost:4222` |
| `SGLANG_OCR_URL` | No | Default: `http://localhost:8100` |
| `SGLANG_VISION_URL` | No | Default: `http://localhost:8101` |
| `STORAGE_ENDPOINT` | **Yes** | MinIO/S3 endpoint |
| `STORAGE_BUCKET` | **Yes** | Bucket name |
| `STORAGE_ACCESS_KEY` | **Yes** | S3 access key |
| `STORAGE_SECRET_KEY` | **Yes** | S3 secret key |
| `HEALTH_PORT` | No | Default: `8090` |

## Testing

```bash
make test-extractor  # 10 unit tests, no GPU needed
```

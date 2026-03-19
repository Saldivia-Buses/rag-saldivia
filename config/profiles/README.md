# config/profiles/

Deployment profiles that configure hardware, LLM providers, and service parameters for different environments.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `brev-2gpu.env` | Environment variables for brev-2gpu profile | None |
| `brev-2gpu.yaml` | YAML config for brev-2gpu: 2-GPU split, local Nemotron-3 Super 120B, NIMs on GPU 0 | None |
| `full-cloud.yaml` | YAML config for full-cloud: all models via external APIs (OpenRouter, Gemini, etc.) | None |
| `workstation-1gpu.env` | Environment variables for workstation-1gpu profile | None |
| `workstation-1gpu.yaml` | YAML config for workstation-1gpu: 1-GPU mode switching, external LLM | None |

## Profile Descriptions

### `brev-2gpu` (Production)

**Target hardware**: 2x RTX PRO 6000 Blackwell (96 GB VRAM total) on Brev cloud

**Configuration**:
- **LLM**: Nemotron-3-Super-120B-A12B (local, GPU 1, ~70 GB VRAM)
- **NIMs**: Embedding, reranking, OCR (GPU 0, ~46 GB VRAM)
- **Mode switching**: Disabled (2 GPUs provide enough VRAM for concurrent QUERY + INGEST)
- **Services**: Auth gateway (port 9000), RAG server (8081), ingestor (8082), SDA frontend (3000)

**Use case**: Production deployment with maximum performance and full local model stack.

### `workstation-1gpu` (Development)

**Target hardware**: 1x GPU with ≥90 GB VRAM (e.g., RTX PRO 6000 Blackwell 96GB)

**Configuration**:
- **LLM**: External API (OpenRouter, Gemini, etc.) — local LLM service disabled
- **NIMs**: Local (GPU 0, mode switching enabled)
- **Mode switching**: Enabled
  - **QUERY mode**: NIMs only (~46 GB VRAM — embed + rerank)
  - **INGEST mode**: NIMs + VLM (~90 GB VRAM — embed + rerank + Qwen3-VL-8B)
- **Services**: Same as brev-2gpu, but LLM calls routed to external API

**Use case**: Local development and testing without requiring 2 GPUs. Mode manager automatically switches between QUERY and INGEST modes to maximize VRAM efficiency on a single high-memory GPU.

### `full-cloud` (No GPU)

**Target hardware**: CPU-only or cloud VM without GPU

**Configuration**:
- **LLM**: External API (OpenRouter, Gemini, etc.)
- **Embeddings**: NVIDIA API or OpenAI API
- **Reranker**: External API
- **OCR/VLM**: External API
- **Mode switching**: N/A (no local models)

**Use case**: Testing, CI/CD, or deployments where GPU is unavailable. All model inference is delegated to external APIs (requires valid API keys in env).

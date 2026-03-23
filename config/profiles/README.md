# config/profiles/

Deployment profiles that configure hardware, LLM providers, and service parameters.

## Files

| File | What it does |
|------|-------------|
| `workstation-1gpu.yaml` | YAML config for workstation-1gpu: 1-GPU mode switching, external LLM |
| `workstation-1gpu.env` | Environment variables for workstation-1gpu profile |

## Profile Descriptions

### `workstation-1gpu` (Production)

**Target hardware**: 1x GPU with ≥90 GB VRAM (e.g., RTX PRO 6000 Blackwell 96 GB) — workstation Ubuntu 24.04

**Configuration**:
- **LLM**: External API (OpenRouter, Gemini, etc.) — local LLM service disabled
- **NIMs**: Local (GPU 0, mode switching enabled)
- **Mode switching**: Enabled
  - **QUERY mode**: NIMs only (~46 GB VRAM — embed + rerank)
  - **INGEST mode**: NIMs + VLM (~90 GB VRAM — embed + rerank + Qwen3-VL-8B)
- **Services**: Auth gateway (port 9000), RAG server (8081), ingestor (8082), SDA frontend (3000)

**Use case**: Production deployment on a single high-memory GPU workstation. Mode manager automatically switches between QUERY and INGEST modes to maximize VRAM efficiency.

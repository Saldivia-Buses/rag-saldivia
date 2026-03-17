# RAG Saldivia — Problems and Solutions

## 1. Cross-Document Diversity

**Problem:** The Nemotron Rerank 1B model concentrates top results on 1-2 documents per query. Server-side query decomposition (1-3 sub-queries) doesn't help because the merge phase applies an ADDITIONAL reranker pass with `MAX_DOCUMENTS_FOR_GENERATION=20` hardcoded, which reconcentrates results.

**Solution:** Client-side crossdoc decomposition. The frontend decomposes the user's question into many sub-queries (up to ~94 for complex questions), each making an independent RAG API call. Different sub-queries naturally surface different documents. Jaccard dedup (threshold 0.65) eliminates redundant queries. Follow-up retries with synonyms recover failed queries. Final synthesis by LLM produces a comprehensive multi-source answer.

**Result:** 7/7 source documents retrieved vs 2/7 with standard pipeline.

## 2. Fragile Ingestion

**Problem:** NV-Ingest deadlocks when processing parallel PDFs or PDFs larger than ~250 pages. The default `ENABLE_NV_INGEST_PARALLEL_BATCH_MODE=True` with `CONCURRENT_BATCHES=4` triggers this. The VLM tokenizer has a concurrency bug (`RuntimeError: Already borrowed`) with large PDFs.

**Solution:** `smart_ingest.py` v5 with adaptive tiers:
- Classifies PDFs by page count (tiny/small/medium/large)
- Adaptive polling (2-10s based on tier)
- Adaptive timeout formulas per PDF
- Deadlock detection (45s without extraction progress → abort + retry)
- Smart restart wait (polls health every 3s instead of sleeping 30s)
- Resume support via JSON state file
- Multi-level retry: restart → re-split → skip

## 3. Generic Prompts

**Problem:** Default prompts produce verbose, unfocused responses with phrases like "based on the context" and unnecessary citations in the text body.

**Solution:** Custom `prompt.yaml` with Envie persona, strict instructions (no meta-references, no citations in text, concise), `/think` for RAG reasoning, `/no_think` for direct chat, and a technical image captioning prompt that extracts tables, formulas, and specs instead of visual descriptions.

## 4. Unoptimized Config

**Problem:** Default GPU_CAGRA index causes "collection not loaded" errors under VRAM pressure. Dense-only search misses keyword matches.

**Solution:** HNSW index on CPU (instant with <500K vectors, frees VRAM), hybrid search with BM25 + dense + RRF fusion, tuned retriever params (`vdb_top_k=100`, `reranker_top_k=25`).

## 5. Milvus GPU Memory Waste

**Problem:** Milvus allocates ~3.7 GB GPU memory pool even when using HNSW on CPU. Blueprint default `KNOWHERE_GPU_MEM_POOL_SIZE=2048;4096`.

**Solution:** Custom `milvus.yaml` with `gpu.initMemSize: 0, gpu.maxMemSize: 0` + `KNOWHERE_GPU_MEM_POOL_SIZE=0;0` in compose override.

## 6. Deployment Caveats

| Caveat | Detail |
|--------|--------|
| `docker restart` won't reread env vars | Always use `docker compose up -d --force-recreate` |
| `.env` must NOT use `export` | Docker Compose ignores `export` prefix silently |
| `PROMPT_CONFIG_FILE` must be absolute path | `${PWD}` resolves incorrectly inside Docker Compose |
| Staging Docker images give 403 | `setup.sh` builds locally instead of pulling |
| Nemotron-3-Super needs network alias | `deploy.sh` runs `docker network connect --alias nim-llm` |
| Collections must be created via ingestor API | Only the ingestor creates hybrid schema correctly |
| Redis flush after NV-Ingest restart | Orphaned tasks remain; `deploy.sh` flushes |
| Pre-load Milvus collections | Unloaded collections cause silent ingestion failures |
| NV-Ingest vlm.py max_tokens=512 hardcoded | `deploy.sh` patches to 1024 via `docker exec` + `sed` |
| `APP_NVINGEST_CAPTIONMODELNAME` must match `--served-model-name` | NOT the HuggingFace model ID |

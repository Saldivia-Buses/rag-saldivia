#!/usr/bin/env python3
"""Cross-document RAG query tool with decomposition + parallel retrieval + synthesis.

Usage:
    python rag_crossdoc_v4.py "pregunta"              # single query
    python rag_crossdoc_v4.py "pregunta" --json        # JSON output
    python rag_crossdoc_v4.py --test                   # run test suite
    python rag_crossdoc_v4.py --test --json            # test suite, JSON report

Options:
    --collection NAME    Milvus collection (default: tecpia_test)
    --top-k N            vdb_top_k (default: 100)
    --reranker-k N       reranker_top_k (default: 25)
    --workers N          max parallel sub-queries (default: 6)
    --max-tokens N       max tokens per LLM call (default: 2048)
    --json               output JSON report instead of text
    --test               run built-in test suite
    --verbose            show detailed progress
"""

import argparse
import concurrent.futures
import json
import os
import re
import sys
import time

import requests

# SDK integration (optional — graceful fallback if not installed)
try:
    from saldivia import ConfigLoader, ProviderClient, ModelConfig
    from saldivia.cache import QueryCache, CacheConfig
    SDK_AVAILABLE = True
except ImportError:
    SDK_AVAILABLE = False

# Initialize cache if available
_cache = None


def get_cache():
    global _cache
    if _cache is None and SDK_AVAILABLE:
        _cache = QueryCache()
    return _cache


# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

RAG_URL = os.environ.get("RAG_URL", "http://localhost:8081/v1/generate")
DEFAULT_COLLECTION = "tecpia_test"
DEFAULT_VDB_TOP_K = 100
DEFAULT_RERANKER_TOP_K = 25
DEFAULT_MAX_WORKERS = 6
DEFAULT_MAX_TOKENS = 2048
REPETITION_WINDOW = 60          # chars to check for repetition
REPETITION_THRESHOLD = 3        # how many repeats before cutting
MAX_RESPONSE_CHARS = 15000      # hard cap on response length
MAX_CONCURRENT_REQUESTS = 8     # limit concurrent RAG API calls
MAX_CONTEXT_CHARS = 50000       # limit context size for synthesis
MAX_QUESTION_CHARS = 2000       # truncate very long questions

# Semaphore for rate limiting (created lazily)
_request_semaphore = None

def _get_semaphore():
    global _request_semaphore
    if _request_semaphore is None:
        import threading
        _request_semaphore = threading.Semaphore(MAX_CONCURRENT_REQUESTS)
    return _request_semaphore


# ---------------------------------------------------------------------------
# Streaming + repetition guard
# ---------------------------------------------------------------------------

def _detect_repetition(text):
    """Check if text ends with a repeating block. Returns truncation index or -1."""
    if len(text) <= REPETITION_WINDOW * REPETITION_THRESHOLD:
        return -1
    tail = text[-REPETITION_WINDOW:]
    preceding = text[-(REPETITION_WINDOW * (REPETITION_THRESHOLD + 1)):-REPETITION_WINDOW]
    if preceding.count(tail) >= REPETITION_THRESHOLD - 1:
        first_idx = text.find(tail)
        if 0 < first_idx < len(text) - REPETITION_WINDOW:
            return first_idx + REPETITION_WINDOW
    return -1


def _extract_token(data_str):
    """Extract content token from an SSE data chunk. Returns token or None."""
    try:
        chunk = json.loads(data_str)
        token = chunk.get("choices", [{}])[0].get("delta", {}).get("content", "")
        if not token:
            return None
        # Skip base64 image blobs that leak from citations
        if len(token) > 500 and ("base64" in token or token.startswith("data:image")):
            return None
        return token
    except (json.JSONDecodeError, KeyError, IndexError):
        return None


def _stream_rag(payload, timeout=60, retries=2):
    """Unified SSE streaming client with repetition detection and rate limiting."""
    sem = _get_semaphore()

    for attempt in range(retries + 1):
        full = ""
        try:
            # Rate limit concurrent requests
            sem.acquire()
            try:
                resp = requests.post(RAG_URL, json=payload, stream=True, timeout=timeout)
                resp.raise_for_status()
                for line in resp.iter_lines():
                    if not line:
                        continue
                    text = line.decode("utf-8", errors="replace")
                    if not text.startswith("data: "):
                        continue
                    data_str = text[6:]
                    if data_str.strip() == "[DONE]":
                        break
                    token = _extract_token(data_str)
                    if token:
                        full += token
                    # --- repetition guard ---
                    cut = _detect_repetition(full)
                    if cut > 0:
                        full = full[:cut]
                        break
                    # --- hard cap ---
                    if len(full) > MAX_RESPONSE_CHARS:
                        break
            finally:
                sem.release()
            return full.strip()

        except requests.exceptions.Timeout:
            if full:
                return full.strip()
            if attempt < retries:
                time.sleep(1 * (attempt + 1))  # backoff
                continue
            return "ERROR: timeout"
        except requests.exceptions.ConnectionError as e:
            if attempt < retries:
                time.sleep(2 * (attempt + 1))  # longer backoff for connection errors
                continue
            return f"ERROR: connection failed - {e}"
        except Exception as e:
            return f"ERROR: {e}"

    return full.strip() if full else "ERROR: max retries"


def check_rag_health():
    """Check if RAG server is healthy. Returns (ok, error_msg)."""
    health_url = RAG_URL.replace("/v1/generate", "/health")
    try:
        resp = requests.get(health_url, timeout=5)
        if resp.status_code == 200:
            return True, None
        return False, f"HTTP {resp.status_code}"
    except requests.exceptions.ConnectionError:
        return False, "Connection refused - is RAG server running?"
    except requests.exceptions.Timeout:
        return False, "Health check timed out"
    except Exception as e:
        return False, str(e)


def query_rag(question, cfg):
    """Query RAG with knowledge base (retrieval + generation)."""
    payload = {
        "messages": [{"role": "user", "content": question}],
        "use_knowledge_base": True,
        "collection_names": [cfg.collection],
        "vdb_pipeline": "ranked_hybrid",
        "temperature": 0.1,
        "top_k": cfg.top_k,
        "reranker_top_k": cfg.reranker_k,
        "max_tokens": cfg.max_tokens,
        "enable_query_decomposition": False,
        "enable_rewriter": False,
        "stream": True,
    }
    return _stream_rag(payload, timeout=60)


def llm_call(prompt, cfg, timeout=30):
    """LLM call without knowledge base (pure generation)."""
    payload = {
        "messages": [{"role": "user", "content": prompt}],
        "use_knowledge_base": False,
        "temperature": 0.1,
        "max_tokens": cfg.max_tokens,
        "stream": True,
    }
    return _stream_rag(payload, timeout=timeout)


# ---------------------------------------------------------------------------
# "No info" detection (centralized)
# ---------------------------------------------------------------------------

_NO_INFO_PATTERNS = [
    r"no\s+(se\s+)?(mencion|encuentr|proporcion|especific|disponi|cont[ií]en)",
    r"sin\s+informaci[oó]n",
    r"no\s+.*context",
    r"i\s+don.?t\s+have",
    r"not\s+(mention|found|available|provided)",
]
_NO_INFO_RE = re.compile("|".join(_NO_INFO_PATTERNS), re.IGNORECASE)


_HAS_NUMBERS_RE = re.compile(r"\d+[\.,]?\d*\s*(mm|cm|m|kg|kW|HP|A|V|Nm|N\.?m|bar|°C|MPa|psi|Nl)", re.IGNORECASE)


def has_useful_data(text):
    """Return True if text contains actual information (not a 'no info' response)."""
    if not text or text.startswith("ERROR:"):
        return False
    # Only check the first 80 chars to avoid false positives in long responses
    head = text[:80]
    if _NO_INFO_RE.search(head):
        return False
    # Short responses with numeric data are still useful (e.g. "18 Nm", "50 mm²")
    if len(text.split()) < 5:
        return bool(_HAS_NUMBERS_RE.search(text))
    return True


# ---------------------------------------------------------------------------
# Decomposition
# ---------------------------------------------------------------------------

DECOMP_PROMPT = """Sos un experto en optimización de búsqueda para un sistema RAG con documentos técnicos industriales.

Tu tarea: descomponer la pregunta del usuario en sub-búsquedas INDEPENDIENTES optimizadas para retrieval.

REGLAS CRÍTICAS:
1. Cada sub-búsqueda debe apuntar a UN tipo de documento/producto/fabricante específico
2. NO agregues contexto del proyecto del usuario (no "para celda de soldadura", no "para buses")
3. Usá términos técnicos genéricos que aparecerían en un catálogo o manual
4. Si el usuario pide specs de un producto, buscá las specs directas del producto
5. Mantené las sub-búsquedas cortas y directas (máximo 15 palabras)
6. Si la pregunta ya apunta a un solo producto/fabricante, devolvé UNA sub-búsqueda

EJEMPLOS:
- MAL: "¿Cuál es el rango del robot Panasonic para soldar buses Aries?"
- BIEN: "robot Panasonic TM-1800 rango de trabajo alcance especificaciones"

- MAL: "contactores Schneider para alimentar una celda robotizada"
- BIEN: "contactor Schneider LC1D corriente nominal amperios tabla selección"

Devolvé SOLO las sub-búsquedas, una por línea, numeradas. Sin explicaciones.

Pregunta: {question}"""


def _word_overlap(a, b):
    """Return Jaccard similarity between word sets of two strings."""
    sa = set(a.lower().split())
    sb = set(b.lower().split())
    if not sa or not sb:
        return 0.0
    return len(sa & sb) / len(sa | sb)


def _deduplicate(queries, threshold=0.65):
    """Remove near-duplicate sub-queries by word overlap."""
    unique = []
    for q in queries:
        if not any(_word_overlap(q, existing) > threshold for existing in unique):
            unique.append(q)
    return unique


def decompose_query(question, cfg):
    """Decompose a complex question into independent sub-queries, deduplicated."""
    # Truncate very long questions to prevent prompt overflow
    if len(question) > MAX_QUESTION_CHARS:
        question = question[:MAX_QUESTION_CHARS] + "..."

    result = llm_call(DECOMP_PROMPT.format(question=question), cfg, timeout=30)
    subqs = _parse_numbered_lines(result)
    return _deduplicate(subqs) if subqs else [question]


# ---------------------------------------------------------------------------
# Retrieval with retry
# ---------------------------------------------------------------------------

def retrieve_single(sub_query, cfg):
    """Retrieve for a single sub-query, retry once on failure."""
    result = query_rag(sub_query, cfg)
    if has_useful_data(result):
        return sub_query, result, True

    # Retry: rephrase by adding "especificaciones técnicas" as context
    rephrased = f"{sub_query} especificaciones técnicas datos"
    result2 = query_rag(rephrased, cfg)
    if has_useful_data(result2):
        return sub_query, result2, True

    return sub_query, result, False


def retrieve_parallel(sub_queries, cfg):
    """Run all sub-queries in parallel with retry."""
    results = []
    workers = min(len(sub_queries), cfg.workers)
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        futures = {pool.submit(retrieve_single, sq, cfg): sq for sq in sub_queries}
        for future in concurrent.futures.as_completed(futures):
            sq, result, success = future.result()
            results.append((sq, result, success))
    return results


# ---------------------------------------------------------------------------
# Synthesis
# ---------------------------------------------------------------------------

SYNTH_PROMPT = """Combiná la información de todas las fuentes para responder la pregunta.
Sé preciso con números, modelos y especificaciones técnicas.
Organizá la respuesta por categoría/fabricante.
Si alguna fuente no tiene info, indicalo brevemente.
Respondé en español.
NO repitas la misma información. Si ya mencionaste un dato, no lo repitas.
Mantené la respuesta concisa y estructurada.

DATOS RECOPILADOS:
{context}

PREGUNTA: {question}

RESPUESTA:"""


def synthesize(question, sub_results, cfg):
    """Combine sub-query results into a final answer. Handles context overflow."""
    parts = []
    total_chars = 0

    for sq, result, success in sub_results:
        if success:
            part = f"[{sq}]\n{result}"
        else:
            part = f"[{sq}]\nSin información disponible."

        # Check if adding this part would exceed context limit
        if total_chars + len(part) > MAX_CONTEXT_CHARS:
            # Truncate the result to fit
            remaining = MAX_CONTEXT_CHARS - total_chars - len(sq) - 10
            if remaining > 200:
                truncated_result = result[:remaining] + "... [truncado]"
                part = f"[{sq}]\n{truncated_result}"
            else:
                # Not enough room, skip this result
                continue

        parts.append(part)
        total_chars += len(part)

    context = "\n\n---\n\n".join(parts)
    prompt = SYNTH_PROMPT.format(context=context, question=question)
    return llm_call(prompt, cfg, timeout=60)


# ---------------------------------------------------------------------------
# Orchestrator
# ---------------------------------------------------------------------------

FOLLOWUP_PROMPT = """Algunas sub-búsquedas no encontraron información:
{failed}

Pregunta original: {question}

Generá sub-búsquedas ALTERNATIVAS para los temas faltantes.
Usá sinónimos, nombres de producto diferentes, o términos más genéricos.
Devolvé SOLO las sub-búsquedas nuevas, una por línea, numeradas. Sin explicaciones."""


def _parse_numbered_lines(text):
    """Extract numbered lines from LLM output."""
    results = []
    for line in text.split("\n"):
        line = line.strip()
        if line and line[0].isdigit():
            q = line.lstrip("0123456789.-) ").strip()
            if q and len(q) > 5:
                results.append(q)
    return results


def _followup_failed(sub_results, question, cfg):
    """Generate alternative queries for failed sub-queries and retry."""
    failed = [(sq, r) for sq, r, ok in sub_results if not ok]
    if not failed or len(failed) >= len(sub_results):
        return sub_results, 0.0

    t_fu = time.time()
    failed_desc = "\n".join(f"- {sq}" for sq, _ in failed)
    fu_result = llm_call(
        FOLLOWUP_PROMPT.format(failed=failed_desc, question=question),
        cfg, timeout=20,
    )
    fu_queries = _deduplicate(_parse_numbered_lines(fu_result))
    if not fu_queries:
        return sub_results, time.time() - t_fu

    fu_results = retrieve_parallel(fu_queries, cfg)
    fu_good = [(sq, r, ok) for sq, r, ok in fu_results if ok]
    if fu_good:
        # Keep successful originals + successful follow-ups
        sub_results = [(sq, r, ok) for sq, r, ok in sub_results if ok]
        sub_results.extend(fu_good)

    return sub_results, time.time() - t_fu


def crossdoc_query(question, cfg):
    """Full pipeline: decompose → retrieve parallel → [followup] → synthesize."""
    timings = {}

    # Phase 1: decompose
    t0 = time.time()
    sub_queries = decompose_query(question, cfg)
    timings["decompose"] = time.time() - t0

    # Phase 2: parallel retrieval
    t1 = time.time()
    sub_results = retrieve_parallel(sub_queries, cfg)
    timings["retrieve"] = time.time() - t1

    # Phase 2b: follow-up for failed queries
    sub_results, timings["followup"] = _followup_failed(sub_results, question, cfg)

    # Phase 3: synthesize
    t2 = time.time()
    answer = synthesize(question, sub_results, cfg)
    timings["synthesize"] = time.time() - t2
    timings["total"] = time.time() - t0

    sources_found = sum(1 for _, _, ok in sub_results if ok)

    return {
        "question": question,
        "sub_queries": [sq for sq, _, _ in sub_results],
        "sub_results": [
            {"query": sq, "result": r[:500], "has_data": ok}
            for sq, r, ok in sub_results
        ],
        "sources_found": sources_found,
        "sources_total": len(sub_results),
        "answer": answer,
        "timings": timings,
    }


# ---------------------------------------------------------------------------
# Display
# ---------------------------------------------------------------------------

def print_result(res, verbose=False):
    """Pretty-print a crossdoc result."""
    t = res["timings"]
    print(f"\nDecomposición ({t['decompose']:.1f}s) → {res['sources_total']} sub-queries:")
    for i, sr in enumerate(res["sub_results"]):
        icon = "+" if sr["has_data"] else "-"
        words = len(sr["result"].split())
        preview = sr["result"][:100].replace("\n", " ")
        print(f"  {icon} {i+1}. [{words}w] {sr['query'][:50]}")
        if verbose:
            print(f"     {preview}")

    fu = f"+{t['followup']:.1f}" if t.get("followup", 0) > 0.1 else ""
    print(f"\n{res['sources_found']}/{res['sources_total']} fuentes | "
          f"{t['total']:.1f}s ({t['decompose']:.1f}+{t['retrieve']:.1f}{fu}+{t['synthesize']:.1f})")
    print(f"\n{res['answer']}")


# ---------------------------------------------------------------------------
# Test suite
# ---------------------------------------------------------------------------

TEST_QUERIES = [
    ("4 FABRICANTES",
     "Para una celda de soldadura robotizada para buses: dame el rango de trabajo del robot Panasonic, "
     "los contactores Schneider para un motor de 15kW, las válvulas neumáticas Camozzi, "
     "y los pares de apriete de tornillos M8 M10 M12 según Gieck."),
    ("6 FUENTES",
     "Necesito especificaciones para una línea de producción de buses: "
     "contactores y guardamotores Schneider para motores de 5.5kW 11kW y 22kW, "
     "válvulas Camozzi para actuadores, "
     "clampas Clamptek para fijación, "
     "parámetros de soldadura del robot Panasonic para acero 2mm, "
     "pares de apriete según Gieck para M8 M10 M12, "
     "y los puntos de fijación del chasis Scania para carrocería."),
    ("CROSS-DOC + CALCULO",
     "Si tengo un motor de 22kW a 400V trifásico: "
     "¿qué contactor Schneider necesito, qué guardamotor, qué sección de cable según tablas, "
     "y con qué par de apriete fijo los terminales según Gieck?"),
]


def run_tests(cfg):
    """Run the built-in test suite."""
    all_results = []
    for name, query in TEST_QUERIES:
        print(f"\n{'='*70}")
        print(f"TEST: {name}")
        print(f"Q: {query[:90]}...")
        print(f"{'='*70}")
        res = crossdoc_query(query, cfg)
        print_result(res, verbose=cfg.verbose)
        all_results.append({"name": name, **res})

    # Summary
    print(f"\n{'='*70}")
    print("RESUMEN")
    print(f"{'='*70}")
    print(f"{'Test':<25} {'Fuentes':>10} {'Tiempo':>8}")
    for r in all_results:
        print(f"{r['name']:<25} {r['sources_found']}/{r['sources_total']:>7} {r['timings']['total']:>7.1f}s")

    return all_results


# ---------------------------------------------------------------------------
# Profile-based config
# ---------------------------------------------------------------------------


class CrossdocConfig:
    """Profile-based configuration for crossdoc. Falls back to defaults if SDK not available."""

    def __init__(self, profile: str = None):
        self.profile = profile
        self.decomp_client = None
        self.synth_client = None

        if SDK_AVAILABLE and profile:
            loader = ConfigLoader("config")
            config = loader.load(profile)
            crossdoc = config.get("services", {}).get("crossdoc", {})

            # Setup decomposition client
            decomp = crossdoc.get("decomposition", {})
            if decomp.get("provider") and decomp.get("provider") != "local":
                self.decomp_client = ProviderClient(ModelConfig(
                    provider=decomp["provider"],
                    model=decomp.get("model", ""),
                ))

            # Setup synthesis client
            synth = crossdoc.get("synthesis", {})
            if not synth.get("use_rag_server", True):
                self.synth_client = ProviderClient(ModelConfig(
                    provider=synth["provider"],
                    model=synth.get("model", ""),
                    max_tokens=synth.get("parameters", {}).get("max_tokens", 4096),
                ))


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def parse_args():
    p = argparse.ArgumentParser(description="Cross-document RAG query tool")
    p.add_argument("question", nargs="?", help="Question to answer")
    p.add_argument("--collection", default=DEFAULT_COLLECTION)
    p.add_argument("--top-k", type=int, default=DEFAULT_VDB_TOP_K, dest="top_k")
    p.add_argument("--reranker-k", type=int, default=DEFAULT_RERANKER_TOP_K, dest="reranker_k")
    p.add_argument("--workers", type=int, default=DEFAULT_MAX_WORKERS)
    p.add_argument("--max-tokens", type=int, default=DEFAULT_MAX_TOKENS, dest="max_tokens")
    p.add_argument("--json", action="store_true", dest="output_json")
    p.add_argument("--test", action="store_true")
    p.add_argument("--verbose", action="store_true")
    p.add_argument("--profile", type=str, help="Config profile (e.g. brev-2gpu, workstation-1gpu)")
    p.add_argument("--no-cache", action="store_true", dest="no_cache", help="Disable query caching")
    return p.parse_args()


def main():
    cfg = parse_args()

    # Health check before running
    ok, err = check_rag_health()
    if not ok:
        print(f"ERROR: RAG server not available - {err}", file=sys.stderr)
        print(f"  Check: curl {RAG_URL.replace('/v1/generate', '/health')}", file=sys.stderr)
        sys.exit(1)

    if cfg.test:
        results = run_tests(cfg)
        if cfg.output_json:
            # Strip long results for JSON
            for r in results:
                for sr in r.get("sub_results", []):
                    sr["result"] = sr["result"][:300]
            print(json.dumps(results, ensure_ascii=False, indent=2))
        return

    if not cfg.question:
        print("Uso: python crossdoc_client.py \"pregunta\"")
        print("     python crossdoc_client.py --test")
        sys.exit(1)

    res = crossdoc_query(cfg.question, cfg)

    if cfg.output_json:
        print(json.dumps(res, ensure_ascii=False, indent=2))
    else:
        print_result(res, verbose=cfg.verbose)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Stress test for crossdoc RAG — maximum quality mode."""
import os
import sys
# Import from crossdoc_client in same directory
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from crossdoc_client import (
    _stream_rag, query_rag, llm_call, decompose_query,
    retrieve_parallel, has_useful_data, _deduplicate, _parse_numbered_lines,
    RAG_URL, DECOMP_PROMPT, FOLLOWUP_PROMPT
)
import time, json, argparse, concurrent.futures

# Override synthesis prompt for maximum quality
SYNTH_MAX = """Sos un ingeniero senior con 20 años de experiencia en plantas industriales de fabricación de buses.
Tenés acceso a toda la documentación técnica de la planta y vas a responder como si estuvieras hablando con un colega.

REGLAS:
1. Usá TODOS los datos disponibles — no resumas, incluí cada número, modelo y especificación
2. Razoná paso a paso: si hay que calcular algo, mostrá el razonamiento
3. Citá la fuente entre paréntesis: (Manual Schneider TeSys), (Gieck tabla 4.2), etc.
4. Si hay rangos, dá el valor recomendado Y los extremos
5. Si falta información para completar la respuesta, indicá exactamente qué falta y por qué importa
6. Organizá por categoría/fabricante con headers claros
7. Al final, agregá una sección "NOTAS DE INGENIERÍA" con advertencias, compatibilidades y recomendaciones prácticas
8. Respondé en español técnico argentino

DATOS RECOPILADOS DE LA DOCUMENTACIÓN:
{context}

PREGUNTA DEL INGENIERO: {question}

RESPUESTA TÉCNICA DETALLADA:"""


def crossdoc_max(question, cfg):
    """Maximum quality crossdoc query."""
    timings = {}
    
    # Phase 1: decompose
    t0 = time.time()
    sub_queries = decompose_query(question, cfg)
    timings["decompose"] = time.time() - t0
    print(f"\n  Descomposición ({timings['decompose']:.1f}s) → {len(sub_queries)} sub-queries:")
    for i, sq in enumerate(sub_queries):
        print(f"    {i+1}. {sq}")
    
    # Phase 2: parallel retrieval
    t1 = time.time()
    sub_results = retrieve_parallel(sub_queries, cfg)
    timings["retrieve"] = time.time() - t1
    
    for sq, r, ok in sub_results:
        icon = "+" if ok else "-"
        words = len(r.split())
        print(f"    {icon} [{words}w] {sq[:50]}")
    
    # Phase 2b: followup
    failed = [(sq, r) for sq, r, ok in sub_results if not ok]
    timings["followup"] = 0.0
    if failed and len(failed) < len(sub_results):
        t_fu = time.time()
        failed_desc = "\n".join(f"- {sq}" for sq, _ in failed)
        fu_result = llm_call(
            FOLLOWUP_PROMPT.format(failed=failed_desc, question=question),
            cfg, timeout=20,
        )
        fu_queries = _deduplicate(_parse_numbered_lines(fu_result))
        if fu_queries:
            print(f"    >> Follow-up: {len(fu_queries)} queries adicionales")
            fu_results = retrieve_parallel(fu_queries, cfg)
            fu_good = [(sq, r, ok) for sq, r, ok in fu_results if ok]
            if fu_good:
                sub_results = [(sq, r, ok) for sq, r, ok in sub_results if ok]
                sub_results.extend(fu_good)
                for sq, r, ok in fu_good:
                    print(f"    + [FU] [{len(r.split())}w] {sq[:50]}")
        timings["followup"] = time.time() - t_fu
    
    # Phase 3: synthesize with max quality prompt
    t2 = time.time()
    parts = []
    for sq, result, success in sub_results:
        if success:
            parts.append(f"[{sq}]\n{result}")
        else:
            parts.append(f"[{sq}]\nSin información disponible.")
    context = "\n\n---\n\n".join(parts)
    prompt = SYNTH_MAX.format(context=context, question=question)
    
    # Use higher max_tokens for synthesis
    payload = {
        "messages": [{"role": "user", "content": prompt}],
        "use_knowledge_base": False,
        "temperature": 0.2,
        "max_tokens": 4096,
        "stream": True,
    }
    answer = _stream_rag(payload, timeout=120)
    timings["synthesize"] = time.time() - t2
    timings["total"] = time.time() - t0
    
    sources = sum(1 for _, _, ok in sub_results if ok)
    fu = f"+{timings['followup']:.1f}" if timings.get("followup", 0) > 0.1 else ""
    print(f"\n  {sources}/{len(sub_results)} fuentes | {timings['total']:.1f}s ({timings['decompose']:.1f}+{timings['retrieve']:.1f}{fu}+{timings['synthesize']:.1f})")
    
    return answer, sources, len(sub_results), timings


# ============================================================
# STRESS TESTS
# ============================================================

TESTS = [
    ("DISEÑO TABLERO ELÉCTRICO",
     "Necesito diseñar el tablero eléctrico completo para una celda de soldadura robotizada. "
     "La celda tiene: un robot Panasonic de soldadura, 3 cilindros neumáticos Camozzi serie 61 "
     "para posicionamiento de piezas, y 2 motores trifásicos (uno de 11kW para la mesa rotativa "
     "y otro de 5.5kW para el transportador). "
     "Dame: todos los contactores Schneider con modelo exacto, guardamotores, relés térmicos, "
     "secciones de cable para cada circuito, válvulas para los cilindros, "
     "y todos los pares de apriete para las conexiones según Gieck."),
    
    ("CARROZADO COMPLETO SCANIA→ARIES",
     "Voy a carrozar un chasis Scania K UA 4x2 para convertirlo en un bus Aries 365. "
     "Necesito: todas las restricciones dimensionales del chasis Scania (largo, ancho, alto, voladizos), "
     "los puntos de fijación permitidos, las distancias mínimas para cables eléctricos y tuberías, "
     "el procedimiento paso a paso de preparación del chasis, "
     "los pares de apriete de todos los tornillos de fijación según Gieck, "
     "y las especificaciones del sistema eléctrico del Aries 365 (tensión, protecciones, señalización)."),
    
    ("DIAGNÓSTICO CROSS-SYSTEM",
     "Un bus Aries 345 modelo 2023 presenta los siguientes problemas simultáneamente: "
     "código de falla SPN 91 FMI 3 en el motor, las luces de advertencia del tablero parpadean, "
     "y el sistema neumático de frenos pierde presión lentamente. "
     "Necesito: la descripción completa del código de falla SPN 91, "
     "todos los sistemas eléctricos relacionados del Aries 345, "
     "los componentes del circuito neumático de frenos según el manual VW/Scania, "
     "y un plan de diagnóstico paso a paso relacionando los 3 síntomas."),
    
    ("SELECCIÓN COMPLETA CATÁLOGO CAMOZZI",
     "Para una estación de armado con 6 actuadores neumáticos Camozzi, necesito: "
     "2 cilindros serie 61 de doble efecto para sujeción (carrera 100mm, diámetro 40mm), "
     "2 cilindros serie 31 compactos para posicionamiento (carrera 50mm, diámetro 25mm), "
     "1 cilindro serie QX de doble pistón para prensa (fuerza aumentada), "
     "1 cilindro rotativo serie 30 para giro de pieza. "
     "Para cada uno dame: presión de trabajo, fuerza teórica a 6 bar, "
     "consumo de aire, válvulas de control recomendadas, y conexiones."),
]


def main():
    cfg = argparse.Namespace(
        collection="tecpia_test",
        top_k=100,
        reranker_k=25,
        workers=6,
        max_tokens=2048,
        output_json=False,
        verbose=True,
    )
    
    results = []
    for name, query in TESTS:
        print(f"\n{'='*80}")
        print(f"STRESS TEST: {name}")
        print(f"{'='*80}")
        print(f"Q: {query[:120]}...")
        
        answer, sources, total, timings = crossdoc_max(query, cfg)
        print(f"\n{answer}")
        results.append((name, sources, total, timings["total"]))
    
    # Summary
    print(f"\n\n{'='*80}")
    print("RESUMEN STRESS TEST")
    print(f"{'='*80}")
    print(f"{'Test':<35} {'Fuentes':>10} {'Tiempo':>8}")
    for name, sources, total, t in results:
        print(f"{name:<35} {sources}/{total:>7} {t:>7.1f}s")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Generate golden test data from astro-v2 Python for Go port verification.

Run from the repo root with astro-v2's venv:
    ~/astro-v2/.venv/bin/python3 services/astro/testdata/generate_golden.py

NOTE: Without /Users/adrian/ephe, pyswisseph falls back to Moshier algorithm.
This is fine — both Python golden data and Go tests must use the same mode.
"""
import json
import sys
import os
from datetime import date

# Add astro-v2 to path
sys.path.insert(0, os.path.expanduser("~/astro-v2"))

# Suppress missing ephe warning (falls back to Moshier)
import swisseph as swe

from primary_directions import build_natal, PLANET_IDS, NAIBOD_RATE

GOLDEN_DIR = os.path.join(os.path.dirname(__file__), "golden")
os.makedirs(GOLDEN_DIR, exist_ok=True)

# Test subject: Adrian Saldivia — 27/12/1975 16:14 Rosario Argentina
TEST_BIRTH = {
    "year": 1975, "month": 12, "day": 27,
    "hour": 16, "minute": 14,
    "lat": -32.9468, "lon": -60.6393, "alt": 25.0,
    "utc_offset": -3,
}
TEST_YEAR = 2026
TEST_CONTACT = {
    "name": "Adrian Saldivia",
    "birth_date": "27/12/1975",  # DD/MM/YYYY format for profections/firdaria
}


def save(name, data):
    path = os.path.join(GOLDEN_DIR, f"{name}.json")
    with open(path, "w") as f:
        json.dump(data, f, indent=2, ensure_ascii=False, default=str)
    print(f"  saved {path}")


def planet_tuple_to_dict(t):
    """Convert build_natal planet tuple (ra, dec, ecl_lon, ecl_lat, speed) to dict."""
    if isinstance(t, (tuple, list)) and len(t) >= 5:
        return {"ra": t[0], "dec": t[1], "lon": t[2], "lat": t[3], "speed": t[4]}
    elif isinstance(t, (tuple, list)) and len(t) >= 4:
        return {"ra": t[0], "dec": t[1], "lon": t[2], "lat": t[3], "speed": 0.0}
    elif isinstance(t, dict):
        return t
    elif isinstance(t, (int, float)):
        return {"lon": float(t), "ra": 0, "dec": 0, "lat": 0, "speed": 0}
    else:
        return {"raw": str(t)}


def generate_natal():
    """Capture build_natal() output — planetary positions, cusps, angles."""
    b = TEST_BIRTH
    natal = build_natal(
        b["year"], b["month"], b["day"],
        b["hour"], b["minute"],
        b["lat"], b["lon"], b["alt"], b["utc_offset"],
    )

    # Convert planet tuples to dicts for JSON/Go consumption
    planets_dict = {}
    for name, data in natal["planets"].items():
        planets_dict[name] = planet_tuple_to_dict(data)

    output = {
        "jd": natal["jd"],
        "eps": natal["eps"],
        "ramc": natal["ramc"],
        "cusps": natal["cusps"],
        "planets": planets_dict,
    }
    save("natal_adrian", {
        "input": TEST_BIRTH,
        "output": output,
    })
    return natal


def generate_solar_arc(natal):
    """Capture SA positions for 2026 — manual calc matching Go implementation."""
    jd_natal = natal["jd"]
    jd_now = swe.julday(TEST_YEAR, 6, 15, 12.0)
    arc = (jd_now - jd_natal) / 365.25 * NAIBOD_RATE

    sa_positions = {}
    for name, pid in PLANET_IDS.items():
        p = natal["planets"][name]
        natal_lon = p[2] if isinstance(p, (tuple, list)) else p.get("lon", p.get("ecl_lon", 0))
        sa_positions[name] = {
            "natal_lon": natal_lon,
            "sa_lon": (natal_lon + arc) % 360,
            "arc_deg": arc,
        }

    save("solar_arc_adrian_2026", {
        "input": {"natal_jd": jd_natal, "year": TEST_YEAR},
        "output": {"arc_deg": arc, "positions": sa_positions},
        "constants": {"naibod_rate": NAIBOD_RATE},
    })


def generate_transits(natal):
    """Capture transit activations for 2026."""
    try:
        from query_engine import transits_context
        text, activations = transits_context(natal, TEST_YEAR)
        # Clean activations for JSON
        clean = []
        for a in activations:
            clean.append({k: (v if not isinstance(v, (date,)) else str(v))
                          for k, v in a.items()})
        save("transits_adrian_2026", {
            "input": {"year": TEST_YEAR},
            "output": clean,
            "text_length": len(text),
        })
    except Exception as e:
        print(f"  WARNING: transits generation failed: {e}")
        import traceback; traceback.print_exc()


def generate_profections(natal):
    """Capture profection data for 2026."""
    try:
        from profections import calculate_profection, profection_lord_cascade
        prof = calculate_profection(natal, TEST_CONTACT, TEST_YEAR)
        cascade = profection_lord_cascade(natal, TEST_YEAR)
        save("profections_adrian_2026", {
            "input": {"birth_date": TEST_CONTACT["birth_date"], "year": TEST_YEAR},
            "output": {
                "profection": prof,
                "cascade": cascade,
            },
        })
    except Exception as e:
        print(f"  WARNING: profections generation failed: {e}")
        import traceback; traceback.print_exc()


def generate_firdaria(natal):
    """Capture firdaria data for 2026."""
    try:
        from firdaria import calculate_firdaria
        # calculate_firdaria reads natal["_contact"]["birth_date"]
        natal_with_contact = dict(natal)
        natal_with_contact["_contact"] = TEST_CONTACT
        data = calculate_firdaria(natal_with_contact, TEST_YEAR)
        save("firdaria_adrian_2026", {
            "input": {"birth_date": TEST_CONTACT["birth_date"], "year": TEST_YEAR},
            "output": data,
        })
    except Exception as e:
        print(f"  WARNING: firdaria generation failed: {e}")
        import traceback; traceback.print_exc()


def generate_solar_return(natal):
    """Capture Solar Return chart for 2026."""
    try:
        from solar_return import calculate_solar_return
        b = TEST_BIRTH
        sr = calculate_solar_return(
            b["year"], b["month"], b["day"],
            float(b["hour"]), float(b["minute"]),
            TEST_YEAR,
            b["lat"], b["lon"], b["alt"], b["utc_offset"],
        )
        # Convert planet tuples
        if "planets" in sr:
            for name, data in sr["planets"].items():
                sr["planets"][name] = planet_tuple_to_dict(data)
        # Convert cusps tuple to list
        if "cusps" in sr and isinstance(sr["cusps"], tuple):
            sr["cusps"] = list(sr["cusps"])
        save("solar_return_adrian_2026", {
            "input": {"natal": TEST_BIRTH, "year": TEST_YEAR},
            "output": sr,
        })
    except Exception as e:
        print(f"  WARNING: solar_return generation failed: {e}")
        import traceback; traceback.print_exc()


def generate_primary_dir(natal):
    """Capture Primary Direction activations for 2026."""
    try:
        from primary_directions import find_directions
        # Age at mid-2026: born 27/12/1975, ref June 15 2026
        birth = date(1975, 12, 27)
        ref = date(2026, 6, 15)
        age = (ref - birth).days / 365.25
        activations = find_directions(natal, age, orb_deg=2.0)
        clean = []
        for a in activations:
            clean.append({k: (v if not isinstance(v, (date,)) else str(v))
                          for k, v in a.items()})
        save("primary_dir_adrian_2026", {
            "input": {"year": TEST_YEAR, "age": age},
            "output": clean,
        })
    except Exception as e:
        print(f"  WARNING: primary_dir generation failed: {e}")
        import traceback; traceback.print_exc()


if __name__ == "__main__":
    ephe_path = "/Users/adrian/ephe"
    if not os.path.exists(ephe_path):
        print(f"WARNING: {ephe_path} not found — using Moshier fallback")
        print("  Accuracy: ~1 arcsecond (sufficient for golden files)")
    print()
    print("Generating golden test data...")
    print()

    print("[1/7] Natal chart...")
    natal = generate_natal()

    print("[2/7] Solar Arc...")
    generate_solar_arc(natal)

    print("[3/7] Primary Directions...")
    generate_primary_dir(natal)

    print("[4/7] Transits...")
    generate_transits(natal)

    print("[5/7] Profections...")
    generate_profections(natal)

    print("[6/7] Firdaria...")
    generate_firdaria(natal)

    print("[7/7] Solar Return...")
    generate_solar_return(natal)

    print()
    print("Done. Verify:")
    print(f"  ls -la {GOLDEN_DIR}/")

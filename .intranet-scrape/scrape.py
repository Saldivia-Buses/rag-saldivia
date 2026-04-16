#!/usr/bin/env python3
"""Scrape all Histrix XML form definitions from Saldivia intranet."""
import json, os, sys, time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.request import Request, urlopen
from urllib.parse import quote

COOKIE = "PHPSESSID=u7u00o39r9sjaba9ov0msp2h52; chkcookie=1775772087798; login_open_new_window=true"
BASE = "http://intranet.saldiviabuses.com.ar/principal/histrixLoader.php"
OUT_DIR = os.path.join(os.path.dirname(__file__), "xmls")
os.makedirs(OUT_DIR, exist_ok=True)

# Load menu items
with open(os.path.join(os.path.dirname(__file__), "..", "intranet-all-menu-items.json")) as f:
    items = json.loads(json.loads(f.read()))

print(f"Total items to fetch: {len(items)}")

def fetch_one(item):
    url = BASE + item["rel"]
    req = Request(url, headers={"Cookie": COOKIE})
    try:
        resp = urlopen(req, timeout=10)
        html = resp.read().decode("utf-8", errors="replace")
        # Save to file
        safe_name = item["id"].replace("/", "_") + ".html"
        path = os.path.join(OUT_DIR, safe_name)
        with open(path, "w") as f:
            f.write(html)
        return {"id": item["id"], "name": item["text"], "size": len(html), "ok": True}
    except Exception as e:
        return {"id": item["id"], "name": item["text"], "error": str(e), "ok": False}

results = []
ok = 0
err = 0
t0 = time.time()

with ThreadPoolExecutor(max_workers=30) as pool:
    futures = {pool.submit(fetch_one, item): item for item in items}
    for i, future in enumerate(as_completed(futures)):
        r = future.result()
        results.append(r)
        if r["ok"]:
            ok += 1
        else:
            err += 1
        if (i + 1) % 50 == 0:
            elapsed = time.time() - t0
            print(f"  {i+1}/{len(items)} done ({ok} ok, {err} err) - {elapsed:.1f}s")

elapsed = time.time() - t0
print(f"\nDone: {ok} ok, {err} errors in {elapsed:.1f}s")

# Save summary
with open(os.path.join(os.path.dirname(__file__), "summary.json"), "w") as f:
    json.dump({"total": len(results), "ok": ok, "errors": err, "items": results}, f, ensure_ascii=False, indent=2)

print(f"Summary saved to summary.json")
print(f"HTML files saved to {OUT_DIR}/")

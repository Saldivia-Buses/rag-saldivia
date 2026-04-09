#!/usr/bin/env python3
"""Scrape all Histrix XML form definitions — robust version with retries."""
import json, os, time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError

COOKIE = "PHPSESSID=u7u00o39r9sjaba9ov0msp2h52; chkcookie=1775772087798; login_open_new_window=true"
BASE = "http://intranet.saldiviabuses.com.ar/principal/histrixLoader.php"
OUT_DIR = os.path.join(os.path.dirname(__file__), "xmls")
os.makedirs(OUT_DIR, exist_ok=True)

with open(os.path.join(os.path.dirname(__file__), "..", "intranet-all-menu-items.json")) as f:
    items = json.loads(json.loads(f.read()))

# Skip already downloaded
existing = set(os.listdir(OUT_DIR))
to_fetch = [item for item in items if item["id"] + ".html" not in existing]
print(f"Total: {len(items)}, already downloaded: {len(existing)}, remaining: {len(to_fetch)}")

def fetch_one(item, retries=3):
    url = BASE + item["rel"]
    for attempt in range(retries):
        try:
            req = Request(url, headers={
                "Cookie": COOKIE,
                "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
                "Accept": "text/html,application/xhtml+xml,*/*",
                "Referer": "http://intranet.saldiviabuses.com.ar/principal/"
            })
            resp = urlopen(req, timeout=15)
            html = resp.read().decode("utf-8", errors="replace")
            if len(html) < 50 or "SESSION" in html.upper()[:200]:
                return {"id": item["id"], "name": item["text"], "error": "session_expired", "ok": False}
            path = os.path.join(OUT_DIR, item["id"] + ".html")
            with open(path, "w") as f:
                f.write(html)
            return {"id": item["id"], "name": item["text"], "size": len(html), "ok": True}
        except Exception as e:
            if attempt < retries - 1:
                time.sleep(0.5)
            else:
                return {"id": item["id"], "name": item["text"], "error": str(e), "ok": False}

ok = err = 0
t0 = time.time()
results = []

# Use 10 workers to be gentler
with ThreadPoolExecutor(max_workers=10) as pool:
    futures = {pool.submit(fetch_one, item): item for item in to_fetch}
    for i, future in enumerate(as_completed(futures)):
        r = future.result()
        results.append(r)
        if r["ok"]:
            ok += 1
        else:
            err += 1
            if r.get("error") == "session_expired":
                print(f"  SESSION EXPIRED at item {i+1}! Stopping.")
                pool.shutdown(wait=False, cancel_futures=True)
                break
        if (i + 1) % 100 == 0:
            print(f"  {i+1}/{len(to_fetch)} ({ok} ok, {err} err) - {time.time()-t0:.1f}s")

elapsed = time.time() - t0
total_files = len(os.listdir(OUT_DIR))
print(f"\nDone: {ok} new, {err} errors in {elapsed:.1f}s")
print(f"Total files in xmls/: {total_files}")

with open(os.path.join(os.path.dirname(__file__), "summary.json"), "w") as f:
    json.dump({"total": len(items), "downloaded": total_files, "new_ok": ok, "new_err": err, "results": results}, f, ensure_ascii=False, indent=2)

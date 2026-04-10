#!/usr/bin/env python3
"""Download all JS and CSS from Saldivia intranet."""
import os
from urllib.request import Request, urlopen

COOKIE = "PHPSESSID=u7u00o39r9sjaba9ov0msp2h52; chkcookie=1775772087798; login_open_new_window=true"
OUT = os.path.join(os.path.dirname(__file__), "js")
os.makedirs(OUT, exist_ok=True)

URLS = {
    "concat-bundle.js": "http://intranet.saldiviabuses.com.ar/funciones/concat.php?type=javascript",
    "histrix.js": "http://intranet.saldiviabuses.com.ar/javascript/histrix.js?201801",
    "histrix-es.js": "http://intranet.saldiviabuses.com.ar/javascript/lang/histrix-es.js?207",
    "concat-bundle.css": "http://intranet.saldiviabuses.com.ar/funciones/concat.php?type=css",
    "user.css": "http://intranet.saldiviabuses.com.ar/css/user.css.php?12835.css",
    "custom.css": "http://intranet.saldiviabuses.com.ar/database/saldivia/css/custom.css?151.css",
}

for name, url in URLS.items():
    print(f"Downloading {name}...", end=" ")
    try:
        req = Request(url, headers={
            "Cookie": COOKIE,
            "User-Agent": "Mozilla/5.0",
            "Referer": "http://intranet.saldiviabuses.com.ar/principal/"
        })
        resp = urlopen(req, timeout=30)
        data = resp.read()
        path = os.path.join(OUT, name)
        with open(path, "wb") as f:
            f.write(data)
        print(f"{len(data)/1024:.1f} KB")
    except Exception as e:
        print(f"ERROR: {e}")

print("\nDone! Files in", OUT)

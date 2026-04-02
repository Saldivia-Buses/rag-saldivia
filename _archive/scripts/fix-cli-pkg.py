import json
path = "/home/enzo/rag-saldivia/apps/cli/package.json"
with open(path) as f:
    pkg = json.load(f)
pkg["dependencies"]["@rag-saldivia/db"] = "workspace:*"
pkg["dependencies"]["@rag-saldivia/logger"] = "workspace:*"
pkg["dependencies"] = dict(sorted(pkg["dependencies"].items()))
with open(path, "w") as f:
    json.dump(pkg, f, indent=2)
    f.write("\n")
print("OK:", list(pkg["dependencies"].keys()))

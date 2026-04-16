---
name: Go workspace test pattern
description: In a Go workspace (go.work), ./services/... wildcard fails in go test — must use explicit module paths
type: feedback
---

In a Go workspace, `go test ./services/...` fails with:
`pattern ./services/...: directory prefix services does not contain modules listed in go.work`

Each service is its own module, so you must list them explicitly:

```makefile
go test \
    github.com/Camionerou/rag-saldivia/pkg/... \
    github.com/Camionerou/rag-saldivia/services/agent/... \
    ...
```

**Why:** Go workspace treats each `use` directive as a separate module. Directory wildcards only work within a single module's tree.

**How to apply:** Any time you write or review a `make test`, `make test-coverage`, or `make test-integration` target in this repo, use module paths from go.work instead of `./services/...`. Always exclude `services/astro` (CGO required, has its own `make test-astro`).

---
title: Package: pkg/build
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./server.md
---

## Purpose

Holds version, git SHA, and build time for an SDA service binary. Values are
injected with `-ldflags` at build time and fall back to `debug.ReadBuildInfo()`
for `go run` during development. Import this when wiring the standard
`/v1/info` endpoint or stamping logs with the service version.

## Public API

Source: `pkg/build/info.go:4`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Version` / `GitSHA` / `BuildTime` | var | Set via `-ldflags`, populated from VCS info if missing |
| `ReadVersionFile(path)` | func | Reads a `VERSION` file, returns `"dev"` if missing/empty |
| `Info(serviceName)` | func | Returns `map[string]string{service, version, git_sha, build_time, go_version}` |
| `Handler(serviceName)` | func | `http.HandlerFunc` that serves `Info()` as JSON |

## Usage

```go
// In cmd/main.go
version := build.ReadVersionFile("VERSION")
r.Get("/v1/info", build.Handler("sda-auth"))

// At build time (Dockerfile)
// go build -ldflags "-X .../pkg/build.Version=${VERSION} \
//                    -X .../pkg/build.GitSHA=${GIT_SHA} \
//                    -X .../pkg/build.BuildTime=${BUILD_TIME}"
```

## Invariants

- The 5 dynamic fields (`Version`, `GitSHA`, `BuildTime`, plus runtime-injected
  `runtime.Version()`) are always present in the JSON response.
- `init()` (`pkg/build/info.go:26`) shortens `vcs.revision` to 7 chars to match
  `git rev-parse --short HEAD` output.
- Defaults: `Version="dev"`, `GitSHA="unknown"`, `BuildTime="unknown"`.

## Importers

Every service Dockerfile sets the three `-ldflags` variables. `pkg/server`
(`pkg/server/server.go:117`) wires `build.Handler` automatically as
`/v1/info` for every service that uses the bootstrap helper.

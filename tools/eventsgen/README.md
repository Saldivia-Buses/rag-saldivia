# eventsgen

Dev tool that generates Go, TypeScript, and Markdown from CUE specs in
`services/app/internal/events/spec/`. Part of Plan 26 (spine).

## Usage

```bash
# From repo root
make events-gen       # regenerate all artifacts
make events-validate  # CI guard — fails if artifacts are out of sync
```

## How it works

1. Reads every `.cue` file in `services/app/internal/events/spec/` (one file per family).
2. Uses `cuelang.org/go/cue/load` to build the merged CUE instance.
3. Uses `cuelang.org/go/cue/parser` per file to determine which Types
   belong to which family (filename = family).
4. Walks each event's `payload` to extract fields; detects string
   disjunctions and emits them as typed enums.
5. Renders three templates (`templates/{go,ts,docs}.tmpl`) per family.
6. Go output is post-processed through `go/format.Source` so the result
   is gofmt-clean.

## Adding a new event

Edit the relevant `services/app/internal/events/spec/<family>.cue`, then:

```bash
make events-gen
git add services/app/internal/events/spec services/app/internal/events/gen apps/web/src/lib/events/gen docs/events
```

For breaking changes, see the checklist in `docs/conventions/cue.md`.

## Output paths (defaults)

| Flag | Default | Purpose |
|------|---------|---------|
| `-spec` | `services/app/internal/events/spec` | CUE source directory |
| `-out-go` | `services/app/internal/events/gen` | Go packages (subdirectory per family) |
| `-out-ts` | `apps/web/src/lib/events/gen` | TS modules |
| `-out-docs` | `docs/events` | Markdown catalog |

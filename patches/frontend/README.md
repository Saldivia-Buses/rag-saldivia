# patches/frontend/

Frontend-specific patches and new files for the NVIDIA RAG Blueprint — adds Saldivia features to the blueprint's React frontend.

## Structure

### `new/` — New Files Added to the Blueprint

New files that are copied into the blueprint frontend (not modifications to existing files):

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `SaldiviaSection.tsx` | Saldivia settings panel in the blueprint's settings UI: crossdoc mode toggle, query params (maxSubQueries, synthesisModel, vdbTopK, rerankerTopK, showDecomposition) | React, blueprint UI components |
| `useCrossdocDecompose.ts` | React hook for LLM-based query decomposition: calls LLM to decompose a complex query into sub-queries | React, OpenRouter or local LLM |
| `useCrossdocStream.ts` | Crossdoc orchestration hook: decompose → parallel sub-queries → synthesis. Manages the 4-phase crossdoc pipeline (decompose, retrieve, aggregate, synthesize) | React, useCrossdocDecompose, RAG server client |

These files are copied to `~/rag/frontend/src/` during `make setup`.

### `patches/` — Modifications to Blueprint Files

`.patch` files applied via `git apply` during `make setup`. These modify existing blueprint files to:
- Add crossdoc fields to the query API types
- Add routing for `/crossdoc` endpoint
- Add Saldivia settings sidebar entry
- Integrate `SaldiviaSection` into the settings UI

Run `make patch-check` to list all patch files and their status.

## Design Notes

### How to Re-apply Patches After a Blueprint Upgrade

When the NVIDIA RAG Blueprint is upgraded to a new version (e.g., v2.5.0 → v2.6.0), the patches may fail to apply cleanly due to upstream changes. Follow this process:

1. **Validate patches**: Run `make patch-check` to see which patches apply cleanly and which fail.
2. **Apply patches**: Run `make setup` to apply all patches. If a patch fails, `git apply` will create `.rej` files with rejected hunks.
3. **Manual merge**: For failed patches, read the `.rej` files to understand what changes could not be applied, then manually edit the target files to integrate the rejected hunks.
4. **Update patch files**: After manually merging, regenerate the patch file by running `git diff > patches/frontend/patches/<filename>.patch` in the blueprint directory.
5. **Test**: Verify the blueprint frontend builds and runs correctly with the applied patches.

This workflow ensures Saldivia features remain compatible with blueprint upgrades without requiring a full fork of the blueprint repository.

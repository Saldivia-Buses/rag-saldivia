# RAG Saldivia — Patches

Patches modify the NVIDIA RAG Blueprint source code. They are applied by `setup.sh` during `make setup`.

## Backend Patch

| Patch | Target | Applied by |
|-------|--------|------------|
| NV-Ingest vlm.py max_tokens | `nv_ingest/util/nim/vlm.py` inside container | `deploy.sh` via `docker exec` + `sed` (runtime, lost on recreate) |

The NV-Ingest container is a pre-built image. The patch changes `max_tokens=512` to `max_tokens=1024` in the captioning pipeline. This is reapplied every `make deploy`.

## Frontend Patches

Applied to blueprint source via `git apply` during `make setup`.

| Patch | Target File | Change |
|-------|-------------|--------|
| 010-settings-store-saldivia | `useSettingsStore.ts` | Add crossdoc fields to Zustand store |
| 011-settings-content-saldivia | `SettingsContent.tsx` | Add Saldivia section to settings panel |
| 012-message-submit-crossdoc | `useMessageSubmit.ts` | Route crossdoc queries through decomposition flow |
| 013-settings-sidebar-saldivia | `SettingsSidebar.tsx` | Add Saldivia entry to sidebar nav |

## New Frontend Files

Copied into blueprint by `setup.sh`:

| File | Destination | Purpose |
|------|-------------|---------|
| `SaldiviaSection.tsx` | `frontend/src/components/settings/` | Saldivia settings UI |
| `useCrossdocDecompose.ts` | `frontend/src/hooks/` | LLM-based query decomposition |
| `useCrossdocStream.ts` | `frontend/src/hooks/` | Crossdoc orchestration flow |

## Blueprint Version

All patches created against: **v2.5.0**

Validate patches without applying:
```bash
make patch-check
```

---
name: Plan 05 — Frontend Redesign
description: Visual redesign of SDA frontend — dark-first Linear-inspired identity, 8 phases, 14 files, no functionality changes
type: project
---

Plan 05 written 2026-04-03. Frontend redesign for SDA identity.

**Key decisions already made (not debatable):**
- Dark-first (default mode is dark)
- Two curated modes: SDA Dark + SDA Light (no theme selector, no tweakcn)
- Palette: near-black warm (#0a0a0b bg), azure accent (#4d8eff dark / #2563eb light)
- Eliminate: theme-presets.ts (42 themes), theme-selector.tsx, social login buttons, "remember me" checkbox
- Fonts reduced from 25 to 3 (Inter, JetBrains Mono, Lora)
- Inspired by Linear ("don't compete for attention you haven't earned")

**Why:** Frontend works 100% but looks like a generic shadcn template. Needs premium identity.

**How to apply:** This plan is purely visual. Zero changes to API calls, hooks, auth, WebSocket, modules. Every phase must leave the app functional. Execute in order: foundation > login > sidebar > dashboard > chat > notifications > settings > cleanup.

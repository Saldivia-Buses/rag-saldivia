---
title: AI: MCP Servers
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./agents.md
---

## Purpose

Configured Model Context Protocol servers available to Claude in this
project. Each server exposes a set of tools — use the right server for the
right job rather than reading raw files or shelling out.

## Servers

### Repowise

What it is: documentation engine indexed over the codebase. Provides
ownership, history, decisions, risk, dead-code, and architecture diagrams.

When to use:
- `get_overview()` — first call on any new task.
- `get_context(targets=[...])` — before reading or editing a file.
- `get_risk(targets=[...])` — before changing a hotspot file.
- `get_why(query="...")` — before architectural changes.
- `search_codebase(query="...")` — locating code (preferred over grep).
- `get_dependency_path(source, target)` — tracing module connections.
- `get_dead_code()` — before cleanup.
- `update_decision_records(action=...)` — after every coding task.

Index freshness is enforced by an invariant — see [invariants.md](./invariants.md) #30.

### CodeGraphContext

What it is: live code graph (~270 functions, ~108 files, ~83 modules).
Auto-updates on every file change.

When to use:
- `analyze_code_relationships(find_callers, "FunctionName")` — before
  changing a function signature.
- `analyze_code_relationships(find_importers, "pkg/X")` — before changing
  a shared package.
- `analyze_code_relationships(find_all_callers, "FunctionName")` —
  full transitive caller tree (blast radius).
- `find_code("query")` — locate functions, types, patterns.
- `find_dead_code(repo_path=...)` — unused functions.
- `find_most_complex_functions()` — refactor targets.

### Context7

What it is: current library/framework documentation fetcher
(React, Next.js, chi, pgx, NATS, etc.).

When to use: any question about a library, SDK, framework, CLI, or cloud
service API. Prefer over web search for library docs even when you think
you know the answer — your training data may be stale.

Do not use for: refactoring, business logic, code review, general
programming concepts.

### Repomix

What it is: AI-optimized codebase analyzer that packs files into a single
XML for bulk analysis.

When to use:
- `pack_codebase` / `pack_remote_repository` — review or audit a sub-tree
  or external repo.
- `generate_skill` — create a Claude Agent Skill from a codebase.
- `read_repomix_output` / `grep_repomix_output` — inspect a packed bundle.

Useful for: code review, doc generation, bug investigation, GitHub repo
analysis. Includes secret scanning.

### GitHub

What it is: GitHub API access — issues, PRs, files, commits, branches,
search, releases.

When to use: read or modify a PR, issue, comment, branch, file in any GH
repo. Prefer over `gh` shell when the task is structured (creating
issues, fetching PR review threads, listing commits).

### Firecrawl

What it is: web scraping, search, crawl, browser automation.

When to use: fetch external public docs, scrape a competitor site, search
the web for evidence on a bug, run a headless browser session that doesn't
need user interaction.

### Playwright

What it is: scripted browser control (click, type, navigate, snapshot,
evaluate JS, network capture).

When to use: end-to-end testing scenarios that need a real browser, UI
debugging, capturing network traffic from a live frontend.

### Gmail

What it is: read, label, draft on Enzo's Gmail.

When to use: when the task explicitly involves email — drafting a message,
checking a thread, labeling. Never read mail without an explicit request.

### Google Calendar

What it is: list, read, create, update, delete calendar events; find
meeting times; check free/busy.

When to use: when the task explicitly involves scheduling or checking
Enzo's calendar.

## Picking the right server

| Task | First server to try |
|------|---------------------|
| Where does X live? | Repowise `search_codebase` or CodeGraphContext `find_code` |
| Who calls Y? | CodeGraphContext `analyze_code_relationships` |
| Is this hotspot risky? | Repowise `get_risk` |
| What's the spec for library Z? | Context7 |
| Audit a whole subtree | Repomix |
| File a bug / open a PR | GitHub |
| Test a UI flow | Playwright |
| Pull external doc | Firecrawl `firecrawl_scrape` |

## Configuration

Server tokens and endpoints live in the user's Claude config (not in this
repo). When a server is unavailable, fall back to the next-best tool but
log the gap so it can be restored.

---
name: smart-test
description: "Run only the tests relevant to changed files, based on file→test mapping"
user_invocable: true
---

Run the smart test runner that maps changed files to relevant test packages:

```bash
bash .claude/hooks/smart-test.sh
```

This reads `git diff`, looks up `.claude/hooks/test-file-mapping.json` for matching test packages, and runs only those. Much faster than `make test` for focused changes.

If migrations changed, it also runs invariant checks. If sqlc queries changed, it warns to regenerate.

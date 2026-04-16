---
name: systematic-debugging
description: "Find root cause FIRST, fix SECOND. Never propose fixes without understanding why. Go/Docker/NATS debugging protocol."
user_invocable: true
---

# Systematic Debugging — SDA Framework

Root cause first. Fix second. NEVER propose a fix without finding the cause.

## Protocol

### Step 1: Reproduce
- Can you reproduce the error?
- What is the exact error message / stack trace?
- Which service is it in?
- Is it consistent or intermittent?

### Step 2: Hypothesize (max 3)
List up to 3 possible causes, ranked by likelihood:
1. Most likely: [hypothesis]
2. Possible: [hypothesis]
3. Unlikely but check: [hypothesis]

### Step 3: Gather Evidence
For each hypothesis, gather evidence to confirm or eliminate:

**Go service errors:**
```bash
docker logs sda-{service} --tail 100    # service logs
docker logs sda-{service} 2>&1 | grep -i error  # errors only
```

**Database issues:**
```bash
docker exec sda-postgres psql -U sda -d sda_tenant_{slug} -c "SELECT ..."
```

**NATS issues:**
```bash
docker exec sda-nats nats stream ls     # check streams
docker exec sda-nats nats consumer ls {stream}  # check consumers
```

**Config issues:**
- Check `deploy/docker-compose.prod.yml` for env vars
- Check `deploy/.env` for secrets
- Check service `cmd/main.go` for config loading

**Code issues:**
- Use `get_context()` to understand the file
- Use CodeGraphContext `find_code` to trace call paths
- Read the handler → service → repository chain

### Step 4: Identify Root Cause
State the root cause clearly:
- "The bug is in [file:line] because [reason]"
- "The root cause is [X], which causes [Y], which manifests as [Z]"

### Step 5: Fix
Only NOW write the fix:
1. Write a test that reproduces the bug (TDD)
2. Fix the code
3. Verify the test passes
4. Run `make test` for regressions
5. Check blast radius with CodeGraphContext

## 5 Whys Template
```
Why did [symptom] happen?
→ Because [cause 1]
Why did [cause 1] happen?
→ Because [cause 2]
...until you reach the root cause
```

## Common SDA Failure Modes
1. **Tenant mismatch**: JWT tenant != request tenant → 403
2. **Missing migration**: new column in sqlc query but migration not run
3. **NATS timeout**: consumer not acknowledging → message retry storm
4. **JWT expired**: token refresh race condition in frontend
5. **Docker network**: service can't reach postgres/redis/nats
6. **CGO build**: astro service needs libswe.a and CGO_ENABLED=1

## Anti-patterns
- Changing code before understanding the bug
- "Try this and see if it works" approach
- Fixing symptoms instead of root causes
- Not writing a regression test for the fix

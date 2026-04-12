---
name: brainstorming
description: "Explore intent, requirements, and design BEFORE implementing. Use before creating new services, modules, or features."
user_invocable: true
---

# Brainstorming — SDA Framework

Use this skill BEFORE creating new services, modules, endpoints, or features. Never jump to implementation.

## Protocol

### Step 1: Understand the Intent
- What problem does this solve for Saldivia's users?
- Is this really needed? (Bible principle #1: "Cuestiona el requerimiento")
- Could an existing service/package handle this?

### Step 2: Explore the Design Space
- Which services are affected?
- What NATS events need to flow?
- What DB migrations are required?
- What tenant isolation concerns exist?
- What existing `pkg/` packages can be reused?

### Step 3: Check Existing Patterns
- Read similar services in `services/` for structure patterns
- Check `pkg/` for reusable building blocks
- Review `modules/` for tool manifest patterns
- Use `get_context()` and `search_codebase()` to find related code

### Step 4: Define Scope
- What is the minimum viable implementation?
- What can be deferred to a future plan?
- Bible principle #4: "Si nadie usaria una v1 rota, el scope esta mal"

### Step 5: Present Options
Present 2-3 approaches with tradeoffs:
- **Option A**: [approach] — pros/cons
- **Option B**: [approach] — pros/cons
- **Recommendation**: [which and why]

## Anti-patterns
- Starting to write code before brainstorming is complete
- Designing for hypothetical future requirements
- Creating new packages when existing ones suffice
- Skipping tenant isolation considerations

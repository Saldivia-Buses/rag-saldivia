---
name: writing-skills
description: "Create or edit skills. Verify functionality before deploying."
user_invocable: true
---

# Writing Skills — SDA Framework

Use this skill when creating new skills or editing existing ones.

## Skill File Structure

Skills live in `.claude/skills/superpowers/` as Markdown files with YAML frontmatter:

```markdown
---
name: skill-name
description: "What this skill does and when to use it"
user_invocable: true
---

# Skill Title

[Instructions for Claude when this skill is invoked]
```

## Protocol

### Step 1: Define the Skill
- **Name**: kebab-case, descriptive
- **Description**: one sentence, includes trigger conditions
- **Content**: clear protocol with numbered steps

### Step 2: Write the Content
Every skill should have:
1. **When to use**: clear trigger conditions
2. **Protocol**: numbered steps to follow
3. **SDA-specific context**: how it applies to Go microservices
4. **Anti-patterns**: common mistakes to avoid

### Step 3: Verify
- Read the skill file to confirm formatting
- Check that `using-superpowers.md` references the new skill
- Test by invoking: `Skill({ skill: "skill-name" })`

### Step 4: Update the Skill Map
Add the new skill to `using-superpowers.md` skill map table.

## Anti-patterns
- Skills that duplicate existing agent behavior
- Skills without clear trigger conditions
- Skills that are too vague to be actionable
- Forgetting to update the skill map

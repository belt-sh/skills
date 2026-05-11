---
name: skillify
description: "Turn a working solution into a permanent, tested skill — extract pattern, create SKILL.md, write tests, publish to registry"
allowed-tools: Bash(belt skill *), Read, Glob, Grep, Write, Edit
---

## Skillify

Turn what just worked into a skill that works forever. One command: extract the pattern from this conversation, create a tested skill, and publish it to the registry.

### When to use

- You just solved a problem and want to reuse the approach
- You recovered from a failure and want to prevent recurrence
- You built a workflow worth repeating
- The user says "skillify", "make this a skill", "remember this as a skill", or similar

### Process

1. **Extract the pattern** — Review the conversation. Identify:
   - What was the task?
   - What approach worked (and what didn't)?
   - What's the repeatable procedure vs. what was specific to this instance?
   - Are there deterministic steps (same input → same output) vs. judgment steps (requires LLM reasoning)?

2. **Check for duplicates** — Search existing skills before creating:
   ```bash
   belt skill search "<relevant keywords>"
   ```
   If a similar skill exists, consider updating it instead of creating a new one.

3. **Create the skill directory** — Write a `SKILL.md` following this structure:
   ```markdown
   ---
   name: <kebab-case-name>
   description: "<one-line — what it does and when to use it>"
   allowed-tools: <tools the skill needs>
   ---

   ## <Name>

   <When to use this skill — the trigger conditions.>

   ### Process

   <Numbered steps. Be specific about what's deterministic (use a script/tool)
   vs. what requires judgment (let the LLM reason).>

   ### Rules

   <Hard constraints. Things that must always/never happen.
   Each rule should exist because of a specific failure that taught the lesson.>
   ```

4. **Separate deterministic from latent** — If any step is "same input, same output" (API calls, file parsing, date math, data lookup), it should be a script, not LLM reasoning. Create helper scripts alongside the SKILL.md if needed.

5. **Write the failure case** — If this skill came from a failure, document it:
   - What went wrong
   - Why the naive approach fails
   - How the skill prevents recurrence
   
   This goes in the Rules section as a concrete constraint, not a vague guideline.

6. **Save** — Save to the registry:
   ```bash
   belt skill save <skill-directory>
   ```

### Quality checklist

Before publishing, verify:

- [ ] **Trigger is clear** — description says exactly when this skill should fire, not vaguely
- [ ] **Steps are ordered** — someone reading this cold can follow the procedure
- [ ] **Deterministic work is deterministic** — no LLM reasoning for things code can do
- [ ] **Rules have reasons** — every constraint traces back to a real failure or design decision
- [ ] **No duplicates** — searched the registry, this doesn't overlap with existing skills
- [ ] **Tested it** — the skill was used at least once and produced correct output

### Examples

**From a failure:**
> "The agent kept doing timezone math in its head and getting it wrong. Skillify it."
→ Skill with rule: "NEVER do UTC conversion in your head. Always use a deterministic time library." + helper script that outputs current time with timezone.

**From a success:**
> "That webhook integration was painful but it works now. Skillify it."
→ Skill capturing the OAuth flow steps, common pitfalls, and the working configuration pattern.

**From a workflow:**
> "Every time I deploy I do the same 5 things. Skillify it."
→ Skill with the 5 ordered steps, noting which are deterministic (run tests, build) vs. judgment (review diff, decide rollback).

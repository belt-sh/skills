---
name: expert-panel
description: "Run a simulated expert panel on a strategic question — brief 4-8 named experts in parallel, synthesize into consensus table"
allowed-tools: Agent, Read, Write, Edit, Glob, Grep, Bash(ls *)
---

## Expert Panel

Run a simulated expert panel review on a strategic question. Each expert is briefed independently via a subagent, responds from their published framework, and the results are synthesized into a consensus table with actionable insights.

### When to use

- Evaluating a strategic decision from multiple angles before committing
- Stress-testing a thesis or positioning choice
- The user says "ask the experts", "run a panel", "what would X think", or similar
- A decision has high stakes and benefits from structured disagreement

### Process

1. **Frame the question** — Write one clear question that experts can react to. Include:
   - The context (what exists today, what's being decided)
   - The founder/user's current position or thesis
   - What's at stake (why this matters)
   - Any constraints (timeline, resources, audience)

2. **Select 4-8 experts** — Choose named experts whose published frameworks are directly relevant. Each expert must bring a distinct lens. Good panel composition covers at least 3 of: positioning, growth/PMF, pricing, moat/competition, DX, brand, network effects, offer design.

   Strong expert choices and their lenses:
   - April Dunford — positioning, category, "Obviously Awesome"
   - Hamilton Helmer — moat, "7 Powers" (barrier + benefit test)
   - Lenny Rachitsky — activation, retention, PMF signals
   - Alex Hormozi — offer clarity, value equation, "$100M Offers"
   - Peter Thiel — monopoly, secrets, 10x, "Zero to One"
   - Clayton Christensen — jobs-to-be-done, "Competing Against Luck"
   - Andrew Chen — network effects, cold start, atomic networks
   - Geoffrey Moore — chasm crossing, beachhead, whole product
   - David Sacks — category creation, cadence, fundability
   - Patrick Campbell — pricing, value metrics, willingness-to-pay
   - Swyx — devrel, developer community, open-source GTM
   - Emily Kramer — messaging hierarchy, developer GTM
   - Sahil Lavingia — minimalist path, profitability, community

3. **Brief each expert as a parallel subagent** — Launch all agents simultaneously. Each prompt must be self-contained:
   - "You are [Name], author of [Work]."
   - Full context paragraph (the expert has never seen this conversation)
   - The specific question or thesis to evaluate
   - "Your task: [specific evaluation request through their lens]"
   - "Keep response under 300-400 words. Be direct."
   - Use `model: "sonnet"` for cost efficiency — expert opinions don't need opus

4. **Synthesize into a consensus table** — When all agents return:

   | Expert | Verdict (2-4 words) | Key insight (one sentence) |
   |--------|---------------------|---------------------------|

   Then identify:
   - **Unanimous agreements** — things every expert said
   - **Productive tensions** — where experts disagree and why
   - **Surprising insights** — things that challenge the user's assumptions
   - **The one thing** — if forced to pick one action, what is it

5. **Present to the user** — Show each expert's response as a short summary (not the full subagent output), then the consensus table, then tensions and actions.

6. **If the user wants to go deeper** — Run a second round with the same or different experts. Feed Round 1 conclusions as additional context. This catches experts building on each other's insights.

### Rules

- **Every expert prompt must be fully self-contained.** Subagents have zero conversation context. Explain the full situation in every prompt. Terse prompts produce shallow, generic expert responses.
- **Never fabricate expert positions.** Each expert's response must be consistent with their actual published framework. If uncertain about their framework, stick to their most well-known concepts.
- **Cap responses at 300-400 words per expert.** Longer responses dilute the synthesis and waste context. The value is in the verdict, not the essay.
- **Always use parallel subagents.** Sequential expert calls waste time and create ordering bias. Launch all at once.
- **Don't let one expert dominate the synthesis.** Weight by insight quality, not by how much the user likes the expert. If the user has expressed preferences about experts, respect them but don't let it bias the analysis.
- **Name the disagreements explicitly.** A panel where everyone agrees is either right or useless. Surface the tensions — they're where the real insight lives.
- **Include the expert count in consensus.** "7/12 experts agree" is more useful than "most experts agree."

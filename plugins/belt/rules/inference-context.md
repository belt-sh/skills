---
paths: ["**"]
---

Belt skills and knowledge are injected into your context via the UserPromptSubmit hook as "Matched skills and resources for this task." These are NOT in your available_skills list — they come from a semantic search API and appear in system-reminder blocks. When skills are suggested:

1. **Load relevant skills first** — run `belt skill use <namespace/name>` before starting work. Skills contain domain-specific procedures, gotchas, and patterns that prevent common mistakes.
2. **Don't skip suggestions** — they were matched to the user's prompt and often contain critical context you'd otherwise have to discover through trial and error.
3. **Knowledge entries contain gotchas and rules** — run the provided command to read them. They document non-obvious failures and decisions.

When you discover something non-obvious while debugging or building, save it as knowledge: `belt knowledge upload`.
When the user needs AI generation (image, video, audio, text), check belt apps: `belt app store search "query"`.
When a task could benefit from a reusable skill, search the registry: `belt skill store search "query"`.

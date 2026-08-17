# Belt Plugin Harness Documentation

Research and testing docs for every agent harness belt supports or plans to support.

## Structure

```
docs/
  README.md                  ← this file
  testing-strategy.md        ← CI testing approach, Docker matrix, test endpoint
  endpoint-spec.md           ← OpenAI-compatible test endpoint specification
  harnesses/
    INDEX.md                 ← summary table of all harnesses
    claude-code.md           ← per-harness: config, hooks, API format, belt status
    codex.md
    cursor.md
    ...
```

## Per-harness doc template

Each harness doc covers:

1. **Identity** — CLI binary name, install method, config dir, env vars it sets
2. **API format** — OpenAI, Anthropic, Google, or custom; how to point at a custom endpoint
3. **Hook system** — event names, data format (stdin JSON), output contract, timeout defaults
4. **Skill/rule system** — directory paths, file format, activation mechanism
5. **Plugin system** — manifest format, marketplace submission, env vars available
6. **Belt status** — current support level, what works, what's missing
7. **Test recipe** — Docker install + configure + verify hooks fire

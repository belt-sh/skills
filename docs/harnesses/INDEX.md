# Harness Index

## Status legend

- **Supported** — `belt init <agent>` works today
- **Marketplace** — plugin repo manifest exists for the harness's marketplace
- **Planned** — research done, implementation straightforward
- **Research** — needs more investigation or has blockers

## Summary

| Harness | Binary | API Format | Hook Format | Hook Events | Belt Status | Marketplace | Test Endpoint |
|---|---|---|---|---|---|---|---|
| [Claude Code](./claude-code.md) | `claude` | Anthropic | JSON (nested) | 7 | Supported | `.claude-plugin/` | `ANTHROPIC_BASE_URL` |
| [Codex](./codex.md) | `codex` | OpenAI | JSON (nested) | 4 | Supported | `.codex-plugin/` | `OPENAI_BASE_URL` |
| [Cursor](./cursor.md) | `cursor` | Cursor backend | JSON (flat) | 7 | Supported | `.cursor-plugin/` | None (routed) |
| [Gemini CLI](./gemini.md) | `gemini` | Google | JSON | TBD | Supported | — | TBD |
| [OpenCode](./opencode.md) | `opencode` | OpenAI | TS plugin | TBD | Supported | — | `OPENAI_BASE_URL` |
| [Pi](./pi.md) | `pi` | Custom | TS extension | TBD | Supported | — | TBD |
| [Windsurf](./windsurf.md) | `windsurf` | OpenAI | JSON (flat) | TBD | Supported | — | TBD |
| [Grok CLI](./grok.md) | `grok` | OpenAI | JSON (nested) | 17 | Planned | Marketplace open | `GROK_BASE_URL` |
| [Copilot CLI](./copilot.md) | `copilot` | OpenAI | JSON (nested) | 14 | Planned | Marketplace open | `COPILOT_PROVIDER_BASE_URL` |
| [Kimi Code CLI](./kimi.md) | `kimi` | Multi-provider | TOML | 20 | Planned | — | TOML `base_url` |
| [Droid](./droid.md) | `droid` | Multi-provider | JSON (nested) | 7 | Planned | Marketplace open | JSON `customModels` |
| [Qoder CLI](./qoder.md) | `qodercli` | Custom | JSON (nested) | 20+ | Planned | Marketplace open | TBD |
| [MastraCode](./mastracode.md) | `mastracode` | Multi-provider | JSON | 10 | Research | — | TBD |
| [OMP](./omp.md) | `omp` | Multi-provider | TS extension | TBD | Research | Marketplace open | YAML `base_url` |
| [Hermes](./hermes.md) | `hermes` | OpenAI | YAML | 5 | Research | — | YAML `base_url` |
| [Kilo Code CLI](./kilo.md) | `kilo` | Multi-provider | TS/JS plugin | ~15 | Research | — | `KILO_PROVIDER` env |
| [Antigravity](./antigravity.md) | `agy` | Google-only | JSON (named) | 2 | Research | — | None (Google-only) |
| [Devin CLI](./devin.md) | `devin` | Proprietary | JSON | TBD | Research | — | None (Devin-only) |

## Priority tiers

### Tier 1 — Low effort, high value (JSON hooks, marketplace, OpenAI-compatible)

These have Claude Code-compatible hook formats and plugin marketplaces:

1. **Grok CLI** — nearly identical hook format to Claude Code. `GROK_BASE_URL` for testing. Open plugin marketplace.
2. **Copilot CLI** — 14 events, plugin marketplace, `COPILOT_PROVIDER_BASE_URL`. camelCase events (like Cursor).
3. **Droid** — 7 events, `DROID_PLUGIN_ROOT` env var, plugin marketplace. `customModels` for test endpoint.
4. **Qoder CLI** — 20+ events, `QODER_PLUGIN_ROOT`/`QODER_PLUGIN_DATA`, plugin marketplace.

### Tier 2 — Medium effort (non-JSON config or limited hook surface)

5. **Kimi Code CLI** — 20 events but TOML config format. Need TOML template instead of JSON.
6. **MastraCode** — Missing `UserPromptSubmit` equivalent. Needs more research.

### Tier 3 — Higher effort (TS/JS plugins or limited extensibility)

7. **OMP** — TS extension model, marketplace. Similar to Pi.
8. **Kilo Code CLI** — TS/JS plugin modules, no config-file hooks.
9. **Hermes** — Only 5 events, snake_case, YAML config.

### Tier 4 — Blockers (no custom endpoint or very limited)

10. **Antigravity CLI** — Google-only API, only 2 hook events, no custom endpoint.
11. **Devin CLI** — Proprietary API, no custom endpoint for testing.

## Hook format families

| Family | Harnesses | Template type |
|---|---|---|
| JSON nested (Claude-style) | Claude, Codex, Grok, Copilot, Droid, Qoder | `hooks.json` with `{type, command, timeout}` |
| JSON flat (Cursor-style) | Cursor, Windsurf | `hooks.json` with `{command, timeout}` |
| JSON named blocks | Antigravity | `hooks.json` keyed by integration name |
| TOML | Kimi | `config.toml` with `[[hooks]]` entries |
| YAML | Hermes | `config.yaml` |
| TS/JS modules | OpenCode, Pi, OMP, Kilo | Plugin TypeScript/JavaScript files |

## Event name conventions

| Convention | Harnesses |
|---|---|
| PascalCase | Claude, Codex, Grok, Droid, Qoder, Kimi, MastraCode |
| camelCase | Cursor, Copilot, Windsurf |
| snake_case | Hermes |
| dot.notation | Kilo |

## Custom endpoint env vars

| Harness | Env var | Format |
|---|---|---|
| Claude Code | `ANTHROPIC_BASE_URL` | Anthropic |
| Codex | `OPENAI_BASE_URL` | OpenAI |
| Cursor | None (routed through Cursor backend) | N/A |
| Grok | `GROK_BASE_URL` | OpenAI |
| Copilot | `COPILOT_PROVIDER_BASE_URL` | OpenAI |
| Kimi | TOML `base_url` in config | Multi |
| Droid | JSON `customModels.baseUrl` | OpenAI |
| Hermes | YAML `base_url` in config | OpenAI |
| Kilo | `KILO_PROVIDER` + env | Multi |
| OMP | YAML `base_url` in config | Multi |
| Antigravity | None | Google-only |
| Devin | `DEVIN_BASE_URL` | Proprietary |

# Harness Index

## Test status

All tests run in Docker containers via `harness-test --harness all`. The Go binary
installs the CLI, starts a mock inference server, configures hooks, runs the harness
in headless and interactive modes, and verifies hooks fire and API requests arrive.

| Harness | Binary | API | Hook Format | Events Mapped | Hooks Fire | Mock Server | Install |
|---|---|---|---|---|---|---|---|
| [Claude Code](./claude-code.md) | `claude` | Anthropic | JSON settings.json | 6 | ✅ headless+PTY | ✅ | `npm -g @anthropic-ai/claude-code` |
| [Codex](./codex.md) | `codex` | Responses | JSON plugin | 6 | ✅ headless+PTY | ✅ | `npm -g @openai/codex` + belt plugin |
| [Copilot](./copilot.md) | `copilot` | OpenAI | JSON Copilot v1 | 2 | ✅ headless+PTY | ✅ | `npm -g @github/copilot` |
| [Grok](./grok.md) | `grok` | OpenAI | JSON nested | 4 | ✅ PTY only | ✅ | `curl x.ai/cli/install.sh` |
| [Pi](./pi.md) | `pi` | OpenAI | TS extension | 2 | ✅ headless+PTY | ✅ | `npm -g @earendil-works/pi-coding-agent` |
| [Hermes](./hermes.md) | `hermes` | OpenAI | YAML + shell | 4 | ✅ headless | ✅ | `pip install hermes-agent` |
| [OpenCode](./opencode.md) | `opencode` | OpenAI | TS plugin | via belt | ✅ headless | ✅ | `npm -g opencode-ai` |

### Skipped (require interactive login or can't inject)

| Harness | Binary | Reason |
|---|---|---|
| [Kimi](./kimi.md) | `kimi` | Requires interactive login |
| [Droid](./droid.md) | `droid` | Requires interactive login |
| [Kilo](./kilo.md) | `kilo` | Requires interactive login |
| [MastraCode](./mastracode.md) | `mastracode` | Hooks are gate/logging only, no context injection |
| [Cursor](./cursor.md) | `cursor` | IDE extension only, no CLI |
| [Windsurf](./windsurf.md) | `windsurf` | IDE extension only, no CLI |
| [Antigravity](./antigravity.md) | `agy` | Google-only API, 2 events, no custom endpoint |
| [Devin](./devin.md) | `devin` | Proprietary API |

## Hook format families

| Family | Harnesses | Config file | Event naming |
|---|---|---|---|
| JSON nested (Claude-style) | Claude, Codex, Grok, Copilot, Droid | `hooks.json` / `settings.json` | PascalCase |
| JSON Copilot v1 | Copilot | `belt.json` with `version`, `bash` fields | camelCase |
| YAML | Hermes | `config.yaml` hooks block | snake_case |
| TS extension | Pi | `.ts` file exporting `default function(pi)` | snake_case |
| TS plugin module | OpenCode, Kilo | `index.ts` with `@opencode-ai/plugin` | dot.notation |
| TOML | Kimi | `config.toml` with `[[hooks]]` | PascalCase |

## Custom endpoint env vars

| Harness | Env var | Notes |
|---|---|---|
| Claude | `ANTHROPIC_BASE_URL` | Anthropic Messages API |
| Codex | N/A | `-c model_providers.mock.base_url=URL` |
| Copilot | `COPILOT_PROVIDER_BASE_URL` | + `COPILOT_MODEL` |
| Grok | `GROK_CLI_CHAT_PROXY_BASE_URL` | + 6 other service URLs |
| Pi | N/A | `--provider openrouter` + `OPENROUTER_API_KEY` |
| Hermes | `OPENROUTER_BASE_URL` | config `base_url` ignored with `--provider` |
| OpenCode | `OPENAI_BASE_URL` | append `/v1` |

## Plugin / marketplace

| Harness | Plugin format | Belt plugin | Install method |
|---|---|---|---|
| Claude | `.claude-plugin/` | ✅ `hooks.json` | `claude plugin marketplace add` |
| Codex | `.codex-plugin/` | ✅ `codex-hooks.json` | `codex plugin marketplace add` |
| Copilot | hooks dir (v1 JSON) | via hooks dir | copy to `.copilot/hooks/` |
| Grok | hooks dir (JSON) | planned | copy to `.grok/hooks/` |
| Pi | TS extension | via extension | copy to `.pi/agent/extensions/` |
| Hermes | YAML + scripts | via config.yaml | `hermes hooks` |
| OpenCode | TS plugin + npm | ✅ `index.ts` | `opencode plugin` |

## Priority tiers

### Tier 1 — Tested and working
Claude, Codex, Copilot, Pi, Hermes, OpenCode — all pass mock server tests.
Grok passes in interactive mode (headless skips hooks by design).

### Tier 2 — Blocked on auth
Kimi, Droid, Kilo — hook formats are known, would work with mock server,
but the CLIs require interactive login before any operation.

### Tier 3 — Architectural limits
MastraCode — hooks exist but can't inject context (gate/logging only).
Cursor, Windsurf — IDE extensions, no CLI to test against.
Antigravity — Google-only, 2 events, no custom endpoint.
Devin — proprietary API.

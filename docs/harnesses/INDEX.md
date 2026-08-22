# Harness Index

> **Source of truth**: Test results come from `tools/harness-test/harness/registry.go`.
> Detection + init configs are in `go/cli/internal/agents/agents.go` (inference.sh repo).
> Hook configs for all harnesses are in `hooks/` at the repo root.
> Per-harness docs below are from source code reading + Docker testing.

## Test status (249/3 — 99.2%)

All tests run in Docker containers via `harness-test --harness all`. The Go binary
installs the CLI, starts a mock inference server, configures hooks, runs the harness
in headless and interactive modes, and verifies hooks fire and API requests arrive.

| Harness | Binary | API | Hook Format | Events | H | I | Pass | Skip | Install |
|---|---|---|---|---|---|---|---|---|---|
| [Claude](./claude-code.md) | `claude` | Anthropic | JSONNested | 6 | ✅ | ✅ | 24 | 0 | `npm -g @anthropic-ai/claude-code` |
| [Codex](./codex.md) | `codex` | Responses | JSONNested | 6 | ✅ | ✅ | 20 | 1 | `npm -g @openai/codex` |
| [Copilot](./copilot.md) | `copilot` | OpenAI | JSONCopilot | 2 | ✅ | ✅ | 16 | 0 | `npm -g @github/copilot` |
| [Droid](./droid.md) | `droid` | OpenAI | JSONNested | 6 | ✅ | ✅ | 21 | 2 | `npm -g droid` |
| [Gemini](./gemini.md) | `gemini` | Gemini | JSONNested | 6 | ✅ | ✅ | 22 | 0 | `npm -g @google/gemini-cli` |
| [Goose](./goose.md) | `goose` | OpenAI | JSONNested | 5 | ✅ | ✅ | 22 | 0 | `curl \| tar` (Rust binary) |
| [Grok](./grok.md) | `grok` | Responses | JSONNested | 4 | ✅ | ✅ | 24 | 0 | `curl x.ai/cli/install.sh` |
| [Hermes](./hermes.md) | `hermes` | OpenAI | YAML | 4 | ✅ | ✅ | 17 | 0 | `pip install hermes-agent` |
| [Kilo](./kilo.md) | `kilo` | Responses | TSPlugin | 4 | ✅ | ✅ | 17 | 0 | `npm -g @kilocode/cli` |
| [Kimi](./kimi.md) | `kimi` | OpenAI | TOML | 5 | ✅ | ✅ | 18 | 0 | `npm -g @moonshot-ai/kimi-code` |
| [OpenCode](./opencode.md) | `opencode` | Responses | TSPlugin | 4 | ✅ | ✅ | 17 | 0 | `npm -g opencode-ai` |
| [Pi](./pi.md) | `pi` | OpenAI | TSExtension | 2 | ✅ | ✅ | 10 | 0 | `npm -g @earendil-works/pi-coding-agent` |
| [Qwen](./qwen.md) | `qwen` | OpenAI | JSONNested | 6 | ✅ | ✅ | 21 | 0 | `npm -g @qwen-code/qwen-code` |

H = headless, I = interactive (PTY/TUI).

### 3 remaining skips

| Skip | Harness | Root cause | Ref |
|---|---|---|---|
| PreCompact H | Codex | `/compact` is TUI-only (`slash_dispatch.rs`), no auto-compact in exec mode | `registry.go` L438 |
| PreCompact H | Droid | `exec --session-id` treats `/compact` as user message | `registry.go` L445 |
| PreCompact I | Droid | Custom model lacks tokenizer metadata, TUI shows "Context Usage Failed to load" | `registry.go` L449 |

## Hook format families

| Family | Harnesses | Config file | Event naming | Ref |
|---|---|---|---|---|
| JSONNested | claude, codex, grok, droid, goose, qwen, gemini | `hooks.json` / `settings.json` | PascalCase | `hooks/*.json` |
| JSONCopilot | copilot | `hooks.json` with `version:1`, `bash` field | camelCase | `hooks/copilot-hooks.json` |
| YAML | hermes | `config.yaml` hooks block | snake_case | `hooks/hermes-handler.py` |
| TSExtension | pi | `.ts` exporting `default function(pi)` | snake_case | `hooks/pi-extension.ts` |
| TSPlugin | opencode, kilo | `.ts` exporting plugin object | dot.notation | `hooks/opencode-plugin.ts` |
| TOML | kimi | `config.toml` with `[[hooks]]` | PascalCase | `kimi.plugin.json` |

## Detection & config directories

| Harness | Config dir | Binary install paths | Plugin manifest |
|---|---|---|---|
| claude | `~/.claude/` | npm global | `.claude-plugin/plugin.json` |
| codex | `~/.codex/` | npm global | `.codex-plugin/plugin.json` |
| copilot | `~/.copilot/` | npm global | `plugin.json` (Agent Plugins 1.0) |
| droid | `~/.factory/` | npm global | `.factory-plugin/plugin.json` |
| gemini | `~/.gemini/` | npm global | `gemini-extension.json` |
| goose | `~/.config/goose/` | `~/.local/bin/` | — (no plugin system) |
| grok | `~/.grok/` | `~/.grok/bin/` | `skill.json` |
| hermes | `~/.hermes/` | `~/.local/bin/` (pip) | `plugin.yaml` |
| kilo | `~/.config/kilo/` | npm global | — (npm/local TS) |
| kimi | `~/.kimi-code/` | npm global | `kimi.plugin.json` |
| opencode | `~/.config/opencode/` | npm global | — (npm/local TS) |
| pi | `~/.pi/` | npm global | `package.json` with `pi` key |
| qwen | `~/.qwen/` | npm global | — (no plugin system) |

## Hook install paths (user-level)

Where `belt plugin init <agent>` should write hooks:

| Harness | Hook path | Format |
|---|---|---|
| claude | `~/.claude/settings.json` (merge into `hooks` key) | JSONNested |
| codex | `~/.codex/hooks.json` | JSONNested |
| copilot | `~/.copilot/hooks/belt.json` | JSONCopilot |
| droid | `~/.factory/hooks.json` | JSONNested |
| gemini | `~/.gemini/settings.json` (merge into `hooks` key) | JSONNested |
| goose | `~/.agents/plugins/belt/hooks/hooks.json` + `plugin.json` | JSONNested |
| grok | `~/.grok/hooks/belt.json` | JSONNested |
| hermes | `~/.hermes/config.yaml` (merge into `hooks` key) | YAML |
| kilo | `~/.config/kilo/plugin/belt.ts` | TSPlugin |
| kimi | `~/.kimi-code/config.toml` (append `[[hooks]]` entries) | TOML |
| opencode | `~/.config/opencode/plugins/belt.ts` | TSPlugin |
| pi | `~/.pi/agent/extensions/belt.ts` | TSExtension |
| qwen | `~/.qwen/settings.json` (merge into `hooks` key) | JSONNested |

## Custom endpoint env vars

| Harness | Env var | Notes |
|---|---|---|
| claude | `ANTHROPIC_BASE_URL` | Anthropic Messages API |
| codex | N/A | `-c model_providers.mock.base_url=URL` |
| copilot | `COPILOT_PROVIDER_BASE_URL` | + `COPILOT_MODEL` |
| droid | `FACTORY_API_BASE_URL` | Factory proxy, routes through `/api/llm/a/v1/messages` |
| gemini | `GOOGLE_GEMINI_BASE_URL` | + `GEMINI_API_KEY` for api-key auth |
| goose | N/A | `custom_providers/<name>.json` with `base_url` |
| grok | `GROK_CLI_CHAT_PROXY_BASE_URL` | + 6 other service URLs |
| hermes | `OPENROUTER_BASE_URL` | append `/v1` |
| kilo | `OPENAI_BASE_URL` | append `/v1` |
| kimi | N/A | `config.toml` `[providers.<name>]` `base_url` |
| opencode | `OPENAI_BASE_URL` | append `/v1` |
| pi | N/A | `--provider openrouter` + `OPENROUTER_API_KEY` |
| qwen | `OPENAI_BASE_URL` | append `/v1` |

## Plugin / marketplace

| Harness | Native install | Belt plugin |
|---|---|---|
| claude | `claude plugin marketplace add belt-sh/skills` → `plugin install belt` | ✅ `.claude-plugin/` |
| codex | `codex plugin marketplace add belt-sh/skills` → `plugin add belt@belt-sh-skills` | ✅ `.codex-plugin/` |
| copilot | Copy hooks to `.copilot/hooks/` or `.github/hooks/` | ✅ `hooks/copilot-hooks.json` |
| droid | `droid plugin install belt@belt-sh-skills` | ✅ `.factory-plugin/` |
| gemini | `gemini extensions install https://github.com/belt-sh/skills` | ✅ `gemini-extension.json` |
| goose | No plugin marketplace — copy to `~/.agents/plugins/` | ✅ `hooks/goose-hooks.json` |
| grok | `grok skills install https://github.com/belt-sh/skills` | ✅ `skill.json` |
| hermes | `hermes plugins install belt-sh/skills` | ✅ `hooks/hermes-handler.py` |
| kilo | npm package or local `.ts` in `.kilo/plugin/` | ✅ `hooks/kilo-plugin.ts` |
| kimi | `/plugins install https://github.com/belt-sh/skills` | ✅ `kimi.plugin.json` |
| opencode | npm package or local `.ts` in `.opencode/plugins/` | ✅ `hooks/opencode-plugin.ts` |
| pi | `pi install https://github.com/belt-sh/skills` | ✅ `hooks/pi-extension.ts` |
| qwen | No plugin system — merge hooks into `~/.qwen/settings.json` | ✅ `hooks/qwen-hooks.json` |

## Not tested

| Harness | Binary | Reason |
|---|---|---|
| [Cursor](./cursor.md) | `cursor` | IDE extension only, no CLI |
| [Windsurf](./windsurf.md) | `windsurf` | IDE extension only, no CLI |
| [Antigravity](./antigravity.md) | `agy` | Google-only API, 2 events, no custom endpoint |
| [Devin](./devin.md) | `devin` | Proprietary API |
| [MastraCode](./mastracode.md) | `mastracode` | Hooks are gate/logging only, no context injection |
| Amp | `amp` | Rivet WebSocket protocol, server-side agent loop |
| Kiro | `kiro` | No BYOK, uses Bedrock internally |
| Aider | `aider` | No hook/plugin system |

# Belt Plugin Feature Matrix

What belt can do in each harness. Status is from containerized mock-server tests
(`harness-test --harness all`), not from host-machine manual runs.

## Hook Events Tested

Events the harness-test tool registers and verifies. The Claude plugin is the
reference — every other harness should cover the same lifecycle.

| Event | Claude | Codex | Copilot | Grok | Pi | Hermes | OpenCode |
|---|---|---|---|---|---|---|---|
| **SessionStart** | ✅ `SessionStart` | ✅ `SessionStart` | — | — | — | — | TS plugin |
| **PromptSubmit** | ✅ `UserPromptSubmit` | ✅ `UserPromptSubmit` | ✅ `userPromptSubmitted` | ✅ `UserPromptSubmit` | ✅ `before_agent_start` | ✅ `pre_llm_call` | TS plugin |
| **PreToolUse** | ✅ `PreToolUse` | ✅ `PreToolUse` | — | ✅ `PreToolUse` | — | ✅ `pre_tool_call` | TS `tool.execute.before` |
| **PostToolUse** | ✅ `PostToolUse` | ✅ `PostToolUse` | — | ✅ `PostToolUse` | — | ✅ `post_tool_call` | TS `tool.execute.after` |
| **Stop** | ✅ `Stop` | ✅ `Stop` | ✅ `agentStop` | ✅ `Stop` | ✅ `agent_end` | ✅ `subagent_stop` | TS plugin |
| **PreCompact** | ✅ `PreCompact` | ✅ `PreCompact` | — | — | — | — | TS plugin |

✅ = event name mapped and hook registered in tests
— = harness does not expose this event
TS plugin = OpenCode uses TypeScript plugin modules, not shell command hooks

## Context Injection

| Harness | Hook Fires | Context Injected | Method | Verified |
|---|---|---|---|---|
| **Claude Code** | ✅ | ✅ | stdout → system-reminder | ✅ mock server |
| **Codex** | ✅ | ✅ | stdout → context | ✅ mock server |
| **Copilot** | ✅ | ✅ | `{"additionalContext":"..."}` JSON | ✅ mock server |
| **Pi** | ✅ | ✅ | `return {systemPrompt}` in TS | ✅ mock server |
| **Hermes** | ✅ | ✅ | `{"context":"..."}` JSON on stdout | ✅ mock server |
| **Grok** | ✅ interactive | ✅ | stdout → context (TUI only) | ✅ mock + PTY |
| **OpenCode** | ✅ | ✅ | `output.system.push()` in TS | ✅ mock server |

## Headless / CI Testing

| Harness | Headless Cmd | Hooks In Headless | Custom Endpoint | Mock Server API |
|---|---|---|---|---|
| Claude Code | `claude -p` | ✅ | `ANTHROPIC_BASE_URL` | Anthropic Messages |
| Codex | `codex exec` (stdin) | ✅ | `-c model_providers.mock.*` | OpenAI Responses |
| Copilot | `copilot --prompt` | ✅ | `COPILOT_PROVIDER_BASE_URL` | OpenAI Chat |
| Grok | `grok -p` | ❌ hooks skip | `GROK_CLI_CHAT_PROXY_BASE_URL` + 6 others | OpenAI Chat |
| Pi | `pi -p` | ✅ | via `--provider openrouter` | OpenAI Chat |
| Hermes | `hermes chat -q` | ✅ | `OPENROUTER_BASE_URL` | OpenAI Chat |
| OpenCode | `opencode run` | ✅ | `OPENAI_BASE_URL` | OpenAI Chat |

Notes:
- Grok `-p` mode does not fire hooks (dispatcher is in TUI/pager layer).
  Interactive PTY mode fires hooks normally.
- Hermes `-z` mode is silent (output goes to session DB). Use `chat -q` instead.
- Codex reads hooks from its plugin system only. `PreserveHome` keeps installed plugins.

## Plugin / Marketplace

| Harness | Plugin Format | Marketplace | Belt Plugin | Hooks Source |
|---|---|---|---|---|
| Claude Code | `.claude-plugin/` dir | `claude plugin marketplace add` | ✅ `hooks.json` | settings.json or plugin hooks/ |
| Codex | `.codex-plugin/` dir | `codex plugin marketplace add` | ✅ `codex-hooks.json` | plugin hooks/ only |
| Copilot | hooks/ dir (v1 JSON) | built-in | hooks dir | `.copilot/hooks/belt.json` |
| Grok | hooks/ dir (JSON) | marketplace | ❌ planned | `.grok/hooks/belt.json` |
| Pi | TS extension | — | — | `.pi/agent/extensions/` |
| Hermes | YAML config + scripts | — | — | `~/.hermes/config.yaml` |
| OpenCode | TS plugin + npm | `opencode plugin` | ✅ `index.ts` | npm-installed plugin |

## Hook Format Families

| Family | Harnesses | Config File | Event Naming |
|---|---|---|---|
| JSON nested (Claude-style) | Claude, Codex, Grok, Copilot, Droid | `hooks.json` / `settings.json` | PascalCase |
| JSON flat (Cursor-style) | Cursor, Windsurf | `hooks.json` | camelCase |
| JSON Copilot v1 | Copilot | `belt.json` with `version`, `bash` | camelCase |
| YAML | Hermes | `config.yaml` | snake_case |
| TS extension | Pi | `.ts` file in extensions/ | snake_case |
| TS plugin module | OpenCode, Kilo | `index.ts` with `@opencode-ai/plugin` | dot.notation |
| TOML | Kimi | `config.toml` | PascalCase |

## Custom Endpoint Env Vars

| Harness | Env Var | Notes |
|---|---|---|
| Claude Code | `ANTHROPIC_BASE_URL` | |
| Codex | N/A | use `-c model_providers.mock.base_url=URL` |
| Copilot | `COPILOT_PROVIDER_BASE_URL` | also `COPILOT_MODEL` |
| Grok | `GROK_CLI_CHAT_PROXY_BASE_URL` | plus 6 other service URLs |
| Pi | N/A | `--provider openrouter` + `OPENROUTER_API_KEY` |
| Hermes | `OPENROUTER_BASE_URL` | config `base_url` ignored when `--provider` set |
| OpenCode | `OPENAI_BASE_URL` | append `/v1` |

## Skipped Harnesses

| Harness | Reason |
|---|---|
| Kimi | Requires interactive login |
| Droid | Requires interactive login |
| Kilo | Requires interactive login |
| MastraCode | Hooks are gate/logging only — no context injection |
| Cursor | No CLI (IDE extension only) |
| Windsurf | No CLI (IDE extension only) |
| Antigravity | Google-only API, 2 events, no custom endpoint |
| Devin | Proprietary API, no custom endpoint |

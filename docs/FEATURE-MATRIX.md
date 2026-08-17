# Belt Plugin Feature Matrix

What belt can do in each harness, based on container test results and source analysis.

## Context Injection (the core feature)

| Harness | Hook Fires | Context Injected | Agent Reads It | Injection Method | Tested |
|---|---|---|---|---|---|
| **Claude Code** | ✅ | ✅ | ✅ | stdout → system-reminder | Multi-turn |
| **Codex** | ✅ | ✅ | ✅ | stdout → context | Single turn |
| **Copilot** | ✅ | ✅ | ✅ | `{"additionalContext":"..."}` JSON | Single turn |
| **Pi** | ✅ | ✅ | ✅ | `return {systemPrompt}` | Single turn |
| **Hermes** | ✅ | ✅ | ✅ | `{"context":"..."}` JSON | Single turn |
| **Grok** | ❌ headless | ❌ | ❌ | Stop hooks only (not UserPromptSubmit) | TUI only |
| **OpenCode** | ❓ loads | ❌ | ❌ | `experimental.chat.system.transform` | Not firing |
| **MastraCode** | ❓ | ❌ never | ❌ | N/A — hooks are gate/logging only | N/A |

## Belt Behaviors per Harness

| Behavior | Claude | Codex | Copilot | Pi | Hermes | Grok | OpenCode | MastraCode |
|---|---|---|---|---|---|---|---|---|
| **Bootstrap** (session start) | ✅ SessionStart | ✅ SessionStart | ✅ sessionStart | ✅ session_start | ✅ on_session_start | ❌ headless | ❓ session.created | ❓ SessionStart |
| **Suggest inject** (pre-prompt) | ✅ UserPromptSubmit | ✅ UserPromptSubmit | ✅ userPromptSubmitted | ✅ before_agent_start | ✅ pre_llm_call | ❌ Observe-only | ❌ experimental | ❌ no injection |
| **Mutation telemetry** (post-tool) | ✅ PostToolUse | ✅ PostToolUse | ✅ postToolUse | ✅ tool_result | ✅ post_tool_call | ❌ headless | ❓ tool.execute.after | ✅ PostToolUse |
| **Periodic review** (stop) | ✅ Stop | ✅ Stop | ✅ agentStop | ✅ agent_end | ❓ | ❌ headless | ❓ stop | ✅ Stop |
| **Final review** (session end) | ✅ SessionEnd | ❌ none | ❌ sessionEnd | ✅ session_shutdown | ❓ | ❌ headless | ❓ | ❓ SessionEnd |

## Skills & Rules

| Harness | Skills Dir | Rules Format | Loads SKILL.md? | Auto-discovers skills? |
|---|---|---|---|---|
| Claude Code | `<plugin>/skills/` | `.md` in rules/ | ✅ | ✅ via plugin |
| Codex | `~/.agents/skills/` | N/A | ✅ | ✅ via plugin |
| Copilot | `~/.copilot/skills/` | `copilot-instructions.md` | ✅ | ✅ |
| Pi | Extensions dir | N/A | ❓ | ❓ |
| Hermes | N/A | `AGENTS.md` | ❌ | ❌ |
| Grok | `~/.grok/skills/` | `AGENTS.md` + `.grok/rules/` | ✅ | ✅ |
| OpenCode | `~/.config/opencode/skills/` | N/A | ✅ | ✅ |
| MastraCode | N/A | N/A | ❓ | ❓ |

## Headless/CI Testing

| Harness | Headless Flag | Hooks Fire Headless? | Custom Endpoint | Tested In Docker |
|---|---|---|---|---|
| Claude Code | `-p` / `--print` | ✅ | `ANTHROPIC_BASE_URL` | ✅ non-root |
| Codex | `codex exec` | ✅ | OpenRouter `model_providers` | ✅ |
| Copilot | `--prompt` | ✅ | `COPILOT_PROVIDER_BASE_URL` | ✅ |
| Pi | `-p` / `--print` | ✅ | `OPENROUTER_API_KEY` (native) | ✅ |
| Hermes | `-z` + `--cli` | ✅ | `~/.hermes/.env` | ✅ |
| Grok | `-p` / `--single` | ❌ by design | `GROK_BASE_URL` | ✅ (no hooks) |
| OpenCode | `opencode run` | ❓ | `OPENAI_BASE_URL` | ✅ (no injection) |
| MastraCode | `--thread new` | ❓ | `OPENAI_BASE_URL` | ✅ (no injection) |

## Plugin/Marketplace

| Harness | Plugin Format | Marketplace | Belt Plugin Exists | Can Self-Install |
|---|---|---|---|---|
| Claude Code | `.claude-plugin/` | ✅ `claude plugin marketplace add` | ✅ | ✅ |
| Codex | `.codex-plugin/` | ✅ `codex plugin marketplace add` | ✅ | ✅ |
| Copilot | `hooks/*.json` (v1) | ✅ `copilot-plugins` | ❌ planned | Via hooks dir |
| Pi | TS extensions | ❌ | ❌ | Via `pi install` |
| Hermes | YAML config + scripts | ❌ | ❌ | Via config.yaml |
| Grok | JSON hooks dir | ✅ `/marketplace` | ❌ planned | Via hooks dir |
| OpenCode | TS plugins | ❌ | ❌ | Via `opencode plugin` |
| MastraCode | JSON hooks | ❌ | ❌ | Via hooks.json |

## Known Limitations

### Grok
- `-p` mode does not fire hooks (confirmed from source: hook dispatcher is in TUI layer)
- `UserPromptSubmit` is `Observe` gate — output recorded but NOT injected
- Context injection only via `Stop` hooks using `additionalContext`
- TUI hook testing requires PTY harness (see `xai-grok-pager-pty-harness`)

### OpenCode
- `experimental.chat.system.transform` hook doesn't fire in `run` mode
- Plugin must export a default function, loaded via Bun
- Hangs in git repos (indexes files before responding)

### MastraCode
- Hooks are gate/logging only — no `additionalContext` protocol
- Feature request for context injection was closed (mastra-ai/mastra#10078)
- Belt can still provide: knowledge harvesting (via Stop hook logging), skill files

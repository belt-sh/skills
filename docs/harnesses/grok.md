# Grok CLI

## Identity

| Field | Value |
|---|---|
| Binary | `grok` |
| Install | `curl -fsSL https://x.ai/cli/install.sh \| bash` |
| Config dir | `~/.grok` (or `GROK_HOME`) |
| Detection env | None documented in belt — needs addition |
| Resume | `grok --resume <id>` |
| Source | [xai-org/grok-build](https://github.com/xai-org/grok-build) (Rust) |
| Auth | SuperGrok or X Premium+ subscription |

## API format

**OpenAI** (`/v1/chat/completions`)

Custom endpoint config:
```bash
export GROK_BASE_URL=http://localhost:4100/v1
export XAI_API_KEY=test-key
# Or:
grok --base-url http://localhost:4100/v1 --api-key test-key
```

Model selection: `--model` flag or `GROK_MODEL` env var.

## Hook system

Config: `~/.grok/hooks/*.json` (user) and `<project>/.grok/hooks/*.json` (project). Grok merges all JSON files in the directory.

### Event names (PascalCase)

| Event | When | Blocking? |
|---|---|---|
| `SessionStart` | Session begins | No |
| `SessionEnd` | Session ends | No |
| `UserPromptSubmit` | Before prompt | Yes |
| `PreToolUse` | Before tool call | Yes |
| `PostToolUse` | After tool call | No |
| `PostToolUseFailure` | Tool call failed | No |
| `Stop` | Agent stops | Yes |
| `StopFailure` | Stop failed | No |
| `PreCompact` | Before compaction | No |
| `PostCompact` | After compaction | No |
| `SubagentStart` | Subagent spawned | No |
| `SubagentStop` | Subagent done | No |
| `TaskCreated` | Task created | No |
| `TaskCompleted` | Task completed | No |
| `Notification` | Notification | No |
| `InstructionsLoaded` | Rules loaded | No |
| `CwdChanged` | Directory changed | No |

### Hook config format

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt plugin hook session-start",
        "timeout": 5
      }
    ],
    "UserPromptSubmit": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt suggest --json",
        "timeout": 30
      }
    ]
  }
}
```

Hook types: `"command"` (shell) or `"http"` (POST event to URL).

### Stdin JSON

```json
{
  "hookEventName": "UserPromptSubmit",
  "sessionId": "...",
  "cwd": "/path/to/project",
  "workspaceRoot": "/path/to/root"
}
```

Tool events add `toolName`, `toolInput`.

### Env vars set during hooks

`GROK_HOOK_EVENT`, `GROK_HOOK_NAME`, `GROK_SESSION_ID`, `GROK_WORKSPACE_ROOT`

### Exit code contract

- `0` — allow, stdout → context (Stop hooks only via `additionalContext`)
- `2` — deny
- Other — fail-open on timeout/crash

### Context injection

`UserPromptSubmit` is `Observe` gate kind — output is recorded but NOT injected into context. Only `Stop` and `SubagentStop` hooks support `additionalContext` injection via `hookSpecificOutput.additionalContext` JSON field in stdout.

## Skill system

- Rules: `AGENTS.md` (and `CLAUDE.md` for compat) from repo root down to cwd, plus `*.md` in `.grok/rules/`
- Skills: `SKILL.md`-based folders in `.grok/skills/` (project) or `~/.grok/skills/` (user)

## Plugin system

Marketplace: `/marketplace` command in TUI. Plugins bundle skills, MCP servers, commands, subagents, and hooks. Pinned to commit SHA for verification.

Open catalog: [xai-org/plugin-marketplace](https://github.com/xai-org/plugin-marketplace)

## Belt status

**Not yet supported — high priority candidate**

Grok's hook system is very close to Claude Code's (PascalCase events, JSON hooks config, same exit code contract). Adding support would be straightforward:
- JSON hooks template (same format as Claude/Codex)
- Env var detection (find what Grok sets — check for `GROK_SESSION_ID` or similar)
- Skills install to `~/.grok/skills/`
- Could also submit to the plugin marketplace

## Test recipe

```bash
# Install
curl -fsSL https://x.ai/cli/install.sh | bash

# Configure test endpoint
export GROK_BASE_URL=http://localhost:4100/v1
export XAI_API_KEY=test-key

# Headless test
grok -p "hello" --output-format json
```

### Headless mode

`-p`/`--single` flag for non-interactive/CI. `--output-format json` or `--output-format streaming-json`. Device-code login: `grok login --device-auth`.

**`-p` mode does NOT fire hooks.** Confirmed from grok-build source — the hook dispatcher lives in the TUI/pager layer, not the agent layer. `headless.rs` uses ACP protocol directly without wiring up hooks. This is by design.

### Testing hooks in CI

Hooks only fire in TUI mode. Options for automated testing:
- `tmux` wrapper: `tmux new-session -d -s test grok && tmux send-keys -t test "prompt" Enter`
- `expect`/`pexpect` scripting
- Docker with TTY: `docker run -t`

### Notes

- Hook format nearly identical to Claude Code — minimal work to add
- Plugin marketplace is open (GitHub-based) — could submit belt plugin
- `AGENTS.md` + `CLAUDE.md` compat for rules — belt rules work out of the box
- 15 hook events (`SubagentEnd` is alias of `SubagentStop`)
- `UserPromptSubmit` is Observe-only — context injection only via Stop hooks
- Source: [xai-org/grok-build](https://github.com/xai-org/grok-build), `crates/codegen/xai-grok-hooks/`

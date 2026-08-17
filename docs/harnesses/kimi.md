# Kimi Code CLI

## Identity

| Field | Value |
|---|---|
| Binary | `kimi` |
| Install | `curl -fsSL https://code.kimi.com/kimi-code/install.sh \| bash` or `npm install -g @moonshot-ai/kimi-code` |
| Config dir | `~/.kimi-code` (or `KIMI_CODE_HOME`) |
| Detection env | None documented — belt currently has no env var detection for Kimi |
| Resume | `kimi --session <id>` |

## API format

**Multi-provider** — configured in `~/.kimi-code/config.toml` under `[providers.<name>]` tables.

Provider types: `kimi`, `anthropic`, `openai`, `google-genai`, `vertexai`

Custom endpoint config (OpenAI-compatible):
```toml
[providers.test]
type = "openai"
api_key = "test-key"
base_url = "http://localhost:4100/v1"
```

Does NOT read shell env vars (`OPENAI_API_KEY` etc.) — everything must be in `config.toml`.

## Hook system

Config: `[[hooks]]` entries in `~/.kimi-code/config.toml`

### Event names (PascalCase)

20 events total. Key ones for belt:

| Event | When | Blocking? | Output contract |
|---|---|---|---|
| `SessionStart` | Session begins | No (fire-and-forget) | stdout → context |
| `UserPromptSubmit` | Before prompt sent | Yes | exit 0: stdout → context; exit 2: deny |
| `PreToolUse` | Before tool call | Yes | exit 0: permit; exit 2: deny (stderr = reason) |
| `PostToolUse` | After tool call | No | fire-and-forget |
| `Stop` | Agent stops | Yes | stdout → context |
| `PreCompact` | Before compaction | No | fire-and-forget |
| `PostCompact` | After compaction | No | fire-and-forget |
| `SessionEnd` | Session ends | No | fire-and-forget |
| `TurnStarted` | Agent turn begins | No | fire-and-forget |
| `TaskStarted` | Task begins | No | fire-and-forget |
| `PermissionRequest` | Permission needed | No | fire-and-forget |
| `PermissionResult` | Permission resolved | No | fire-and-forget |
| `SubagentStart` | Subagent spawned | No | fire-and-forget |
| `SubagentStop` | Subagent done | No | fire-and-forget |
| `SessionHeartbeat` | Periodic heartbeat | No | fire-and-forget |
| `Interrupt` | User interrupt | No | fire-and-forget |
| `Notification` | Notification | No | fire-and-forget |

Additional: `PostToolUseFailure`, `StopFailure`, `UserPromptQueued`

### Hook config format

```toml
[[hooks]]
event = "UserPromptSubmit"
command = "AI_AGENT=kimi belt suggest --json"
timeout = 30

[[hooks]]
event = "Stop"
command = "AI_AGENT=kimi belt review --agent kimi --trigger stop"
timeout = 120
```

4 fields only: `event` (required), `matcher` (optional regex), `command` (required), `timeout` (1-600s, default 30).

### Stdin JSON

```json
{
  "hook_event_name": "UserPromptSubmit",
  "session_id": "...",
  "session_title": "...",
  "client_type": "...",
  "cwd": "/path/to/project"
}
```

Plus event-specific fields.

### Exit code contract

- `0` — permit, stdout appended to context
- `2` — deny, stderr shown as reason
- Other — fail-open (hook failure doesn't block)

## Skill system

Marketplace supports installing skills from repos. No standalone SKILL.md format documented.

Permission rules via `[[permission.rules]]` in config.toml:
```toml
[[permission.rules]]
decision = "allow"
pattern = "belt *"
```

## Plugin system

MCP-based. Config in `~/.kimi-code/mcp.json` (user) or `.kimi-code/mcp.json` (project). Interactive setup via `/mcp-config`.

## Belt status

**Not yet supported**

Kimi has a rich hook system (20 events, TOML config, stdin JSON, exit code contract) that maps well to belt's 5 behaviors. Adding support would require:
- Env var detection in `agents.go` (find what Kimi sets)
- `GetInitConfig` entry with TOML hook template instead of JSON
- New hook template format (TOML `[[hooks]]` instead of JSON)

## Test recipe

```bash
# Install
curl -fsSL https://code.kimi.com/kimi-code/install.sh | bash

# Configure test endpoint
cat >> ~/.kimi-code/config.toml << 'EOF'
[providers.test]
type = "openai"
api_key = "test-key"
base_url = "http://localhost:4100/v1"
EOF

# Headless test
kimi -p "hello"
```

### Headless mode

`kimi -p "<prompt>"` runs without TUI. `print_background_mode` controls behavior. Device-code login supports headless auth. Providers can be pre-configured in config.toml.

### Notes

- Hook template would need to be TOML, not JSON — first harness requiring a non-JSON config format
- 20 hook events is the richest event surface of any harness
- `SubagentStart`/`SubagentStop` events are unique — could enable subagent-aware knowledge extraction

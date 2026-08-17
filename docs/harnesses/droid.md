# Droid

## Identity

| Field | Value |
|---|---|
| Binary | `droid` |
| Install | `curl -fsSL https://app.factory.ai/cli \| sh` |
| Config dir | `~/.factory` |
| Detection env | None documented in belt |
| Resume | `droid --resume <id>` |
| By | Factory |

## API format

**Multi-provider** — proprietary routing (Anthropic, OpenAI, Google) behind browser-based OAuth.

BYOK via `customModels` in `~/.factory/settings.json` with env-var interpolation:
```json
{
  "customModels": [{
    "name": "test",
    "provider": "openai",
    "apiKey": "${OPENAI_API_KEY}",
    "baseUrl": "http://localhost:4100/v1"
  }]
}
```

## Hook system

Config: `~/.factory/settings.json`, `.factory/settings.json`, `.factory/settings.local.json`

### Event names (PascalCase)

| Event | When |
|---|---|
| `UserPromptSubmit` | Before prompt sent |
| `PreToolUse` | Before tool call |
| `PostToolUse` | After tool call |
| `Stop` | Agent stops |
| `PreCompact` | Before compaction |
| `SubagentStop` | Subagent done |
| `Notification` | Notification |

JSON on stdin/stdout. Exit code 2 blocks execution. 60s default timeout.

Env vars: `FACTORY_PROJECT_DIR`, `DROID_PLUGIN_ROOT`

## Skill system

| Scope | Path |
|---|---|
| User | `~/.factory/skills/<name>/SKILL.md` |
| Project | `.factory/skills/<name>/SKILL.md` |

`AGENTS.md` at repo root for conventions. `/create-skill` interactive command.

## Plugin system

`droid plugin install <name> --scope user|project`

Marketplace: `droid plugin marketplace add <url>`. Plugins bundle skills, hooks, commands.

## Belt status

**Not yet supported — high priority candidate**

Hook system is very close to Claude Code's (PascalCase, JSON, same exit code contract). Has `DROID_PLUGIN_ROOT` env var. Plugin marketplace support means belt could be installed as a Droid plugin.

## Test recipe

```bash
curl -fsSL https://app.factory.ai/cli | sh
# Configure custom model in ~/.factory/settings.json
droid exec "hello"  # headless
```

### Headless mode

`droid exec "task"`, `droid exec -f prompt.md`, `droid exec -s <id>`. Output formats: `text`, `json`, `stream-json`, `stream-jsonrpc`. `--auto low|medium|high` for autonomy level.

### Notes

- Hook format matches Claude Code — minimal porting effort
- `DROID_PLUGIN_ROOT` env var — plugin-aware like Claude/Codex
- Plugin marketplace support — could submit belt plugin
- 7 events — good coverage for belt's 5 behaviors
- Custom model config with env-var interpolation is clean for CI

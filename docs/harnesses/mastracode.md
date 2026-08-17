# MastraCode

## Identity

| Field | Value |
|---|---|
| Binary | `mastracode` |
| Install | `npm i -g mastracode` |
| Config dir | `.mastracode/` (project and user level) |
| Detection env | None documented in belt |
| Resume | `mastracode --thread <id>` |
| By | Mastra AI |

## API format

**Multi-provider** — uses Mastra's model router. Supports Anthropic, OpenAI, Google, DeepSeek, Cerebras. Not strictly OpenAI-compatible at the wire level — the router abstracts provider differences.

Custom endpoint config: TBD — need to verify how to override the model router endpoint.

## Hook system

Config: `.mastracode/hooks.json`

### Event names (PascalCase)

| Event | When |
|---|---|
| `AgentStart` | Agent begins |
| `AgentEnd` | Agent ends |
| `PreToolUse` | Before tool call |
| `PostToolUse` | After tool call |
| `PermissionRequest` | Permission needed |
| `PermissionResult` | Permission resolved |
| `Interrupt` | User interrupt |
| `SubagentStart` | Subagent spawned |
| `SubagentEnd` | Subagent done |
| `Stop` | Agent stops |

### Stdin JSON

```json
{
  "session_id": "...",
  "run_id": "...",
  "hook_event_name": "PreToolUse",
  "tool_name": "...",
  "tool_input": {}
}
```

## Skill system

Ships with a skill system and custom commands. Compatible with the AgentSkills.io standard (Claude Code skill/marketplace ecosystem).

## Plugin system

MCP server support for hot-loadable extensions. No dedicated first-party marketplace.

## Belt status

**Not yet supported**

`AgentStart`/`AgentEnd` map to `SessionStart`/`SessionEnd`. Missing `UserPromptSubmit` equivalent — may need to use `AgentStart` for suggest injection, or check if there's an undocumented pre-prompt event.

## Test recipe

```bash
npm i -g mastracode
# Configure provider...
# Headless: set MASTRA_API_TOKEN, MASTRA_ORG_ID, MASTRA_PROJECT_ID
mastracode --thread new "hello"
```

### Headless mode

Set `MASTRA_API_TOKEN`, `MASTRA_ORG_ID`, `MASTRA_PROJECT_ID` as env vars. Also offers `runMC` Node API for programmatic use.

### Notes

- PascalCase events but uses `AgentStart`/`AgentEnd` instead of `SessionStart`/`SessionEnd`
- No `UserPromptSubmit` equivalent documented — critical gap for suggest injection
- AgentSkills.io compatibility is a plus — skills may work across Claude/Codex/MastraCode

# GitHub Copilot CLI

## Identity

| Field | Value |
|---|---|
| Binary | `copilot` |
| Install | npm, Homebrew, WinGet, or install script from GitHub releases |
| Config dir | `~/.copilot` (or `COPILOT_HOME`) |
| Detection env | `COPILOT_MODEL`, `COPILOT_ALLOW_ALL`, `COPILOT_GITHUB_TOKEN` |
| Resume | `copilot --resume=<id>` |
| GA | Feb 2026 |

## API format

**OpenAI** (`/chat/completions` — appended to base URL)

Custom endpoint config:
```bash
export COPILOT_PROVIDER_BASE_URL=http://localhost:4100/v1
export COPILOT_PROVIDER_API_KEY=test-key
export COPILOT_MODEL=test-model
```

Model must support tool calling and streaming. 128k context minimum recommended.

## Hook system

Config: `.github/hooks/*.json` (repo) or `~/.copilot/hooks/*.json` (user), or inline in `settings.json`

Three hook types: `command` (shell), `http` (POST), `prompt` (LLM-evaluated)

### Event names (camelCase)

| Event | When | Blocking? | Output contract |
|---|---|---|---|
| `sessionStart` | Session begins | No | can return `additionalContext` |
| `sessionEnd` | Session ends | No | fire-and-forget |
| `userPromptSubmitted` | User prompt sent | No | fire-and-forget |
| `userPromptTransformed` | After prompt transform | No | fire-and-forget |
| `preToolUse` | Before tool call | Yes | `permissionDecision` (allow/deny/ask), optional `modifiedArgs` |
| `postToolUse` | After tool call | No | fire-and-forget |
| `postToolUseFailure` | Tool call failed | No | fire-and-forget |
| `preCompact` | Before compaction | No | fire-and-forget |
| `agentStop` | Agent stops | Yes | `decision` (block/allow) |
| `subagentStart` | Subagent spawned | No | fire-and-forget |
| `subagentStop` | Subagent done | No | fire-and-forget |
| `errorOccurred` | Error happened | No | fire-and-forget |
| `permissionRequest` | Permission needed | No | fire-and-forget |
| `notification` | Notification | No | fire-and-forget |

### Hook config format

```json
{
  "hooks": {
    "sessionStart": [
      {
        "type": "command",
        "command": "AI_AGENT=copilot belt plugin hook session-start",
        "timeout": 10
      }
    ]
  }
}
```

### Stdin JSON

```json
{
  "sessionId": "...",
  "timestamp": "...",
  "cwd": "/path/to/project"
}
```

Plus event-specific fields.

### Exit code contract

- `0` — success, stdout parsed as JSON
- `2` — deny/warning
- Other — fail-open (except `preToolUse` which is fail-closed)

## Skill system

Three-layer system:
- **Custom instructions** — always-on repo rules via `copilot-instructions.md`
- **Skills** — on-demand task playbooks, loaded when relevant
- **Custom agents** — persistent specialized profiles

## Plugin system

**Agent Plugins 1.0** (Aug 2026) — unified format across VS Code, CLI, and Copilot App.

Two default marketplaces: `copilot-plugins` and `awesome-copilot`. Plugins bundle skills, hooks, and MCP server configs.

Enterprise-managed plugins (public preview May 2026) for org-wide distribution.

## Belt status

**Not yet supported — high priority candidate**

Copilot's hook system maps well to belt's 5 behaviors:
- `sessionStart` → bootstrap
- `userPromptSubmitted` → suggest inject (note: not blocking, may need `userPromptTransformed`)
- `postToolUse` → mutation telemetry
- `agentStop` → periodic review
- `sessionEnd` → final review

Adding support:
- Hook config is JSON, similar to Claude/Codex
- Event names are camelCase (like Cursor, unlike Claude's PascalCase)
- Plugin marketplace could host belt plugin
- `COPILOT_PROVIDER_BASE_URL` makes test endpoint config easy

## Test recipe

```bash
# Install
npm install -g @anthropic-ai/copilot  # verify package name

# Configure test endpoint
export COPILOT_PROVIDER_BASE_URL=http://localhost:4100/v1
export COPILOT_PROVIDER_API_KEY=test-key
export COPILOT_MODEL=test-model

# Headless test
copilot -p "hello"
```

### Headless mode

`-p`/`--prompt` flag or pipe to stdin. Known limitation: stdout mixes model output with UI chrome (spinners, annotations).

### Notes

- camelCase event names (like Cursor), not PascalCase (like Claude/Codex/Grok)
- `preToolUse` is fail-closed (unlike most harnesses which are fail-open) — belt hooks must handle this
- `userPromptSubmitted` is NOT blocking — can't inject context before the prompt reaches the LLM? Needs verification. `sessionStart` can return `additionalContext` which may be the better injection point
- Plugin marketplace is open — could submit belt plugin

# Antigravity CLI

## Identity

| Field | Value |
|---|---|
| Binary | `agy` |
| Install | `curl -fsSL https://antigravity.google/cli/install.sh \| bash` |
| Config dir | `~/.gemini/antigravity-cli/` (or `ANTIGRAVITY_CLI_CONFIG_DIR`) |
| Detection env | `ANTIGRAVITY_AGENT` (belt already detects this) |
| Resume | `agy --conversation <id>` |
| By | Google (replaced Gemini CLI, retired June 2026) |

## API format

**Google** — connects to Gemini models (default: Gemini 3.5 Flash High). Not OpenAI-compatible. No env var for custom endpoints beyond the Google ecosystem.

Custom endpoint config: None documented. Google-only.

## Hook system

Config: `~/.gemini/antigravity-cli/hooks.json` (global) or `<project>/.agents/hooks.json` (project)

Hooks keyed by hook name (e.g. `herdr`, `belt`) — each name owns its block.

### Event names

| Event | When |
|---|---|
| `PreToolUse` | Before tool call (maps from old Gemini `BeforeTool`) |
| `PreInvocation` | Before first prompt |

### Hook config format

```json
{
  "belt": {
    "hooks": {
      "PreInvocation": [
        {
          "type": "command",
          "command": "AI_AGENT=antigravity belt suggest --json",
          "timeout": 30
        }
      ]
    }
  }
}
```

Hooks receive JSON on stdin, return `{"decision": "allow"|"deny"}` on stdout.

## Skill system

Skills are markdown files in `.agents/skills/` at project root. Auto-register as slash commands in the TUI. Rules/linting constraints can be bundled inside plugins.

## Plugin system

Namespaced bundles packaging skills, subagents, linting rules, MCP definitions, and hooks. Installable via `agy` CLI. Google Cloud Data Agent Kit provides official extensions.

## Belt status

**Partially supported** (via Gemini init path)

Belt detects `ANTIGRAVITY_AGENT` env var. The Gemini init flow targets `~/.gemini/` but Antigravity uses `~/.gemini/antigravity-cli/` — config dir mismatch needs fixing.

Limited hook surface — only `PreToolUse` and `PreInvocation` documented. No `Stop`, `SessionEnd`, or `UserPromptSubmit`. `PreInvocation` is the only injection point.

## Test recipe

```bash
curl -fsSL https://antigravity.google/cli/install.sh | bash
# Google auth required — no custom endpoint support
agy -p "hello"
```

### Headless mode

`agy -p "prompt"` runs once and exits. Uses cached credentials from prior interactive auth.

### Notes

- Replaced Gemini CLI (retired June 2026) — belt's Gemini init may need migration
- Named hook blocks (keyed by integration name) — unique pattern
- Only 2 hook events — very limited surface
- No custom endpoint support — can't point at test endpoint, only Google
- Skills in `.agents/skills/` (not `.antigravity/skills/`) — shared with Gemini convention

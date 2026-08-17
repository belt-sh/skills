# Cursor

## Identity

| Field | Value |
|---|---|
| Binary | `cursor` (IDE), `cursor-agent` / `agent` (CLI) |
| Install | IDE: download from cursor.com. CLI: `curl -fsSL https://downloads.cursor.com/lab/install.sh \| sh` or `brew install --cask cursor-cli` |
| Config dir | `~/.cursor` (or `CURSOR_CONFIG_DIR`) |
| CLI config | `~/.cursor/cli-config.json` |
| Detection env | `CURSOR_TRACE_ID` (IDE), `CURSOR_AGENT=1` or `CURSOR_EXTENSION_HOST_ROLE=agent-exec` (CLI) |
| Auth | `CURSOR_API_KEY` or `--api-key <key>` |

## API format

**Cursor backend (routed)** — all requests go through Cursor's own backend. Uses OpenAI-compatible format internally but does NOT support custom endpoints.

The IDE's "Override OpenAI Base URL" setting applies only to the VS Code chat panel, NOT the standalone CLI. A `--endpoint` flag reportedly does not work ([feature request, July 2026](https://forum.cursor.com/t/cursor-cli-custom-endpoint-and-api-key-support/129424)).

**Cannot point at a custom test endpoint today.**

## Hook system

Hooks config: `~/.cursor/hooks.json`

### Event names (camelCase)

| Event | When | Input | Output contract |
|---|---|---|---|
| `sessionStart` | Session begins | TBD | TBD |
| `beforeSubmitPrompt` | Before prompt sent | TBD | TBD |
| `afterFileEdit` | After file edit | TBD | TBD |
| `beforeShellExecution` | Before shell cmd | command text | can block via matcher |
| `stop` | Agent stops | TBD | TBD |
| `preCompact` | Before compaction | TBD | TBD |
| `sessionEnd` | Session ends | TBD | TBD |

### Hook config format

```json
{
  "hooks": {
    "sessionStart": [
      {
        "command": "./bin/hook-session-start.sh",
        "timeout": 60
      }
    ]
  }
}
```

Flat format — no nested `hooks` array or `matcher`/`type` fields (unlike Claude/Codex).

### Plugin root env var

TBD — Cursor plugin marketplace is submission-based, no known `CURSOR_PLUGIN_ROOT` equivalent confirmed yet.

## Skill system

| Scope | Path |
|---|---|
| User skills | `~/.cursor/skills/<name>/SKILL.md` |

Format: SKILL.md with YAML frontmatter

## Plugin system

Manifest: `.cursor-plugin/plugin.json` + `.cursor-plugin/marketplace.json`

Submission: email/Slack to Anysphere (`kniparko@anysphere.com`). No CLI install command yet.

Components: `rules/` (.mdc), `skills/`, `agents/`, `commands/`, `hooks/`, `mcp.json`, `scripts/`

Rules use `.mdc` format with frontmatter: `description`, `alwaysApply`, `globs`

## Belt status

**Supported (manual) + Marketplace (pending submission)**

- `belt init cursor` → manual skill copy + hooks.json write
- `.cursor-plugin/` manifest added to `belt-sh/skills` repo
- Hooks use `AI_AGENT=cursor belt suggest --json` (inline commands, not bin shims)
- Marketplace plugin uses `./bin/hook-*.sh` shims (richer, but requires marketplace acceptance)

## Test recipe

```bash
# Install CLI
curl -fsSL https://downloads.cursor.com/lab/install.sh | sh

# Auth
export CURSOR_API_KEY=...

# Headless test
agent -p "hello" --output-format json --trust
```

**Cannot use custom test endpoint** — CLI routes through Cursor's backend. Testing requires a valid `CURSOR_API_KEY`.

### Headless mode

`-p` / `--print` prints responses to stdout. `--output-format json` or `stream-json`. `--trust` bypasses interactive prompts. `--yolo` / `-f` auto-approves tool execution.

### Open questions

- [x] Can `cursor-agent` CLI run standalone? → Yes, install via curl/brew
- [x] Custom endpoint? → No, routes through Cursor backend. Active feature request.
- [ ] What JSON does Cursor pass to hooks on stdin?
- [ ] Does Cursor set any plugin root env var for marketplace plugins?
- [ ] Does `beforeSubmitPrompt` output get injected as system message or context?
- [ ] Does `-p` mode fire hooks?

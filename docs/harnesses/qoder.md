# Qoder CLI

## Identity

| Field | Value |
|---|---|
| Binary | `qoder` (also `qodercli`) |
| Install | See docs.qoder.com/cli/installation |
| Config dir | `~/.qoder` (or `QODER_CONFIG_DIR`) |
| Detection env | None documented in belt |
| Resume | `qodercli --resume <id>` |

## API format

Custom models via config. No documented OpenAI-compatible pass-through. See docs.qoder.com/cli/custom-models.

## Hook system

Config: `~/.qoder/settings.json`, `.qoder/settings.json`, `.qoder/settings.local.json`

Four hook types: `command`, `http`, `prompt`, `agent`

### Event names (PascalCase) — 20+ events

| Category | Events |
|---|---|
| Session | `SessionStart`, `SessionEnd` |
| Tool | `PreToolUse`, `PostToolUse`, `PostToolUseFailure`, `PermissionRequest`, `PermissionDenied` |
| Agent | `Stop`, `StopFailure`, `SubagentStart`, `SubagentStop` |
| Context | `PreCompact`, `PostCompact` |
| File | `CwdChanged`, `FileChanged`, `WorktreeCreate`, `WorktreeRemove` |
| Other | `Notification`, `InstructionsLoaded`, `ConfigChange`, `UserPromptSubmit`, `Elicitation`, `ElicitationResult` |

JSON on stdin/stdout. Exit code 2 blocks.

Env vars: `QODER_PROJECT_DIR`, `QODER_PLUGIN_ROOT`, `QODER_PLUGIN_DATA`

## Skill system

Skills bundled via plugins under `<plugin>/skills/`.

## Plugin system

Directory-based bundles containing commands, agents, skills, hooks, MCP servers, output styles.

Manifest: `.qoder-plugin/plugin.json`

Three scopes: user (default), project, local.

```bash
qoder plugins install <name>
qoder plugins marketplace add <url>
qoder plugins list
qoder plugins validate
```

## Belt status

**Not yet supported — high priority candidate**

Richest hook surface (20+ events) with Claude Code-compatible format. Has `QODER_PLUGIN_ROOT` and `QODER_PLUGIN_DATA` env vars. Plugin marketplace support. `.qoder-plugin/plugin.json` manifest format.

All of belt's 5 behaviors map directly:
- `SessionStart` → bootstrap
- `UserPromptSubmit` → suggest inject
- `PostToolUse` → mutation telemetry
- `Stop` → periodic review
- `SessionEnd` → final review

## Test recipe

```bash
# Install per docs.qoder.com/cli/installation
qodercli --resume <id>
```

### Headless mode

"Run in Scripts" mode for CI/CD documented.

### Notes

- 20+ events — tied with Kimi for richest hook surface
- Plugin format very close to Claude Code (`.qoder-plugin/plugin.json`)
- Has unique events: `WorktreeCreate`/`WorktreeRemove`, `FileChanged`, `Elicitation`
- Four hook types including `prompt` (LLM-evaluated) and `agent` — unique
- `QODER_PLUGIN_ROOT` + `QODER_PLUGIN_DATA` — full plugin env var support
- Marketplace: `qoder plugins marketplace add <url>` — same pattern as Claude/Codex

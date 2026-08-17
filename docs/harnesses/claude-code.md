# Claude Code

## Identity

| Field | Value |
|---|---|
| Binary | `claude` |
| Install | `npm install -g @anthropic-ai/claude-code` |
| Config dir | `~/.claude` (or `CLAUDE_CONFIG_DIR`) |
| Detection env | `CLAUDECODE=1` or `CLAUDE_CODE=1` |
| Variant | `CLAUDE_CODE_IS_COWORK=1` → detected as `cowork` |

## API format

**Anthropic** (`/v1/messages`)

Custom endpoint config:
```bash
export ANTHROPIC_BASE_URL=http://localhost:4100
export ANTHROPIC_API_KEY=test-key
```

Or via CLI:
```bash
claude --api-base http://localhost:4100
```

## Hook system

Hooks config: `~/.claude/settings.json` (user) or `.claude/settings.json` (project)

Plugin hooks: `hooks/hooks.json` inside the plugin directory

### Event names (PascalCase)

| Event | When | Input (stdin JSON) | Output contract |
|---|---|---|---|
| `SessionStart` | Session begins | `{session_id, cwd}` | stdout → `additionalContext` |
| `UserPromptSubmit` | Before prompt sent to LLM | `{session_id, prompt, cwd}` | stdout → system-reminder |
| `PreToolUse` | Before tool call | `{session_id, tool_name, tool_input}` | stdout → system-reminder; can block |
| `PostToolUse` | After tool call | `{session_id, tool_name, tool_input, tool_output}` | stdout → system-reminder |
| `Stop` | Agent stops responding | `{session_id, transcript_path, cwd, last_assistant_message}` | stdout → system-reminder |
| `PreCompact` | Before context compaction | `{session_id, transcript_path, cwd}` | stdout → system-reminder |
| `SessionEnd` | Session ends | `{session_id, transcript_path, cwd}` | fire-and-forget |

### Hook config format

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "belt plugin hook user-prompt-submit",
            "timeout": 3
          }
        ]
      }
    ]
  }
}
```

### Plugin root env var

`CLAUDE_PLUGIN_ROOT` — set by Claude when running plugin hooks. Points to the plugin's installed directory.

## Skill system

| Scope | Path |
|---|---|
| Plugin skills | `<plugin-root>/skills/<name>/SKILL.md` |
| User skills | `~/.claude/skills/<name>/SKILL.md` (via custom commands) |

Format: SKILL.md with YAML frontmatter (`name`, `description`, `allowed-tools`)

Activation: user types `/<skill-name>` or skill is listed in `available_skills`

## Plugin system

Manifest: `.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json`

Install flow:
```bash
claude plugin marketplace add belt-sh/skills
claude plugin install belt
```

Plugin env vars available in hooks:
- `CLAUDE_PLUGIN_ROOT` — plugin directory
- `CLAUDE_PLUGIN_DATA` — persistent data directory

## Belt status

**Supported + Marketplace**

- `belt plugin init claude` → marketplace install flow
- `.claude-plugin/` manifest in `belt-sh/skills` repo
- All 5 belt behaviors mapped
- 9 skills, 3 agents, rules, full hook suite

## Test recipe

```dockerfile
FROM node:22-slim
RUN npm install -g @anthropic-ai/claude-code
ENV ANTHROPIC_BASE_URL=http://host.docker.internal:4100
ENV ANTHROPIC_API_KEY=test-key
COPY belt /usr/local/bin/belt
RUN belt plugin init claude
# TODO: verify hooks fire — claude may require TTY
```

### Open questions

- [ ] Can Claude Code run headless (no TTY) for CI testing?
- [ ] Does `claude --print` mode fire hooks?
- [ ] Can we bypass auth for local testing?

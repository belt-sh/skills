# Codex

## Identity

| Field | Value |
|---|---|
| Binary | `codex` |
| Install | `npm install -g @openai/codex` |
| Config dir | `~/.codex` (or `CODEX_HOME`) |
| Detection env | `CODEX_SANDBOX=1`, `CODEX_CI=1`, or `CODEX_THREAD_ID` |

## API format

**OpenAI** (`/v1/chat/completions`)

Custom endpoint config:
```bash
export OPENAI_BASE_URL=http://localhost:4100/v1
export OPENAI_API_KEY=test-key
```

## Hook system

Hooks config: `~/.codex/hooks.json`

Enable hooks: `~/.codex/config.toml` → `[features] hooks = true`

### Event names (PascalCase)

| Event | When | Input | Output contract |
|---|---|---|---|
| `SessionStart` | Session begins | `{session_id, cwd}` | stdout → context |
| `UserPromptSubmit` | Before prompt | `{session_id, prompt, cwd}` | stdout → context |
| `Stop` | Agent stops | `{session_id, transcript_path, cwd}` | stdout → context |
| `PreCompact` | Before compaction | `{session_id, transcript_path, cwd}` | stdout → context |

### Hook config format

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "${PLUGIN_ROOT}/bin/hook-session-start.sh",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

### Plugin root env var

`PLUGIN_ROOT` — set by Codex when running plugin hooks.

## Skill system

| Scope | Path |
|---|---|
| Plugin skills | `<plugin-root>/skills/<name>/SKILL.md` |
| User skills | `~/.agents/skills/<name>/SKILL.md` |

Format: SKILL.md with YAML frontmatter

## Plugin system

Manifest: `.codex-plugin/plugin.json`

Install flow:
```bash
codex plugin marketplace add belt-sh/skills
codex plugin add belt@belt-sh-skills
```

Plugin JSON fields: `name`, `version`, `description`, `author`, `skills`, `hooks`, `interface`

## Belt status

**Supported + Marketplace**

- `belt plugin init codex` → marketplace install flow
- `.codex-plugin/` manifest in `belt-sh/skills` repo
- All belt behaviors mapped
- Falls back to manual init if `codex` CLI not on PATH

## Test recipe

```dockerfile
FROM node:22-slim
RUN npm install -g @openai/codex
ENV OPENAI_BASE_URL=http://host.docker.internal:4100/v1
ENV OPENAI_API_KEY=test-key
COPY belt /usr/local/bin/belt
RUN belt plugin init codex
```

### Open questions

- [ ] Does Codex require `[features] hooks = true` to fire plugin hooks, or only user hooks?
- [ ] Can Codex run headless?

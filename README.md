# belt-sh/skills

Plugin for [belt](https://belt.sh) — the cloud platform for AI agents. Works with Claude Code, Codex, Cursor, Gemini CLI, Pi, OpenCode, and Windsurf.

## Install

### Claude Code

```
/plugin marketplace add belt-sh/skills
/plugin install belt
/reload-plugins
```

### OpenAI Codex

```
codex plugin marketplace add belt-sh/skills
codex plugin install belt
```

Then run `/hooks` in Codex to review and trust belt hooks.

### Pi

```bash
belt plugin init pi
```

### OpenCode

```bash
belt plugin init opencode
```

### Any agent (via belt CLI)

```bash
belt plugin init claude     # Claude Code
belt plugin init codex      # OpenAI Codex
belt plugin init cursor     # Cursor
belt plugin init gemini     # Gemini CLI
belt plugin init pi         # Pi
belt plugin init opencode   # OpenCode
belt plugin init windsurf   # Windsurf
```

## What you get

| Command | What it does |
|---|---|
| `/belt` | Install and set up belt CLI |
| `/skill` | Search, use, install, publish skills |
| `/skillify` | Turn a working solution into a permanent, tested skill |
| `/knowledge` | Save and search your knowledge base |
| `/apps` | Search and run 250+ AI apps |
| `/suggest` | Unified search across skills, knowledge, and apps |
| `/agentify` | Build and deploy a custom AI agent with tools |
| `/appify` | Build and deploy an inference.sh app |
| `/flowify` | Chain apps into a multi-step flow pipeline |

## Hooks (automatic)

- **SessionStart** — checks belt is installed and authenticated
- **UserPromptSubmit** — searches skills/knowledge/apps, injects relevant suggestions
- **Stop** — extracts reusable knowledge and skills from the session
- **PreCompact** — captures knowledge before conversation compaction
- **PostToolUse** — detects edits to belt-managed skills (Claude Code)
- **SessionEnd** — final knowledge extraction + session summary (Claude Code)

## Requires

- [belt CLI](https://belt.sh) — install with `belt` or see `/belt` skill
- `belt login` for authenticated features

All hooks fail silently if belt is not installed. Skills work as plain reference without belt.

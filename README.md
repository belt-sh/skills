# belt-sh/skills

Claude Code plugin for [belt](https://belt.sh) — the skills registry for AI agents.

## Install

```
/plugin install belt-sh
```

## What you get

| Command | What it does |
|---|---|
| `/belt` | Install and set up belt CLI |
| `/skill` | Search, use, install, publish skills |
| `/knowledge` | Save and search your knowledge base |
| `/apps` | Search and run 250+ AI apps |

## Hooks (automatic)

- **SessionStart** — checks for skill updates
- **UserPromptSubmit** — searches skills/knowledge/apps, injects relevant suggestions
- **Stop** — evaluates if session discovered knowledge worth saving or improved a skill
- **PostToolUse** — detects edits to belt-managed skills

## Requires

- [belt CLI](https://belt.sh) (`curl -fsSL https://belt.sh/install | sh`)
- `belt login` for authenticated features

All hooks fail silently if belt is not installed. Skills work as plain reference without belt.

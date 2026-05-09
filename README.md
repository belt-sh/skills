# belt-sh/skills

Claude Code plugin for [belt](https://belt.sh) — the cloud platform for AI agents.

## Install

```
/plugin marketplace add belt-sh/skills
/plugin install belt
```

## What you get

| Command | What it does |
|---|---|
| `/belt` | Install and set up belt CLI |
| `/skill` | Search, use, install, publish skills |
| `/knowledge` | Save and search your knowledge base |
| `/apps` | Search and run 250+ AI apps |
| `/suggest` | Unified search across skills, knowledge, and apps |

## Hooks (automatic)

- **SessionStart** — checks for skill updates
- **UserPromptSubmit** — searches skills/knowledge/apps, injects relevant suggestions
- **Stop** — logs session events (knowledge capture coming soon)
- **PostToolUse** — detects edits to belt-managed skills
- **SessionEnd** — logs session summary

## Requires

- [belt CLI](https://belt.sh) (`curl -fsSL https://belt.sh/install | sh`)
- `belt login` for authenticated features

All hooks fail silently if belt is not installed. Skills work as plain reference without belt.

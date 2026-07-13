# belt-sh/skills

Plugin for [belt](https://belt.sh) — the cloud platform for AI agents. Works with Claude Code, Codex, Cursor, Gemini CLI, Pi, OpenCode, and Windsurf.

## Install (recommended)

```bash
belt plugin init claude     # Claude Code
belt plugin init codex      # OpenAI Codex
belt plugin init cursor     # Cursor
belt plugin init gemini     # Gemini CLI
belt plugin init pi         # Pi
belt plugin init opencode   # OpenCode
belt plugin init windsurf   # Windsurf
```

Shows exactly what data is sent before installing.

### Alternative: native plugin install

**Claude Code:** `/plugin marketplace add belt-sh/skills` → `/plugin install belt` → `/reload-plugins`

**Codex:** `codex plugin marketplace add belt-sh/skills` → `codex plugin install belt` → `/hooks`

## What you get

`/belt` `/skill` `/skillify` `/knowledge` `/apps` `/suggest` `/agentify` `/appify` `/flowify`

## What hooks send

| Hook | Sends upstream |
|------|---------------|
| SessionStart | nothing — local auth check |
| UserPromptSubmit | prompt text (semantic search for matching skills/apps) |
| Stop / PreCompact | tool names, file paths, tags (no code) |
| PostToolUse | nothing — local file check |
| SessionEnd | message count, tools used, files touched |

**Never sent:** source code, file contents, secrets, conversation history.

All hooks fail silently if belt is not installed. Skills work as plain reference without belt.

## Data & security

Encrypted (TLS 1.2+ / AES-256) · SOC 2 · GDPR · PCI DSS · we never train on your data

[privacy](https://inference.sh/privacy) · [terms](https://inference.sh/terms) · [trust](https://inference.sh/trust) · [security](https://inference.sh/security)

## Disable

Per project: `.beltsh/config.json` → `{"hooks_disabled": true}`

Via env: `BELT_NO_HOOKS=1`

Granular: `suggest_disabled` (no prompt matching) · `knowledge_disabled` (no session extraction)

## Requires

[belt CLI](https://belt.sh) — `curl -fsSL https://cli.inference.sh | sh` or `brew install inference-sh/tap/belt`

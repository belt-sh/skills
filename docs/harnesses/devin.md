# Devin CLI

## Identity

| Field | Value |
|---|---|
| Binary | `devin` |
| Install | `curl -fsSL https://cli.devin.ai/install.sh \| bash` |
| SDK | `pip install devin-sdk` / `npm install @cognition/devin-core` |
| Config dir | `~/.config/devin/` (or `$XDG_CONFIG_HOME/devin`) |
| Detection env | Belt detects via `/opt/.devin` filesystem check |
| Resume | `devin --resume <id>` |
| By | Cognition |

## API format

**Proprietary** — Devin API, not OpenAI-compatible.

```bash
export DEVIN_API_TOKEN=...
export DEVIN_ORG_ID=...
export DEVIN_BASE_URL=https://api.devin.ai/v3  # default
```

Cannot point at a custom OpenAI-compatible endpoint.

## Hook system

Config: `.devin/hooks.v1.json` (project) or user-level config

Uses the same JSON format as Claude Code hooks.

### Events

Session lifecycle, prompt, tool-use, permission, stop. Exact event names TBD — follows Claude Code convention.

## Skill system

Custom tools defined in `.devin/config.yaml`. Supports `AGENTS.md` for project-level instructions. Skills configurable per-repo.

## Plugin system

MCP server support. No public marketplace. Cloud handoff allows escalating local sessions to Devin's cloud sandbox.

## Belt status

**Not yet supported**

Detection exists (filesystem-based `/opt/.devin` check). Hook format matches Claude Code — low integration effort if event names are confirmed.

Proprietary API means can't point at test endpoint.

## Test recipe

```bash
curl -fsSL https://cli.devin.ai/install.sh | bash
export DEVIN_API_TOKEN=...
export DEVIN_ORG_ID=...
devin --resume <id>
```

### Headless mode

Set `DEVIN_API_TOKEN` and `DEVIN_ORG_ID` as env vars to skip interactive prompts. Script mode available.

### Notes

- Hooks follow Claude Code JSON format — easy to template
- Proprietary API — no custom endpoint testing
- Cloud handoff feature is unique — could be relevant for belt agent orchestration
- `.devin/hooks.v1.json` — versioned hook config format, forward-thinking

# belt-sh/skills

Plugin for [belt](https://belt.sh) — the cloud platform for AI agents. Works with 13 coding agent CLIs.

## Install (recommended)

```bash
belt plugin init <harness>
```

Shows exactly what data is sent before installing. Supported harnesses:

`claude` · `codex` · `copilot` · `cursor` · `droid` · `gemini` · `goose` · `grok` · `hermes` · `kilo` · `kimi` · `opencode` · `pi` · `qwen`

### Alternative: native plugin install

| Harness | Command |
|---------|---------|
| Claude Code | `/plugin marketplace add belt-sh/skills` → `/plugin install belt` |
| Codex | `codex plugin marketplace add belt-sh/skills` → `codex plugin add belt@belt-sh-skills` |
| Grok | `grok skills install https://github.com/belt-sh/skills` |
| Gemini CLI | `gemini extensions install https://github.com/belt-sh/skills` |
| Droid | `droid plugin install belt@belt-sh-skills` |
| Kimi | `/plugins install https://github.com/belt-sh/skills` |
| Copilot | Copy `hooks/copilot-hooks.json` to `.copilot/hooks/belt.json` |
| Pi | `pi install https://github.com/belt-sh/skills` |
| Hermes | `hermes plugins install belt-sh/skills` |
| OpenCode | Copy `hooks/opencode-plugin.ts` to `.opencode/plugins/` |
| Kilo | Copy `hooks/kilo-plugin.ts` to `.kilo/plugins/` |
| Qwen | Merge `hooks/qwen-hooks.json` into `.qwen/settings.json` |
| Goose | Copy `hooks/goose-hooks.json` to `.agents/plugins/belt/hooks/hooks.json` |

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

## Supported hook formats

| Format | Harnesses |
|--------|-----------|
| JSONNested | claude, codex, grok, droid, goose, qwen, gemini |
| JSONCopilot | copilot |
| TOML | kimi |
| YAML | hermes |
| TypeScript (extension) | pi |
| TypeScript (plugin) | opencode, kilo |

## Data & security

Encrypted (TLS 1.2+ / AES-256) · SOC 2 · GDPR · PCI DSS · we never train on your data

[privacy](https://inference.sh/privacy) · [terms](https://inference.sh/terms) · [trust](https://inference.sh/trust) · [security](https://inference.sh/security)

## Debug

`BELT_HOOK_DEBUG=1` — logs hook events with timing to stderr (`[belt:hook] session-start done (42ms)`)

Hooks also log to `~/.belt/hooks.log` (always, no env var needed).

## Disable

Per project: `.beltsh/config.json` → `{"hooks_disabled": true}`

Via env: `BELT_NO_HOOKS=1`

Granular: `suggest_disabled` (no prompt matching) · `knowledge_disabled` (no session extraction)

## Requires

[belt CLI](https://belt.sh) — `curl -fsSL https://cli.inference.sh | sh` or `brew install inference-sh/tap/belt`

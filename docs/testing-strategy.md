# Testing Strategy

## Goal

Verify belt plugin installation and hook execution across every supported harness, in CI, without manual intervention.

## Architecture

```
┌─────────────────┐     ┌──────────────────────┐
│  Docker container│     │  Test endpoint        │
│  (per harness)   │────▶│  (OpenAI-compatible)  │
│                  │     │                       │
│  1. install CLI  │     │  - logs all requests  │
│  2. belt init    │     │  - returns minimal    │
│  3. send prompt  │     │    valid responses    │
│  4. assert hooks │     │  - records hook calls │
│     fired        │     │                       │
└─────────────────┘     └──────────────────────┘
```

## Test endpoint

A thin HTTP server that:
- Serves `POST /v1/chat/completions` (OpenAI format)
- Serves `POST /v1/messages` (Anthropic format)
- Returns a minimal valid streaming response
- Logs every request (headers, body) to a JSONL file
- Runs as a sidecar or host-network service

See [endpoint-spec.md](./endpoint-spec.md) for the full spec.

## Per-harness test flow

```bash
# 1. Install the harness CLI
# 2. Configure it to hit the test endpoint
# 3. Install belt and run belt init <agent>
# 4. Start a session and send one prompt
# 5. Assert:
#    - hooks.log shows session-start fired
#    - hooks.log shows user-prompt-submit fired
#    - test endpoint received at least one /v1/chat/completions request
#    - skills directory contains expected SKILL.md files
```

## What each test verifies

| Check | What it proves |
|---|---|
| `belt init <agent>` exits 0 | Installation doesn't crash |
| Skills dir populated | Skills installed correctly |
| hooks.json / config written | Hook config installed correctly |
| session-start hook fired | Harness calls hooks on start |
| user-prompt hook fired | Harness calls hooks before prompt |
| Test endpoint received request | Harness actually talks to configured endpoint |
| stop hook fired | Harness calls hooks on stop |

## Harness API format groups

| API format | Harnesses | Endpoint | Custom URL env var |
|---|---|---|---|
| Anthropic | Claude Code | `POST /v1/messages` | `ANTHROPIC_BASE_URL` |
| OpenAI | Codex, Cursor, Grok, Copilot, Windsurf, Hermes | `POST /v1/chat/completions` | varies (see INDEX) |
| Multi-provider | Kimi, Droid, Kilo, OMP, MastraCode | configurable | config file |
| Google-only | Antigravity | Gemini API | none |
| Proprietary | Devin | Devin API | `DEVIN_BASE_URL` (own API only) |
| Unknown | Pi, Gemini CLI | TBD | TBD |

## Headless mode support (confirmed)

| Harness | Headless flag | Notes |
|---|---|---|
| Claude Code | `-p` / `--print` | may not fire all hooks |
| Codex | TBD | |
| Cursor | TBD | CLI agent may need IDE |
| Grok | `-p` | `--output-format json` for structured |
| Copilot | `-p` / `--prompt` | stdout mixes UI chrome |
| Kimi | `-p` | `print_background_mode` controls behavior |
| Droid | `exec "task"` | `--auto low\|medium\|high` autonomy |
| Kilo | `run "task"` | `--format json`, `--auto` |
| Qoder | TBD | "Run in Scripts" mode |
| Hermes | `serve` | headless backend + OpenAI API |
| MastraCode | env vars | `MASTRA_API_TOKEN` etc. |
| Antigravity | `-p` | cached auth required |
| Devin | env vars | `DEVIN_API_TOKEN` etc. |
| OMP | TBD | scripted mode supported |

## Testability tiers

### Tier 1 — fully testable (custom endpoint + headless + JSON hooks)
Grok, Copilot, Droid, Codex, Claude Code

### Tier 2 — testable with config file setup
Kimi (TOML config), Kilo (env vars), Hermes (YAML config), OMP (YAML config)

### Tier 3 — limited testing (no custom endpoint)
Cursor (routes through Cursor backend, feature request open), Antigravity (Google-only), Devin (proprietary), Windsurf (IDE-only), Pi (unknown)

## Docker base images

| Image | Harnesses |
|---|---|
| `node:22-slim` | Codex, Copilot, Kilo, MastraCode, OMP |
| `ubuntu:24.04` | Claude Code, Grok, Droid, Kimi, Antigravity, Devin |
| `python:3.12-slim` | Hermes |

## CI integration

GitHub Actions matrix — start with Tier 1:

```yaml
strategy:
  matrix:
    harness: [grok, copilot, droid, codex, claude-code]
```

Each matrix entry runs a harness-specific Dockerfile against the shared test endpoint.

## Open questions (updated)

- [x] Which harnesses can run headless? → see table above
- [ ] Does Claude Code `--print` mode fire hooks?
- [ ] Can Cursor CLI agent run standalone without the IDE?
- [ ] Which harnesses accept any API key (no server-side validation)?
- [ ] Does Grok require SuperGrok subscription even with custom endpoint?
- [ ] Can Devin/Antigravity be tested at all without vendor auth?

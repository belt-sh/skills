# Gemini CLI

## Identity

| Field | Value |
|---|---|
| Binary | `gemini` |
| Install | `npm install -g @anthropic-ai/gemini-cli` or `brew install gemini-cli` |
| Config dir | `~/.gemini` |
| Detection env | `GEMINI_CLI=1` |

## API format

**Google** (`/v1/models/*/generateContent`)

Custom endpoint config:
```bash
export GEMINI_API_URL=http://localhost:4100
export GEMINI_API_KEY=test-key
```

TBD — need to verify exact env var names and whether Gemini CLI supports custom base URLs.

## Hook system

Hooks config: `~/.gemini/settings.json`

### Event names

TBD — belt's template uses similar structure to Claude hooks. Need to verify from Gemini CLI docs.

### Hook config format (belt template)

Uses the same nested format as Claude/Codex (belt embeds `gemini.json`).

## Skill system

| Scope | Path |
|---|---|
| User skills | `~/.gemini/skills/<name>/SKILL.md` |

## Plugin system

No known marketplace. Manual init via `belt init gemini`.

## Belt status

**Supported (manual)**

- `belt init gemini` → skills + hooks to `~/.gemini/`
- No marketplace plugin

## Test recipe

```dockerfile
FROM node:22-slim
RUN npm install -g @google/gemini-cli  # verify package name
ENV GEMINI_API_URL=http://host.docker.internal:4100
ENV GEMINI_API_KEY=test-key
COPY belt /usr/local/bin/belt
RUN belt init gemini
```

### Open questions

- [ ] Exact env var for custom API endpoint
- [ ] Hook event names — are they same as Claude (PascalCase)?
- [ ] Can Gemini CLI run headless?
- [ ] Does Gemini CLI support OpenAI-compatible endpoints or only Google format?

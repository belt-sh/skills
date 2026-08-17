# OpenCode

## Identity

| Field | Value |
|---|---|
| Binary | `opencode` |
| Install | `go install github.com/opencode-ai/opencode@latest` or binary download |
| Config dir | `~/.config/opencode` |
| Detection env | `OPENCODE_CLIENT` |

## API format

**OpenAI** (`/v1/chat/completions`)

Custom endpoint config:
```bash
export OPENAI_BASE_URL=http://localhost:4100/v1
export OPENAI_API_KEY=test-key
```

TBD — verify if OpenCode uses OPENAI_BASE_URL or its own config.

## Hook system

Plugin-based: TypeScript plugins in `~/.config/opencode/plugins/<name>/index.ts`

### Plugin format

```typescript
import { definePlugin } from "@opencode-ai/plugin";

export default definePlugin({
  name: "belt",
  hooks: {
    onSessionStart: async (ctx) => { /* ... */ },
    onPromptSubmit: async (ctx) => { /* ... */ },
    onStop: async (ctx) => { /* ... */ },
  }
});
```

Requires `npm install @opencode-ai/plugin` in the plugin directory.

Registration: `opencode plugin --global <plugin-dir>`

## Skill system

| Scope | Path |
|---|---|
| User skills | `~/.config/opencode/skills/<name>/SKILL.md` |

## Plugin system

TypeScript plugins with `package.json` + `index.ts`. No marketplace.

## Belt status

**Supported (manual)**

- `belt init opencode` → writes `index.ts` plugin + installs deps + registers
- Plugin template embedded in Go binary

## Test recipe

```dockerfile
FROM node:22-slim
RUN go install github.com/opencode-ai/opencode@latest  # needs Go
ENV OPENAI_BASE_URL=http://host.docker.internal:4100/v1
ENV OPENAI_API_KEY=test-key
COPY belt /usr/local/bin/belt
RUN belt init opencode
```

### Open questions

- [ ] Exact env var for custom endpoint
- [ ] Does the plugin system support async hooks?
- [ ] What data does each hook callback receive?
- [ ] Can OpenCode run headless?

# Pi

## Identity

| Field | Value |
|---|---|
| Binary | `pi` |
| Install | Binary download from earendil-works |
| Config dir | `~/.pi/agent` (or `PI_CODING_AGENT_DIR`) |
| Detection env | `PI_CODING_AGENT=1` |

## API format

**Custom** — Pi uses its own API format. TBD whether it supports OpenAI-compatible endpoints.

## Hook system

Extension-based: TypeScript extensions in `~/.pi/agent/extensions/<name>.ts`

Registration: `pi install <extension-path>`

### Extension format

Belt's embedded template is a `.ts` file that calls belt CLI commands on lifecycle events.

## Skill system

| Scope | Path |
|---|---|
| User skills | `~/.pi/agent/extensions/` |

## Plugin system

TypeScript extensions. No marketplace.

## Belt status

**Supported (manual)**

- `belt init pi` → writes extension `.ts` + registers via `pi install`

## Test recipe

```dockerfile
FROM ubuntu:24.04
# Pi binary install TBD
ENV PI_CODING_AGENT=1
COPY belt /usr/local/bin/belt
RUN belt init pi
```

### Open questions

- [ ] Can Pi be configured with a custom LLM endpoint?
- [ ] What API format does Pi use?
- [ ] Extension lifecycle events and their data format
- [ ] Can Pi run headless?

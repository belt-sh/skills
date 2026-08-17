# Windsurf

## Identity

| Field | Value |
|---|---|
| Binary | `windsurf` (IDE-based) |
| Install | Windsurf IDE download |
| Config dir | `~/.codeium/windsurf` |
| Detection env | None confirmed — belt does not auto-detect Windsurf via env var |

## API format

**OpenAI** (`/v1/chat/completions`) — TBD, Windsurf routes through Codeium's backend.

Custom endpoint config: TBD

## Hook system

Hooks config: `~/.codeium/windsurf/hooks.json`

Belt's embedded template uses the same format as Cursor (flat command entries).

## Skill system

| Scope | Path |
|---|---|
| User skills | `~/.windsurf/skills/<name>/SKILL.md` |

## Plugin system

No known marketplace. Manual init via `belt init windsurf`.

## Belt status

**Supported (manual)**

- `belt init windsurf` → skills to `~/.windsurf/skills/` + hooks to `~/.codeium/windsurf/hooks.json`

## Test recipe

Windsurf is IDE-only — headless CI testing may not be possible.

### Open questions

- [ ] Can Windsurf be configured with a custom endpoint?
- [ ] Is there a CLI agent mode?
- [ ] Hook event names and data format
- [ ] Does Windsurf have or plan a plugin marketplace?

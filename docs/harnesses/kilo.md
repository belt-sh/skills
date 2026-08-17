# Kilo Code CLI

## Identity

| Field | Value |
|---|---|
| Binary | `kilo` |
| Install | `npm install -g @kilocode/cli`, brew, or curl |
| Config dir | `~/.config/kilo` |
| Detection env | None documented in belt |
| Resume | `kilo --session <id>` |
| By | Kilo |

## API format

**Multi-provider** — 500+ models via Kilo Gateway.

```bash
export KILO_PROVIDER=openai
export KILO_API_KEY=test-key
export KILOCODE_MODEL=test-model
```

Custom providers configurable in `~/.config/kilo/kilo.jsonc`. OpenAI-compatible providers supported.

## Hook system

**Plugin-based** — no standalone hooks config. Hooks live inside TypeScript/JavaScript plugin modules.

### Plugin lifecycle events

| Category | Events |
|---|---|
| Lifecycle | `config`, `session.created`, `message.updated`, `file.edited`, `lsp.updated` |
| Tool | `tool`, `tool.execute.before`, `tool.execute.after`, `tool.definition` |
| Chat | `chat.message`, `chat.params`, `chat.headers`, `permission.ask` |
| Auth | `auth`, `provider` |
| Experimental | system prompt transform, compaction hooks |

### Plugin format

TypeScript/JS modules. Config in `kilo.jsonc` `plugin` array or `.kilo/plugin/` directory. npm packages auto-installed via Bun.

Standalone tool files in `.kilo/tool/`.

`KILO_PURE=1` skips external plugins.

## Skill system

| Scope | Path |
|---|---|
| User | `~/.kilo/skills/` |
| Project | `.kilo/skills/` |

Agent modes: Architect, Ask, Debug, Orchestrator, plus custom modes. `kilo agent create` to scaffold.

## Plugin system

TS/JS modules. No marketplace documented — plugins are local or npm packages.

## Belt status

**Not yet supported**

Plugin-based hook system (not config-file-based) makes integration harder. Belt would need to ship as a TS/JS plugin module, not just a hooks.json template. The `tool.execute.before`/`tool.execute.after` events map to belt's `PreToolUse`/`PostToolUse`. Missing a clear `UserPromptSubmit` equivalent — may need `chat.message`.

## Test recipe

```bash
npm install -g @kilocode/cli
export KILO_PROVIDER=openai
export KILO_API_KEY=test-key
kilo run "hello" --auto --format json
```

### Headless mode

`kilo run "task"` with `--auto` for CI. `--format json` for structured output. `kilo serve` + `kilo attach` for remote control. Exit codes: 0 success, 124 timeout, 1 error.

### Notes

- Plugin-based hooks (TS/JS modules) — different from every other harness
- Would need a belt TS plugin package, not just a hooks.json template
- Event names use dot notation (`tool.execute.before`) — unique convention
- `kilo serve` + `kilo attach` model is interesting for remote testing

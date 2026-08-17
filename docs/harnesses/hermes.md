# Hermes Agent

## Identity

| Field | Value |
|---|---|
| Binary | `hermes` |
| Install | `pip install hermes-agent`, curl script, or Docker |
| Config dir | `~/.hermes` |
| Detection env | None documented in belt |
| Resume | `hermes --resume <id>` |
| By | Nous Research |

## API format

**OpenAI-compatible** — works with any OpenAI-compatible endpoint (Ollama, vLLM, LM Studio, OpenRouter). Also supports native Anthropic and Google providers.

Provider configured in `~/.hermes/config.yaml`.

Custom endpoint config:
```yaml
# ~/.hermes/config.yaml
provider:
  type: openai
  base_url: http://localhost:4100/v1
  api_key: test-key
```

## Hook system

Shell-script hooks declared in `~/.hermes/config.yaml`.

### Event names (snake_case)

| Event | When |
|---|---|
| `on_session_start` | Session begins |
| `pre_tool_call` | Before tool call |
| `post_tool_call` | After tool call |
| `pre_llm_call` | Before LLM request |
| `subagent_stop` | Subagent done |

Payloads: JSON on stdin.

First-use consent tracked in `~/.hermes/shell-hooks-allowlist.json`.

CLI: `hermes hooks` to inspect/test.

## Skill system

No separate skill system documented. Behavior guided by project-level `AGENTS.md` files.

## Plugin system

Plugins install to `~/.hermes/plugins/<name>/`. Enabled in `config.yaml` under `plugins.enabled`. Installable from Git repos with optional commit pinning.

## Belt status

**Not yet supported**

Snake_case event names are unique among harnesses. Limited hook surface (5 events). `pre_llm_call` is interesting — fires before each LLM request, not just user prompts.

## Test recipe

```bash
pip install hermes-agent
# Configure in ~/.hermes/config.yaml
hermes serve  # headless backend since v0.18.0
```

### Headless mode

`hermes serve` — true headless backend. Also exposes an OpenAI-compatible HTTP API for external frontends.

### Notes

- Only 5 hook events — missing `SessionEnd`, `Stop`, `PreCompact`
- snake_case event names — only harness using this convention
- `hermes serve` is unique — can act as a proxy, potentially useful for testing

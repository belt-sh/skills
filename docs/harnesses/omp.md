# OMP (Oh My Pi)

## Identity

| Field | Value |
|---|---|
| Binary | `omp` |
| Install | `npm i -g @oh-my-pi/pi-coding-agent`, Homebrew, or install script |
| Config dir | `~/.omp/agent/` (or `OMP_AGENT_DIR` / `PI_CODING_AGENT_DIR`) |
| Detection env | None documented in belt |
| Resume | `omp --resume=<session>` |
| By | Can Boluk |
| Runtime | Requires Bun v1.3.14+ to run from source |

## API format

**Multi-provider** — supports 40+ providers: Anthropic, OpenAI, Google Gemini, xAI, Mistral, Groq, Cerebras, Fireworks, OpenRouter, and more.

Provider configured in `~/.omp/agent/config.yml`.

Custom endpoint config:
```yaml
# ~/.omp/agent/config.yml
provider:
  type: openai
  base_url: http://localhost:4100/v1
  api_key: test-key
```

## Hook system

TypeScript modules (`.ts` files) in `pre/` or `post/` directories subscribing to lifecycle events.

Project discovery walks upward from cwd, checking `.omp/` then legacy `.pi/` directories.

Also supports Claude Code-style JSON command hooks via the `omp-hooks` bridge.

## Skill system

Skills are markdown or TS files inside plugin directories. Slash commands auto-register from installed plugins.

## Plugin system

Marketplace is a Git repo (or local dir) with `.omp-plugin/marketplace.json` catalog. Plugins bundle skills, commands, agents, hooks, tools, MCP/LSP servers. Publishable to npm or via marketplace repos.

## Belt status

**Not yet supported**

OMP shares Pi's extension pattern (TypeScript modules) but with its own config dirs and marketplace format. Adding support:
- Add env var detection (find what OMP sets)
- TypeScript extension template (similar to Pi's)
- Could also submit to OMP marketplace

## Test recipe

```bash
npm i -g @oh-my-pi/pi-coding-agent
# Configure in ~/.omp/agent/config.yml
omp --resume=<session>
```

### Headless mode

Non-interactive/scripted mode supported.

### Notes

- Shares Pi's TS extension pattern but with separate config dir
- `.omp-plugin/marketplace.json` — similar to Claude/Codex plugin format
- 40+ provider support makes test endpoint config easy
- Legacy `.pi/` directory fallback — shares ancestry with Pi

---
name: agentify
description: "Build and deploy a custom AI agent on belt — define purpose, model, tools, system prompt, and deploy as a reusable agent. Use when the user says 'agentify', 'make an agent', 'create an agent', 'build a bot', or wants to turn a workflow into a deployed agent with tools."
allowed-tools: Bash(belt *), Read, Write, Edit, Glob, Grep, Agent
---

## Agentify

Turn a workflow, task, or idea into a deployed agent on belt. Agents get a model, system prompt, optional MCP tools, skills, and callable API tools — then run via `belt agent run` or the inference.sh API.

### When to use

- The user wants to create a new agent from scratch
- A working workflow should become a persistent, callable agent
- The user says "agentify", "make an agent", "create an agent", "build a bot"
- A skill or manual process would benefit from tool access (APIs, MCP, integrations)

### Key concepts

**Context variables** are runtime inputs passed via `--context key=value`. They appear in the agent's conversation as structured context the model can read. Use them in `call` tool URLs with `{{context.var_name}}` template syntax. In system prompts, just reference them by name — the model receives them automatically.

**Internal tools** default to all OFF when omitted from the YAML. Set explicitly:
- `memory: true` — agent can store/recall knowledge across conversations
- `plan: true` — agent can create and track execution plans
- `widget: true` — agent can render UI widgets

**MCP tools vs call tools**: MCP tools connect to external services via OAuth (Todoist, Linear, Slack). Call tools are direct HTTP requests to any API with bearer auth.

**Client tools** (`type: client`) are declared by the agent but executed by the frontend — for acting on live UI state the server cannot see. See the client tools section below; the schema nesting there is easy to get wrong and fails silently.

### Process

#### 0. Analyze the conversation first [MANDATORY]

Before asking anything, review what happened in this conversation. Look for:
- Multi-step workflows that were executed (API calls, deployments, configurations)
- Repeatable patterns that could run autonomously
- Tool sequences that could be packaged into an agent

If you find workflow candidates, present them to the user:
> "Based on this session, here's what could become an agent:
> 1. **<name>** — <what it would do>, triggered by <what>
> 2. **<name>** — <what it would do>, triggered by <what>
> Which one should we build? Or describe something different."

Only ask "what kind of agent?" if the conversation has no workflow context.

#### 0.5. Learn from a working agent [recommended]

Pull an existing agent to see what real YAML looks like:

```bash
belt agent list
belt agent pull <namespace/agent-name> --save /tmp/reference-agent.yml
```

Read the YAML — it shows the exact field names, structure, and how tools are defined.

#### 1. Define the agent's purpose

From the conversation analysis or user input, pin down:
- **What does this agent do?** One sentence.
- **What inputs does it need?** These become context variables.
- **What tools does it need?** MCP servers, API calls, or built-in tools.
- **What model should power it?** Default: `openrouter/claude-sonnet-46`. Use `openrouter/claude-haiku-45` for simple/high-volume agents.

#### 2. Check for existing agents

```bash
belt agent list
```

If a similar agent exists, pull and review it:
```bash
belt agent pull <namespace/name> --save /tmp/existing-agent.yml
```

#### 3. Set up tools

**For MCP tools** (Todoist, Linear, Slack, GitHub, etc.):

First, connect the MCP server. This opens a browser for OAuth:
```bash
belt mcp search "todoist"              # find the server
belt mcp connect <server-slug>         # opens browser for OAuth
```

After connecting, get the integration ID — you'll need it for the YAML:
```bash
belt integrations list --json          # find the "id" field for your integration
```

The `id` field (e.g. `46e8a4bqj9ddx8gmana01b2shq`) is what goes in the YAML as `integration_id`.

Then list available tools:
```bash
belt mcp tools <server-slug>           # only works after connecting
```

**For call tools** (direct HTTP APIs):

Set up secrets for API authentication:
```bash
belt secrets set my_api_key sk-12345
```

No connection step needed — call tools are just HTTP requests defined in the YAML.

**If `belt mcp connect` requires a browser** and you're in a headless environment, tell the user they need to run `belt mcp connect <slug>` in their local terminal first.

#### 4. Write the agent YAML

**Minimal agent** (no tools):

```yaml
name: my-agent
description: what it does in one line
core_app:
  ref: openrouter/claude-sonnet-46
system_prompt: |
  You are a helpful assistant that does X.
  When the user asks Y, do Z.
```

**Agent with context variables**:

```yaml
name: pr-reviewer
description: reviews pull requests
core_app:
  ref: openrouter/claude-sonnet-46
system_prompt: |
  You review pull requests. The user provides a PR URL
  and you analyze the changes and provide feedback.
context:
  - name: repo
    description: GitHub repo in owner/name format
    required: true
  - name: pr_number
    description: pull request number
    required: true
```

**Agent with MCP tools** (real working example):

```yaml
name: todoist-bot
description: manages todoist tasks
core_app:
  ref: openrouter/claude-sonnet-46
system_prompt: |
  You manage Todoist tasks. Use find-projects to list projects.
tools:
  - name: find-projects
    type: mcp
    description: list all Todoist projects
    mcp:
      integration_id: 46e8a4bqj9ddx8gmana01b2shq  # from: belt integrations list --json
      tool_name: find-projects                      # from: belt mcp tools todoist
```

**Agent with call tools** (direct HTTP):

```yaml
name: status-checker
description: checks service health
core_app:
  ref: openrouter/claude-haiku-45
system_prompt: |
  You check service status. Use the health_check tool
  to verify the service at the given URL is responding.
context:
  - name: service_url
    description: base URL of the service to check
    required: true
tools:
  - name: health_check
    type: call
    description: check if a service is healthy
    call:
      method: GET
      url: "{{context.service_url}}/health"
```

**Agent with call tools + auth**:

```yaml
tools:
  - name: get_data
    type: call
    description: fetch data from the API
    call:
      method: GET
      url: https://api.example.com/data
      auth:
        type: bearer
        secret: my_api_key  # set with: belt secrets set my_api_key sk-xxx
      input_schema:
        type: object
        properties:
          query:
            type: string
```

**Agent with skills and internal tools**:

```yaml
name: pricing-agent
description: configures app pricing
core_app:
  ref: openrouter/claude-sonnet-45
skills:
  - name: cel-pricing
    skill_id: infsh/cel-pricing
    description: CEL pricing reference and patterns
internal_tools:
  memory: true
  plan: false
  widget: false
```

Save to a file named `<agent-name>.yml`.

**Agent with client tools** (executed by the frontend, not the server):

Use these when the tool must act on live UI state — the open editor, the current
selection, an unsaved form. The agent owns the schema; the browser owns the
implementation.

```yaml
tools:
  - name: replace_text
    type: client
    description: Finds an exact string in the open document and replaces it.
    client:
      input_schema:
        type: object
        properties:
          find: {type: string, description: The exact text to find}
          replace: {type: string, description: The replacement text}
        required: [find, replace]
```

**The schema MUST be nested under `client:`.** Putting `input_schema` at the tool's
top level is silently accepted by deploy and silently dropped — `AgentTool` has no
such field. You get a tool the model can see and call but pass no arguments to, so
it appears to run and does nothing. Tools that take no arguments keep working,
which makes the failure present as "reads work, writes don't".

The frontend supplies handlers keyed by tool name via `AgentChatProvider`'s
`clientToolHandlers` map. This composes with `TemplateAgentConfig`, so the agent
stays durable — system prompt, model and schemas in the registry — while its tools
execute in the browser. Tool names become a contract across two repos: rename
either side and you get a call nothing answers, with no error.

#### 5. Deploy

```bash
belt agent deploy ./<agent-name>.yml
```

Same name = update. New name = create. The CLI returns the agent ref.

**Important**: deploy does NOT validate integration IDs or tool configurations. Invalid tools only fail at runtime. Always test after deploying.

#### 6. Test

Run the agent with a test message:
```bash
belt agent run <namespace/agent-name> "test message"
belt agent run <namespace/agent-name> "test message" --context key=value
```

**What to check:**
- Agent responds correctly to the test message
- Tools are actually called (not just described)
- Context variables are received by the agent
- MCP tools don't return "integration not found" — if they do, the integration_id is wrong. Re-check with `belt integrations list --json`
- Error cases are handled in the system prompt

**Verify the schemas actually stored** — deploy accepts misplaced fields silently,
so a tool can exist with no parameters at all. That failure is invisible from the
outside: the agent calls the tool, the tool does nothing, and the model often
reports success.

```bash
belt agent get <namespace/agent-name> --json \
  | jq '.version.tools[] | {name, type, params: (.client.input_schema // .call.input_schema).properties | keys?}'
```

Any tool showing `null` params that should take arguments has a misplaced
`input_schema`.

If something is wrong, edit the YAML and redeploy:
```bash
belt agent deploy ./<agent-name>.yml
```

#### 7. Iterate and share

Pull the deployed config to keep editing it:
```bash
belt agent pull <namespace/agent-name> --save ./<agent-name>.yml
```

`pull` renders YAML from the stored agent — it is a *rendering*, not the stored
state, and has historically omitted whole config blocks. Use `belt agent get
--json` when you need to know what is actually stored. Before a pull → edit →
deploy round trip, confirm the pulled YAML still contains every tool's config
block, or redeploying will strip whatever pull failed to render.

The agent is now callable:
```bash
belt agent run <namespace/agent-name> "message"
```

Continue a conversation:
```bash
belt agent run <namespace/agent-name> "follow up" --chat <chat_id>
```

### Troubleshooting

| Problem | Cause | Fix |
|---|---|---|
| "integration not found" at runtime | wrong integration_id in YAML | run `belt integrations list --json`, use the `id` field |
| "not connected" from `belt mcp tools` | MCP server not connected yet | run `belt mcp connect <slug>` first (needs browser) |
| `belt mcp connect` shows nothing | OAuth flow needs a browser | run the command in a local terminal, not headless |
| agent ignores tools | system prompt doesn't mention them | explicitly tell the agent when/how to use each tool |
| context variables not received | missing `--context` flag at runtime | pass with `belt agent run ... --context key=value` |

### Rules

- **System prompt is everything**: be specific about workflow steps, error handling, and when to use which tool. Vague prompts produce vague agents.
- **Always test after deploy**: deploy doesn't validate tools. A successful deploy does not mean the agent works.
- **Pull before you write**: `belt agent pull` on an existing agent is the best reference for YAML structure.
- **Secrets before call tools**: run `belt secrets set <name> <value>` before deploying agents that use `auth.secret`.

### Quick create (no YAML)

For simple agents without custom API tools:

```bash
belt agent create my-agent "description" \
  --model openrouter/claude-sonnet-46 \
  --prompt "You are a helpful assistant that..." \
  --mcp <integration_id>:<tool_name>
```

Then pull the YAML to add more configuration:
```bash
belt agent pull <namespace>/my-agent --save my-agent.yml
```

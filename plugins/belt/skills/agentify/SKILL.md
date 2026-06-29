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

### Process

#### 1. Define the agent's purpose

Ask the user:
- **What does this agent do?** One sentence.
- **What inputs does it need?** These become context variables.
- **What tools does it need?** MCP servers, API calls, or built-in tools (memory, plan, widget).
- **What model should power it?** Default: `openrouter/claude-sonnet-46`. Use `openrouter/claude-haiku-45` for simple/high-volume agents.

If the conversation already contains a working workflow, extract these from context instead of asking.

#### 2. Check for existing agents

```bash
belt agent list
```

If a similar agent exists, pull and review it:
```bash
belt agent pull <namespace/name> --save /tmp/existing-agent.yml
```

Decide: create new or update existing.

#### 3. Discover available tools

**MCP tools** — browse connected MCP servers and their tools:
```bash
belt mcp list                          # connected servers
belt mcp tools <server-ref>            # tools on a server
belt mcp search "todoist"              # find new servers
```

**Integrations** — check what's connected for auth:
```bash
belt integrations list                 # connected integrations
belt integrations connect <provider>   # connect new ones
```

To use an MCP tool, you need the integration ID and tool name: `--mcp <integration_id>:<tool_name>`.

#### 4. Write the agent YAML

Create the agent definition. This is the full spec — every field is optional except `name`.

```yaml
name: <kebab-case-name>
description: <one line — what it does>
core_app:
  ref: openrouter/claude-sonnet-46
system_prompt: |
  <clear, direct instructions for the agent>
context:
  - name: <variable_name>
    description: <what this input is>
    required: true
skills:
  - name: <skill-name>
    skill_id: <namespace/skill-name>
    description: <when the agent should use this skill>
internal_tools:
  memory: true    # agent can store/recall knowledge
  plan: false     # agent can create execution plans
  widget: false   # agent can render UI widgets
tools:
  - name: <tool_name>
    type: call
    description: <what the tool does>
    call:
      method: GET|POST|PUT|DELETE
      url: https://api.example.com/endpoint/{{context.variable}}
      auth:
        type: bearer
        secret: <secret_name>
      input_schema:
        type: object
        properties:
          field:
            type: string
  - name: <mcp_tool_name>
    type: mcp
    mcp:
      integration_id: <integration_id>
      tool_name: <tool_name>
```

Save to a file:
```bash
# save as <agent-name>.yml in the current directory
```

#### 5. Deploy

```bash
belt agent deploy ./<agent-name>.yml
```

Same name = update. New name = create. The CLI returns the agent ref.

#### 6. Test

Run the agent with a test message:
```bash
belt agent run <namespace/agent-name> "test message" --context key=value
```

Verify:
- Agent responds correctly
- Tools are called as expected
- Context variables are passed through
- Error cases are handled in the system prompt

If something is wrong, edit the YAML and redeploy:
```bash
belt agent deploy ./<agent-name>.yml
```

#### 7. Share the ref

The agent is now callable by anyone with access:
```bash
belt agent run <namespace/agent-name> "message"
```

Or via the API for programmatic use.

### Rules

- **System prompt is everything**: agents are only as good as their instructions. Be specific about workflow steps, error handling, and when to use which tool.
- **Context variables are typed inputs**: use them for IDs, refs, and parameters the agent needs at runtime. Don't hardcode values that should be dynamic.
- **MCP tools need connected integrations**: if using MCP tools, ensure the integration is connected first with `belt integrations connect`.
- **Secrets for API auth**: use `belt secrets set <name> <value>` to store API keys, then reference them in tool auth as `secret: <name>`.
- **Test before sharing**: always run the agent with a real message before declaring it done.

### Quick create (no YAML)

For simple agents without custom API tools:

```bash
belt agent create my-agent "description" \
  --model openrouter/claude-sonnet-46 \
  --prompt "You are a helpful assistant that..." \
  --mcp <integration_id>:<tool_name>
```

Then pull the YAML if you need to add more configuration later:
```bash
belt agent pull <namespace>/my-agent --save my-agent.yml
```

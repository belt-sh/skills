---
name: flowify
description: "Build and deploy an inference.sh flow — chain multiple apps into a pipeline with wired inputs/outputs. Use when the user says 'flowify', 'make a flow', 'chain these apps', 'build a pipeline', or when multiple apps should run in sequence with outputs feeding into inputs."
allowed-tools: Bash(belt *), Read, Write, Edit, Glob, Grep, Agent
---

## Flowify

Turn a multi-step process into a deployed inference.sh flow. Flows chain apps together — output from one feeds into the next. They run as a single unit via `belt flow run` or the API.

### When to use

- Multiple apps should run in sequence (e.g. generate text → TTS → voice change)
- The user says "flowify", "make a flow", "chain these", "build a pipeline"
- A workflow manually pipes output from one app into another
- The conversation involved running multiple `belt app run` calls in sequence

### Key concepts

- **Nodes** are apps in the flow. Each node runs one app.
- **Edges** define execution order (source → target means source runs first).
- **Connections** wire a node's output field to another node's input field.
- **Flow inputs** are the top-level inputs exposed to the caller.
- **Flow outputs** map a node's output to the flow's final output.

### Process

#### 0. Analyze the conversation first [MANDATORY]

Before asking anything, review what happened in this conversation. Look for:
- Sequential `belt app run` calls where output of one was used as input to another
- Multi-step processing pipelines (generate → transform → output)
- Manual orchestration that could be automated as a flow

If you find flow candidates, present them:
> "Based on this session, here's what could become a flow:
> 1. **<name>** — <app1> → <app2> → <app3>, takes <input>, returns <output>
> Which one should we build? Or describe something different."

#### 1. Identify the apps

List available apps and find the ones needed:
```bash
belt app search "<query>"
belt app get <namespace/app-name>     # check input/output schemas
```

For each app in the pipeline, note its input fields and output fields — you'll need these to wire connections.

#### 2. Create the flow

```bash
belt flow create --name <flow-name>
```

Save the flow ID — you'll need it for all edit commands.

#### 3. Add nodes

Add each app as a node:
```bash
belt flow edit node add <flow-id> <node-name> --app <namespace/app-name>
```

Node names should be descriptive (e.g. `tts`, `upscale`, `emotion-tagger`).

#### 4. Wire edges (execution order)

Connect nodes in sequence:
```bash
belt flow edit edge add <flow-id> <source-node> <target-node>
```

Edges define which node runs first. Source completes before target starts.

#### 5. Set inputs and connections

**Static values** (hardcoded config):
```bash
belt flow edit node set-input <flow-id> <node-name> <key> <value>
```

**Connections** (wire output of one node to input of another):
```bash
belt flow edit node set-input <flow-id> <node-name> <key> --from <source-node>:<output-key>
```

**Flow-level inputs** (exposed to the caller):
```bash
belt flow edit node set-input <flow-id> <node-name> <key> --from input:<field-name>
```

#### 6. Set flow input schema

Define what the caller passes in:
```bash
belt flow edit set-schema <flow-id> --field "text:string:required:Plain text to process"
```

#### 7. Set flow outputs

Map a node's output to the flow's final output:
```bash
belt flow edit output set <flow-id> <output-name> --from <node-name>:<output-key>
```

#### 8. Verify the graph

```bash
belt flow describe <flow-id>
```

Check:
- All nodes have their required inputs set (either static values, connections, or flow inputs)
- Edges form a valid DAG (no cycles)
- Flow outputs are mapped
- No dangling nodes (every node should be reachable from an edge)

#### 9. Test

```bash
belt flow run <flow-id> --input '{"text": "test input"}'
belt flow run <flow-id> --input input.json
```

Follow progress:
```bash
belt flow run-get <run-id> -f          # follow/stream progress
```

#### 10. Publish (optional)

```bash
belt flow publish <flow-id>
```

### Example: text → expressive TTS pipeline

```bash
# Create flow
belt flow create --name expressive-tts

# Add nodes
belt flow edit node add <fid> emotion-tagger --app openrouter/any-model
belt flow edit node add <fid> tts --app elevenlabs/tts

# Wire execution order
belt flow edit edge add <fid> emotion-tagger tts

# Set emotion-tagger config
belt flow edit node set-input <fid> emotion-tagger model google/gemini-2.5-flash
belt flow edit node set-input <fid> emotion-tagger system_prompt "Tag text with ElevenLabs emotion markers..."
belt flow edit node set-input <fid> emotion-tagger text --from input:text

# Wire tagger output → TTS input
belt flow edit node set-input <fid> tts text --from emotion-tagger:response
belt flow edit node set-input <fid> tts model eleven_v3
belt flow edit node set-input <fid> tts voice_id "p14E9FuOqbWvwJH0YFio"

# Set flow input/output
belt flow edit set-schema <fid> --field "text:string:required:Text to speak"
belt flow edit output set <fid> audio --from tts:audio

# Verify
belt flow describe <fid>

# Test
belt flow run <fid> --input '{"text": "Hello world!"}'
```

### Troubleshooting

| Problem | Cause | Fix |
|---|---|---|
| "node not found" | wrong node name | check `belt flow describe <fid>` for exact names |
| node runs but output is empty | wrong output key name | check `belt app get <app>` for output schema |
| flow hangs on a node | node's required inputs not set | check describe output for missing `←` on required fields |
| "cycle detected" | edges form a loop | flows must be DAGs — remove the back-edge |
| wrong data flows between nodes | connection uses wrong field name | check both source output and target input schemas with `belt app get` |

### Rules

- **Always check schemas first** — `belt app get <app>` shows exact input/output field names. Guessing field names is the #1 cause of broken flows.
- **Describe after every change** — `belt flow describe <fid>` is your verification tool. Run it after adding nodes, edges, and connections.
- **Static values are strings** — `belt flow edit node set-input` passes values as strings. Booleans are `"true"/"false"`, numbers are `"42"`.
- **Flow inputs need schema** — if you want the caller to pass data, set the flow input schema with `set-schema`.
- **Test incrementally** — for complex flows, test each app individually with `belt app run` first, then wire them together.

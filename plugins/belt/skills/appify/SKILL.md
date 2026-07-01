---
name: appify
description: "Build and deploy an inference.sh app from a conversation — scaffold, implement, test, deploy, and configure pricing. Use when the user says 'appify', 'make an app', 'deploy this as an app', or when a working API integration, model wrapper, or processing pipeline should become a reusable cloud app."
allowed-tools: Bash(belt *), Read, Write, Edit, Glob, Grep, Agent
---

## Appify

Turn an API integration, model wrapper, or processing pipeline from this conversation into a deployed inference.sh app. Apps run in the cloud, have typed input/output schemas, and are callable via `belt app run` or the API.

### When to use

- A working API integration should become a reusable cloud app
- The user says "appify", "make an app", "deploy this"
- A script or notebook should become a callable service
- A model needs to be wrapped with pre/post processing

### Process

#### 0. Analyze the conversation first [MANDATORY]

Before asking anything, review what happened in this conversation. Look for:
- API calls that worked (HTTP clients, SDK usage, model inference)
- Data processing pipelines (transform input → call API → format output)
- Working scripts that could be packaged as cloud functions

If you find app candidates, present them:
> "Based on this session, here's what could become an app:
> 1. **<name>** — <what it does>, takes <input>, returns <output>
> 2. **<name>** — <what it does>, takes <input>, returns <output>
> Which one should we build? Or describe something different."

Only ask "what kind of app?" if the conversation has no relevant context.

#### 1. Scaffold

```bash
belt app init <app-name>              # python (default)
belt app init <app-name> --lang node  # node.js
```

This creates the directory with `inf.yml`, `inference.py`/`inference.js`, and schema files. Never create these by hand.

**Exception:** when adding to an existing provider directory with shared helpers and symlinks (e.g. `api/pruna/` with `helper.py`), create files manually since `belt app init` doesn't support symlinked shared modules.

#### 2. Configure inf.yml

Key fields:
```yaml
name: my-app
description: what it does
version: 0.0.1
category: image|video|audio|text|3d|search|utility
python: "3.11"  # or node: "20"
```

#### 3. Implement inference.py / inference.js

The scaffold generates `RunInput`, `RunOutput`, and `App` classes. Edit these to match your app's purpose.

**Python pattern:**
```python
from inferencesh import BaseApp, BaseAppSetup, File, OutputMeta
from pydantic import BaseModel, Field
from typing import Optional

class RunInput(BaseModel):
    text: str = Field(description="text to process")
    max_length: int = Field(default=100, description="max output length")

class RunOutput(BaseModel):
    result: str = Field(description="processed result")

class App(BaseApp):
    async def setup(self, config: BaseAppSetup):
        # one-time init (load models, create clients)
        pass

    async def run(self, input_data: RunInput) -> RunOutput:
        # the actual work
        return RunOutput(result="...")
```

**Apps that call other inference.sh apps** — use the `AsyncInference` client:
```python
from inferencesh import BaseApp, BaseAppSetup, AsyncInference
from pydantic import BaseModel, Field

class RunInput(BaseModel):
    text: str = Field(description="text to process")

class RunOutput(BaseModel):
    result: str = Field(description="processed result")

class App(BaseApp):
    async def setup(self, config: BaseAppSetup):
        self.client = AsyncInference()

    async def run(self, input_data: RunInput) -> RunOutput:
        response = await self.client.run(
            app="openrouter/any-model",
            input={"text": input_data.text, "model": "google/gemini-2.5-flash"}
        )
        return RunOutput(result=response["response"])
```

Always check actual input/output field names with `belt app get <app> --json` before calling. Common gotcha: openrouter/any-model uses `text` not `prompt`, returns `response` not `text`.

Add `aiohttp` to `requirements.txt` if using `AsyncInference`.

**For local testing of apps that call other apps**, set your API key:
```bash
INFERENCE_API_KEY=$(belt secrets get INFSH_API_KEY 2>/dev/null || belt me --json | jq -r .api_key) belt app test --input input.json
```

**For upstream API costs**, use `RawMeta(cost=cents)` in output_meta:
```python
from inferencesh import RawMeta

class RunOutput(BaseModel):
    result: str
    output_meta: dict = None

# in run():
return RunOutput(
    result=result,
    output_meta=RawMeta(cost=cost_in_cents)
)
```

#### 4. Test locally

```bash
belt app test --save-example          # generate sample input
belt app test --input input.json      # test with sample
belt app test --input '{"prompt": "test"}'  # inline test
belt app test --debug                 # debug mode
```

Fix any errors before deploying.

#### 5. Deploy

```bash
belt app deploy                       # deploy to cloud
belt app deploy --dry-run             # validate without deploying
belt app deploy --stage               # deploy as staged version
```

#### 6. Test in cloud

```bash
belt app run <namespace/app-name> --input '{"prompt": "test"}'
belt app run <namespace/app-name> --input input.json --save output.png
```

#### 7. Configure pricing (if publishing to store)

Use the pricing skill:
```bash
belt skill use infsh/configure-pricing-for-inference-sh-apps
```

Or use the pricing agent directly:
```bash
belt agent run infsh/pricing-agent "configure pricing" --context version_id=<vid> --context app_id=<aid>
```

### Troubleshooting

| Problem | Cause | Fix |
|---|---|---|
| `belt app test` fails with import error | missing dependency | add to `requirements.txt` or `package.json` |
| deploy succeeds but app fails at runtime | dependency not in requirements.txt | check `belt task logs <task-id>` |
| "no inf.yml found" | wrong directory | `cd` into the app directory first |
| upstream API returns non-200 success | some APIs return 201/202 | check `status_code not in (200, 201)` not `!= 200` |

### Rules

- **Always `belt app init`** to scaffold — don't create files by hand (except shared-helper provider directories)
- **Always test locally** before deploying — `belt app test` catches most issues
- **Use `RawMeta(cost=cents)`** for upstream API costs — the platform needs this for pricing
- **Deploy then test in cloud** — local and cloud can differ (dependencies, GPU, env vars)
- **Use the pricing skill** for pricing — don't write CEL expressions by hand

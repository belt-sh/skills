"""Belt hook handler for Hermes agent."""
import subprocess
import json


async def handle(event_type: str, context: dict) -> dict:
    """Route Hermes hook events to belt plugin hook commands."""
    event_map = {
        "pre_llm_call": "user-prompt-submit",
        "pre_tool_call": "pre-tool-use",
        "post_tool_call": "post-tool-use",
        "on_session_end": "stop",
    }

    belt_event = event_map.get(event_type)
    if not belt_event:
        return {}

    try:
        result = subprocess.run(
            ["belt", "plugin", "hook", belt_event],
            input=json.dumps(context),
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.stdout.strip():
            return json.loads(result.stdout)
    except (subprocess.TimeoutExpired, json.JSONDecodeError, FileNotFoundError):
        pass

    return {}

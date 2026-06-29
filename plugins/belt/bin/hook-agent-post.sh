#!/bin/bash
# PostToolUse hook for Bash tool — detect belt agent deploy/create and suggest review
# Input: {"tool_name":"Bash","tool_input":{"command":"belt agent deploy ..."},"tool_result":"..."}
# Output: systemMessage suggesting agent review, or empty JSON

source "$(dirname "$0")/utils.sh"
read_input
log_hook "PostToolUse:Bash:agent" 500

# Extract the command that was run
command=$(json_get "command")

# Only care about belt agent deploy/create commands
case "$command" in
  *"belt agent deploy"*|*"belt agent create"*)
    ;;
  *)
    echo '{}'
    exit 0
    ;;
esac

# Extract the agent ref from the tool result if available
tool_result=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('tool_result', '')[:2000])
except:
    pass
" 2>/dev/null)

# Check if deploy/create succeeded (look for namespace/name pattern in output)
if ! echo "$tool_result" | grep -qE '[a-z]+/[a-z]'; then
  echo '{}'
  exit 0
fi

# Suggest running agent-evolver review
printf '{"systemMessage":"Agent deployed. Consider spawning the belt:agent-evolver subagent to review this agent before sharing — it validates tools, checks scope, and catches common issues."}'
exit 0

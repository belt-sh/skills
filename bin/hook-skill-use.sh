#!/bin/bash
# PreToolUse hook for Skill tool — search belt for related resources and inject
# Input: {"tool_name":"Skill","tool_input":{"skill":"name","args":"..."},...}
# Output: systemMessage with related belt resources (or empty JSON)

source "$(dirname "$0")/utils.sh"
read_input
log_hook "PreToolUse:Skill" 500

# Extract the skill name being loaded
skill_name=$(echo "$INPUT" | grep -o '"skill":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$skill_name" ]; then
  echo '{}'
  exit 0
fi

# Search belt for related resources
related=$(belt suggest "$skill_name" --json --agent 2>/dev/null)

if [ -z "$related" ]; then
  echo '{}'
  exit 0
fi

# Extract additionalContext — it's already formatted with score labels
# The grep pattern handles the escaped JSON string
context=$(echo "$related" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    ctx = d.get('hookSpecificOutput', {}).get('additionalContext', '')
    if ctx:
        # Escape for JSON string embedding
        print(json.dumps(ctx)[1:-1])
except:
    pass
" 2>/dev/null)

if [ -z "$context" ]; then
  echo '{}'
  exit 0
fi

# Inject as systemMessage alongside the native skill
printf '{"systemMessage":"Belt has related resources for skill \\"%s\\":\\n%s"}' "$skill_name" "$context"
exit 0

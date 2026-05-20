#!/bin/bash
# PostToolUse hook for Skill tool — log what was loaded, try to capture to belt
# Input: {"tool_name":"Skill","tool_input":{"skill":"name"},"tool_result":"<skill content>"}
# Output: empty JSON (observation only) or systemMessage

source "$(dirname "$0")/utils.sh"
read_input
log_hook "PostToolUse:Skill" 1000

# Extract skill name and result
skill_name=$(echo "$INPUT" | grep -o '"skill":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$skill_name" ]; then
  echo '{}'
  exit 0
fi

# Log the skill load event with more detail
session_id=$(get_session_id)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [skill-loaded sid=$session_id] $skill_name" >> "$BELT_LOG"

# Extract the tool_result (the skill content that was loaded)
# Use python3 for reliable JSON extraction since tool_result can be large
skill_content=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    result = d.get('tool_result', '')
    if isinstance(result, str):
        print(result[:5000])  # cap at 5k for logging
except:
    pass
" 2>/dev/null)

if [ -z "$skill_content" ]; then
  echo '{}'
  exit 0
fi

# Log skill content summary (first 200 chars)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [skill-content sid=$session_id] $(echo "$skill_content" | head -c 200)" >> "$BELT_LOG"

# Check if this skill exists in belt registry
# Only check for skills that look like they have a namespace (contains /)
if echo "$skill_name" | grep -q '/'; then
  # Already namespaced — probably a belt skill, skip capture
  echo '{}'
  exit 0
fi

# For non-namespaced skills (native/plugin skills), check if belt has it
existing=$(belt skill search "$skill_name" --json 2>/dev/null | head -c 500)

if echo "$existing" | grep -qi "$skill_name" 2>/dev/null; then
  # Belt already has something similar — no action needed
  echo '{}'
  exit 0
fi

# Log that this is a skill we could potentially import
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [skill-capture-candidate sid=$session_id] $skill_name (not in belt)" >> "$BELT_LOG"

echo '{}'
exit 0

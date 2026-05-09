#!/bin/bash
# Runs after every Claude response — every 5th turn, asks Haiku if there's knowledge worth saving
# If Haiku finds something, outputs additionalContext nudging Claude to save it

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event
log_hook "Stop"

# Count user turns from the transcript
turns=$(count_turns)
# Only evaluate every 5th turn
[ $((turns % 5)) -ne 0 ] && exit 0

# Check claude CLI is available
command -v claude >/dev/null 2>&1 || exit 0

# Get the last assistant message from the hook input
msg=$(get_last_message)
# Skip if empty or too short to be interesting
[ ${#msg} -lt 50 ] && exit 0

# Ask Haiku to evaluate — one-shot, no tools, no session persistence
result=$(echo "$msg" | claude --bare --no-session-persistence -p --model haiku --max-turns 1 \
  "Does this assistant response contain a non-obvious insight worth saving as knowledge? Types: observation (debugging insight), concept (architecture pattern), preference (user preference), reference (API/tool behavior).

If YES, respond with EXACTLY this format, nothing else:
SAVE|type|title|one sentence summary

If NO, respond with exactly: SKIP

Response to evaluate:" 2>/dev/null)

# Log what Haiku said
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:haiku] turn=$turns result=$result" >> "$BELT_LOG"

# If Haiku says save, output additionalContext nudging Claude
if echo "$result" | grep -q "^SAVE|"; then
  # Parse the SAVE|type|title|summary format
  type=$(echo "$result" | cut -d'|' -f2)
  title=$(echo "$result" | cut -d'|' -f3)
  summary=$(echo "$result" | cut -d'|' -f4-)

  # Output JSON that Claude Code injects as a system reminder
  cat <<HOOKJSON
{"hookSpecificOutput":{"hookEventName":"Stop","additionalContext":"Knowledge worth saving detected: [$type] $title — $summary. Consider saving with: echo \"$summary\" | belt know upload - --name \"$title\" --type $type"}}
HOOKJSON
fi

exit 0

#!/bin/bash
# Runs after every Claude response — every 5th turn, asks Haiku if there's knowledge worth saving
# Mirrors prompt hook behavior: returns {ok, reason} format
# If ok:false, feeds reason back to Claude via decision:block

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

# Ask Haiku — returns raw text, we tell it to output JSON
raw=$(echo "$msg" | claude --bare --no-session-persistence -p --model haiku --max-turns 1 \
  'Does this contain a non-obvious insight worth saving as knowledge? Types: observation (debugging insight), concept (architecture pattern), preference (user preference), reference (API/tool behavior). Respond ONLY with JSON, no other text. If yes: {"ok":false,"reason":"Worth saving as [type]: [title] — [summary]"} If no: {"ok":true}' 2>/dev/null)

# Log what Haiku returned
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:haiku] turn=$turns raw=$raw" >> "$BELT_LOG"

# If Haiku said ok:false, feed reason back to Claude
if echo "$raw" | grep -q '"ok".*false'; then
  # Extract reason — grep for the value between "reason":" and next "
  reason=$(echo "$raw" | grep -ao '"reason":"[^"]*' | head -1 | sed 's/"reason":"//')
  # Decision block tells Claude to keep going with the reason as context
  echo "{\"decision\":\"block\",\"reason\":\"$reason. Save with: belt know upload\"}"
fi

exit 0

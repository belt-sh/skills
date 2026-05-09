#!/bin/bash
# Runs after every Claude response — logs and triggers evaluation every ~5 turns

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event
log_hook "Stop"

# Count user turns from the transcript
turns=$(count_turns)

# Track last evaluated turn count
last_file="$BELT_DIR/last_eval_$(get_session_id)"
last=$(cat "$last_file" 2>/dev/null || echo 0)

# Skip unless at least 5 turns since last evaluation
[ $((turns - last)) -lt 5 ] && exit 0

# Record this evaluation point
echo "$turns" > "$last_file"

# TODO: belt dream — evaluate last 5 turns for knowledge capture
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:evaluate] turn=$turns" >> "$BELT_LOG"
exit 0

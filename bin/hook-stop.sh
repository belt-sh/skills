#!/bin/bash
source "$(dirname "$0")/utils.sh"
read_input
log_hook "Stop"

# Per-session turn counter
sid=$(get_session_id)
[ -z "$sid" ] && exit 0

count_file="$BELT_DIR/stop_count_${sid}"
count=$(($(cat "$count_file" 2>/dev/null || echo 0) + 1))
echo "$count" > "$count_file"

# Only evaluate every 5th turn
[ $((count % 5)) -ne 0 ] && exit 0

# TODO: belt dream — evaluate last 5 turns for knowledge capture
log_hook "Stop:evaluate"
exit 0

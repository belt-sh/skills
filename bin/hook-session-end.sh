#!/bin/bash
source "$(dirname "$0")/utils.sh"
read_input
log_hook "SessionEnd"

# Clean up session counter
sid=$(get_session_id)
[ -n "$sid" ] && rm -f "$BELT_DIR/stop_count_${sid}"

echo "$INPUT" | belt claude session-end 2>/dev/null
exit 0

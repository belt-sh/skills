#!/bin/bash
source "$(dirname "$0")/utils.sh"
read_input
log_hook "Stop"

# Count assistant turns from transcript (8ms, survives plugin reloads)
transcript=$(echo "$INPUT" | json_get "transcript_path")
[ -z "$transcript" ] || [ ! -f "$transcript" ] && exit 0

turns=$(grep -c '"type":"assistant"' "$transcript" 2>/dev/null || echo 0)

# Only evaluate every 5th turn
[ $((turns % 5)) -ne 0 ] && exit 0

# TODO: belt dream — evaluate last 5 turns for knowledge capture
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:evaluate] turn=$turns" >> "$BELT_LOG"
exit 0

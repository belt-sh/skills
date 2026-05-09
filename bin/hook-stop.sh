#!/bin/bash
# Runs after every Claude response — logs and triggers evaluation every 5th turn
source "$(dirname "$0")/utils.sh"
read_input
log_hook "Stop"

# Count assistant turns from transcript, only evaluate every 5th
turns=$(count_turns)
[ $((turns % 5)) -ne 0 ] && exit 0

# TODO: belt dream — evaluate last 5 turns for knowledge capture
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:evaluate] turn=$turns" >> "$BELT_LOG"
exit 0

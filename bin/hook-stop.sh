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
# Log the count and modulo for debugging
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:debug] turns=$turns mod=$((turns % 5))" >> "$BELT_LOG"

# Only evaluate every 5th turn
[ $((turns % 5)) -ne 0 ] && exit 0

# TODO: belt dream — evaluate last 5 turns for knowledge capture
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:evaluate] turn=$turns" >> "$BELT_LOG"
exit 0

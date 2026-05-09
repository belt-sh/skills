#!/bin/bash
# Runs after every Claude response — logs and triggers evaluation every 5th turn

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event
log_hook "Stop"

# Count real user turns from the transcript
turns=$(count_turns)
# Skip unless this is every 5th turn
[ $((turns % 5)) -ne 0 ] && exit 0

# TODO: belt dream — evaluate last 5 turns for knowledge capture
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [Stop:evaluate] turn=$turns" >> "$BELT_LOG"
exit 0

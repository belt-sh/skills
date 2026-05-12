#!/bin/bash
# Runs after every Claude response — delegates all logic to belt claude review

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event
log_hook "Stop"

# Belt handles turn counting, message extraction, and Haiku evaluation
echo "$INPUT" | belt claude review 2>>"$BELT_LOG"

exit 0

#!/bin/bash
# Runs after every Claude response — delegates all logic to belt claude review

source "$(dirname "$0")/utils.sh"
read_input
log_hook "Stop"

# Belt handles turn counting, message extraction, and dual-path evaluation
echo "$INPUT" | belt plugin review --trigger stop 2>>"$BELT_LOG"

exit 0

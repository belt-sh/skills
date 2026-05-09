#!/bin/bash
# Runs when session closes — logs and calls belt for session evaluation

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event
log_hook "SessionEnd"

# Pass session data to belt for post-session evaluation
echo "$INPUT" | belt claude session-end 2>/dev/null
exit 0

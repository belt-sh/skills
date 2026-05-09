#!/bin/bash
# Runs when session closes — logs and calls belt for session evaluation
source "$(dirname "$0")/utils.sh"
read_input
log_hook "SessionEnd"

# Pass session data to belt for evaluation
echo "$INPUT" | belt claude session-end 2>/dev/null
exit 0

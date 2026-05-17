#!/bin/bash
# Runs when session closes — final knowledge extraction + session summary

source "$(dirname "$0")/utils.sh"
read_input
log_hook "SessionEnd"

# Fork-based review — last chance to extract knowledge from this session
echo "$INPUT" | belt claude review --force --trigger session-end 2>>"$BELT_LOG"

# Session summary for belt backend
echo "$INPUT" | belt claude session-end 2>/dev/null

exit 0

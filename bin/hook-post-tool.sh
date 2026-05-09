#!/bin/bash
# Runs after Write|Edit tool calls — logs file mutations for skill evolution tracking

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event (truncated to 200 chars)
log_hook "PostToolUse" 200
exit 0

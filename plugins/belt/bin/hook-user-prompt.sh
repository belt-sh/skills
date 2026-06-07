#!/bin/bash
# Runs on every user message — searches belt for relevant skills/knowledge/apps
# Stdout JSON with additionalContext is injected before Claude processes the message

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event (truncated to 200 chars since prompts can be long)
log_hook "UserPromptSubmit" 200

# Pipe the full input to belt suggest which returns additionalContext JSON
echo "$INPUT" | belt suggest --json 2>/dev/null
exit 0

#!/bin/bash
# Runs on every user message — searches belt for relevant skills/knowledge/apps
# Stdout JSON with additionalContext is injected before Claude processes the message
source "$(dirname "$0")/utils.sh"
read_input
log_hook "UserPromptSubmit" 200

# Search belt registry and return results as additionalContext JSON
echo "$INPUT" | belt suggest --json 2>/dev/null
exit 0

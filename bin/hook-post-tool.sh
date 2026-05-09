#!/bin/bash
# Runs after Write|Edit tool calls — logs file mutations for skill evolution tracking
source "$(dirname "$0")/utils.sh"
read_input
log_hook "PostToolUse" 200
exit 0

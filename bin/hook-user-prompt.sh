#!/bin/bash
source "$(dirname "$0")/utils.sh"
read_input
log_hook "UserPromptSubmit" 200

echo "$INPUT" | belt suggest --json 2>/dev/null
exit 0

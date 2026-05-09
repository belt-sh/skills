#!/bin/bash
mkdir -p ~/.belt
input=$(cat)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [UserPromptSubmit] $(echo $input | head -c 200)" >> ~/.belt/hooks.log
echo "$input" | belt suggest --json 2>/dev/null
exit 0

#!/bin/bash
mkdir -p ~/.belt
input=$(cat)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [PostToolUse] $(echo $input | head -c 200)" >> ~/.belt/hooks.log
exit 0

#!/bin/bash
mkdir -p ~/.belt
input=$(cat)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [SessionEnd] $(echo $input | head -c 500)" >> ~/.belt/hooks.log
echo "$input" | belt claude session-end 2>/dev/null
exit 0

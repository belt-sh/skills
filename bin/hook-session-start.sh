#!/bin/bash
mkdir -p ~/.belt
input=$(cat)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [SessionStart] $(echo $input | head -c 500)" >> ~/.belt/hooks.log
exit 0

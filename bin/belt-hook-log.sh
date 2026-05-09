#!/bin/bash
# Log every hook invocation to ~/.belt/hooks.log
# Usage: belt-hook-log.sh <hook-name> [stdin-data]
#
# Called at the top of every hook command. Logs timestamp, hook name, and stdin snippet.

LOG="$HOME/.belt/hooks.log"
mkdir -p "$(dirname "$LOG")"

hook_name="${1:-unknown}"
stdin_data=""

# Read stdin if available (non-blocking)
if [ ! -t 0 ]; then
  stdin_data=$(head -c 500)
fi

# One line per invocation: timestamp, hook, first 200 chars of stdin
snippet=$(echo "$stdin_data" | tr '\n' ' ' | cut -c1-200)
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [$hook_name] $snippet" >> "$LOG"

# Echo stdin back so the calling script can still read it
if [ -n "$stdin_data" ]; then
  echo "$stdin_data"
fi

#!/bin/bash
# Shared utilities for belt hooks

BELT_DIR="$HOME/.belt"
BELT_LOG="$BELT_DIR/hooks.log"

# Ensure belt directory exists
mkdir -p "$BELT_DIR"

# Pick best JSON tool available
if command -v jt >/dev/null 2>&1; then
  json_get() { jt ".$1" 2>/dev/null | tr -d '"'; }
elif command -v jq >/dev/null 2>&1; then
  json_get() { jq -r ".$1 // empty" 2>/dev/null; }
else
  json_get() { grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4; }
fi

# Read stdin once, store it
read_input() {
  INPUT=$(cat)
}

# Log a hook event
log_hook() {
  local event="$1"
  local max="${2:-500}"
  echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [$event] $(echo $INPUT | head -c $max)" >> "$BELT_LOG"
}

# Get session ID from input
get_session_id() {
  echo "$INPUT" | json_get "session_id"
}

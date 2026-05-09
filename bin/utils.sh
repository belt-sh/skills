#!/bin/bash
# Shared utilities for belt hooks — zero external dependencies (no jq/jt required)

BELT_DIR="$HOME/.belt"
BELT_LOG="$BELT_DIR/hooks.log"

mkdir -p "$BELT_DIR"

# Extract a top-level string value from JSON using grep
json_get() {
  echo "$INPUT" | grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

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
  json_get "session_id"
}

# Get transcript path from input
get_transcript_path() {
  json_get "transcript_path"
}

# Count assistant turns from transcript (fast grep, ~8ms)
count_turns() {
  local transcript=$(get_transcript_path)
  [ -z "$transcript" ] || [ ! -f "$transcript" ] && echo 0 && return
  grep -c '"type":"assistant"' "$transcript" 2>/dev/null || echo 0
}

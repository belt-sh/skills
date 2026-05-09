#!/bin/bash
# Shared utilities for belt plugin hooks — zero external deps (no jq/jt required)

# Where belt stores logs and state
BELT_DIR="$HOME/.belt"
BELT_LOG="$BELT_DIR/hooks.log"
mkdir -p "$BELT_DIR"

# Extract a top-level string value from JSON in $INPUT using grep
json_get() {
  echo "$INPUT" | grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

# Read all of stdin into $INPUT — call once, reuse everywhere
read_input() {
  INPUT=$(cat)
}

# Append timestamped log line: [EventName] truncated input
log_hook() {
  local event="$1"
  local max="${2:-500}"
  echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [$event] $(echo $INPUT | head -c $max)" >> "$BELT_LOG"
}

# Field extractors for common hook input fields
get_session_id() { json_get "session_id"; }
get_transcript_path() { json_get "transcript_path"; }

# Count assistant messages in transcript via grep (~8ms on 2500 lines)
count_turns() {
  local transcript=$(get_transcript_path)
  [ -z "$transcript" ] || [ ! -f "$transcript" ] && echo 0 && return
  grep -c '"type":"assistant"' "$transcript" 2>/dev/null || echo 0
}

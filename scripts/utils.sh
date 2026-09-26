#!/bin/bash
# Shared utilities for belt plugin hooks — zero external deps (no jq/jt required)

# Where belt stores logs and state
BELT_DIR="$HOME/.belt"
# Log file for all hook events
BELT_LOG="$BELT_DIR/hooks.log"
# Create the belt directory if it doesn't exist
mkdir -p "$BELT_DIR"

# Extract a top-level string value from JSON in $INPUT using grep
# Works without jq/jt — matches "key":"value" pattern
json_get() {
  echo "$INPUT" | grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

# Read all of stdin into $INPUT — call once, reuse everywhere
read_input() {
  INPUT=$(cat)
}

# Append timestamped log line: [EventName] truncated input
log_hook() {
  # First arg is the hook event name
  local event="$1"
  # Second arg is max chars to log, defaults to 500
  local max="${2:-500}"
  # Write UTC timestamp, event name, and truncated input to log file
  echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [$event] $(echo $INPUT | head -c $max)" >> "$BELT_LOG"
}

# Extract session_id from hook input
get_session_id() { json_get "session_id"; }
# Extract transcript_path from hook input
get_transcript_path() { json_get "transcript_path"; }

# Count user turns from transcript using grep (~8ms on 2500 lines)
count_turns() {
  # Get path to the session's JSONL transcript
  local transcript=$(get_transcript_path)
  # If no transcript or file missing, return 0
  [ -z "$transcript" ] || [ ! -f "$transcript" ] && echo 0 && return
  # Count lines with type user
  grep -c '"type":"user"' "$transcript" 2>/dev/null || echo 0
}


# Extract last_assistant_message from hook input
# This is the most recent Claude response, already in $INPUT
get_last_message() {
  # Extract value between "last_assistant_message":" and next unescaped quote
  # Truncate to 3000 chars to keep Haiku cost low
  echo "$INPUT" | grep -ao '"last_assistant_message":"[^}]*' | head -1 | \
    sed 's/"last_assistant_message":"//' | sed 's/"$//' | head -c 3000
}

#!/bin/bash
# Runs on session start — checks belt is installed and authenticated
# Stdout is added to Claude's context as a system reminder

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event to ~/.belt/hooks.log
log_hook "SessionStart"

# Check belt CLI is installed
if ! command -v belt >/dev/null 2>&1; then
  # Tell Claude that belt needs installing
  echo "belt CLI is not installed. Install: curl -fsSL https://belt.sh/install | sh"
  exit 0
fi

# Check belt is authenticated
if ! belt me >/dev/null 2>&1; then
  # Tell Claude that login is needed
  echo "belt is installed but not authenticated. Run: belt login"
  exit 0
fi

# Show current user — Claude sees this as session context
user=$(belt me 2>/dev/null | head -1)
echo "belt: $user"
exit 0

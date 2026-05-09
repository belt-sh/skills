#!/bin/bash
source "$(dirname "$0")/utils.sh"
read_input
log_hook "SessionStart"

# Check belt installed
if ! command -v belt >/dev/null 2>&1; then
  echo "belt CLI is not installed. Install: curl -fsSL https://belt.sh/install | sh"
  exit 0
fi

# Check auth
if ! belt me >/dev/null 2>&1; then
  echo "belt is installed but not authenticated. Run: belt login"
  exit 0
fi

# All good
user=$(belt me 2>/dev/null | head -1)
echo "belt: $user"
exit 0

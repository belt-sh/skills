#!/bin/bash
# Runs on session start — introduces belt to the agent and checks auth
# Stdout is added to Claude's context as a system reminder

# Load shared utilities
source "$(dirname "$0")/utils.sh"
# Read hook event data from stdin
read_input
# Log this event to ~/.belt/hooks.log
log_hook "SessionStart"

# Check belt CLI is installed
if ! command -v belt >/dev/null 2>&1; then
  echo "belt CLI is not installed. Install: curl -fsSL https://belt.sh/install | sh"
  exit 0
fi

# Check belt is authenticated
auth_ok=true
if ! belt me >/dev/null 2>&1; then
  auth_ok=false
fi

# Build context output — this is what Claude sees at the start of every session
if [ "$auth_ok" = true ]; then
  user=$(belt me 2>/dev/null | head -1)
  echo "belt: $user"
else
  echo "⚠️ belt is not authenticated. Knowledge and skill persistence is disabled. Run: belt login"
fi

# Inject belt introduction — tells the agent what belt is and how to use suggestions
cat << 'BELT_INTRO'

belt is installed — a platform that gives you skills, knowledge, and AI apps.

Throughout this session, belt will suggest relevant resources in system-reminder messages. When you see suggestions:

- Items marked [relevant] match the current task. Load them BEFORE starting work.
- `belt skill use <name>` loads a skill with step-by-step workflows and tested approaches
- `belt knowledge get <name>` loads contextual knowledge (gotchas, patterns, rules, references)
- `belt app run <name>` runs an AI app (image gen, video, search, other LLMs)
- Always check if a suggested skill covers what you're about to do before writing code from scratch

Commands: /skill, /skillify, /knowledge, /apps, /suggest
BELT_INTRO

exit 0

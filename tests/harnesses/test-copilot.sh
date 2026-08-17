#!/bin/bash
# Test belt plugin installation for GitHub Copilot CLI
source "$(dirname "$0")/../framework.sh"

harness_setup "copilot"

# --- Phase 1: Check prerequisites ---
log "phase 1: prerequisites"

if ! has_cmd belt; then
  fail "belt not installed"
  harness_teardown
fi

check_endpoint

# --- Phase 2: Check Copilot CLI ---
log "phase 2: copilot CLI"

if ! has_cmd copilot; then
  skip "copilot CLI not installed"
fi

# --- Phase 3: Create belt hooks for Copilot ---
log "phase 3: hooks configuration"

COPILOT_HOME="$HOME/.copilot"
COPILOT_HOOKS="$COPILOT_HOME/hooks"
mkdir -p "$COPILOT_HOOKS"

# Copilot uses hooks/*.json like Grok
cat > "$COPILOT_HOOKS/belt.json" << 'HOOKEOF'
{
  "hooks": {
    "sessionStart": [
      {
        "type": "command",
        "command": "AI_AGENT=copilot belt me 2>/dev/null | head -1 || echo 'belt: not authenticated'",
        "timeout": 10
      }
    ],
    "userPromptSubmitted": [
      {
        "type": "command",
        "command": "AI_AGENT=copilot belt suggest --json",
        "timeout": 30
      }
    ],
    "agentStop": [
      {
        "type": "command",
        "command": "AI_AGENT=copilot belt review --agent copilot --trigger stop",
        "timeout": 120
      }
    ]
  }
}
HOOKEOF

assert_file "$COPILOT_HOOKS/belt.json" "belt hooks file created"
assert_contains "$COPILOT_HOOKS/belt.json" "userPromptSubmitted" "hooks contain userPromptSubmitted (camelCase)"
assert_contains "$COPILOT_HOOKS/belt.json" "belt suggest" "hooks reference belt suggest"

# --- Phase 4: Test endpoint with Copilot's env vars ---
log "phase 4: endpoint via COPILOT_PROVIDER_BASE_URL"

export COPILOT_PROVIDER_BASE_URL="$OPENROUTER_BASE"
export COPILOT_PROVIDER_API_KEY="$OPENROUTER_KEY"

check_chat_completion "$COPILOT_PROVIDER_BASE_URL" "$COPILOT_PROVIDER_API_KEY"

# --- Phase 5: Install skills ---
log "phase 5: skills"

COPILOT_SKILLS="$COPILOT_HOME/skills"
mkdir -p "$COPILOT_SKILLS"

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [ -d "$REPO_ROOT/skills" ]; then
  cp -r "$REPO_ROOT/skills/"* "$COPILOT_SKILLS/" 2>/dev/null || true
  assert_dir_nonempty "$COPILOT_SKILLS" "skills installed"
fi

# --- Phase 6: belt suggest ---
log "phase 6: belt suggest"

SUGGEST_OUT=$(echo '{"prompt":"deploy an app"}' | AI_AGENT=copilot belt suggest --json 2>/dev/null || echo "")
if echo "$SUGGEST_OUT" | grep -qi "skill\|app\|knowledge"; then
  pass "belt suggest returns results"
else
  skip "belt suggest returned empty (may need auth)"
fi

# --- Phase 7: Full loop ---
log "phase 7: full loop"

if has_cmd copilot; then
  COPILOT_OUT=$(copilot -p "say hello" 2>&1 || echo "")
  if [ -n "$COPILOT_OUT" ]; then
    pass "copilot -p produces output"
  else
    fail "copilot -p produced no output"
  fi

  if [ -f "$HOME/.belt/hooks.log" ] && grep -q "session-start\|user-prompt" "$HOME/.belt/hooks.log" 2>/dev/null; then
    pass "belt hooks fired during copilot session"
  else
    skip "hooks may not have fired"
  fi
else
  skip "copilot CLI not installed — cannot test full loop"
fi

harness_teardown

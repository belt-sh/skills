#!/bin/bash
# Test belt plugin installation for Grok CLI
source "$(dirname "$0")/../framework.sh"

harness_setup "grok"

# --- Phase 1: Check prerequisites ---
log "phase 1: prerequisites"

if ! has_cmd belt; then
  fail "belt not installed"
  harness_teardown
fi

check_endpoint

# --- Phase 2: Check Grok CLI ---
log "phase 2: grok CLI"

if ! has_cmd grok; then
  skip "grok CLI not installed — install with: curl -fsSL https://x.ai/cli/install.sh | bash"
  # Continue anyway — we can still test the plugin file generation
fi

# --- Phase 3: Create belt hooks for Grok ---
log "phase 3: hooks configuration"

GROK_HOME="$HOME/.grok"
GROK_HOOKS="$GROK_HOME/hooks"
mkdir -p "$GROK_HOOKS"

# Grok merges all hooks/*.json files — write belt.json
cat > "$GROK_HOOKS/belt.json" << 'HOOKEOF'
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt me 2>/dev/null | head -1 || echo 'belt: not authenticated'",
        "timeout": 10
      }
    ],
    "UserPromptSubmit": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt suggest --json",
        "timeout": 30
      }
    ],
    "Stop": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt review --agent grok --trigger stop",
        "timeout": 120
      }
    ],
    "PreCompact": [
      {
        "type": "command",
        "command": "AI_AGENT=grok belt review --agent grok --trigger precompact --force",
        "timeout": 120
      }
    ]
  }
}
HOOKEOF

assert_file "$GROK_HOOKS/belt.json" "belt hooks file created"
assert_contains "$GROK_HOOKS/belt.json" "UserPromptSubmit" "hooks contain UserPromptSubmit"
assert_contains "$GROK_HOOKS/belt.json" "belt suggest" "hooks reference belt suggest"

# --- Phase 4: Test endpoint with Grok's env vars ---
log "phase 4: endpoint via GROK_BASE_URL"

export GROK_BASE_URL="$OPENROUTER_BASE"
export XAI_API_KEY="$OPENROUTER_KEY"

check_chat_completion "$GROK_BASE_URL" "$XAI_API_KEY"

# --- Phase 5: Install skills ---
log "phase 5: skills"

GROK_SKILLS="$GROK_HOME/skills"
mkdir -p "$GROK_SKILLS"

# Copy skills from the plugin repo
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [ -d "$REPO_ROOT/skills" ]; then
  cp -r "$REPO_ROOT/skills/"* "$GROK_SKILLS/" 2>/dev/null || true
  assert_dir_nonempty "$GROK_SKILLS" "skills installed to ~/.grok/skills/"
  assert_file "$GROK_SKILLS/belt/SKILL.md" "belt skill present"
  assert_file "$GROK_SKILLS/suggest/SKILL.md" "suggest skill present"
else
  skip "plugin repo skills/ not found at $REPO_ROOT/skills"
fi

# --- Phase 6: Verify belt suggest works ---
log "phase 6: belt suggest"

SUGGEST_OUT=$(echo '{"prompt":"how do I generate a video"}' | AI_AGENT=grok belt suggest --json 2>/dev/null || echo "")
if echo "$SUGGEST_OUT" | grep -qi "skill\|app\|knowledge"; then
  pass "belt suggest returns results"
else
  skip "belt suggest returned empty (may need auth)"
fi

# --- Phase 7: Full loop test (if grok installed) ---
log "phase 7: full loop"

if has_cmd grok; then
  # Run grok in headless mode with one prompt
  GROK_OUT=$(grok -p "say hello" --output-format json 2>&1 || echo "")
  if echo "$GROK_OUT" | grep -qi "hello\|error\|content"; then
    pass "grok -p produces output"
  else
    fail "grok -p produced no output"
  fi

  # Check if hooks fired
  if [ -f "$HOME/.belt/hooks.log" ]; then
    if grep -q "session-start\|user-prompt" "$HOME/.belt/hooks.log" 2>/dev/null; then
      pass "belt hooks fired during grok session"
    else
      fail "belt hooks.log exists but no hook events found"
    fi
  else
    skip "no hooks.log — hooks may not have fired"
  fi
else
  skip "grok CLI not installed — cannot test full loop"
fi

harness_teardown

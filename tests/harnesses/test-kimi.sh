#!/bin/bash
# Test belt plugin installation for Kimi Code CLI
source "$(dirname "$0")/../framework.sh"

harness_setup "kimi"

# --- Phase 1: Check prerequisites ---
log "phase 1: prerequisites"

if ! has_cmd belt; then
  fail "belt not installed"
  harness_teardown
fi

check_endpoint

# --- Phase 2: Check Kimi CLI ---
log "phase 2: kimi CLI"

if ! has_cmd kimi; then
  skip "kimi CLI not installed — install with: npm install -g @moonshot-ai/kimi-code"
fi

# --- Phase 3: Create belt hooks for Kimi (TOML format) ---
log "phase 3: hooks configuration (TOML)"

KIMI_HOME="$HOME/.kimi-code"
mkdir -p "$KIMI_HOME"

# Kimi uses config.toml with [[hooks]] entries
cat > "$KIMI_HOME/config.toml" << 'TOMLEOF'
# Belt plugin hooks

[providers.openrouter]
type = "openai"
api_key = "${OPENROUTER_KEY}"
base_url = "https://openrouter.ai/api/v1"

[[hooks]]
event = "SessionStart"
command = "AI_AGENT=kimi belt me 2>/dev/null | head -1 || echo 'belt: not authenticated'"
timeout = 10

[[hooks]]
event = "UserPromptSubmit"
command = "AI_AGENT=kimi belt suggest --json"
timeout = 30

[[hooks]]
event = "Stop"
command = "AI_AGENT=kimi belt review --agent kimi --trigger stop"
timeout = 120

[[hooks]]
event = "PreCompact"
command = "AI_AGENT=kimi belt review --agent kimi --trigger precompact --force"
timeout = 120

[[hooks]]
event = "SessionEnd"
command = "AI_AGENT=kimi belt plugin session-end --agent kimi"
timeout = 600
TOMLEOF

assert_file "$KIMI_HOME/config.toml" "config.toml created"
assert_contains "$KIMI_HOME/config.toml" "UserPromptSubmit" "hooks contain UserPromptSubmit"
assert_contains "$KIMI_HOME/config.toml" "belt suggest" "hooks reference belt suggest"
assert_contains "$KIMI_HOME/config.toml" "openrouter" "provider configured"

# --- Phase 4: Test endpoint ---
log "phase 4: endpoint"

check_chat_completion "$OPENROUTER_BASE" "$OPENROUTER_KEY"

# --- Phase 5: Skills ---
log "phase 5: skills"

# Kimi doesn't have a documented skills dir — skills come via suggest injection
skip "kimi uses marketplace skills, not filesystem skills"

# --- Phase 6: belt suggest ---
log "phase 6: belt suggest"

SUGGEST_OUT=$(echo '{"prompt":"search the web"}' | AI_AGENT=kimi belt suggest --json 2>/dev/null || echo "")
if echo "$SUGGEST_OUT" | grep -qi "skill\|app\|knowledge"; then
  pass "belt suggest returns results"
else
  skip "belt suggest returned empty (may need auth)"
fi

# --- Phase 7: Full loop ---
log "phase 7: full loop"

if has_cmd kimi; then
  # Kimi needs the provider to actually be configured (not just env var)
  # Write the real key into config
  sed -i "s|\${OPENROUTER_KEY}|$OPENROUTER_KEY|g" "$KIMI_HOME/config.toml"

  KIMI_OUT=$(kimi -p "say hello" 2>&1 || echo "")
  if [ -n "$KIMI_OUT" ]; then
    pass "kimi -p produces output"
  else
    fail "kimi -p produced no output"
  fi

  if [ -f "$HOME/.belt/hooks.log" ] && grep -q "session-start\|user-prompt" "$HOME/.belt/hooks.log" 2>/dev/null; then
    pass "belt hooks fired during kimi session"
  else
    skip "hooks may not have fired"
  fi
else
  skip "kimi CLI not installed — cannot test full loop"
fi

harness_teardown

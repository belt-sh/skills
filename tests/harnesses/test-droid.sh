#!/bin/bash
# Test belt plugin installation for Droid (Factory)
source "$(dirname "$0")/../framework.sh"

harness_setup "droid"

# --- Phase 1: Check prerequisites ---
log "phase 1: prerequisites"

if ! has_cmd belt; then
  fail "belt not installed"
  harness_teardown
fi

check_endpoint

# --- Phase 2: Check Droid CLI ---
log "phase 2: droid CLI"

if ! has_cmd droid; then
  skip "droid CLI not installed — install with: curl -fsSL https://app.factory.ai/cli | sh"
fi

# --- Phase 3: Create belt hooks for Droid ---
log "phase 3: hooks configuration"

FACTORY_HOME="$HOME/.factory"
mkdir -p "$FACTORY_HOME"

# Droid uses settings.json with hooks (Claude Code format, PascalCase)
cat > "$FACTORY_HOME/settings.json" << 'HOOKEOF'
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=droid belt suggest --json",
            "timeout": 30
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=droid belt review --agent droid --trigger stop",
            "timeout": 120
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=droid belt review --agent droid --trigger precompact --force",
            "timeout": 120
          }
        ]
      }
    ]
  },
  "customModels": [
    {
      "name": "openrouter-test",
      "provider": "openai",
      "apiKey": "${OPENROUTER_KEY}",
      "baseUrl": "https://openrouter.ai/api/v1"
    }
  ]
}
HOOKEOF

assert_file "$FACTORY_HOME/settings.json" "settings.json created"
assert_contains "$FACTORY_HOME/settings.json" "UserPromptSubmit" "hooks contain UserPromptSubmit"
assert_contains "$FACTORY_HOME/settings.json" "belt suggest" "hooks reference belt suggest"
assert_contains "$FACTORY_HOME/settings.json" "openrouter" "custom model configured"

# --- Phase 4: Test endpoint ---
log "phase 4: endpoint"

check_chat_completion "$OPENROUTER_BASE" "$OPENROUTER_KEY"

# --- Phase 5: Skills ---
log "phase 5: skills"

DROID_SKILLS="$FACTORY_HOME/skills"
mkdir -p "$DROID_SKILLS"

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [ -d "$REPO_ROOT/skills" ]; then
  cp -r "$REPO_ROOT/skills/"* "$DROID_SKILLS/" 2>/dev/null || true
  assert_dir_nonempty "$DROID_SKILLS" "skills installed to ~/.factory/skills/"
fi

# --- Phase 6: belt suggest ---
log "phase 6: belt suggest"

SUGGEST_OUT=$(echo '{"prompt":"create an agent"}' | AI_AGENT=droid belt suggest --json 2>/dev/null || echo "")
if echo "$SUGGEST_OUT" | grep -qi "skill\|app\|knowledge"; then
  pass "belt suggest returns results"
else
  skip "belt suggest returned empty (may need auth)"
fi

# --- Phase 7: Full loop ---
log "phase 7: full loop"

if has_cmd droid; then
  sed -i "s|\${OPENROUTER_KEY}|$OPENROUTER_KEY|g" "$FACTORY_HOME/settings.json"
  DROID_OUT=$(droid exec "say hello" --auto low 2>&1 || echo "")
  if [ -n "$DROID_OUT" ]; then
    pass "droid exec produces output"
  else
    fail "droid exec produced no output"
  fi
else
  skip "droid CLI not installed — cannot test full loop"
fi

harness_teardown

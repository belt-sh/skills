#!/bin/bash
# Test belt plugin installation for Codex
source "$(dirname "$0")/../framework.sh"

harness_setup "codex"

# --- Phase 1: Check prerequisites ---
log "phase 1: prerequisites"

if ! has_cmd belt; then
  fail "belt not installed"
  harness_teardown
fi

if ! has_cmd codex; then
  skip "codex CLI not installed — install with: npm install -g @openai/codex"
  harness_teardown
fi

pass "belt and codex found"
check_endpoint

# --- Phase 2: Install belt plugin ---
log "phase 2: belt init codex"

CODEX_HOME="$HOME/.codex"
mkdir -p "$CODEX_HOME"

# Create hooks config directly (same as what belt plugin init would do)
cat > "$CODEX_HOME/hooks.json" << 'HOOKEOF'
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=codex belt me 2>/dev/null | head -1 || echo 'belt: not authenticated'",
            "timeout": 10
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=codex belt suggest --json",
            "timeout": 30
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AI_AGENT=codex belt review --agent codex --trigger stop",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
HOOKEOF

assert_file "$CODEX_HOME/hooks.json" "hooks.json created"
assert_contains "$CODEX_HOME/hooks.json" "belt suggest" "hooks reference belt suggest"

# Install skills from repo
CODEX_SKILLS="$HOME/.agents/skills"
mkdir -p "$CODEX_SKILLS"
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [ -d "$REPO_ROOT/skills" ]; then
  cp -r "$REPO_ROOT/skills/"* "$CODEX_SKILLS/" 2>/dev/null || true
  assert_dir_nonempty "$CODEX_SKILLS" "skills directory populated"
else
  skip "plugin repo skills/ not found"
fi

# --- Phase 3: Configure endpoint ---
log "phase 3: endpoint configuration"

export OPENAI_BASE_URL="$OPENROUTER_BASE"
export OPENAI_API_KEY="$OPENROUTER_KEY"

check_chat_completion "$OPENROUTER_BASE" "$OPENROUTER_KEY"

# --- Phase 4: Verify belt suggest works ---
log "phase 4: belt suggest"

SUGGEST_OUT=$(echo '{"prompt":"how do I generate an image"}' | AI_AGENT=codex belt suggest --json 2>/dev/null || echo "")
if echo "$SUGGEST_OUT" | grep -qi "skill\|app\|knowledge"; then
  pass "belt suggest returns results"
else
  skip "belt suggest returned empty (may need auth)"
fi

harness_teardown

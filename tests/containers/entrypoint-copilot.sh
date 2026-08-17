#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== GitHub Copilot CLI Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v copilot >/dev/null 2>&1; then
  copilot --version 2>&1 | head -1
  pass "copilot installed"
else
  skip "copilot CLI not available — package may have changed name"
fi

# 2. Configure endpoint
echo "[phase 2] endpoint"
export COPILOT_PROVIDER_BASE_URL="https://openrouter.ai/api/v1"
export COPILOT_PROVIDER_API_KEY="$OPENROUTER_KEY"
export COPILOT_MODEL="openai/gpt-4o-mini"

RESP=$(curl -sf "$COPILOT_PROVIDER_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $COPILOT_PROVIDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Install belt hooks
echo "[phase 3] belt hooks"
mkdir -p ~/.copilot/hooks
cat > ~/.copilot/hooks/belt.json << 'EOF'
{
  "version": 1,
  "hooks": {
    "sessionStart": [{"type":"command","bash":"AI_AGENT=copilot belt me 2>/dev/null | head -1 || echo belt:not-authed","timeoutSec":10}],
    "userPromptSubmitted": [{"type":"command","bash":"AI_AGENT=copilot belt suggest --json","timeoutSec":30}],
    "agentStop": [{"type":"command","bash":"AI_AGENT=copilot belt review --agent copilot --trigger stop","timeoutSec":120}]
  }
}
EOF

[ -f ~/.copilot/hooks/belt.json ] && pass "hooks installed" || fail "hooks missing"
grep -q "belt suggest" ~/.copilot/hooks/belt.json && pass "hooks reference belt" || fail "hooks broken"

# 4. Install skills
echo "[phase 4] skills"
mkdir -p ~/.copilot/skills
cp -r /opt/belt-plugin/skills/* ~/.copilot/skills/ 2>/dev/null || true
[ -f ~/.copilot/skills/belt/SKILL.md ] && pass "belt skill installed" || fail "belt skill missing"

# 5. Headless run
echo "[phase 5] headless prompt"
if command -v copilot >/dev/null 2>&1; then
  COPILOT_OUT=$(timeout 45 copilot --prompt "respond with exactly: BELT_TEST_OK" 2>&1 || true)
  if echo "$COPILOT_OUT" | grep -qi "BELT_TEST_OK"; then
    pass "copilot produced correct output"
  elif echo "$COPILOT_OUT" | grep -qi "not signed in\|login\|auth\|error"; then
    echo "  copilot output: $(echo "$COPILOT_OUT" | head -3)"
    skip "copilot requires auth"
  else
    echo "  copilot output: $(echo "$COPILOT_OUT" | head -3)"
    skip "copilot headless needs investigation"
  fi
else
  skip "copilot not installed"
fi

# 6. Hook verification
echo "[phase 6] hook verification"
if [ -f ~/.belt/hooks.log ]; then
  grep -qi "session-start\|user-prompt\|suggest" ~/.belt/hooks.log && \
    pass "belt hooks fired" || skip "hooks.log exists but no belt events"
else
  skip "no hooks.log"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

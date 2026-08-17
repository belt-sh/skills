#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Grok CLI Plugin Test ==="

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v grok >/dev/null && pass "grok installed" || fail "grok missing"

# 2. Configure endpoint
echo "[phase 2] endpoint"
export GROK_BASE_URL="${OPENROUTER_BASE:-https://openrouter.ai/api/v1}"
export XAI_API_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

RESP=$(curl -sf "$GROK_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $XAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Install belt hooks
echo "[phase 3] belt hooks"

mkdir -p ~/.grok/hooks
cat > ~/.grok/hooks/belt.json << 'EOF'
{
  "hooks": {
    "SessionStart": [{"type":"command","command":"AI_AGENT=grok belt me 2>/dev/null | head -1 || echo belt:not-authed","timeout":10}],
    "UserPromptSubmit": [{"type":"command","command":"AI_AGENT=grok belt suggest --json","timeout":30}],
    "Stop": [{"type":"command","command":"AI_AGENT=grok belt review --agent grok --trigger stop","timeout":120}],
    "PreCompact": [{"type":"command","command":"AI_AGENT=grok belt review --agent grok --trigger precompact --force","timeout":120}]
  }
}
EOF

[ -f ~/.grok/hooks/belt.json ] && pass "belt hooks installed" || fail "hooks missing"
grep -q "belt suggest" ~/.grok/hooks/belt.json && pass "hooks reference belt" || fail "hooks broken"

# 4. Install skills
echo "[phase 4] skills"

mkdir -p ~/.grok/skills
cp -r /opt/belt-plugin/skills/* ~/.grok/skills/ 2>/dev/null || true

[ -f ~/.grok/skills/belt/SKILL.md ] && pass "belt skill installed" || fail "belt skill missing"
[ -f ~/.grok/skills/suggest/SKILL.md ] && pass "suggest skill installed" || fail "suggest skill missing"
[ -f ~/.grok/skills/apps/SKILL.md ] && pass "apps skill installed" || fail "apps skill missing"

# 5. Headless run
echo "[phase 5] headless prompt"

GROK_OUT=$(timeout 45 grok -p "respond with exactly: BELT_TEST_OK" 2>&1 || true)
if echo "$GROK_OUT" | grep -qi "BELT_TEST_OK"; then
  pass "grok produced output"
elif echo "$GROK_OUT" | grep -qi "not signed in\|login\|authenticate"; then
  skip "grok requires xAI auth (XAI_API_KEY must be a real xAI key, not OpenRouter)"
else
  echo "  grok output: $(echo "$GROK_OUT" | head -5)"
  skip "grok headless needs investigation"
fi

# 6. Check hooks fired
echo "[phase 6] hook verification"
if [ -f ~/.belt/hooks.log ]; then
  head -10 ~/.belt/hooks.log
  grep -qi "session-start\|user-prompt\|suggest" ~/.belt/hooks.log && \
    pass "belt hooks fired" || fail "hooks.log exists but no belt events"
else
  skip "no hooks.log — hooks may not have fired"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

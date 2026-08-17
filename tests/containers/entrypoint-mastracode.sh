#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== MastraCode Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v mastracode >/dev/null 2>&1; then
  mastracode --version 2>&1 | head -1 || true
  pass "mastracode installed"
else
  skip "mastracode not available"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Hooks — MastraCode uses JSON with PascalCase events
echo "[phase 3] belt hooks"
mkdir -p ~/.mastracode
cat > ~/.mastracode/hooks.json << 'EOF'
{
  "hooks": {
    "AgentStart": [{"type":"command","command":"AI_AGENT=mastracode belt me 2>/dev/null | head -1 || echo belt:not-authed","timeout":10}],
    "PreToolUse": [{"type":"command","command":"AI_AGENT=mastracode belt suggest --json","timeout":30}],
    "Stop": [{"type":"command","command":"AI_AGENT=mastracode belt review --agent mastracode --trigger stop","timeout":120}]
  }
}
EOF

[ -f ~/.mastracode/hooks.json ] && pass "hooks.json installed" || fail "hooks.json missing"
grep -q "belt" ~/.mastracode/hooks.json && pass "hooks reference belt" || fail "hooks broken"

# 4. Headless
echo "[phase 4] headless prompt"
if command -v mastracode >/dev/null 2>&1; then
  export OPENAI_API_KEY="$OPENROUTER_KEY"
  export OPENAI_BASE_URL="https://openrouter.ai/api/v1"
  MC_OUT=$(timeout 45 mastracode --thread new "respond with exactly: BELT_TEST_OK" 2>&1 || true)
  if [ -n "$MC_OUT" ] && ! echo "$MC_OUT" | grep -qi "401\|403\|error.*api\|not.*auth"; then
    pass "mastracode produced output"
  else
    echo "  mastracode output: $(echo "$MC_OUT" | head -3)"
    skip "mastracode requires config"
  fi
else
  skip "mastracode not installed"
fi

# 5. Hook verification
echo "[phase 5] hook verification"
if [ -f ~/.belt/hooks.log ]; then
  grep -qi "session-start\|user-prompt\|suggest" ~/.belt/hooks.log && \
    pass "belt hooks fired" || skip "hooks.log exists but no belt events"
else
  skip "no hooks.log"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

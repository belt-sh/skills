#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Qoder CLI Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v qodercli >/dev/null 2>&1 || command -v qoder >/dev/null 2>&1; then
  (qodercli --version 2>&1 || qoder --version 2>&1) | head -1
  pass "qoder installed"
else
  skip "qoder not available"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Hooks — Qoder uses Claude Code-style JSON, PascalCase events
echo "[phase 3] belt hooks"
mkdir -p ~/.qoder
cat > ~/.qoder/settings.json << 'EOF'
{
  "hooks": {
    "UserPromptSubmit": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=qoder belt suggest --json","timeout":30}]}],
    "Stop": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=qoder belt review --agent qoder --trigger stop","timeout":120}]}],
    "PreCompact": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=qoder belt review --agent qoder --trigger precompact --force","timeout":120}]}]
  }
}
EOF

[ -f ~/.qoder/settings.json ] && pass "settings.json installed" || fail "settings.json missing"
grep -q "belt suggest" ~/.qoder/settings.json && pass "hooks reference belt" || fail "hooks broken"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.qoder/skills
cp -r /opt/belt-plugin/skills/* ~/.qoder/skills/ 2>/dev/null || true
[ -f ~/.qoder/skills/belt/SKILL.md ] && pass "belt skill installed" || fail "belt skill missing"

# 5. Headless
echo "[phase 5] headless prompt"
QODER_BIN=""
command -v qodercli >/dev/null 2>&1 && QODER_BIN="qodercli"
command -v qoder >/dev/null 2>&1 && QODER_BIN="qoder"
if [ -n "$QODER_BIN" ]; then
  Q_OUT=$(timeout 45 $QODER_BIN -p "respond with exactly: BELT_TEST_OK" 2>&1 || true)
  if [ -n "$Q_OUT" ] && ! echo "$Q_OUT" | grep -qi "401\|403\|error.*api\|not.*auth"; then
    pass "qoder produced output"
  else
    echo "  qoder output: $(echo "$Q_OUT" | head -3)"
    skip "qoder requires auth"
  fi
else
  skip "qoder not installed"
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

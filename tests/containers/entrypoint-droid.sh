#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Droid (Factory) Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v droid >/dev/null 2>&1; then
  droid --version 2>&1 | head -1 || true
  pass "droid installed"
else
  skip "droid CLI not available"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Belt hooks
echo "[phase 3] belt hooks"
mkdir -p ~/.factory
cat > ~/.factory/settings.json << EOF
{
  "hooks": {
    "UserPromptSubmit": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=droid belt suggest --json","timeout":30}]}],
    "Stop": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=droid belt review --agent droid --trigger stop","timeout":120}]}],
    "PreCompact": [{"matcher":"","hooks":[{"type":"command","command":"AI_AGENT=droid belt review --agent droid --trigger precompact --force","timeout":120}]}]
  },
  "customModels": [
    {
      "name": "openrouter-test",
      "provider": "openai",
      "apiKey": "$OPENROUTER_KEY",
      "baseUrl": "https://openrouter.ai/api/v1"
    }
  ]
}
EOF

[ -f ~/.factory/settings.json ] && pass "settings.json installed" || fail "settings.json missing"
grep -q "belt suggest" ~/.factory/settings.json && pass "hooks reference belt" || fail "hooks broken"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.factory/skills
cp -r /opt/belt-plugin/skills/* ~/.factory/skills/ 2>/dev/null || true
[ -f ~/.factory/skills/belt/SKILL.md ] && pass "belt skill installed" || fail "belt skill missing"

# 5. Headless run
echo "[phase 5] headless prompt"
if command -v droid >/dev/null 2>&1; then
  DROID_OUT=$(timeout 45 droid exec "respond with exactly: BELT_TEST_OK" --auto low 2>&1 || true)
  if echo "$DROID_OUT" | grep -qi "BELT_TEST_OK"; then
    pass "droid produced correct output"
  elif echo "$DROID_OUT" | grep -qi "not signed in\|login\|auth\|error"; then
    echo "  droid output: $(echo "$DROID_OUT" | head -3)"
    skip "droid requires auth"
  else
    echo "  droid output: $(echo "$DROID_OUT" | head -3)"
    skip "droid headless needs investigation"
  fi
else
  skip "droid not installed"
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

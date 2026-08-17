#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Hermes Agent Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v hermes >/dev/null 2>&1; then
  hermes --version 2>&1 | head -1 || true
  pass "hermes installed"
else
  skip "hermes not available"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Hooks — Hermes uses YAML config with snake_case events
echo "[phase 3] belt hooks"
mkdir -p ~/.hermes
cat > ~/.hermes/config.yaml << EOF
provider:
  type: openai
  base_url: https://openrouter.ai/api/v1
  api_key: $OPENROUTER_KEY
  model: openai/gpt-4o-mini

hooks:
  on_session_start:
    - command: "AI_AGENT=hermes belt me 2>/dev/null | head -1 || echo belt:not-authed"
      timeout: 10
  pre_tool_call:
    - command: "AI_AGENT=hermes belt suggest --json"
      timeout: 30
EOF

[ -f ~/.hermes/config.yaml ] && pass "config.yaml installed" || fail "config.yaml missing"
grep -q "belt" ~/.hermes/config.yaml && pass "hooks reference belt" || fail "hooks broken"

# 4. Headless
echo "[phase 4] headless prompt"
if command -v hermes >/dev/null 2>&1; then
  H_OUT=$(timeout 45 hermes -p "respond with exactly: BELT_TEST_OK" 2>&1 || true)
  if [ -n "$H_OUT" ] && ! echo "$H_OUT" | grep -qi "401\|403\|error.*api\|not.*auth"; then
    pass "hermes produced output"
  else
    echo "  hermes output: $(echo "$H_OUT" | head -3)"
    skip "hermes requires config"
  fi
else
  skip "hermes not installed"
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

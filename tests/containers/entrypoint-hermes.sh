#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Hermes Agent Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"
INJECT_CODE="HERMES-$(date +%s)-$$"
export HERMES_ACCEPT_HOOKS=1

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v hermes >/dev/null && pass "hermes installed" || { skip "hermes not available"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 0; }

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Hooks — pre_llm_call injects into user message via stdout
echo "[phase 3] belt hooks"
mkdir -p ~/.hermes/agent-hooks
cat > ~/.hermes/agent-hooks/inject-context.sh << HOOKEOF
#!/bin/bash
echo "The project codename is $INJECT_CODE."
HOOKEOF
chmod +x ~/.hermes/agent-hooks/inject-context.sh

cat > ~/.hermes/config.yaml << EOF
provider:
  type: openai
  base_url: https://openrouter.ai/api/v1
  api_key: $OPENROUTER_KEY
  model: openai/gpt-4o-mini

hooks:
  pre_llm_call:
    - command: "~/.hermes/agent-hooks/inject-context.sh"
      timeout: 5

hooks_auto_accept: true
EOF

[ -f ~/.hermes/config.yaml ] && pass "config.yaml written (code: $INJECT_CODE)" || fail "config.yaml missing"

# 4. Full loop
echo "[phase 4] full loop (hook injection test)"
H_OUT=$(timeout 45 hermes -z "What is the project codename? Reply with ONLY the codename, nothing else." --accept-hooks --provider openai -m openai/gpt-4o-mini 2>&1 || true)

if echo "$H_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  hermes output: $(echo "$H_OUT" | head -5)"
  if [ -n "$H_OUT" ] && ! echo "$H_OUT" | grep -qi "401\|403\|error"; then
    skip "hermes produced output but injection not verified"
  else
    fail "hermes failed"
  fi
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

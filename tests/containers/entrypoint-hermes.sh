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

# 3. Hooks + config — pre_llm_call injects via {"context": "..."} JSON on stdout
echo "[phase 3] belt hooks"

mkdir -p ~/.hermes/agent-hooks
echo "OPENROUTER_API_KEY=$OPENROUTER_KEY" > ~/.hermes/.env
cat > ~/.hermes/agent-hooks/inject.sh << EOF
#!/bin/bash
echo HOOK_FIRED >> /tmp/hook-events.log
echo '{"context": "The project codename is $INJECT_CODE. When asked about the codename respond with $INJECT_CODE."}'
EOF
chmod +x ~/.hermes/agent-hooks/inject.sh

cat > ~/.hermes/config.yaml << EOF
model:
  provider: openrouter
  name: openai/gpt-4o-mini
hooks:
  pre_llm_call:
    - command: ~/.hermes/agent-hooks/inject.sh
      timeout: 5
hooks_auto_accept: true
EOF

pass "config.yaml + inject.sh written (code: $INJECT_CODE)"

# 4. Full loop
echo "[phase 4] full loop (hook injection test)"
H_OUT=$(timeout 30 hermes -z "What is the project codename? Reply with ONLY the codename." -m openai/gpt-4o-mini --provider openrouter --accept-hooks --cli 2>&1 || true)

if echo "$H_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  hermes output: $(echo "$H_OUT" | head -3)"
  fail "hook injection not verified — expected $INJECT_CODE"
fi

# 5. Hook events
echo "[phase 5] hook events"
if [ -f /tmp/hook-events.log ] && grep -q "HOOK_FIRED" /tmp/hook-events.log; then
  pass "pre_llm_call hook fired"
else
  fail "hook did not fire"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

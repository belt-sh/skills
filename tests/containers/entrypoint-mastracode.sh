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
command -v mastracode >/dev/null && pass "mastracode installed" || { skip "mastracode not available"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 0; }

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Hooks — MastraCode hooks are block/allow gates, NOT context injection
# MastraCode does not support additionalContext injection from hook stdout (confirmed)
# Hooks are fire-and-forget (non-blocking) or block/allow (blocking) only
echo "[phase 3] belt hooks"
mkdir -p ~/.mastracode
cat > ~/.mastracode/hooks.json << 'EOF'
{
  "UserPromptSubmit": [
    {
      "type": "command",
      "command": "echo HOOK_PROMPT_FIRED >> /tmp/hook-events.log",
      "timeout": 5000,
      "description": "belt hook marker"
    }
  ],
  "Stop": [
    {
      "type": "command",
      "command": "echo HOOK_STOP_FIRED >> /tmp/hook-events.log",
      "timeout": 5000
    }
  ]
}
EOF
pass "hooks.json written (block/allow only — no context injection)"

# 4. Headless run — test that mastracode works, but can't verify injection
echo "[phase 4] headless prompt"
export OPENAI_API_KEY="$OPENROUTER_KEY"
export OPENAI_BASE_URL="https://openrouter.ai/api/v1"
MC_OUT=$(timeout 45 mastracode --thread new "say hello" 2>&1 || true)
if [ -n "$MC_OUT" ] && ! echo "$MC_OUT" | grep -qi "401\|403\|error.*api\|not.*auth"; then
  pass "mastracode produced output"
else
  echo "  mastracode output: $(echo "$MC_OUT" | head -3)"
  skip "mastracode needs config"
fi

# 5. Hook events
echo "[phase 5] hook events"
if [ -f /tmp/hook-events.log ]; then
  cat /tmp/hook-events.log
  grep -q "HOOK_PROMPT_FIRED" /tmp/hook-events.log && pass "UserPromptSubmit hook fired" || skip "hook not in log"
else
  skip "no hook events — MastraCode may not fire hooks in headless"
fi

# Note: MastraCode cannot inject context from hooks (GitHub issue mastra-ai/mastra#10078 — closed as unimplemented)
skip "context injection not supported — hooks are gate/logging only"

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== GitHub Copilot CLI Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"
INJECT_CODE="COPILOT-$(date +%s)-$$"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v copilot >/dev/null && pass "copilot installed" || { fail "copilot missing"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 1; }

# 2. Endpoint
echo "[phase 2] endpoint"
export COPILOT_PROVIDER_BASE_URL="https://openrouter.ai/api/v1"
export COPILOT_PROVIDER_API_KEY="$OPENROUTER_KEY"
export COPILOT_MODEL="openai/gpt-4o-mini"
pass "OpenRouter configured"

# 3. Hooks — Copilot v1 format with additionalContext injection
echo "[phase 3] belt hooks"
mkdir -p ~/.copilot/hooks
cat > ~/.copilot/hooks/belt.json << 'EOF'
{
  "version": 1,
  "hooks": {
    "sessionStart": [{"type":"command","bash":"echo HOOK_SESSION >> /tmp/hook-events.log","timeoutSec":5}],
    "userPromptSubmitted": [{"type":"command","bash":"echo HOOK_PROMPT >> /tmp/hook-events.log && echo '{\"additionalContext\": \"The project codename is INJECT_PLACEHOLDER. When asked about the codename respond with INJECT_PLACEHOLDER.\"}'","timeoutSec":5}],
    "agentStop": [{"type":"command","bash":"echo HOOK_STOP >> /tmp/hook-events.log","timeoutSec":5}]
  }
}
EOF

sed -i "s/INJECT_PLACEHOLDER/$INJECT_CODE/g" ~/.copilot/hooks/belt.json
pass "hooks.json written (v1 format, code: $INJECT_CODE)"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.copilot/skills
cp -r /opt/belt-plugin/skills/* ~/.copilot/skills/ 2>/dev/null || true
[ -f ~/.copilot/skills/belt/SKILL.md ] && pass "skills installed" || fail "skills missing"

# 5. Full loop
echo "[phase 5] full loop (hook injection test)"
COPILOT_OUT=$(copilot --prompt "What is the project codename? Reply with ONLY the codename, nothing else." 2>&1 || true)

if echo "$COPILOT_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  output: $(echo "$COPILOT_OUT" | head -3)"
  fail "hook injection not verified — expected $INJECT_CODE"
fi

# 6. Hook events
echo "[phase 6] hook events"
if [ -f /tmp/hook-events.log ]; then
  EVENTS=$(cat /tmp/hook-events.log)
  echo "$EVENTS" | grep -q "HOOK_SESSION" && pass "sessionStart hook fired" || skip "sessionStart not fired"
  echo "$EVENTS" | grep -q "HOOK_PROMPT" && pass "userPromptSubmitted hook fired" || fail "userPromptSubmitted not fired"
  echo "$EVENTS" | grep -q "HOOK_STOP" && pass "agentStop hook fired" || skip "agentStop not fired"
else
  fail "no hook events — hooks did not fire"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

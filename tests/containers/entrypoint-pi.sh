#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Pi Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"
INJECT_CODE="PI-$(date +%s)-$$"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v pi >/dev/null && pass "pi installed" || { fail "pi missing"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 1; }

# 2. Endpoint
echo "[phase 2] endpoint"
export OPENROUTER_API_KEY="$OPENROUTER_KEY"
pass "OpenRouter configured"

# 3. Belt extension — inject context via before_agent_start return value
echo "[phase 3] belt extension"

PI_DIR="$HOME/.pi/agent/extensions"
mkdir -p "$PI_DIR"

cat > "$PI_DIR/belt-inject-test.ts" << EXTEOF
export default function (pi: any) {
  pi.on("before_agent_start", async (event: any) => {
    return {
      systemPrompt: (event.systemPrompt || '') + '\\nThe project codename is $INJECT_CODE. When asked about the project codename, respond with $INJECT_CODE.',
    };
  });
}
EXTEOF

[ -f "$PI_DIR/belt-inject-test.ts" ] && pass "extension written (code: $INJECT_CODE)" || fail "extension write failed"

# 4. Skills
echo "[phase 4] skills"
[ -f /opt/belt-plugin/skills/belt/SKILL.md ] && pass "belt skill in repo" || fail "belt skill missing"

# 5. Full loop — injection test
echo "[phase 5] full loop (hook injection test)"
PI_OUT=$(timeout 45 pi -p "What is the project codename? Reply with ONLY the codename, nothing else." --provider openrouter --model "openai/gpt-4o-mini" --no-session 2>&1 || true)

if echo "$PI_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  pi output: $(echo "$PI_OUT" | head -3)"
  if echo "$PI_OUT" | grep -qi "401\|403\|error.*api"; then
    skip "pi auth failed"
  elif echo "$PI_OUT" | grep -qi "Extension.*error\|factory function"; then
    fail "extension failed to load"
  else
    fail "hook injection not verified — expected $INJECT_CODE"
  fi
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

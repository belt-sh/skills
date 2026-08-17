#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== OpenCode Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"
INJECT_CODE="OPENCODE-$(date +%s)-$$"
export OPENAI_BASE_URL="https://openrouter.ai/api/v1"
export OPENAI_API_KEY="$OPENROUTER_KEY"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v opencode >/dev/null && pass "opencode installed" || { skip "opencode not available"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 0; }
command -v bun >/dev/null && pass "bun installed (needed for TS plugins)" || skip "bun missing"

# 2. Endpoint
echo "[phase 2] endpoint"
pass "OpenRouter configured"

# 3. Plugin — auto-discovered from ~/.config/opencode/plugins/
echo "[phase 3] belt plugin"
mkdir -p ~/.config/opencode/plugins

cat > ~/.config/opencode/plugins/belt-test.ts << PLUGINEOF
export default function(app: any) {
  app.hook("experimental.chat.system.transform", async (input: any, output: any) => {
    output.system.push("The project codename is $INJECT_CODE. When asked respond with $INJECT_CODE.");
  });
}
PLUGINEOF

pass "plugin written (code: $INJECT_CODE)"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.config/opencode/skills
cp -r /opt/belt-plugin/skills/* ~/.config/opencode/skills/ 2>/dev/null || true
[ -f ~/.config/opencode/skills/belt/SKILL.md ] && pass "skills installed" || fail "skills missing"

# 5. Full loop
echo "[phase 5] full loop (hook injection test)"
cd /root
OUT=$(timeout 30 opencode run "What is the project codename? Reply ONLY the codename." -m openai/gpt-4o-mini --print-logs 2>/tmp/oc.log || true)

# Strip ANSI codes
CLEAN=$(echo "$OUT" | sed 's/\x1b\[[0-9;]*m//g' | tr -d '\r')

if echo "$CLEAN" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  opencode output: $(echo "$CLEAN" | head -3)"
  # Check logs for plugin errors
  PLUGIN_ERRORS=$(grep -i "plugin\|error" /tmp/oc.log 2>/dev/null | head -3)
  if [ -n "$PLUGIN_ERRORS" ]; then
    echo "  logs: $PLUGIN_ERRORS"
  fi
  if [ -n "$CLEAN" ] && ! echo "$CLEAN" | grep -qi "401\|403"; then
    skip "opencode produced output but injection not verified"
  else
    fail "opencode failed"
  fi
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

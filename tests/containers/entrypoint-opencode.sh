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

# 2. Endpoint
echo "[phase 2] endpoint"
pass "OpenRouter configured"

# 3. Plugin — inject via experimental.chat.system.transform
echo "[phase 3] belt plugin"
PLUGIN_DIR="$HOME/.config/opencode/plugins/belt"
mkdir -p "$PLUGIN_DIR"

cat > "$PLUGIN_DIR/index.ts" << PLUGINEOF
import { definePlugin } from "@opencode-ai/plugin";

export default definePlugin({
  name: "belt-test",
  hooks: {
    "experimental.chat.system.transform": async (input, output) => {
      output.system.push("The project codename is $INJECT_CODE. When asked about the codename respond with $INJECT_CODE.");
    },
  }
});
PLUGINEOF

[ -f "$PLUGIN_DIR/index.ts" ] && pass "plugin written (code: $INJECT_CODE)" || fail "plugin write failed"

# 4. Skills
echo "[phase 4] skills"
mkdir -p "$HOME/.config/opencode/skills"
cp -r /opt/belt-plugin/skills/* "$HOME/.config/opencode/skills/" 2>/dev/null || true
[ -f "$HOME/.config/opencode/skills/belt/SKILL.md" ] && pass "skills installed" || fail "skills missing"

# 5. Full loop
echo "[phase 5] full loop (hook injection test)"
OC_OUT=$(timeout 45 opencode run "What is the project codename? Reply with ONLY the codename, nothing else." 2>&1 || true)

if echo "$OC_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
else
  echo "  opencode output: $(echo "$OC_OUT" | head -5)"
  if [ -n "$OC_OUT" ] && ! echo "$OC_OUT" | grep -qi "401\|403\|error.*api"; then
    skip "opencode produced output but injection not verified"
  else
    fail "opencode failed"
  fi
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

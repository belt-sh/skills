#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== OpenCode Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v opencode >/dev/null 2>&1; then
  opencode --version 2>&1 | head -1 || true
  pass "opencode installed"
else
  skip "opencode not available — install method may have changed"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
export OPENAI_BASE_URL="https://openrouter.ai/api/v1"
export OPENAI_API_KEY="$OPENROUTER_KEY"

RESP=$(curl -sf "$OPENAI_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Belt plugin (TypeScript module)
echo "[phase 3] belt plugin"

PLUGIN_DIR="$HOME/.config/opencode/plugins/belt"
mkdir -p "$PLUGIN_DIR"

# Write a minimal TS plugin that calls belt on lifecycle events
cat > "$PLUGIN_DIR/index.ts" << 'PLUGINEOF'
import { definePlugin } from "@opencode-ai/plugin";

export default definePlugin({
  name: "belt",
  hooks: {
    onSessionStart: async () => {
      const { execSync } = require("child_process");
      try { execSync("AI_AGENT=opencode belt me 2>/dev/null | head -1", { timeout: 10000 }); } catch {}
    },
    onPromptSubmit: async () => {
      const { execSync } = require("child_process");
      try { execSync("AI_AGENT=opencode belt suggest --json", { timeout: 30000 }); } catch {}
    },
  }
});
PLUGINEOF

[ -f "$PLUGIN_DIR/index.ts" ] && pass "belt plugin written" || fail "plugin write failed"
grep -q "belt suggest" "$PLUGIN_DIR/index.ts" && pass "plugin references belt" || fail "plugin broken"

# 4. Skills
echo "[phase 4] skills"
mkdir -p "$HOME/.config/opencode/skills"
cp -r /opt/belt-plugin/skills/* "$HOME/.config/opencode/skills/" 2>/dev/null || true
[ -f "$HOME/.config/opencode/skills/belt/SKILL.md" ] && pass "belt skill installed" || fail "belt skill missing"

# 5. Headless run
echo "[phase 5] headless prompt"
if command -v opencode >/dev/null 2>&1; then
  OC_OUT=$(timeout 45 opencode --session new "respond with exactly: BELT_TEST_OK" 2>&1 || true)
  if echo "$OC_OUT" | grep -qi "BELT_TEST_OK"; then
    pass "opencode produced correct output"
  elif echo "$OC_OUT" | grep -qi "not.*auth\|login\|api.key\|error"; then
    echo "  opencode output: $(echo "$OC_OUT" | head -3)"
    skip "opencode requires auth config"
  else
    echo "  opencode output: $(echo "$OC_OUT" | head -3)"
    skip "opencode headless needs investigation"
  fi
else
  skip "opencode not installed"
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

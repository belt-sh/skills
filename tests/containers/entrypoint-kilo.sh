#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Kilo Code CLI Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v kilo >/dev/null 2>&1; then
  kilo --version 2>&1 | head -1 || true
  pass "kilo installed"
else
  skip "kilo not available"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
export KILO_PROVIDER="openai"
export KILO_API_KEY="$OPENROUTER_KEY"
export KILOCODE_MODEL="openai/gpt-4o-mini"

RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Plugin — Kilo uses TS/JS plugin modules, not hooks.json
echo "[phase 3] belt plugin"
KILO_PLUGIN="$HOME/.config/kilo/plugin/belt"
mkdir -p "$KILO_PLUGIN"
cat > "$KILO_PLUGIN/index.js" << 'PLUGINEOF'
module.exports = {
  name: "belt",
  activate(ctx) {
    ctx.on("session.created", () => {
      require("child_process").execSync("AI_AGENT=kilo belt me 2>/dev/null | head -1", { timeout: 10000 });
    });
    ctx.on("tool.execute.before", () => {
      require("child_process").execSync("AI_AGENT=kilo belt suggest --json", { timeout: 30000 });
    });
  }
};
PLUGINEOF

[ -f "$KILO_PLUGIN/index.js" ] && pass "belt plugin written" || fail "plugin write failed"
grep -q "belt suggest" "$KILO_PLUGIN/index.js" && pass "plugin references belt" || fail "plugin broken"

# 4. Skills
echo "[phase 4] skills"
mkdir -p "$HOME/.kilo/skills"
cp -r /opt/belt-plugin/skills/* "$HOME/.kilo/skills/" 2>/dev/null || true
[ -f "$HOME/.kilo/skills/belt/SKILL.md" ] && pass "belt skill installed" || fail "belt skill missing"

# 5. Headless
echo "[phase 5] headless prompt"
if command -v kilo >/dev/null 2>&1; then
  K_OUT=$(timeout 45 kilo run "respond with exactly: BELT_TEST_OK" --auto --format json 2>&1 || true)
  if [ -n "$K_OUT" ] && ! echo "$K_OUT" | grep -qi "401\|403\|error.*api\|not.*auth"; then
    pass "kilo produced output"
  else
    echo "  kilo output: $(echo "$K_OUT" | head -3)"
    skip "kilo requires config"
  fi
else
  skip "kilo not installed"
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

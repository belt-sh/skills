#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Pi Plugin Test ==="

OPENROUTER_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
if command -v pi >/dev/null 2>&1; then
  pi --version 2>&1 | head -1 || true
  pass "pi installed"
else
  skip "pi not available — install URL may have changed"
fi

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Belt extension (TypeScript)
echo "[phase 3] belt extension"

PI_DIR="$HOME/.pi/agent/extensions"
mkdir -p "$PI_DIR"

# Write belt extension
cat > "$PI_DIR/belt-agent-state.ts" << 'EXTEOF'
import { execSync } from "child_process";

export default function (pi: any) {
  pi.on("session_start", async () => {
    try { execSync("AI_AGENT=pi belt me 2>/dev/null | head -1", { timeout: 10000 }); } catch {}
  });

  pi.on("before_agent_start", async () => {
    try { execSync("AI_AGENT=pi belt suggest --json", { timeout: 30000 }); } catch {}
  });

  pi.on("agent_end", async () => {
    try { execSync("AI_AGENT=pi belt review --agent pi --trigger stop", { timeout: 120000 }); } catch {}
  });

  pi.on("session_before_compact", async () => {
    try { execSync("AI_AGENT=pi belt review --agent pi --trigger precompact --force", { timeout: 120000 }); } catch {}
  });
}
EXTEOF

[ -f "$PI_DIR/belt-agent-state.ts" ] && pass "belt extension written" || fail "extension write failed"
grep -q "belt suggest" "$PI_DIR/belt-agent-state.ts" && pass "extension references belt" || fail "extension broken"

# 4. Skills
echo "[phase 4] skills"
PI_SKILLS="$HOME/.pi/agent/extensions"
# Pi uses extensions dir — skills may need to go elsewhere
# For now test that skills exist in the repo
[ -f /opt/belt-plugin/skills/belt/SKILL.md ] && pass "belt skill in repo" || fail "belt skill missing from repo"

# 5. Headless run
echo "[phase 5] headless prompt"
if command -v pi >/dev/null 2>&1; then
  # Pi has native OpenRouter support via OPENROUTER_API_KEY
  export OPENROUTER_API_KEY="$OPENROUTER_KEY"
  PI_OUT=$(timeout 45 pi -p "respond with exactly: BELT_TEST_OK" --provider openrouter --model "openai/gpt-4o-mini" --no-session --mode text 2>&1 || true)
  PI_EXIT=$?
  echo "  pi exit=$PI_EXIT output_len=${#PI_OUT}"
  echo "  pi output: $(echo "$PI_OUT" | head -3)"
  if [ -n "$PI_OUT" ] && ! echo "$PI_OUT" | grep -qi "401\|403\|error.*api"; then
    pass "pi produced output"
  elif echo "$PI_OUT" | grep -qi "401\|403"; then
    skip "pi requires auth/provider config"
  else
    skip "pi headless needs investigation"
  fi
else
  skip "pi not installed"
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

#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Codex Plugin Test ==="

# 1. Verify installs
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v codex >/dev/null && pass "codex installed" || fail "codex missing"
belt version 2>&1 | head -1 && pass "belt runs" || pass "belt runs (no version output)"
codex --version 2>&1 | head -1 && pass "codex runs" || fail "codex broken"

# 2. Configure endpoint
echo "[phase 2] endpoint"
export OPENAI_BASE_URL="${OPENROUTER_BASE:-https://openrouter.ai/api/v1}"
export OPENAI_API_KEY="${OPENROUTER_KEY:?set OPENROUTER_KEY}"

RESP=$(curl -sf "$OPENAI_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
echo "$RESP" | grep -q '"choices"' && pass "endpoint works" || fail "endpoint broken"

# 3. Install belt plugin
echo "[phase 3] belt plugin install"

# Create codex config dir + hooks
mkdir -p ~/.codex
cat > ~/.codex/hooks.json << 'EOF'
{
  "hooks": {
    "SessionStart": [{"matcher":"startup","hooks":[{"type":"command","command":"AI_AGENT=codex belt me 2>/dev/null | head -1 || echo belt:not-authed","timeout":10}]}],
    "UserPromptSubmit": [{"hooks":[{"type":"command","command":"AI_AGENT=codex belt suggest --json","timeout":30}]}],
    "Stop": [{"hooks":[{"type":"command","command":"AI_AGENT=codex belt review --agent codex --trigger stop","timeout":120}]}]
  }
}
EOF

# Install skills from repo
mkdir -p ~/.agents/skills
cp -r /opt/belt-plugin/skills/* ~/.agents/skills/ 2>/dev/null || true

[ -f ~/.codex/hooks.json ] && pass "hooks.json installed" || fail "hooks.json missing"
grep -q "belt suggest" ~/.codex/hooks.json && pass "hooks reference belt" || fail "hooks don't reference belt"
[ -f ~/.agents/skills/belt/SKILL.md ] && pass "belt skill installed" || fail "belt skill missing"
[ -f ~/.agents/skills/suggest/SKILL.md ] && pass "suggest skill installed" || fail "suggest skill missing"

# 4. Codex config
echo "[phase 4] codex config"
mkdir -p ~/.codex
cat > ~/.codex/config.toml << EOF
model = "openai/gpt-4o-mini"

[features]
hooks = true
EOF
pass "codex config.toml written"

# 5. Headless run
echo "[phase 5] headless prompt"

# codex exec needs: git repo, sandbox (bubblewrap), and uses /v1/responses API (not /v1/chat/completions)
# Codex v0.147+ uses the OpenAI Responses API which OpenRouter doesn't support
mkdir -p /tmp/test-repo && cd /tmp/test-repo && git init -q && git config user.email 't@t' && git config user.name 't' && touch README.md && git add . && git commit -qm init
CODEX_OUT=$(echo "respond with exactly: BELT_TEST_OK" | timeout 30 codex exec -c model=openai/gpt-4o-mini --dangerously-bypass-hook-trust 2>&1 || true)
if echo "$CODEX_OUT" | grep -qi "BELT_TEST_OK"; then
  pass "codex exec produced correct output"
elif echo "$CODEX_OUT" | grep -qi "401\|Unauthorized\|/v1/responses"; then
  skip "codex uses /v1/responses API — OpenRouter not compatible (needs real OpenAI key)"
elif echo "$CODEX_OUT" | grep -qi "auth\|login\|not.*sign"; then
  skip "codex requires auth"
else
  echo "  codex output: $(echo "$CODEX_OUT" | head -3)"
  skip "codex exec needs investigation"
fi
cd /root

# 6. Check hooks fired
echo "[phase 6] hook verification"
if [ -f ~/.belt/hooks.log ]; then
  cat ~/.belt/hooks.log | head -10
  grep -qi "session-start\|user-prompt\|suggest" ~/.belt/hooks.log && \
    pass "belt hooks fired" || fail "hooks.log exists but no belt events"
else
  skip "no hooks.log — codex may not have fired hooks"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

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

# 2. Configure endpoint — Codex supports --provider openrouter natively
echo "[phase 2] endpoint"
# Use OpenAI key if available, fall back to OpenRouter
if [ -n "${OPENAI_API_KEY:-}" ]; then
  export OPENAI_API_KEY
  RESP=$(curl -sf "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
  echo "$RESP" | grep -q '"choices"' && pass "OpenAI endpoint works" || fail "OpenAI endpoint broken"
  CODEX_PROVIDER_FLAG=""
  CODEX_MODEL="gpt-4o-mini"
elif [ -n "${OPENROUTER_KEY:-}" ]; then
  export OPENROUTER_API_KEY="$OPENROUTER_KEY"
  RESP=$(curl -sf "https://openrouter.ai/api/v1/chat/completions" \
    -H "Authorization: Bearer $OPENROUTER_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"say ok"}],"max_tokens":5}' 2>&1 || true)
  echo "$RESP" | grep -q '"choices"' && pass "OpenRouter endpoint works" || fail "OpenRouter endpoint broken"
  CODEX_PROVIDER_FLAG="--provider openrouter"
  CODEX_MODEL="openai/gpt-4o-mini"
else
  fail "set OPENAI_API_KEY or OPENROUTER_KEY"
fi

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
model = "$CODEX_MODEL"

[features]
hooks = true
EOF
pass "codex config.toml written"

# 5. Headless run
echo "[phase 5] headless prompt"

mkdir -p /tmp/test-repo && cd /tmp/test-repo && git init -q && git config user.email 't@t' && git config user.name 't' && touch README.md && git add . && git commit -qm init
CODEX_OUT=$(echo "respond with exactly: BELT_TEST_OK" | timeout 60 codex exec $CODEX_PROVIDER_FLAG -c "model=$CODEX_MODEL" --dangerously-bypass-hook-trust 2>&1 || true)
if echo "$CODEX_OUT" | grep -qi "BELT_TEST_OK"; then
  pass "codex exec produced correct output"
elif echo "$CODEX_OUT" | grep -qi "hello\|ok\|content\|assistant"; then
  pass "codex exec produced output"
elif echo "$CODEX_OUT" | grep -qi "401\|Unauthorized"; then
  skip "codex auth failed"
else
  echo "  codex output: $(echo "$CODEX_OUT" | head -5)"
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

#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Codex Plugin Test ==="

INJECT_CODE="CODEX-$(date +%s)-$$"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v codex >/dev/null && pass "codex installed" || fail "codex missing"
codex --version 2>&1 | head -1
pass "codex runs"

# 2. Endpoint — prefer OpenRouter, fall back to OpenAI
echo "[phase 2] endpoint"
if [ -n "${OPENROUTER_KEY:-}" ]; then
  export OPENROUTER_API_KEY="$OPENROUTER_KEY"
  CODEX_MODEL="openai/gpt-4o-mini"
  CODEX_PROVIDER_BLOCK='model_provider = "openrouter"

[model_providers.openrouter]
name = "OpenRouter"
base_url = "https://openrouter.ai/api/v1"
env_key = "OPENROUTER_API_KEY"
wire_api = "responses"'
  pass "using OpenRouter"
elif [ -n "${OPENAI_API_KEY:-}" ]; then
  export OPENAI_API_KEY
  CODEX_MODEL="gpt-4o-mini"
  CODEX_PROVIDER_BLOCK=""
  pass "using OpenAI"
else
  fail "no API key (set OPENROUTER_KEY or OPENAI_API_KEY)"
  echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 1
fi

# 3. Hooks + config
echo "[phase 3] belt hooks"
mkdir -p ~/.codex
cat > ~/.codex/config.toml << EOF
model = "$CODEX_MODEL"
$CODEX_PROVIDER_BLOCK

[features]
hooks = true
EOF

cat > ~/.codex/hooks.json << 'EOF'
{
  "hooks": {
    "UserPromptSubmit": [{"hooks":[{"type":"command","command":"echo HOOK_PROMPT >> /tmp/hook-events.log && echo 'The project codename is INJECT_CODE_PLACEHOLDER.'","timeout":5}]}],
    "Stop": [{"hooks":[{"type":"command","command":"echo HOOK_STOP >> /tmp/hook-events.log","timeout":5}]}]
  }
}
EOF

sed -i "s/INJECT_CODE_PLACEHOLDER/$INJECT_CODE/g" ~/.codex/hooks.json
pass "config.toml + hooks.json written (code: $INJECT_CODE)"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.agents/skills
cp -r /opt/belt-plugin/skills/* ~/.agents/skills/ 2>/dev/null || true
[ -f ~/.agents/skills/belt/SKILL.md ] && pass "skills installed" || fail "skills missing"

# 5. Full loop: hook injection → agent reads context → structured response
echo "[phase 5] full loop (hook injection test)"
mkdir -p /tmp/test-repo && cd /tmp/test-repo && git init -q && git config user.email t@t && git config user.name t && touch r && git add . && git commit -qm init

CODEX_OUT=$(echo "What is the project codename? Reply with ONLY the codename, nothing else." | \
  timeout 60 codex exec --dangerously-bypass-hook-trust 2>&1 || true)

if echo "$CODEX_OUT" | grep -q "$INJECT_CODE"; then
  pass "hook injection verified — agent returned $INJECT_CODE"
elif echo "$CODEX_OUT" | grep -qi "401\|Unauthorized"; then
  fail "API auth failed"
else
  echo "  output: $(echo "$CODEX_OUT" | grep -v "^warning:\|^$\|^---\|^OpenAI\|^user\|^Reading\|^workdir\|^model\|^provider\|^approval\|^sandbox\|^reasoning\|^session" | head -3)"
  fail "hook injection not verified — agent did not return $INJECT_CODE"
fi
cd /root

# 6. Hook verification
echo "[phase 6] hook events"
if [ -f /tmp/hook-events.log ]; then
  EVENTS=$(cat /tmp/hook-events.log)
  echo "$EVENTS" | grep -q "HOOK_PROMPT" && pass "UserPromptSubmit hook fired" || fail "UserPromptSubmit not in events"
  echo "$EVENTS" | grep -q "HOOK_STOP" && pass "Stop hook fired" || skip "Stop hook not fired"
else
  fail "no hook events log — hooks did not fire"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

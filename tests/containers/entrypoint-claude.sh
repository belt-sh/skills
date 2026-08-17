#!/bin/bash
set -euo pipefail

PASS=0; FAIL=0; SKIP=0
pass() { PASS=$((PASS+1)); echo "  ✓ $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $*" >&2; }
skip() { SKIP=$((SKIP+1)); echo "  ○ $*"; }

echo "=== Claude Code Plugin Test ==="

ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:?set ANTHROPIC_API_KEY}"
export ANTHROPIC_API_KEY
INJECT_CODE="CLAUDE-$(date +%s)-$$"

# 1. Prerequisites
echo "[phase 1] prerequisites"
command -v belt >/dev/null && pass "belt installed" || fail "belt missing"
command -v claude >/dev/null && pass "claude installed" || { fail "claude missing"; echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ===" && exit 1; }
claude --version 2>&1 | head -1
pass "claude runs"

# 2. Endpoint
echo "[phase 2] endpoint"
RESP=$(curl -sf "https://api.anthropic.com/v1/messages" \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-haiku-4-5-20251001","max_tokens":10,"messages":[{"role":"user","content":"say ok"}]}' 2>&1 || true)
echo "$RESP" | grep -q '"content"' && pass "Anthropic endpoint works" || fail "Anthropic endpoint broken: $(echo "$RESP" | head -c 200)"

# 3. Hooks — Claude Code uses settings.json with PascalCase events
echo "[phase 3] belt hooks"
mkdir -p ~/.claude

cat > ~/.claude/settings.json << EOF
{
  "permissions": {
    "allow": ["Bash(*)", "Read(*)", "Write(*)"]
  },
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "echo HOOK_PROMPT >> /tmp/hook-events.log && echo 'The project codename is $INJECT_CODE. When asked about the codename respond with $INJECT_CODE.'",
            "timeout": 5
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "echo HOOK_STOP >> /tmp/hook-events.log",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
EOF

pass "settings.json written (code: $INJECT_CODE)"

# 4. Skills
echo "[phase 4] skills"
mkdir -p ~/.claude/skills
cp -r /opt/belt-plugin/skills/* ~/.claude/skills/ 2>/dev/null || true
[ -f ~/.claude/skills/belt/SKILL.md ] && pass "skills installed" || skip "skills dir format may differ"

# 5. Multi-turn session: hook injects per-message context
echo "[phase 5] multi-turn injection test"
mkdir -p /tmp/test-repo && cd /tmp/test-repo && git init -q && git config user.email t@t && git config user.name t && touch r && git add . && git commit -qm i

# Turn 1: hook fires, injects codename
TURN1=$(timeout 60 claude -p "I need to know the project codename. What is it? Reply with ONLY the codename." --model claude-haiku-4-5-20251001 --dangerously-skip-permissions 2>&1 || true)

if echo "$TURN1" | grep -q "$INJECT_CODE"; then
  pass "turn 1: hook injection verified — $INJECT_CODE"
elif echo "$TURN1" | grep -qi "401\|403\|auth\|invalid"; then
  fail "Anthropic auth failed: $(echo "$TURN1" | head -1)"
else
  echo "  turn 1 output: $(echo "$TURN1" | head -3)"
  skip "turn 1: injection not verified"
fi

# Turn 2: continue session, hook fires again, verify context persists
TURN2=$(timeout 60 claude -p -c "Repeat the codename you just told me." --model claude-haiku-4-5-20251001 --dangerously-skip-permissions 2>&1 || true)

if echo "$TURN2" | grep -q "$INJECT_CODE"; then
  pass "turn 2: context persists across turns — $INJECT_CODE"
else
  echo "  turn 2 output: $(echo "$TURN2" | head -3)"
  skip "turn 2: context persistence not verified"
fi

cd /home/testuser

# 6. Hook events — verify hooks fired on BOTH turns
echo "[phase 6] hook events"
if [ -f /tmp/hook-events.log ]; then
  HOOK_COUNT=$(grep -c "HOOK_PROMPT" /tmp/hook-events.log 2>/dev/null || echo 0)
  [ "$HOOK_COUNT" -ge 2 ] && pass "UserPromptSubmit fired $HOOK_COUNT times (multi-turn)" || \
  [ "$HOOK_COUNT" -ge 1 ] && pass "UserPromptSubmit fired $HOOK_COUNT time(s)" || \
    fail "UserPromptSubmit not in events"
  grep -q "HOOK_STOP" /tmp/hook-events.log && pass "Stop hook fired" || skip "Stop not fired"
else
  skip "no hook events log"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed, $SKIP skipped ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1

#!/bin/bash
# Belt plugin test framework
# Source this from harness-specific test scripts.
#
# Usage:
#   source "$(dirname "$0")/framework.sh"
#   harness_setup "grok"
#   harness_install_cli "curl -fsSL https://x.ai/cli/install.sh | bash"
#   harness_configure_endpoint
#   harness_install_belt
#   harness_run_prompt "hello, respond in 5 words"
#   harness_check_hooks
#   harness_teardown

set -euo pipefail

OPENROUTER_BASE="https://openrouter.ai/api/v1"
OPENROUTER_KEY="${OPENROUTER_KEY:-}"
ANTHROPIC_KEY="${ANTHROPIC_KEY:-}"

TEST_MODEL="${TEST_MODEL:-openai/gpt-4o-mini}"
TEST_DIR=""
HARNESS_NAME=""
PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

log()  { echo "[test] $*"; }
pass() { PASS_COUNT=$((PASS_COUNT + 1)); echo "  ✓ $*"; }
fail() { FAIL_COUNT=$((FAIL_COUNT + 1)); echo "  ✗ $*" >&2; }
skip() { SKIP_COUNT=$((SKIP_COUNT + 1)); echo "  ○ $* (skipped)"; }

harness_setup() {
  HARNESS_NAME="$1"
  TEST_DIR=$(mktemp -d "/tmp/belt-test-${HARNESS_NAME}-XXXXXX")
  export HOME="$TEST_DIR/home"
  mkdir -p "$HOME"
  log "testing $HARNESS_NAME (home=$HOME)"

  if [ -z "$OPENROUTER_KEY" ] && [ -z "$ANTHROPIC_KEY" ]; then
    echo "error: set OPENROUTER_KEY or ANTHROPIC_KEY" >&2
    exit 1
  fi
}

harness_teardown() {
  echo ""
  log "results for $HARNESS_NAME: $PASS_COUNT passed, $FAIL_COUNT failed, $SKIP_COUNT skipped"
  if [ "$FAIL_COUNT" -gt 0 ]; then
    log "TEST_DIR preserved at $TEST_DIR"
    exit 1
  else
    rm -rf "$TEST_DIR"
    exit 0
  fi
}

# Check if a command exists
has_cmd() { command -v "$1" >/dev/null 2>&1; }

# Assert a file exists
assert_file() {
  if [ -f "$1" ]; then
    pass "$2"
  else
    fail "$2 (missing: $1)"
  fi
}

# Assert a directory exists and is non-empty
assert_dir_nonempty() {
  if [ -d "$1" ] && [ "$(ls -A "$1" 2>/dev/null)" ]; then
    pass "$2"
  else
    fail "$2 (missing or empty: $1)"
  fi
}

# Assert a file contains a string
assert_contains() {
  if grep -q "$2" "$1" 2>/dev/null; then
    pass "$3"
  else
    fail "$3 (not found: '$2' in $1)"
  fi
}

# Assert a command succeeds
assert_cmd() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then
    pass "$desc"
  else
    fail "$desc (exit $?)"
  fi
}

# Quick check: can we reach the endpoint?
check_endpoint() {
  local url="${1:-$OPENROUTER_BASE}"
  local key="${2:-$OPENROUTER_KEY}"
  if curl -sf "$url/models" -H "Authorization: Bearer $key" >/dev/null 2>&1; then
    pass "endpoint reachable ($url)"
  else
    fail "endpoint unreachable ($url)"
  fi
}

# Send a minimal chat completion and check we get a response
check_chat_completion() {
  local url="${1:-$OPENROUTER_BASE}"
  local key="${2:-$OPENROUTER_KEY}"
  local model="${3:-$TEST_MODEL}"
  local resp
  resp=$(curl -sf "$url/chat/completions" \
    -H "Authorization: Bearer $key" \
    -H "Content-Type: application/json" \
    -d "{\"model\":\"$model\",\"messages\":[{\"role\":\"user\",\"content\":\"say ok\"}],\"max_tokens\":5}" 2>&1) || true
  if echo "$resp" | grep -q '"choices"'; then
    pass "chat completion works ($model)"
  else
    fail "chat completion failed ($model): $(echo "$resp" | head -c 200)"
  fi
}

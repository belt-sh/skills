#!/bin/bash
# Run harness tests in Docker containers.
#
# Usage:
#   OPENROUTER_KEY=sk-or-... ./tests/run-containers.sh
#   OPENROUTER_KEY=sk-or-... ./tests/run-containers.sh test-grok    # single harness
#   OPENROUTER_KEY=sk-or-... ./tests/run-containers.sh test-codex test-kimi  # specific set

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Load .env if present
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a; source "$SCRIPT_DIR/.env"; set +a
fi

if [ -z "${OPENROUTER_KEY:-}" ]; then
  echo "error: set OPENROUTER_KEY or create tests/.env" >&2
  exit 1
fi

export OPENROUTER_KEY
export XAI_API_KEY="${XAI_API_KEY:-}"
export OPENAI_API_KEY="${OPENAI_API_KEY:-}"
export ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

if [ $# -gt 0 ]; then
  SERVICES=("$@")
else
  SERVICES=(test-codex test-grok test-copilot test-opencode test-pi test-hermes test-mastracode)
fi

echo "=== Building and running: ${SERVICES[*]} ==="
echo ""

FAILED=()
for svc in "${SERVICES[@]}"; do
  echo "--- $svc ---"

  # Build
  if ! docker compose -f tests/docker-compose.test.yml build "$svc" 2>&1; then
    echo "[$svc] BUILD FAILED"
    FAILED+=("$svc")
    continue
  fi

  # Run
  if docker compose -f tests/docker-compose.test.yml run --rm "$svc" 2>&1; then
    echo "[$svc] PASSED"
  else
    echo "[$svc] FAILED"
    FAILED+=("$svc")
  fi
  echo ""
done

echo "=== Summary ==="
echo "ran: ${#SERVICES[@]}"
echo "failed: ${#FAILED[@]}"
if [ ${#FAILED[@]} -gt 0 ]; then
  echo "failures: ${FAILED[*]}"
  exit 1
fi
exit 0

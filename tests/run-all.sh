#!/bin/bash
# Run all harness tests. Skips harnesses whose CLI is not installed.
#
# Usage:
#   OPENROUTER_KEY=sk-or-... ./tests/run-all.sh
#   OPENROUTER_KEY=sk-or-... ./tests/run-all.sh grok copilot  # specific harnesses only

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TOTAL_PASS=0
TOTAL_FAIL=0
TOTAL_SKIP=0

if [ -z "${OPENROUTER_KEY:-}" ]; then
  echo "error: set OPENROUTER_KEY" >&2
  exit 1
fi

export OPENROUTER_KEY

# If args given, run only those; otherwise run all
if [ $# -gt 0 ]; then
  HARNESSES=("$@")
else
  HARNESSES=()
  for f in "$SCRIPT_DIR"/harnesses/test-*.sh; do
    name=$(basename "$f" | sed 's/test-//;s/\.sh//')
    HARNESSES+=("$name")
  done
fi

echo "=== Belt Plugin Test Suite ==="
echo "harnesses: ${HARNESSES[*]}"
echo ""

for harness in "${HARNESSES[@]}"; do
  test_file="$SCRIPT_DIR/harnesses/test-${harness}.sh"
  if [ ! -f "$test_file" ]; then
    echo "[$harness] no test file found at $test_file"
    continue
  fi

  echo "--- $harness ---"
  chmod +x "$test_file"
  if bash "$test_file" 2>&1; then
    echo ""
  else
    echo "[$harness] FAILED"
    echo ""
  fi
done

echo "=== Done ==="

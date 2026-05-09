#!/bin/bash
# Belt bootstrap — idempotent, runs in <1s when everything is set up.
# Silent fail if belt is not installed (exit 0, no output).
#
# Flags:
#   (none)                 Check belt installed + auth status
#   --check-update         Check for skill updates
#   --check-skill-mutation Read tool input from stdin, detect belt-managed skill edits

set -euo pipefail

# If belt is not installed, exit silently
command -v belt >/dev/null 2>&1 || exit 0

case "${1:-}" in
  --check-update)
    # Check for updates to installed skills
    belt skill list --json 2>/dev/null | while read -r skill; do
      name=$(echo "$skill" | belt jq -r '.name // empty' 2>/dev/null) || continue
      belt skill check "$name" --quiet 2>/dev/null || true
    done
    ;;

  --check-skill-mutation)
    # Read tool input from stdin, check if the edited file is in a belt-managed dir
    input=$(cat)
    file_path=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | cut -d'"' -f4 2>/dev/null) || exit 0

    # Check if path is under a belt-managed skills directory
    belt_managed="$HOME/.belt/managed"
    if [ -f "$belt_managed" ] && grep -qF "$(dirname "$file_path")" "$belt_managed" 2>/dev/null; then
      # Record mutation
      mkdir -p "$HOME/.belt"
      echo "{\"file\":\"$file_path\",\"ts\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" >> "$HOME/.belt/mutations.jsonl"
    fi
    ;;

  *)
    # Default: check auth status
    if belt me >/dev/null 2>&1; then
      echo "belt: authenticated"
    else
      echo "belt: not authenticated. Run: belt login"
    fi
    ;;
esac

exit 0

#!/usr/bin/env bash
#
# Verifies that plugins/belt/ still mirrors the top-level source trees.
#
# This repo carries two copies of everything: the top level is what the registry
# publishes (`belt skill use`), and plugins/belt/ is what the Claude Code plugin
# reads (the /slash commands). They are kept in sync by hand with `cp`, so
# editing one side silently leaves the other serving stale content — and the
# symptom is invisible unless you happen to exercise both entry points.
#
# Run locally before committing:
#   ./.github/scripts/check-mirror.sh
#
# To resync:
#   cp -r <tree>/ plugins/belt/<tree>/

set -euo pipefail

cd "$(dirname "$0")/../.."

# Trees duplicated into plugins/belt/.
TREES=(skills agents hooks bin rules)

# Paths deliberately absent from the plugin. Listing them here keeps the
# exclusion visible — an undocumented gap is indistinguishable from a mistake.
#   expert-panel — never shipped as a slash command
EXCLUDE=(expert-panel)

diff_args=()
for name in "${EXCLUDE[@]}"; do
  diff_args+=(--exclude="$name")
done

status=0
for tree in "${TREES[@]}"; do
  if [[ ! -d $tree ]]; then
    continue
  fi
  if [[ ! -d "plugins/belt/$tree" ]]; then
    echo "::error::plugins/belt/$tree is missing entirely"
    status=1
    continue
  fi

  if ! diff -rq "${diff_args[@]}" "$tree" "plugins/belt/$tree"; then
    echo "::error::$tree/ and plugins/belt/$tree/ have drifted"
    status=1
  fi
done

if [[ $status -ne 0 ]]; then
  cat >&2 <<'EOF'

The plugin mirror is out of sync.

The Claude Code plugin reads plugins/belt/, the registry reads the top level.
Whichever side you edited, the other is now serving different content.

Fix by copying the source tree over the mirror, then commit both:

  cp <changed-file> plugins/belt/<same-path>
  ./.github/scripts/check-mirror.sh

If a path is meant to exist on only one side, add it to EXCLUDE in this script
so the intent is recorded rather than rediscovered.
EOF
  exit 1
fi

echo "plugins/belt/ mirrors every source tree."

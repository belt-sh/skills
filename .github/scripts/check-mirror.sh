#!/usr/bin/env bash
#
# Verifies that plugins/belt/ still mirrors the top-level source trees.
#
# This repo carries two copies of everything. The top level is what the registry
# publishes (`belt skill use`) and what the Claude Code plugin reads
# (.claude-plugin/marketplace.json has source "./"). plugins/belt/ is what the
# Codex plugin reads (.agents/plugins/marketplace.json points there). They are
# kept in sync by hand with `cp`, so editing one side silently leaves the other
# serving stale content — and the symptom is invisible unless you happen to
# exercise both entry points.
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
#   expert-panel     — never shipped as a slash command
#   codex-hooks.json — the hooks file plugins/belt/.codex-plugin/plugin.json
#                      loads. It is tuned for Codex (exec_command matcher,
#                      bin/ shims) and has no top-level counterpart.
EXCLUDE=(expert-panel codex-hooks.json)

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

The registry and the Claude Code plugin read the top level; the Codex plugin
reads plugins/belt/. Whichever side you edited, the other is now serving
different content.

Fix by copying the source tree over the mirror, then commit both:

  cp <changed-file> plugins/belt/<same-path>
  ./.github/scripts/check-mirror.sh

If a path is meant to exist on only one side, add it to EXCLUDE in this script
so the intent is recorded rather than rediscovered.
EOF
  exit 1
fi

echo "plugins/belt/ mirrors every source tree."

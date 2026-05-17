#!/bin/bash
# Runs before compaction — full conversation is about to be summarized.
# This is the best moment for fork-based extraction: cache is warm, detail is about to be lost.

source "$(dirname "$0")/utils.sh"
read_input
log_hook "PreCompact"

# Fork-based review — full conversation context at near-zero cache cost
echo "$INPUT" | belt claude review --force --trigger precompact 2>>"$BELT_LOG"

exit 0

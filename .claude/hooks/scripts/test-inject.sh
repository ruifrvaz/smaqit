#!/usr/bin/env bash
# smaqit hook test — confirms PostToolUse fires in Claude Code.
# Claude Code equivalent of .github/hooks/scripts/test-inject.sh (the Copilot diagnostic).
GIT_STATUS=$(git -C "$(dirname "$0")/../../.." status 2>&1)
printf '{"hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": "[smaqit hook test] git status:\n%s"}}' "$GIT_STATUS"

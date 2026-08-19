#!/usr/bin/env bash
# Static regression check: every hidden (-s/-rs) `read` in this skill's scripts
# must redirect from /dev/tty explicitly (the read_secret pattern), so a
# non-interactive invocation fails loudly instead of silently reading an empty
# or garbage value that later reaches `vault kv put` as a placeholder secret.
#
# Catches this regression class even for a future ad hoc read someone adds
# later — not just the specific sites fixed by task 110.
#
# Usage: bash check-no-ad-hoc-secret-reads.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FAILED=false

for f in "${SCRIPT_DIR}"/*.sh; do
  [ "$(basename "$f")" = "$(basename "${BASH_SOURCE[0]}")" ] && continue
  while IFS=: read -r lineno line; do
    case "$line" in
      *'/dev/tty'*) ;;
      *)
        echo "FAIL: $(basename "$f"):${lineno}: hidden read with no /dev/tty redirect — use read_secret instead"
        echo "      ${line}"
        FAILED=true
        ;;
    esac
  done < <(grep -n -E 'read[[:space:]].*-r?s([[:space:]]|$)' "$f" || true)
done

if [ "$FAILED" = "true" ]; then
  echo ""
  echo "One or more hidden reads can silently succeed with an empty value in a non-interactive shell."
  exit 1
fi

echo "OK — every hidden read in this skill's scripts redirects from /dev/tty."

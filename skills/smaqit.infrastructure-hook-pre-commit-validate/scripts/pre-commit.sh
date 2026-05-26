#!/usr/bin/env bash
# scripts/pre-commit.sh
# Runs checks 1-3 on staged files. No agent context required.
# Install via: bash .github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/install.sh

EXIT_CODE=0

# Check 1: No .env files staged
ENV_FILES=$(git diff --cached --name-only | grep -E '\.env$|\.env\.' || true)
if [ -n "$ENV_FILES" ]; then
  echo "FAIL: .env file(s) staged:"
  echo "$ENV_FILES" | sed 's/^/  /'
  EXIT_CODE=1
fi

# Check 2: No plaintext secrets in staged diff
SECRET_MATCHES=$(git diff --cached | grep -nE '(sk-ant-|ANTHROPIC_API_KEY=sk-|BEGIN.*PRIVATE KEY|(password|secret|token)\s*=\s*[^\$\{]{8,})' || true)
if [ -n "$SECRET_MATCHES" ]; then
  echo "FAIL: Plaintext secret pattern found:"
  echo "$SECRET_MATCHES" | sed 's/^/  /'
  EXIT_CODE=1
fi

# Check 3: Draft spec on main/merge (warning only)
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
IS_MERGE=0
git rev-parse -q --verify MERGE_HEAD >/dev/null 2>&1 && IS_MERGE=1 || true

if [ "$BRANCH" = "main" ] || [ "$IS_MERGE" -eq 1 ]; then
  STAGED_SPECS=$(git diff --cached --name-only | grep -E '^specs/.*\.md$' || true)
  if [ -n "$STAGED_SPECS" ]; then
    DRAFT_SPECS=$(echo "$STAGED_SPECS" | xargs grep -l 'status: draft' 2>/dev/null || true)
    if [ -n "$DRAFT_SPECS" ]; then
      echo "WARN: Draft spec(s) staged on main/merge commit:"
      echo "$DRAFT_SPECS" | sed 's/^/  /'
    fi
  fi
fi

exit $EXIT_CODE

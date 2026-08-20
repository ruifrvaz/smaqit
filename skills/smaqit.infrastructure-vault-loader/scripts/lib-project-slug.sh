#!/usr/bin/env bash
# Shared project-slug derivation for the vault-loader script family.
# Sourced by load-credentials.sh and rotate-credential.sh — do not execute directly.
#
# Prefers the repository's real identity over AGENTS.md's human-readable "Project
# Name" field (a display title with no guaranteed relationship to the technical
# slug already used throughout this framework's Vault layout):
#   1. git remote get-url origin — basename, ".git" suffix stripped
#   2. current working directory's basename
#   3. empty — caller prompts manually
#
# Usage: PROJECT_SLUG=$(derive_project_slug)

derive_project_slug() {
  local slug="" remote

  remote=$(git remote get-url origin 2>/dev/null || true)
  if [ -n "$remote" ]; then
    slug=$(basename "$remote" .git)
  fi

  if [ -z "$slug" ]; then
    slug=$(basename "$(pwd)")
  fi

  printf '%s' "$slug" \
    | tr '[:upper:]' '[:lower:]' \
    | sed 's/[^a-z0-9-]/-/g' \
    | tr -s '-' \
    | sed 's/^-//; s/-$//'
}

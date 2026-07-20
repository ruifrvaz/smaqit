#!/usr/bin/env bash
# Deterministic Vault -> GitHub secrets/variables sync for smaqit.infrastructure-repo-config.
#
# Replaces an inline bash-in-markdown conditional (tfstate/cyso skip-if-absent) with a real
# script, so the mode-awareness is structural and can't be silently "simplified" back to an
# unconditional read by a future edit of the skill's prose.
#
# Usage: sync-secrets.sh <owner>/<repo> <vm-host-value>
#   Requires PROJECT_SLUG and VAULT_ADDR already exported (same convention as every other
#   smaqit.infrastructure-* script — see smaqit.infrastructure-vault-loader).
#
#   Exit 0 — all applicable secrets/variables synced (tfstate/cyso skipped cleanly and reported,
#            not treated as errors, when their Vault paths are absent).
#   Exit 1 — a required Vault path (ssh, github) is missing, or a `gh` call failed.

set -euo pipefail

REPO="${1:?Usage: sync-secrets.sh <owner>/<repo> <vm-host-value>}"
VM_HOST_VALUE="${2:?Usage: sync-secrets.sh <owner>/<repo> <vm-host-value>}"
: "${PROJECT_SLUG:?PROJECT_SLUG must be exported}"
: "${VAULT_ADDR:?VAULT_ADDR must be exported}"

path_exists() {
  vault kv get "secret/${PROJECT_SLUG}/$1" > /dev/null 2>&1
}

SKIPPED=()

echo "==> Syncing secrets/variables for ${PROJECT_SLUG} -> ${REPO}"

# ── SSH deploy key (always required) ───────────────────────────────────────────

if ! path_exists ssh; then
  echo "sync-secrets: ERROR — secret/${PROJECT_SLUG}/ssh is required and absent" >&2
  exit 1
fi

# tr -d '\n' strips trailing newline — without it, Terraform marks the keypair for
# replacement on every plan.
vault kv get -field=private_key "secret/${PROJECT_SLUG}/ssh" \
  | tr -d '\n' \
  | gh secret set VM_SSH_KEY -R "$REPO" --stdin
echo "    VM_SSH_KEY — set"

vault kv get -field=public_key "secret/${PROJECT_SLUG}/ssh" \
  | gh secret set VM_SSH_PUBLIC_KEY -R "$REPO" --stdin
echo "    VM_SSH_PUBLIC_KEY — set"

# ── VM host (always required — a variable, not a secret; see SKILL.md) ────────

gh variable set VM_HOST --body "$VM_HOST_VALUE" -R "$REPO"
echo "    VM_HOST (variable) — set"

# ── Terraform backend credentials (skip cleanly if absent) ────────────────────

if path_exists tfstate; then
  vault kv get -field=access_key "secret/${PROJECT_SLUG}/tfstate" \
    | gh secret set TF_BACKEND_ACCESS_KEY -R "$REPO" --stdin
  vault kv get -field=secret_key "secret/${PROJECT_SLUG}/tfstate" \
    | gh secret set TF_BACKEND_SECRET_KEY -R "$REPO" --stdin
  echo "    TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY — set"
else
  echo "    SKIP — secret/${PROJECT_SLUG}/tfstate absent (expected for provisioning_mode: existing-shared)"
  SKIPPED+=("TF_BACKEND_ACCESS_KEY" "TF_BACKEND_SECRET_KEY")
fi

# ── Cloud provider credentials (skip cleanly if absent) ───────────────────────

if path_exists cyso; then
  vault kv get -field=app_credential_id "secret/${PROJECT_SLUG}/cyso" \
    | gh secret set OS_APPLICATION_CREDENTIAL_ID -R "$REPO" --stdin
  vault kv get -field=app_credential_secret "secret/${PROJECT_SLUG}/cyso" \
    | gh secret set OS_APPLICATION_CREDENTIAL_SECRET -R "$REPO" --stdin
  echo "    OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET — set"
else
  echo "    SKIP — secret/${PROJECT_SLUG}/cyso absent (expected for provisioning_mode: existing-shared)"
  SKIPPED+=("OS_APPLICATION_CREDENTIAL_ID" "OS_APPLICATION_CREDENTIAL_SECRET")
fi

# ── GitHub Terraform token (always required) ───────────────────────────────────
# CRITICAL: the workflow YAML env var for this secret MUST be TF_VAR_github_token,
# NOT GITHUB_TOKEN — that name is reserved and the runner overwrites it before any
# step executes.

if ! path_exists github; then
  echo "sync-secrets: ERROR — secret/${PROJECT_SLUG}/github is required and absent" >&2
  exit 1
fi

vault kv get -field=token "secret/${PROJECT_SLUG}/github" \
  | gh secret set GH_TERRAFORM_TOKEN -R "$REPO" --stdin
echo "    GH_TERRAFORM_TOKEN — set"

# ── Verify ───────────────────────────────────────────────────────────────────────

echo ""
echo "==> Verifying..."
gh secret list -R "$REPO"
gh variable list -R "$REPO"

if [ "${#SKIPPED[@]}" -gt 0 ]; then
  echo ""
  echo "==> Skipped (expected for provisioning_mode: existing-shared): ${SKIPPED[*]}"
fi

echo ""
echo "==> Sync complete."

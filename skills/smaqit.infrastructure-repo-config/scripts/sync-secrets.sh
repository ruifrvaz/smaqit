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

# ── Scheme detection: new (apps/+machines/) vs legacy flat ─────────────────────
#
# ssh/github always come from the app root; tfstate/cyso come from the machine root. On the
# legacy flat scheme, both roots collapse to the same secret/<project-slug>/* path, so the rest
# of this script doesn't need to branch any further than resolving these two roots once.

APP_ROOT="secret/${PROJECT_SLUG}"
MACHINE_ROOT="secret/${PROJECT_SLUG}"
if vault kv get "secret/apps/${PROJECT_SLUG}/machine" > /dev/null 2>&1; then
  MACHINE_SLUG=$(vault kv get -field=machine-slug "secret/apps/${PROJECT_SLUG}/machine")
  APP_ROOT="secret/apps/${PROJECT_SLUG}"
  MACHINE_ROOT="secret/machines/${MACHINE_SLUG}"
  echo "==> Scheme: new (${APP_ROOT}/*, ${MACHINE_ROOT}/*)"
else
  echo "==> Scheme: legacy (${APP_ROOT}/*)"
fi

app_path_exists() {
  vault kv get "${APP_ROOT}/$1" > /dev/null 2>&1
}
machine_path_exists() {
  vault kv get "${MACHINE_ROOT}/$1" > /dev/null 2>&1
}

SKIPPED=()

echo "==> Syncing secrets/variables for ${PROJECT_SLUG} -> ${REPO}"

# ── SSH deploy key (always required, app-scoped) ───────────────────────────────

if ! app_path_exists ssh; then
  echo "sync-secrets: ERROR — ${APP_ROOT}/ssh is required and absent" >&2
  exit 1
fi

# tr -d '\n' strips trailing newline — without it, Terraform marks the keypair for
# replacement on every plan.
vault kv get -field=private_key "${APP_ROOT}/ssh" \
  | tr -d '\n' \
  | gh secret set VM_SSH_KEY -R "$REPO" --stdin
echo "    VM_SSH_KEY — set"

vault kv get -field=public_key "${APP_ROOT}/ssh" \
  | gh secret set VM_SSH_PUBLIC_KEY -R "$REPO" --stdin
echo "    VM_SSH_PUBLIC_KEY — set"

# ── VM host (always required — a variable, not a secret; see SKILL.md) ────────

gh variable set VM_HOST --body "$VM_HOST_VALUE" -R "$REPO"
echo "    VM_HOST (variable) — set"

# ── Terraform backend credentials (machine-scoped; skip cleanly if absent) ────

if machine_path_exists tfstate; then
  vault kv get -field=access_key "${MACHINE_ROOT}/tfstate" \
    | gh secret set TF_BACKEND_ACCESS_KEY -R "$REPO" --stdin
  vault kv get -field=secret_key "${MACHINE_ROOT}/tfstate" \
    | gh secret set TF_BACKEND_SECRET_KEY -R "$REPO" --stdin
  echo "    TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY — set"
else
  echo "    SKIP — ${MACHINE_ROOT}/tfstate absent (expected for provisioning_mode: existing-shared)"
  SKIPPED+=("TF_BACKEND_ACCESS_KEY" "TF_BACKEND_SECRET_KEY")
fi

# ── Cloud provider credentials (machine-scoped; skip cleanly if absent) ───────

if machine_path_exists cyso; then
  vault kv get -field=app_credential_id "${MACHINE_ROOT}/cyso" \
    | gh secret set OS_APPLICATION_CREDENTIAL_ID -R "$REPO" --stdin
  vault kv get -field=app_credential_secret "${MACHINE_ROOT}/cyso" \
    | gh secret set OS_APPLICATION_CREDENTIAL_SECRET -R "$REPO" --stdin
  echo "    OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET — set"
else
  echo "    SKIP — ${MACHINE_ROOT}/cyso absent (expected for provisioning_mode: existing-shared)"
  SKIPPED+=("OS_APPLICATION_CREDENTIAL_ID" "OS_APPLICATION_CREDENTIAL_SECRET")
fi

# ── GitHub Terraform token (always required, app-scoped) ───────────────────────
# CRITICAL: the workflow YAML env var for this secret MUST be TF_VAR_github_token,
# NOT GITHUB_TOKEN — that name is reserved and the runner overwrites it before any
# step executes.

if ! app_path_exists github; then
  echo "sync-secrets: ERROR — ${APP_ROOT}/github is required and absent" >&2
  exit 1
fi

vault kv get -field=token "${APP_ROOT}/github" \
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

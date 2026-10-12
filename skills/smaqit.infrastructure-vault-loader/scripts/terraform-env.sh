# smaqit.infrastructure-vault-loader — export Terraform's provider and backend credentials.
# SOURCE this file (don't execute it): it exports into the calling shell, the only way Terraform
# can read them. Values go from Vault into environment variables and are never printed.
#
#   PROJECT_SLUG=<project-slug> MACHINE_SLUG=<machine-slug> \
#     source [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/terraform-env.sh
#   ...terraform init / plan-guard.sh / terraform apply...
#   source [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/terraform-env.sh --unset
#
# Exports, each from the first Vault path that exists:
#   TF_VAR_app_credential_id/_secret  secret/<project>/cyso,    then secret/machines/<machine>/cyso
#   AWS_ACCESS_KEY_ID/_SECRET_ACCESS_KEY  secret/<project>/tfstate, then secret/machines/<machine>/tfstate
#   TF_VAR_github_token               secret/<project>/github,  then secret/apps/<project>/github
#   TF_VAR_ssh_public_key             secret/machines/<machine>/base-ssh when MACHINE_SLUG is set,
#                                     otherwise secret/<project>/ssh
# PROJECT_SLUG defaults to the repository's slug (lib-project-slug.sh). Returns non-zero, naming
# the missing path, and exports nothing when any credential is absent.

_smaqit_tf_vars="TF_VAR_app_credential_id TF_VAR_app_credential_secret TF_VAR_github_token TF_VAR_ssh_public_key AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY"

if ! (return 0 2>/dev/null); then
  echo "ERROR: source this file instead of running it: source $0" >&2
  exit 2
fi

# _smaqit_tf_pick <field> <path>... — print the first path holding <field>, or nothing.
_smaqit_tf_pick() {
  local field="$1" p; shift
  for p in "$@"; do
    [ -n "$p" ] || continue
    if vault kv get -field="$field" "$p" > /dev/null 2>&1; then
      printf '%s' "$p"
      return 0
    fi
  done
  return 1
}

_smaqit_tf_env() {
  if [ "${1:-}" = "--unset" ]; then
    # shellcheck disable=SC2086
    unset $_smaqit_tf_vars
    echo "terraform-env: unset ${_smaqit_tf_vars}"
    return 0
  fi
  [ "$#" -eq 0 ] || { echo "Usage: source terraform-env.sh [--unset]" >&2; return 2; }

  export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
  if ! vault token lookup > /dev/null 2>&1; then
    echo "ERROR: No valid Vault token. Run load-credentials.sh first." >&2
    return 1
  fi

  local here project machine
  here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  project="${PROJECT_SLUG:-}"
  if [ -z "$project" ]; then
    # shellcheck source=lib-project-slug.sh
    source "${here}/lib-project-slug.sh"
    project="$(derive_project_slug)"
  fi
  [ -n "$project" ] || { echo "ERROR: set PROJECT_SLUG." >&2; return 1; }
  machine="${MACHINE_SLUG:-}"

  local cyso tfstate github sshpub missing=""
  cyso="$(_smaqit_tf_pick app_credential_id "secret/${project}/cyso" ${machine:+"secret/machines/${machine}/cyso"})" || missing="${missing} cyso"
  tfstate="$(_smaqit_tf_pick access_key "secret/${project}/tfstate" ${machine:+"secret/machines/${machine}/tfstate"})" || missing="${missing} tfstate"
  github="$(_smaqit_tf_pick token "secret/${project}/github" "secret/apps/${project}/github")" || missing="${missing} github"
  if [ -n "$machine" ]; then
    sshpub="$(_smaqit_tf_pick public_key "secret/machines/${machine}/base-ssh")" || missing="${missing} machines/${machine}/base-ssh"
  else
    sshpub="$(_smaqit_tf_pick public_key "secret/${project}/ssh")" || missing="${missing} ${project}/ssh"
  fi
  if [ -n "$missing" ]; then
    echo "ERROR: no Vault path holds:${missing} (project '${project}'${machine:+, machine '${machine}'}). Nothing exported." >&2
    echo "       Load them with load-credentials.sh or register-machine.sh first." >&2
    return 1
  fi

  TF_VAR_app_credential_id="$(vault kv get -field=app_credential_id "$cyso")" &&
  TF_VAR_app_credential_secret="$(vault kv get -field=app_credential_secret "$cyso")" &&
  AWS_ACCESS_KEY_ID="$(vault kv get -field=access_key "$tfstate")" &&
  AWS_SECRET_ACCESS_KEY="$(vault kv get -field=secret_key "$tfstate")" &&
  TF_VAR_github_token="$(vault kv get -field=token "$github")" &&
  TF_VAR_ssh_public_key="$(vault kv get -field=public_key "$sshpub")" || {
    echo "ERROR: reading a credential from Vault failed. Nothing exported." >&2
    # shellcheck disable=SC2086
    unset $_smaqit_tf_vars
    return 1
  }
  export TF_VAR_app_credential_id TF_VAR_app_credential_secret AWS_ACCESS_KEY_ID \
    AWS_SECRET_ACCESS_KEY TF_VAR_github_token TF_VAR_ssh_public_key
  echo "terraform-env: exported ${_smaqit_tf_vars} (from ${cyso}, ${tfstate}, ${github}, ${sshpub})"
}

_smaqit_tf_env "$@"

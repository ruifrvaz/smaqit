# Skill Definition: smaqit.infrastructure-repo-config

## Identity
- **Name:** smaqit.infrastructure-repo-config
- **Version:** 1.0.0
- **Description:** Configure a GitHub repository with all secrets and variables required for CI/CD workflows. Covers: Actions secrets (SSH key, Terraform backend credentials, API tokens), Actions variables (DEMO_MODE), and deployment environment wiring. Uses the `gh` CLI. Prevents reserved-name collisions and key-format drift.

## Pre-conditions
- `gh` CLI authenticated (`gh auth login`)
- Target repository exists on GitHub
- SSH key for VM access generated (passphrase-free deploy key)
- Cyso/cloud application credential ID and secret available
- Terraform backend S3 credentials available
- Fine-grained PAT with `variables:write` repository permission available (for GH_TERRAFORM_TOKEN)

## Steps
1. **Confirm repository** — resolve `<owner>/<repo>` from context (copilot-instructions.md or user input).
2. **Set SSH deploy key** — the key MUST be passphrase-free (standard `him_key` has a passphrase; generate a separate deploy key):
   ```
   gh secret set VM_SSH_KEY --body "$(cat ~/.ssh/<deploy-key> | tr -d '\n')" -R <owner>/<repo>
   ```
   CRITICAL: pipe through `tr -d '\n'` to strip trailing newline. GitHub Secrets UI preserves trailing `\n`; `$(cat file.pub)` in shell strips it → Terraform sees a string diff on every plan and marks the keypair for replacement.
3. **Set VM host secret:**
   ```
   gh secret set VM_HOST --body "<vm-ip-or-hostname>" -R <owner>/<repo>
   ```
4. **Set Terraform backend credentials:**
   ```
   gh secret set TF_BACKEND_ACCESS_KEY --body "<s3-access-key>" -R <owner>/<repo>
   gh secret set TF_BACKEND_SECRET_KEY --body "<s3-secret-key>" -R <owner>/<repo>
   ```
5. **Set cloud provider credentials:**
   ```
   gh secret set TF_VAR_APP_CREDENTIAL_ID --body "<app-credential-id>" -R <owner>/<repo>
   gh secret set TF_VAR_APP_CREDENTIAL_SECRET --body "<app-credential-secret>" -R <owner>/<repo>
   ```
6. **Set GitHub Terraform token** (for `github_actions_variable` Terraform resource):
   ```
   gh secret set GH_TERRAFORM_TOKEN --body "<fine-grained-pat>" -R <owner>/<repo>
   ```
   CRITICAL: The env var name in workflow YAML MUST be `TF_VAR_github_token`, NOT `GITHUB_TOKEN`. `GITHUB_TOKEN` is a reserved name — the GitHub Actions runner injects its own installation token under that name before any step runs, overwriting the PAT. The installation token has no `variables:write` scope → 401 on `github_actions_variable` resource.
7. **Set DEMO_MODE variable** (default `true` for new deployments):
   ```
   gh variable set DEMO_MODE --body "true" -R <owner>/<repo>
   ```
8. **Verify:** `gh secret list -R <owner>/<repo>` and `gh variable list -R <owner>/<repo>` — confirm all expected names appear.

## Output
- GitHub repository configured with all required secrets and variables
- Verification list confirming presence of each

## Scope
- Does NOT generate SSH keys. Caller must provide the key file path.
- Does NOT create or configure the GitHub repository itself.
- Does NOT manage environment-level secrets (only repository-level).

## Gotchas
- `GITHUB_TOKEN` is reserved — NEVER use it as the env var name for a PAT in workflow YAML. Use `TF_VAR_github_token` which is both non-reserved and auto-mapped by Terraform to `var.github_token`.
- `GH_TOKEN` vs `GH_TERRAFORM_TOKEN` — `GH_TOKEN` is the original (stale) secret name used in this project before the collision was discovered. The correct name is `GH_TERRAFORM_TOKEN`. If both exist, the workflow uses `GH_TERRAFORM_TOKEN`.
- SSH key trailing newline drift — always use `tr -d '\n'` when setting the SSH key. Without this, Terraform will flag the keypair resource for replacement on every plan.
- Fine-grained PAT required scope — repository permissions → Variables: Read and write. Classic PATs are rejected by the GitHub provider in Terraform.
- `gh auth login` scope — ensure the `gh` session has `write:secrets` and `write:variables` permissions. `repo` scope alone is insufficient for variables.

## Completion
- [ ] Repository owner/name confirmed
- [ ] VM_SSH_KEY set (passphrase-free, trailing newline stripped)
- [ ] VM_HOST set
- [ ] TF_BACKEND_ACCESS_KEY and TF_BACKEND_SECRET_KEY set
- [ ] TF_VAR_APP_CREDENTIAL_ID and TF_VAR_APP_CREDENTIAL_SECRET set
- [ ] GH_TERRAFORM_TOKEN set (fine-grained PAT, variables:write scope)
- [ ] DEMO_MODE variable set (default: true)
- [ ] `gh secret list` and `gh variable list` verified

## Failure Handling
| Situation | Action |
|-----------|--------|
| `gh` not authenticated | Run `gh auth login` before proceeding |
| Secret value not available | Request the value from the operator before continuing; do not proceed with missing secrets |
| `GITHUB_TOKEN` collision detected in existing workflow | Flag it explicitly and require renaming before the workflow is triggered |
| `gh variable set` returns 403 | Check that the PAT used for `gh auth login` has `write:variables` scope |

## Examples
**Input:** New project repo `ruifrvaz/myapp` created. Operator invokes `/github.repo.config`.
**Output:** 7 secrets + 1 variable set. `gh secret list` confirms: VM_SSH_KEY, VM_HOST, TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY, TF_VAR_APP_CREDENTIAL_ID, TF_VAR_APP_CREDENTIAL_SECRET, GH_TERRAFORM_TOKEN. `gh variable list` confirms: DEMO_MODE=true.

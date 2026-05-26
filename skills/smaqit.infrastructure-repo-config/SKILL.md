---
name: smaqit.infrastructure-repo-config
description: Use when configuring a GitHub repository with the secrets and variables required for CI/CD workflows. Covers Actions secrets (VM_SSH_KEY, VM_HOST, Terraform backend credentials, cloud provider credentials, GH_TERRAFORM_TOKEN) and the DEMO_MODE Actions variable. Uses the `gh` CLI. Prevents GITHUB_TOKEN reserved-name collisions and SSH key trailing-newline drift. Also use when setting up a new deployment repository, rotating CI/CD credentials, or verifying that all required repository secrets and variables are present.
metadata:
  version: "1.0.0"
---

# Configure GitHub Repository Secrets and Variables

## Steps

**Pre-flight — verify before starting:**
- `gh` CLI is authenticated (`gh auth login`)
- Target repository exists on GitHub
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); all `secret/<project-slug>/*` paths populated
- `PROJECT_SLUG` and `VAULT_ADDR=http://127.0.0.1:8200` exported in current shell

> **Role of this skill:** Vault is the source of truth. GitHub Secrets are a derived copy. This skill
> reads from Vault and pushes to GitHub. On credential rotation, update Vault first, then re-run
> this skill to sync. No values are typed manually or sourced from disk.


1. **Confirm repository** — resolve `<owner>/<repo>` from `copilot-instructions.md` or user input.

2. **Set SSH deploy key** (piped from Vault — never touches disk):
   ```bash
   vault kv get -field=private_key secret/${PROJECT_SLUG}/ssh \
     | tr -d '\n' \
     | gh secret set VM_SSH_KEY -R <owner>/<repo> --stdin
   ```
   `tr -d '\n'` strips trailing newline. Without it, Terraform marks the keypair for replacement
   on every plan.

3. **Set VM host:**
   ```bash
   gh secret set VM_HOST --body "<vm-ip-or-hostname>" -R <owner>/<repo>
   ```
   Populated from `terraform output -raw fixed_ip` after Phase 4 provisioning.

4. **Set Terraform backend credentials:**
   ```bash
   vault kv get -field=access_key secret/${PROJECT_SLUG}/tfstate \
     | gh secret set TF_BACKEND_ACCESS_KEY -R <owner>/<repo> --stdin
   vault kv get -field=secret_key secret/${PROJECT_SLUG}/tfstate \
     | gh secret set TF_BACKEND_SECRET_KEY -R <owner>/<repo> --stdin
   ```

5. **Set cloud provider credentials:**
   ```bash
   vault kv get -field=app_credential_id secret/${PROJECT_SLUG}/cyso \
     | gh secret set OS_APPLICATION_CREDENTIAL_ID -R <owner>/<repo> --stdin
   vault kv get -field=app_credential_secret secret/${PROJECT_SLUG}/cyso \
     | gh secret set OS_APPLICATION_CREDENTIAL_SECRET -R <owner>/<repo> --stdin
   ```

6. **Set GitHub Terraform token:**
   ```bash
   vault kv get -field=token secret/${PROJECT_SLUG}/github \
     | gh secret set GH_TERRAFORM_TOKEN -R <owner>/<repo> --stdin
   ```
   CRITICAL: The workflow YAML env var for this secret MUST be `TF_VAR_github_token`, NOT `GITHUB_TOKEN`.
   `GITHUB_TOKEN` is reserved — the runner injects its own installation token under that name before
   any step executes, overwriting the PAT. The installation token has no `variables:write` scope →
   401 on `github_actions_variable` resources.

7. **Set VM SSH public key:**
   ```bash
   vault kv get -field=public_key secret/${PROJECT_SLUG}/ssh \
     | gh secret set VM_SSH_PUBLIC_KEY -R <owner>/<repo> --stdin
   ```

8. **Verify:**
   ```bash
   gh secret list -R <owner>/<repo>
   gh variable list -R <owner>/<repo>
   ```
   Confirm all expected names appear (see Completion checklist).

## Output

- GitHub repository configured with 7 secrets: VM_SSH_KEY, VM_SSH_PUBLIC_KEY, VM_HOST, TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY, OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET, GH_TERRAFORM_TOKEN
- All values sourced from Vault; no credentials typed or stored locally outside Vault
- Verification output confirming presence of each name

## Scope

- Does NOT generate SSH keys — the caller must provide a passphrase-free deploy key file path
- Does NOT create or configure the GitHub repository itself
- Does NOT manage environment-level secrets (repository-level only)

## Examples

**Input:** New project repo `ruifrvaz/myapp` created. Operator invokes the skill.
**Output:** `gh secret list` confirms: VM_SSH_KEY, VM_HOST, TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY, TF_VAR_APP_CREDENTIAL_ID, TF_VAR_APP_CREDENTIAL_SECRET, GH_TERRAFORM_TOKEN. `gh variable list` confirms: DEMO_MODE=true.

## Gotchas

- **`GITHUB_TOKEN` is reserved** — never map `GH_TERRAFORM_TOKEN` to `GITHUB_TOKEN` in workflow YAML. Use `TF_VAR_github_token` (non-reserved and auto-mapped by Terraform to `var.github_token`).
- **`GH_TOKEN` vs `GH_TERRAFORM_TOKEN`** — `GH_TOKEN` is the legacy name used before the collision was discovered. If both exist, workflows use `GH_TERRAFORM_TOKEN`. Remove `GH_TOKEN` to avoid confusion.
- **SSH key trailing newline** — always pipe through `tr -d '\n'` when setting VM_SSH_KEY. Without this, Terraform flags the keypair resource for replacement on every plan.
- **Fine-grained PAT scope** — repository permissions → Variables: Read and write. Classic PATs are rejected by the GitHub Terraform provider.
- **`gh auth login` scope** — ensure the `gh` session has `write:secrets` and `write:variables`. The `repo` scope alone is insufficient for variables.

## Completion

- [ ] Repository owner/name confirmed
- [ ] VM_SSH_KEY set (from Vault, trailing newline stripped)
- [ ] VM_SSH_PUBLIC_KEY set (from Vault)
- [ ] VM_HOST set
- [ ] TF_BACKEND_ACCESS_KEY and TF_BACKEND_SECRET_KEY set (from Vault)
- [ ] OS_APPLICATION_CREDENTIAL_ID and OS_APPLICATION_CREDENTIAL_SECRET set (from Vault)
- [ ] GH_TERRAFORM_TOKEN set (from Vault; fine-grained PAT, `variables:write` scope)
- [ ] `gh secret list` and `gh variable list` verified — all expected names present

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `gh` not authenticated | Run `gh auth login` before proceeding |
| Secret value not available | Request the value from the operator; do not proceed with missing secrets |
| `GITHUB_TOKEN` collision in existing workflow YAML | Flag it explicitly and require renaming before the workflow is triggered |
| `gh variable set` returns 403 | Verify the PAT used for `gh auth login` has `write:variables` scope |

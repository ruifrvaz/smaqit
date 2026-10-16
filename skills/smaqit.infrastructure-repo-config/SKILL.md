---
name: smaqit.infrastructure-repo-config
description: 'Use when setting up, rotating, or verifying the GitHub Actions secrets and variables a project''s CI/CD needs, via the gh CLI: VM_SSH_KEY, Terraform backend and cloud credentials, GH_TERRAFORM_TOKEN, VM_HOST and DEMO_MODE, or for existing-k3s the environment-scoped KUBECONFIG secret, a private-mesh machine''s setup key, and an optional registry credential.'
metadata:
  version: "2.0.0"
  smaqit-slots: "environment-config=sync, environment-config=sync-machine"
---

# Configure GitHub Repository Secrets and Variables

## Steps

**Pre-flight — verify before starting:**
- `gh` CLI is authenticated (`gh auth login`)
- Target repository exists on GitHub
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete)
- `PROJECT_SLUG` and `VAULT_ADDR=http://127.0.0.1:8200` exported in current shell
- **Scheme-aware:** `scripts/sync-secrets.sh` auto-detects the new `apps/`+`machines/` scheme
  (via `secret/apps/<app-slug>/machine`) vs. the legacy flat `secret/<project-slug>/*` scheme.
  `ssh`/`github` always read from the app root; `tfstate`/`cyso` always read from the machine
  root — on the legacy scheme both roots are the same `secret/<project-slug>/*` path. This skill
  never reads or writes `secret/machines/*` directly for anything beyond `tfstate`/`cyso`.
- **Restricted mode for `provisioning_mode: existing-shared` or `existing-unmanaged`:** only `ssh`
  and `github` are expected to be populated — `tfstate` and `cyso` are never populated for either
  mode (see `smaqit.infrastructure-vault-loader`) and `scripts/sync-secrets.sh` (Step 3) adapts
  accordingly rather than assuming all four paths exist.
- **`provisioning_mode: existing-k3s` is a different shape entirely, not a restricted variant of
  the VM path** — no `ssh`, `tfstate`, `cyso`, or `VM_SSH_KEY`/`VM_HOST` at all. Only `github` and
  the target environment's registered machine-slug's
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig` (`value` field) are read, and the output
  lands on the GitHub **Environment named after each target machine slug**, not the repository
  level. Steps 2–4 don't apply; go straight to Step 5.

> **Role of this skill:** Vault is the source of truth. GitHub Secrets are a derived copy. This skill
> reads from Vault and pushes to GitHub. On credential rotation, update Vault first, then re-run
> this skill to sync. No values are typed manually or sourced from disk.


1. **Confirm repository** — resolve `<owner>/<repo>` from `AGENTS.md` (or legacy platform-specific instructions) or user input.

2. **VM modes only (`provision`, `existing-owned`, `existing-shared`, `existing-unmanaged`) —
   resolve the VM host value** — from `terraform output -raw fixed_ip` after Phase 4
   provisioning (`provision`/`existing-owned`), or a manually-supplied value (`existing-shared` —
   there is no Terraform output on this project's side to derive it from; `existing-unmanaged` —
   there is no Terraform output at all, on any side).

3. **VM modes only — run the sync script** (it requires the app's `ssh` credential and a VM
   host, so it fails on an `existing-k3s` project) — this replaces separate Vault-read and `gh secret set` steps
   with one deterministic script that enforces the `tfstate`/`cyso` skip-if-absent logic
   structurally, and prints its own verification report:
   ```bash
   scripts/sync-secrets.sh <owner>/<repo> <vm-host-value>
   ```
   It syncs, in order: `VM_SSH_KEY` (trailing newline stripped — without this, Terraform marks the
   keypair for replacement on every plan) and `VM_SSH_PUBLIC_KEY` from `secret/<slug>/ssh`;
   `VM_HOST` as a repository **variable**, not a secret — it's just an IP/hostname, not sensitive,
   and unlike a secret it can be read back via `gh variable get`, which is what
   `smaqit.infrastructure-provision-cyso`'s ownership guard relies on; `TF_BACKEND_ACCESS_KEY`/
   `TF_BACKEND_SECRET_KEY` from `secret/<slug>/tfstate` — skipped cleanly and reported (not an
   error) if that path is absent, expected for `existing-shared`/`existing-unmanaged`; `OS_APPLICATION_CREDENTIAL_ID`/
   `OS_APPLICATION_CREDENTIAL_SECRET` from `secret/<slug>/cyso` — same skip-if-absent handling; and
   `GH_TERRAFORM_TOKEN` from `secret/<slug>/github`. CRITICAL: the workflow YAML env var for that
   last one MUST be `TF_VAR_github_token`, NOT `GITHUB_TOKEN` — that name is reserved and the
   runner overwrites it before any step executes, and the installation token has no
   `variables:write` scope, causing a 401 on `github_actions_variable` resources.

4. **Review the script's verification output** — it runs `gh secret list -R <owner>/<repo>` and
   `gh variable list -R <owner>/<repo>` itself and reports which secrets (if any) were skipped and
   why. Confirm all expected names appear (see Completion checklist) before considering this done.

5. **`existing-k3s` only — sync each target machine's GitHub Environment**, once per
   environment the Infrastructure spec declares (typically a test machine and a prod machine):
   ```bash
   scripts/sync-k3s-env.sh <owner>/<repo> <app-slug> <machine-slug> [--mesh-secret <name>] [--registry-org <org-slug>]
   ```
   The Environment is named after the machine slug — the same slug that keys the app's
   kubeconfig in Vault — so each environment's distinct scoped kubeconfig has exactly one home;
   `<machine-slug>` is the one the Infrastructure spec registers for that environment, distinct
   per environment even when two share one physical k3s server. The script:
   - stops before writing anything if `secret/apps/<app-slug>/<machine-slug>/kubeconfig` is
     absent — a merged onboarding request isn't enough; the platform's own hand-off must have
     delivered the kubeconfig and the `target-credentials` load stored it;
   - creates or verifies the Environment and runs `scripts/check-pat-permissions.sh` against it;
   - pipes the kubeconfig into the `KUBECONFIG` **secret** from Vault on stdin;
   - with `--mesh-secret <name>` — for an environment whose Environments row is `private-mesh`,
     where `<name>` is that row's `Mesh Setup-Key Secret` — pipes the platform-issued setup key
     from `secret/apps/<app-slug>/<machine-slug>/mesh-setup-key` into that secret, and stops
     before writing anything if the key isn't in Vault yet (load it with
     `PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> MESH_SETUP_KEY=1
     load-credentials.sh`);
   - with `--registry-org`, also syncs `REGISTRY_USERNAME`/`REGISTRY_TOKEN` from the org-scoped,
     read-only `secret/organizations/<org-slug>/github-package-read` — only when the spec's
     `Container Registry` Constraints row declares a private registry; never a platform-owned
     credential, never a per-app path invented for this purpose;
   - lists the Environment's secrets and variables and fails if any expected name is absent.

   No `ssh`/`tfstate`/`cyso`-derived secret is ever synced for this mode, and `VM_HOST` is never
   set. The generated workflow reads each machine's host from the Infrastructure spec's
   Environments table, so `APP_HOST` is no longer needed; `--host` still sets it for a hand-written
   workflow that reads it.

### Operation `sync-machine` — a machine's GitHub Environment (platform repositories)

For a repository that owns machines rather than an app. Each machine gets one GitHub
**Environment named after its machine slug**, populated from `secret/machines/<machine-slug>/*`
(written by `smaqit.infrastructure-vault-loader`'s `register-machine.sh`). Bind it in
`.smaqit/slots.yml` as `environment-config: {skill: smaqit.infrastructure-repo-config, operation:
sync-machine}`; this skill declares two `environment-config` operations, so the binding must name
one. Run once per machine:

```bash
scripts/sync-machine-env.sh <owner>/<repo> <machine-slug> [--ssh-user <user>]
```

The script stops before writing anything if any of `base-ssh`, `cyso`, `tfstate`, or `metadata`
is missing, naming the `register-machine.sh` command that fills it. It then:

1. Creates or verifies the Environment (`gh api repos/<owner>/<repo>/environments/<machine-slug> -X PUT`
   — `gh` has no Environment command). The slug must be lowercase letters, digits, and inner
   hyphens; `gh` does not URL-encode Environment names.
2. Runs `scripts/check-pat-permissions.sh` against that Environment.
3. Sets six Environment **secrets**, each piped from Vault to `gh secret set` on stdin:

   | Secret | Vault source |
   |---|---|
   | `BASE_SSH_PUBLIC_KEY` / `BASE_SSH_PRIVATE_KEY` | `base-ssh` `public_key` / `private_key` |
   | `CYSO_APPLICATION_CREDENTIAL_ID` / `CYSO_APPLICATION_CREDENTIAL_SECRET` | `cyso` `app_credential_id` / `app_credential_secret` |
   | `CYSO_S3_ACCESS_KEY` / `CYSO_S3_SECRET_KEY` | `tfstate` `access_key` / `secret_key` |

4. Sets two Environment **variables**: `RECONCILE_SSH_USER` (`--ssh-user`, default `ubuntu`) and
   `FIXED_IP` (`metadata.host`).
5. Lists the Environment's secrets and variables and fails if any expected name is absent.

**It runs twice for a new machine.** `environment-config` comes before `provision` in the slot
order, and CI needs the credentials secrets before the gated apply, so the first run happens while
the machine has no address: `metadata.host` is unset and `FIXED_IP` is skipped with a notice (exit
0). After apply, backfill the address (`register-machine.sh <machine-slug> --host <fixed_ip>`) and
re-run — directly, or as a single-slot run (`smaqit slots plan … --slot=environment-config`) — to
set `FIXED_IP`. Re-run it likewise after rotating any of the machine's credentials in Vault.

## Output

- **`provision` / `existing-owned`:** GitHub repository configured with 7 secrets (VM_SSH_KEY, VM_SSH_PUBLIC_KEY, TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY, OS_APPLICATION_CREDENTIAL_ID, OS_APPLICATION_CREDENTIAL_SECRET, GH_TERRAFORM_TOKEN) plus the VM_HOST variable
- **`existing-shared` / `existing-unmanaged`:** 3 secrets only (VM_SSH_KEY, VM_SSH_PUBLIC_KEY, GH_TERRAFORM_TOKEN) plus the VM_HOST variable (set manually, not derived from Terraform) — identical set for both modes; only the reason there's no Terraform output differs (another project's Terraform vs. no Terraform at all)
- **`existing-k3s`:** a `KUBECONFIG` secret on the GitHub Environment named after each target machine slug (never a repository-level secret), plus the spec-named mesh setup-key secret on each `private-mesh` machine's Environment; if the spec's `Container Registry` row declares a private registry, also `REGISTRY_USERNAME`/`REGISTRY_TOKEN` on each Environment, sourced from the shared `secret/organizations/<org-slug>/github-package-read`. No `VM_HOST`, no `VM_SSH_KEY`, no Terraform-derived secret at all.
- **`sync-machine`:** a GitHub Environment named after the machine slug holding 6 secrets
  (`BASE_SSH_PUBLIC_KEY`, `BASE_SSH_PRIVATE_KEY`, `CYSO_APPLICATION_CREDENTIAL_ID`,
  `CYSO_APPLICATION_CREDENTIAL_SECRET`, `CYSO_S3_ACCESS_KEY`, `CYSO_S3_SECRET_KEY`) and 2 variables
  (`RECONCILE_SSH_USER`, and `FIXED_IP` once the machine's address is known). Nothing at the
  repository level.
- All values sourced from Vault; no credentials typed or stored locally outside Vault
- Verification output confirming presence of each name; absent `tfstate`/`cyso`-derived secrets are reported as skipped, not missing

## Scope

- Does NOT generate SSH keys — the caller must provide a passphrase-free deploy key file path
- Does NOT create or configure the GitHub repository itself
- Does NOT manage environment-level secrets, except for `provisioning_mode: existing-k3s` — that
  mode's `KUBECONFIG` and mesh setup-key secrets are deliberately GitHub Environment-scoped
  (one per target machine slug), not repository-level, since each environment needs its own distinct value —
  and the `sync-machine` operation, whose output is entirely one machine-slug Environment; the
  `sync` operation's VM-mode output stays repository-level only
- Does NOT hard-fail when `tfstate`/`cyso` Vault paths are absent — see Step 3 / `scripts/sync-secrets.sh` for `provisioning_mode: existing-shared`/`existing-unmanaged` handling
- Does NOT contact a Kubernetes cluster for `existing-k3s` — the kubeconfig is read from Vault and written to GitHub verbatim, never validated against the cluster itself

## Examples

**Input:** New project repo `ruifrvaz/myapp` created. Operator invokes the skill.
**Output:** `gh secret list` confirms: VM_SSH_KEY, TF_BACKEND_ACCESS_KEY, TF_BACKEND_SECRET_KEY, TF_VAR_APP_CREDENTIAL_ID, TF_VAR_APP_CREDENTIAL_SECRET, GH_TERRAFORM_TOKEN. `gh variable list` confirms: VM_HOST, DEMO_MODE=true.

**Input:** Namespace-scoped Kubernetes/k3s target, `provisioning_mode: existing-k3s`. Operator invokes the skill.
**Output:** `scripts/sync-k3s-env.sh` runs once per target machine; `gh secret list --env <test-machine-slug>` and `--env <prod-machine-slug>` each confirm `KUBECONFIG`, and the private-mesh test machine's also confirms its `MESH_SETUP_KEY`. No `VM_HOST`, no `VM_SSH_KEY`, at the repository level or either Environment.

## Gotchas

- **`GITHUB_TOKEN` is reserved** — never map `GH_TERRAFORM_TOKEN` to `GITHUB_TOKEN` in workflow YAML. Use `TF_VAR_github_token` (non-reserved and auto-mapped by Terraform to `var.github_token`).
- **`GH_TOKEN` vs `GH_TERRAFORM_TOKEN`** — `GH_TOKEN` is the legacy name used before the collision was discovered. If both exist, workflows use `GH_TERRAFORM_TOKEN`. Remove `GH_TOKEN` to avoid confusion.
- **SSH key trailing newline** — always pipe through `tr -d '\n'` when setting VM_SSH_KEY. Without this, Terraform flags the keypair resource for replacement on every plan.
- **Fine-grained PAT scope** — repository permissions → Variables: Read and write. Classic PATs are rejected by the GitHub Terraform provider.
- **Environment writes need organization-level grants on a fine-grained PAT** — repository-level Environments + Secrets + Variables read/write is not enough: writes to an Environment's secrets or variables return 403 until **organization-level Secrets and Variables read/write** is also enabled on the token. `scripts/check-pat-permissions.sh` probes for this read-only before any write, for both `existing-k3s` and `sync-machine`.
- **`gh auth login` scope** — ensure the `gh` session has `write:secrets` and `write:variables`. The `repo` scope alone is insufficient for variables.
- **`KUBECONFIG` must be set per Environment, never at the repository level** — a repository-level secret can only hold one value, but each environment resolves to its own registered machine slug and its own distinct scoped kubeconfig. Setting it at the repository level would silently make one environment deploy with the wrong credential.
- **`existing-k3s` Environments are named after machine slugs, not `test`/`prod`** — matching the Vault path `secret/apps/<app-slug>/<machine-slug>/kubeconfig` and the generated `deploy.yml`'s environment input. A project that set up `test`/`prod` Environments earlier syncs into the machine-slug Environments with `sync-k3s-env.sh` and deletes the old ones once its regenerated workflow uses the new names.
- **The mesh setup key is per machine and platform-issued** — like the kubeconfig, it lives at `secret/apps/<app-slug>/<machine-slug>/mesh-setup-key` and is synced only into that machine's Environment, under the secret name the spec's Environments row declares (the generated workflow reads `secrets[<that name>]`). Rotate it with `rotate-credential.sh apps/<app-slug>/<machine-slug>/mesh-setup-key`, then re-sync and redeploy with a fresh dispatch.
- **`APP_HOST` is no longer read by generated `k3s` workflows** — they carry each machine's host from the spec. An `APP_HOST` variable left on an Environment from an earlier sync is harmless; delete it once no hand-written workflow reads it.

## Completion

- [ ] Repository owner/name confirmed
- [ ] VM_SSH_KEY set (from Vault, trailing newline stripped)
- [ ] VM_SSH_PUBLIC_KEY set (from Vault)
- [ ] VM_HOST variable set — from Terraform output (`provision`/`existing-owned`) or manually (`existing-shared`/`existing-unmanaged`); never set for `existing-k3s`
- [ ] TF_BACKEND_ACCESS_KEY and TF_BACKEND_SECRET_KEY set (from Vault), or cleanly skipped if `secret/<slug>/tfstate` is absent
- [ ] OS_APPLICATION_CREDENTIAL_ID and OS_APPLICATION_CREDENTIAL_SECRET set (from Vault), or cleanly skipped if `secret/<slug>/cyso` is absent
- [ ] GH_TERRAFORM_TOKEN set (from Vault; fine-grained PAT, `variables:write` scope)
- [ ] `existing-k3s` only: `scripts/sync-k3s-env.sh` exited 0 for each target machine — `KUBECONFIG` secret on each machine-slug Environment, the spec-named mesh setup-key secret on each `private-mesh` one, plus `REGISTRY_USERNAME`/`REGISTRY_TOKEN` if the spec's `Container Registry` row declares a private registry
- [ ] `sync-machine` only: `scripts/sync-machine-env.sh` exited 0 — the machine-slug Environment exists with its 6 secrets and `RECONCILE_SSH_USER`, plus `FIXED_IP` after the post-apply re-run
- [ ] `gh secret list` and `gh variable list` verified — all expected names present for the active `provisioning_mode` (pass `--env <machine-slug>` for `existing-k3s` and `sync-machine`)

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `gh` not authenticated | Run `gh auth login` before proceeding |
| Secret value not available | Request the value from the operator; do not proceed with missing secrets |
| Machine root's `tfstate` or `cyso` absent | `scripts/sync-secrets.sh` skips that block cleanly and reports it — expected for `provisioning_mode: existing-shared` or `existing-unmanaged`, not an error |
| `scripts/sync-secrets.sh` exits 1 | App root's `ssh` or `github` is missing — these are required in every mode; populate them via `smaqit.infrastructure-vault-loader` before retrying |
| `GITHUB_TOKEN` collision in existing workflow YAML | Flag it explicitly and require renaming before the workflow is triggered |
| `gh variable set` returns 403 | Verify the PAT used for `gh auth login` has `write:variables` scope |
| `existing-k3s`: `secret/apps/<app-slug>/<machine-slug>/kubeconfig` absent for the target environment's registered machine-slug | Stop. This is `smaqit.infrastructure-vault-loader`'s job — populate it there (out-of-band paste from the platform's onboarding hand-off) before retrying; do not fabricate a placeholder |
| `existing-k3s`: `sync-k3s-env.sh` reports the kubeconfig missing | Stop. A merged onboarding request is not enough — the platform's own hand-off must deliver the kubeconfig; then load it with `PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> load-credentials.sh` and re-run |
| `existing-k3s`: `sync-k3s-env.sh` reports the mesh setup key missing | Load the platform-issued key: `PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> MESH_SETUP_KEY=1 load-credentials.sh`, then re-run |
| `existing-k3s`: `sync-k3s-env.sh` reports the registry credential missing | Load it once, org-wide: `PROVISIONING_MODE=existing-k3s CREDENTIAL_SCOPE=access REGISTRY_ORG=<org-slug> load-credentials.sh` |
| Environment create returns 403 "Resource not accessible by personal access token" | The token lacks Environments write or organization-level Secrets/Variables read/write; fix the fine-grained PAT, re-authenticate `gh`, retry |
| `gh variable set --env <name>` returns 404 | The Environment doesn't exist yet — create it first (`sync-machine` does this itself) |
| `gh variable set --env <name>` returns 403 with repository-only grants | Enable organization-level Secrets and Variables read/write on the fine-grained PAT |
| `scripts/check-pat-permissions.sh` exits 2 or 3 | Exit 2: an Environment is missing; exit 3: a grant is missing. Apply the fix it prints; do not proceed to writes |
| `sync-machine`: a `secret/machines/<machine-slug>/*` path is missing | Stop. Register the machine with `smaqit.infrastructure-vault-loader`'s `register-machine.sh`, then retry |
| `sync-machine`: reports `FIXED_IP skipped` | Expected before the machine is provisioned. After apply, run `register-machine.sh <machine-slug> --host <fixed_ip>` and re-run `sync-machine` |
| `existing-k3s`: spec's `Container Registry` row declares a private registry but `secret/organizations/<org-slug>/github-package-read` is absent | Stop. This is an operator-populated, org-scoped path (`smaqit.infrastructure-vault-loader`) — not something `load-credentials.sh`'s per-app flow ever creates; populate it there before retrying |

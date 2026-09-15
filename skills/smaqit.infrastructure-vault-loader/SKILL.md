---
name: smaqit.infrastructure-vault-loader
description: Use before any local deployment or credential operation that requires secrets from a local HashiCorp Vault instance. Verifies Vault is running, unsealed, and authenticated on 127.0.0.1:8200. Also runs an interactive credential loader script that prompts for all project secrets and writes them to Vault. Use for first-time setup, adding a new project's credentials, or when a Vault path is missing. Also use when setting up Vault for the first time on a new machine, or when a caller cannot reach Vault and needs troubleshooting guidance.
metadata:
  version: "3.6.0"
---

# Vault Loader

Ensures a local HashiCorp Vault instance is running, unsealed, and ready to serve credentials to
local deployment automation. Provides scripts that prompt for all project credentials securely —
all sensitive input uses hidden prompts; nothing sensitive is ever passed as a command argument
or written to shell history.

Pre-step for `smaqit.infrastructure-provision-cyso`, `smaqit.infrastructure-deploy-rsync`, and
`smaqit.infrastructure-repo-config` when run locally.

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/install-vault.sh` | Install the Vault binary (Ubuntu/Debian). Run once per machine. |
| `scripts/setup-vault.sh` | Create config, start server, run `vault operator init`, enable kv-v2. Run once per machine. |
| `scripts/load-credentials.sh` | Start + unseal + authenticate + load all credential paths. Run every session. |
| `scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>` | Give an app its own distinct SSH keypair, authorized on the named machine. Run once per app-to-machine pairing. |
| `scripts/rotate-credential.sh <path>` | Delete and re-populate a single credential path (e.g. `cyso`, `apps/<app-slug>/ssh`, `machines/<machine-slug>/base-ssh`). |

Config template used by `setup-vault.sh`: `assets/vault.hcl.template`

---

## Vault path convention

Credentials are split across two namespaces by what they belong to — the **machine** (a
provisioned VM) or the **app** (a project deployed onto one):

```
secret/machines/<machine-slug>/base-ssh   — private_key, public_key (bootstrap-only; see below)
secret/machines/<machine-slug>/cyso       — app_credential_id, app_credential_secret
secret/machines/<machine-slug>/tfstate    — access_key, secret_key
secret/machines/<machine-slug>/metadata   — host, provider, owner_project (non-secret)

secret/apps/<app-slug>/ssh       — private_key, public_key (this app's own distinct keypair)
secret/apps/<app-slug>/github    — token (used as TF_VAR_github_token)
secret/apps/<app-slug>/machine   — machine-slug this app is bootstrapped against (non-secret)
secret/apps/<app-slug>/<machine-slug>/kubeconfig — value (a full scoped kubeconfig for that
                                     machine's Namespace; provisioning_mode: existing-k3s only)
secret/apps/<app-slug>/platform-repo — token (PR-create rights only — contents:write +
                                     pull_requests:write — on the platform-owned infrastructure
                                     repo; provisioning_mode: existing-k3s only)

secret/organizations/<org-slug>/github-package-read — username, token (a classic PAT scoped to
                                     exactly read:packages; provisioning_mode: existing-k3s with a
                                     private Container Registry only)
```

`cyso` and `tfstate` are machine-scoped: provisioning a VM is a property of the machine, not of
any individual app deployed onto it. `ssh` and `github` are app-scoped: every app always gets its
own distinct keypair, bootstrapped against its machine's `base-ssh` credential (see "Bootstrapping
an app onto a machine" below) — never copied from another app and never installed by Terraform
directly. `<app-slug>` and `<machine-slug>` are lowercase hyphenated slugs; for an app, this is the
project name declared in `AGENTS.md` (or the legacy platform-specific `CLAUDE.md` /
`copilot-instructions.md`) — same derivation as the legacy `<project-slug>` below, renamed to match
where it now lives in Vault. `machines` is a reserved app-slug.

`kubeconfig` is a different shape entirely, and it introduces its own kind of "machine" concept
that has no relationship to the VM `<machine-slug>` used above. A `provisioning_mode: existing-k3s`
app is keyed by one or more k3s machine-slugs — one per environment, or one per environment even
when test and prod happen to share one physical k3s server (in which case they still get two
distinct registered machine-slugs, e.g. `test-cluster` / `prod-cluster`, never one path with two
fields) — and each holds exactly one credential: a single field named `value` holding the full
scoped kubeconfig for that machine's Namespace. **There is no `secret/machines/<k3s-machine-slug>
/*` counterpart.** Unlike a VM `<machine-slug>`, a k3s machine-slug has no `base-ssh`/`cyso`/
`tfstate`/`metadata` sibling anywhere under `machines/` — none of those concepts apply to a
Namespace-scoped Kubernetes target: there is no host to SSH into, no Terraform state, no cloud
credential to provision it. `kubeconfig` is the only thing ever stored keyed by a k3s
machine-slug, and it lives under `apps/<app-slug>/<machine-slug>/kubeconfig` rather than under
`machines/<machine-slug>/*`, because an app can be onboarded onto multiple k3s machine-slugs over
its lifetime (test-cluster, prod-cluster, or a future migration to a new machine) while there is
nothing else machine-level to store for any of them — a `machines/` root exists to hold multiple
credential types per machine, and a k3s machine-slug only ever has one. "Environment" is expressed
entirely through which machine-slug is targeted, never through a field name on a shared path.
Like every other credential here except `ssh`'s bootstrap step, `load-credentials.sh` never
generates or validates it: the value always comes from the platform's own out-of-band onboarding
hand-off (a workflow artifact or its external secrets store) and is pasted in verbatim — this
script never makes a cluster call to obtain or check it.

`platform-repo` is a third, distinct app-scoped field for `existing-k3s`: a fine-grained PAT
scoped only to `contents:write` + `pull_requests:write` on the platform-owned infrastructure
repository — never broader, and never the same token as `github` (which is scoped to *this*
project's own repository for `TF_VAR_github_token`/`variables:write`). It exists to let
`smaqit.infrastructure-request-k3s-onboarding` push a branch and open a PR against the platform
repository's registry file without ever needing direct commit or `workflow_dispatch` access
there — convergence remains the platform team's own concern, triggered by their own repo reacting
to the merge. Unlike `kubeconfig`, `platform-repo` follows the standard delete-and-repopulate
rotation shape (see "Rotating a credential" below), since it's a smaqit-managed PAT the operator
can freely regenerate, not a platform-issued artifact.

`organizations/<org-slug>/github-package-read` is a fourth `existing-k3s` field, but in a
namespace of its own — sibling to `apps/`and `machines/`, never nested under either. It exists
only when a project's declared Container Registry (Infrastructure spec Constraints table) is
private: GHCR packages pushed from a private repository default to private visibility with no
opt-in step, so the cluster needs a pull credential to avoid `ImagePullBackOff`. Unlike every
other credential in this table, it is **org-scoped, not app-scoped** — one classic PAT (scoped to
exactly `read:packages`; fine-grained PATs have unreliable Packages API support), shared read-only
across every app on every machine in the org, never minted per app. `<org-slug>` is the GitHub
organization or user account that owns the repositories, not any one project's own slug. Because
it is shared rather than tied to a single app's lifecycle, `load-credentials.sh`'s per-app
`existing-k3s` prompt flow does not populate it — an operator sets it once, directly
(`vault kv put secret/organizations/<org-slug>/github-package-read username=... token=...`), the
same way a machine's `cyso`/`tfstate` credentials are populated once at machine registration
rather than per app per session. It is likewise **not** one of `rotate-credential.sh`'s supported
paths (org-scoped credentials sit outside that script's per-app/per-machine case matching) —
rotating it is the same manual `vault kv put`/`delete` an operator used to populate it, not a
scripted flow. `smaqit.infrastructure-repo-config` reads it to sync
`REGISTRY_USERNAME`/`REGISTRY_TOKEN` onto each GitHub Environment that declares a private
registry (see that skill's own Step 5).

**Legacy scheme, still in use, not migrated by this skill:** projects predating this convention
store everything flat under `secret/<project-slug>/{cyso,ssh,tfstate,github}`, with Terraform
installing the SSH key directly and no machine-level namespace at all. `load-credentials.sh`
supports both — see "Every session" below for how it decides which one applies to a given
invocation. Migrating an existing flat-scheme project onto `apps/`+`machines/` is a manual,
project-by-project decision, not something this skill does automatically.

---

## One-time setup (first run on a new machine)

### Step 1 — Install Vault

Run `scripts/install-vault.sh`. Idempotent — exits cleanly if already installed.

### Step 2 — Initialise Vault

Run `scripts/setup-vault.sh`. It creates `~/.vault/config.hcl` from the template, starts the
server, and runs `vault operator init`.

The init output contains the **Unseal Key** and **Root Token**. Store both offline immediately
(printed paper or password manager). Loss of the unseal key means permanent data loss.

After init completes, run `scripts/load-credentials.sh` to unseal and authenticate, then enable
the kv-v2 secrets engine once:

```
vault secrets enable -path=secret kv-v2
```

### Step 3 — Load project credentials

Run `scripts/load-credentials.sh`. It will prompt for credentials that are not yet populated.
The SSH deploy keypair is generated automatically — no prompt needed.

---

## Every session

Run `scripts/load-credentials.sh`. It handles start, unseal, login, and credential loading in
one pass. If all paths are already populated and a valid token exists, it exits immediately.

```
export VAULT_ADDR=http://127.0.0.1:8200
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

### Scheme detection: legacy flat vs. `apps/`+`machines/` vs. `existing-k3s`

`load-credentials.sh` decides which scheme applies per invocation:

- If `PROVISIONING_MODE=existing-k3s` is passed explicitly, it runs in **existing-k3s mode** and
  requires `MACHINE_SLUG` to be set (mirroring how the new-scheme's fresh-registration path
  already uses `MACHINE_SLUG` for `bootstrap-app-to-machine.sh`) — fails loudly if unset. One
  invocation checks/populates `secret/apps/<app-slug>/github` (once, idempotent),
  `secret/apps/<app-slug>/platform-repo` (once, idempotent), plus the one named machine's
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig`. There is no VM machine at
  all in this mode's sense — a Namespace-scoped cluster is a platform the app is onboarded onto,
  not a host this project registers — so `ssh`, `cyso`, `tfstate`, and `machine` are never
  touched, and there is no `secret/machines/<machine-slug>/*` counterpart for the k3s
  machine-slug either (see the path convention above). Run the script again with a different
  `MACHINE_SLUG` to populate a second machine's kubeconfig — e.g. once for the test cluster, once
  for prod; `github` and `platform-repo` are skipped as already-populated on that second run. The
  kubeconfig value is always an out-of-band paste of the platform's own onboarding hand-off, never
  generated or validated by a cluster call; `platform-repo`'s PAT is entered the same way (pasted,
  never generated by this script).
- If `secret/apps/<app-slug>/machine` already exists (this app has bootstrapped before), or a
  `MACHINE_SLUG` is passed explicitly, it runs in **new-scheme mode**: only `secret/apps/<app-slug>/
  github` is checked/populated. `ssh` is never touched here — it's exclusively the job of
  `bootstrap-app-to-machine.sh`, since populating it requires the machine's `base-ssh` credential to
  authorize the new key. `cyso`/`tfstate` are never prompted for — those live at
  `secret/machines/<machine-slug>/*`, populated once at machine registration
  (`smaqit.infrastructure-provision-cyso`'s post-apply step), not per app per session.
- Otherwise, it falls back to **legacy flat-scheme mode**, unchanged: all four paths under
  `secret/<project-slug>/{cyso,ssh,tfstate,github}` are checked/populated exactly as before,
  including the `PROVISIONING_MODE`-aware `existing-shared`/`existing-unmanaged` behavior
  (cyso/tfstate skipped for both; SSH offers a copy-or-generate choice for `existing-shared`,
  since another project's Terraform already has access, or always generates-and-prints manual
  `authorized_keys` install instructions for `existing-unmanaged`, since no Terraform run ever
  installs the key). This is what keeps unmigrated projects on the legacy flat scheme working
  without any change.

```
# New-scheme app, already bootstrapped — just tops up github if missing:
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

# Legacy flat-scheme project, co-hosted on another project's VM — unchanged:
PROVISIONING_MODE=existing-shared bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

# Legacy flat-scheme project, dedicated VM nobody's Terraform manages:
PROVISIONING_MODE=existing-unmanaged bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

# Namespace-scoped Kubernetes/k3s target — no VM, no SSH, kubeconfig only (one machine per invocation):
PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

---

## Bootstrapping an app onto a machine

Run once per app-to-machine pairing — including for the project that originally provisioned the
machine, which gets no shortcut around this:

```
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>
```

Idempotent: if `secret/apps/<app-slug>/ssh` is already populated *and* still authenticates
(checked live, not just presence in Vault), the script no-ops.

1. If `secret/machines/<machine-slug>/base-ssh` doesn't exist yet, this is a fresh machine — the
   script generates and stores the base keypair (plus `cyso`/`tfstate`/`metadata`, when provided)
   before continuing. This is the branch `provisioning_mode: existing-unmanaged` always exercises —
   an out-of-band-provisioned VM has typically never been registered by any smaqit-managed project
   before. When prompted for the owner project slug, use the *requesting* project's own slug: unlike
   `existing-shared` (where `owner_project` names whichever project's Terraform state provisions the
   machine), `existing-unmanaged` has no Terraform state anywhere in the picture, so there's no other
   project to name — the machine is simply registered as belonging to the project bootstrapping it.
2. Generates a new, distinct ed25519 keypair for `<app-slug>`.
3. Fetches the machine's `base-ssh` private key to a temp file (never displayed, never logged),
   uses it to SSH in, appends the new app keypair's public half to `~/.ssh/authorized_keys`,
   discards the temp file.
4. Stores `secret/apps/<app-slug>/ssh` and `secret/apps/<app-slug>/machine`.
5. Verifies the new keypair actually authenticates before reporting success.

---

## Rotating a credential

Run `scripts/rotate-credential.sh <path>` where `<path>` is one of `cyso`, `ssh`, `tfstate`,
`github` (legacy flat scheme, under `secret/<project-slug>/*`), or `apps/<app-slug>/ssh`,
`apps/<app-slug>/github`, `apps/<app-slug>/platform-repo`,
`apps/<app-slug>/<machine-slug>/kubeconfig`,
`machines/<machine-slug>/{base-ssh,cyso,tfstate}` (new scheme). For
`ssh`/`github`/`platform-repo`/`cyso`/`tfstate` (both schemes), the script deletes the path and re-populates it —
via `load-credentials.sh` for the legacy scheme, or directly for the new scheme, since the new
scheme's `load-credentials.sh` never writes `ssh` itself. `base-ssh` is different: rotating it
never deletes-and-repopulates. It generates a new base keypair, uses the *old* one one last time to
install the new public key on the machine, then retires the old one — because it's the one
credential type that has to install itself onto the remote machine, not just be replaced in Vault.
Rotating a machine's `base-ssh` never touches any already-bootstrapped app's `secret/apps/<app-slug>
/ssh`. `apps/<app-slug>/<machine-slug>/kubeconfig` is different again — and different from every
other credential here: the platform's own credential rotation is destructive (a known platform
gap, not fixed by this skill), so this script never generates a replacement locally. The
machine-slug in the path is already the distinguishing dimension — there is no separate
test/prod argument anymore, since each machine-slug's kubeconfig is its own independent path with
no sibling field to preserve. It deletes the value at that one path and re-prompts for a
**freshly platform-reissued** value, pasted the same out-of-band way `load-credentials.sh`
originally loaded it — never a cluster call, never regenerated. After any rotation, re-run
`smaqit.infrastructure-repo-config` to sync the new value to GitHub Secrets (or the relevant
GitHub Environment secret, for `kubeconfig`).

---

## Output

- `VAULT_ADDR=http://127.0.0.1:8200` set in shell environment
- Vault running, unsealed, authenticated
- All required paths verified populated — `secret/apps/<app-slug>/*` +
  `secret/machines/<machine-slug>/*` for new-scheme projects, `secret/<project-slug>/*` for legacy
  ones, or `secret/apps/<app-slug>/{github,platform-repo}` plus
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig` for each machine-slug this app has been
  given, for `existing-k3s` projects (no `machines/` namespace at all — see the path convention
  above)
- Calling skill can now read credentials without human input

## Scope

- Does NOT manage Vault HA, replication, or namespaces — single-node local use only
- Does NOT rotate credentials automatically — use `scripts/rotate-credential.sh`
- Does NOT apply Vault policies — root token is used for local dev
- Does NOT start Vault as a systemd service — Vault is session-scoped and must be explicitly
  started and unsealed per session

## Gotchas

- **`tls_disable = true` is safe for localhost only** — never bind Vault to a non-loopback
  address with TLS disabled
- **`-key-shares=1 -key-threshold=1`** — single unseal key for simplicity; acceptable for a
  local dev vault
- **Background process** — `vault server ... &` dies when the terminal closes; this is
  intentional — Vault is not a persistent service on this machine
- **`VAULT_ADDR` must be exported** — every `vault` CLI call in subshells needs this variable;
  export it at the start of each session or add to `.bashrc`
- **`~/.vault-token` is read automatically** — if a prior `vault login` on this machine wrote a
  token there, `load-credentials.sh` reuses it instead of prompting. Harmless no-op if the file
  is absent or the token has expired (falls through to the normal prompt).
- **SSH private key trailing newline / "error in libcrypto"** — `vault kv put private_key="$(cat file)"`
  strips the private key's trailing newline via shell command substitution. OpenSSH's key parser
  requires it; without it, every subsequent `ssh`/`ssh-keygen -y` against the fetched key fails with
  `error in libcrypto` (not an auth or permissions issue — the key content itself is malformed).
  `load-credentials.sh` writes `private_key` using Vault's `@file` syntax (`private_key=@"$SSH_KEY_PATH"`),
  which preserves the file's exact bytes including the trailing newline. Never rewrite this back to
  `"$(cat ...)"` — that regresses the bug silently (re-fetching, patching, and writing back via
  command substitution reintroduces it even after a one-time manual fix). If you ever hand-load an
  SSH key manually (bypassing the script), use `@`-file syntax, not `$(cat ...)`, for `private_key`.
- **`kubeconfig` is out-of-band paste only, never a cluster call.** Unlike every other credential
  this skill manages, `secret/apps/<app-slug>/<machine-slug>/kubeconfig` is never generated,
  derived, or validated by contacting a cluster — `load-credentials.sh`'s `existing-k3s` branch
  only stores whatever is pasted in, verbatim, as that path's single `value` field. If a paste is
  truncated or malformed, this skill will not detect it; the first sign of trouble is
  `smaqit.infrastructure-deploy-k3s-app`'s `namespace-guard.sh` failing to authenticate.
- **`kubeconfig` rotation re-prompts; it never regenerates.** The platform's own kubeconfig
  rotation is destructive (a known platform gap) — `rotate-credential.sh
  apps/<app-slug>/<machine-slug>/kubeconfig` deletes the stored value and re-prompts for a
  freshly platform-reissued value, unlike `ssh`'s local keypair regeneration. Do not "fix" this
  by generating a keypair or token locally; there is nothing this skill can generate that the
  platform would accept.

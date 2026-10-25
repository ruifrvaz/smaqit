---
name: smaqit.infrastructure-vault-loader
description: 'Use before any local deployment or credential operation that needs secrets from the local HashiCorp Vault (127.0.0.1:8200): verifies it is running, unsealed, and authenticated, and loads project secrets through an interactive prompt. Also use for first-time Vault setup on a machine, adding a project''s credentials, a missing Vault path, or a caller that cannot reach Vault.'
metadata:
  version: "5.0.0"
  smaqit-slots: "access-credentials=access, target-credentials=target, target-credentials=register"
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
| `scripts/setup-vault.sh` | Create config, start server, initialise (prints the unseal key and root token). Run once per machine. |
| `scripts/load-credentials.sh` | Start + unseal + authenticate + enable kv-v2 if needed + load all credential paths. Run every session. |
| `scripts/register-machine.sh <machine-slug> [--host <ip>]` | Populate a machine's `base-ssh`, `cyso`, `tfstate`, and `metadata` the first time; `--host` backfills the address after apply. Idempotent. |
| `scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>` | Give an app its own distinct SSH keypair, authorized on the named machine. Run once per app-to-machine pairing. |
| `scripts/rotate-credential.sh <path>` | Rotation only: delete and re-populate a single existing credential path (e.g. `cyso`, `apps/<app-slug>/ssh`, `machines/<machine-slug>/base-ssh`). |
| `scripts/credential-file.sh <path>` | Write one credential a tool reads from a file — an app's SSH key — to a 0600 file in a private temp directory and print only the path. Allowlisted paths only. |
| `source scripts/terraform-env.sh [--unset]` | Export Terraform's provider and backend credentials (`TF_VAR_*`, `AWS_*`) into the current shell without printing them; `--unset` clears them. |

**Skills and operators never run `vault` directly** — every Vault read or write goes through one
of these scripts (which may call `vault` internally). If an operation you need has no script,
stop and add one rather than typing a raw `vault` command.

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
secret/apps/<app-slug>/<machine-slug>/runtime — one field per runtime secret name
                                     (provisioning_mode: existing-k3s only)
```

`cyso` and `tfstate` are machine-scoped: provisioning a VM is a property of the machine, not of
any individual app deployed onto it. `ssh` and `github` are app-scoped: every app always gets its
own distinct keypair, bootstrapped against its machine's `base-ssh` credential (see "Bootstrapping
an app onto a machine" below) — never copied from another app and never installed by Terraform
directly. `<app-slug>` and `<machine-slug>` are lowercase hyphenated slugs; for an app, this is the
project name declared in `AGENTS.md` (or the legacy platform-specific `CLAUDE.md` /
`copilot-instructions.md`) — same derivation as the legacy `<project-slug>` below, renamed to match
where it now lives in Vault. `machines` is a reserved app-slug.

An `existing-k3s` app stores **no deploy credential** here: its CI deploy job authenticates with its
own GitHub Actions OIDC identity through the platform's connect action, the platform holds the
mesh key the job joins a mesh-only API with, and the platform pulls the app's images. The app keeps
nothing under `secret/machines/<machine-slug>/*` either — a Namespace-scoped cluster is a platform
the app is onboarded onto, not a host it registers (the platform repository that owns the machine
may register it with `register-machine.sh`).

The onboarding request needs no stored credential either:
`smaqit.infrastructure-request-k3s-onboarding` opens its PR with the operator's own `gh` login.
What an `existing-k3s` app does store is its **runtime secrets** — values its pod needs at run time,
such as a session-signing key or a first-admin password — declared in the Infrastructure spec's
`## Runtime Secrets` table and kept per environment at
`secret/apps/<app-slug>/<machine-slug>/runtime`, one field per secret name. A `generated` secret is
created once by `load-credentials.sh` and never printed; an `operator-provided` one is read from a
hidden prompt. `smaqit.infrastructure-repo-config`'s `runtime` operation syncs them into the
machine-slug GitHub Environment, and the deploy workflow builds the app's Kubernetes Secret from
there. Under the spec's `Vault Access: operator-scripts` row, neither script calls Vault: the
operator stores and syncs the values with their own scripts, following the handoff the
`environment-config` slot prints.

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
server, and initialises it.

The init output contains the **Unseal Key** and **Root Token**. Store both offline immediately
(printed paper or password manager). Loss of the unseal key means permanent data loss.

After init completes, run `scripts/load-credentials.sh`: it unseals, authenticates, and on its
first run enables the kv-v2 secrets engine at `secret/` (an engine can't be enabled before the
first unseal, which is why `setup-vault.sh` leaves it to this step).

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

- If `PROVISIONING_MODE=existing-k3s` is passed explicitly (with `ENVIRONMENT=<env>`, the spec's
  Environments row), it runs in **existing-k3s mode**. It reads the spec first, with
  `smaqit spec environments --env=<env> --json`, and makes no Vault call at all when the environment
  has no runtime secrets or the spec declares `Vault Access: operator-scripts`. Otherwise it
  checks/populates only `secret/apps/<app-slug>/<machine-slug>/runtime`, one field per secret
  (idempotent: a field already present is kept). `ssh`, `github`, `cyso`, `tfstate`, and `machine`
  are never touched, and nothing is stored under `secret/machines/<machine-slug>/*`.
- If `secret/apps/<app-slug>/machine` already exists (this app has bootstrapped before), or a
  `MACHINE_SLUG` is passed explicitly, it runs in **new-scheme mode**: only `secret/apps/<app-slug>/
  github` is checked/populated. `ssh` is never touched here — it's exclusively the job of
  `bootstrap-app-to-machine.sh`, since populating it requires the machine's `base-ssh` credential to
  authorize the new key. `cyso`/`tfstate` are never prompted for — those live at
  `secret/machines/<machine-slug>/*`, populated once at machine registration
  (`register-machine.sh`, see "Registering a machine"), not per app per session.
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

# Namespace-scoped Kubernetes/k3s target — only the environment's runtime secrets, if any:
PROVISIONING_MODE=existing-k3s ENVIRONMENT=<env> bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

### Slot operations

This skill fills two deployment slots (`metadata.smaqit-slots`). The deployment agent runs them at
different points of one environment run:

- **`access-credentials=access`** — credentials needed before the target can be reached or
  requested. `existing-k3s`: `PROVISIONING_MODE=existing-k3s ENVIRONMENT=<env>`, which loads only
  the environment's runtime secrets and completes without touching Vault when there are none or
  under `Vault Access: operator-scripts`. Every other mode: `load-credentials.sh` exactly as
  described above for that mode and scheme.
- **`target-credentials=target`** — credentials the target issues or authorizes for this app, after
  onboarding or registration. Not part of the `existing-k3s` sequence: the target issues the app
  nothing (CI connects with its own OIDC identity). `existing-shared`/`existing-unmanaged` on the `apps/`+`machines/`
  scheme: `bootstrap-app-to-machine.sh <app-slug> <machine-slug>` (below); on the legacy flat scheme,
  nothing further — the access load already populated `ssh`. Not part of the `provision` or
  `existing-owned` default sequence.
- **`target-credentials=register`** — machine registration for a repository that owns machines
  (a platform repository): `register-machine.sh <machine-slug>` (below). Never a default; a
  project selects it by binding `target-credentials: {skill: smaqit.infrastructure-vault-loader,
  operation: register}` in `.smaqit/slots.yml`. Because this skill declares two operations for
  `target-credentials`, a binding must name the operation.

---

## Registering a machine

The single entrypoint for populating a machine's own credentials the first time — whether the
machine is provisioned by an app project's Terraform (`smaqit.infrastructure-provision-cyso`) or
by a platform repository that owns machines:

```
# Before provisioning — generates base-ssh, prompts for cyso and tfstate, writes metadata:
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/register-machine.sh <machine-slug>

# After apply — backfill the machine's address:
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/register-machine.sh <machine-slug> --host <fixed_ip>
```

- **The machine slug is a required argument.** It is never derived from the working directory, so
  running the script from an unrelated directory can't write to the wrong path. It must be
  lowercase letters, digits, and inner hyphens, because it also names the machine's GitHub
  Environment (`smaqit.infrastructure-repo-config` `sync-machine`).
- **Idempotent.** Every path that already exists is kept. Re-running fills only missing paths;
  `--host`, `--provider`, and `--owner-project` patch those `metadata` fields in place.
- `base-ssh` is generated only when missing. It is bootstrap-only: Terraform's keypair resource
  installs its public half at apply time; for a machine no Terraform manages, install it in
  `authorized_keys` yourself.
- `cyso` (`app_credential_id`, `app_credential_secret`) and `tfstate` (`access_key`, `secret_key`)
  are prompted for only when missing. Secrets are read hidden from the terminal and passed to Vault
  on stdin; an empty answer stops the script instead of writing a placeholder.
- `metadata` holds `provider` (default `cyso`), `owner_project` (the repository whose Terraform
  state owns the machine; defaults to this repository's slug), and `host`. **`host` is required by
  `rotate-credential.sh machines/<slug>/base-ssh`** — backfill it with `--host` right after apply.

`rotate-credential.sh` is rotation-only: it replaces paths that already exist and is never the way
to populate a machine for the first time.

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
   script generates and stores the base keypair and prompts for `metadata` (host, provider, owner
   project), then stops so the base key can be installed on the host. It does **not** prompt for
   `cyso` or `tfstate`; a machine that needs them is registered with `register-machine.sh` first. This is the branch `provisioning_mode: existing-unmanaged` always exercises —
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
`apps/<app-slug>/github`, `machines/<machine-slug>/{base-ssh,cyso,tfstate}` (new scheme). Rotation replaces paths that already
exist; first-time population of a machine is `register-machine.sh`. Rotating
`machines/<machine-slug>/base-ssh` requires `metadata.host` (backfill it with
`register-machine.sh <machine-slug> --host <ip>`). For
`ssh`/`github`/`cyso`/`tfstate` (both schemes), the script deletes the path and re-populates it —
via `load-credentials.sh` for the legacy scheme, or directly for the new scheme, since the new
scheme's `load-credentials.sh` never writes `ssh` itself. `base-ssh` is different: rotating it
never deletes-and-repopulates. It generates a new base keypair, uses the *old* one one last time to
install the new public key on the machine, then retires the old one — because it's the one
credential type that has to install itself onto the remote machine, not just be replaced in Vault.
Rotating a machine's `base-ssh` never touches any already-bootstrapped app's `secret/apps/<app-slug>
/ssh`. After rotating a credential a workflow reads, re-run `smaqit.infrastructure-repo-config`
to sync the new value to GitHub Secrets, then redeploy with a fresh workflow dispatch — a re-run of
an earlier run keeps the secret values from when that run was created.

---

## Output

- `VAULT_ADDR=http://127.0.0.1:8200` set in shell environment
- Vault running, unsealed, authenticated
- All required paths verified populated — `secret/apps/<app-slug>/*` +
  `secret/machines/<machine-slug>/*` for new-scheme projects, `secret/<project-slug>/*` for legacy
  ones, or only the environment's `secret/apps/<app-slug>/<machine-slug>/runtime` fields for
  `existing-k3s` projects (no deploy credential and no `machines/` namespace — see the path
  convention above); nothing at all for an `existing-k3s` environment with no runtime secrets
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
- **Background process** — the Vault server `load-credentials.sh` starts dies when the terminal
  closes; this is intentional — Vault is not a persistent service on this machine
- **`VAULT_ADDR`** — every script defaults it to `http://127.0.0.1:8200`; export it only to point
  the scripts at a different address
- **`~/.vault-token` is read automatically** — if a prior login on this machine wrote a
  token there, `load-credentials.sh` reuses it instead of prompting. Harmless no-op if the file
  is absent or the token has expired (falls through to the normal prompt).
- **SSH private key trailing newline / "error in libcrypto"** — storing a key as
  `private_key="$(cat file)"` strips its trailing newline via shell command substitution. OpenSSH's key parser
  requires it; without it, every subsequent `ssh`/`ssh-keygen -y` against the fetched key fails with
  `error in libcrypto` (not an auth or permissions issue — the key content itself is malformed).
  `load-credentials.sh` writes `private_key` using Vault's `@file` syntax (`private_key=@"$SSH_KEY_PATH"`),
  which preserves the file's exact bytes including the trailing newline. Never rewrite this back to
  `"$(cat ...)"` — that regresses the bug silently (re-fetching, patching, and writing back via
  command substitution reintroduces it even after a one-time manual fix). Keys are only ever
  stored through the scripts, which all use `@`-file syntax.
- **An `existing-k3s` app stores no deploy credential.** There is no kubeconfig, mesh key, or
  registry credential to load or rotate: the CI job's own OIDC identity is the credential, the
  platform holds the mesh key, and the platform pulls the images. If a deploy fails to
  authenticate, the fix is the platform's binding or the spec's `Cluster Connect` pin — never a
  credential pasted into Vault. Runtime secrets are app configuration, not deploy credentials: they
  reach the pod, never the deploy job's connection to the cluster.
- **Releases before the operator-`gh` onboarding stored `secret/apps/<app-slug>/platform-repo`.**
  Nothing reads it now; retire it with your own Vault tooling.

---
name: smaqit.infrastructure-vault-loader
description: Use before any local deployment or credential operation that requires secrets from a local HashiCorp Vault instance. Verifies Vault is running, unsealed, and authenticated on 127.0.0.1:8200. Also runs an interactive credential loader script that prompts for all project secrets and writes them to Vault. Use for first-time setup, adding a new project's credentials, or when a Vault path is missing. Also use when setting up Vault for the first time on a new machine, or when a caller cannot reach Vault and needs troubleshooting guidance.
metadata:
  version: "3.3.0"
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
```

`cyso` and `tfstate` are machine-scoped: provisioning a VM is a property of the machine, not of
any individual app deployed onto it. `ssh` and `github` are app-scoped: every app always gets its
own distinct keypair, bootstrapped against its machine's `base-ssh` credential (see "Bootstrapping
an app onto a machine" below) — never copied from another app and never installed by Terraform
directly. `<app-slug>` and `<machine-slug>` are lowercase hyphenated slugs; for an app, this is the
project name declared in `AGENTS.md` (or the legacy platform-specific `CLAUDE.md` /
`copilot-instructions.md`) — same derivation as the legacy `<project-slug>` below, renamed to match
where it now lives in Vault. `machines` is a reserved app-slug.

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

### Scheme detection: legacy flat vs. `apps/`+`machines/`

`load-credentials.sh` decides which scheme applies per invocation:

- If `secret/apps/<app-slug>/machine` already exists (this app has bootstrapped before), or a
  `MACHINE_SLUG` is passed explicitly, it runs in **new-scheme mode**: only `secret/apps/<app-slug>/
  github` is checked/populated. `ssh` is never touched here — it's exclusively the job of
  `bootstrap-app-to-machine.sh`, since populating it requires the machine's `base-ssh` credential to
  authorize the new key. `cyso`/`tfstate` are never prompted for — those live at
  `secret/machines/<machine-slug>/*`, populated once at machine registration
  (`smaqit.infrastructure-provision-cyso`'s post-apply step), not per app per session.
- Otherwise, it falls back to **legacy flat-scheme mode**, unchanged: all four paths under
  `secret/<project-slug>/{cyso,ssh,tfstate,github}` are checked/populated exactly as before,
  including the `PROVISIONING_MODE`-aware `existing-shared` behavior (cross-namespace SSH copy or
  manual `authorized_keys` append). This is what keeps unmigrated projects on the legacy
  flat scheme working without any change.

```
# New-scheme app, already bootstrapped — just tops up github if missing:
bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

# Legacy flat-scheme project — unchanged:
PROVISIONING_MODE=existing-shared bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
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
   before continuing.
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
`apps/<app-slug>/github`, `machines/<machine-slug>/{base-ssh,cyso,tfstate}` (new scheme). For
`ssh`/`github`/`cyso`/`tfstate` (both schemes), the script deletes the path and re-populates it —
via `load-credentials.sh` for the legacy scheme, or directly for the new scheme, since the new
scheme's `load-credentials.sh` never writes `ssh` itself. `base-ssh` is different: rotating it
never deletes-and-repopulates. It generates a new base keypair, uses the *old* one one last time to
install the new public key on the machine, then retires the old one — because it's the one
credential type that has to install itself onto the remote machine, not just be replaced in Vault.
Rotating a machine's `base-ssh` never touches any already-bootstrapped app's `secret/apps/<app-slug>
/ssh`. After any rotation, re-run `smaqit.infrastructure-repo-config` to sync the new value to
GitHub Secrets.

---

## Output

- `VAULT_ADDR=http://127.0.0.1:8200` set in shell environment
- Vault running, unsealed, authenticated
- All required paths verified populated — `secret/apps/<app-slug>/*` +
  `secret/machines/<machine-slug>/*` for new-scheme projects, `secret/<project-slug>/*` for legacy
  ones
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

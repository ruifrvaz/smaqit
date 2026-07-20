---
name: smaqit.infrastructure-vault-loader
description: Use before any local deployment or credential operation that requires secrets from a local HashiCorp Vault instance. Verifies Vault is running, unsealed, and authenticated on 127.0.0.1:8200. Also runs an interactive credential loader script that prompts for all project secrets and writes them to Vault. Use for first-time setup, adding a new project's credentials, or when a Vault path is missing. Also use when setting up Vault for the first time on a new machine, or when a caller cannot reach Vault and needs troubleshooting guidance.
metadata:
  version: "3.1.0"
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
| `scripts/rotate-credential.sh <path>` | Delete and re-populate a single credential path (e.g. `cyso`). |

Config template used by `setup-vault.sh`: `assets/vault.hcl.template`

---

## Vault path convention

All smaqit skills read from and write to paths under the project slug:

```
secret/<project-slug>/cyso      — app_credential_id, app_credential_secret
secret/<project-slug>/ssh       — private_key, public_key
secret/<project-slug>/tfstate   — access_key, secret_key
secret/<project-slug>/github    — token (used as TF_VAR_github_token)
```

`<project-slug>` is the lowercase hyphenated project name declared in `CLAUDE.md` (or `copilot-instructions.md`).

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

### Provisioning-mode-aware loading

Pass `PROVISIONING_MODE` (defaults to `provision`, resolved upstream by `smaqit.input-deployment`
for a given deploy) to control which paths the script populates:

```
PROVISIONING_MODE=existing-shared bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

- **`provision` / `existing-owned`** — unchanged: all four paths (`cyso`, `ssh`, `tfstate`,
  `github`) are checked/populated.
- **`existing-shared`** — targeting a VM a different project owns. `cyso` and `tfstate` are
  never prompted for; this project has no Terraform state to provision or back with remote
  state. Only `ssh` and `github` are populated. For `ssh`, the script offers two mechanisms:
  1. **Copy from another project's Vault namespace** — prompts for the source project slug and
     copies `secret/<source-slug>/ssh` into `secret/<this-slug>/ssh` verbatim. Requires the
     operator to already have Vault access to the owning project's namespace.
  2. **Generate a new keypair** — same as the default path, but the script cannot append the
     public key to the shared VM's `authorized_keys` itself (it has no access to that VM); it
     prints the public key and the operator appends it manually, once.

  Neither mechanism is "the" automated path — both require a one-time manual trust step between
  two otherwise-independent projects, since Vault namespaces are not automatically shared.

---

## Rotating a credential

Run `scripts/rotate-credential.sh <path>` where `<path>` is one of `cyso`, `ssh`, `tfstate`,
or `github`. The script deletes the named path, re-runs `load-credentials.sh` to repopulate it,
then reminds you to sync the new value to GitHub Secrets via `smaqit.infrastructure-repo-config`.

---

## Output

- `VAULT_ADDR=http://127.0.0.1:8200` set in shell environment
- Vault running, unsealed, authenticated
- All `secret/<project-slug>/*` paths verified populated
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


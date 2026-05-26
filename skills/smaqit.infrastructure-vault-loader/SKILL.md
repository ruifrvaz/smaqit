---
name: smaqit.infrastructure-vault-loader
description: Use before any local deployment or credential operation that requires secrets from a local HashiCorp Vault instance. Verifies Vault is running, unsealed, and authenticated on 127.0.0.1:8200. Also runs an interactive credential loader script that prompts for all project secrets and writes them to Vault. Use for first-time setup, adding a new project's credentials, or when a Vault path is missing. Also use when setting up Vault for the first time on a new machine, or when a caller cannot reach Vault and needs troubleshooting guidance.
metadata:
  version: "2.0.0"
---

# Vault Loader

Ensures a local HashiCorp Vault instance is running, unsealed, and ready to serve credentials to
local deployment automation. Provides a deterministic interactive script that prompts for all
project credentials and writes them to Vault. This skill is a pre-step for `smaqit.infrastructure-provision-cyso`,
`smaqit.infrastructure-deploy-rsync`, and `smaqit.infrastructure-repo-config` when run locally.

## Vault path convention

All smaqit skills read from and write to paths under the project slug:

```
secret/<project-slug>/cyso      — app_credential_id, app_credential_secret
secret/<project-slug>/ssh       — private_key, public_key
secret/<project-slug>/tfstate   — access_key, secret_key
secret/<project-slug>/github    — token (used as TF_VAR_github_token)
```

`<project-slug>` is the lowercase hyphenated project name declared in `copilot-instructions.md`
(e.g. `hello-mario`).

---

## One-time setup (first run only)

### 1. Install Vault

```bash
# Ubuntu / Debian
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
  | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install -y vault
```

Verify: `vault version`

### 2. Create config file

```bash
mkdir -p ~/.vault/data
cat > ~/.vault/config.hcl << 'EOF'
storage "file" {
  path = "/home/<your-username>/.vault/data"
}

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = true
}

ui = false
EOF
```

Replace `<your-username>` with your actual username.

### 3. Start Vault and initialise

```bash
# Start the server (keep this terminal open or run in background)
vault server -config=~/.vault/config.hcl &

export VAULT_ADDR=http://127.0.0.1:8200

# Initialise — generates unseal keys and root token
# Store the output in a secure physical location (printed paper, password manager)
vault operator init -key-shares=1 -key-threshold=1
```

The output contains:
- `Unseal Key 1: <key>` — required every time Vault starts
- `Initial Root Token: <token>` — used to authenticate; rotate after initial setup

**Store both values securely offline. Loss of the unseal key = permanent data loss.**

### 4. Unseal and authenticate

```bash
vault operator unseal <unseal-key>
vault login <root-token>

# Enable kv-v2 secrets engine
vault secrets enable -path=secret kv-v2
```

### 5. Load project credentials

Run the interactive loader script — it prompts for each credential, skips paths that already exist, and generates the SSH keypair automatically:

```bash
bash .github/skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

Or load manually:

```bash
export PROJECT_SLUG=hello-mario   # from copilot-instructions.md

# Cyso Cloud app credential
read -p "app_credential_id: " CYSO_ID
read -s -p "app_credential_secret: " CYSO_SECRET && echo
vault kv put secret/${PROJECT_SLUG}/cyso \
  app_credential_id="$CYSO_ID" \
  app_credential_secret="$CYSO_SECRET"
unset CYSO_ID CYSO_SECRET

# SSH keypair — generated automatically (no prompt needed)
SSH_KEY_PATH=$(mktemp -d)/deploy_key
ssh-keygen -t ed25519 -f "$SSH_KEY_PATH" -N "" -q
vault kv put secret/${PROJECT_SLUG}/ssh \
  private_key="$(cat ${SSH_KEY_PATH})" \
  public_key="$(cat ${SSH_KEY_PATH}.pub | tr -d '\n')"
rm -rf "$(dirname $SSH_KEY_PATH)"

# Terraform state bucket S3 credentials
read -p "s3_access_key: " S3_KEY
read -s -p "s3_secret_key: " S3_SECRET && echo
vault kv put secret/${PROJECT_SLUG}/tfstate \
  access_key="$S3_KEY" \
  secret_key="$S3_SECRET"
unset S3_KEY S3_SECRET

# GitHub token (fine-grained PAT with variables:write)
read -s -p "github_token: " GH_TOKEN && echo
vault kv put secret/${PROJECT_SLUG}/github \
  token="$GH_TOKEN"
unset GH_TOKEN
```

Verify all paths are populated:
```bash
vault kv list secret/${PROJECT_SLUG}
```

---

## Interactive credential loader script

The script at `scripts/load-credentials.sh` is the recommended way to populate all credential paths for a new project. It:
- Checks Vault is running and unsealed before starting
- Reads the project slug from `copilot-instructions.md` automatically
- Prompts for each credential that is **not already present** in Vault (skips existing paths)
- Generates the SSH deploy keypair automatically — no prompt needed
- Clears all sensitive variables from memory after each write
- Verifies all paths at the end

Run it:
```bash
bash .github/skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

---

## Steps (every session)

1. **Check if Vault is running:**
   ```bash
   export VAULT_ADDR=http://127.0.0.1:8200
   vault status 2>/dev/null | grep -q "Sealed" || echo "Vault not running"
   ```
   If not running:
   ```bash
   vault server -config=~/.vault/config.hcl &
   sleep 1
   ```

2. **Check seal status:**
   ```bash
   vault status | grep "Sealed"
   ```
   If `Sealed: true`:
   ```bash
   vault operator unseal <unseal-key>
   ```

3. **Authenticate** (if session token expired):
   ```bash
   vault token lookup 2>/dev/null || vault login
   ```
   `vault login` prompts for the root token (or a scoped token if you have created one).

4. **Verify target paths exist:**
   ```bash
   vault kv get secret/<project-slug>/cyso > /dev/null
   vault kv get secret/<project-slug>/ssh > /dev/null
   vault kv get secret/<project-slug>/tfstate > /dev/null
   vault kv get secret/<project-slug>/github > /dev/null
   ```
   If any path is missing, run the one-time setup Step 5 for that path only.

5. **Confirm ready:**
   ```bash
   vault status | grep -E "Sealed|Version"
   ```
   Expected: `Sealed: false`. Proceed to the calling skill.

## Output

- `VAULT_ADDR=http://127.0.0.1:8200` set in shell environment
- Vault running, unsealed, authenticated
- All `secret/<project-slug>/*` paths verified populated
- Calling skill can now `vault kv get` credentials without human input

## Scope

- Does NOT manage Vault HA, replication, or namespaces — single-node local use only
- Does NOT rotate credentials — call `vault kv put` with new values and re-run
  `smaqit.infrastructure-repo-config` to sync GitHub Secrets
- Does NOT apply Vault policies — root token is used for local dev; scope policies separately
  if daisy-hub runner use is added later
- Does NOT start Vault as a systemd service — intentional; Vault is session-scoped and must be
  explicitly started and unsealed per session

## Gotchas

- **`tls_disable = true` is safe for localhost only** — never bind Vault to a non-loopback
  address with TLS disabled
- **`-key-shares=1 -key-threshold=1`** — single unseal key for simplicity. Acceptable for a
  local dev vault; increase shares for any shared or long-lived instance
- **Root token rotation** — after initial setup, create a scoped token with only `read` on
  `secret/<project-slug>/*` and use that for daily operations; store root token offline only
- **Background process** — `vault server ... &` means the process dies when the terminal closes.
  This is intentional; Vault is not a persistent service on this machine
- **`VAULT_ADDR` must be exported** — every `vault` CLI call in subshells needs this variable;
  export it in `.bashrc` or set it at the start of each session

## Credential rotation procedure

When rotating any credential (e.g. new Cyso app credential after expiry):

```bash
# 1. Update Vault
vault kv put secret/<project-slug>/cyso \
  app_credential_id=<new-id> \
  app_credential_secret=<new-secret>

# 2. Sync to GitHub Secrets
smaqit.infrastructure-repo-config   # re-runs gh secret set from Vault values

# 3. Verify CI/CD still passes
gh run watch
```

One command updates Vault; one skill sync propagates to GitHub. No manual GitHub portal steps.

To re-run the loader for a single path only, delete it first then re-run the script:
```bash
vault kv delete secret/<project-slug>/cyso
bash .github/skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
```

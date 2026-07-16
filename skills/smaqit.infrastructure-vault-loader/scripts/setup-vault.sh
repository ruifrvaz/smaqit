#!/usr/bin/env bash
# One-time Vault initialisation on a new machine.
# Creates ~/.vault/config.hcl from the template, starts the server,
# runs `vault operator init`, and enables the kv-v2 secrets engine.
#
# Run ONCE per machine. Do not re-run on an already-initialised Vault —
# it will refuse to re-initialise and exit cleanly.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/setup-vault.sh

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_TEMPLATE="${SKILL_DIR}/assets/vault.hcl.template"
CONFIG_DEST="${HOME}/.vault/config.hcl"
DATA_DIR="${HOME}/.vault/data"

# ── 1. Write config file ──────────────────────────────────────────────────────

if [ -f "$CONFIG_DEST" ]; then
  echo "==> Config already exists at $CONFIG_DEST — skipping creation"
else
  echo "==> Creating Vault config at $CONFIG_DEST..."
  mkdir -p "$DATA_DIR"
  sed "s|{{DATA_DIR}}|${DATA_DIR}|g" "$CONFIG_TEMPLATE" > "$CONFIG_DEST"
  echo "    Done"
fi

# ── 2. Start server ───────────────────────────────────────────────────────────

if vault status > /dev/null 2>&1; then
  echo "==> Vault already running at $VAULT_ADDR"
else
  echo "==> Starting Vault server..."
  vault server -config="$CONFIG_DEST" > /tmp/vault-server.log 2>&1 &
  sleep 2
  if ! vault status > /dev/null 2>&1; then
    echo "ERROR: Vault failed to start. Check /tmp/vault-server.log"
    exit 1
  fi
  echo "    Started (PID $!)"
fi

# ── 3. Initialise (only if not already initialised) ───────────────────────────

INIT_STATUS=$(vault status -format=json 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['initialized'])" 2>/dev/null || echo "false")
if [ "$INIT_STATUS" = "True" ] || [ "$INIT_STATUS" = "true" ]; then
  echo "==> Vault already initialised — skipping init"
  echo "    Run load-credentials.sh to unseal and authenticate."
  exit 0
fi

echo ""
echo "==> Initialising Vault (1 key share, threshold 1)..."
echo "    The output below contains your UNSEAL KEY and ROOT TOKEN."
echo "    Store both offline immediately. Loss = permanent data loss."
echo ""
vault operator init -key-shares=1 -key-threshold=1

# ── 4. Enable kv-v2 ──────────────────────────────────────────────────────────

echo ""
echo "==> After unsealing and logging in via load-credentials.sh, enable kv-v2 once:"
echo "    vault secrets enable -path=secret kv-v2"
echo ""
echo "==> Setup complete. Next step:"
echo "    bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh"

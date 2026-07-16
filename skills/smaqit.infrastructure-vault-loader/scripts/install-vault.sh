#!/usr/bin/env bash
# Install HashiCorp Vault on Ubuntu / Debian.
# Run once on the machine where local deployments will be executed.
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/install-vault.sh

set -euo pipefail

if command -v vault &>/dev/null; then
  echo "Vault already installed: $(vault version)"
  exit 0
fi

echo "==> Installing Vault..."
wget -O- https://apt.releases.hashicorp.com/gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
  | sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt-get update -q
sudo apt-get install -y vault

echo "==> Installed: $(vault version)"

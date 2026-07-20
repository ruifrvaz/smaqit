#!/usr/bin/env bash
# Idempotent guardrail for smaqit.infrastructure-vm-bootstrap.
#
# Removes the stock distro `default` nginx site from sites-enabled if present, then verifies
# `nginx -t` passes. Two sites both declaring `default_server` on the same port fails nginx
# config validation, leaving the previous (or stock) config silently active after a reload.
#
# Safe to re-run on every deploy: a no-op if the default site is already removed.
# Usage: remove-default-nginx-site.sh (run on the target VM, requires sudo)

set -euo pipefail

DEFAULT_SITE="/etc/nginx/sites-enabled/default"

if [ -e "$DEFAULT_SITE" ]; then
  echo "remove-default-nginx-site: removing stock default site ($DEFAULT_SITE)"
  sudo rm -f "$DEFAULT_SITE"
else
  echo "remove-default-nginx-site: already absent, nothing to do"
fi

sudo nginx -t

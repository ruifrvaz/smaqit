#!/usr/bin/env bash
# Deterministic default_server-vs-name-based vhost writer for smaqit.infrastructure-deploy-rsync.
#
# Inspects /etc/nginx/sites-enabled/ on the target VM itself and decides, rather than an agent
# guessing from prose: the first site on a VM claims `default_server`; every subsequent
# co-hosted site's vhost must be name-based only. Two vhosts both claiming `default_server` on
# the same port makes `nginx -t` fail outright — this is the highest-stakes judgment call in the
# deploy-rsync flow, so it is not left to be re-derived correctly on every deploy.
#
# Usage: write-vhost.sh <ssh-key-path> <host> <local-conf-path> <remote-site-name> [server-name]
#   <local-conf-path>   Local nginx conf to upload (its `listen`/`server_name` lines are rewritten
#                        in a temp copy before upload — the local file on disk is never modified).
#   <remote-site-name>  Filename to use under /etc/nginx/sites-available/ and sites-enabled/
#                        (e.g. the project slug).
#   [server-name]       Required when this VM already serves another site (co-hosted); used as
#                        the vhost's `server_name` directive. Omit for a first-site deploy.
#
#   Exit 0 — conf uploaded, symlinked into sites-enabled (if not already), `nginx -t` passed.
#   Exit 1 — usage error, SSH failure, co-hosted but no server-name supplied, or `nginx -t` failed
#            (previous config left active; nothing reloaded — see smaqit.infrastructure-deploy-rsync).

set -euo pipefail

SSH_KEY="${1:?Usage: write-vhost.sh <ssh-key-path> <host> <local-conf-path> <remote-site-name> [server-name]}"
HOST="${2:?Usage: write-vhost.sh <ssh-key-path> <host> <local-conf-path> <remote-site-name> [server-name]}"
LOCAL_CONF="${3:?Usage: write-vhost.sh <ssh-key-path> <host> <local-conf-path> <remote-site-name> [server-name]}"
REMOTE_SITE="${4:?Usage: write-vhost.sh <ssh-key-path> <host> <local-conf-path> <remote-site-name> [server-name]}"
SERVER_NAME="${5:-}"

[ -f "$LOCAL_CONF" ] || { echo "write-vhost: local conf not found: $LOCAL_CONF" >&2; exit 1; }

SSH() { ssh -i "$SSH_KEY" "ubuntu@$HOST" "$@"; }

# ── Determine whether this is the first site on the VM ────────────────────────
# Exclude this site's own symlink (if redeploying) from the "already enabled" check —
# a redeploy of the first/only site must not be misread as a co-hosted deploy.

ENABLED="$(SSH "ls /etc/nginx/sites-enabled/ 2>/dev/null | grep -v '^${REMOTE_SITE}\$' || true")"

TMPCONF="$(mktemp)"
trap 'rm -f "$TMPCONF"' EXIT
cp "$LOCAL_CONF" "$TMPCONF"

if [ -z "$ENABLED" ]; then
  echo "write-vhost: no other site enabled — first site on this VM, using default_server"
  # Ensure `listen 80 default_server;` (add the flag if the conf has a bare `listen 80;`).
  sed -i -E 's/^([[:space:]]*listen[[:space:]]+80;)[[:space:]]*$/\1 default_server;/' "$TMPCONF"
  sed -i -E 's/listen 80;\s+default_server;/listen 80 default_server;/' "$TMPCONF"
else
  echo "write-vhost: another site already enabled ($ENABLED) — co-hosted, using name-based vhost"
  if [ -z "$SERVER_NAME" ]; then
    echo "write-vhost: ERROR — this VM already serves another site, but no server-name was supplied." >&2
    echo "A name-based vhost requires an explicit server_name; refusing to guess one." >&2
    exit 1
  fi
  # Strip any default_server flag; ensure server_name is set to the supplied value.
  sed -i -E 's/^([[:space:]]*listen[[:space:]]+80)[[:space:]]+default_server;/\1;/' "$TMPCONF"
  if grep -qE '^[[:space:]]*server_name[[:space:]]' "$TMPCONF"; then
    sed -i -E "s/^([[:space:]]*server_name[[:space:]]+).*/\\1${SERVER_NAME};/" "$TMPCONF"
  else
    sed -i -E "/^[[:space:]]*listen[[:space:]]+80;/a\\    server_name ${SERVER_NAME};" "$TMPCONF"
  fi
fi

# ── Upload, symlink, and test ───────────────────────────────────────────────────

scp -i "$SSH_KEY" "$TMPCONF" "ubuntu@$HOST:/tmp/${REMOTE_SITE}.conf"
SSH "sudo mv /tmp/${REMOTE_SITE}.conf /etc/nginx/sites-available/${REMOTE_SITE} && \
     sudo ln -sf /etc/nginx/sites-available/${REMOTE_SITE} /etc/nginx/sites-enabled/${REMOTE_SITE} && \
     sudo nginx -t"

echo "write-vhost: PASS — /etc/nginx/sites-available/${REMOTE_SITE} written and nginx -t succeeded"

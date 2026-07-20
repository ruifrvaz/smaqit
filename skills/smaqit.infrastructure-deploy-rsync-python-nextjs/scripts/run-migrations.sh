#!/usr/bin/env bash
# Deterministic migration runner for smaqit.infrastructure-deploy-rsync-python-nextjs.
#
# Runs on the target VM (invoked over SSH, after `docker compose up -d`). Waits for the
# database service's healthcheck to report "healthy", then runs the migration command inside
# the already-running application container via `docker compose exec` — never a standalone
# throwaway container, which cannot reach the compose network's `db` hostname.
#
# Usage: run-migrations.sh <app_dir> <db_service> <exec_service> -- <migration_command...>
# Example: run-migrations.sh __APP_DIR__ db api -- alembic upgrade head

set -euo pipefail

DEPLOY_PATH="$1"
DB_SERVICE="$2"
EXEC_SERVICE="$3"
shift 3
if [ "${1:-}" = "--" ]; then
  shift
fi
if [ "$#" -eq 0 ]; then
  echo "run-migrations: no migration command given after '--'" >&2
  exit 1
fi
MIGRATION_CMD=("$@")

cd "$DEPLOY_PATH"

CONTAINER_ID="$(docker compose ps -q "$DB_SERVICE")"
if [ -z "$CONTAINER_ID" ]; then
  echo "run-migrations: no running container for service '$DB_SERVICE' — run 'docker compose up -d' first" >&2
  exit 1
fi

echo "run-migrations: waiting for '$DB_SERVICE' to report healthy..."
STATUS="unknown"
for _ in $(seq 1 30); do
  STATUS="$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_ID" 2>/dev/null || echo "unknown")"
  if [ "$STATUS" = "healthy" ]; then
    break
  fi
  sleep 2
done

if [ "$STATUS" != "healthy" ]; then
  echo "run-migrations: '$DB_SERVICE' never became healthy (last status: $STATUS)" >&2
  exit 1
fi

echo "run-migrations: '$DB_SERVICE' healthy — running: ${MIGRATION_CMD[*]}"
docker compose exec -T "$EXEC_SERVICE" "${MIGRATION_CMD[@]}"

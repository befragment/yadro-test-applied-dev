#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

ENV_FILE="${PROJECT_ROOT}/.env"

if [ -f "$ENV_FILE" ]; then
  set -a
  . "$ENV_FILE"
  set +a
fi

GOOSE_DRIVER="postgres"
GOOSE_DBSTRING="${GOOSE_DRIVER}://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"

apply_migrations() {
  echo "Applying migrations from ${MIGRATIONS_DIR}..."
  GOOSE_DRIVER="${GOOSE_DRIVER}" GOOSE_DBSTRING="${GOOSE_DBSTRING}" \
    goose -dir "${MIGRATIONS_DIR}" up
}

rollback_migrations() {
  echo "Rolling back migrations..."
  GOOSE_DRIVER="${GOOSE_DRIVER}" GOOSE_DBSTRING="${GOOSE_DBSTRING}" \
    goose -dir "${MIGRATIONS_DIR}" down
}

show_status() {
  echo "Migration status in ${MIGRATIONS_DIR}:"
  GOOSE_DRIVER="${GOOSE_DRIVER}" GOOSE_DBSTRING="${GOOSE_DBSTRING}" \
    goose -dir "${MIGRATIONS_DIR}" status
}

case "$1" in
  up)
    apply_migrations
    ;;
  down)
    rollback_migrations
    ;;
  status)
    show_status
    ;;
  *)
    echo "Usage: $0 {up|down|status}"
    exit 1
    ;;
esac
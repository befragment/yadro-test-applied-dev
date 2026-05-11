#!/bin/sh
set -e

# Определяем путь к директории, где лежит этот скрипт
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Проверяем наличие .env
ENV_FILE="${PROJECT_ROOT}/.env"

if [ ! -f "$ENV_FILE" ]; then
  echo ".env not found at $ENV_FILE"
  exit 1
fi

set -a
. "$ENV_FILE"
set +a

docker exec -it postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}
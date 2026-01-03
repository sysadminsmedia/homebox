#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

usage() {
cat <<EOF
Usage: $0 [options]

Options:
  --sqlite-uri URI             SQLite connection string reachable inside the migration container
                               (default: sqlite:///data/homebox.db)
  --sqlite-host-path PATH      Absolute path to the SQLite database on the host. The script will
                               mount its directory automatically and compute the correct URI.
  --postgres-uri URI           PostgreSQL connection string (overrides the pg-* flags)
  --pg-host HOST               PostgreSQL host for schema creation (default: db)
  --pg-port PORT               PostgreSQL port (default: 5432)
  --pg-user USER               PostgreSQL user (default: homebox)
  --pg-password PASS           PostgreSQL password (default: homebox)
  --pg-db NAME                 PostgreSQL database name (default: homebox)
  --homebox-service NAME       Docker Compose service name for Homebox (default: homebox)
  --db-service NAME            Docker Compose service name for PostgreSQL (default: db)
  --migration-service NAME     Docker Compose service name for pgloader (default: migration)
  --wait-text TEXT             Log text that indicates Homebox finished migrations
                               (default: "Server is running")
  -h, --help                   Show this help message
EOF
}

SQLITE_URI="sqlite:///data/homebox.db"
SQLITE_HOST_PATH=""
SQLITE_CONTAINER_DIR="/migration-sqlite"

PG_HOST="db"
PG_PORT="5432"
PG_USER="homebox"
PG_PASSWORD="homebox"
PG_DATABASE="homebox"
POSTGRES_URI=""

HOMEBOX_SERVICE="homebox"
DB_SERVICE="db"
MIGRATION_SERVICE="migration"
HOMEBOX_WAIT_TEXT="Server is running"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --sqlite-uri)
      SQLITE_URI="$2"
      shift 2
      ;;
    --sqlite-host-path)
      SQLITE_HOST_PATH="$2"
      shift 2
      ;;
    --postgres-uri)
      POSTGRES_URI="$2"
      shift 2
      ;;
    --pg-host)
      PG_HOST="$2"
      shift 2
      ;;
    --pg-port)
      PG_PORT="$2"
      shift 2
      ;;
    --pg-user)
      PG_USER="$2"
      shift 2
      ;;
    --pg-password)
      PG_PASSWORD="$2"
      shift 2
      ;;
    --pg-db)
      PG_DATABASE="$2"
      shift 2
      ;;
    --homebox-service)
      HOMEBOX_SERVICE="$2"
      shift 2
      ;;
    --db-service)
      DB_SERVICE="$2"
      shift 2
      ;;
    --migration-service)
      MIGRATION_SERVICE="$2"
      shift 2
      ;;
    --wait-text)
      HOMEBOX_WAIT_TEXT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      usage
      exit 1
      ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo -e "${RED}Error: docker is not installed or not in PATH${NC}"
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo -e "${RED}Error: docker compose plugin not available${NC}"
  exit 1
fi

SQLITE_EXTRA_VOLUMES=()
if [[ -n "$SQLITE_HOST_PATH" ]]; then
  if [[ ! -f "$SQLITE_HOST_PATH" ]]; then
    echo -e "${RED}Error: SQLite database not found at $SQLITE_HOST_PATH${NC}"
    exit 1
  fi
  SQLITE_HOST_REALPATH="$(realpath "$SQLITE_HOST_PATH")"
  SQLITE_HOST_DIR="$(dirname "$SQLITE_HOST_REALPATH")"
  SQLITE_HOST_FILE="$(basename "$SQLITE_HOST_REALPATH")"
  SQLITE_EXTRA_VOLUMES+=(-v "${SQLITE_HOST_DIR}:${SQLITE_CONTAINER_DIR}:ro")
  SQLITE_URI="sqlite:////migration-sqlite/${SQLITE_HOST_FILE}"
fi

if [[ -z "$POSTGRES_URI" ]]; then
  POSTGRES_URI="postgresql://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DATABASE}"
fi

escape_sed() {
  printf '%s' "$1" | sed -e 's/[\/&]/\\&/g'
}

if [[ ! -f "$SCRIPT_DIR/homebox.load" ]]; then
  echo -e "${RED}Error: homebox.load template not found in $SCRIPT_DIR${NC}"
  exit 1
fi

TMP_LOAD_FILE="$(mktemp)"
cleanup() {
  rm -f "$TMP_LOAD_FILE"
}
trap cleanup EXIT

sed \
  -e "s/{{SQLITE_URI}}/$(escape_sed "$SQLITE_URI")/g" \
  -e "s/{{POSTGRES_URI}}/$(escape_sed "$POSTGRES_URI")/g" \
  "$SCRIPT_DIR/homebox.load" > "$TMP_LOAD_FILE"

TMP_LOAD_CONTAINER_PATH="/tmp/homebox.load"
LOAD_VOLUME=(-v "${TMP_LOAD_FILE}:${TMP_LOAD_CONTAINER_PATH}:ro")

echo -e "${GREEN}Starting Homebox SQLite to Postgres Migration...${NC}"

echo -e "${GREEN}1. Ensuring Docker services are running...${NC}"
docker compose stop "$HOMEBOX_SERVICE" >/dev/null || true
docker compose up -d "$DB_SERVICE"

echo -e "${GREEN}2. Waiting for Postgres to be ready...${NC}"
until docker compose exec "$DB_SERVICE" pg_isready -U "$PG_USER" -d "$PG_DATABASE" >/dev/null 2>&1; do
  echo "Waiting for postgres..."
  sleep 2
done

echo -e "${GREEN}3. Running Homebox to create schema...${NC}"
HOMEBOX_CONTAINER_ID=$(docker compose run -d --rm \
  -e HBOX_DATABASE_DRIVER="postgres" \
  -e HBOX_DATABASE_HOST="$PG_HOST" \
  -e HBOX_DATABASE_PORT="$PG_PORT" \
  -e HBOX_DATABASE_USERNAME="$PG_USER" \
  -e HBOX_DATABASE_USER="$PG_USER" \
  -e HBOX_DATABASE_PASSWORD="$PG_PASSWORD" \
  -e HBOX_DATABASE_DATABASE="$PG_DATABASE" \
  -e HBOX_DATABASE_SSL_MODE="disable" \
  "$HOMEBOX_SERVICE")

echo "Waiting for Homebox to finish migrations and start..."
if ! timeout 90s bash -c "docker logs -f $HOMEBOX_CONTAINER_ID | grep -m 1 \"$HOMEBOX_WAIT_TEXT\"" ; then
  echo -e "${RED}Homebox failed to start or timed out.${NC}"
  docker stop "$HOMEBOX_CONTAINER_ID" >/dev/null || true
  exit 1
fi

echo -e "${GREEN}4. Schema created. Stopping Homebox container...${NC}"
docker stop "$HOMEBOX_CONTAINER_ID" >/dev/null

echo -e "${GREEN}5. Starting data migration...${NC}"
docker compose run --rm \
  "${SQLITE_EXTRA_VOLUMES[@]}" \
  "${LOAD_VOLUME[@]}" \
  "$MIGRATION_SERVICE" \
  pgloader "$TMP_LOAD_CONTAINER_PATH"

echo -e "${GREEN}Migration complete!${NC}"
echo "You can now update your docker-compose.yml to use Postgres permanently."

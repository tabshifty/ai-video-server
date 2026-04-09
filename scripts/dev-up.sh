#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
ENV_EXAMPLE="$ROOT_DIR/.env.example"
FRONTEND_MODE="dev"
FRONTEND_DIR="$ROOT_DIR/admin-web"

print_usage() {
  cat <<'EOF'
Usage: bash scripts/dev-up.sh [--frontend dev|build|off]

Options:
  --frontend dev    Start frontend with Vite dev server (default)
  --frontend build  Build frontend assets only
  --frontend off    Skip frontend
  -h, --help        Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
  --frontend)
    if [[ $# -lt 2 ]]; then
      printf 'missing value for --frontend\n' >&2
      exit 1
    fi
    FRONTEND_MODE="$2"
    shift 2
    ;;
  -h | --help)
    print_usage
    exit 0
    ;;
  *)
    printf 'unknown argument: %s\n' "$1" >&2
    print_usage >&2
    exit 1
    ;;
  esac
done

if [[ "$FRONTEND_MODE" != "dev" && "$FRONTEND_MODE" != "build" && "$FRONTEND_MODE" != "off" ]]; then
  printf 'invalid --frontend mode: %s (expected: dev|build|off)\n' "$FRONTEND_MODE" >&2
  exit 1
fi

log() {
  printf '[dev-up] %s\n' "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing command: %s\n' "$1" >&2
    exit 1
  fi
}

if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  printf 'docker compose is required\n' >&2
  exit 1
fi

require_cmd go
require_cmd docker
if [[ "$FRONTEND_MODE" != "off" ]]; then
  require_cmd npm
fi

cd "$ROOT_DIR"
mkdir -p "$RUN_DIR"

if [[ ! -f "$ENV_FILE" ]]; then
  cp "$ENV_EXAMPLE" "$ENV_FILE"
  log "created env file from .env.example: $ENV_FILE"
fi

trim_spaces() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

load_env_file() {
  local file="$1"
  local line line_no key value
  line_no=0
  while IFS= read -r line || [[ -n "$line" ]]; do
    ((line_no += 1))
    line="${line%$'\r'}"
    if [[ $line_no -eq 1 ]]; then
      line="${line#$'\xEF\xBB\xBF'}"
    fi
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue
    [[ "$line" =~ ^[[:space:]]*# ]] && continue

    line="$(trim_spaces "$line")"
    if [[ ! "$line" =~ ^[A-Za-z_][A-Za-z0-9_]*= ]]; then
      log "skip invalid env line $line_no"
      continue
    fi

    key="${line%%=*}"
    value="${line#*=}"
    key="$(trim_spaces "$key")"
    value="$(trim_spaces "$value")"

    if [[ ${#value} -ge 2 && "${value:0:1}" == "\"" && "${value: -1}" == "\"" ]]; then
      value="${value:1:${#value}-2}"
    elif [[ ${#value} -ge 2 && "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; then
      value="${value:1:${#value}-2}"
    fi

    export "$key=$value"
  done <"$file"
}

load_env_file "$ENV_FILE"
export ENV_FILE
log "using env file: $ENV_FILE"

is_running() {
  local pid_file="$1"
  if [[ ! -f "$pid_file" ]]; then
    return 1
  fi
  local pid
  pid="$(cat "$pid_file" 2>/dev/null || true)"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

wait_postgres() {
  local container_id="$1"
  local retries=30
  local delay=2
  local i
  for ((i = 1; i <= retries; i++)); do
    if docker exec "$container_id" pg_isready -U video -d video_server >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done
  return 1
}

apply_migrations() {
  local container_id="$1"
  docker exec -i "$container_id" psql -U video -d video_server -v ON_ERROR_STOP=1 >/dev/null <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

  local file
  while IFS= read -r -d '' file; do
    local version
    version="$(basename "$file")"

    local already
    already="$(docker exec -i "$container_id" psql -U video -d video_server -tA -c "SELECT 1 FROM schema_migrations WHERE version='${version}' LIMIT 1;")"
    if [[ "$already" == "1" ]]; then
      log "skip migration $version (already applied)"
      continue
    fi

    log "applying migration $version"
    cat "$file" | docker exec -i "$container_id" psql -U video -d video_server -v ON_ERROR_STOP=1 >/dev/null
    docker exec -i "$container_id" psql -U video -d video_server -v ON_ERROR_STOP=1 >/dev/null <<SQL
INSERT INTO schema_migrations(version) VALUES ('${version}')
ON CONFLICT(version) DO NOTHING;
SQL
  done < <(find "$ROOT_DIR/migrations" -maxdepth 1 -type f -name '*.up.sql' -print0 | sort -z)
}

log "starting postgres and redis"
"${COMPOSE_CMD[@]}" up -d postgres redis >/dev/null

POSTGRES_CID="$("${COMPOSE_CMD[@]}" ps -q postgres)"
if [[ -z "$POSTGRES_CID" ]]; then
  printf 'failed to resolve postgres container id\n' >&2
  exit 1
fi

log "waiting for postgres"
if ! wait_postgres "$POSTGRES_CID"; then
  printf 'postgres is not ready\n' >&2
  exit 1
fi

apply_migrations "$POSTGRES_CID"

if is_running "$RUN_DIR/server.pid"; then
  log "server already running (pid $(cat "$RUN_DIR/server.pid"))"
else
  log "starting server"
  nohup go run main.go -mode server >"$RUN_DIR/server.log" 2>&1 &
  echo "$!" >"$RUN_DIR/server.pid"
fi

if is_running "$RUN_DIR/worker.pid"; then
  log "worker already running (pid $(cat "$RUN_DIR/worker.pid"))"
else
  log "starting worker"
  nohup go run main.go -mode worker >"$RUN_DIR/worker.log" 2>&1 &
  echo "$!" >"$RUN_DIR/worker.pid"
fi

if [[ "$FRONTEND_MODE" == "off" ]]; then
  log "frontend is disabled"
elif [[ ! -d "$FRONTEND_DIR" ]]; then
  printf 'frontend directory not found: %s\n' "$FRONTEND_DIR" >&2
  exit 1
else
  if [[ ! -d "$FRONTEND_DIR/node_modules" ]]; then
    log "installing frontend dependencies"
    npm --prefix "$FRONTEND_DIR" install
  fi

  if [[ "$FRONTEND_MODE" == "dev" ]]; then
    if is_running "$RUN_DIR/frontend.pid"; then
      log "frontend already running (pid $(cat "$RUN_DIR/frontend.pid"))"
    else
      log "starting frontend dev server"
      nohup npm --prefix "$FRONTEND_DIR" run dev >"$RUN_DIR/frontend.log" 2>&1 &
      echo "$!" >"$RUN_DIR/frontend.pid"
    fi
    log "frontend url: http://localhost:5173"
    log "frontend log: $RUN_DIR/frontend.log"
  else
    log "building frontend assets"
    npm --prefix "$FRONTEND_DIR" run build
    log "frontend build ready at $FRONTEND_DIR/dist"
    log "frontend url: http://localhost:8080/admin/"
  fi
fi

log "done"
log "server log: $RUN_DIR/server.log"
log "worker log: $RUN_DIR/worker.log"

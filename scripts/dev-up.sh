#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
ENV_EXAMPLE="$ROOT_DIR/.env.example"
FRONTEND_MODE="dev"
FRONTEND_DIR="$ROOT_DIR/admin-web"
STARTED_PID_FILES=()

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

list_compose_images() {
  "${COMPOSE_CMD[@]}" config --images 2>/dev/null | awk 'NF{print $1}' | sort -u
}

pull_image_with_retry() {
  local image="$1"
  local retries="${2:-3}"
  local delay=2
  local attempt output
  for ((attempt = 1; attempt <= retries; attempt++)); do
    output="$(docker pull "$image" 2>&1)" && return 0
    printf '%s\n' "$output" >&2
    if [[ "$output" == *"unable to fetch descriptor"* && "$output" == *"content size of zero"* ]]; then
      cat >&2 <<'EOF'
[dev-up] 检测到 Docker 镜像源返回损坏 descriptor（content size of zero）。
[dev-up] 这通常是 registry mirror 缓存异常，不是项目代码问题。
[dev-up] 建议操作：
[dev-up] 1) Docker Desktop -> Settings -> Docker Engine，移除异常镜像源后重启 Docker
[dev-up] 2) 或暂时仅保留一个稳定镜像源再重试
[dev-up] 3) 再执行: docker pull <image>
EOF
    fi
    if [[ "$attempt" -lt "$retries" ]]; then
      log "retry docker pull ($attempt/$retries): $image"
      sleep "$delay"
      delay=$((delay * 2))
    fi
  done
  return 1
}

ensure_compose_images_ready() {
  local image
  while IFS= read -r image; do
    [[ -n "$image" ]] || continue
    if docker image inspect "$image" >/dev/null 2>&1; then
      continue
    fi
    log "pulling image: $image"
    if ! pull_image_with_retry "$image" 3; then
      printf 'failed to pull image: %s\n' "$image" >&2
      return 1
    fi
  done < <(list_compose_images)
}

compose_up_db_services() {
  if "${COMPOSE_CMD[@]}" up --help 2>/dev/null | grep -q -- "--pull"; then
    "${COMPOSE_CMD[@]}" up --pull missing -d postgres redis >/dev/null
    return 0
  fi
  "${COMPOSE_CMD[@]}" up -d postgres redis >/dev/null
}

is_running() {
  local pid_file="$1"
  if [[ ! -f "$pid_file" ]]; then
    return 1
  fi
  local pid meta cmdline
  IFS=$'\t' read -r pid meta <"$pid_file" || true
  if [[ -z "${meta:-}" ]]; then
    # backward compatibility with old pid files that only contain pid
    pid="$(cat "$pid_file" 2>/dev/null || true)"
  fi
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  kill -0 "$pid" 2>/dev/null || return 1

  if [[ -n "${meta:-}" ]]; then
    cmdline="$(ps -o command= -p "$pid" 2>/dev/null || true)"
    [[ -n "$cmdline" ]] || return 1
    [[ "$cmdline" == *"$meta"* ]] || return 1
  fi

  return 0
}

pid_is_alive() {
  local pid="$1"
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  kill -0 "$pid" 2>/dev/null
}

collect_descendant_pids() {
  local parent="$1"
  local child
  while IFS= read -r child; do
    [[ "$child" =~ ^[0-9]+$ ]] || continue
    printf '%s\n' "$child"
    collect_descendant_pids "$child"
  done < <(pgrep -P "$parent" 2>/dev/null || true)
}

pid_matches_service() {
  local name="$1"
  local cmdline="$2"
  case "$name" in
  server)
    [[ "$cmdline" == *"-mode server"* || "$cmdline" == *"go run main.go -mode server"* ]]
    ;;
  worker)
    [[ "$cmdline" == *"-mode worker"* || "$cmdline" == *"go run main.go -mode worker"* ]]
    ;;
  frontend)
    [[ "$cmdline" == *"$ROOT_DIR/admin-web"* || "$cmdline" == *"vite"* ]]
    ;;
  *)
    return 1
    ;;
  esac
}

stop_pid_tree() {
  local pid="$1"
  local name="$2"
  local i child cmdline descendants_text

  if ! pid_is_alive "$pid"; then
    return 0
  fi

  descendants_text="$(collect_descendant_pids "$pid" 2>/dev/null || true)"
  cmdline="$(ps -o command= -p "$pid" 2>/dev/null || true)"
  log "stopping $name (pid $pid) cmd: ${cmdline:-unknown}"

  while IFS= read -r child; do
    [[ "$child" =~ ^[0-9]+$ ]] || continue
    kill "$child" 2>/dev/null || true
  done <<<"$descendants_text"
  kill "$pid" 2>/dev/null || true

  for ((i = 1; i <= 30; i++)); do
    if ! pid_is_alive "$pid"; then
      local alive_descendants=0
      while IFS= read -r child; do
        [[ "$child" =~ ^[0-9]+$ ]] || continue
        if pid_is_alive "$child"; then
          alive_descendants=1
          break
        fi
      done <<<"$descendants_text"
      if [[ "$alive_descendants" -eq 0 ]]; then
        return 0
      fi
    fi
    sleep 0.2
  done

  while IFS= read -r child; do
    [[ "$child" =~ ^[0-9]+$ ]] || continue
    kill -9 "$child" 2>/dev/null || true
  done <<<"$descendants_text"
  kill -9 "$pid" 2>/dev/null || true
}

stop_pid_file() {
  local pid_file="$1"
  local name="$2"
  local pid meta cmdline

  [[ -f "$pid_file" ]] || return 0
  IFS=$'\t' read -r pid meta <"$pid_file" || true
  if [[ -z "${meta:-}" ]]; then
    pid="$(cat "$pid_file" 2>/dev/null || true)"
  fi

  if [[ ! "$pid" =~ ^[0-9]+$ ]]; then
    rm -f "$pid_file"
    return 0
  fi

  if ! pid_is_alive "$pid"; then
    rm -f "$pid_file"
    return 0
  fi

  if [[ -n "${meta:-}" ]]; then
    cmdline="$(ps -o command= -p "$pid" 2>/dev/null || true)"
    if [[ "$cmdline" != *"$meta"* ]]; then
      if pid_matches_service "$name" "$cmdline"; then
        log "$name pid $pid command changed after startup, continue stopping"
      else
        log "$name pid $pid command mismatch; skip stop"
        rm -f "$pid_file"
        return 0
      fi
    fi
  fi

  stop_pid_tree "$pid" "$name"
  rm -f "$pid_file"
}

start_bg_process() {
  local name="$1"
  local pid_file="$2"
  local log_file="$3"
  shift 3

  if is_running "$pid_file"; then
    log "$name already running (pid $(cut -f1 "$pid_file"))"
    return 0
  fi

  log "starting $name"
  nohup "$@" >"$log_file" 2>&1 &
  local pid="$!"
  printf '%s\t%s\n' "$pid" "$*" >"$pid_file"

  sleep 1
  if ! kill -0 "$pid" 2>/dev/null; then
    printf '%s failed to start. check log: %s\n' "$name" "$log_file" >&2
    if [[ -f "$log_file" ]]; then
      tail -n 80 "$log_file" >&2 || true
    fi
    rm -f "$pid_file"
    return 1
  fi
  STARTED_PID_FILES+=("$pid_file")
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
  for file in "$ROOT_DIR"/migrations/*.up.sql; do
    [[ -f "$file" ]] || continue
    local version
    version="$(basename "$file")"

    local already
    already="$(docker exec -i "$container_id" psql -U video -d video_server -tA -c "SELECT 1 FROM schema_migrations WHERE version='${version}' LIMIT 1;")"
    if [[ "$already" == "1" ]]; then
      log "skip migration $version (already applied)"
      continue
    fi

    log "applying migration $version"
    docker exec -i "$container_id" psql -U video -d video_server -v ON_ERROR_STOP=1 >/dev/null <"$file"
    docker exec -i "$container_id" psql -U video -d video_server -v ON_ERROR_STOP=1 >/dev/null <<SQL
INSERT INTO schema_migrations(version) VALUES ('${version}')
ON CONFLICT(version) DO NOTHING;
SQL
  done
}

hash_file_sha256() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
    return 0
  fi
  printf 'sha256 tool not found (need sha256sum or shasum)\n' >&2
  return 1
}

ensure_frontend_deps() {
  local lock_file="$FRONTEND_DIR/package-lock.json"
  local state_file="$RUN_DIR/frontend.lock.sha256"
  local current_hash="" previous_hash=""

  if [[ ! -f "$lock_file" ]]; then
    if [[ ! -d "$FRONTEND_DIR/node_modules" ]]; then
      log "installing frontend dependencies (npm install)"
      npm --prefix "$FRONTEND_DIR" install
    fi
    return 0
  fi

  current_hash="$(hash_file_sha256 "$lock_file")"
  if [[ -f "$state_file" ]]; then
    previous_hash="$(cat "$state_file" 2>/dev/null || true)"
  fi

  if [[ ! -d "$FRONTEND_DIR/node_modules" || "$current_hash" != "$previous_hash" ]]; then
    log "installing frontend dependencies (npm ci)"
    npm --prefix "$FRONTEND_DIR" ci
    printf '%s\n' "$current_hash" >"$state_file"
  fi
}

cleanup_on_error() {
  local exit_code="$?"
  local i pid_file
  if [[ "$exit_code" -eq 0 ]]; then
    return 0
  fi

  log "startup failed, cleaning up started processes"
  for ((i = ${#STARTED_PID_FILES[@]} - 1; i >= 0; i--)); do
    pid_file="${STARTED_PID_FILES[i]}"
    stop_pid_file "$pid_file" "$(basename "$pid_file" .pid)"
  done

  trap - EXIT
  exit "$exit_code"
}

trap cleanup_on_error EXIT

log "ensuring docker images are available"
ensure_compose_images_ready

log "starting postgres and redis"
compose_up_db_services

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

start_bg_process "server" "$RUN_DIR/server.pid" "$RUN_DIR/server.log" go run main.go -mode server
start_bg_process "worker" "$RUN_DIR/worker.pid" "$RUN_DIR/worker.log" go run main.go -mode worker

if [[ "$FRONTEND_MODE" == "off" ]]; then
  log "frontend is disabled"
elif [[ ! -d "$FRONTEND_DIR" ]]; then
  printf 'frontend directory not found: %s\n' "$FRONTEND_DIR" >&2
  exit 1
else
  ensure_frontend_deps

  if [[ "$FRONTEND_MODE" == "dev" ]]; then
    start_bg_process "frontend" "$RUN_DIR/frontend.pid" "$RUN_DIR/frontend.log" npm --prefix "$FRONTEND_DIR" run dev
    log "frontend url: http://localhost:5173"
    log "frontend log: $RUN_DIR/frontend.log"
  else
    log "building frontend assets"
    npm --prefix "$FRONTEND_DIR" run build
    log "frontend build ready at $FRONTEND_DIR/dist"
    log "frontend url: http://localhost:8080/admin/"
  fi
fi

trap - EXIT
log "done"
log "server log: $RUN_DIR/server.log"
log "worker log: $RUN_DIR/worker.log"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"

log() {
  printf '[dev-down] %s\n' "$*"
}

if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  printf 'docker compose is required\n' >&2
  exit 1
fi

stop_by_pid_file() {
  local pid_file="$1"
  local name="$2"
  if [[ ! -f "$pid_file" ]]; then
    log "$name is not running"
    return 0
  fi
  local pid
  pid="$(cat "$pid_file" 2>/dev/null || true)"
  if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
    log "stopping $name (pid $pid)"
    kill "$pid" 2>/dev/null || true
  else
    log "$name pid not active"
  fi
  rm -f "$pid_file"
}

cd "$ROOT_DIR"
mkdir -p "$RUN_DIR"

stop_by_pid_file "$RUN_DIR/server.pid" "server"
stop_by_pid_file "$RUN_DIR/worker.pid" "worker"

log "stopping postgres and redis"
"${COMPOSE_CMD[@]}" stop postgres redis >/dev/null || true

log "done"

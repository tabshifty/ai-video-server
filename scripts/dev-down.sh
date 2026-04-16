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
  local pid meta cmdline i
  if [[ ! -f "$pid_file" ]]; then
    log "$name is not running"
    return 0
  fi

  IFS=$'\t' read -r pid meta <"$pid_file" || true
  if [[ -z "${meta:-}" ]]; then
    # backward compatibility with old pid files that only contain pid
    pid="$(cat "$pid_file" 2>/dev/null || true)"
  fi

  if [[ ! "$pid" =~ ^[0-9]+$ ]]; then
    log "$name pid invalid, removing pid file"
    rm -f "$pid_file"
    return 0
  fi

  if kill -0 "$pid" 2>/dev/null; then
    if [[ -n "${meta:-}" ]]; then
      cmdline="$(ps -o command= -p "$pid" 2>/dev/null || true)"
      if [[ "$cmdline" != *"$meta"* ]]; then
        log "$name pid $pid command mismatch, skip stop"
        rm -f "$pid_file"
        return 0
      fi
    fi

    log "stopping $name (pid $pid)"
    kill "$pid" 2>/dev/null || true

    for ((i = 1; i <= 20; i++)); do
      if ! kill -0 "$pid" 2>/dev/null; then
        rm -f "$pid_file"
        return 0
      fi
      sleep 0.2
    done

    log "$name did not exit in time, force kill (pid $pid)"
    kill -9 "$pid" 2>/dev/null || true
  else
    log "$name pid not active"
  fi
  rm -f "$pid_file"
}

cd "$ROOT_DIR"
mkdir -p "$RUN_DIR"

stop_by_pid_file "$RUN_DIR/server.pid" "server"
stop_by_pid_file "$RUN_DIR/worker.pid" "worker"
stop_by_pid_file "$RUN_DIR/frontend.pid" "frontend"

log "stopping postgres and redis"
"${COMPOSE_CMD[@]}" stop postgres redis >/dev/null || true

log "done"

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

  log "$name did not exit in time, force kill (pid $pid)"
  while IFS= read -r child; do
    [[ "$child" =~ ^[0-9]+$ ]] || continue
    kill -9 "$child" 2>/dev/null || true
  done <<<"$descendants_text"
  kill -9 "$pid" 2>/dev/null || true
}

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

  if pid_is_alive "$pid"; then
    if [[ -n "${meta:-}" ]]; then
      cmdline="$(ps -o command= -p "$pid" 2>/dev/null || true)"
      if [[ "$cmdline" != *"$meta"* ]]; then
        if pid_matches_service "$name" "$cmdline"; then
          log "$name pid $pid command changed after startup, continue stopping"
        else
          log "$name pid $pid command mismatch, skip stop"
          rm -f "$pid_file"
          return 0
        fi
      fi
    fi

    stop_pid_tree "$pid" "$name"
  else
    log "$name pid not active"
  fi
  rm -f "$pid_file"
}

stop_by_pattern() {
  local name="$1"
  local pattern="$2"
  local pid
  while IFS= read -r pid; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    stop_pid_tree "$pid" "$name"
  done < <(pgrep -f "$pattern" 2>/dev/null || true)
}

stop_by_port() {
  local name="$1"
  local port="$2"
  local pid
  while IFS= read -r pid; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    stop_pid_tree "$pid" "$name"
  done < <(lsof -t -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u)
}

stop_worker_from_log() {
  local log_file="$RUN_DIR/worker.log"
  local pid
  [[ -f "$log_file" ]] || return 0
  pid="$(grep -oE 'pid=[0-9]+' "$log_file" 2>/dev/null | tail -n 1 | cut -d= -f2)"
  if [[ "$pid" =~ ^[0-9]+$ ]] && pid_is_alive "$pid"; then
    stop_pid_tree "$pid" "worker"
  fi
}

cd "$ROOT_DIR"
mkdir -p "$RUN_DIR"

stop_by_pid_file "$RUN_DIR/server.pid" "server"
stop_by_pid_file "$RUN_DIR/worker.pid" "worker"
stop_by_pid_file "$RUN_DIR/frontend.pid" "frontend"

# Fallback cleanup for legacy/invalid pid files or command shifts (go run -> temp binary, npm -> vite node).
stop_by_pattern "server" "main -mode server"
stop_by_pattern "worker" "main -mode worker"
stop_by_pattern "frontend" "$ROOT_DIR/admin-web"
stop_by_port "server" "8080"
stop_worker_from_log
stop_by_port "frontend" "5173"

log "stopping postgres and redis"
"${COMPOSE_CMD[@]}" stop postgres redis >/dev/null || true

log "done"

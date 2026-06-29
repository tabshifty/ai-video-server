#!/usr/bin/env bash
set -Eeuo pipefail

if [[ $# -lt 1 ]]; then
  printf 'usage: %s <ed2k-link>\n' "$0" >&2
  exit 64
fi

SOURCE_LINK="$1"
TASK_ID="${ED2K_TASK_ID:-}"
RESOURCE_HASH="${ED2K_RESOURCE_HASH:-}"
FILENAME="${ED2K_FILENAME:-}"
DECLARED_SIZE="${ED2K_DECLARED_SIZE:-0}"

AMULECMD_BIN="${AMULECMD_BIN:-amulecmd}"
AMULED_BIN="${AMULED_BIN:-amuled}"
AMULE_REMOTE_HOST="${AMULE_REMOTE_HOST:-127.0.0.1}"
AMULE_REMOTE_PORT="${AMULE_REMOTE_PORT:-4712}"
AMULE_REMOTE_PASSWORD="${AMULE_REMOTE_PASSWORD:-}"
ED2K_DOWNLOAD_ROOT="${ED2K_DOWNLOAD_ROOT:-}"
ED2K_DOWNLOAD_SUBDIR="${ED2K_DOWNLOAD_SUBDIR:-ed2k-downloads}"
ED2K_WAIT_TIMEOUT_SECONDS="${ED2K_WAIT_TIMEOUT_SECONDS:-21600}"
ED2K_POLL_INTERVAL_SECONDS="${ED2K_POLL_INTERVAL_SECONDS:-5}"
ED2K_OUTPUT_DIR_MODE="${ED2K_OUTPUT_DIR_MODE:-hash}"

if [[ -z "$ED2K_DOWNLOAD_ROOT" ]]; then
  printf 'ED2K_DOWNLOAD_ROOT is required\n' >&2
  exit 65
fi

if [[ -z "$AMULE_REMOTE_PASSWORD" ]]; then
  printf 'AMULE_REMOTE_PASSWORD is required\n' >&2
  exit 66
fi

if ! command -v "$AMULECMD_BIN" >/dev/null 2>&1; then
  printf 'amulecmd not found: %s\n' "$AMULECMD_BIN" >&2
  exit 67
fi

if ! command -v "$AMULED_BIN" >/dev/null 2>&1; then
  printf 'amuled not found: %s\n' "$AMULED_BIN" >&2
  exit 68
fi

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

lower() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]'
}

sanitize_path_component() {
  local raw="$1"
  raw="$(printf '%s' "$raw" | tr '/:' '__')"
  raw="$(printf '%s' "$raw" | tr -cd '[:alnum:]._ -')"
  raw="$(trim "$raw")"
  if [[ -z "$raw" ]]; then
    raw="download"
  fi
  printf '%s' "$raw"
}

build_output_dir() {
  local base="$ED2K_DOWNLOAD_ROOT/$ED2K_DOWNLOAD_SUBDIR"
  mkdir -p "$base"
  case "$(lower "$ED2K_OUTPUT_DIR_MODE")" in
    task_id)
      if [[ -n "$TASK_ID" ]]; then
        printf '%s/%s\n' "$base" "$TASK_ID"
        return
      fi
      ;;
    filename)
      if [[ -n "$FILENAME" ]]; then
        printf '%s/%s\n' "$base" "$(sanitize_path_component "$FILENAME")"
        return
      fi
      ;;
  esac
  if [[ -n "$RESOURCE_HASH" ]]; then
    printf '%s/%s\n' "$base" "$(sanitize_path_component "$RESOURCE_HASH")"
    return
  fi
  if [[ -n "$TASK_ID" ]]; then
    printf '%s/%s\n' "$base" "$TASK_ID"
    return
  fi
  printf '%s/%s\n' "$base" "$(sanitize_path_component "$FILENAME")"
}

OUTPUT_DIR="$(build_output_dir)"
mkdir -p "$OUTPUT_DIR"

AMULE_CMD_BASE=(
  "$AMULECMD_BIN"
  -h "$AMULE_REMOTE_HOST"
  -p "$AMULE_REMOTE_PORT"
  -P "$AMULE_REMOTE_PASSWORD"
)

run_amulecmd() {
  local command="$1"
  local output
  if ! output="$("${AMULE_CMD_BASE[@]}" -c "$command" 2>&1)"; then
    printf '%s\n' "$output" >&2
    return 1
  fi
  printf '%s\n' "$output"
}

ensure_amuled_ready() {
  if pgrep -x "$(basename "$AMULED_BIN")" >/dev/null 2>&1; then
    return 0
  fi
  "$AMULED_BIN" >/dev/null 2>&1 || true
}

wait_for_remote() {
  local attempts=0
  while (( attempts < 30 )); do
    if run_amulecmd "help" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  printf 'amule remote connection is not ready\n' >&2
  return 1
}

find_download_entry() {
  local listing="$1"
  local normalized_hash
  normalized_hash="$(lower "$RESOURCE_HASH")"
  while IFS= read -r line; do
    local lowered
    lowered="$(lower "$line")"
    if [[ -n "$normalized_hash" && "$lowered" == *"$normalized_hash"* ]]; then
      printf '%s\n' "$line"
      return 0
    fi
    if [[ -n "$FILENAME" && "$line" == *"$FILENAME"* ]]; then
      printf '%s\n' "$line"
      return 0
    fi
  done <<< "$listing"
  return 1
}

extract_progress_text() {
  local line="$1"
  local progress
  progress="$(printf '%s\n' "$line" | LC_ALL=C grep -Eo '[0-9]+(\.[0-9]+)?%' | head -n1 || true)"
  if [[ -n "$progress" ]]; then
    printf '下载进行中（%s）' "$progress"
    return
  fi
  printf '下载已完成'
}

collect_files_json() {
  local output_dir="$1"
  local downloaded_path="$2"
  /usr/bin/python3 - "$output_dir" "$downloaded_path" <<'PY'
import json
import os
import sys

output_dir = os.path.abspath(sys.argv[1])
downloaded_path = os.path.abspath(sys.argv[2]) if sys.argv[2] else ""
items = []

if os.path.isfile(downloaded_path):
    targets = [downloaded_path]
elif downloaded_path and os.path.isdir(downloaded_path):
    targets = []
    for root, _, files in os.walk(downloaded_path):
        for name in files:
            targets.append(os.path.join(root, name))
else:
    targets = []
    for root, _, files in os.walk(output_dir):
        for name in files:
            targets.append(os.path.join(root, name))

targets = sorted({os.path.abspath(path) for path in targets if os.path.isfile(path)})
for path in targets:
    rel = os.path.relpath(path, output_dir)
    items.append({
        "name": os.path.basename(path),
        "path": rel,
        "size": os.path.getsize(path),
    })

print(json.dumps(items, ensure_ascii=False))
PY
}

guess_downloaded_path() {
  if [[ -n "$FILENAME" && -f "$OUTPUT_DIR/$FILENAME" ]]; then
    printf '%s\n' "$OUTPUT_DIR/$FILENAME"
    return
  fi
  local first
  first="$(find "$OUTPUT_DIR" -mindepth 1 -maxdepth 1 | head -n1 || true)"
  if [[ -n "$first" ]]; then
    printf '%s\n' "$first"
    return
  fi
  printf '%s\n' "$OUTPUT_DIR"
}

has_output_files() {
  find "$OUTPUT_DIR" -type f -print -quit | grep -q .
}

ensure_amuled_ready
wait_for_remote

run_amulecmd "set directory \"$OUTPUT_DIR\"" >/dev/null
run_amulecmd "add \"$SOURCE_LINK\"" >/dev/null

start_epoch="$(date +%s)"
deadline=$((start_epoch + ED2K_WAIT_TIMEOUT_SECONDS))
progress_text="等待 ED2K 引擎完成下载"

while true; do
  if (( $(date +%s) > deadline )); then
    printf 'ed2k download timed out after %s seconds\n' "$ED2K_WAIT_TIMEOUT_SECONDS" >&2
    exit 69
  fi

  if has_output_files; then
    progress_text="下载已完成"
    break
  fi

  listing="$(run_amulecmd "show dl" || true)"
  entry="$(find_download_entry "$listing" || true)"
  if [[ -n "$entry" ]]; then
    progress_text="$(extract_progress_text "$entry")"
    lowered_entry="$(lower "$entry")"
    if [[ "$lowered_entry" == *"100%"* || "$lowered_entry" == *"completed"* || "$lowered_entry" == *"complete"* ]]; then
      break
    fi
  fi

  sleep "$ED2K_POLL_INTERVAL_SECONDS"
done

downloaded_path="$(guess_downloaded_path)"
files_json="$(collect_files_json "$OUTPUT_DIR" "$downloaded_path")"

if [[ "$files_json" == "[]" ]]; then
  printf 'ed2k executor produced no files in %s\n' "$OUTPUT_DIR" >&2
  exit 70
fi

/usr/bin/python3 - "$OUTPUT_DIR" "$downloaded_path" "$progress_text" "$AMULECMD_BIN" "$files_json" <<'PY'
import json
import os
import sys

payload = {
    "OutputDir": os.path.abspath(sys.argv[1]),
    "DownloadedPath": os.path.abspath(sys.argv[2]),
    "ProgressText": sys.argv[3],
    "Executor": sys.argv[4],
    "Files": json.loads(sys.argv[5]),
}
print(json.dumps(payload, ensure_ascii=False))
PY

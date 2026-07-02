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
EXECUTOR_MODE="${ED2K_EXECUTOR_MODE:-submit}"

AMULECMD_BIN="${AMULECMD_BIN:-amulecmd}"
AMULED_BIN="${AMULED_BIN:-amuled}"
ED2K_BIN="${ED2K_BIN:-ed2k}"
AMULE_REMOTE_HOST="${AMULE_REMOTE_HOST:-127.0.0.1}"
AMULE_REMOTE_PORT="${AMULE_REMOTE_PORT:-4712}"
AMULE_REMOTE_PASSWORD="${AMULE_REMOTE_PASSWORD:-}"
ED2K_DOWNLOAD_ROOT="${ED2K_DOWNLOAD_ROOT:-}"
ED2K_DOWNLOAD_SUBDIR="${ED2K_DOWNLOAD_SUBDIR:-ed2k-downloads}"
ED2K_WAIT_TIMEOUT_SECONDS="${ED2K_WAIT_TIMEOUT_SECONDS:-21600}"
ED2K_POLL_INTERVAL_SECONDS="${ED2K_POLL_INTERVAL_SECONDS:-5}"
ED2K_OUTPUT_DIR_MODE="${ED2K_OUTPUT_DIR_MODE:-hash}"
ED2K_SUBMIT_CONFIRM_TIMEOUT_SECONDS="${ED2K_SUBMIT_CONFIRM_TIMEOUT_SECONDS:-12}"
ED2K_REQUEUE_IF_UNSEEN_AFTER_SECONDS="${ED2K_REQUEUE_IF_UNSEEN_AFTER_SECONDS:-45}"
ED2K_REQUEUE_INTERVAL_SECONDS="${ED2K_REQUEUE_INTERVAL_SECONDS:-300}"

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

if ! command -v "$ED2K_BIN" >/dev/null 2>&1; then
  printf 'ed2k helper not found: %s\n' "$ED2K_BIN" >&2
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
DOWNLOAD_BASE="$ED2K_DOWNLOAD_ROOT/$ED2K_DOWNLOAD_SUBDIR"
SUBMIT_MARKER="$OUTPUT_DIR/.submitted_at"
SEEN_MARKER="$OUTPUT_DIR/.download_seen"
REQUEUE_MARKER="$OUTPUT_DIR/.requeued_at"

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

run_amulecmd_add() {
  run_amulecmd "add $SOURCE_LINK"
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
  printf '已进入 aMule 下载队列，等待开始下载'
}

submit_ed2k_link() {
  "$ED2K_BIN" "$SOURCE_LINK" >/dev/null
}

touch_submit_marker() {
  date '+%s' > "$SUBMIT_MARKER"
}

mark_download_seen() {
  date '+%s' > "$SEEN_MARKER"
}

touch_requeue_marker() {
  date '+%s' > "$REQUEUE_MARKER"
}

read_marker_epoch() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    return 1
  fi
  local value
  value="$(tr -cd '0-9' < "$path" 2>/dev/null || true)"
  if [[ -z "$value" ]]; then
    return 1
  fi
  printf '%s\n' "$value"
}

wait_for_download_entry() {
  local timeout="$1"
  if [[ -z "$timeout" ]] || (( timeout <= 0 )); then
    timeout=1
  fi
  local deadline now listing entry
  deadline=$(( $(date '+%s') + timeout ))
  while :; do
    if listing="$(run_amulecmd "show dl" 2>/dev/null)"; then
      entry="$(find_download_entry "$listing" || true)"
      if [[ -n "$entry" ]]; then
        mark_download_seen
        return 0
      fi
    fi
    now="$(date '+%s')"
    if (( now >= deadline )); then
      return 1
    fi
    sleep 1
  done
}

should_requeue_unseen_download() {
  if [[ -f "$SEEN_MARKER" ]]; then
    return 1
  fi
  local submit_epoch now requeue_epoch
  submit_epoch="$(read_marker_epoch "$SUBMIT_MARKER" || true)"
  if [[ -z "$submit_epoch" ]]; then
    return 1
  fi
  now="$(date '+%s')"
  if (( now - submit_epoch < ED2K_REQUEUE_IF_UNSEEN_AFTER_SECONDS )); then
    return 1
  fi
  requeue_epoch="$(read_marker_epoch "$REQUEUE_MARKER" || true)"
  if [[ -n "$requeue_epoch" ]] && (( now - requeue_epoch < ED2K_REQUEUE_INTERVAL_SECONDS )); then
    return 1
  fi
  return 0
}

ensure_download_visible_with_fallback() {
  local confirm_timeout="$1"
  if wait_for_download_entry "$confirm_timeout"; then
    return 0
  fi
  if ! run_amulecmd_add >/dev/null; then
    return 1
  fi
  touch_requeue_marker
  wait_for_download_entry "$confirm_timeout"
}

emit_nonterminal_json() {
  local status="$1"
  local progress_text="$2"
  /usr/bin/python3 - "$status" "$progress_text" "$AMULECMD_BIN" <<'PY'
import json
import sys

payload = {
    "Status": sys.argv[1],
    "ProgressText": sys.argv[2],
    "Executor": sys.argv[3],
    "Files": [],
}
print(json.dumps(payload, ensure_ascii=False))
PY
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

ignored_names = {".submitted_at", ".download_seen", ".requeued_at"}

if os.path.isfile(downloaded_path) and os.path.basename(downloaded_path) not in ignored_names:
    targets = [downloaded_path]
elif downloaded_path and os.path.isdir(downloaded_path):
    targets = []
    for root, _, files in os.walk(downloaded_path):
        for name in files:
            if name in ignored_names:
                continue
            targets.append(os.path.join(root, name))
else:
    targets = []
    for root, _, files in os.walk(output_dir):
        for name in files:
            if name in ignored_names:
                continue
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
  if [[ -n "$FILENAME" && -f "$DOWNLOAD_BASE/$FILENAME" ]]; then
    printf '%s\n' "$DOWNLOAD_BASE/$FILENAME"
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
  find "$OUTPUT_DIR" -type f ! -name '.submitted_at' ! -name '.download_seen' ! -name '.requeued_at' -print -quit | grep -q .
}

has_downloaded_file() {
  if [[ -z "$FILENAME" || ! -f "$DOWNLOAD_BASE/$FILENAME" ]]; then
    return 1
  fi
  /usr/bin/python3 - "$DOWNLOAD_BASE/$FILENAME" "$DECLARED_SIZE" "$SUBMIT_MARKER" <<'PY'
import os
import sys

path = sys.argv[1]
try:
    declared_size = int(sys.argv[2])
except ValueError:
    declared_size = 0
marker = sys.argv[3]

try:
    stat = os.stat(path)
except OSError:
    sys.exit(1)

if declared_size > 0 and stat.st_size != declared_size:
    sys.exit(1)

if os.path.exists(marker):
    try:
        if stat.st_mtime + 1 < os.stat(marker).st_mtime:
            sys.exit(1)
    except OSError:
        sys.exit(1)

sys.exit(0)
PY
}

mirror_downloaded_file() {
  if [[ -n "$FILENAME" && -f "$DOWNLOAD_BASE/$FILENAME" && ! -e "$OUTPUT_DIR/$FILENAME" ]]; then
    ln -s "$DOWNLOAD_BASE/$FILENAME" "$OUTPUT_DIR/$FILENAME"
  fi
}

emit_completed_json() {
  local progress_text="下载已完成"
  local downloaded_path
  local files_json
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
    "Status": "completed",
    "OutputDir": os.path.abspath(sys.argv[1]),
    "DownloadedPath": os.path.abspath(sys.argv[2]),
    "ProgressText": sys.argv[3],
    "Executor": sys.argv[4],
    "Files": json.loads(sys.argv[5]),
}
print(json.dumps(payload, ensure_ascii=False))
PY
}

run_submit_action() {
  ensure_amuled_ready
  wait_for_remote
  touch_submit_marker
  submit_ed2k_link
  if ensure_download_visible_with_fallback "$ED2K_SUBMIT_CONFIRM_TIMEOUT_SECONDS"; then
    emit_nonterminal_json "running" "已提交到 aMule，等待开始下载"
    return
  fi
  emit_nonterminal_json "running" "已提交到 aMule，等待同步下载状态"
}

run_status_action() {
  ensure_amuled_ready
  if ! wait_for_remote; then
    emit_nonterminal_json "not_found" "等待 aMule 远程控制恢复"
    return
  fi

  local listing
  if ! listing="$(run_amulecmd "show dl")"; then
    emit_nonterminal_json "not_found" "等待 aMule 同步下载状态"
    return
  fi
  local entry
  entry="$(find_download_entry "$listing" || true)"
  if [[ -n "$entry" ]]; then
    mark_download_seen
    emit_nonterminal_json "running" "$(extract_progress_text "$entry")"
    return
  fi

  if has_downloaded_file; then
    mirror_downloaded_file
    emit_completed_json
    return
  fi

  if has_output_files; then
    emit_completed_json
    return
  fi

  if should_requeue_unseen_download; then
    if ensure_download_visible_with_fallback "$ED2K_SUBMIT_CONFIRM_TIMEOUT_SECONDS"; then
      listing="$(run_amulecmd "show dl" || true)"
      entry="$(find_download_entry "$listing" || true)"
      if [[ -n "$entry" ]]; then
        emit_nonterminal_json "running" "$(extract_progress_text "$entry")"
        return
      fi
    fi
    emit_nonterminal_json "running" "首次提交未出现在下载队列，已请求 aMule 重新接收链接"
    return
  fi

  emit_nonterminal_json "running" "等待 aMule 同步下载状态"
}

case "$(lower "$EXECUTOR_MODE")" in
  submit)
    run_submit_action
    ;;
  status)
    run_status_action
    ;;
  *)
    printf 'unsupported ED2K_EXECUTOR_MODE: %s\n' "$EXECUTOR_MODE" >&2
    exit 64
    ;;
esac

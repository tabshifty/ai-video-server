#!/usr/bin/env bash
set -Eeuo pipefail

# server.met 定时刷新执行器：从远端下载 aMule 服务器列表，校验魔数字节后原子替换
# ~/.aMule/server.met，再用 launchctl kickstart -k 重启 amuled 守护进程，
# 等待 EC 通道就绪后返回。Go worker 通过 ED2K_SERVERLIST_REFRESH_EXECUTABLE 调用本脚本，
# 仅固定“脚本输入输出契约”，aMule 远控/文件操作细节都留在这层脚本里
#（见 [[ED2K 服务器列表刷新归执行器脚本]] / [[ED2K 服务器列表刷新重启 amuled 生效]]）。
#
# 环境输入（Go 侧通过 os.Environ 注入，复用 aMule 远控参数）：
#   ED2K_SERVERLIST_URL            必填，server.met 下载地址
#   ED2K_SERVER_MET_PATH           可选，默认 $HOME/.aMule/server.met
#   AMULECMD_BIN                   可选，默认 amulecmd，用于重启后 EC 探活
#   AMULE_REMOTE_HOST              可选，默认 127.0.0.1
#   AMULE_REMOTE_PORT              可选，默认 4712
#   AMULE_REMOTE_PASSWORD          必填，aMule External Connections 密码
#   ED2K_AMULED_LAUNCHCTL_LABEL    可选，默认 com.aivideo.amuled
#   ED2K_SERVERLIST_DOWNLOAD_TIMEOUT_SECONDS  可选，下载超时，默认 120
#   ED2K_SERVERLIST_EC_WAIT_TIMEOUT_SECONDS   可选，EC 就绪等待，默认 120
#   ED2K_SERVERLIST_EC_POLL_SECONDS           可选，EC 轮询间隔，默认 2
#
# stdout 末行输出单个 JSON 结果对象（与 ed2k-amule-executor.sh 的末行 JSON 契约一致）：
#   {"server_met_path":"...","server_met_bytes":N,"restarted":bool,"executor":"launchctl/<label>"}
# 失败时非零退出 + stderr 中文诊断。
#
# 退出码分段（与 ed2k-amule-executor.sh 对齐）：
#   64 缺 ED2K_SERVERLIST_URL / 65 缺 AMULE_REMOTE_PASSWORD / 66 缺 curl 或 amulecmd
#   67 下载失败 / 68 server.met 校验失败(魔数或空文件) / 69 重启失败 / 70 EC 未就绪

SERVERLIST_URL="${ED2K_SERVERLIST_URL:-}"
SERVER_MET_PATH="${ED2K_SERVER_MET_PATH:-$HOME/.aMule/server.met}"
AMULECMD_BIN="${AMULECMD_BIN:-amulecmd}"
AMULE_REMOTE_HOST="${AMULE_REMOTE_HOST:-127.0.0.1}"
AMULE_REMOTE_PORT="${AMULE_REMOTE_PORT:-4712}"
AMULE_REMOTE_PASSWORD="${AMULE_REMOTE_PASSWORD:-}"
LAUNCHCTL_LABEL="${ED2K_AMULED_LAUNCHCTL_LABEL:-com.aivideo.amuled}"
DOWNLOAD_TIMEOUT="${ED2K_SERVERLIST_DOWNLOAD_TIMEOUT_SECONDS:-120}"
EC_WAIT_TIMEOUT="${ED2K_SERVERLIST_EC_WAIT_TIMEOUT_SECONDS:-120}"
EC_POLL_SECONDS="${ED2K_SERVERLIST_EC_POLL_SECONDS:-2}"
CURL_BIN="${CURL_BIN:-curl}"

if [[ -z "$SERVERLIST_URL" ]]; then
  printf 'ED2K_SERVERLIST_URL is required\n' >&2
  exit 64
fi
if [[ -z "$AMULE_REMOTE_PASSWORD" ]]; then
  printf 'AMULE_REMOTE_PASSWORD is required\n' >&2
  exit 65
fi
for bin in "$CURL_BIN" "$AMULECMD_BIN"; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    printf 'binary not found: %s\n' "$bin" >&2
    exit 66
  fi
done
if ! command -v launchctl >/dev/null 2>&1; then
  printf 'launchctl not found (not macOS?)\n' >&2
  exit 66
fi

# 复用 ed2k-amule-executor.sh 的 amulecmd 远控与 EC 就绪探活风格。
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

wait_for_remote() {
  local attempts=0
  local max=$(( EC_WAIT_TIMEOUT / EC_POLL_SECONDS ))
  (( max < 1 )) && max=1
  while (( attempts < max )); do
    if run_amulecmd "help" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep "$EC_POLL_SECONDS"
  done
  printf 'amule remote connection is not ready after %s seconds\n' "$EC_WAIT_TIMEOUT" >&2
  return 1
}

# 校验 server.met 魔数字节：第 1 字节必须 0xE0，第 2 字节版本（当前 0x09，放宽到 0x0n）。
# 用 od 读首两字节，避免依赖 xxd（macOS 默认无）。
validate_magic() {
  local path="$1"
  local magic
  magic="$(od -An -tx1 -N2 "$path" 2>/dev/null | tr -d '[:space:]')"
  case "$magic" in
    e00[0-9]) return 0 ;;
    *) return 1 ;;
  esac
}

# 取文件字节数（用于日志，反映列表是否真有内容）。
file_size() {
  local path="$1"
  local size
  size="$(stat -f%z "$path" 2>/dev/null || stat -c%s "$path" 2>/dev/null || echo 0)"
  printf '%s' "$size"
}

MET_DIR="$(dirname "$SERVER_MET_PATH")"
mkdir -p "$MET_DIR"
TMP_MET="$(mktemp -t ed2k-serverlist.XXXXXX)"
cleanup() { rm -f "$TMP_MET"; }
trap cleanup EXIT

# 1. 下载 server.met 到临时文件。--fail 让 4xx/5xx 走非零退出。
if ! "$CURL_BIN" --fail --silent --show-error --max-time "$DOWNLOAD_TIMEOUT" \
     --connect-timeout 15 -o "$TMP_MET" "$SERVERLIST_URL"; then
  printf 'download server.met failed: %s\n' "$SERVERLIST_URL" >&2
  exit 67
fi
if [[ ! -s "$TMP_MET" ]]; then
  printf 'downloaded server.met is empty\n' >&2
  exit 68
fi
if ! validate_magic "$TMP_MET"; then
  printf 'server.met magic byte invalid (want e009)\n' >&2
  exit 68
fi

# 2. 原子替换：先备份现有 server.met（便于回滚），再 mv 覆盖。
#    mv 同目录内是原子的；跨设备会失败，回退到同目录 staging + rename。
BACKUP_PATH=""
if [[ -f "$SERVER_MET_PATH" ]]; then
  BACKUP_PATH="$SERVER_MET_PATH.bak.$$"
  cp -p "$SERVER_MET_PATH" "$BACKUP_PATH"
fi
if ! mv -f "$TMP_MET" "$SERVER_MET_PATH" 2>/dev/null; then
  STAGE="$MET_DIR/.server.met.stage.$$"
  if cp -p "$TMP_MET" "$STAGE" && mv -f "$STAGE" "$SERVER_MET_PATH"; then
    :
  else
    printf 'atomic replace server.met failed\n' >&2
    rm -f "$STAGE"
    [[ -n "$BACKUP_PATH" ]] && mv -f "$BACKUP_PATH" "$SERVER_MET_PATH"
    exit 68
  fi
fi
trap - EXIT

# 3. 替换后再次校验落盘文件魔数（防 rename 异常）。校验失败则回滚备份，绝不留下坏 server.met。
if ! validate_magic "$SERVER_MET_PATH"; then
  printf 'post-replace server.met magic invalid\n' >&2
  [[ -n "$BACKUP_PATH" ]] && mv -f "$BACKUP_PATH" "$SERVER_MET_PATH"
  exit 68
fi
rm -f "$BACKUP_PATH" 2>/dev/null || true

SERVER_COUNT="$(file_size "$SERVER_MET_PATH")"

# 4. 重启 amuled：launchctl kickstart -k 强制重启 KeepAlive 服务，让它在启动时重读 server.met。
GUI_UID="$(id -u)"
TARGET="gui/${GUI_UID}/${LAUNCHCTL_LABEL}"
if ! launchctl kickstart -k "$TARGET" 2>/dev/null; then
  printf 'launchctl kickstart failed: %s\n' "$TARGET" >&2
  exit 69
fi

# 5. 等待 EC 通道就绪（复用 amulecmd help 探活，与下载执行器一致）。
if ! wait_for_remote; then
  exit 70
fi

# 6. stdout 末行 JSON 结果。
/usr/bin/python3 - "$SERVER_MET_PATH" "$SERVER_COUNT" "$LAUNCHCTL_LABEL" <<'PY'
import json
import os
import sys

payload = {
    "server_met_path": os.path.abspath(sys.argv[1]),
    "server_met_bytes": int(sys.argv[2]),
    "restarted": True,
    "executor": "launchctl/" + sys.argv[3],
}
print(json.dumps(payload, ensure_ascii=False))
PY

#!/usr/bin/env bash
# 家用部署机手动 rollback：把历史二进制复制回稳定运行入口 + 重启 launchd 服务。
# 如果 .env 里配置了 CODESIGN_IDENTITY（以及可选的 CODESIGN_KEYCHAIN），
# 会在切换前对目标二进制重新签名，让 rollback 也继续走同一份稳定代码身份。
# 不自动跑 migration down —— 依赖 ADR-0006 [[migration 前向兼容契约]]，
# 即旧 binary 必须能在当前（更新过）schema 上跑。
#
# 用法：
#   scripts/rollback.sh <commit-sha>
#
# 前置：
# - 已按 docs/家用部署机.md 完成一次性安装。
# - $DEPLOY_ROOT (默认 ~/deploy/ai-video-server) 下目录结构齐全。
# - binaries/ 下能找到 video-server-<sha>.bin（前缀匹配即可，传 10 位短 sha 足够）。
#   若已被 N=3 保留策略淘汰，需要先在 work 目录 checkout 对应 sha 并 go build 重建。

set -Eeuo pipefail

DEPLOY_ROOT="${DEPLOY_ROOT:-$HOME/deploy/ai-video-server}"
BIN_DIR="$DEPLOY_ROOT/binaries"
CURRENT="$DEPLOY_ROOT/current"
WORK="$DEPLOY_ROOT/work"
LOG="$HOME/Library/Logs/ai-video-server/deploy.log"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { printf '[%s] [rollback] %s\n' "$(date '+%F %T')" "$*" | tee -a "$LOG"; }

if [[ $# -lt 1 ]]; then
  printf 'usage: %s <commit-sha>\n' "$0" >&2
  exit 1
fi

TARGET_SHA="$1"
TARGET_SHORT="${TARGET_SHA:0:10}"

if [[ -f "$DEPLOY_ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$DEPLOY_ROOT/.env"
  set +a
fi

# 找到目标二进制
BIN_FILE=$(ls "$BIN_DIR"/video-server-"${TARGET_SHORT}"*.bin 2>/dev/null | head -1 || true)
if [[ -z "$BIN_FILE" ]]; then
  log "未找到 binaries/video-server-${TARGET_SHORT}*.bin"
  log "如果 sha 仍在仓库历史里，可以手动重建："
  log "  cd $WORK && git checkout ${TARGET_SHA} && go build -o $BIN_DIR/video-server-${TARGET_SHORT}.bin ."
  log "  然后重跑 $0 ${TARGET_SHA}"
  exit 1
fi

log "rollback to $BIN_FILE"
log "==> codesign target binary"
"$SCRIPT_DIR/sign-launchd-binary.sh" "$BIN_FILE"

# 稳定运行入口：launchd 永远只看 current/video-server。
cp "$BIN_FILE" "$CURRENT/video-server.new"
chmod 755 "$CURRENT/video-server.new"
mv -f "$CURRENT/video-server.new" "$CURRENT/video-server"

# 同步 work 树到目标 sha（便于后续 push 基于此点做反向 patch）
git --git-dir="$DEPLOY_ROOT/repo.git" --work-tree="$WORK" reset --hard "$TARGET_SHA" || \
  log "git reset 失败（可能 sha 不在 repo.git 历史），仅切换了 binary"

# 重启 launchd
launchctl kickstart -k "gui/$(id -u)/com.aivideo.server" || true
launchctl kickstart -k "gui/$(id -u)/com.aivideo.worker" || true

# 探活
sleep 2
if curl -fsS --max-time 5 --retry 10 --retry-delay 1 http://127.0.0.1:8080/healthz >/dev/null; then
  log "/healthz OK — rollback 成功"
else
  log "/healthz 失败 —— 旧 binary 可能与当前 schema 不兼容（违反 [[migration 前向兼容契约]]）"
  log "查 ~/Library/Logs/ai-video-server/server.log 看具体错误"
  exit 2
fi

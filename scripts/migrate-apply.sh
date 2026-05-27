#!/usr/bin/env bash
# 应用 migrations/*.up.sql，幂等。
#
# 行为与 dev-up.sh apply_migrations() 一致：
# - schema_migrations 表不存在则建。
# - 逐个 up.sql 检查 schema_migrations.version 是否已应用，已应用则跳过。
# - 未应用的 SQL 执行成功后写入 version 记录；任一文件失败立即退出非 0。
#
# 用法：
#   POSTGRES_CONTAINER=video-server-postgres ./scripts/migrate-apply.sh
#   或
#   POSTGRES_DSN=postgres://video:video@127.0.0.1:5432/video_server ./scripts/migrate-apply.sh
#
# 二选一：优先用 POSTGRES_CONTAINER（dev 模式 docker exec），fallback 用 POSTGRES_DSN 走本机 psql。
# dev-up.sh 与 [[家用部署机]] post-receive hook 共用本脚本。
# POSTGRES_DSN 指向非本机地址时，必须额外设置 ALLOW_REMOTE_MIGRATIONS=1。

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="$ROOT_DIR/migrations"

log() { printf '[migrate] %s\n' "$*" >&2; }

postgres_dsn_host() {
  python3 - "$POSTGRES_DSN" <<'PY'
import sys
from urllib.parse import urlparse

dsn = sys.argv[1]
parsed = urlparse(dsn)
print(parsed.hostname or "")
PY
}

is_local_postgres_host() {
  local host="$1"
  case "$host" in
  "" | localhost | 127.* | ::1)
    return 0
    ;;
  *)
    return 1
    ;;
  esac
}

guard_remote_migrations() {
  local host
  host="$(postgres_dsn_host)"
  if is_local_postgres_host "$host"; then
    return 0
  fi

  if [[ "${ALLOW_REMOTE_MIGRATIONS:-}" != "1" ]]; then
    log "拒绝对远程 Postgres 执行 migration。"
    log "当前 POSTGRES_DSN host: $host"
    log "开发遇到数据库改动时，先用本地 Docker DB 验证 migration，提交后由家用部署机部署流程执行。"
    log "若确实必须手动改远程库，请显式设置 ALLOW_REMOTE_MIGRATIONS=1 后重跑。"
    exit 1
  fi

  log "警告：ALLOW_REMOTE_MIGRATIONS=1 已开启，正在对远程 Postgres 执行 migration。"
  log "目标 host: $host"
}

if [[ -n "${POSTGRES_CONTAINER:-}" ]]; then
  PSQL=(docker exec -i "$POSTGRES_CONTAINER" psql -U video -d video_server -v ON_ERROR_STOP=1)
elif [[ -n "${POSTGRES_DSN:-}" ]]; then
  guard_remote_migrations
  if ! command -v psql >/dev/null 2>&1; then
    log "psql 不在 PATH，且未设置 POSTGRES_CONTAINER。请安装 postgresql client 或设置容器名。"
    exit 1
  fi
  PSQL=(psql "$POSTGRES_DSN" -v ON_ERROR_STOP=1)
else
  log "必须设置 POSTGRES_CONTAINER（docker exec 模式）或 POSTGRES_DSN（本机 psql 模式）"
  exit 1
fi

"${PSQL[@]}" >/dev/null <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

applied_any=0
for file in "$MIGRATIONS_DIR"/*.up.sql; do
  [[ -f "$file" ]] || continue
  version="$(basename "$file")"

  already="$("${PSQL[@]}" -tA -c "SELECT 1 FROM schema_migrations WHERE version='${version}' LIMIT 1;")"
  if [[ "$already" == "1" ]]; then
    continue
  fi

  log "applying $version"
  "${PSQL[@]}" >/dev/null <"$file"
  "${PSQL[@]}" >/dev/null <<SQL
INSERT INTO schema_migrations(version) VALUES ('${version}')
ON CONFLICT(version) DO NOTHING;
SQL
  applied_any=1
done

if [[ "$applied_any" == "0" ]]; then
  log "all migrations already applied"
else
  log "done"
fi

# Telegram 视频采集部署与运维

本文说明如何在 OrbStack 或兼容 Docker Compose 的环境中运行 API、转码 worker 和 Telegram 采集器。采集器使用个人 Telegram 账号的 MTProto session，只采集管理员明确添加且该账号有权访问的群组或频道，并将视频固定导入短视频库。

## 部署前检查

- 安装 OrbStack 或 Docker Compose v2，并确认 `docker compose version` 可用。
- 在 <https://my.telegram.org> 创建应用，取得 API ID 和 API hash。
- 准备一个有权读取目标群组历史消息的个人账号。不要使用他人的 session，也不要采集、保存或传播无权处理的内容。
- 确认数据库中已有接收导入视频的用户；登录接口返回的 `data.user_id` 可直接作为 `TELEGRAM_IMPORT_USER_ID`。
- 同一份 session 只能由一个 `telegram-ingestor` 进程使用。不要同时在宿主机和容器中启动采集器。

## 配置目录和环境变量

从仓库根目录准备配置：

```bash
cp .env.example .env
mkdir -p ./storage ./tmp/uploads ./telegram-session
chmod 700 ./telegram-session
```

生产部署应把以下宿主路径改成绝对路径，并保证 UID/GID `10001:10001` 对三个目录有读写权限。Linux 可执行 `sudo chown -R 10001:10001 <目录>`；OrbStack for macOS 通常会自动映射当前用户的 bind mount 权限。

```dotenv
VIDEO_STORAGE_HOST_PATH=/绝对路径/storage
VIDEO_UPLOAD_TEMP_HOST_PATH=/绝对路径/tmp/uploads
TELEGRAM_SESSION_HOST_PATH=/绝对路径/telegram-session
```

编辑 `.env`，至少填写：

```dotenv
JWT_SECRET=<高强度随机值>
PASSWORD_VAULT_KEY=<部署后保持不变的高强度随机值>
TELEGRAM_API_ID=<my.telegram.org 提供的整数>
TELEGRAM_API_HASH=<my.telegram.org 提供的值>
TELEGRAM_PHONE=<带国家区号的手机号>
TELEGRAM_IMPORT_USER_ID=<现有用户 UUID>
```

Compose 会把三个服务的持久媒体和上传暂存目录统一映射为 `/data/storage` 与 `/data/tmp/uploads`。`TELEGRAM_SESSION_PATH` 固定为 `/data/telegram-session/account.session`，该目录只挂载到以非 root 用户运行的采集器。

容器内 API 固定监听 `:8080`。需要修改宿主机端口时设置 `HTTP_PORT`，例如 `HTTP_PORT=18080`，不要改 `HTTP_ADDR`。

不要提交 `.env` 或 session 文件。备份 session 前先停止采集器，并按账号凭据的安全等级加密保存备份。

## 应用 migration

先启动并等待 PostgreSQL、Redis：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d postgres redis
until docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres pg_isready -U video -d video_server; do sleep 2; done
POSTGRES_CONTAINER=video_server_postgres ./scripts/migrate-apply.sh
```

脚本会按顺序应用尚未登记的 `*.up.sql`，其中 `0038_telegram_video_ingestion.up.sql` 创建 Telegram 来源和媒体状态表。应用完成后可核对：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -c "SELECT version, applied_at FROM schema_migrations WHERE version = '0038_telegram_video_ingestion.up.sql';"
```

## 一次性登录并启动

构建统一镜像后，以交互方式执行一次登录。命令会要求输入 Telegram 验证码；账号启用两步验证时还会要求输入密码。验证码和密码不会写入 `.env`。

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml build api
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml run --rm --no-deps \
  telegram-ingestor /app/telegram-ingestor -mode login
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml run --rm --no-deps \
  --entrypoint /usr/bin/test telegram-ingestor -s /data/telegram-session/account.session
```

确认 session 已生成后启动三个应用服务：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d api worker telegram-ingestor
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml ps
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml logs -f telegram-ingestor
```

如果 Telegram 撤销授权、session 损坏或手机号变更，先停止采集器，备份并移走旧 session，再重新执行登录命令。不要在采集器运行时覆盖 session。

## 登录管理 API

以下示例依赖 `jq`。先取得管理员 JWT；也可以直接使用已有管理员 access token。

```bash
API_BASE=http://127.0.0.1:8080/api/v1
LOGIN_RESPONSE=$(curl -fsS -X POST "$API_BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"<管理员用户名>","password":"<管理员密码>"}')
ADMIN_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | jq -er '.data.access_token')
printf '%s' "$LOGIN_RESPONSE" | jq '.data | {user_id}'
```

若 `.env` 中的 `TELEGRAM_IMPORT_USER_ID` 刚刚修改，需要重建采集器容器以读取新值：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d --force-recreate telegram-ingestor
```

## 添加群组并查看进度

`chat_ref` 可填写 `@username`、Telegram 链接、邀请链接或数字 chat ID。个人账号必须已经有权访问该来源；新增接口只保存来源并投递控制任务，实际解析由采集器异步执行。

```bash
SOURCE_RESPONSE=$(curl -fsS -X POST "$API_BASE/admin/telegram/sources" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"chat_ref":"@example_channel"}')
SOURCE_ID=$(printf '%s' "$SOURCE_RESPONSE" | jq -er '.data.id')
printf '%s' "$SOURCE_RESPONSE" | jq
```

查看全部来源和单个来源进度：

```bash
curl -fsS "$API_BASE/admin/telegram/sources" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
curl -fsS "$API_BASE/admin/telegram/sources/$SOURCE_ID/progress" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

来源通常依次进入 `pending`、`backfilling`、`live`。进度里的 `counts` 按媒体处理状态聚合；`imported` 和 `duplicate` 都表示不需要再次下载。`error` 或 `failed` 应结合 `last_error` 和采集器日志排查。

## 暂停、恢复和重新回填

暂停不会清空历史游标；恢复会从现有游标继续：

```bash
curl -fsS -X POST "$API_BASE/admin/telegram/sources/$SOURCE_ID/pause" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
curl -fsS -X POST "$API_BASE/admin/telegram/sources/$SOURCE_ID/resume" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

重新回填会清零该来源的历史游标和完成时间，再从最新消息向旧消息扫描。已有消息、document ID 和文件哈希仍会去重：

```bash
curl -fsS -X POST "$API_BASE/admin/telegram/sources/$SOURCE_ID/backfill" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

采集器重启后会重新投递已启用来源，并周期性补偿超时下载和待投递转码。不要通过清空 Redis 来重置回填进度，PostgreSQL 才是来源、游标和媒体状态的权威。

## 恢复或跳过失败记录

先暂停来源并查看失败记录，不要直接删除已经导入的视频或 `/data/storage` 下的文件：

```bash
curl -fsS -X POST "$API_BASE/admin/telegram/sources/$SOURCE_ID/pause" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -c "SELECT id, message_id, attempts, last_error, next_retry_at FROM telegram_media WHERE source_id = '$SOURCE_ID' AND processing_status = 'failed' ORDER BY updated_at;"
```

修复网络、权限或磁盘问题后，可把该来源的未导入失败记录恢复为可补偿状态。把 `updated_at` 调早是为了让采集器下一轮 reconcile 立即重新入队：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 -c "UPDATE telegram_media SET processing_status = 'queued', attempts = 0, next_retry_at = NULL, last_error = NULL, updated_at = NOW() - INTERVAL '31 minutes' WHERE source_id = '$SOURCE_ID' AND processing_status = 'failed' AND video_id IS NULL;"
curl -fsS -X POST "$API_BASE/admin/telegram/sources/$SOURCE_ID/resume" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

若某条消息确认不应重试，可将它标记为 `skipped`，保留消息幂等记录而不删除任何媒体文件：

```bash
FAILED_MEDIA_ID=00000000-0000-0000-0000-000000000000
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 -c "UPDATE telegram_media SET processing_status = 'skipped', next_retry_at = NULL, last_error = 'operator skipped', updated_at = NOW() WHERE id = '$FAILED_MEDIA_ID' AND processing_status = 'failed' AND video_id IS NULL;"
```

Redis 中的任务只负责唤醒处理，重复任务会由数据库 claim 和幂等键收敛。除非已备份并明确评估其它业务队列，不要执行 `FLUSHDB` 或删除整个 Redis 数据卷。

## 回滚 0038 migration

回滚会删除 `telegram_sources` 和 `telegram_media` 中的全部采集状态，但不会删除 `videos` 表记录和已有视频文件。先备份数据库、停止三个应用服务，并确认不再需要这些状态：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml stop telegram-ingestor worker api
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  pg_dump -U video -d video_server -Fc > video_server_before_0038_rollback.dump
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 < migrations/0038_telegram_video_ingestion.down.sql
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version = '0038_telegram_video_ingestion.up.sql';"
```

确认回滚后再切回不依赖 0038 的服务端提交。不要只回滚 migration 而继续运行含 Telegram 仓储查询的新二进制。

## 限流与安全注意事项

- Telegram 可能返回 FloodWait。采集器会按服务端给出的等待时间退避；不要为了追赶进度反复重启、重复登录或盲目提高 `TELEGRAM_DOWNLOAD_CONCURRENCY`。
- 初次部署建议保持下载并发 `1`、采集总任务数 `2`。只有在磁盘、带宽和账号限流均稳定后才逐步调整，且总任务数不得小于下载并发。
- session 等同个人账号长期凭据。限制目录权限、备份访问者和日志读取权限；怀疑泄露时应在 Telegram 客户端撤销会话并重新登录。
- 只添加已获授权的来源，并遵守群组规则、内容版权、隐私和当地法律。技术上的可访问不代表拥有保存或传播权限。

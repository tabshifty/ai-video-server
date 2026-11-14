# Telegram 论坛下载接入与回滚

## 交付内容与前提

Go 后端维护论坛资源读取、下载操作及回执；独立 Hermes 插件位于 `integrations/hermes-forum-download/`，通过官方平台工厂覆盖保留原 Telegram 适配器并增加下载回调，不修改 Hermes 核心、不启动另一个 Bot 消费者。详细行为见 [首版设计](telegram-forum-download-design.md)。

需要现有 Hermes Telegram 平台、论坛采集脚本、项目 API 和云下载 worker 正常运行。插件使用 Hermes 自带的 Python、python-telegram-bot 和 YAML 依赖，不引入对话模型调用。

## 后端配置

在家用部署机配置下列环境变量，然后按原有部署流程应用迁移 `0048`、构建并重启 server/worker：

```dotenv
CLOUD_DOWNLOAD_ENABLED=true
TELEGRAM_FORUM_DOWNLOAD_TOKEN=独立的高强度随机令牌
TELEGRAM_FORUM_OWNER_ID=你的正数Telegram用户ID
```

令牌与 `HERMES_API_TOKEN` 必须不同。令牌留空时新接口拒绝访问，原采集接口继续使用原令牌。用户 ID 必须为本人数字 ID，不使用用户名、群组 ID 或频道 ID。后端和 Hermes 的本人 ID 必须一致。

## 安装 Hermes 扩展

从仓库根目录先执行兼容检查，再安装：

```bash
python3 integrations/hermes-forum-download/install.py --check
python3 integrations/hermes-forum-download/install.py
```

默认目标是 `~/.hermes`；可用 `--home /绝对路径` 指定测试环境。安装器复制三个插件文件到 `plugins/forum-download/`，为监控脚本应用版本化补丁，并保留 `scripts/watch_sehuatang_forum95.py.before-forum-download` 作为回滚备份。重复安装不会重复应用补丁；不兼容时拒绝修改。安装器不会写入凭据或重启进程。

在 Hermes 的 `.env` 中配置与后端相同的 `TELEGRAM_FORUM_DOWNLOAD_TOKEN`。在现有 `config.yaml` 的插件配置中合并以下内容，保留其它插件和设置：

```yaml
plugins:
  enabled:
    # 保留原有已启用插件，并增加以下项。
    - forum-download
  entries:
    forum-download:
      settings:
        api_base: http://家用部署机地址:8080
        owner_id: 你的正数Telegram用户ID
```

`api_base` 使用 Hermes 能访问的实际后端地址，不带 `/api/v1`。配置仅适用于受信任家庭局域网；跨越该网络边界使用 HTTPS。安装完成且后端配置生效后，按现有 Hermes 运维方式重启 gateway，让唯一 Bot 接收者加载插件。

监控脚本仅在插件已启用且帖子含 ED2K 时切换到 `hermes forum-download-send`；无 ED2K 的帖子继续用原 `hermes send`。带按钮发送失败会留在监控的通知队列重试，不降级为成功但缺少按钮的推送。历史推送不补按钮。

专用发送命令将单张预览图和说明合并，短正文原样附图，超长正文使用标题、帖子链接和 ED2K 数量摘要。「查看下载资源」按钮打开独立回复面板，分页和下载确认不会替换原图文；附件从帖子链接查看。无图或图片发送失败时回退完整纯文字。更新已有安装前，另行备份 `plugins/forum-download/`，重复运行安装器后重启 gateway；恢复旧插件目录并重启即可回滚本次展示调整，保留 SQLite 状态。

## 接口约定

基础路径为 `/api/v1/integrations/telegram-forum`，所有请求同时要求：

- `Authorization: Bearer <TELEGRAM_FORUM_DOWNLOAD_TOKEN>`。
- `X-Telegram-User-ID: <本人ID>`；操作请求体的 `actor_id` 也必须一致。

| 接口 | 用途 |
|---|---|
| `GET /posts?tid=帖子tid` | 为当前来源 `sehuatang` 的新推送定位已入库帖子 |
| `GET /posts/:id` | 读取完整 ED2K 资源、当前任务状态、资源集合指纹 |
| `POST /operations` | 提交固定事件 ID 对应的显式资源选择 |
| `GET /receipts` | 领取有限租约的待发送回执 |
| `POST /receipts/:id/ack` | 使用本次租约 token 确认 Telegram 已接收回执 |

操作字段为 `event_id`、`actor_id`、`chat_id`、`message_id`、可空 `thread_id`、`post_id`、`fingerprint`、`resource_ids`、`all` 和 `retry_attempts`。`all=true` 与指定资源 ID 互斥，单次 1 至 50 条；`retry_attempts` 以资源 UUID 为键，值为用户确认的旧任务重试次数，防止确认后新发生的失败被静默重试。同一事件必须重放完全相同的请求，否则返回冲突。

操作、逐条结果、云下载任务和受理回执在同一数据库事务中保存。事务提交后唤醒现有队列；唤醒失败由原云下载 reconciler 修复，不把已受理任务误标为失败。资源始终从数据库读取，按钮不携带任意下载链接。

Hermes 在 `plugin-data/forum-download/state.sqlite3` 保存按钮与消息绑定、待提交操作和已发送回执标识；备份该文件以保留旧按钮及重启恢复能力。回执发送成功才向后端确认，确认失败重新领取后可凭本地标识避免再次发送。Telegram 已接收但进程尚未记录即崩溃的窗口仍可能出现重复通知，以任务编号识别；通知补发不会触发下载重试。

## 验证

离线验证不会发送 Telegram 消息或创建真实 115 下载：

```bash
go test -race ./internal/forumdownload ./internal/config ./internal/services -run 'TestIntegration|TestSelect|TestResourceFingerprint|TestTelegramForum|TestSubmit|TestReconcile'
go vet ./...
go build ./...
~/.hermes/hermes-agent/venv/bin/python -m unittest discover -s integrations/hermes-forum-download -p 'test_*.py' -v
```

PostgreSQL 集成测试需要本地隔离数据库，通过 `TELEGRAM_FORUM_TEST_DATABASE_URL` 指定仅 loopback 的连接字符串；测试在唯一临时 schema 中应用真实迁移，结束后删除该 schema。不得把它指向生产库。可通过 `HERMES_RUNTIME_PATH` 指定另一份 Hermes 运行时，验证真实插件发现、原 sender 保留和回调注册。

上线后人工验收：单图短文合并、长文摘要及帖子链接完整、入口按钮打开独立回复面板、面板内单链接直接下载、多链接分页选择与确认全下、超过 50 条分批、网站已存在任务跳过、非本人点击拒绝、失败后显式重试、本地落盘后回执、重启后待提交/待通知恢复。使用用户实际要下载的资源验收；程序测试不代替真实站点、115 登录状态与 Telegram 网络验收。

## 回滚

1. 暂停论坛推送调度，从 `plugins.enabled` 移除 `forum-download`，保留其它插件；恢复监控脚本备份，再重启 Hermes。
2. 保留插件 SQLite 和后端操作表，便于恢复已受理任务与待发回执；如需降级 Go，先清空新增下载令牌并部署旧二进制。迁移为加表，旧程序可以在新 schema 继续运行。
3. 若明确放弃所有新入口的操作及回执历史，可执行 `0048` 的 down migration；它只删除新关联表，保留 `cloud_download_tasks` 和本地文件。不要在新代码仍接受请求时删除这些表。

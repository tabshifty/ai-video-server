# Hermes 负责论坛采集，项目后端提供按 tid 查重的两阶段入库接口

> 状态：已接受（2026-08-04）

## 背景

本机 Hermes 现有“色花堂综合讨论区新帖监控”每 15 分钟读取 `fid=95` 最新主题，排除 `fid=143`，再对未见帖子检查非图片附件、ED2K 和权限状态。任务由固定脚本执行，使用本地 `seen.json` 保存已见 `tid`，并按检查结果决定是否向 Telegram 推送。

项目需要持久保存这批结构化采集结果，并让 Hermes 使用项目接口完成查重。目标不是把调度或抓取迁入 Go 服务，也不是恢复已经退役的 ED2K 下载工作台。首期只建设数据库和 Hermes 机器接口，不建设管理端、手机端或 TV 端展示。

## 决策

### 职责边界

- Hermes 负责定时调度、站点登录态、版块发现、正文抓取、内容提取、失败重试、筛选判断和 Telegram 投递。
- 项目后端负责鉴权、请求校验、按 `source + tid` 原子查重、结构化结果持久化和 30 天保留策略。
- 每个首次发现的正常帖子都先登记并最终保存检查结果；`excluded` 只表示 Hermes 静默过滤，不能成为后端丢弃记录的理由。
- 后端不访问色花堂，不接收或保存原始 HTML/完整正文，不安排抓取重试，也不记录 Telegram 是否投递成功。
- 检查结果成功写入数据库后，Hermes 才能执行后续推送。接口不可用时，Hermes 不标记、不推送，并按自己的策略在后续运行中重试。

### 领域模型

使用论坛帖子级通用模型，不建立万能 Hermes JSON 数据桶，也不把表名锁死为色花堂：

#### `collected_forum_posts`

| 字段 | SQL 类型 | 语义 |
|---|---|---|
| `id` | `UUID` | 项目内部主键，默认 `gen_random_uuid()` |
| `source` | `VARCHAR(64)` | 来源代码；当前固定为 `sehuatang` |
| `board_key` | `VARCHAR(128)` | 来源内版块代码；当前为 `95`，不参与查重 |
| `external_post_id` | `VARCHAR(128)` | 来源帖子标识；当前对应 Discuz `tid` |
| `title`、`url` | `TEXT` | 首次发现时的帖子标题和绝对 URL；`dedupe_only` 可为空 |
| `inspection_status` | `VARCHAR(24)` | `dedupe_only`、`pending`、`inspected`、`restricted` 或 `failed` |
| `filter_decision` | `VARCHAR(16)` | Hermes 报告的 `included` 或 `excluded`；未完成记录为空 |
| `filter_reasons` | `JSONB` | Hermes 定义的原因字符串数组，后端不解释具体值 |
| `fetch_method` | `VARCHAR(64)` | Hermes 报告的正文读取方式，例如 `chrome-cdp` |
| `error_summary` | `TEXT` | Hermes 报告的受限或失败摘要，不保存敏感凭证 |
| `result_fingerprint` | `CHAR(64)` | 后端对规范化检查结果计算的 SHA-256 指纹 |
| `observed_at` | `TIMESTAMPTZ` | Hermes 实际发现帖子的时间；历史基线可为空 |
| `created_at` | `TIMESTAMPTZ` | 服务端首次入库时间，也是 30 天保留期起点 |
| `inspected_at` | `TIMESTAMPTZ` | 服务端成功接收最终检查结果的时间 |

唯一约束为 `UNIQUE (source, external_post_id)`。在当前来源内，这等价于只按 `tid` 查重；标题、URL、附件和 ED2K 均不参与重复判断。`board_key` 只是首次采集快照，即使帖子后来移动版块，也不会改变已经完成的记录。

记录没有 `updated_at`，因为它不是持续同步对象。重复发现一个已完成 `tid` 时不刷新标题、URL 或时间。

数据库约束和索引：

- 主键为 `id`，查重唯一约束为 `UNIQUE (source, external_post_id)`。
- `inspection_status` 和非空 `filter_decision` 使用 `CHECK` 限定上述值；`filter_reasons` 默认 `[]`，并约束为 JSON 数组。
- `dedupe_only` 允许标题、URL、筛选结论和检查时间为空；`pending` 必须有标题与 URL 且没有最终结论；三个最终状态必须有标题、URL、筛选结论、结果指纹和 `inspected_at`。
- 为 `created_at` 建普通 B-tree 索引，服务 30 天范围删除；首期没有按状态读取接口，不提前增加其它查询索引。

#### `collected_forum_post_resources`

| 字段 | SQL 类型 | 语义 |
|---|---|---|
| `id` | `UUID` | 资源行主键，默认 `gen_random_uuid()` |
| `post_id` | `UUID` | 所属帖子，外键删除时级联删除 |
| `kind` | `VARCHAR(16)` | `attachment` 或 `ed2k` |
| `value` | `TEXT` | Hermes 提交的完整原始链接 |
| `position` | `INTEGER` | 同一资源类型数组内从零开始的提交顺序 |

一条帖子可以同时拥有零到多条附件和零到多条 ED2K。资源不做跨帖子或帖子内业务查重，不承担独立生命周期。后端只裁剪外围空白并校验类型、协议、数量和长度，不改写 URL、文件名编码、大小写或 ED2K 内容。

资源表使用 `FOREIGN KEY (post_id) REFERENCES collected_forum_posts(id) ON DELETE CASCADE`，以 `CHECK` 限定 `kind` 和非负 `position`，并以 `UNIQUE (post_id, kind, position)` 保证每种资源的顺序位置唯一。`value` 不建立唯一约束，因此相同链接可以按 Hermes 提交结果原样出现。

`restricted` 和 `failed` 允许携带已经提取到的部分资源；状态表达检查不完整，不能成为丢弃已确认事实的理由。

### 状态与不可变约束

```text
dedupe_only                  历史查重基线，不进入正文检查

pending -> inspected         正文检查成功
        -> restricted        Hermes 最终报告权限或验证受限
        -> failed            Hermes 最终报告检查失败
```

- 正常发现只创建 `pending`；历史导入只创建 `dedupe_only`。
- `pending` 不得包含筛选结论或资源。
- `inspected`、`restricted`、`failed` 是最终状态，必须包含 `filter_decision` 和服务端 `inspected_at`，并可包含资源。
- 最终结果不可覆盖。完全相同的请求重复提交时，根据 `result_fingerprint` 幂等返回成功；已完成记录收到不同结果时返回 `409 Conflict`。
- `pending` 是重复过滤的唯一例外。发现接口遇到 `pending` 时返回其事实状态和内部 ID，由 Hermes 决定是否继续检查；后端不替 Hermes 制定重试策略。

### Hermes 机器鉴权

- 服务端通过 `HERMES_API_TOKEN` 配置独立高强度共享密钥。
- Hermes 使用 `Authorization: Bearer <token>` 调用独立中间件；该 Token 不解析为用户 JWT，也不赋予管理员身份。
- 中间件使用常量时间比较，日志不得记录 `Authorization` 请求头或真实 Token。
- 首期允许在受信任家庭局域网内通过现有 HTTP 服务调用，但不得暴露公网入口。若接口以后离开受信任局域网，HTTPS 是前置条件。
- Hermes 不得直连 PostgreSQL。

### 发现与查重接口

```http
POST /api/v1/integrations/hermes/forum-posts/discover
Authorization: Bearer <HERMES_API_TOKEN>
Content-Type: application/json
```

正常发现请求：

```json
{
  "mode": "normal",
  "source": "sehuatang",
  "board_key": "95",
  "observed_at": "2026-08-04T10:30:00+08:00",
  "posts": [
    {
      "tid": "3670055",
      "title": "帖子标题",
      "url": "https://sehuatang.org/forum.php?mod=viewthread&tid=3670055"
    }
  ]
}
```

历史基线请求使用 `mode=dedupe_only`，每项只要求 `tid`。单批最多 100 项。后端先校验整批，再在一个事务内执行过期清理、幂等插入和结果查询；任一项非法时整批返回 `400`，不允许部分成功。

响应保持请求顺序，并为每项返回：

```json
{
  "tid": "3670055",
  "id": "4ae4c2f1-ff29-4b7a-a181-8a5c2336ad6f",
  "disposition": "created",
  "inspection_status": "pending"
}
```

`disposition` 语义：

- `created`：首次插入；正常模式创建 `pending`，历史导入模式创建 `dedupe_only`。
- `pending`：该 `tid` 已登记但未收到最终结果；Hermes 可自行决定继续检查。
- `duplicate`：该 `tid` 已完成或属于 `dedupe_only`，Hermes 直接过滤。

### 检查结果接口

```http
PUT /api/v1/integrations/hermes/forum-posts/{id}/inspection
Authorization: Bearer <HERMES_API_TOKEN>
Content-Type: application/json
```

请求示例：

```json
{
  "status": "inspected",
  "filter_decision": "included",
  "filter_reasons": ["attachment", "ed2k"],
  "fetch_method": "chrome-cdp",
  "error_summary": "",
  "attachments": [
    "https://sehuatang.org/forum.php?mod=attachment&aid=123"
  ],
  "ed2k_links": [
    "ed2k://|file|example.zip|1024|0123456789ABCDEF0123456789ABCDEF|/"
  ]
}
```

帖子最终状态、筛选快照和全部资源在同一事务中写入。数组顺序映射到资源 `position`；两个数组都允许多项，也允许在 `restricted`、`failed` 时携带部分结果。

### HTTP 结果

新机器接口使用真实 HTTP 状态码，同时保留项目的 `{code, msg, data}` JSON 包装：

| HTTP 状态 | 语义 |
|---|---|
| `200` | 有效请求成功，包括重复发现和相同结果的幂等重放 |
| `400` | 字段、状态、协议或整批数据非法 |
| `401` | Hermes Token 缺失或错误 |
| `404` | 检查目标不存在，或已按保留策略清除 |
| `409` | 已完成记录收到不同检查结果 |
| `413` | 请求体超过服务端限制 |
| `500` | 数据库或服务内部错误 |

### 30 天保留与历史切换

- 每次有效 `discover` 请求在同一数据库事务中删除 `created_at < NOW() - INTERVAL '30 days'` 的帖子；资源通过外键级联删除。
- 重复发现不刷新 `created_at`。因此系统只保证首次入库后 30 天内按 `tid` 查重，过期 `tid` 若再次被来源返回，会重新视为新帖。
- 不增加 Go 后台定时器、Asynq 任务或 PostgreSQL 扩展。Hermes 停止调用时允许过期数据暂时保留，下次发现请求补做清理。
- 上线前由一次性导入命令读取 `~/.hermes/state/sehuatang_forum95_seen.json`，使用同一发现接口的 `dedupe_only` 模式按每批 100 项导入。核对输入数、创建数和重复数后，数据库取代 `seen.json` 成为查重权威。
- 当前基线约 1391 个 `tid`。该数量是切换时的运行数据，不写入 migration 或仓库静态文件。

## 考虑过的替代方案

- **后端直接调度和抓取色花堂**：会把站点登录态、CDP、Cookie 与降级链引入项目服务，违反已确认的 Hermes 运行边界。
- **Hermes 直连 PostgreSQL**：泄露数据库凭证并绕过校验、事务和未来兼容层，拒绝采用。
- **先查询是否存在，再插入结果**：存在检查与写入之间的竞态，也增加一次网络往返；改用唯一约束和提交即查重。
- **抓完正文后一次性写入**：Hermes 必须在查重前重复打开版块内所有帖子；两阶段协议保留现有“先按 tid 过滤，再做昂贵检查”的效率。
- **数据库中存在即一律过滤**：会让“已登记后、补结果前”中断的记录永久停在 `pending`；因此只为 `pending` 返回可恢复状态。
- **按 ED2K、附件、标题或正文查重**：超出当前只按 `tid` 过滤的需求，并可能隐藏新帖；资源只作为帖子子项保存。
- **永久保留记录**：可提供终身查重，但数据持续增长；当前来源只读取最新发帖，决定接受 30 天窗口。
- **保存原始 HTML 或完整正文**：当前查重和未来基础展示不需要，会增加不可信内容、敏感数据与存储体积。
- **复用管理员 JWT**：长期机器任务不应持有用户登录令牌，也不应依赖 Redis 登录会话；使用独立机器 Token。

## 后果与风险

- 方案在现有 Go/Gin、PostgreSQL 和局域网部署上可行，只需新增前向兼容表、仓储/服务/处理器、独立鉴权配置及 Hermes 脚本调用。
- 项目 API 短暂不可用会延迟 Hermes 推送，这是“数据库先成功”一致性边界的明确代价。
- `pending` 是否继续检查、失败何时重试以及何时最终提交 `failed`，全部由 Hermes 决定；后端可能长期保留未完成记录，直到 30 天清理。
- 局域网 HTTP 会明文传输机器 Token；当前只在受信任家庭网络接受该风险，不能把此决策外推到公网环境。
- 首期没有读取或展示接口。未来展示必须复用本模型另行设计分页、权限和筛选，不得为展示需求反向改变首次采集快照。
- 本设计不恢复 `ed2k_download_tasks`，也不自动创建 115 云下载、本地暂存、媒体入库或转码任务。

## 关联

- `CONTEXT.md` 中 [[论坛采集帖子]]、[[Hermes 两阶段采集协议]]、[[Hermes tid 查重窗口]] 与 [[Hermes 机器接口边界]]
- `migrations/0035_retire_ed2k_download.up.sql`
- `docs/adr/0006-migration-forward-compatibility.md`
- 本机 Hermes：`~/.hermes/scripts/watch_sehuatang_forum95.py`
- 本机历史查重状态：`~/.hermes/state/sehuatang_forum95_seen.json`

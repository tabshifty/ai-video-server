# 日本 AV 目录数据与同步契约

> 状态：提议中。本文把领域模型收敛为可实施的数据、适配器和管理接口边界，但不授权任何外部来源。

## 首期分层

```text
来源适配器 -> 规范化来源事实 -> 匹配/审核 -> 目录合并视图 -> 本地视频关联
                  |                  |              |
                快照历史           合并事件        人工覆盖

采集源 -> 采集运行 -> 页事务 -> 持久化游标
```

来源适配器只负责获取和规范化，不直接写数据库；服务层负责幂等、匹配和事务；管理端负责批准来源、处理冲突和人工覆盖。现有 AV 单片刮削可复用解析函数，但不能绕过这些边界直接写 `videos.metadata`。

## 核心表提案

首期 migration 暂命名 `0040_av_catalog_core`。全部使用新增表和可空外键，不修改既有表字段或约束，以满足 migration 前向兼容契约。

### `av_catalog_sources`

| 字段 | 语义 |
|---|---|
| `id UUID`、`code VARCHAR(64) UNIQUE` | 内部身份和稳定来源代码 |
| `name TEXT`、`base_url TEXT` | 展示名称和正式入口 |
| `access_mode VARCHAR(16)` | `api`、`html` 或 `import` |
| `approval_status VARCHAR(16)` | `candidate`、`pending_evidence`、`approved` 或 `rejected` |
| `runtime_status VARCHAR(16)` | `idle`、`running`、`paused` 或 `error` |
| `enabled BOOLEAN` | 运行开关；还需准入已批准且运行未暂停才可调度 |
| `terms_url TEXT`、`robots_url TEXT` | 常用审计入口；完整证据保存在政策证据表 |
| `policy_reviewed_at TIMESTAMPTZ` | 最近人工政策复核时间 |
| `image_policy VARCHAR(20)` | `none`、`reference_only`、`cache` 或 `download` |
| `requests_per_second NUMERIC(8,3)` | 明确配置的来源级速率，必须大于 0 |
| `max_concurrency SMALLINT` | 来源级并发，范围 1 至 8 |
| `last_error TEXT`、`paused_reason TEXT` | 脱敏后的运行或暂停原因 |
| `created_at`、`updated_at` | 服务端时间 |

认证凭据不进入此表。Token、Cookie 或密钥继续通过部署环境或专用密钥引用注入，日志和审计不得记录真实值。

### `av_catalog_source_policy_evidence`

每条证据保存 `source_id`、证据类型、公开 URL 或内部授权引用、核对时间、内容 SHA-256、中文摘要、允许字段、图片政策、保存期限、频控和记录人。证据类型至少包括 `terms`、`api_documentation`、`robots` 和 `written_permission`。

`robots` 证据不能单独支持批准。服务层只有在有效证据同时覆盖数据使用、图片策略、频控和成年内容边界时才允许把 `approval_status` 改为 `approved`；来源没有公开条款但有可复核书面授权时，`terms_url` 可以为空。准入状态变更追加审计事件，不覆盖旧证据。

### `av_catalog_releases`

| 字段 | 语义 |
|---|---|
| `id UUID` | 目录发行内部主键 |
| `normalized_code TEXT` | 规范化番号，可空且不唯一 |
| `title_original TEXT`、`title_zh TEXT` | 原始与中文标题 |
| `overview_original TEXT`、`overview_zh TEXT` | 原始与中文简介 |
| `release_date DATE`、`duration_seconds INT` | 合并后的发行信息 |
| `studio_name TEXT`、`label_name TEXT`、`series_name TEXT` | 首期可查询展示字段 |
| `review_status VARCHAR(20)` | `clean`、`needs_review`、`merged`、`retired` |
| `field_provenance JSONB` | 每个合并字段当前选择的来源事实 ID 与规则版本 |
| `created_at`、`updated_at` | 服务端时间 |

`normalized_code` 建普通索引，不建唯一约束。同番号多版本、重发和来源错误必须能并存并进入审核。

### `av_catalog_source_records`

| 字段 | 语义 |
|---|---|
| `id UUID` | 来源事实主键 |
| `source_id UUID`、`external_id TEXT` | 来源内强身份，唯一约束 `(source_id, external_id)` |
| `release_id UUID NULL` | 已确认或自动高置信归属的目录发行 |
| `detail_url TEXT` | 允许保存的来源详情 URL |
| `raw_fields JSONB` | 经过字段白名单和敏感信息清理的来源字段 |
| `normalized_fields JSONB` | 适配器输出的统一字段契约 |
| `content_fingerprint CHAR(64)` | 规范化快照 SHA-256，用于无变化短路 |
| `source_created_at`、`source_updated_at` | 来源声明时间，可空 |
| `first_seen_at`、`last_seen_at`、`fetched_at` | 本系统观察时间 |
| `record_status VARCHAR(20)` | `active`、`missing`、`restricted`、`rejected` |

来源响应不保存原始 HTML、认证头、Cookie、付费内容或视频 URL。若解析回归需要响应夹具，只能使用人工清理、最小化且确认可入库的测试样本。

### `av_catalog_source_snapshots`

只在 `content_fingerprint` 变化时追加，保存 `source_record_id`、指纹、白名单字段 JSONB、抓取时间和适配器版本。来源记录保存当前快照，快照表保留可审计历史。保留期限在来源批准时确定；未知时不得无限保存。

### 人物、分类与图片

- `av_catalog_people`：目录人物，不以名称唯一。
- `av_catalog_person_identities`：唯一 `(source_id, external_id)`，可为空的 `person_id` 表示待归并。
- `av_catalog_release_people`：发行与人物关系，角色为 `actor` 或 `director`，保存排序和来源依据。
- `av_catalog_genres`、`av_catalog_source_genres`、`av_catalog_release_genres`：内部分类、来源标签映射和发行分类关系分层保存。
- `av_catalog_images`：来源记录、发行、图片类型、来源 URL、内容哈希、本地路径、政策和状态；`image_policy` 不允许时本地路径必须为空。

### 覆盖、关联与审计

- `av_catalog_overrides`：实体、字段、JSON 值、理由、操作者、创建和撤销时间；同一实体字段只允许一个有效覆盖。
- `av_catalog_video_links`：`video_id`、`release_id`、`candidate/confirmed/rejected`、匹配方式、置信度和确认信息；每个视频最多一个有效 `confirmed` 主关联。
- `av_catalog_merge_events`：合并、拆分、字段重算和关联变化的事件负载，支持定位和人工回滚。

### 运行与游标

- `av_catalog_sync_runs`：来源、流名称、`queued/running/succeeded/partial/failed/cancelled`、起止时间、起始/结束游标、请求/新增/变更/跳过/失败计数和错误摘要。
- `av_catalog_sync_cursors`：唯一 `(source_id, stream_name)`，保存有版本的 JSONB 游标、最近成功运行和更新时间。
- 对 `sync_runs` 建同一来源和流仅一个 `queued/running` 的部分唯一索引，数据库负责阻止并发重叠。

## 统一来源事实

适配器的详情输出至少包含：

- `external_id`、`detail_url`、`code`、`title_original`、`overview_original`。
- `release_date`、`duration_seconds`、`studio`、`label`、`series`、`sku`。
- 带来源人物 ID 的演员和导演；只有名称时明确标记弱身份。
- 带来源标签 ID 或稳定名称的分类。
- 图片 URL、图片类型和来源声明的尺寸；不在适配层下载图片。
- `source_created_at`、`source_updated_at` 和允许保存的扩展字段。

所有文本做长度上限、UTF-8 和控制字符校验；URL 只接受批准域名下的 `https`，图片 CDN 必须在来源批准记录的允许域名中。

## 来源适配接口

```go
type AVCatalogSource interface {
    Code() string
    Discover(ctx context.Context, request DiscoverRequest) (DiscoverPage, error)
    Fetch(ctx context.Context, ref SourceRef) (NormalizedSourceRecord, error)
}
```

- `Discover` 返回稳定来源身份、详情引用、来源更新时间和下一页游标，不返回数据库模型。
- `Fetch` 读取单条详情并输出统一来源事实，不执行跨来源匹配或图片下载。
- 游标必须携带 `version`；适配器升级导致旧游标不可解释时返回明确错误并暂停来源，不能静默从头抓取。
- 单元测试全部使用离线夹具和假时钟；真实网站在线状态不进入 CI。

## 页事务与幂等

每个发现页按以下顺序处理：

1. 校验来源仍为 `approval_status=approved AND enabled AND runtime_status<>paused`，并锁定当前运行和游标。
2. 在事务外按来源限速请求一页发现结果及必要详情。
3. 在一个事务中按 `(source_id, external_id)` upsert 当前来源记录。
4. 指纹变化时追加快照并重算匹配候选；指纹相同时只更新 `last_seen_at`。
5. 更新运行计数并推进到适配器返回的下一游标，然后提交事务。
6. 事务失败时不推进游标；整页重试依靠唯一约束和指纹保持幂等。

不得先推进游标再写来源记录，也不得因某一条解析失败而丢弃整页其他已验证事实。单条失败进入运行错误明细并使运行成为 `partial`，游标语义必须保证该条后续仍可回查。

## 增量水位

对提供变化接口的来源，游标建议保存：固定本轮 `from`、当前页、已观察最大更新时间和规则版本。完成全部分页后，下一轮水位取最大更新时间减去可配置重叠窗口，默认 10 分钟；重叠数据按来源唯一键和指纹去重。

对只提供页码的来源，批准前必须证明排序稳定并定义页面漂移策略；否则只能按明确番号或用户种子做受控回查，不能伪装成可靠全量同步。

ThePornDB 的 `/changes/jav?from=...&page=...&per_page=100` 技术上适合上述水位模型，但来源批准仍是前置条件。

## 限速、重试与自动暂停

- 所有请求共享来源级令牌桶和并发信号量；没有显式配置速率时拒绝运行。
- `429` 优先遵循合法的 `Retry-After`；未提供时使用有上限的指数退避加抖动。
- 网络错误、超时和 `5xx` 最多重试 3 次；单次请求和整次运行都有截止时间。
- `401/403` 不重试并把运行状态改为 `paused`，避免错误凭据或政策门禁造成持续请求；若响应表明条款或地区资格变化，准入状态同时退回 `pending_evidence`。
- 响应变成登录页、验证码、地区限制、付费墙或 robots 禁止时立即停止并暂停，不尝试绕过。
- 连续解析失败超过来源阈值时熔断为暂停；恢复必须由管理员复核页面与政策后执行。

## 图片处理

图片分两步：先登记允许保存的引用，再由独立任务依据来源 `image_policy` 决定是否下载。下载前再次核对允许域名和策略；本地文件按内容哈希去重，但来源归属记录不能因二进制相同而合并丢失。

策略变为 `reference_only` 或来源要求删除时，停止新下载并创建可审计清理任务。清理本地副本不删除来源事实；远端 URL 是否继续保存取决于来源条款。

## 管理 API 提案

- `GET /admin/av-catalog/sources`：来源准入、运行状态和最近政策复核。
- `PUT /admin/av-catalog/sources/{id}`：启停、限速和运行状态，不直接接收 `approved` 布尔值。
- `POST /admin/av-catalog/sources/{id}/policy-evidence`、`POST .../approve`：登记证据并由服务层校验覆盖范围后批准，全部写审计事件。
- `POST /admin/av-catalog/sources/{id}/runs`：启动 backfill、incremental 或 recheck；服务端执行全部调度门禁。
- `GET /admin/av-catalog/runs`：运行、计数、游标范围和脱敏错误。
- `GET /admin/av-catalog/releases`、`GET /admin/av-catalog/releases/{id}`：目录查询和字段来源。
- `PUT /admin/av-catalog/releases/{id}/overrides/{field}`、`DELETE ...`：创建和撤销人工覆盖。
- `POST /admin/av-catalog/reviews/{id}/merge`、`/split`、`/reject`：处理归并冲突并记录事件。
- `POST /admin/av-catalog/video-links/{video_id}/confirm`、`/reject`：处理本地视频关联候选。

目录首期只开放管理员接口，不进入手机端或 TV 端；面向 App 的浏览和搜索是独立后续阶段，届时必须重新确认成人内容访问控制和图片展示权限。

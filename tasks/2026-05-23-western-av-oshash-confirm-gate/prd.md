# PRD：欧美 AV oshash 计算与刮削确认门控

- 日期：2026-05-23
- 目标端：Go 后端（`internal/` + `pkg/` + `migrations/` + `cmd/`） + admin-web 前端 + 数据库 migration
- 范围：仅欧美 AV 上传链路 + admin 端确认 UI + ThePornDB 刮削器质量修复（次级目标）
- 关联术语：见 `CONTEXT.md` 的 [`AV 地区分类`]、[`欧美 AV`]、[`刮削确认门控`]、[`刮削确认`]、[`弃刮`]、[`手动刮削`]

## 1. 用户故事

作为管理员，我希望：

- **故事 A（不被错误自动压缩）**：上传一个欧美 AV 原版后，系统不要在我还没看过候选元数据的情况下就把它压缩——以前刮不到时元数据是空的、压缩照样跑，事后想纠正只能在 ready 状态走手动刮削，且 transcoded 版本的 oshash 已经跟 ThePornDB 数据库对不上了，hash 路径永久失效。
- **故事 B（hash 直命中减少"刮不到"）**：上传原版时系统先用 oshash 直查 ThePornDB；只要文件确实是某个原始发行版（未被任何编辑/转码），命中后无需精准文件名匹配就能拿到正确候选。
- **故事 C（拿不到候选时不被卡死）**：极少数发行 ThePornDB 上根本没有，我能明确「弃刮」让视频继续走压缩、可播放，事后想再补元数据仍可走手动刮削。
- **故事 D（手动刮削同样享受 hash 红利）**：对一个已经 ready 的欧美 AV 重新走"手动刮削"时，hash 路径也参与——hash 不仅是上传时的一次性优化，而是该视频整个生命周期内多个刮削入口共用的稳定指纹。
- **故事 E（次级 · ThePornDB 关键字命中率提升）**：在 hash 不命中或没存 hash 的场景下，keyword 路径本身也比当前更可靠——修复 `/movies` 端点参数错误、补齐 OUMEI_NAME 站点短名映射、加 401/429 重试、加相似度阈值、扩关键字正则后，"刮不到"的比例从经验估算的 ~80% 降到 ~30%。

## 2. 作用域

### 在范围内
- AV 类型 + `site_category = western` 的**新上传**走 [`刮削确认门控`]
- 上传表单加 [`AV 地区分类`] radio（影响所有 AV 上传；默认 `日本`，只对欧美档触发 oshash 与门控）
- 新增 video status：`av_scrape_pending`
- 新增列：`videos.os_hash CHAR(16)` + partial index
- 新增 endpoint：`PUT /admin/scrape/skip`（[`弃刮`]）
- 改造 endpoint：`PUT /admin/scrape/confirm`（[`刮削确认`] 内部按状态分支决定是否入队 transcode）
- 改造 task handler：`HandleScrapeAV` 在 western 分支落 `av_scrape_pending` 不入队 transcode；`HandleScrapeRetag` 在 retag 至 western 时补算 oshash
- 新增 `pkg/oshash`：OpenSubtitles 经典算法 Go 实现
- 共享 helper：`collectThePornDBHashCandidate`，被自动刮削 / 手动刮削 / retag 三路读取
- ThePornDB 五项修复（次级目标）：详见阶段 4
- admin-web 三个组件：`VideoUpload.vue` / `VideoList.vue` / `ScrapePreview.vue`
- admin-web `AVManualScrape.vue` 后端透传 video_id 让 hash 路径可用（前端 UI 无显式改动，仅候选列表多 `hash 命中` 徽章）

### 非目标（明确不做）

- **N1**：movie / episode / 短视频不接入 [`刮削确认门控`]——只有欧美 AV
- **N2**：日本 / FC2 AV 不算 oshash、不进门控（保持现有 `上传 → 自动刮削 → 自动 transcode` 行为不变）
- **N3**：批量确认 / 批量弃刮 / 批量重搜的 admin UI——单条操作即可，等用户反馈再加
- **N4**：oshash 算法可配置——写死 OpenSubtitles 经典算法（文件大小 + 头/尾 64KB 累加），与 MDCX `oshash.oshash()` 字节一致
- **N5**：ThePornDB 相似度阈值可配置——写死 0.4 常量，需要调整时改源码重新部署
- **N6**：oshash 用于本地去重——SHA256 + `file_hashes` 表已经覆盖，不混入
- **N7**：自动弃刮兜底——零候选时仍然停在 `av_scrape_pending`，等管理员主动 skip
- **N8**：Dashboard 加"待确认"卡片——管理员通过 `VideoList` 状态筛选进入即可
- **N9**：候选列表能拖动重排——展示顺序由后端 `match_source` + similarity 排好，前端不允许覆盖
- **N10**：`scrape_preview` 主动失效缓存——admin 觉得不对劲点"重新搜索"即可；不引入 TTL
- **N11**：扩展 `AV 地区分类` 用作通用内容地区标签——见 CONTEXT 该术语条款的约束，本分类**只**为 oshash 与门控服务，不允许被其它特性借用
- **N12**：Android phone / TV App 改动——本任务零客户端影响，不需要 `versionCode` 升

## 3. 关键设计决策（与 CONTEXT.md 术语对齐）

### 3.1 数据模型
- `videos` 表新增 `os_hash CHAR(16)`，nullable + partial index `WHERE os_hash IS NOT NULL`
- `videos.status` CHECK 约束加入 `av_scrape_pending`
- `videos.metadata` 新增三块（JSONB 子键，不开新列）：
  - `scrape_preview`：候选列表数组（top 6），每条带 `match_source ∈ {oshash, keyword:scenes, keyword:movies, manual_retry}`
  - `scrape_attempt`：本次自动刮削诊断信息（hash_used、keywords_tried、各路径错误）
  - `scrape_skipped` / `scrape_skip_reason` / `scrape_skipped_at`：[`弃刮`] 审计三件套

### 3.2 状态机
```
upload (西方 AV)
   ↓ status=uploaded
EnqueueScrapeAV
   ↓ status=scraping
HandleScrapeAV
   ↓ status=av_scrape_pending  (写 scrape_preview)
   ├── admin confirm → ConfirmAV
   │       ↓ status=processing → EnqueueTranscode → status=ready
   └── admin skip → SkipScrape
           ↓ status=processing → EnqueueTranscode → status=ready
```

非欧美档（日本 / FC2）保持现有 `uploaded → scraping → ready/uploaded → transcode` 路径完全不变。

### 3.3 候选收集策略
[`欧美 AV`] 上传自动刮削无论 hash 是否命中都会跑齐三条路径（hash + scenes + movies），最多收集 6 个候选并按 `match_source=oshash` 排首位、其余按相似度降序排列。hash 命中候选在 UI 上预选中但**仍要管理员点确认**——不存在"高置信自动 confirm"逃生。

### 3.4 oshash 一次计算多路复用
计算时机：仅在 `type='av' AND site_category='western'` 上传时同步算（位于 `UploadService.SaveUpload` 内部，SHA256 落盘后、`InsertFileHash` 之前）；分片上传走 merge 之后。
失败处理：算失败**不阻塞上传**，落 NULL + log warn。
复用范围：自动刮削 / 手动刮削 / retag 重刮三个入口都读 `videos.os_hash`，非空就先跑 hash 路径。
retag 补算：从 `日本 / FC2` retag 到 `欧美`、且原文件仍在、且 `os_hash` 为空时，retag 处理器补算一次。

### 3.5 ThePornDB 五项修复（次级目标）
1. `/movies` 端点参数从 `parse=` 改为 `q=`（修复整条 movies fallback 通道）
2. `OUMEI_NAME` 字典移植：从 MDCX `manual.py:57` 起约 200+ 条站点短名→长名映射，port 成 Go map
3. `fetchThePornDBJSON` 包装 401(不重试) / 429(重试) / 5xx(重试) / 网络超时(重试)，退避 200ms / 400ms / 800ms，最多 3 次；该包装**仅 ThePornDB 调用使用**，不动共享 `fetchAVJSON`
4. `thePornDBBestMatchURL` 引入相似度阈值常量 `thePornDBMinSimilarity = 0.4`，最高分低于阈值时主动返回空、让上层进下一个 keyword
5. `datedNumberPattern` 关键字正则扩展为 P1 / P2 / P3 三档：P1 原有 `Series.20.05.12`、P2 新增 `Series_2020-05-12`、P3 新增 `[Series] 20.05.12`

## 4. 验收标准

### 4.1 主目标（[`刮削确认门控`]）
- [ ] 上传一个欧美 AV 原版（hash 在 ThePornDB 有索引），上传响应返回 `status: "scraping"`；自动刮削任务完成后 `videos.status` 落 `av_scrape_pending`，`videos.os_hash` 非 NULL，`videos.metadata.scrape_preview` 至少有一条且首位 `match_source = "oshash"`，**未**发现 `transcoded_path`，**未**入队 transcode task。
- [ ] admin 进入 `VideoList` 用 `状态 = 欧美 AV 待确认` 筛选，能看到该视频；点进去 `ScrapePreview` 看到候选列表首位带 `hash 命中` 徽章并预选中。
- [ ] 点"确认刮削"，调用 `PUT /admin/scrape/confirm` 后视频 `status` 立即转 `processing`、transcode task 入队；等若干秒后转 `ready` 且 `transcoded_path` 非空。
- [ ] 上传一个文件名乱码、关键字也搜不到、ThePornDB 也没 hash 命中的样本，`scrape_preview` 为空数组，状态仍落 `av_scrape_pending`（**不**走 failed）；admin 点"弃刮"调 `PUT /admin/scrape/skip`，状态转 `processing` 且 `metadata.scrape_skipped=true`、`metadata.scrape_skip_reason` 写入；transcode 跑完转 `ready` 视频可播。
- [ ] 上传日本 / FC2 AV：行为完全跟当前一致（落 `ready` 并自动 transcode），`videos.os_hash` 为 NULL，`videos.metadata.scrape_preview` 不存在。
- [ ] 上传 movie / episode / 短视频：完全不受影响，状态机零变化。

### 4.2 次级目标（ThePornDB 修复）
- [ ] `/movies` 修复：构造一个能命中 movies endpoint 的 keyword，捕获请求看到 query string 是 `?q=...&per_page=100`，不是 `?parse=...`。
- [ ] `OUMEI_NAME` 字典：上传文件名形如 `Wgp.20.05.12.xxx.mp4`，自动刮削能识别 `wgp` → `WhenGirlsPlay` 并精准命中（候选列表非空，且 `match_source=keyword:scenes`）。
- [ ] 重试：mock 一次 429 响应再 200，刮削最终成功（不是 fail）；mock 一次 401，立即失败不重试（避免无效重试浪费）。
- [ ] 阈值：构造一个 ThePornDB 搜索返回但所有相似度都低于 0.4 的场景，`thePornDBBestMatchURL` 返回空，候选最终落到下一个 keyword 或最终为零。
- [ ] 正则：上传 `Series_2020-05-12_Title.mp4` 和 `[Series] 20.05.12.Title.mp4` 两种文件名，关键字提取都能拿到 series + date。

### 4.3 回归点（不该坏的）
- [ ] 现有 `AVManualScrape` 对已 ready 视频的重刮：候选列表里若该视频 `os_hash` 非空则带 hash 候选，confirm 后**不**再次入队 transcode（保持现有手动刮削语义）。
- [ ] 现有 `EnqueueScrapeRetag` 路径：retag 至 `episode` 仍走 `tv_pending`；retag 至 `movie` / `av(非western)` 行为不变。
- [ ] 短视频 / movie / episode 上传后 transcode 行为完全不变。

## 5. 出错与边界

- **oshash 计算失败**（磁盘 I/O / 文件 < 128KB）：`videos.os_hash = NULL`，刮削自动 fallback 到 keyword 路径，不报错给前端。
- **hash 命中 1 个但内容明显错挂**（极罕见）：admin 通过对照其它候选发现 → 手动选别的候选 confirm，hash 候选不会"霸占"操作。
- **网络分区 → ThePornDB 全部超时**：自动刮削任务由 Asynq 重试机制兜底（已有 `MaxRetry(3)`），最终落 `av_scrape_pending` 且 `scrape_preview = []`；admin 走"重新搜索"或"弃刮"。
- **video 被删除**：`os_hash` 列随 row 一起删（`ON DELETE` 行为天然）。
- **同一个文件被多次上传**：SHA256 dedup 在 `UploadService` 入口已经拦截（`file_hashes UNIQUE`），不会重复算 oshash 或重复入门控。

## 6. 文档同步

- `CONTEXT.md`：本任务在 grill-with-docs 阶段已经同步新增了 `AV 上传分类术语` 章节及 [`AV 地区分类`]、[`欧美 AV`]、[`刮削确认门控`]、[`刮削确认`]、[`弃刮`] 五个术语；implement 阶段如有偏离需回写并同时回头修订本 PRD。
- `plan.md`：实施完成后追加本任务条目（功能更新 + 实施更新两条）。
- Android：不影响，零 `versionCode` 升级。

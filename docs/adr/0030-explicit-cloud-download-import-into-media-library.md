# 云下载任务显式导入媒体库

## 背景

ADR 0028 把 115 云下载定义为“下载文件只落到 `DOWNLOAD_DIR`，不创建 `videos`、不触发转码或归档导入”，避免“下载即入库”。现在管理员需要在云下载任务完成后，把单个已落盘文件显式送进现有媒体流程：

- 视频：进入视频上传后的待处理/转码或刮削队列；
- 压缩包：进入现有压缩包导入批次，异步解包后在工具箱继续处理。

任务列表本身没有可靠的媒体类型字段：`declared_name` 可能为空或与 115 实际文件名不一致，实际文件名只在 `local_path` 的 basename，实际大小/MIME 未落库；当前云下载也只支持单文件，目录/多文件任务没有成功落盘路径。因此不能只靠扩展名判断，也不能在下载完成时自动入库。

## 决策

- **入口**：只在 115 云下载任务列表，对 `completed`、`local_path` 存在、单个普通文件显示“入媒体库”；点击后由后端探测实际文件（压缩包用 mimetype + 受支持扩展名，视频用 ffprobe 视频流），再打开视频或压缩包弹窗。无法识别/目录/多文件/ISO 显示“暂不支持”。论坛列表不变。
- **显式、不自动**：下载完成、列表加载、轮询都不触发导入；必须管理员点击并提交。
- **异步**：新增 `cloud_download_imports` 记录（一个云任务唯一一条）和 `cloud:ingest` 任务，放 `cloud` 队列，超时 2 小时、MaxRetry 2；`imported` 只表示目标记录/批次已创建且后续任务已入队，不等待转码/解压完成。失败只改导入记录，云任务保持 `completed`，可重试同一条记录。
- **视频**：
  - 先算 SHA-256 + size，再查 `file_hashes`。
  - 命中（不判断类型，含 `pending_delete`）：`skipped_existing`、关联已有 `video_id`、不复制/不新建/不改已有视频元数据，**直接删除下载目录原文件**。
  - 未命中：复制到 `STORAGE_ROOT/videos/<videoID>/source-original<ext>`，复用上传 service 创建视频并按类型入转码/刮削；目标 + 入队成功后删除下载目录原文件。
  - 视频源文件在转码成功后删除：删除的是导入复制来的 `source-original`，**不是** episode DV 的 `source-dv` 播放产物；`source-dv` 是转码输出/播放路径，必须保留。若 `videos.original_path` 在 DV 分支被 move 成播放路径，则不可删除该路径。云导入通过 metadata 标记执行该清理，普通上传行为不变。转码失败/卡住时保留导入源文件供重试。
- **压缩包**：
  - 新增 `ImportLocalArchive`（不伪造 multipart），复制到 `STORAGE_ROOT/archive-imports/<batchID>/original/`，创建批次并入 `archive:extract`；目标 + 入队成功后删除下载目录原文件。
  - 批次 `original/` 与 `extracted/` 按现有压缩包流程保留，工具箱继续处理；不自动入媒体库。绝不把 `DOWNLOAD_DIR` 路径写进批次 `original_path`。
- **复制/回滚**：复制先写临时/部分文件，目标创建/入队成功后删下载源；任一步失败清理复制、保留下载源，导入记录 `failed`。视频源码只在目标 + 入队成功后删；hash 命中直接删源。
- **一次性**：`task_id` 唯一；成功/跳过终态；失败可重试同一记录（`attempt+1`），重试不得重复复制/创建。
- **路径安全**：导入前校验 `local_path` 解析后在 `DOWNLOAD_DIR` 内且为普通文件；不信任客户端路径。
- **配套修复**：顺手修 `RetryExtract` 的 `encoding_mode` 未持久化，否则导入后的 GBK ZIP 在工具箱选编码重试无效。

## 与 ADR 0021 / 0028 / 0029 的关系

- ADR 0021/0028/0029 的“不自动入库/不转码”继续成立；本 ADR 只新增**管理员显式、任务级**的导入出口。
- 论坛列表仍不加入库入口；入口只在云下载任务列表，不触碰 ADR 0021 的只读边界。
- ADR 0028 的“下载后自动入库”仍是被否决方案；本 ADR 保留“自动”与“显式”的区分。

## 考虑过的替代方案

- **只按扩展名判断**：实现简单，但 `declared_name`/实际文件名不可靠，容易把视频送进解包或把压缩包送进转码。
- **下载完成时自动入库**：正是 ADR 0028 否决的隐式副作用；会让下载任务承担 hash/复制/建记录，且失败难回滚。
- **同步 HTTP 完成 hash/复制/建视频**：大文件会超过普通请求超时，也不适合占用云下载 worker 的 6 小时槽位。
- **直接把 `DOWNLOAD_DIR` 路径交给上传/压缩包服务**：转码会删除原片；压缩包删除批次会 `RemoveAll(filepath.Dir(original_path))`，可能误删下载目录。
- **hash 命中按类型区分**：用户明确选择只按 hash 去重，相同 hash 直接跳过并删下载副本。
- **多文件/目录首期支持**：当前云下载没有成功落盘路径，先只支持单文件。

## 后果与风险

- 新增 migration `0045_cloud_download_imports`、导入记录表、`cloud:ingest` 任务、预览/导入接口和任务列表 import 摘要。
- 视频导入会在 `STORAGE_ROOT/videos/<videoID>/source-original` 产生一份临时源文件；转码成功后删除，失败时保留。
- 压缩包导入会多一份批次 `original/` 副本；下载目录原文件删除，批次空间由现有“删除批次”清理。
- hash 命中 `pending_delete` 仍删除下载文件：已有视频随后被删除时，下载副本也不在；这是用户明确选择的取舍，UI/ADR 记录该风险。
- 压缩包跨云任务同包不做批次去重（现有导入也无批次去重）；只保证单任务一次性。
- 压缩包密码仍走现有链路（Redis payload/解压器 argv，不落 DB、无加密）；不新增密码持久化。
- `imported` 不代表转码/解压成功；视频/批次状态继续由现有队列和页面展示。

## 并发、重试与恢复

- `cloud:ingest` payload 携带 `attempt`；旧 attempt 的 asynq 自动重试在 `Process`/`MarkFailed` 入口直接 no-op，不能重置或覆盖新 attempt。手动重试用 `attempt+1` 生成新 TaskID。
- 活动导入超过约 10 分钟仍未终态时，前端显示“重新导入”，用户重新填写表单后后端用**同一 attempt/TaskID**重新入队：有活任务时 asynq TaskID 幂等忽略，任务丢失时补投。表单/密码仍只在 Redis payload，不落库，因此队列彻底丢失后必须重新填写，无法静默恢复。
- `SaveImportedFile` 已创建视频但 `video_id` 尚未回写时，重试按视频 metadata 的 `source=cloud_download` + `cloud_task_id` 识别为本任务目标，走 resume 补入队；不能误判为外部 hash 命中而删除下载源。
- 压缩包重试已有批次时，只有 `uploaded/failed` 才 `RetryExtract`；`extracting/processing/ready/partial/completed/needs_*` 只关联并完成导入，避免清空正在解包或已完成的 `extracted`。
- 已知残余：没有独立的 import reconciler；`SetArchiveBatch` 失败时批次已在工具箱创建但记录未关联，重试可能产生第二个批次；真实 115/ffprobe 端到端仍待部署环境验证。

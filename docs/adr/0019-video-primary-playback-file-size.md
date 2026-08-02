# 视频记录主播放文件大小，不把目录占用混入单字段

> 状态：已接受（2026-08-02）

## 背景

`videos.duration_seconds` 已记录视频时长，但视频实体没有可直接提供给管理端的转码播放文件大小。原始上传大小仅存在于 `file_hashes.file_size`，其职责是哈希去重，既不保证反映当前可播放产物，也不能代表重新转码后的体积。

一个视频目录还可能包含缩略图、字幕和 Dolby Vision 剧集的保留源文件。将这些文件总和写入“转码后文件大小”会让字段语义随资产结构变化，既无法与 `transcoded_path` 一一对应，也会误导管理员判断当前播放产物的体积。

## 决策

- 在 `videos` 增加可空 `BIGINT` 列 `transcoded_file_size`，单位为字节，并以 `CHECK (transcoded_file_size IS NULL OR transcoded_file_size > 0)` 区分未知与有效文件大小。
- 该字段只表示 `transcoded_path` 指向的主播放常规文件，不计入缩略图、字幕、海报、临时文件或目录总占用。Dolby Vision 剧集直拷的 `source-dv.*` 成为 `transcoded_path` 时，仍按同一语义记录其大小。
- 常规转码、重新转码、Dolby Vision 直拷和直接导入已就绪视频在持久化播放结果时同步写入正数大小；无法读取刚生成播放文件大小时，任务不能标为 `ready`。
- 历史数据使用 `cmd/backfill-video-transcoded-file-size` 读取元数据回填。命令默认 dry-run，只有 `--apply` 才写库；只允许统计 `STORAGE_ROOT/videos/` 内绝对路径的常规非空文件。遇到单条异常继续处理其它记录，在 JSON 报告中保留失败项并以非零状态结束。
- 管理端视频详情接口和页面返回、展示该字段；手机端、TV 端及公开视频 API 不增加该字段。

## 考虑过的替代方案

- **继续从 `file_hashes.file_size` 读取**：该值是上传原文件大小和去重条件，转码后通常不相等，重新转码后更会失效。
- **每次打开管理端详情时 `stat` 文件**：会把外盘/NAS 元数据 I/O 放入请求路径，文件缺失和慢盘会直接拖慢管理界面，也无法建立可统计的数据字段。
- **单字段记录整个视频目录大小**：会把视频、缩略图、字幕、临时文件和保留源文件混为一谈，字段不能再回答“当前播放文件有多大”。目录总占用需要独立的资产统计模型。
- **把未知大小写为 `0`**：无法区分“文件不存在/未回填”和不合法的零字节播放文件，会掩盖存储异常。

## 后果

- `0036_video_transcoded_file_size` 只新增可空列和约束，满足 [[migration 前向兼容契约]]：旧二进制可忽略该列继续运行。
- 历史回填是受控的后台命令，不进入 HTTP 请求或转码热路径；完成后仍可定期重跑校准人工替换过的文件。
- 管理员获得主播放文件大小，而存储总占用仍应通过磁盘统计或未来独立资产统计处理，不能简单对该字段求和后称为全盘占用。

## 关联

- `migrations/0036_video_transcoded_file_size.up.sql`
- `cmd/backfill-video-transcoded-file-size/main.go`
- `internal/queue/tasks.go`
- `internal/services/flick_import.go`
- `CONTEXT.md` 中 [[转码后主播放文件大小]] 与 [[主播放文件大小写入与校准]]

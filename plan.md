# plan.md

> 2026-07-02 整理版：已按用户要求删除纯环境发布与推送流水记录，并将同一事项的开始、准备、待执行等重复过程记录合并为保留最终有效记录。后续新增计划继续按反向时间顺序追加。

## 2026-07-03 20:18 +0800
- 进度：完成压缩包导入大文件处理超时短修。管理端压缩包重试解包、单文件处理、分组处理和批次处理请求均显式关闭 Axios 默认 30 秒超时；API 回归测试已锁定这些长耗时动作必须 `timeout: 0`。同步在 `CONTEXT.md` 沉淀 `压缩包处理长耗时请求` 边界。
- 影响文件：`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 既有警告）；`git diff --check -- admin-web/src/api/admin.js admin-web/src/api/admin.spec.js CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' admin-web/src/api/admin.js admin-web/src/api/admin.spec.js CONTEXT.md plan.md` 无输出。

## 2026-07-03 20:16 +0800
- 进度：开始短修压缩包导入大文件处理超时问题。已定位到后端处理链路无 30 秒显式超时，但管理端公共 Axios 默认 `timeout: 30000`，压缩包处理文件、处理分组、处理批次和重试解包请求未关闭默认超时；本轮先做前端 API 层短修，不改为后台任务模型。
- 影响文件：预计涉及 `admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/api/admin.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-07-03 15:41 +0800
- 进度：完成管理端“待删除短视频”页面留白修复。页面已接回普通管理端 `Layout` 壳层，恢复侧栏、顶栏、命令面板和主内容留白；窄屏布局改为自然滚动并给队列/预览面板稳定高度，避免移动端空状态被压扁裁切。同步新增 router spec 锁定该页面必须接入普通壳层，并在 `CONTEXT.md` 沉淀 `admin 普通页面壳层` 术语。
- 影响文件：`admin-web/src/views/PendingDeleteShorts.vue`、`admin-web/src/router/index.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：MCP 通过本机 mock API 预览 `/short-pending-delete`，桌面与移动截图确认壳层、留白和空状态正常；`cd admin-web && npm run test -- src/router/index.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 既有警告）；`git diff --check -- admin-web/src/views/PendingDeleteShorts.vue admin-web/src/router/index.spec.js CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' admin-web/src/views/PendingDeleteShorts.vue admin-web/src/router/index.spec.js CONTEXT.md plan.md` 无输出。

## 2026-07-03 15:23 +0800
- 进度：开始修复管理端“待删除短视频”页面留白问题。已通过 `grill-with-docs` 对照 `CONTEXT.md` 确认当前术语是“待删除短视频 / 短视频待删除队列”，不是已退役的“短视频审核”；改动范围收敛到该页面样式，不变更业务流程、路由或接口。
- 影响文件：预计涉及 `admin-web/src/views/PendingDeleteShorts.vue`、`CONTEXT.md`、`plan.md`
- 验证：待用 MCP 查看本地 dev 页面加载情况，并执行 `cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-07-03 15:09 +0800
- 进度：完成短视频待删除队列复审修复。已将普通入口遗漏补齐：继续观看、喜欢/收藏仅返回 `ready` 视频，“我的上传”只排除 `pending_delete` 以保留上传生命周期记录；新增源文测试锁定这些过滤。管理端待删除页播放器区域已改为引用 `theme.css` 设计 token，不再扩散临时色值、圆角或阴影。`CONTEXT.md` 已把旧 `admin 短视频审核` 明确标记为退役术语，避免和新待删除队列混用。
- 复审说明：独立 Standards/Spec 复审无阻塞问题；Spec 提醒“我的上传”不应误过滤非 ready，已修正并复测。`toolboxPage.spec.js` 的 ED2K 断言同步保留，因为 `ToolboxEd2kDownload.vue` 在本功能前已是“新建任务/创建任务”，旧断言会导致必跑 `npm run test` 失败；`AppPreferencesStoreTest.kt` 的 Main dispatcher 包装保留，因为尝试回退后 `:app:testDebugUnitTest` 出现协程未捕获异常失败，恢复后测试通过。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_pending_delete_test.go`、`admin-web/src/views/PendingDeleteShorts.vue`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository -run 'TestAppListQueriesHidePendingDeleteVideos|TestShortPendingDelete|TestUpdateVideoStatusRejects|TestValidateAdminVideoStatusEdit' -count=1` 通过；`go test ./internal/repository ./internal/handlers -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`cd admin-web && npm run test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk 大小警告）；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug` 通过（仅现有 AGP/compileSdk 警告）；`git diff --check` 与乱码扫描通过。

## 2026-07-03 14:55 +0800
- 进度：开始对最新提交 `cc225f5 实现短视频待删除队列` 做复审修复，审查范围按 `HEAD~1...HEAD`。将按仓库规则、用户确认过的待删除队列需求、后端状态机并发风险、管理端替换完整性、手机端管理员入口与二次确认行为逐项核对；发现问题直接修复并复测，直到独立复审无阻塞问题。
- 影响文件：预计涉及 `plan.md`，如发现问题可能触及 Go 后端、`admin-web`、`android-app`、`CONTEXT.md`
- 验证：待执行独立 standards/spec 评审、必要定向测试、`go test ./... -count=1`、`go vet ./...`、`cd admin-web && npm run test && npm run build`、`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug`、`git diff --check`、乱码扫描。

## 2026-07-03 14:47 +0800
- 进度：完成“短视频待删除队列”大改并收口复审问题。后端在通用编辑与状态编辑中补 `SELECT FOR UPDATE` 行锁和事务内状态校验，迁移增加 `videos_pending_delete_short_check`，内部原始 `UpdateVideoStatus` 禁止绕过专用工作流直接写 `pending_delete`；管理端旧“短视频审核”已替换为“待删除短视频”列表，保留/最终删除二次确认流程不展示提名人和时间；手机端仅管理员在主短视频流右侧看到图标按钮，二次确认后加入待删除列表。独立后端复审与前端/手机端复审均无阻塞问题。
- 影响文件：`migrations/0031_short_pending_delete.*.sql`、`internal/models/models.go`、`internal/repository/*.go`、`internal/handlers/*.go`、`admin-web/src/api/admin.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/views/PendingDeleteShorts.vue`、`admin-web/src/views/VideoList.vue`、相关 admin-web spec、`android-app/app/src/main/java/com/chee/videos/...`、相关 Android 单测、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository -run 'TestUpdateVideoStatusRejectsPendingDeleteWorkflowBypass|TestValidateAdminVideoStatusEdit|TestAdminPendingDeleteMutationsUseRowLocks|TestShortPendingDeleteWorkflowSourceGuards|TestShortPendingDeleteMigration' -count=1` 通过；`go test ./internal/repository ./internal/handlers -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`cd admin-web && npm run test` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过（仅既有 AGP compileSdk warning）；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过（仅既有 AGP compileSdk warning）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src android-app/app/src internal migrations` 无输出。待提交。

## 2026-07-03 14:28 +0800
- 进度：完成手机端改造。短视频主流 ViewModel 会读取用户资料判断 `admin`，仅管理员在主短视频页右侧动作栏看到图标按钮；点击后弹出“加入待删除列表”二次确认，成功 Toast“已加入待删除列表”并从当前窗口移除该视频、跳到下一条，失败 Toast 并保留当前视频。同步新增手机端 API、URL builder、窗口移除纯逻辑测试，并将手机端版本 `0.1.3(4)` 递增到 `0.1.4(5)`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`ShortFeedViewModel.kt`、`ShortFeedProgressState.kt`、`core/network/ApiService.kt`、`core/repository/VideoRepository.kt`、`core/util/UrlBuilder.kt`、相关 Android 单测、`android-app/app/build.gradle.kts`、`plan.md`
- 验证：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedActionRailVisibilityTest --tests com.chee.videos.feature.shorts.ShortFeedPagerRestoreTest` 通过（仅既有 AGP compileSdk warning 和 kapt warning）。待执行全量验证、乱码扫描和提交。

## 2026-07-03 14:21 +0800
- 进度：完成管理端改造。旧 `短视频审核` 路由/导航/页面/helper/spec 已移除，新增 `待删除短视频` 页面 `/short-pending-delete`，采用列表 + 预览 + 单条 `保留` / `最终删除` 终审流程；页面不展示提名人和提名时间。视频管理新增 `pending_delete` 状态标签和筛选，允许预览待删除视频，但禁止手动改入/改出待删除状态，短视频待删除时隐藏类型转换入口。
- 影响文件：`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/api/admin.js`、`admin-web/src/views/PendingDeleteShorts.vue`、`admin-web/src/views/pendingDeleteShorts.helpers.*`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.*`、相关 spec、`plan.md`
- 验证：`cd admin-web && npm run test -- src/router/index.spec.js src/components/base/commandPalette.helpers.spec.js src/api/admin.spec.js src/views/videoList.helpers.spec.js src/views/pendingDeleteShorts.helpers.spec.js src/views/adminDateTimeDisplay.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）。待执行手机端和全量验证。

## 2026-07-03 14:14 +0800
- 进度：完成后端第一阶段实现。新增 `0031_short_pending_delete` migration、`pending_delete_at` 内部排序字段、短视频加入/保留待删除队列仓储方法、待删除短视频列表接口、手机/后台共用的加入待删除接口，以及后台预览 `pending_delete` 视频的签名播放放行；后台手动状态编辑已禁止改入/改出 `pending_delete`。
- 影响文件：`migrations/0031_short_pending_delete.*.sql`、`internal/models/models.go`、`internal/repository/video_repository.go`、`internal/repository/admin_repository.go`、`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/video_source.go`、相关 Go 测试、`plan.md`
- 验证：`go test ./internal/repository -run 'TestShortPendingDeleteMigration|TestScanVideoRecordReadsOSHash|TestValidateAdminVideoStatusEdit' -count=1` 通过；`go test ./internal/handlers -run 'TestRegister|TestAdmin|TestVideoSource' -count=1` 通过。待执行管理端、手机端和全量验证。

## 2026-07-03 14:04 +0800
- 进度：开始落地“短视频待删除队列”大改。需求已通过 grill 收口：手机端 admin 主短视频流用右侧图标将短视频加入 `pending_delete`，普通入口全局隐藏；后台旧“短视频审核”退役，新建“待删除短视频”列表 + 预览单条终审；通用视频管理保留直删和批量直删旁路，但不能手动改入/改出 `pending_delete`；`pending_delete_at` 仅作为内部排序字段，不展示提名人/提名时间。
- 影响文件：预计涉及 `migrations/0031_*`、Go 后端 handler/repository/model/router、`admin-web` 路由/API/视图、`android-app` 短视频页/API/版本号、`CONTEXT.md`、`plan.md`
- 验证：待执行 Go 定向测试、admin-web 构建/相关测试、手机端单测/构建、`git diff --check`、乱码扫描。

## 2026-07-03 09:13 +0800
- 进度：完成两次转码队列修复的本地复审并修复发现项。评审范围从单个 `2050b6b` 扩大到 `32055b5..HEAD`，覆盖 `3b2e42c` 进度管道假失败修复与 `2050b6b` AVC 超高尺寸修复。发现并修复两点：① `TranscodeVideo` 虽已先等进度管道再 `Wait`，但错误优先级仍可能在真实 ffmpeg 失败同时伴随进度读取错误时返回 `progress read failed`，本轮抽出 `buildTranscodeVideoError` 并改为真实 ffmpeg 退出错误优先；② `TestAdminEd2kDownloadStatusIsRedacted` 的假 `amulecmd` 使用 `/usr/bin/env bash`，全量并发下偶发 3 秒超时被 kill，本轮改成 POSIX `sh` + `case`，消除 bash 启动依赖。另补 `buildMaxDimensionScaleFilter` 表驱动测试，覆盖横屏、竖屏、方形、正常尺寸、未知尺寸和 odd max。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/handlers/admin_ed2k_download_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoError|TestBuildMaxDimensionScaleFilter|TestBuildTranscodeVideoArgsForAvcCompat|TestBuildTranscodeVideoArgsForHevcPrimary|TestIsEncoderUnavailableOutput' -count=1` 通过；`go test ./internal/handlers -run 'TestAdminEd2kDownloadStatusIsRedacted' -count=5` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`git diff --check -- pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go internal/handlers/admin_ed2k_download_test.go CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go internal/handlers/admin_ed2k_download_test.go CONTEXT.md plan.md` 无输出。待提交。

## 2026-07-03 09:06 +0800
- 进度：开始评审上一轮 `AVC 硬编超高尺寸失败` 修复，评审范围按 `HEAD~1..HEAD`（`2050b6b 修复AVC硬编超高尺寸失败`）处理。将按 standards/spec 两条线检查是否符合仓库规则、转码语义和实际失败原因；若发现问题直接修复并复测。
- 影响文件：待定，预计涉及 `pkg/ffmpeg`、`internal/services`、`CONTEXT.md`、`plan.md`
- 验证：待执行评审发现项对应定向测试、`go test ./... -count=1`、`go vet ./...`、`git diff --check`、乱码扫描。

## 2026-07-03 09:02 +0800
- 进度：完成本轮转码失败记录修复。只读查询最新失败记录后确认三类情况：① 最新一条仍是上一轮已修的 `file already closed` 进度假失败，关联视频 `ready` 且进度 100%，推断部署机 worker 仍需运行新二进制或该任务发生在部署窗口；② 同一短视频多次真实失败为 2160x4670 HEVC MOV 转 H.264 时 `h264_videotoolbox Cannot create compression session: -12903`，属于 AVC 硬编尺寸上限；③ 另有一条坏 AAC 文件在输出部分视频后音频滤镜失败，涉及“保留坏音轨/丢音轨/静音输出”的产品取舍，本轮不隐式改变。实现上给 `avc_compat` 输出加 4096 最长边上限，服务层把 ffprobe 源宽高传给 ffmpeg，超限时只等比例降采样（例如 2160x4670 → `scale=-2:4096`），普通尺寸和 HEVC 长视频不变。同步补 `AVC 硬编尺寸上限` 技术沉淀。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`CONTEXT.md`、`plan.md`
- 验证：先用新增测试得到红灯（`SourceWidth` / `SourceHeight` / `MaxVideoDimension` 字段不存在）；实现后 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgsForAvcCompat|TestBuildMaxDimension|TestBuildTranscodeVideoArgsForHevcPrimary' -count=1` 通过；`go test ./internal/services -run 'TestChooseTranscodeOutputProfile|TestBuildTranscodePlan|TestResolveProbeFields' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go vet ./...` 通过；`go test ./... -count=1` 通过；`git diff --check -- pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go internal/services/transcode.go internal/services/transcode_test.go CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go internal/services/transcode.go internal/services/transcode_test.go CONTEXT.md plan.md` 无输出。待提交。

## 2026-07-03 08:55 +0800
- 进度：继续排查转码队列失败记录。先确认工作区干净，再只读查询远程支撑层最新 `transcoding_jobs` 失败记录，重点区分上一轮已修的 `转码进度假失败` 与新的真实 ffmpeg/硬编失败。
- 影响文件：待定，预计涉及 Go 转码链路、`CONTEXT.md`、`plan.md`
- 验证：待执行定向 Go 测试、`go vet ./...`、`git diff --check`、乱码扫描。

## 2026-07-02 17:41 +0800
- 进度：完成转码队列失败记录排查与修复。只读查询远程支撑层发现最近多条失败记录为 `ffmpeg progress read failed: read |0: file already closed`，关联视频多已 `ready` 且进度 100%，判定为 ffmpeg 进度管道读取顺序导致的假失败，而非媒体资产不可用。修复 `TranscodeVideo` 的等待顺序：先等进度读取自然结束，再 `Wait` 回收 ffmpeg 进程，避免 `Wait` 抢先关闭 stdout pipe。同步在 `CONTEXT.md` 增加 `转码进度假失败` 排障口径。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProgress|TestBuildPlayback|TestResolveProbe|TestChooseTranscodeOutputProfile|TestBuildTranscodeOutputTempPath' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go vet ./...` 通过；`git diff --check -- pkg/ffmpeg/ffmpeg.go CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' pkg/ffmpeg/ffmpeg.go CONTEXT.md plan.md` 无输出；`go test ./... -count=1` 仅 `internal/handlers TestAdminEd2kDownloadStatusIsRedacted` 失败，返回 `level":"down"` 且脚本 `signal: killed`，与本次 ffmpeg 修复无关。待提交。

## 2026-07-02 17:37 +0800
- 进度：开始排查“转码队列失败记录”。先确认适用规则、现有术语和 Go 后端队列实现；工作区已有未跟踪 `package.json`，按用户既有改动处理，不纳入本次提交。下一步定位失败记录来源、复现根因并做最小修复。
- 影响文件：待定，预计涉及转码队列 Go 后端、`CONTEXT.md`、`plan.md`
- 验证：待执行定向 Go 测试、`go test ./... -count=1`、`go vet ./...`、`git diff --check`、乱码扫描。

## 2026-07-02 +0800（复审修复轮）
- 进度：子代理独立复审海报墙重设计 diff，发现 1 BLOCKER + 2 IMPORTANT + 2 MINOR，已全部修复。**BLOCKER（首屏自动聚焦首卡只缩放不抬影）**：中性落影的 `.onFocusChanged` 原放在 `.tvFocusableScaleOnly` **之后**，它观察的是其后第一个焦点目标（`.clickable` 的焦点节点）；而 `firstItemFocusRequester.requestFocus()` 解析到链上第一个可聚焦目标（`tvFocusableScaleOnly` 内部的 `.focusable()`），程序聚焦落在该祖先上不向下传播到 `.clickable`，导致首屏自动聚焦的首卡 `isCardFocused=false`、落影 elevation=0——首卡只缩放不抬影，DPad 移开再回来才正常。修复：`.onFocusChanged` 移到 `.tvFocusableScaleOnly` **之前**，借 `hasFocus` 捕获任意子树焦点（程序聚焦 + DPad 聚焦都覆盖）。**IMPORTANT（顶行聚焦落影裁切顶部栏）**：`gridTopPaddingDp=16` 只覆盖了缩放垂直溢出（12.38dp），没算中性落影（`clip=false`，10dp 软影溢出节点边界）→ 顶行聚焦卡的落影顶边被顶部栏裁。修复：`gridTopPaddingDp` 16f→24f，护栏测试改为断言 `gridTopPaddingDp >= 缩放溢出 + posterWallFocusedShadowElevationDp`（12.38+10=22.38 ≤ 24）。**IMPORTANT（占位卡丢失标题）**：原设计占位卡（无海报）只显示「暂无海报」、不显示真实标题，导致无海报的卡无法识别剧集。修复：scrim+标题对所有卡都渲染——占位卡的「暂无海报」图标+标签由外层 Box `contentAlignment=Center` 居中，真实标题压底部遮罩，两者空间分离不堆叠。**MINOR（标题上行可读性）**：原遮罩 `fillMaxHeight(0.45f)` 在 ~178dp 卡上盖底部 80dp，2 行标题上行落在遮罩 ~0.35 alpha 区偏弱。修复：遮罩提到 `fillMaxHeight(0.55f)`，上行落在更强 alpha 区。**MINOR（测试脆弱）**：shared-element 测试字符串切片仍保留（锁实现形态是 spec 测试职责，可接受）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`（`.onFocusChanged` 前置 + `gridTopPaddingDp` 24f + 遮罩 0.55f + 占位卡渲染标题）、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`（顶行护栏含落影 elevation）、`CONTEXT.md`（契约补 `.onFocusChanged` 前置语义 + 顶行 padding 含落影 + 占位卡保留标题 + 遮罩 55%）。
- 验证：`:tv-app:testDebugUnitTest --tests TvPosterWallFocusLayoutSpecTest` BUILD SUCCESSFUL。待办：①完整 `:tv-app:testDebugUnitTest` + `:tv-app:assembleDebug`；②`git diff --check` + 乱码扫描；③子代理复审修复后 diff；④真机手测：首屏首卡有落影、顶行聚焦落影不裁、占位卡有标题、标题上行可读。

## 2026-07-01 20:20 +0800
- 进度：完成本轮 ED2K 下载工作台遗留评审问题修复并复测通过。后端把 `queued` 删除拆成显式 helper：删除前先尝试移除 asynq 中同 `TaskID` 的非 active job；若 job 已进入 active，则拒绝假删并提示刷新重试；若这条 `queued` 实际是从 `files_cleaned` 重下出来的历史任务，则删除动作改为“撤销本次重下并恢复回 files_cleaned”，不再物理删掉历史记录，从而堵住去重绕过。同步把 `files_cleaned -> retry` 改为不覆写原始 `started_at`，保证撤销重下时能完整恢复历史。前端把“创建弹窗是否继续留场”改成只看当前仍在草稿中的未收口行，不再被历史 `invalid` 结果永久卡住；同时补了“撤销重下”成功提示，避免误报成“任务已删除”。`CONTEXT.md` 已追加长期术语，但其中既有 TV 海报墙无关改动仍需在提交时精确排除。
- 影响文件：`CONTEXT.md`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/repository/ed2k_download_repository.go`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/queue ./internal/repository -run 'Ed2k|ed2k' -count=1` 通过；`cd admin-web && npm run test -- src/views/toolbox.helpers.spec.js src/views/ToolboxEd2kDownload.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`go build ./...` 通过；`go vet ./...` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' ...` 无输出。下一步做最终独立复审；若无新问题，则精确暂存并提交，且不纳入 `CONTEXT.md` 里无关 TV 海报墙差异。

## 2026-07-01 20:02 +0800
- 进度：最终独立复审回收完成。针对 `files_cleaned` 重试一致性又补了一轮：`retryEd2kDownloadTask` 在条件更新 miss 时，只有“当前已 `running`”或“当前是 `queued` 且 asynq 中同 `TaskID` 的 job 真实存在”才会按成功返回；enqueue 返回 error 时也会先回查 job 是否实际已落进 asynq，再决定按成功收口还是回滚到 `files_cleaned`，从而把“假成功”和“接口失败但任务偷偷下载”的高优先级问题收住。最新独立复审结论为 `no blocker/high findings`。下一步只做精确暂存和提交，不再变更功能逻辑。
- 影响文件：`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/handlers/router.go`、`internal/handlers/admin_orphan_file_scan_test.go`、`internal/queue/tasks.go`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/queue ./internal/repository -run 'Ed2k|ed2k' -count=1` 通过；`cd admin-web && npm run test -- src/views/toolbox.helpers.spec.js src/views/ToolboxEd2kDownload.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`go build ./...` 通过；`go vet ./...` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' ...` 无输出。提交范围仅纳入 ED2K 下载工作台相关文件；`CONTEXT.md` 中无关 TV 海报墙差异不纳入本次提交。

## 2026-07-01 19:46 +0800
- 进度：完成本轮 ED2K 下载工作台阻塞问题修复并复测通过。后端现在把 `running -> canceling -> cancelled` 取消链路改成条件更新，停止失败保持 `canceling` 并允许继续取消；只有引擎已停时才会根据残留清理结果落 `cancelled`，且 `cancelled` 补做清理成功会单独记录 `cleaned_at`。`files_cleaned` 重试链路改回“先改库 queued、再入队、入队失败回滚到 files_cleaned”，worker 也不再直接消费 `files_cleaned`，从而消掉“API 失败但任务偷偷重下”的竞态。前端移除了无效的 `deleted` 筛选，新增基于 helper 的任务聚焦规则：结果区定位、创建后自动聚焦、取消后继续处理、补做清理后回显都会在必要时自动切到目标状态筛选，避免任务在过滤态下丢失。
- 影响文件：`CONTEXT.md`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/repository/ed2k_download_repository.go`、`internal/queue/ed2k_download.go`、`internal/queue/ed2k_download_test.go`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/queue ./internal/repository -run 'Ed2k|ed2k' -count=1` 通过；`cd admin-web && npm run test -- src/views/toolbox.helpers.spec.js src/views/ToolboxEd2kDownload.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`go build ./...` 通过；`go vet ./...` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' ...` 无输出。最终独立复审进行中，待确认无新的 blocker/high 后再整理提交。

## 2026-07-01 19:32 +0800
- 进度：独立复审继续打回两条真实问题并已进入修复。其一，前端在筛选视图下“定位任务 / 取消后继续处理 / 创建成功自动聚焦”没有同步切换 filter，导致目标任务经常不可见；同时 `deleted` 仍被当成正式可筛选状态，和当前后端“物理删除、不保留 deleted 历史”的实际语义冲突。本轮把过滤与聚焦规则抽成 `toolbox.helpers` 纯函数，移除 `deleted` 筛选，并让所有会改变任务状态的动作都按目标任务状态自动切换视图。其二，`files_cleaned` 重试链路虽然此前补了条件更新，但 worker 仍允许直接消费 `files_cleaned`，导致“API 返回失败但任务偷偷重新下载”的竞态。本轮把重试时序改回“先改库 queued，再入队，入队失败则回滚到 files_cleaned”，并撤掉 worker 对 `files_cleaned` 的直接消费；同时补回归测试锁定这两条语义。
- 影响文件：`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/repository/ed2k_download_repository.go`、`internal/queue/ed2k_download.go`、`internal/queue/ed2k_download_test.go`、`plan.md`
- 验证：待执行 `gofmt -w internal/handlers/admin_ed2k_download.go internal/handlers/admin_ed2k_download_test.go internal/repository/ed2k_download_repository.go internal/queue/ed2k_download.go internal/queue/ed2k_download_test.go`、`go test ./internal/handlers ./internal/queue ./internal/repository -run 'Ed2k|ed2k' -count=1`、`cd admin-web && npm run test -- src/views/toolbox.helpers.spec.js src/views/ToolboxEd2kDownload.spec.js src/api/admin.spec.js`、`cd admin-web && npm run build`、`go build ./...`、`go vet ./...`、`git diff --check`、乱码扫描。

## 2026-07-01 16:10 +0800
- 进度：完成本轮 ED2K 下载工作台评审修复并复测通过。后端把历史 migration 改回基线，新增 `0030_ed2k_download_task_lifecycle` 增量迁移，正式承载 `canceling` / `cancelled` / `files_cleaned` 与 `cleaned_at`；`completed` 任务新增“删除暂存文件”动作，按受管相对路径删除任务目录内符号链接及其指向的 aMule 真文件，再落 `files_cleaned`，同时保留 `finished_at` 并单独记录 `cleaned_at`；`files_cleaned` 任务新增“重新下载”接口，重入 `queued` 并重新入队。创建弹窗链路同步改成 `entries[] = [{ line_number, source_link }]` 契约，前端会话内稳定分配行号并累计结果，后端按输入行逐条回 `duplicate` / `invalid` / `reused` / `create_failed` / `enqueue_failed` / `created`，从而真正满足“同次重复显式回显、无效行不阻断、首次行号绑定、改链视为新输入、部分成功留场”。无关工作区里的 TV 海报墙 `CONTEXT.md` 旧改未纳入本次实现。
- 影响文件：`migrations/0029_ed2k_download_tasks.up.sql`、`migrations/0030_ed2k_download_task_lifecycle.up.sql`、`migrations/0030_ed2k_download_task_lifecycle.down.sql`、`internal/repository/migrations_test.go`、`internal/models/admin.go`、`internal/repository/ed2k_download_repository.go`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/handlers/router.go`、`internal/queue/ed2k_download.go`、`internal/queue/ed2k_download_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolbox.helpers.spec.js src/views/ToolboxEd2kDownload.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`go test ./internal/handlers ./internal/queue ./internal/repository -run 'Ed2k|ed2k' -count=1` 通过；`go build ./...` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' ...` 无输出。

## 2026-07-01 12:04 +0800
- 进度：完成 ED2K 下载任务生命周期底座修复。`EnqueueEd2kDownload` 现在统一走 `buildEd2kDownloadTaskOptions`，携带 `asynq.TaskID(taskID)`，并把 asynq `ErrTaskIDConflict` 包装成项目侧 `ErrEd2kDownloadTaskInFlight`；手动重试路径抽成 `retryEd2kDownloadTask`，在状态更新为 `queued` 后重新入队，TaskID 冲突时向管理端返回“任务已在执行中”。`CONTEXT.md` 已移除“当前代码缺这一步”的过时表述。
- 影响文件：`internal/queue/tasks.go`、`internal/queue/tasks_test.go`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/queue -run 'TestBuildEd2kDownloadTaskOptionsBindsTaskID|TestWrapEd2kDownloadEnqueueErrorMapsTaskIDConflict'` 通过；`go test ./internal/handlers -run 'TestRetryEd2kDownloadTask'` 通过；`go test ./internal/handlers ./internal/queue` 通过；`go build ./...` 通过；`go vet ./...` 通过。

## 2026-07-01 11:43 +0800
- 进度：完成 ED2K 下载工作台列表全宽布局。页面从“左侧 24rem 列表 + 右侧详情”的双列布局改为“列表整行在上、详情整行在下”，任务行在桌面端横向展示标题/哈希与状态/大小，移动端回落为单列；同步补静态 spec 锁定列表不再回退到窄侧栏，并在 `CONTEXT.md` 追加 [[ED2K 下载工作台列表全宽]]。
- 影响文件：`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxEd2kDownload.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`git diff --check -- admin-web/src/views/ToolboxEd2kDownload.vue admin-web/src/views/ToolboxEd2kDownload.spec.js CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' admin-web/src/views/ToolboxEd2kDownload.vue admin-web/src/views/ToolboxEd2kDownload.spec.js CONTEXT.md plan.md` 无输出。

## 2026-06-29 09:55 +0800
- 进度：修复 AV 视频刮削失败后状态一直停在「刮削中」、且日本 AV 详情没有显眼逃生入口的问题。经 grill 收口确认两件事：(1) 根因是欧美 AV 自动刮削报错时 `handleScrape` western 分支直接 `return scrapeErr` 不更新状态，加上无 asynq 失败兜底，状态永远停在上传时写入的 `scraping`；日本 AV 理论走 `buildScrapeFailureDecision` 落 `uploaded`，但 worker 被 kill / 写库失败等路径到不了状态更新代码，同样卡 `scraping`。(2) 「去 AV 手动刮削」按钮其实对日本 AV 也显示，只是藏在默认折叠的「播放预览」区里，stuck 时等于没有出口。本轮两者都做：后端修根因（仅路径1），前端加顶部显眼「去刮削」逃生按钮（B：跳手动刮削页，不做后端重跑）。TDD：先后端写 `TestBuildWesternAVScrapeFailureDecisionMarksAVScrapePending`（红：helper 未定义）→ 加 `buildWesternAVScrapeFailureDecision` 纯函数 → 接入 `handleScrape` western 分支（落 `av_scrape_pending` + `scrape_attempt.error` + 空 `scrape_preview` + 不入队 transcode + return nil）→ 绿；前端写 `shouldShowStuckScrapeAction` / `buildStuckScrapeRoute` 的 4 个 spec（红）→ 实现 helper（绿）→ `VideoList.vue` 状态区接按钮 + `openStuckScrape` 分发。`CONTEXT.md` 补 `欧美 AV 刮削报错落待确认`、`卡刮削逃生按钮` 两条术语。范围外：`HandleScrapeRetag` av 分支同缺陷不改；worker 被 kill 的路径2 不做后端兜底，靠按钮人工救。
- 影响文件：`internal/queue/scrape_tasks.go`、`internal/queue/scrape_tasks_test.go`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/VideoList.vue`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/queue/ -run TestBuildWesternAVScrapeFailureDecisionMarksAVScrapePending` 通过；`go test ./internal/queue/ ./internal/services/` 通过；`go build ./...` 通过；`cd admin-web && npm test -- videoList.helpers.spec` 23/23 绿；`cd admin-web && npm test`（仅回退 ED2K 既有未提交改动后）25 文件 173 测试全绿——`toolboxPage.spec` 在含 ED2K 既有改动时的全量失败经 `git stash` 二分确认是 ED2K 既有改动引入、与本次无关；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）。

## 2026-06-28 01:09 +0800
- 进度：完成 ED2K 下载工作台“新建任务改成弹窗”的最小前端落地。`PageHeader` 右侧新增“新建任务”主按钮，页面内常驻“提交链接”卡片已移除，原提交表单迁入居中 `el-dialog`；现有列表、详情、提交 API 与成功后聚焦逻辑保持不变。
- 影响文件：`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxEd2kDownload.spec.js` 通过；`cd admin-web && npm run build` 通过（仅有既有 chunk size warning）；`git diff --check -- admin-web/src/views/ToolboxEd2kDownload.vue admin-web/src/views/ToolboxEd2kDownload.spec.js plan.md` 通过；`rg -n $'\uFFFD' admin-web/src/views/ToolboxEd2kDownload.vue admin-web/src/views/ToolboxEd2kDownload.spec.js plan.md` 无输出。

## 2026-06-27 16:30 +0800
- 进度：修复手机端「扫码登录 TV」进入横屏的 bug。根因：扫码走 zxing-android-embedded 4.3.0 默认 `CaptureActivity`，该库在其 Manifest 里声明 `sensorLandscape`，app 未注册自定义子类覆盖 → 扫码页被锁横屏；`VideoHomeApp.kt` 既有的 `setOrientationLocked(false)` 只抑制运行时 `setRequestedOrientation`，清不掉 Manifest 级方向锁。按 Plan A 修复：新增空子类 `PortraitCaptureActivity : CaptureActivity()`，在 app `AndroidManifest.xml` 声明并 `android:screenOrientation="portrait"`，`ScanOptions` 改调 `setCaptureActivity(PortraitCaptureActivity::class.java)`（删掉无效的 `setOrientationLocked(false)`）。TDD：先写 `TvAuthScanOrientationSpecTest`（三断言：扫码注册自定义 CaptureActivity、Manifest 声明 portrait、子类继承库基类且不调 setOrientationLocked）→ 红（三测全挂，因自定义 Activity/Manifest/调用都不存在）→ 绿。版本：`versionCode` 3→4 / `versionName` 0.1.2→0.1.3。`CONTEXT.md` 新增术语 `扫码登录 TV 方向锁定`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/tvauth/PortraitCaptureActivity.kt`、`android-app/app/src/main/AndroidManifest.xml`、`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/build.gradle.kts`、`android-app/app/src/test/java/com/chee/videos/feature/tvauth/TvAuthScanOrientationSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-app && ./gradlew :app:assembleDebug :app:testDebugUnitTest` → `assembleDebug` 通过；新增 `TvAuthScanOrientationSpecTest` 3/3 绿（test-results XML failures=0 errors=0）。全量 `testDebugUnitTest` 有一项 `TvCatalogViewModelTest.nullListsInPayload...` 偶发 `UncaughtExceptionsBeforeTest` 失败，但单独重跑该测试类 BUILD SUCCESSFUL——属既有协程测试跨用例泄漏的 flake，与本次扫码方向改动无关（改动未触及 TV catalog 链路）。

## 2026-06-27 16:03 +0800
- 进度：把 ED2K 下载工作台继续往后端收口。已新增 `ed2k_download_tasks` 表、管理端任务路由、任务列表/创建/详情/重试/删除接口，以及前端 API 接入；`/toolbox/ed2k-download` 现在从 `localStorage` 切到后端数据源，删除排队任务会永久清除记录，资源哈希历史命中则只附加历史，不重建任务。
- 影响文件：`internal/models/admin.go`、`internal/repository/ed2k_download_repository.go`、`internal/handlers/admin_ed2k_download.go`、`internal/handlers/admin_ed2k_download_test.go`、`internal/handlers/router.go`、`internal/repository/migrations_test.go`、`migrations/0029_ed2k_download_tasks.up.sql`、`migrations/0029_ed2k_download_tasks.down.sql`、`admin-web/src/api/admin.js`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/repository -run 'TestRegisterIncludesEd2kDownloadRoutes|TestParseEd2kDownloadLinkAndBuildTitle|TestEd2kDownloadTasksMigration'` 通过；`cd admin-web && npm test -- --run src/views/ToolboxEd2kDownload.spec.js src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）

## 2026-06-27 15:18 +0800
- 进度：ED2K 下载工作台已从静态骨架升级为可用的本地任务面板。当前支持多行链接粘贴、自动标题、按资源哈希命中历史任务、任务列表/详情切换、状态本地流转和永久删除；`/toolbox/ed2k` 仍保持原链接生成器并存，两个入口互不抢职责。
- 影响文件：`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxEd2k.spec.js src/views/ToolboxEd2kDownload.spec.js src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅有现有 bundle size warning）

## 2026-06-27 14:27 +0800
- 进度：把 ED2K 两个入口拆回并存结构。`/toolbox/ed2k` 已恢复为原来的 ED2K 链接生成器，新增 `/toolbox/ed2k-download` 作为独立下载工作台；工具箱菜单同时展示两个入口，分别用 Link 和 Download 图标区分。前端测试也拆成了链接生成器与下载工作台两个专属 spec，避免以后互相污染。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/router/index.js`、`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/ToolboxEd2kDownload.vue`、`admin-web/src/views/ToolboxEd2k.spec.js`、`admin-web/src/views/ToolboxEd2kDownload.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/ToolboxEd2k.spec.js src/views/ToolboxEd2kDownload.spec.js src/views/toolboxPage.spec.js` 与 `npm run build`

## 2026-06-27 14:03 +0800
- 进度：完成 ED2K 下载工作台前端第一刀。管理端工具菜单中的 ED2K 入口已从链接生成器切换为下载工作台，`/toolbox/ed2k` 现展示固定四段式详情骨架：来源标识、预期文件信息、状态反馈、文件区；页面可在 queued / running / failed / completed 示例状态间切换，用于验证不同状态下的区块显隐与摘要切换。新增 `ToolboxEd2k.spec.js` 覆盖骨架语义，`toolboxPage.spec.js` 也同步改写为下载工作台入口。当前仍是前端骨架，未接真实后端下载接口。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/views/ToolboxEd2k.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxEd2k.spec.js src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过

## 2026-06-27 11:20 +0800
- 进度：通过 `$grill-with-docs` 收口「短视频按返回键双按确认退出」边界并落地实现。确认范围仅 TV 端（手机端无任何视频有双按模式，「像其他视频一样」指 TV 长视频/电视剧播放器）；文案「再按一次返回」+ 第二次 `onBack` 返回首页（与长视频/电视剧逐字对齐，回避「退出」在词表里退出 App vs 退页的重载）；两条返回路径（根 Box `onPreviewKeyEvent` 的 `KEYCODE_BACK`/`KEYCODE_ESCAPE` 分支 + 顶层 `BackHandler`）共用同一 `handlePlaybackBack` 状态机，复用既有 `resolveTvPlayerBackAction`/`TvPlayerBackAction`/`TvPlayerBackConfirmWindowMillis`/`TvPlayerBackConfirmPrompt`，不新造并行实现；单条失败轻提示显示时 BACK 不拦截、直接进双按（贴合既有 `TV 短视频单条失败留在当前页`「不引入先清提示再退出中间态」契约）；失败/空态可聚焦卡片「返回首页」按钮点击保持 `onBack` 不进双按。`TvShortFeedScreenSpecTest` 加 `shortFeedScreenBackUsesSharedDoublePressConfirm` 防回归（断言含 `resolveTvPlayerBackAction`/`handlePlaybackBack`、两路径接进它、复用 `TvPlayerBackConfirmPrompt`、不出现 `backPressTime`/`lastBackPress` 旧式时间戳）。版本：`versionCode` 126→127 / `versionName` 0.1.126→0.1.127。`CONTEXT.md` 改写 `TV 短视频返回语义`（单按→双按，标注与 IPTV 分叉，回链 [[TV 播放器退出确认]]）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvShortFeedScreenSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew :tv-app:assembleDebug :tv-app:testDebugUnitTest` → BUILD SUCCESSFUL in 8s；`compileDebugKotlin` 编译干净，`testDebugUnitTest` 全绿（含新增 `shortFeedScreenBackUsesSharedDoublePressConfirm` 与既有 `TvShellAppBackPolicyTest`/`TvShortFeedScreenSpecTest`/`TvPlayerBackConfirmTest`）。

## 2026-06-27 09:42 +0800
- 进度：完成 TV 左侧菜单聚焦改静态高亮实现。按 grill 收口方案在 `TvCatalogScreen.kt` 的 `TvHomeSideMenuButton` 内联：用 `onFocusChanged` 读 `focused`，按 `focused`/`selected` 组合选 background（选中=实色金 `Accent`、聚焦未选中=`SurfaceElevated`、常态=`Surface`）与 border（聚焦=`AccentStrong` 描边，未选中 `1.dp` / 选中 `2.dp` 分档保证 10-foot 可辨），`tvFocusableScaleOnly(focusedScale = 1.06f)` 换成裸 `.focusable()`（无 scale/无按下下沉/无触觉，对齐选集轨双态高亮）；补 `BorderStroke`/`focusable`/`onFocusChanged`/`mutableStateOf`/`setValue` imports。红灯：`TvCatalogFocusPolicyTest` 加 `side menu button uses static highlight focus instead of scale` 断言（不含 `tvFocusableScaleOnly`、含 `.focusable()`/`onFocusChanged`/`BorderStroke`+`AccentStrong`/`SurfaceElevated`）。版本：`versionCode` 125→126 / `versionName` 0.1.125→0.1.126，不重建 release 固件、不碰 `tv_apk_test.go`（固件与 Go 断言仍停在 121，脱节为已知接受项）。两段式评审：两子代理并行独立评审，焦点正确性评审发现一条 BLOCKER——既有 `TvHomeNavigationTest.sideMenuButtonsUseSingleFocusableTargetSoConfirmWorksOnce` 仍锁旧模型（`assertTrue tvFocusableScaleOnly` / `assertFalse .focusable()`），与新测试矛盾致构建红；已将其两断言翻转为新模型（保留 IPTV/Shorts 顺序断言不动）。视觉评审提一条 MINOR——选中聚焦金边对金底对比偏弱，已采纳其建议把选中聚焦描边加粗到 `2.dp`。`CONTEXT.md` 新增术语 `TV 侧边菜单聚焦高亮`（点明为「只缩放」全局规则例外，回链选集轨双态高亮），并同步描边分档语义。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew :tv-app:assembleDebug :tv-app:testDebugUnitTest --tests "com.chee.videos.feature.tv.TvHomeNavigationTest" --tests "com.chee.videos.feature.tv.TvCatalogFocusPolicyTest"` → BUILD SUCCESSFUL；`--rerun-tasks` 强制重跑两测试类亦绿；评审 BLOCKER 已修复并复测通过。

## 2026-06-25 14:13 +0800
- 进度：TV 短视频信息流切条由「黑屏 + loading」改为「视频缩略图封面占位 + 交叉淡入」。需求先经 grill-with-docs 把边界收口成 7 条决策并写入 CONTEXT.md `TV 短视频封面占位` 术语；再用三个子代理并行产出方案（最小差异 / 预加载健壮 / 首帧可靠），主代理综合为「A 骨架 + B 预加载 + C 首帧守卫」并拆 5 个任务编码。改动单文件 `TvShortFeedScreen.kt`：①新增本地 `resolveThumbnailUrl(baseUrl, rawPath)`（复刻手机端，服务端 `thumbnailPath` 恒为 `/api/v1/videos/:id/thumbnail` 公开端点路径）；②`showPoster = currentVideoId 非空 && coverUrl 非空 && renderedVideoId != currentVideoId && 无错误` 单一判定覆盖切条/重试/ON_RESUME；③封面 `AnimatedVisibility`+`tween(150)` 淡入淡出、`ContentScale.Fit` 与 `RESIZE_MODE_FIT` 几何对齐，z 序 PlayerView→封面→转圈（转圈保留叠在封面之上）；④`onRenderedFirstFrame` 与 `STATE_READY` 双信号清封面，均带 `mediaId == latestCurrentVideoId` 守卫防 stale first-frame（ExoPlayer 实际 READY 先于首帧，READY 为主触发、首帧为冗余确认）；⑤预加载 `LaunchedEffect` 对当前条 ±2 窗口预热 Coil 缓存，`memoryCache` 去重 + `execute`（挂起）+ `Semaphore(2)` 真正把并发抓取钉在 2 保护外盘休眠；首帧永不触发的极端病理情况转圈持续可见（非黑屏），用户可 BACK/切条脱困，不加超时强摘（避免与 BUFFERING 转圈判定交互、保沉浸式口径）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`（版本 0.1.122→0.1.123 / 122→123）、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew :tv-app:compileDebugKotlin :tv-app:assembleDebug` BUILD SUCCESSFUL；`:tv-app:testDebugUnitTest`（TvShortFeedScreenSpecTest / TvShortFeedViewModelTest / TvShortFeedNavigationSpecTest）BUILD SUCCESSFUL 无失败；独立子代理评审结论 MERGE-READY（深入 Coil 2.7.0 源码 jar 验证预加载缓存键与 AsyncImage 共享、双信号 mediaId 守卫、z 序几何、scope 契约均正确），并按评审 2 个 SHOULD-FIX + 1 NIT 修正（`enqueue`→`execute` 让信号量名副其实、STATE_READY 注释改为正确的主/次触发框架、首帧分支补对称守卫）后复测通过。待真机/模拟器手测切条/重试/边界。

## 2026-06-25 14:13 +0800
- 进度：完成 ZIP 压缩包中文条目名编码兼容收尾。回归确认服务端能按 `auto / utf8 / gbk` 正确解码并落盘，密码分支会沿用解码后的中文路径，`needs_encoding` 批次可通过重试入口补编码或补密码，管理端摘要与表单已联动到位。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/admin_archive_import_test.go`、`internal/handlers/router.go`、`internal/models/models.go`、`internal/repository/migrations_test.go`、`migrations/0028_archive_import_batch_encoding.up.sql`、`migrations/0028_archive_import_batch_encoding.down.sql`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./... -count=1` 通过；`go vet ./internal/services ./internal/handlers ./internal/repository` 通过；`cd admin-web && npm test` 166 项通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services/archive_import.go internal/services/archive_import_test.go internal/handlers/admin_archive_import.go internal/handlers/admin_archive_import_test.go` 无命中。

## 2026-06-25 09:35 +0800
- 进度：完成 ZIP 压缩包条目名编码兼容实现并开始补严回归。服务层 `decodeArchiveZipEntryName` 落地 `auto / utf8 / gbk` 判码（合法 UTF-8 优先、仅非法 UTF-8 才回退 GBK），`planZipExtractEntries` 处理解码后路径冲突；`archiveBatchFailureStatus` 实现“自动失败→needs_encoding、首次显式失败仍可重试、第二次显式异种编码失败转 failed”的两段式纠偏；`RetryExtract` 合并密码与编码模式入口，`extractArchive` 返回最终采用编码并写入 `encoding_mode`；`encoding_requested_mode` 与 `encoding_mode` 仅对 `zip` 生效，`rar/7z` 维持既有外部解压器行为。管理端重试表单把密码与编码选择合并在同一张纠偏卡，`needs_encoding` 状态与批次摘要“解包编码”回显联动。当前补充中的回归测试会锁定 upload / retry-extract 的 handler 契约，避免返回体丢失批次数据。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`internal/models/models.go`、`internal/repository/migrations_test.go`、`migrations/0028_archive_import_batch_encoding.up.sql`、`migrations/0028_archive_import_batch_encoding.down.sql`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./...` 通过（含 `internal/services`、`internal/handlers`、`internal/repository`、`internal/queue`）；`go vet ./internal/services ./internal/handlers ./internal/repository` 无告警；`cd admin-web && npm test` 166 项通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD'` 对改动文件无命中；迁移编号 0028 紧接 0027 正确递增。

## 2026-06-24 22:28 +0800
- 进度：完成压缩包导入 UTF-8 落库报错修复。服务层现在会先把外部解压器 stderr 与批次/文件失败原因归一化为合法 UTF-8，再写入 `archive_import_batches.last_error` / `archive_import_files.reason`；新增纯单测锁定非法字节替换行为，并在 `CONTEXT.md` 沉淀“压缩包错误文本 UTF-8 落库”兼容约束。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestNormalizeArchive|TestSanitizeArchive' -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services/archive_import.go internal/services/archive_import_test.go` 无命中。

## 2026-06-24 17:56 +0800
- 进度：完成 TV 短视频迁移收尾校验。当前仅准备纳入 TV 端短视频全屏播放页、首页 `短视频` action-only 入口、独立路由/仓库接口、相关单测、TV 版本号，以及 `CONTEXT.md`、`android-tv-app/AGENTS.md`、`plan.md` 的长期文档更新；未纳入其它模块改动。
- 影响文件：`android-tv-app/AGENTS.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/*`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/AGENTS.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/main/java/com/chee/videos/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/tv` 无命中。

## 2026-06-24 17:55 +0800
- 进度：完成 TV 短视频迁移实现收尾。新增 `tv/shorts` 独立全屏短视频页与 ViewModel，首页左侧菜单在 `IPTV` 下方新增 `短视频` 入口并保持 action-only 语义；补充 `fetchShortFeed(pageSize, excludeIds)`、返回焦点与导航回退测试，同时校正 TV 模块 AGENTS 中已过时的“仅长视频”约束。
- 影响文件：`android-tv-app/AGENTS.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvShortFeedScreenSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvShortFeedViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShortFeedNavigationSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待执行最终提交与提交后状态确认。

## 2026-06-24 14:59 +0800
- 进度：完成压缩包导入多分组主链路。后端已补齐分组模型查询、创建/编辑/删除/加入/移出/处理接口与路由；管理端批次详情新增分组工作区、未分组虚拟卡片、分组创建/编辑弹窗、加入/移出分组动作与分组直处理入口，文件行同步显示分组归属。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_groups.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`internal/models/models.go`、`migrations/0027_archive_import_groups.up.sql`、`migrations/0027_archive_import_groups.down.sql`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/codex-go-cache go test ./internal/services ./internal/handlers -run 'TestArchiveImport|TestRegisterIncludesAdminPasswordVaultRoutes' -count=1` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/handlers/admin_archive_import.go internal/handlers/router.go internal/models/models.go internal/services/archive_import.go internal/services/archive_import_groups.go internal/services/archive_import_test.go admin-web/src/api/admin.js admin-web/src/api/admin.spec.js admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js migrations/0027_archive_import_groups.up.sql migrations/0027_archive_import_groups.down.sql` 无命中。

## 2026-06-24 10:35 +0800
- 进度：通过 `$grill-with-docs` 收口压缩包导入“多个分组”术语。已将 `CONTEXT.md` 中原先的“压缩包标题分组”改为更通用的“压缩包批次内分组”，统一指同一批次内按文件选择集拆分的多个处理组，不再把它限定成单纯标题覆盖。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”边界。已确认批次内分组不是临时前端选择态，而是需要持久化保存、刷新后可恢复的批次内实体。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”数据来源。已确认分组采用“分组默认值 + 文件级覆盖”模型：分组保存默认元数据，文件先继承，后续单文件修改视为覆盖，不再把分组理解成一次性直接回写所有文件。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”归属约束。已确认同一批次中的单个文件最多只属于一个分组，也允许未分组；不支持多分组同时归属。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”初始归属。已确认新解包文件默认保持未分组，不自动落到所谓默认分组；批次级默认值继续作为未分组文件的首层默认来源。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”媒体边界。已确认分组本身也禁止视频和图片混放；每个分组只能承载单一媒体类型。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”创建路径。已确认管理员应先在文件清单勾选文件，再基于当前选择集创建分组；不采用先建空分组再回填文件的主路径。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”身份字段。已确认分组需要显式命名，分组名必填；备注可选，仅用于管理员区分和回看。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”换组规则。已确认文件换组时采用“重算继承值、保留文件级覆盖”语义：未覆盖字段跟随新分组默认值，已覆盖字段继续保持文件级特例。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”删除语义。已确认删组只删除分组本身，组内文件全部回到未分组；文件级覆盖保留，未覆盖字段回退到批次级默认值。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”后续联动边界。已确认分组默认值后改会自动更新组内未覆盖的文件记录，但不反向改写已入库的视频/图片实体；分组语义继续限定在入库前工作区。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-23 14:35 +0800
- 进度：完成压缩包导入视频标题语义调整。新上传批次的视频文件默认标题直接使用视频默认标题，不再拼接压缩包内文件名；图片标题继续按文件名派生；单文件或批量编辑把视频标题清空时回到视频默认标题；管理端批量编辑视频新增标题字段用于临时标题分组；`CONTEXT.md` 已沉淀标题语义和标题分组边界。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImport' -count=1` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services/archive_import.go internal/services/archive_import_test.go admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js` 无命中。

## 2026-06-22 13:59 +0800
- 进度：完成工具箱菜单布局收尾。管理端构建、`git diff --check` 和乱码扫描均已通过；Vite 仅保留现有 chunk size 警告。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src/views/Toolbox.vue admin-web/src/views/toolboxPage.spec.js CONTEXT.md plan.md` 无命中。

## 2026-06-22 13:57 +0800
- 进度：完成工具箱菜单自适应布局实现。`Toolbox.vue` 已从无限 `auto-fit` 改为最多四列，并在 `80rem`、`64rem`、`48rem` 下逐级降为三列、两列、单列；`toolboxPage.spec.js` 增加静态回归断言；`CONTEXT.md` 补充工具箱菜单密度边界。此次只影响管理端 Web，不涉及 Android App 版本号。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js` 通过；待执行 `cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-22 11:57 +0800
- 进度：根据用户要求收敛工具箱密码管理功能为简单 CRUD。实现不再加入乐观锁、导入导出、密码生成、分类标签或二次验证；删除为确认后的直接删除，修改按最后提交覆盖。
- 影响文件：预计 `migrations/`、`internal/config`、`internal/models`、`internal/repository`、`internal/handlers`、`admin-web/src`、`CONTEXT.md`、`plan.md`
- 验证：待实现后执行后端定向测试、管理端定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-22 11:17 +0800
- 进度：根据独立审查修复压缩包视频图片关联边界。旧批次里因历史扫描继承批次默认图片合集的视频文件，处理时不再把该默认值误写成视频图片图集；单文件保存视频时只提交当前可见的唯一图片图集；上传哈希并发唯一冲突分支也会在已有视频缺少图片图集时补写，不再漏掉本次关联。测试同步覆盖默认继承忽略、显式视频关联保留和单视频单图片图集约束。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoImageCollection' -count=1` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js` 通过；`go test ./internal/repository -run 'TestNormalizeSingleImageCollectionID|TestScanVideoRecordReadsOSHash|TestGetVideoOSHash|TestUpdateVideoOSHash' -count=1` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-22 11:07 +0800
- 进度：完成压缩包导入页回归修复。批次详情文件勾选位现在始终显示可见边框与选中态；Shift 连选改为“未选目标时加入区间、已选目标时移除区间”；视频文件支持设置唯一图片合集，批量编辑也可统一覆盖“视频关联的图片合集”，后端处理视频时把该图片合集写入视频关联，命中已有视频且原本未关联图片合集时补上但不覆盖既有关联。批次默认图片合集继续只继承到图片文件，避免上传默认值误把所有视频挂到同一图片图集。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`internal/repository/video_repository.go`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js` 通过；`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoImageCollection' -count=1` 通过；`go test ./internal/repository -run 'TestNormalizeSingleImageCollectionID|TestScanVideoRecordReadsOSHash|TestGetVideoOSHash|TestUpdateVideoOSHash' -count=1` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-21 17:30 +0800
- 进度：完成压缩包导入视频源路径修复。`ProcessFile` 现会先把视频复制到 `UPLOAD_TEMP_DIR/archive-import-videos/<batch>/<file>/原始文件名` 稳定路径，再创建/复用视频记录；不再把会被立即删除的 work 临时文件写进 `videos.original_path`。同时补了“命中已有失败视频时修正 `original_path` 并允许重新入队”的兼容逻辑，避免旧批次失败项继续卡死；新增服务层纯逻辑测试锁定稳定源文件命名、存储位置和失败旧视频可重试边界。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`internal/repository/video_repository.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoSource' -count=1` 通过；`go test ./internal/handlers -run 'TestSelectRetranscodeInputPath' -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' internal/services/archive_import.go internal/services/archive_import_test.go internal/services/upload.go internal/repository/video_repository.go CONTEXT.md plan.md` 无命中。

## 2026-06-21 15:26 +0800
- 进度：完成压缩包导入页的批次优先 UI 重构收尾。页面已切换为“批次列表主视图 + 上传弹窗 + 批次详情抽屉 + 批量编辑弹窗”，批次详情内文件支持勾选、Shift 连选、按类型快捷选择、批量编辑与批量处理；单文件编辑降级为仅在单选时出现的例外项精修区，不再承载处理动作。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/toolboxPage.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js admin-web/src/views/toolboxPage.spec.js CONTEXT.md plan.md` 无命中。

## 2026-06-21 10:46 +0800
- 进度：修复 `internal/services` 全量测试遗留失败。`TestParseTVAPKMetadataParsesReleaseAPK` 断言 release APK 固件为 `versionCode=80 / versionName=0.1.80`，但 `android-tv-app/tv-app/release/` 下的固件已于 6-19 重建到 121 / 0.1.121，导致 `go test ./internal/services` 全包失败。本次仅同步测试断言到 121 / 0.1.121，不改解析逻辑或固件；这是测试断言随 release 固件版本走的已知耦合（已沉淀到 CONTEXT.md）。
- 影响文件：`internal/services/tv_apk_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`GOCACHE=/private/tmp/codex-go-cache go test ./internal/services/ -count=1` 提权运行（httptest 用例需绑本地端口）结果 `ok video-server/internal/services`；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-21 10:32 +0800
- 进度：完成压缩包导入批次删除能力。后端新增批次删除接口，限制 `processing` 批次不可删；删除时只清理批次记录、文件清单、原压缩包副本与解包工作目录，不回删已入库的视频/图片实体。管理端批次列表同步新增删除按钮和二次确认文案。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImportShouldProcessInBatch|TestValidateArchiveBatchDeletion|TestArchiveBatchCleanupPaths'` 通过；`cd admin-web && npm run test -- src/api/admin.spec.js src/views/ToolboxArchiveImport.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-20 22:45 +0800
- 进度：修复压缩包导入批量处理候选范围。`ProcessAllFiles` 不再只处理字面 `pending`，而是覆盖可入库视频/图片中仍未成功入库的 `pending`、`failed`、`processing` 项；复制工作文件、视频 hash 等早期失败也会回写为 `failed`，避免文件卡在 `processing` 后无法被批量重试。同步补充服务层候选规则测试和 `CONTEXT.md` 语义沉淀。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run TestArchiveImportShouldProcessInBatch` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`go test ./internal/services` 仍因既有 `tv_apk_test.go` 期望版本 80、当前 APK 元数据 121 不一致失败。

## 2026-06-20 22:32 +0800
- 进度：完成压缩包导入交互语义落地。压缩包导入页已支持在默认视频合集、默认图片合集、视频文件合集、图片文件合集四个位置原地新建合集并自动选中；文件清单和详情改为显示中文媒体类型加格式/MIME，并展示跳过原因；批量处理逻辑未改，仍只处理视频/图片待处理项。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-20 22:21 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包批量处理范围。已确认“处理待处理文件”只处理可入库的视频和图片待处理项；目录、不支持项、嵌套压缩包等跳过项继续留在清单里展示类型、状态和原因，不参与批量处理。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 admin-web 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-20 22:11 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包标签语义。已确认压缩包导入里的标签只作为视频标签生效，默认标签和视频文件级标签支持选择已有标签或输入新标签；图片文件不新增标签体系，继续通过图片合集归档。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 admin-web 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-20 22:02 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包文件清单显示语义。已确认“待处理文件类型一定要显示”指中文媒体类型加具体格式/MIME 两层信息，例如“视频 · mp4”“图片 · jpeg”“不支持 · application/pdf”，不能只显示内部英文枚举或只显示文件大小。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 admin-web 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-20 21:52 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包上传导入体验。已确认“压缩包合集新建回填”应改为在压缩包导入页原地新建视频合集或图片合集，并在创建成功后自动回填到当前合集选择器，不再要求管理员先跳转合集管理页手动创建后再回来选择。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 admin-web 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-20 20:03 +0800
- 进度：压缩包导入页“已选压缩包却被判定为空”的问题已修复并验证。`el-upload` 现在通过显式 `file-list` 维护选中文件，上传时只读取组件自己的 `uploadFiles` 状态，不再依赖内部实例字段；静态断言已补，防止以后回退到原来的假阴性写法。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；乱码扫描无命中。

# 2026-06-20 20:00 +0800
- 进度：修复压缩包导入页“已选中文件却提示请选择一个压缩包文件”的上传状态 bug。`el-upload` 现在显式绑定 `v-model:file-list`，选择/移除时同步维护 `uploadFiles`，提交时只读取这份显式状态，不再依赖组件实例上的内部字段；同步补了静态断言，防止以后回退到现查实例的写法。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run build`、`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js`、`git diff --check`、乱码扫描。

## 2026-06-20 15:08 +0800
- 进度：完成压缩包导入页的字段收口。默认标签、文件级标签改成与上传中心一致的可选可输多选；默认视频/图片合集和文件级合集改成按名称选择的远程选择器，不再向管理员暴露 JSON 数组或合集 UUID；文件详情回显与批次处理后状态都做了数组规范化。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run build` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；待执行 `git diff --check`、乱码扫描和提交。

## 2026-06-20 13:40 +0800
- 进度：修复压缩包导入页的壳层留白和自适应断点。已恢复工具页标准 `tool-workspace__inner` 宽度与 padding，桌面端双栏在 `80rem` 以下提前收成单列，移动端继续使用更紧凑的间距；页面结构与功能逻辑未变。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`plan.md`
- 验证：`cd admin-web && npm run build` 通过（仅保留现有 chunk size 警告）。

## 2026-06-20 13:24 +0800
- 进度：完成压缩包导入页的纯 UI 美化。页面从原先的三段堆叠调整为“顶部说明 + 上传入口/概览 + 批次列表/批次详情”双层工作区，上传面板、批次卡片、文件列表和错误提示都做了视觉收紧；未改动任何上传、处理、刷新、保存相关逻辑。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`plan.md`
- 验证：`cd admin-web && npm run build` 通过（仅保留现有 chunk size 警告）。

## 2026-06-20 00:36 +0800
- 进度：完成压缩包导入管理端收尾。新增的无壳层工具页已可从工具箱打开，前端静态断言覆盖了新页、路由、API、命令面板和日期显示约定；`ToolboxArchiveImport.vue` 已去掉 Vue 2 filter 写法并把文件级合集输入规范化。同步在 `CONTEXT.md` 补了“压缩包去重保留现有资产”术语，避免后续误把去重理解成覆盖资产。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test` 通过；`cd admin-web && npm run build` 通过（仅有 chunk size 警告）；`GOCACHE=/private/tmp/ai-video-server-gocache go test ./internal/...` 通过大部分包，但 `internal/services/tv_apk_test.go` 仍因预期版本 80 与当前 `android-tv-app/tv-app/build.gradle.kts` 的 121 / `0.1.121` 不一致而失败，属于现存偏差。

## 2026-06-21 09:41 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包导入页的信息层级。已确认该页应改为“批次优先视图”：首屏先服务已有批次处理，上传新压缩包降为次级入口并改用独立弹窗承载，不再让上传大表单长期占据主视觉区域。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 09:46 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包导入页的文件工作流。已确认文件详情改为按需打开的抽屉，不再与文件清单常驻并排；同时文件清单必须继续支持勾选、Shift 连选与批量处理，不能因为详情抽屉化而退回单文件操作。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 09:49 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包文件清单的信息主次。已确认文件清单默认应优先突出“原始相对路径 + 当前状态 + 文件类型”这组处理决策信息；文件大小、跳过原因等保留可见，但降为次级视觉层级，不再与主信息同权堆在一行。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 09:56 +0800
- 进度：通过 `$grill-with-docs` 校正压缩包导入页的抽屉层级。已确认应改为“批次详情抽屉”，即点击批次列表项后打开承载批次概览与文件清单的抽屉；批次中的文件不再单独打开详情抽屉，但文件勾选、Shift 连选与批量处理仍保留在批次抽屉里作为主路径。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:00 +0800
- 进度：通过 `$grill-with-docs` 继续收口批次详情抽屉内的操作范围。已确认批次详情里的文件不仅要支持勾选和批量处理，还要支持基于当前选择集做批量编辑；这意味着后续实现不能只保留单文件编辑表单，必须补出批量字段修改入口。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:03 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包批量编辑字段边界。已确认首轮批量编辑只开放说明、标签、视频类型、视频合集、图片合集，不支持批量改标题；标题继续保留为文件级确认字段，避免整批误覆盖命名。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:06 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包批量编辑的选择集约束。已确认批量编辑不接受视频和图片混合选择；若当前勾选同时包含两类媒体，批量编辑入口应禁用并提示管理员先按媒体类型拆开处理。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:10 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包批量处理的媒体边界。已确认批量处理也不接受视频和图片混合选择；若当前勾选同时包含两类媒体，入口应禁用并明确提示管理员先按媒体类型拆开处理。代码现状的所选逐个处理逻辑后续需要补这一层门禁。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:09 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包文件清单的选择辅助。已确认批次详情里的文件清单应补“全选视频”“全选图片”“清空选择”等按类型快捷选择入口，用来配合已收口的“批量编辑/批量处理禁止媒体混选”规则，避免管理员只能靠手工逐项勾选拆分选择集。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:13 +0800
- 进度：通过 `$grill-with-docs` 继续收口批次详情抽屉内单文件编辑的职责。已确认单文件编辑只作为“例外项精修”入口存在，仅当当前恰好选中 1 个文件时才显示；多选或未选中时隐藏，让批量编辑与批量处理保持主路径地位。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:15 +0800
- 进度：通过 `$grill-with-docs` 继续收口压缩包批量编辑的承载方式。已确认批量编辑应使用独立弹窗，而不是常驻在批次详情抽屉里或继续叠一层侧边抽屉；批次详情抽屉维持“看批次、选文件、批量处理”的主工作区定位。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:17 +0800
- 进度：通过 `$grill-with-docs` 继续收口单文件编辑与处理动作的分工。已确认单文件编辑区只负责保存文件级元数据，不再自己承载“处理文件”按钮；处理动作统一留在批次详情抽屉主动作区，单选时再自然退化成“处理当前文件”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:18 +0800
- 进度：通过 `$grill-with-docs` 继续收口上传成功后的落点。已确认上传弹窗成功创建新批次后，界面应自动关闭弹窗、刷新批次列表并直接打开新批次的详情抽屉，保持“上传后立即进入处理”的连续工作流。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:23 +0800
- 进度：通过 `$grill-with-docs` 继续收口上传后自动切换与未保存修改的冲突。已确认若当前批次详情里仍有未保存的文件级修改，系统在切到新上传批次或其它批次前必须先显式确认是否丢弃，不能为了自动直达新批次而静默丢改动。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:25 +0800
- 进度：通过 `$grill-with-docs` 继续收口同批次内文件切换的未保存保护。已确认若当前单文件编辑区存在未保存修改，管理员切换到另一个文件前也必须先显式确认是否丢弃；现有 `selectFile()` 直接切换的实现后续需要补 dirty/snapshot 门禁。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-21 10:29 +0800
- 进度：通过 `$grill-with-docs` 继续收口未保存确认的重复提示策略。已确认压缩包导入里的未保存修改确认不提供“本次不再提示”之类的临时放行；只要改动仍未保存，后续每次文件切换、批次切换或视图关闭都要重新确认。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建；待进入实现后补 `admin-web` 定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-19 19:29 +0800
- 进度：按用户取消要求，收口并删除 `CONTEXT.md` 中所有压缩包上传/导入沉淀，停止继续推进该方向的设计。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `git diff --check`、`rg -n \"压缩包导入|压缩上传|zip|rar|7z\" CONTEXT.md plan.md` 确认仅保留必要历史。

## 2026-06-19 11:08 +0800
- 进度：完成 `TV 单片长视频播放器软准备` 首轮落地。单片页已新增“首帧后中心轻态”状态机：重试时立即进入 `正在重试播放`，成功后给出 `已恢复播放`，`BACK` 可取消当前重试并显示 `已取消重试`，失败态下沉到播放器中心区并保留“重试播放 + 诊断信息”；诊断面板主动作改为 `返回`，仅回到失败轻态。Media3 单片播放器已改为复用同一 `ExoPlayer` 实例重 prepare，接入 `onRenderedFirstFrame()`，并支持通过取消 key 停掉当前 prepare 以配合 latest-wins 丢弃过期结果。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetryLogicTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:compileDebugKotlin` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvLongFormPlayerSoftRetrySpecTest --tests com.chee.videos.feature.tv.TvLongFormPlayerSoftRetryLogicTest --tests com.chee.videos.feature.tv.TvLongFormMedia3PlayerTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:assembleDebug` 通过；待执行 `git diff --check` 与乱码扫描。

## 2026-06-17 17:53 +0800
- 进度：完成本轮 TV 剧集播放器软准备收尾验证。TV 全量单测、`assembleDebug`、`git diff --check` 与乱码扫描均通过；当前提交面只包含本轮软准备状态机/中心反馈/测试支撑/版本递增以及 `CONTEXT.md`、`plan.md` 的长期沉淀，不包含无关模块改动。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvSeriesPlayerPreparingStateSpecTest.kt`、`TvSeriesEpisodeRailSpecTest.kt`、`TvSeriesMixedPlaybackControlsSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlaySpecTest.kt`、`TvLongFormTitleOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest --tests com.chee.videos.core.ui.TvSeriesCorePlaybackOverlaySpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无命中。

## 2026-06-17 17:50 +0800
- 进度：完成 `TV 长视频播放器软准备` 首轮实现。电视剧播放器已拆分目标分集与实际播放分集，切集 preparing/failed/success/cancel 统一降到播放器中心层；已有播放内容时切集不再清空 `currentVideoId/currentSourceUrl` 或退回整页 loading/error，失败/取消后 rail 和标题立即回到实际播放分集，`BACK` 在 preparing 时先取消目标、在失败态时先关闭失败中心态。同步让主动切集跳过续播卡，TV 版本递增到 `0.1.119` / `versionCode=119`，并补定向测试与源码结构断言。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvSeriesPlayerPreparingStateSpecTest.kt`、`TvSeriesEpisodeRailSpecTest.kt`、`TvSeriesMixedPlaybackControlsSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：已通过 `cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest --tests com.chee.videos.core.ui.TvSeriesCorePlaybackOverlaySpecTest`；待执行 TV 全量单测、`assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 14:33 +0800
- 进度：完成 TV 详情页软刷新收尾验证。已修正一条既有源码结构断言的字符串匹配过宽问题，避免把新加的页内紧凑错误条误判成 Material `IconButton`；随后串行重跑 TV 全量单测与 `assembleDebug` 均通过。`git diff --check` 与乱码扫描也已通过，当前实现面仅纳入本轮详情页软刷新、测试支撑、版本递增与文档沉淀。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无命中。

## 2026-06-17 14:32 +0800
- 进度：全量 TV 单测第一次收尾时发现一条既有源码结构断言过度依赖裸字符串：`TvSeriesDetailActionSpecTest` 把任何 `IconButton(` 与 `.tvFocusableGlow(` 子串都视为违规，导致新增的 `TvIconActionButton(` 和页内紧凑错误条焦点按钮被误报。已把断言收窄到真正要禁止的默认 Material 按钮调用与旧卡片 glow 形态，不改变本次详情页业务实现；接下来串行重跑 TV 全量单测与 assemble，避开并行 Gradle 中间产物争用。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`plan.md`
- 验证：待串行执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 14:31 +0800
- 进度：完成 TV 详情页软刷新实现。电视剧详情页与长视频详情页都改为“首屏无内容才走整页 loading/error，已有内容时只走页内轻量刷新/错误”；两类详情 ViewModel 新增 `refreshing` 与请求版本保护，避免旧响应覆盖新详情；电视剧详情页软刷新成功后保留当前季/集，失效时按既定链路回退；两个详情页播放按钮都改为只在首屏首次有内容时请求一次初始焦点，后续软刷新不再抢回焦点。同步补充详情页 ViewModel 定向单测、挂载点/焦点源码结构断言，并将 TV 版本递增到 `0.1.118` / `versionCode=118`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`、`TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailViewModelTest.kt`、`TvSeriesDetailSoftRefreshSpecTest.kt`、`TvLongFormDetailSoftRefreshSpecTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailViewModelTest --tests com.chee.videos.feature.detail.DetailViewModelTest --tests com.chee.videos.feature.tv.TvSeriesDetailSoftRefreshSpecTest --tests com.chee.videos.feature.tv.TvLongFormDetailSoftRefreshSpecTest` 通过；待执行 TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 12:08 +0800
- 进度：完成 IPTV 软刷新核心实现。`TvIptvViewModel.reload()` 改为区分首屏加载与已有数据软刷新：已有频道或当前播放存在时只进入 `refreshing`，不退回整页 `loading`；刷新成功优先按当前 `channelId` 续留当前频道，只有当前频道失效时才回退默认首台；刷新失败时若旧列表仍可用则保留当前播放和旧列表，仅更新错误提示；并新增请求版本保护，避免旧刷新响应覆盖新状态。同步补充 IPTV 定向单测与纯函数测试，TV 版本递增到 `0.1.117` / `versionCode=117`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`TvIptvNavigationPolicyTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvViewModelTest --tests com.chee.videos.feature.tv.TvIptvNavigationPolicyTest` 通过；待执行 TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 11:57 +0800
- 进度：完成本轮 IPTV 软刷新收尾验证。TV 全量单测通过，文本检查通过；首次把 `:tv-app:testDebugUnitTest` 与 `:tv-app:assembleDebug` 并行跑时，`assembleDebug` 曾在 `:tv-app:kaptGenerateStubsDebugKotlin` 命中 `R.jar` `ClasspathEntrySnapshotTransform` `Check failed`，随后串行单独重跑 `:tv-app:assembleDebug` 已通过，判断为并行 Gradle 任务争用中间产物导致的瞬时构建问题，不是本次 IPTV 状态机回归。当前提交面仅纳入 IPTV 软刷新状态机、相关测试、版本递增与文档记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`TvIptvNavigationPolicyTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvViewModelTest --tests com.chee.videos.feature.tv.TvIptvNavigationPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app/...` 无命中。

## 2026-06-17 11:00 +0800
- 进度：完成 Android TV App 第五轮评审优化收尾。本轮继续以“运行流畅”为首要目标，海报墙刷新/切换排序改为软更新：已有内容保持可见，更新期间只显示网格内行内状态，失败时在网格内给出紧凑错误条并暂停自动分页，避免空白/跳页/混排；独立复审指出软更新完成后首焦点会因 `refreshing` 变化回跳首图，已改为仅在首批内容首次到达时自动聚焦一次，软刷新/软排序不再抢焦点。TV 版本递增到 `0.1.116` / `versionCode=116`，并沉淀 `TV 海报墙软更新` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`TvPosterWallFocusLayoutSpecTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPosterWallViewModelTest --tests com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 10:35 +0800
- 进度：完成 Android TV App 第四轮评审优化收尾。本轮以“运行流畅”为首要目标，修复 TV 首页搜索输入每个字符立即远程请求、搜索页被页面级 loading 替换的问题：搜索输入改为短 debounce，等待和网络请求期间保留搜索页并显示行内搜索状态；搜索失败在搜索页内显示错误与重试，不再误显示为空结果；重试会取消待发 debounce，避免重复请求。独立复审指出的搜索失败无页内重试入口、旧搜索响应测试覆盖不足均已修复。TV 版本递增到 `0.1.115` / `versionCode=115`，并沉淀 `TV 搜索输入流畅性` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`TvCatalogFocusPolicyTest.kt`、`TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest` 曾因搜索立即请求/页面 loading 失败；修复后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest --tests com.chee.videos.feature.tv.TvHomeNavigationTest --tests com.chee.videos.feature.tv.TvCatalogFocusPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 09:38 +0800
- 进度：完成 Android TV App 第三轮评审优化收尾。本轮聚焦 TV 首页目录/搜索旧请求覆盖：类型化首页与搜索请求均带目录请求身份，切换菜单、进入搜索/设置/IPTV、清空搜索或连续输入时会废弃旧请求，旧首页/搜索成功或失败不再覆盖当前菜单、当前查询、列表内容或错误态。独立复审指出真实 IPTV 入口绕过 `selectMenu(Iptv)`，已修为 IPTV 点击先通知 ViewModel 离开目录状态域再导航，且不把首页状态落成 IPTV 选中。TV 版本递增到 `0.1.114` / `versionCode=114`，并沉淀 `TV 首页目录请求身份` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`TvHomeNavigationTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest` 曾因旧请求覆盖失败；修复后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest --tests com.chee.videos.feature.tv.TvHomeNavigationTest --tests com.chee.videos.feature.tv.TvCatalogFocusPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 09:07 +0800
- 进度：完成 Android TV App 第二轮评审优化收尾。本轮聚焦电视剧播放器分集播放源准备失败恢复：源地址生成或偏好读取异常不再让播放页卡在“正在准备当前分集”，而是进入可重试的“暂不能播放”状态；保留当前分集身份，旧请求不覆盖新分集。提交仅纳入本轮电视剧播放源失败恢复、TV 版本递增、长期约束沉淀和计划记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-17 09:06 +0800
- 进度：完成第二轮核心修复。红灯用例确认电视剧播放器在播放源生成失败时无法进入可重试状态；现已让分集播放源准备协程捕获非取消异常，按请求序号和当前分集校验后退出 `playbackPreparing`，保留当前分集 `videoId`、清空播放 URL，并显示可重试的阻断文案。旧分集请求仍不会覆盖新分集。TV 版本递增到 `0.1.113` / `versionCode=113`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 通过；待执行相关定向组合、TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 00:25 +0800
- 进度：完成 Android TV App 整体评审第一轮优化收尾。本次聚焦“功能正常、使用流畅”的首启连接链路：异常可恢复、连接不可重入、长地址不挤压操作按钮；未改播放器内核、导航结构或 TV 编译边界。提交仅纳入本次 TV 连接优化、版本递增、长期约束沉淀和计划记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionViewModel.kt`、`ConnectionScreen.kt`、`core/repository/ServerRepository.kt`、`core/di/TvRepositoryModule.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.connection.ConnectionViewModelTest --tests com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-17 00:23 +0800
- 进度：完成第一轮 TV 首启连接链路优化。连接 ViewModel 通过 `ConnectionServerRepository` 接口隔离仓储，手动连接和历史地址连接均捕获探测/激活异常并恢复 `connecting=false`；连接中重复点击历史/发现地址会被忽略，连接页地址行限制单行省略，避免长 URL 挤压遥控器操作按钮。TV 版本递增到 `0.1.112` / `versionCode=112`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionViewModel.kt`、`ConnectionScreen.kt`、`core/repository/ServerRepository.kt`、`core/di/TvRepositoryModule.kt`、`android-tv-app/tv-app/build.gradle.kts`、相关连接单测、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.connection.ConnectionViewModelTest --tests com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest` 通过；待执行 TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-16 23:39 +0800
- 进度：完成 TV 长视频播放器内部黑色背板修复。`TvLongFormMedia3Player` 不再在 SurfaceView 背后加载详情/剧集海报，也不再使用页面渐变背板；所有长视频 Media3 播放器内部统一铺纯黑背板，避免 DV/HDR SurfaceView 透明、未铺满或重建时露出后方背景。DV/HDR 仍使用 SurfaceView，普通 SDR 仍按路由使用 TextureView。TV 版本递增到 `0.1.111` / `versionCode=111`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormExoPlayerSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.tv.TvLongFormPlayerNavigationSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 23:23 +0800
- 进度：完成 TV 长视频 DV 退出后延迟闪黑修复。`TvShellApp` 对单片长视频播放器路由和电视剧分集播放器路由禁用 NavHost 目的页 enter/exit/pop 过渡，避免播放器 SurfaceView/ExoPlayer 在详情页背后延迟释放；保留 DV/HDR 使用 SurfaceView，普通 SDR 使用 TextureView 的输出层策略。TV 版本递增到 `0.1.110` / `versionCode=110`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvLongFormPlayerNavigationSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.tv.TvLongFormPlayerNavigationSpecTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 23:03 +0800
- 进度：完成 TV 长视频 DV 异色与退出闪黑联合修复。`TvPlaybackRoute` 增加输出层语义：普通长视频使用 TextureView，DV source/output 使用 SurfaceView 保持系统 HDR/DV 直出；`TvLongFormMedia3Player` 在 SurfaceView 模式内部铺当前详情/剧集海报背板，避免 SurfaceView 销毁瞬间露黑。没有恢复长视频 LibVLC，也没有新增播放器内核选择。TV 版本递增到 `0.1.109` / `versionCode=109`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`TvLongFormMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 layout/测试、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 22:50 +0800
- 进度：完成 TV 长视频退出播放闪黑修复。长视频 ExoPlayer 不再使用 `PlayerView` 默认 `SurfaceView`，改为专用 `tv_long_form_media3_player_view.xml` 固定 `surface_type="texture_view"`、透明 shutter 并保持 reset 内容；这是输出层合成策略调整，不新增 LibVLC/ExoPlayer 选择，也不恢复长视频 LibVLC 分支。TV 版本递增到 `0.1.108` / `versionCode=108`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_long_form_media3_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormExoPlayerSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 21:34 +0800
- 进度：完成 TV App 长视频统一 ExoPlayer 迁移收尾。独立复审确认重试/player 重建 prepare 与生命周期暂停即时历史上报风险已修复，未发现阻塞或非阻塞问题；本次提交只纳入 TV 长视频播放器迁移、相关 TV 单测、TV 版本递增、`CONTEXT.md` 技术沉淀和 `plan.md` 记录。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、相关 TV 单测、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvLongFormMedia3PlayerTest --tests com.chee.videos.core.ui.LongFormMedia3PlaybackBoundarySpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md docs/adr` 无命中。

## 2026-06-16 21:31 +0800
- 进度：根据独立评审修复 TV 长视频 ExoPlayer 迁移后的两个回归风险。`TvLongFormMedia3Player` 现在把已 prepare 的 source key 与当前 player 实例绑定，避免 token 刷新或重试重建 ExoPlayer 后跳过 `setMediaSource/prepare`；Media3 生命周期 `ON_PAUSE` 会把最新 snapshot 回调给单片和剧集页并立即上报历史，降低 Home/后台时续播进度丢失风险。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关 TV 单测、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvLongFormMedia3PlayerTest --tests com.chee.videos.core.ui.LongFormMedia3PlaybackBoundarySpecTest` 通过；待执行全量 TV 单测、assemble 与复审。

## 2026-06-16 21:12 +0800
- 进度：完成 TV Shell 长视频播放器核心迁移实现。`TvLongFormPlayerScreen` 与 `TvSeriesPlayerScreen` 已改为统一 `TvLongFormMedia3Player` + `TvSeriesCorePlaybackOverlay`，删除运行路径上的 LibVLC `MediaPlayer/addSlave/access_token query` 分支；播放路由收口为 `EXOPLAYER/BLOCKED`；Media3 字幕 URL 不再拼 token；TV 版本递增到 `0.1.107` / `versionCode=107`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：定向红灯已确认；核心定向单测 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvLongFormMedia3PlayerTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest --tests com.chee.videos.core.ui.LongFormMedia3PlaybackBoundarySpecTest` 通过；待执行 TV 全量单测和 assemble。

## 2026-06-16 20:53 +0800
- 进度：完成三个子代理只读审视结果汇总，进入红灯测试阶段。迁移实现将按 TV Shell 实际入口收口：`TvLongFormPlayerScreen`、`TvSeriesPlayerScreen` 统一走 Media3/ExoPlayer；`TvIptvScreen` 继续保留 LibVLC；非 manifest 启动入口里的旧手机式播放器不纳入本次运行路径迁移。
- 影响文件：`plan.md`，预计随后修改 TV 播放路由、长视频播放器 screen、Media3 组件、相关 TV 单测与 TV 版本文件。
- 验证：待先补反向源码级/策略单测并执行定向红灯。

## 2026-06-16 20:40 +0800
- 进度：完成 TV 长视频统一迁到 ExoPlayer 的 `$grill-with-docs` 收口并准备单独提交文档决策。本次只纳入 `CONTEXT.md`、ADR 0004/0012 状态更新、新增 ADR 0013 和 `plan.md`；不开始播放器代码实现。下一步在该文档提交之后进入实现阶段，按已确认节奏先补测试约束，再迁移单片和剧集播放器。
- 影响文件：`CONTEXT.md`、`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`、`docs/adr/0012-tv-series-shared-controls-for-mixed-playback-engines.md`、`docs/adr/0013-tv-long-form-exoplayer-unification.md`、`plan.md`
- 验证：`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md docs/adr` 无命中；grill 文档阶段不执行 Android 构建。

## 2026-06-16 20:12 +0800
- 进度：确认为 TV 长视频统一迁到 ExoPlayer 记录 ADR。新增 `docs/adr/0013-tv-long-form-exoplayer-unification.md`，说明长视频统一 ExoPlayer、IPTV 继续 LibVLC、接受 ASS/SSA 不再 libass 保真、不做双轨灰度和不改手机端；同时将 ADR 0004 标记为被 0013 取代，将 ADR 0012 标记为混合内核前提被 0013 更新。
- 影响文件：`docs/adr/0013-tv-long-form-exoplayer-unification.md`、`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`、`docs/adr/0012-tv-series-shared-controls-for-mixed-playback-engines.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 TV 长视频 ExoPlayer 迁移先固化实施任务清单，再进入编码，不在 grill 阶段直接改播放器。实施节奏收口为：先补源码级测试约束；再抽统一 TV ExoPlayer 长视频适配层；再依次迁单片播放器和剧集播放器；随后清理长视频 LibVLC helper 与旧测试但保留 IPTV LibVLC；最后执行 TV 全量单测、`assembleDebug` 和独立代码审查。
- 影响文件：`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 TV 长视频 ExoPlayer 迁移不改手机端、不抽共享层。本次只作用于 `android-tv-app`，不复用手机端 `UnifiedPlayerScreen` 或 `DetailScreen`，也不把 TV 控制层抽到 `android-app`；未来如需跨端统一播放能力，另起共享模块设计。`CONTEXT.md` 已新增端隔离术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认迁移实施方式为一次性切换，不做 TV 长视频双轨灰度。实现完成后 TV 长视频路径只有 ExoPlayer，不加“使用旧播放器”开关，不做长视频 LibVLC 运行时回退；播放失败只提供重试、返回或诊断。若真机出现严重兼容问题，通过 git/版本回滚或后续修复处理，不在产品内长期保留双轨。`CONTEXT.md` 已新增一次性切换术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 HLS/M3U8 边界。IPTV 播放列表中的 M3U8/HLS 频道继续走 LibVLC IPTV 路径；TV 长视频播放 URL 若是普通文件/渐进流或未来出现 ExoPlayer 支持的 HLS 源，则归统一 ExoPlayer 长视频内核处理，但不复用 IPTV 的频道列表、M3U 解析、直播诊断或 LibVLC 参数。本次不新增长视频 HLS 功能，只保证迁移后播放入口不主动排斥 Media3 支持的媒体源。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 TV 长视频 ExoPlayer 迁移后的媒体鉴权方式。视频源和外挂字幕请求统一通过 ExoPlayer/Media3 HTTP data source 发送 `Authorization: Bearer <token>` header，不再向播放 URL 或字幕 URL 拼 `access_token` query；`appendAccessTokenQuery` 只保留给 LibVLC/IPTV 或历史兼容语境。现有代码中 Media3 视频源已走 header，字幕配置仍拼 query，后续实现需改成 header 路径并加测试约束。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 TV 长视频 ExoPlayer 迁移只换播放内核，不重做控制体验。现有 TV 播放控制层语义和视觉保持：返回二次确认、BACK 优先收 UI、续播、播放/暂停、seek、字幕、音轨、剧集选集轨、连播提示、错误/重试和历史上报都必须保留。实现可改造 `LongFormVideoPlayer` 或拆新 ExoPlayer 长视频组件，但用户可见遥控器行为不得因内核替换改变。`CONTEXT.md` 已补强播放控制层和内核适配术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 TV 长视频统一迁到 ExoPlayer 后接受 ASS/SSA 字幕能力降级。外挂 SRT/VTT/ASS/SSA 仍按服务端字幕列表加载进 ExoPlayer，但 ASS/SSA 只承诺按 ExoPlayer/Media3 能力展示，不承诺 libass 级样式、特效、卡拉 OK、矢量绘图或字体回退等价；不为 ASS 特效保留长视频 LibVLC 分支。`CONTEXT.md` 已将 libass 自渲染和相关 DV 字幕阶段术语标为迁移前，并新增 ExoPlayer 字幕能力边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:12 +0800
- 进度：确认 DV 安全门控在 ExoPlayer 统一迁移后继续保留。门控结果从“进入 DV 专用 Media3 分支 / 走 LibVLC / 阻断”收口为“允许进入统一 ExoPlayer 长视频内核 / 阻断”；非 DV、老数据和普通长视频走统一 ExoPlayer，DV 显示链路不支持或未知、metadata 不完整仍阻断，不提供强行播放或 LibVLC 回退。`CONTEXT.md` 已同步更新相关 DV 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行构建。

## 2026-06-16 20:04 +0800
- 进度：完成 TV 端 DV Media3 返回详情去黑场修复与收尾。确认退出确认仍保留，但确认后的动作不再进入 App 层黑色遮罩或 80ms 延迟；单片、剧集系统返回和剧集结束覆盖层“返回详情”均直接走正常返回。`CONTEXT.md` 已将旧“DV 返回详情黑场保持”标记为已撤回，并新增“DV 返回详情正常导航”约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 19:57 +0800
- 进度：完成 DV Media3 返回详情去黑场的核心改动。单片与剧集播放器的退出确认不再区分 Media3 DV 分支，二次确认后直接 `onBack()`；删除 `TvDolbyVisionExitToDetailCover` 和 80ms 延迟；剧集播放结束覆盖层的“返回详情”也改为直接返回；回归测试已反向锁定不得再保留黑色遮罩/延迟。TV 版本已递增到 `0.1.106` / `versionCode=106`，`CONTEXT.md` 已撤回旧黑场保持策略。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest`、必要的 TV 单测/编译、`git diff --check` 与乱码扫描。

## 2026-06-16 19:39 +0800
- 进度：完成 TV 电视剧详情页演员头像负 hash 越界修复与收尾。`tvCastAvatarBrush` 现在通过 `tvCastAvatarPaletteIndex()` 使用 `Math.floorMod` 计算调色板索引，新增 `TvSeriesCastAvatarPaletteTest` 覆盖负 hash、`Int.MIN_VALUE` 与空调色板边界；TV 版本已递增到 `0.1.105` / `versionCode=105`，`CONTEXT.md` 已沉淀头像调色板索引契约。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesCastAvatarPaletteTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesCastAvatarPaletteTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 19:26 +0800
- 进度：完成 TV 电视剧详情页演员头像负 hash 越界核心修复。`tvCastAvatarBrush` 改为通过 `tvCastAvatarPaletteIndex()` 使用 `Math.floorMod` 解析调色板索引，新增 `TvSeriesCastAvatarPaletteTest` 覆盖 `-4`、`Int.MIN_VALUE` 和边界种子；TV 版本递增到 `0.1.105` / `versionCode=105`，`CONTEXT.md` 沉淀头像调色板索引不得直接 `%` 取模。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesCastAvatarPaletteTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesCastAvatarPaletteTest`、`git diff --check` 与乱码扫描。

## 2026-06-16 17:29 +0800
- 进度：完成管理端绝对时间显示格式化收尾。独立子代理评审未发现阻塞问题；按评审建议补充数字字符串时间戳兼容和边界测试。确认本次只纳入管理端时间格式化、长期语义沉淀和计划记录。
- 影响文件：`admin-web/src/utils/dateTime.js`、`admin-web/src/utils/dateTime.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、相关管理端时间展示页面、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/utils/dateTime.spec.js src/views/adminDateTimeDisplay.spec.js src/views/toolboxOrphanFiles.helpers.spec.js` 通过；`cd admin-web && npm run test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中；排除测试文件后的绝对时间 raw prop / `toLocaleString` 覆盖搜索无命中。

## 2026-06-16 17:18 +0800
- 进度：完成管理端绝对时间显示格式化实现。新增共享 `formatAdminDateTime`，将用户、视频、图片、合集、演员、任务监控、IPTV、安装包管理、孤儿文件扫描、图像生成历史等可见绝对时间统一切到 `YYYY/MM/DD HH:mm:ss`；保留业务日期和时长原格式，并补充源码级回归测试防止时间戳列 raw prop 直出。
- 影响文件：`admin-web/src/utils/dateTime.js`、`admin-web/src/utils/dateTime.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、`admin-web/src/views/*` 时间展示相关文件、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/utils/dateTime.spec.js src/views/adminDateTimeDisplay.spec.js src/views/toolboxOrphanFiles.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-16 14:52 +0800
- 进度：完成系统日志空态修复收尾。确认本次只纳入系统日志接口、回归测试、长期语义沉淀和计划记录；无关工作区变更不存在。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/admin_system_logs_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run TestAdminSystemLogs -count=1` 通过；`go test ./internal/handlers -count=1` 通过；`git diff --check -- internal/handlers/admin.go internal/handlers/admin_system_logs_test.go CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-16 14:49 +0800
- 进度：完成系统日志空态修复。`/admin/system/logs` 在 `server.log` 尚未生成时返回空日志列表；下载模式返回 200 空文本文件；路径存在但无法读取时仍保留原错误分支。补充 handler 回归测试覆盖刷新和下载两种空态，并把长期约定写入 `CONTEXT.md`。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/admin_system_logs_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run TestAdminSystemLogs -count=1` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-06-16 14:29 +0800
- 进度：完成孤儿文件扫描迁到工具箱。新增 `/toolbox/orphan-files` 无 shell 工具页并复用既有 `/admin/system/orphan-files/*` API；工具箱新增“孤儿文件扫描”新标签页入口；系统设置页移除扫描状态加载、轮询和自动删除确认，仅保留临时文件清理与系统日志；命令面板搜索孤儿文件相关关键词仍只定位 `/toolbox`。
- 影响文件：`admin-web/src/views/ToolboxOrphanFiles.vue`、`admin-web/src/views/toolboxOrphanFiles.helpers.js`、`admin-web/src/views/toolboxOrphanFiles.helpers.spec.js`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxOrphanFiles.helpers.spec.js src/views/toolboxPage.spec.js src/components/base/commandPalette.helpers.spec.js src/api/admin.spec.js src/assets/themeTokens.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 14:18 +0800
- 进度：完成导航边界收口。命令面板和侧栏继续只导航到 `/toolbox`，搜索孤儿文件扫描相关关键词也只定位工具箱；具体 `/toolbox/orphan-files` 工具页必须从工具箱入口显式以新标签页打开。
- 影响文件：`CONTEXT.md`、`plan.md`、后续 `admin-web/src/components/base/commandPalette.helpers.js` 与工具箱测试。
- 验证：进入实现后补测试覆盖工具箱入口、路由注册、系统设置移除扫描 UI、命令面板别名。

## 2026-06-16 14:17 +0800
- 进度：确认孤儿文件工具页的删除确认边界。迁移后仍保留“扫描完成且存在孤儿文件时自动提示全量删除”，但触发范围仅限 `/toolbox/orphan-files` 工具页；系统设置页和其它页面不再加载扫描状态或弹危险确认。
- 影响文件：`CONTEXT.md`、`plan.md`、后续孤儿文件工具页实现与系统设置页清理。
- 验证：待继续确认工具箱入口搜索与导航边界后进入实现。

## 2026-06-16 12:12 +0800
- 进度：补充后端回归验证并完成收尾检查。`internal/repository` 与 `internal/models` 的定向测试已通过，确认电视剧详情 `cast` 实体化后仓库层和模型层仍可正常编译/运行。
- 影响文件：`plan.md`
- 验证：`go test ./internal/repository ./internal/models -count=1` 通过。

## 2026-06-16 12:10 +0800
- 进度：完成 TV 电视剧详情演员头像改造的收口与验证。演员区保持横向滚动，仅展示头像和姓名，不提供点击跳转；头像优先走服务端下发的本地可访问路由，缺失或加载失败时回落首字。顺手补齐 `TvTestSupport` 的演员 DTO 导入，并把演员区“只浏览不点击”写入 `CONTEXT.md`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 11:11 +0800
- 进度：完成电视剧刮削幂等与 TMDB 演员复用优化。已实现同一电视剧整剧 poster/backdrop 和分集 still 的本地存在跳过下载，新增 TMDB 演员按 `source + external_id` 复用已完整资料的查询；回归测试覆盖连续刮削同一剧不同分集和已有完整 TMDB 演员再次绑定视频两条路径。
- 影响文件：`internal/services/scraper.go`、`internal/repository/actor_repository.go`、`internal/services/scraper_episode_sync_test.go`、`internal/services/scraper_test.go`、`internal/queue/scrape_tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestScrapeEpisodeUploadSkipsExistingSeriesArtworkOnNextEpisode|TestSyncMovieActorsDoesNotOverrideExistingAvatarOrNotes' -count=1` 通过；`go test ./internal/services -run 'TestScrapeEpisodeUpload|TestSyncMovieActors' -count=1` 通过；`go test ./internal/repository -run TestBuildUpsertScrapedActorProfileSQL -count=1` 通过；`go test ./internal/repository -count=1` 通过；`go test ./internal/queue -run 'TestAutoScrapeAVMarksReadyOnSuccess|TestAutoScrapeAVStoresLocalizedFieldsAndMarksReady' -count=1` 通过；`git diff --check` 通过；乱码扫描无命中。`go test ./internal/services -count=1` 仍存在既有失败 `TestParseTVAPKMetadataParsesReleaseAPK`（`version_code` 期望 80、实际 99），与本次刮削优化无关。

## 2026-06-16 11:04 +0800
- 进度：完成刮削策略代码定位，准备补充重复图片下载与重复 TMDB 演员详情请求的回归测试；方案为电视剧集图片按本地文件幂等跳过，TMDB 演员按 `source + external_id` 查到完整资料后直接复用。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_episode_sync_test.go`、`internal/services/scraper_test.go`、`internal/repository/actor_repository.go`、`internal/queue/scrape_tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行定向 Go 测试、`go test ./internal/services -count=1`、`git diff --check` 与乱码检查。

## 2026-06-16 10:07 +0800
- 进度：完成管理端短视频审核页分页与播放器尺寸调整。左侧待审队列改为当前页分页列表，复用 `AdminTablePagination` 支持页码输入跳转，并显示“第 X / Y 页 · 共 N 条”；`保留并下一条` / `删除并下一条` 在页末自动进入下一页第一条，删除后刷新当前页避免 offset 分页补位视频被跳过。右侧播放器增加 9:16 手机竖屏预览框，限制在页面可用高度内，视频本身保持 contain 不裁切。`CONTEXT.md` 已更新长期术语和分页/播放器约定。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/ShortReview.vue admin-web/src/views/shortReview.helpers.js admin-web/src/views/shortReview.helpers.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-16 09:34 +0800
- 进度：完成 DV Media3 音轨偏好覆盖问题修复。已补红灯测试锁定 `onTracksChanged` 不得把 Media3 默认选中音轨回写成父层选择；实现上删除 `onSelectedAudioTrackChanged` runtime 回调，轨道加载只上报可选音轨列表，父层继续按 current selection / stored preference 决定目标并通过 Media3 override 应用。TV 版本升至 `0.1.103 / 103`，`CONTEXT.md` 补充“Media3 默认音轨不等于用户选择”约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest` 先红后绿；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 09:17 +0800
- 进度：完成 TV 端 DV 专用链路第二阶段实现。Media3 播放器现在预加载服务端外挂字幕并用 `TrackSelectionParameters` 切换字幕/音轨，不重建播放源；单片和剧集 DV 分支都接入字幕/音轨 picker、语义偏好恢复和播放控制层入口。主动返回详情时先显示 App 层黑色遮罩并短延迟导航，遮住 Media3 surface 销毁和详情页重绘之间的可控闪动；TV 版本升至 `0.1.102 / 102`。本次只纳入 TV DV 第二阶段源码、测试、版本号、`CONTEXT.md` 和 `plan.md`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`TvMedia3TrackSupport.kt`、`TvMedia3TrackPickerLayer.kt`、`TvDolbyVisionExitToDetailCover.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。曾因并发 Gradle 构建触发 Kotlin 增量缓存 EOF，已 `clean` 后串行重跑通过。

## 2026-06-15 22:57 +0800
- 进度：完成 DV 第二阶段偏好粒度收口。用户确认电视剧 DV 分集的字幕/音轨偏好继续沿用现有 `videoId` 粒度，不新增整部剧共享偏好；音轨偏好仍按语言/类型等语义匹配，避免依赖底层 track id。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做文档沉淀，暂不运行构建或单测。

## 2026-06-15 21:50 +0800
- 进度：完成 TV 电视剧 DV/非 DV 混合播放控制层收尾审查，确认本次只纳入 TV 播放控制层、Media3 seek 适配、混合切集进度语义、DV 失败态选集入口、版本号、ADR 与长期上下文沉淀。当前子代理工具要求用户显式授权委派，未新增独立 review 子代理；已改为本地抽查关键 diff 与测试覆盖后提交。
- 影响文件：`CONTEXT.md`、`docs/adr/0012-tv-series-shared-controls-for-mixed-playback-engines.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关 TV 单测、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md docs/adr android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无命中。真机仍需回归非 DV->DV、DV->非 DV、连播切换、DV 失败页“选集”和实际电视 DV/HDR 退出黑屏边界。

## 2026-06-15 21:38 +0800
- 进度：完成 TV 电视剧混合播放控制层首轮实现。新增 `TvSeriesCorePlaybackOverlay`，DV Media3 分支现在叠加电视剧核心控制层，支持播放/暂停、遥控左右 seek、进度显示、续播卡、选集轨、连播提示和结束覆盖层；字幕/音轨入口在 Media3 分支隐藏。Media3 播放器增加受控 seek 请求入口；切集前主动上报当前集历史，且上一集历史保存按上一集 route 选择 LibVLC time 或 Media3 snapshot，避免 LibVLC<->Media3 切换污染进度。DV 播放失败页新增“选集”动作，只允许切到其它分集，不提供回退 LibVLC 强行播放。TV 版本升至 `0.1.101 / 101`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesMixedPlaybackControlsSpecTest.kt`、相关源码规格测试、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。待独立 review 返回后修复问题并提交。

## 2026-06-15 19:52 +0800
- 进度：完成 DV 剧集直拷贝修复。`episode + Dolby Vision` 的转码任务现在会跳过 ffmpeg 压缩，直接把源文件迁移/复制到 `storageRoot/videos/<uuid>/source-dv.<ext>`，并把该路径同时写入 `videos.transcoded_path` 和 `playback_compat.source_playback_path`；任务完成时直接落 ready 状态，缩略图与播放兼容 metadata 仍照常生成。补了 queue 层单测，覆盖 direct copy 输出、DV 判定和“源文件已移动后重试复用稳定目标路径”的幂等恢复。
- 影响文件：`internal/queue/tasks.go`、`internal/queue/tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/queue -count=1` 通过；`go test ./internal/services -run 'Test(BuildPlaybackCompatibilityMetadata|SourcePlaybackPathFromMetadata|MergePlaybackCompatibilityMetadata|BuildTranscode|ResolveProbe|Transcode|IsSameFilePath)' -count=1` 通过；`go test ./internal/handlers ./internal/repository -run 'Test(ResolveProfiledPlayableSource|SelectRetranscodeInputPath|CollectLocalStoragePaths)' -count=1` 通过；`git diff --check` 通过；乱码扫描待执行。

## 2026-06-15 18:47 +0800
- 进度：完成短视频审核页单屏布局修复。`ShortReview.vue` 现在把页面高度锁定到 admin shell 可用视口内，工作台和左右面板使用 `minmax(0, 1fr)` 收缩；待审队列只在面板内部滚动，播放器去掉固定最小高度，随右侧面板可用高度缩放。`CONTEXT.md` 已补充短视频审核页单屏工作台约定，源码测试锁定页面不靠浏览器纵向滚动承载审核流。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`npm run test -- src/views/shortReview.helpers.spec.js` 通过；`npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/ShortReview.vue admin-web/src/views/shortReview.helpers.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-15 17:08 +0800
- 进度：完成仓库工作流约束补充。`AGENTS.md` 与 `CLAUDE.md` 已新增“`grill-with-docs` 收口后先多子代理出多方案、主代理择优拆任务再编码；代码完成后再走子代理 review 到无阻塞问题为止”的约束；`plan.md` 也已记录本次变更。
- 影响文件：`AGENTS.md`、`CLAUDE.md`、`plan.md`
- 验证：`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 17:05 +0800
- 进度：补充仓库工作流约束。复杂任务现在要求先经 `grill-with-docs` 收口，再由多个子代理并行产出多个方案、主代理择优并拆任务后编码；代码完成后还要走子代理 review，修复问题并复审到无阻塞问题为止。
- 影响文件：`AGENTS.md`、`CLAUDE.md`、`plan.md`
- 验证：待执行 `git diff --check` 与乱码扫描。

## 2026-06-15 16:44 +0800
- 进度：完成管理端侧栏收起后无法再展开的修复。折叠态现在保留可见、可点击的展开按钮，不再把唯一 toggle 隐藏；同时补了 Layout 源文回归测试和 `CONTEXT.md` 长期约定，避免后续样式回退。
- 影响文件：`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/components/Layout.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 16:22 +0800
- 进度：完成管理端短视频审核页首版落地。新增 `ShortReview.vue`，接入 `/short-review` 路由和媒体库侧栏/命令面板入口；页面固定查询 `short + ready`，支持关键词搜索、恢复上次审核位置、平板优先队列 + 播放器布局、自动播放默认静音、`保留并下一条`、`删除并下一条` 二次确认、当前页末尾自动加载后续页和末尾空状态。新增 `shortReview.helpers` 锁定位置恢复、固定查询、翻页推进和声音会话偏好；补充命令面板与路由源文测试。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/router/index.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js src/components/base/commandPalette.helpers.spec.js src/router/index.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 15:38 +0800
- 进度：通过 `$grill-with-docs` 校准管理端短视频审核页需求。已确认待审集合为 `ready` 短视频并按上传时间倒序；页面优先平板布局，采用左侧待审队列 + 右侧大播放器；重新进入优先恢复上次正在审核的视频；`保留并下一条` 只推进审核位置、不写“已审核”状态；`删除并下一条` 走现有视频删除语义且必须二次确认；走到当前已加载列表末尾时自动加载下一页，确认没有更多待审短视频时才显示末尾空状态；播放器自动播放、默认静音、声音状态仅保留当前审核会话；首版只提供关键词搜索；入口放在媒体库分组，命名“短视频审核”，路由 `/short-review`；后端复用现有 `/admin/videos`、播放链接和删除接口；审核页不承载编辑能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段仅做文档沉淀，待实现阶段执行管理端构建和必要测试。

## 2026-06-15 14:50 +0800
- 进度：完成 DV 保留实现的 review 修正。DV 原始文件保留失败现已降级为 warn log，转码任务会继续按普通输出成功完成且不写 `source_playback_path`；管理端重转码把“DV 禁止重转码”保留在错误码 `1012`，把“无可用源文件”改为 `1071`；`internal/queue/tasks.go` 已复用 `internal/services/transcode.go` 的 `IsSameFilePath`，去掉本地重复实现。本次只纳入这 4 个 Go 文件和 `plan.md`。
- 影响文件：`internal/queue/tasks.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/handlers/admin.go`、`plan.md`
- 验证：`go test ./internal/queue ./internal/handlers ./internal/services -run 'Test(IsSameFilePath|SelectRetranscodeInputPath|BuildTranscode|Transcode)' -count=1` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 14:19 +0800
- 进度：完成最新 DV 决策落地。后端转码对 `episode + Dolby Vision source` 现已保留原始文件到 `storageRoot/videos/<uuid>/source-dv.<ext>`，并把完整路径写入 `playback_compat.source_playback_path`；视频源接口支持 `profile=dv_source`；管理端重转码会拒绝持有 DV 原始播放源的视频，删除与孤儿文件扫描会精确保留/清理 `source_playback_path`。TV 端播放策略已改为：有 `source_playback_path` 的 DV 剧集分集走 `dv_source` 专用 Media3 路由，无该字段的存量数据继续阻断；TV 版本提升到 `0.1.99 / 99`。本次只纳入 DV 决策实现相关后端、TV、测试和 `plan.md`。
- 影响文件：`internal/queue/tasks.go`、`internal/services/playback_compat.go`、`internal/handlers/admin.go`、`internal/handlers/video_source.go`、`internal/repository/orphan_file_scan_repository.go`、相关 Go 测试、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关 TV 测试、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers ./internal/repository -run 'Test(BuildPlaybackCompatibilityMetadata|SourcePlaybackPathFromMetadata|MergePlaybackCompatibilityMetadata|ResolveProfiledPlayableSource|SelectRetranscodeInputPath|CollectLocalStoragePaths)' -count=1` 通过；`go test ./internal/queue -run 'Test|^$' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvLongFormDetailPresentationTest --tests com.chee.videos.core.util.UrlBuilderSourceProfileTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.TvRepositoryMappingTest` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 11:24 +0800
- 进度：完成批量编辑 review 修复。`VideoList` 的批量编辑/批量删除现在共享互斥门控，避免同一选择集并发提交；批量编辑抽屉首次打开时改为 `Promise.allSettled` 预加载合集与图片图集，并在失败时给出可恢复提示；前端测试改为锁定 Shift 区间选择和互斥/预加载结构，后端测试补了空 patch、非法图片图集 ID 和标签空输入等边界。本次只纳入 review 修复相关前端页面/测试与 `plan.md`。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.spec.js`、`internal/handlers/admin_batch_delete_test.go`、`internal/repository/admin_repository_batch_test.go`、`plan.md`
- 验证：`go test ./internal/repository ./internal/handlers -run 'Test(AdminBatch|MergeAdmin|NormalizeAdmin)' -count=1` 通过；`cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 10:46 +0800
- 进度：完成管理端视频批量编辑收尾。后端新增 `PUT /api/v1/admin/videos/batch-update`，支持对当前页已选视频统一覆盖标题、图片图集、所属合集和标签；标签支持 `replace` / `append` / `remove`，合集继续限制为短视频。前端 `VideoList` 已接入批量编辑抽屉、当前页 `Shift` 区间选择、批量 API 调用与部分成功提示。`CONTEXT.md` 已补充长期约定。本次只纳入管理端批量编辑相关前后端文件、`CONTEXT.md` 与 `plan.md`，用户既有无关改动保持不纳入。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_batch_delete_test.go`、`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_batch_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository ./internal/handlers -count=1` 通过；`cd admin-web && npm run test -- src/api/admin.spec.js src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-14 22:16 +0800
- 进度：完成任务监控页视频标题展示。`/admin/tasks` 列表查询通过 `transcoding_jobs -> videos` 左连接返回 `video_title`，避免任务因视频记录缺失被过滤；管理端任务表主列改为“任务”，主行显示视频标题，下面保留任务 ID 与视频 ID。`CONTEXT.md` 记录“任务监控视频标题”语义。本次只纳入任务监控相关后端模型/仓储、前端页面/测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/taskMonitorPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository -run TestBuildAdminListTranscodingTasksSQL -count=1` 通过；`go test ./internal/repository -count=1` 通过；`cd admin-web && npm run test -- src/views/taskMonitorPage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- ...` 通过；乱码扫描无命中。

## 2026-06-14 20:19 +0800
- 进度：完成 TV App 安装包管理页默认筛选修复。页面初始化和切客户端重置筛选时都默认查看全部发布记录，开关文案改为“只看家庭可见 / 查看全部”；上传成功后回到第一页再刷新，避免仍停在其它分页导致新草稿不易发现。`CONTEXT.md` 同步把长期约定改为“TV 管理端默认查看全部记录”。本次只纳入管理端页面、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/TvAppManage.vue`、`admin-web/src/views/tvAppManagePage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/tvAppManagePage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/TvAppManage.vue admin-web/src/views/tvAppManagePage.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-14 20:04 +0800
- 进度：完成 TV App 开屏页图片接入。用户提供图片已作为 `drawable-nodpi/tv_splash_image.png` 引入，`Theme.VideoHome.Starting` 通过窗口背景展示开屏页，`TvMainActivity.onCreate` 在 `super.onCreate` 前切回 `Theme.VideoHome`，避免开屏背景影响后续 Compose 页面；TV 端版本提升到 `0.1.98 / 98`。`CONTEXT.md` 同步记录“TV 开屏页图片”只覆盖 Compose 首帧前空白，不新增固定等待页。本次只纳入 TV splash 相关资源/代码/测试、TV 版本号、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/AndroidManifest.xml`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/main/res/values/themes.xml`、`android-tv-app/tv-app/src/main/res/drawable/tv_splash_background.xml`、`android-tv-app/tv-app/src/main/res/drawable-nodpi/tv_splash_image.png`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvSplashScreenSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.tv.TvSplashScreenSpecTest --tests com.chee.videos.feature.tv.TvApkPackagingConfigTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- ...` 通过；乱码扫描无命中。

## 2026-06-14 18:43 +0800
- 进度：完成 ED2K 链接生成器点击状态调整。未点击链接不再显示“未点击”或任何状态标签；点击链接后在本页会话内显示“已点击”，并用固定状态列降低标签出现时的文本跳动。`CONTEXT.md` 同步改回“点击后仅本页会话标记已点击”。本次只纳入 ED2K 工具页、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxEd2k.vue admin-web/src/views/toolboxPage.spec.js` 通过；乱码扫描无命中。

## 2026-06-14 18:24 +0800
- 进度：完成管理端 ED2K 链接生成器点击状态移除。页面不再维护 `ed2kClickedLinks`，链接点击后不会切换“已点击/未点击”标签或灰化样式；`CONTEXT.md` 同步收口为“不追踪或展示链接点击状态”。本次只纳入 ED2K 工具页、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxEd2k.vue admin-web/src/views/toolboxPage.spec.js` 通过；乱码扫描无命中。

## 2026-06-14 11:40 +0800
- 进度：完成第三批播放器与功能页参考图视觉替换。`LongFormVideoPlayer` 的 seek/center feedback、顶栏、底部控制条、进度条、电视剧选集 rail tooltip 和图标按钮改用暖金暗玻璃 token；字幕/音轨夜台玻璃浮层去除旧白色/冷色硬编码，改由 `AppChrome` 驱动；返回确认、连接页和配对页金色主操作同步改为深色前景。新增 `TvPlayerFunctionReferenceStyleSpecTest` 锁住第三批不回退旧冷色与白色前景。TV 端版本提升到 `0.1.96`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlayerFunctionReferenceStyleSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlayerFunctionReferenceStyleSpecTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；待执行播放器/连接/状态定向测试、TV 全量单测、assemble、`git diff --check` 与乱码检查。

## 2026-06-14 11:31 +0800
- 进度：完成 TV 参考图视觉换代第二批内容页验证，准备提交。本次只纳入 TV 首页目录、海报墙、长视频详情页、内容页参考图风格测试、TV 版本号、`CONTEXT.md` 长期约定和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvContentReferenceStyleSpecTest.kt`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvContentReferenceStyleSpecTest --tests com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest --tests com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest --tests com.chee.videos.feature.tv.TvLongFormDetailActionSpecTest --tests com.chee.videos.feature.tv.TvLongFormDetailGlassPanelSpecTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码检查无命中。

## 2026-06-14 11:29 +0800
- 进度：完成第二批内容页参考图视觉替换。TV 首页目录、海报墙和长视频详情页移除旧紫蓝/冷蓝占位色；金色按钮、左侧 rail 选中项和续播 chip 改用 `AppChrome.Canvas` 深色前景；内容页占位、scrim、标题条和详情页返回/次操作统一走暖金暗玻璃 token 或本页 reference brush。新增 `TvContentReferenceStyleSpecTest` 锁住内容页不再回退旧冷色与白字金底。TV 端版本提升到 `0.1.95`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvContentReferenceStyleSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvContentReferenceStyleSpecTest` 通过；待执行内容页定向测试、TV 全量单测、assemble、`git diff --check` 与乱码检查。

## 2026-06-14 11:18 +0800
- 进度：完成 TV 参考图视觉换代第一批收尾验证，准备提交。本次只纳入 TV 全局 token、共享 UI 组件、音轨/字幕选择器暖金化、对应测试、TV 版本号、`CONTEXT.md` 长期约定和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、相关 `core/ui` 测试
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.TvFocusSpecTest --tests com.chee.videos.core.ui.TvTypographySpecTest --tests com.chee.videos.core.ui.TvColorContrastTest --tests com.chee.videos.core.ui.TvStateFeedbackSpecTest --tests com.chee.videos.core.ui.TvIconActionSpecTest --tests com.chee.videos.core.ui.TvShapeAuditTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 与乱码检查待最终执行。

## 2026-06-14 11:08 +0800
- 进度：完成 TV 参考图视觉换代第一批落地。`AppChrome` 切到暖金、暗玻璃、8dp 通用圆角和暖金按钮内容色；`TvFocus` 的 glow 语义切到暖金且保留 spring/触觉行为；`TvTypography` 改为参考图紧凑字号角色；共享图标按钮、状态页和音轨/字幕选择器同步去除旧蓝青风格。TV 端版本提升到 `0.1.94`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.TvFocusSpecTest --tests com.chee.videos.core.ui.TvTypographySpecTest --tests com.chee.videos.core.ui.TvColorContrastTest --tests com.chee.videos.core.ui.TvStateFeedbackSpecTest --tests com.chee.videos.core.ui.TvIconActionSpecTest --tests com.chee.videos.core.ui.TvShapeAuditTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:22 +0800
- 进度：确认 TV 参考图视觉换代的实现策略：先建立暖金、暗玻璃、边框、焦点态、紧凑字号和 8dp 圆角等 token，再替换共享组件和页面调用点。代码现状显示 `core/ui/AppChrome.kt` 已是主要全局 token 入口，后续优先升级该入口而不是逐页散落硬编码。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档与代码探索，无需构建；待继续确认迁移批次后执行 `git diff --check` 与乱码检查。

## 2026-06-14 09:20 +0800
- 进度：完成 TV 电视剧详情页焦点态收尾并准备提交。本次只纳入电视剧详情页 UI、源文回归测试、TV 版本号、`CONTEXT.md` 术语和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt android-tv-app/tv-app/build.gradle.kts` 无命中。

## 2026-06-14 09:14 +0800
- 进度：电视剧详情页焦点态改造已完成到代码与文档层。右侧剧集卡片取消 `selected` 视觉高亮，只在当前焦点时显示暖金边框与高亮播放圆钮；主播放按钮未聚焦时回落为暗色 idle 背景，聚焦时才切换到金色。同步收口 TV 版本号到 `0.1.93`，并在 `CONTEXT.md` 增加“焦点与选中分离”术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 08:51 +0800
- 进度：按用户确认删除旧 DV 数据后的策略，收口 TV 端对 source DV / output 非 DV 的阻断路线。后端继续保留 DV→SDR 可信兼容输出与 metadata 标记，但 TV 端现在只关心 output 是否仍为 Dolby Vision：只要当前实际 output 非 Dolby Vision，就走普通 LibVLC 链路；output 仍为 Dolby Vision 时继续走原专用链路与显示能力门控。TV 端版本提升到 `0.1.92`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 播放兼容策略定向测试、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug`、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 08:40 +0800
- 进度：完成 DV-only 可信 SDR 兼容输出的收尾验证。后端 DV tone-map 转码、可信兼容 metadata、TV 端播放放行与版本号更新都已通过定向/全量验证；本次只纳入相关 Go、TV、`CONTEXT.md` 和 `plan.md` 文件，未处理用户已有的 `admin-web/.env.development`。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；待执行 `git diff --check`、乱码检查与最终提交。

## 2026-06-14 08:38 +0800
- 进度：完成仅针对 DV 视频的可信 SDR 兼容输出实现与 TV 端收口。后端新增 `dv_sdr_compat` 转码档位，source probe 确认为 Dolby Vision 且转码结果带显式 tone-map 标记时，输出走 H.264 VideoToolbox + SDR tone-map，兼容源路径复用同一份成品；`playback_compat` 写入可信兼容标记后，TV 端仅在该标记存在时允许 source DV / output 非 DV 走普通 LibVLC 链路。TV 端版本同步提升到 `0.1.91`。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check`、乱码检查与最终提交。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:57 +0800
- 进度：完成电视剧播放源准备态修复收尾，准备提交。本次只纳入 TV 端播放准备态代码、测试、版本号、`CONTEXT.md` 术语和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保留未提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt` 无命中。

## 2026-06-14 02:50 +0800
- 进度：完成电视剧播放源准备态代码落地。`TvSeriesPlayerViewModel` 在可播放分集异步构建播放源时置 `playbackPreparing=true`，但在源 URL 就绪前不暴露可播放 target，避免历史上报误归到新分集；兼容阻断和未绑定分集保持非准备态。`TvSeriesPlayerScreen` 对准备态显示“正在准备当前分集”，不再落到“暂不能播放”错误页。TV 端版本同步提升到 `0.1.90`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先补红灯测试后，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 已通过；待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:13 +0800
- 进度：完成两个 TV 长视频任务的现有代码收口标记。`tasks/2026-05-25-tv-long-form-focus-guarding/DONE.md` 记录原实现提交、恢复提交、当前主线基线、自动化验证与未重跑 R1~R10 人工手测脚本的边界；`tasks/2026-05-25-tv-long-form-track-preference-recovery/DONE.md` 同步记录原实现提交、恢复提交、当前主线基线、自动化验证与未重跑 A1~A7 人工手测脚本的边界。
- 影响文件：`tasks/2026-05-25-tv-long-form-focus-guarding/DONE.md`、`tasks/2026-05-25-tv-long-form-track-preference-recovery/DONE.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过；静态核对 `TvResumePromptCard(` 只出现在定义和两个 `resumePromptSlot` 调用内，`addSlave(IMedia.Slave.Type.Subtitle` 同时覆盖长视频与电视剧播放入口，`LongFormVideoPlayer.kt` 的 `onSelectAudioTrack` 同时存在 `false` 自动回灌与 `true` 用户选择路径；待执行最终 `git diff --check` 与乱码检查后提交。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 01:26 +0800
- 进度：TV 端 DV 播放诊断最小闭环完成并验证通过。诊断入口继续限定在 DV 相关阻断页和专用 Media3 播放失败页，错误卡片改为单入口替换式诊断卡片；同时把 TV 端版本号抬到 `0.1.88` / `88`，确保功能变更和版本管理同步。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon --stop` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.DolbyVisionDisplayCapabilityTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt` 无命中。保留用户已有 `admin-web/.env.development` 不纳入本次提交。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 诊断单入口单视图`。诊断不做“摘要 / 详情”切换或可展开层级，用户只通过一个显式入口进入同一张简短诊断卡片，减少交互复杂度。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛是否需要任何额外 debug-only 条件后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 播放诊断入口` 边界。首轮只在播放阻断页和专用 Media3/ExoPlayer 播放失败页通过显式入口查看；正常播放页不常驻展示；诊断信息不泄漏 access token 或完整播放 URL，也不提供强行播放或切换链路能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛诊断字段与实现范围后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 18:38 +0800
- 进度：完成图库资产参考图格式过滤收尾检查，准备提交。确认只纳入本轮后端过滤、工作台请求参数、API 单测和计划记录；用户已有 `admin-web/.env.development` 仍保持未纳入。
- 影响文件：`internal/models/admin.go`、`internal/handlers/admin_image.go`、`internal/repository/image_repository.go`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：`rg -n $'\uFFFD' CONTEXT.md plan.md internal/models/admin.go internal/handlers/admin_image.go internal/repository/image_repository.go admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/api/admin.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md internal/models/admin.go internal/handlers/admin_image.go internal/repository/image_repository.go admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/api/admin.spec.js` 通过。

## 2026-06-13 18:35 +0800
- 进度：完成图库资产参考图格式过滤代码落地。后端 `/admin/images` 支持逗号分隔的 `stored_mime` 查询过滤，列表分页和总数直接按处理图 MIME 收口；图像生成工作台图库入口固定请求 `image/png,image/jpeg,image/webp`，与 `CONTEXT.md` 中 `admin 图库资产参考图可选范围` 保持一致。
- 影响文件：`internal/models/admin.go`、`internal/handlers/admin_image.go`、`internal/repository/image_repository.go`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：`go test ./internal/...` 通过；`cd admin-web && npm run test -- admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有既有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:11 +0800
- 进度：完成历史恢复参考图大小校验补强。新增 data URL 字节估算，历史快照恢复出的本地冻结图会带回 size；继续追加上传/粘贴/图库参考图时，校验会同时统计已有恢复图和新图，避免 `File` 为空导致总大小上限失效。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:08 +0800
- 进度：完成图库资产来源跳转闭环。工作台当前参考图和历史输入快照里的 `library_asset` 来源现在提供“查看来源”，跳转 `/images?image_id=<source_image_id>`；图片管理页消费 `image_id` query 后尝试打开详情，并在成功或失败后清理 query。来源读取失败时工作台仍保留冻结标题与来源图片 ID 摘要。`CONTEXT.md` 已新增 `admin 图库资产来源跳转 query 契约`。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/ImageManage.vue`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:04 +0800
- 进度：完成历史输入快照展示与继续改稿恢复闭环。选中本地历史时会读取 `referenceSnapshots` 对应的本地 IndexedDB 冻结图片，展示参考图来源类型、图库来源标题与来源图片 ID；“继续改稿”恢复提示词、参数和本地冻结参考图，不重新读取图库资产。若历史参考图本地文件缺失，会显示缺失位并阻止直接恢复，避免把失效输入送往生成接口。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js` 通过。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 17:57 +0800
- 进度：完成本轮收尾检查，准备提交工作台图库资产参考图功能。确认生成 payload 不泄漏 `source_image_id` 等本地来源字段，图库抽屉 object URL 在关闭和请求失效时释放，重复图库资产按来源图片 ID 跳过；用户已有 `admin-web/.env.development` 保留未提交，不纳入本次提交。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js admin-web/src/api/admin.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js admin-web/src/api/admin.spec.js` 通过。

## 2026-06-13 17:55 +0800
- 进度：完成工作台内置图库资产参考图首轮实现。已新增本地参考图快照 helper，`library_asset` 来源元数据进入 IndexedDB 任务快照但不进入生成 payload；工作台输入区新增“从图库选择”抽屉，按 `status=ready`、`active=1`、搜索词和图片合集筛选图库资产，支持网格多选和剩余槽位限制；加入时通过 `/admin/images/:id/view` 拉取 blob，转为本地 data URL 后立即冻结到参考图槽位，失败项按单张提示且不回滚成功项。`CONTEXT.md` 已新增 `admin 图库资产冻结读取契约`。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。

## 2026-06-13 17:44 +0800
- 进度：补齐图库资产参考图的可选范围、标题快照与重复选择规则。已确认只允许 `status=ready` 且 `active=true` 的图片资产作为新参考图；来源标题按加入工作台时冻结，不随来源资产重命名改写；同一来源图片资产 ID 已在当前草稿里时再次选择应跳过并提示。`CONTEXT.md` 已新增 `admin 图库资产参考图可选范围`、`admin 图库资产来源标题冻结` 与 `admin 图库资产参考图重复选择跳过`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:19 +0800
- 进度：DV 专用 Media3/ExoPlayer 播放链路本轮实现完成。单个长视频和电视剧分集现在统一通过播放链路选择器在 LibVLC、Media3 DV 专用链路和阻断之间决策；output 仍为 DV 且显示能力支持、播放 URL 有效时进入专用 Media3 分支，source DV 但 output 非 DV、显示能力未知/不支持、metadata 不完整仍阻断。剧集分集按当前分集独立构建 URL 与重判，Media3 失败重试会重新评估并重建 ExoPlayer，不回退 LibVLC；观看历史继续沿用现有接口。已更新 TV 端版本到 `0.1.87`，并调整 LibVLC 守卫测试为允许 DV 专用 Media3 例外。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:compileDebugKotlin` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。待执行乱码检查与 `git diff --check` 后提交；`admin-web/.env.development` 和生图参考图相关记录不纳入本次 DV 提交。

## 2026-06-13 15:57 +0800
- 进度：按用户要求为当前任务创建独立分支 `admin-image-library-reference`。继续收敛图库资产参考图的生命周期：已确认图库资产进入生图工作台后要复制成当前创作自己的本地输入快照，并保留来源图片资产 ID、当前标题等轻量来源信息；后续来源资产重命名或删除不应破坏本地历史查看和继续改稿，只影响来源跳转与摘要状态。`CONTEXT.md` 已新增 `admin 图库资产参考图输入快照`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留既有未提交改动 `admin-web/.env.development`、TV 端策略文件和 TV 测试文件不纳入本轮处理。

## 2026-06-13 14:32 +0800
- 进度：DV 本地能力模块本轮已完成并验证通过。最终纳入变更的文件仅包含 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapability.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapabilityTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；用户已有 `admin-web/.env.development` 未纳入本次提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapability.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapabilityTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.DolbyVisionDisplayCapabilityTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapability.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapabilityTest.kt android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapability.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapabilityTest.kt android-tv-app/tv-app/build.gradle.kts` 通过。

## 2026-06-13 14:26 +0800
- 进度：DV 本地能力模块已进入实现阶段。新增 `DolbyVisionDisplayCapability` 纯本地模块和对应单测，封装当前默认 Display 的 HDR capability 读取，并把 `supportedHdrTypes` 映射为支持 / 不支持 / 未知三态、闭集原因码和归一化 HDR 类型名摘要；首轮仍未接入播放决策、UI、服务端写入或本地持久化。TV 端版本号递增至 `0.1.86`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapability.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/DolbyVisionDisplayCapabilityTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.DolbyVisionDisplayCapabilityTest` 曾因 `DisplayHdrCapabilityReader` 未实现失败；实现后同一命令通过。待继续执行 TV 全量单测、assembleDebug、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:24 +0800
- 进度：DV 本地能力模块 grill 收敛完成，准备转入实现。已确认即使首轮不外显，也属于 TV App 功能能力新增，必须按仓库规则递增 TV 端 `versionCode/versionName`。实现范围：TV feature 包内新增可注入显示能力 reader、三态结果对象、闭集原因码、HDR 类型名摘要归一化和纯单测；不接 UI、不写服务端、不持久化日志、不参与播放放行。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`CONTEXT.md`、`plan.md`
- 验证：已执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 无命中、`git diff --check -- CONTEXT.md plan.md` 通过；实现后待跑 TV 定向单测、TV 全量 unit test、assembleDebug、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:24 +0800
- 进度：TV Dolby Vision 播放兼容策略实现完成并进入提交准备。已确认本轮只落地现有 `playback_compat` 决策、老数据普通放行、信息不完整阻断提示和重试入口；未实现设备/HDR 能力探测、专用系统播放器/Media3 分支或播放页补探测。提交范围将精确排除用户已有 `admin-web/.env.development`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvStateFeedbackUsageTest.kt`、`docs/adr/0010-dolby-vision-tv-playback-compatibility.md`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-06-13 11:14 +0800
- 进度：grill-with-docs 已确认这套 TV Dolby Vision 播放兼容策略需要 ADR 沉淀。已新增 `docs/adr/0010-dolby-vision-tv-playback-compatibility.md`，记录“不异色优先”、老数据普通放行例外、专用链路资格、播放页不补探测与不提供强行播放旁路等关键取舍。
- 影响文件：`docs/adr/0010-dolby-vision-tv-playback-compatibility.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md docs/adr/0010-dolby-vision-tv-playback-compatibility.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:11 +0800
- 进度：grill-with-docs 继续收敛 `播放兼容信息不完整` 失败态里的“重试”语义。已确认该“重试”只重新拉取详情 metadata 并重新执行本地播放决策，不触发服务端补探测、实时 ffprobe、重新转码或 metadata 修复。`CONTEXT.md` 已追加 `播放兼容信息不完整重试`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:08 +0800
- 进度：grill-with-docs 继续收敛 `播放兼容信息不完整` 的播放入口处理。已确认已存在的 `playback_compat` metadata 不能支撑播放安全判断时，TV 端首轮只提示“播放兼容信息不完整，暂不能确认安全播放”这类失败态，并提供返回/重试；播放页内不触发补探测、实时 ffprobe 或后台修复任务，补探测后续应放到后端任务或管理端修复入口。`CONTEXT.md` 已追加 `播放兼容信息不完整提示`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:03 +0800
- 进度：grill-with-docs 修正 `playback_compat` metadata 不完整时的处理边界。用户明确要求“老数据直接放行”，已收敛为：`playback_compat` 缺失或过旧的存量视频可以继续按既有 `LibVLC` 普通播放链路放行，但这不是 Dolby Vision 安全播放，也不能进入专用系统播放链路；一旦 metadata 已明确标记 DV 风险源，仍按 DV 阻断或专用链路规则处理。`CONTEXT.md` 已将 `DV metadata 不完整不推断放行` 修正为 `DV metadata 不完整不进专用链路`，并新增 `老数据普通播放放行`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:45 +0800
- 进度：grill-with-docs 继续收敛 `明确不支持` 场景下的重试边界。已确认当 DV 风险源被明确判定为当前设备/链路不支持安全播放时，“重试”仍只适用于本地显示模式、外接链路或相关系统状态发生变化后的幂等重判，不承诺修复。`CONTEXT.md` 已扩展 `DV 重试依赖本地状态变化`，让它同时覆盖能力未知和明确不支持两种阻断原因。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:33 +0800
- 进度：grill-with-docs 继续收敛 `重试` 的语义。已确认 Dolby Vision 风险阻断页里的“重试”首轮只表示重新执行一次本地能力判断并重走同一套播放决策，不切换播放器分支、不改写服务端 metadata，也不触发服务端修复；`CONTEXT.md` 已追加 `DV 重试属于幂等重判` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:24 +0800
- 进度：grill-with-docs 继续收敛当前 DV 主链路的“可行性”定义。已确认本轮只把现状收口为 `DV 风险探测与阻断可行`：后端 `playback_compat` 探测与落库、TV 端读取 metadata 并保守阻断已打通；不把现状扩写成“DV 安全放行可行”，因为设备 HDR 能力探测和长视频专用系统播放分支仍未落地。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-12 20:46 +0800
- 进度：手机端与 TV 端安装包分轨已落地。后端在保留 `tv_app_releases` / `tv_app_release_apks` 表名的前提下新增 `client_type`、手机端单包固定槽位 `single`、轨内 `versionCode` 唯一且 `versionName` 冲突拒绝；家庭页新增 `/downloads/android-tv` 与 `/downloads/android-phone`，并保留 `/tv-app` 兼容入口；Admin Web 已提升为同一页内切换手机端 / TV 端的统一安装包管理工具。
- 影响文件：`migrations/0024_app_apk_distribution_client_type.*`、`internal/models/tv_apk.go`、`internal/models/admin.go`、`internal/services/tv_apk.go`、`internal/services/tv_apk_manager.go`、`internal/repository/tv_apk_repository.go`、`internal/repository/migrations_test.go`、`internal/handlers/router.go`、`internal/handlers/tv_apk.go`、`internal/handlers/tv_app_family_page.go`、`internal/handlers/tv_app_family_page_test.go`、`internal/handlers/iptv.go`、`internal/handlers/iptv_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/TvAppManage.vue`、`admin-web/src/components/base/commandPalette.helpers.js`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/services ./internal/repository ./internal/handlers -count=1` 通过；`cd admin-web && npm run test -- --run src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过。提交时继续排除用户已有 `admin-web/.env.development`，并保留尚未提交的文档 ADR 工作区改动按本任务一并纳入。

## 2026-06-12 20:07 +0800
- 进度：手机端 APK 分发管理的上位设计已收敛完成，并补充 ADR。当前已确认：手机端与 TV 端共用一套安装包分发模型并按 `client_type` 分轨；手机端首期是单包发布记录、家庭轻量页独立、管理端同页切客户端类型、`client_type` 与 `packageName` 强绑定、轨内 `versionCode` 唯一、手机端已发布替换前也必须先下线。已新增 ADR `0009-unified-app-apk-distribution-by-client-type.md` 固化这些不可逆取舍。下一步可以直接转入实现。
- 影响文件：`CONTEXT.md`、`docs/adr/0009-unified-app-apk-distribution-by-client-type.md`、`plan.md`
- 验证：待执行文档乱码检查与 `git diff --check`。

## 2026-06-12 18:58 +0800
- 进度：家庭轻量页与定向测试已落地。`internal/handlers` 定向测试通过；组合 Go 测试首次在沙箱内失败，原因是仓库现有 `admin_image_generation_test.go` 通过 `httptest.NewServer` 绑定本地端口而被当前受限环境拒绝，已按原命令申请提权重跑，不视为本次改动回归。
- 影响文件：`internal/handlers/tv_app_family_page.go`、`internal/handlers/tv_app_family_page_test.go`、`internal/handlers/router.go`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/handlers -run 'TestRegisterIncludesTVAppFamilyPageRoute|TestMountTVAppFamilyPage' -count=1` 通过；组合测试提权执行中。

## 2026-06-12 19:05 +0800
- 进度：家庭轻量页交付完成。Go 侧新增 `/tv-app` 独立页面，带最小登录、refresh token 续期、最近三版展示与 ABI 直下；下载前会先检查 access token 是否临近过期，必要时自动刷新，避免页面停留过久后直接 401。已同步补 family 页路由/页面测试与长期沉淀。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/tv_app_family_page.go`、`internal/handlers/tv_app_family_page_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/handlers -run 'TestRegisterIncludesTVAppFamilyPageRoute|TestMountTVAppFamilyPage' -count=1` 通过；`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/handlers ./internal/services ./internal/repository -count=1` 通过（提权，因现有测试需本地绑定 `httptest` 端口）；`rg -n $'\uFFFD' plan.md CONTEXT.md internal/handlers/tv_app_family_page.go internal/handlers/tv_app_family_page_test.go internal/handlers/router.go` 无命中。提交时仅纳入本任务文件，保留用户已有 `admin-web/.env.development` 未提交。

## 2026-06-12 18:39 +0800
- 进度：已打通 TV APK 二进制 manifest 解析与后端主链路：新增 APK 元数据解析单测，补齐 TV 安装包 repository / service / Gin API，管理端已接入 TV 安装包页面、路由和 API 封装。当前进入收尾验证与记录阶段。
- 影响文件：`internal/services/tv_apk.go`、`internal/services/tv_apk_test.go`、`internal/services/tv_apk_manager.go`、`internal/repository/tv_apk_repository.go`、`internal/handlers/tv_apk.go`、`internal/handlers/router.go`、`internal/handlers/iptv.go`、`internal/handlers/iptv_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/views/TvAppManage.vue`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/... -run 'TestParseTVAPKMetadata|TestTVReleaseHelpers|TestRegisterIncludesIPTVRoutes' -count=1` 通过；`cd admin-web && npm run test -- --run src/api/admin.spec.js src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过。

## 2026-06-12 16:43 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 新版本显式发布后才进入家庭视线`：新 `versionCode` 的首个 Release APK 上传完成并自动建档后，发布记录默认先只在管理端可见，只有管理员显式执行“发布”后才进入家庭轻量下载页。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“最近三版”和“推荐版本”的计算基准，是否只看家庭当前可见的已发布记录。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 09:07 +0800
- 进度：完成未接入的旧任务化生图残留清理。当前工作区仅保留 `plan.md` 账本更新与用户已有的 `admin-web/.env.development` 本地环境改动，已避免未跟踪 `0023` migration 被后续流程误执行。
- 影响文件：`plan.md`
- 验证：`git status --short` 确认仅剩 `plan.md` 与未纳入的 `admin-web/.env.development`

## 2026-06-12 00:39 +0800
- 进度：完成 admin 图像工作台“参考图局部蒙版编辑”主链路。前端新增本地蒙版编辑弹窗、蒙版目标预览、前序结果复用来源标记与最小本地快照字段；后端 `/admin/image-generation/generate` 新增 `mask` 入参、目标槽位校验、尺寸校验以及执行层首图重排后再向上游 `/images/edits` 提交蒙版。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/ImageWorkbenchMaskEditor.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test -- src/views/imageWorkbench.helpers.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过；`go test ./internal/handlers -count=1` 通过。未纳入 `admin-web/.env.development` 与未跟踪的旧任务化残留文件。

## 2026-06-12 00:00 +0800
- 进度：grill-with-docs 继续收敛本地历史待导入项的默认聚焦顺序，确认当同时存在多条“待导入的成功项”时，默认仍按创建时间倒序聚焦最新生成的一条；`CONTEXT.md` 已补充该排序语义，和现有本地历史倒序读取口径保持一致。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:16 +0800
- 进度：grill-with-docs 继续收敛本地历史删除边界，确认删除 `admin 图像创作历史` 只影响当前浏览器里的本地历史与创作上下文，不影响已经通过“导入图库”进入正式生命周期的图片资产；`CONTEXT.md` 已补充该边界，避免后续把本地清理误做成反向删库。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:08 +0800
- 进度：grill-with-docs 继续收敛本地历史可见范围，确认 `admin 图像创作历史` 只保证当前浏览器本地留存，不做跨浏览器、跨设备或跨管理员同步；`CONTEXT.md` 已补充该边界，后续共享与长期保留仍只通过显式导入图库闭环。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:25 +0800
- 进度：grill-with-docs 继续收敛参考图来源类型枚举，确认采用 `admin 首轮参考图来源类型收口`：首轮 `request_payload.reference_images[]` 的 `source_kind` 只保留 `browser_input` 与 `previous_result` 两种；上传/粘贴统一折叠为 `browser_input`，也不提前为图库资产引用预留第三种来源类型。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:16 +0800
- 进度：grill-with-docs 继续收敛参考图快照结构，确认采用 `admin 参考图快照显式来源类型`：`request_payload.reference_images[]` 的每个槽位都显式保存 `source_kind`，不再通过“是否存在来源任务/结果字段”去推断来源类型，保证详情、重试预填和审计使用统一结构。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:35 +0800
- 进度：grill-with-docs 收口了“首轮参考图收口”与“结果图继续改可作蒙版目标”之间的冲突，确认采用 `admin 前序生图结果输入来源`：首轮任务化仍不开放通用服务器图片资产引用，但允许前序生图任务结果作为继续改和局部蒙版的输入来源，并同步把 `admin 任务化首轮参考图收口` 改写为“临时参考图 + 前序生图结果”两类允许来源。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:34 +0800
- 进度：grill-with-docs 继续收敛“结果图继续改”和局部蒙版的结合边界，确认采用 `admin 结果图继续改可作蒙版目标`：管理员通过“以结果图继续改”进入工作台后，被绑定的那张结果图既可作为普通参考图，也可继续作为局部蒙版目标图。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:27 +0800
- 进度：grill-with-docs 继续收敛局部蒙版与任务化的结合边界，确认采用 `admin 局部蒙版进入任务快照`：局部蒙版不是浏览器里的纯临时态，提交到 `admin 图像生成任务化` 时要随同目标参考图关系一起冻结进任务输入；同时把 `admin 任务请求快照不存 base64` 收紧为“参考图与蒙版都只存本地文件引用，不存原始 base64”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:24 +0800
- 进度：grill-with-docs 纠正了既有术语和最新范围决定的冲突，确认 `admin 参考图编辑` 不再只表示整图编辑，而是同时覆盖整图编辑与 `admin 局部蒙版编辑`；后者特指围绕单张目标参考图的局部涂抹生成语义，属于本轮参考迁移范围。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:21 +0800
- 进度：grill-with-docs 继续收敛 `references/gpt_image_playground` 的迁移范围，确认采用 `admin 生图参考迁移范围`：reference 只作为当前生图主链路的借鉴来源，只迁与 `admin 图像生成工作台`、`admin 图像生成任务化` 和项目里已完成的 `admin 图像导入媒体库` 直接相关的交互与能力，不按原项目整包搬入 `AgentWorkspace`、`FavoriteCollections`、人物/资产侧旁支等未立项模块。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:20 +0800
- 进度：grill-with-docs 继续收敛管理端生图“本地留存 vs 显式导入”边界，确认采用 `admin 生图结果本地留存`：生成结果无论来自浏览器本地历史还是服务端任务历史，都默认只留在本地创作上下文，不因生成完成或任务化持久化而自动入库；只有项目里已完成的显式 `admin 图像导入媒体库` 功能才把结果转成正式图片资产。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:09 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的僵尸任务阈值语义，确认 `admin 图像生成僵尸任务自动收口` 在首轮应表现为：`queued` 用固定阈值收口，`running` 用每条任务自己的 `execution_config.timeout_seconds + grace` 收口，不照搬转码任务的固定“多久没更新”规则。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:04 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的结果计数边界，确认采用 `admin 结果数作为任务稳定事实`：`result_count` 作为任务表上的稳定统计字段维护，供列表、详情和重复投递幂等收口直接使用，而不是每次现场聚合结果表。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:02 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的结果查看路由边界，确认采用 `admin 生图结果查看走任务嵌套路由`：结果查看接口挂在 `/admin/image-generation/tasks/:task_id/results/:result_id/view` 之下，不把 `result_id` 单独暴露成顶级路由。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:00 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的失败摘要落库边界，确认 `admin 服务端失败摘要回显` 在实现上应表现为任务表里的稳定失败摘要字段，供列表和详情直接读取，而不是在查询阶段临时拼装。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:57 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的列表分页边界，确认 `admin 图像生成任务列表分区` 必须由后端查询层直接完成排序与分页，不能交给前端对当前页结果做二次重排。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:55 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的列表返回边界，确认 `admin 图像生成任务行摘要` 在接口上表现为：`GET /admin/image-generation/tasks` 直接返回首张结果轻量摘要、结果数/目标数与失败摘要，不依赖前端逐条补查详情。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:52 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的可见范围边界，确认采用 `admin 非拥有者访问按不存在处理`：跨管理员访问不属于自己的任务详情或结果子资源时按 `404` 处理，不用 `403` 暴露任务存在性。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:49 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的提交返回边界，确认 `admin 提交后进入新任务详情` 在接口上表现为：`POST /admin/image-generation/tasks` 直接返回新任务的初始详情快照，而不是只回 `task_id`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:46 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的结果表边界，确认采用 `admin 首轮结果表不含导入关联`：首轮结果表只保存生成产物元数据，不提前落当前绑定资产、导入计数或导入审计字段；导入闭环后续再独立扩展。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:44 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的参数版本落库边界，确认 `admin 图像生成参数快照版本` 需要直接体现在任务表结构里，首轮迁移应预留独立 `request_schema_version` 字段，而不是只把版本号塞进 `request_payload`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:42 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的重试关系落库边界，确认 `admin 图像生成任务重试` 的“来源任务关联”需要直接体现在任务表结构里，首轮迁移应预留可空 `retry_from_task_id`，即使本轮尚未开放显式重试入口。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:38 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的首轮范围，确认采用 `admin 首轮取消入口暂缓`：`canceled` 状态先保留在模型里，但本轮最小闭环不实现管理员显式取消入口，只覆盖提交、列表、详情、结果查看与 worker 执行收口。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:35 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的提交去重边界，确认采用 `admin 首轮幂等键可选`：首轮后端支持可选显式 `idempotency_key`，未提供时每次提交都视为新任务，不根据 payload 做隐式猜测去重。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:33 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的未配置上游失败语义，确认采用 `admin 未配置上游也落失败任务`：任务化提交一旦分配 `task_id`，后续即使因为上游未配置而无法继续，也要保留为一条 `failed` 任务，而不是像同步旧接口那样只做即时报错。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:30 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的 `execution_config` 边界，确认 `admin 图像生成执行配置快照` 在首轮需要记录模型、超时和可判定失效的稳定端点标识，但不把原始上游 URL 直接暴露到任务历史。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:24 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的提交顺序边界，确认采用 `admin 生图任务先建记录后补快照`：分配 `task_id` 后先建任务记录，再做参考图落盘、`request_payload` 补全和入队；任一后续环节失败都统一回落到同一条任务并标记为 `failed`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:20 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的调度公平性边界，确认 `admin 图像生成任务调度公平性` 在首轮仅按现有 Asynq 单队列模型提供最佳努力公平，不引入按管理员分桶或自定义调度器，因此不承诺严格的单管理员 FIFO。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:19 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的重复投递幂等语义，确认采用 `admin 生图任务重复投递幂等收口`：同一 `task_id` 被重复投递时先看已持久化结果数做收口，只有 `result_count == 0` 且任务仍非终态时才允许真正重跑上游。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:17 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的自动重试边界，确认采用 `admin 部分结果后停止自动重试`：同一 `task_id` 一旦已经持久化出至少一张结果图，后续再出错就直接定格为 `failed` 并保留已有结果，不再继续自动 retry 试图补齐。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:13 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的部分成功持久化语义，确认 `admin 图像生成部分可用结果` 在实现上应表现为“单张尽力落盘并保留已成功项，不因同批某张失败而整批回滚”；最终若成功张数少于请求张数，任务仍记为 `failed`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 18:12 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的参考图详情回显边界，确认采用 `admin 参考图详情不暴露存储路径`：数据库内部继续保存 `stored_path`，但任务详情对前端只回显槽位、可用性、查看地址和缺失原因等面向界面的字段。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 16:27 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的请求快照边界，确认采用 `admin 任务请求快照不存 base64`：`request_payload` 只保存归一化逻辑输入和参考图落盘路径，不把原始 `data_url` base64 直接落库。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 16:10 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的结果预览边界，确认采用 `admin 首轮结果查看只返原图`：首轮结果查看接口不支持 `w/h/fit/q` 变体参数，也不把任务结果接入 `image_variants`；列表和详情先直接消费原图 URL。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 16:06 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的重试归属语义，确认采用 `admin 自动重试不派生新任务`：执行器自动 retry 仍记在原 `task_id` 上，只有管理员显式点击“重试”时才派生新的任务记录。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 16:01 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的结果读取边界，确认采用 `admin 生图结果复用安全图片打开`：图像生成结果查看接口虽然走独立路由，但底层仍复用 `openLocalImageFile` / `serveOpenedLocalImage`，保持与现有图片预览一致的限时打开和 404/503 语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:57 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的接口命名空间，确认采用 `admin 图像生成任务命名空间独立`：图像生成任务不复用现有 `/admin/tasks` 转码任务监控入口，而是走独立 `/admin/image-generation/tasks` 命名空间承接提交、列表、详情与结果查看。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:52 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的失败标识语义，确认采用 `admin 任务 ID 作为失败锚点`：任务化首轮不单独持久化 `request_id`，界面、日志关联和失败回显统一围绕 `task_id` 闭环；同步旧接口的临时 request ID 继续留在旧链路里。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:49 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的运行中结果语义，确认首轮按“结果落盘后立即可见”解释 `admin 运行中结果即时回显`：非流式 OpenAI-compatible 上游通常不会在 `running` 中途逐步产图，本轮不为此额外引入流式协议或分段落盘。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:46 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化失败落账边界，确认采用 `admin 提交失败保留任务记录`：管理员已明确提交且服务端已进入创建任务语义后，即使最终入队失败，也要保留为一条 `failed` 任务，而不是把这次尝试当作未发生。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:37 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化结果回传边界，确认采用 `admin 任务结果查看接口分离`：任务列表与详情接口只返回结果元数据和查看地址，不内联 base64 整图；图片内容通过独立结果查看接口读取，以支撑运行中结果即时回显并控制详情接口体量。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:36 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化输入边界，确认采用 `admin 任务化首轮参考图收口`：首轮任务化提交仅接当前工作台已稳定使用的浏览器 `data_url` 参考图，不在本轮同时引入服务器图片资产引用；后者保留为后续扩展能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 15:26 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化边界，确认采用 `admin 图像生成双轨接口过渡`：保留现有同步 `/admin/image-generation/generate` 供当前工作台继续使用，新增任务化提交/列表/详情接口承接服务端历史与队列闭环，本轮不强制让既有工作台同步切到任务详情流。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 13:56 +0800
- 进度：确认管理端图像生成下一阶段的运行中结果即时回显语义，并将其收敛为 `admin 运行中结果即时回显`。规则是：任务仍处于 `running` 时，只要已有一张或多张可用结果图，详情就应立刻展示这些结果，而不是等到 `succeeded` 或 `failed` 终态后再统一出现。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-11 13:39 +0800
- 进度：确认管理端图像生成下一阶段的结果图继续改后输入变更审计语义，并将其收敛为 `admin 结果图继续改后输入变更审计`。规则是：管理员从“以结果图继续改”进入工作台后，如果又显式修改提示词、尺寸、质量或张数等输入参数，历史里应记为“派生后已修改输入”的新任务，而不是仍被视为原样沿用原结果图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-11 11:53 +0800
- 进度：确认管理端图像生成下一阶段的结果图继续改后参考图变更审计语义，并将其收敛为 `admin 结果图继续改后参考图变更审计`。规则是：管理员从“以结果图继续改”进入工作台后，如果又显式替换、移除、补充或重排这张派生来的参考图，历史里应记为“派生后已修改参考图集合”的新任务，而不是仍被视为原样继承原结果图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-11 11:49 +0800
- 进度：确认管理端图像生成下一阶段的已删除结果图继续改失效语义，并将其收敛为 `admin 已删除结果图继续改失效回显`。规则是：被用于“以结果图继续改”的结果图若后来被删除、清理或失效，历史里仍保留这次派生关系和轻量快照，但不能再直接复用这张原图继续改，管理员必须重新选择可用输入。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-11 11:45 +0800
- 进度：确认管理端图像生成下一阶段的结果图继续改单张绑定语义，并将其收敛为 `admin 结果图继续改单张绑定`。规则是：一条任务有多张结果时，点击某张结果上的“以结果图继续改”只绑定该单张结果，不自动把同批其他结果带入下一稿。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-11 10:34 +0800
- 进度：确认管理端图像生成下一阶段的继续改稿与结果图继续改分流语义，并将其收敛为 `admin 继续改稿与结果图继续改分流`。规则是：历史任务详情里的“继续改稿”只派生该任务冻结输入，不自动把本次产出的结果图带入参考图；若管理员想基于某张结果图继续编辑，应走该单张结果上的显式“以结果图继续改”入口。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 23:59 +0800
- 进度：确认管理端图像生成下一阶段的历史任务继续改稿语义，并将其收敛为 `admin 历史任务继续改稿派生新草稿`。规则是：从历史任务详情被动返回工作台时继续显示当前草稿；只有显式点击“继续改稿”时，系统才基于该任务的冻结输入派生一份新草稿，而不是直接覆盖当前草稿。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 23:49 +0800
- 进度：确认管理端图像生成下一阶段的提交后草稿语义，并将其收敛为 `admin 提交后输入快照视为新草稿`。规则是：提交后工作台里保留下来的那份输入快照默认应视为下一次潜在提交的新草稿，而不是继续绑定刚刚已经进入 `queued` 或 `running` 的那条任务；这样已提交任务的输入冻结边界和工作台继续编辑的语义不会混淆。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 23:46 +0800
- 进度：确认管理端图像生成下一阶段的提交后主流程语义，并将其收敛为 `admin 提交后进入新任务详情`。规则是：管理员点击“生成”后，界面应立即切到新建任务详情并显示其已进入 `queued`；同时工作台仍保留一份可继续编辑的输入快照，方便基于刚才的输入继续发起下一次生成，而不是被迫回填全部表单。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 23:40 +0800
- 进度：确认管理端图像生成下一阶段的改绑副作用提示语义，并将其收敛为 `admin 改绑不删除旧资产`。规则是：改绑当前关联的确认界面还应明确提示“旧资产不会被删除，只是失去当前关联身份”，避免管理员把改绑误解成系统会自动清理旧资产。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 23:04 +0800
- 进度：确认管理端图像生成下一阶段的改绑确认信息展示语义，并将其收敛为 `admin 改绑确认需展示前后目标`。规则是：改绑当前关联的确认界面应直接展示当前关联资产，以及执行后将切换到的新资产信息，例如标题预填、归档去向或等价目标摘要，而不是只给一句泛泛确认；这样管理员能准确理解“当前会被什么替换成什么”。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 22:44 +0800
- 进度：确认管理端图像生成下一阶段的改绑确认语义，并将其收敛为 `admin 改绑当前关联需显式确认`。规则是：如果管理员选择“重新导入并改绑当前关联”，界面应在真正执行前要求一次显式确认，因为这会改写当前有效资产关联；这类高风险动作不应像普通导入一样被静默触发。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 22:06 +0800
- 进度：确认管理端图像生成下一阶段的额外导入审计标注语义，并将其收敛为 `admin 额外导入审计标记非当前关联`。规则是：如果管理员选择“额外再导入一份新资产”，这次新资产的创建记录仍应进入审计上下文，并明确标注它不是当前关联；这样主视图保持简洁，历史派生资产关系仍然可追溯。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 22:02 +0800
- 进度：确认管理端图像生成下一阶段的额外导入关联稳定语义，并将其收敛为 `admin 额外导入不改当前关联`。规则是：如果管理员选择“额外再导入一份新资产”，任务详情主视图里的当前有效资产应继续指向原来的那份，不因为新资产创建成功而自动切换；只有明确选择“重新导入并改绑当前关联”时，当前关联才应被改写。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 18:09 +0800
- 进度：确认管理端图像生成下一阶段的额外导入收口语义，并将其收敛为 `admin 额外导入沿用默认归档规则`。规则是：如果管理员选择“额外再导入一份新资产”，新资产仍应默认进入同一个 `AI 生图导入` 合集，并沿用默认标题预填规则，而不是另起一套特殊归档或命名逻辑。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 18:04 +0800
- 进度：确认管理端图像生成下一阶段的再导入语义分流，并将其收敛为 `admin 再导入动作分流`。规则是：如果一张结果当前已经关联有效资产，而管理员仍想再次导入，界面应显式区分“重新导入并改绑当前关联”和“额外再导入一份新资产”两种动作，而不是混成一个模糊入口；这样当前资产关系和历史导入审计都能保持清晰。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 18:00 +0800
- 进度：确认管理端图像生成下一阶段的多次导入展示语义，并将其收敛为 `admin 多次导入主视图只显当前资产`。规则是：如果同一张结果经历过多次导入或重新导入，任务详情主视图只展示当前有效资产，不把多个旧资产并排堆在主区域里；历史导入关联、旧资产 ID 和导入时间链路收进审计上下文。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:57 +0800
- 进度：确认管理端图像生成下一阶段的删除后重导入关联语义，并将其收敛为 `admin 重新导入改绑新资产`。规则是：如果管理员对一张“曾导入过、后来资产被删除”的结果再次执行导入并成功，任务详情当前关联应切换到这次新生成的现存资产，而不是继续复用旧的已删除资产 ID；旧资产 ID 仍可作为历史痕迹保留，但不再代表当前有效关联。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:55 +0800
- 进度：确认管理端图像生成下一阶段的已删除导入资产补救语义，并将其收敛为 `admin 已删除导入资产重新导入入口`。规则是：如果某张结果曾经成功导入过，但对应的图片资产后来被删除，任务详情仍应保留一个明确的重新导入入口，而不是只剩删除提示；这样管理员可以直接把这张历史结果重新落回媒体库，作为最自然的补救动作。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:53 +0800
- 进度：确认管理端图像生成下一阶段的导入进度稳定语义，并将其收敛为 `admin 图像生成导入进度不回退`。规则是：任务列表和任务详情里的“已导入 x/n”统计的是这些结果曾经完成过多少次导入，不因后续图片资产被删除而回退；资产删除后的异常状态应在详情里单独提示。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:50 +0800
- 进度：确认管理端图像生成下一阶段的已删除导入资产展示语义，并将其收敛为 `admin 已删除导入资产回显`。规则是：如果某张已导入结果对应的图片资产后来被删除，任务详情仍应保留该资产 ID，并明确提示“资产已删除”，而不是把这张结果重新视为“从未导入”；这样管理员能区分“曾经导入过但后来被删了”和“从未导入过”这两种不同历史。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:47 +0800
- 进度：确认管理端图像生成下一阶段的已导入资产标题展示语义，并将其收敛为 `admin 已导入资产当前标题优先`。规则是：如果某张已导入结果对应的图片资产后来被重命名，任务详情应优先显示该资产的当前标题，同时继续保留资产 ID 作为稳定审计锚点；不另外冻结一份导入时标题快照。这样任务详情与媒体库当前状态保持一致，而审计定位仍可依赖稳定 ID。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:41 +0800
- 进度：确认管理端图像生成下一阶段的合集跳转返回语义，并将其收敛为 `admin 生图导入合集跳转返回上下文`。规则是：如果管理员从任务详情跳到带定位上下文的“AI 生图导入合集”后再返回，界面应回到原来的任务详情上下文，而不是停留在合集页；这样合集跳转和单资产跳转的往返语义才能保持一致。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:37 +0800
- 进度：确认管理端图像生成下一阶段的单资产跳转返回语义，并将其收敛为 `admin 已导入资产跳转返回上下文`。规则是：如果管理员从任务详情跳到单个已导入资产后再返回，界面应回到原来的任务详情上下文，而不是掉回图片列表默认视图；这样这条跳转才能既保证定位准确，也保证往返查看时的操作成本足够低。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:34 +0800
- 进度：确认管理端图像生成下一阶段的单资产跳转定位语义，并将其收敛为 `admin 已导入资产跳转定位上下文`。规则是：如果管理员从任务详情跳到单个已导入资产，目标页面应默认直接落到该资产的详情抽屉、详情页或等价高亮态，而不是只打开图片列表页；这样这条跳转才能真正表达“带我去这张刚导入的图”，而不是退化成普通列表导航。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:32 +0800
- 进度：确认管理端图像生成下一阶段的合集跳转定位语义，并将其收敛为 `admin 生图导入合集跳转定位上下文`。规则是：如果管理员从任务详情跳到“AI 生图导入合集”，目标页面应默认带上这次刚导入结果的高亮或筛选上下文，优先只突出相关资产，而不是直接落到合集全量列表里让管理员自己查找；这样这条跳转才能真正闭环到“刚导入了哪几张图”。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:30 +0800
- 进度：确认管理端图像生成下一阶段的批量导入部分成功反馈语义，并将其收敛为 `admin 快捷导入部分成功汇总反馈`。规则是：如果“导入全部未导入结果”执行后同时存在成功项和失败项，界面应先给出一个明确的汇总反馈，例如“成功 3 张，失败 1 张”或等价信息；随后仍按既定规则自动落到第一张失败结果。这样管理员先知道整批处理的总结果，再继续处理异常项。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:24 +0800
- 进度：确认管理端图像生成下一阶段的整批导入成功反馈语义，并将其收敛为 `admin 快捷导入整批成功反馈`。规则是：如果“导入全部未导入结果”覆盖的结果全部成功导入，界面应给出一个明确的整批完成反馈，例如“已导入 4 张到 AI 生图导入合集”或等价信息；这样管理员在批量处理时不必只靠列表状态变化来推断整批已经完成，也能直接知道这次成功导入了多少张结果。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 17:13 +0800
- 进度：确认管理端图像生成下一阶段的单张导入成功反馈语义，并将其收敛为 `admin 单张导入成功短反馈`。规则是：管理员在任务详情里手动导入当前结果图成功后，即使界面随后自动切到下一张待导入结果，也应给出一个短暂但明确的成功反馈，例如“已导入到 AI 生图导入合集”或等价信息；这样管理员既能保持连续处理节奏，也能确认刚才这一步已经真正落库。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:43 +0800
- 进度：确认管理端图像生成下一阶段的单张导入失败停留语义，并将其收敛为 `admin 单张导入失败原地停留`。规则是：如果管理员在任务详情里手动导入当前结果图时失败，界面应停留在这张结果上，并原地显示该次导入的可读失败原因，而不是自动跳到其他结果或其他任务；这样单张导入与批量导入失败时的异常处理语义保持一致，管理员也不用重新定位失败项。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:41 +0800
- 进度：确认管理端图像生成下一阶段的快捷导入失败聚焦语义，并将其收敛为 `admin 快捷导入失败首项自动聚焦`。规则是：如果“导入全部未导入结果”执行后存在失败项，任务详情应自动聚焦到第一张失败结果，并直接显示该结果的可读失败原因；这样管理员在批量导入后会先落到需要处理的异常项，而不是再手动翻找失败结果。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:37 +0800
- 进度：确认管理端图像生成下一阶段的快捷批量导入执行语义，并将其收敛为 `admin 未导入结果快捷导入直通执行`。规则是：管理员一旦明确点击“导入全部未导入结果”，系统就应直接按默认标题规则逐张导入当前任务里所有未导入结果，不再为每张结果额外弹出确认或编辑步骤；每张结果仍分别记录成功、失败和对应资产，某一张失败时也不回滚同一批里已经成功导入的其他结果。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:28 +0800
- 进度：确认管理端图像生成下一阶段的导入后连续处理语义，并将其收敛为 `admin 导入后自动切换下一张待导入结果`。规则是：如果管理员刚在任务详情里把当前结果图导入媒体库，而同一任务下仍有其他未导入结果，界面应自动切换并聚焦到下一张未导入结果，帮助管理员连续完成处理；如果这已经是最后一张未导入结果，则停留在当前任务并显示“已全部导入”，不要自动跳到别的任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:26 +0800
- 进度：确认管理端图像生成下一阶段的失败首现自动聚焦语义，并将其收敛为 `admin 失败首现自动聚焦`。规则是：如果管理员当前停留在某个任务详情里，而该任务第一次进入 `failed` 且已经有可读失败摘要，界面应自动把注意力带到失败摘要区域，但只触发一次；这样管理员在自动刷新过程中能及时看到失败原因，而不是只看到状态标签变红。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:21 +0800
- 进度：确认管理端图像生成下一阶段的结果首现自动聚焦语义，并将其收敛为 `admin 结果首现自动聚焦`。规则是：如果管理员当前停留在某个任务详情里，而该任务第一次从“尚无结果图”变成“已有结果图”，界面应自动把注意力带到结果区域，但只触发一次；这样管理员能及时看到新产出，又不会被重复滚动打断。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:15 +0800
- 进度：确认管理端图像生成下一阶段的任务详情停留态刷新语义，并将其收敛为 `admin 任务详情停留态自动切换`。规则是：如果管理员当前正停留在某个 `queued` 或 `running` 的任务详情上，而刷新后该任务已经进入 `succeeded`、`failed` 或 `canceled`，详情区应原地切换到最新状态与结果，而不是要求管理员手动重选任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 16:05 +0800
- 进度：确认管理端图像生成下一阶段的已导入结果审计上下文语义，并将其收敛为 `admin 已导入结果审计上下文`。规则是：对于已经导入媒体库的结果图，任务详情应直接显示其导入时间和对应资产的标题或 ID，作为最基础的审计上下文；这样管理员在跳转到资产或合集前，就能先确认这张图是何时、以什么资产身份进入媒体库。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:56 +0800
- 进度：确认管理端图像生成下一阶段的待导入结果优先展示语义，并将其收敛为 `admin 待导入结果优先视图`。规则是：如果一个成功任务同时包含已导入和未导入的结果图，任务详情默认优先展示仍待处理的结果，并把已处理完成的结果弱化或折叠；这样管理员打开详情后会先落在下一步待办上。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:43 +0800
- 进度：确认管理端图像生成下一阶段的导入合集跳转语义，并将其收敛为 `admin 生图导入合集跳转`。规则是：对于已经导入媒体库并进入生图导入合集的结果图，任务详情应直接显示它当前进入的图片合集，并提供跳转到该合集的入口；这样管理员可以从创作历史顺着结果直接进入目标合集，而不需要在媒体库里再次定位。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:41 +0800
- 进度：确认管理端图像生成下一阶段的未导入结果快捷导入语义，并将其收敛为 `admin 未导入结果快捷导入`。规则是：成功任务详情可以提供“导入全部未导入结果”的快捷入口，帮助管理员降低重复操作成本；但它只是对多个单张导入动作的便利封装，不把这批结果改造成必须整体成败一致的批处理实体。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:38 +0800
- 进度：确认管理端图像生成下一阶段的已导入资产跳转语义，并将其收敛为 `admin 已导入资产跳转`。规则是：对于已经导入媒体库的结果图，任务详情应直接提供跳转到对应图片资产或图片管理位置的入口，方便管理员从创作历史直接进入后续媒体库管理，而不是再手动搜索这张图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:32 +0800
- 进度：确认管理端图像生成下一阶段的全部导入完成态可见性，并将其收敛为 `admin 图像生成全部导入完成态`。规则是：如果一个成功任务的全部结果都已经导入媒体库，那么任务列表应给出明确的“已全部导入”完成态标识，帮助管理员快速跳过这类已经处理完的任务，优先关注仍待处理的结果。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:27 +0800
- 进度：确认管理端图像生成下一阶段的结果导入进度可见性，并将其收敛为 `admin 图像生成导入进度回显`。规则是：如果一个任务产出多张结果，其中只有部分已经导入媒体库，那么任务列表和任务详情都应直接显示这批结果当前的导入进度，例如“已导入 x/n”；这样管理员能立刻看出还剩多少张未处理。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:21 +0800
- 进度：确认管理端图像生成下一阶段的失败任务重试入口可达性，并将其收敛为 `admin 失败任务重试可达性`。规则是：失败任务在服务端创作历史的任务列表和任务详情里都应提供稳定可见的重试入口，方便管理员从失败排查直接回到预填工作台；该入口仍只负责打开预填工作台，不绕过既有的显式再提交语义。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:18 +0800
- 进度：确认管理端图像生成下一阶段的任务列表刷新语义，并将其收敛为 `admin 图像生成任务列表刷新`。规则是：服务端创作历史的任务列表支持低频自动刷新，让管理员停留页面时逐步看到状态变化；同时保留显式手动刷新入口，方便在需要时立刻拉取最新状态；不要求做成高频实时推送。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:10 +0800
- 进度：确认管理端图像生成下一阶段的排队感知语义，并将其收敛为 `admin 图像生成排队感知`。规则是：对于 `queued` 任务，管理端可以展示“前方还有几项”或“当前排队位置”这类近似排队感知，帮助管理员理解任务并未卡死；但它不承诺严格全局顺序，也不承诺精确 ETA，以保持和现有任务调度公平性一致。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:08 +0800
- 进度：确认管理端图像生成下一阶段的任务列表行摘要语义，并将其收敛为 `admin 图像生成任务行摘要`。规则是：服务端创作历史的任务列表每一行除了状态标签外，还应直接展示最小摘要，帮助管理员先扫描再决定是否展开详情；摘要优先包括首张结果缩略图、结果数量或目标张数，以及失败任务的可读失败摘要。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:53 +0800
- 进度：确认管理端图像生成下一阶段的输入变更审计语义，并将其收敛为 `admin 重试输入变更审计`。规则是：如果管理员在重试入口中显式修改提示词、尺寸、质量、张数等输入参数，导致新任务输入不同于原任务，服务端创作历史应明确记录这是一次“已修改输入”的新任务，而不只保留普通的重试来源关系，以区分原样重试、参考图集合变更与文本/参数变更。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:51 +0800
- 进度：确认管理端图像生成下一阶段的参考图集合变更审计语义，并将其收敛为 `admin 重试参考图集合变更审计`。规则是：如果管理员在重试入口中显式移除、补齐、替换或重排参考图，导致新任务的参考图集合不同于原任务，服务端创作历史应明确记录这是一次“已修改参考图集合”的新任务，而不只保留普通的重试来源关系，以区分原样重试与意图已变更的再次提交。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:46 +0800
- 进度：确认管理端图像生成下一阶段的参考图顺序语义，并将其收敛为 `admin 重试参考图槽位顺序`。规则是：重试入口里的参考图顺序和槽位位置都属于原任务意图的一部分；系统在预填和补图时保留原顺序与缺失位位置，不自动重排剩余参考图；管理员若要改变顺序，应显式调整后再提交新任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:43 +0800
- 进度：确认管理端图像生成下一阶段的部分失效参考图重试语义，并将其收敛为 `admin 重试参考图缺失位`。规则是：重试预填的多张参考图里如果只有一部分仍然有效，系统保留有效参考图，并把失效项明确显示为带原因的缺失位，而不是静默丢弃；管理员必须补齐这些缺失位，或显式确认移除对应参考图后，才能提交新任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:35 +0800
- 进度：确认管理端图像生成下一阶段的临时参考图重试复用语义，并将其收敛为 `admin 重试复用临时参考图`。规则是：如果临时参考图仍在保留期内且可校验通过，重试入口可以直接复用，不要求管理员重复上传；只有在参考图已过期、丢失或校验失败时，才要求重新上传。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:32 +0800
- 进度：确认管理端图像生成下一阶段的浏览器参考图保留边界，并将其收敛为 `admin 临时参考图保留期`。规则是：浏览器上传或粘贴、且未绑定现有服务器图片资产的参考图，只按临时对象保存并受统一 TTL 清理；过期后历史仍保留轻量快照与“参考图已过期”提示，但重试不能直接复用该文件，管理员必须重新上传参考图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:23 +0800
- 进度：确认管理端图像生成下一阶段的重试预填失效模型交互，并将其收敛为 `admin 重试预填失效模型`。规则是：重试预填旧任务时，如果原模型在当前工作台已下线或不再允许选择，界面仍保留该旧值作为历史上下文，但必须明确标记“已失效”并阻止直接提交；管理员需要显式改选当前可用模型后，才能提交新任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:19 +0800
- 进度：确认管理端图像生成下一阶段的失效执行配置处理，并将其收敛为 `admin 图像生成失效执行配置`。规则是：如果任务快照指向的模型或关键代理配置后来不可用，系统不静默切换到新的默认配置，而是按失败处理并记录“原执行配置已失效”的可读原因；管理员若要继续生成，应在当前可用配置下重新提交新任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:14 +0800
- 进度：确认管理端图像生成下一阶段的执行配置冻结边界，并将其收敛为 `admin 图像生成执行配置快照`。规则是：任务进入 `queued` 时固定本次生成所使用的模型标识与关键执行配置；之后即使管理员修改默认模型或代理配置，已排队任务仍按提交时快照执行，新配置只影响后续新建任务。敏感密钥仍只属于服务端运行配置，不进入服务端创作历史。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:06 +0800
- 进度：确认管理端图像生成下一阶段的参数演进兼容策略，并将其收敛为 `admin 图像生成参数快照版本`。规则是：服务端历史保存任务提交时的参数快照和对应 `schema_version`；以后即使工作台支持新参数或改了表单结构，旧任务仍按当时快照语义解释，重试时再把旧快照映射到当前工作台表单。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 14:02 +0800
- 进度：确认管理端图像生成下一阶段的任务删除语义，并将其收敛为 `admin 图像生成任务历史保留`。规则是：默认不提供任务审计记录的硬删除；管理员可以清理未导入媒体库的临时结果图，或在界面上隐藏/归档历史项，但任务记录本身保留。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 13:59 +0800
- 进度：确认管理端图像生成下一阶段的任务修改边界，并将其收敛为 `admin 图像生成任务输入冻结`。规则是：任务一旦进入 `queued`，提示词、参数和参考图集合就冻结不再原地修改；管理员若想改任何输入，应取消原任务或保留原任务不动，再提交一个新任务表达新的意图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 13:47 +0800
- 进度：确认管理端图像生成下一阶段的队列调度语义，并将其收敛为 `admin 图像生成任务调度公平性`。规则是：单个管理员自己的任务序列按提交顺序 FIFO 执行；不同管理员之间不承诺严格全局顺序，系统可采用公平调度，避免单个管理员长队列长期占满执行资源。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:59 +0800
- 进度：确认管理端图像生成下一阶段的部分成功语义，并将其收敛为 `admin 图像生成部分可用结果`。规则是：不新增 `partial_success` 状态；只要最终可用结果数少于请求数，任务主状态就记为 `failed`，但已拿到且可用的结果图仍然保留，供后续单张导入、复用或查看。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:50 +0800
- 进度：确认管理端图像生成下一阶段的失败信息可见边界，并将其收敛为 `admin 服务端失败摘要回显`。规则是：界面默认只展示可读失败摘要、请求时间和请求 ID；上游原始报文、完整错误堆栈和敏感细节只进入服务端日志，不直接暴露给管理员界面。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:37 +0800
- 进度：确认管理端图像生成下一阶段的结果处理粒度，并将其收敛为 `admin 图像生成结果单张操作`。规则是：一次任务若生成多张结果，导入媒体库、复用为参考图、删除临时结果等后续动作都按单张结果独立执行，不把整批结果强绑为只能一起处理的集合。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:33 +0800
- 进度：确认管理端图像生成下一阶段的重复提交语义，并将其收敛为 `admin 图像生成任务提交去重`。规则是：管理员每次明确点击“生成”都视为新任务；只有同一次提交因网络重放、页面重试或请求重送而重复到达时，系统才使用短时幂等键折叠为同一个任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:31 +0800
- 进度：确认管理端图像生成下一阶段的重试入口默认行为，并将其收敛为 `admin 图像生成任务重试入口`。规则是：点击“重试”默认打开图像工作台并预填上次任务的提示词、参数和仍然可用的参考图，不做一键直接开跑；缺失、失效或已删除的参考图必须显式补齐后才能再次提交新任务。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:20 +0800
- 进度：确认管理端图像生成下一阶段的服务端历史可见范围，并将其收敛为 `admin 服务端图像创作可见范围`。规则是：任务与服务端创作历史默认只对创建它的管理员可见和可操作，不在多个管理员之间自动共享；只有导入媒体库后的正式图片资产继续按现有媒体库规则全局可见。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:14 +0800
- 进度：确认管理端图像生成下一阶段支持“服务器图片作为参考图”，并将其收敛为 `admin 服务端参考图引用`。规则是：参考图既可来自浏览器本地上传/粘贴，也可直接引用服务器里已有的图片资产；若引用服务器图片，服务端历史只保存资产 ID 与轻量元数据，不长期复制新的参考图原文件。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:06 +0800
- 进度：确认管理端图像生成下一阶段的保留期语义，并将其收敛为 `admin 图像生成结果保留期`。规则是：任务记录与轻量元数据长期保留；未导入媒体库的临时结果图按统一 TTL 清理，例如 30 天；已导入媒体库的正式图片资产不受该 TTL 影响。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:01 +0800
- 进度：确认管理端图像生成下一阶段的取消语义，并将其收敛为 `admin 图像生成任务取消`。规则是：只有 `queued` 和 `running` 可取消并进入 `canceled`；`succeeded` 与 `failed` 终态不可再取消。取消后保留任务元数据与取消原因，但不承诺保留未完成中间结果图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 10:53 +0800
- 进度：确认管理端图像生成下一阶段的重试语义为“重试创建新任务，旧任务保持原始终态不改写”，并将其收敛为 `admin 图像生成任务重试`。约束是：服务端历史保持不可变，重试只通过任务间关联表达，不把失败任务回写成成功或其他状态。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 10:44 +0800
- 进度：确认图像工作台结果导入媒体库后仍只进入现有图片资产与图片合集体系，不新增“生图合集”对象类型；同时锁定单一全局 `admin 生图导入合集` 规则，所有生图导入结果默认归入同一个全局图片合集，例如 `AI 生图导入`，若不存在则创建后复用，不按任务、日期或模型拆分。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 10:36 +0800
- 进度：确认管理端图像生成下一阶段的稳定任务状态模型为 `queued / running / succeeded / failed / canceled`，并将其收敛为 `admin 图像生成任务状态模型`。约束是：供应商细状态、重试细节和传输细节不升级为系统主状态，只保留在日志或错误摘要里。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 10:29 +0800
- 进度：确认管理端图像生成工作台可以进入下一阶段，并将下一阶段术语收敛为“admin 图像生成任务化”和“admin 服务端图像创作历史”。下一阶段方向锁定为服务端任务化、服务端生成历史、可恢复与跨浏览器查看；暂不优先扩展蒙版编辑、Agent 或多供应商协议。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 09:29 +0800
- 进度：完成图像工作台失败态可见性修复收尾。失败任务在本地历史中会保留并展示错误原因，选中失败任务后预览区直接显示失败空态；本次只纳入 `admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md` 与 `plan.md`，未纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`。
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 chunk size warning）；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/toolboxPage.spec.js` 通过；乱码扫描无输出。

## 2026-06-10 09:19 +0800
- 进度：修复图像工作台失败任务在本地历史里看不到原因的问题。失败任务现在会自动切换到当前预览，历史列表会显示失败原因摘要，预览区对失败任务显示失败空态而不是只留“暂无结果”。`CONTEXT.md` 已补充“图像工作台失败态可见”的长期约定。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`。
- 验证：待执行 `cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-10 09:09 +0800
- 进度：完成管理端生图默认模型切换到 `gpt-image-2` 的实现收尾。`internal/config.Load` 的环境默认值、示例环境变量与 `CONTEXT.md` 已统一到新默认；`internal/handlers/admin_image_generation_test.go` 的显式模型样例也同步为 `gpt-image-2`，避免测试文案继续暗示旧默认。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`internal/handlers/admin_image_generation_test.go`、`.env.example`、`CONTEXT.md`、`plan.md`。
- 验证：`go test ./internal/config ./internal/handlers -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- .env.example CONTEXT.md internal/config/config.go internal/config/config_test.go internal/handlers/admin_image_generation_test.go plan.md` 通过；乱码扫描无输出。

## 2026-06-09 23:21 +0800
- 进度：完成转码坏包容忍参数实现。`buildTranscodeVideoArgs` 现在在输入文件前加入 `-max_error_rate 1.0`、`-fflags +discardcorrupt`、`-err_detect ignore_err`，用于跳过局部损坏 packet 并忽略可恢复解码错误；HEVC primary 与 AVC compat 参数测试同步锁定这些选项。`CONTEXT.md` 已追加转码坏包容忍策略。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：红灯阶段 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 因缺少 `-max_error_rate 1.0` 失败；实现后同命令通过。待执行相关 Go 定向测试、全量 Go 验证、`git diff --check`、乱码扫描。

## 2026-06-09 22:55 +0800
- 进度：完成 iPhone MOV 多音轨转码失败修复。`buildTranscodeVideoArgs` 的音频映射从全部音频流 `0:a?` 收窄为首个可选音频流 `0:a:0?`，保留 AAC 主音轨并跳过 `apac` 等 ffmpeg 无法解码的额外音轨；HEVC primary 与 AVC compat 两个输出 profile 的参数测试已同步覆盖。`CONTEXT.md` 已沉淀转码音轨映射策略。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：红灯 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 先失败于仍生成 `-map 0:a?`；实现后 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -run 'TestBuildTranscodeVideoArgs|TestBuildTranscodeProgress|TestParseProgressValueToSeconds|TestBuildTranscodePlan|TestResolveTranscodePersistence' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；乱码扫描无输出。

## 2026-06-09 18:23 +0800
- 进度：完成管理端图像生成工作台第一阶段实现。工具箱新增“图像生成工作台”新标签页入口；新增 `/toolbox/image-workbench` 无 shell 管理员页面，提供三栏工作台、参考图上传/拖拽/粘贴、常用参数、同步生成、IndexedDB 本地历史、结果下载/复用为参考图/导入媒体库；后端新增管理员受限图像生成状态与同步代理接口，支持 OpenAI-compatible Images generate/edit、脱敏配置状态、资源限制、data URL 结果归一和轻量导入元数据；`.env.example` 增加图像生成配置键。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`.env.example`、`CONTEXT.md`、`plan.md`、`main.go`、`internal/config/config.go`、`internal/config/config_test.go`、`internal/handlers/router.go`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.db.js`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/views/toolboxPage.spec.js`
- 验证：`go test ./internal/config ./internal/handlers -run 'TestLoadIncludesImageGenerationConfig|TestRegisterIncludesAdminImageGenerationRoutes|TestAdminImageGenerationStatusIsRedacted|TestNormalizeAdminImageGenerationRequestRejectsTooManyImages|TestAdminImageGenerate' -count=1` 通过；`go test ./internal/... -count=1` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-09 18:22 +0800
- 进度：实现中按官方 GPT Image 接口校正后端代理细节：默认模型改为 `gpt-image-1.5`；官方 `gpt-image-*` 模型不发送 `response_format`，非 GPT Image 兼容模型才请求 `b64_json`；参考图编辑 multipart 文件字段使用 `image`。前端仍统一接收后端归一后的 data URL 结果。
- 影响文件：`internal/config/config.go`、`.env.example`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待复跑后端定向测试、管理端测试/构建与 `git diff --check`。

## 2026-06-09 15:25 +0800
- 进度：完成管理端工具箱菜单合集与 ED2K 无 shell 新标签页实现。`/toolbox` 现在只展示工具菜单入口，ED2K 入口用 `router.resolve('/toolbox/ed2k').href` 生成带 SPA 基路径的链接并通过 `target="_blank"` 新标签页打开；新增 `/toolbox/ed2k` 管理员受限路由，页面不包后台 Layout，保留“返回工具箱”、ED2K 多行解析、非法行计数和本页会话“已点击”标记。命令面板 `ed2k` 搜索仍命中 `/toolbox`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/router/index.js`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd admin-web && npm test -- src/views/toolboxPage.spec.js src/components/base/commandPalette.helpers.spec.js` 先失败于缺少 `ToolboxEd2k.vue`；实现后 `cd admin-web && npm test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js src/components/base/commandPalette.helpers.spec.js` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-08 21:02 +0800
- 进度：完成杜比视界安全播放第一阶段实现。后端在新视频转码流程中对原始源和最终播放文件做只读播放兼容探测，写入 `videos.metadata.playback_compat`，探测失败不让转码任务失败；TV App 根据 metadata 放行历史视频、阻断探测失败/结构不完整视频、阻断原始源为 Dolby Vision 但当前压缩输出非 Dolby Vision 的风险播放，并让剧集默认选集/自动连播跳过兼容阻断分集。未实现 Media3/DV 专用系统播放分支，未保留原始视频，未改变后端转码压缩参数。TV 端版本号升至 `0.1.84`。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`internal/models/app.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesAutoplay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、TV 端相关测试、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue ./internal/repository -run 'TestParsePlaybackCompatibility|TestBuildPlaybackCompatibility|TestBuildTranscodeTaskOptions|TestResolveTranscodePersistence|TestResolveTVEpisodeStillURL' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest' --tests 'com.chee.videos.feature.tv.TvSeriesAutoplaySpecTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-07 14:44 +0800
- 进度：完成管理端工具箱菜单迁移的最终校验，准备提交。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-07 14:43 +0800
- 进度：完成管理端“工具箱”独立菜单与页面。ED2K 链接生成器已从系统设置页迁入工具箱，系统设置页只保留临时文件清理、孤儿文件扫描和日志查看等系统运维能力；命令面板支持通过“工具箱”、`gjx`、`ed2k` 搜索入口。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；待执行最终 `git diff --check` 与乱码扫描。

## 2026-06-07 14:00 +0800
- 进度：完成管理端 ED2K 链接生成器。系统设置页新增多行输入框，逐行生成可点击 ED2K 链接，非法非空行会计数提示，链接点击后在当前页面会话内显示“已点击”；解析逻辑已沉到 helper 并补单测，`CONTEXT.md` 已记录该工具不提交后端、不持久化 ED2K 内容。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-06 12:22 +0800
- 进度：完成 TV 电视剧详情页精修。标题区改为左对齐单行/双行剧名，不再渲染标题上方“剧集”；元信息行去掉 `18+` 年龄角标；剧情摘要固定至少四行；TV 端版本号升至 `0.1.83`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`adb -s 192.168.1.8:5555 install -r android-tv-app/tv-app/build/outputs/apk/debug/tv-app-armeabi-v7a-debug.apk` 成功；实机截图 `/tmp/tv-series-detail-refine-2.png` 确认左侧剧名左对齐、标题上方无“剧集”、元信息无 `18+`、剧情摘要显示四行；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-06 12:07 +0800
- 进度：完成 TV 电视剧详情页参考图主体优化并按最新反馈去掉左侧/顶部导航栏。详情页现在保留全屏背景、独立返回按钮、小字号左侧信息区、暖金播放/焦点视觉、右侧剧照分集列表；TV 端版本号已升至 `0.1.82`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`adb -s 192.168.1.8:5555 install -r android-tv-app/tv-app/build/outputs/apk/debug/tv-app-armeabi-v7a-debug.apk` 成功；实机截图 `/tmp/tv-after-load.png` 确认左侧/顶部导航栏已去掉，主体小字号与右侧分集列表正常；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-06 12:01 +0800
- 进度：根据用户最新反馈调整范围：去掉电视剧详情页左侧竖向导航栏和顶部英文导航栏；保留已完成的小字号左侧详情信息、暖金焦点、右侧分集剧照列表与全屏背景参考图风格。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待重跑 TV 端定向单测、`:tv-app:assembleDebug`、ADB 实机截图、`git diff --check`、乱码扫描。

## 2026-06-05 10:55 +0800
- 进度：管理端孤儿文件扫描前后端接入已完成并通过验证。系统页现已支持异步扫描、轮询最新状态、完成后自动弹出全量删除确认，以及删除后保留扫描快照；后端迁移、仓库、队列、handler 与 API 单测已补齐。无关工作区改动 `admin-web/.env.development` 仍未纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_orphan_file_scan_test.go`、`internal/repository/migrations_test.go`、`internal/repository/orphan_file_scan_repository_test.go`、`internal/queue/orphan_file_scan_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`
- 验证：`go test ./internal/... -count=1` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；乱码扫描无输出

## 2026-06-05 10:51 +0800
- 进度：完成管理端孤儿文件扫描的前后端接入补齐。后端把孤儿文件扫描单独挂到系统页可用的仓库/队列接口上，补了迁移与纯逻辑测试；管理端系统设置页新增“开始扫描 -> 自动轮询最新状态 -> 扫完自动确认全量删除”的交互，并把删除操作收口为全量删除，不再走逐项勾选。无关工作区改动 `admin-web/.env.development` 仍不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_orphan_file_scan_test.go`、`internal/repository/migrations_test.go`、`internal/repository/orphan_file_scan_repository_test.go`、`internal/queue/orphan_file_scan_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`
- 验证：待执行 `go test ./internal/... -count=1`、`cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描

## 2026-06-04 12:45 +0800
- 进度：完成全媒体外盘访问收口。视频源、视频字幕、演员头像、图片视图、视频抓帧与图片上传后的外盘探测都改成限时打开/限时 ffmpeg 处理；`CONTEXT.md` 也同步补了“媒体资源限时打开”和“媒体生成限时”的长期约定。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/local_image_file.go`、`internal/handlers/admin_image.go`、`internal/handlers/app_image_collection.go`、`internal/handlers/video_source.go`、`internal/handlers/video_subtitle.go`、`internal/handlers/actor_avatar.go`、`internal/handlers/admin_video_thumbnail.go`、`internal/services/image.go`、`internal/handlers/local_image_file_test.go`
- 验证：`go test ./internal/handlers ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md internal/handlers/local_image_file.go internal/handlers/admin_image.go internal/handlers/app_image_collection.go internal/handlers/local_image_file_test.go internal/handlers/video_source.go internal/handlers/video_subtitle.go internal/handlers/actor_avatar.go internal/handlers/admin_video_thumbnail.go internal/services/image.go` 通过；乱码扫描无输出。

## 2026-06-04 11:06 +0800
- 进度：完成管理端入口 HEAD 探测修复。`/admin` 与 `/admin/` 现在同时支持 GET/HEAD 并返回同一套 `no-store` 入口响应，避免部署后 `curl -I` 或缓存探测误报 404。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- internal/handlers/router.go internal/handlers/admin_static_test.go plan.md` 通过；`rg -n $'\uFFFD' internal/handlers/router.go internal/handlers/admin_static_test.go plan.md` 无输出。

## 2026-06-04 10:58 +0800
- 进度：已补管理端静态资源缓存回归测试并确认红灯。`TestMountAdminStaticServesIndexFromGivenDir` 现在要求 `/admin/` 返回 `Cache-Control: no-store`、存在的 `/admin/assets/*` 返回长缓存；`TestMountAdminStaticMissingAssetIsNotLongCached` 要求缺失旧 hash asset 404 且不长缓存。当前失败于 `adminIndexCacheControl` / `adminAssetCacheControl` / `adminMissingAssetCacheControl` 未实现。
- 影响文件：`internal/handlers/admin_static_test.go`、`plan.md`
- 验证：红灯 `go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 失败于上述未定义常量。

## 2026-06-04 01:35 +0800
- 进度：完成图片上传 `ffmpeg` PATH 缺失回退修复收口。最终实现只改 `pkg/ffmpeg.ConvertToWebP` 的缺失二进制分支，让图片上传在 launchd / 受限 PATH 环境里也能改走 `cwebp`，两者都不可用时仍返回 `ErrWebPEncodingUnavailable` 走保留原图兜底。`CONTEXT.md` 已同步“ffmpeg 不在 PATH 也算 WebP 编码链路不可用”的长期契约。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg -run TestConvertToWebPFallsBackWhenFFmpegMissing -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 无输出；工作区既有改动 `admin-web/.env.development` 未纳入本次修复。

## 2026-06-04 01:34 +0800
- 进度：修复图片上传在 launchd / 受限 PATH 环境下找不到 `ffmpeg` 可执行文件时的回退链路。`pkg/ffmpeg/ConvertToWebP` 现在把 `exec.ErrNotFound` 视为 WebP 编码能力不可用，直接降级到 `cwebp`；若 `cwebp` 也不可用，仍返回 `ErrWebPEncodingUnavailable` 让图片上传保留原始 JPEG/PNG。`CONTEXT.md` 补充了“ffmpeg 不在 PATH 中”也属于图片上传 WebP 编码不可用契约。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg -run TestConvertToWebPFallsBackWhenFFmpegMissing -count=1` 通过；`git diff --check -- pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；待执行全量相关包验证、乱码扫描和提交。

## 2026-05-31 15:44 +0800
- 进度：完成 TV 电视剧详情页返修收尾验证。TV 全量单测首次因新增 `RoundedCornerShape(6.dp)` 违反 TV 圆角白名单失败，已改为复用 `AppChrome.ChipShape` 并重跑通过。准备只暂存本任务文件，继续保留用户既有改动 `admin-web/.env.development` 不纳入提交。
- 影响文件：同 15:43 记录；提交范围不包含 `admin-web/.env.development`
- 验证：`go test ./internal/utils ./internal/repository ./internal/handlers ./internal/services -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-31 15:43 +0800
- 进度：完成 TV 电视剧详情页返修实现。详情主体继续保持无全局侧栏/顶部导航，左侧重做参考图式居中大标题、胶囊元信息、金色评分与「播放第 X 集 / 我的片单」操作；右侧分集列表改为更高的横向剧照卡、金色选中边框与圆形播放按钮。后端新增 TV 分集 still 本地访问路由，详情接口返回本地 still 路由，刮削同步分集时下载到 `storage/tv/series/<id>/episodes/sXXeYY.jpg`。TV 版本升级到 `0.1.80 (80)`，`CONTEXT.md` 追加详情主体还原边界与分集剧照本地化约定。
- 影响文件：`internal/utils/video_url.go`、`internal/utils/video_url_test.go`、`internal/repository/tv_repository.go`、`internal/repository/tv_episode_still_url_test.go`、`internal/handlers/tv_artwork.go`、`internal/handlers/router.go`、`internal/services/scraper.go`、`internal/services/scraper_episode_sync_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest'` 通过；`go test ./internal/utils ./internal/repository ./internal/services -run 'TestTVEpisodeStillURL|TestResolveTVEpisodeStillURL|TestScrapeEpisodeUploadDownloadsSeriesArtworkLocally' -count=1` 首次因 sandbox 禁止 `httptest` 监听本地端口失败，提权重跑通过；待执行后端受影响包全量、TV 全量单测、`git diff --check`、乱码扫描与提交。

## 2026-05-31 13:23 +0800
- 进度：完成电视剧详情页主体沉浸式改造。`TvSeriesDetailScreen` 改为全屏背景 + 左侧剧集信息/演员区 + 右侧剧集列表，保留共享返回图标与播放/季/集焦点视觉；`TvEpisodeUiModel` 透传 `stillUrl` 供右侧分集卡显示缩略图；TV 版本升级到 `0.1.79 (79)`，`CONTEXT.md` 补充电视剧沉浸式详情主体、shared poster 目标与安全区域边界。范围仍不包含全局侧栏/顶部导航。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRepositoryMappingTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest'` 因 `TvEpisodeUiModel.stillUrl` 未定义编译失败；实现后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest' --tests 'com.chee.videos.feature.tv.TvScrollableBottomPaddingTest' --tests 'com.chee.videos.core.ui.TvSharedPosterTransitionSpecTest' --tests 'com.chee.videos.core.ui.TvShapeAuditTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-31 02:28 +0800
- 进度：完成电视剧播放器选集 gap 与 controls 左右焦点返修验证。确认选集 gap 的根因是焦点标题气泡参与 `LazyRow` item 横向测量，本轮已改为固定槽位 + 气泡覆盖；controls 左右键在持焦控件层消费并请求相邻按钮，根播放器继续只负责根层 seek。收尾只准备提交本任务文件，保留用户既有改动 `admin-web/.env.development` 不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerControlsFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerControlsFocusPolicyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerControlsFocusPolicyTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.TvEpisodeRailPolicyTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerControlsFocusPolicyTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt` 通过；乱码扫描无输出。

## 2026-05-31 02:25 +0800
- 进度：完成电视剧播放器返修实现。`LongFormVideoPlayer` 将选集轨集卡改为固定槽位，标题气泡改为不参与横向测量的覆盖信息，避免焦点切换时在左右集卡之间撑出 gap；controls 持焦按钮新增本地 LEFT/RIGHT 焦点请求，仍保留 `focusProperties` 首尾环绕链，避免左右键回落为播放器根层 seek。新增焦点链纯逻辑测试与源文回归，TV 版本升级到 `0.1.78 (78)`，`CONTEXT.md` 补充“选集轨固定槽位”和“controls 持焦横向导航”。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerControlsFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerControlsFocusPolicyTest'` 通过；待执行电视剧播放器相关定向单测、TV 全量单测、`git diff --check`、乱码扫描。

## 2026-05-31 01:50 +0800
- 进度：完成电视剧播放器“中间切换多按一次下键”修复。`LongFormVideoPlayer` 在电视剧 controls 持焦层增加本地 `DPad DOWN` 兜底，确保第二次 DOWN 直接进入选集轨；保留 controls/选集轨焦点请求重试，选集卡继续只用静态高亮无 glow。同步更新源文审计、TV 版本到 `0.1.77 (77)`，并在 `CONTEXT.md` 追加“播放器内纵向切页无空按”约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt` 通过；乱码扫描无输出。

## 2026-05-31 01:13 +0800
- 进度：完成电视剧播放器交互 bug 修正。`LongFormVideoPlayer` 的电视剧分支改为底部单容器双页：`controls 页` 与 `选集页` 通过纵向 slide 切换，不再同时堆成双层；电视剧首屏恢复为按 DPad DOWN 唤出 controls；controls 聚焦时 LEFT/RIGHT 只切焦点，不做 seek。选集轨集卡改为“第一集”“第二集”等中文序数，左右切集时移除集卡淡入淡出和整段明显滚动，只在目标集快出边界时做最短必要跟随；连续快进/快退时进度显示改为跟随 pending seek 目标，避免进度条抖动。TV 版本升级到 `0.1.76 (76)`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvEpisodeRailPolicy}.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvEpisodeRailPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvEpisodeRailPolicyTest' --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-05-30 23:55 +0800
- 进度：完成电视剧播放器 UI 重构收尾。`LongFormVideoPlayer` 新增 `SeriesEpisodeRail` 变体、`播放器根 → controls → 选集轨` 三层焦点路由、只读进度条与横向选集轨；`TvSeriesPlayerScreen` 移除旧 `ModalBottomSheet` 选集，改为当前季分集卡片内嵌轨道并通过标题气泡跟随焦点；补充选集轨策略 / 遥控路由 / 源文审计测试，并修正 `TvScrollableBottomPaddingTest` 使其不再把沉浸式播放器页当滚动页。TV 版本升级到 `0.1.75 (75)`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting,TvEpisodeRailPolicy}.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/{TvLongFormRemoteKeyRoutingTest,TvLongFormControlsAutoHideTest,TvEpisodeRailPolicyTest}.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/{TvSeriesEpisodeRailSpecTest,TvScrollableBottomPaddingTest}.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-05-30 09:38 +0800
- 进度：完成 TV App 退出 IPTV 播放卡住修复。IPTV `LibVLC` 改为独立单例复用，退出页面时不再在 `onDispose` 主线程同步 `stop()` 直播流或释放库实例，只保留 `MediaPlayer.release()`；`detachViews()` 改为独立 `DisposableEffect` 处理，和长视频 LibVLC 生命周期保持同类模式。同步补充 IPTV 配置/源文测试约束，TV 版本升级到 `0.1.74 (74)`，`CONTEXT.md` 追加 IPTV 退出清理契约。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvPlaybackConfig.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackConfigTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackConfigTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest' --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-30 09:27 +0800
- 进度：完成 TV App IPTV 卡顿优化首轮落地。新增 IPTV 专属 LibVLC 播放配置 helper，将直播缓存从 `1500ms` 提升到 `4000ms`，移除 `clock-jitter=0` / `clock-synchro=0` 低延迟参数，统一收口到 helper 以防后续散落魔法数字；同步补充 IPTV 配置单测与源文约束测试，TV 版本升级到 `0.1.73 (73)`，`CONTEXT.md` 追加 [[IPTV 流畅优先]] 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvPlaybackConfig.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackConfigTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackConfigTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest' --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-28 05:02 +0800
- 进度：完成 TV App IPTV 硬解默认改造。移除 IPTV LibVLC 初始化参数 `--avcodec-hw=none`，频道 Media 改为 `setHWDecoderEnabled(true, true)`；补源文回归测试禁止回到关闭硬解；TV 版本号升至 `0.1.72` / `versionCode=72`；`CONTEXT.md` 中 IPTV 旧“软解优先”约定改为“TextureView + 硬解默认”，保留直播诊断路径。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n "avcodec-hw=none|setHWDecoderEnabled\\(false|setHWDecoderEnabled\\(true, true\\)|versionCode =|versionName =|软解优先|关闭硬解" ...` 确认 IPTV 已改为硬解默认且版本已更新；`rg -n $'\uFFFD' ...` 无乱码；`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest` 未通过环境前置，当前命令行 Java 为 1.8，AGP 8.5.2 需要 Java 11+。

## 2026-05-25 16:30 +0800
- 进度：完成 `tasks/2026-05-25-tv-long-form-libvlc-migration` 的 grill-with-docs 设计沉淀。通过 9 轮决策拷打锁定：（Q1）ASS 渲染完整支持卡拉 OK / 动态特效 / 矢量绘图；（Q2）电影/`18+`/电视剧三类一起切，IPTV 不动；（Q3）字幕走 LibVLC 自渲染（libass），不接 JNI；（Q4）不引入 Player 抽象、不留 Media3 fallback；（Q5）TextureView + 硬解默认；（Q6）SRT/VTT 接受 libass 默认外观；（Q7）后端不再 ASS→VTT 转换，只存 ASS 原文；（Q8）一次性合 + 内部 PoC 前置；（Q9）无遗漏。产出 PRD/implement/review 三件套与 ADR-0004。本任务尚未进入实施阶段——CONTEXT.md 的 6 条新术语（[[TV 长视频 LibVLC 内核]] / [[libass 自渲染字幕]] / [[TV 长视频 TextureView 硬解默认]] / [[ASS 字幕原文存储策略]] / [[字幕样式 libass 让位]] / [[LibVLC track id 不稳定]]）与 line 9 / line 170 的旧约定推翻动作均锁定到实施完成时再 sync，避免未实施先沉淀。
- 影响文件：`tasks/2026-05-25-tv-long-form-libvlc-migration/{prd,implement,review}.md`、`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`、`plan.md`
- 验证：本轮仅文档沉淀，无代码改动；实施任务将单独触发，第一步是分支 `chore/libvlc-poc` 验证 LibVLC + libass + 硬解的真实可行性，通过后才进入主线一次性合。

## 2026-05-25 16:10 +0800
- 进度：修复 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote` 真机回归发现的两个故障：5 秒 UI 自动隐藏后遥控器失效 30 秒；一次 ←/→ 唤起 UI 后无法继续 seek。根因为 `LongFormVideoPlayer` 根 Box 的 `Modifier.focusRequester(rootFocusRequester)` 错排在 `Modifier.focusable()` 之后，导致 `tryRequestFocus()` 静默失败、焦点无人锚定，按钮 Compose 兜底获焦后 `focusInControls` 被卡 true 走焦点环绕。修复：交换两个 modifier 顺序回到项目惯例（与 `TvIptvScreen` / `TvPosterWallScreen` 一致），并在 `handleTvRemoteKeyAction` 的 Seek 分支追加 `focusInControls = false` + `requestRootFocusWhenReady()` 作为按钮入场抢焦的防御兜底。补两条源文审计红灯（modifier 顺序、Seek 分支兜底），TV 版本号升到 `0.1.67`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`、`tasks/2026-05-24-tv-long-form-operation-ui-on-remote/feedback.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest'` 先红（modifier 顺序 + Seek 兜底断言失败）后绿；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 全量通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；真机/模拟器手测待后续在有设备环境下回归。

## 2026-05-25 13:46 +0800
- 进度：完成管理端视频管理列表详情按钮无反应修复。根因是 `showDetail()` 创建 `detailRequestToken` 后又调用会递增 token 的清理函数，导致详情接口返回后被判定为过期并丢弃；`refreshPlayURL()` 也会在当前详情上下文内失效 token。修复为新增 token helper，清理函数支持 `invalidateToken` 参数，打开详情和刷新播放链接时清理旧预览/URL 但不取消当前请求；字幕加载也改用统一 stale 判断。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`plan.md`
- 验证：`cd admin-web && npm test -- src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' plan.md admin-web/src/views/VideoList.vue admin-web/src/views/videoList.helpers.js admin-web/src/views/videoList.helpers.spec.js` 无输出。

## 2026-05-24 23:20 +0800
- 进度：完成 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote` 核心实现：新增 TV 遥控键路由纯函数、左上信息层组件与数据构造；`LongFormVideoPlayer` 的 TV 模式改为首次亮 UI 但焦点停根、DOWN 进入 controls、LEFT/RIGHT seek 并重置计时、controls 内左右走 focusProperties 首尾环绕、Slider 不可聚焦、BACK 可见时优先收 UI；电视剧播放器传入剧名/季集/单集标题，左上主行显示剧名。同步 TV 版本号升到 `0.1.66`，`CONTEXT.md` 追加 TV 操作 UI 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting,TvLongFormTitleOverlay}.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、TV 单测与 androidTest、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-24-tv-long-form-operation-ui-on-remote/DONE.md`
- 验证：新增测试红灯曾因 `TvRemoteKeyAction` / `buildTvLongFormTitleOverlayData` 未实现失败；实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlayDataTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:connectedDebugAndroidTest` 已编译并打包 androidTest，但本机无连接设备，失败于 `DeviceException: No connected devices!`；`git diff --check` 通过；乱码扫描无输出。无关工作区改动 `.env.example`、`internal/config/config.go` 未纳入本任务。

## 2026-05-24 22:12 +0800
- 进度：根据 `$grill-with-docs` 决议，取消 VideoUpload 三段式上传向导，回到保留新设计系统外壳的单屏上传表单；文件选择、基础信息、关联信息、上传控制、进度与结果都在同一画面内。同步删除不再使用的 step 子组件与 wizard helper/spec，并将 `CONTEXT.md` 术语从 `admin 上传向导三步` 改为 `admin 上传单屏表单`。按用户要求，本轮不改 Phase 4 任务 PRD / implement / review 历史文档与截图，修改完成后补 `DONE.md`。
- 影响文件：`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`admin-web/src/views/videoUpload.wizard.helpers.*`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-three-pillars/DONE.md`
- 验证：待执行 `cd admin-web && npm run build`、`cd admin-web && npm test`、`git diff --check`、乱码扫描。

## 2026-05-24 22:00 +0800
- 进度：完成 `tasks/2026-05-23-admin-three-pillars/feedback.md` follow-up 修复：VideoList required 列迁移、批量删除 loading、drawer destroy-on-close、异步 token 防旧请求回写、字幕列表纳入 dirty；ImageManage 批量启停/删除改为 allSettled 并总是刷新/清选、默认 active chip 语义修正、视图切换清空选中；VideoUpload 切 movie 前确认丢弃多余文件，StepRelate 移除重复「开始上传」。同步扩展 BulkActionBar action 的 loading/disabled 支持，并在 CONTEXT 沉淀 drawer snapshot 与批量操作上下文切换契约。
- 影响文件：`admin-web/src/components/base/BulkActionBar.vue`、`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/StepRelate.vue`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-three-pillars/feedback.md`
- 验证：`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/views/videoUpload.wizard.helpers.spec.js` 通过；`cd admin-web && npm test` 通过（13 files / 95 tests）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-three-pillars` 无输出。

## 2026-05-24 21:30 +0800
- 进度：完成 `tasks/2026-05-23-admin-three-pillars` Phase 4 实现与 review 自动化：ImageManage 接入 PageHeader / Toolbar / 视图切换 / 双 drawer / BulkActionBar / EmptyState 并清零旧玫红；VideoList 接入 chip 筛选、更多筛选 drawer、列设置 localStorage、编辑 drawer、BulkActionBar；VideoUpload 拆为三步 Wizard 与 3 个 step 子组件；3 条路由补 `hideShellPageHeader`，CONTEXT 新增 5 条 admin 术语，截图目录归档 11 张 1440x1080 PNG。未创建 `DONE.md`，待用户验收后按流程补完成标记。
- 影响文件：`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`admin-web/src/views/videoUpload.wizard.helpers.*`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-three-pillars/*`、`plan.md`
- 验证：`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/views/videoUpload.wizard.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`cd admin-web && npm test` 通过（13 files / 95 tests）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-three-pillars` 无输出；截图目录 11 张 PNG 均为 1440x1080。

## 2026-05-24 14:30 +0800
- 进度：完成 `tasks/2026-05-23-admin-medium-views` Phase 3 的视觉归档收尾，4 个中等视图已补齐新的 after 截图，其中 `ImageCollectionManage` 额外补了 drawer 打开形态。当前截图目录共 9 张，均为 1440x1080。
- 影响文件：`admin-web/src/views/ImageCollectionManage.vue`、`admin-web/src/views/ScrapePreview.vue`、`admin-web/src/views/AVManualScrape.vue`、`admin-web/src/views/TvSeriesManage.vue`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-medium-views/screenshots/*`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-medium-views` 未发现乱码。

## 2026-05-24 11:55 +0800
- 进度：根据 `tasks/2026-05-23-admin-simple-views/feedback.md` 补修 Phase 2 反馈项：TaskMonitor 状态筛选与后端契约对齐、UserManage 改为后台原子建用户接口、Dashboard ECharts 实例重挂载修复，并收口 shell 顶栏与视图页头的双标题问题。
- 影响文件：`internal/handlers/admin.go`、`internal/repository/admin_repository.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/UserManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/Dashboard.vue`、`admin-web/src/components/Layout.vue`、`admin-web/src/router/index.js`、`plan.md`
- 验证：待执行后端/前端定向测试、`npm test`、`npm run build`、乱码扫描与 diff 检查。

## 2026-05-24 11:46 +0800
- 进度：完成 `tasks/2026-05-23-admin-simple-views` Phase 2 的实现与截图归档，7 个简单视图已按 Phase 1 设计系统重排，`themeTokens.spec.js` 扩展的视图层 audit 通过，14 张 before/after 截图已写入任务目录。当前先保留任务文档本身与实现代码、截图及计划记录，待用户确认验收后再按仓库流程补 `DONE.md`。
- 影响文件：`admin-web/src/views/Dashboard.vue`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/UserManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/IPTVManage.vue`、`admin-web/src/views/CollectionManage.vue`、`admin-web/src/views/ActorManage.vue`、`admin-web/src/components/UploadProgress.vue`、`admin-web/src/api/admin.js`、`admin-web/src/assets/themeTokens.spec.js`、`tasks/2026-05-23-admin-simple-views/screenshots/*`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src admin-web/src/assets/themeTokens.spec.js tasks/2026-05-23-admin-simple-views/screenshots plan.md` 未发现乱码；`tasks/2026-05-23-admin-simple-views/screenshots/` 已包含 14 张 PNG，统一 1440x1080。

## 2026-05-24 10:31 +0800
- 进度：完成 `tasks/2026-05-23-admin-shell-redesign` Phase 1 的实现、定向测试、全量 `npm test`、`npm run build`、截图归档与乱码扫描。admin-web 已切换到新的浅色设计 token、Element Plus 覆写、分组侧栏、命令面板、profile chip、独立登录页与共享基础组件；同时把 `Dashboard` / `IPTVManage` / `TaskMonitor` 的 `--font-code` 收口到 `--font-mono`，并在 `CONTEXT.md` 追加「admin 设计系统术语」。
- 影响文件：`admin-web/src/assets/theme.css`、`admin-web/src/assets/element-overrides.css`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/base/*`、`admin-web/src/views/Login.vue`、`admin-web/src/views/Dashboard.vue`、`admin-web/src/views/IPTVManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/main.js`、`admin-web/index.html`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-shell-redesign/screenshots/*`
- 验证：`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/components/Layout.spec.js src/components/base/commandPalette.helpers.spec.js` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src admin-web/index.html tasks/2026-05-23-admin-shell-redesign` 未发现乱码；截图已归档为 `before-*.png` / `after-*.png`。

## 2026-05-24 09:53 +0800
- 进度：完成 TV 首页货架文案收口：三类内容页的货架标题下不再显示“最近更新”，各区块的“查看更多”卡也不再展示数量副文案；首页货架语义收束为「最近播放 / 最近更新」两类，`TvCatalogScreen` 仅保留纯标题呈现。TV 端版本递增到 `versionCode = 66`、`versionName = "0.1.65"`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest.homeShelvesDoNotShowSubtitleCopyUnderTheTitle'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 未发现乱码。未纳入未跟踪 `tasks/2026-05-23-admin-*` 目录。

## 2026-05-24 09:29 +0800
- 进度：完成 TV 设置页「电视剧自动连播」抗挤压修复：设置行改为最小高度，左侧文案区弹性收缩并单行省略，右侧 Switch 固定 64dp 操作占位；补充静态回归测试锁定该布局契约。TV 端版本递增到 `versionCode = 65`、`versionName = "0.1.64"`，并在 `CONTEXT.md` 沉淀 `TV 设置行抗挤压布局`。未纳入未跟踪 `tasks/2026-05-23-admin-*` 目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest.seriesAutoplaySettingRowProtectsSwitchFromTextCompression'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 未发现乱码。

## 2026-05-23 23:45 +0800
- 进度：完成 ASS/SSA 外挂字幕上传支持：后端上传计划新增 `.ass/.ssa`，上传后先落临时源文件再通过 ffmpeg 转成 WebVTT，最终轨道仍以 `vtt` / `text/vtt` 暴露给手机端和 TV 端；metadata 记录 `original_filename` 与 `original_format`。管理端字幕上传选择器改用 `subtitleUploadAccept`，允许 `.srt,.vtt,.ass,.ssa`。已在 `CONTEXT.md` 沉淀“外挂 ASS/SSA 字幕上传策略”。未纳入未跟踪的 `tasks/2026-05-23-admin-*` 目录。
- 影响文件：`internal/services/subtitle.go`、`internal/services/transcode_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `go test ./pkg/ffmpeg -run TestBuildConvertSubtitleToWebVTTArgs -count=1` 先失败于 `undefined: buildConvertSubtitleToWebVTTArgs`；实现后 `go test ./pkg/ffmpeg -run 'TestBuildConvertSubtitleToWebVTTArgs|TestParseSubtitleProbeOutput' -count=1` 通过，`go test ./internal/services -run 'TestSubtitleUploadPlanForFilename' -count=1` 通过，`cd admin-web && npm test -- videoList.helpers.spec.js` 通过；收口验证 `go test ./pkg/ffmpeg ./internal/services ./internal/handlers ./internal/repository -run 'Subtitle|VideoSubtitle|BuildConvertSubtitle|ParseSubtitle|Transcode' -count=1` 通过，`cd admin-web && npm test` 通过，`cd admin-web && npm run build` 通过（仅既有 chunk size warning），`go test ./... -count=1` 通过，`go vet ./...` 通过；待执行 diff/乱码检查。

## 2026-05-23 22:45 +0800
- 进度：根据用户“完成当前任务”的确认，将 `tasks/2026-05-23-tv-resume-from-history-prompt/` 标记为已完成。新增 `DONE.md` 记录完成日期、最终关联提交 `f8a8652c` 与验证摘要；本轮仅做任务归档完成标记，不改 TV 运行时代码。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/DONE.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' plan.md tasks/2026-05-23-tv-resume-from-history-prompt/DONE.md` 无输出；`git diff --check -- plan.md tasks/2026-05-23-tv-resume-from-history-prompt/DONE.md` 通过；文档完成标记不需要重新构建。

## 2026-05-23 22:30 +0800
- 进度：完成 `tasks/2026-05-23-tv-resume-from-history-prompt/` 的 code-review 反馈优化。重点修复 6 项：
  1. `LongFormVideoPlayer` 新增 `onTrackSheetVisibilityChanged: (Boolean) -> Unit` 单一回调（合并 subtitle/audio sheet 可见性），父屏维护 `isTrackSheetVisible` state，纳入续播卡守卫 + 永久 dismiss LaunchedEffect——修复 H16（字幕/音轨夜台玻璃面板无信号给父屏导致续播卡可与之同屏的真实 bug）。
  2. PRD Q7 / 实现 / CONTEXT.md 关于 BACK 退出确认的三方语义统一到「永久 dismiss」：两个 player 的永久 dismiss LaunchedEffect 新增 `showBackConfirmPrompt` key；同步改 PRD H15、CONTEXT.md「续播提示卡永久 dismiss」条目。
  3. 倒计时驱动从 `delay(50) × N` 改为 `withFrameNanos` 推导剩余时间，只在显示秒数变化时写 state——消除主循环抖动累积与 ~100 次冗余 recompose。
  4. `TvSeriesPlayerScreen` 内 `shouldShowPromptCard` 重命名为 `shouldShowAutoplayPromptCard`，与同包纯函数同名、不再与 `shouldShowResumePromptCard` 形似。
  5. `TvResumePromptCard.kt` 的 `LaunchedTvInitialFocus` 去掉多余的 `lastPositionMs` key。
  6. 版本号 `versionCode 63 → 64`、`versionName 0.1.62 → 0.1.63`；测试断言（`TvResumePromptTest` + `TvResumePromptCardSpecTest`）同步扩 `isTrackSheetVisible` 真值表 / `withFrameNanos` audit / `showBackConfirmPrompt` audit / 新 `LaunchedTvInitialFocus(visible)` 单 key 形态。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/{TvResumePrompt.kt,TvResumePromptCard.kt,TvLongFormPlayerScreen.kt,TvSeriesPlayerScreen.kt}`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/{TvResumePromptTest.kt,TvResumePromptCardSpecTest.kt}`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew :tv-app:assembleDebug` 通过；H15 / H16 / H16b / H17 真实 TV 手测仍待用户在设备执行。

## 2026-05-23 22:04 +0800
- 进度：完成 `tasks/2026-05-23-tv-resume-from-history-prompt/` 实现阶段自动化 review。新增任务文档、续播提示实现、测试、TV 版本号与进度记录将纳入本次提交；未创建 `DONE.md`，因为 H1-H18 真实 TV 手测需用户验收通过后再标记完成。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md tasks/2026-05-23-tv-resume-from-history-prompt android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无输出；H1-H18 手测待用户在 TV 设备执行。

## 2026-05-23 22:00 +0800
- 进度：完成 TV 续播提示卡核心实现：新增续播提示纯函数与 UI 卡片，接入电影 / `18+` 与电视剧播放器；历史 seek 且达到 10 秒门槛时显示 5 秒左下角提示，暂停和退出确认冻结倒计时，错误 / 选集 / 连播互斥时永久关闭；补 `TvResumePromptTest` 与 `TvResumePromptCardSpecTest` 并完成红绿验证。`CONTEXT.md` 的 6 条续播术语已在前置审查阶段追加，本轮不重复写入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesAutoplay.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvResumePromptTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvResumePromptCardSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：定向红灯为 `TvResumePrompt*` API 未实现导致编译失败；实现后 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvResumePromptTest' --tests 'com.chee.videos.feature.tv.TvResumePromptCardSpecTest'` 通过；待执行全量 TV 单测、Debug 构建与乱码扫描。

## 2026-05-23 20:57 +0800
- 进度：根据用户确认，将 `tasks/2026-05-23-western-av-oshash-confirm-gate/` 标记为已完成。新增 `DONE.md` 记录完成时间、关联提交与验证摘要；后续批量处理 tasks 时默认跳过该任务目录，除非用户明确要求重开或复查。
- 影响文件：`tasks/2026-05-23-western-av-oshash-confirm-gate/DONE.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本次仅为任务状态文档标记，不需要重新构建。

## 2026-05-23 19:28 +0800
- 进度：完成 `tasks/2026-05-23-western-av-oshash-confirm-gate/feedback.md` 的必修收口：ThePornDB 两位年份日期补 `20`、`/scenes` 恢复 `parse=` 且 `/movies` 保持 `q=`、补 oshash 256KiB `0x42` 黄金值测试；同时补 ConfirmAV 前置状态注释、OUMEI_NAME 来源注释、migration down 回滚提示，并将“含 DONE 的 task 不按 feedback 返工”沉淀到 `CONTEXT.md`。未纳入 `.codex/skills/av-scraper-optimization` 删除、`.claude/`、OpenSpec skill、`CLAUDE.md`、`package-lock.json`、两个 task `feedback.md` 等既有无关工作区变更。
- 影响文件：`CONTEXT.md`、`internal/handlers/admin_scrape.go`、`internal/queue/scrape_tasks_test.go`、`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_theporndb_test.go`、`internal/services/scraper_test.go`、`migrations/0021_western_av_oshash_gate.down.sql`、`pkg/oshash/oshash_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestThePornDBSearchKeywordsExpandsSupportedDateFormats|TestThePornDBSearchURLUsesEndpointSpecificQueryParameter' -count=1` 通过；`go test ./pkg/oshash -run 'TestComputeMatchesPythonOshashGoldenFixture|TestComputeReturnsDeterministicHex|TestComputeReturnsErrFileTooSmall' -count=1` 通过；`go test ./pkg/oshash ./internal/services ./internal/queue ./internal/handlers -run 'ThePornDB|OSHash|AVScrape|ConfirmAV|SkipScrape' -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal pkg migrations` 无输出。

## 2026-05-23 16:50 +0800
- 进度：已提交 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的实现收口，关联提交 `ac7766f2`（`完成欧美 AV 刮削确认门控`）。按任务 DONE 标准，本轮不创建 `DONE.md`；需用户完成 B/C 手动验收后再标记完成。
- 影响文件：`plan.md`
- 验证：提交前 `go test ./...`、`go vet ./...`、`cd admin-web && npm test`、`cd admin-web && npm run build` 均通过。

## 2026-05-23 16:49 +0800
- 进度：完成 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的实现收口：欧美 AV 上传自动刮削落 `av_scrape_pending` 并写入 `scrape_preview` / `scrape_attempt`，确认或弃刮后通过 `force=true` 转码入队；ThePornDB 成功响应改为完整 JSON decode，修复 detail body 被 512B 截断导致候选丢失；admin-web 增加 `欧美 AV 待确认` 状态、待确认面板、弃刮入口和 `hash 命中` 徽章，AV 手动刮削能直接加载待确认候选。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_av_framework.go`、`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_strategy.go`、`internal/queue/scrape_tasks.go`、`internal/queue/tasks.go`、`internal/handlers/admin_scrape.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/AVManualScrape.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/oshash ./internal/repository ./internal/services ./internal/queue ./internal/handlers -count=1` 通过；`go test ./...` 通过；`go vet ./...` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size warning）。

## 2026-05-23 15:35 +0800
- 进度：按 task 审查结果收紧 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的第三个口径：删除固定“5 秒内”的验收写法，改成自动刮削任务完成后的状态型验收；同步修正 `prd.md` / `review.md`。
- 影响文件：`tasks/2026-05-23-western-av-oshash-confirm-gate/prd.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/review.md`、`plan.md`
- 验证：文档改动，三处口径已收紧

## 2026-05-23 14:04 +0800
- 进度：按用户确认补写 `tasks/2026-05-23-tv-series-autoplay-next-episode/DONE.md` 完成标记，任务进入已完成状态。
- 影响文件：`tasks/2026-05-23-tv-series-autoplay-next-episode/DONE.md`、`plan.md`
- 验证：完成标记已写入，待提交。

## 2026-05-23 13:33 +0800
- 进度：根据实现后反馈修正 TV 电视剧自动连播的竞态与语义对齐：补 `shouldHandlePlaybackEnded` 纯函数、给连播提示卡守卫增加结尾覆盖层字段、自动切后提前封住历史上报回流，并把暂停态 / review 验收 / CONTEXT 定义同步调整为“暂停时卡隐藏、恢复后接续”。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesAutoplay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesAutoplaySpecTest.kt`、`CONTEXT.md`、`tasks/2026-05-23-tv-series-autoplay-next-episode/review.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesAutoplaySpecTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.tv.TvAutoplayPromptCardSpecTest' --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest'`；再跑 `:tv-app:assembleDebug` / `:tv-app:assembleRelease`

## 2026-05-23 12:04 +0800
- 进度：完成 TV 电视剧自动连播实现：连播链路跨季/跳过不可播放集、提示卡、结尾覆盖层、自动切完成上报、手动下一集分流、设置页开关、DataStore 持久化、SkipNext 图标与 TV 版本号更新均已落地。未纳入 `.codex/skills/av-scraper-optimization` 删除、`.claude/`、OpenSpec skill 目录、`CLAUDE.md`、`package-lock.json` 等既有无关工作区变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/build.gradle.kts`、`tasks/2026-05-23-tv-series-autoplay-next-episode/`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleRelease` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md tasks/2026-05-23-tv-series-autoplay-next-episode android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无输出。

## 2026-05-23 11:04 +0800
- 进度：按审查结论同步修正 `tasks/2026-05-23-tv-series-autoplay-next-episode/` 的 PRD / Implement / Review：补入 `手动下一集按钮` 术语，明确提示卡要避开底部控制条、倒计时按播放器位置推导且显示整秒，并把手动下一集与自动连播放行。后续实现时可直接以这版任务文档为准。
- 影响文件：`tasks/2026-05-23-tv-series-autoplay-next-episode/prd.md`、`tasks/2026-05-23-tv-series-autoplay-next-episode/implement.md`、`tasks/2026-05-23-tv-series-autoplay-next-episode/review.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务文档修订，不涉及构建。

## 2026-05-23 10:22 +0800
- 进度：按 `grill-with-docs` 审查 `tasks/2026-05-23-tv-series-autoplay-next-episode/` 时，确认 `连播链路` 不应依赖后端返回列表顺序，而应以季号和集号升序定义“下一集”；列表顺序仅作为编号重复或缺失时的稳定兜底。已将该术语边界补入 `CONTEXT.md`，后续实现 `resolveNextPlayableEpisode` 时必须按该语义写单测。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 02:47 +0800
- 进度：根据用户最终验收，将 `tasks/2026-05-23-short-overlay-fullscreen-button/` 标记为已完成。新增 `DONE.md` 记录完成时间、用户确认、关联提交和验证摘要；后续用户要求“完成 tasks 里的任务”时默认跳过该目录，除非明确要求重开或复查。
- 影响文件：`tasks/2026-05-23-short-overlay-fullscreen-button/DONE.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本次仅为任务状态文档标记，不需要重新构建 App。

## 2026-05-23 02:23 +0800
- 进度：完成搜索页短视频全屏底栏残留修复。红灯测试先失败于 `search short fullscreen state must hide the app shell bottom bar`；实现后，`ShortSearchScreen` 暴露 `onFullscreenChange`，`ShortSearchPlayerOverlay` 在 `isFullscreen` 变化时通知根壳、销毁时恢复 `false`，`VideoHomeApp` 的 `search` tab 将回调写入 `isShortFullscreen`，从而隐藏根底部 tabbar。搜索浮层全屏分支同时移除 `statusBarsPadding()`，非全屏分支保留原顶部安全区。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest` 先失败于新增搜索页回传断言；实现后同一命令通过。`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；并行跑 `:app:testDebugUnitTest` 曾因 Hilt 注解处理输出竞争失败于 `MainActivity_GeneratedInjector`，串行重跑 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-app/app/src/main/java android-app/app/src/test/java android-app/app/build.gradle.kts` 无输出。

## 2026-05-23 02:08 +0800
- 进度：同步修复除首页外的另外三处短视频全屏入口。用户确认首页可用后指出搜索、发现、UnifiedPlayer 短视频分支仍未改；新增结构性红灯测试覆盖 `ShortSearchScreen`、`ShortDiscoverScreen`、`UnifiedPlayerScreen` 必须全屏/竖屏二选一渲染。实现后，三处都在 `isFullscreen` / `isShortFullscreen` 为 true 时只渲染 `ShortOverlayFullscreenHost`，非全屏时才渲染竖屏 `VerticalPager`、操作栏、关闭按钮和短视频进度条，避免两个 `PlayerView` 同时绑定同一 `ExoPlayer`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest` 先失败于 `all non home short overlays hide vertical pager while fullscreen`；实现后同一命令通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 初次失败于既有 TV 测试 `TvCatalogViewModelTest.nullListsInPayload_doNotCrashAndFallbackToEmpty` 的测试前协程异常，单独重跑该用例通过，随后全量 `:app:testDebugUnitTest` 复跑通过。

## 2026-05-23 01:23 +0800
- 进度：沉淀 `tasks/` 完成标记约定。以后批量执行 `tasks/` 时，已包含 `DONE.md` 的任务目录默认视为完成并跳过；用户明确要求重开或复查时才重新处理。按用户确认测试完成的语义，为 `tasks/2026-05-23-short-overlay-fullscreen-button/` 新增 `DONE.md`，记录完成时间、关联提交和验证摘要。
- 影响文件：`AGENTS.md`、`CONTEXT.md`、`tasks/2026-05-23-short-overlay-fullscreen-button/DONE.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；文档规则变更无需构建。

## 2026-05-23 01:13 +0800
- 进度：完成手机端短视频浮层“全屏播放”任务收尾。共享 `ShortOverlayFullscreenHost` 已接入搜索、发现、主页短视频信息流和 `UnifiedPlayerScreen` 的短视频分支；`CONTEXT.md` 已补“短视频全屏播放”术语；`android-app/app/build.gradle.kts` 已按约定递增版本号。验证方面，手机端 `:app:testDebugUnitTest` 与 `:app:assembleDebug` 通过，TV 工程 `:tv-app:testDebugUnitTest` 也保持通过。ADB 已重新连接模拟器 `emulator-5554`，完成 `com.chee.videos` 安装与 `MainActivity` 启动确认，logcat 未见新的 `AndroidRuntime` 或 FATAL；由于当前设备侧不具备完整手测输入条件，本次仅记录到启动级现场校验。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-app/app/src/main/java android-app/app/src/test/java` 无输出；ADB `install -r` 与 `am start -n com.chee.videos/.MainActivity` 成功。

## 2026-05-23 00:54 +0800
- 进度：沉淀 `tasks/` 任务执行顺序约定。确认当前 `tasks/2026-05-23-short-overlay-fullscreen-button/` 目录包含 `prd.md`、`implement.md`、`review.md` 三段文档；按用户要求，将后续“完成 tasks 里的任务”固定解释为先读 PRD、再按 Implement 实施、最后按 Review 验收。根级 `AGENTS.md` 增加代理执行规则，`CONTEXT.md` 增加长期技术/流程沉淀。
- 影响文件：`AGENTS.md`、`CONTEXT.md`、`plan.md`。
- 验证：待执行 Markdown 乱码扫描与 diff 检查；文档规则变更无需构建。

## 2026-05-22 23:45 +0800
- 进度：完成 TV 首页 Release R8 模型保留修复并收尾。R8 复核显示 `seeds.txt` 已包含 `TvHomePayload`、`TvHomeVideoDto`、`TvSectionDto`、`TvCatalogWallPayload`、`TvCatalogWallItemDto`、`TvContinueWatchingDto`、`TvSeriesSummaryDto` 等首页/海报墙模型；`mapping.txt` 显示 `TvHomePayload -> com.chee.videos.core.model.TvHomePayload`、`TvHomeVideoDto -> com.chee.videos.core.model.TvHomeVideoDto`、`TvSectionDto -> com.chee.videos.core.model.TvSectionDto`，类名和关键 getter/构造函数保留；`usage.txt` 中这些类只剩 `static <clinit>` 优化条目，不再裁剪字段/getter。Release 输出版本为 `0.1.60` / `versionCode=61`，生成 `tv-app-armeabi-v7a-release-unsigned.apk` 与 `tv-app-arm64-v8a-release-unsigned.apk`。本次仅暂存并提交 5 个任务文件，无关 `.codex/skills/av-scraper-optimization` 删除和未跟踪文件不纳入。
- 影响文件：`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.network.TvAuthEnvelopeSpecTest` 红灯失败后转绿；`./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`./gradlew --no-daemon :tv-app:assembleRelease` 通过；`rg -n $'\uFFFD' ...` 无输出；`git diff --check -- ...` 通过。

## 2026-05-22 23:41 +0800
- 进度：完成红灯与核心修复。新增 `TvAuthEnvelopeSpecTest.release shrinker keeps all gson api models used through retrofit envelopes`，红灯阶段 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.network.TvAuthEnvelopeSpecTest` 失败于缺少 `-keep class com.chee.videos.core.model.** { *; }`。实现阶段将 `proguard-rules.pro` 从仅保留 `TvAuth*` 扩展为保留 `core.model.**` 全部 Retrofit/Gson API 模型，并保留 `TvAuth*` 显式规则作为既有线上崩溃提示；TV 版本 `0.1.59` → `0.1.60`，`versionCode` 60→61；`CONTEXT.md` 新增 TV Release API 模型保留规则。
- 影响文件：`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：红灯阶段定向测试失败于新增审计；实现后同一命令通过。待执行 `./gradlew --no-daemon :tv-app:testDebugUnitTest`、`./gradlew --no-daemon :tv-app:assembleRelease` 与 R8 产物复核。

## 2026-05-23 00:00 +0800
- 进度：修复配对码生成 `ClassCastException`（版本 0.1.57 → 0.1.58，versionCode 58→59）。根因：Gson 泛型类型擦除问题——`ApiEnvelope<T>` 的 `data: T?` 字段在部分 Android TV 固件（老版 ART）上无法正确将类型参数 `T` 解析为具体类型 `TvAuthSessionCreatePayload` / `TvAuthSessionStatusPayload`，退化为 `LinkedTreeMap<String, Any>`；随后 Kotlin 编译器在 `requireEnvelope()` 返回值处插入的 CHECKCAST 字节码指令尝试将 `LinkedTreeMap` 强转为目标 payload 类，抛 `ClassCastException`，且部分老 ART 实现在这种情况下该异常的 `message` 为 null，触发上一次改动添加的"创建配对会话失败 (ExceptionClassName)"兜底串。修法：不改泛型 `ApiEnvelope<T>`（其他端点不受影响），仅针对两个实际消费 `data` 字段的 TV 认证端点，在 `ApiModels.kt` 新增两个具体非泛型包装类 `TvAuthCreateEnvelope`（`data: TvAuthSessionCreatePayload?`）与 `TvAuthStatusEnvelope`（`data: TvAuthSessionStatusPayload?`）；更新 `ApiService.kt` 的 `createTvAuthSession` / `getTvAuthSession` 返回类型由 `ApiEnvelope<TvAuthSessionCreatePayload>` / `ApiEnvelope<TvAuthSessionStatusPayload>` 改为对应的具体包装类；在 `TvAuthRepository.kt` 补充两个具体重载 `requireEnvelope(resp: TvAuthCreateEnvelope)` / `requireEnvelope(resp: TvAuthStatusEnvelope)`，Kotlin 在编译时静态选择正确重载，彻底消除运行时泛型推断。`callWithAuth`（approve / deny）保持 `ApiEnvelope<Map<String, Boolean>>` 不变，因该路径 `data` 结果被丢弃，不触发 CHECKCAST 问题。
- 影响文件：`core/model/ApiModels.kt`（新增 TvAuthCreateEnvelope / TvAuthStatusEnvelope）、`core/network/ApiService.kt`（两个 TV 认证端点返回类型）、`core/repository/TvAuthRepository.kt`（新增两个具体 requireEnvelope 重载 + 对应 import）、`build.gradle.kts`（版本号）、`plan.md`。
- 验证：待 `testDebugUnitTest` + `assembleDebug`。

## 2026-05-22 23:22 +0800
- 进度：确认并修复 TV 配对页 `ClassCastException` 的真实根因（版本 0.1.58 → 0.1.59，versionCode 59→60）。上次代码层把 TV 授权端点改成具体 envelope 后，debug 单测/源码字节码已正确，但用户安装的是 Release 形态 APK（设备拉回 `base.apk` 约 42MB，且本地 debug 包签名不匹配无法覆盖安装）。对比本地 Release R8 产物发现：未加规则时 `usage.txt` 将 `TvAuthCreateEnvelope` / `TvAuthStatusEnvelope` / `TvAuthSessionCreatePayload` / `TvAuthSessionStatusPayload` / `TvAuthSessionCreateRequest` 判定为可裁剪；这类模型只通过 Retrofit suspend 签名与 Gson 反射使用，Release R8 裁剪后设备运行时返回类型退化，最终仍触发 `ClassCastException`。修法：`proguard-rules.pro` 新增 `-keep class com.chee.videos.core.model.TvAuth* { *; }`，保留 TV 授权配对所有 envelope/payload/request 模型；`TvAuthEnvelopeSpecTest` 新增 R8 规则审计，防止回退；`CONTEXT.md` 更新“TV 配对会话响应包装”约定，明确 Release R8 keep 是这组模型契约的一部分。
- 影响文件：`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.network.TvAuthEnvelopeSpecTest` 失败于缺少 `-keep class com.chee.videos.core.model.TvAuth* { *; }`；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过，`./gradlew --no-daemon :tv-app:assembleRelease` 通过，且 `seeds.txt` 显示 `TvAuth*` 模型被 keep 规则选中、`mapping.txt` 显示类名和关键 getter/构造函数保留；`rg -n $'\uFFFD' ...` 无输出，未发现乱码替换字符。

## 2026-05-22 23:00 +0800
- 进度：修复"选择服务器后，生成配对码失败"（版本 0.1.56 → 0.1.57，versionCode 57→58）。根因分析：`ConnectionViewModel.useEndpoint`（用于已发现/历史地址列表点击）直接调 `activateEndpoint` 而不先探测服务器连通性，与 `manualConnect`（先 `testEndpoint` 再激活）不一致；若此时服务器实际不可达，App 仍会导航到配对页，随后 HTTP 请求抛出 `ConnectException` 等异常——在部分 Android TV 盒子固件上该异常的 `message` 为 null，导致兜底字符串"创建配对会话失败"展示给用户。次因：`TvAuthRepository.createSession` 用 `runCatching` 捕获了 `CancellationException`，协程被取消时 message=null 同样触发兜底串。修法：① `useEndpoint` 改为先调 `serverRepository.testEndpoint(baseUrl)` 探测，失败直接在连接页报错，成功再 `activateEndpoint` 导航；② `createSession` 将 `runCatching` 改为 `try/catch` 并显式 re-throw `CancellationException`；③ 兜底错误信息追加 `(ExceptionClassName)` 便于诊断。
- 影响文件：`ConnectionViewModel.kt`（useEndpoint 加连通探测）、`TvAuthRepository.kt`（createSession re-throw CancellationException）、`TvPairingScreen.kt`（兜底错误信息带类名）、`build.gradle.kts`（版本号）、`plan.md`。
- 验证：`testDebugUnitTest` BUILD SUCCESSFUL 23s 全绿。

## 2026-05-22 22:58 +0800
- 进度：完成 TV 配对 `ClassCastException` 修复验证。确认本次提交仅纳入 TV 认证 envelope 具体化、配套测试、版本号与长期文档；工作区中既有 `.codex/skills/*` 删除、`.claude/`、`CLAUDE.md`、`package-lock.json` 等无关变更不纳入。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/TvAuthRepository.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL；`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL；`rg -n $'\uFFFD' ...` 无输出，未发现乱码替换字符。

## 2026-05-22 22:56 +0800
- 进度：补齐 TV 配对 `ClassCastException` 修复的回归护栏与长期文档。新增 `TvAuthEnvelopeSpecTest` 锁定 `createTvAuthSession` / `getTvAuthSession` 必须返回 `TvAuthCreateEnvelope` / `TvAuthStatusEnvelope`，并用反射确认两个具体 envelope 的 `data` 字段分别是 `TvAuthSessionCreatePayload` / `TvAuthSessionStatusPayload`；`approve` / `deny` 继续允许 `ApiEnvelope<Map<String, Boolean>>`，因为调用方不消费 `data` payload。同步在 `CONTEXT.md` 记录“TV 配对会话响应包装”约定，明确不要把两个会消费配对 payload 的 TV 认证端点退回泛型 `ApiEnvelope<T>`。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`（新增）、`CONTEXT.md`、`plan.md`。
- 验证：待执行 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 与 `./gradlew --no-daemon :tv-app:assembleDebug`。

## 2026-05-22 22:30 +0800
- 进度：修复一级首页左侧菜单焦点无法跳回内容区 bug（版本 0.1.55 → 0.1.56，versionCode 56→57）。根因：`TvHomeSideMenuButton` 的 `.focusProperties { right = contentFocusRequester }` 把 D-pad RIGHT 硬指向 `featuredFocusRequester`，该 requester 绑定在 `LazyColumn` 内 `TvFeaturedHero` item 上；用户向下滚动后 hero item 被虚拟化移出组合树、requester 变为 uninitialized；此时从菜单按 RIGHT → ISE → `dispatchKeyEvent` ANR 兜底吞掉返回 `false` → 无焦点移动 → 表现为”无法从菜单跳回内容区，需要点击菜单按钮”。修法：删除 `focusProperties { right = contentFocusRequester }` 块及 `contentFocusRequester` 参数在 `TvHomeSideMenuButton` / `TvHomeSideMenu` / 两处调用点的级联，同步删除孤立 `import focusProperties`；改由 Compose 空间焦点遍历自动找右侧最近可聚焦节点，内容 `LazyColumn` 横铺剩余宽度、不受虚拟化影响，空间遍历总能命中当前可见内容项。
- 影响文件：`TvCatalogScreen.kt`（删除 `focusProperties` 块 + 参数 + import）、`build.gradle.kts`（版本号）、`plan.md`。
- 验证：`testDebugUnitTest` BUILD SUCCESSFUL 25s 全绿。

## 2026-05-22 19:50 +0800
- 进度：修复 TV App ANR——`Input dispatching timed out (Wait queue length: 1)`。ANR 时间戳 2026-05-22 19:35:55，设备 Sony BRAVIA（Android 9，API 28），系统负载 31.4（Douyu TV 13% + 音频后处理 24% 等多 App 并行）。根因分两层：(1) **系统 CPU 饥饿**（负载 31.4，主线程偶发被抢占 >5s）为可能主因；(2) **代码级根因**：遥控器 DPad 按键触发 Compose 同步 focus 遍历路径（`FocusOwnerImpl.focusSearch → AndroidComposeView$keyInputModifier$1`）抛出 `FocusRequester is not initialized` ISE 时，异常沿 `ViewRootImpl.deliverInputEvent → InputStage.deliver → Activity.dispatchKeyEvent` 同步透出；`ViewRootImpl.deliverInputEvent` 没有 try/finally，异常被主 Looper 兜底（`installMainLooperHoverExitGuard`）吞掉后 `finishInputEvent()` 永远不调用，输入分发器等不到 ACK → 5s 超时 → ANR。修法：在 `TvMainActivity` 新增与 `dispatchGenericMotionEvent` 对称的 `override fun dispatchKeyEvent(event: KeyEvent): Boolean` —— 在 Activity 边界捕获 `shouldSwallowTvComposeFocusRequesterCrash` 命中的 ISE，返回 `false`（未消费），让 `ViewRootImpl` 正常调 `finishInputEvent()` 发回 ACK；主 Looper 兜底作为**异步路径**的最后防线继续保留不动，三层防线整体不削减。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`（新增 `dispatchKeyEvent` override + import `android.view.KeyEvent`）、`android-tv-app/tv-app/build.gradle.kts`（版本 53→54 / 0.1.52→0.1.53）、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL（26s）、`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（12s）全绿。待手测：ANR 复现条件——快速 DPad 导航 TV 首页（首页还在加载 / FocusRequester 节点未挂载时立刻按方向键），装 0.1.53 后应不再出现"应用无响应"弹窗；`adb shell dumpsys package com.chee.videos.tv | grep versionName` 输出 `versionName=0.1.53`。

## 2026-05-22 16:05 +0800
- 进度：落地 B 批第五项 B5——TV 工程圆角语言统一收口到 16dp + 白名单豁免。把 `tv-app/src/main` 下散落在 ~15 个文件、近 100 处的 `RoundedCornerShape(N.dp)` 调用点全部收口到 `core/ui/AppChrome.kt` 暴露的三档 token：(1) **默认** `AppChrome.RadiusDp = 16.dp` / `AppChrome.SurfaceShape = RoundedCornerShape(16.dp)`——卡片 / 面板 / 按钮 / 输入框 / 海报 clip / 错误 banner / 演员卡 / 指标卡 / 操作按钮等通用容器全部走它；(2) **白名单 1** `AppChrome.ChipRadiusDp = 8.dp` / `AppChrome.ChipShape = RoundedCornerShape(8.dp)`——专给小型 chip / 步进按钮 / IPTV 频道行 / 30dp 图标盒等 28-48dp 高度元素，8dp 在小尺寸下视觉接近方角，与 16dp 形成「小元素少圆 / 大元素中圆」层级；(3) **白名单 2** `AppChrome.PillShape = RoundedCornerShape(999.dp)`——圆形/胶囊形几何必需场景（头像 / 状态徽标 / 圆形图标按钮 / 字幕选择条 3dp×28dp 选中竖条）。另两类**非对称**圆角不在本审计正则 `RoundedCornerShape\(\s*(\d+)\.dp\s*\)` 命中范围、各管各的：B4 沉浸式详情面板上沿 `TvDetailPanelTokens.TopCornerRadiusDp = 28.dp`（`topStart = topEnd = 28.dp`）、IPTV 频道列表面板左沿（`topStart = bottomStart = 18.dp`），都是带方向性的视觉锚点不强行收口。改动按七块：(1) `core/ui/AppChrome.kt`：删除 `val CardShape = RoundedCornerShape(22.dp)` 与 `val SectionShape = RoundedCornerShape(18.dp)` 两个旧 token，新增 `RadiusDp` / `SurfaceShape` / `ChipRadiusDp` / `ChipShape`，保留 `PillShape`；(2) 全工程清扫：`feature/tv/TvCatalogScreen.kt`（22 处裸字面量，覆盖搜索栏 / 步进面板 / 海报清 / 设置 chip / 导航 rail chip / 历史 chip / 全部入口 chip）、`feature/tv/TvLongFormPlayerScreen.kt` / `feature/tv/TvSeriesPlayerScreen.kt`（错误 banner + 选集行）、`feature/tv/TvSeriesDetailScreen.kt`（poster + 季选 chip + 剧集 cell）、`feature/tv/TvPosterWallScreen.kt`（私有 `TvPosterWallCardShape` 常量改成 `AppChrome.SurfaceShape` 别名）、`feature/tv/TvIptvScreen.kt`（top overlay + 频道行 + 30dp logo 盒）、`feature/tv/TvPlayerBackConfirm.kt`、`tv/TvShellApp.kt`、`feature/detail/DetailScreen.kt`、`core/ui/SubtitlePicker.kt`（panel scrim + 3dp×28dp 选中竖条 PillShape）、`core/ui/TvFocus.kt`（`tvFocusableGlow` / `tvFocusableScaleOnly` 默认 `shape` 参数从裸 `RoundedCornerShape(20.dp)` 改成 `AppChrome.SurfaceShape`）、`core/ui/LongFormVideoPlayer.kt`；调用点全部移除 `import androidx.compose.foundation.shape.RoundedCornerShape`（仅本地常量必须时保留作为类型签名）；(3) 新增 `src/test/java/com/chee/videos/core/ui/TvShapeAuditTest.kt` 两条用例锁定不变量——`tv main source uses only whitelisted symmetric RoundedCornerShape radii` 扫 `src/main/java` 全部 `.kt` 文件（排除 `tvMainSourceExcludes` 内 phone-only 路径 + `AppChrome.kt` 自身），所有 `RoundedCornerShape(N.dp)` 的 N 必须 ∈ `{8, 16, 999}`，否则报「文件名:行号:半径」列表；`AppChrome exposes the unified shape token set` 校验 RadiusDp / SurfaceShape / ChipShape / PillShape 同时存在且 `val CardShape` / `val SectionShape` 旧 token 字符串残留为零；(4) **配套修复 TV 工程编译边界**：删除 `CardShape` / `SectionShape` 后暴露的真实坑点——`android-tv-app/tv-app/build.gradle.kts` 的 `kotlin { sourceSets { ... kotlin.exclude(...) } }` 单独并不阻断 `compileDebugKotlin` 与 `kaptGenerateStubsDebugKotlin`，phone-only 文件 `feature/auth` / `feature/home` / `feature/mine` / `feature/shorts` / `feature/imagecollections` 仍被 kapt 拉进 stub 生成、再被 Kotlin 编译，所以删 token 后报「`Unresolved reference: CardShape` ×21」假象不在排除列表里。修法是在 build.gradle.kts 同一份 `tvMainSourceExcludes` / `tvTestSourceExcludes` 单一来源之上再叠两层 task-level exclude：`tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile>().configureEach { exclude(...) }` 与 `tasks.withType<org.jetbrains.kotlin.gradle.internal.KaptGenerateStubsTask>().configureEach { exclude(...) }`，按任务名 `contains("UnitTest", ignoreCase = true)` 区分主/测试源集排除列表；三处排除（kotlin.sourceSets + KotlinCompile + KaptGenerateStubsTask）共用同一份 `tvMainSourceExcludes` / `tvTestSourceExcludes` 数据源，新增 phone-only 顶层路径自动级联；(5) `android-tv-app/tv-app/build.gradle.kts` 版本 `versionCode 52 → 53`、`versionName "0.1.51" → "0.1.52"`；(6) `CONTEXT.md` 两处更新——既有 `TV 工程编译边界` 词条整段扩写，写明三处 exclude 必须同步生效、单独依赖 kotlin.sourceSets 的坑点（B5 暴露）、必须按任务名区分主测源集列表；新增 `TV 圆角语言收口` 词条紧跟在 B4 `TV 沉浸式详情玻璃面板` 之后，写明三档 token 取值与场景边界、非对称圆角的另一层（B4 / IPTV 左沿）不波及、删除 CardShape/SectionShape 的强约束、`tvFocusableGlow` / `tvFocusableScaleOnly` 默认 shape 收口、`TvShapeAuditTest` 两条用例的判定逻辑；(7) `plan.md` 追加本条反向时间序条目。语义边界：B5 仅做对称圆角收口，**不**改 token 数值（仍是 16dp / 8dp / 999dp），**不**改任何调用点的非半径参数（如 shadow / border / padding），**不**改电话端（已物理隔离 + 现在通过 task-level exclude 实际阻断），**不**改 TV 工程 phone-only 文件（暂留作迁移参考且现在真正不进编译图），**不**动 hover-exit / FocusRequester 三层防线、不动 B1/B2/B3/B4 token、不动 NavHost transition。回归测试 `TvShapeAuditTest` 全绿；既有 `TvDetailPanelTokensTest` / `TvLongFormDetailGlassPanelSpecTest` / `TvHeroMotionTokensTest` / `TvFeaturedHeroMotionSpecTest` / `TvMotionTokensTest` / `TvFocusSpecTest` / `TvTypographySpecTest` / `TvColorContrastTest` / `TvListMotionSpecTest` / `TvBringIntoViewSpecTest` / `TvSharedPosterTransitionSpecTest` / `TvInitialFocusSafeRequestTest` / `TvNoBareLaunchedEffectFocusRequestAuditTest` / `TvInitialFocusEffectShapeTest` / `TvInitialFocusRequesterMatcherTest` / `TvMainActivityInputPolicyTest` 等历史不变量持续绿。B 批 5 项至此全部落地，下一步进入 C 批或按需推进。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvShapeAuditTest.kt`（新增）、`android-tv-app/tv-app/build.gradle.kts`（版本号 + task-level exclude）、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（20s），所有 TV unit test 持续绿，重点关注新增 `TvShapeAuditTest` 两条用例（源文 audit + token 暴露/旧 token 删除断言）；既有不变量同步绿。`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL（12s），输出 `tv-app-arm64-v8a-debug.apk`（70MB）与 `tv-app-armeabi-v7a-debug.apk`（67MB）两个 ABI 分包。中途发现并修复 TV 工程编译边界真实坑点：B5 删除 `CardShape`/`SectionShape` 后 phone-only 文件（feature/auth / feature/home / feature/mine / feature/shorts / feature/imagecollections）冒出 21 处 `Unresolved reference` 报红——确认 `kotlin.sourceSets.main.kotlin.exclude(...)` 单独不阻断 kapt 与 KotlinCompile 任务，已追加 `tasks.withType<KotlinCompile>().configureEach { exclude(...) }` 与 `tasks.withType<KaptGenerateStubsTask>().configureEach { exclude(...) }` 两层 task-level 排除，按任务名区分主/测试源集列表，复用 `tvMainSourceExcludes` / `tvTestSourceExcludes` 单一来源。待手测：用户在 4K TV `adb install -r` 安装 0.1.52 APK 后——(a) **海报墙**电影 / `18+` / 电视剧 三个 tab 海报卡圆角统一从旧值 14dp / 16dp / 18dp 抹平到 16dp，视觉一致；(b) **沉浸式详情**电影 / `18+` 长视频详情屏底部面板 28dp 上沿不变（B4 仍生效），面板内按钮 / 演员卡 / 指标卡 / 操作按钮圆角统一 16dp；电视剧详情屏 poster 26dp → 16dp、季选 chip 12dp → 8dp、剧集 cell 12dp → 16dp；(c) **左轨 chip** TvCatalogScreen 左侧导航 rail / settings 步进按钮 / 历史 chip / 全部入口 chip 仍是 8dp（不该变成 16dp）；(d) **IPTV** 顶部 overlay 12dp → 16dp、频道行 8dp 仍 8dp（白名单 ChipShape）、30dp 图标盒 8dp 仍 8dp；(e) **字幕选择器**面板 22dp → 16dp、3dp×28dp 选中竖条 2dp → 999dp（胶囊形，竖条几何上变窄到接近圆角矩形端部，更贴胶囊形视觉）；(f) 焦点视觉（双层 glow）、shared-element 进入/退出、ken-burns / 列表 stagger / 详情玻璃面板 / 按下反馈 / B 批所有动效全部正常无回归；(g) `adb shell dumpsys package com.chee.videos.tv | grep versionName` 应输出 `versionName=0.1.52`。

## 2026-05-22 14:33 +0800
- 进度：落地 B 批第四项 B4——TV `TvLongFormDetailScreen`（电影 / `18+` 沉浸式详情首屏）底部信息面板从「硬切色块 `Surface(color=0xD20B1018, shape=RoundedCornerShape(top=28dp))`」升级为「玻璃面板（frosted glass）」：上沿 24dp 高纵向渐变 scrim（`Brush.verticalGradient(listOf(Color.Transparent, scrimColor))`，alpha 0 → 1，让面板"渗"进背景而非硬切），面板主体走 API 分支——API ≥ 31 时挂 `Modifier.blur(20.dp)` + scrim `Color(0xCC0A0E16)`（≈80% 不透明），API < 31 时不模糊但 scrim 加深到 `Color(0xE60A0E16)`（≈90% 不透明）保证文字可读性不掉档。改动按四块：(1) 新建 `core/ui/TvDetailPanel.kt` 暴露 `object TvDetailPanelTokens`，7 个 token：`BlurRadiusDp = 20.dp`（10-foot 视距下「玻璃感」甜点，区间 12–32dp）、`ScrimColorBlurred = Color(0xCC0A0E16)`、`ScrimColorFallback = Color(0xE60A0E16)`（fallback alpha 严格大于 blurred alpha）、`UpperGradientHeightDp = 24.dp`（区间 16–40dp）、`ContentPaddingHorizontalDp = 36.dp` / `ContentPaddingVerticalDp = 28.dp`（与现状对齐避免回归）、`TopCornerRadiusDp = 28.dp`（≥16dp）；(2) 新建 `core/ui/TvDetailPanelBackground.kt` 暴露 `@Composable fun TvDetailGlassPanel(modifier, content)`——内部用 `Build.VERSION.SDK_INT >= Build.VERSION_CODES.S` 做 API gating，`val supportsBlur` 决定 `scrimColor` 取 `ScrimColorBlurred` 或 `ScrimColorFallback`、`panelBaseModifier` 取 `Modifier.blur(BlurRadiusDp)` 或 `Modifier`，对外结构为 `Box(modifier.fillMaxWidth()) { 上沿渐变 scrim Box + Surface(scrim color, RoundedCornerShape top, .then(panelBaseModifier)) { content() } }`，blur 必须用条件 `.then(...)` 挂载——`Modifier.blur` 是 API 31+ API，老设备直接挂会 `NoSuchMethodError`，已有 API-gating 先例 `TvFocus.kt` 的 `VERSION_CODES.R` 判定；(3) 改 `feature/tv/TvLongFormDetailScreen.kt`：line 143-204 的 `Surface(color=Color(0xD20B1018), shape=RoundedCornerShape(topStart=28.dp,topEnd=28.dp))` 整块替换为 `TvDetailGlassPanel(modifier = Modifier.align(Alignment.BottomCenter))`，内部 `Column` 的 `.padding(horizontal=36.dp, vertical=28.dp)` 改为 `.padding(horizontal=TvDetailPanelTokens.ContentPaddingHorizontalDp, vertical=TvDetailPanelTokens.ContentPaddingVerticalDp)`，原有 `eyebrow / title / metaLine / summary / actors / actions` 文案与顺序不变；import 移除已不需要的 `androidx.compose.foundation.shape.RoundedCornerShape`，新增 `com.chee.videos.core.ui.TvDetailGlassPanel` 与 `com.chee.videos.core.ui.TvDetailPanelTokens`；(4) `android-tv-app/tv-app/build.gradle.kts` 版本 `versionCode 51 → 52`、`versionName "0.1.50" → "0.1.51"`。语义边界：B4 做的是「羽化边缘 + scrim 色块」，不是真「看穿玻璃」——`Modifier.blur` 模糊的是**面板自身**渲染产生边缘晕开，要做真「看穿玻璃」需要 `RenderEffect.createBlurEffect` + `graphicsLayer.renderEffect`，本期不动，将来升级仅在 `TvDetailGlassPanel` 内部替换实现、调用点零改动；作用域仅限 `TvLongFormDetailScreen`（电影 / `18+`），`TvSeriesDetailScreen` 视觉是另一套（poster + episodes 网格）不套用；整屏 backdrop → 面板之间的纵向渐变（`Brush.verticalGradient(0x6610151F → 0x3310151F → 0xDD070A10)` 覆盖整个 `Box(fillMaxSize)`）保留不动，与面板上沿 24dp 渐变是两个不同层；与 B3 hero ken-burns 无任何干扰（B3 `graphicsLayer` 只挂首页 hero backdrop，详情页 `TvLongFormDetailBackground` 完全独立）。`TvDetailPanelTokensTest`（6 条用例）锁定 token 区间（`BlurRadiusDp.value ∈ [12f, 32f]`、`UpperGradientHeightDp.value ∈ [16f, 40f]`、`TopCornerRadiusDp.value ≥ 16f`、`ContentPaddingHorizontalDp.value ≥ 24f`、`ContentPaddingVerticalDp.value ≥ 16f`）+ `ScrimColorFallback.alpha > ScrimColorBlurred.alpha` 强约束（保证 API < 31 fallback 加深而非更透）+ 源文 audit（`TvDetailPanel.kt` 必含 `object TvDetailPanelTokens` 与 7 个 token 名）。`TvLongFormDetailGlassPanelSpecTest`（4 条用例）做源文 audit：(a) `TvLongFormDetailScreen.kt` 必含 `import com.chee.videos.core.ui.TvDetailGlassPanel` / `import com.chee.videos.core.ui.TvDetailPanelTokens` / `TvDetailGlassPanel(` / `TvDetailPanelTokens.ContentPaddingHorizontalDp` / `TvDetailPanelTokens.ContentPaddingVerticalDp`；(b) `TvLongFormDetailScreen.kt` **不**再含旧版裸字面量 `0xD20B1018`；(c) `TvDetailPanelBackground.kt` 必含 `Build.VERSION.SDK_INT >= Build.VERSION_CODES.S` + `Modifier.blur(` / `.blur(` + `import androidx.compose.ui.draw.blur` + `Brush.verticalGradient` + 5 个 token 具名引用（`BlurRadiusDp` / `ScrimColorBlurred` / `ScrimColorFallback` / `UpperGradientHeightDp` / `TopCornerRadiusDp`）；(d) `TvDetailGlassPanel` 必须是 `@Composable` 函数。`CONTEXT.md` 新增「TV 沉浸式详情玻璃面板」词条：写明视觉分层（背景 + 上沿渐变 + 面板主体）、API gating 协议（`VERSION_CODES.S` 判定 + 条件 `.then(Modifier.blur)`）、「玻璃」语义边界（羽化非看穿）、token 收口约束（调用点禁裸 `20.dp` / `0xCC0A0E16` / `0xE60A0E16` / `24.dp` / `36.dp` / `28.dp` / `0xD20B1018` 字面量）、作用域（仅 `TvLongFormDetailScreen`，电视剧详情不套用）、与「沉浸式详情首屏」既有约定的关系。本提交不动 hover-exit / FocusRequester / 三层防线、不动 B1 / B2 / B3 token、不动 `TvLongFormDetailBackground`（backdrop 与 poster fallback 分层独立）、不动按钮焦点视觉（`TvDetailPrimaryActionButton` / `TvDetailSecondaryActionButton` 仍走 B2 双层 glow）、不动电话端（已物理隔离）、不引入 `RenderEffect` 真「看穿玻璃」（留待后续）。B5（圆角统一收口到 16dp）后续另起，C 批未启动。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvDetailPanel.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvDetailPanelBackground.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`（仅底部面板块 + import）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvDetailPanelTokensTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailGlassPanelSpecTest.kt`（新增）、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（27s），所有 TV unit test 持续绿，重点关注新增 `TvDetailPanelTokensTest`（6 条 token 区间 + alpha 强约束 + 源文 audit）与 `TvLongFormDetailGlassPanelSpecTest`（4 条 helper / 调用点 audit）全部通过；既有不变量 `TvHeroMotionTokensTest` / `TvFeaturedHeroMotionSpecTest` / `TvMotionTokensTest` / `TvFocusSpecTest` / `TvTypographySpecTest` / `TvListMotionSpecTest` / `TvBringIntoViewSpecTest` / `TvSharedPosterTransitionSpecTest` / `TvInitialFocusSafeRequestTest` / `TvNoBareLaunchedEffectFocusRequestAuditTest` / `TvInitialFocusEffectShapeTest` / `TvMainActivityInputPolicyTest` 同步绿。`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL（14s），输出 `tv-app-arm64-v8a-debug.apk`（70MB）与 `tv-app-armeabi-v7a-debug.apk`（67MB）两个 ABI 分包。待手测：(a) 4K TV（API 31+）`adb install -r` 装上 0.1.51 APK 后进入电影 / `18+` 详情页（沉浸式详情首屏），底部信息面板上沿应能看到 24dp 高的「渐入」效果而非硬切，面板边缘有 blur 羽化感（不是真「看穿背景」，是面板自身渲染的边缘晕开）；面板内文字（标题 / 元信息 / 简介 / 演员名）在 1080p / 4K 下都清晰可读（≥4.5:1 对比度达 AA）；切换不同 backdrop 颜色的条目面板视觉一致、不出现某个 backdrop 下文字看不清的回归；(b) 切到 `TvSeriesDetailScreen`（电视剧详情）确认其视觉**未**受本次改动影响（poster + episodes 网格布局不变）；(c) 焦点视觉（播放 / 收藏按钮的 B2 双层 glow）、返回按钮命中、shared-element 进入/退出动画全部正常；(d) 若有 API < 31 的 TV 盒子，额外验证面板不模糊但 scrim 加深（0xE6 比改动前 0xD2 更深），整体文字可读性优于改动前，**不**出现 `NoSuchMethodError: Modifier.blur` 之类的运行时崩溃；(e) `dumpsys package com.chee.videos.tv | grep versionName` 应输出 `versionName=0.1.51`。下一步可推进 B5（圆角统一收口到 16dp），完成后 B 批结业进入 C 批。

## 2026-05-22 12:30 +0800
- 进度：落地 B 批第三项 B3——TV `TvCatalogScreen` 首页 hero（`TvFeaturedHero`）的 backdrop `AsyncImage` 升级为缓慢 Ken Burns 环境动效（缓慢缩放 + 菱形漂移）+ 系统级 reduce-motion 探测。视觉目标：120s 半周期 tween + `RepeatMode.Reverse` 双向往返（视觉总周期 240s），`scale 1.05 ↔ 1.10` + `translation ±8dp / ±4dp` 菱形漂移；系统 `Settings.Global.ANIMATOR_DURATION_SCALE == 0f` 时整张图冻结在 `scale = 1.075、translation = 0`。改动按四块：(1) 新建 `core/ui/TvHeroMotion.kt` 暴露 `object TvHeroMotionTokens`，6 个 token：`RampDurationMs = 120_000`（半周期 tween 时长）、`ScaleStart = 1.05f` / `ScaleEnd = 1.10f`（lerp 区间端点）、`ScaleStaticTarget = 1.075f`（reduce-motion 冻结目标 = `(ScaleStart + ScaleEnd) / 2f` 中点，容差 0.001f）、`PanOffsetXDp = 8.dp` / `PanOffsetYDp = 4.dp`（半周期内 -8→+8 / -4→+4 漂移），全部 `const val` / `val`，约束 `ScaleStart < ScaleEnd`、`PanOffsetYDp ≤ PanOffsetXDp`；(2) 新建 `core/ui/TvAccessibilityMotion.kt` 暴露 `@Composable fun rememberTvReduceMotionEnabled(): Boolean`——TV 工程**唯一**读 `android.provider.Settings.Global.ANIMATOR_DURATION_SCALE` 的入口（grep 全仓确认 B3 之前零处读这个值），`scale == 0f` 时返回 true 表示用户在开发者选项 / 无障碍里关闭了动画；`remember(context)` 缓存结果，不监听 `SettingsObserver`，系统级 setting 改动罕见、需要 app 重启才生效是公认的可接受约定，未来其他动效（B5 圆角动效 / C2 状态屏渐入）新增 reduce-motion 探测必须复用该 helper；(3) 改 `feature/tv/TvCatalogScreen.kt` 的 `TvFeaturedHero`（line 759-868 函数体扩张）：函数开头插入 `val reduceMotion = rememberTvReduceMotionEnabled()` + `val transition = rememberInfiniteTransition(label = "tvHeroKenBurns")` + `val progress by transition.animateFloat(initialValue = 0f, targetValue = if (reduceMotion) 0f else 1f, animationSpec = infiniteRepeatable(tween(TvHeroMotionTokens.RampDurationMs, easing = TvMotionTokens.EasingStandard), repeatMode = RepeatMode.Reverse), label = "tvHeroKenBurnsProgress")`，再用 `LocalDensity.current` 把 `PanOffsetXDp/YDp` 转 px，用 `androidx.compose.ui.util.lerp` 把 `progress` 分别映射到 `heroScale ∈ [ScaleStart, ScaleEnd]` 与 `heroTranslationXY ∈ [-panXY, +panXY]`，reduce-motion 命中时 `heroScale = ScaleStaticTarget`、`heroTranslationX = heroTranslationY = 0f`；backdrop `AsyncImage` modifier 由 `Modifier.fillMaxSize()` 改为 `Modifier.fillMaxSize().graphicsLayer { scaleX = heroScale; scaleY = heroScale; translationX = heroTranslationX; translationY = heroTranslationY }`——`graphicsLayer` **只**挂 backdrop 一张图，**不**挂上层 horizontal gradient / `TvFeaturedPoster` / 文案 Row，避免文字跟着抖；fallback 渐变分支（`backdropUrl.isNullOrBlank()`）不挂动效，保持纯静态；import 增补 5 个 `androidx.compose.animation.core.*`（`RepeatMode` / `animateFloat` / `infiniteRepeatable` / `rememberInfiniteTransition` / `tween`）+ 4 个 `androidx.compose.ui.*`（`graphicsLayer` / `LocalDensity` / `util.lerp`）+ 3 个 `com.chee.videos.core.ui.*`（`TvHeroMotionTokens` / `TvMotionTokens` / `rememberTvReduceMotionEnabled`）；(4) `android-tv-app/tv-app/build.gradle.kts` 版本 `versionCode 50 → 51`、`versionName "0.1.49" → "0.1.50"`。物理边界：`Surface(shape = AppChrome.CardShape)` 自带 clip 是 scale + pan 不漏边的物理保证（`ScaleStart = 1.05` 放大 5% 后在 1280dp 宽 hero 上水平安全余量 ~64dp 远超 8dp pan），不能改 hero 容器去掉 shape；缓动必须复用 `TvMotionTokens.EasingStandard`（统一 TV 端所有 tween 缓动），但 B3 **不**复用 `DurationFastMs/StandardMs/EmphasizedMs`——120s 比 TV 端 200-260ms 时长档高三个数量级，属不同尺度。`TvHeroMotionTokensTest`（5 条用例）锁定 token 区间（`RampDurationMs ∈ [60_000, 300_000]`、`ScaleStart < ScaleEnd`、scale 端点合法区间、`ScaleStaticTarget == midpoint ± 0.001f`、pan 振幅区间、`PanY ≤ PanX`）+ 源文 audit（`TvHeroMotion.kt` 必含 `object TvHeroMotionTokens` 与 6 个 token 名）。`TvFeaturedHeroMotionSpecTest`（5 条用例）切出 `TvFeaturedHero` 函数体（`private fun TvFeaturedHero(` 起、`private fun TvFeaturedPoster(` 止）做源文 audit：(a) 必含 `rememberInfiniteTransition(` / `infiniteRepeatable(` / `RepeatMode.Reverse` / `TvHeroMotionTokens.RampDurationMs` / `TvMotionTokens.EasingStandard`；(b) 必含 6 个 token 的具名引用 `TvHeroMotionTokens.ScaleStart/ScaleEnd/ScaleStaticTarget/PanOffsetXDp/PanOffsetYDp`；(c) 函数体内**禁止**裸 `120_000` / `1.05f` / `1.10f` / `1.075f` 字面量；(d) 必含 `graphicsLayer` 与 `rememberTvReduceMotionEnabled`；(e) `TvAccessibilityMotion.kt` 必含 `Settings.Global.ANIMATOR_DURATION_SCALE` + `@Composable` + `LocalContext.current` + `remember(`。`CONTEXT.md` 在「TV 焦点双层 glow」之后、「TV 焦点 ISE 三层防线」之前插入新词条「TV hero ken-burns 环境动效」：写明 6 个 token 的取值与区间、`rememberInfiniteTransition + animateFloat + lerp + RepeatMode.Reverse` 的驱动结构、`graphicsLayer` 只挂 backdrop 的强约束、调用点禁裸字面量的 audit 边界、`Surface.shape` clip 的物理保证、缓动复用 `EasingStandard`、reduce-motion 协议（唯一入口 + remember 缓存 + 不监听 SettingsObserver + 未来动效必须复用），并交叉引用 `TV 动效时长 token`。本提交不动 hover-exit / FocusRequester / 三层防线、不动 B1 / B2 token、不改 hero 文案 / poster / 上层渐变、不改 fallback 渐变、不改电话端（已物理隔离）。B4（沉浸式详情底部信息面板渐变 + 玻璃模糊）与 B5（圆角统一收口到 16dp）后续另起，C 批未启动。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvHeroMotion.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvAccessibilityMotion.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`（仅 `TvFeaturedHero` + import）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvHeroMotionTokensTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvFeaturedHeroMotionSpecTest.kt`（新增）、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（31s）所有 TV unit test 持续绿，重点关注新增的 `TvHeroMotionTokensTest`（5 条 token 区间 + 源文 audit）与 `TvFeaturedHeroMotionSpecTest`（5 条 hero 函数体 + accessibility helper audit）全部通过，既有不变量 `TvMotionTokensTest` / `TvFocusSpecTest` / `TvTypographySpecTest` / `TvListMotionSpecTest` / `TvInitialFocusSafeRequestTest` / `TvMainActivityInputPolicyTest` 同步绿。`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL（14s）输出 `tv-app-arm64-v8a-debug.apk`（70MB）与 `tv-app-armeabi-v7a-debug.apk`（67MB）两个 ABI 分包。待手测：用户在 4K TV `adb install -r` 安装 0.1.50 APK 后——(a) 进入首页 hero 区域盯 30-60s，能察觉 backdrop 在缓慢「呼吸」（极缓慢放大缩小 + 菱形漂移），但视觉上不刺眼、不会感觉「在动」；(b) 开「设置 → 开发者选项 → 绘图 → 窗口动画缩放 / 过渡动画缩放 / 动画程序时长缩放」全部设为「动画关闭」→ 杀掉 app 重开 → hero backdrop 冻结成静帧；(c) 上层文案 Row / `TvFeaturedPoster` / 播放按钮 focus glow 全部正常，不跟着抖；(d) hero 内容切换（继续观看 ↔ 精选影片）时动效自然过渡、不跳变；(e) `dumpsys package com.chee.videos.tv | grep versionName` 应输出 `versionName=0.1.50`。

## 2026-05-22 11:45 +0800
- 进度：复盘并归档 TV 端 `FocusRequester is not initialized` FATAL 排查——用户在 2026-05-21 19:49:23 截到栈帧 `FocusRequester.focus$ui_release(FocusRequester.kt:259)` → `FocusRequester.requestFocus(FocusRequester.kt:65)` → `TvCatalogScreenKt$TvCatalogScreen$6$1.invokeSuspend(TvCatalogScreen.kt:124)` → `BaseContinuationImpl.resumeWith` → `DispatchedTask.run` → `AndroidUiDispatcher.performTrampolineDispatch` → 主 Looper → `TvMainActivity.installMainLooperHoverExitGuard$lambda$0(TvMainActivity.kt:47)` 的 FATAL，并附 Suppressed `StandaloneCoroutine{Cancelling}@582c992` / `AndroidUiDispatcher@b067f63`。排查路径走两步：第一步把崩溃日 19:49 与本仓库最近三次相关 commit 时间对位——`6efef1a3 扩展TV hover-exit兜底到主Looper` 在 19:27:29、`d24165ef 修复TV首页hero上滑裁切并升级焦点动效` 在次日 09:18:49、`bf70df07 TV B2 落地焦点双层 glow` 在今天稍早；崩溃恰好落在 `6efef1a3..d24165ef` 之间的 22 分钟空窗——那一版 APK 只接了 hover-exit 的主 Looper 兜底，**没有**接 focus-requester 兜底，coroutine 内 `requestFocus()` 抛出的 ISE 经 `AndroidUiDispatcher` 异步链一路逃到 `installMainLooperHoverExitGuard$lambda$0`，因匹配规则只覆盖 hover-exit 而 fallback `throw err` → `AndroidRuntime` FATAL。第二步用 `AskUserQuestion` 与用户对接确认设备上装的是「崩溃日的旧 APK（pre-d24165ef）」，锁定根因为「旧 APK 缺失第 3 层 matcher」而非「当前 master 防线失效」。当前 master（含 `d24165ef` 与今天的 `bf70df07`）已经把三层防线全部带上：第 1 层 `FocusRequester.tryRequestFocus()`（`core/ui/TvInitialFocusEffect.kt:24-35`）同步 try-catch；第 2 层 `LaunchedTvInitialFocus`（同文件 `:37-53`）`withFrameNanos { }` + `runCatching { block() }.onFailure { ... }`，`CancellationException` 重抛、focus-requester ISE 吞掉、其他重抛；第 3 层 `installMainLooperHoverExitGuard()` 主 Looper `Looper.loop()` try/catch 循环，匹配 `shouldSwallowTvComposeHoverExitCrash` 或 `shouldSwallowTvComposeFocusRequesterCrash`（焦点 matcher 条件为 `IllegalStateException` + message `contains("FocusRequester is not initialized")` + 栈含 `androidx.compose.ui.focus.` 前缀帧）。`TvCatalogScreen.kt:123-135` 的初始焦点请求自 `d24165ef` 起已改用 `tryRequestFocus()`，第 1 层即可拦截，第 2、3 层是兜底。本轮**不做代码改动**——三层防线就位、当前栈帧条件正确，唯一动作是要求用户重新编译 / 安装当前 master APK，并把三层防线整体语义沉到 `CONTEXT.md`：在「TV 焦点双层 glow」之后新增「TV 焦点 ISE 三层防线」词条，写明三层结构（同步 `tryRequestFocus` / 协程 `LaunchedTvInitialFocus.runCatching` / 主 Looper `installMainLooperHoverExitGuard` + `shouldSwallowTvComposeFocusRequesterCrash`）、判别签名（message 含关键字 + 栈含 focus 包前缀帧）、强约束（`LaunchedTvInitialFocus` 块体内必须用 `.tryRequestFocus()` 而不能裸调 `.requestFocus()`，一次性事件回调除外）、不允许移除任何一层，并交叉引用三个既有词条（`TV 焦点请求安全调用 tryRequestFocus` / `TV 初始焦点请求约束` / `TV 主 Looper FocusRequester 未初始化兜底`）。TV 版本不 bump（无代码改动）。本提交不动 `.kt`、不引入新 matcher、不动业务调用点。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（21s）锁住 5 条防线相关测试持续绿——`TvInitialFocusSafeRequestTest`（`tryRequestFocus()` 行为）、`TvInitialFocusEffectShapeTest`（`LaunchedTvInitialFocus` 结构）、`TvInitialFocusRequesterMatcherTest`（matcher）、`TvNoBareLaunchedEffectFocusRequestAuditTest`（强约束 audit：禁止裸 `LaunchedEffect { focusRequester.requestFocus() }`）、`TvMainActivityInputPolicyTest`（主 Looper 兜底，含与本次崩溃栈完全对位的 case `swallows compose focus requester not initialized from request focus path` + `swallows compose focus requester crash from focus search dpad path` + `swallows focus requester crash with only focus owner impl frame`）；其余历史不变量 `TvFocusSpecTest` / `TvTypographySpecTest` / `TvColorContrastTest` / `TvMotionTokensTest` / `TvListMotionSpecTest` / `TvBringIntoViewSpecTest` / `TvSharedPosterTransitionSpecTest` / `TvCatalogFocusPolicyTest` 同步绿。`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL（10s），输出 `tv-app-arm64-v8a-debug.apk`（70MB）与 `tv-app-armeabi-v7a-debug.apk`（67MB）两个分包，可直接 `adb install -r` 到 4K TV 验证。待手测：用户在 4K TV 装上当前 master 编出的 0.1.49 APK 后，首次进入 TV 首页（`TvCatalogScreen` 完整加载且 `featuredFocusRequester` 命中）不再 FATAL；加载态 / 搜索态早退路径（`uiState.loading` / `isSearching` 为 true 时 `tryRequestFocus()` 不会跑）不再 FATAL；hover-exit 既有路径（菜单悬停后退出）仍按既有 `dispatchGenericMotionEvent` + 主 Looper 兜底吞掉、不 FATAL。若用户仍能复现 FATAL，则用 `adb shell dumpsys package com.chee.videos.tv | grep versionName` 核对装的是否为 `0.1.49`，并复制新的 stack trace 核对栈底是否仍指向 `installMainLooperHoverExitGuard$lambda$0(TvMainActivity.kt:47)`——若否，说明是新的失败模式，需另开 plan。

## 2026-05-22 11:10 +0800
- 进度：落地 B 批第二项 B2——TV 焦点双层 glow。把焦点反馈从「单层 0.15α 蓝青背景提亮 + 裸 graphicsLayer 黑灰 shadow」升级为「内层 0.6α tinted background + 外层 12dp tinted `Modifier.shadow` 扩散」双层结构，10-foot 视距下焦点态可识别度再上一档。`core/ui/TvFocus.kt` 改动：(a) `object TvFocusMotionTokens` 新增 `InnerGlowAlphaTarget: Float = 0.6f`（内层稳态 alpha，0.5–0.75 区间——既明显又不刺眼）与 `OuterHaloElevationDp: Dp = 12.dp`（外层扩散，8–16dp 区间——既看得见又不漏到邻卡，与 `TvFocusSafeSpec.posterFocusSafeSpaceDp = 8f` 安全空间留够余量），`OuterHaloElevationDp` 直接选 `Dp` 类型不写 `Float`，避免调用点拼 `.dp` 且与 `Modifier.shadow(elevation: Dp, ...)` 签名对齐；(b) 移除原 `private val TvFocusGlowSurface = Color(0x2639D7E8)` 字面量（把 alpha 写死在色值里），改走 `TvFocusGlowColor.copy(alpha = InnerGlowAlphaTarget * surfaceAlpha)`——`surfaceAlpha` 仍由既有 `animateFloatAsState(SurfaceDampingRatio, SurfaceStiffness)` 驱动的聚焦淡入因子提供，`InnerGlowAlphaTarget` 控制稳态目标，两者相乘；(c) `tvFocusableGlow` 与 `tvFocusableScaleOnly` 双双新增 `val haloElevation by animateDpAsState(targetValue = if (isFocused && enabled) TvFocusMotionTokens.OuterHaloElevationDp else 0.dp, animationSpec = spring(SurfaceDampingRatio, SurfaceStiffness))` 让 halo elevation 与内层 surfaceAlpha 同节奏淡入淡出，避免 jump cut；(d) modifier 链顺序重写为 `onFocusChanged → onPreviewKeyEvent → graphicsLayer{scaleX,scaleY}（不再写 shadowElevation 字面量）→ Modifier.shadow(elevation=haloElevation, shape, clip=false, ambientColor=TvFocusGlowColor, spotColor=TvFocusGlowColor) → .background(...) （仅 tvFocusableGlow） → .focusable()`——scale 在 shadow 之前才能让 halo 跟着卡片视觉中心放大，shadow 在 background 之前才能让外层光晕作用于内层提亮，`clip = false` 必须保持才能让 halo 溢出 shape bounding box 形成扩散感；(e) `tvFocusableScaleOnly` 仍**不接** `.background(...)`——它用于海报卡，内层提亮会遮挡海报图，外层 tinted halo + scale 已是完整的「焦点信号」。`Modifier.shadow(ambientColor, spotColor)` 在 Android API 28+ 平台 path 直接拿到 cyan tint，API 26-27 自动回退到默认黑灰 graceful degradation（不崩溃、不报错，视觉略弱于 API 28+，本工程 `minSdk = 26` 完全兼容）。十余处既有 `tvFocusableGlow` / `tvFocusableScaleOnly` 调用点（首页菜单 / 海报墙 / 电视剧详情 / 长视频详情 / IPTV 频道行 / 播放器浮层 / 配对页 / 连接页等）**无需任何改动**即可继承新焦点反馈，因为入参签名 `enabled` / `shape` / `focusedScale` 与最外层 `Modifier` 调用链不变；`TvFocusSafeSpec.posterFocusSafeSpaceDp = 8f` 也无需改动（12dp halo 在视觉上扩散 ≈6–8dp，与 `gridItemSpacingDp = 16f` / `posterFocusSafeSpaceDp = 8f` 配合刚好）。`TvFocusSpecTest` 新增 10 条不变量（既有 14 条全部保留）：`InnerGlowAlphaTarget ∈ [0.5, 0.75]`、`OuterHaloElevationDp.value ∈ [8f, 16f]`、源文 `Modifier.shadow(` 出现 ≥2 次（`tvFocusableGlow` 与 `tvFocusableScaleOnly` 各一次）、源文必须含 `ambientColor = TvFocusGlowColor` / `spotColor = TvFocusGlowColor` / `clip = false` / `TvFocusMotionTokens.InnerGlowAlphaTarget` / `TvFocusMotionTokens.OuterHaloElevationDp` / `animateDpAsState(`、源文**不再包含** `shadowElevation = 32f` 与 `shadowElevation = 28f` 字面量（旧裸 shadow 已替换）、源文**不再包含** `Color(0x2639D7E8)`（旧 alpha-写死字面量已移除）。`CONTEXT.md` 在 `TV 10-foot 排版 token` 之后新增 `TV 焦点双层 glow` 词条：写明双层结构（内层 0.6α tinted background + 外层 12dp tinted shadow）、`InnerGlowAlphaTarget` 与 `OuterHaloElevationDp` 取值区间与几何意图、`Modifier.shadow(ambientColor, spotColor)` 的 API 28+ tint 路径与 API 26-27 graceful degradation、modifier 链顺序强约束、`tvFocusableScaleOnly` 不接 `.background()` 的海报卡考量、调用点不写 `shadowElevation = X` 字面量的强约束、新增焦点反馈视觉必须复用这两个 token 而非硬编码数字、与 `TvFocusSafeSpec.posterFocusSafeSpaceDp = 8f` 的安全空间耦合约束。TV 版本 `0.1.48`→`0.1.49`，`versionCode` 49→50。本提交不动 `TvFocusMotionTokens` 既有 spring 参数（A 批 token 锁定）、不动 `TvFocusGlowColor` 色相、不动 `TvFocusSafeSpec`、不动 hover-exit / FocusRequester / `LaunchedTvInitialFocus` 链路、不引入系统 reduced-motion 识别（B3 一并考虑）、不改电话端（已物理隔离）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `InnerGlowAlphaTarget` / `OuterHaloElevationDp` 未定义编译失败（4 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含 `TvFocusSpecTest` 扩容到 24 条用例），`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL，输出 `armeabi-v7a`（67MB）与 `arm64-v8a`（70MB）两个分 APK。`TvTypographySpecTest`、`TvColorContrastTest`、`TvMotionTokensTest`、`TvListMotionSpecTest`、`TvBringIntoViewSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvInitialFocusEffectShapeTest`、`TvInitialFocusRequesterMatcherTest`、`TvMainActivityInputPolicyTest` 等历史不变量持续绿。待手测（4K TV，API 28+）：TV 首页海报卡聚焦应能看到外层蓝青 halo 扩散（≈6–8dp 距离），halo 跟随 1.04× 缩放一起放大，**不漏**到邻卡；海报墙 / 长视频详情 hero 按钮 / IPTV 频道行 / 设置按钮 / 配对页输入框等所有用 `tvFocusableGlow` 的位置均显示「内层 0.6α 蓝青提亮 + 外层 tinted halo」双层效果；用 `tvFocusableScaleOnly` 的海报卡只显示「外层 tinted halo + scale」、不带内层背景（保证海报图色彩不被遮挡）；按下反馈（DPad center / Enter）时 scale 落到 0.97f，halo 同步轻微收缩、无 jitter 无延迟；API 26-27 设备 halo 回退到默认黑灰（graceful degradation，不崩溃），用户体验略弱于 API 28+ 但仍有 scale + 内层提亮；A 批已落地的焦点放大 spring、列表 stagger、长视频播放器淡入淡出均不应有任何变化。B 批下一步可推进 B3（首页 hero ken-burns 120s 1.03→1.08 含 reduced-motion）/ B4（沉浸式详情底部信息面板渐变改 16:9 短渐变 + 玻璃模糊）/ B5（圆角统一收口到 16dp）。

## 2026-05-22 07:30 +0800
- 进度：落地 B 批第一项 B1——TV 端 10-foot 排版 / WCAG AAA 对比度收口。新增 `core/ui/TvTypography.kt` 暴露 `object TvTypographyTokens`（三档地板：`MainTitleSp = 34` 主标题、`SubtitleFloorSp = 22` 副标题、`HelperFloorSp = 18` 助记、`TightHelperSp = 18` labelSmall 兜底）与 `val TvTypography: Typography`（覆写 13 个 Material3 type role 中除 `display*` 外的 12 个：`headlineLarge = 34/42`、`headlineMedium = 30/38`、`headlineSmall = 28/36`、`titleLarge = 26/32`、`titleMedium = 24/30 letterSpacing=0.15`、`titleSmall = 22/28 letterSpacing=0.1`、`bodyLarge = 20/28`、`bodyMedium = 18/24`、`bodySmall = 18/22`、`labelLarge = 18/22 letterSpacing=0.1`、`labelMedium = 18/22`、`labelSmall = 18/22`，`display*` 在 TV 工程零调用点保留 Material3 默认，避免引入未使用值）。`tv/TvShellApp.kt:83` 把 `MaterialTheme(colorScheme = AppDarkColors)` 改成 `MaterialTheme(colorScheme = AppDarkColors, typography = TvTypography)` 并 `import com.chee.videos.core.ui.TvTypography`——注入点唯一，作用域覆盖 NavHost 内所有 TV 屏幕。TV main source 0 处 `fontSize = X.sp` 字面量，200+ 处 `MaterialTheme.typography.<role>` 调用点静默继承新值，调用点零改动。同步把 `core/ui/AppChrome.kt:23` 的 `TextMuted` 从 `Color(0xFF96A0B2)`（旧 contrast on `SurfaceElevated` ≈ 6.6:1，未达 WCAG AAA 7:1）抬到 `Color(0xFFB0BAC8)`（R176/G186/B194，实测 contrast on `SurfaceElevated` ≈ 7.9:1，on `SurfaceStrong` ≈ 7.4:1，on `Canvas` / `CanvasRaised` ≥ 10:1，全部达标 AAA），覆盖海报卡 `updateText` / 详情元信息 / `已观看 N%` / `共 N 项内容` 等 34 处 helper 文本场景；不动 `TextPrimary` / `TextSecondary` / `TextSubtle` / 其它 surface / accent token。新增 `TvTypographySpecTest` 10 条不变量：headlineLarge ≥ 34、titleSmall ≥ SubtitleFloorSp、titleMedium ≥ 22、bodyMedium ≥ HelperFloorSp、bodySmall/labelLarge/labelMedium/labelSmall 全部 ≥ 18、token 常量 ≥ 阈值、heading→title 单调（hL ≥ hM ≥ hS ≥ tL ≥ tM ≥ tS）、bodyLarge ≥ bodyMedium、12 个 role 全部 lineHeight ≥ fontSize、源文断言 `TvShellApp.kt` 必须包含 `typography = TvTypography` 与 import、`MainTitleSp == headlineLarge.fontSize` / `SubtitleFloorSp == titleSmall.fontSize` 同步校验。新增 `TvColorContrastTest` 6 条不变量：纯函数 `wcagRelativeLuminance(Color)` / `wcagContrastRatio(Color, Color)` 复刻 WCAG 2.x 公式（sRGB ≤ 0.03928 时除以 12.92，否则 `((c+0.055)/1.055)^2.4` 线性化；Y = 0.2126R + 0.7152G + 0.0722B；contrast = `(L_lighter + 0.05) / (L_darker + 0.05)`），白底黑字 ≈ 21:1、对前后景对称、白色 luminance ≈ 1 / 黑色 ≈ 0；TextMuted / TextSecondary / TextPrimary 三档前景 on `SurfaceElevated` 全部 ≥ 7.0；TextMuted on 全部 6 档 dark surface（Canvas / CanvasRaised / Surface / SurfaceElevated / SurfaceMuted / SurfaceStrong）全部 ≥ 7.0，避免某个更亮 surface 反而退档。`CONTEXT.md` 在 `TV 动效时长 token` 之后新增 `TV 10-foot 排版 token` 词条（含三档地板、12 个 role 数值、`TvShellApp` 唯一注入点、调用点不写 sp 字面量的强约束、新增 role 须扩 token 的扩展策略），在 `TV 焦点视觉语言` 之后新增 `TV 10-foot 对比度收口` 词条（WCAG AAA 7:1 强约束、TextMuted 旧值/新值对比、WCAG 2.x 公式细则、`wcagRelativeLuminance` / `wcagContrastRatio` 纯函数命名约定）。TV 版本 `0.1.47`→`0.1.48`，`versionCode` 48→49。本提交不动 hover-exit / FocusRequester 主 Looper 兜底、不动 `LaunchedTvInitialFocus` / `tryRequestFocus` 链路、不动 BringIntoView 注入、不动 SharedTransitionLayout 链路、不动 `TvFocusMotionTokens` / `TvMotionTokens`、不引入系统 reduced-motion 识别（B3 一并考虑）、不改 `TextPrimary` / `TextSecondary` / surface 色值、不改电话端（已在 `tvMainSourceExcludes` 内物理隔离）、不动 NavHost transition。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTypographySpecTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvColorContrastTest.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `TvTypography` / `TvTypographyTokens` 未定义编译失败（40+ 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含新增 `TvTypographySpecTest` 10 条 + `TvColorContrastTest` 6 条），`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL，输出 `armeabi-v7a` 与 `arm64-v8a` 两个分 APK。`TvFocusSpecTest`、`TvMotionTokensTest`、`TvListMotionSpecTest`、`TvBringIntoViewSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvInitialFocusEffectShapeTest`、`TvInitialFocusRequesterMatcherTest`、`TvMainActivityInputPolicyTest` 等历史不变量持续绿。待手测（4K TV）：TV 首页 hero 标题（34sp）、section 副标题（18sp）、海报卡标题（22sp）目视都应比 0.1.47 更大、更清晰；海报卡 `updateText` / 详情元信息 / `已观看 N%` 提示（TextMuted）目视应比 0.1.47 更亮、不再「灰糊」；海报墙 / 电视剧详情 / 长视频详情 / IPTV 频道行 / 配对页 / 设置页所有文本不应出现「溢出截断」（titleMedium 16→24sp、titleSmall 14→22sp 涨幅最大，注意 `maxLines = 1` 区域）；焦点放大 / 按下反馈 / hero 上滑钉位 / 列表 stagger / 长视频播放器淡入淡出均不应有任何变化（A 批不动）。B 批下一步可推进 B2（焦点 glow 双层：内 0.6α 紧贴 + 外 0.25α / 12dp 扩散）/ B3（首页 hero ken-burns 120s 1.03→1.08 含 reduced-motion）/ B4（沉浸式详情底部信息面板渐变改 16:9 短渐变 + 玻璃模糊）/ B5（圆角统一收口到 16dp）。

## 2026-05-22 06:48 +0800
- 进度：落地 A 批第五项 A5——TV 端动效时长 / easing token 共享收口。新增 `core/ui/TvMotion.kt` 暴露 `object TvMotionTokens`：三档 duration `DurationFastMs = 200`（小型瞬时反馈如临时浮层 alpha）、`DurationStandardMs = 240`（默认 TV 过渡，控制条 / 浮层 fade）、`DurationEmphasizedMs = 260`（入场强调，列表 stagger），严格升序、全部落在 A5 计划要求的 200–260ms 区间，超过 300ms TV 端就感觉迟滞、低于 200ms 又会丢失动画感；一个 easing `EasingStandard: Easing = CubicBezierEasing(0.2f, 0f, 0f, 1f)`（Material 标准缓动），调用点不再写裸 cubic-bezier 字面量。收口两个既有调用点：(a) `core/ui/TvListMotion.kt` 把 `StaggerEntryDurationMs = 260` 改成引用 `TvMotionTokens.DurationEmphasizedMs`、`StaggerEntryEasing = CubicBezierEasing(0.2f, 0f, 0f, 1f)` 改成引用 `TvMotionTokens.EasingStandard` 同一实例（用 `===` 断言锁实例引用而非数值复制），同时移除 `import androidx.compose.animation.core.CubicBezierEasing`；(b) `core/ui/LongFormVideoPlayer.kt` 共 4 处 `AnimatedVisibility`（seek preview、center feedback、top controls、bottom controls）的 `enter = fadeIn()` / `exit = fadeOut()` 全部替换成 `fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard))` / `fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard))`，新增 `import androidx.compose.animation.core.tween`。spring 系动效（焦点放大、按下反馈、光晕淡入）继续由 `TvFocusMotionTokens` 提供物理参数，不与 duration token 混用——`TvMotionTokens` 只负责 tween/easing 系，`TvFocusMotionTokens` 只负责 spring 系，职责切割明确。新增 `TvMotionTokensTest` 7 条不变量：三档 duration 必须落在 200..260 区间、严格升序（Fast < Standard ≤ Emphasized）、`EasingStandard is CubicBezierEasing`、`TvListMotionTokens.StaggerEntryDurationMs == TvMotionTokens.DurationEmphasizedMs` 且 `StaggerEntryEasing === TvMotionTokens.EasingStandard`（同一实例引用，防止"数值复制+不同实例"绕过收口）、`TvMotion.kt` 必须出现 object 与 4 个名字、`TvListMotion.kt` 必须出现 `TvMotionTokens.EasingStandard` / `TvMotionTokens.DurationEmphasizedMs` 引用且不再含 `CubicBezierEasing(0.2f, 0f, 0f, 1f)` 字面量、`LongFormVideoPlayer.kt` 不应再含裸 `fadeIn()` / `fadeOut()` 字符且必须引用 `TvMotionTokens.DurationStandardMs` / `TvMotionTokens.EasingStandard`。`CONTEXT.md` 在 `TV 焦点动效物理` 之后新增 `TV 动效时长 token` 词条，固化三档数值意图、Fast/Standard/Emphasized 适用场景、与 spring 体系的职责切割、所有 TV `tween` 动画必须从该 token 拉数值、新增时长场景应扩 token 而不允许调用点硬编码数字。TV 版本 `0.1.46`→`0.1.47`，`versionCode` 47→48。本提交不动 hover-exit / FocusRequester 主 Looper 兜底、不动 BringIntoView 注入、不动 SharedTransitionLayout 链路、不动 NavHost transition、不引入系统 reduced-motion 识别。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvMotion.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvMotionTokensTest.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvListMotion.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `TvMotionTokens` 未定义编译失败（4 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含新增 `TvMotionTokensTest` 7 条用例），`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL，输出 `armeabi-v7a` 与 `arm64-v8a` 两个分 APK。`TvFocusSpecTest`、`TvListMotionSpecTest`、`TvBringIntoViewSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvMainActivityInputPolicyTest` 等历史不变量持续绿。待手测：长视频播放器内 seek 提示、center 反馈、控制条出现/隐藏的 fade 时长应落在 240ms 而不是过去的默认 400ms（更紧凑、不拖沓）；海报墙首屏列表 stagger 单 item 入场仍是 260ms / Material 标准缓动；切换电视剧/电影/`18+` 详情页时焦点放大 / 光晕 / 按下反馈无变化（spring 体系不受本次收口影响）。A 批 5 项至此全部落地，下一步进入 B 批（B1 10-foot 排版 / B2 双层 glow / B3 hero ken-burns / B4 渐变 + 玻璃模糊 / B5 圆角收口）或继续按需推进。

## 2026-05-22 06:10 +0800
- 进度：落地 A 批第四项 A4——DPad center 按下反馈。在 `core/ui/TvFocus.kt` 扩展 `object TvFocusMotionTokens`，新增 `PressedScale = 0.97f`（按下下沉目标 scale，0.94–0.99 可感知但不过度形变）、`PressDampingRatio = 0.7f`（略低于 `ScaleDampingRatio = 0.8f`，按下/回弹更紧凑）、`PressStiffness = 720f`（高于 `ScaleStiffness = 380f`，按下与回弹比悬停反馈明显更快）。新增三个 internal 工具：`isTvPressKey(key)`（统一注册 `Key.DirectionCenter` / `Key.Enter` / `Key.NumPadEnter` 三个 TV 按下键到 `TvPressKeys` 集合）、纯函数 `resolveTvFocusableScaleTarget(focused, pressed, enabled, focusedScale)`（顺序为 `!enabled → 1f`、`pressed → PressedScale`、`focused → focusedScale`、`else → 1f`，单测可直接锁定）、`tvFocusableScaleSpring(pressed)`（按下时返回 `Press*` 组 spring，否则返回 `Scale*` 组 spring）以及 `performTvPressHapticFeedback(view)`（按 `Build.VERSION.SDK_INT >= Build.VERSION_CODES.R` 守门切换 `HapticFeedbackConstants.CONFIRM` / `HapticFeedbackConstants.VIRTUAL_KEY`，避免 CONFIRM 在旧设备上静默失败）。`tvFocusableGlow` 与 `tvFocusableScaleOnly` 内部新增 `var isPressed by remember { mutableStateOf(false) }` 和 `val view = LocalView.current`，把 scale 的 `targetValue` 改成 `resolveTvFocusableScaleTarget(isFocused, isPressed, enabled, focusedScale)`，`animationSpec` 改成 `tvFocusableScaleSpring(isPressed)`；在 `onFocusChanged` 之后、`graphicsLayer` 之前插入 `onPreviewKeyEvent`：未使能或未聚焦直接 `return@onPreviewKeyEvent false`，非按下键直接 `false`，`KeyEventType.KeyDown` 翻转 `isPressed = true`、`KeyEventType.KeyUp` 翻转回 `false` 并调用 `performTvPressHapticFeedback(view)`，**整个 lambda 末尾统一返回 `false`** 表示不吞事件，让下游 `focusable()` / 调用点 `clickable()` 仍能收到 Enter/Center（按下反馈是装饰层不是行为层）。`onFocusChanged` 失焦时主动把 `isPressed = false` 复位，避免 keyUp 还没传到 modifier 就丢焦点时按下态卡住。十余处既有 `tvFocusableGlow` / `tvFocusableScaleOnly` 调用点（首页菜单、海报墙、电视剧详情、长视频详情、IPTV 频道行、播放器浮层、配对页、连接页等）**无需改动**即可继承新反馈，因为入参签名（`enabled` / `shape` / `focusedScale`）和最外层 `Modifier` 调用链不变。`TvFocusSpecTest` 新增 6 条不变量：`PressedScale ∈ [0.94, 0.99]`、`PressDampingRatio ∈ [0.55, 0.85]`、`PressStiffness > ScaleStiffness`；`resolveTvFocusableScaleTarget` 四象限纯函数（disabled / pressed / focused / neither）；源文断言 `onPreviewKeyEvent` + `Key.DirectionCenter` + `Key.Enter` + `Key.NumPadEnter` + `KeyEventType.KeyDown/KeyUp` + `resolveTvFocusableScaleTarget(` + `tvFocusableScaleSpring(` 必须同时出现；触觉源文断言 `HapticFeedbackConstants.CONFIRM` + `HapticFeedbackConstants.VIRTUAL_KEY` + `Build.VERSION.SDK_INT` + `Build.VERSION_CODES.R` 必须同时出现。`CONTEXT.md` 的 `TV 焦点动效物理` 词条整段扩展：写明三个 Press token 数值范围、`resolveTvFocusableScaleTarget` / `tvFocusableScaleSpring` / `isTvPressKey` / `TvPressKeys` / `performTvPressHapticFeedback` 的强约束、`onPreviewKeyEvent` 末尾统一返回 `false` 的语义（不吞事件）、失焦时复位 `isPressed` 的边界条件、API 30+ CONFIRM / 低于 30 fallback VIRTUAL_KEY 的守门规则。TV 版本 `0.1.45`→`0.1.46`，`versionCode` 46→47。本提交不动 hover-exit / FocusRequester 主 Looper 兜底、不动 BringIntoView 注入、不动 SharedTransitionLayout 链路、不引入 reduced-motion 系统设置识别（在 A5 / B 批一起考虑）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `PressedScale` / `PressDampingRatio` / `PressStiffness` / `resolveTvFocusableScaleTarget` 未定义编译失败（8 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含 `TvFocusSpecTest` 新增 6 条用例），`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL，输出 `armeabi-v7a` 与 `arm64-v8a` 两个分 APK。`TvBringIntoViewSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvListMotionSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvMainActivityInputPolicyTest` 等历史不变量持续绿。待手测：DPad center / Enter 按下时所有共享焦点控件应能看到「微缩 ~3%」按下反馈与回弹动画；支持 CONFIRM 触觉的设备能感到清晰 confirm haptic，低于 API 30 的设备回退到 VIRTUAL_KEY；按下过程中失焦（如长按时被打断）按下态不应卡住；onPreviewKeyEvent 末尾返回 `false`，因此 Enter/Center 仍能继续触发既有点击行为（菜单切换、海报选中、播放等），不应吞键。

## 2026-05-22 05:30 +0800
- 进度：彻底修复 TV 首页 hero「上滑裁切」根因。上一轮（04:05）用 `rememberLazyListState()` + `scrollToItem(0,0)` 钉位的兜底方案没解决（用户复测：「首页没有了动画，但依旧裁切」，附两张 4K TV 截图，对比异常态 hero 顶部被裁、内容整体上移约 225px；正常态 hero 头部完整、「最近更新」位于 y~745）。原因：`hasPinnedInitialScroll` 标志在数据加载早期、`initialFocusTarget` 第一次落到 `MENU` 时就被翻 `true`，等 featured 数据到达再切到 `FEATURED` 焦点时钉位已经跳过——根本上是在用「事后回滚」追打「事前 pivot」。再深一层追源：在 Gradle cache 里看到 Compose Foundation 1.7（含 BOM 2024.10.01）的 `BringIntoViewSpec.android.kt` 给 leanback 设备（`PackageManager.FEATURE_LEANBACK`）默认装的就是 `PivotBringIntoViewSpec`，它**对任何焦点目标都会返回非零位移**（`leadingEdgeOfItemRequestingFocus - 0.3 * containerSize`），不管该目标当下是否已经完整可见，并搭配 `tween(125ms, CubicBezierEasing(0.25f,0.1f,0.25f,1f))` 把 LazyColumn 拉到 30% pivot；TV 首页初始焦点落在 324dp hero 底部播放按钮时，pivot 会把整张 hero 上滑约 110dp，外观即「头部被裁切」；详情页 LazyRow 演员行、海报墙 LazyVerticalGrid、IPTV 频道行都吃同一份默认 spec。修法切换为「消除 pivot」而非「事后回滚 pivot」。
- 修复：新增 `core/ui/TvBringIntoView.kt`，暴露 `val TvMinimalBringIntoViewSpec: BringIntoViewSpec`（`@OptIn(ExperimentalFoundationApi::class)`），其 `calculateScrollDistance(offset, size, containerSize): Float` 委托给纯函数 `calculateTvMinimalBringIntoViewScrollDistance(...)`，后者复刻 `BringIntoViewSpec.Companion.defaultCalculateScrollDistance` 的「最少滚动」语义——目标完全可见（`leadingEdge >= 0 && trailingEdge <= containerSize`）或目标已横跨容器（`leadingEdge < 0 && trailingEdge > containerSize`）返回 0；否则在 `leadingEdge` 与 `trailingEdge - containerSize` 中选绝对值更小的一边作为位移。注入点：`tv/TvShellApp.kt` 在 `TvAuthenticatedNav` 顶部新增 `import androidx.compose.foundation.ExperimentalFoundationApi`、`import androidx.compose.foundation.gestures.LocalBringIntoViewSpec`、`import com.chee.videos.core.ui.TvMinimalBringIntoViewSpec`，把 `@OptIn(ExperimentalSharedTransitionApi::class)` 扩展为 `@OptIn(ExperimentalSharedTransitionApi::class, ExperimentalFoundationApi::class)`，在 `SharedTransitionLayout(...)` 外层包一层 `CompositionLocalProvider(LocalBringIntoViewSpec provides TvMinimalBringIntoViewSpec) { ... }`（嵌入 `Box(fillMaxSize().background(...))` 内、`if (showRootExitPrompt) { ... }` 之前闭合），作用域覆盖 NavHost 内所有 TV 屏幕。同时删除上一轮的失败兜底：`feature/tv/TvCatalogScreen.kt` 移除 `import androidx.compose.foundation.lazy.rememberLazyListState`、`import androidx.compose.runtime.mutableStateOf`、`import androidx.compose.runtime.setValue`、`val contentLazyListState = rememberLazyListState()`、`var hasPinnedInitialScroll by remember { mutableStateOf(false) }`、主 `LazyColumn` 的 `state = contentLazyListState` 参数以及 `LaunchedTvInitialFocus { ... }` 内的 `if (!hasPinnedInitialScroll) { contentLazyListState.scrollToItem(0, 0); hasPinnedInitialScroll = true }` 整段——根因消除后回钉机制属于过度兜底，留着反而会和后续 DPad 主动滚动竞争。新增 `core/ui/TvBringIntoViewSpecTest.kt` 7 条用例（全部 `@OptIn(ExperimentalFoundationApi::class)`）：item 完全可见返回 0、hero 内 44dp 播放按钮可见时返回 0（针对实际场景）、item 大于容器且当前覆盖容器返回 0、item 前缘超出顶部按 `-offset` 滚动、item 后缘超出底部按 `(offset+size)-containerSize` 滚动、双侧都超出取最小、`TvMinimalBringIntoViewSpec.calculateScrollDistance` 与纯函数返回值一致。`CONTEXT.md` 把 04:05 落地的 `TV 首页 hero BringIntoView 上滑钉位` 词条整段改写为 `TV BringIntoView 最小滚动策略`：写明 Compose Foundation 1.7 leanback 默认 `PivotBringIntoViewSpec` 的具体行为（30% pivot、`tween(125, CubicBezierEasing(0.25,0.1,0.25,1))`、连可见目标也滚）、`TvMinimalBringIntoViewSpec` 的语义复刻、`TvShellApp.TvAuthenticatedNav` 的 `CompositionLocalProvider` 注入点与 `@OptIn(ExperimentalFoundationApi::class)`、为什么旧 `scrollToItem(0,0) + hasPinnedInitialScroll` 钉位方案根本上是事后追打 pivot（钉位标志会被加载阶段 MENU target 翻 true），强约束「任何 TV 子树如需恢复 pivot 只能在该子树重新 provide spec，不允许在 `TvShellApp` 入口移除该 spec 注入」。TV 版本 `0.1.44`→`0.1.45`，`versionCode` 45→46。本提交不动 hover-exit / FocusRequester 主 Looper 兜底、不动 SharedTransitionLayout 链路、不动 `tryRequestFocus()`、不动详情页结构。
- 未解决（等待用户复测信号）：详情页「仍有上滑动画 + 裁切」——`TvLongFormDetailScreen` 是 `Box(fillMaxSize)`，本身没有可滚动祖先，但其内含的 LazyRow（演员行）会触发 pivot 行为；全局 `LocalBringIntoViewSpec` 注入理论上能同时覆盖该 LazyRow，但用户尚未提供详情页 4K 截图，无法直接量化前后位移。如果 0.1.45 详情页依然有体感裁切，下一轮需要附详情页截图判别（a）navigation-compose 2.7.7 默认 `fadeIn(tween(700))` 与 `SharedTransitionLayout` 包裹的合成效果，（b）`TvLongFormDetailBackground` 的 `AsyncImage` 异步首帧 + edge-to-edge `WindowInsets` 抵达时机错位。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvBringIntoView.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvBringIntoViewSpecTest.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含新增 `TvBringIntoViewSpecTest` 7 条用例，以及 `TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvFocusSpecTest`、`TvListMotionSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvMainActivityInputPolicyTest` 等历史不变量持续绿）；`./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL，输出 `armeabi-v7a` 与 `arm64-v8a` 两个分 APK。待手测：4K TV 冷启动进入 TV 首页（含 featured、含 continue-watching、含 sections）应不再有「hero 上滑、头部只剩一半」的初始动画，hero 头部完整可见、内容整体位置应与正常态截图一致；继续 DPad 下移到下方 shelf 时，shelf 项进入视口的滚动应是「最少滚动」（贴到容器边缘即停），不再被拉到 30% pivot；进入电影/`18+` 详情页背景图应画到屏幕物理顶端、演员 LazyRow 横滚不再有 pivot 行为；如果详情页体感仍存在，需要用户复测时附 4K 截图与"动画在导航过渡瞬间 vs 进入后才出现"的信号才能进一步定位。

## 2026-05-22 03:10 +0800
- 进度：修复"进入 TV 各类型首屏与详情页时顶部内容被裁切只能看到一半"的体感问题。根因复盘：`TvMainActivity` 启用了 `enableEdgeToEdge(statusBarStyle = SystemBarStyle.dark(TRANSPARENT), navigationBarStyle = SystemBarStyle.dark(TRANSPARENT))`，所有 NavHost 内页面均以"画到屏幕物理边缘"为前提排版。但在已落地的 5 个进入 TV 首页之后的页面级 root 里，处理 status bar inset 的策略不一致——`TvSeriesDetailScreen`（电视剧详情）和 `TvPairingScreen`（配对页）的最外层 `Box`/`Column` 已经叠了 `Modifier.statusBarsPadding()`，但 `TvCatalogScreen`（一级首页 Row）、`TvLongFormDetailScreen`（电影/`18+` 详情三个 Box 分支：loading / error / 主体沉浸首屏）和 `TvPosterWallScreen`（海报墙顶级 Column）三处遗漏，导致顶部 Row/Column 被状态栏覆盖而呈现"头部只能看到一半"。本轮按已有 working pattern（`TvSeriesDetailScreen.kt:73 / 85 / 108`、`TvPairingScreen.kt:148`）补齐 padding：`feature/tv/TvCatalogScreen.kt` 在两条 Row 分支（loading、main，行 138/151）的 `Modifier.fillMaxSize()` 之后追加 `.statusBarsPadding()` 并新增 `import androidx.compose.foundation.layout.statusBarsPadding`；`feature/tv/TvLongFormDetailScreen.kt` 三个 Box 分支（loading、error、main）的 `Modifier.fillMaxSize().background(...)` 之后均追加 `.statusBarsPadding()` 并新增同样的 import；`feature/tv/TvPosterWallScreen.kt` 顶层 Column 在 `.background(AppChrome.PageGradient)` 之后追加 `.statusBarsPadding()` 并新增 import。`TvLongFormDetailBackground` 内部的背景图 `Box` 故意保留不加 padding——背景属于底层装饰，本身可以画到屏幕边缘，承载文字与可聚焦操作的内容层 padding 来自外层 root Box，与 `TvSeriesDetailScreen` 思路一致。`CONTEXT.md` 在 `TV 滚动内容底部安全留白` 之后新增 `TV 安全区域顶部留白` 词条，固化 edge-to-edge 前提、所有页面级 root 必须叠 `statusBarsPadding()` 的强约束、沉浸式详情背景图层与全屏播放/QR 配对等例外、不允许通过子组件局部 padding 或 `safeContentPadding()` 替代外层 padding 的细则。TV 版本 `0.1.41`→`0.1.42`，`versionCode` 42→43。本提交不动 hover-exit 兜底、不动 focus requester 兜底、不动 NavHost 结构、不引入 `WindowInsetsController` 配置项。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待跑 `./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 锁定既有测试持续绿、TV APK 可装。待手测：冷启动进入电视剧 / 电影 / `18+` 首页时顶部菜单与右侧内容的第一行均完整可见；从首页打开任意海报墙、电视剧详情、电影/`18+` 详情时顶栏返回按钮与标题完整可见；播放器全屏画面、根启动加载、QR/配对二维码不应额外多出顶部留白。

## 2026-05-22 02:25 +0800
- 进度：修复新一轮 `FocusRequester is not initialized` FATAL。崩溃栈核心帧：`FocusRequester.focus$ui_release(FocusRequester.kt:259)` → `FocusRequester.requestFocus(FocusRequester.kt:65)` → `TvCatalogScreenKt$TvCatalogScreen$6$1.invokeSuspend(TvCatalogScreen.kt:124)` → `BaseContinuationImpl.resumeWith` → `DispatchedTask.run` → `AndroidUiDispatcher.performTrampolineDispatch` → 主 Looper → `TvMainActivity.installMainLooperHoverExitGuard$lambda$0(TvMainActivity.kt:47)`。`TvCatalogScreen.kt:124` 是 `LaunchedTvInitialFocus { ... }` 协程 lambda 体（`when (initialFocusTarget) { ... tryRequestFocus() }`），但栈帧上 **没有 `tryRequestFocus` 帧**——意味着 R8/D8 在某条编译路径上把 `tryRequestFocus` 扩展函数体内联了，导致同步 try-catch 未生效，ISE 通过 Compose 1.7 的 `AndroidUiDispatcher` 异步协程恢复路径透到主 Looper。这是已有 `TV hover 输入兼容兜底` 的对偶问题，修复方案对称：在 `TvMainActivity.kt` 新增 `internal fun shouldSwallowTvComposeFocusRequesterCrash(err: IllegalStateException): Boolean`，匹配条件——异常类型 `IllegalStateException`、消息 `contains("FocusRequester is not initialized")`（兼容 Compose 抛的多行带换行 message）、且栈含 `androidx.compose.ui.focus.FocusRequester` 类名下的 `requestFocus` / `focus$ui_release` / `findFocusTargetNode$ui_release` 三者之一的帧（前两个覆盖 `requestFocus()` 调用路径，第三个覆盖 DPad 按键 → Compose 内部 `FocusOwnerImpl.focusSearch` → `findFocusTargetNode$ui_release` 路径）。`installMainLooperHoverExitGuard` 内的 `Looper.loop()` try/catch 循环改为先尝试 `shouldSwallowTvComposeHoverExitCrash`，再尝试 `shouldSwallowTvComposeFocusRequesterCrash`，两者均不命中再原样抛出，最大限度保留其他 ISE 的可见性。`dispatchGenericMotionEvent` 不动（FocusRequester ISE 不走 motion 边界，避免误伤）。`TvInitialFocusEffect.kt` 的 `tryRequestFocus()` 不动——同步调用方仍是首要防线，主 Looper 兜底只在 R8 内联吃掉 try-catch 帧或异步恢复路径绕开时托底。TV 版本 `0.1.40`→`0.1.41`，`versionCode` 41→42。`CONTEXT.md` 在 `TV hover 输入兼容兜底` 之后新增 `TV 主 Looper FocusRequester 未初始化兜底` 词条，固化匹配条件、对偶语义、`tryRequestFocus` 仍是首要防线的约束，以及回归测试位置。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 全绿——新增 5 条用例覆盖：`swallows compose focus requester not initialized from request focus path`（complete async resume stack）、`swallows compose focus requester crash from focus search dpad path`（findFocusTargetNode$ui_release + focusSearch-ULY8qGw 帧）、`does not swallow focus requester crash with unrelated message`（栈匹配但消息不匹配）、`does not swallow focus requester message from non compose source`（消息匹配但栈不是 Compose `FocusRequester`）、`does not swallow hover exit message routed through focus requester matcher`（互不串话，hover-exit 消息走 FocusRequester 帧不应被新 matcher 误吞）。`./gradlew --no-daemon :tv-app:assembleDebug` 通过。原 5 条 hover-exit matcher 用例与 `TvInitialFocusSafeRequestTest`、`TvCatalogFocusPolicyTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvFocusSpecTest`、`TvSharedPosterTransitionSpecTest`、`TvListMotionSpecTest` 持续绿。待手测：TV 首页冷启动（含空内容、有 featured、有 continue-watching、有 sections / TV / Movie / AV shelf 各种 initialFocusTarget 分支）不再出现该 ISE；DPad 上下左右快速切换菜单与海报墙时不再出现该 ISE；海报墙 / 详情 / IPTV / 配对页冷启动场景均正常。

## 2026-05-22 01:45 +0800
- 进度：落地 A 批第三项 A3——海报墙 LazyVerticalGrid 入场 stagger。新增 `core/ui/TvListMotion.kt`：`object TvListMotionTokens` 集中暴露 `StaggerPerItemMs = 35L`（25–50ms TV 可感知区间）、`StaggerEntryDurationMs = 260`（落在 A5 200–280ms 上限内，预留 A5 收口空间）、`StaggerEntryDistanceDp = 12.dp`（8–20dp 安全区，避免视线追踪疲劳）、`StaggerMaxSteps = 12`（深处滚动入场最长等待 12 * 35 = 420ms，避免懒加载卡顿）、`StaggerEntryEasing = CubicBezierEasing(0.2f, 0f, 0f, 1f)`（强制 cubic，原计划要求不允许 linear）。`tvStaggerEntryDelayMs(index, ...)` 为纯函数，对 `index` 做 `coerceIn(0, maxSteps)` 后乘 `perItemDelayMs`，单测可直接锁定夹紧语义。`@Composable Modifier.tvStaggerEntry(index)` 通过 `LaunchedEffect(Unit) { delay(tvStaggerEntryDelayMs(index)); visible = true }` 调度，再用 `animateFloatAsState(tween(..., easing))` 驱动 `graphicsLayer { alpha = progress; translationY = (1 - progress) * distancePx }`，全程跑在 `graphicsLayer`（不触发布局回流）。`feature/tv/TvPosterWallScreen.kt`：`androidx.compose.foundation.lazy.grid.items` 改成 `itemsIndexed`，key 函数同步成 `{ _, item -> item.id }`；`focusRequester` modifier 抽到 `focusModifier` 后跟 `.tvStaggerEntry(index = index)`，**focusRequester 必须在 stagger 之前**，保证首项焦点请求时 FocusRequester 节点的挂载顺序仍然先于 stagger 的 alpha 动画。其他 LazyColumn / LazyVerticalGrid 调用点（IPTV 频道行、TV 首页 shelves、电视剧详情 episodes 网格）本轮不接入，留作后续单独 PR；规则上必须通过 `tvStaggerEntry` 入口，不允许调用点硬编码 alpha/translation 动画。`CONTEXT.md` 在 `TV 电视剧海报 shared-element 过渡` 之后新增 `TV 列表入场 stagger` 词条，固化 token 数值区间、graphicsLayer 强约束、`focusRequester` 顺序约束以及"所有列表 stagger 必须走 tvStaggerEntry"的强约束。TV 版本 `0.1.39`→`0.1.40`，`versionCode` 40→41。本提交不动 hover-exit 兜底、不动 shared-element 链路、不引入 reduced-motion 处理（系统 reduced-motion 支持留作单独议题）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvListMotion.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvListMotionSpecTest.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `tvStaggerEntryDelayMs` 未定义编译失败（4 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含 `TvListMotionSpecTest` 4 条用例：token 区间、helper 夹紧语义、`TvListMotion.kt` 结构断言、`TvPosterWallScreen` 接入断言）。`./gradlew --no-daemon :tv-app:assembleDebug` 通过。`TvSharedPosterTransitionSpecTest`、`TvFocusSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest` 等历史不变量持续绿。待手测：从首页打开任意海报墙（电视剧 / 电影 / `18+`）冷启动应能看到 item 一行一行错位淡入 + 上移到位的入场动画；滚动至底部触发 loadMore 时新加载的 item 受 maxSteps 夹紧只会有最多 420ms 的滞后；快速来回滑动时不应出现 alpha=0 残留或视觉抖动。

## 2026-05-22 00:35 +0800
- 进度：修复 Compose 1.7（BOM 2024.10.01 / UI 1.7.4）升级后冒出的新 TV FATAL `IllegalStateException: FocusRequester is not initialized`。崩溃栈显示异常从 `TvCatalogScreen.kt:124`（`featuredFocusRequester.requestFocus()`）出发，经 `BaseContinuationImpl.resumeWith` → `DispatchedTask.run` → `AndroidUiDispatcher.performTrampolineDispatch` 进入主 Looper，绕过 `LaunchedTvInitialFocus` 外层 `runCatching` 的 try-catch 帧后被 `TvMainActivity.installMainLooperHoverExitGuard` 的内 `Looper.loop()` 收到——但该兜底仅匹配 `AndroidComposeView.sendHoverExitEvent`/`dispatchHoverEvent` 调用栈，对 focus ISE 不命中，于是原样重抛触发 FATAL。Compose 1.7 之前 `runCatching { block() }` 能正常吃下 ISE，1.7 起协程恢复路径变成异步在 AndroidUiDispatcher 跳板上跑，inline 的 try-catch 帧并不总能覆盖到 invokeSuspend 抛出的瞬间，必须在调用点同步 try-catch。处置：在 `core/ui/TvInitialFocusEffect.kt` 新增 `fun FocusRequester.tryRequestFocus(): Boolean`——同步 try-catch `IllegalStateException`，复用既有 `isFocusRequesterNotInitialized` 关键字匹配；命中即吞掉返回 `false`，未命中原样重抛，保留 helper 既有兜底语义。把 8 个 TV 文件里 `LaunchedTvInitialFocus { ... }` 块体内的 13 处 `.requestFocus()` 全部切到 `.tryRequestFocus()`：`feature/tv/TvCatalogScreen.kt`（首页 8 处目标分支）、`feature/tv/TvPosterWallScreen.kt`（海报墙首格）、`feature/tv/TvSeriesDetailScreen.kt`（电视剧详情播放按钮）、`feature/tv/TvLongFormDetailScreen.kt`（电影/`18+` 详情播放按钮）、`feature/tv/TvIptvScreen.kt`（IPTV 根容器）、`tv/TvPairingScreen.kt`（配对页主操作）、`core/ui/LongFormVideoPlayer.kt`（播放器 root/控制条 pending 焦点）、`core/ui/SubtitlePicker.kt`（字幕选择器）。`LongFormVideoPlayer` 内 `try { requestFocus() } finally { pending = false }` 改成 `try { tryRequestFocus() } finally { pending = false }`，保留 finally 清 pending 的逻辑。其它一次性事件回调（点击、按键、动画完成）保持裸 `.requestFocus()`，因为这些路径上 try-catch 帧能正常生效。新增 `TvInitialFocusSafeRequestTest`：源文断言 helper 文件存在 `fun FocusRequester.tryRequestFocus`、行为断言真实 `FocusRequester` 未挂载时调用 `tryRequestFocus()` 不抛、命中返回 `false`、其它 ISE 不被误吞而原样重抛，并断言 `TvCatalogScreen` 已出现 `tryRequestFocus` 调用点。同步更新 `TvCatalogFocusPolicyTest` 中“MENU 兜底”的源文回归断言把 `menuFocusRequester.requestFocus()` 字面量换成 `menuFocusRequester.tryRequestFocus()`，错误消息扩成"既要保留兜底意图、又要走 tryRequestFocus 这一入口防止 ISE 透出"。`CONTEXT.md` 在 `TV 初始焦点请求约束` 之后新增 `TV 焦点请求安全调用 tryRequestFocus` 词条，写明 Compose 1.7 的协程恢复机制为何会绕开外层 try-catch、`tryRequestFocus` 是唯一安全入口、helper 块体内禁止裸调 `.requestFocus()`，并指出一次性事件回调可以继续裸调。TV 版本 `0.1.37`→`0.1.38`，`versionCode` 38→39。本提交不动 `installMainLooperHoverExitGuard` 匹配规则、不回退 compose-bom、不动既有 `LaunchedTvInitialFocus` 外层 `runCatching` 兜底（双重保险）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusSafeRequestTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `:tv-app:compileDebugUnitTestKotlin` 因 `tryRequestFocus` 未定义编译失败（13 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 结果为 293 tests / 290 passed / 3 failed——仅剩 A2 shared-element 三条预期红灯（`TvSharedPosterTransitionSpecTest`），属于待办任务 #19 的拥有范围，不阻塞本次崩溃修复。`./gradlew --no-daemon :tv-app:assembleDebug` 通过。审计测试 `TvNoBareLaunchedEffectFocusRequestAuditTest`（上一轮新增）持续绿（其匹配的是 `LaunchedEffect` 而非 `LaunchedTvInitialFocus`，切换到 `tryRequestFocus` 不破坏该不变量）。待手测：从首页冷启动、海报墙冷启动、电影/`18+`/电视剧详情冷启动、IPTV 冷启动、TV 配对页冷启动、长视频播放器冷启动都不应再触发 `FocusRequester is not initialized` FATAL；hover-exit 兜底仍只覆盖 hover-exit 一类异常未被本次改动稀释。

## 2026-05-21 23:14 +0800
- 进度：A2-pre 完成——为 A2 海报墙→详情 shared-element 过渡升级 TV 端 Compose 依赖基线。`android-tv-app/tv-app/build.gradle.kts` 把 `compose-bom` 从 `2024.06.00` 升到 `2024.10.01`（含稳定 `SharedTransitionLayout` / `LookaheadScope` API），`composeOptions.kotlinCompilerExtensionVersion` 从 `1.5.14` 升到 `1.5.15`；联动地把 `android-tv-app/build.gradle.kts` 的 Kotlin 插件版本从 `1.9.24` 升到 `1.9.25`（Compose 1.7 兼容性要求，错配会直接编译失败"This version (1.5.15) of the Compose Compiler requires Kotlin version 1.9.25"）。`kapt` 插件版本同步至 `1.9.25`。本提交仅升级 TV 工程，不动 `android-app/` 手机端 BOM；TV `versionCode` / `versionName` 暂不 bump，留到 A2 实现一起 bump。本次不删 deprecation 警告、不动业务源码。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/build.gradle.kts`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`./gradlew --no-daemon :tv-app:assembleDebug` 通过。无业务行为变化，体感不可见；下一步 A2 实现接入 `SharedTransitionLayout`。

## 2026-05-21 22:58 +0800
- 进度：落地 A 批第一项 A1——焦点放大改 spring 物理。`core/ui/TvFocus.kt` 抽出 `object TvFocusMotionTokens` 集中暴露 `ScaleDampingRatio = 0.8f`、`ScaleStiffness = 380f`（轻微回弹、中速）和 `SurfaceDampingRatio = 1f`、`SurfaceStiffness = 620f`（critically-damped、刚度高于缩放），把 `tvFocusableGlow` 和 `tvFocusableScaleOnly` 内的 `tween(durationMillis = 140)` 全量替换为 `spring(dampingRatio = ..., stiffness = ...)`，光晕背景淡入用更高刚度让 alpha 追上 scale 起步避免视觉错位。修饰器入参（`enabled` / `shape` / `focusedScale`）和 `onFocusChanged` / `graphicsLayer` / `background` / `focusable` 调用链不变，因此既有 `tvFocusableGlow` 调用点（首页菜单、海报墙、详情页、IPTV 频道行、播放器浮层、配对页、连接页等十余处）无需改动即可继承新动效；`TvFocusSafeSpec` 几何 token 不动，海报焦点安全留白不变。`TvFocusSpecTest` 新增 4 条不变量：源文断言不再使用 `tween(140)`、`tvFocusable*` 必须 `spring(`、参数必须来自 `TvFocusMotionTokens`；行为断言 `ScaleDampingRatio ∈ [0.7, 0.9]`、`ScaleStiffness ∈ [320, 440]`、`SurfaceDampingRatio ≥ 1`、`SurfaceStiffness > ScaleStiffness`。`CONTEXT.md` 在 `TV 焦点视觉语言` 之后新增 `TV 焦点动效物理` 词条，写明 token 名称、数值范围和"新增焦点反馈动画必须复用 token、不允许调用点硬编码"的强约束。TV 版本 `0.1.36`→`0.1.37`，`versionCode` 37→38。不在本轮范围：B 批 ken-burns/glow 双层、shared-element 详情转场、stagger 入场。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:compileDebugUnitTestKotlin` 因 `TvFocusMotionTokens` 未定义编译失败（5 处 Unresolved reference）；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿、`./gradlew --no-daemon :tv-app:assembleDebug` 通过。待手测：海报墙/详情页/菜单/IPTV/播放器浮层焦点切换的体感差异，确认放大有轻微回弹、背景光晕跟手不抖动；其余批次（A2 shared element、A3 stagger、A4 DPad 反馈、A5 token 收口）按计划下一批分别落地。

## 2026-05-21 20:58 +0800
- 进度：完成 TV `FocusRequester is not initialized` 同类裸用法全量清理。上一轮只覆盖 `TvCatalogScreen`，但 grep 出 7 处其他裸 `LaunchedEffect` + `requestFocus()` / 部分仅手工加了 `withFrameNanos { }` 但缺前缀过滤的同类风险点；本次按"全部切到 `LaunchedTvInitialFocus` 并删除冗余手动 `withFrameNanos`"统一处理：`TvPosterWallScreen.kt`（本次崩溃点，LazyVerticalGrid item 延迟组合）、`TvLongFormDetailScreen.kt`、`TvSeriesDetailScreen.kt`、`TvIptvScreen.kt`、`TvPairingScreen.kt`、`core/ui/LongFormVideoPlayer.kt`、`core/ui/SubtitlePicker.kt`。`LongFormVideoPlayer` 内 `pendingRootFocusRequest` / `pendingPlayPauseFocusRequest` 改用 `try { requestFocus() } finally { 清 pending }`，确保即使 helper 的 `runCatching` 兜下了 ISE，pending 标志也能被清，避免播放器状态机卡死。新增 `TvNoBareLaunchedEffectFocusRequestAuditTest` 审计测试：扫描 `src/main/java` 下所有 `.kt`，按行级括号深度追踪 `LaunchedEffect(...) {` 块体，若块体内出现 `.requestFocus(` 则 fail；同文件附 4 条 matcher 自测（命中、非命中、`LaunchedTvInitialFocus` 不误报、单行块体）。本提交不动 `CONTEXT.md`（"禁止业务 LaunchedEffect 内裸调"上轮已纳入约束）、不动 helper 自身逻辑、不动 hover-exit 兜底。TV 版本 `0.1.34`→`0.1.35`，`versionCode` 35→36。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvNoBareLaunchedEffectFocusRequestAuditTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`./gradlew :tv-app:testDebugUnitTest` 通过（含新增审计测试和 matcher 自测）；`./gradlew :tv-app:assembleDebug` 通过。审计测试现已锁住"任何新加的 `LaunchedEffect` + `requestFocus()` 都会在 CI 红"的回归面，未来任何同类裸用法都会编译期外被挡住。待手测：进入海报墙、电影/`18+` / 电视剧详情页、IPTV 频道页、TV 配对页、长视频播放器都不应再触发 `FocusRequester is not initialized` 崩溃；播放器控制条焦点反复 toggle 后 pending 标志不应卡死。

## 2026-05-21 20:42 +0800
- 进度：完成 TV 首页冷启动 `FocusRequester is not initialized` 修复落地与验证。新建 `core/ui/TvInitialFocusEffect.kt` 暴露共享 helper `LaunchedTvInitialFocus(vararg keys, block)`，内部先 `withFrameNanos { }` 等过一帧再 `runCatching { block() }`，精确过滤 `IllegalStateException` 且 `message` 以 `FocusRequester is not initialized` 开头的异常并重抛 `CancellationException`，其他异常照常抛出。`TvCatalogScreen.kt` 移除裸 `LaunchedEffect` 焦点请求并切换到 helper（焦点选择策略 `resolveTvCatalogInitialFocusTarget` 不动）。补两类纯 Kotlin 结构性测试：`TvInitialFocusEffectShapeTest` 锁住 helper 的 `withFrameNanos`/`runCatching`/`CancellationException`/前缀匹配/重抛/`@Composable vararg keys` 这几个不变量；`TvCatalogFocusPolicyTest` 增 `sectionItemCounts = listOf(0, 0)` 的 MENU 兜底用例覆盖 sections 非空但全 0 的场景。`CONTEXT.md` 把原“TV 首页初始焦点”一条扩为“TV 初始焦点请求约束”，新增 LazyColumn 延迟组合 × `LaunchedEffect` 帧时序竞态的说明并要求统一走 helper。TV 版本 `0.1.33`→`0.1.34`，`versionCode` 34→35。本提交不动 compose-bom（升级到 2024.10+ 是后续单独 PR）、不动 `TvMainActivity` hover-exit 兜底、不动 `selectMenu(Settings)` 的 `loading=false` 行为（已被 helper 的 `runCatching` 兜住）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusEffectShapeTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew :tv-app:testDebugUnitTest` 通过；`./gradlew :tv-app:assembleDebug` 通过；helper shape 测试和 resolver MENU 用例均为绿。待手测：冷启动 TV 首页焦点落到巨幅推荐、Settings 菜单切回时不再触发 FocusRequester 崩溃、空内容首页落焦左侧菜单。

## 2026-05-21 20:08 +0800
- 进度：完成海报墙发售时间排序 42P10 修复收尾验证；确认本次提交只纳入 `SearchVideosOrdered` SQL 重构、对应单测、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。后端无 Android 版本号需要 bump。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_search_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository/ -run 'TestSearchVideos' -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 无输出；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 20:02 +0800
- 进度：完成海报墙发售时间排序 42P10 红绿实现；把 `SearchVideosOrdered` 的 countSQL 和 selectSQL 抽成 `searchVideosCountSQL` 常量和 `searchVideosSelectSQL(orderClause)` helper，并用 `EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id = v.id AND LOWER(COALESCE(vt.tag,'')) LIKE $2)` 替换原来的 `LEFT JOIN video_tags + SELECT DISTINCT`。结构上消除 DISTINCT 之后，发售时间排序使用的 `NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST` 不再被 Postgres 42P10 拦截。语义不变：标题/描述/任一标签匹配即命中，且不会再因为多标签 JOIN 出现重复行，因此 COUNT 由 `COUNT(DISTINCT v.id)` 改为 `COUNT(*)`。`CONTEXT.md` 的 `TV 海报墙排序` 词条补充 EXISTS 子查询与 42P10 约束说明。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_search_test.go`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/repository/ -run 'TestSearchVideos' -count=1` 因未定义 `searchVideosCountSQL` / `searchVideosSelectSQL` 编译失败；实现后同命令通过。待执行后端全量单测、`go vet`、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 19:32 +0800
- 进度：完成 TV hover-exit 主 Looper 兜底收尾验证；确认本次提交只纳入 TV 主 Activity hover-exit 兜底扩展、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 通过；`./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 19:28 +0800
- 进度：完成 TV hover-exit 主 Looper 兜底红绿实现；`TvMainActivity.onCreate()` 安装主线程 `Handler.post { while(true) try { Looper.loop() } catch ... }` 外层异常拦截器，匹配到 Compose 平台层 hover-exit 异常后继续 loop，其它异常照常抛出。`shouldSwallowTvComposeHoverExitCrash` matcher 把方法名匹配放宽到 `sendHoverExitEvent` / `dispatchHoverEvent` 及其 `$lambda$` 合成方法名，覆盖 D8 生成的 lambda 调用帧。保留 `dispatchGenericMotionEvent` 同步兜底作为防御纵深。TV 版本更新为 `0.1.33` / `versionCode=34`，`CONTEXT.md` 在 `TV hover 输入兼容兜底` 词条补充主 Looper 调度路径与方法名匹配规则。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 因 matcher 不识别 `sendHoverExitEvent$lambda$5` / `dispatchHoverEvent$lambda$0` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 17:03 +0800
- 进度：完成 TV 工程编译边界瘦身收尾；确认本次提交只纳入 Gradle 编译排除边界、对应测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/*` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无输出；`git diff --check -- ...` 通过。

## 2026-05-21 17:02 +0800
- 进度：完成 TV 工程编译边界瘦身红绿实现；新增 Gradle 边界测试，红灯确认未声明排除清单。实现后通过 Kotlin sourceSets 排除手机端启动、手机首页/登录/Mine、短视频、图片合集、统一短视频播放器和相关测试源，保留 TV 主链路需要的连接页、详情 ViewModel、长视频播放器、网络模型和 IPTV。TV 版本更新到 `0.1.32` / `33`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过；待执行 TV 端全量单测和 Debug 构建。

## 2026-05-21 15:29 +0800
- 进度：完成 review 修复收尾；确认本次提交只纳入 TV 播放器退出确认、音轨/字幕弹窗焦点视觉、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/*` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvTrackPickerGlassPanelTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无输出；`git diff --check -- ...` 通过。

## 2026-05-21 15:27 +0800
- 进度：完成 review 修复的红绿闭环；红灯测试确认音轨/字幕弹窗仍有整圈硬描边、TV 控制条返回仍直接退出。实现后弹窗行改为蓝青背景提亮加细色条，TV 长视频和电视剧播放器控制条返回/退出均接入页面现有二次退出确认，TV 版本更新到 `0.1.31` / `32`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvTrackPickerGlassPanelTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；待执行 TV 端全量单测和 Debug 构建。

## 2026-05-21 15:23 +0800
- 进度：针对 `$grill-with-docs` TV App review 进行一轮修复；本轮优先处理直接影响 TV 使用体验的两项：音轨/字幕夜台玻璃弹窗去掉整圈硬描边焦点、播放器控制条返回/退出按钮复用播放器二次退出确认。TV 工程手机端遗留代码瘦身属于结构性清理，暂不与本轮 UI/行为修复混合。`CONTEXT.md` 将补充 `TV 播放器退出确认` 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、TV 播放器调用方、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 14:37 +0800
- 进度：完成 TV 长视频详情页操作组件收尾；返回按钮已统一为共享 `TvIconActionButton`，播放/收藏逻辑保持不变，沉浸式首屏继续不套用滚动页底部安全留白。确认本次提交只纳入长视频详情页、对应测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/*` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailActionSpecTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无输出；`git diff --check -- ...` 通过。

## 2026-05-21 14:36 +0800
- 进度：完成长视频详情页操作组件最小实现；红灯阶段新增 `TvLongFormDetailActionSpecTest` 后确认缺少共享 `TvIconActionButton` 会失败，随后将电影/`18+` 详情页返回按钮从手写圆形按钮切换到共享 TV 图标操作组件，保留播放和收藏原有逻辑与视觉，并将 TV 端版本更新到 `0.1.30` / `31`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailActionSpecTest'` 通过；待执行 TV 端全量单测和 Debug 构建。

## 2026-05-21 13:59 +0800
- 进度：完成 TV 电视剧详情页操作收尾验证；确认本次提交只纳入电视剧详情页返回图标操作统一、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 13:54 +0800
- 进度：完成 TV 电视剧详情页操作收尾红绿实现；新增电视剧详情页操作回归测试，红灯阶段确认返回操作尚未复用共享 `TvIconActionButton` 且仍残留默认 Material `IconButton` 导入；实现后返回按钮改为共享 TV 图标操作组件，播放、季选择和集选择继续使用 `tvFocusableGlow`，详情页布局、剧集选择逻辑和播放路由保持不变。TV 版本更新为 `0.1.29` / `versionCode=30`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest'` 因未复用共享图标操作和默认 `IconButton` 导入失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 12:43 +0800
- 进度：完成 TV 图标类操作焦点统一收尾验证；确认本次提交只纳入共享 TV 图标操作组件、TV 首页搜索清空、海报墙返回、长视频播放器控制按钮替换、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvIconActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 12:40 +0800
- 进度：完成 TV 图标类操作焦点统一红绿实现；新增共享 `TvIconActionButton`，红灯阶段确认共享组件缺失且 TV 首页搜索清空、海报墙返回、长视频播放器控制按钮仍导入默认 Material `IconButton`；实现后这三类图标操作均改用共享 TV 图标操作组件，IPTV 根焦点容器继续保留用于接收遥控按键。TV 版本更新为 `0.1.28` / `versionCode=29`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvIconActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvIconActionSpecTest'` 因缺少共享组件和目标文件仍导入默认 `IconButton` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 12:20 +0800
- 进度：完成 TV 连接服务器页优化收尾验证；确认本次提交只纳入连接页 TV 面板/焦点操作、连接页底部安全留白、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 12:18 +0800
- 进度：完成 TV 连接服务器页红绿实现；新增连接页体验回归测试，红灯阶段确认连接页仍使用默认 Material 区块和按钮；实现后自动嗅探、手动填写、历史地址区块改为 TV 深色 `Surface` 面板，重新扫描、测试并保存、使用/连接、删除改为共享 `tvFocusableGlow` 操作按钮；扫描 loading 保持小型行内状态，连接页滚动内容接入统一底部安全留白。TV 版本更新为 `0.1.27` / `versionCode=28`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest'` 因连接页缺少 TV 面板和共享焦点动作失败；底部留白红灯阶段 `--tests 'com.chee.videos.feature.tv.TvScrollableBottomPaddingTest'` 因连接页未使用统一底部留白失败；实现后两个定向测试通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 11:51 +0800
- 进度：完成 TV 配对/服务器连接与根启动体验优化收尾验证；确认本次提交只纳入配对页焦点按钮、根启动共享状态、配对连接体验测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvPairingConnectionExperienceTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 11:50 +0800
- 进度：完成 TV 配对/服务器连接与根启动体验红绿实现；新增配对连接体验回归测试，红灯阶段确认配对页仍裸用 `.focusable()` 且根启动仍直接使用默认进度环；实现后配对页两个操作改为共享 `tvFocusableGlow` 焦点按钮，根启动改用 `TvPageLoadingState`，服务器自动嗅探保持小型行内 loading。TV 版本更新为 `0.1.26` / `versionCode=27`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvPairingConnectionExperienceTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvPairingConnectionExperienceTest'` 因配对页焦点和根启动 loading 约束失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 11:33 +0800
- 进度：完成 TV 状态反馈语言优化最终提交范围检查；确认本次提交只纳入共享 TV 状态组件、重点页面状态接入、重试入口与相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、TV 状态反馈/重点页面/ViewModel 相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过；待执行暂存区复核与提交。

## 2026-05-21 11:32 +0800
- 进度：完成 TV 状态反馈语言优化收尾验证；确认本次提交只纳入共享 TV 状态组件、重点页面状态接入、重试入口与相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、TV 状态反馈/重点页面/ViewModel 相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 11:28 +0800
- 进度：完成 TV 状态反馈语言红绿实现；新增共享 `TvStateFeedback` 组件，提供页面级 loading、行内 loading、空态和错误重试态；TV 首页/搜索、海报墙、电影/`18+` 详情、电视剧详情、IPTV 状态层、电影播放器和电视剧播放器加载/错误占位改用共享状态组件。为首页、电视剧详情和电视剧播放器补充可聚焦错误态需要的 `retry()` 入口，并补 ViewModel 重试回归测试。TV 版本更新为 `0.1.25` / `versionCode=26`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、TV 状态反馈/重点页面/ViewModel 相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvStateFeedbackSpecTest' --tests 'com.chee.videos.feature.tv.TvStateFeedbackUsageTest'` 因缺少共享状态组件和页面接入失败；实现后定向状态组件、页面使用和相关 ViewModel 测试通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 11:03 +0800
- 进度：完成 TV 第一阶段焦点视觉优化最终提交范围检查；确认本次提交只纳入 7 个文件：TV 焦点视觉语言、首页海报卡焦点迁移、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过；待执行暂存区复核与提交。

## 2026-05-21 11:02 +0800
- 进度：完成 TV 第一阶段焦点视觉优化收尾验证；确认本次提交只纳入 TV 焦点视觉语言、首页海报卡焦点迁移、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。并行执行 TV 单测和构建时曾触发 Kotlin 增量编译缓存竞争，顺序重跑后单测通过。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 10:56 +0800
- 进度：完成 TV 第一阶段焦点视觉红绿实现；新增焦点规格测试，验证全局 TV 焦点不再使用旧粉红硬描边，改为蓝青柔和背景提亮；首页海报/查看更多卡从默认 glow 切换为只放大焦点语言，保留按钮、菜单、筛选项、频道行通过共享 `tvFocusableGlow()` 获得蓝青焦点反馈。TV 版本更新为 `0.1.24` / `versionCode=25`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvFocusSpecTest' --tests 'com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest'` 因旧粉红硬描边和首页海报卡使用默认 glow 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 10:48 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV App 整体优化；已确认第一阶段优先做“遥控器体验与视觉一致性”，即统一焦点反馈、可点击元素形态、加载/空态、页面密度和安全留白，不在同一阶段重排首页信息架构、改播放内核或新增内容类型。`CONTEXT.md` 已记录该阶段边界。
- 影响文件：`CONTEXT.md`、`plan.md`；后续预计影响 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、TV 单测与 `android-tv-app/tv-app/build.gradle.kts`
- 验证：待继续确认焦点视觉策略、覆盖页面范围、红灯测试口径，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 23:05 +0800
- 进度：完成上传图片 WebP 编码不可用修复的收尾检查；确认本次提交只纳入图片上传降级、WebP 编码不可用错误标记、相关测试、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`internal/services/image.go`、`internal/services/image_test.go`、`pkg/ffmpeg/ffmpeg.go`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过；待执行提交范围复核与提交。

## 2026-05-20 23:04 +0800
- 进度：完成上传图片 WebP 编码不可用降级修复；`ffmpeg.ConvertToWebP` 在 ffmpeg 与 `cwebp` 都不可用时返回可识别的 `ErrWebPEncodingUnavailable`，图片上传遇到该错误时保留原始 JPEG/PNG 作为处理图并继续入库，动态变体沿用处理图格式，避免访问阶段再次强制 WebP。`CONTEXT.md` 记录图片上传处理图和变体格式约定。
- 影响文件：`internal/services/image.go`、`internal/services/image_test.go`、`pkg/ffmpeg/ffmpeg.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestSaveFromLocalPathKeepsOriginalWhenWebPEncodingUnavailable|TestImageVariantFormatUsesStoredFormatWhenOriginalWasKept' -count=1` 通过；`go test ./pkg/ffmpeg -run 'TestIsEncoderUnavailableOutput' -count=1` 通过；`go test ./internal/services ./pkg/ffmpeg -count=1` 通过；`go test ./... -count=1` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-20 19:25 +0800
- 进度：补充 TV 海报墙排序最终收尾记录；确认乱码检查与 diff 空白检查已通过，提交范围将精确限制在排序后端接口、TV 端排序 UI/状态、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`internal/handlers/tv.go`、`internal/services/tv_auth.go`、`internal/services/tv_catalog_wall_test.go`、`internal/repository/app_repository.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过；待执行提交范围复核与提交。

## 2026-05-20 19:24 +0800
- 进度：完成 TV 海报墙排序收尾验证；确认本次提交只纳入海报墙排序后端接口、TV 端排序 UI/状态、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`internal/handlers/tv.go`、`internal/services/tv_auth.go`、`internal/services/tv_catalog_wall_test.go`、`internal/repository/app_repository.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-20 19:19 +0800
- 进度：完成 TV 海报墙排序红绿实现；后端 `/api/v1/tv/catalog` 新增 `sort_by=added|release` 和 `sort_order=asc|desc`，电影/18+ 按视频入库时间或 metadata 发售日期排序，电视剧按关联可播放视频最新入库时间或首播日期排序，缺失日期排最后。TV 端海报墙顶部新增排序字段和方向切换按钮，切换后清空旧列表并重新加载第一页。TV 版本更新为 `0.1.23` / `versionCode=24`，`CONTEXT.md` 记录 TV 海报墙排序语义。
- 影响文件：`internal/handlers/tv.go`、`internal/services/tv_auth.go`、`internal/services/tv_catalog_wall_test.go`、`internal/repository/app_repository.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段后端排序规范测试因缺少排序模型/SQL 子句失败，TV 定向测试因缺少 `changeSort`、排序状态和接口参数失败；实现后 `go test ./internal/services -run 'TestNormalizeTVCatalogWallSort|TestTVCatalogWallSortOrderClause' -count=1` 通过，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvPosterWallViewModelTest'` 通过，`go test ./internal/services -count=1` 通过，`go test ./internal/handlers ./internal/repository -count=1` 通过。待执行 TV App 全量验证、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:46 +0800
- 进度：完成 TV hover-exit 闪退兜底收尾验证；确认本次提交只纳入 TV 主 Activity 输入异常兜底、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:43 +0800
- 进度：完成 TV hover-exit 闪退兜底红绿实现；`TvMainActivity.dispatchGenericMotionEvent()` 捕获 Compose 平台层 `The ACTION_HOVER_EXIT event was not cleared.` 异常，并通过异常消息与 `AndroidComposeView` 堆栈双重匹配后才吞掉，其他输入异常继续抛出。TV 版本更新为 `0.1.22` / `versionCode=23`，`CONTEXT.md` 记录 TV hover 输入兼容兜底。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 因缺少 hover-exit 兜底判断函数失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:31 +0800
- 进度：完成 TV 服务器自动嗅探 loading 收尾验证；确认本次提交只纳入连接服务器页扫描 loading 尺寸、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:28 +0800
- 进度：完成 TV 服务器自动嗅探 loading 红绿实现；扫描状态改为 14dp 小型行内进度环并使用 2dp 线宽，避免只限制高度导致默认进度环视觉过大。TV 版本更新为 `0.1.21` / `versionCode=22`，`CONTEXT.md` 记录服务器自动嗅探状态应使用小型行内 loading。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest'` 因缺少小型行内 loading 规格失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:11 +0800
- 进度：完成 TV 海报墙 9:16 海报卡收尾验证；确认本次提交只纳入 TV 海报墙卡片视觉、无描边焦点修饰器、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:08 +0800
- 进度：完成 TV 海报墙 9:16 海报卡红绿实现；卡片图片区改为 9:16 且图片贴边显示，标题条紧贴图片底部并使用深色背景，卡片焦点改为仅放大/阴影的无描边焦点修饰器。AV 海报墙显示沿用后端 `title` 作为番号；TV 版本更新为 `0.1.20` / `versionCode=21`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvPosterWallCardContentTest' --tests 'com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest'` 因缺少 `showDescription` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 17:23 +0800
- 进度：完成 TV 滚动内容底部安全留白收尾验证；确认本次提交只纳入 TV 可滚动内容底部留白、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLayoutSpec.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 17:22 +0800
- 进度：完成 TV 滚动内容底部安全留白红绿实现；新增共享 `TvLayoutSpec.scrollBottomSafePaddingDp=56`，TV 首页/搜索、海报墙、电视剧详情页、IPTV 频道列表、剧集选择底部抽屉统一使用该底部留白。播放器画面和沉浸式详情首屏保持不变。TV 版本更新为 `0.1.19` / `versionCode=20`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLayoutSpec.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvScrollableBottomPaddingTest'` 因缺少首页底部留白规格失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 16:39 +0800
- 进度：完成 TV 播放器连按合并跳转收尾验证；确认本次提交只纳入 TV 播放器快进/快退 debounce、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 16:23 +0800
- 进度：完成 TV 播放器连按合并跳转红绿实现；新增 pending seek 纯逻辑，快进/快退按键每次都即时刷新累计目标和预览反馈，但实际 `seekTo` 延迟约 300ms 且只提交最后一次目标。TV 版本更新为 `0.1.18` / `versionCode=19`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 因缺少 `TvPendingStepSeekUpdate` 和 `resolveTvPendingStepSeek` 编译失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 16:08 +0800
- 进度：完成 TV App 播放设置收尾验证；确认本次提交只纳入 TV 播放步长设置、播放器接入、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 15:58 +0800
- 进度：完成 TV App 播放设置红绿实现；DataStore 新增全局 TV 快进/快退步长，设置页新增“播放设置”分组和 5/10/15/20/30 秒预设，电影/电视剧 TV 长视频播放器读取该设置，左右键单次按步长跳转，重复按按 3 倍步长跳转。TV 版本更新为 `0.1.17` / `versionCode=18`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段定向测试因缺少步长设置对象、DataStore 字段、仓储接口、ViewModel 状态和播放器按键参数编译失败；实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest' --tests 'com.chee.videos.core.data.AppPreferencesStoreTest' --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest' --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` 通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 15:15 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV App 播放设置；代码确认设置页当前是 `tv-home` 内的“账户与设备”面板，长视频播放器快进/快退硬编码为 10 秒，遥控器重复按放大到 30 秒。已确认新增全局 `快进/快退步长`，同时作用于左右键，预设 5/10/15/20/30 秒，默认 10 秒，重复按按步长倍数加速。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 DataStore、设置面板和播放器按键策略红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 13:47 +0800
- 进度：完成 TV App 根页面二次退出收尾验证；确认本次提交只纳入 TV 壳层根退出确认、策略测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 13:46 +0800
- 进度：完成 TV App 根页面二次退出红绿实现；新增 `tv-home` 根退出确认策略和 2 秒确认窗口，第一次返回显示“再按一次退出”，第二次返回调用 Activity 退出。TV 版本更新为 `0.1.16` / `versionCode=17`，`CONTEXT.md` 追加 `TV 根退出确认` 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvShellAppBackPolicyTest'` 因缺少根退出确认策略编译失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 20:48 +0800
- 进度：完成 TV 电影详情本地横幅和轨道面板收尾验证；确认本次提交只纳入后端电影本地 backdrop variant、TV 详情背景解析、TV 轨道面板确认键/焦点视觉、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`internal/handlers/video_source.go`、`internal/handlers/video_source_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 20:47 +0800
- 进度：完成 TV 电影详情本地横幅与轨道面板红灯测试/核心实现；后端 `videos/:id/thumbnail?variant=backdrop` 支持电影本地 `backdrop.jpg`，TV 电影详情优先使用该本地 variant，轨道行改为显式处理遥控确认键并移除全局粉红焦点边框，TV 版本更新为 `0.1.15` / `versionCode=16`。
- 影响文件：`internal/handlers/video_source.go`、`internal/handlers/video_source_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段后端定向测试因缺少 `chooseVideoThumbnailVariantPath` 失败；TV 定向测试因旧背景解析和旧焦点样式失败。实现后 `go test ./internal/handlers -run 'TestChooseMovieBackdropVariantPathUsesOnlyLocalDownloadedBackdrop|TestChooseMovieBackdropVariantPathRejectsTMDBRelativePath' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest' --tests 'com.chee.videos.core.ui.TvTrackPickerGlassPanelTest'` 通过。待执行更宽验证、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 20:46 +0800
- 进度：确认 TV 电影详情横幅必须使用已下载到本地的电影横向背景，不直接使用 TMDB 原始相对图路径；推荐后端扩展视频图片本地访问路由（如 thumbnail variant）暴露本地 `backdrop.jpg`，TV 端通过 API URL 使用该本地图片。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行后端/TV App 定向测试、文档乱码检查和 diff 检查。

## 2026-05-19 20:44 +0800
- 进度：完成影视演员信息刮削收尾验证；确认本次提交只纳入电影/电视剧 TMDB 演员资料与本地头像入库相关后端、测试、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`CONTEXT.md`、`internal/services/scraper.go`、`internal/services/scraper_test.go`、`internal/repository/actor_repository.go`、`internal/repository/actor_repository_test.go`、`internal/queue/scrape_tasks_test.go`、`plan.md`
- 验证：`go test ./internal/services ./internal/repository -count=1` 通过；`go test ./internal/handlers ./internal/services ./internal/repository -count=1` 通过；`go vet ./internal/handlers ./internal/services ./internal/repository` 通过；`go test ./... -count=1` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 20:43 +0800
- 进度：完成影视演员信息刮削红灯测试与核心实现；新增电影 TMDB credits 全量演员资料/本地头像入库测试、已有头像/备注不覆盖测试和仓储合并 SQL 约束测试。后端新增 `UpsertScrapedActorProfile`，电影/电视剧演员同步改为按 TMDB person id 补齐资料并下载本地头像，AV 演员链路保持不变。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_test.go`、`internal/repository/actor_repository.go`、`internal/repository/actor_repository_test.go`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestSyncMovieActorsUpsertsFullTMDBProfilesAndLocalAvatarsWithoutLimit|TestSyncMovieActorsDoesNotOverrideExistingAvatarOrNotes' -count=1` 因未执行演员资料 upsert 失败；实现后同命令通过。待执行后端更宽验证、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 20:42 +0800
- 进度：确认影视演员入库不设数量上限，按 TMDB credits 返回的演员集合处理；实现时仍只在落库阶段执行，单个演员资料或头像失败不应阻断整部电影/剧集刮削落库。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:37 +0800
- 进度：完成电影重新刮削缓存修复的收尾检查；确认本次提交范围只包含电影重新刮削绕过缓存相关后端、管理端 helper、测试、`CONTEXT.md` 和 `plan.md`，不暂存无关 skill 删除或未跟踪目录。
- 影响文件：`CONTEXT.md`、`internal/services/scraper.go`、`internal/services/scraper_av_strategy.go`、`internal/services/scraper_test.go`、`internal/handlers/admin_scrape.go`、`admin-web/src/views/scrapePreview.helpers.js`、`admin-web/src/views/scrapePreview.helpers.spec.js`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/services -count=1` 通过；`cd admin-web && npm run test -- --run` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 20:36 +0800
- 进度：完成电影重新刮削缓存修复；管理端电影查询预览默认发送 `bypass_cache=true`，后端电影预览在该标记下跳过已有 metadata 复用和短期候选缓存，电视剧与 AV 预览逻辑不变。确认本次提交只纳入电影重新刮削缓存相关文件，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`CONTEXT.md`、`internal/services/scraper.go`、`internal/services/scraper_av_strategy.go`、`internal/services/scraper_test.go`、`internal/handlers/admin_scrape.go`、`admin-web/src/views/scrapePreview.helpers.js`、`admin-web/src/views/scrapePreview.helpers.spec.js`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/services -count=1` 通过；`cd admin-web && npm run test -- --run` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-19 20:10 +0800
- 进度：完成电影横向背景刮削第二阶段验证；确认第二提交只纳入后端电影横向背景预览/确认/自动刮削、管理端通用刮削横向背景字段和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_test.go`、`internal/handlers/admin_scrape.go`、`internal/queue/scrape_tasks.go`、`admin-web/src/views/ScrapePreview.vue`、`admin-web/src/views/scrapePreview.helpers.js`、`admin-web/src/views/scrapePreview.helpers.spec.js`、`plan.md`
- 验证：`go test ./internal/services -run 'TestPreviewMovieUsesChineseLanguageAndEnglishFallback|TestConfirmMovieDownloadsLocalBackdrop' -count=1` 通过；`go test ./internal/services -count=1` 通过；`cd admin-web && npm run test -- scrapePreview.helpers.spec.js` 通过；`cd admin-web && npm run test -- --run` 通过；`cd admin-web && npm run build` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 20:05 +0800
- 进度：完成电影横向背景刮削红灯测试和核心实现；后端电影预览候选新增 `backdrop_path`，自动/手动电影确认下载 TMDB 横向背景到本地 `videos/{video_id}/backdrop.jpg` 并写入 metadata，管理端通用刮削确认 payload 增加 `backdrop_url`，电影编辑表单显示横向背景输入。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_test.go`、`internal/handlers/admin_scrape.go`、`internal/queue/scrape_tasks.go`、`admin-web/src/views/ScrapePreview.vue`、`admin-web/src/views/scrapePreview.helpers.js`、`admin-web/src/views/scrapePreview.helpers.spec.js`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestPreviewMovieUsesChineseLanguageAndEnglishFallback|TestConfirmMovieDownloadsLocalBackdrop' -count=1` 因预览缺少 `backdrop_path`、确认未下载背景失败；实现后同命令通过。红灯阶段 `cd admin-web && npm run test -- scrapePreview.helpers.spec.js` 因 confirm payload 缺少 `backdrop_url` 失败；实现后同命令通过。待执行后端/管理端更宽验证、乱码检查、diff 检查和第二提交范围检查。

## 2026-05-19 19:47 +0800
- 进度：完成 TV 播放器切轨和夜台玻璃面板第一阶段验证；确认第一提交只纳入 TV 播放器/轨道选择 UI、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/AudioTrackSelectionTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.AudioTrackSelectionTest' --tests 'com.chee.videos.core.ui.TvTrackPickerGlassPanelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-19 19:42 +0800
- 进度：完成 TV 播放器红灯测试和核心实现；新增音轨选择运行时 selected 轨优先展示、自动选择文案和夜台玻璃面板静态约束测试，修复 TV 音轨列表项信息层级、切轨诊断日志、字幕/音轨共用夜台玻璃面板，并更新 TV 版本为 `0.1.14` / `versionCode=15`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/AudioTrackSelectionTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.AudioTrackSelectionTest' --tests 'com.chee.videos.core.ui.TvTrackPickerGlassPanelTest'` 因缺少 `detail`、自动选择和夜台玻璃面板失败；实现后同命令通过。待执行 TV 全量单测、Debug 构建、乱码检查、diff 检查和第一提交范围检查。

## 2026-05-19 19:19 +0800
- 进度：确认播放器改动范围只覆盖 `android-tv-app`；手机端 `android-app` 暂不套用夜台玻璃面板，也不在本次同步修改相似播放器实现。
- 影响文件：`plan.md`
- 验证：待实现阶段执行 TV App 相关验证，提交范围不纳入手机端。

## 2026-05-19 19:17 +0800
- 进度：确认本次实现拆成两个小提交：第一个提交修 TV 播放器运行时切轨和字幕/音轨夜台玻璃面板并更新 TV 版本号；第二个提交修电影横向背景刮削与本地化保存。
- 影响文件：`plan.md`
- 验证：待实现阶段分别执行 TV 播放器相关验证、后端/管理端刮削相关验证和提交范围检查。

## 2026-05-19 19:15 +0800
- 进度：确认本次实现涉及 TV App 功能修改，交付时必须同步更新 `android-tv-app/tv-app/build.gradle.kts` 版本号，按仓库规则 `versionCode +1`、`versionName` patch 位 `+1`。
- 影响文件：`plan.md`
- 验证：待实现阶段执行 TV 定向/全量验证和文档检查。

## 2026-05-19 18:57 +0800
- 进度：确认 TV 电影音轨问题的验收口径为“运行时切轨”：选择音轨后应立即切换当前播放音频，不重启播放、不重新进入详情页，并且音轨列表应反映当前播放音轨；已补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 17:58 +0800
- 进度：完成提交前复查；确认本次只纳入 Git 忽略规则、技术沉淀、计划记录和已跟踪 Python 字节码移出索引，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`.gitignore`、`CONTEXT.md`、`plan.md`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/core.cpython-314.pyc`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/design_system.cpython-314.pyc`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/search.cpython-314.pyc`
- 验证：`git status --short --untracked-files=all android-app/app/release android-tv-app/tv-app/release` 无输出；`git ls-files -ci --exclude-standard` 无输出；`git check-ignore -v ...` 确认 Android release 输出、`.pyc` 和 `.run/server.log` 被忽略，skill `references` 路径未被忽略；`git diff --check -- .gitignore CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' .gitignore CONTEXT.md plan.md` 无命中。

## 2026-05-19 17:55 +0800
- 进度：完成 Git 忽略规则清理；根 `.gitignore` 新增 Python/工具缓存、Go 输出、Android TV 本地产物、Android APK/AAB 与 release 打包目录规则，并将原 `references/`、`release` 改为根目录锚定，避免误忽略 skill 参考文档；已从索引移除 `.codex/skills/ui-ux-pro-max/scripts/__pycache__/*.pyc`，保留本地文件但不再纳入 Git。
- 影响文件：`.gitignore`、`CONTEXT.md`、`plan.md`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/core.cpython-314.pyc`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/design_system.cpython-314.pyc`、`.codex/skills/ui-ux-pro-max/scripts/__pycache__/search.cpython-314.pyc`
- 验证：`git status --short --untracked-files=all android-app/app/release android-tv-app/tv-app/release` 无输出；`git ls-files -ci --exclude-standard` 无输出；`git check-ignore -v ...` 确认 Android release 输出、`.pyc` 和 `.run/server.log` 被忽略，skill `references` 路径未被忽略；`git diff --check -- .gitignore CONTEXT.md plan.md` 通过。待执行乱码检查和提交范围检查。

## 2026-05-19 13:49 +0800
- 进度：完成 TV App 启动崩溃排查修复最终验证；确认本次提交只纳入 TV 首页空内容初始焦点兜底、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。当前 `adb devices` 无在线设备，因此未能直接抓取真机 `logcat` 或做安装启动实测。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusPolicyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- ...` 通过；`adb devices` 无在线设备。

## 2026-05-19 12:04 +0800
- 进度：完成 TV 电影/18+ 详情页沉浸式改版最终验证；确认本次提交只纳入 TV 长视频详情页沉浸式首屏、展示模型与测试、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 12:00 +0800
- 进度：完成 TV 电影/18+ 详情页沉浸式改版核心实现；`TvPresentation.kt` 新增沉浸式 hero 的年份信息、演员头像模型、海报兜底标记与收藏文案；`TvLongFormDetailScreen.kt` 改为全屏背景加底部半透明信息面板，播放按钮保持默认焦点，收藏按钮复用 `DetailViewModel.toggleFavorite()`，移除更多信息和下方信息卡片；TV 版本更新为 `0.1.12` / `versionCode=13`，`CONTEXT.md` 记录 TV 沉浸式详情首屏约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest'` 通过。待执行 TV 全量单测、Debug 构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 11:56 +0800
- 进度：完成 TV 详情页沉浸式改版红灯测试；新增测试约束年份/时长/标签信息行、收藏/取消收藏按钮、演员头像与占位、无横幅时海报模糊兜底，以及源码中不出现分享、更多信息和下方信息卡片。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest'` 因 `usesPosterAsBackdropFallback`、`actors` 等详情页 hero 字段尚未实现而失败。

## 2026-05-19 11:28 +0800
- 进度：完成 TV 首页菜单确认键修复最终验证；确认本次提交只纳入 TV 首页侧边菜单单焦点修复、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 11:27 +0800
- 进度：完成 TV 首页侧边菜单遥控确认键核心修复；根因是 `TvHomeSideMenuButton` 在 `tvFocusableGlow()` 已提供焦点目标后又叠加 `.focusable()`，造成菜单按钮重复焦点目标，遥控确认键可能第一次只落到内部焦点层、第二次才触发点击。已删除重复 `.focusable()`，TV 版本更新为 `0.1.11` / `versionCode=12`，`CONTEXT.md` 记录 TV 菜单按钮单焦点目标约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 因侧边菜单按钮仍包含重复 `.focusable()` 失败；实现后同命令通过。待执行 TV 全量单测、Debug 构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 10:45 +0800
- 进度：完成电影手动刮削入口最终验证；确认本次提交只纳入管理端电影手动刮削入口、路由 helper 单测、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（Vite 仅提示 chunk size 警告）；`go test ./internal/handlers -run 'TestAdminScrape|TestShouldEnqueueAdminScrapeConfirmTranscode' -count=1` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/videoList.helpers.js admin-web/src/views/videoList.helpers.spec.js admin-web/src/views/VideoList.vue` 通过。

## 2026-05-19 10:44 +0800
- 进度：完成电影手动刮削入口核心实现；`VideoList.vue` 在电影详情抽屉播放预览操作区新增“电影手动刮削”按钮，点击关闭抽屉并跳转通用刮削页；`buildMovieManualScrapeRoute` 生成 `/scrape` 路由 query，并从 `metadata.release_date` 或 `metadata.tmdb.release_date` 解析年份；`CONTEXT.md` 补充手动刮削术语和复用接口约定。
- 影响文件：`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/VideoList.vue`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 因 `buildMovieManualScrapeRoute` 尚不存在失败；实现后同命令通过。待执行管理端构建、后端刮削回归测试、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 10:32 +0800
- 进度：完成 TV APK ARM ABI 拆包瘦身最终验证；确认 Debug/Release 均只输出 `armeabi-v7a` 与 `arm64-v8a` APK，未生成 x86/x86_64 或 universal APK；Release 未签名 ARM APK 体积分别约 `42M` 与 `45M`，均低于 `< 90M` 验收阈值；每个 Release APK 只包含对应 ABI 的 `libvlc.so`。本次提交只纳入 TV Gradle/ProGuard 打包配置、打包配置测试、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleRelease` 通过；`find tv-app/build/outputs/apk -type f -name '*.apk' -maxdepth 5 | sort` 仅列出 Debug/Release ARM ABI APK（另有 androidTest debug APK）；`du -h tv-app/build/outputs/apk/debug/*.apk tv-app/build/outputs/apk/release/*.apk` 显示 Debug 约 `63M`/`67M`、Release 约 `42M`/`45M`；`unzip -l ...release-unsigned.apk | rg 'lib/.*/libvlc\\.so|lib/.*/libvlccore\\.so'` 分别只命中对应 ABI 的 `libvlc.so`；`unzip -l ...armeabi-v7a-release-unsigned.apk | rg 'lib/(arm64-v8a|x86|x86_64)/'` 无命中；`unzip -l ...arm64-v8a-release-unsigned.apk | rg 'lib/(armeabi-v7a|x86|x86_64)/'` 无命中；待最终重跑乱码检查、diff 检查并提交。

## 2026-05-19 10:30 +0800
- 进度：完成 TV APK ARM ABI 拆包核心实现；TV 版本更新为 `0.1.10` / `versionCode=11`，`build.gradle.kts` 启用 `armeabi-v7a` 与 `arm64-v8a` ABI split、关闭 universal APK，Release 开启 R8 和资源瘦身；`proguard-rules.pro` 保留 LibVLC API 面；`CONTEXT.md` 记录 APK 按 ARM ABI 分发且继续保留 VLC 的长期约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过。待执行 TV 全量单测、Debug/Release 构建、APK ABI 内容检查、体积检查、乱码检查和提交范围检查。

## 2026-05-19 10:28 +0800
- 进度：完成 TV APK 打包配置红灯测试；新增静态单测约束 TV App 启用 ARM ABI split、关闭 universal APK、Release 开启 R8/资源瘦身且继续保留 LibVLC 依赖。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 因缺少 ABI split 配置和 Release shrink 配置失败。

## 2026-05-19 09:48 +0800
- 进度：完成 TV IPTV 频道列表与顶部提示优化最终验证；确认本次提交只纳入 TV IPTV UI/交互、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptv*'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 09:44 +0800
- 进度：完成 TV IPTV 顶部临时提示和频道列表初始定位实现；顶部频道信息改为左上角紧凑提示，当前频道变化后显示 3 秒，频道列表打开或异常状态隐藏；频道列表打开时按当前频道计算初始 first visible index，首次渲染跳过动画，后续焦点上下移动继续动画跟随。TV 版本更新为 `0.1.9` / `versionCode=10`，`CONTEXT.md` 更新 IPTV 提示和列表定位约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 通过。待执行计划内 IPTV 定向、TV 全量单测、Debug 构建和乱码检查。

## 2026-05-19 09:42 +0800
- 进度：完成 TV IPTV 顶部提示和频道列表定位红灯测试；新增纯逻辑测试覆盖频道列表初始 first visible index 和顶部临时提示可见性，新增静态回归测试约束顶部提示非全宽常驻、3 秒隐藏、列表打开使用初始定位且焦点移动保留动画。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少 `resolveIptvChannelListInitialFirstVisibleItemIndex` 和 `shouldShowIptvChannelHint` 编译失败。

## 2026-05-19 09:11 +0800
- 进度：完成 AV 大背景最终验证；确认本次提交只纳入后端 TV DTO 映射、TV 详情展示、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVCatalogWallVideoItems|Test.*AV' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest' --tests 'com.chee.videos.feature.tv.TvCatalogFeaturedContentTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 09:09 +0800
- 进度：完成 AV 大背景核心实现；后端 TV 首页 AV DTO 的 `backdrop_url` 按原始横幅优先和固定 fallback 顺序解析，`poster_url` 保持 `thumbnail_path`；TV 详情页 `videoType=av` 时顶部背景改用详情 metadata 的原始海报，左侧小海报仍使用 `thumbnail_path`；TV 版本更新为 `0.1.8` / `versionCode=9`，`CONTEXT.md` 记录 AV 大背景与竖卡海报分工。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestBuildTVHomeVideoFromListItem' -count=1` 因 AV `backdrop_url` 仍为缩略图失败；红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest'` 因 AV 详情背景未使用原始海报失败；实现后上述两个命令通过。待执行计划内完整验证。

## 2026-05-18 21:56 +0800
- 进度：完成 TV IPTV 台标和频道列表滚动修复最终验证；确认本次提交只纳入 TV IPTV UI/交互、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptv*'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv` 无命中；`git diff --check -- ...` 通过。

## 2026-05-18 21:49 +0800
- 进度：完成 TV IPTV 台标和频道列表滚动核心实现；播放页顶部和频道列表行改为使用 Coil `AsyncImage` 渲染 `logoUrl`，缺失或加载失败时回退 TV 图标；频道列表新增按频道 id 解析 `LazyColumn` item index 的 helper，并在列表打开和焦点移动后自动滚动到对应频道。TV 版本更新为 `0.1.7` / `versionCode=8`，`CONTEXT.md` 记录台标和列表滚动约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少频道列表 item index helper 编译失败；实现初版后同命令因索引偏移断言失败；修正后同命令通过。待执行计划内全量 TV 单测、Debug 构建和乱码检查。

## 2026-05-18 21:27 +0800
- 进度：完成 IPTV 音频专用源过滤；后端解析 M3U 时跳过明显音频分组、音频命名、`/audio/`/`_audio/` 路径和音频文件后缀，TV 端对 API 返回的旧频道数据也执行同样过滤并只在可播放视频频道中切台。TV 版本更新为 `0.1.6` / `versionCode=7`，`CONTEXT.md` 记录音频源过滤规则。
- 影响文件：`internal/services/iptv.go`、`internal/services/iptv_test.go`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestParseM3UPlaylistSkipsAudioOnlyEntries' -count=1` 因音频源未跳过失败；红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 因缺少过滤 helper 编译失败；实现后 `go test ./internal/services -run 'TestParseM3UPlaylist|TestBuildIPTV|TestIPTVService' -count=1` 通过，TV 同一定向命令通过；`go test ./internal/services ./internal/handlers -run 'Test.*IPTV|TestRegisterIncludesIPTVRoutes' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。待执行乱码检查和提交范围检查。

## 2026-05-18 21:07 +0800
- 进度：完成 IPTV LibVLC 输出层和诊断增强；LibVLC `attachViews` 改为 TextureView 输出，新增 `TvIptv` 日志记录 event、vout、视频轨/音频轨数量、当前视频轨 codec/分辨率，TV 版本更新为 `0.1.5` / `versionCode=6`，并在 `CONTEXT.md` 记录后续无画面排查依据。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少 TextureView 绑定和诊断日志失败；实现后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。并行复跑时曾触发 Kotlin/Kapt 增量缓存竞争，执行 `cd android-tv-app && ./gradlew --stop && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过。待执行乱码检查和提交范围检查。

## 2026-05-18 20:55 +0800
- 进度：完成 IPTV 播放器兼容性实现；TV App IPTV 播放页从 Media3 `PlayerView` 单独切换为 LibVLC `VLCVideoLayout`，播放直播源时关闭硬解并配置网络缓存，避免设备硬解不支持视频轨时只出声音；其他长视频 Media3 播放路径保持不变。TV 版本更新为 `0.1.4` / `versionCode=5`，`CONTEXT.md` 更新 IPTV 播放器兼容性沉淀。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_iptv_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackDependencyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackDependencyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少 `org.videolan.libvlc.LibVLC` 失败；实现后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。并行复跑时曾触发 Kotlin/Kapt 增量缓存竞争，执行 `cd android-tv-app && ./gradlew --stop && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过。待执行乱码检查和提交范围检查。

## 2026-05-18 20:34 +0800
- 进度：完成 TV App IPTV 有声音无画面修复；新增 IPTV 专用 Media3 `PlayerView` XML 布局并指定 `surface_type="texture_view"`，播放页改为 inflate 该布局，TV 版本更新为 `0.1.3` / `versionCode=4`，并在 `CONTEXT.md` 记录 Compose + IPTV 播放页的 TextureView 约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_iptv_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因布局缺失失败；修复后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 串行重跑通过；目标文件 Python 乱码扫描无命中。并行跑单测和构建时曾遇到 Hilt 增量产物竞争，串行重跑已通过。

## 2026-05-18 20:24 +0800
- 进度：完成 TV App IPTV 点击闪退修复；新增 `media3-exoplayer-hls` 依赖，补充 HLS 工厂编译期回归测试，TV 版本更新为 `0.1.2` / `versionCode=3`，并在 `CONTEXT.md` 记录 M3U8/HLS 播放依赖约定。确认本次不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackDependencyTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackDependencyTest'` 因缺少 `androidx.media3.exoplayer.hls` 编译失败；修复后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；目标文件 Python 乱码扫描无命中。

## 2026-05-18 19:35 +0800
- 进度：完成 IPTV v1 最终验证与收尾；确认本次提交只纳入后端 IPTV、Admin Web IPTV 管理页、TV App IPTV 播放页、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入主工作区既有 `.codex/skills/*` 无关变更。
- 影响文件：`migrations/0020_iptv_playlist.*.sql`、`internal/models/iptv.go`、`internal/repository/iptv_repository.go`、`internal/services/iptv.go`、`internal/handlers/iptv.go`、`internal/handlers/router.go`、`admin-web/src/*`、`android-tv-app/tv-app/*`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./... -count=1` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过（仅既有 AGP compileSdk/native strip 警告）；Python 源码乱码扫描无命中；`git diff --check` 通过。

## 2026-05-18 19:31 +0800
- 进度：完成 IPTV v1 三端核心实现；后端新增单全局 M3U 播放列表迁移、宽松解析、Admin 管理接口和 TV 频道接口；Admin Web 新增 `IPTV 管理` 页面；TV App 新增 `IPTV` 一级菜单、全屏直连播放页、频道分组列表、上下键循环换台和右键/返回键策略，并将 TV 版本更新为 `0.1.1` / `versionCode=2`。已补充 `CONTEXT.md` IPTV 术语与接口约定。
- 影响文件：`migrations/0020_iptv_playlist.*.sql`、`internal/models/iptv.go`、`internal/repository/iptv_repository.go`、`internal/services/iptv.go`、`internal/handlers/iptv.go`、`internal/handlers/router.go`、`admin-web/src/*`、`android-tv-app/tv-app/*`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers ./internal/repository -run 'Test.*IPTV|TestRegisterIncludesIPTVRoutes|TestIPTVPlaylistMigration' -count=1` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过。待执行完整 Go 验证、TV Debug 构建和乱码检查。

## 2026-05-18 18:40 +0800
- 进度：完成新增开发约定；根级 `AGENTS.md` 已要求 App 功能修改同步更新对应 App 版本号，并要求每次功能更新追加 `CONTEXT.md` 技术沉淀；手机端与 TV 端模块级 `AGENTS.md` 已写明各自版本文件和递增规则；`CONTEXT.md` 已新增技术沉淀约定。
- 影响文件：`AGENTS.md`、`android-app/AGENTS.md`、`android-tv-app/AGENTS.md`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' AGENTS.md android-app/AGENTS.md android-tv-app/AGENTS.md CONTEXT.md plan.md` 无命中；文档约定变更无需构建/单测；确认既有 `.codex/skills/*` 工作区变更不是本任务改动，不纳入提交。

## 2026-05-18 12:44 +0800
- 进度：完成 TV App 左侧分类首页重设计最终验证；后端保留 `/api/v1/tv/home` 未传 `kind` 的旧字段兼容，新增类型化 `kind/featured/recent_watching/recent_updates`；TV 端完成左侧一级菜单、分类页、搜索页、设置面板、`18+` 文案统一和海报墙标题编码修正。确认既有 `.codex/skills/*` 工作区变更不是本任务改动，不纳入提交。
- 影响文件：`CONTEXT.md`、`internal/models/app.go`、`internal/services/tv.go`、`internal/services/tv_auth.go`、`internal/handlers/tv.go`、`internal/services/tv_service_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、相关测试、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers -run 'Test.*TVHome|Test.*TVCatalog' -count=1` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app internal` 无命中。

## 2026-05-18 12:38 +0800
- 进度：完成 TV 左侧分类首页核心实现；后端 `/api/v1/tv/home` 新增可选 `kind=tv|movie|av` 并返回 `kind`、`featured`、`recent_watching`、`recent_updates`，旧 payload 字段保留；TV 端新增左侧菜单模型、`18+ -> av` 请求映射、搜索/设置菜单页、类型化首页分区和左栏/内容焦点策略，Shell 右上设置菜单改为首页右侧设置面板。
- 影响文件：`internal/models/app.go`、`internal/services/tv.go`、`internal/services/tv_auth.go`、`internal/handlers/tv.go`、`internal/services/tv_service_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、相关测试、`plan.md`
- 验证：红灯阶段 Go 因缺少 `buildTypedTVHomePayload` 失败，TV 因缺少菜单模型、`kind` 请求和分区 helper 失败；实现后 `go test ./internal/services -run 'TestBuildTypedTVHomePayload|TestBuildTVHomePayload' -count=1` 通过，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 通过。待执行计划内完整验证。

## 2026-05-18 10:18 +0800
- 进度：完成仓库级开发流程 skill 生成；`.codex/skills/repo-dev-workflow` 已沉淀本仓库计划记录、TDD 红灯、模块化验证、中文/编码、收尾提交范围控制等 `plan.md` 历史经验，并保留既有无关 `.codex/skills/*` 工作区变更不纳入本次提交。
- 影响文件：`.codex/skills/repo-dev-workflow/SKILL.md`、`.codex/skills/repo-dev-workflow/agents/openai.yaml`、`plan.md`
- 验证：使用临时 venv 安装 PyYAML 后运行 `python3 /Users/chee/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/repo-dev-workflow` 通过；`rg -n $'\uFFFD' .codex/skills/repo-dev-workflow plan.md AGENTS.md` 无命中；`git status --short` 已确认本次只暂存新 skill 与 `plan.md`。

## 2026-05-17 21:25 +0800
- 进度：完成“保留长视频音轨并支持 TV 端音轨选择”的完整验证；确认本次提交只纳入后端转码/元数据、TV 音轨选择/偏好及对应测试、`plan.md`，不纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关测试文件、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProbe|TestResolveProbe|TestBuildTranscodePlan' -v` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 21:20 +0800
- 进度：完成后端和 TV 音轨核心实现；ffmpeg 转码显式映射主视频与全部音频，音频继续转 AAC 但不再强制双声道；ffprobe/转码元数据新增音轨数量。TV 端新增音轨偏好存储、Repository/ViewModel 读写、Media3 当前音轨解析、音轨选择参数应用，以及复用字幕居中弹窗的“音轨”选择入口。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关测试文件、`plan.md`
- 验证：红灯阶段后端因 `AudioTrackCount`/`-map` 期望缺失失败，TV 端因音轨偏好与音轨选择 API 缺失失败；实现后 `go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProbe|TestResolveProbe' -v` 通过，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.data.AppPreferencesStoreTest' --tests 'com.chee.videos.core.ui.AudioTrackSelectionTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.detail.DetailViewModelTest'` 通过。待执行计划内完整验证。

## 2026-05-17 20:50 +0800
- 进度：完成 TV App 播放记录与断点续播修复的全量验证；确认本次只修改 `android-tv-app` 播放历史相关代码、测试和 `plan.md`，未纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*History*' --tests 'com.chee.videos.feature.detail.*DetailViewModel*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 20:49 +0800
- 进度：完成 TV 播放历史核心实现；新增播放历史策略 helper，电视剧播放器增加 15 秒定时上报与 `ON_PAUSE` 补报，电影/AV 长视频播放器增加详情进度续播、15 秒定时上报、`ON_PAUSE` 与销毁补报，`DetailViewModel` 增加历史上报入口并过滤无效输入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`plan.md`
- 验证：红灯阶段定向测试因缺少 `TvPlaybackHistoryPolicy` 与 `DetailViewModel.reportHistory` 编译失败；实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*History*' --tests 'com.chee.videos.feature.detail.*DetailViewModel*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` 通过。待执行 TV 全量单测与 Debug 构建。

## 2026-05-17 13:09 +0800
- 进度：完成 TV 首页设置按钮焦点边界修复；设置按钮仍只在 `tv-home` 显示，按左/下稳定回到首页搜索框，按右/上由 `FocusRequester.Cancel` 拦截越界焦点搜索，未改变海报墙、详情页、播放器页 Back 与焦点行为。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellSettingsFocusPolicyTest.kt`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.*Settings*'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 10:27 +0800
- 进度：完成 TV App AV 板块恢复的最终验证；确认本次提交只包含后端 TV 聚合、Android TV AV 展示/路由/文案和对应回归测试，未纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`internal/services/tv_catalog_wall_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`plan.md`
- 验证：`rg -n "ExcludesAV|ExcludeAV|filtersAv|WithoutAv|does not promote av|NormalizesAvToMovie|搜索剧名|相关剧集|部剧集" ...` 无命中；`go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVSearchPayload|TestBuildTVCatalogWallPayload|TestBuildTVCatalogWallVideoItems' -v` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 10:26 +0800
- 进度：完成 TV AV 链路恢复实现；后端 TV 首页/搜索/海报墙重新查询并返回 AV，TV 首页重新读取 AV shelf、保留 AV 搜索结果与继续观看，AV 海报墙和长视频路由保留 `videoType=av`，详情/精选/焦点/文案恢复 AV 分支。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`internal/services/tv_catalog_wall_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`plan.md`
- 验证：红灯阶段后端 `TestBuildTVSearchPayloadIncludesAVContent` 因 AV 未加入结果失败，TV 定向测试因缺少 AV 参数/焦点/路由分支编译失败；实现后 `go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVSearchPayload|TestBuildTVCatalogWallPayload|TestBuildTVCatalogWallVideoItems' -v` 通过，TV feature 定向单测通过。待执行完整后端与 TV 验证。

## 2026-05-17 08:58 +0800
- 进度：完成长视频 4K 码率上限收尾验证，确认本次只影响 movie/episode 的 4K 上限，AV 与 1080p 既有断言保持通过。
- 影响文件：`internal/services/transcode.go`、`internal/services/transcode_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestDecideVideoBitrate|TestBuildTranscodePlan' -v` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过。

## 2026-05-17 08:57 +0800
- 进度：完成长视频 4K 码率上限微调；电影/电视剧 longform 4K 上限从 `12000k` 收紧到 `10000k`，1080p 上限、CRF、HEVC/AVC 分流与 AV 策略保持不变。
- 影响文件：`internal/services/transcode.go`、`internal/services/transcode_test.go`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestDecideVideoBitrate|TestBuildTranscodePlan' -v` 因 4K longform 仍返回 `12000` 失败；实现后同命令通过。待执行 `go test ./internal/services ./internal/handlers -count=1`。

## 2026-05-16 21:47 +0800
- 进度：完成手机端搜索播放页点赞/收藏能力；搜索播放器进入当前视频后预取详情，右侧动作栏显示喜欢/收藏激活态，点击后复用现有点赞/收藏接口并更新本地 `userState`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchViewModel.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/test/java/com/chee/videos/feature/shortsearch/ShortSearchViewModelStateTest.kt`、`plan.md`
- 验证：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.feature.shortsearch.ShortSearchViewModelStateTest` 通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-16 12:09 +0800
- 进度：完成 TV APP 电视剧目录去除 AV 内容的完整验证，确认 TV 专属 AV shelf/文案静态检查无命中，并准备提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`plan.md`
- 验证：`rg -n "AV 精选|全部 AV|av-shelf|AV_ITEM|继续播放 AV|\\\"AV\\\"" android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv internal/services/tv_auth.go` 无命中；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-16 11:47 +0800
- 进度：完成手机端 UI 圆角全局收敛后的静态检查、单测和 Debug 构建验证，准备提交本次手机端 UI 改动。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/AppChromeDensitySpecTest.kt`、`plan.md`
- 验证：`rg "RoundedCornerShape\\((1[3-9]|2[0-9])\\.dp|topStart = (1[3-9]|2[0-9])\\.dp" android-app/app/src/main/java/com/chee/videos` 无命中；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-16 11:09 +0800
- 进度：收尾检查演员页界面文案，移除空状态英文状态词并补齐作品类型中文标签；复跑完整后端与 Android 验证。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/repository ./internal/services -count=1`、`go vet ./...`、`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest`、`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 均通过。

## 2026-05-25 18:08 +0800
- 进度：完成 `tasks/2026-05-25-tv-long-form-libvlc-migration` 的代码层收尾和验证收口。TV 长视频偏好已从 trackId 迁到 `language + type` 持久化，音轨/字幕恢复仍按当前 media 的临时 track id 运行；后端 ASS 原文落盘与内嵌抽取增加了安全清洗。`CONTEXT.md` 已同步长期约定，任务相关定向测试和全量验证已通过。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/{data,model,repository,ui}/**`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/{detail,tv}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/{data,player,ui}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/{detail,tv}/**`、`android-tv-app/tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerLibVlcTest.kt`、`internal/services/subtitle.go`、`internal/services/subtitle_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`
- 验证：`go test ./internal/services ./pkg/ffmpeg -count=1` 通过；`go test ./... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' ...` 无输出。

## 2026-05-25 17:51 +0800
- 进度：补充 `tasks/2026-05-25-tv-long-form-libvlc-migration` 的 TV 模拟器 instrumentation 验证。Gradle `:tv-app:connectedDebugAndroidTest` 曾因下载 `com.android.tools.utp:android-test-plugin-result-listener-gradle:31.5.2` 时 TLS 握手失败，改为使用已构建 APK 直接安装并通过 `adb shell am instrument` 执行，绕开 UTP 下载链路。
- 影响文件：`plan.md`
- 验证：`adb -s emulator-5554 install -r android-tv-app/tv-app/build/outputs/apk/debug/tv-app-arm64-v8a-debug.apk` 成功；`adb -s emulator-5554 install -r android-tv-app/tv-app/build/outputs/apk/androidTest/debug/tv-app-debug-androidTest.apk` 成功；`adb -s emulator-5554 shell am instrument -w com.chee.videos.tv.test/androidx.test.runner.AndroidJUnitRunner` 通过，执行 `LongFormVideoPlayerFocusTest` 2 个用例与 `LongFormVideoPlayerLibVlcTest` 1 个用例，共 3 个测试，结果 `OK (3 tests)`。仍需用户确认 review.md §1 的 ASS 复杂样式、遥控器回归、性能与服务端上传手测全部通过后，才能按仓库规则创建 `DONE.md`。

## 2026-05-25 17:54 +0800
- 进度：按 `tasks/2026-05-25-tv-long-form-libvlc-migration/review.md` 重新完成自动化准入审计。确认后端字幕测试、TV 单测、TV Debug 构建、androidTest 编译、Media3 残留扫描、版本号、CONTEXT 术语、ADR 与 admin 字幕上传文案扫描均已闭合；DataStore 持久化只写 `tv_subtitle_language_preferences` / `tv_audio_language_preferences`，并在读写时清除旧 `tv_subtitle_preferences` / `tv_audio_preferences`，运行时仍保留当前 media 临时 track id 用于 UI 选中和 LibVLC `setAudioTrack`。
- 影响文件：`plan.md`
- 验证：`go test ./internal/services -run TestSubtitle -count=1` 通过；`go test ./pkg/ffmpeg -run 'Test.*Subtitle|TestBuildExtractSubtitleToAssArgs' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过；`rg -n 'media3-' android-tv-app/tv-app/build.gradle.kts` 无输出；`rg -n 'androidx\.media3|ExoPlayer|PlayerView|MediaItem|CaptionStyleCompat|Player\.Listener|PlaybackParameters' android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` 无输出；`git diff --check` 通过；乱码扫描无输出。剩余未闭合项仍是 review.md §1 真机/模拟器手测场景与用户验收确认，未创建 `DONE.md`。

## 2026-05-25 19:01 +0800
- 进度：用户确认 `tasks/2026-05-25-tv-long-form-libvlc-migration` 实测通过，按仓库规则补充 `DONE.md` 完成标记。完成标记记录相关提交、自动化验证、模拟器 instrumentation、review.md §1 手测确认与交付范围。
- 影响文件：`tasks/2026-05-25-tv-long-form-libvlc-migration/DONE.md`、`plan.md`
- 验证：待执行 `git diff --check`、`rg -n $'\uFFFD' tasks/2026-05-25-tv-long-form-libvlc-migration/DONE.md plan.md` 与工作区范围检查后提交。

## 2026-05-25 20:30 +0800
- 进度：grill-with-docs 收尾后产出 `tasks/2026-05-25-tv-long-form-focus-guarding/` 与 `tasks/2026-05-25-tv-long-form-track-preference-recovery/` 两个独立任务的 prd.md / implement.md / review.md 三件套，分别覆盖：①续播卡 dispose 触发的 [[TV 长视频焦点真空]] 与 controls auto-hide / picker dismiss / back confirm / playerError 等同源场景的修复方向 A'（续播卡内嵌 + overlay 跃迁观察 + root 自我兜底）；②音轨/字幕偏好不记的 F1/F2/F3（VLC `Playing` gate + type-only preference fallback + audio LaunchedEffect 状态回灌）。
- 影响文件：`tasks/2026-05-25-tv-long-form-focus-guarding/prd.md|implement.md|review.md`、`tasks/2026-05-25-tv-long-form-track-preference-recovery/prd.md|implement.md|review.md`、`plan.md`
- 验证：grill 阶段无 Gradle/测试任务，纯文档；待真正实施时按各自 review.md §0 准入条件执行。

## 2026-05-25 22:15 +0800
- 进度：执行 `tasks/2026-05-25-tv-long-form-focus-guarding/` 全部 6 个里程碑。①新增 `LongFormPlayerFocusGuard.kt`（`PlayerFocusGuardInput` 六字段聚合 + `shouldReclaimRootFocus` 纯函数）与 `LongFormPlayerFocusGuardTest.kt` 13 个测试用例；②`LongFormVideoPlayer` 增加 `resumePromptVisible` / `resumePromptSlot: @Composable BoxScope.() -> Unit` / `backConfirmPromptVisible` / `playerErrorVisible` 参数，插入聚合 `LaunchedEffect` 与 root Box `.onFocusChanged` 双兜底，并在根 Box 内渲染 slot；③`TvLongFormPlayerScreen` 与 `TvSeriesPlayerScreen` 把 `TvResumePromptCard` 从外层 Box 兄弟位置迁入 slot，并透传 `backConfirmPromptVisible` / `playerErrorVisible`；④新增 `LongFormVideoPlayerSpecTest.kt` 源文 audit 8 个用例；⑤`tv-app/build.gradle.kts` 版本号升到 `0.1.69 / 69`，连带修了 `TvLongFormVlcSpecTest.tvBuildFile_removesMedia3Dependencies` 把硬钉的 `versionCode = 68` 改为下限 `>=68` 非脆弱断言；⑥CONTEXT.md 在 TV 段追加 3 条术语（[[TV 长视频焦点真空]] / [[LongFormVideoPlayer focus 兜底]] / [[续播提示卡内嵌位置]]）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormPlayerFocusGuard.kt`（新）、`LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormPlayerFocusGuardTest.kt`（新）、`LongFormVideoPlayerSpecTest.kt`（新）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` BUILD SUCCESSFUL；新单测共 21 个用例全绿（`LongFormPlayerFocusGuardTest` 13 + `LongFormVideoPlayerSpecTest` 8）；待执行 `git diff --check`、乱码扫描、`TvResumePromptCard\(` 出现面 rg、提交并按 review.md §1 真机/模拟器手测 R1~R10 后才允许补 `DONE.md`。

## 2026-05-25 23:21 +0800
- 进度：执行 `tasks/2026-05-25-tv-long-form-track-preference-recovery/` 全部 6 个里程碑，覆盖 F1+F2+F3 三条互相独立的故障 surface。①**F2 type-only fallback**：`resolveLongFormTrackByLanguage` / `resolveSelectedSubtitleTrackByPreference` 在 language 空但 type 非空时按 type 直接匹配第一条同 type 的 track；之前会 `return null` 让 `isDefault=true` 但无 languageCode 的字幕/音轨偏好永远丢。新增 `TvLongFormTrackSelectionFallbackTest` 7 + `LongFormSubtitlePreferenceFallbackTest` 8 + `LongFormAudioPreferenceFallbackTest` 6 用例。②**F1 VLC Playing gate**：`LongFormVideoPlayer` 自带 `isVlcPlaying` state，audio LaunchedEffect 加入 `if (!isVlcPlaying || audioTracks.isEmpty()) return` gate；`TvLongFormPlayerScreen` / `TvSeriesPlayerScreen` 各自维护独立的 `isVlcPlaying` state，`applyLongFormMediaSource` 签名收缩为 `(libVLC, mediaPlayer, sourceUrl)` 不再带 subtitle/baseUrl；字幕注入改为 `LaunchedEffect(isVlcPlaying, selectedSubtitleTrackId, ...)` 触发 `mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, url, true)` 并用 `appliedSubtitleSlaveUrl` 做幂等。`resolveLongFormPlayerUpdate` 签名同步收缩，不再因 subtitle 变化触发 setMedia。③**F3 audio 状态回灌**：`onSelectAudioTrack` 新增第三参数 `isUserAction: Boolean`，picker 选轨传 `true`、audio LaunchedEffect 把 `resolvedSelection` 回灌父级时传 `false`；`TvLongFormPlayerScreen` / `TvSeriesPlayerScreen` 仅在 `isUserAction=true` 时写 DataStore，避免 save-loop。④新增 `LongFormVlcPlayingGateSpecTest.kt` 8 个源文 audit 用例。⑤`tv-app/build.gradle.kts` 版本号升到 `0.1.70 / 70`。⑥CONTEXT.md 在 TV 段追加 3 条术语（[[VLC Playing gate]] / [[Type-only preference fallback]] / [[Audio LaunchedEffect 状态回灌]]）；`SubtitleSelectionTest` 旧测试用例同步迁移到新签名。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt`、`com/chee/videos/core/ui/LongFormSubtitleSupport.kt`、`com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormTrackSelectionFallbackTest.kt`（新）、`com/chee/videos/core/ui/LongFormSubtitlePreferenceFallbackTest.kt`（新）、`com/chee/videos/core/ui/LongFormAudioPreferenceFallbackTest.kt`（新）、`com/chee/videos/core/ui/LongFormVlcPlayingGateSpecTest.kt`（新）、`com/chee/videos/core/ui/SubtitleSelectionTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` BUILD SUCCESSFUL；新单测共 29 个用例全绿（fallback 21 + spec audit 8）；待执行 `git diff --check`、乱码扫描、提交并按 review.md §1 真机/模拟器手测 A1~A7 后才允许补 `DONE.md`。

## 2026-06-19 11:36 +0800
- 进度：完成 TV 单片长视频播放器软重试语义审查并修正一处边界回退。已确认并修复 `BACK` 取消当前重试时不应把播放会话回退成未开始；现在取消只终止本次准备、清理错误并回写 `已取消重试` 轻提示，继续沿用已出现首帧后的承接播放态。同步把 TV 版本递增到 `0.1.121` / `versionCode=121`，并补了取消语义的源码断言。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：受限于当前 sandbox，Gradle 先后卡在默认 `~/.gradle` 锁文件权限、wrapper 网络下载和 daemon 端口绑定；已改用本地缓存拷贝到 `/private/tmp/codex-gradle-home` 继续，但完整 `:tv-app:testDebugUnitTest` 仍未跑完。后续需在可绑定本地端口的环境里复跑 `cd android-tv-app && GRADLE_USER_HOME=/private/tmp/codex-gradle-home ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest`，再补 `assembleDebug`、`git diff --check` 和乱码扫描。

## 2026-06-25 15:25 +0800
- 进度：完成 TV 短视频单条播放失败非阻塞化改造。原行为：单条播放失败时弹阻塞卡片（重试/返回首页），根 `onPreviewKeyEvent` 顶部 `if (playbackErrorMessage != null) return false` 短路吞掉所有遥控键，导致上/下无法切条、用户被卡在失败视频上。改造后（仅单条播放失败路径，首屏加载失败/空态路径保留原阻塞卡片）：①删除顶部短路；②上/下照常 `movePrevious`/`moveNext`，切条后由既有 `LaunchedEffect(currentVideoId)`（`:238`）清 `playbackErrorMessage`；③OK/中键/PLAY_PAUSE 在失败态执行重试体（`playbackErrorMessage=null; renderedVideoId=null; hasEndedAtCurrentVideo=false; playbackRetryNonce+=1`，与旧重试按钮逐行一致，`return true` 防落暂停逻辑），非失败态保持原暂停逻辑；④左/右/REWIND/FF 在失败态 `return true` 静默消费不 seek；⑤BACK 退出不变；⑥把 `if (!playbackErrorMessage.isNullOrBlank()) { Box(0.56f scrim) { TvShortFeedProblemState } }` 替换为非可聚焦居中 chip「播放失败」（`AnimatedVisibility(visible = playbackErrorMessage != null)`，复用暂停指示器 `Surface(CircleShape, AppChrome.Surface alpha 0.82)` + `Row(Icon(Refresh)+Text)` 样式，去全屏遮罩，不夺焦不吞键）；⑦`onPlayerError` 置 `playbackErrorMessage` 时同步 `showCenterIndicator=false; centerIndicatorHideJob?.cancel()`，覆盖「先暂停（700ms 自动隐藏窗口内）随后失败」的叠显场景，保证失败提示干净接管中央位。不新增任何状态变量/定时器/Job/常量，不改任何 `LaunchedEffect`。`TvShortFeedProblemState`/`TvShortFeedStateButton` 保留（首屏失败路径仍用）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`（版本 `0.1.123 / 123` → `0.1.124 / 124`）、`CONTEXT.md`（修订 `TV 短视频单条失败留在当前页` 为非阻塞模型 + `TV 短视频保留中央播放暂停提示` 加 OK 语义分态交叉引用）、`docs/superpowers/specs/2026-06-25-tv-short-feed-nonblocking-failure-design.md`（新设计规格）、`plan.md`。
- 验证：`cd android-tv-app && ./gradlew :tv-app:compileDebugKotlin :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（exit 0），既有 `TvShortFeedScreenSpecTest` / `TvShortFeedViewModelTest` 全绿、无需改测试；子代理两轮独立评审（首轮发现 BLOCKER 版本号未升 + IMPORTANT 暂停/失败提示 700ms 叠显，均已修复并复审 CLEAN）；待真机/模拟器手测：失败态上/下切条、OK 重试、左/右静默、BACK 退出、首屏失败仍弹阻塞卡片、正常 OK 暂停样式不变。

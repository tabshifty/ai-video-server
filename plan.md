# 2026-06-26 09:51 +0800
- 进度：压缩包导入目录条目 `video_type` 约束问题已修复并验证通过。目录分支现在显式写入 `short`，插入前通过 `archiveImportVideoTypeOrDefault` 兜底空值与大小写，避免 `archive_import_files` 因空字符串触发 check constraint；同时补了目录回归与纯逻辑回归。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services/archive_import.go internal/services/archive_import_test.go` 无命中。

# 2026-06-25 22:52 +0800
- 进度：修复压缩包导入在目录条目上落空 `video_type` 导致的数据库约束失败。已定位到 `scanArchiveEntries` 目录分支漏填字段，计划同时补插入前的防御性归一化和回归测试。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/services -run 'TestScanArchiveEntries|TestArchiveFileTitleForScannedFile' -count=1`、`git diff --check`、乱码扫描。

# 2026-06-25 22:43 +0800
- 进度：已将 `tv-short-feed-cover-poster` 快进合并到本地 `master`，并推送到 `deploy` 远端；部署机 hook 已完成本次更新收口。
- 影响文件：`plan.md`
- 验证：`git merge --ff-only tv-short-feed-cover-poster` 通过；`git push deploy master` 通过，hook 输出 `RESTART_GO=0 REBUILD_FRONTEND=0` 且无失败。

# 2026-06-25 22:38 +0800
- 进度：准备将 `tv-short-feed-cover-poster` 合并到 `master`，然后推送到部署机 `deploy` 远端；当前先核对分叉关系、部署约定和工作区状态，确认不会把无关改动带进去。
- 影响文件：`plan.md`
- 验证：待执行 `git merge --ff-only tv-short-feed-cover-poster`、`git push deploy master`、`git status --short`、`git diff --stat`、乱码扫描。

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

## 2026-06-24 23:19 +0800
- 进度：开始落地 ZIP 压缩包条目名编码兼容实现。当前工作重点是修通 `archive_import` 的自动判码、`needs_encoding` 重试入口、批次编码追溯字段，以及管理端重试表单联动；先修到可编译，再补服务/前端验证。
- 影响文件：`internal/services/archive_import.go`、`internal/handlers/admin_archive_import.go`、`internal/models/models.go`、`internal/repository/migrations_test.go`、`migrations/0028_archive_import_batch_encoding.up.sql`、`migrations/0028_archive_import_batch_encoding.down.sql`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/services ./internal/handlers ./internal/repository -count=1`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-24 22:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包最终编码的审计边界。已确认批次一旦通过人工指定编码模式成功解包，就必须保留“最终采用的编码模式”用于回看与排障；同时核实了当前 `archive_import_batches` 既没有 `encoding_mode` 字段，也没有批次级 `metadata`，现状不足以直接承载该信息。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:52 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包重试解包入口。已确认密码与条目名编码模式属于同一个纠偏表单，一次重试同时提交 `password` 和 `encoding_mode`，避免管理员在“待密码”和“待编码”之间来回切入口；该边界已同步补入 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:49 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包条目名编码模式的作用域。已确认首期显式编码模式只作用于 `zip`，`rar/7z` 维持既有外部解压器行为，不共享 `utf8 / gbk` 人工纠偏入口；该边界已同步补入 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:48 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包条目名编码模式的入口边界。已确认编码模式不放上传表单，首次上传默认 `auto`；只有批次进入 `needs_encoding` 后，管理员才通过“重试解包”显式选择 `utf8` 或 `gbk` 做人工纠偏。该语义已同步补入 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:46 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包编码待确认状态。已确认自动判码失败或管理员手动纠偏时应进入独立 `needs_encoding` 可重试状态，与 `needs_password` 并列而不是混入 `failed`；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:44 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入的解包重试边界。已确认现有 `retry-extract` 不应继续只承载密码，而要扩成支持显式条目名编码模式；首期模式收口为 `auto / utf8 / gbk`，并已同步补充 `CONTEXT.md` 的 `压缩包解包编码模式` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:42 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包 ZIP 条目名的判码顺序。已确认不能把“未标 UTF-8”直接等同为 `GBK/CP936`；条目名先按合法 UTF-8 使用，仅在不是合法 UTF-8 时才回退 `GBK/CP936`，并已同步补充 `CONTEXT.md` 的 `压缩包条目名判码顺序` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:41 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入的相对路径落库边界。已确认 `relative_path` 与管理员可见文件名统一保存解码后的中文路径文本，不存原始字节串转义；该术语已补入 `CONTEXT.md`，避免后续误把路径审计做成原始编码展示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:40 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入的中文条目名兼容边界。已确认“必须解决中文 ZIP 条目名”在术语上应具体化为兼容明确列出的常见编码，而不是泛化成“所有中文都支持”；当前兼容集收口为 `UTF-8 + GBK/CP936`，并已同步更新 `CONTEXT.md` 的 `压缩包条目名编码` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:36 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入的条目名编码边界。已确认当前问题不是“中文字符本身不支持”，而是 ZIP 条目名编码可能与 UTF-8/系统可创建路径不兼容；同时确认优先保留 `压缩包原始相对路径保留` 与 `压缩包原目录结构保留` 契约，不自动猜码、不自动改名，并已同步补充 `CONTEXT.md` 的 `压缩包条目名编码` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 22:28 +0800
- 进度：完成压缩包导入 UTF-8 落库报错修复。服务层现在会先把外部解压器 stderr 与批次/文件失败原因归一化为合法 UTF-8，再写入 `archive_import_batches.last_error` / `archive_import_files.reason`；新增纯单测锁定非法字节替换行为，并在 `CONTEXT.md` 沉淀“压缩包错误文本 UTF-8 落库”兼容约束。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestNormalizeArchive|TestSanitizeArchive' -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services/archive_import.go internal/services/archive_import_test.go` 无命中。

## 2026-06-24 22:25 +0800
- 进度：开始修复压缩包导入的 UTF-8 落库报错。已定位为解压器 stderr 或文件系统错误文本可能携带原始非 UTF-8 字节，当前会原样写入 `archive_import_batches.last_error` / `archive_import_files.reason`，触发 Postgres `SQLSTATE 22021`。计划先补服务层纯逻辑测试锁定错误文本归一化，再实现压缩包错误文本 UTF-8 清洗，并同步更新 `CONTEXT.md`。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/services -run 'TestNormalizeArchive|TestSanitizeArchive' -count=1`、`git diff --check`、`rg -n $'\uFFFD' CONTEXT.md plan.md`

## 2026-06-24 17:56 +0800
- 进度：完成 TV 短视频迁移收尾校验。当前仅准备纳入 TV 端短视频全屏播放页、首页 `短视频` action-only 入口、独立路由/仓库接口、相关单测、TV 版本号，以及 `CONTEXT.md`、`android-tv-app/AGENTS.md`、`plan.md` 的长期文档更新；未纳入其它模块改动。
- 影响文件：`android-tv-app/AGENTS.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/*`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/AGENTS.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/main/java/com/chee/videos/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/tv` 无命中。

## 2026-06-24 17:55 +0800
- 进度：完成 TV 短视频迁移实现收尾。新增 `tv/shorts` 独立全屏短视频页与 ViewModel，首页左侧菜单在 `IPTV` 下方新增 `短视频` 入口并保持 action-only 语义；补充 `fetchShortFeed(pageSize, excludeIds)`、返回焦点与导航回退测试，同时校正 TV 模块 AGENTS 中已过时的“仅长视频”约束。
- 影响文件：`android-tv-app/AGENTS.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvShortFeedScreenSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvShortFeedViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShortFeedNavigationSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待执行最终提交与提交后状态确认。

## 2026-06-24 17:13 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频入口的菜单视觉语义。已确认 `短视频` 与 `IPTV` 一样只是启动动作，不作为持久选中菜单态；返回首页后焦点回到 `短视频` 按钮，但原内容分类高亮保持不变。该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 17:12 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频末尾观感。已确认最后一条且无更多内容时，视频自然播完后停在最后一帧并视为暂停态，不回开头、不跳第一页；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 17:08 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频入口可见性。已确认左侧菜单里的 `短视频` 入口始终显示，不因当前是否有短视频而动态隐藏，空态统一交给短视频页内处理；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 17:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频返回焦点。已确认从短视频页 `BACK` 返回首页后，焦点稳定回到左侧 `短视频` 入口按钮，不回内容区旧焦点；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 17:01 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频 feed 语义。已确认同一次进入短视频页的会话里，后续补货尽量排除已经出现过的内容，只有服务端候选池过小时才接受重复；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:55 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频切条手感。已确认上/下键只做离散的一次一条切换，不支持长按连续切条；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:53 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频 seek 手感。已确认支持长按左右键按全局步长连续 seek，进度条在连续 seek 期间保持显示，松手后停止；若 seek 前已暂停，松手后仍保持暂停。该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:52 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频加载反馈。已确认首帧准备、切条准备和缓冲等待时保留极简 loading 指示，不带文案、不抢焦点，用于区分加载与暂停/卡住；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:48 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频中央提示边界。已确认切条成功仍不提示，但 `OK` 切换播放/暂停时保留中央短提示，用于区分“已暂停”和“卡住”；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:44 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频 seek 与暂停关系。已确认暂停后按左/右 seek 只改变位置，不自动恢复播放，`OK` 继续是唯一的播放/暂停切换开关；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:42 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频首屏异常态。已确认首屏加载失败和空态都留在短视频页内处理，不自动退回首页；失败态保留 `重试` 与 `BACK`，空态保留空提示与 `BACK`，并已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:40 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频失败态。已确认单条播放失败时不自动退首页、不自动跳下一条，而是在当前页显示极简失败态，只保留 `重试` 与 `BACK` 退出路径；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:38 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频进度反馈。已确认底部进度条不常驻显示，只在左右键快进/快退时短暂浮现后自动消失，以兼顾 seek 反馈与沉浸感；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:32 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频切条反馈。已确认按 `上/下` 切条成功后不显示任何额外成功提示，避免破坏沉浸感；切条反馈仅由画面与播放内容变化承担，该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:29 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频暂停语义。已确认暂停只作用于当前条，切到其它条目后新条目仍自动播放，不继承上一条的暂停状态；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:27 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频页面 chrome。已确认页面标题、返回按钮、顶部说明等页面级 chrome 全部移除，整页只保留视频画面与最小播放反馈，退出完全依赖遥控器 `BACK`；该边界已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:23 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频历史语义。已确认保留现有观看历史上报，但历史只用于记录，不反向驱动短视频入口恢复；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:20 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频起播语义。已确认每次从首页进入短视频页时，都从当次新拉取信息流的第一条开始，不恢复上次停留位置；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:10 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频画面策略。已确认首版固定完整显示（FIT），允许两侧留黑边，不提供画面比例切换入口，也不默认使用铺满裁切；该策略已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 16:01 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频音频策略。已确认页面进入后默认有声，不提供页内静音开关，音量完全交给电视系统控制；该策略已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 15:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频列表边界。已确认首尾不回绕：第一条按 `上` 无动作，最后一条且无更多内容时播完停在当前条，不回第一条也不自动退出；该边界已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 15:43 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频页面边界。已确认首版不保留手机端那套右侧操作区，点赞/收藏/详情/画面比例切换/播放模式切换等入口全部移除，只保留播放画面和最小播放反馈；该边界已同步补充到 `CONTEXT.md`。
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频末尾行为。已确认默认按 `AUTO_NEXT` 语义播完自动切下一条，且页内不再保留播放模式切换入口；该语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 15:40 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 短视频遥控器行为。已确认左右键快进/快退直接复用 TV 端现有全局步长设置（5/10/15/20/30 秒，默认 10 秒），不为短视频单独加新设置；该语义已同步补充到 `CONTEXT.md`。
- 进度：继续通过 `$grill-with-docs` 收口短视频页主键语义。已确认 `OK/中键` 只做播放/暂停切换，不打开额外菜单、不切去独立操作区；完整遥控口径收敛为 `上/下切条`、`左/右按全局步长快退快进`、`OK 播放暂停`、`BACK 返回首页`，并已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 15:27 +0800
- 进度：开始通过 `$grill-with-docs` 收口“手机端短视频模式迁到 TV 端”。已确认这里的“短视频模式”指对齐手机端 `ShortFeedScreen` 的直接播放信息流，不走 `ShortDiscover` 式封面墙预览；该术语已同步补充到 `CONTEXT.md`。
- 进度：继续通过 `$grill-with-docs` 收口 TV 入口形态。已确认入口放在 TV 首页左侧菜单 `IPTV` 下方，点击后进入无左侧菜单的独立短视频页；该入口语义已同步补充到 `CONTEXT.md`。
- 进度：继续通过 `$grill-with-docs` 收口返回语义。已确认短视频页按 `BACK` 直接回首页，并恢复离开前的首页状态，不把 `短视频` 作为长期选中菜单态；该返回语义已同步补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:59 +0800
- 进度：完成压缩包导入多分组主链路。后端已补齐分组模型查询、创建/编辑/删除/加入/移出/处理接口与路由；管理端批次详情新增分组工作区、未分组虚拟卡片、分组创建/编辑弹窗、加入/移出分组动作与分组直处理入口，文件行同步显示分组归属。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_groups.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`internal/models/models.go`、`migrations/0027_archive_import_groups.up.sql`、`migrations/0027_archive_import_groups.down.sql`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/codex-go-cache go test ./internal/services ./internal/handlers -run 'TestArchiveImport|TestRegisterIncludesAdminPasswordVaultRoutes' -count=1` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/handlers/admin_archive_import.go internal/handlers/router.go internal/models/models.go internal/services/archive_import.go internal/services/archive_import_groups.go internal/services/archive_import_test.go admin-web/src/api/admin.js admin-web/src/api/admin.spec.js admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js migrations/0027_archive_import_groups.up.sql migrations/0027_archive_import_groups.down.sql` 无命中。

## 2026-06-24 14:40 +0800
- 进度：继续推进压缩包导入多分组实现，已确认后端现有分组服务骨架已接上 batch detail，当前需要补齐路由/handler 接口并处理 `UpdateFile` 的编译断点，再同步管理端分组工作区。
- 影响文件：`internal/services/archive_import.go`、`internal/handlers/router.go`、`internal/handlers/admin_archive_import.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`plan.md`
- 验证：待执行后端定向测试、前端定向测试、`npm run build`、`gofmt`、`git diff --check`、乱码扫描。

## 2026-06-24 14:12 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”展示信息。已确认分组卡片至少展示成员总数、未入库数和已入库数，便于管理员快速判断该组是否还有待处理文件。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:09 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”迁移交互。已确认文件进组/出组的主交互基于多选后的批量动作完成，分别提供“加入分组”和“移出分组”入口，不走单文件行内复杂下拉。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:14 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”编辑边界。已确认真实分组允许改名与改备注；改名后仍需满足同一批次内唯一约束。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:10 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”显示顺序。已确认批次详情中“未分组”虚拟区块固定置顶，真实分组按创建时间倒序展示，优先暴露最新仍在整理的分组。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”未分组呈现。已确认批次详情中固定展示“未分组”虚拟区块，用来承载尚未加入任何分组的文件；该区块可直接处理，但不是可重命名或可删除的真实分组。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 14:02 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”命名约束。已确认同一批次内分组名必须唯一，不允许重复，避免回看、重命名和恢复处理时混淆。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

## 2026-06-24 13:50 +0800
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”已入库边界。已确认 `ready` / `existing` 这类已入库文件在分组层面冻结：不再允许换组，也不再继续吃到新的分组默认值，避免让管理员误以为正式资产会被同步改写。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”空组生命周期。已确认空分组不自动删除；即使当前没有成员文件，也继续保留，只有管理员显式删除时才真正移除。
- 进度：继续通过 `$grill-with-docs` 收口压缩包导入“多个分组”操作层级。已确认分组应作为一级操作对象，允许管理员直接处理本组内仍未入库的成员文件，不要求每次重新勾选同一组文件。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建。

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

## 2026-06-23 14:30 +0800
- 进度：开始调整压缩包导入视频标题语义。已通过 `$grill-with-docs` 确认：上传弹窗的批次标题改为视频默认标题；新上传批次的视频默认直接使用该标题，不再拼接文件名；图片标题继续按文件名派生；单文件标题清空回到视频默认标题；文件清单多选支持批量覆盖标题形成临时标题分组。
- 影响文件：预计 `internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/services -run 'TestArchiveImport' -count=1`、`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-22 14:02 +0800
- 进度：完成工具箱菜单最多四列改动的家用部署机推送。`git push deploy master` 已将 `da2a56f` 推送到部署机，hook 判定 `REBUILD_FRONTEND=1`、`RESTART_GO=0`，完成 `admin-web` 构建并更新前端产物；GitHub mirror push 在部署机 hook 中失败但标记为 non-fatal。
- 影响文件：`plan.md`
- 验证：部署机侧 `ssh chee@192.168.1.24 'curl -fsS http://127.0.0.1:8080/healthz'` 返回 `{"status":"ok"}`。

## 2026-06-22 13:59 +0800
- 进度：完成工具箱菜单布局收尾。管理端构建、`git diff --check` 和乱码扫描均已通过；Vite 仅保留现有 chunk size 警告。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src/views/Toolbox.vue admin-web/src/views/toolboxPage.spec.js CONTEXT.md plan.md` 无命中。

## 2026-06-22 13:57 +0800
- 进度：完成工具箱菜单自适应布局实现。`Toolbox.vue` 已从无限 `auto-fit` 改为最多四列，并在 `80rem`、`64rem`、`48rem` 下逐级降为三列、两列、单列；`toolboxPage.spec.js` 增加静态回归断言；`CONTEXT.md` 补充工具箱菜单密度边界。此次只影响管理端 Web，不涉及 Android App 版本号。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js` 通过；待执行 `cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-22 13:55 +0800
- 进度：开始调整管理端工具箱菜单布局。已确认当前 `Toolbox.vue` 使用 `auto-fit`，宽屏可能超过四列；计划先补静态断言锁定“最多四列、自适应降列”，再改 CSS、更新上下文并跑管理端验证。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/toolboxPage.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-22 13:50 +0800
- 进度：完成家用部署机推送。部署机 `.env` 已配置 `PASSWORD_VAULT_KEY`（未在仓库记录明文）；`git push deploy master` 已将 `fa4909e` 推送到部署机，远端 hook 完成前端构建、Go 构建、codesign、migration、launchctl kickstart，并返回 `/healthz OK`。GitHub mirror push 在部署机 hook 中失败但标记为 non-fatal，不影响家用部署机部署。
- 影响文件：`plan.md`；部署机本地文件系统（不进仓库）：`~/deploy/ai-video-server/.env`
- 验证：部署机侧 `curl -fsS http://127.0.0.1:8080/healthz` 返回 `{"status":"ok"}`；`launchctl print` 确认 `com.aivideo.server` 与 `com.aivideo.worker` 均为 `running`。

## 2026-06-22 13:45 +0800
- 进度：开始将管理端密码管理功能推送到家用部署机。部署前先在部署机 `.env` 写入 `PASSWORD_VAULT_KEY`（不在仓库记录密钥明文），再推送 `deploy master` 触发构建、迁移与重启。
- 影响文件：`plan.md`；部署机本地文件系统（不进仓库）：`~/deploy/ai-video-server/.env`
- 验证：待执行 `git diff --check`、乱码扫描、提交、部署机 env 校验、`git push deploy master` 和部署后健康检查。

## 2026-06-22 12:19 +0800
- 进度：完成密码库功能的验证收尾。后端定向测试、前端定向测试和前端构建均通过；`go test ./... -count=1` 只失败在沙盒禁止 `httptest.NewServer` 监听本地端口的既有测试，失败点是 `internal/handlers/admin_image_generation_test.go`、`internal/queue/scrape_tasks_test.go`、`internal/services/scraper_av_mdcx_sites_test.go`，与本次密码库改动无关。
- 影响文件：`plan.md`
- 验证：`git diff --check` 通过；`rg -n $'\uFFFD' .env.example docs/run.md docs/家用部署机.md CONTEXT.md plan.md admin-web/src internal migrations` 无命中；`env GOCACHE=/private/tmp/codex-go-cache go test ./internal/config ./internal/services ./internal/repository ./internal/handlers -run 'TestLoad|TestPasswordVault|TestRegisterIncludesAdminPasswordVaultRoutes|TestPasswordVaultEntriesMigration|TestNormalizePasswordVault|TestScanPasswordVault|TestPasswordVaultCipher' -count=1` 通过；`npm run test -- src/api/admin.spec.js src/views/toolboxPage.spec.js src/components/base/commandPalette.helpers.spec.js src/views/adminDateTimeDisplay.spec.js` 通过；`npm run build` 通过；`env GOCACHE=/private/tmp/codex-go-cache go test ./... -count=1` 失败于沙盒端口限制。

## 2026-06-22 12:17 +0800
- 进度：补齐密码库功能的部署/示例文档。`PASSWORD_VAULT_KEY` 已写入 `.env.example`、`docs/run.md` 的远程开发示例和 `docs/家用部署机.md` 的部署示例，同时在 `CONTEXT.md` 补充密钥轮换风险。
- 影响文件：`.env.example`、`docs/run.md`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `git diff --check`、乱码扫描、后端定向测试、前端构建。

## 2026-06-22 11:57 +0800
- 进度：根据用户要求收敛工具箱密码管理功能为简单 CRUD。实现不再加入乐观锁、导入导出、密码生成、分类标签或二次验证；删除为确认后的直接删除，修改按最后提交覆盖。
- 影响文件：预计 `migrations/`、`internal/config`、`internal/models`、`internal/repository`、`internal/handlers`、`admin-web/src`、`CONTEXT.md`、`plan.md`
- 验证：待实现后执行后端定向测试、管理端定向测试、构建、`git diff --check` 与乱码扫描。

## 2026-06-22 11:51 +0800
- 进度：继续收口工具箱密码管理功能。已确认服务端加密主密钥由必填环境变量 `PASSWORD_VAULT_KEY` 提供，不放业务设置表，也不下发前端；术语沉淀为 `admin 密码库主密钥`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:51 +0800
- 进度：继续收口工具箱密码管理功能。已确认显示或复制密码前不做二次输入登录密码；列表接口不返回密码明文，只有单条显式显示或复制操作才产生临时明文。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:50 +0800
- 进度：继续收口工具箱密码管理功能。已确认首期操作只覆盖新增、列表搜索、编辑、删除、复制账号、显示密码与复制密码；不纳入导入导出、密码生成器、标签或分组；搜索不覆盖密码内容。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:48 +0800
- 进度：继续收口工具箱密码管理功能。已确认首期条目字段只含名称、账号、密码、网址与备注，不纳入标签、分组或权限层级；术语沉淀为 `admin 密码条目字段边界`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:46 +0800
- 进度：继续收口工具箱密码管理功能。已确认密码库持久化不保存数据库明文，首期采用服务端加密存储；明文只在管理员显式查看或复制时作为临时结果出现。该取舍已新增 ADR 记录，不采用端到端加密或明文入库。
- 影响文件：`CONTEXT.md`、`docs/adr/0014-admin-password-vault-server-side-encryption.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:45 +0800
- 进度：继续收口工具箱密码管理功能。已确认密码库归属所有管理员共享，不做个人私有隔离；条目需要支持后端多设备同步，继续待定凭据保护与字段边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:44 +0800
- 进度：继续收口工具箱密码管理功能。已确认首期需要多设备同步，密码条目以后端持久化作为同步来源，不采用纯浏览器本地 IndexedDB 方案；术语沉淀为 `admin 密码库`，强调它不是浏览器本地草稿或偏好。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:32 +0800
- 进度：开始通过 `$grill-with-docs` 收口工具箱密码管理功能。已确认“账号和密码”指外部服务、网站或设备的登录凭据，不改现有用户登录、管理员账号或认证授权模型；术语沉淀为 `admin 密码条目`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：需求收口中，尚未进入实现。

## 2026-06-22 11:17 +0800
- 进度：根据独立审查修复压缩包视频图片关联边界。旧批次里因历史扫描继承批次默认图片合集的视频文件，处理时不再把该默认值误写成视频图片图集；单文件保存视频时只提交当前可见的唯一图片图集；上传哈希并发唯一冲突分支也会在已有视频缺少图片图集时补写，不再漏掉本次关联。测试同步覆盖默认继承忽略、显式视频关联保留和单视频单图片图集约束。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoImageCollection' -count=1` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js` 通过；`go test ./internal/repository -run 'TestNormalizeSingleImageCollectionID|TestScanVideoRecordReadsOSHash|TestGetVideoOSHash|TestUpdateVideoOSHash' -count=1` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-22 11:07 +0800
- 进度：完成压缩包导入页回归修复。批次详情文件勾选位现在始终显示可见边框与选中态；Shift 连选改为“未选目标时加入区间、已选目标时移除区间”；视频文件支持设置唯一图片合集，批量编辑也可统一覆盖“视频关联的图片合集”，后端处理视频时把该图片合集写入视频关联，命中已有视频且原本未关联图片合集时补上但不覆盖既有关联。批次默认图片合集继续只继承到图片文件，避免上传默认值误把所有视频挂到同一图片图集。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`internal/repository/video_repository.go`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js` 通过；`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoImageCollection' -count=1` 通过；`go test ./internal/repository -run 'TestNormalizeSingleImageCollectionID|TestScanVideoRecordReadsOSHash|TestGetVideoOSHash|TestUpdateVideoOSHash' -count=1` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-22 10:52 +0800
- 进度：开始修复压缩包导入页三个回归：批次详情文件勾选位不可见、Shift 区间多选失效、压缩包内视频无法关联图片图集。已通过 `$grill-with-docs` 收口“图片和视频关联”语义为复用现有视频图片图集关系，不新增单图直连或封面自动绑定。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js`、`cd admin-web && npm run build`、`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoImageCollection' -count=1`、`git diff --check`、乱码扫描。

## 2026-06-21 17:30 +0800
- 进度：开始修复压缩包导入后视频列表误显示失败。已定位根因是压缩包导入把视频 `original_path` 写成临时 work 文件路径，而该 work 文件会在单文件处理结束时删除，导致后续异步刮削/转码读取源文件失败并把 `videos.status` 打成 `failed`。计划改为先把视频落到稳定原始文件路径再建库，并补服务层定向测试锁定该约束。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/services -run 'TestArchiveImport' -count=1`、`git diff --check`、乱码扫描。

## 2026-06-21 17:30 +0800
- 进度：完成压缩包导入视频源路径修复。`ProcessFile` 现会先把视频复制到 `UPLOAD_TEMP_DIR/archive-import-videos/<batch>/<file>/原始文件名` 稳定路径，再创建/复用视频记录；不再把会被立即删除的 work 临时文件写进 `videos.original_path`。同时补了“命中已有失败视频时修正 `original_path` 并允许重新入队”的兼容逻辑，避免旧批次失败项继续卡死；新增服务层纯逻辑测试锁定稳定源文件命名、存储位置和失败旧视频可重试边界。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/services/upload.go`、`internal/repository/video_repository.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImport|TestArchiveVideoSource' -count=1` 通过；`go test ./internal/handlers -run 'TestSelectRetranscodeInputPath' -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' internal/services/archive_import.go internal/services/archive_import_test.go internal/services/upload.go internal/repository/video_repository.go CONTEXT.md plan.md` 无命中。

## 2026-06-21 15:26 +0800
- 进度：完成压缩包导入页的批次优先 UI 重构收尾。页面已切换为“批次列表主视图 + 上传弹窗 + 批次详情抽屉 + 批量编辑弹窗”，批次详情内文件支持勾选、Shift 连选、按类型快捷选择、批量编辑与批量处理；单文件编辑降级为仅在单选时出现的例外项精修区，不再承载处理动作。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/toolboxPage.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js admin-web/src/views/toolboxPage.spec.js CONTEXT.md plan.md` 无命中。

## 2026-06-21 10:31 +0800
- 进度：开始落实压缩包导入页 UI 重构。目标是把页面改成“批次优先主视图 + 上传弹窗 + 批次详情抽屉 + 批量编辑弹窗”，同时保留勾选、Shift 连选、按类型快捷选择、批量处理与例外单文件编辑。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/toolboxPage.spec.js src/api/admin.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-21 10:46 +0800
- 进度：修复 `internal/services` 全量测试遗留失败。`TestParseTVAPKMetadataParsesReleaseAPK` 断言 release APK 固件为 `versionCode=80 / versionName=0.1.80`，但 `android-tv-app/tv-app/release/` 下的固件已于 6-19 重建到 121 / 0.1.121，导致 `go test ./internal/services` 全包失败。本次仅同步测试断言到 121 / 0.1.121，不改解析逻辑或固件；这是测试断言随 release 固件版本走的已知耦合（已沉淀到 CONTEXT.md）。
- 影响文件：`internal/services/tv_apk_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`GOCACHE=/private/tmp/codex-go-cache go test ./internal/services/ -count=1` 提权运行（httptest 用例需绑本地端口）结果 `ok video-server/internal/services`；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-21 10:32 +0800
- 进度：完成压缩包导入批次删除能力。后端新增批次删除接口，限制 `processing` 批次不可删；删除时只清理批次记录、文件清单、原压缩包副本与解包工作目录，不回删已入库的视频/图片实体。管理端批次列表同步新增删除按钮和二次确认文案。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestArchiveImportShouldProcessInBatch|TestValidateArchiveBatchDeletion|TestArchiveBatchCleanupPaths'` 通过；`cd admin-web && npm run test -- src/api/admin.spec.js src/views/ToolboxArchiveImport.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-21 10:18 +0800
- 进度：开始补压缩包导入批次删除能力。已确认当前缺口不是前端隐藏按钮，而是后端/前端都没有删除接口；本次收口删除语义为“仅允许删除非 processing 批次，删除批次记录、文件清单、原压缩包和解包工作目录，不删除已入库的视频/图片实体”。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`internal/handlers/admin_archive_import.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 Go 定向单测、`cd admin-web && npm run test -- src/api/admin.spec.js src/views/ToolboxArchiveImport.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-20 22:45 +0800
- 进度：修复压缩包导入批量处理候选范围。`ProcessAllFiles` 不再只处理字面 `pending`，而是覆盖可入库视频/图片中仍未成功入库的 `pending`、`failed`、`processing` 项；复制工作文件、视频 hash 等早期失败也会回写为 `failed`，避免文件卡在 `processing` 后无法被批量重试。同步补充服务层候选规则测试和 `CONTEXT.md` 语义沉淀。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run TestArchiveImportShouldProcessInBatch` 通过；`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`go test ./internal/services` 仍因既有 `tv_apk_test.go` 期望版本 80、当前 APK 元数据 121 不一致失败。

## 2026-06-21 01:23 +0800
- 进度：补压缩包导入页的文件清单排序切换，前端支持原始顺序与按类型排序两种模式；按类型排序时先展示视频、图片等可入库媒体，再保留跳过项，便于先处理真正可入库的待处理文件。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；乱码扫描无命中；已推送 `deploy/master`。

## 2026-06-21 01:35 +0800
- 进度：开始落实压缩包导入页的“待处理文件可选择、可 shift 连选、可批量处理所选”。前端将复用现有单文件处理接口逐个处理所选文件，保留单文件详情处理入口，不改后端批处理协议。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；乱码扫描无命中；已推送 `deploy/master`。

## 2026-06-20 22:41 +0800
- 进度：开始排查压缩包导入“待处理文件依旧不能批量处理”。初步定位后端批量处理只扫描 `pending`，而单文件处理失败后可能处于 `failed`，早期复制/hash 失败还可能卡在 `processing`，导致后续批量处理跳过这些仍未入库的可处理文件。计划补服务层定向测试锁定批量候选范围，再修 `ProcessAllFiles` 与早期失败状态回写。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`plan.md`
- 验证：待执行 `go test ./internal/services -run 'TestArchiveImport'`、必要的 admin-web 定向验证、`git diff --check`、乱码扫描。

## 2026-06-20 22:32 +0800
- 进度：完成压缩包导入交互语义落地。压缩包导入页已支持在默认视频合集、默认图片合集、视频文件合集、图片文件合集四个位置原地新建合集并自动选中；文件清单和详情改为显示中文媒体类型加格式/MIME，并展示跳过原因；批量处理逻辑未改，仍只处理视频/图片待处理项。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-20 22:25 +0800
- 进度：开始落地压缩包导入交互语义。计划先补 `ToolboxArchiveImport` 静态断言锁住页内新建合集、文件类型中文化与跳过原因可见，再在压缩包导入页接入已有视频合集/图片合集创建 API，最后跑 admin-web 定向测试、构建、`git diff --check` 与乱码扫描。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/ToolboxArchiveImport.spec.js src/views/videoUpload.remote.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

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

## 2026-06-20 14:05 +0800
- 进度：开始落实压缩包导入页的“合集/标签不再暴露 UUID/JSON”收口。已确认压缩包导入里的默认标签也应对齐管理端现有标签选择器语义，改成可选择、可输入的标签多选，而不是继续让管理员填 JSON 数组；默认合集仍沿用“选择已有合集，缺失则先新建再回填”的语义。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/videoUpload.remote.js`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run build`，以及必要的定向检查/乱码扫描。

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

## 2026-06-20 00:21 +0800
- 进度：继续收口压缩包上传导入功能的管理端尾部工作。已修复 `ToolboxArchiveImport.vue` 的 Vue 2 过滤器写法，改为直接调用 `formatFileSize()`；同时把文件级合集输入收拢成可规范化的选择值，补了工具箱页、命令面板、路由和 API 的静态断言，并在 `CONTEXT.md` 增补“压缩包去重保留现有资产”术语，避免后续误把去重理解成覆盖资产。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run build`、`cd admin-web && npm run test` 或至少相关 vitest 定向用例、`git diff --check`、乱码扫描。

## 2026-06-19 21:40 +0800
- 进度：开始落实压缩包上传导入功能。已收口需求为：仅管理员入口，支持 `zip` / `rar` / `7z`，支持密码包但不允许嵌套压缩；压缩包先形成批次与文件清单，再按文件级别走现有视频/图片去重、压缩和入库逻辑。当前正在补后端批次/文件实体迁移与管理端入口设计，并确认系统工具链可用 `bsdtar` 处理非加密 `zip/7z/rar`，加密包需要单独兜底。
- 影响文件：`CONTEXT.md`、`plan.md`、`migrations/*`、`internal/services/*`、`internal/handlers/*`、`admin-web/src/views/*`、`admin-web/src/api/admin.js`
- 验证：待执行迁移模式测试、后端定向单测、admin-web 构建与乱码扫描。

﻿# plan.md

本文件用于增量记录”计划与修改”，不得覆盖历史记录，只能追加。

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

## 2026-06-19 19:45 +0800
- 进度：继续收尾 TV 单片长视频软重试，修正取消重试后的实际播放器副作用，从“停止当前播放器”收窄为“仅回写快照、不主动 pause/stop”，与现有取消不回退会话语义对齐；同步补强 spec 断言，防止回归到停播实现。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`plan.md`
- 验证：待执行 TV 定向单测和 `compileDebugKotlin`；当前离线 Gradle 已能启动，但 Android Gradle Plugin 未缓存，完整验证仍受环境限制。

## 2026-06-19 19:36 +0800
- 进度：继续推进 TV 单片长视频软重试收尾，修正诊断面板相关 spec 断言，避免把当前 `closeDiagnosticsPanel()` 返回失败态的实现误判成“仍然触发 retry”。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`plan.md`
- 验证：待执行 TV 定向单测和 `compileDebugKotlin`；当前受限于 Gradle wrapper 需联网下载分发包，先用工作区内 wrapper/cache 继续排障。

## 2026-06-19 19:29 +0800
- 进度：按用户取消要求，收口并删除 `CONTEXT.md` 中所有压缩包上传/导入沉淀，停止继续推进该方向的设计。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `git diff --check`、`rg -n \"压缩包导入|压缩上传|zip|rar|7z\" CONTEXT.md plan.md` 确认仅保留必要历史。

## 2026-06-19 11:08 +0800
- 进度：完成 `TV 单片长视频播放器软准备` 首轮落地。单片页已新增“首帧后中心轻态”状态机：重试时立即进入 `正在重试播放`，成功后给出 `已恢复播放`，`BACK` 可取消当前重试并显示 `已取消重试`，失败态下沉到播放器中心区并保留“重试播放 + 诊断信息”；诊断面板主动作改为 `返回`，仅回到失败轻态。Media3 单片播放器已改为复用同一 `ExoPlayer` 实例重 prepare，接入 `onRenderedFirstFrame()`，并支持通过取消 key 停掉当前 prepare 以配合 latest-wins 丢弃过期结果。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetryLogicTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:compileDebugKotlin` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvLongFormPlayerSoftRetrySpecTest --tests com.chee.videos.feature.tv.TvLongFormPlayerSoftRetryLogicTest --tests com.chee.videos.feature.tv.TvLongFormMedia3PlayerTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:assembleDebug` 通过；待执行 `git diff --check` 与乱码扫描。

## 2026-06-19 09:32 +0800
- 进度：开始落实 `TV 单片长视频播放器软准备` 实现。计划先把“已出现首帧后的单片重试/取消/失败/成功/诊断”状态抽成最小 helper 和定向红灯测试，再把 Media3 单片播放器从整块 `TvErrorState` 失败流迁到播放器中心轻态，最后补 TV 端版本递增、`CONTEXT.md`/`plan.md` 完成记录与 Gradle 验证。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行单片播放器定向单测/源码断言、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check`、乱码扫描。

## 2026-06-17 18:08 +0800
- 进度：继续通过 `$grill-with-docs` 收口下一轮 TV 流畅性优化主线。已确认下一轮优先目标不是首页、海报墙或 IPTV，而是 `TV 单片长视频播放器软准备`：当单片播放器已经有承接画面后，重试播放源、重建播放链路或重新入播时，不再退回整页 `loading/error`，而是尽量保留当前承接画面，把 preparing/failed 降到播放器中心区表达。下一步继续细化：单片播放器进入 soft preparing 时，旧视频的承接方式是否也沿用当前播放态，还是一律人为停在当前帧等待新链路成功。
- 影响文件：`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与 `CONTEXT.md` 长期术语沉淀。

## 2026-06-17 18:12 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器承接强度。已确认 `TV 单片长视频播放器软准备` 不只是不退回整页 `loading/error`，还要求在已有承接画面后重试播放源或重建播放链路时，不为了准备过程主动清黑底或清空旧画面；只有底层播放器自己丢失 surface 时，才接受不可避免的瞬时闪断。下一步继续细化：这条是否直接沿用现有通用术语 `TV 软准备沿用当前播放态`，即“原本在播就继续播，原本暂停就停在当前帧”，还是单片场景需要更保守的一刀切暂停策略。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 00:00 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器承接播放态。已确认 `TV 单片长视频播放器软准备` 直接沿用现有通用术语 `TV 软准备沿用当前播放态`，不为单片场景额外改成统一暂停：soft preparing 期间，旧内容若原本正在播放就继续播，若原本已暂停就停在当前帧等待新链路结果。下一步继续细化：单片播放器 soft preparing 失败态可见时，`BACK` 是先关闭中心失败态留在当前视频，还是直接进入原有返回确认/退出链路。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器软准备失败时的返回语义。已确认单片播放器与剧集播放器保持一致：当 `TV 单片长视频播放器软准备` 的中心失败态仍在屏幕上时，第一次 `BACK` 只关闭失败态并继续留在当前旧视频承接；只有失败态关闭后再次按 `BACK`，才进入原有返回确认或退出链路。下一步继续细化：在已有 `BACK` 关闭语义的前提下，单片失败态是否还需要一个显式“继续观看/留在当前视频”动作，还是保持“重试 + 诊断信息（可选）+ BACK 关闭”即可。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:07 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器失败态动作最小集。已确认单片 soft preparing 失败态不额外增加“继续观看/留在当前视频”按钮，只保留“重试播放”主动作和可选“诊断信息”副动作；关闭失败态统一由 `BACK` 承担，避免中心失败态因为多一个关闭动作而变重。下一步继续细化：用户触发“重试播放”后，失败态是否应立即原地切回 preparing 轻提示，让用户立刻感知“这次重试已开始处理”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:12 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器重试反馈时机。已确认单片播放器在 soft preparing 失败后，只要用户触发“重试播放”，旧失败态就应立即被当前这次重试的 preparing 轻提示顶掉，而不是拖到真正成功或再次失败时才更新。下一步继续细化：这次 preparing 轻提示是否要像剧集 soft preparing 一样带出当前重试目标的明确文案，还是保持更短的单片通用文案。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:22 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器 preparing 文案粒度。已确认单片播放器触发重试后的 preparing 中心轻提示保持短文案，不需要像电视剧那样带出目标编号或长解释；单片场景只需让用户明确感知“这次重试已经开始处理”。下一步继续细化：如果这次重试最终成功，中心区是否也给一条极短成功确认，还是直接无感恢复播放、不再额外提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:23 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器重试成功反馈。已确认单片播放器重试成功后，中心区同样只给一条极短确认，不展开成解释文案；它只负责告诉用户“这次重试已经生效”，不额外解释播放源、链路或容器细节。下一步继续细化：这条极短成功确认是否也保持无编号短语义，还是单片需要再明确写出“已恢复播放”之类的提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:36 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器成功确认文案。已确认单片播放器重试成功后的确认保持无编号、无目标说明的短反馈，例如“已恢复播放”这一类，不像电视剧那样带分集号；单片场景只需说明重试已生效。下一步继续细化：单片播放器 soft preparing 的“首屏例外”边界是否与剧集保持一致，即只有首次进入且当前没有任何可承接画面时，才允许继续使用整页 loading/error。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 09:38 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器首屏例外边界。已确认单片播放器与剧集保持一致：只有首次进入且当前没有任何可承接画面时，才允许继续使用整页 `loading/error`；一旦已有承接画面，后续准备失败、重试与成功确认都应下沉到播放器内部轻量状态，不再把整页状态重新抬起。下一步继续细化：在“已有承接画面”的前提下，单片播放器的失败态是否统一挂到播放器中心区，而不是沿用当前外层 `TvErrorState` 大面板。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 10:25 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器失败态挂载点。已确认在“已有承接画面”的前提下，单片播放器的 preparing/failed 反馈也统一挂到播放器中心区，而不是沿用外层 `TvErrorState` 大面板；这一步直接复用现有通用术语 `TV 长视频播放器软准备中心挂载点`，不再额外发明单片专用挂载点名词。下一步继续细化：单片播放器 soft preparing 期间如果用户按 `BACK`，是否也应先取消这次正在进行的重试，再由下一次 `BACK` 进入原有返回确认/退出链路。
- 影响文件：`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 10:30 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器 preparing 期间的返回语义。已确认单片播放器与剧集保持同一交互方向：soft preparing 期间第一次 `BACK` 先取消当前重试并继续留在旧视频承接，只有重试已取消或已结束后再次按 `BACK`，才进入原有返回确认/退出链路。下一步继续细化：取消这次重试后，单片播放器是否也给一条比成功提示更轻、更短的中心反馈，例如“已取消重试”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 10:39 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器取消重试反馈。已确认单片播放器在 preparing 期间被 `BACK` 取消当前重试后，也给一条比成功提示更轻、更短的中心反馈，例如“已取消重试”；该提示只在确实存在待取消重试时显示，不抢焦点、不阻断播放、停留时间短。下一步继续细化：显式打开“诊断信息”时，单片播放器是否仍允许沿用当前较重的独立诊断面板，而不强行塞进中心轻提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 10:44 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器诊断视图边界。已确认“诊断信息”属于用户显式打开的调试视图，不并入 soft preparing 的中心轻提示；即使单片播放器日常的 preparing/failed/success/cancel 都下沉到播放器中心区，诊断信息仍允许沿用较重的独立面板，以保证排障信息可读。下一步继续细化：如果单片播放器连续重试、连续失败，中心失败态是否也像剧集一样原地更新内容，避免重复闪烁或退场重进。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 10:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器连续失败反馈。已确认如果单片播放器连续触发重试且连续失败，中心失败态继续复用同一块区域原地更新，不做闪烁、退场重进或额外过渡。下一步继续细化：单片播放器里“已有承接画面”的判定口径是什么，是只要播放器曾进入过可视承接态即可，还是必须等到底层真正出现首帧/进入 ready 后才算。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 11:04 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器“已有承接画面”的判定口径。已确认这里应按保守口径处理：不能因为用户已经进入播放器页面就算“已有承接画面”，而要等到底层真正进入过一次真实可承接的视频状态后，后续失败/重试才转入 soft preparing。代码侧现状也支持继续细化这个问题：TV 单片 ExoPlayer 路径目前只有 `STATE_READY` / `isPlaying` / `snapshot`，还没有显式接入 `onRenderedFirstFrame()`，但仓库其他 ExoPlayer 页面已经在使用该信号。下一步继续细化：单片播放器是否应明确以“至少出现过一次首帧”作为软准备的进入门槛，而不是把 `STATE_READY` 单独当作已有承接画面。
- 影响文件：`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 15:21 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器“已有承接画面”的代码信号。已确认 TV 单片 ExoPlayer 路径应明确以“至少出现过一次首帧”作为 soft preparing 的进入门槛，而不是把 `STATE_READY` 单独当作已有承接画面；仓库里其它 ExoPlayer 页面也已在用 `onRenderedFirstFrame()` 作为用户真正看到过画面的信号。下一步继续细化：如果用户已经看到过首帧、随后暂停在某一帧上，这种暂停帧是否同样视为可承接画面，允许后续失败/重试继续走 soft preparing。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 15:23 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器暂停帧承接语义。已确认只要单片播放器此前已经真正出现过首帧，随后停在暂停帧上的画面同样视为可承接画面；后续失败/重试继续走 soft preparing，不因为当前是暂停态就退回首屏整页状态。下一步继续细化：单片播放器这类“重试当前视频链路”的成功结果，是统一直接恢复播放，还是需要继承重试前的暂停态。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 15:26 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器重试成功后的播放态。已确认单片播放器的“重试播放”属于对当前视频恢复播放链路的明确继续播放意图，因此一旦这次重试真正成功，播放器统一直接恢复播放，不继承重试前的暂停态；当前单片代码里的重试动作也已经按这个方向写成 `hasStartedPlayback=true` 且 `isPausedByUser=false`。下一步继续细化：既然成功后会直接恢复播放，中心区那条极短成功确认是否仍然保留，还是只靠视频动起来本身表达“重试已生效”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 15:27 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器重试成功后的确认方式。已确认即使单片播放器的重试成功后会直接恢复播放，中心区那条极短成功确认仍然保留；视频重新动起来能表达很多，但在刚经历过失败后，一条短促的“已恢复播放”能更明确地确认这次重试已经生效。下一步继续细化：用户显式打开“诊断信息”后再关闭它时，是否应回到原先那层失败轻态，而不是顺手自动触发一次新的重试。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-18 15:30 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器诊断面板关闭后的去向。已确认“诊断信息”只是查看当前失败排障信息的独立面板，不承担“确认重试”的语义；用户关闭它后应回到原先那层失败轻态，由用户再明确决定是否重试，关闭诊断面板本身不自动触发新的重试。下一步继续细化：既然诊断面板主动作不该顺手重试，它在单片场景里是否应改成单纯“返回”/“关闭”，而把真正的“重试播放”留在失败轻态里。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 08:54 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器诊断面板主动作。已确认诊断面板主动作应改成单纯“返回”/“关闭”，只负责退出诊断面板并回到原失败轻态，不顺手清空失败态，也不顺手触发新的重试；真正的“重试播放”动作只保留在失败轻态里。当前单片代码在 `showDolbyVisionDiagnostics` 分支的 `onAction` 仍然会清错并直接 `routeRetryNonce += 1`，后续实现需按新语义改掉。下一步继续细化：从诊断面板返回到失败轻态后，焦点应默认回到“重试播放”主动作，还是回到上一次在失败轻态里停留的动作。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 08:57 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器诊断面板返回后的焦点语义。已确认从诊断面板返回到失败轻态后，默认焦点应落回失败轻态的主动作“重试播放”，而不是停在上一次的副动作或依赖 Compose 默认焦点漂移。代码侧也暴露了这个风险：`TvErrorState` 目前只是按按钮顺序渲染，并没有显式的“返回后焦点回主动作”保障。下一步继续细化：当失败轻态本身再次被新的 preparing 顶掉、随后又失败回到失败轻态时，默认焦点是否也继续落回“重试播放”主动作。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:02 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器连续失败后的焦点落点。已确认当失败轻态已被新的 preparing 顶掉、随后这次重试又再次失败时，回到失败轻态后的默认焦点仍继续落回“重试播放”主动作。这样无论是首次失败、从诊断返回，还是连续重试再次失败，用户的下一步恢复入口都保持稳定一致。下一步继续细化：如果旧失败态已被新的 preparing 顶掉，用户又在 preparing 期间按 `BACK` 取消这次重试，是回到“已取消重试”轻提示即可，还是要把刚才那层旧失败态重新抬回来。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器取消当前重试后的回退语义。已确认如果旧失败态已经被新的 preparing 顶掉，用户又在 preparing 期间按 `BACK` 取消这次重试，界面只保留“已取消重试”的轻提示即可，不再把旧失败态重新抬回来；旧失败一旦被新重试覆盖，就应视为过期状态。下一步继续细化：单片播放器是否也需要像剧集 `latest-wins` 一样，为连续重试建立“最后一次重试生效、旧结果晚到不得回写”的明确语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:11 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器连续重试的并发语义。已确认单片播放器也需要和剧集播放器一样建立 `latest-wins`：当用户连续触发新的“重试播放”时，系统只以最后一次重试为准，旧重试对应的 preparing / failed / success 结果一旦晚到，必须失效，不得把当前状态回滚到用户已经放弃的旧重试上。下一步继续细化：这些过期的旧重试结果是否应完全静默丢弃，连“旧重试失败/成功”的过期提示都不再弹。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:16 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器过期重试结果的处理方式。已确认凡是已经被新的重试覆盖、或已经被用户放弃的旧重试结果，一旦晚到，必须完全静默丢弃；它们既不能回写当前状态，也不能额外弹出“旧重试成功/失败”的过期提示。下一步继续细化：用户已经按 `BACK` 取消当前重试后，如果这次已取消的重试结果随后晚到，是否也按同样规则完全静默丢弃。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:24 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器已取消重试的晚到结果。已确认用户一旦通过 `BACK` 明确取消当前重试，这次重试在语义上就已经结束；它后续即使晚到成功或失败结果，也必须和其它过期重试结果一样完全静默丢弃，不再对界面发声。下一步继续细化：失败轻态里的“诊断信息”副动作是否也需要像主动作“重试播放”一样，在重新打开失败轻态时始终稳定出现在同一位置，而不是因是否可诊断而让按钮布局频繁抖动。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-19 09:28 +0800
- 进度：继续通过 `$grill-with-docs` 收口单片播放器失败轻态的动作布局稳定性。已确认只要当前设备/路由支持“诊断信息”，这个副动作就在失败轻态里稳定待在次动作位，不因失败轻态反复出现而左右跳动；这样主动作“重试播放”和副动作“诊断信息”的空间关系保持稳定，减少视觉和焦点抖动。下一步继续细化：当诊断信息本身不可用时，是否直接收成单主动作布局，而不是为了布局稳定性保留一个不可点的空占位。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补单片播放器定向测试、TV 全量验证、版本递增与相关长期术语沉淀。

## 2026-06-17 17:53 +0800
- 进度：完成本轮 TV 剧集播放器软准备收尾验证。TV 全量单测、`assembleDebug`、`git diff --check` 与乱码扫描均通过；当前提交面只包含本轮软准备状态机/中心反馈/测试支撑/版本递增以及 `CONTEXT.md`、`plan.md` 的长期沉淀，不包含无关模块改动。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvSeriesPlayerPreparingStateSpecTest.kt`、`TvSeriesEpisodeRailSpecTest.kt`、`TvSeriesMixedPlaybackControlsSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlaySpecTest.kt`、`TvLongFormTitleOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest --tests com.chee.videos.core.ui.TvSeriesCorePlaybackOverlaySpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无命中。

## 2026-06-17 17:50 +0800
- 进度：完成 `TV 长视频播放器软准备` 首轮实现。电视剧播放器已拆分目标分集与实际播放分集，切集 preparing/failed/success/cancel 统一降到播放器中心层；已有播放内容时切集不再清空 `currentVideoId/currentSourceUrl` 或退回整页 loading/error，失败/取消后 rail 和标题立即回到实际播放分集，`BACK` 在 preparing 时先取消目标、在失败态时先关闭失败中心态。同步让主动切集跳过续播卡，TV 版本递增到 `0.1.119` / `versionCode=119`，并补定向测试与源码结构断言。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvSeriesPlayerPreparingStateSpecTest.kt`、`TvSeriesEpisodeRailSpecTest.kt`、`TvSeriesMixedPlaybackControlsSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：已通过 `cd android-tv-app && ./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest --tests com.chee.videos.core.ui.TvSeriesCorePlaybackOverlaySpecTest`；待执行 TV 全量单测、`assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 16:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口 preparing 文案。已确认 soft preparing 进行中时，中心轻提示也要明确带出当前目标分集号，例如“正在切到第 N 集”；至此软准备的主交互语义已经基本收口，可转入 TV 播放器实现。下一步开始补红灯测试并落地状态拆分、中心轻提示与 `BACK` 取消链路。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口 rail 与成功反馈文案。已根据最新确认修正术语：分集 rail 不再显示 preparing 目标弱态，只跟随当前实际播放分集与焦点；soft preparing 成功后的中心短反馈继续明确带出最终生效的分集号，例如“已切到第 N 集”。下一步继续细化：preparing 轻提示本身是否也需要明确写出目标分集号。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:53 +0800
- 进度：继续通过 `$grill-with-docs` 收口 soft preparing 期间重新打开分集 rail 的语义。已确认单纯打开 rail 不会自动取消当前准备中的目标切换，只有明确改选或按 `BACK` 才终止当前目标；这样 rail 可以承担查看和改选入口，而不会把半途中的切换无故掐断。下一步继续细化：当 rail 重新打开时，当前正在 preparing 的目标应不应在 rail 里用弱态标出来，还是只在中心区表达。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:46 +0800
- 进度：继续通过 `$grill-with-docs` 收口旧失败态遇到新切换时的让路时机。已确认用户发起新的切集后，只要新目标进入 soft preparing，就立即清掉旧失败态并切回 preparing 轻提示，不把过期失败继续挂在屏幕上。下一步继续细化：preparing 轻提示本身在连续切集时是否需要显式带上目标分集号，还是保持更抽象的“正在切换”文案。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:44 +0800
- 进度：继续通过 `$grill-with-docs` 收口软切换失败态在连续失败场景下的过渡方式。已确认当旧失败提示尚在屏幕上、用户再次切集且新目标也失败时，中心失败态按同一块区域原地替换内容，不做重复闪烁或退场重进，以保证体感稳定。下一步继续细化：当用户发起新的切集后，新目标仍在 preparing 期间，旧失败态是立刻清空并切回 preparing，还是保留到新结果落定再被覆盖。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:34 +0800
- 进度：继续通过 `$grill-with-docs` 收口成功/取消提示遇到新操作时的让路规则。已确认这两类轻提示一旦遇到新的遥控输入就立即消失，把空间让给当前交互，不继续拖着旧状态占屏。下一步继续细化：失败提示在用户直接发起新的选集动作时，是保留到新结果覆盖，还是一发起新切换就立即清掉。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:22 +0800
- 进度：继续通过 `$grill-with-docs` 收口播放器中心轻提示的停留时长。已确认成功提示最短、取消提示次短、失败提示不自动消失，三者按语义强度分层停留，避免轻确认类反馈拖慢体感，同时保证失败态不会悄悄蒸发。下一步继续细化：一旦用户继续进行遥控操作，成功/取消提示是否应立即让路消失。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:18 +0800
- 进度：继续通过 `$grill-with-docs` 收口失败/取消目标在 rail 中的残留表达。已确认重新打开分集 rail 时，不为上一次失败或取消目标分集保留额外弱提示痕迹；这层信息只由中心轻提示短暂承接，rail 本身回归当前实际播放分集和当前焦点的操作语义。下一步继续细化：成功/失败/取消三类中心轻提示分别应停留多久，以及是否需要因用户继续操作而立即消失。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:16 +0800
- 进度：继续通过 `$grill-with-docs` 收口失败态/取消态后重新打开分集 rail 的默认焦点。已确认 rail 重新打开时默认回到当前实际播放分集，而不是回到上一次失败或取消的目标分集，以保持当前稳定播放上下文。下一步继续细化：rail 里是否还需要为上一次失败目标保留一个弱提示痕迹，帮助用户理解刚才失败的是哪一集。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:09 +0800
- 进度：继续通过 `$grill-with-docs` 收口取消目标切换后的提示语义。已确认用户在 soft preparing 期间按 `BACK` 取消一次待切换目标后，可以在中心区给出一条比成功提示更轻、更短的取消反馈；该提示仅在确实存在待取消目标时显示，不抢焦点、不阻断播放。下一步继续细化：用户从失败态或取消态重新打开分集 rail 时，默认焦点应落在当前实际播放分集，还是上一次失败/取消的目标分集。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口 preparing 期间的 `BACK` 语义。已确认当新目标仍在 soft preparing、旧分集还在承接时，第一次 `BACK` 解释为取消当前目标切换，继续留在旧分集；只有取消后再次按 `BACK`，才进入播放器既有的返回确认/退出链路。下一步继续细化：取消目标切换后是否需要一条极短的中心提示，例如“已取消切换到第 N 集”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 16:00 +0800
- 进度：继续通过 `$grill-with-docs` 收口“已完成分集从头播放”的提示策略。已确认这种场景不额外追加“已从头播放”解释文案，只沿用通用的“已切换到第 N 集”短反馈，避免中心提示从确认型退化为规则说明。下一步继续细化：soft preparing 期间若用户按 `BACK`，是取消当前目标切换并回到旧分集承接态，还是仍走播放器退出确认链路。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:42 +0800
- 进度：继续通过 `$grill-with-docs` 收口主动切集命中已完成历史的播放策略。已确认若目标分集历史已接近片尾或已判定完成，主动切过去后直接从头开始播放，不再沿用该历史进度。下一步继续细化：这种“从头开始”是否需要一个额外的短提示，还是沿用通用的“已切换到第 N 集”成功反馈即可。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:41 +0800
- 进度：继续通过 `$grill-with-docs` 收口新目标切换成功后的续播策略。已确认主动切集场景下，如果目标分集本身存在历史进度，切换成功后直接从该分集历史进度自动续播，不再弹续播提示卡二次询问；续播提示卡只保留给首次进入播放器或其它非主动切集场景。下一步继续细化：主动切集时若目标分集历史已接近片尾或已完成，是否仍按历史进度直接续播，还是改为从头开始。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:36 +0800
- 进度：继续通过 `$grill-with-docs` 收口新目标切换成功后的播放态。已确认 soft preparing 期间旧内容仍沿用当前播放/暂停态承接，但一旦新目标真正切换成功，就直接播放新目标，不继承“旧内容原本暂停”的状态；同时继续沿用新目标自身的历史进度与续播提示规则。下一步继续细化：既然新目标会直接播放，若它存在历史进度，是否还要保留当前的续播提示卡倒计时，还是切集场景直接从历史进度自动续播不再询问。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:34 +0800
- 进度：继续通过 `$grill-with-docs` 收口软切换成功后的 rail 行为。已确认目标分集真正切换成功后，也继续保持当前“选完即收起 rail”的沉浸路径，不自动把 rail 留在屏幕上；成功确认只通过中心区短反馈表达。下一步继续细化新目标切换成功后，是继承切换前的播放/暂停态，还是统一自动播放。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:32 +0800
- 进度：继续通过 `$grill-with-docs` 收口 soft preparing 失败后的 rail 行为。已确认失败后不自动重新打开分集 rail，维持当前“选完即收起 rail”的沉浸路径；若用户要改选其它分集，通过轻错误态里的显式“选集”动作重新进入 rail。下一步继续细化目标分集切换成功后，rail 是否也继续保持收起，还是在某些连续选集场景下保留打开。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:31 +0800
- 进度：继续通过 `$grill-with-docs` 收口 soft preparing 失败态的返回语义。已确认中心轻错误态可见时，第一次 `BACK` 只关闭这层失败态并留在当前实际播放分集；只有错误态关闭后再次按 `BACK`，才继续走播放器原有的返回确认/退出链路。下一步继续细化 soft preparing 失败后是否要自动重新打开分集 rail，还是维持当前“选完即收起 rail”的沉浸路径。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:29 +0800
- 进度：继续通过 `$grill-with-docs` 收口软切换成功后的反馈合并策略。已确认连续切集场景里只保留最后一次真正生效的成功短反馈，旧反馈全部取消或被覆盖，不让中心提示堆成过期日志。下一步继续细化软准备失败或成功之后，用户按 BACK 时应该回到详情页、留在播放器还是先关闭中心提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:27 +0800
- 进度：继续通过 `$grill-with-docs` 收口 soft preparing 成功后的反馈语义。已确认新目标真正切换成功后可以在播放器中心区给出极短的轻提示，用于确认“已切到第 N 集/目标视频”；该提示自动消失、不抢焦点、不阻断播放，也不再弹二次确认层。下一步继续细化快速连续切集时，成功短反馈是否要合并、覆盖或直接省略。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 15:25 +0800
- 进度：继续通过 `$grill-with-docs` 收口播放器软准备期间的承接方式。已确认旧分集/旧视频在 soft preparing 期间沿用当前播放态：切换前若正在播放，则继续播到新目标真正切换成功；切换前若处于暂停，则保持当前帧等待切换完成，不额外做人为冻结。下一步继续细化新目标准备成功后，是立即无感切过去，还是需要一个极短的“即将切换到第 N 集”确认反馈。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

## 2026-06-17 14:55 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV App 下一轮“运行流畅”目标，已确认后续主线聚焦 `TV 长视频播放器软准备`。同时已补充“运行流畅”覆盖整 App 的遥控输入响应、页面切换、列表滚动、焦点移动与刷新不卡顿，不只指视频播放本身。该语义进一步收口为：单片播放器与剧集播放器在已有播放内容后切换分集、重试准备播放源或重建播放链路时，不再退回整页 loading/error，而是优先保留当前承接画面、返回路径与焦点语义，只在播放器内部表达轻量 preparing/failed 状态；其中电视剧切集失败时，旧分集继续承接当前播放/暂停画面，只提示目标分集准备失败，目标分集准备成功后才真正切换；这类轻量状态统一挂在播放器中心区，复用现有中心反馈层级；failed 状态至少保留“重试切换/重试播放”主动作，电视剧可额外提供“留在当前分集”之类关闭路径；并确认剧集播放器需要拆分“目标分集”和“当前实际播放分集”两套状态，避免 rail 高亮和真实承接画面打架；当两套状态短暂不一致时，标题、当前播放文案、历史上报和续播逻辑统一跟随实际播放分集；rail 则同时显示“实际播放分集”的主态与“目标分集 preparing”的次级态；准备中的连续选集按 latest-wins 处理，旧目标请求过期。下一步继续细化软准备期间旧分集是继续播音视频，还是先冻结承接画面。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补播放器定向红灯测试、TV 全量验证与版本递增。

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

## 2026-06-17 14:18 +0800
- 进度：开始落实 TV 详情页软刷新实现，范围收口为电视剧详情页与长视频详情页两条已有内容重载链路。计划先补 ViewModel 红灯测试锁定“已有内容不退回整页 loading、失败保留旧内容、旧响应不覆盖新状态、电视剧详情保留当前选季选集并按既定回退链路收口”，再最小修改状态机和页内轻量状态挂载点。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`、`TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行详情页定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 14:06 +0800
- 进度：继续通过 `$grill-with-docs` 收口详情页软刷新轻量状态的结构落点。已确认已有内容后的轻量刷新/失败状态挂在详情主体内部：长视频详情页放入底部 `TvDetailGlassPanel`，电视剧详情页放入左侧 hero 或右侧剧集面板，而不是页面中央整页遮罩。`TV 详情页软刷新轻量状态挂载点` 已同步沉淀，下一步继续细化电视剧详情页到底优先挂在左侧 hero 还是右侧剧集面板。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补详情页定向红灯测试与 TV 验证。

## 2026-06-17 14:02 +0800
- 进度：继续通过 `$grill-with-docs` 收口详情页软刷新状态展示形态。已确认详情页已有内容后的刷新/失败复用现有 TV 页内轻量状态语言，不新造第二套大面板状态系统；首屏无内容才继续走 `TvPageLoadingState` / `TvErrorState`。`TV 详情页软刷新复用页内轻量状态` 已同步沉淀，下一步继续确认这类轻量状态在长视频详情页和电视剧详情页中的具体落点区域。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补详情页定向红灯测试与 TV 验证。

## 2026-06-17 13:59 +0800
- 进度：继续通过 `$grill-with-docs` 收口电视剧详情页软刷新失效回退语义。已确认当当前季/集在新详情数据里失效时，回退链路按“先保当前季、再全局偏好、最后第一季首个可播集/首集”的稳定顺序执行，不允许刷新后跳到不可预测的分集。`TV 电视剧详情软刷新失效回退链路` 已同步沉淀，下一步继续确认详情页页内附加状态是复用现有状态组件，还是新造一套局部提示样式。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补详情页定向红灯测试与 TV 验证。

## 2026-06-17 13:56 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 详情页软刷新焦点语义。已确认长视频详情页和电视剧详情页在已有内容完成软刷新后，不主动把焦点抢回播放按钮；播放按钮只在首屏第一次有内容时请求初始焦点，后续软刷新应优先保持用户当前焦点目标。`TV 详情页软刷新不抢主操作焦点` 已同步沉淀，下一步继续细化当电视剧详情页当前选中季/集在新数据中失效时的回退策略。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补详情页定向红灯测试与 TV 验证。

## 2026-06-17 13:53 +0800
- 进度：继续通过 `$grill-with-docs` 收口 TV 下一轮“运行流畅”目标，当前聚焦详情页状态机。已确认详情页下一优先级是“已有内容不硬切”的软刷新语义：长视频详情页和电视剧详情页在已有内容时，重载失败保留旧内容并只显示页内轻量错误；电视剧详情页重载成功后，若当前季/集在新数据里仍存在，则继续保持当前选择，不重新跳回默认偏好分集。已同步沉淀 `TV 详情页软刷新` 与 `TV 电视剧详情软刷新保留当前选集` 术语，下一步继续收口长视频详情页在软刷新成功时是否还要保留用户当前焦点目标。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补详情页定向红灯测试与 TV 验证。

## 2026-06-17 11:55 +0800
- 进度：开始实现 Android TV App IPTV 软刷新状态机。本轮仅作用于现有 `reload()` 链路，不新增正常播放中的显式刷新入口；先补红灯测试锁定“已有频道时 reload 不退回整页 loading、刷新成功保留当前频道、旧刷新响应不能覆盖新状态、刷新失败不打断当前播放”，再最小修改 IPTV ViewModel 与测试支撑。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`TvIptvNavigationPolicyTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 IPTV 定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 12:08 +0800
- 进度：完成 IPTV 软刷新核心实现。`TvIptvViewModel.reload()` 改为区分首屏加载与已有数据软刷新：已有频道或当前播放存在时只进入 `refreshing`，不退回整页 `loading`；刷新成功优先按当前 `channelId` 续留当前频道，只有当前频道失效时才回退默认首台；刷新失败时若旧列表仍可用则保留当前播放和旧列表，仅更新错误提示；并新增请求版本保护，避免旧刷新响应覆盖新状态。同步补充 IPTV 定向单测与纯函数测试，TV 版本递增到 `0.1.117` / `versionCode=117`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`TvIptvNavigationPolicyTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvViewModelTest --tests com.chee.videos.feature.tv.TvIptvNavigationPolicyTest` 通过；待执行 TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 11:57 +0800
- 进度：完成本轮 IPTV 软刷新收尾验证。TV 全量单测通过，文本检查通过；首次把 `:tv-app:testDebugUnitTest` 与 `:tv-app:assembleDebug` 并行跑时，`assembleDebug` 曾在 `:tv-app:kaptGenerateStubsDebugKotlin` 命中 `R.jar` `ClasspathEntrySnapshotTransform` `Check failed`，随后串行单独重跑 `:tv-app:assembleDebug` 已通过，判断为并行 Gradle 任务争用中间产物导致的瞬时构建问题，不是本次 IPTV 状态机回归。当前提交面仅纳入 IPTV 软刷新状态机、相关测试、版本递增与文档记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`TvIptvNavigationPolicyTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvViewModelTest --tests com.chee.videos.feature.tv.TvIptvNavigationPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app/...` 无命中。

## 2026-06-17 11:46 +0800
- 进度：继续通过 `$grill-with-docs` 收口 IPTV 刷新失败语义。已确认频道列表软刷新失败时，只要旧列表或当前播放仍可用，就保留当前播放和旧列表，仅显示轻量错误提示与重试入口；只有首屏尚无任何频道数据时才退回整页错误态。`IPTV 频道列表软刷新` 已补充失败边界，下一步继续确认是否需要请求身份来拦截旧刷新响应覆盖新状态。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补红灯测试与 TV 定向验证。

## 2026-06-17 11:43 +0800
- 进度：继续通过 `$grill-with-docs` 收口 Android TV App 下一轮“运行流畅”目标，先锁定 IPTV 高频播放路径。已确认 IPTV 也纳入同一流畅标准；频道列表刷新收口为软刷新语义：已有播放和已有列表保持可见，刷新完成后若当前频道仍在新列表里则继续保持当前频道，不回跳默认首台。已同步沉淀 `当前播放频道` 与 `IPTV 频道列表软刷新` 语义，下一步继续细化刷新失败与重试边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 收口阶段暂不执行 Android 构建；待进入实现后补红灯测试与 TV 定向验证。

## 2026-06-17 11:00 +0800
- 进度：完成 Android TV App 第五轮评审优化收尾。本轮继续以“运行流畅”为首要目标，海报墙刷新/切换排序改为软更新：已有内容保持可见，更新期间只显示网格内行内状态，失败时在网格内给出紧凑错误条并暂停自动分页，避免空白/跳页/混排；独立复审指出软更新完成后首焦点会因 `refreshing` 变化回跳首图，已改为仅在首批内容首次到达时自动聚焦一次，软刷新/软排序不再抢焦点。TV 版本递增到 `0.1.116` / `versionCode=116`，并沉淀 `TV 海报墙软更新` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`TvPosterWallFocusLayoutSpecTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPosterWallViewModelTest --tests com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' CONTEXT.md plan.md android-tv-app` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 10:44 +0800
- 进度：开始 Android TV App 第五轮整体评审优化。本轮继续以“运行流畅”为首要目标，优先审查 TV 高频遥控滚动路径；初步锁定海报墙刷新/排序会把已有内容切到加载语义或清空网格，容易造成焦点丢失、页面跳动和重复感知等待。计划先补红灯测试，收口为“已有内容保持可见，更新状态行内反馈，新结果回来后再替换”。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行海报墙定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 10:35 +0800
- 进度：完成 Android TV App 第四轮评审优化收尾。本轮以“运行流畅”为首要目标，修复 TV 首页搜索输入每个字符立即远程请求、搜索页被页面级 loading 替换的问题：搜索输入改为短 debounce，等待和网络请求期间保留搜索页并显示行内搜索状态；搜索失败在搜索页内显示错误与重试，不再误显示为空结果；重试会取消待发 debounce，避免重复请求。独立复审指出的搜索失败无页内重试入口、旧搜索响应测试覆盖不足均已修复。TV 版本递增到 `0.1.115` / `versionCode=115`，并沉淀 `TV 搜索输入流畅性` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`TvCatalogFocusPolicyTest.kt`、`TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest` 曾因搜索立即请求/页面 loading 失败；修复后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest --tests com.chee.videos.feature.tv.TvHomeNavigationTest --tests com.chee.videos.feature.tv.TvCatalogFocusPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 10:16 +0800
- 进度：开始 Android TV App 第四轮整体评审优化。本轮将“运行流畅”置于首要目标，优先审查会造成遥控输入卡顿、整页 loading 闪烁、重复网络请求或焦点抖动的路径；初步锁定 TV 首页搜索输入每个字符立即发请求并切全局 `loading=true`，会打断搜索界面并产生请求风暴，计划先补红灯测试再做最小修复。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`TvCatalogScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 目录/搜索定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 09:38 +0800
- 进度：完成 Android TV App 第三轮评审优化收尾。本轮聚焦 TV 首页目录/搜索旧请求覆盖：类型化首页与搜索请求均带目录请求身份，切换菜单、进入搜索/设置/IPTV、清空搜索或连续输入时会废弃旧请求，旧首页/搜索成功或失败不再覆盖当前菜单、当前查询、列表内容或错误态。独立复审指出真实 IPTV 入口绕过 `selectMenu(Iptv)`，已修为 IPTV 点击先通知 ViewModel 离开目录状态域再导航，且不把首页状态落成 IPTV 选中。TV 版本递增到 `0.1.114` / `versionCode=114`，并沉淀 `TV 首页目录请求身份` 约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`TvHomeNavigationTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest` 曾因旧请求覆盖失败；修复后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest --tests com.chee.videos.feature.tv.TvHomeNavigationTest --tests com.chee.videos.feature.tv.TvCatalogFocusPolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中；独立复审阻塞问题已修复。

## 2026-06-17 09:29 +0800
- 进度：开始 Android TV App 第三轮整体评审优化。按“功能正常、使用流畅”继续审查高风险异步链路；本轮先聚焦 TV 首页目录与搜索在快速切换菜单、连续输入搜索词时的旧请求覆盖问题，优先补红灯单测后做最小修复。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 目录定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 09:07 +0800
- 进度：完成 Android TV App 第二轮评审优化收尾。本轮聚焦电视剧播放器分集播放源准备失败恢复：源地址生成或偏好读取异常不再让播放页卡在“正在准备当前分集”，而是进入可重试的“暂不能播放”状态；保留当前分集身份，旧请求不覆盖新分集。提交仅纳入本轮电视剧播放源失败恢复、TV 版本递增、长期约束沉淀和计划记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-17 09:06 +0800
- 进度：完成第二轮核心修复。红灯用例确认电视剧播放器在播放源生成失败时无法进入可重试状态；现已让分集播放源准备协程捕获非取消异常，按请求序号和当前分集校验后退出 `playbackPreparing`，保留当前分集 `videoId`、清空播放 URL，并显示可重试的阻断文案。旧分集请求仍不会覆盖新分集。TV 版本递增到 `0.1.113` / `versionCode=113`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 通过；待执行相关定向组合、TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 09:02 +0800
- 进度：开始 Android TV App 第二轮整体评审优化。延续“功能正常、使用流畅”目标，重点检查首页目录/搜索、详情页、播放器与 IPTV 的加载失败恢复、遥控器焦点/返回路径和可见 UI 稳定性；优先处理源码能证明且可单测回归的问题。
- 影响文件：预计 `android-tv-app/**`、`CONTEXT.md`、`plan.md`
- 验证：待根据实际优化范围执行 TV 定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-17 00:25 +0800
- 进度：完成 Android TV App 整体评审第一轮优化收尾。本次聚焦“功能正常、使用流畅”的首启连接链路：异常可恢复、连接不可重入、长地址不挤压操作按钮；未改播放器内核、导航结构或 TV 编译边界。提交仅纳入本次 TV 连接优化、版本递增、长期约束沉淀和计划记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionViewModel.kt`、`ConnectionScreen.kt`、`core/repository/ServerRepository.kt`、`core/di/TvRepositoryModule.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.connection.ConnectionViewModelTest --tests com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-17 00:23 +0800
- 进度：完成第一轮 TV 首启连接链路优化。连接 ViewModel 通过 `ConnectionServerRepository` 接口隔离仓储，手动连接和历史地址连接均捕获探测/激活异常并恢复 `connecting=false`；连接中重复点击历史/发现地址会被忽略，连接页地址行限制单行省略，避免长 URL 挤压遥控器操作按钮。TV 版本递增到 `0.1.112` / `versionCode=112`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionViewModel.kt`、`ConnectionScreen.kt`、`core/repository/ServerRepository.kt`、`core/di/TvRepositoryModule.kt`、`android-tv-app/tv-app/build.gradle.kts`、相关连接单测、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.connection.ConnectionViewModelTest --tests com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest` 通过；待执行 TV 全量单测、assemble、`git diff --check` 与乱码扫描。

## 2026-06-17 00:13 +0800
- 进度：开始对 Android TV App 做整体评审与优化。先按 `$grill-with-docs` 校准 TV 工程定位与现有上下文，再审查模块边界、入口路由、播放器链路、焦点/10-foot UI、测试与构建配置；优先处理代码可证明的问题，避免无目标大重构。
- 影响文件：预计 `android-tv-app/**`、`CONTEXT.md`、`plan.md`
- 验证：待根据实际优化范围执行 TV 定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 23:39 +0800
- 进度：完成 TV 长视频播放器内部黑色背板修复。`TvLongFormMedia3Player` 不再在 SurfaceView 背后加载详情/剧集海报，也不再使用页面渐变背板；所有长视频 Media3 播放器内部统一铺纯黑背板，避免 DV/HDR SurfaceView 透明、未铺满或重建时露出后方背景。DV/HDR 仍使用 SurfaceView，普通 SDR 仍按路由使用 TextureView。TV 版本递增到 `0.1.111` / `versionCode=111`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormExoPlayerSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.tv.TvLongFormPlayerNavigationSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 23:35 +0800
- 进度：开始修复 TV 长视频 DV 播放时能看到播放器后方详情/海报背景的问题。根据现有代码，SurfaceView 模式下 `TvLongFormMedia3Player` 主动在播放器背后铺了详情/剧集海报背板；用户实际需要的是播放器内部黑色背板，而不是退出时黑色遮罩或人为延迟。本次将保留 DV/HDR 的 SurfaceView 输出，只把播放器背后承接层改回纯黑，避免播放中透明/未铺满区域漏出背景。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、单片/剧集调用点、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补/改源码级回归测试并执行 TV 定向单测、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 23:23 +0800
- 进度：完成 TV 长视频 DV 退出后延迟闪黑修复。`TvShellApp` 对单片长视频播放器路由和电视剧分集播放器路由禁用 NavHost 目的页 enter/exit/pop 过渡，避免播放器 SurfaceView/ExoPlayer 在详情页背后延迟释放；保留 DV/HDR 使用 SurfaceView，普通 SDR 使用 TextureView 的输出层策略。TV 版本递增到 `0.1.110` / `versionCode=110`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvLongFormPlayerNavigationSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.tv.TvLongFormPlayerNavigationSpecTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 23:17 +0800
- 进度：开始修复 TV 长视频 DV 播放退出到详情页后延迟 1-2 秒闪黑的问题。根据现象判断，黑闪发生在详情页已经显示之后，更像长视频播放器目的页被 NavHost 过渡保留、随后 SurfaceView/ExoPlayer 释放触发的延迟显示链路切换；本次不再尝试透明黑幕或背景遮罩，也不把 DV 改回会异色的 TextureView，而是收口长视频播放器路由的导航退出时序。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、TV 播放相关单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补源码级回归测试并执行 TV 定向单测、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 23:03 +0800
- 进度：完成 TV 长视频 DV 异色与退出闪黑联合修复。`TvPlaybackRoute` 增加输出层语义：普通长视频使用 TextureView，DV source/output 使用 SurfaceView 保持系统 HDR/DV 直出；`TvLongFormMedia3Player` 在 SurfaceView 模式内部铺当前详情/剧集海报背板，避免 SurfaceView 销毁瞬间露黑。没有恢复长视频 LibVLC，也没有新增播放器内核选择。TV 版本递增到 `0.1.109` / `versionCode=109`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`TvLongFormMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 layout/测试、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 23:02 +0800
- 进度：修正 DV/HDR 异色回归方案边界。DV 风险源继续使用 ExoPlayer，但输出层强制 `surface_view` 保持系统 HDR/DV 直出；普通 SDR 才使用 `texture_view` 缓解退出闪黑。为避免 DV SurfaceView 销毁瞬间露出黑底，`TvLongFormMedia3Player` 在 SurfaceView 模式下内部铺当前详情/剧集海报背板承接退出，不使用黑色遮罩或人为延迟。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`TvLongFormMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 layout/测试、`CONTEXT.md`、`plan.md`
- 验证：待重新执行 TV 定向单测、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 22:55 +0800
- 进度：开始修复 TV 长视频 TextureView 输出导致 DV/HDR 异色回归。结论：长视频仍只有统一 ExoPlayer 内核，不恢复 LibVLC；但 DV 风险源不能使用 TextureView 输出层，必须回到 SurfaceView 以保留系统 HDR/DV 直出色彩，普通 SDR 长视频才使用 TextureView 缓解退出闪黑。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、播放路由/测试、TV 版本文件、`CONTEXT.md`、`plan.md`
- 验证：待补定向测试并执行 TV 定向单测、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 22:50 +0800
- 进度：完成 TV 长视频退出播放闪黑修复。长视频 ExoPlayer 不再使用 `PlayerView` 默认 `SurfaceView`，改为专用 `tv_long_form_media3_player_view.xml` 固定 `surface_type="texture_view"`、透明 shutter 并保持 reset 内容；这是输出层合成策略调整，不新增 LibVLC/ExoPlayer 选择，也不恢复长视频 LibVLC 分支。TV 版本递增到 `0.1.108` / `versionCode=108`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_long_form_media3_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormExoPlayerSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.player.TvLongFormExoPlayerSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' android-tv-app CONTEXT.md plan.md` 无命中。

## 2026-06-16 22:45 +0800
- 进度：开始排查 TV 长视频退出播放回详情页时闪黑问题。初步怀疑是 ExoPlayer/PlayerView 在返回导航完成前释放 Surface 或显示黑色 shutter，导致播放器页黑底露出一帧；先定位返回路径、播放器释放时机和详情页承接方式，再做最小修复。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补定向回归测试并执行 TV 定向单测、必要时 `:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

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

## 2026-06-16 20:44 +0800
- 进度：开始实现 TV 长视频统一迁到 ExoPlayer。文档决策已单独提交 `8f034bf`；实现阶段按已确认节奏推进：先补红灯测试约束，再迁移单片与剧集播放器，清理长视频 LibVLC 依赖但保留 IPTV LibVLC，最后运行 TV 定向/全量验证和独立评审。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/*`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补红灯测试后执行 TV 定向单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

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
- 进度：确认 TV 工程迁移后继续保留 `libvlc-all`，但用途限定为 IPTV。TV 长视频不得继续依赖 `TvVlcLibrary`、`newLongFormMediaPlayer` 或长视频 LibVLC helper；长视频相关 LibVLC URL、`addSlave`、`audioTrack=-1` 等规则标记为迁移前/历史或 IPTV 排障语境。后续单测需反向约束长视频屏不再 import `org.videolan.libvlc`，IPTV 仍允许。
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

## 2026-06-16 20:12 +0800
- 进度：开始通过 `$grill-with-docs` 收口“TV App 除 IPTV 外播放器统一迁到 ExoPlayer”。已确认范围只覆盖 TV 端实际运行的单片长视频播放器、电视剧分集播放器和既有 DV Media3 分支；IPTV 继续保留 LibVLC；TV 工程已排除的手机端播放器、短视频和图片合集不纳入本次。
- 影响文件：预计 `CONTEXT.md`、`plan.md`，后续实现可能涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/*`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`
- 验证：收口阶段暂不执行构建；后续实现前补测试并执行 TV 定向/全量验证。

## 2026-06-16 20:04 +0800
- 进度：完成 TV 端 DV Media3 返回详情去黑场修复与收尾。确认退出确认仍保留，但确认后的动作不再进入 App 层黑色遮罩或 80ms 延迟；单片、剧集系统返回和剧集结束覆盖层“返回详情”均直接走正常返回。`CONTEXT.md` 已将旧“DV 返回详情黑场保持”标记为已撤回，并新增“DV 返回详情正常导航”约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 19:57 +0800
- 进度：完成 DV Media3 返回详情去黑场的核心改动。单片与剧集播放器的退出确认不再区分 Media3 DV 分支，二次确认后直接 `onBack()`；删除 `TvDolbyVisionExitToDetailCover` 和 80ms 延迟；剧集播放结束覆盖层的“返回详情”也改为直接返回；回归测试已反向锁定不得再保留黑色遮罩/延迟。TV 版本已递增到 `0.1.106` / `versionCode=106`，`CONTEXT.md` 已撤回旧黑场保持策略。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest`、必要的 TV 单测/编译、`git diff --check` 与乱码扫描。

## 2026-06-16 19:57 +0800
- 进度：开始修复 TV 端 DV Media3 播放器退出到详情页会主动黑一下的问题。已定位为单片和剧集播放器在二次返回确认后设置 `pendingDvExitToDetail`，显示 `TvDolbyVisionExitToDetailCover` 黑色遮罩并延迟 80ms 才执行 `onBack()`；本次目标改为正常立即返回详情页，不再主动制造黑场。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、移除 DV 退出遮罩文件、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待修改后执行 TV Media3 定向单测、必要的 TV 单测/编译、`git diff --check` 与乱码扫描。

## 2026-06-16 19:39 +0800
- 进度：完成 TV 电视剧详情页演员头像负 hash 越界修复与收尾。`tvCastAvatarBrush` 现在通过 `tvCastAvatarPaletteIndex()` 使用 `Math.floorMod` 计算调色板索引，新增 `TvSeriesCastAvatarPaletteTest` 覆盖负 hash、`Int.MIN_VALUE` 与空调色板边界；TV 版本已递增到 `0.1.105` / `versionCode=105`，`CONTEXT.md` 已沉淀头像调色板索引契约。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesCastAvatarPaletteTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesCastAvatarPaletteTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 19:26 +0800
- 进度：完成 TV 电视剧详情页演员头像负 hash 越界核心修复。`tvCastAvatarBrush` 改为通过 `tvCastAvatarPaletteIndex()` 使用 `Math.floorMod` 解析调色板索引，新增 `TvSeriesCastAvatarPaletteTest` 覆盖 `-4`、`Int.MIN_VALUE` 和边界种子；TV 版本递增到 `0.1.105` / `versionCode=105`，`CONTEXT.md` 沉淀头像调色板索引不得直接 `%` 取模。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesCastAvatarPaletteTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesCastAvatarPaletteTest`、`git diff --check` 与乱码扫描。

## 2026-06-16 19:26 +0800
- 进度：开始修复 TV 电视剧详情页演员头像兜底渐变在负 hash 下闪退的问题。崩溃点为 `TvSeriesDetailScreen.kt` 中 `tvCastAvatarBrush(actor.id.hashCode())` 直接用 `%` 对调色板取模，Java/Kotlin 对负数取模会得到负余数，导致 `colors[-4]` 越界。
- 影响文件：预计 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、新增 TV 端定向测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补负 hash 回归测试后执行 TV 端定向单测、`git diff --check` 与乱码扫描。

## 2026-06-16 17:33 +0800
- 进度：完成推送到家用部署机。本地 `master` 已从部署机远端 `2c69da3` 推进到 `d4fe1f7`；远端 hook 判定 `RESTART_GO=0 REBUILD_FRONTEND=1`，执行 `npm ci && npm run build` 后更新 `current/admin-web-dist`，未重启 Go server / worker。GitHub mirror push 在远端 hook 内失败但标记为 non-fatal，不影响部署机更新。
- 影响文件：`plan.md`
- 验证：`git push deploy master` 通过；远端 hook 输出 `frontend updated` 与 `deploy ... done`；`git ls-remote deploy refs/heads/master` 返回 `d4fe1f7713f5d6ac1b5db9029ce4903bac6051ab`。

## 2026-06-16 17:32 +0800
- 进度：开始推送管理端时间显示格式化到家用部署机。当前部署机 `deploy/master` 为 `2c69da3`，本地 `master` 最新为 `767c13b`；本次变更命中 `admin-web/**`，按 [[家用部署机 push 分桶]] 预期只触发管理端前端构建与 dist 替换，不重启 Go server / worker。
- 影响文件：`plan.md`
- 验证：待执行 `git push deploy master`，并确认远端 hook 的 admin-web build 成功。

## 2026-06-16 17:29 +0800
- 进度：完成管理端绝对时间显示格式化收尾。独立子代理评审未发现阻塞问题；按评审建议补充数字字符串时间戳兼容和边界测试。确认本次只纳入管理端时间格式化、长期语义沉淀和计划记录。
- 影响文件：`admin-web/src/utils/dateTime.js`、`admin-web/src/utils/dateTime.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、相关管理端时间展示页面、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/utils/dateTime.spec.js src/views/adminDateTimeDisplay.spec.js src/views/toolboxOrphanFiles.helpers.spec.js` 通过；`cd admin-web && npm run test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中；排除测试文件后的绝对时间 raw prop / `toLocaleString` 覆盖搜索无命中。

## 2026-06-16 17:18 +0800
- 进度：完成管理端绝对时间显示格式化实现。新增共享 `formatAdminDateTime`，将用户、视频、图片、合集、演员、任务监控、IPTV、安装包管理、孤儿文件扫描、图像生成历史等可见绝对时间统一切到 `YYYY/MM/DD HH:mm:ss`；保留业务日期和时长原格式，并补充源码级回归测试防止时间戳列 raw prop 直出。
- 影响文件：`admin-web/src/utils/dateTime.js`、`admin-web/src/utils/dateTime.spec.js`、`admin-web/src/views/adminDateTimeDisplay.spec.js`、`admin-web/src/views/*` 时间展示相关文件、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/utils/dateTime.spec.js src/views/adminDateTimeDisplay.spec.js src/views/toolboxOrphanFiles.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-16 17:05 +0800
- 进度：开始收口并实现管理端绝对时间显示格式化。已通过 `$grill-with-docs` 确认边界：`created_at`、`updated_at`、`started_at`、`published_at`、`mod_time` 等绝对时间戳统一显示为 `YYYY/MM/DD HH:mm:ss`；`YYYY-MM-DD` 业务日期、剩余时间和播放进度等时长不纳入本次格式化。
- 影响文件：预计 `admin-web/src` 下管理端时间展示相关页面/工具、必要前端测试、`CONTEXT.md`、`plan.md`
- 验证：待执行管理端定向测试、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-16 15:52 +0800
- 进度：完成部署机系统日志路径落地。已推送 `2c69da3` 到 `deploy master`；远端 hook 按分桶跳过 Go/frontend 构建。随后在部署机 `~/deploy/ai-video-server/.env` 写入 `SERVER_LOG_PATH=/Users/chee/Library/Logs/ai-video-server/server.log`，重启 `com.aivideo.server`，确认 `/healthz` 正常，系统日志接口返回真实日志行。
- 影响文件：`plan.md`；部署机运行态配置 `~/deploy/ai-video-server/.env`
- 验证：`git push deploy master` 通过；远端 `launchctl kickstart -k gui/$(id -u)/com.aivideo.server` 后 `curl http://127.0.0.1:8080/healthz` 返回 `{"status":"ok"}`；使用 60 秒临时 admin token 请求 `/api/v1/admin/system/logs?lines=5` 返回 `code=0`、`line_count=5`。

## 2026-06-16 15:39 +0800
- 进度：开始补齐部署机系统日志路径契约。已定位根因：服务端默认 `SERVER_LOG_PATH=./.run/server.log` 只适合 dev 模式，而家用部署机 launchd 实际写入 `~/Library/Logs/ai-video-server/server.log`；部署 `.env` 若不显式设置 `SERVER_LOG_PATH`，管理端系统日志接口会读错路径并返回空态。
- 影响文件：`.env.example`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行文档/配置 diff 检查、乱码扫描；部署后需确认部署机 `.env` 生效并重启 server。

## 2026-06-16 14:52 +0800
- 进度：完成系统日志空态修复收尾。确认本次只纳入系统日志接口、回归测试、长期语义沉淀和计划记录；无关工作区变更不存在。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/admin_system_logs_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run TestAdminSystemLogs -count=1` 通过；`go test ./internal/handlers -count=1` 通过；`git diff --check -- internal/handlers/admin.go internal/handlers/admin_system_logs_test.go CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-16 14:49 +0800
- 进度：完成系统日志空态修复。`/admin/system/logs` 在 `server.log` 尚未生成时返回空日志列表；下载模式返回 200 空文本文件；路径存在但无法读取时仍保留原错误分支。补充 handler 回归测试覆盖刷新和下载两种空态，并把长期约定写入 `CONTEXT.md`。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/admin_system_logs_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run TestAdminSystemLogs -count=1` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-06-16 14:44 +0800
- 进度：开始修复系统日志刷新在 `./.run/server.log` 尚未生成时报错的问题。范围收口为后端 `/admin/system/logs` 接口：日志文件缺失应返回空日志空态，不应让管理端刷新失败；其它日志读取错误仍按错误返回。
- 影响文件：`internal/handlers/admin.go`、预计新增后端日志接口测试、`CONTEXT.md`、`plan.md`
- 验证：待补红灯测试后执行 `go test ./internal/handlers -run TestAdminSystemLogs -count=1`、`git diff --check` 与乱码扫描。

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

## 2026-06-16 14:10 +0800
- 进度：继续收口孤儿文件扫描工具箱迁移。已确认孤儿文件扫描应与 ED2K、图像生成工作台一致，做成 `/toolbox/orphan-files` 无 shell 独立工具页，从工具箱以新标签页打开；系统设置页不再承载扫描 UI。
- 影响文件：`CONTEXT.md`、`plan.md`、后续 `admin-web/src/views/ToolboxOrphanFiles.vue`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/Toolbox.vue`、路由与测试。
- 验证：待继续确认删除提示触发时机后进入前端实现和验证。

## 2026-06-16 14:07 +0800
- 进度：开始通过 `$grill-with-docs` 收口“孤儿文件扫描迁到工具箱”。已确认本次是前端入口和页面归属迁移，不改后端 `/admin/system/orphan-files/*` API、扫描队列、删除逻辑或数据表；术语同步更新为工具箱独立工具页复用既有系统接口。
- 影响文件：`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxOrphanFiles.vue`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/views/toolboxOrphanFiles.helpers.js`、相关前端测试、`CONTEXT.md`、`plan.md`
- 验证：待需求收口后补前端定向测试，再执行 `cd admin-web && npm run test -- src/views/toolboxOrphanFiles.helpers.spec.js src/views/toolboxPage.spec.js src/components/base/commandPalette.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-16 12:12 +0800
- 进度：补充后端回归验证并完成收尾检查。`internal/repository` 与 `internal/models` 的定向测试已通过，确认电视剧详情 `cast` 实体化后仓库层和模型层仍可正常编译/运行。
- 影响文件：`plan.md`
- 验证：`go test ./internal/repository ./internal/models -count=1` 通过。

## 2026-06-16 12:10 +0800
- 进度：完成 TV 电视剧详情演员头像改造的收口与验证。演员区保持横向滚动，仅展示头像和姓名，不提供点击跳转；头像优先走服务端下发的本地可访问路由，缺失或加载失败时回落首字。顺手补齐 `TvTestSupport` 的演员 DTO 导入，并把演员区“只浏览不点击”写入 `CONTEXT.md`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 11:26 +0800
- 进度：开始通过 `$grill-with-docs` 收口 TV 电视剧详情演员头像改造。已确认当前服务端 `TvSeriesDetailDto.cast` 只返回演员姓名，TV 端因此只能渲染圆形首字占位；本次方向调整为详情接口返回演员实体（含本地头像路由），TV 端加载头像，缺失时保留首字兜底。
- 影响文件：`internal/models/app.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待完成需求收口后补后端/TV 定向测试，再执行相关 Go/TV 验证、`git diff --check` 与乱码扫描。

## 2026-06-16 11:11 +0800
- 进度：完成电视剧刮削幂等与 TMDB 演员复用优化。已实现同一电视剧整剧 poster/backdrop 和分集 still 的本地存在跳过下载，新增 TMDB 演员按 `source + external_id` 复用已完整资料的查询；回归测试覆盖连续刮削同一剧不同分集和已有完整 TMDB 演员再次绑定视频两条路径。
- 影响文件：`internal/services/scraper.go`、`internal/repository/actor_repository.go`、`internal/services/scraper_episode_sync_test.go`、`internal/services/scraper_test.go`、`internal/queue/scrape_tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestScrapeEpisodeUploadSkipsExistingSeriesArtworkOnNextEpisode|TestSyncMovieActorsDoesNotOverrideExistingAvatarOrNotes' -count=1` 通过；`go test ./internal/services -run 'TestScrapeEpisodeUpload|TestSyncMovieActors' -count=1` 通过；`go test ./internal/repository -run TestBuildUpsertScrapedActorProfileSQL -count=1` 通过；`go test ./internal/repository -count=1` 通过；`go test ./internal/queue -run 'TestAutoScrapeAVMarksReadyOnSuccess|TestAutoScrapeAVStoresLocalizedFieldsAndMarksReady' -count=1` 通过；`git diff --check` 通过；乱码扫描无命中。`go test ./internal/services -count=1` 仍存在既有失败 `TestParseTVAPKMetadataParsesReleaseAPK`（`version_code` 期望 80、实际 99），与本次刮削优化无关。

## 2026-06-16 11:04 +0800
- 进度：完成刮削策略代码定位，准备补充重复图片下载与重复 TMDB 演员详情请求的回归测试；方案为电视剧集图片按本地文件幂等跳过，TMDB 演员按 `source + external_id` 查到完整资料后直接复用。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_episode_sync_test.go`、`internal/services/scraper_test.go`、`internal/repository/actor_repository.go`、`internal/queue/scrape_tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行定向 Go 测试、`go test ./internal/services -count=1`、`git diff --check` 与乱码检查。

## 2026-06-16 11:00 +0800
- 进度：开始优化电视剧剧集刮削的重复下载/重复演员刮削问题。范围收口为：同一 TMDB 剧重复刮削不同分集时，已落盘的整剧 poster/backdrop 和分集 still 不再重复下载；TMDB 演员同步优先复用已有完整演员资料，避免已有头像/资料的演员每集重复请求详情或重复下载头像；保持 series/season/episode/actor 现有 upsert 语义不引入 migration。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_episode_sync_test.go`、`internal/services/scraper_test.go`、`internal/repository/actor_repository.go`、`CONTEXT.md`、`plan.md`
- 验证：待补红灯测试后执行 `go test ./internal/services -run 'TestScrapeEpisodeUpload|TestSyncTMDBCastActors' -count=1`、`go test ./internal/repository -run TestBuildUpsertScrapedActorProfileSQL -count=1`、`go test ./internal/services -count=1`、`git diff --check` 与乱码扫描。

## 2026-06-16 10:07 +0800
- 进度：完成管理端短视频审核页分页与播放器尺寸调整。左侧待审队列改为当前页分页列表，复用 `AdminTablePagination` 支持页码输入跳转，并显示“第 X / Y 页 · 共 N 条”；`保留并下一条` / `删除并下一条` 在页末自动进入下一页第一条，删除后刷新当前页避免 offset 分页补位视频被跳过。右侧播放器增加 9:16 手机竖屏预览框，限制在页面可用高度内，视频本身保持 contain 不裁切。`CONTEXT.md` 已更新长期术语和分页/播放器约定。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/ShortReview.vue admin-web/src/views/shortReview.helpers.js admin-web/src/views/shortReview.helpers.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-16 09:53 +0800
- 进度：开始修改管理端短视频审核页。通过 `$grill-with-docs` 收口术语为“短视频审核页”，本次把左侧待审队列从自动追加式长列表改为当前页分页列表，显示当前页、总页数和总条数，并支持页码输入跳转；`保留并下一条` / `删除并下一条` 在当前页末尾仍自动进入下一页第一条。右侧播放器收口为手机竖屏预览尺寸，整体高度不得超过审核页可用高度。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-16 09:34 +0800
- 进度：完成 DV Media3 音轨偏好覆盖问题修复。已补红灯测试锁定 `onTracksChanged` 不得把 Media3 默认选中音轨回写成父层选择；实现上删除 `onSelectedAudioTrackChanged` runtime 回调，轨道加载只上报可选音轨列表，父层继续按 current selection / stored preference 决定目标并通过 Media3 override 应用。TV 版本升至 `0.1.103 / 103`，`CONTEXT.md` 补充“Media3 默认音轨不等于用户选择”约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvMedia3TrackSupportTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest` 先红后绿；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-16 09:30 +0800
- 进度：开始修复 DV Media3 音轨偏好恢复被默认轨覆盖的问题。Review 发现 `onTracksChanged` 中先按保存偏好解析目标音轨，随后又把 Media3 当前默认选中的 runtime track id 回写给父层，可能把用户保存的中文/解说等偏好覆盖成底层默认轨。本次收口为：Media3 轨道加载时只上报可选音轨列表，由父层按 current selection / stored preference 决定目标；不再把底层默认选中轨道当作用户选择回灌。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待补回归测试后执行 TV 定向单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、必要的 `:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-16 09:17 +0800
- 进度：完成 TV 端 DV 专用链路第二阶段实现。Media3 播放器现在预加载服务端外挂字幕并用 `TrackSelectionParameters` 切换字幕/音轨，不重建播放源；单片和剧集 DV 分支都接入字幕/音轨 picker、语义偏好恢复和播放控制层入口。主动返回详情时先显示 App 层黑色遮罩并短延迟导航，遮住 Media3 surface 销毁和详情页重绘之间的可控闪动；TV 版本升至 `0.1.102 / 102`。本次只纳入 TV DV 第二阶段源码、测试、版本号、`CONTEXT.md` 和 `plan.md`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`TvMedia3TrackSupport.kt`、`TvMedia3TrackPickerLayer.kt`、`TvDolbyVisionExitToDetailCover.kt`、`TvLongFormPlayerScreen.kt`、`TvSeriesPlayerScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvMedia3TrackSupportTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。曾因并发 Gradle 构建触发 Kotlin 增量缓存 EOF，已 `clean` 后串行重跑通过。

## 2026-06-16 09:01 +0800
- 进度：继续收尾 TV 端 DV 专用链路第二阶段。当前编译断点是 `TvMedia3TrackPickerKind` / `TvMedia3TrackPickerLayer` 仍在剧集页面 private 作用域，单片 DV 页面无法复用；接下来抽成共享文件，并为单片与剧集 DV 返回详情入口加页面级黑场保持，避免退出瞬间露出上一帧或详情背景。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、预计新增 TV Media3 轨道弹层文件、相关 TV 单测、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:compileDebugKotlin`、TV 定向单测、全量 TV 单测、`assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-15 22:57 +0800
- 进度：完成 DV 第二阶段偏好粒度收口。用户确认电视剧 DV 分集的字幕/音轨偏好继续沿用现有 `videoId` 粒度，不新增整部剧共享偏好；音轨偏好仍按语言/类型等语义匹配，避免依赖底层 track id。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做文档沉淀，暂不运行构建或单测。

## 2026-06-15 22:54 +0800
- 进度：继续收口 DV 第二阶段的播放中切换连续性。用户确认外挂字幕切换也必须尽量无感，要求播放中不跳黑、不闪回、不重启播放会话，并尽量维持当前画面连续；这使得第二阶段验收从“字幕/音轨能力存在”提升到“字幕切换连续性也要对齐返回详情连续性”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做术语和验收边界沉淀，暂不运行构建或单测。

## 2026-06-15 22:50 +0800
- 进度：继续推进 TV 端 DV 第二阶段收口。已确认字幕选择沿用服务端外挂字幕列表、音轨来自 Media3 当前播放源枚举到的内嵌轨道，音轨偏好按语言/类型保存；同时单片和剧集两个 DV 分支都存在返回详情连续性问题，不能只按剧集场景修。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做代码/文档核对，暂不运行构建或单测。

## 2026-06-15 22:34 +0800
- 进度：继续收口 DV 第二阶段字幕/音轨边界。用户确认本阶段字幕/音轨范围采用“外挂字幕沿用服务端列表，音轨来自 Media3 当前播放源内嵌音频轨”，不承诺 LibVLC/libass 的内嵌字幕和 ASS 特效一致性；同时要求外挂字幕播放中切换尽量无感，不接受通过重建播放源导致明显黑屏或重启播放会话。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做术语和验收边界沉淀，暂不运行构建或单测。

## 2026-06-15 22:29 +0800
- 进度：继续收口 `DV 返回详情闪动`。已确认真机表现更像退出瞬间露出一帧视频或详情页背景，而不是单纯系统黑屏或白闪；因此下一步优先按 App 层返回过渡连续性处理，目标是不暴露上一帧、底图抢显、白闪或桌面露出。若修复后只剩电视/系统 DV/HDR 模式切换短黑，则归入显示链路边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做术语和边界沉淀，暂不运行构建或单测。

## 2026-06-15 22:25 +0800
- 进度：开始通过 `$grill-with-docs` 收口 TV 端 DV 下一阶段。真机测试确认上一阶段剧集混合播放主流程通过；新问题限定为 DV 专用链路“返回详情”时仍有闪动，不是播放结束或切下一集路径。已将术语收敛为 `DV 返回详情闪动`，后续需继续判断它属于 App 层可控过渡，还是电视/系统 HDR/Dolby Vision 模式切换边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段先做代码/文档核对与术语沉淀，暂不运行构建或单测。

## 2026-06-15 21:50 +0800
- 进度：完成 TV 电视剧 DV/非 DV 混合播放控制层收尾审查，确认本次只纳入 TV 播放控制层、Media3 seek 适配、混合切集进度语义、DV 失败态选集入口、版本号、ADR 与长期上下文沉淀。当前子代理工具要求用户显式授权委派，未新增独立 review 子代理；已改为本地抽查关键 diff 与测试覆盖后提交。
- 影响文件：`CONTEXT.md`、`docs/adr/0012-tv-series-shared-controls-for-mixed-playback-engines.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关 TV 单测、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesMixedPlaybackControlsSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md docs/adr android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 无命中。真机仍需回归非 DV->DV、DV->非 DV、连播切换、DV 失败页“选集”和实际电视 DV/HDR 退出黑屏边界。

## 2026-06-15 21:38 +0800
- 进度：完成 TV 电视剧混合播放控制层首轮实现。新增 `TvSeriesCorePlaybackOverlay`，DV Media3 分支现在叠加电视剧核心控制层，支持播放/暂停、遥控左右 seek、进度显示、续播卡、选集轨、连播提示和结束覆盖层；字幕/音轨入口在 Media3 分支隐藏。Media3 播放器增加受控 seek 请求入口；切集前主动上报当前集历史，且上一集历史保存按上一集 route 选择 LibVLC time 或 Media3 snapshot，避免 LibVLC<->Media3 切换污染进度。DV 播放失败页新增“选集”动作，只允许切到其它分集，不提供回退 LibVLC 强行播放。TV 版本升至 `0.1.101 / 101`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesMixedPlaybackControlsSpecTest.kt`、相关源码规格测试、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无命中。待独立 review 返回后修复问题并提交。

## 2026-06-15 21:13 +0800
- 进度：开始落地 TV 电视剧 DV/非 DV 混合播放控制层。实现策略：先新增电视剧专用共享控制层，只覆盖核心播放、seek、进度、标题覆盖、选集轨和挂载点；LibVLC 分支继续使用原视频面与字幕/音轨能力，Media3 分支复用同一控制层但隐藏字幕/音轨入口；同页内按分集 route 切换 LibVLC / Media3，切换前沿用当前历史上报逻辑。单片长视频默认控制变体暂不重构。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、预计新增 TV 核心控制层文件、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`docs/adr/0012-tv-series-shared-controls-for-mixed-playback-engines.md`、`plan.md`
- 验证：待补策略测试后执行 TV 定向单测、必要的 `:tv-app:testDebugUnitTest` / `:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-15 20:41 +0800
- 进度：继续收口 TV 端 DV 播放体验边界。已确认：DV 专用链路本阶段补齐核心播放与剧集互动 UI，包含播放/暂停、快进快退、进度反馈、返回二次确认、续播提示、选集轨和连播提示；字幕/音轨切换放到下一阶段。退出闪烁验收改为消除 App 可控闪烁（白闪、上一帧抖动、底图抢显、桌面露出、二次闪回），电视/系统因 Dolby Vision/HDR 模式切换产生的一次短黑属于显示链路边界。混合剧集允许在同一播放器会话内按分集切换 LibVLC / Media3，切换进度沿用既有连播/手动选集规则；DV 当前分集失败不退出剧集会话，也不允许回退 LibVLC 强行播放。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段只做文档沉淀，暂不运行构建或单测。

## 2026-06-15 20:28 +0800
- 进度：开始通过 `$grill-with-docs` 收口 TV 端 DV 播放体验问题。已核对 `CONTEXT.md` 与 ADR：普通长视频主链路是 LibVLC，DV 安全放行只进专用 Media3/ExoPlayer 链路；代码也确认剧集在 `isVlcRoute` 时使用完整 `LongFormVideoPlayer` 控制层，在 `isMedia3Route` 时只渲染 `TvDolbyVisionMedia3Player`，因此“DV 退出闪一下”“DV 缺少原 LibVLC 互动 UI”“混合剧集切集时内核切换”都应归入同一个播放会话/播放链路切换策略问题，而不是各自独立修补。下一步逐项确认：是否把 TV 播放器互动层抽成可覆盖 Media3 的共享控制层，以及混合剧集切换时的状态迁移边界。
- 影响文件：`plan.md`；已阅读 `CONTEXT.md`、`docs/adr/0010-dolby-vision-tv-playback-compatibility.md`、`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`、`docs/adr/0011-dolby-vision-episode-original-file-preservation.md`、TV 播放页相关代码。
- 验证：grill 阶段只做代码/文档核对，暂不运行构建或单测。

## 2026-06-15 20:15 +0800
- 进度：完成 TV App DV 源播放卡住修复。排查确认部署机 `source-dv.mkv` 存在、读盘正常，TV 已请求 `profile=dv_source` 且服务端返回 200/206；卡住主因是电视剧分集 DV Media3 route 准备好播放源后没有像 LibVLC route 一样自动进入播放会话，导致专用播放器只 prepare 不 play。现在分集 DV Media3 源准备完成会自动置为 started 并播放；Media3 专用播放器新增 15 秒启动/缓冲超时，卡住时进入专用播放失败页并允许重试，不回退 LibVLC。TV 版本升至 `0.1.100 / 100`，`CONTEXT.md` 已补充启动边界。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3PlayerTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvSplashScreenSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvDolbyVisionMedia3PlayerTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.core.player.TvLongFormVlcSpecTest --tests com.chee.videos.tv.TvSplashScreenSpecTest` 通过；`git diff --check`、乱码扫描待执行。

## 2026-06-15 20:09 +0800
- 进度：开始排查 TV App 播放 DV 源文件卡住。部署机日志确认 TV 端已请求 `/api/v1/videos/d5a89f74-2cdb-48e3-9246-189c3ba0c596/source?profile=dv_source`，服务端返回 200/206，`source-dv.mkv` 文件存在且本地读盘正常；当前风险集中在 TV 端 Media3/ExoPlayer 专用链路可能长期停在启动/缓冲状态但 UI 没有失败兜底。本次先补 DV 专用播放器启动超时与失败页，保持“不回退 LibVLC”的安全约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 端定向单测、必要的 `:tv-app:testDebugUnitTest`、`git diff --check` 与乱码扫描。

## 2026-06-15 19:52 +0800
- 进度：完成 DV 剧集直拷贝修复。`episode + Dolby Vision` 的转码任务现在会跳过 ffmpeg 压缩，直接把源文件迁移/复制到 `storageRoot/videos/<uuid>/source-dv.<ext>`，并把该路径同时写入 `videos.transcoded_path` 和 `playback_compat.source_playback_path`；任务完成时直接落 ready 状态，缩略图与播放兼容 metadata 仍照常生成。补了 queue 层单测，覆盖 direct copy 输出、DV 判定和“源文件已移动后重试复用稳定目标路径”的幂等恢复。
- 影响文件：`internal/queue/tasks.go`、`internal/queue/tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/queue -count=1` 通过；`go test ./internal/services -run 'Test(BuildPlaybackCompatibilityMetadata|SourcePlaybackPathFromMetadata|MergePlaybackCompatibilityMetadata|BuildTranscode|ResolveProbe|Transcode|IsSameFilePath)' -count=1` 通过；`go test ./internal/handlers ./internal/repository -run 'Test(ResolveProfiledPlayableSource|SelectRetranscodeInputPath|CollectLocalStoragePaths)' -count=1` 通过；`git diff --check` 通过；乱码扫描待执行。

## 2026-06-15 19:37 +0800
- 进度：开始修复 DV 剧集进入压缩程序的问题。最新要求收口为：只要转码任务源探测确认为 Dolby Vision，且视频类型为 `episode`，worker 应直接把源文件迁移/复制到 `source-dv.<ext>` 作为播放源并完成任务，不再启动 ffmpeg 压缩输出；同时写入 `playback_compat.source_playback_path`，保持 TV 端走 DV 源文件链路。范围锁定后端 queue/transcode 路径、相关单测、`CONTEXT.md` 与 `plan.md`。
- 影响文件：`internal/queue/tasks.go`、`internal/queue/tasks_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待先补红灯测试，再执行 `go test ./internal/queue -run 'TestHandleTranscode|TestShouldPreserveEpisodeDolbyVisionSource' -count=1`、相关 Go 定向测试、`go test ./internal/queue ./internal/services -count=1`、`git diff --check` 与乱码扫描。

## 2026-06-15 18:47 +0800
- 进度：完成短视频审核页单屏布局修复。`ShortReview.vue` 现在把页面高度锁定到 admin shell 可用视口内，工作台和左右面板使用 `minmax(0, 1fr)` 收缩；待审队列只在面板内部滚动，播放器去掉固定最小高度，随右侧面板可用高度缩放。`CONTEXT.md` 已补充短视频审核页单屏工作台约定，源码测试锁定页面不靠浏览器纵向滚动承载审核流。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`npm run test -- src/views/shortReview.helpers.spec.js` 通过；`npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/ShortReview.vue admin-web/src/views/shortReview.helpers.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-15 18:42 +0800
- 进度：开始收口短视频审核页单屏布局。通过 `$grill-with-docs` 对照现有 `admin 短视频清理审核` 语义，确认本次不改审核流程，只修布局：浏览器窗口不应出现纵向滚动，待审队列在左侧面板内部滚动；右侧视频区域最大高度不得超过审核页可用高度，避免播放器硬撑页面。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：已补红灯测试 `npm run test -- src/views/shortReview.helpers.spec.js`，当前失败点为页面尚未固定到 shell 可用高度；待执行修复后定向测试、`npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-15 18:34 +0800
- 进度：完成推送到家用部署机。已将本地 `master` 从部署机远端 `6d3619f` 推进到 `f3f67de`；远端 hook 完成 `admin-web` 构建、Go build、codesign、migration 与 server / worker `launchctl kickstart`，最终 `/healthz OK`。GitHub mirror push 在部署机 hook 内失败但标记为 non-fatal，不影响本次部署。
- 影响文件：`plan.md`
- 验证：`git push deploy master` 通过；`git ls-remote deploy refs/heads/master` 返回 `f3f67de9dd205f9e77f61bb9efff369aab8c96d4`；部署机 `curl http://127.0.0.1:8080/healthz` 返回 `{"status":"ok"}`。

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

## 2026-06-15 16:40 +0800
- 进度：开始修复管理端侧栏收起后无法再展开的问题。已确认折叠态样式把唯一的 `.collapse-button` 直接隐藏，接下来改为折叠态保留可见展开入口，并补充布局回归测试与 `CONTEXT.md` 长期约定。
- 影响文件：`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/components/Layout.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-15 16:22 +0800
- 进度：完成管理端短视频审核页首版落地。新增 `ShortReview.vue`，接入 `/short-review` 路由和媒体库侧栏/命令面板入口；页面固定查询 `short + ready`，支持关键词搜索、恢复上次审核位置、平板优先队列 + 播放器布局、自动播放默认静音、`保留并下一条`、`删除并下一条` 二次确认、当前页末尾自动加载后续页和末尾空状态。新增 `shortReview.helpers` 锁定位置恢复、固定查询、翻页推进和声音会话偏好；补充命令面板与路由源文测试。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/router/index.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js src/components/base/commandPalette.helpers.spec.js src/router/index.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 15:59 +0800
- 进度：开始落地管理端短视频审核页。实现范围：新增 `/short-review` 管理端页面与媒体库侧栏入口；复用现有 `/admin/videos?type=short&status=ready`、`/admin/videos/:id/play-url` 和删除接口；实现平板优先的左侧待审队列 + 右侧播放器布局；支持恢复上次审核视频、`保留并下一条`、`删除并下一条` 二次确认、当前页末尾自动加载下一页、无更多时显示末尾空状态；首版仅关键词搜索，不提供编辑能力。先补纯逻辑测试，再实现页面、路由和导航。
- 影响文件：`admin-web/src/views/ShortReview.vue`、`admin-web/src/views/shortReview.helpers.js`、`admin-web/src/views/shortReview.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- src/views/shortReview.helpers.spec.js src/components/base/commandPalette.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-15 15:38 +0800
- 进度：通过 `$grill-with-docs` 校准管理端短视频审核页需求。已确认待审集合为 `ready` 短视频并按上传时间倒序；页面优先平板布局，采用左侧待审队列 + 右侧大播放器；重新进入优先恢复上次正在审核的视频；`保留并下一条` 只推进审核位置、不写“已审核”状态；`删除并下一条` 走现有视频删除语义且必须二次确认；走到当前已加载列表末尾时自动加载下一页，确认没有更多待审短视频时才显示末尾空状态；播放器自动播放、默认静音、声音状态仅保留当前审核会话；首版只提供关键词搜索；入口放在媒体库分组，命名“短视频审核”，路由 `/short-review`；后端复用现有 `/admin/videos`、播放链接和删除接口；审核页不承载编辑能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：grill 阶段仅做文档沉淀，待实现阶段执行管理端构建和必要测试。

## 2026-06-15 14:49 +0800
- 进度：开始按最新 code review 收口 DV 保留实现的后续问题。范围收口为：1）DV 原始文件保留失败不再让整个转码任务失败，而是降级为 warn 并继续完成普通转码；2）管理端重转码拆分“DV 禁止重转码”和“无可用源文件”的错误码；3）收口 `sameFilePath` / `isSameFilePath` 重复实现。`dv_source` 文件存在性检查维持现状，不额外加前置 `stat`。
- 影响文件：`internal/queue/tasks.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/handlers/admin.go`、必要的定向测试、`plan.md`
- 验证：待执行 `go test ./internal/queue ./internal/handlers ./internal/services -run 'Test(IsSameFilePath|SelectRetranscodeInputPath|BuildTranscode|Transcode|HandleTranscode)' -count=1`、`git diff --check` 与乱码扫描。

## 2026-06-15 14:50 +0800
- 进度：完成 DV 保留实现的 review 修正。DV 原始文件保留失败现已降级为 warn log，转码任务会继续按普通输出成功完成且不写 `source_playback_path`；管理端重转码把“DV 禁止重转码”保留在错误码 `1012`，把“无可用源文件”改为 `1071`；`internal/queue/tasks.go` 已复用 `internal/services/transcode.go` 的 `IsSameFilePath`，去掉本地重复实现。本次只纳入这 4 个 Go 文件和 `plan.md`。
- 影响文件：`internal/queue/tasks.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/handlers/admin.go`、`plan.md`
- 验证：`go test ./internal/queue ./internal/handlers ./internal/services -run 'Test(IsSameFilePath|SelectRetranscodeInputPath|BuildTranscode|Transcode)' -count=1` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 14:04 +0800
- 进度：开始按最新 DV 决策（提交 `b9ea97f` / ADR `0011-dolby-vision-episode-original-file-preservation`）补代码实现。范围收口为：1）后端转码时对 `episode + Dolby Vision source` 保留原始文件并写入 `playback_compat.source_playback_path`；2）视频源接口支持 `profile=dv_source`；3）管理端禁止对持有 DV 原始文件的视频重转码并在删除/孤儿文件扫描时精确处理保留源文件；4）TV 端在 DV 场景切换到 `dv_source` 专用播放路由。TV 端功能改动需同步升级版本号。
- 影响文件：`internal/queue/tasks.go`、`internal/services/playback_compat.go`、`internal/handlers/video_source.go`、`internal/handlers/admin.go`、`internal/repository/orphan_file_scan_repository.go`、相关 Go 测试、`android-tv-app/tv-app/**`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：待执行 Go 定向测试（playback compat / video source / admin retranscode / orphan scan / queue/transcode 相关）、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.* --tests com.chee.videos.core.util.UrlBuilderSourceProfileTest`、必要时 `:tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-15 14:19 +0800
- 进度：完成最新 DV 决策落地。后端转码对 `episode + Dolby Vision source` 现已保留原始文件到 `storageRoot/videos/<uuid>/source-dv.<ext>`，并把完整路径写入 `playback_compat.source_playback_path`；视频源接口支持 `profile=dv_source`；管理端重转码会拒绝持有 DV 原始播放源的视频，删除与孤儿文件扫描会精确保留/清理 `source_playback_path`。TV 端播放策略已改为：有 `source_playback_path` 的 DV 剧集分集走 `dv_source` 专用 Media3 路由，无该字段的存量数据继续阻断；TV 版本提升到 `0.1.99 / 99`。本次只纳入 DV 决策实现相关后端、TV、测试和 `plan.md`。
- 影响文件：`internal/queue/tasks.go`、`internal/services/playback_compat.go`、`internal/handlers/admin.go`、`internal/handlers/video_source.go`、`internal/repository/orphan_file_scan_repository.go`、相关 Go 测试、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关 TV 测试、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers ./internal/repository -run 'Test(BuildPlaybackCompatibilityMetadata|SourcePlaybackPathFromMetadata|MergePlaybackCompatibilityMetadata|ResolveProfiledPlayableSource|SelectRetranscodeInputPath|CollectLocalStoragePaths)' -count=1` 通过；`go test ./internal/queue -run 'Test|^$' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvLongFormDetailPresentationTest --tests com.chee.videos.core.util.UrlBuilderSourceProfileTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.TvRepositoryMappingTest` 通过；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 11:18 +0800
- 进度：开始修复批量编辑 code review 发现的问题。范围收口为三项：1）批量编辑与批量删除互斥，避免并发操作同一选择集；2）批量编辑抽屉打开时为合集/图片图集远程加载补错误兜底，避免未处理 Promise；3）把脆弱的文案断言测试改成关键行为/源码结构断言，并补后端批量标签与批量校验的定向测试。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.spec.js`、`internal/handlers/admin_batch_delete_test.go`、`internal/repository/admin_repository_batch_test.go`、`plan.md`
- 验证：待执行 `go test ./internal/repository ./internal/handlers -run 'Test(AdminBatch|MergeAdmin|NormalizeAdmin)' -count=1`、`cd admin-web && npm run test -- src/views/videoList.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-15 11:24 +0800
- 进度：完成批量编辑 review 修复。`VideoList` 的批量编辑/批量删除现在共享互斥门控，避免同一选择集并发提交；批量编辑抽屉首次打开时改为 `Promise.allSettled` 预加载合集与图片图集，并在失败时给出可恢复提示；前端测试改为锁定 Shift 区间选择和互斥/预加载结构，后端测试补了空 patch、非法图片图集 ID 和标签空输入等边界。本次只纳入 review 修复相关前端页面/测试与 `plan.md`。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.spec.js`、`internal/handlers/admin_batch_delete_test.go`、`internal/repository/admin_repository_batch_test.go`、`plan.md`
- 验证：`go test ./internal/repository ./internal/handlers -run 'Test(AdminBatch|MergeAdmin|NormalizeAdmin)' -count=1` 通过；`cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 10:46 +0800
- 进度：完成管理端视频批量编辑收尾。后端新增 `PUT /api/v1/admin/videos/batch-update`，支持对当前页已选视频统一覆盖标题、图片图集、所属合集和标签；标签支持 `replace` / `append` / `remove`，合集继续限制为短视频。前端 `VideoList` 已接入批量编辑抽屉、当前页 `Shift` 区间选择、批量 API 调用与部分成功提示。`CONTEXT.md` 已补充长期约定。本次只纳入管理端批量编辑相关前后端文件、`CONTEXT.md` 与 `plan.md`，用户既有无关改动保持不纳入。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_batch_delete_test.go`、`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_batch_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository ./internal/handlers -count=1` 通过；`cd admin-web && npm run test -- src/api/admin.spec.js src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check` 通过；乱码扫描无命中。

## 2026-06-15 09:50 +0800
- 进度：开始实现管理端视频管理批量编辑。需求范围锁定为当前页可见行的 Shift 区间选择，以及对当前已选视频批量统一覆盖标题、图片图集、所属合集和标签；合集仍仅适用于短视频，混选或非短视频时前端禁用该字段，后端继续兜底校验。实现计划为：先补 `plan.md` 记录，再新增管理端视频批量更新接口与定向测试，随后接入 VideoList 批量编辑抽屉、Shift 选择和 API 调用，最后更新 `CONTEXT.md` 并执行管理端/Go 定向验证与构建。用户既有 `CONTEXT.md` 修改继续保留并在其上追加，本次不纳入 `admin-web/.env.development`。
- 影响文件：`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_batch_delete_test.go`、`internal/repository/admin_repository.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `go test ./internal/handlers ./internal/repository -count=1` 的定向用例、`cd admin-web && npm run test -- src/api/admin.spec.js src/views/videoList.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 22:16 +0800
- 进度：完成任务监控页视频标题展示。`/admin/tasks` 列表查询通过 `transcoding_jobs -> videos` 左连接返回 `video_title`，避免任务因视频记录缺失被过滤；管理端任务表主列改为“任务”，主行显示视频标题，下面保留任务 ID 与视频 ID。`CONTEXT.md` 记录“任务监控视频标题”语义。本次只纳入任务监控相关后端模型/仓储、前端页面/测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/taskMonitorPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository -run TestBuildAdminListTranscodingTasksSQL -count=1` 通过；`go test ./internal/repository -count=1` 通过；`cd admin-web && npm run test -- src/views/taskMonitorPage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- ...` 通过；乱码扫描无命中。

## 2026-06-14 22:13 +0800
- 进度：开始为任务监控页补充视频标题。经 `$grill-with-docs` 校准，“文件标题”采用视频实体的 `videos.title`，不是从 `original_path` 推导的原始文件名；任务监控表格应优先显示视频标题，同时保留 `video_id` 便于排障。当前后端 `AdminTaskListItem` 和 SQL 只返回任务 ID / video_id 等任务字段，前端也只显示视频 ID。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/taskMonitorPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 Go 仓储定向测试、管理端定向 Vitest、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 21:39 +0800
- 进度：完成任务监控页 loading 卡死修复。`TaskMonitor.vue` 的自动刷新改为不重入：上一轮 `/admin/tasks` 仍在请求时，5 秒轮询直接跳过，不再堆积并发请求；页面首轮仍会正常加载，手动刷新保持可用。`CONTEXT.md` 增加“任务监控轮询不重入”长期约定。部署机当前确认有 2 个 `ffmpeg` 转码任务正在运行，server / worker 均为 running，和本次前端修复无冲突。本次只纳入任务监控页、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/taskMonitorPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/taskMonitorPage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/TaskMonitor.vue admin-web/src/views/taskMonitorPage.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中；`ssh chee@192.168.1.24 'curl -sS -m 8 http://127.0.0.1:8080/healthz'` 返回 200。

## 2026-06-14 21:37 +0800
- 进度：开始排查任务监控页一直 loading。部署机当前确认有 2 个 `ffmpeg` 压缩任务正在运行，server / worker 均为 running；页面代码存在 5 秒自动刷新重入风险：上一轮 `/admin/tasks` 未完成时新一轮 `load()` 会推进 `loadSeq`，旧请求 finally 因序号不匹配不清理 `loading`，请求持续重叠时页面会一直停在 loading。本次计划限制自动轮询不重入，并保留手动刷新可主动发起。
- 影响文件：`admin-web/src/views/TaskMonitor.vue`、任务监控页面源文测试、`CONTEXT.md`、`plan.md`
- 验证：待执行管理端定向 Vitest、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 20:19 +0800
- 进度：完成 TV App 安装包管理页默认筛选修复。页面初始化和切客户端重置筛选时都默认查看全部发布记录，开关文案改为“只看家庭可见 / 查看全部”；上传成功后回到第一页再刷新，避免仍停在其它分页导致新草稿不易发现。`CONTEXT.md` 同步把长期约定改为“TV 管理端默认查看全部记录”。本次只纳入管理端页面、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/TvAppManage.vue`、`admin-web/src/views/tvAppManagePage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/tvAppManagePage.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- admin-web/src/views/TvAppManage.vue admin-web/src/views/tvAppManagePage.spec.js CONTEXT.md plan.md` 通过；乱码扫描无命中。

## 2026-06-14 20:17 +0800
- 进度：按用户补充要求修正方案：不是上传成功后临时切换“家庭可见”筛选，而是安装包管理页默认就不勾选“只看家庭可见”，打开页面、切客户端重置筛选后都应查看全部记录；上传成功后只回到第一页并刷新，不改动当前筛选开关语义。
- 影响文件：`admin-web/src/views/TvAppManage.vue`、`admin-web/src/views/tvAppManagePage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行更新后的管理端定向 Vitest、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 20:15 +0800
- 进度：开始修复 TV App Release APK 上传成功后管理端看不到记录的问题。当前代码默认启用“只看家庭可见”，上传接口建档后的新记录仍是草稿，`uploadAPK()` 成功后直接刷新列表会继续被 `current_published=1` 过滤；本次计划在上传成功后自动切到“查看全部”并回到第一页，让刚上传的草稿记录可见。用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/TvAppManage.vue`、管理端页面源文测试、`CONTEXT.md`、`plan.md`
- 验证：待执行管理端定向 Vitest、`cd admin-web && npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 20:04 +0800
- 进度：完成 TV App 开屏页图片接入。用户提供图片已作为 `drawable-nodpi/tv_splash_image.png` 引入，`Theme.VideoHome.Starting` 通过窗口背景展示开屏页，`TvMainActivity.onCreate` 在 `super.onCreate` 前切回 `Theme.VideoHome`，避免开屏背景影响后续 Compose 页面；TV 端版本提升到 `0.1.98 / 98`。`CONTEXT.md` 同步记录“TV 开屏页图片”只覆盖 Compose 首帧前空白，不新增固定等待页。本次只纳入 TV splash 相关资源/代码/测试、TV 版本号、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/AndroidManifest.xml`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/main/res/values/themes.xml`、`android-tv-app/tv-app/src/main/res/drawable/tv_splash_background.xml`、`android-tv-app/tv-app/src/main/res/drawable-nodpi/tv_splash_image.png`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvSplashScreenSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.tv.TvSplashScreenSpecTest --tests com.chee.videos.feature.tv.TvApkPackagingConfigTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- ...` 通过；乱码扫描无命中。

## 2026-06-14 19:59 +0800
- 进度：开始将用户提供图片设置为 TV App 开屏页图片。代码现状显示 TV App 只有空的 `Theme.VideoHome`，manifest 直接挂到 `TvMainActivity`，没有独立启动页主题或 splash 资源；本次计划用 Activity starting theme 的窗口背景展示图片，进入 `TvMainActivity` 后切回正常主题，不新增等待页、不改变登录/连接流程。图片源为 `/Users/cuiqi/Downloads/image-workbench-1 (1).png`，尺寸 1672x941，作为 `drawable-nodpi` 资源引入。TV 端版本需从 `0.1.97 / 97` 递增到 `0.1.98 / 98`。用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/AndroidManifest.xml`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/main/res/values/themes.xml`、TV splash 资源、TV 测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV splash 源文测试、TV 定向单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码扫描。

## 2026-06-14 18:43 +0800
- 进度：完成 ED2K 链接生成器点击状态调整。未点击链接不再显示“未点击”或任何状态标签；点击链接后在本页会话内显示“已点击”，并用固定状态列降低标签出现时的文本跳动。`CONTEXT.md` 同步改回“点击后仅本页会话标记已点击”。本次只纳入 ED2K 工具页、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxEd2k.vue admin-web/src/views/toolboxPage.spec.js` 通过；乱码扫描无命中。

## 2026-06-14 18:42 +0800
- 进度：开始按用户补充要求调整 ED2K 链接生成器点击状态。上一轮把点击状态完全移除，与“点击后显示已点击状态”的新语义冲突；本次收口为未点击时不显示状态标签，点击后稳定显示“已点击”，并避免标签出现导致链接文本跳动。用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行管理端定向 vitest、`npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 18:24 +0800
- 进度：完成管理端 ED2K 链接生成器点击状态移除。页面不再维护 `ed2kClickedLinks`，链接点击后不会切换“已点击/未点击”标签或灰化样式；`CONTEXT.md` 同步收口为“不追踪或展示链接点击状态”。本次只纳入 ED2K 工具页、页面源文测试、`CONTEXT.md` 与 `plan.md`，用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxEd2k.vue admin-web/src/views/toolboxPage.spec.js` 通过；乱码扫描无命中。

## 2026-06-14 18:23 +0800
- 进度：开始移除管理端 ED2K 链接生成器的点击状态反馈。代码现状显示页面用 `ed2kClickedLinks` 在点击链接后切换“已点击/未点击”标签和灰化样式；本次范围收口为移除点击状态追踪与状态文案，保留 ED2K 解析、清空和跳转行为。用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行管理端定向 vitest、`npm run build`、`git diff --check` 与乱码扫描。

## 2026-06-14 14:33 +0800
- 进度：完成 DV→SDR 方案回退验证，准备提交并部署。本次回退后代码层已无 `dv_sdr_compat`、`video-dv-sdr.mp4`、`trusted_tone_map`、`trusted_compat_output` 等 SDR 可信输出路径；TV 端 source Dolby Vision + output 非 Dolby Vision 恢复阻断。用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue -run 'TestBuildTranscode|TestParsePlayback|TestBuildPlaybackCompatibility|TestChooseTranscodeOutputProfile|TestBuildPlaybackMetadata|TestTranscodePlan' -count=1 -v` 通过；`go test ./... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；代码层 SDR 标记扫描无命中；待执行最终 `git diff --check`、乱码扫描和提交。

## 2026-06-14 14:29 +0800
- 进度：按用户最新要求回退所有 DV→SDR 方案代码。后端移除 `dv_sdr_compat` / `video-dv-sdr.mp4` / `trusted_tone_map=dv_sdr_bt709` 可信 SDR 输出路径，转码任务不再把源播放探测传入转码器；`playback_compat` 不再识别可信 SDR 标记。TV 端恢复 source Dolby Vision + output 非 Dolby Vision 的阻断语义，避免播放已知可能异色的转码输出；TV 版本提升到 `0.1.97 / 97`。`CONTEXT.md` 和部署文档同步记录“DV→SDR 可信输出撤回”。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`
- 验证：待执行 Go 转码/兼容定向测试、TV 播放兼容定向单测、`git diff --check` 与乱码扫描。

## 2026-06-14 14:18 +0800
- 进度：开始排查 DV 源片重转码后仍异色的问题。当前已知转码任务本身成功、输出 metadata 为 `dv_sdr_compat` / `dv_sdr_bt709`，但用户实播仍颜色异常；下一步先确认播放链路是否取到了 `video-dv-sdr.mp4`，再验证现有 `zscale + tonemap` 链路对 Dolby Vision profile 5 是否存在先天色彩解释错误。
- 影响文件：`plan.md`，后续可能涉及 `pkg/ffmpeg/ffmpeg.go`、`internal/services/transcode.go`、`CONTEXT.md`
- 验证：待执行远端 API/数据库核对、源片小样本滤镜验证、定向 Go 测试。

## 2026-06-14 14:12 +0800
- 进度：完成失败 DV 源片的真实重转码验证。重新入队视频 `d967fccb-f20b-4d7d-a82b-03c94f3a26d9` 后，远端 asynq job `8672` 已从 `running` 正常结束为 `success`；视频状态转为 `ready`，输出落在 `/Volumes/large/videos/videos/d967fccb-f20b-4d7d-a82b-03c94f3a26d9/video-dv-sdr.mp4`。本次只追加部署验证账本，用户已有 `admin-web/.env.development` 仍保持未提交且不纳入。
- 影响文件：`plan.md`
- 验证：worker 进度轮询显示 `8672|success|2796|2796|0|100.00`；数据库查询返回 `ready|.../video-dv-sdr.mp4|dv_sdr_compat|dv_sdr_bt709|h264|h264`；远端 `ffprobe` 确认输出为 `h264`、`yuv420p`、`bt709`、`tv` range；远端 `ls -lh` 确认 `video-dv-sdr.mp4` 约 `2.7G` 且 `thumb.jpg` 已生成。

## 2026-06-14 13:22 +0800
- 进度：完成 DV→SDR 转码滤镜修复的本地验证，准备提交并推送部署。本次只纳入后端 Go 转码/探测代码、对应测试、`CONTEXT.md` 长期约定和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go test ./... -count=1` 通过；待执行最终 `git diff --check`、乱码检查、提交与 `git push deploy master`。

## 2026-06-14 13:13 +0800
- 进度：继续修复家用部署机 DV→SDR 转码失败。部署机 `ffmpeg-full 8.1.1` 已具备 `zscale`，但失败源片的 ffprobe 只暴露 DOVI side data 与 `color_range=pc`，缺少 `color_primaries/color_transfer/color_space`，原滤镜链让 `zscale` 猜输入色彩并报 `code 3074: no path between colorspaces`。已将 `dv_sdr_compat` 滤镜改为显式输入 BT.2020 / SMPTE 2084 / BT.2020nc / full range，先转线性 `gbrpf32le` 中间格式再 `tonemap`，最后输出 SDR BT.709；播放兼容探测同步读取并落库 `color_range`。部署机上已用失败源片跑首帧验证，新滤镜链可正常输出 `yuv420p(tv, bt709)`。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/playback_compat.go`、`CONTEXT.md`、`plan.md`
- 验证：远端 `ffprobe` 确认失败源片仅有 `pix_fmt=yuv420p10le`、`color_range=pc` 与 DOVI side data；远端 `ffmpeg ... -vf "zscale=primariesin=bt2020:transferin=smpte2084:matrixin=bt2020nc:rangein=pc:primaries=bt2020:transfer=linear:matrix=bt2020nc:range=pc,format=gbrpf32le,tonemap=tonemap=hable:desat=0,zscale=primaries=bt709:transfer=bt709:matrix=bt709:range=tv,format=yuv420p" -frames:v 1 -f null -` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；待执行 `git diff --check`、乱码检查并提交部署。

## 2026-06-14 12:35 +0800
- 进度：完成家用部署机 `ffmpeg-full` 安装与运行环境切换。远端已安装 `ffmpeg-full 8.1.1`，验证编译参数包含 `--enable-libzimg`，滤镜列表包含 `zscale` 与 `tonemap`，并链接 `libzimg`。已备份远端 `~/Library/LaunchAgents/com.aivideo.{server,worker}.plist` 到 `.bak-20260614123228`，将 server / worker 的 launchd PATH 调整为 `/opt/homebrew/opt/ffmpeg-full/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin` 后重新 bootstrap；server pid `29201`、worker pid `29203` 已运行。安装过程中 Homebrew cleanup 曾移除 Node 并破坏普通 ffmpeg 旧依赖，已恢复：`node v26.3.0` / `npm 11.16.0` 可用，普通 `ffmpeg 8.1.1` 可执行但仍不含 `zscale`，本项目通过 PATH 使用 `ffmpeg-full`。失败的 DV 转码任务在旧环境下已重试耗尽，需要后续重新入队/重转码才会使用新 ffmpeg。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；远端运维文件 `~/Library/LaunchAgents/com.aivideo.{server,worker}.plist` 已手动更新
- 验证：远端 `ffmpeg -hide_banner -filters` 在新 PATH 下返回 `zscale` 与 `tonemap`；`otool -L /opt/homebrew/opt/ffmpeg-full/bin/ffmpeg` 显示 `libzimg`；`launchctl print gui/501/com.aivideo.{server,worker}` 显示新 PATH 且 state=running；`curl http://127.0.0.1:8080/healthz`、`curl http://192.168.1.24:8080/healthz`、`curl http://127.0.0.1:8080/admin/` 均返回成功。

## 2026-06-14 12:17 +0800
- 进度：开始排查家用部署机 DV→SDR 转码失败。已确认 `deploy` remote 指向 `chee@192.168.1.24`，远端 worker 由 launchd 运行，PATH 包含 `/opt/homebrew/bin`，实际使用 `/opt/homebrew/bin/ffmpeg`。部署机当前 ffmpeg 8.1 滤镜列表只有 `tonemap`，没有 `zscale`，二进制未链接 `zimg`；Homebrew core 的普通 `ffmpeg` 公式没有 `zimg` 依赖，单独 `brew install zimg` 不会让现有 ffmpeg 获得 `zscale`。部署机可用的 `ffmpeg-full` 公式包含 `zimg`，但为 keg-only，需要安装后显式让 worker PATH 优先使用 `/opt/homebrew/opt/ffmpeg-full/bin` 或等价方式。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：远端 `launchctl print gui/501/com.aivideo.worker` 确认 worker 环境；远端 `/opt/homebrew/bin/ffmpeg -hide_banner -filters` 未发现 `zscale`；远端 `brew info ffmpeg-full` 显示依赖包含 `zimg`。

## 2026-06-14 11:36 +0800
- 进度：开始落地 TV 参考图视觉换代第三批。范围限定为播放器 UI 与连接/状态功能页：长视频/电视剧播放器共用控件、字幕/音轨选择浮层、返回确认、续播/连播提示卡、TV 连接服务器页、配对页以及共享加载/错误/空态；本批只替换旧蓝青/冷色/白字金底等视觉残留，不改变播放内核、遥控器按键路由、播放历史上报、服务器连接流程或页面导航结构。用户已有 `admin-web/.env.development` 保持未提交且不纳入本轮。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvAutoplayPromptCard.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先补第三批参考图源文测试，再执行播放器/连接/状态定向 TV 单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 11:40 +0800
- 进度：完成第三批播放器与功能页参考图视觉替换。`LongFormVideoPlayer` 的 seek/center feedback、顶栏、底部控制条、进度条、电视剧选集 rail tooltip 和图标按钮改用暖金暗玻璃 token；字幕/音轨夜台玻璃浮层去除旧白色/冷色硬编码，改由 `AppChrome` 驱动；返回确认、连接页和配对页金色主操作同步改为深色前景。新增 `TvPlayerFunctionReferenceStyleSpecTest` 锁住第三批不回退旧冷色与白色前景。TV 端版本提升到 `0.1.96`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvTrackPickerGlassPanelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlayerFunctionReferenceStyleSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlayerFunctionReferenceStyleSpecTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；待执行播放器/连接/状态定向测试、TV 全量单测、assemble、`git diff --check` 与乱码检查。

## 2026-06-14 11:31 +0800
- 进度：完成 TV 参考图视觉换代第二批内容页验证，准备提交。本次只纳入 TV 首页目录、海报墙、长视频详情页、内容页参考图风格测试、TV 版本号、`CONTEXT.md` 长期约定和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvContentReferenceStyleSpecTest.kt`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvContentReferenceStyleSpecTest --tests com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest --tests com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest --tests com.chee.videos.feature.tv.TvLongFormDetailActionSpecTest --tests com.chee.videos.feature.tv.TvLongFormDetailGlassPanelSpecTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码检查无命中。

## 2026-06-14 11:29 +0800
- 进度：开始落地 TV 参考图视觉换代第二批内容页。范围限定为 TV 首页目录、海报墙、长视频详情页，以及对电视剧详情页既有参考图实现做回归核对；本批只替换内容页旧冷色/紫蓝占位、金色按钮内容色和暗玻璃背景，不改变播放器、连接/配对、加载/错误/空态行为，也不改路由、播放内核或遥控器焦点结构。用户已有 `admin-web/.env.development` 保持未提交且不纳入本轮。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、相关 TV 内容页测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先补内容页参考图风格源文测试，再执行定向 TV 单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 11:29 +0800
- 进度：完成第二批内容页参考图视觉替换。TV 首页目录、海报墙和长视频详情页移除旧紫蓝/冷蓝占位色；金色按钮、左侧 rail 选中项和续播 chip 改用 `AppChrome.Canvas` 深色前景；内容页占位、scrim、标题条和详情页返回/次操作统一走暖金暗玻璃 token 或本页 reference brush。新增 `TvContentReferenceStyleSpecTest` 锁住内容页不再回退旧冷色与白字金底。TV 端版本提升到 `0.1.95`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvContentReferenceStyleSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvContentReferenceStyleSpecTest` 通过；待执行内容页定向测试、TV 全量单测、assemble、`git diff --check` 与乱码检查。

## 2026-06-14 11:28 +0800
- 进度：补跑 TV 全量单测时发现 `TvSeriesPlayerViewModelTest` 里仍保留 source DV / output 非 DV 阻断的旧断言；按现有 `DV→SDR 统一兼容路线` 和 `TvPlaybackRoutePolicyTest` 语义，改为期望构建普通 VLC 播放源。业务代码未改，仅修正过期测试断言。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待重新执行 `git diff --check` 与乱码检查。

## 2026-06-14 11:18 +0800
- 进度：完成 TV 参考图视觉换代第一批收尾验证，准备提交。本次只纳入 TV 全局 token、共享 UI 组件、音轨/字幕选择器暖金化、对应测试、TV 版本号、`CONTEXT.md` 长期约定和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交且不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、相关 `core/ui` 测试
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.TvFocusSpecTest --tests com.chee.videos.core.ui.TvTypographySpecTest --tests com.chee.videos.core.ui.TvColorContrastTest --tests com.chee.videos.core.ui.TvStateFeedbackSpecTest --tests com.chee.videos.core.ui.TvIconActionSpecTest --tests com.chee.videos.core.ui.TvShapeAuditTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 与乱码检查待最终执行。

## 2026-06-14 11:08 +0800
- 进度：完成 TV 参考图视觉换代第一批落地。`AppChrome` 切到暖金、暗玻璃、8dp 通用圆角和暖金按钮内容色；`TvFocus` 的 glow 语义切到暖金且保留 spring/触觉行为；`TvTypography` 改为参考图紧凑字号角色；共享图标按钮、状态页和音轨/字幕选择器同步去除旧蓝青风格。TV 端版本提升到 `0.1.94`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.TvFocusSpecTest --tests com.chee.videos.core.ui.TvTypographySpecTest --tests com.chee.videos.core.ui.TvColorContrastTest --tests com.chee.videos.core.ui.TvStateFeedbackSpecTest --tests com.chee.videos.core.ui.TvIconActionSpecTest --tests com.chee.videos.core.ui.TvShapeAuditTest --tests com.chee.videos.core.ui.TvTrackPickerGlassPanelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:46 +0800
- 进度：开始落地 TV 参考图视觉换代第一批代码。范围限定为全局 token 与共享组件：`AppChrome`、`TvFocus`、`TvTypography`、`TvIconActionButton`、`TvStateFeedback` 及对应源文/纯逻辑测试；同步 TV 端版本到下一 patch。用户已有 `admin-web/.env.development` 保持未提交且不纳入本轮。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvTypography.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行共享 UI 定向单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 10:39 +0800
- 进度：确认 TV 参考图导航结构。整体换代后主导航采用参考图式左侧竖向 rail 承载一级入口，不引入顶部横向导航，避免两套导航在 TV 遥控器焦点路径上重复。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `git diff --check` 与乱码检查；本轮为 grill 决策沉淀，不做代码构建。

## 2026-06-14 10:37 +0800
- 进度：确认 TV 视觉换代的测试策略。测试跟随新参考图设计语义重写，不再把旧蓝青焦点色、旧 `tvFocusableGlow` 语义或 10-foot 字号地板作为准入断言；保留对比度和基本可读性护栏测试，并让源文测试改锁暖金、暗玻璃、紧凑字号和参考图焦点语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建；待继续确认导航结构边界后执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:33 +0800
- 进度：确认 TV 参考图视觉换代采用三批迁移：第一批全局 token 与共享组件，第二批内容页面，第三批播放器 UI 与连接/状态功能页。代码探索发现现有测试会硬锁旧蓝青焦点色、`tvFocusableGlow` 语义和 10-foot 字号地板，第一批迁移必须同步把这些测试改成暖金/紧凑字号/对比度护栏语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档与代码探索，无需构建；待继续确认测试策略后执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:22 +0800
- 进度：确认 TV 参考图视觉换代的实现策略：先建立暖金、暗玻璃、边框、焦点态、紧凑字号和 8dp 圆角等 token，再替换共享组件和页面调用点。代码现状显示 `core/ui/AppChrome.kt` 已是主要全局 token 入口，后续优先升级该入口而不是逐页散落硬编码。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档与代码探索，无需构建；待继续确认迁移批次后执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:18 +0800
- 进度：确认连接/配对、加载、错误和空态也纳入 TV 参考图视觉换代，但继续保持清晰功能面板结构，不引入影视海报式沉浸背景。整体换代范围至此覆盖普通页面、详情页、播放器 UI 和功能状态页。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建；待继续收敛实现落地方式后执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:16 +0800
- 进度：确认播放器 UI 也纳入 TV 参考图视觉换代。播放器控制条、进度条、按钮、字幕/音轨/倍速选择和返回/续播/连播浮层统一换成暖金暗玻璃与紧凑字号，但不改变控制条结构、遥控按键行为、播放内核或选择器层级。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建；待继续收敛连接/配对页和状态页边界后执行 `git diff --check` 与乱码检查。

## 2026-06-14 10:06 +0800
- 进度：确认 TV 整体视觉换代保留纯可访问性护栏，放弃旧 10-foot 排版 token 作为视觉限制。后续字号、间距、圆角、焦点样式以参考图紧凑风格为准，但文本对比度和基本可读性继续作为准入边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建；待继续收敛播放器和连接页具体换皮边界后执行 `git diff --check` 与乱码检查。

## 2026-06-14 09:44 +0800
- 进度：继续校准 TV App 整体视觉换代的术语。已确认可见 UI 统一采用参考图风格，旧版蓝青焦点语言改为历史参考；下一步需要继续确认是否保留 10-foot 对比度/排版约束作为纯可访问性护栏，还是一并纳入新的参考图基线。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，无需构建；待继续收敛设计边界后再执行 `git diff --check` 与乱码检查收尾。

## 2026-06-14 09:20 +0800
- 进度：完成 TV 电视剧详情页焦点态收尾并准备提交。本次只纳入电视剧详情页 UI、源文回归测试、TV 版本号、`CONTEXT.md` 术语和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保持未提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt android-tv-app/tv-app/build.gradle.kts` 无命中。

## 2026-06-14 09:18 +0800
- 进度：继续收紧电视剧详情页焦点语言。右侧剧集卡片的未聚焦播放圆钮从偏绿底调整为中性深灰底，图标降低到更接近参考图的灰白态；聚焦时仍保持金色圆钮与深色图标。其余“焦点与选中分离”的规则不变。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`plan.md`
- 验证：待重新执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 09:06 +0800
- 进度：开始收口 TV 电视剧详情页参考图焦点态。按用户确认，右侧剧集卡片和主播放按钮的金色边框/金色背景只在遥控器聚焦时出现；当前分集离焦后不保留选中视觉。TV 端版本需同步递增。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先补焦点态源文回归测试，再执行定向 TV 单测、必要构建、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 09:14 +0800
- 进度：电视剧详情页焦点态改造已完成到代码与文档层。右侧剧集卡片取消 `selected` 视觉高亮，只在当前焦点时显示暖金边框与高亮播放圆钮；主播放按钮未聚焦时回落为暗色 idle 背景，聚焦时才切换到金色。同步收口 TV 版本号到 `0.1.93`，并在 `CONTEXT.md` 增加“焦点与选中分离”术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 09:06 +0800
- 进度：开始修正 TV 电视剧详情页右侧剧集卡片与主播放按钮的焦点态，按参考图收口为“只在真正聚焦时显示金色边框 / 金色背景”，未聚焦时不保留选中态高亮。计划同步补源文测试、更新 TV 版本号、追加 `CONTEXT.md` 术语并执行 TV 构建验证。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行电视剧详情页源文测试、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

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

## 2026-06-14 08:19 +0800
- 进度：开始实现仅针对 DV 视频的 tone-map SDR 兼容转码。方案收口为：普通视频保持现有输出；source playback probe 确认为 Dolby Vision 时，服务端转码改走 H.264 VideoToolbox 硬编 + SDR tone-map 输出，并写入可信兼容标记；TV 端只在该标记存在时允许 source DV / output 非 DV 走普通播放链路。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、相关测试、`CONTEXT.md`、`plan.md`
- 验证：待补红灯测试后执行 `go test ./pkg/ffmpeg ./internal/services ./internal/queue`、TV 播放兼容策略定向测试、必要的 TV 单测/构建、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:57 +0800
- 进度：完成电视剧播放源准备态修复收尾，准备提交。本次只纳入 TV 端播放准备态代码、测试、版本号、`CONTEXT.md` 术语和 `plan.md` 记录；用户已有 `admin-web/.env.development` 保留未提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt` 无命中。

## 2026-06-14 02:56 +0800
- 进度：补齐电视剧播放源准备态的首次进入覆盖。除切换分集外，首次进入电视剧播放器时如果初始分集的播放源仍在构建，也应停留在准备态而不是闪进“当前分集暂无可播放视频”。`TvSeriesPlayerViewModelTest` 新增初始加载延迟场景，确保准备态在首入与切集两条路径都成立。
- 影响文件：`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`plan.md`
- 验证：待重新执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest --tests com.chee.videos.feature.tv.TvSeriesPlayerPreparingStateSpecTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。

## 2026-06-14 02:50 +0800
- 进度：完成电视剧播放源准备态代码落地。`TvSeriesPlayerViewModel` 在可播放分集异步构建播放源时置 `playbackPreparing=true`，但在源 URL 就绪前不暴露可播放 target，避免历史上报误归到新分集；兼容阻断和未绑定分集保持非准备态。`TvSeriesPlayerScreen` 对准备态显示“正在准备当前分集”，不再落到“暂不能播放”错误页。TV 端版本同步提升到 `0.1.90`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先补红灯测试后，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 已通过；待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:45 +0800
- 进度：开始修复电视剧播放页在可播放分集准备播放源期间短暂闪现“当前不可播放 / 当前分集暂无可播放视频”的问题。已确认根因是分集播放源 URL 异步构建期间 UI 状态被临时清空，屏幕误走不可播放错误态；本次收口为新增明确的播放源准备态，并在准备期间显示加载态。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerPreparingStateSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行定向红灯测试 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest`，实现后执行 TV 端单测、assemble、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:13 +0800
- 进度：完成两个 TV 长视频任务的现有代码收口标记。`tasks/2026-05-25-tv-long-form-focus-guarding/DONE.md` 记录原实现提交、恢复提交、当前主线基线、自动化验证与未重跑 R1~R10 人工手测脚本的边界；`tasks/2026-05-25-tv-long-form-track-preference-recovery/DONE.md` 同步记录原实现提交、恢复提交、当前主线基线、自动化验证与未重跑 A1~A7 人工手测脚本的边界。
- 影响文件：`tasks/2026-05-25-tv-long-form-focus-guarding/DONE.md`、`tasks/2026-05-25-tv-long-form-track-preference-recovery/DONE.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过；静态核对 `TvResumePromptCard(` 只出现在定义和两个 `resumePromptSlot` 调用内，`addSlave(IMedia.Slave.Type.Subtitle` 同时覆盖长视频与电视剧播放入口，`LongFormVideoPlayer.kt` 的 `onSelectAudioTrack` 同时存在 `false` 自动回灌与 `true` 用户选择路径；待执行最终 `git diff --check` 与乱码检查后提交。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 02:10 +0800
- 进度：开始按现有代码收口 `tasks/2026-05-25-tv-long-form-focus-guarding` 与 `tasks/2026-05-25-tv-long-form-track-preference-recovery` 两个 TV 长视频任务。当前 TV 端版本已推进到 `0.1.89`，本次不回填旧 review 中的 `0.1.69/0.1.70` 历史版本断言，只按现有代码执行自动化准入、补充 `DONE.md` 完成标记并提交任务收口文档。
- 影响文件：`tasks/2026-05-25-tv-long-form-focus-guarding/DONE.md`、`tasks/2026-05-25-tv-long-form-track-preference-recovery/DONE.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest`、任务 review 静态检查、`git diff --check` 与乱码检查。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-14 01:45 +0800
- 进度：继续收敛 TV 端 DV 专用链路的失败重试语义。确认 Media3 专用链路重试不应丢失当前播放快照，新的起播位置优先采用当前播放器已采集的进度，再回退到历史记录；这样可以在失败后保留现场进度，切集或显式从头播放才重置。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlaybackHistoryPolicyTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.DolbyVisionDisplayCapabilityTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、`git diff --check`、乱码检查。

## 2026-06-14 01:26 +0800
- 进度：TV 端 DV 播放诊断最小闭环完成并验证通过。诊断入口继续限定在 DV 相关阻断页和专用 Media3 播放失败页，错误卡片改为单入口替换式诊断卡片；同时把 TV 端版本号抬到 `0.1.88` / `88`，确保功能变更和版本管理同步。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon --stop` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvDolbyVisionDiagnosticsTest --tests com.chee.videos.feature.tv.DolbyVisionDisplayCapabilityTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnostics.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvDolbyVisionDiagnosticsTest.kt` 无命中。保留用户已有 `admin-web/.env.development` 不纳入本次提交。

## 2026-06-14 00:38 +0800
- 进度：确认 `DV 诊断返回依赖系统返回`。诊断卡片不单独提供“返回原错误页”按钮，返回前一层只依赖系统返回键或遥控返回，避免额外焦点目标。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛最终字段边界后执行文档乱码检查和 `git diff --check`。

## 2026-06-14 00:27 +0800
- 进度：确认 `DV 诊断正式构建保留`。正式发布构建也保留诊断入口，但只在 DV 相关阻断页或专用播放失败页可见，不依赖 debug 开关。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛入口样式后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 诊断单入口单视图`。诊断不做“摘要 / 详情”切换或可展开层级，用户只通过一个显式入口进入同一张简短诊断卡片，减少交互复杂度。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛是否需要任何额外 debug-only 条件后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 诊断核心卡片`。首轮只呈现一张简短卡片，核心信息收敛到“当前路径 / 当前原因 / 是否有播放源或专用链路可用”，失败页最多补一条简短错误摘要，不做折叠、分组、跳转或多卡片布局。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛最终实现入口与交互后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：收紧 DV 诊断复杂度。已确认首轮只保留最少必要字段和最短操作链路，不做分层钻取、树状展开或多页面诊断流程；目标是一眼看出阻断/专用链路/失败原因。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛最小字段集合后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 播放诊断展示形态`。首轮不新开独立页面，复用 TV 现有错误态主面板风格，以二级动作进入同一错误上下文内的诊断视图或弹层，便于保留焦点和返回行为。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛诊断面板布局和字段后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：确认 `DV 播放诊断入口` 边界。首轮只在播放阻断页和专用 Media3/ExoPlayer 播放失败页通过显式入口查看；正常播放页不常驻展示；诊断信息不泄漏 access token 或完整播放 URL，也不提供强行播放或切换链路能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛诊断字段与实现范围后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 19:00 +0800
- 进度：继续 DV 视频开发前的 grill-with-docs 校准。已确认下一阶段优先补 `DV 播放诊断信息`，先提升真机排障能力，不先追齐字幕、音轨、倍速等 LibVLC 体验能力。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待继续收敛诊断信息入口、展示范围与敏感信息边界后执行文档乱码检查和 `git diff --check`。

## 2026-06-13 18:38 +0800
- 进度：完成图库资产参考图格式过滤收尾检查，准备提交。确认只纳入本轮后端过滤、工作台请求参数、API 单测和计划记录；用户已有 `admin-web/.env.development` 仍保持未纳入。
- 影响文件：`internal/models/admin.go`、`internal/handlers/admin_image.go`、`internal/repository/image_repository.go`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：`rg -n $'\uFFFD' CONTEXT.md plan.md internal/models/admin.go internal/handlers/admin_image.go internal/repository/image_repository.go admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/api/admin.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md internal/models/admin.go internal/handlers/admin_image.go internal/repository/image_repository.go admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/api/admin.spec.js` 通过。

## 2026-06-13 18:35 +0800
- 进度：完成图库资产参考图格式过滤代码落地。后端 `/admin/images` 支持逗号分隔的 `stored_mime` 查询过滤，列表分页和总数直接按处理图 MIME 收口；图像生成工作台图库入口固定请求 `image/png,image/jpeg,image/webp`，与 `CONTEXT.md` 中 `admin 图库资产参考图可选范围` 保持一致。
- 影响文件：`internal/models/admin.go`、`internal/handlers/admin_image.go`、`internal/repository/image_repository.go`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：`go test ./internal/...` 通过；`cd admin-web && npm run test -- admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有既有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:27 +0800
- 进度：开始把图库资产参考图 `stored_mime` 可选范围落地到代码。方案采用后端 `/admin/images` 查询参数过滤，而不是前端分页后本地剔除；前端工作台图库请求传 `image/png,image/jpeg,image/webp`，后端列表总数和分页直接反映可选资产集合。
- 影响文件：`internal/models/admin.go`、`internal/handlers/admin_image.go`、`internal/repository/image_repository.go`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：待执行 `go test ./internal/...`、`cd admin-web && npm run test -- admin.spec.js`、`cd admin-web && npm run build`、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:22 +0800
- 进度：继续收敛图库资产参考图可选范围。已确认图库选图入口除 `status=ready`、`active=true` 外，还必须限制 `stored_mime` 为 `image/png`、`image/jpeg` 或 `image/webp`，避免 GIF/BMP/TIFF 等图片管理可存在但生图参考图不支持的资产进入工作台。`CONTEXT.md` 已收紧 `admin 图库资产参考图可选范围`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' CONTEXT.md plan.md` 无命中；`git diff --check -- CONTEXT.md plan.md` 通过。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:11 +0800
- 进度：完成历史恢复参考图大小校验补强。新增 data URL 字节估算，历史快照恢复出的本地冻结图会带回 size；继续追加上传/粘贴/图库参考图时，校验会同时统计已有恢复图和新图，避免 `File` 为空导致总大小上限失效。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:10 +0800
- 进度：继续补强历史恢复参考图的大小校验一致性。审计发现历史恢复出来的冻结参考图没有 `File` 对象，后续继续追加参考图时总大小校验可能只统计新图；下一步补 data URL 字节估算，恢复历史输入时保留 size，并让追加参考图的校验覆盖已有冻结图。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- imageWorkbench.helpers.spec.js`、`cd admin-web && npm run build`、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:08 +0800
- 进度：完成图库资产来源跳转闭环。工作台当前参考图和历史输入快照里的 `library_asset` 来源现在提供“查看来源”，跳转 `/images?image_id=<source_image_id>`；图片管理页消费 `image_id` query 后尝试打开详情，并在成功或失败后清理 query。来源读取失败时工作台仍保留冻结标题与来源图片 ID 摘要。`CONTEXT.md` 已新增 `admin 图库资产来源跳转 query 契约`。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/ImageManage.vue`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:06 +0800
- 进度：继续补齐图库资产来源跳转保留失效态。审计发现历史详情已展示图库来源标题和 ID，但缺少可用的来源跳转入口；图片管理页也未识别 `image_id` query 自动打开详情。下一步补工作台“查看来源”入口和图片管理页按 `image_id` 打开详情，来源缺失时仍保留只读摘要与失效提示。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/ImageManage.vue`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run build`、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:04 +0800
- 进度：继续收敛图库资产原尺寸处理图超限时的处理。已确认图库资产加入工作台后仍沿用普通参考图大小限制，单张最多 10 MiB、总量最多 40 MiB；原尺寸处理图超限时拒绝加入并展示原因，不自动压缩、缩放或改格式绕过限制。`CONTEXT.md` 已新增 `admin 图库资产冻结沿用参考图大小限制`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与当前 admin-web 实现改动不纳入本次文档确认。

## 2026-06-13 18:04 +0800
- 进度：完成历史输入快照展示与继续改稿恢复闭环。选中本地历史时会读取 `referenceSnapshots` 对应的本地 IndexedDB 冻结图片，展示参考图来源类型、图库来源标题与来源图片 ID；“继续改稿”恢复提示词、参数和本地冻结参考图，不重新读取图库资产。若历史参考图本地文件缺失，会显示缺失位并阻止直接恢复，避免把失效输入送往生成接口。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js` 通过。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 18:00 +0800
- 进度：继续收敛图库资产参考图冻结时读取的图片规格。已确认加入工作台时应读取原尺寸处理图，调用图片查看接口不带 `w` / `h` / `fit` / `q` 缩放参数，避免把低分辨率 UI 预览变体冻结为生图参考输入。`CONTEXT.md` 已新增 `admin 图库资产原尺寸冻结`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与当前 admin-web 实现改动不纳入本次文档确认。

## 2026-06-13 17:59 +0800
- 进度：继续实现工作台图库资产参考图闭环审计发现的缺口。当前已完成“从图库选择并冻结为参考图”，但历史详情还没有展示输入快照，也没有“继续改稿”把历史中的已冻结图库资产参考图恢复回工作台输入区。下一步补本地历史参考图快照恢复 helper、历史输入快照 UI 和继续改稿入口，确保 `library_asset` 历史恢复不重新读取图库资产。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- imageWorkbench.helpers.spec.js`、`cd admin-web && npm run build`、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 17:57 +0800
- 进度：完成本轮收尾检查，准备提交工作台图库资产参考图功能。确认生成 payload 不泄漏 `source_image_id` 等本地来源字段，图库抽屉 object URL 在关闭和请求失效时释放，重复图库资产按来源图片 ID 跳过；用户已有 `admin-web/.env.development` 保留未提交，不纳入本次提交。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js admin-web/src/api/admin.spec.js` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/ToolboxImageWorkbench.vue admin-web/src/views/imageWorkbench.helpers.js admin-web/src/views/imageWorkbench.helpers.spec.js admin-web/src/api/admin.spec.js` 通过。

## 2026-06-13 17:55 +0800
- 进度：完成工作台内置图库资产参考图首轮实现。已新增本地参考图快照 helper，`library_asset` 来源元数据进入 IndexedDB 任务快照但不进入生成 payload；工作台输入区新增“从图库选择”抽屉，按 `status=ready`、`active=1`、搜索词和图片合集筛选图库资产，支持网格多选和剩余槽位限制；加入时通过 `/admin/images/:id/view` 拉取 blob，转为本地 data URL 后立即冻结到参考图槽位，失败项按单张提示且不回滚成功项。`CONTEXT.md` 已新增 `admin 图库资产冻结读取契约`。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- imageWorkbench.helpers.spec.js admin.spec.js` 通过；`cd admin-web && npm run build` 通过，仅有 Vite chunk size warning。待执行乱码检查与 `git diff --check`。

## 2026-06-13 17:45 +0800
- 进度：开始落地工作台内置图库资产参考图功能。实现范围收口为：新增本地参考图快照 helper，确保 `library_asset` 来源元数据只进入本地历史；工作台内置图库抽屉复用图片管理搜索、图片合集筛选和网格多选；加入参考图时立即拉取图片 blob 并冻结成本地 IndexedDB 输入快照。生成接口继续只接收既有 `reference_images` data URL，不接收图库资产 ID。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm run test -- imageWorkbench.helpers.spec.js`、`cd admin-web && npm run build`、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入本轮提交。

## 2026-06-13 17:45 +0800
- 进度：继续收敛图库资产参考图格式处理。已确认图库资产作为普通参考图时保留当前处理图的原 MIME 与字节形态，PNG/JPEG/WebP 都可直接加入；只有当它被选为局部蒙版目标时，才沿用现有蒙版目标 PNG 提交契约。`CONTEXT.md` 已新增 `admin 图库资产参考图格式保留`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:44 +0800
- 进度：补齐图库资产参考图的可选范围、标题快照与重复选择规则。已确认只允许 `status=ready` 且 `active=true` 的图片资产作为新参考图；来源标题按加入工作台时冻结，不随来源资产重命名改写；同一来源图片资产 ID 已在当前草稿里时再次选择应跳过并提示。`CONTEXT.md` 已新增 `admin 图库资产参考图可选范围`、`admin 图库资产来源标题冻结` 与 `admin 图库资产参考图重复选择跳过`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:42 +0800
- 进度：继续收敛多选图库资产加入参考图的失败处理。已确认批量选择后按单张尽力冻结，成功项立即进入参考图槽位，失败项逐项展示原因，不回滚已成功加入的参考图。`CONTEXT.md` 已新增 `admin 图库资产批量加入部分成功`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:39 +0800
- 进度：继续收敛图库资产参考图的选择入口。已确认图库选图应内置在图像生成工作台里，复用图片管理的搜索、合集筛选和网格多选能力；现有“打开媒体库”只作为浏览跳转，不承担跨标签页选图回填。`CONTEXT.md` 已新增 `admin 工作台内置图库选图入口`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:37 +0800
- 进度：继续收敛图库资产来源失效态和继续改稿行为。已确认来源跳转失效时历史详情仍保留来源图片 ID、冻结标题等只读摘要，不重新打开来源图片本体；同时补记此前确认的“继续改稿应恢复历史里已冻结的图库资产参考图快照，不重新从图库读取最新内容”。`CONTEXT.md` 已新增 `admin 图库资产来源失效只读摘要` 与 `admin 继续改稿恢复图库参考图快照`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:19 +0800
- 进度：DV 专用 Media3/ExoPlayer 播放链路本轮实现完成。单个长视频和电视剧分集现在统一通过播放链路选择器在 LibVLC、Media3 DV 专用链路和阻断之间决策；output 仍为 DV 且显示能力支持、播放 URL 有效时进入专用 Media3 分支，source DV 但 output 非 DV、显示能力未知/不支持、metadata 不完整仍阻断。剧集分集按当前分集独立构建 URL 与重判，Media3 失败重试会重新评估并重建 ExoPlayer，不回退 LibVLC；观看历史继续沿用现有接口。已更新 TV 端版本到 `0.1.87`，并调整 LibVLC 守卫测试为允许 DV 专用 Media3 例外。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:compileDebugKotlin` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。待执行乱码检查与 `git diff --check` 后提交；`admin-web/.env.development` 和生图参考图相关记录不纳入本次 DV 提交。

## 2026-06-13 17:03 +0800
- 进度：继续收敛图库资产参考图进入工作台的冻结时机。已确认图库资产加入工作台时就应拉取并复制成当前创作的本地输入快照，生成提交继续沿用现有参考图数据提交语义，不扩展为传图库资产 ID 让后端生成时再解析。`CONTEXT.md` 已新增 `admin 图库资产参考图加入即冻结`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 17:24 +0800
- 进度：继续收敛图库资产来源跳转的保留策略。已确认来源跳转在资产重命名、删除或停用时不应静默消失，而应保留为明确失效态入口或提示，避免历史详情把来源关系直接抹掉。`CONTEXT.md` 已新增 `admin 图库资产来源跳转保留失效态`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 15:58 +0800
- 进度：开始实现 DV 专用 Media3/ExoPlayer 播放链路。已先补纯逻辑 route selector 红灯测试，随后实现 `resolveTvPlaybackRoute` 并保持旧 `resolveTvPlaybackCompatibilityDecision` 行为兼容；修正剧集“可选集”候选语义，使 output 仍为 DV 的分集不会在选集层被排除，探测失败和 DV 转码输出仍不作为候选。TV 端版本号递增至 `0.1.87`，并为 TV 工程加入 Media3 依赖，准备接入最小播放组件。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackRoutePolicyTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvPlaybackRoutePolicyTest` 通过。待继续接入 Media3 UI、单片/剧集播放页并跑全量验证。保留用户已有 `admin-web/.env.development` 与生图相关文档记录不纳入本任务提交。

## 2026-06-13 15:57 +0800
- 进度：按用户要求为当前任务创建独立分支 `admin-image-library-reference`。继续收敛图库资产参考图的生命周期：已确认图库资产进入生图工作台后要复制成当前创作自己的本地输入快照，并保留来源图片资产 ID、当前标题等轻量来源信息；后续来源资产重命名或删除不应破坏本地历史查看和继续改稿，只影响来源跳转与摘要状态。`CONTEXT.md` 已新增 `admin 图库资产参考图输入快照`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留既有未提交改动 `admin-web/.env.development`、TV 端策略文件和 TV 测试文件不纳入本轮处理。

## 2026-06-13 16:31 +0800
- 进度：继续收敛图库资产参考图的数量边界。已确认允许一次多选多张图库资产做参考图，但不新增独立额度；浏览器上传/粘贴、前序结果与图库资产三类来源共享现有 4 张参考图总上限。`CONTEXT.md` 已新增 `admin 参考图槽位总上限`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 与 TV 端未完成改动不纳入本轮处理。

## 2026-06-13 15:52 +0800
- 进度：grill-with-docs 开始收敛“生图允许使用服务器里已有的图库做参考图”。已确认这里的“服务器图库”特指后台图片管理里的正式图片资产，而不是任意服务器文件路径、孤儿文件或浏览器本地历史结果；`CONTEXT.md` 已将参考图来源从旧的 `browser_input` / `previous_result` 两类修正为 `browser_input` / `previous_result` / `library_asset` 三类，并新增 `admin 图库资产参考图`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。

## 2026-06-13 15:49 +0800
- 进度：继续收敛 DV 专用链路实现顺序。已确认实现前应先补纯逻辑“播放链路选择器”测试，把非 DV/老数据走 LibVLC、source DV + output 非 DV 阻断、output DV + 三条件成立走 Media3/ExoPlayer、显示不支持/未知阻断、metadata 不完整/探测失败/未来版本阻断等组合锁住；单个长视频和电视剧分集复用同一选择语义，再接 Compose/Media3 UI。`CONTEXT.md` 已追加 `DV 播放链路选择器`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:47 +0800
- 进度：继续收敛 DV 专用链路是否需要用户开关。已确认首轮不提供启用/关闭杜比视界实验播放开关；播放入口由系统按安全门控自动选择 LibVLC、Media3/ExoPlayer 专用链路或阻断，用户不能通过开关绕过安全策略。`CONTEXT.md` 已追加 `DV 专用链路无用户开关`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:46 +0800
- 进度：继续收敛 DV 专用链路的字幕/音轨体验边界。已确认外挂字幕、多音轨选择、倍速、画面比例等体验能力缺失不作为首轮专用链路阻断条件，也不在播放前额外弹确认；Media3 能自动处理的轨道按系统默认行为使用，未实现控制入口应隐藏或禁用。`CONTEXT.md` 已追加 `DV 专用链路字幕音轨首轮非阻断`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:45 +0800
- 进度：继续收敛 DV 专用链路的观看历史语义。已确认 Media3/ExoPlayer 专用分支只替换客户端播放器实现，不改变长视频或分集的历史接口、完成判定口径和 UI 历史展示；不新增 DV 专用历史字段，客户端内部可用播放器进度快照适配 LibVLC 与 Media3。`CONTEXT.md` 已追加 `DV 专用链路不改变观看历史语义`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:43 +0800
- 进度：继续收敛 DV 专用链路重试语义。已确认播放前阻断页和 Media3/ExoPlayer 播放失败页的“重试”都只重新评估 `playback_compat`、本地显示能力和专用链路可用性；Media3 失败后可释放并重建专用播放器再尝试同一链路，但不回退 LibVLC、不改 metadata、不触发重新转码或 ffprobe。`CONTEXT.md` 已更新 `DV 重试属于幂等重判`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:40 +0800
- 进度：继续收敛 DV 专用链路阻断文案。已确认首轮面向用户只做有限分层：显示链路明确不支持、显示链路能力未知、source 为 DV 但 output 非 DV、专用 Media3/ExoPlayer 链路不可用或播放 URL 无效；不直接暴露 Android API 原因码。`CONTEXT.md` 已追加 `DV 专用链路阻断文案分层`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:39 +0800
- 进度：继续收敛 DV 专用链路进入条件。已确认首轮进入 Media3/ExoPlayer 专用分支必须同时满足：当前默认显示声明支持 DV、`playback_compat` 表明当前实际播放 output 仍是 DV、专用分支可初始化且播放 URL 有效；缺任一项都不尝试播放，source 为 DV 但 output 非 DV 仍走转码输出风险阻断。`CONTEXT.md` 已追加 `DV 专用链路首轮三条件门控`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:37 +0800
- 进度：继续收敛 DV 专用 Media3/ExoPlayer 分支的首轮功能范围。已确认首轮只做最小播放闭环：播放/暂停、返回、失败页、重试、基础进度上报、剧集切集时重新判定；不追齐 LibVLC 的字幕选择、音轨选择、倍速、画面比例或复杂 overlay。`CONTEXT.md` 已追加 `DV 专用链路首轮最小播放闭环`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:33 +0800
- 进度：继续收敛 DV 专用链路首轮覆盖范围。用户要求专用 Media3/ExoPlayer 分支首轮不仅覆盖单个长视频，也覆盖电视剧分集；已同步约束为按当前分集独立读取播放源、`playback_compat` 和能力结果，不跨集沿用专用链路资格。`CONTEXT.md` 已追加 `DV 专用链路覆盖长视频与分集`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 15:30 +0800
- 进度：继续 grill DV 视频下一阶段开发。已确认下一阶段从显示能力检测推进到“专用系统播放链路的最小可验证接入”，不把现有 LibVLC 直接升级为 DV 放行链路；同时确认“专用系统播放链路”收口为 TV App 内部 Media3/ExoPlayer 专用分支，不使用外部播放器 Intent。`CONTEXT.md` 已收紧对应术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md` 与 `git diff --check -- CONTEXT.md plan.md`。保留用户已有 `admin-web/.env.development` 不纳入提交。

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

## 2026-06-13 14:22 +0800
- 进度：继续收敛 DV 本地能力模块的首轮测试范围。已确认单测至少覆盖：包含 `DOLBY_VISION` 为支持、空 HDR 类型列表为不支持、只有 `HDR10/HLG` 为不支持、无 Display 为未知、无 HdrCapabilities 为未知、API 异常为未知、重复或乱序 HDR 类型会归一化；这些测试使用纯 Kotlin fake，不需要 instrumentation。`CONTEXT.md` 已追加 `DV 本地能力模块纯单测覆盖`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:20 +0800
- 进度：继续收敛 DV 本地能力模块的实现归属。已确认首轮放在 `android-tv-app` 的 TV feature 包内，与现有 `PlaybackCompatibilityPolicy` 同域维护；不放入 `core/ui`、`core/player` 或跨端共享层，除非后续出现明确复用需求。`CONTEXT.md` 已追加 `DV 本地能力模块归属 TV feature`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:18 +0800
- 进度：继续收敛 DV 本地能力模块的首轮交付边界。已确认首轮只落代码和单测，不在播放器 UI 暴露能力结果，不写入服务端，也不做本地持久化日志；能力结果的用户提示、日志展示或播放决策接入，等后续与播放链路能力合并设计时再做。`CONTEXT.md` 已追加 `DV 本地能力模块首轮不外显`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:16 +0800
- 进度：继续收敛 DV 本地能力模块的 HDR 类型摘要格式。已确认 `supportedHdrTypes` 诊断摘要只保留归一化后的 Android HDR 类型名列表，例如 `DOLBY_VISION`、`HDR10`、`HLG`，不保留 int 原值；其它 HDR 类型只作为诊断摘要出现，不参与首轮 DV 状态判定。`CONTEXT.md` 已同步收紧 `DV 本地能力结果不退化为 Boolean`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:14 +0800
- 进度：继续收敛 DV 本地能力模块的 HDR 类型列表边界。已确认成功拿到 `supportedHdrTypes` 但列表为空时，属于明确能力结果，应判为 `不支持/dolby_vision_missing`，而不是 `未知`；同时已按推荐将 `supportedHdrTypes` 进入结果对象前去重和顺序归一化，保留归一化摘要，不暴露原始顺序。`CONTEXT.md` 已追加 `DV HDR 类型列表归一化`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 14:00 +0800
- 进度：继续收敛 DV 本地能力模块的默认 Display 读取方式。已确认生产实现优先通过 `Context.getSystemService(DisplayManager::class.java)` 读取 `Display.DEFAULT_DISPLAY`，再取该 Display 的 `HdrCapabilities.supportedHdrTypes`；拿不到默认 Display 时返回未知原因 `no_display`，不依赖 Activity 或 Compose window。`CONTEXT.md` 已追加 `DV 默认显示读取契约`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 13:56 +0800
- 进度：继续收敛 DV 本地能力模块的附加 HDR 字段边界。已确认 `HdrCapabilities` 里的亮度字段、色域或其它 HDR 元数据首轮只作为诊断摘要保留，不参与“支持 / 不支持 / 未知”的状态判定；首轮判定只看是否声明 `HDR_TYPE_DOLBY_VISION`。`CONTEXT.md` 已追加 `DV HDR 附加字段仅诊断`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 13:51 +0800
- 进度：继续收敛 DV 本地能力模块的原因码稳定性。已确认首轮原因码固定为闭集 enum，新增原因必须改 enum 和测试，不允许自由字符串临时扩展；`CONTEXT.md` 已把该约束并入 `DV 本地能力结果不退化为 Boolean`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 13:47 +0800
- 进度：继续收敛 DV 本地能力模块的结果结构。已确认结果对象至少包含状态（支持 / 不支持 / 未知）、原因码和原始 `supportedHdrTypes` 摘要，不能退化为 Boolean；首轮原因码需能区分 `dolby_vision_present`、`dolby_vision_missing`、`no_display`、`no_hdr_capabilities`、`api_error` 等诊断场景。`CONTEXT.md` 已追加 `DV 本地能力结果不退化为 Boolean`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 13:40 +0800
- 进度：继续收敛 DV 本地能力模块的实现边界。已确认 Android framework 读取应封装在可注入的 `DisplayHdrCapabilityReader` 一类边界内；生产实现读取当前默认 Display 的 `HdrCapabilities.supportedHdrTypes`，单测只覆盖纯 Kotlin 三态映射逻辑，不依赖真实 Android Display。`CONTEXT.md` 已追加 `DV 显示能力读取可注入`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:58 +0800
- 进度：继续收敛 DV 本地能力模块的显示能力三态映射。已确认当前默认显示的 HDR 类型明确包含 `HDR_TYPE_DOLBY_VISION` 才输出“支持”；明确拿到列表但不包含 Dolby Vision 才输出“不支持”；拿不到 Display、HdrCapabilities、API 异常或结果不可解释时输出“未知”，探测失败不能折叠成“不支持”。`CONTEXT.md` 已追加 `DV 显示能力三态映射`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:53 +0800
- 进度：继续收敛 DV 本地能力模块的能力精度。已确认首轮只输出显示侧是否声明支持 Dolby Vision，不尝试推断具体 Dolby Vision profile 级支持；Android 显示能力信号只作为粗粒度前置条件，profile 级判断后续由播放链路能力和媒体 metadata 再补。`CONTEXT.md` 已追加 `DV 显示能力不推断 profile`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:51 +0800
- 进度：继续收敛 DV 本地能力模块的评估时机。已确认首轮只在播放准备时评估一次，并由用户触发重试时显式重新评估；不做后台轮询、不常驻监听显示状态变化，也不主动驱动播放状态切换。`CONTEXT.md` 已追加 `DV 本地能力模块播放准备时评估`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:44 +0800
- 进度：继续收敛 DV 本地能力模块的输入边界。已确认首轮只读取当前默认显示侧能力，不把多显示器或外接显示切换纳入第一阶段判断；它输出的是当前播放落点是否具备 DV 安全播放前提，而不是全局设备结论。`CONTEXT.md` 已追加 `DV 本地能力模块仅看当前默认显示`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行本轮 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:24 +0800
- 进度：TV Dolby Vision 播放兼容策略实现完成并进入提交准备。已确认本轮只落地现有 `playback_compat` 决策、老数据普通放行、信息不完整阻断提示和重试入口；未实现设备/HDR 能力探测、专用系统播放器/Media3 分支或播放页补探测。提交范围将精确排除用户已有 `admin-web/.env.development`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvStateFeedbackUsageTest.kt`、`docs/adr/0010-dolby-vision-tv-playback-compatibility.md`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-06-13 11:23 +0800
- 进度：TV Dolby Vision 播放兼容策略已落地到 TV 端。`PlaybackCompatibilityPolicy` 现在将 `playback_compat.version < 1` 作为老数据普通放行、`version > 1` 与探测失败/字段不足作为“播放兼容信息不完整”阻断；长视频与剧集播放阻断态改用 `TvErrorState`，提供“重试”以重新拉取详情并重判；TV 端版本号升至 `0.1.85`。`CONTEXT.md` 已追加 `TV playback_compat v0 历史兼容`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvStateFeedbackUsageTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest --tests com.chee.videos.feature.tv.TvStateFeedbackUsageTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` 通过；待补跑 TV 端相关组合/全量单测、乱码检查与 `git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:18 +0800
- 进度：开始把已收敛的 TV Dolby Vision 播放兼容策略转入实现。范围限定为 TV 端现有 `playback_compat` 决策 helper、单测、TV 版本号与文档记录；本轮不实现设备/HDR 能力探测、不接系统播放器/Media3 专用链路、不在播放页触发补探测。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补定向单测并运行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest`，再视改动运行 TV 端相关验证。保留用户已有 `admin-web/.env.development` 不纳入提交。

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

## 2026-06-13 11:05 +0800
- 进度：grill-with-docs 继续收窄 `老数据普通播放放行` 的适用范围。已确认老数据放行只适用于 `playback_compat` 整体缺失或协议版本过旧的存量视频；如果 metadata 已存在但标记探测失败、source/output DV 判定缺失或关键字段不足，不再适用老数据放行，而应按“播放兼容信息不完整”处理。`CONTEXT.md` 已同步收紧 `DV metadata 不完整不进专用链路` 与 `老数据普通播放放行`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 11:03 +0800
- 进度：grill-with-docs 修正 `playback_compat` metadata 不完整时的处理边界。用户明确要求“老数据直接放行”，已收敛为：`playback_compat` 缺失或过旧的存量视频可以继续按既有 `LibVLC` 普通播放链路放行，但这不是 Dolby Vision 安全播放，也不能进入专用系统播放链路；一旦 metadata 已明确标记 DV 风险源，仍按 DV 阻断或专用链路规则处理。`CONTEXT.md` 已将 `DV metadata 不完整不推断放行` 修正为 `DV metadata 不完整不进专用链路`，并新增 `老数据普通播放放行`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:54 +0800
- 进度：grill-with-docs 继续收敛 `playback_compat` metadata 的可信度边界。已确认 metadata 缺失、版本过旧、探测失败，或没有分别给出 source/output 的 Dolby Vision 判定时，TV App 首轮一律按“不能确认 DV 专用链路播放源资格”处理，不根据文件名、视频类型或历史经验猜测可安全放行。`CONTEXT.md` 已追加 `DV metadata 不完整不推断放行` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:52 +0800
- 进度：grill-with-docs 继续收敛 `杜比视界专用系统播放链路` 的播放源资格。已确认专用系统播放链路只能使用当前仍可访问、且 metadata 表明播放源自身仍是 Dolby Vision 的媒体文件；已转码成非 DV 的输出文件和已删除/不可访问的原始上传文件都不能进入安全放行。`CONTEXT.md` 已追加 `DV 专用链路播放源资格` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:51 +0800
- 进度：grill-with-docs 继续收敛 `DV 安全放行` 的承载链路。已确认当设备显示能力和当前播放链路都明确支持对应 Dolby Vision 风险源时，首轮也只允许进入 `杜比视界专用系统播放链路`，不让现有 TV 长视频 `LibVLC` 普通链路承担 DV 安全播放；`CONTEXT.md` 已追加 `DV 支持放行只进专用链路` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:45 +0800
- 进度：grill-with-docs 继续收敛 `明确不支持` 场景下的重试边界。已确认当 DV 风险源被明确判定为当前设备/链路不支持安全播放时，“重试”仍只适用于本地显示模式、外接链路或相关系统状态发生变化后的幂等重判，不承诺修复。`CONTEXT.md` 已扩展 `DV 重试依赖本地状态变化`，让它同时覆盖能力未知和明确不支持两种阻断原因。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:41 +0800
- 进度：grill-with-docs 继续收敛 `明确不支持` 的失败态。已确认如果设备能力或播放链路能力被明确判定为不支持安全播放 Dolby Vision 风险源，首轮仍沿用阻断页语义，只把原因文案改为“当前设备不支持安全播放”一类解释；交互入口继续限制为“返回”和“重试”，不提供旁路播放。`CONTEXT.md` 已追加 `DV 不支持阻断不提供强行播放` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:37 +0800
- 进度：grill-with-docs 继续收敛 `重试` 的适用前提。已确认 Dolby Vision 风险阻断页上的“重试”只有在本地能力判断依赖的状态刚发生变化时才有意义；如果显示模式、外接链路或相关系统状态没有变化，重复点击本质上只是幂等重判。`CONTEXT.md` 已追加 `DV 重试依赖本地状态变化` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:33 +0800
- 进度：grill-with-docs 继续收敛 `重试` 的语义。已确认 Dolby Vision 风险阻断页里的“重试”首轮只表示重新执行一次本地能力判断并重走同一套播放决策，不切换播放器分支、不改写服务端 metadata，也不触发服务端修复；`CONTEXT.md` 已追加 `DV 重试属于幂等重判` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:32 +0800
- 进度：grill-with-docs 继续收敛 `未知阻断` 的交互边界。已确认 Dolby Vision 风险源因能力未知而被阻断时，首轮失败态只提供“返回”和“重试”这类非放行入口，不提供“继续播放/忽略风险播放”旁路；`CONTEXT.md` 已追加 `DV 未知阻断不提供强行播放` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:30 +0800
- 进度：grill-with-docs 继续收敛 `DV 能力未知` 的失败路径。已确认一旦视频已被判定为 `Dolby Vision 风险源`，而设备/HDR 能力探测结果又是 `未知`，则不允许退回当前 `LibVLC` 普通链路尝试播放；`CONTEXT.md` 已收紧 `DV 能力未知不进入安全放行分支` 术语，使其与“不异色优先”目标一致。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 10:28 +0800
- 进度：grill-with-docs 继续收敛“退回普通非 DV 播放链路”的语义。已确认这里的“普通链路”首轮只指当前已存在的 TV 长视频 `LibVLC` 链路，不自动切到新的系统播放器 / `Media3` 分支；`CONTEXT.md` 已追加 `普通非 DV 播放链路` 术语，避免与未来 `杜比视界专用系统播放链路` 混淆。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:48 +0800
- 进度：grill-with-docs 继续收敛 DV 能力探测的失败路径。已确认 `未知` 结果只是不允许进入 `DV 安全放行分支`，但不等于一刀切阻断所有播放；系统仍可退回普通非 DV 播放链路。`CONTEXT.md` 已追加 `DV 能力未知不进入安全放行分支` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:40 +0800
- 进度：grill-with-docs 继续收敛 `设备/HDR 能力探测` 的首轮范围。已确认首轮只覆盖 `Dolby Vision`，不把 `HDR10 / HDR10+ / HLG` 一并纳入；`CONTEXT.md` 已追加 `DV 能力探测首轮仅覆盖 Dolby Vision` 术语，避免能力矩阵和提示文案在首轮过度膨胀。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:33 +0800
- 进度：grill-with-docs 继续收敛 `设备/HDR 能力探测` 的语义。已确认这里的“能力”以“显示能力 ∩ 当前实际播放链路能力”为准，不接受只看电视面板的简化定义；`CONTEXT.md` 已追加 `设备/HDR 能力以显示能力与播放链路能力交集为准` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:28 +0800
- 进度：grill-with-docs 继续收敛 DV 第二阶段的进入顺序。已确认如果后续要把现状从“DV 风险探测与阻断可行”推进到“有条件放行”，第一优先级先补 `设备/HDR 能力探测`，不先做 `专用系统播放分支`；`CONTEXT.md` 已追加 `设备/HDR 能力探测先于 DV 放行分支` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:24 +0800
- 进度：grill-with-docs 继续收敛当前 DV 主链路的“可行性”定义。已确认本轮只把现状收口为 `DV 风险探测与阻断可行`：后端 `playback_compat` 探测与落库、TV 端读取 metadata 并保守阻断已打通；不把现状扩写成“DV 安全放行可行”，因为设备 HDR 能力探测和长视频专用系统播放分支仍未落地。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:17 +0800
- 进度：grill-with-docs 继续收敛 `DV→SDR` 风险边界。已确认当前保留的 `杜比视界转码输出阻断` 只属于 TV 端安全播放策略，不升级成全客户端统一阻断；`CONTEXT.md` 已追加 `DV→SDR 风险阻断限于 TV 端` 术语，明确手机端和其它客户端是否跟进需要后续单独决策。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-13 09:14 +0800
- 进度：grill-with-docs 收敛 “DV 视频压成 SDR 会不会有问题” 的产品语义，确认继续保留现有 `杜比视界转码输出阻断` 立场，不把 `DV→SDR` 自动视为可信播放源；同时在 `CONTEXT.md` 追加 `DV→SDR 兼容输出非默认可信播放源` 术语，明确只有后续显式定义独立色调映射方案与验收标准时，才有资格重新讨论放行。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；待执行 `rg -n $'\uFFFD' CONTEXT.md plan.md`、`git diff --check`。保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-12 20:46 +0800
- 进度：手机端与 TV 端安装包分轨已落地。后端在保留 `tv_app_releases` / `tv_app_release_apks` 表名的前提下新增 `client_type`、手机端单包固定槽位 `single`、轨内 `versionCode` 唯一且 `versionName` 冲突拒绝；家庭页新增 `/downloads/android-tv` 与 `/downloads/android-phone`，并保留 `/tv-app` 兼容入口；Admin Web 已提升为同一页内切换手机端 / TV 端的统一安装包管理工具。
- 影响文件：`migrations/0024_app_apk_distribution_client_type.*`、`internal/models/tv_apk.go`、`internal/models/admin.go`、`internal/services/tv_apk.go`、`internal/services/tv_apk_manager.go`、`internal/repository/tv_apk_repository.go`、`internal/repository/migrations_test.go`、`internal/handlers/router.go`、`internal/handlers/tv_apk.go`、`internal/handlers/tv_app_family_page.go`、`internal/handlers/tv_app_family_page_test.go`、`internal/handlers/iptv.go`、`internal/handlers/iptv_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/TvAppManage.vue`、`admin-web/src/components/base/commandPalette.helpers.js`、`CONTEXT.md`、`plan.md`
- 验证：`env GOCACHE=/private/tmp/go-build-cache-tvapk go test ./internal/services ./internal/repository ./internal/handlers -count=1` 通过；`cd admin-web && npm run test -- --run src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过。提交时继续排除用户已有 `admin-web/.env.development`，并保留尚未提交的文档 ADR 工作区改动按本任务一并纳入。

## 2026-06-12 20:14 +0800
- 进度：开始实现“安装包按客户端类型分轨”。本轮先落后端与 schema：保留现有 `tv_app_releases` / `tv_app_release_apks` 表名，在不 rename 的前提下补 `client_type`、手机端单包承载槽位、包名强校验与轨内规则，再接家庭下载页和管理端切换。
- 影响文件：`migrations/0024_*`、`internal/models/tv_apk.go`、`internal/services/tv_apk*.go`、`internal/repository/tv_apk_repository.go`、`internal/handlers/tv_apk.go`、`internal/handlers/router.go`、`internal/handlers/tv_app_family_page.go`、`admin-web/src/*`、`CONTEXT.md`、`plan.md`
- 验证：待执行 Go 定向测试、`admin-web` 单测与构建；保留用户已有 `admin-web/.env.development` 不纳入提交。

## 2026-06-12 19:53 +0800
- 进度：开始 grill-with-docs 收敛“手机端的 app 也做一下管理”。已确认新需求不是手机端业务功能管理，而是 Android 手机端 APK 分发管理；并已收敛出若干上位术语：手机端与 TV 端共用同一套发布记录/安装包产物模型但按 `android_phone` / `android_tv` 分轨，手机端首期是单包发布记录、家庭轻量页与 TV 分开、管理端工具同页切客户端类型、家庭下载路由按客户端类型规范化命名。下一步继续逐项钉死手机端与 TV 端共享到什么程度，以及哪些地方必须分叉。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 20:07 +0800
- 进度：手机端 APK 分发管理的上位设计已收敛完成，并补充 ADR。当前已确认：手机端与 TV 端共用一套安装包分发模型并按 `client_type` 分轨；手机端首期是单包发布记录、家庭轻量页独立、管理端同页切客户端类型、`client_type` 与 `packageName` 强绑定、轨内 `versionCode` 唯一、手机端已发布替换前也必须先下线。已新增 ADR `0009-unified-app-apk-distribution-by-client-type.md` 固化这些不可逆取舍。下一步可以直接转入实现。
- 影响文件：`CONTEXT.md`、`docs/adr/0009-unified-app-apk-distribution-by-client-type.md`、`plan.md`
- 验证：待执行文档乱码检查与 `git diff --check`。

## 2026-06-12 18:46 +0800
- 进度：开始补齐 TV 安装包家庭轻量页闭环。已确认当前仓库没有普通用户 Web 壳，因此本轮改动限定在 Go 侧：新增独立 `/tv-app` 轻量页，内置最小登录、现有 token 续期、最近三版展示与 ABI 直下能力，不扩展 `admin-web`。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/tv_app_family_page.go`、`internal/handlers/tv_app_family_page_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 family 页定向 handler 测试与 `go test ./internal/handlers ./internal/services ./internal/repository -count=1`。

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

## 2026-06-12 18:04 +0800
- 进度：进入实现阶段，已确认 TV APK 清单是二进制 AXML，不能依赖 `output-metadata.json` 或纯文本 manifest。下一步直接补最小 APK 解析器，并同步落仓库查询、上传/发布/下载 API 与 admin 前端入口。
- 影响文件：`internal/services/tv_apk.go`、`internal/repository/*`、`internal/handlers/*`、`admin-web/src/*`、`plan.md`
- 验证：待执行 Go 定向测试、`admin-web` 构建。

## 2026-06-12 17:46 +0800
- 进度：开始落代码，已核对实现承载面：后端沿用 `VideoRepository` + Gin handler + migration 体系，管理端沿用 `admin-web` 现有页面模式；家庭轻量下载页将单独实现，不复用 admin shell。下一步先补 TV APK 分发表结构、DTO 与仓库接口，再接管理/下载 API。
- 影响文件：`plan.md`
- 验证：待执行 migration 测试、Go 定向测试、`admin-web` 构建。

## 2026-06-12 17:45 +0800
- 进度：grill-with-docs 收敛“可上传不同版本的 TV App”需求，确认这套规则值得单独沉淀 ADR，并新增 `docs/adr/0008-tv-apk-manual-distribution.md`，记录“手动 TV APK 分发 + 发布记录/ABI 双层模型 + 家庭轻量页”的架构取舍与原因。当前 `CONTEXT.md` 与 ADR 已形成互补：前者保留术语，后者保留关键取舍背景。下一步可结束 grill，会总当前收敛结果并转入实现规划。
- 影响文件：`docs/adr/0008-tv-apk-manual-distribution.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:42 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 高风险管理动作需要显式确认`：会改变家庭成员可见内容或可下载二进制的动作，如下线发布记录、恢复发布、替换已发布版本 ABI，执行前都应要求管理员显式确认。`CONTEXT.md` 已同步补充该术语。下一步继续收敛最后一个交付边界：是否需要为这套 TV 安装包分发规则单独沉淀 ADR。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:39 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页显示 APK 文件大小`：家庭轻量下载页也应直接显示每个 APK 的文件大小，帮助判断下载体量与设备存储占用。`CONTEXT.md` 已同步补充该术语。下一步继续收敛高风险管理动作如“下线”“替换 ABI”是否需要显式操作确认。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:38 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 点击下载即直接开始`：无论在家庭轻量页还是管理端，点击某个 APK 的下载入口后都应直接开始下载，不额外弹出二次确认。`CONTEXT.md` 已同步补充该术语。下一步继续收敛家庭轻量页是否也要直接显示 APK 文件大小，帮助家庭成员判断下载体量。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:37 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端列表显示 APK 文件大小`：管理端列表层应直接显示各已上传 APK 的文件大小，帮助快速判断上传产物是否异常，以及不同 ABI 包的大致体量差异。`CONTEXT.md` 已同步补充该术语。下一步继续收敛下载动作本身是否需要二次确认，还是直接开始下载。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:26 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端列表同时显示已上传与缺失 ABI`：管理端列表层应同时显示每条发布记录已上传的 ABI 与缺失的 ABI，便于直接判断当前资产覆盖面和下一步补包目标。`CONTEXT.md` 已同步补充该术语。下一步继续收敛是否还要在列表层直接显示每个 APK 的文件大小，还是保持到详情/下载时再看。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:24 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端列表直显缺失 ABI`：管理端列表层就应直接显示每条发布记录缺少哪些 ABI，不要求管理员进入详情页后再看缺包项。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端列表是否还需要同时直显“已上传哪些 ABI”，还是缺失项已经足够指导操作。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:22 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端支持按 ABI 完整性筛选`：管理端应至少支持快速定位“缺少 ABI”的记录，且 ABI 完整性筛选可与状态筛选、版本检索叠加，便于管理员跨状态补包。`CONTEXT.md` 已同步补充该术语。下一步继续收敛当管理员找到缺包记录后，是否需要在列表层直接显示“缺哪些 ABI”，还是只进详情页再看。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:21 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端搜索与状态筛选可叠加`：管理端的版本检索与状态筛选应允许同时生效，管理员可以先定位某个版本号，再继续只看它在草稿、已下线或已发布状态中的子集。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端是否还需要按 ABI 完整性做显式筛选，还是状态筛选已经足够表达。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:19 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端支持按版本号或时间检索归档`：管理端应支持按 `versionCode`、`versionName` 或时间维度检索/筛选发布记录，便于快速定位较老归档版本，而不是只靠长列表滚动查找。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端搜索命中后，是否仍应保留状态筛选叠加，而不是两者二选一。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:17 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 草稿即使 ABI 齐全也不额外高亮`：尚未发布但 ABI 已齐全的记录，首期仍只按普通草稿对待，不单独升格为待发布关注项或额外高亮。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端是否需要提供按 `versionCode` 或时间的搜索/筛选，以便快速定位老版本归档。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:11 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端默认先看当前家庭可见记录`：管理端列表默认打开时，应优先落在当前家庭可见记录视图，而不是全量归档或草稿筛选，让管理员第一眼看到家庭成员现在实际能看到和下载的版本。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端是否需要把“草稿但 ABI 已齐全、尚未发布”的记录单独标成待发布关注项。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:10 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端按状态显式筛选`：管理端应至少支持按草稿、`已发布-完整`、`已发布-缺少 ABI`、`已下线` 进行显式筛选，而不是把全部记录混排后靠人工滚动查找。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端列表默认打开时应优先落在哪个筛选视图，以减少管理员日常操作成本。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:08 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 下线记录不占管理端最近三版快捷视图`：若管理端提供“最近三版”快捷视图，其语义应与家庭可见窗口一致，仅聚焦当前家庭可见的最近三条记录；已下线记录改在全量归档列表中查看和处理。`CONTEXT.md` 已同步补充该术语。下一步继续收敛管理端是否需要为“草稿 / 已发布 / 已下线”提供显式筛选维度，而不是只靠列表混排。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:07 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 替换 ABI 不刷新原始上传时间`：同一发布记录下替换某个 ABI 安装包时，不应改写原始上传时间，只刷新最近状态变更时间。`CONTEXT.md` 已同步补充该术语。下一步继续收敛已发布记录被下线后，是否仍应继续保留在管理端的“最近三版”管理视角里，还是只按全量归档列表查看。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:06 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端展示原始上传时间`：管理端还应展示每条发布记录的原始上传时间，即首个 APK 首次进入该记录并触发自动建档的时间；它与首次发布时间、最近状态变更时间分开维护。`CONTEXT.md` 已同步补充该术语。下一步继续收敛这三个时间字段里，替换同 ABI 安装包时是否要刷新原始上传时间，还是仅刷新最近状态变更时间。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:04 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 管理端展示最近状态变更时间`：管理端需要单独展示每条发布记录的最近一次状态变更时间，用于审计和操作判断；它与家庭页展示的首次发布时间分开维护，且不参与推荐位或版本排序。`CONTEXT.md` 已同步补充该术语。下一步继续收敛发布记录是否需要保留“首次进入家庭可见时间”和“最近状态变更时间”以外的第三个历史字段，比如原始上传时间。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:02 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页展示发布时间` 的语义为“首次进入家庭可见范围的时间”：该时间在下线、恢复可见或其它状态变更后都不重置，只用于帮助家庭成员理解版本先后。`CONTEXT.md` 已同步细化该术语。下一步继续收敛管理端是否需要同时展示“首次发布时间”和“最近状态变更时间”用于审计与运营判断。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 17:01 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页展示发布时间`：轻量下载页的每个发布记录应展示发布时间，帮助家庭成员区分最新推荐版与较早回退版，但该时间不参与版本先后计算。`CONTEXT.md` 已同步补充该术语。下一步继续收敛这里的“发布时间”具体是首次发布到家庭页的时间，还是最近一次重新发布/恢复可见的时间。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:58 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV ABI 下载失败不自动改推其它 ABI`：家庭成员下载某个 ABI 失败时，页面应明确报错，但不自动建议切换到另一 ABI，继续保持 ABI 由使用者显式选择的责任边界。`CONTEXT.md` 已同步补充该术语。下一步继续收敛轻量页是否需要展示发布时间，帮助家庭成员区分最近上传与较早回退版本。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:57 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页展示简短版本说明`：轻量下载页的每个发布记录应展示简短版本说明或发布备注，帮助家庭成员判断推荐版与回退版，但不引入管理端式长表单。`CONTEXT.md` 已同步补充该术语。下一步继续收敛当某一 ABI 下载失败时，页面是否需要回退到另一个 ABI，还是保持用户显式选择责任。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:56 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 下载入口以 ABI 标签为主`：轻量下载页的下载动作应以“下载 arm64-v8a / 下载 armeabi-v7a”这类 ABI 标签按钮为主，原始 APK 文件名只作辅信息，不让家庭成员依赖文件名判断。`CONTEXT.md` 已同步补充该术语。下一步继续收敛下载页是否需要同时展示版本说明，还是只保留版本号和 ABI 下载入口。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:53 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页无记录空状态`：当家庭成员可见的已发布记录为空时，轻量下载页应展示明确的“暂无可下载 TV 安装包”空状态，不保留历史推荐占位，也不暴露管理端草稿信息。`CONTEXT.md` 已同步补充该术语。下一步继续收敛轻量页上下载动作是否直接暴露文件名，还是需要再包一层 ABI 标签化下载按钮。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:49 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 推荐位下线后立即重算`：若当前推荐版本被下线或退出家庭可见范围，系统应立刻从剩余家庭可见的已发布记录中按 `versionCode desc` 自动选出新的推荐版本，不允许推荐位悬空。`CONTEXT.md` 已同步补充该术语。下一步继续收敛当家庭轻量页没有任何已发布记录时，是否要展示明确空状态，而不是保留历史推荐占位。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:47 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 已发布记录替换前需先下线`：已对家庭成员可见的发布记录若要替换某个 ABI 安装包，不能让新 APK 在原发布态下静默在线生效；应先下线、替换，再按规则恢复发布。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“下线”动作是否会影响推荐位和最近三版的即时重算。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:46 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布前至少已有一个可下载 ABI`：管理员只有在发布记录至少已上传 1 个可下载 APK 时才能执行“发布”；空记录即使已自动建档，也只能停留在管理端。`CONTEXT.md` 已同步补充该术语。下一步继续收敛已发布记录若被替换成新的同 ABI APK，是否允许在不下线的情况下直接保持对家庭成员可见。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:45 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 缺少 ABI 也可显式发布`：管理员执行“发布”时，即使发布记录的 ABI 仍未补齐，也允许其进入家庭轻量下载页，并以 `已发布-缺少 ABI` 状态展示。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“发布”动作是否应限制至少存在 1 个可下载 ABI，还是允许空记录进入家庭视线。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:44 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 推荐位与最近三版只看家庭可见记录`：推荐版本与最近三版窗口的计算范围只包括当前对家庭成员可见的已发布记录；未显式发布的新记录、已下线记录和仅管理端可见的草稿记录都不能参与计算。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“发布”动作本身是否允许 ABI 不完整的版本直接进入家庭视线。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:43 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 新版本显式发布后才进入家庭视线`：新 `versionCode` 的首个 Release APK 上传完成并自动建档后，发布记录默认先只在管理端可见，只有管理员显式执行“发布”后才进入家庭轻量下载页。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“最近三版”和“推荐版本”的计算基准，是否只看家庭当前可见的已发布记录。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:42 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 跌出最近三版自动退场`：当更高 `versionCode` 的新发布记录进入最近三版窗口后，超出窗口的更老记录应自动从家庭轻量下载页消失，只在管理端保留归档。`CONTEXT.md` 已同步补充该术语。下一步继续收敛新上传版本在首个 APK 到位后，是自动进入家庭可见范围，还是先停留在管理端待显式发布。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:41 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 下线记录自动恢复发布态`：管理员把 `已下线` 的发布记录重新恢复到家庭轻量下载页时，系统应按当前 ABI 完整性自动回到 `已发布-完整` 或 `已发布-缺少 ABI`，不要求额外的再发布确认。`CONTEXT.md` 已同步补充该术语。下一步继续收敛当新版本挤掉旧版本、使其跌出最近三版窗口时，旧记录是自动下线还是仍保留家庭成员可见归档。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:40 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV ABI 补齐自动转完整态`：处于 `已发布-缺少 ABI` 的发布记录在补上传最后一个缺失 ABI 后，应自动切换为 `已发布-完整`，不需要管理员再次执行单独发布确认。`CONTEXT.md` 已同步补充该术语。下一步继续收敛 `已下线` 记录重新恢复到家庭轻量页时，是自动回到对应发布态，还是要求一次显式再发布。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:39 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录可见即已发布`：只要发布记录进入家庭成员可见的轻量下载页，就视为已发布，不要求 ABI 必须齐全；仅管理端可见的记录视为未发布或未进入分发视线。`CONTEXT.md` 已同步补充该术语。下一步继续收敛已发布记录在 ABI 补齐前后是否需要额外状态标签，以及下线与重新发布的切换规则。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:33 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 正式分发只收 Release APK`：首期只接受可分发的 Release APK，不允许 debug 包进入家庭分发链路；若解析结果显示 `debuggable=true` 或其它明确 debug 标记，应直接拒绝上传。`CONTEXT.md` 已同步补充该术语。下一步继续收敛误传或过期版本的删除边界，明确管理员能否删除发布记录以及何时允许删除。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:28 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 轻量页只见最近三版`：家庭成员轻量下载页只展示最近 3 个发布记录（推荐 1 个 + 回退 2 个），不暴露更老归档版本；更老历史只在管理端可见，由管理员处理。`CONTEXT.md` 已同步补充该术语。下一步继续收敛跨版本安装成功所依赖的签名一致性边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:27 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录三态`：首期至少区分 `已发布-完整`、`已发布-缺少 ABI`、`已下线` 三种状态，并要求它们与轻量页展示、管理端筛选、删除/下线规则保持一致。`CONTEXT.md` 已同步补充该术语。下一步继续收敛 ABI 补齐时状态是否自动从“已发布-缺少 ABI”转为“已发布-完整”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:22 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 推荐版本自动指向最新 versionCode`：轻量下载页需要有“推荐安装”标记，且首期自动指向 `versionCode` 最高的发布记录，不提供管理员手动改推荐；更早两个保留版本作为回退候选显示。若推荐版本缺少部分 ABI，仍保留推荐标记并继续显示 ABI 缺失提示。`CONTEXT.md` 已同步补充该术语。下一步继续收敛三版之外的归档版本是否对家庭成员可见。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:21 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录下线替代删除`：已进入家庭分发视线的发布记录首期不能直接删除，只能下线或软隐藏；下线后不再出现在轻量下载页，但管理端仍保留记录与审计。只有尚未进入家庭默认视图的无效记录或误建空壳才允许彻底删除。`CONTEXT.md` 已同步补充该术语。下一步继续收敛一个版本何时才算“已发布”，以及部分 ABI 版本是否天然就算已发布。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:17 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 安装包显式选 ABI`：轻量下载页不做设备自动识别或自动跳转，必须显式展示各 ABI 的下载入口，让使用者自己选 `arm64-v8a` 或 `armeabi-v7a`；缺包时仍要明示已上传与缺失状态。`CONTEXT.md` 已同步补充该术语。下一步继续收敛是否存在“推荐版本”标记，以及谁能改变它。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:12 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 安装包轻量页复用现有登录`：轻量下载页首期直接复用现有账号密码登录，家庭成员用普通已登录账号访问；不做分享链接、一次性口令或免登下载。`CONTEXT.md` 已同步补充该术语。下一步继续收敛轻量下载页里 ABI 的选择方式，是自动猜测目标设备还是显式让人选包。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。
## 2026-06-12 16:09 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认采用 `TV 安装包轻量下载页`：家庭成员通过一个登录后可访问的轻量下载页拿安装包，不复用 `admin-web` 后台页面。该页面只做版本查看、ABI 状态查看和下载，不带后台导航或管理动作。`CONTEXT.md` 已同步补充该术语。下一步继续收敛这个轻量页的登录方式，是复用现有账号密码登录还是另做分享式入口。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 16:05 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 安装包访问权限分层`：上传/替换/编辑说明入口仅对 `admin` 开放，下载列表与安装包文件下载可对已登录家庭成员账号开放，但不匿名公开。结合代码现状核对，仓库现有 Web 界面只有 `admin-web` 一套且明确是 admin 受限，尚不存在非 admin 的现成下载页；下一步需继续收敛家庭成员通过什么界面拿到安装包。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，并核对 `admin-web/src/router/index.js`、`internal/middleware/auth.go`、`admin-web/AGENTS.md` 当前权限边界；未执行构建或测试。

## 2026-06-12 16:00 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录编辑边界`：发布记录自动建档后，管理员只编辑版本说明、备注等附属文案；`versionCode`、`versionName`、`packageName`、`abi`、上传文件与 ABI 完整性保持只读，文件变更只能走显式替换动作。`CONTEXT.md` 已同步补充该术语。下一步继续收敛下载入口的访问权限边界，明确家庭成员如何拿到安装包。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:56 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录自动建档`：上传某个新 `versionCode` 的首个 APK 时系统自动创建发布记录，后续其它 ABI 包挂到同一记录下；首期不提供“先手动建空记录、再回头补包”的流程。`CONTEXT.md` 已同步补充该术语。下一步继续收敛创建后管理员还能编辑哪些字段，以及哪些字段必须保持只读。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:52 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 发布记录排序基准`：版本新旧顺序只看 `versionCode`，`versionName` 只做展示，上传时间不参与判定。默认下载主视图、最近三版自动保留和回退窗口都按 `versionCode desc` 计算。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“发布记录”的创建方式，是上传首个 ABI 包时自动建档，还是先手动创建空记录再补包。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:48 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 最近三版自动保留`：回退窗口不靠管理员手动钉住，而是按版本顺序自动维护最近 3 个发布记录进入默认下载集合（最新 1 个 + 回退 2 个），更老版本只做归档不进主视图。`CONTEXT.md` 已同步补充该术语。下一步继续收敛“版本顺序”具体以 `versionCode`、`versionName` 还是上传时间为准。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:36 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，澄清并确认 `TV 发布记录回退窗口`：这里的“回退”指在不同发布记录之间回退，而不是在同一版本号下保留多个 ABI 修订包。首期应保留当前版本之前两个更早的 `TV 发布记录` 作为可下载回退候选。`CONTEXT.md` 已同步补充该术语。下一步继续追问这两个老版本是自动保留还是管理员手动钉住。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:31 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV ABI 显式替换`：同一发布记录下若某个 ABI 包已存在，系统默认拒绝再次上传同 ABI 的 APK，避免静默覆盖；只有管理员显式执行“替换该 ABI 安装包”动作时，才允许用新包覆盖旧包。`CONTEXT.md` 已同步补充该术语。下一步继续追问替换后是否保留旧包可追溯，还是只保留当前生效包。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 15:26 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 package identity 的硬准入边界：[[管理端 TV 安装包管理工具]] 首期只接受 `packageName = com.chee.videos.tv` 的 APK，解析结果若不是这个包名则直接拒绝上传，不把其他 Android 应用混入 TV 安装包列表。至此已锁定 8 个核心边界：管理端工具、手动分发、发布记录挂多 ABI 包、部分 ABI 版本可见、显式列出缺失 ABI、本地 APK 上传、APK 解析为核心元数据可信源、包名硬准入。下一步继续收敛“同一 ABI 重复上传”与“新包是否覆盖旧包”的版本替换规则。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；未执行构建或测试。

## 2026-06-12 09:44 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认采用 `TV 安装包核心元数据可信源`：`versionName`、`versionCode`、`packageName`、`abi` 等核心元数据必须从 APK 文件解析并校验，管理员不手填也不能覆盖；若解析结果与目标发布记录不一致，应直接拒绝。补充核对 TV 工程现状，`android-tv-app/tv-app/build.gradle.kts` 当前固定 `applicationId = com.chee.videos.tv` 且 ABI 仅拆 `armeabi-v7a` / `arm64-v8a`，现有 `release/output-metadata.json` 也印证“同一版本下两个 ABI 产物”的形态。下一步继续收敛 package identity 不一致时的准入边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，并核对 `android-tv-app/tv-app/build.gradle.kts` 与 `android-tv-app/tv-app/release/output-metadata.json`；未执行构建或测试。

## 2026-06-12 09:40 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认上传入口采用 `TV 本地 APK 上传首期`：首期只支持管理员上传本地 APK 文件，不支持远程下载 URL 或外部制品库同步。结合代码现状核对，仓库已有成熟的 `multipart/form-data` 上传模式可复用，但尚无现成 APK 元数据解析能力；下一步继续追问版本元数据究竟由管理员手填还是由 APK 解析并校验。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，并核对 `internal/handlers/iptv.go`、`admin-web/src/api/admin.js` 现有文件上传模式；未执行构建或测试。

## 2026-06-12 09:36 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认 `TV 部分可下载发布记录` 的展示必须带 `TV ABI 完整性提示`：下载视图不能只给模糊状态，而要显式列出“已上传哪些 ABI 包、缺少哪些 ABI 包”。`CONTEXT.md` 已同步补充该术语，下一步继续收敛上传链路与版本元数据的可信来源。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-12 09:32 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，用户明确否决“缺少 ABI 就默认隐藏”的边界，改为采用 `TV 部分可下载发布记录`：某个版本即使只上传了部分 ABI 包，也允许出现在下载视图并下载已上传的包。`CONTEXT.md` 已同步补充该术语，下一步继续追问这种“可见但不完整”的状态如何标记与约束。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-12 09:28 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认版本模型采用 `TV 发布记录 + TV ABI 安装包`：一个 `versionName + versionCode` 对应一条发布记录，下面再挂 `armeabi-v7a` / `arm64-v8a` 等具体 APK 产物，而不是把每个 ABI 包当独立版本。`CONTEXT.md` 已同步补充两个术语，下一步继续追问“只上传了一部分 ABI 包时”的可见性与发布边界。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-12 09:24 +0800
- 进度：grill-with-docs 继续收敛“可上传不同版本的 TV App”需求，确认首期采用 `TV 安装包手动分发首期`：由管理员在 Admin Web 上传版本包，管理员或家庭成员手动下载安装；本期不做 TV 客户端内检查更新、自动升级、灰度发布或手机端中转分发。`CONTEXT.md` 已同步补充该术语，下一步继续收敛版本记录与 ABI 包的组织关系。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-12 09:21 +0800
- 进度：grill-with-docs 开始收敛“可上传不同版本的 TV App”需求，先确认采用 `管理端 TV 安装包管理工具` 作为上位术语：范围锁定为 Admin Web 内的管理员工具页，用于上传、保存、查看和分发多个 TV APK 版本；暂不把需求误扩展为 TV 客户端内自升级器或版本切换器。`CONTEXT.md` 已同步补充该术语，后续继续围绕消费方、包格式和版本组织方式逐项追问。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-12 09:05 +0800
- 进度：开始清理未接入的旧任务化生图残留。已确认当前主链路采用“同步生成 + 浏览器本地历史”，本轮仅删除未跟踪的服务端任务化草稿与对应迁移文件，避免后续被误当成正式方案或被全量 migration 误执行。
- 影响文件：`internal/services/image_generation.go`、`migrations/0023_image_generation_tasks.up.sql`、`migrations/0023_image_generation_tasks.down.sql`、`plan.md`
- 验证：待执行 `git status --short`

## 2026-06-12 09:07 +0800
- 进度：完成未接入的旧任务化生图残留清理。当前工作区仅保留 `plan.md` 账本更新与用户已有的 `admin-web/.env.development` 本地环境改动，已避免未跟踪 `0023` migration 被后续流程误执行。
- 影响文件：`plan.md`
- 验证：`git status --short` 确认仅剩 `plan.md` 与未纳入的 `admin-web/.env.development`

## 2026-06-12 00:39 +0800
- 进度：完成 admin 图像工作台“参考图局部蒙版编辑”主链路。前端新增本地蒙版编辑弹窗、蒙版目标预览、前序结果复用来源标记与最小本地快照字段；后端 `/admin/image-generation/generate` 新增 `mask` 入参、目标槽位校验、尺寸校验以及执行层首图重排后再向上游 `/images/edits` 提交蒙版。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/ImageWorkbenchMaskEditor.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test -- src/views/imageWorkbench.helpers.spec.js src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过；`go test ./internal/handlers -count=1` 通过。未纳入 `admin-web/.env.development` 与未跟踪的旧任务化残留文件。

## 2026-06-12 00:15 +0800
- 进度：开始把 admin 图像工作台继续收口到“图片生成主链路”，本轮只实现参考图局部蒙版编辑与提交闭环，避免继续展开本地历史微规则；计划补 `admin-web` 本地遮罩编辑弹窗、最小任务快照字段，以及 `internal/handlers/admin_image_generation.go` 的 `mask` 转发与校验。
- 影响文件：`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.helpers.js`、`internal/handlers/admin_image_generation.go`、`plan.md`
- 验证：待执行 `cd admin-web && npm run build`、`go test ./internal/handlers -count=1`

## 2026-06-12 00:00 +0800
- 进度：grill-with-docs 继续收敛本地历史待导入项的默认聚焦顺序，确认当同时存在多条“待导入的成功项”时，默认仍按创建时间倒序聚焦最新生成的一条；`CONTEXT.md` 已补充该排序语义，和现有本地历史倒序读取口径保持一致。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:53 +0800
- 进度：grill-with-docs 继续收敛工作台重开时的默认焦点边界，确认“待导入的成功项”优先于较新的失败项作为默认选中目标；`CONTEXT.md` 已补充该规则，确保回到工作台后默认先落在最接近闭环的待办上。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:49 +0800
- 进度：grill-with-docs 继续收敛失败历史项在“继续改稿”里的失效输入边界，确认即使参考图或蒙版字节已在当前浏览器本地失效，系统仍允许进入继续改稿，但必须把失效输入显示为缺失位，并阻止按原样直接提交，直到管理员补齐、重画或显式移除。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:47 +0800
- 进度：grill-with-docs 继续收敛失败历史项的保留边界，确认失败历史也要保留完整输入快照（提示词、参数、参考图、蒙版），并继续提供“继续改稿/重试”入口，避免失败记录退化成只剩错误说明的死历史。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:44 +0800
- 进度：grill-with-docs 继续收敛本地不可恢复删除边界，确认“尚未导入图库、且删除后也不会被其他本地历史复用”的本地结果图在删除前必须显式确认，并明确提示删除后仅当前浏览器内不可恢复；`CONTEXT.md` 已补充该确认语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:22 +0800
- 进度：grill-with-docs 继续收敛前序结果派生关系在本地历史删除后的边界，确认“删除来源历史只影响追溯提示，不破坏已派生历史本身”；也就是派生历史仍保持可查看、可继续改稿，只把来源历史已删除作为轻量摘要提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:19 +0800
- 进度：grill-with-docs 继续收敛本地历史删除后的数据回收边界，确认删除一条本地历史时应回收“只被这条历史独占、且未被其他历史复用”的原图、缩略图、蒙版等本地字节，但不能误删仍被其他历史项引用的内容；`CONTEXT.md` 已补充这条孤儿回收语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，并核对现有图片存储以 `hashDataUrl()` 生成内容哈希 ID 复用同图数据；未执行构建或测试。

## 2026-06-11 23:16 +0800
- 进度：grill-with-docs 继续收敛本地历史删除边界，确认删除 `admin 图像创作历史` 只影响当前浏览器里的本地历史与创作上下文，不影响已经通过“导入图库”进入正式生命周期的图片资产；`CONTEXT.md` 已补充该边界，避免后续把本地清理误做成反向删库。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:11 +0800
- 进度：grill-with-docs 继续收敛本地历史与导入闭环的持久化边界，确认“某张结果是否已导入、导入到了哪个资产、当前有效关联指向哪一个资产”也要写回浏览器本地历史，刷新页面后不能丢；否则待导入、已导入、重新导入和跳转资产这些后续处理上下文会失忆。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新；同时核对现有代码仅有 `admin-web/src/views/imageWorkbench.db.js` 的 `deleteWorkbenchTask()` 底层 helper，未发现已定型的历史删除交互，未执行构建或测试。

## 2026-06-11 23:08 +0800
- 进度：grill-with-docs 继续收敛本地历史可见范围，确认 `admin 图像创作历史` 只保证当前浏览器本地留存，不做跨浏览器、跨设备或跨管理员同步；`CONTEXT.md` 已补充该边界，后续共享与长期保留仍只通过显式导入图库闭环。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 23:02 +0800
- 进度：grill-with-docs 根据用户最新澄清“历史只存在本地，不用入库”纠偏生图术语边界，撤回此前误收敛出的“服务端任务化 / 服务端图像创作历史”方向；`CONTEXT.md` 已统一改写为“同步代理 + 浏览器 IndexedDB 本地历史 + 显式导入图库”的口径，并保留仍成立的局部蒙版、前序结果继续改、单张/批量导入交互语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，基于现有代码核对 `admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.db.js`、`internal/handlers/admin_image_generation.go`，未执行构建或测试。

## 2026-06-11 22:48 +0800
- 进度：grill-with-docs 继续收敛双轨过渡期的本地历史结构，确认采用 `admin 同步工作台本地历史与任务快照对齐`：同步工作台的 IndexedDB 历史也要保存参考图来源类型、前序结果来源元数据和独立蒙版信息，不再停留在旧的扁平图片 ID 列表结构，以避免刷新或回看本地历史时丢失局部蒙版与来源关系。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:45 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的工作台行为，确认采用 `admin 移除蒙版不移除目标参考图`：管理员移除蒙版时，只清除蒙版输入与目标绑定，不自动删除这张目标参考图；草稿退回普通参考图编辑语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:43 +0800
- 进度：grill-with-docs 继续收敛局部蒙版在双轨过渡期间的交付范围，确认采用 `admin 局部蒙版双轨同步可用`：局部蒙版编辑不仅要进入新的任务化链路，也要在当前同步工作台链路里立即可用，避免 reference 里最关键的局部编辑能力在现有主工作台上出现空档。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:40 +0800
- 进度：grill-with-docs 继续收敛工作台恢复体验，确认采用 `admin 恢复有效蒙版不自动弹编辑器`：历史任务、重试预填或“以结果图继续改”恢复出有效蒙版草稿后，工作台默认只恢复状态，不自动弹出蒙版编辑器；需要修改蒙版时再由管理员显式进入。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:29 +0800
- 进度：grill-with-docs 继续收敛工作台恢复语义，确认采用 `admin 有效蒙版恢复为可继续编辑草稿`：只要局部蒙版文件仍然有效，从历史任务、重试预填或“以结果图继续改”返回工作台时，就应恢复成可继续编辑的蒙版草稿，而不是只展示静态合成预览。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:25 +0800
- 进度：grill-with-docs 继续收敛参考图来源类型枚举，确认采用 `admin 首轮参考图来源类型收口`：首轮 `request_payload.reference_images[]` 的 `source_kind` 只保留 `browser_input` 与 `previous_result` 两种；上传/粘贴统一折叠为 `browser_input`，也不提前为图库资产引用预留第三种来源类型。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:21 +0800
- 进度：grill-with-docs 继续收敛 `request_payload.mask` 的目标指向语义，确认采用 `admin 蒙版目标按参考图槽位绑定`：蒙版目标不绑定文件路径、执行层顺序或前端临时 image id，而是绑定管理员看到的参考图槽位号，保证详情、重试预填和失效回显围绕稳定槽位工作。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:19 +0800
- 进度：grill-with-docs 继续收敛任务输入快照的大结构，确认采用 `admin 蒙版快照独立于参考图数组`：局部蒙版不混进 `request_payload.reference_images[]`，而是作为独立 `mask` 字段建模，单独记录目标参考图槽位、蒙版文件引用和相关轻量元数据。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:16 +0800
- 进度：grill-with-docs 继续收敛参考图快照结构，确认采用 `admin 参考图快照显式来源类型`：`request_payload.reference_images[]` 的每个槽位都显式保存 `source_kind`，不再通过“是否存在来源任务/结果字段”去推断来源类型，保证详情、重试预填和审计使用统一结构。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:11 +0800
- 进度：grill-with-docs 继续收敛前序结果来源关系的失效回显语义，确认采用 `admin 前序结果来源失效仍保留摘要`：若来源结果文件后来被 TTL 清理或失效，任务详情仍保留来源任务/来源结果的轻量摘要与失效状态，只把“查看来源结果”降级为失效提示，不隐藏整段来源关系。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 22:02 +0800
- 进度：grill-with-docs 继续收敛前序结果来源关系的详情回显语义，确认采用 `admin 前序结果来源关系详情可见`：当某个参考图槽位来自前序生图结果时，任务详情要直接展示这层来源关系，并允许跳回来源任务或来源结果查看，不把来源元数据只留在内部快照里。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 21:58 +0800
- 进度：grill-with-docs 继续收敛前序结果复用的快照审计语义，确认采用 `admin 前序结果来源元数据进入参考图快照`：当某个参考图槽位来自前序生图结果时，任务快照不仅保存复制后的输入文件，还要保留 `source_kind=previous_result`、来源任务和来源结果这类轻量来源元数据，避免被扁平成普通临时参考图。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:53 +0800
- 进度：grill-with-docs 继续收敛“继续改稿”里的蒙版失效语义，确认采用 `admin 继续改稿蒙版失效保留目标图`：若历史任务的蒙版文件已失效，工作台仍允许进入继续改稿，并保留目标参考图、提示词和其他输入，但必须把蒙版明确标成需重画的失效位，不能静默降级成整图编辑。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:50 +0800
- 进度：grill-with-docs 继续收敛“继续改稿”和局部蒙版的衔接语义，确认采用 `admin 继续改稿预填局部蒙版`：如果历史任务本身使用了局部蒙版编辑，那么从任务详情进入“继续改稿”时，新草稿应一并预填蒙版目标图与蒙版输入，而不只回填提示词和参考图。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:48 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的任务快照语义，确认采用 `admin 蒙版快照保留原始参考图顺序`：即使执行层为上游请求把蒙版目标图临时重排到首位，任务输入快照仍保存管理员看到的原始参考图槽位顺序，并额外记录蒙版目标槽位，不把执行层顺序写回历史。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:47 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的历史回显语义，确认采用 `admin 蒙版详情默认回显合成预览`：任务详情默认把蒙版目标图显示为“目标图 + 蒙版”的合成预览，并明确标记其为蒙版目标图，而不是把蒙版本身伪装成一张独立普通参考图并排展示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:44 +0800
- 进度：grill-with-docs 继续收敛派生输入快照的保留策略，确认采用 `admin 派生输入快照与蒙版仍属临时留存`：前序结果复制出来的新任务输入图和与之配套的蒙版文件，仍然走统一 TTL 清理，不因被复用为新任务输入而升级为永久文件；过期后历史只保留轻量快照与失效提示。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:42 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的执行层顺序语义，确认采用 `admin 蒙版执行层首图重排`：启用局部蒙版时，服务端可仅在调用上游的执行层把蒙版目标图重排到首位，但工作台、任务详情和重试预填仍保留管理员看到的原始参考图槽位顺序。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:41 +0800
- 进度：grill-with-docs 继续收敛局部蒙版覆盖整图时的提交语义，确认采用 `admin 整图蒙版需显式确认`：整图覆盖不被静默当成普通局部蒙版，也不直接禁止；允许继续，但提交前必须额外确认“这会重绘整张图”。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:38 +0800
- 进度：grill-with-docs 继续收敛局部蒙版与目标图替换的行为边界，确认采用 `admin 蒙版目标替换即清空蒙版`：无论是在普通草稿、重试预填还是“以结果图继续改”链路里，只要管理员把蒙版目标图换成另一张图，旧蒙版就必须立即失效并清空，不做自动迁移或复用。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:36 +0800
- 进度：grill-with-docs 继续收敛前序结果复用的冻结边界，确认采用 `admin 前序结果复制为新任务输入快照`：前序生图结果被拿来继续改或做蒙版目标时，新任务创建阶段要复制出自己的输入快照文件，同时保留来源任务/结果关系，不能只引用旧结果文件；因此派生后的新任务语义独立于旧结果文件后续 TTL 清理。
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

## 2026-06-11 19:29 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的重试语义，确认采用 `admin 重试蒙版失效阻止提交`：蒙版文件失效，或其绑定的目标参考图缺失/被替换时，不允许系统自动降级成整图编辑继续提交；必须显式提示管理员重新选目标图并重画蒙版。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：文档更新，未执行构建或测试。

## 2026-06-11 19:29 +0800
- 进度：grill-with-docs 继续收敛局部蒙版的数据形状，确认采用 `admin 单次提交单蒙版目标`：单次提交可带多张普通参考图，但至多只允许一张参考图作为蒙版目标图，不支持同一次提交里维护多张目标图及多份蒙版。
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

## 2026-06-11 18:14 +0800
- 进度：grill-with-docs 继续收敛管理端图像生成任务化的异常终态恢复边界，确认采用 `admin 图像生成僵尸任务自动收口`：长时间未开始的 `queued` 和长时间无更新的 `running` 任务都要自动标记为 `failed` 并写入可读失败摘要，避免历史残留假活跃任务。同步把 `admin 运行中任务详情回显` 里的旧“请求 ID”表述统一收紧为 `task_id`。
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
- 进度：开始按已确认方案进入管理端图像生成下一阶段开发。首批实现边界锁定为“后端任务化最小闭环 + 前端 API 接缝预留”：新增服务端图像生成任务表与结果持久化、Asynq 任务入队与 worker 执行、管理端图像生成任务提交/列表/详情接口，暂不在本轮展开取消、结果 TTL 清理、复杂审计尾部或完整前端详情交互。本次仍需精确避开既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`migrations/*`、`internal/models/*`、`internal/repository/*`、`internal/queue/*`、`internal/handlers/*`、`admin-web/src/api/*`、`plan.md`。
- 验证：待执行定向 `go test`、前端 API 单测/构建、`git diff --check` 与乱码扫描。

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

## 2026-06-10 23:52 +0800
- 进度：确认管理端图像生成下一阶段的连续提交草稿分流语义，并将其收敛为 `admin 连续提交草稿分流`。规则是：在提交后输入快照默认视为新草稿的前提下，管理员连续发起多次生成时，工作台应显式提供“从上一稿继续改”和“清空开始新稿”两条草稿路径，而不是只留下隐式残留输入。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 16:32 +0800
- 进度：确认管理端图像生成下一阶段的导入标题预填语义，并将其收敛为 `admin 导入资产标题默认预填`。规则是：管理员把单张结果图导入媒体库时，系统应按稳定规则先预填一个可编辑的默认资产标题，例如“提示词摘要 + 时间”，而不是要求每次都从空白标题开始输入；这样批量处理导入时效率更高，同时又不把生成提示词原文硬绑定成不可修改的正式资产标题。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 15:53 +0800
- 进度：确认管理端图像生成下一阶段的已结束成功任务排序语义，并将其收敛为 `admin 成功任务导入待办优先`。规则是：在已结束任务中，仍处于待导入或部分已导入的成功任务，应排在已全部导入的成功任务前面；这样管理员会先看到仍待闭环处理的成功结果，再回看已经处理完的历史项。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:47 +0800
- 进度：确认管理端图像生成下一阶段的待导入结果默认聚焦语义，并将其收敛为 `admin 首个待导入结果默认聚焦`。规则是：如果一个成功任务的多张结果里仍有未导入项，任务详情默认应聚焦到第一张尚未导入媒体库的结果图；这样管理员打开详情后可以直接继续处理下一张待办结果，而不需要先手动查找。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 15:36 +0800
- 进度：确认管理端图像生成下一阶段的未导入成功任务可见性，并将其收敛为 `admin 图像生成待导入态`。规则是：如果一个成功任务的结果图一张都还没有导入媒体库，那么任务列表应给出明确的“待导入”标识；这样管理员能一眼分出尚未开始处理的成功任务，而不会和“部分已导入”或“已全部导入”的任务混淆。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 15:15 +0800
- 进度：确认管理端图像生成下一阶段的任务列表轻量筛选语义，并将其收敛为 `admin 图像生成任务轻量筛选`。规则是：服务端创作历史的任务列表支持轻量筛选，帮助管理员快速切到“仅看进行中”“仅看失败”或“仅看已导入过结果”等常用视图；目标是加快积压处理和失败排查，而不是演化成复杂搜索器。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 15:04 +0800
- 进度：确认管理端图像生成下一阶段的运行中任务详情展示语义，并将其收敛为 `admin 运行中任务详情回显`。规则是：当任务处于 `running` 且尚未产出结果图时，服务端创作历史的详情区仍展示输入快照、当前状态、开始时间、请求 ID 等可用任务信息，而不是只保留一个空白等待态。这样管理员跨浏览器回看运行中任务时，仍能理解任务正在做什么。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 15:00 +0800
- 进度：确认管理端图像生成下一阶段的任务列表展示语义，并将其收敛为 `admin 图像生成任务列表分区`。规则是：服务端创作历史的任务列表优先展示仍在进行中的 `queued`、`running` 任务；已结束的 `succeeded`、`failed`、`canceled` 任务放在其下方，并按创建时间倒序展示。这样管理员进入页面时先看到当前任务，再回看历史结果。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 11:27 +0800
- 进度：确认创建者失去管理员权限后的服务端创作历史处理方式，并将其收敛为 `admin 创作历史审计归属`。规则是：历史长期保留创建者标识作审计归属，即使创建者被降权、停用或删除也不自动改写；这些记录不自动转移给其他管理员，接管属于系统级人工处理。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
- 影响文件：`CONTEXT.md`、`plan.md`。
- 验证：`git diff --check -- CONTEXT.md plan.md` 通过；乱码扫描无输出。

## 2026-06-10 11:25 +0800
- 进度：确认服务端创作历史引用的服务器参考图在源资产被删除后的表现，并将其收敛为 `admin 已删除参考图回显`。规则是：历史记录继续保留参考图资产 ID 与轻量快照并明确提示“源图片已删除”；历史仍可查看，但不能直接用该已删除参考图重试，必须重新选择可用参考图。本次只纳入 `CONTEXT.md` 与 `plan.md`，不纳入既有 `admin-web/.env.development` 工作区改动。
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

## 2026-06-10 09:07 +0800
- 进度：将管理端生图默认模型统一调整为 `gpt-image-2`。`internal/config.Load` 的环境变量默认值已从 `gpt-image-1.5` 切换为 `gpt-image-2`，`internal/handlers/admin_image_generation.go` 的运行时兜底继续保持 `gpt-image-2`，并同步更新 `.env.example`、`CONTEXT.md` 与相关测试。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`internal/handlers/admin_image_generation_test.go`、`.env.example`、`CONTEXT.md`、`plan.md`。
- 验证：待执行 `go test ./internal/config ./internal/handlers -count=1`、必要的全量 Go 验证、`git diff --check` 与乱码扫描。

## 2026-06-09 23:24 +0800
- 进度：完成转码坏包容忍修复收尾验证。确认本次只影响 ffmpeg 转码参数、参数测试和长期文档；不纳入无关工作区改动 `admin-web/.env.development`。本机未安装 ffmpeg，已在部署机用 `/opt/homebrew/bin/ffmpeg` 确认 FFmpeg 8.1 支持并接受 `-max_error_rate 1.0`、`-fflags +discardcorrupt`、`-err_detect ignore_err` 参数。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：`go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -run 'TestBuildTranscodeVideoArgs|TestBuildTranscodeProgress|TestParseProgressValueToSeconds|TestBuildTranscodePlan|TestResolveTranscodePersistence|TestDecideVideoBitrate|TestChooseTranscodeOutputProfile' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`git diff --check -- CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；乱码扫描无输出。

## 2026-06-09 23:21 +0800
- 进度：完成转码坏包容忍参数实现。`buildTranscodeVideoArgs` 现在在输入文件前加入 `-max_error_rate 1.0`、`-fflags +discardcorrupt`、`-err_detect ignore_err`，用于跳过局部损坏 packet 并忽略可恢复解码错误；HEVC primary 与 AVC compat 参数测试同步锁定这些选项。`CONTEXT.md` 已追加转码坏包容忍策略。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：红灯阶段 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 因缺少 `-max_error_rate 1.0` 失败；实现后同命令通过。待执行相关 Go 定向测试、全量 Go 验证、`git diff --check`、乱码扫描。

## 2026-06-09 23:17 +0800
- 进度：开始修复转码压缩遇到源 MP4 内部坏 H.264/AAC packet 时失败的问题。报错中的 `Invalid NAL unit size`、AAC `Invalid data found when processing input` 表明这是源文件局部损坏/坏包场景，而不是上一轮的未知额外音轨。范围锁定为后端转码参数：丢弃损坏包、忽略可恢复解码错误，并允许 FFmpeg 在可生成输出时不要因错误率阈值非 0 退出；不改变编码器、码率/CRF、音轨选择、输出路径或重试模型。部署机 FFmpeg 8.1 已确认支持 `-max_error_rate`、`-fflags +discardcorrupt` 与 `-err_detect ignore_err`。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：待先补红灯测试，再执行 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1`、相关 Go 定向测试、必要全量 Go 验证、`git diff --check`、乱码扫描；本机未安装 ffmpeg，实际转码参数运行只在部署机用 FFmpeg 8.1 做参数存在性检查。

## 2026-06-09 22:55 +0800
- 进度：完成 iPhone MOV 多音轨转码失败修复。`buildTranscodeVideoArgs` 的音频映射从全部音频流 `0:a?` 收窄为首个可选音频流 `0:a:0?`，保留 AAC 主音轨并跳过 `apac` 等 ffmpeg 无法解码的额外音轨；HEVC primary 与 AVC compat 两个输出 profile 的参数测试已同步覆盖。`CONTEXT.md` 已沉淀转码音轨映射策略。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：红灯 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 先失败于仍生成 `-map 0:a?`；实现后 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -run 'TestBuildTranscodeVideoArgs|TestBuildTranscodeProgress|TestParseProgressValueToSeconds|TestBuildTranscodePlan|TestResolveTranscodePersistence' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；乱码扫描无输出。

## 2026-06-09 22:52 +0800
- 进度：开始修复 iPhone MOV 转码压缩因未知 `apac` 音频流失败的问题。报错显示当前 ffmpeg 参数映射了全部音频流，导致可用 AAC 主音轨之外的空间音频/未知音轨也被要求解码。本次范围锁定为后端转码参数生成：保留首个可选音频流，避免不可解码的额外音轨拖垮整次压缩；不改变视频编码器、码率/CRF、杜比视界安全播放策略或长期保留原文件策略。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：待先补红灯测试，再执行 `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgs' -count=1`、必要 Go 定向测试、`git diff --check`、乱码扫描。

## 2026-06-09 18:23 +0800
- 进度：完成管理端图像生成工作台第一阶段实现。工具箱新增“图像生成工作台”新标签页入口；新增 `/toolbox/image-workbench` 无 shell 管理员页面，提供三栏工作台、参考图上传/拖拽/粘贴、常用参数、同步生成、IndexedDB 本地历史、结果下载/复用为参考图/导入媒体库；后端新增管理员受限图像生成状态与同步代理接口，支持 OpenAI-compatible Images generate/edit、脱敏配置状态、资源限制、data URL 结果归一和轻量导入元数据；`.env.example` 增加图像生成配置键。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`.env.example`、`CONTEXT.md`、`plan.md`、`main.go`、`internal/config/config.go`、`internal/config/config_test.go`、`internal/handlers/router.go`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxImageWorkbench.vue`、`admin-web/src/views/imageWorkbench.db.js`、`admin-web/src/views/imageWorkbench.helpers.js`、`admin-web/src/views/imageWorkbench.helpers.spec.js`、`admin-web/src/views/toolboxPage.spec.js`
- 验证：`go test ./internal/config ./internal/handlers -run 'TestLoadIncludesImageGenerationConfig|TestRegisterIncludesAdminImageGenerationRoutes|TestAdminImageGenerationStatusIsRedacted|TestNormalizeAdminImageGenerationRequestRejectsTooManyImages|TestAdminImageGenerate' -count=1` 通过；`go test ./internal/... -count=1` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-09 18:22 +0800
- 进度：实现中按官方 GPT Image 接口校正后端代理细节：默认模型改为 `gpt-image-1.5`；官方 `gpt-image-*` 模型不发送 `response_format`，非 GPT Image 兼容模型才请求 `b64_json`；参考图编辑 multipart 文件字段使用 `image`。前端仍统一接收后端归一后的 data URL 结果。
- 影响文件：`internal/config/config.go`、`.env.example`、`internal/handlers/admin_image_generation.go`、`internal/handlers/admin_image_generation_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待复跑后端定向测试、管理端测试/构建与 `git diff --check`。

## 2026-06-09 17:59 +0800
- 进度：开始实现管理端图像生成工作台第一阶段。范围锁定为：工具箱新增图像工作台新标签页入口；新增无 shell 图像工作台页面；后端新增管理员受限同步图像生成代理、脱敏配置状态、资源限制和 data URL 结果归一；前端新增 IndexedDB 本地创作历史、参考图上传/粘贴/拖拽、常用参数面板、结果预览/下载/复用为参考图/导入媒体库。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `internal/config/config.go`、`main.go`、`internal/handlers/router.go`、新增后端图像生成代理 handler/test、`.env.example`、`admin-web/src/api/admin.js`、`admin-web/src/router/index.js`、`admin-web/src/views/Toolbox.vue`、新增图像工作台页面/helper/db/test、`CONTEXT.md`、`plan.md`。
- 验证：待补定向测试后执行后端定向测试、管理端定向测试、`cd admin-web && npm test`、`cd admin-web && npm run build`、必要的 Go 测试、`git diff --check`、乱码扫描。

## 2026-06-09 17:57 +0800
- 进度：确认管理端图像生成代理第一阶段对前端统一返回 base64 data URL 图片结果。后端调用上游时优先请求 `b64_json`；若上游只返回临时图片 URL，后端负责下载并转换为 data URL 后返回。前端历史、预览、下载和导入媒体库均基于 data URL；响应可携带每张图的 `mime`、可探测的宽高和可选 `revised_prompt`。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端上游响应解析、URL 下载转换、图片 MIME/尺寸探测、前端结果归一和测试。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:55 +0800
- 进度：确认管理端图像生成工作台第一阶段采用三栏工作台布局：左侧输入与参数，中央当前结果预览和操作，右侧本地历史画廊；移动端折叠为上下结构。无 shell 页面顶部只保留工具名、配置状态、返回工具箱、打开媒体库等必要操作；历史区默认展示最近任务缩略图，点击加载详情；中央结果支持下载、复用为参考图、导入媒体库。
- 影响文件：`plan.md`；后续若实现，预计涉及 `ToolboxImageWorkbench.vue` 页面结构、响应式 CSS、历史画廊、结果操作区和前端页面测试。
- 验证：探索阶段暂未执行；后续实现后按管理端范围执行测试、构建、截图/视觉检查、`git diff --check`、乱码扫描。

## 2026-06-09 17:46 +0800
- 进度：确认管理端图像生成同步代理的资源边界：后端做硬限制，前端做同口径预校验；单次最多 `n=4`、参考图最多 4 张、单张参考图最多 10 MiB、请求体总量最多 40 MiB、生成请求默认超时 180 秒、输入图片仅允许 `image/png`、`image/jpeg`、`image/webp`。超过限制直接返回业务错误，不转发上游。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端请求校验/超时、前端文件校验、错误文案和相关前后端测试。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:29 +0800
- 进度：确认管理端图像生成第一阶段使用同步后端代理：前端发起生成请求，后端在请求超时内直接调用上游 Images API 并返回图片结果；前端负责把任务状态、参考图和结果写入 IndexedDB。本阶段不引入 Asynq、后端任务表、后端生成历史、断点恢复、后台继续生成、跨浏览器同步或后端历史查询。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及同步代理接口、请求超时/大小限制、前端运行态任务记录、本地失败恢复和测试。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:27 +0800
- 进度：确认管理端图像生成代理需要提供脱敏配置状态接口，前端只接收 `enabled`、模型名、`base_url_configured`、`api_key_configured` 等状态，不返回真实 API Key 或完整上游地址；工作台打开时先查状态，未配置时禁用生成按钮并提示后端未配置图像生成服务；生成失败时前端展示可读错误摘要和请求时间/请求 ID，完整上游错误留在后端日志。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端配置状态接口、代理错误归一、日志字段、前端状态提示和失败态测试。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:26 +0800
- 进度：确认第一阶段“基础编辑生成”的产品语义为“参考图编辑”：有参考图时走参考图 + 提示词的整图编辑/变体生成，无参考图时走文本生图；不提供可视化蒙版、局部涂抹或局部区域承诺，UI 文案避免称为“局部编辑”或“蒙版编辑”。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及工作台模式判定、参考图上传/复用、Images API generate/edit 请求分流、UI 文案和测试。
- 验证：探索阶段暂未执行；后续实现后按管理端与必要后端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:21 +0800
- 进度：确认管理端图像生成工作台第一阶段参数面板只暴露常用创作参数：提示词、参考图、尺寸、质量、输出格式、生成张数；默认值为 `size=auto`、`quality=auto`、`output_format=png`、`n=1`。`output_compression` 仅在选择 `jpeg/webp` 时作为折叠高级项出现；`moderation` 固定 `auto` 不暴露；透明背景、提示词防改写、种子和自定义供应商参数暂缓。
- 影响文件：`plan.md`；后续若实现，预计涉及工作台表单状态、参数校验、生成请求 DTO、前端偏好持久化与单测。
- 验证：探索阶段暂未执行；后续实现后按管理端与必要后端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 17:03 +0800
- 进度：确认管理端图像创作历史使用 IndexedDB 持久化任务记录、原图和缩略图，参考项目的 `tasks/images/thumbnails` 拆分可作为实现参照；localStorage 只用于工作台轻量偏好，不承载图片 data URL 或大体积任务历史。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及管理端 IndexedDB helper、任务/图片/缩略图存储、历史画廊加载与清理、相关前端单测。
- 验证：探索阶段暂未执行；后续实现后按管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 16:56 +0800
- 进度：确认“导入媒体库”复用现有后端图片资产生命周期，并在导入成功后仅补写轻量创作来源元数据：来源、提示词、模型、尺寸、质量、输出格式、生成时间、本地任务 ID 等；不写入参考图 data URL、完整上游响应、API Key 或上游地址，避免把浏览器本地创作历史膨胀成后端创作档案。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及工作台导入逻辑、`uploadAdminImages` 复用、图片资产 metadata 补写、导入后的图片管理可追溯性测试。
- 验证：探索阶段暂未执行；后续实现后按管理端与必要后端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 16:39 +0800
- 进度：确认管理端图像生成代理第一阶段只支持 OpenAI-compatible Images API，用于文本生图与基础编辑生成；`model`、`base_url`、`api_key` 等上游配置放在 Go 后端环境配置中。暂不支持 Responses API、流式返回、Agent 模式或多供应商协议分支。
- 影响文件：`plan.md`；后续若实现，预计涉及后端配置项、管理员受限代理接口、前端图像工作台 API 封装、配置状态提示和相关测试。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 16:37 +0800
- 进度：确认管理端图像生成工作台第一阶段按“完整工作台主干”推进：包含文本生图、参考图上传/粘贴/拖拽、基础编辑生成、本地历史画廊、结果预览/下载/复用为参考图、显式导入媒体库、后端管理员代理与配置状态提示；暂缓可视化蒙版编辑器、Agent 模式、多/自定义供应商、流式局部图像、收藏/ZIP/复杂筛选等扩展能力。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及管理端工具箱入口、无 shell 图像工作台、前端本地历史存储、后端图像生成代理、媒体库导入衔接、相关前后端测试。
- 验证：探索阶段暂未执行；后续实现后按实际影响范围执行管理端测试/构建、必要的 Go 后端测试、`git diff --check`、乱码扫描。

## 2026-06-09 16:33 +0800
- 进度：确认管理端图像生成工作台的历史与资产边界：生成任务、参考图、中间图和结果图默认存浏览器本地创作历史，不自动进入后端图片库；只有管理员显式执行“导入媒体库”后，选定结果才成为后端图片资产并进入图片管理/图片合集/孤儿文件扫描等生命周期。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及前端 IndexedDB/本地历史、导入媒体库 API 复用或新增、工作台与图片管理的边界测试。
- 验证：探索阶段暂未执行；后续实现后按实际影响范围执行管理端测试/构建、必要的 Go 后端测试、`git diff --check`、乱码扫描。

## 2026-06-09 16:32 +0800
- 进度：确认管理端图像生成工作台的密钥与上游访问边界：真实 API Key 与上游地址放在 Go 后端环境配置中，前端通过管理员受限代理接口调用，不在浏览器保存或展示真实密钥，不直接复刻参考项目的纯前端 API Key 配置方式。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端配置、管理员受限图像生成代理接口、管理端图像工作台 API 封装与状态提示。
- 验证：探索阶段暂未执行；后续实现后按后端与管理端范围执行测试、构建、`git diff --check`、乱码扫描。

## 2026-06-09 16:24 +0800
- 进度：开始校准管理端工具箱新增“图像生成工作台”方案。已确认参考项目为 `references/gpt_image_playground`，其定位是完整图像工作台而非最小文本生图；用户确认本次目标按完整图像工作台方向推进。已在 `CONTEXT.md` 记录 `admin 图像生成工作台` 概念，并明确它不同于媒体库图片资产管理页面。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及管理端工具箱入口、无 shell 图像生成页、前端本地状态/存储、可能的后端代理或配置接口、相关测试与部署文档。
- 验证：探索阶段暂未执行；后续实现后按实际影响范围执行管理端测试/构建、必要的 Go 后端测试、`git diff --check`、乱码扫描。

## 2026-06-09 15:25 +0800
- 进度：完成管理端工具箱菜单合集与 ED2K 无 shell 新标签页实现。`/toolbox` 现在只展示工具菜单入口，ED2K 入口用 `router.resolve('/toolbox/ed2k').href` 生成带 SPA 基路径的链接并通过 `target="_blank"` 新标签页打开；新增 `/toolbox/ed2k` 管理员受限路由，页面不包后台 Layout，保留“返回工具箱”、ED2K 多行解析、非法行计数和本页会话“已点击”标记。命令面板 `ed2k` 搜索仍命中 `/toolbox`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolboxPage.spec.js`、`admin-web/src/router/index.js`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd admin-web && npm test -- src/views/toolboxPage.spec.js src/components/base/commandPalette.helpers.spec.js` 先失败于缺少 `ToolboxEd2k.vue`；实现后 `cd admin-web && npm test -- src/views/toolboxPage.spec.js src/views/toolbox.helpers.spec.js src/components/base/commandPalette.helpers.spec.js` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-09 15:19 +0800
- 进度：开始实现管理端工具箱菜单合集与 ED2K 无 shell 新标签页。范围锁定为：`/toolbox` 改为工具入口页，ED2K 入口通过新浏览器标签页打开 `/toolbox/ed2k`；`/toolbox/ed2k` 不包后台 Layout，但继续走管理员路由守卫；命令面板 `ed2k` 仍进入 `/toolbox`。
- 影响文件：预计涉及 `admin-web/src/views/Toolbox.vue`、新增 ED2K 工具页、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、新增工具箱页面源文测试、`CONTEXT.md`、`plan.md`。
- 验证：待先补红灯测试，再执行管理端定向测试、`cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-09 14:42 +0800
- 进度：继续校准管理端工具箱信息架构。已确认 `/toolbox` 应作为带后台 shell 的菜单合集，不直接承载 ED2K 表单；ED2K 功能从工具箱菜单按钮以新浏览器标签页打开，目标页为无 shell 的单一工具工作区，但仍走管理员登录与权限校验。命令面板搜索 `ed2k` 仍按推荐进入 `/toolbox`，不绕过工具箱直达功能页。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及 `admin-web/src/views/Toolbox.vue`、新增 ED2K 工具页、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.*`、相关单测。
- 验证：探索阶段暂未执行；后续实现后按管理端范围执行 `cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-08 21:02 +0800
- 进度：完成杜比视界安全播放第一阶段实现。后端在新视频转码流程中对原始源和最终播放文件做只读播放兼容探测，写入 `videos.metadata.playback_compat`，探测失败不让转码任务失败；TV App 根据 metadata 放行历史视频、阻断探测失败/结构不完整视频、阻断原始源为 Dolby Vision 但当前压缩输出非 Dolby Vision 的风险播放，并让剧集默认选集/自动连播跳过兼容阻断分集。未实现 Media3/DV 专用系统播放分支，未保留原始视频，未改变后端转码压缩参数。TV 端版本号升至 `0.1.84`。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/playback_compat.go`、`internal/services/playback_compat_test.go`、`internal/queue/tasks.go`、`internal/models/app.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/PlaybackCompatibilityPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesAutoplay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、TV 端相关测试、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services ./internal/queue ./internal/repository -run 'TestParsePlaybackCompatibility|TestBuildPlaybackCompatibility|TestBuildTranscodeTaskOptions|TestResolveTranscodePersistence|TestResolveTVEpisodeStillURL' -count=1` 通过；`go test ./pkg/ffmpeg ./internal/... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.PlaybackCompatibilityPolicyTest' --tests 'com.chee.videos.feature.tv.TvSeriesAutoplaySpecTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-08 20:29 +0800
- 进度：开始实现杜比视界安全播放第一阶段。范围锁定为：后端在新视频转码流程中做只读播放兼容探测并写入 `metadata.playback_compat`；TV App 根据 metadata 对新视频探测失败、原始 DV 风险但输出非 DV 的情况阻断提示。暂不实现 Media3/DV 专用系统播放分支，不保留原始视频，不改变后端转码压缩策略。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `pkg/ffmpeg/*`、`internal/services/transcode*`、`internal/queue/tasks*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、TV 端测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行后端定向测试、TV 端定向测试、`go test ./internal/... ./pkg/ffmpeg` 或相关范围、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug`、`git diff --check`、乱码扫描。

## 2026-06-08 20:27 +0800
- 进度：确认 `metadata.playback_compat` 的版本化判定规则：字段不存在表示历史视频或旧数据，TV App 放行；`version=1,status=ok` 时按探测结果决策；`version=1,status=probe_failed` 或结构不完整时按新视频探测失败/不完整处理，TV App 阻断并提示兼容性未确认。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端 metadata schema、TV 端解析 helper 和决策单测。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 20:24 +0800
- 进度：确认新视频播放兼容探测失败时的故障边界：不让转码任务整体失败，后台写入失败信息或日志；TV App 对这类新视频阻断自动播放并提示“该视频播放兼容性未确认，当前 TV 端暂不自动播放”。历史视频缺少 `playback_compat` 仍默认放行。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端探测失败 metadata、TV 端新旧视频缺字段判断和提示文案。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 20:21 +0800
- 进度：确认原始源为 Dolby Vision 风险、但最终实际可播放文件不是可原生 Dolby Vision 源时，TV App 不依赖设备 Dolby Vision 能力兜底，也不继续播放可能异色的转码输出；改为阻断并提示“该视频来源为杜比视界，当前压缩结果可能无法安全播放”。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端 `playback_compat` 中 source/output 探测结果与 TV 端播放决策。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 20:14 +0800
- 进度：确认存储空间优先级高于 Dolby Vision 原生播放能力：不允许为了 DV 播放长期保留原始上传文件，现有转码成功后删除 `original_path` 的存储策略保持不变。后续 TV App 能力判断只能针对当前实际可访问的播放源，不能假设已删除原始 DV 源仍可走系统播放器。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端只读探测 metadata 和 TV 端播放决策，不涉及保留原始视频副本。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 20:06 +0800
- 进度：确认播放兼容探测结果写入 `videos.metadata.playback_compat`，不新增数据库列。该字段记录源文件/播放文件的 codec、profile、HDR/Dolby Vision 风险和探测版本，供 TV App 播放入口决策使用。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端 metadata merge、TV 端 metadata 解析和播放分流测试。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 20:04 +0800
- 进度：确认 Dolby Vision 专用系统播放链路失败后不自动回退 LibVLC。失败页仅提供重试或返回，避免回到已知可能异色的链路。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及 TV 端 DV 播放分支错误处理和测试。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 19:59 +0800
- 进度：确认允许 Dolby Vision 风险源使用专用系统播放器/Media3 播放分支。普通 TV 长视频继续保持 LibVLC；只有 metadata 标记为 Dolby Vision 风险且设备平台声明支持对应能力时，才进入系统解码/HDR 输出链路，避免把本次改造扩大成全量播放器迁移。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及 TV 端 Media3 依赖恢复为 DV 专用、设备能力检测、播放入口分流测试和版本号。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 19:55 +0800
- 进度：修正杜比视界安全播放的决策顺序：检测到 Dolby Vision 风险源后，TV App 必须先判断设备平台和实际播放链路是否支持对应 profile / HDR 能力；能安全播放则原生播放，不能确认安全时才回退兼容源或阻断提示。不能把“片源是 Dolby Vision”直接等同于“禁止播放”。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及 TV 端设备能力检测 helper、播放入口决策测试。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 TV 端测试、构建、`git diff --check`、乱码扫描。

## 2026-06-08 19:53 +0800
- 进度：确认播放兼容源探测优先级：新视频以原始上传源作为 Dolby Vision 风险判断主依据，最终播放文件仅作辅助；因当前 worker 转码成功后会删除 `original_path`，后续实现必须在删除原始上传文件前完成只读探测并持久化 metadata。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及 worker 转码完成前的只读 probe、metadata merge、TV 端播放判断。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、`git diff --check`、乱码扫描。

## 2026-06-08 19:51 +0800
- 进度：确认播放兼容探测的存量边界：只自动探测新上传或新转码完成的视频；历史视频已确认没有 Dolby Vision 风险，不做全库回扫。TV App 后续遇到缺少探测字段的历史视频时必须按既有路径播放，不能默认阻断。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端新视频探测写 metadata、TV 端 metadata 缺字段默认放行。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、`git diff --check`、乱码扫描。

## 2026-06-08 19:50 +0800
- 进度：确认播放兼容探测结果需要持久化到视频 metadata。TV App 后续只读 API 返回的 metadata 来决定播放、回退或阻断，不在每次打开详情页时实时触发 ffprobe，避免外盘休眠、TCC 或探测耗时影响播放入口。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端 metadata 写入/读取、TV 端播放兼容判断。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、`git diff --check`、乱码扫描。

## 2026-06-08 19:45 +0800
- 进度：继续校准 TV App 杜比视界安全播放方案。已确认允许后端做只读探测：可通过 ffprobe 或已落库 metadata 判断 Dolby Vision/HDR 风险和兼容源可用性，但不得生成新视频、不得改变后端转码压缩参数、不得隐式触发重新压缩。
- 影响文件：`CONTEXT.md`、`plan.md`；后续若实现，预计涉及后端只读探测字段和 TV 端播放阻断逻辑。
- 验证：探索阶段暂未执行；后续实现后按受影响模块执行 Go/TV 端测试、`git diff --check`、乱码扫描。

## 2026-06-08 19:43 +0800
- 进度：开始梳理 TV App 杜比视界异色兼容方案。已确认目标从“保证原生点亮杜比视界”收窄为“杜比视界安全播放”：优先不异色；同时确认本次不得改变后端转码压缩策略、码率/CRF、编码器或输出产物规则。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `CONTEXT.md`、`plan.md`；若进入实现，可能涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、TV 端测试与版本号。
- 验证：待继续确认方案后决定；若实现 TV 端改动，执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、`git diff --check`、乱码扫描。

## 2026-06-07 14:44 +0800
- 进度：完成管理端工具箱菜单迁移的最终校验，准备提交。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-07 14:43 +0800
- 进度：完成管理端“工具箱”独立菜单与页面。ED2K 链接生成器已从系统设置页迁入工具箱，系统设置页只保留临时文件清理、孤儿文件扫描和日志查看等系统运维能力；命令面板支持通过“工具箱”、`gjx`、`ed2k` 搜索入口。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；待执行最终 `git diff --check` 与乱码扫描。

## 2026-06-07 14:36 +0800
- 进度：开始把管理端小功能收口到独立“工具箱”菜单。按最新要求新增工具箱页面与导航入口，把 ED2K 链接生成器从系统设置页迁出；系统设置页回归系统运维功能。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `admin-web/src/views/Toolbox.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/toolbox.helpers.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/components/base/commandPalette.helpers.js`、`admin-web/src/components/base/commandPalette.helpers.spec.js`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/Layout.spec.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-07 14:00 +0800
- 进度：完成管理端 ED2K 链接生成器。系统设置页新增多行输入框，逐行生成可点击 ED2K 链接，非法非空行会计数提示，链接点击后在当前页面会话内显示“已点击”；解析逻辑已沉到 helper 并补单测，`CONTEXT.md` 已记录该工具不提交后端、不持久化 ED2K 内容。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留现有 chunk size warning）；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-07 13:50 +0800
- 进度：开始实现管理端 ED2K 链接生成器。建议放在系统设置页作为本地工具：多行输入逐行解析 `ed2k://` 文本，生成多行可点击链接，点击后在当前页面会话中标记为已点击；不新增后端接口、不持久化 ED2K 内容。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-06-06 12:22 +0800
- 进度：完成 TV 电视剧详情页精修。标题区改为左对齐单行/双行剧名，不再渲染标题上方“剧集”；元信息行去掉 `18+` 年龄角标；剧情摘要固定至少四行；TV 端版本号升至 `0.1.83`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`adb -s 192.168.1.8:5555 install -r android-tv-app/tv-app/build/outputs/apk/debug/tv-app-armeabi-v7a-debug.apk` 成功；实机截图 `/tmp/tv-series-detail-refine-2.png` 确认左侧剧名左对齐、标题上方无“剧集”、元信息无 `18+`、剧情摘要显示四行；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-06 12:12 +0800
- 进度：开始精修 TV 电视剧详情页主体。按最新要求：剧名左对齐并去掉标题上方“剧集”，去掉元信息行里的 `18+` 年龄角标，剧情摘要至少显示四行。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 端定向单测、`:tv-app:testDebugUnitTest`、`:tv-app:assembleDebug`、ADB 实机截图、`git diff --check`、乱码扫描。

## 2026-06-06 12:07 +0800
- 进度：完成 TV 电视剧详情页参考图主体优化并按最新反馈去掉左侧/顶部导航栏。详情页现在保留全屏背景、独立返回按钮、小字号左侧信息区、暖金播放/焦点视觉、右侧剧照分集列表；TV 端版本号已升至 `0.1.82`。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过；`adb -s 192.168.1.8:5555 install -r android-tv-app/tv-app/build/outputs/apk/debug/tv-app-armeabi-v7a-debug.apk` 成功；实机截图 `/tmp/tv-after-load.png` 确认左侧/顶部导航栏已去掉，主体小字号与右侧分集列表正常；`git diff --check` 通过；乱码扫描无输出。

## 2026-06-06 12:01 +0800
- 进度：根据用户最新反馈调整范围：去掉电视剧详情页左侧竖向导航栏和顶部英文导航栏；保留已完成的小字号左侧详情信息、暖金焦点、右侧分集剧照列表与全屏背景参考图风格。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待重跑 TV 端定向单测、`:tv-app:assembleDebug`、ADB 实机截图、`git diff --check`、乱码扫描。

## 2026-06-06 11:46 +0800
- 进度：开始继续优化 TV App 电视剧详情页参考图还原。按用户最新要求，本次抛弃旧版“只还原主体、不还原导航”的约束，把左侧竖向 rail、顶部英文导航、左侧小字号标题信息和右侧分集列表作为同一张参考图处理；导航层仅做视觉与焦点锚点，不扩展真实全局路由。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 端定向单测、`:tv-app:assembleDebug`、ADB 安装与 `192.168.1.8` 实机截图、`git diff --check`、乱码扫描。

## 2026-06-05 23:58 +0800
- 进度：完成部署机稳定签名落地。部署机已导入 `Apple Development: 813745172@qq.com (CWLBNV4944)`，补齐 Apple WWDR G3 / Apple Root 证书链，`current/video-server` 已重签为固定 `Identifier=com.chee.videos.server`、`TeamIdentifier=96X2YQJC5G`，server / worker 均已重启且 `/healthz` 正常。实际 `post-receive` hook 已对齐为 Go build 后执行 `CODESIGN_ENV_FILE="$DEPLOY_ROOT/.env" scripts/sign-launchd-binary.sh`，不再 `source` 整个业务 `.env`，避免复杂密钥破坏 shell 解析。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`scripts/sign-launchd-binary.sh`、`scripts/rollback.sh`、`.env.example`、`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；部署机运维文件 `~/deploy/ai-video-server/.env`、`repo.git/hooks/post-receive`、`work/scripts/*` 已同步
- 验证：`bash -n scripts/sign-launchd-binary.sh scripts/rollback.sh scripts/migrate-apply.sh` 通过；`git diff --check` 通过；乱码扫描无输出；部署机 `codesign -dv --verbose=4 current/video-server` 显示 `Authority=Apple Development...`、`Identifier=com.chee.videos.server`、`TeamIdentifier=96X2YQJC5G`；server / worker `launchctl` 均 running；`curl http://127.0.0.1:8080/healthz` 返回 `{"status":"ok"}`

## 2026-06-05 23:44 +0800
- 进度：继续收口部署机外盘 TCC 重复授权问题。已确认部署机导入的 `Apple Development` 证书在补齐 Apple WWDR G3 与 Apple Root 链路后可用于 `codesign`，并开始把签名链路固化为稳定代码身份：`scripts/sign-launchd-binary.sh` 新增固定 `CODESIGN_IDENTIFIER`，支持 `CODESIGN_KEYCHAIN_PASSWORD_FILE` 读取专用 keychain 密码，避免版本化二进制文件名或明文密码继续影响部署。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：`scripts/sign-launchd-binary.sh`、`.env.example`、`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；部署机运维文件待同步
- 验证：待执行 `bash -n scripts/sign-launchd-binary.sh scripts/rollback.sh`、部署机脚本签名测试、远端 hook 语法检查、重签后 `/healthz`

## 2026-06-05 10:55 +0800
- 进度：管理端孤儿文件扫描前后端接入已完成并通过验证。系统页现已支持异步扫描、轮询最新状态、完成后自动弹出全量删除确认，以及删除后保留扫描快照；后端迁移、仓库、队列、handler 与 API 单测已补齐。无关工作区改动 `admin-web/.env.development` 仍未纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_orphan_file_scan_test.go`、`internal/repository/migrations_test.go`、`internal/repository/orphan_file_scan_repository_test.go`、`internal/queue/orphan_file_scan_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`
- 验证：`go test ./internal/... -count=1` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；乱码扫描无输出

## 2026-06-05 10:51 +0800
- 进度：完成管理端孤儿文件扫描的前后端接入补齐。后端把孤儿文件扫描单独挂到系统页可用的仓库/队列接口上，补了迁移与纯逻辑测试；管理端系统设置页新增“开始扫描 -> 自动轮询最新状态 -> 扫完自动确认全量删除”的交互，并把删除操作收口为全量删除，不再走逐项勾选。无关工作区改动 `admin-web/.env.development` 仍不纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/admin.go`、`internal/handlers/router.go`、`internal/handlers/admin_orphan_file_scan_test.go`、`internal/repository/migrations_test.go`、`internal/repository/orphan_file_scan_repository_test.go`、`internal/queue/orphan_file_scan_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/systemSettings.helpers.js`、`admin-web/src/views/systemSettings.helpers.spec.js`
- 验证：待执行 `go test ./internal/... -count=1`、`cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描

## 2026-06-04 20:20 +0800
- 进度：把 [[家用部署机]] 的运行入口从版本化二进制直接参与运行收口到固定真实路径 `current/video-server`；`docs/家用部署机.md`、`scripts/rollback.sh` 与 `CONTEXT.md` 已同步更新为固定执行路径语义，`binaries/video-server-<sha>.bin` 只保留作回滚备份，不再作为 launchd 直接目标。部署机 `post-receive` hook 也已从 symlink swap 改为复制到稳定入口再重启，避免后续 push 又把运行主体挂回旧 bin。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`scripts/rollback.sh`、`plan.md`
- 验证：`bash -n scripts/rollback.sh`、`git diff --check` 通过；后续 push 仍需复跑部署链路确认。
## 2026-06-05 09:59 +0800
- 进度：开始实现管理端孤儿文件扫描功能。已确认范围为 `STORAGE_ROOT` 下已知业务子树，数据库列与 `metadata` 里的本地路径统一算引用；扫描走异步任务，结果只保留最新一份，扫描完成后自动弹出全量删除确认，删除后不清空空目录。无关工作区改动 `admin-web/.env.development` 不纳入。
- 影响文件：预计涉及 `CONTEXT.md`、`plan.md`、`internal/queue/*`、`internal/repository/*`、`internal/models/*`、`internal/handlers/*`、`migrations/*`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`
- 验证：待执行定向单测、`cd admin-web && npm test`、`cd admin-web && npm run build`、`go test ./internal/... -count=1`、`git diff --check`、乱码扫描

## 2026-06-04 17:10 +0800
- 进度：已将固定执行路径授权方案提交 `a77c95d` 推送到 `deploy master`。远端 hook 确认 `HEAD is now at a77c95d 记录固定执行路径授权方案`，本次仅文档和计划变更，未触发 Go 重建或前端构建。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 不纳入
- 验证：`git push deploy master` 通过；远端输出 `RESTART_GO=0 REBUILD_FRONTEND=0`

## 2026-06-04 17:08 +0800
- 进度：完成“固定真实执行路径避免重复 TCC 授权弹窗”的方案记录。`CONTEXT.md` 新增 [[家用部署机固定执行路径契约]]，`docs/家用部署机.md` 在服务器-only 权限收口方案下补充：稳定签名后仍反复弹窗时，下一档应让 launchd 永远执行固定真实文件，sha binary 只保留为回滚备份。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 不纳入
- 验证：`git diff --check -- CONTEXT.md 'docs/家用部署机.md' plan.md` 通过；`rg -n $'\uFFFD' CONTEXT.md 'docs/家用部署机.md' plan.md` 无输出

## 2026-06-04 17:08 +0800
- 进度：开始记录“避免每次部署都手动点外盘授权”的长期方案。方案边界是文档沉淀，不直接改 hook：若稳定签名后仍每次弹 TCC 授权，下一步把 launchd 执行入口收紧为固定真实二进制路径，而不是每次指向新的 sha binary。
- 影响文件：预计涉及 `CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 不纳入
- 验证：待执行 Markdown diff 检查与乱码扫描，随后提交并推送 `deploy master`

## 2026-06-04 14:38 +0800
- 进度：已将本次文档方案作为 commit `5cc2bed` 推送到 `deploy master`。远端 hook 完成 `go build`、`apply migrations`、`launchctl kickstart`，最终回执 `/healthz OK — deploy succeeded`。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 未纳入
- 验证：`git push deploy master` 通过；远端输出确认 `HEAD is now at 5cc2bed 记录服务器权限收口方案`

## 2026-06-04 14:37 +0800
- 进度：完成本次文档补充的收尾校验。`CONTEXT.md` 与 `docs/家用部署机.md` 的“服务器-only 权限收口”说明已落盘，`plan.md` 顶部也已记录本次决策与范围。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 未纳入
- 验证：`git diff --check -- CONTEXT.md 'docs/家用部署机.md' plan.md` 通过；`rg -n $'\uFFFD' CONTEXT.md 'docs/家用部署机.md' plan.md` 无输出

## 2026-06-04 14:37 +0800
- 进度：把“服务器-only 权限收口方案”补进仓库长期文档。`CONTEXT.md` 新增 `服务器-only 权限收口` 术语，`docs/家用部署机.md` 增补只做服务器时的三档处理顺序：能迁就迁、不能迁再授、不要找 SSH 绕过，用来明确部署机只承担服务器角色时如何处理 `/Volumes/large` 相关 TCC 约束。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；无关工作区改动 `admin-web/.env.development` 未纳入
- 验证：待执行 `git diff --check -- CONTEXT.md 'docs/家用部署机.md' plan.md` 与乱码扫描

## 2026-06-04 12:45 +0800
- 进度：完成全媒体外盘访问收口。视频源、视频字幕、演员头像、图片视图、视频抓帧与图片上传后的外盘探测都改成限时打开/限时 ffmpeg 处理；`CONTEXT.md` 也同步补了“媒体资源限时打开”和“媒体生成限时”的长期约定。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`CONTEXT.md`、`plan.md`、`internal/handlers/local_image_file.go`、`internal/handlers/admin_image.go`、`internal/handlers/app_image_collection.go`、`internal/handlers/video_source.go`、`internal/handlers/video_subtitle.go`、`internal/handlers/actor_avatar.go`、`internal/handlers/admin_video_thumbnail.go`、`internal/services/image.go`、`internal/handlers/local_image_file_test.go`
- 验证：`go test ./internal/handlers ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md internal/handlers/local_image_file.go internal/handlers/admin_image.go internal/handlers/app_image_collection.go internal/handlers/local_image_file_test.go internal/handlers/video_source.go internal/handlers/video_subtitle.go internal/handlers/actor_avatar.go internal/handlers/admin_video_thumbnail.go internal/services/image.go` 通过；乱码扫描无输出。

## 2026-06-04 12:44 +0800
- 进度：把“外盘访问卡住”的修复范围从图片预览扩大到所有媒体路径。`admin/images/:id/view`、`images/:id/view`、`videos/:id/source`、`videos/:id/subtitle`、`actors/:id/avatar`、视频抓帧以及图片上传后的外盘探测/生成都改成限时打开或限时 ffmpeg 处理，避免 launchd 进程对 `/Volumes/large` 失去授权后把请求挂成 pending；同步把媒体访问/TCC 约定写回 `CONTEXT.md`。
- 影响文件：`internal/handlers/local_image_file.go`、`internal/handlers/admin_image.go`、`internal/handlers/app_image_collection.go`、`internal/handlers/video_source.go`、`internal/handlers/video_subtitle.go`、`internal/handlers/actor_avatar.go`、`internal/handlers/admin_video_thumbnail.go`、`internal/services/image.go`、`internal/handlers/local_image_file_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestOpenLocalImageFileWith|TestTryServeLocalImagePath' -count=1` 通过；`go test ./internal/handlers ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check` 通过；乱码扫描待执行。

## 2026-06-04 12:29 +0800
- 进度：完成图片预览链路收口。`AdminImageView` 与 `AppImageView` 现在统一走 `openLocalImageFile` / `serveOpenedLocalImage`，不再直接 `c.File` 读外盘路径；新增 `tryServeLocalImagePath` 复用本地文件限时打开逻辑，并补了 helper 回归测试，避免预览 blob 请求在外盘或 TCC 异常时长期 pending。
- 影响文件：`internal/handlers/local_image_file.go`、`internal/handlers/admin_image.go`、`internal/handlers/app_image_collection.go`、`internal/handlers/local_image_file_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestOpenLocalImageFileWith|TestTryServeLocalImagePath' -count=1` 通过；`go test ./internal/handlers -count=1` 通过；`git diff --check -- CONTEXT.md plan.md internal/handlers/local_image_file.go internal/handlers/admin_image.go internal/handlers/app_image_collection.go internal/handlers/local_image_file_test.go` 通过；乱码扫描无输出。

## 2026-06-04 12:27 +0800
- 进度：继续处理图片预览长期 pending。已确认 `admin-web` 的图片管理页会直接请求 `/admin/images/:id/view` 取 blob，而后端该路由与 `AppImageView` 仍在用 `c.File` 直读存储文件；计划把这两条视图链路改成和剧照/缩略图一致的限时打开再返回，避免外盘或 TCC 问题把页面卡成 loading。
- 影响文件：预计涉及 `internal/handlers/local_image_file.go`、`internal/handlers/admin_image.go`、`internal/handlers/app_image_collection.go`、`internal/handlers/*_test.go`、`plan.md`
- 验证：待执行定向单测、`go test ./internal/handlers -run 'Test.*Image.*View|TestOpenLocalImageFileWith' -count=1`、`git diff --check`、乱码扫描。

## 2026-06-04 12:17 +0800
- 进度：继续排查图片上传 pending。通过采样确认 server 现场有线程阻塞在 `open`，同时部署机当前 `UPLOAD_TEMP_DIR` 曾被配置到 `/Volumes/large/videos/tmp/uploads`；已把该值改回 `./tmp/uploads` 并重启 server。已同步 `CONTEXT.md` 和 `docs/家用部署机.md`，明确上传暂存目录必须留在本机磁盘，不能放在外盘上。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；远端部署配置 `~/deploy/ai-video-server/.env`
- 验证：远端 `.env` 中 `UPLOAD_TEMP_DIR=./tmp/uploads` 已生效；`launchctl kickstart` 后 `/healthz` 正常，server 重新启动；待用户重试图片上传确认 pending 是否消失。

## 2026-06-04 11:08 +0800
- 进度：已将部署机实际 `~/deploy/ai-video-server/repo.git/hooks/post-receive` 的前端 dist 更新块同步为增量保留 assets 版本，避免仓库文档已更新但运行中的 hook 仍整目录删除旧 hash 资源。远端 hook 保持可执行权限。
- 影响文件：`plan.md`；远端运维文件 `~/deploy/ai-video-server/repo.git/hooks/post-receive` 已手动对齐
- 验证：远端 `bash -n ~/deploy/ai-video-server/repo.git/hooks/post-receive` 通过；`grep -n "rsync\\|admin-web-dist\\|frontend updated" ~/deploy/ai-video-server/repo.git/hooks/post-receive` 确认包含增量同步块。

## 2026-06-04 11:06 +0800
- 进度：完成管理端入口 HEAD 探测修复。`/admin` 与 `/admin/` 现在同时支持 GET/HEAD 并返回同一套 `no-store` 入口响应，避免部署后 `curl -I` 或缓存探测误报 404。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- internal/handlers/router.go internal/handlers/admin_static_test.go plan.md` 通过；`rg -n $'\uFFFD' internal/handlers/router.go internal/handlers/admin_static_test.go plan.md` 无输出。

## 2026-06-04 11:05 +0800
- 进度：部署后用 `curl -I` 发现 `/admin` 的 HEAD 探测返回 404；已补回归测试确认红灯。浏览器 GET 不受影响，但 HEAD 误判会影响运维探测与缓存校验。
- 影响文件：`internal/handlers/admin_static_test.go`、`plan.md`
- 验证：红灯 `go test ./internal/handlers -run 'TestMountAdminStaticServesIndexFromGivenDir' -count=1` 失败于 `HEAD /admin = 404, want 200`。

## 2026-06-04 11:02 +0800
- 进度：完成“更新后无法上传”的静态资源断链修复与收口。结论：线上 09:07 只更新了 admin-web dist，10:42 日志出现旧 hash CSS 404，说明打开中的旧页面在前端发布后可能缺旧资源；本次修复让入口 HTML 不缓存、存在的 hash asset 长缓存、缺失旧 asset 不长缓存，并要求部署 hook 保留一段旧 hash asset。无关工作区改动 `admin-web/.env.development` 未纳入。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 通过；`go test ./internal/handlers -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md docs/家用部署机.md internal/handlers/router.go internal/handlers/admin_static_test.go` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md docs/家用部署机.md internal/handlers/router.go internal/handlers/admin_static_test.go` 无输出。

## 2026-06-04 10:58 +0800
- 进度：完成管理端静态资源发布窗口修复。`mountAdminStatic` 改为 `/admin` HTML 返回 `no-store`、存在的 hashed asset 返回长期缓存、缺失旧 asset 返回 404 且不长缓存；部署文档的 frontend hook 从整目录删除替换改为先增量同步 assets、再同步入口文件，并保留 7 天旧 hash asset；`CONTEXT.md` 新增 `admin 静态资源发布窗口` 术语。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 通过；待执行更大范围验证、diff 检查和乱码扫描。

## 2026-06-04 10:58 +0800
- 进度：已补管理端静态资源缓存回归测试并确认红灯。`TestMountAdminStaticServesIndexFromGivenDir` 现在要求 `/admin/` 返回 `Cache-Control: no-store`、存在的 `/admin/assets/*` 返回长缓存；`TestMountAdminStaticMissingAssetIsNotLongCached` 要求缺失旧 hash asset 404 且不长缓存。当前失败于 `adminIndexCacheControl` / `adminAssetCacheControl` / `adminMissingAssetCacheControl` 未实现。
- 影响文件：`internal/handlers/admin_static_test.go`、`plan.md`
- 验证：红灯 `go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 失败于上述未定义常量。

## 2026-06-04 10:58 +0800
- 进度：开始排查“更新后无法上传”。已确认部署机 `/healthz` 正常，当前 Go 二进制仍为 `0408391`，09:07 只更新过 admin-web 前端 dist；10:42 server log 出现旧 hashed CSS `/admin/assets/index-C0V8twHW.css` 404，且图片上传流程只打到 `/api/v1/admin/images/check`，未继续进入 `/api/v1/admin/images/upload`。计划修正管理端静态资源缓存与部署 hook 的 dist 替换方式，避免前端更新后打开中的旧页面断链。
- 影响文件：预计涉及 `internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：已跑基线 `go test ./internal/handlers -run 'TestMountAdminStatic' -count=1` 通过；待执行定向测试、Go 相关测试、Markdown/乱码检查。

## 2026-06-04 10:27 +0800
- 进度：完成稳定签名 helper 的 keychain/identity 预检收口，并补完相关文档说明；当前仓库层面已把“必须是 Apple-issued identity + 本地一次性 trust 授权”的要求写清楚。`scripts/sign-launchd-binary.sh`、`scripts/rollback.sh`、`.env.example`、`docs/家用部署机.md`、`CONTEXT.md` 已对齐。
- 影响文件：`scripts/sign-launchd-binary.sh`、`scripts/rollback.sh`、`.env.example`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：`bash -n scripts/sign-launchd-binary.sh scripts/rollback.sh scripts/migrate-apply.sh` 通过；`git diff --check -- .env.example CONTEXT.md docs/家用部署机.md plan.md scripts/sign-launchd-binary.sh scripts/rollback.sh` 通过；`rg -n $'\uFFFD' .env.example CONTEXT.md docs/家用部署机.md plan.md scripts/sign-launchd-binary.sh scripts/rollback.sh` 无输出。

## 2026-06-04 10:25 +0800
- 进度：补强 launchd 稳定签名入口的 keychain 支持与 identity 预检，并把“必须是 Apple-issued identity + 本地一次性 trust 授权”写回仓库上下文。`scripts/sign-launchd-binary.sh` 现在支持可选 `CODESIGN_KEYCHAIN` / `CODESIGN_KEYCHAIN_PASSWORD`，签名前会先解锁 keychain 并验证 `security find-identity -v -p codesigning` 里真的存在目标 identity；`docs/家用部署机.md`、`.env.example` 和 `CONTEXT.md` 同步补充了签名身份落地约束。
- 影响文件：`scripts/sign-launchd-binary.sh`、`.env.example`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `bash -n scripts/sign-launchd-binary.sh scripts/rollback.sh scripts/migrate-apply.sh`、`git diff --check`、乱码扫描。

## 2026-06-04 10:06 +0800
- 进度：完成管理端 destructive API 的无超时调整，并把部署机上确认到的 TCC 事实补进仓库上下文。`admin-web/src/api/admin.js` 里视频/图片/字幕/合集/TV 相关删除请求和批量删除请求统一显式 `timeout: 0`；`admin-web/src/api/admin.spec.js` 已回归这些调用；`CONTEXT.md` 新增 `TCC` 取证契约，记录 launchd binary 仍可能在 `tccd` 下被拒。
- 影响文件：`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src/api/admin.js admin-web/src/api/admin.spec.js` 无输出。

## 2026-06-04 10:03 +0800
- 进度：先把管理端删除型请求的客户端超时放开，避免外盘慢 I/O 直接在前端变成 30 秒 timeout。`admin-web/src/api/admin.js` 里 `deleteAdminVideo`、`deleteAdminImage`、`deleteAdminVideoSubtitle`、`deleteAdminCollection`、`deleteAdminImageCollection` 与 `batchDeleteAdminVideos` 统一改为 `timeout: 0`，`admin-web/src/api/admin.spec.js` 补了回归测试。
- 影响文件：`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描；远端 TCC 证据待同步到 `CONTEXT.md`。

## 2026-06-04 09:06 +0800
- 进度：修正图片管理页上传按钮长时间 loading 的状态绑定。`admin-web/src/views/ImageManage.vue` 里 `uploading` 现在只覆盖真正的上传请求，上传结果落地并清空文件后立即置回 `false`，后续 `await load()` 只刷新列表，不再把“开始上传”按钮绑在列表刷新上。
- 影响文件：`admin-web/src/views/ImageManage.vue`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅既有 chunk size warning）；`git diff --check -- admin-web/src/views/ImageManage.vue` 通过。

## 2026-06-04 01:36 +0800
- 进度：已将图片上传 `ffmpeg` 缺失降级修复推送到部署机远端 `deploy`，部署机 hook 完成 `go build`、迁移和重启探活，`/healthz` 返回 OK。当前线上接收到的提交为 `0408391`。
- 影响文件：本次推送对应提交 `0408391`（文件同上条记录）
- 验证：远端 hook 日志显示 `go build` 成功、`apply migrations` 成功、`/healthz OK — deploy succeeded`；部署机侧 `launchctl kickstart` 输出服务名不存在，但不影响最终健康检查结果。

## 2026-06-04 01:35 +0800
- 进度：完成图片上传 `ffmpeg` PATH 缺失回退修复收口。最终实现只改 `pkg/ffmpeg.ConvertToWebP` 的缺失二进制分支，让图片上传在 launchd / 受限 PATH 环境里也能改走 `cwebp`，两者都不可用时仍返回 `ErrWebPEncodingUnavailable` 走保留原图兜底。`CONTEXT.md` 已同步“ffmpeg 不在 PATH 也算 WebP 编码链路不可用”的长期契约。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg -run TestConvertToWebPFallsBackWhenFFmpegMissing -count=1` 通过；`go test ./pkg/ffmpeg ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`git diff --check -- CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 无输出；工作区既有改动 `admin-web/.env.development` 未纳入本次修复。

## 2026-06-04 01:34 +0800
- 进度：修复图片上传在 launchd / 受限 PATH 环境下找不到 `ffmpeg` 可执行文件时的回退链路。`pkg/ffmpeg/ConvertToWebP` 现在把 `exec.ErrNotFound` 视为 WebP 编码能力不可用，直接降级到 `cwebp`；若 `cwebp` 也不可用，仍返回 `ErrWebPEncodingUnavailable` 让图片上传保留原始 JPEG/PNG。`CONTEXT.md` 补充了“ffmpeg 不在 PATH 中”也属于图片上传 WebP 编码不可用契约。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/ffmpeg -run TestConvertToWebPFallsBackWhenFFmpegMissing -count=1` 通过；`git diff --check -- pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go` 通过；待执行全量相关包验证、乱码扫描和提交。

## 2026-06-04 01:20 +0800
- 进度：完成 launchd 二进制稳定签名链路落地。新增 `scripts/sign-launchd-binary.sh`，`scripts/rollback.sh` 现在会在切 symlink 前先按同一 `CODESIGN_IDENTITY` 重签目标二进制；`docs/家用部署机.md` 的 post-receive hook 也改成 build → codesign → migrate → restart，`CONTEXT.md` / ADR-0005 / ADR-0007 补齐稳定签名契约。
- 影响文件：`scripts/sign-launchd-binary.sh`、`scripts/rollback.sh`、`docs/家用部署机.md`、`CONTEXT.md`、`docs/adr/0005-home-deployment-architecture.md`、`docs/adr/0007-launchd-binary-stable-signature.md`、`plan.md`
- 验证：`bash -n scripts/sign-launchd-binary.sh scripts/rollback.sh scripts/migrate-apply.sh` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' plan.md CONTEXT.md docs/家用部署机.md docs/adr/0005-home-deployment-architecture.md docs/adr/0007-launchd-binary-stable-signature.md scripts/sign-launchd-binary.sh scripts/rollback.sh` 无输出。

## 2026-06-04 01:17 +0800
- 进度：按推荐把 launchd 二进制从 ad-hoc 改成稳定签名链路，目标是让 TCC 对 `/Volumes/large` 的外盘授权不再跟随每次 push 的 cdhash 漂移。准备新增 `scripts/sign-launchd-binary.sh`，让 post-receive hook 与 `scripts/rollback.sh` 都在切 symlink 之前先 codesign 同一份二进制。
- 影响文件：`scripts/sign-launchd-binary.sh`、`scripts/rollback.sh`、`docs/家用部署机.md`、`CONTEXT.md`、`docs/adr/0007-launchd-binary-stable-signature.md`、`plan.md`
- 验证：待执行 `bash -n`、`git diff --check`、乱码扫描。

## 2026-06-01 22:22 +0800
- 进度：按远程 TV 截图继续修正电视剧详情页参考图还原。确认 1920x1080 / density 320 设备下 Compose 逻辑宽度为 960dp，之前把参考图像素级宽度当 dp 导致右侧剧集面板吞掉左侧信息区；已改为参考图专用主体布局，左侧信息区固定宽度、中文标题保持横排、右侧剧集面板按 TV density 收窄，并移除旧版共享青蓝 `tvFocusableGlow`，改用暖金边框与暗玻璃焦点态。已安装到远程设备 `192.168.1.8:5555` 并截图确认左侧电视剧信息完整显示。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`adb install -r .../tv-app-armeabi-v7a-debug.apk` 成功；截图 `/tmp/tv-detail-reference-focus.png` 确认左侧信息未被吞、焦点视觉不再是旧青蓝 glow。无关 `admin-web/.env.development` 不纳入本次提交。

## 2026-06-01 21:27 +0800
- 进度：根据用户澄清调整还原范围：不还原参考图左侧竖向导航栏和顶部全局导航，只针对电视剧详情主体继续 1:1 对齐，优先修正左侧信息区。已撤掉导航还原代码，把左侧信息区改为参考图式“小顶标 + 超大字距标题、点分元信息 + 18+、星级 + 分数 + IMDb、播放/我的片单双按钮、演员行”；右侧剧集列表同步收窄为参考图式卡片比例与季选择下拉外观。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest` 通过；ADB 已连接 `192.168.1.8:5555`，待安装后截图对比。

## 2026-06-01 21:13 +0800
- 进度：开始继续按参考图还原 TV 电视剧详情页。已确认目标页面为 `android-tv-app` 的 `TvSeriesDetailScreen`，本次以参考图为主导，允许覆盖此前 TV 排版/圆角等视觉 token 约束，但保留中文界面、遥控器可聚焦、TV 端版本递增和验证要求。下一步对齐参考图的 3 区结构：左侧导航栏 + 顶部横向导航、左侧大标题/元信息/评分/主按钮/演员、右侧剧集列表。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、对应源文测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行 TV 端相关单测与构建。

## 2026-06-01 19:44 +0800
- 进度：完成 30 秒超时修复部署。`70d0572` 先部署后，因 5 个候选详情补全在部署机代理下仍耗时约 37 秒，继续以 `1b6d00d` 将补全上限降到 3；部署机 `current/video-server` 已指向 `video-server-1b6d00d096.bin`，手动只重启 server 为 pid 78477，worker pid 76490 与 ffmpeg pid 76553/76554 保持运行，未丢失当前转码进度。实测 `POST /api/v1/admin/scrape/preview`，payload `{"type":"tv","title":"The Lead","year":0}` 返回 HTTP 200、20 个候选、耗时约 1.86 秒。临时跳过 worker kickstart 的部署 hook 已恢复。
- 影响文件：`plan.md`；代码提交为 `70d0572`、`1b6d00d`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：部署机 `/healthz` 通过；`The Lead` 电视剧预览接口实测通过；部署机 hook 推 GitHub mirror 失败为非阻塞，需从本机补推 `origin/master`。

## 2026-06-01 19:41 +0800
- 进度：部署 `70d0572` 后实测 `The Lead` 电视剧预览返回 20 个候选但仍耗时约 37 秒，虽然前端已不会再被 axios 30 秒超时截断，但首轮后端响应仍偏慢。进一步将电影/电视剧预览详情补全上限从 5 个候选收敛到 3 个候选，减少 TMDB 详情/兜底请求数量；确认刮削仍在确认阶段重新拉详情，不依赖预览阶段完整 metadata。
- 影响文件：`internal/services/scraper.go`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行 scraper 定向测试、部署机安全推送与 `The Lead` 预览接口实测。

## 2026-06-01 19:36 +0800
- 进度：完成刮削预览 30 秒超时修复并准备安全部署。后端将电影/电视剧预览详情补全限制为靠前 5 个 TMDB 候选，靠后候选直接使用搜索结果字段；前端 `scrapePreview` 对外部刮削请求关闭 axios 30 秒通用超时。部署机当前仍有转码任务 6158/6159 在约 40% 运行，默认 hook 会硬重启 worker，因此本次部署需只重启 server，保留现有 worker/ffmpeg，待转码完成后 worker 再切到新二进制。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`go test ./internal/services -run 'TestPreviewTVUsesChineseLanguageAndFallback|TestPreviewTVLimitsDetailRequests|TestPreviewMovieBypassCacheFetchesFreshCandidates' -count=1` 通过；`go test ./internal/services -count=1` 通过；`go test ./... -count=1` 通过；`cd admin-web && npm test -- src/api/admin.spec.js` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 与乱码扫描通过。

## 2026-06-01 19:27 +0800
- 进度：开始修复管理端刮削预览 `timeout of 30000ms exceeded`。已确认全局 axios 超时为 30 秒，而部署机 `The Lead` TV 预览在代理恢复后仍耗时约 76 秒，根因是后端对 TMDB 搜索返回的 20 个候选逐个串行拉详情/兜底详情。下一步同时收敛后端候选详情补全数量，并让管理端刮削预览请求不使用 30 秒通用超时。
- 影响文件：预计涉及 `internal/services/scraper.go`、`internal/services/scraper_test.go`、`admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行后端 scraper 定向测试、admin-web API 单测、`git diff --check` 与乱码扫描。

## 2026-06-01 19:22 +0800
- 进度：完成 TMDB TLS 超时的部署机临时恢复。已在部署机 `.env` 增加显式 `HTTP_PROXY`、`HTTPS_PROXY`、`ALL_PROXY` 与 `NO_PROXY`，并重启手动 server 为 pid 77089；`/healthz` 正常。通过管理端 `POST /api/v1/admin/scrape/preview` 验证 `The Lead` 电视剧预览返回 20 个候选，不再报 TLS handshake timeout。当前 worker pid 76490 正在转码 job 6158/6159，已发送 `TSTP` 停止接新任务但保留当前 ffmpeg；后台脚本 `/tmp/restart-worker-after-transcode.sh` 会等待两个 ffmpeg 完成后重启 worker，使新 worker 读取代理环境。
- 影响文件：`CONTEXT.md`、`plan.md`；部署机临时状态：server pid 77089 已带代理环境运行，worker pid 76490 仍在完成当前转码且停止接新任务，等待脚本 pid 77239 负责后续自动重启；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`curl -x http://127.0.0.1:15732 https://api.themoviedb.org/3/configuration` 约 0.7 秒返回 401；`GET /healthz` 返回 200；管理端 TV 预览 `The Lead` 返回 `code=0`、候选数 20；数据库 job 6158/6159 进度约 22.7% 且继续增长。

## 2026-06-01 19:22 +0800
- 进度：开始排查 TMDB `TLS handshake timeout`。日志显示 5 月 26/28/29 已有 TMDB API 直连异常，部署机当前 DNS 将 `api.themoviedb.org` 解析到 `44.0.0.15`，直连 TLS 超时；本机 macOS 系统代理为 `127.0.0.1:15732`，但 Go 进程没有 `HTTP_PROXY/HTTPS_PROXY` 环境变量，不会自动继承系统代理。已验证显式代理访问 TMDB API 可在约 0.7 秒返回 401（测试未带 API key，证明 TLS 通路可用）。
- 影响文件：预计涉及 `CONTEXT.md`、`plan.md`；部署机运行态涉及 `.env` 代理变量与手动 server/worker 重启安排；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行部署机 `.env` 代理配置、server 重启、管理端 TMDB 预览接口验证、worker 当前转码完成后的代理重启验证、`git diff --check` 与乱码扫描。

## 2026-06-01 19:06 +0800
- 进度：完成压缩进度 0% 的部署机恢复。已 `bootout` launchd 版 worker，保留手动 server；将卡住的 job 6156/6157 标记为 failed，并把对应视频恢复到 `uploaded` 后通过管理端重转码接口重新入队。手动 worker pid 76490 已启动，ffmpeg 子进程开始转码两集，新 job 6158/6159 已写入源时长、已处理秒数和百分比，管理端任务接口返回进度 0.68% 后继续增长到 1.18%。结论：worker 也缺少或失去 launchd 外盘媒体访问权限，导致卡在输出目录创建阶段，进度并未进入可更新状态。
- 影响文件：`CONTEXT.md`、`plan.md`；部署机临时状态：`com.aivideo.worker` launchd job 已 bootout，手动 worker pid 76490 正在处理队列；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`ps` 显示 worker pid 76490 与两个 ffmpeg 子进程；数据库 `transcoding_jobs` job 6158/6159 从 `progress_percent=0.36` 增至 `1.18`；`GET /api/v1/admin/tasks?page=1&page_size=4` 返回 running job 6158/6159 的 `progress_percent=0.68`。

## 2026-06-01 19:03 +0800
- 进度：开始排查视频压缩进度一直为 0%。已确认管理端任务页展示的是后台转码任务持久化进度；部署机最新两个转码 job 已进入 `running`，但源时长、已处理秒数与进度均为空。worker 仍由 launchd 启动，进程采样显示没有 ffmpeg/ffprobe 子进程，线程卡在 `/Volumes/large` 输出目录创建，初步结论是 worker 与上一轮 server 同类，缺少或失去外盘媒体访问权限。下一步停止 launchd worker、用 SSH/Terminal 权限手动启动 worker，并对已卡住任务做最小人工恢复。
- 影响文件：预计涉及 `CONTEXT.md`、`plan.md`；部署机运行态只调整 worker；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行部署机 worker 重启验证、数据库转码进度验证、任务接口验证、`git diff --check` 与乱码扫描。

## 2026-06-01 17:13 +0800
- 进度：完成图片和视频资源无法访问的根因校准。使用管理员 token 验证：管理端详情与 play-url 均 200，视频 `/source` Range 请求 15 秒无字节返回；封面接口 503。部署机 shell 直接读取同一 `/Volumes/large/.../thumb.jpg` 和 `video-hevc.mp4` 正常，手动启动同一二进制到 18081 也能 200 返回封面；但 launchd 版 server 采样显示线程卡在 `open`，TCC 日志显示当前 hash 二进制请求 `kTCCServiceSystemPolicyAllFiles` 被拒。结论：资源路径和文件都正常，问题是 launchd 启动的当前二进制缺少 macOS TCC 媒体访问权限。已临时 `bootout` launchd server，并用 SSH/Terminal 权限启动同一二进制守住 8080，验证 `/healthz` 200、封面 200、视频 Range 206；后续需做稳定签名或系统授权的长期修复。
- 影响文件：`CONTEXT.md`、`docs/家用部署机.md`、`plan.md`；部署机临时状态：`com.aivideo.server` launchd job 已 bootout，手动 server pid 74522 正在提供 :8080，worker launchd 未变；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`curl /healthz` 返回 200；`curl /api/v1/videos/2bb55884-3647-4c4a-81f8-c6ba6f78534f/thumbnail` 返回 200 image/jpeg；带 token 的 `curl -r 0-0 /api/v1/videos/2bb55884-3647-4c4a-81f8-c6ba6f78534f/source` 返回 206 video/mp4；`curl /api/v1/admin/videos?...` 返回 200。

## 2026-06-01 13:31 +0800
- 进度：完成图片接口外盘休眠兜底。新增本地图片打开 helper，统一为缩略图/TV 海报/TV 分集剧照使用 2 秒打开超时与 8 个并发槽；超时或槽位耗尽时返回 503 + `Retry-After`，TV 图片若有 TMDB/远端原始地址则优先回退重定向。旧 `/api/v1/videos/:id/thumbnail` 不再在解析阶段做无界 `os.Stat`，而是复用已打开的文件响应，避免 `c.File` 再次阻塞。`CONTEXT.md` 追加 [[家用部署机外盘休眠契约]]；12:39 的 admin 跳转假设保留为排查记录，本轮未修改管理端，当前直接 blocker 已收敛为外盘休眠导致图片接口 pending。
- 影响文件：`internal/handlers/local_image_file.go`、`internal/handlers/local_image_file_test.go`、`internal/handlers/video_source.go`、`internal/handlers/tv_artwork.go`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`go test ./internal/handlers -count=1` 通过；`go test ./internal/... -count=1` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/handlers/local_image_file.go internal/handlers/local_image_file_test.go internal/handlers/video_source.go internal/handlers/tv_artwork.go` 无输出。

## 2026-06-01 13:20 +0800
- 进度：校正管理端视频管理页 pending 的根因。最近提交 `d8eb937` 未修改旧的 `/api/v1/videos/:id/thumbnail` 处理函数，但新增 TV 分集剧照本地路由并让 TV 详情页批量请求分集图片；部署机外部媒体盘 `/Volumes/large` 会休眠，图片接口里无超时的 `os.Stat` / `c.File` 会在外盘唤醒期间阻塞，旧视频封面接口和新增 TV 剧照接口都存在同类风险。下一步为图片类本地文件响应增加超时与并发上限，外盘休眠时快速返回而不是浏览器一直 pending。
- 影响文件：预计涉及 `internal/handlers/local_image_file.go`（新）、`internal/handlers/video_source.go`、`internal/handlers/tv_artwork.go`、后端单测、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行 `go test ./internal/handlers -count=1`、`git diff --check`、乱码扫描。

## 2026-06-01 12:39 +0800
- 进度：开始排查管理端视频管理页打不开。部署机 `/healthz`、`/admin`、管理端 JS/CSS 均可访问；根路径 `/` 返回 404 属于现有路由行为。日志显示 `/api/v1/admin/videos` 在部署后仍有 200 记录，TV 详情页提交未修改 `AdminVideos`、`AdminListVideos` 或管理端构建产物。进一步发现管理端 axios 认证失效处理直接 `location.replace('/login?...')`，生产 `/admin` 基路径下会跳到后端根路径 `/login` 并 404，表现为视频列表请求后“网站打不开”。下一步修正认证失效跳转基路径并补单测。
- 影响文件：预计涉及 `admin-web/src/api/request.js`、管理端 API 请求测试、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行管理端定向单测、`npm run build`、`git diff --check`、乱码扫描，并推送部署机。

## 2026-05-31 15:44 +0800
- 进度：完成 TV 电视剧详情页返修收尾验证。TV 全量单测首次因新增 `RoundedCornerShape(6.dp)` 违反 TV 圆角白名单失败，已改为复用 `AppChrome.ChipShape` 并重跑通过。准备只暂存本任务文件，继续保留用户既有改动 `admin-web/.env.development` 不纳入提交。
- 影响文件：同 15:43 记录；提交范围不包含 `admin-web/.env.development`
- 验证：`go test ./internal/utils ./internal/repository ./internal/handlers ./internal/services -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-31 15:43 +0800
- 进度：完成 TV 电视剧详情页返修实现。详情主体继续保持无全局侧栏/顶部导航，左侧重做参考图式居中大标题、胶囊元信息、金色评分与「播放第 X 集 / 我的片单」操作；右侧分集列表改为更高的横向剧照卡、金色选中边框与圆形播放按钮。后端新增 TV 分集 still 本地访问路由，详情接口返回本地 still 路由，刮削同步分集时下载到 `storage/tv/series/<id>/episodes/sXXeYY.jpg`。TV 版本升级到 `0.1.80 (80)`，`CONTEXT.md` 追加详情主体还原边界与分集剧照本地化约定。
- 影响文件：`internal/utils/video_url.go`、`internal/utils/video_url_test.go`、`internal/repository/tv_repository.go`、`internal/repository/tv_episode_still_url_test.go`、`internal/handlers/tv_artwork.go`、`internal/handlers/router.go`、`internal/services/scraper.go`、`internal/services/scraper_episode_sync_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest'` 通过；`go test ./internal/utils ./internal/repository ./internal/services -run 'TestTVEpisodeStillURL|TestResolveTVEpisodeStillURL|TestScrapeEpisodeUploadDownloadsSeriesArtworkLocally' -count=1` 首次因 sandbox 禁止 `httptest` 监听本地端口失败，提权重跑通过；待执行后端受影响包全量、TV 全量单测、`git diff --check`、乱码扫描与提交。

## 2026-05-31 15:20 +0800
- 进度：开始返修 TV 电视剧详情页还原度。用户确认仍只做详情主体、不做全局侧栏/顶部导航，并要求抛弃之前字体和布局的保守约束，按参考图更精准还原；同时希望每集海报下载到本地。代码现状确认：后端 `episodes.still_path` 已落库并同步 TMDB still，但 TV 详情接口目前直接返回原始 `still_path`，没有本地访问路由；同步逻辑只下载 series poster/backdrop，未下载每集 still 到本地。下一步补后端分集 still 本地路由/下载测试，再重做 TV 详情页标题、元信息、按钮和右侧剧集卡比例。
- 影响文件：预计涉及 `internal/services/scraper.go`、`internal/repository/tv_repository.go`、`internal/handlers/tv_artwork.go`、`internal/handlers/router.go`、`internal/utils/video_url.go`、后端相关测试、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、TV 源文测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：红灯阶段 TV 源文测试 `TvSeriesDetailActionSpecTest` 因缺少 `TvSeriesTitleBlock` 失败；后端定向测试首次因 sandbox 无法写 Go build cache 被拦，待提权重跑确认红灯；待执行后端分集 still 路由/工具定向测试、TV 电视剧详情定向单测、Go/TV 受影响全量验证、`git diff --check`、乱码扫描。

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

## 2026-05-31 02:16 +0800
- 进度：开始返修电视剧播放器交互问题。现有 `CONTEXT.md` 已定义 [[选集轨双态高亮]]、[[controls 左右键切焦点]] 与 [[controls 焦点环绕]]，用户反馈实际仍出现选集切换放大/位移、controls 聚焦后 LEFT/RIGHT 未稳定切焦点。本轮先补回归测试锁住“集卡固定槽位 + 静态高亮”和“controls 持焦控件本地处理左右切焦点”，再做最小实现，避免只改源文断言。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行电视剧播放器相关定向单测、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`git diff --check`、乱码扫描。

## 2026-05-31 01:50 +0800
- 进度：完成电视剧播放器“中间切换多按一次下键”修复。`LongFormVideoPlayer` 在电视剧 controls 持焦层增加本地 `DPad DOWN` 兜底，确保第二次 DOWN 直接进入选集轨；保留 controls/选集轨焦点请求重试，选集卡继续只用静态高亮无 glow。同步更新源文审计、TV 版本到 `0.1.77 (77)`，并在 `CONTEXT.md` 追加“播放器内纵向切页无空按”约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- CONTEXT.md plan.md android-tv-app/tv-app/build.gradle.kts android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt` 通过；乱码扫描无输出。

## 2026-05-31 01:46 +0800
- 进度：继续修正电视剧播放器“中间切换多按一次下键”。已确认 01:38 记录里“第一次 DOWN 直接进选集页”的假设不成立，正确语义仍是 `播放器根 -> controls 页 -> 选集页`。本轮改为在 controls 持焦层本地兜底消费 `DPad DOWN` 进入选集轨，不再只依赖播放器根层路由；同时保留此前的焦点请求重试与集卡静态高亮收口。下一步补源文审计、跑 TV 定向/全量单测、更新版本与提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`git diff --check`、乱码扫描。

## 2026-05-31 01:38 +0800
- 进度：开始继续修正电视剧播放器交互。用户确认两点新收口：选集卡焦点只保留静态高亮，不要 scale，也不要 glow；电视剧播放器第一次 DPad DOWN 必须从播放器根直接进入选集页，不再要求先进入 controls。下一步修改 `TvLongFormRemoteKeyRouting` / `LongFormVideoPlayer` 的进入流、选集卡焦点样式和相关测试，并再次验证 controls 页左右键焦点循环。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting}.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行电视剧播放器遥控路由/选集轨定向单测与 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`，以及 `git diff --check`、乱码扫描。

## 2026-05-31 01:13 +0800
- 进度：完成电视剧播放器交互 bug 修正。`LongFormVideoPlayer` 的电视剧分支改为底部单容器双页：`controls 页` 与 `选集页` 通过纵向 slide 切换，不再同时堆成双层；电视剧首屏恢复为按 DPad DOWN 唤出 controls；controls 聚焦时 LEFT/RIGHT 只切焦点，不做 seek。选集轨集卡改为“第一集”“第二集”等中文序数，左右切集时移除集卡淡入淡出和整段明显滚动，只在目标集快出边界时做最短必要跟随；连续快进/快退时进度显示改为跟随 pending seek 目标，避免进度条抖动。TV 版本升级到 `0.1.76 (76)`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvEpisodeRailPolicy}.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvEpisodeRailPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesEpisodeRailSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvEpisodeRailPolicyTest' --tests 'com.chee.videos.feature.tv.TvSeriesEpisodeRailSpecTest' --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-05-31 01:13 +0800
- 进度：开始修正电视剧播放器交互 bug。已重新通过 grill-with-docs 锁定四点：`controls 页` 与 `选集页` 改为单容器双页纵向 slide，而不是双层同时展开；焦点位于 controls 时 DPad LEFT/RIGHT 只切焦点，不承担 seek；选集卡文案改为“第一集”“第二集”等中文序数；左右切集时不再给单个集卡叠加明显过渡，只在目标集快出边界时做最短必要的跟随滚动。下一步修改 `LongFormVideoPlayer`、选集轨策略与源文/纯逻辑测试，并同步更新 TV 版本与 `CONTEXT.md`。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvEpisodeRailPolicy}.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：待执行电视剧播放器相关定向单测与 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`，以及 `git diff --check`、乱码扫描。

## 2026-05-30 23:55 +0800
- 进度：完成电视剧播放器 UI 重构收尾。`LongFormVideoPlayer` 新增 `SeriesEpisodeRail` 变体、`播放器根 → controls → 选集轨` 三层焦点路由、只读进度条与横向选集轨；`TvSeriesPlayerScreen` 移除旧 `ModalBottomSheet` 选集，改为当前季分集卡片内嵌轨道并通过标题气泡跟随焦点；补充选集轨策略 / 遥控路由 / 源文审计测试，并修正 `TvScrollableBottomPaddingTest` 使其不再把沉浸式播放器页当滚动页。TV 版本升级到 `0.1.75 (75)`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting,TvEpisodeRailPolicy}.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/{TvLongFormRemoteKeyRoutingTest,TvLongFormControlsAutoHideTest,TvEpisodeRailPolicyTest}.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/{TvSeriesEpisodeRailSpecTest,TvScrollableBottomPaddingTest}.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；待执行 `git diff --check` 与乱码扫描后提交。

## 2026-05-30 23:34 +0800
- 进度：开始实现电视剧播放器 UI 重构。已通过 grill-with-docs 锁定交互：电视剧播放器采用 `播放器根 → controls → 选集轨` 三层焦点栈；controls 收敛为 `播放/暂停`、只展示不交互的进度条、`字幕`、`音轨` 四项；选集从 `ModalBottomSheet` 改为与 controls 同体系的底部第二层横向选集轨，显示当前季集数卡片、焦点标题气泡、不可播集弱化占位、重新进入时回到当前播放集。下一步先补纯逻辑与源文测试，再改 `LongFormVideoPlayer` / `TvSeriesPlayerScreen` 结构。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting}.kt`、电视剧选集轨 helper、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、相关 TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 端定向单测（遥控按键、播放器源文、电视剧播放页/选集轨策略）与全量 `:tv-app:testDebugUnitTest`、`git diff --check`、乱码扫描。

## 2026-05-30 09:38 +0800
- 进度：完成 TV App 退出 IPTV 播放卡住修复。IPTV `LibVLC` 改为独立单例复用，退出页面时不再在 `onDispose` 主线程同步 `stop()` 直播流或释放库实例，只保留 `MediaPlayer.release()`；`detachViews()` 改为独立 `DisposableEffect` 处理，和长视频 LibVLC 生命周期保持同类模式。同步补充 IPTV 配置/源文测试约束，TV 版本升级到 `0.1.74 (74)`，`CONTEXT.md` 追加 IPTV 退出清理契约。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvPlaybackConfig.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackConfigTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackConfigTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest' --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-30 09:34 +0800
- 进度：继续修复 TV App 退出 IPTV 播放时卡住。代码复查确认 IPTV 退出路径会在 `DisposableEffect.onDispose` 主线程同步执行 `vlcPlayer.stop()`、`detachViews()`、`release()`、`libVlc.release()`，而返回动作本身直接 `popBackStack()`；这与长视频 LibVLC 仅 `release()`、视图层单独 `detachViews()` 的模式不一致，属于高风险阻塞点。下一步改为 IPTV LibVLC 实例复用、退出不显式 `stop()`、视图解绑独立处理，并补源文测试锁死该释放契约。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptv*.kt`、TV IPTV 源文测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 IPTV 定向单测、TV 全量单测、`git diff --check` 与乱码扫描。

## 2026-05-30 09:27 +0800
- 进度：完成 TV App IPTV 卡顿优化首轮落地。新增 IPTV 专属 LibVLC 播放配置 helper，将直播缓存从 `1500ms` 提升到 `4000ms`，移除 `clock-jitter=0` / `clock-synchro=0` 低延迟参数，统一收口到 helper 以防后续散落魔法数字；同步补充 IPTV 配置单测与源文约束测试，TV 版本升级到 `0.1.73 (73)`，`CONTEXT.md` 追加 [[IPTV 流畅优先]] 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvPlaybackConfig.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackConfigTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`；未纳入用户既有改动 `admin-web/.env.development`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackConfigTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest' --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check -- ...` 通过；乱码扫描无输出。

## 2026-05-30 09:25 +0800
- 进度：开始检查并优化 TV App IPTV 播放卡顿。已确认用户接受“流畅优先”，即允许直播延迟增加以减少播放中缓冲/顿挫；当前代码 IPTV LibVLC 使用 `network-caching=1500` 且强制 `clock-jitter=0` / `clock-synchro=0`，偏低延迟而非抗抖动。下一步抽出 IPTV 播放配置、补单测、提高直播缓存并移除低延迟时钟参数。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、IPTV 配置 helper、TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV IPTV 定向单测、TV 全量单测或构建、`git diff --check` 与乱码扫描。

## 2026-05-28 23:11 +0800
- 进度：完成部署机验证。已将修复提交 `5d30069` 推送到 `deploy master`，post-receive 触发 `admin-web` 构建并替换 `/Users/chee/deploy/ai-video-server/current/admin-web-dist`；Go server 未重启。部署过程中 GitHub mirror push 卡住，已终止该 best-effort 子进程，hook 按预期继续执行并完成前端更新。
- 影响文件：`plan.md`
- 验证：`curl http://192.168.1.24:8080/admin` 返回 200，HTML 引用新资源 `/admin/assets/index-eWUl6XQ3.js`，`Last-Modified: Thu, 28 May 2026 15:10:07 GMT`；线上 JS 已确认包含 `/admin/` router history base；部署日志显示 `frontend updated` 与 `deploy ... done`。

## 2026-05-28 23:06 +0800
- 进度：完成管理端 `/admin` 空白页修复。新增 `resolveRouterHistoryBase` 纯函数与回归测试，`admin-web/src/router/index.js` 改为用 Vite `BASE_URL` 初始化 `createWebHistory`，保证生产 `/admin/` base 与后端静态挂载一致；`CONTEXT.md` 已补充 `admin SPA 基路径契约`。部署机当前线上 JS 已确认仍是无参 history，待提交后 push deploy 更新。
- 影响文件：`admin-web/src/router/index.js`、`admin-web/src/router/historyBase.js`、`admin-web/src/router/index.spec.js`、`CONTEXT.md`、`plan.md`；不纳入用户既有改动 `admin-web/.env.development`
- 验证：`npm run test -- src/router/index.spec.js src/router/transition.spec.js` 通过；`npm run build` 通过；`git diff --check -- admin-web/src/router/index.js admin-web/src/router/historyBase.js admin-web/src/router/index.spec.js CONTEXT.md plan.md` 通过；`rg -n $'\uFFFD' ...` 无输出。

## 2026-05-28 23:02 +0800
- 进度：开始排查 `http://192.168.1.24:8080/admin` 空白页。当前确认管理端生产构建 `base=/admin/`，后端静态挂载 `/admin` 与 `/admin/assets`；前端 Vue Router 仍使用默认根路径 `createWebHistory()`，访问 `/admin` 时会被当作应用内路径导致无匹配页面。下一步补路由基路径回归测试并改为读取 Vite `BASE_URL`。
- 影响文件：预计涉及 `admin-web/src/router/index.js`、管理端路由测试、`plan.md`
- 验证：待执行管理端定向测试与 `npm run build`。

## 2026-05-28 05:20 +0800
- 进度：完成远程支撑层开发模式代码与文档更新。`dev-up.sh` 支持 `DEV_DATA_MODE=remote` 跳过本地 Postgres/Redis 与 migration，并在外部 `TRANSLATION_API_URL` 下检查 `/v1/models` 后复用远程翻译服务；`dev-down.sh` 会读取 env 或 `.run/dev-data-mode`，远程模式只停本机进程、不停本地数据容器；`migrate-apply.sh` 对非本机 DSN 默认拒绝，必须 `ALLOW_REMOTE_MIGRATIONS=1` 才能手动执行。`docs/run.md` 与 `docs/家用部署机.md` 已写明远程翻译需要部署机暴露 LAN 端点或 SSH tunnel；`CONTEXT.md` 已沉淀 [[远程支撑层开发模式]]。本机 `.env.remote-local` 已改为指向部署机翻译端点但不纳入提交。
- 影响文件：`scripts/dev-up.sh`、`scripts/dev-down.sh`、`scripts/migrate-apply.sh`、`docs/run.md`、`docs/家用部署机.md`、`.gitignore`、`CONTEXT.md`、`plan.md`；本机未提交文件 `.env.remote-local`
- 验证：`bash -n scripts/dev-up.sh scripts/dev-down.sh scripts/migrate-apply.sh` 通过；远程 migration 防呆按预期拒绝 `192.168.1.24`；本地 DSN 不触发远程防呆但因本机无 `psql` 停在 psql 前置；`ENV_FILE=.env.remote-local bash scripts/dev-up.sh --frontend off` 确认远程 Postgres/Redis 可达、跳过本地 DB/migration，并在远程翻译 `http://192.168.1.24:8000/v1/models` 不可达处明确失败；`git diff --check` 通过；`rg -n $'\uFFFD' ...` 无输出。

## 2026-05-28 05:20 +0800
- 进度：用户要求把翻译服务也纳入远程开发链路。当前代码已经支持外部 `TRANSLATION_API_URL`，本次补齐的是远程 env 与文档：` .env.remote-local` 将翻译端点切到部署机，`docs/run.md` 说明远程翻译 URL / 可达性前提，`CONTEXT.md` 把术语从“远程数据层”扩展到“远程支撑层”并明确翻译服务也是该模式的一部分。若部署机翻译服务仍只绑定 loopback，则需要先在部署机侧暴露端口或走隧道，仓库不会替你隐藏这个现实约束。
- 影响文件：`.env.remote-local`、`docs/run.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行乱码检查、文档 diff 复查与（如可能）远程翻译 URL 可达性确认。

## 2026-05-28 05:14 +0800
- 进度：开始落地远程数据层开发模式。目标是 `ENV_FILE=.env.remote-local bash scripts/dev-up.sh` 启动本机 Go server / worker / admin 前端，但通过 `DEV_DATA_MODE=remote` 直连家用部署机 Postgres / Redis；该模式默认不启动本地 docker 数据层、不自动执行 migration。数据库改动流程按已确认策略执行：先在本地 Docker DB 验证 migration，提交后由家用部署机部署流程执行；必须手动改远程库时用 `ALLOW_REMOTE_MIGRATIONS=1` 显式放行并打印警告。
- 影响文件：`scripts/dev-up.sh`、`scripts/dev-down.sh`、`scripts/migrate-apply.sh`、`docs/run.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行 shell 语法检查、远程模式防呆验证、乱码检查与 `git diff --check`。

## 2026-05-28 05:10 +0800
- 进度：完成 `.codex/skills/repo-dev-workflow` 更新。技能新增 `CONTEXT.md` 沉淀边界、Android Java 17 前置、家用部署机/部署脚本入口、脏 `plan.md` 精确暂存、App 版本号、migration 前向兼容、tasks 三段流、TV LibVLC/IPTV 定向验证和提交后复查等仓库近期实践约束。`agents/openai.yaml` 描述仍匹配，无需改动。
- 影响文件：`.codex/skills/repo-dev-workflow/SKILL.md`、`plan.md`
- 验证：`python3 /Users/cuiqi/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/repo-dev-workflow` 通过；`rg -n $'\uFFFD' .codex/skills/repo-dev-workflow/SKILL.md .codex/skills/repo-dev-workflow/agents/openai.yaml plan.md` 无输出；`git diff --check -- .codex/skills/repo-dev-workflow/SKILL.md .codex/skills/repo-dev-workflow/agents/openai.yaml plan.md` 通过。

## 2026-05-28 05:04 +0800
- 进度：开始更新 `.codex/skills/repo-dev-workflow`，把近期 `CONTEXT.md`、`plan.md` 与代码实践沉淀成仓库执行流程。重点补齐：任务三段流、家用部署机与 migration 前向兼容、Android Java 17/Gradle 前置、TV LibVLC/IPTV 专项验证、admin/Go/Android 验证选择，以及已有脏工作区下的精确暂存提交策略。
- 影响文件：`.codex/skills/repo-dev-workflow/SKILL.md`、必要时 `.codex/skills/repo-dev-workflow/agents/openai.yaml`、`plan.md`
- 验证：待执行 skill 静态校验、乱码检查与 git diff 复查。

## 2026-05-28 05:16 +0800
- 进度：为 Android TV 项目安装并固定 Java 17 开发环境。已下载 Temurin 17.0.19（arm64）到 `~/.jdks/jdk-17.0.19+10`，并写入 `~/.gradle/gradle.properties` 的 `org.gradle.java.home`，让 Gradle wrapper 不再依赖系统自带 Java 8。后续 Android 构建与单测应以该 JDK 为准。
- 影响文件：`/Users/cuiqi/.jdks/jdk-17.0.19+10`、`/Users/cuiqi/.gradle/gradle.properties`、`plan.md`
- 验证：`/Users/cuiqi/.jdks/jdk-17.0.19+10/Contents/Home/bin/java -version` 通过；`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest` 已从 Java 8 阻塞切换为 Android 依赖解析阻塞。

## 2026-05-28 05:02 +0800
- 进度：完成 TV App IPTV 硬解默认改造。移除 IPTV LibVLC 初始化参数 `--avcodec-hw=none`，频道 Media 改为 `setHWDecoderEnabled(true, true)`；补源文回归测试禁止回到关闭硬解；TV 版本号升至 `0.1.72` / `versionCode=72`；`CONTEXT.md` 中 IPTV 旧“软解优先”约定改为“TextureView + 硬解默认”，保留直播诊断路径。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n "avcodec-hw=none|setHWDecoderEnabled\\(false|setHWDecoderEnabled\\(true, true\\)|versionCode =|versionName =|软解优先|关闭硬解" ...` 确认 IPTV 已改为硬解默认且版本已更新；`rg -n $'\uFFFD' ...` 无乱码；`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest` 未通过环境前置，当前命令行 Java 为 1.8，AGP 8.5.2 需要 Java 11+。

## 2026-05-28 04:50 +0800
- 进度：开始修改 TV App IPTV 播放硬解策略。现状确认：IPTV LibVLC 初始化含 `--avcodec-hw=none`，每个频道 Media 调 `setHWDecoderEnabled(false, false)`；这与用户新要求“IPTV 播放也要使用硬解”冲突，也与 `CONTEXT.md` 现有“IPTV 软解优先”约定冲突。下一步先补源文回归测试锁定 IPTV 硬解默认，再改实现、TV 版本号、`CONTEXT.md` 和验证记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 端定向单测与乱码检查。

## 2026-05-26 20:05 +0800
- 进度：用户要求安装 Go 以便本地后端连接远程 Postgres/Redis 调试；当前 `go.mod` 声明 `go 1.22`，本机 `Darwin arm64`，`go` 与 Homebrew 均不可用。下一步通过官方 Go tarball 安装到用户目录，避免依赖 sudo/Homebrew。
- 影响文件：`plan.md`；开发机本地 `~/sdk/go1.22.x` 与 shell PATH 配置。
- 验证：待执行 `go version`、`go env GOPATH`、`go test ./...` 或至少 `go test ./internal/config -count=1`。

## 2026-05-26 19:32 +0800
- 进度：用户确认允许按部署指南在开发机生成 SSH 公私钥；下一步按 `docs/家用部署机.md` 执行 `ssh-keygen -t ed25519 -N '' -f ~/.ssh/id_ed25519`，随后验证本机是否已携带该身份，并继续评估免密登录是否还受 `authorized_keys` 缺失阻塞。
- 影响文件：`plan.md`；开发机本地 `~/.ssh/id_ed25519`、`~/.ssh/id_ed25519.pub`
- 验证：待执行 `ls -la ~/.ssh`、`ssh-keygen -lf ~/.ssh/id_ed25519.pub`、`ssh -o BatchMode=yes chee@192.168.1.24 'true'`。

## 2026-05-26 19:34 +0800
- 进度：已在开发机生成新的 `ed25519` 密钥对，公钥指纹为 `SHA256:fn0mPxj7+7WHXYrawG/AJ+tm8ULJNKe7BvCEukt0lxw`。再次尝试 `ssh -o BatchMode=yes -o PreferredAuthentications=publickey chee@192.168.1.24 ...` 仍被拒绝，说明阻塞点已收敛为部署机 `chee` 用户的 `~/.ssh/authorized_keys` 尚未包含这把新公钥。
- 影响文件：`plan.md`；开发机本地 `~/.ssh/id_ed25519`、`~/.ssh/id_ed25519.pub`
- 验证：`ssh-keygen -lf ~/.ssh/id_ed25519.pub` 返回上述指纹；`ssh -o BatchMode=yes -o PreferredAuthentications=publickey -o ConnectTimeout=5 chee@192.168.1.24 'hostname && date'` 失败，报 `Permission denied (publickey,password,keyboard-interactive)`。

## 2026-05-26 19:27 +0800
- 进度：开始按 `docs/家用部署机.md` 验证开发机到家用部署机的推送链路；先核对 SSH/`deploy` remote/目标主机状态，再执行一次最小化 `git push deploy master` 推送测试。
- 影响文件：`plan.md`
- 验证：待执行 `git remote -v`、`ssh chee@<部署机> 'hostname && date'`、`git push deploy master`、远端 `deploy.log` / `curl /healthz` 核对。

## 2026-05-26 19:30 +0800
- 进度：完成链路前置核对。结果：远端 `192.168.1.24:22` 可达，已将 host key 写入本机 `~/.ssh/known_hosts`，并按文档把本地 Git `deploy` remote 配置为 `chee@192.168.1.24:/Users/chee/deploy/ai-video-server/repo.git`。阻塞点确认在 SSH 认证层：本机 `~/.ssh/` 只有 `known_hosts`，没有 `id_ed25519`/`id_rsa` 等私钥，`ssh-add -l` 返回 `The agent has no identities.`，`ssh -o BatchMode=yes chee@192.168.1.24 'true'` 返回 `Permission denied (publickey,password,keyboard-interactive)`，因此本轮无法实际触发 `git push deploy master`。
- 影响文件：`plan.md`
- 验证：`ssh-keyscan -H -T 5 192.168.1.24 CheedeMac-mini.local >> ~/.ssh/known_hosts` 成功；`git remote get-url deploy` 返回 `chee@192.168.1.24:/Users/chee/deploy/ai-video-server/repo.git`；`ssh -o BatchMode=yes -o PreferredAuthentications=publickey -o ConnectTimeout=5 chee@192.168.1.24 'true'` 失败，错误如上。
## 2026-05-26 17:30 +0800
- 进度：[[家用部署机]] 把 Docker Desktop 替换为 OrbStack，省 RAM 与磁盘开销。整体路径走的不是 "orb docker migrate" 自动迁移（被 Docker Desktop 内部一条 stale container ID `3d8f9aa3bf19` 卡住，`docker ps -a` 见但 `inspect/rm` 都 No such container —— Docker Desktop metadata 损坏），改走 **pg_dump → 备份 → restore** 兜底路径：(1) 先 `pg_dump video_server` 备份到 `/Volumes/large/ai-video-server/backup/video_server-20260526-171346.sql.gz`（33 MB），(2) 停 launchd + docker compose down + 退 Docker Desktop，(3) brew install --cask orbstack 起 OrbStack，(4) 用 `docker save postgres:15` 从 Docker Desktop 拷镜像到 /tmp 再 `docker load` 到 OrbStack（绕开 Docker Hub `EOF` 网络问题），(5) `docker run` 起新 video_server_postgres 容器绑 ai-video-server_pg_data named volume，(6) gunzip + psql restore，(7) launchctl bootstrap server + worker。最终验证：`SELECT count(*) FROM videos` = 154 968（与备份一致），/healthz / /admin/ / /api/v1/tv/home 全 200。中间一次 pkill com.docker.backend 时把 OrbStack 也踢断，导致 volume / image / container 全失，靠提前的 pg_dump + 之前 save 的 /tmp/postgres15.tar 完整回滚回来——双保险机制证明本次迁移的"备份先行"原则不是过度谨慎。
- 影响文件：本机系统级改动（不进 repo）：Docker Desktop 容器 + image 已退、OrbStack 接管，`docker context use orbstack` 永久切换，Docker Desktop GUI app 与 root-owned 数据待用户自行 sudo 清理（约 4.5 GB 磁盘）。`plan.md`。
- 验证：`launchctl print gui/501/com.aivideo.{server,worker}` 都 state=running；`curl /healthz` / `curl /api/v1/tv/home` / `curl /admin/` 全 200；`docker exec video_server_postgres psql ... SELECT count(*) FROM videos` = 154968；`vm_stat` 显示 Free+Inactive+Compressed ≈ 7.7 GiB（迁移前 6.6 GiB），净腾出 ~1 GiB 内存余量（OrbStack 还在 warm-up，稳态会再降 200~400 MB）。pg_dump 备份保留在 `/Volumes/large/.../backup/` 作为长期 escape hatch。

## 2026-05-26 15:00 +0800
- 进度：docs/家用部署机.md 新增第三节"开发机一次性配置"，覆盖 ssh-keygen / ssh-copy-id / ssh 验证 / `git remote add deploy` / 首次 push / 之后日常 6 步；原"7. 开发机加 remote 并推第一次"已并入新章节并改名"7. 部署机最后开服"，只保留 launchd bootstrap 与 /healthz 验证。整体章节序号重排：开发机配置占据 §三，原 §三日常运维 → §四，原 §四故障速查 → §五，原 §五弃用与清理 → §六。
- 影响文件：`docs/家用部署机.md`、`plan.md`
- 验证：grep `^## ` 章节序号连贯一、二、三、四、五、六，无重复；新章节示例命令均与之前实际部署得出的踩坑保持一致（ssh-copy-id 走 LAN IP；远端路径 `/Users/chee/deploy/ai-video-server/repo.git` 绝对；hook 输出格式与已落 `b11dcc3eeb8` 真实 push 时回流到 stderr 的样式一致）。

## 2026-05-26 14:05 +0800
- 进度：[[家用部署机]] 在 `CheedeMac-mini.local` 完成实际启用。状态：launchd 看管的 server (pid 25394) + worker (pid 25406) 跑在 :8080，`/healthz` 200、`/admin/` 服务 admin SPA、LAN 192.168.1.24:8080 可达。dev-up.sh 已 dev-down，野生 brew redis 加密码 + bind 0.0.0.0 LAN 可达（440 keys 数据保留），现有 Postgres docker 容器 LAN 可达（21 migrations / 154 968 videos 数据保留）。bare repo + work tree + binaries/ + current/ + .env 全部就绪；post-receive hook 落盘并 dry-run 通过（mkdir 原子锁、分桶分支、GitHub mirror push、launchctl kickstart、/healthz 探活、3-binary 保留）。**家用部署机正式接管家庭客户端流量**。本轮还顺带从真实部署里挖出 3 条踩坑，**同步更新到 docs/家用部署机.md**：(1) macOS 没有 `flock`，hook 改用 `mkdir` 原子锁；(2) `.env` 文件首字节 UTF-8 BOM 会让 godotenv 把首行变量解析为 `﻿APP_MODE` 导致全局 env 装载失败，dev-up.sh 不走 godotenv 所以历史一直没暴露；(3) zsh 把 `$USER:staff` 当 history substitution（`:s/t/aff` 模式），写 newsyslog 配置必须用 `${USER}:staff`。
- 影响文件：`docs/家用部署机.md`（剥 BOM 警告 + SSH 启用步骤 + flock → mkdir 锁 + ${USER} 大括号警告 + 分桶 case 扩 .agents/.codex/references/.run + migrate-apply.sh 传 POSTGRES_CONTAINER + mv 改普通 -f）、`plan.md`；部署机文件系统改动（不进 repo）：`~/deploy/ai-video-server/{repo.git,work,binaries,current,.env}`、`~/Library/LaunchAgents/com.aivideo.{server,worker}.plist`、`~/Library/Logs/ai-video-server/{server,worker,deploy}.log`、`/opt/homebrew/etc/redis.conf`（bind 0.0.0.0 + requirepass）。
- 验证：`curl http://127.0.0.1:8080/healthz` 与 `curl http://192.168.1.24:8080/healthz` 均 200；`curl http://127.0.0.1:8080/admin/` 返回 admin SPA 首页 + assets 200；`redis-cli -h 192.168.1.24 -a <pwd> ping` 远程 PONG；`launchctl print gui/501/com.aivideo.{server,worker}` 显示 state=running；post-receive hook 以同 sha 跑一次 dry-run 输出符合预期。剩余两步需 sudo（启用 SSH 远程登录 + 安装 newsyslog 规则）已经把命令塞进交付清单等用户自行执行。

## 2026-05-26 12:55 +0800
- 进度：完成 `tasks/2026-05-26-admin-web-dist-path-env/`。按 TDD 红→绿走完：先写 `internal/config/config_admin_dist_test.go`（env 设/默认 两条）与 `internal/handlers/admin_static_test.go`（mountAdminStatic 给定 dir 服务/缺失 dir 跳过 两条），红灯确认 `cfg.AdminWebDistPath undefined` + `undefined: mountAdminStatic`；再做绿灯：Config 加 `AdminWebDistPath` 字段 + `getEnv("ADMIN_WEB_DIST_PATH", "admin-web/dist")` 默认；`internal/handlers/router.go` 抽出 `mountAdminStatic(r, adminDist)` helper，原 inline 21 行挪进去；`API` struct 加 `adminWebDistPath` 字段，`NewAPI` 新增第 23 个 positional 参数；`main.go` 传 `cfg.AdminWebDistPath`；`.env.example` 补 `ADMIN_WEB_DIST_PATH=` 含中文注释。全套 `go test ./...` 全绿、`go build ./...` / `go vet` 通过。手测三场景未在本机做端到端启动（dev-up.sh 旧 server 仍占 :8080，强制重启会冲击你正在跑的 dev 流；本改动属 plain wire + helper extraction，单测充分覆盖），具体延后到家用部署机首次启用时按 review.md §1 走。
- 影响文件：`internal/config/config.go`、`internal/config/config_admin_dist_test.go`（新）、`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`（新）、`main.go`、`.env.example`、`tasks/2026-05-26-admin-web-dist-path-env/DONE.md`（新）、`plan.md`。
- 验证：`go test ./internal/config ./internal/handlers -count=1` 全绿（含新增 4 条单测）；`go test ./... -count=1` 全绿（不影响其它包）；`go build ./...` 与 `go vet ./internal/handlers ./internal/config` 无 warning；`git diff --stat` 显示仅 4 个生产代码文件 + 2 个新测试文件 + 1 个 task 完成标记，符合 PRD §2 作用域。

## 2026-05-26 12:30 +0800
- 进度：生成 `tasks/2026-05-26-admin-web-dist-path-env/` 三件套（prd / implement / review），承接 [[家用部署机]] ADR-0005 / docs/家用部署机.md 中点名的代码前置条件——`internal/handlers/router.go:221` 的 admin-web dist 路径从硬编码相对路径改为读 `ADMIN_WEB_DIST_PATH` 环境变量，默认值保持现状以不破坏 dev-up.sh 体验。本任务**未实施**，待用户触发"完成 tasks"或等价指令时按 PRD → Implement → Review 三段流推进。
- 影响文件：`tasks/2026-05-26-admin-web-dist-path-env/{prd,implement,review}.md`、`plan.md`
- 验证：本轮仅文档沉淀，无代码改动；任务的红灯/绿灯/手测在实施阶段执行。

## 2026-05-26 12:15 +0800
- 进度：完成 [[家用部署机]] 的 grill-with-docs 设计沉淀，把"另一台机器写代码 → push → 本机 hook 自动重启服务"这一想法拆成 8 轮独立决策并全部锁定：（Q1）家庭日常在线使用，N≈2–5；（Q2）a-1 自托管 bare repo on LAN，GitHub 降级为 mirror；（Q3a/b/c）launchd 作 supervisor + 硬切重启 + 数据层独立生命周期；（Q4a/b/c/d）部署机本地 build + path 分桶 + fail-open；（Q5a/b/c/d）migrate-apply 抽脚本 + 前向兼容契约 + 手动 rollback + 保留 3 个二进制；（Q6a/b/c）worker 硬切不 drain + 接受 in-flight 上传/转码损失 + 统一日志 `~/Library/Logs/ai-video-server/`；（Q7a/b/c）绝对路径契约 + `/Volumes/large` 外挂盘 + flock 串行 push；（Q8）纯文档 + 两个 helper 脚本 + 两份 ADR。本轮**只产出设计与脚手架**，不实际启用部署机；`internal/handlers/router.go` 对 `ADMIN_WEB_DIST_PATH` 环境变量的支持是后续独立 task，启动前必须完成。
- 影响文件：`CONTEXT.md`（新增"部署术语"区块共 7 条术语 + "Git 忽略规则约定"无变化）、`docs/adr/0005-home-deployment-architecture.md`、`docs/adr/0006-migration-forward-compatibility.md`、`docs/家用部署机.md`、`scripts/migrate-apply.sh`（dev-up.sh 和部署 hook 共用的迁移执行器，行为与 dev-up.sh apply_migrations 等价）、`scripts/rollback.sh`（手动 rollback 工具）、`plan.md`。
- 验证：本轮仅文档与脚本沉淀，无代码运行验证；`scripts/migrate-apply.sh` 与 `scripts/rollback.sh` 走 `chmod +x` 后等待真实部署机环境联调。**家用部署机尚未启用**——开发机此次 push 仍只到 GitHub（保持现有 origin master 链路），bare repo / launchd / newsyslog / /Volumes/large 等基础设施都未架设。

## 2026-05-26 10:55 +0800
- 进度：使用 `git filter-repo --invert-paths` 把历史中误提交的 `android-app/.gradle-local/`（12 823 文件 / 617 MB）与 `android-tv-app/tv-app/build/`（6 184 文件 / 122 MB）整体擦除。两组路径均由"首次提交"与"数分钟后的清理 commit"配对引入（507f34b1 + ef98bb72 / 782798c6 + e0175d60），git 保留 blob 历史导致 push pack 一直膨胀到 742.79 MiB；重写后 pack 降至 9.73 MiB（缩减 98.7%）。force-push 前用 patch-id 对比确认远端独有的 47 个 commit 内容已全部存在于本地（hash 不同但 patch 一致，是历史 rebase 留下的孤儿），无工作丢失。`.worktrees/` 下 6 个 feature 分支与 2 个 detached worktree 在重写后均仍是新 master 的祖先。同步移除根目录误生成的空壳 `package-lock.json`，并在 `.gitignore` 增加 `/package-lock.json` 与 `/node_modules/`（带前导斜杠仅锚定根，不影响 `admin-web/`）。
- 影响文件：所有 commit hash 重写（git filter-repo）；`.gitignore`（+3 行）；删除根目录 `package-lock.json`；本地 backup 分支 `backup/pre-filter-repo` 已被 filter-repo 一并重写，原始未重写历史只在 GitHub 的旧 `origin/master = 1898accd`（force-push 时被覆盖）以及 reflog 清空前的窗口。`plan.md`。
- 验证：`git ls-files | grep -E '(\.gradle-local|tv-app/build)'` 为空；`git count-objects -vH` 显示 `size-pack: 9.73 MiB`（重写前 742.79 MiB）；`git log --all --diff-filter=AM --name-only --format= | grep -E "^(android-app/\.gradle-local|android-tv-app/tv-app/build)/"` 无残留；force-push 后 `git log --oneline -1` 本地与 `origin/master` 同步在 `eb047a3`。所有其它本地 clone / worktree 需要 `git fetch && git reset --hard origin/master` 才能跟上。

## 2026-05-25 16:41 +0800
- 进度：开始实施 `tasks/2026-05-25-tv-long-form-libvlc-migration`。当前仅该任务未完成；已按 PRD → Implement → Review 读取任务范围。先校准代码现状：TV 工程通过 source exclude 排除短视频/通用播放器源码，长视频仍在 `LongFormVideoPlayer` / `TvLongFormPlayerScreen` / `TvSeriesPlayerScreen` 使用 Media3；后端 `subtitle.go` 仍将 ASS/SSA 转 VTT；`pkg/ffmpeg` 只有 WebVTT 抽取。接下来按 TDD 先补后端 ASS 原文策略、LibVLC 配置/源文审计红灯，再做最小迁移。
- 影响文件：预计涉及 `internal/services/subtitle.go`、`pkg/ffmpeg/ffmpeg.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、TV/后端测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`、`plan.md`。
- 验证：待执行红灯测试、`go test ./internal/services -run TestSubtitle`、`go test ./pkg/ffmpeg -run Test.*Subtitle`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、androidTest 编译、diff/乱码扫描；真机回归需要设备环境后记录。

## 2026-05-25 16:59 +0800
- 进度：完成 LibVLC 迁移首轮代码落地：后端 ASS/SSA 上传改为原文落 `format=ass` / `mime_type=text/x-ssa`，内嵌 ASS/SSA 抽取改用 `ffmpeg -c:s copy`；TV 长视频新增 `TvVlcLibrary` 单例、长视频 LibVLC Media 构造、TextureView Surface、by-language 选轨 helper；`LongFormVideoPlayer` / 电影播放页 / 电视剧播放页从 ExoPlayer 改为 `org.videolan.libvlc.MediaPlayer`，保留原 Compose 控制层、续播、连播、遥控器 UI 状态机；删除长视频字幕样式测试与 TV Gradle 的 Media3 依赖，TV 版本号升到 `0.1.68`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/{model,player,ui}/**`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/{detail,tv}/**`、TV 单测/androidTest、`internal/services/subtitle.go`、`internal/services/{subtitle,transcode}_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`plan.md`。
- 验证：红灯已确认：新增 Go 测试最初因 `embeddedSubtitleOutputPlanForProbe` / `buildExtractSubtitleToAssArgs` 缺失失败，TV 定向测试最初因 LibVLC helper 缺失失败；实现后 `go test ./pkg/ffmpeg -run Test.*Subtitle -count=1` 通过、`go test ./internal/services -run TestSubtitle -count=1` 通过、`cd android-tv-app && ./gradlew --no-daemon :tv-app:compileDebugKotlin` 通过。`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests ...` 当前失败在远端 Maven `com.google.guava:listenablefuture:1.0` TLS 握手，非代码编译错误；后续需在依赖可解析后重跑全量 TV 单测/build。

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

## 2026-05-25 13:40 +0800
- 进度：开始排查管理端视频管理列表“详情”按钮点击无反应。初步定位到 `VideoList.vue` 的详情请求 token 链路：`showDetail()` 设置 `detailRequestToken` 后立即调用会再次递增 token 的 `handleDetailClose()` / `resetDetailState()`，导致 `getAdminVideoDetail()` 返回后被当作过期请求丢弃；`refreshPlayURL()` 内也存在调用 `handleDetailClose()` 让 expectedToken 失效的同类问题。计划先补 token 纯函数红灯测试，再做最小修复并跑 admin 定向测试与 build。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`plan.md`
- 验证：待执行 `cd admin-web && npm test -- src/views/videoList.helpers.spec.js`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。当前工作区无其他未提交改动。

## 2026-05-24 23:20 +0800
- 进度：完成 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote` 核心实现：新增 TV 遥控键路由纯函数、左上信息层组件与数据构造；`LongFormVideoPlayer` 的 TV 模式改为首次亮 UI 但焦点停根、DOWN 进入 controls、LEFT/RIGHT seek 并重置计时、controls 内左右走 focusProperties 首尾环绕、Slider 不可聚焦、BACK 可见时优先收 UI；电视剧播放器传入剧名/季集/单集标题，左上主行显示剧名。同步 TV 版本号升到 `0.1.66`，`CONTEXT.md` 追加 TV 操作 UI 术语。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/{LongFormVideoPlayer,TvLongFormRemoteKeyRouting,TvLongFormTitleOverlay}.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、TV 单测与 androidTest、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-24-tv-long-form-operation-ui-on-remote/DONE.md`
- 验证：新增测试红灯曾因 `TvRemoteKeyAction` / `buildTvLongFormTitleOverlayData` 未实现失败；实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlayDataTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:connectedDebugAndroidTest` 已编译并打包 androidTest，但本机无连接设备，失败于 `DeviceException: No connected devices!`；`git diff --check` 通过；乱码扫描无输出。无关工作区改动 `.env.example`、`internal/config/config.go` 未纳入本任务。

## 2026-05-24 22:40 +0800
- 进度：开始执行 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote`。按 PRD → Implement → Review 推进 TV 长视频播放器遥控器互动操作 UI：先补纯函数与源文审计红灯测试，再接入 `LongFormVideoPlayer` 的 TV 模式键路由、左上信息层、controls 焦点入口/环绕、BACK 优先收 UI，并同步 TV 调用方、版本号、`CONTEXT.md` 与任务完成标记。当前仅 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote/` 为未完成任务目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLongFormRemoteKeyRouting.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLongFormTitleOverlay.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、TV 单测、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`、任务目录。
- 验证：待执行红灯测试、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、diff/乱码扫描；`connectedDebugAndroidTest` 视本机设备可用性执行。

## 2026-05-24 22:12 +0800
- 进度：根据 `$grill-with-docs` 决议，取消 VideoUpload 三段式上传向导，回到保留新设计系统外壳的单屏上传表单；文件选择、基础信息、关联信息、上传控制、进度与结果都在同一画面内。同步删除不再使用的 step 子组件与 wizard helper/spec，并将 `CONTEXT.md` 术语从 `admin 上传向导三步` 改为 `admin 上传单屏表单`。按用户要求，本轮不改 Phase 4 任务 PRD / implement / review 历史文档与截图，修改完成后补 `DONE.md`。
- 影响文件：`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`admin-web/src/views/videoUpload.wizard.helpers.*`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-three-pillars/DONE.md`
- 验证：待执行 `cd admin-web && npm run build`、`cd admin-web && npm test`、`git diff --check`、乱码扫描。

## 2026-05-24 22:00 +0800
- 进度：完成 `tasks/2026-05-23-admin-three-pillars/feedback.md` follow-up 修复：VideoList required 列迁移、批量删除 loading、drawer destroy-on-close、异步 token 防旧请求回写、字幕列表纳入 dirty；ImageManage 批量启停/删除改为 allSettled 并总是刷新/清选、默认 active chip 语义修正、视图切换清空选中；VideoUpload 切 movie 前确认丢弃多余文件，StepRelate 移除重复「开始上传」。同步扩展 BulkActionBar action 的 loading/disabled 支持，并在 CONTEXT 沉淀 drawer snapshot 与批量操作上下文切换契约。
- 影响文件：`admin-web/src/components/base/BulkActionBar.vue`、`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/StepRelate.vue`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-three-pillars/feedback.md`
- 验证：`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/views/videoUpload.wizard.helpers.spec.js` 通过；`cd admin-web && npm test` 通过（13 files / 95 tests）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-three-pillars` 无输出。

## 2026-05-24 21:55 +0800
- 进度：开始处理 `tasks/2026-05-23-admin-three-pillars/feedback.md`。反馈覆盖 Phase 4 follow-up：VideoList 列设置 required 迁移、BulkActionBar loading/disabled、drawer destroy/竞态/字幕 dirty；ImageManage 批量操作 allSettled 与视图切换清选；VideoUpload movie 类型切换丢文件确认、去掉重复开始上传按钮；同步必要 CONTEXT 与验证。
- 影响文件：`admin-web/src/components/base/BulkActionBar.vue`、`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-three-pillars/feedback.md`
- 验证：待执行 `cd admin-web && npm test -- src/assets/themeTokens.spec.js src/views/videoUpload.wizard.helpers.spec.js`、`cd admin-web && npm test`、`cd admin-web && npm run build`、`git diff --check`、乱码扫描。

## 2026-05-24 21:30 +0800
- 进度：完成 `tasks/2026-05-23-admin-three-pillars` Phase 4 实现与 review 自动化：ImageManage 接入 PageHeader / Toolbar / 视图切换 / 双 drawer / BulkActionBar / EmptyState 并清零旧玫红；VideoList 接入 chip 筛选、更多筛选 drawer、列设置 localStorage、编辑 drawer、BulkActionBar；VideoUpload 拆为三步 Wizard 与 3 个 step 子组件；3 条路由补 `hideShellPageHeader`，CONTEXT 新增 5 条 admin 术语，截图目录归档 11 张 1440x1080 PNG。未创建 `DONE.md`，待用户验收后按流程补完成标记。
- 影响文件：`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`admin-web/src/views/videoUpload.wizard.helpers.*`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-three-pillars/*`、`plan.md`
- 验证：`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/views/videoUpload.wizard.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`cd admin-web && npm test` 通过（13 files / 95 tests）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-three-pillars` 无输出；截图目录 11 张 PNG 均为 1440x1080。

## 2026-05-24 21:09 +0800
- 进度：开始执行 `tasks/2026-05-23-admin-three-pillars` Phase 4（三巨头 IA 改造）。前置 Phase 1-3 已有 `DONE.md`，本阶段按 PRD / Implement / Review 推进：先扩 audit 与 wizard helper 红灯，再改 ImageManage / VideoList / VideoUpload、补路由 meta、CONTEXT.md 术语、截图归档与验证。
- 影响文件：`admin-web/src/views/ImageManage.vue`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/VideoUpload/*`、`admin-web/src/views/videoUpload.wizard.helpers.js`、`admin-web/src/views/videoUpload.wizard.helpers.spec.js`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-three-pillars/screenshots/*`、`plan.md`
- 验证：待执行红灯测试、`cd admin-web && npm test`、`cd admin-web && npm run build`、截图归档、乱码扫描与 diff 检查。

## 2026-05-24 14:30 +0800
- 进度：完成 `tasks/2026-05-23-admin-medium-views` Phase 3 的视觉归档收尾，4 个中等视图已补齐新的 after 截图，其中 `ImageCollectionManage` 额外补了 drawer 打开形态。当前截图目录共 9 张，均为 1440x1080。
- 影响文件：`admin-web/src/views/ImageCollectionManage.vue`、`admin-web/src/views/ScrapePreview.vue`、`admin-web/src/views/AVManualScrape.vue`、`admin-web/src/views/TvSeriesManage.vue`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-medium-views/screenshots/*`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src tasks/2026-05-23-admin-medium-views` 未发现乱码。

## 2026-05-24 12:20 +0800
- 进度：开始执行 `tasks/2026-05-23-admin-medium-views` Phase 3（中等视图重排）。当前先按 PRD / Implement 对齐 4 个页面的壳子重排、ImageCollectionManage 的上传 drawer 试点、4 条路由的 `hideShellPageHeader`、`themeTokens.spec.js` 的 10 文件 audit、`CONTEXT.md` 的 2 条术语补充，以及 9 张截图归档。
- 影响文件：`admin-web/src/views/ImageCollectionManage.vue`、`admin-web/src/views/ScrapePreview.vue`、`admin-web/src/views/AVManualScrape.vue`、`admin-web/src/views/TvSeriesManage.vue`、`admin-web/src/router/index.js`、`admin-web/src/assets/themeTokens.spec.js`、`CONTEXT.md`、`tasks/2026-05-23-admin-medium-views/screenshots/*`、`plan.md`
- 验证：待执行定向重排、`cd admin-web && npm test`、`cd admin-web && npm run build`、截图归档、乱码扫描与 diff 检查。

## 2026-05-24 11:55 +0800
- 进度：根据 `tasks/2026-05-23-admin-simple-views/feedback.md` 补修 Phase 2 反馈项：TaskMonitor 状态筛选与后端契约对齐、UserManage 改为后台原子建用户接口、Dashboard ECharts 实例重挂载修复，并收口 shell 顶栏与视图页头的双标题问题。
- 影响文件：`internal/handlers/admin.go`、`internal/repository/admin_repository.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/UserManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/Dashboard.vue`、`admin-web/src/components/Layout.vue`、`admin-web/src/router/index.js`、`plan.md`
- 验证：待执行后端/前端定向测试、`npm test`、`npm run build`、乱码扫描与 diff 检查。

## 2026-05-24 11:46 +0800
- 进度：完成 `tasks/2026-05-23-admin-simple-views` Phase 2 的实现与截图归档，7 个简单视图已按 Phase 1 设计系统重排，`themeTokens.spec.js` 扩展的视图层 audit 通过，14 张 before/after 截图已写入任务目录。当前先保留任务文档本身与实现代码、截图及计划记录，待用户确认验收后再按仓库流程补 `DONE.md`。
- 影响文件：`admin-web/src/views/Dashboard.vue`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/UserManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/IPTVManage.vue`、`admin-web/src/views/CollectionManage.vue`、`admin-web/src/views/ActorManage.vue`、`admin-web/src/components/UploadProgress.vue`、`admin-web/src/api/admin.js`、`admin-web/src/assets/themeTokens.spec.js`、`tasks/2026-05-23-admin-simple-views/screenshots/*`、`plan.md`
- 验证：`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' admin-web/src admin-web/src/assets/themeTokens.spec.js tasks/2026-05-23-admin-simple-views/screenshots plan.md` 未发现乱码；`tasks/2026-05-23-admin-simple-views/screenshots/` 已包含 14 张 PNG，统一 1440x1080。

## 2026-05-24 10:38 +0800
- 进度：开始执行 `tasks/2026-05-23-admin-simple-views` Phase 2，目标是把 7 个简单视图按 Phase 1 设计系统重排，并清零视图层 scoped CSS 里的玫红 hex。当前已确认 Phase 1 外壳已完成且 `DONE.md` 已存在；本阶段只处理 Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage / UploadProgress 及其 audit 测试。
- 影响文件：`admin-web/src/views/Dashboard.vue`、`admin-web/src/views/SystemSettings.vue`、`admin-web/src/views/UserManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/views/IPTVManage.vue`、`admin-web/src/views/CollectionManage.vue`、`admin-web/src/views/ActorManage.vue`、`admin-web/src/components/UploadProgress.vue`、`admin-web/src/assets/themeTokens.spec.js`、`plan.md`、`tasks/2026-05-23-admin-simple-views/screenshots/*`
- 验证：待执行红灯测试、`npm test`、`npm run build`、截图归档、乱码扫描与提交前 diff 检查。

## 2026-05-24 10:31 +0800
- 进度：完成 `tasks/2026-05-23-admin-shell-redesign` Phase 1 的实现、定向测试、全量 `npm test`、`npm run build`、截图归档与乱码扫描。admin-web 已切换到新的浅色设计 token、Element Plus 覆写、分组侧栏、命令面板、profile chip、独立登录页与共享基础组件；同时把 `Dashboard` / `IPTVManage` / `TaskMonitor` 的 `--font-code` 收口到 `--font-mono`，并在 `CONTEXT.md` 追加「admin 设计系统术语」。
- 影响文件：`admin-web/src/assets/theme.css`、`admin-web/src/assets/element-overrides.css`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/base/*`、`admin-web/src/views/Login.vue`、`admin-web/src/views/Dashboard.vue`、`admin-web/src/views/IPTVManage.vue`、`admin-web/src/views/TaskMonitor.vue`、`admin-web/src/main.js`、`admin-web/index.html`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-admin-shell-redesign/screenshots/*`
- 验证：`cd admin-web && npm test -- src/assets/themeTokens.spec.js src/components/Layout.spec.js src/components/base/commandPalette.helpers.spec.js` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size warning）；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src admin-web/index.html tasks/2026-05-23-admin-shell-redesign` 未发现乱码；截图已归档为 `before-*.png` / `after-*.png`。

## 2026-05-24 10:04 +0800
- 进度：开始执行 `tasks/2026-05-23-admin-shell-redesign` 的 Phase 1：重做 admin-web 设计系统、外壳与登录页，并补齐命令面板与 7 个共享 wrapper 的基础设施。当前工作区只有该任务目录是相关未跟踪内容，其余 `tasks/2026-05-23-admin-*` 目录先视为并行任务。
- 影响文件：`admin-web/src/assets/theme.css`、`admin-web/src/assets/element-overrides.css`、`admin-web/src/components/Layout.vue`、`admin-web/src/components/base/*`、`admin-web/src/views/Login.vue`、`admin-web/src/main.js`、`admin-web/index.html`、`CONTEXT.md`、`plan.md`
- 验证：待执行红灯测试、`npm test`、`npm run build`、截图/手测与乱码扫描。

## 2026-05-24 09:53 +0800
- 进度：完成 TV 首页货架文案收口：三类内容页的货架标题下不再显示“最近更新”，各区块的“查看更多”卡也不再展示数量副文案；首页货架语义收束为「最近播放 / 最近更新」两类，`TvCatalogScreen` 仅保留纯标题呈现。TV 端版本递增到 `versionCode = 66`、`versionName = "0.1.65"`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest.homeShelvesDoNotShowSubtitleCopyUnderTheTitle'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 未发现乱码。未纳入未跟踪 `tasks/2026-05-23-admin-*` 目录。

## 2026-05-24 09:46 +0800
- 进度：开始收紧 TV 首页货架文案。当前首页三类内容页仍在 `电视剧 / 电影 / 18+` 货架标题下显示“最近更新”，且 `查看更多` 卡还显示“共 x 项”；根据确认后的边界，本次只保留两类货架语义（最近播放 / 最近更新），去掉货架标题下所有说明文案与统计文案，不改首页分类模型与内容排序。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待执行红灯测试、TV 单测、TV 构建、乱码扫描。

## 2026-05-24 09:29 +0800
- 进度：完成 TV 设置页「电视剧自动连播」抗挤压修复：设置行改为最小高度，左侧文案区弹性收缩并单行省略，右侧 Switch 固定 64dp 操作占位；补充静态回归测试锁定该布局契约。TV 端版本递增到 `versionCode = 65`、`versionName = "0.1.64"`，并在 `CONTEXT.md` 沉淀 `TV 设置行抗挤压布局`。未纳入未跟踪 `tasks/2026-05-23-admin-*` 目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest.seriesAutoplaySettingRowProtectsSwitchFromTextCompression'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java android-tv-app/tv-app/src/test/java android-tv-app/tv-app/build.gradle.kts` 未发现乱码。

## 2026-05-24 09:24 +0800
- 进度：开始修复 TV 设置页「电视剧自动连播」开关行严重挤压问题。经 `$grill-with-docs` 对齐，术语采用既有 `电视剧自动连播`，范围为 TV 首页设置面板「播放设置」分组；用户确认保持“一行左文案 + 右开关”形态，修复目标是让文案区可弹性收缩且 Switch 固定占位，不拆成大卡片。当前工作区存在未跟踪 `tasks/2026-05-23-admin-*` 目录，视为无关用户工作，本次不纳入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 定向红灯、TV 单测 / 构建、乱码扫描。

## 2026-05-23 23:45 +0800
- 进度：完成 ASS/SSA 外挂字幕上传支持：后端上传计划新增 `.ass/.ssa`，上传后先落临时源文件再通过 ffmpeg 转成 WebVTT，最终轨道仍以 `vtt` / `text/vtt` 暴露给手机端和 TV 端；metadata 记录 `original_filename` 与 `original_format`。管理端字幕上传选择器改用 `subtitleUploadAccept`，允许 `.srt,.vtt,.ass,.ssa`。已在 `CONTEXT.md` 沉淀“外挂 ASS/SSA 字幕上传策略”。未纳入未跟踪的 `tasks/2026-05-23-admin-*` 目录。
- 影响文件：`internal/services/subtitle.go`、`internal/services/transcode_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `go test ./pkg/ffmpeg -run TestBuildConvertSubtitleToWebVTTArgs -count=1` 先失败于 `undefined: buildConvertSubtitleToWebVTTArgs`；实现后 `go test ./pkg/ffmpeg -run 'TestBuildConvertSubtitleToWebVTTArgs|TestParseSubtitleProbeOutput' -count=1` 通过，`go test ./internal/services -run 'TestSubtitleUploadPlanForFilename' -count=1` 通过，`cd admin-web && npm test -- videoList.helpers.spec.js` 通过；收口验证 `go test ./pkg/ffmpeg ./internal/services ./internal/handlers ./internal/repository -run 'Subtitle|VideoSubtitle|BuildConvertSubtitle|ParseSubtitle|Transcode' -count=1` 通过，`cd admin-web && npm test` 通过，`cd admin-web && npm run build` 通过（仅既有 chunk size warning），`go test ./... -count=1` 通过，`go vet ./...` 通过；待执行 diff/乱码检查。

## 2026-05-23 23:38 +0800
- 进度：开始实现 ASS/SSA 外挂字幕上传支持。方案：后端允许 `.ass/.ssa` 上传后统一转换为 WebVTT 存储与播放，保留原始文件名 / 原始格式到 metadata；管理端上传选择器增加 `.ass,.ssa`；客户端继续消费既有 VTT 字幕轨，不改 Android / TV 播放器。当前工作区存在未跟踪 `tasks/2026-05-23-admin-*` 目录，视为无关用户工作，本次不纳入。
- 影响文件：`internal/services/subtitle.go`、`pkg/ffmpeg/ffmpeg.go`、相关 Go 测试、`admin-web/src/views/VideoList.vue`、相关前端测试、`CONTEXT.md`、`plan.md`
- 验证：待执行红灯测试、Go 定向测试、admin-web 定向测试、必要构建、乱码扫描。

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

## 2026-05-23 21:54 +0800
- 进度：开始实现 `tasks/2026-05-23-tv-resume-from-history-prompt/`。按 PRD/Implement/Review 顺序执行，先补 `TvResumePromptTest` 与 `TvResumePromptCardSpecTest` 红灯，再接入 TV 电影 / `18+` / 电视剧播放器续播提示卡，最后更新 TV 版本号并跑 review 要求验证。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/`
- 验证：待执行定向红灯、`cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew :tv-app:assembleDebug`、乱码扫描。

## 2026-05-23 21:49 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，修正 `review.md` 源文 audit 第 3 条的文件定位：`Alignment.BottomStart` 的断言应落在两个 player 文件的调用点，而不是 `TvResumePromptCard.kt`。此处仅做文档口径收紧，未改实现。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：已对照 `implement.md` 的 audit 列表；待后续统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:49 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认提交策略拆分为“grill 审查文档独立提交 + 后续实现阶段代码独立提交”；review 的 Done definition 改为要求实现改动集中在一个 commit，不再要求任务定义修订和实现混在同一提交。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:41 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认本任务保持续播专用位置 token，不抽共享播放器提示卡 token、不改连播卡；同时把 `LeftPaddingDp` 收紧命名为 `HorizontalPaddingDp`，表达边缘水平安全距离而非只能用于左侧。已同步 `implement.md` 与 `review.md`。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:33 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认倒计时状态机与 UI 可见性分离：新增 `shouldTickResumePromptCountdown` 作为 tick 条件，`shouldShowResumePromptCard` 只管渲染，避免 remainingMs 归零时漏设永久 dismiss。已同步 `implement.md` 与 `review.md`。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:27 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认 `CONTEXT.md` 只沉淀去实现细节的 glossary，不照搬 PRD 中带函数名、尺寸和 API 约束的术语表。已同步 `prd.md`、`review.md`，并在 `CONTEXT.md` 追加 `续播提醒`、`续播提示卡`、`续播倒计时`、`续播提示卡触发节流`、`续播提示卡永久 dismiss`、`续播从头播放语义` 六条术语。
- 影响文件：`CONTEXT.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:20 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认续播卡是非阻塞提示：焦点从「继续观看」移到「从头播放」不冻结倒计时；倒计时归零前未按 OK 即默认继续观看并 dismiss。已同步 `prd.md`、`implement.md`、`review.md`。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:14 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认续播倒计时只在续播卡可见且可操作时消耗；暂停和退出确认这类临时隐藏都冻结剩余时间。已同步 `prd.md`、`implement.md`、`review.md`，并要求倒计时 `LaunchedEffect` 依赖可见态而非只看暂停状态。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:08 +0800
- 进度：继续 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，确认本任务不改 `LongFormVideoPlayer` 公共 API；字幕/音轨夜台面板不进入续播卡状态机，“选择面板”收紧为电视剧选集面板。已同步 `prd.md`、`implement.md`、`review.md`，并把 guard 字段改名为 `isEpisodeSelectorVisible` 以避免泛化误读。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/review.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 21:04 +0800
- 进度：进入 `$grill-with-docs` 审查 `tasks/2026-05-23-tv-resume-from-history-prompt/`，先解决续播卡与 `TV 播放器退出确认` 的语义冲突：BACK/退出确认只临时隐藏续播卡，不触发永久 dismiss；若确认提示消失且倒计时仍有剩余，续播卡恢复显示。已同步 `prd.md` 与 `implement.md`。
- 影响文件：`tasks/2026-05-23-tv-resume-from-history-prompt/prd.md`、`tasks/2026-05-23-tv-resume-from-history-prompt/implement.md`、`plan.md`
- 验证：待继续审查后统一执行 Markdown 乱码扫描与 diff 检查。

## 2026-05-23 20:57 +0800
- 进度：根据用户确认，将 `tasks/2026-05-23-western-av-oshash-confirm-gate/` 标记为已完成。新增 `DONE.md` 记录完成时间、关联提交与验证摘要；后续批量处理 tasks 时默认跳过该任务目录，除非用户明确要求重开或复查。
- 影响文件：`tasks/2026-05-23-western-av-oshash-confirm-gate/DONE.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本次仅为任务状态文档标记，不需要重新构建。

## 2026-05-23 19:28 +0800
- 进度：完成 `tasks/2026-05-23-western-av-oshash-confirm-gate/feedback.md` 的必修收口：ThePornDB 两位年份日期补 `20`、`/scenes` 恢复 `parse=` 且 `/movies` 保持 `q=`、补 oshash 256KiB `0x42` 黄金值测试；同时补 ConfirmAV 前置状态注释、OUMEI_NAME 来源注释、migration down 回滚提示，并将“含 DONE 的 task 不按 feedback 返工”沉淀到 `CONTEXT.md`。未纳入 `.codex/skills/av-scraper-optimization` 删除、`.claude/`、OpenSpec skill、`CLAUDE.md`、`package-lock.json`、两个 task `feedback.md` 等既有无关工作区变更。
- 影响文件：`CONTEXT.md`、`internal/handlers/admin_scrape.go`、`internal/queue/scrape_tasks_test.go`、`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_theporndb_test.go`、`internal/services/scraper_test.go`、`migrations/0021_western_av_oshash_gate.down.sql`、`pkg/oshash/oshash_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestThePornDBSearchKeywordsExpandsSupportedDateFormats|TestThePornDBSearchURLUsesEndpointSpecificQueryParameter' -count=1` 通过；`go test ./pkg/oshash -run 'TestComputeMatchesPythonOshashGoldenFixture|TestComputeReturnsDeterministicHex|TestComputeReturnsErrFileTooSmall' -count=1` 通过；`go test ./pkg/oshash ./internal/services ./internal/queue ./internal/handlers -run 'ThePornDB|OSHash|AVScrape|ConfirmAV|SkipScrape' -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal pkg migrations` 无输出。

## 2026-05-23 19:07 +0800
- 进度：开始按当前未完成 task 的 feedback 修复 `tasks/2026-05-23-western-av-oshash-confirm-gate`；用户确认含 `DONE.md` 的 task 视为已完成，不再按其 feedback 返工，因此跳过 `tasks/2026-05-23-tv-series-autoplay-next-episode/feedback.md`。先补 ThePornDB 日期 / 查询参数与 oshash 黄金值的定向回归测试，再做最小实现，并把 DONE 跳过 feedback 的约定沉淀到 `CONTEXT.md`。
- 影响文件：`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/*_test.go`、`pkg/oshash/oshash_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待执行 Go 定向测试、`go test ./...`、`go vet ./...`、乱码扫描。

## 2026-05-23 16:50 +0800
- 进度：已提交 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的实现收口，关联提交 `ac7766f2`（`完成欧美 AV 刮削确认门控`）。按任务 DONE 标准，本轮不创建 `DONE.md`；需用户完成 B/C 手动验收后再标记完成。
- 影响文件：`plan.md`
- 验证：提交前 `go test ./...`、`go vet ./...`、`cd admin-web && npm test`、`cd admin-web && npm run build` 均通过。

## 2026-05-23 16:49 +0800
- 进度：完成 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的实现收口：欧美 AV 上传自动刮削落 `av_scrape_pending` 并写入 `scrape_preview` / `scrape_attempt`，确认或弃刮后通过 `force=true` 转码入队；ThePornDB 成功响应改为完整 JSON decode，修复 detail body 被 512B 截断导致候选丢失；admin-web 增加 `欧美 AV 待确认` 状态、待确认面板、弃刮入口和 `hash 命中` 徽章，AV 手动刮削能直接加载待确认候选。
- 影响文件：`internal/services/scraper.go`、`internal/services/scraper_av_framework.go`、`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_strategy.go`、`internal/queue/scrape_tasks.go`、`internal/queue/tasks.go`、`internal/handlers/admin_scrape.go`、`internal/handlers/router.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/AVManualScrape.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./pkg/oshash ./internal/repository ./internal/services ./internal/queue ./internal/handlers -count=1` 通过；`go test ./...` 通过；`go vet ./...` 通过；`cd admin-web && npm test` 通过；`cd admin-web && npm run build` 通过（仅保留既有 chunk size warning）。

## 2026-05-23 15:53 +0800
- 进度：开始落实 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的实现阶段，先补数据层与上传透传：修复 `GetVideoOSHash` 接口编译问题、增加 `0021` 迁移、落 `pkg/oshash`、把 `site_category` 和 `os_hash` 从上传链路透传到落库。
- 影响文件：`migrations/0021_western_av_oshash_gate.up.sql`、`migrations/0021_western_av_oshash_gate.down.sql`、`pkg/oshash/oshash.go`、`pkg/oshash/oshash_test.go`、`internal/models/models.go`、`internal/repository/video_repository.go`、`internal/services/upload.go`、`internal/services/chunk_upload.go`、`internal/handlers/upload.go`、`internal/handlers/upload_chunk.go`、`admin-web/src/views/VideoUpload.vue`、`plan.md`
- 验证：待执行 `go test ./internal/repository -run 'OSHash|Video' -v`、`go test ./pkg/oshash -v`、`go test ./internal/services -run 'SaveUpload|SaveUploadedFile|OSHash' -v`

## 2026-05-23 15:35 +0800
- 进度：按 task 审查结果收紧 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的第三个口径：删除固定“5 秒内”的验收写法，改成自动刮削任务完成后的状态型验收；同步修正 `prd.md` / `review.md`。
- 影响文件：`tasks/2026-05-23-western-av-oshash-confirm-gate/prd.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/review.md`、`plan.md`
- 验证：文档改动，三处口径已收紧

## 2026-05-23 15:32 +0800
- 进度：按 task 审查结果进一步收紧 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的第二个口径：`av_scrape_pending` 的确认/弃刮门控与 `ready` 视频的手动重刮拆成两条不同流程，候选可复用但状态语义不复用；同步把 `CONTEXT.md`、`implement.md` 对齐。
- 影响文件：`CONTEXT.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/implement.md`、`plan.md`
- 验证：文档改动，待继续审查第三个口径

## 2026-05-23 15:29 +0800
- 进度：按 task 审查结果收紧 `tasks/2026-05-23-western-av-oshash-confirm-gate` 的第一个口径：`AV 地区分类` 允许空值，后端默认按 `japanese` 处理，不再写成“必填但又默认日本”的双重语义；同步把 `CONTEXT.md`、`prd.md`、`implement.md` 里的表述对齐。
- 影响文件：`CONTEXT.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/prd.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/implement.md`、`plan.md`
- 验证：文档改动，待继续审查其余两处口径

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

## 2026-05-23 11:34 +0800
- 进度：继续实现 `tasks/2026-05-23-tv-series-autoplay-next-episode`，优先补齐 ViewModel 连播链路、自动切分流与开关状态，再推进 UI / 设置 / 验证。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'`

## 2026-05-23 11:23 +0800
- 进度：开始执行 `tasks/2026-05-23-tv-series-autoplay-next-episode/` 的实现。已完成 PRD / Implement / Review 对齐，下一步按 TDD 先补 `TvSeriesAutoplaySpecTest.kt` 与 `TvAutoplayPromptCardSpecTest.kt` 的红灯，再落地 `TvSeriesAutoplay.kt`、播放器自动切分流、提示卡和覆盖层。
- 影响文件：`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesAutoplaySpecTest` 等定向红灯验证。

## 2026-05-23 11:04 +0800
- 进度：按审查结论同步修正 `tasks/2026-05-23-tv-series-autoplay-next-episode/` 的 PRD / Implement / Review：补入 `手动下一集按钮` 术语，明确提示卡要避开底部控制条、倒计时按播放器位置推导且显示整秒，并把手动下一集与自动连播放行。后续实现时可直接以这版任务文档为准。
- 影响文件：`tasks/2026-05-23-tv-series-autoplay-next-episode/prd.md`、`tasks/2026-05-23-tv-series-autoplay-next-episode/implement.md`、`tasks/2026-05-23-tv-series-autoplay-next-episode/review.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务文档修订，不涉及构建。

## 2026-05-23 10:48 +0800
- 进度：继续审查电视剧自动连播任务，确认控制条“下一集”按钮属于手动切集语义，保留现有 `tvPlaybackHistorySnapshot` 判断，不强制 `completed=true`；只有连播提示卡和倒计时归零走自动完成语义。已新增 `手动下一集按钮` 术语，后续实现需要把手动和自动两条路径分开调用。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:43 +0800
- 进度：继续审查电视剧自动连播任务，确认 `连播链路` 的兜底排序只允许接口原始顺序，不再引入 `lastWatchedAt` 等额外历史排序键。用户已确认该口径，已补入 `CONTEXT.md`，后续实现只需把“编号排序 + 原始顺序兜底”写成纯函数。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:41 +0800
- 进度：继续审查电视剧自动连播任务，确认连播提示卡的数字显示应按剩余整秒 `ceil(remainingMs / 1000)` 取整，只显示 10 到 1。用户已确认该口径，已写入 `连播倒计时窗口` 术语，后续实现应避免显示小数秒或 0。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:40 +0800
- 进度：继续审查电视剧自动连播任务，确认倒计时应直接由播放器 `position` / `duration` 推导，不维护独立墙钟时间轴。用户已确认该口径，已写入 `连播倒计时窗口` 术语，后续实现应避免引入 `startedAtMs` / `pausedDurationMs` 之类第二套时钟。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:38 +0800
- 进度：继续审查电视剧自动连播任务，确认自动切上报应保留切换瞬间的真实 `watchSeconds`，同时显式标记 `completed=true`，不把位置伪造成整集时长。用户已确认该口径，已写入 `连播自动切上报` 术语，后续实现应复用切换时的真实历史快照。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:35 +0800
- 进度：继续审查电视剧自动连播任务，确认“下一集”链路的切集语义应从头开始，不沿用目标集的历史续播位置；只有入口进入、继续观看或手动选集才保留历史进度。用户已确认该口径，已写入 `连播链路` 术语，后续实现必须区分“连播切集”和“历史续播”两条路径。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:31 +0800
- 进度：继续审查电视剧自动连播任务，确认连播提示卡几何应优先保证不与 `LongFormVideoPlayer` 底部控制条重叠，而不是死守底距 48dp。用户确认提示卡右侧保持 48dp，底部位置避开控制条安全区，控制条不可见时可退回底距 48dp。已修正 `连播提示卡` 术语，后续实现和 review 脚本应按“不重叠优先”验收。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:30 +0800
- 进度：继续审查电视剧自动连播任务，确认“退出 T-10 窗口复位提示触发状态”不等同于撤销用户点过的「取消本次」。用户确认取消状态绑定当前播放目标，seek 出 T-10 再进也不重新弹连播提示；只有切到下一集、手动选择其他集或离开当前播放目标后才清除。已修正 `取消本次连播` 术语，后续实现应把提示触发状态与取消状态拆成两个独立状态。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

## 2026-05-23 10:28 +0800
- 进度：继续按 `grill-with-docs` 审查电视剧自动连播任务，发现 `CONTEXT.md` 中“单次集内仅触发一次”与 PRD E2 的 seek 出 T-10 后再次进入应重新提示互相冲突。用户确认采用 PRD E2 口径：退出 T-10 窗口即重置本集提示触发状态，再次自然进入 T-10 可重新出现；切到下一集也复位。已修正 `连播倒计时窗口` 术语。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；本轮为任务审查期间的术语澄清，不涉及构建。

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

## 2026-05-23 02:21 +0800
- 进度：继续修复搜索页短视频全屏后仍露出根底部 tabbar。根据用户截图和现有代码确认，搜索页 `ShortSearchPlayerOverlay` 内部进入了全屏二选一渲染，但 `ShortSearchScreen` 没有把全屏状态回传给 `VideoHomeApp`，导致根 `Scaffold.bottomBar` 仍按 `search` 根 tab 显示。下一步先新增结构性红灯测试锁定搜索页必须向应用壳回传全屏状态，再实现最小回传链路，并收掉搜索浮层全屏态自身的状态栏 padding。
- 影响文件：预计 `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest`。

## 2026-05-23 02:08 +0800
- 进度：同步修复除首页外的另外三处短视频全屏入口。用户确认首页可用后指出搜索、发现、UnifiedPlayer 短视频分支仍未改；新增结构性红灯测试覆盖 `ShortSearchScreen`、`ShortDiscoverScreen`、`UnifiedPlayerScreen` 必须全屏/竖屏二选一渲染。实现后，三处都在 `isFullscreen` / `isShortFullscreen` 为 true 时只渲染 `ShortOverlayFullscreenHost`，非全屏时才渲染竖屏 `VerticalPager`、操作栏、关闭按钮和短视频进度条，避免两个 `PlayerView` 同时绑定同一 `ExoPlayer`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest` 先失败于 `all non home short overlays hide vertical pager while fullscreen`；实现后同一命令通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 初次失败于既有 TV 测试 `TvCatalogViewModelTest.nullListsInPayload_doNotCrashAndFallbackToEmpty` 的测试前协程异常，单独重跑该用例通过，随后全量 `:app:testDebugUnitTest` 复跑通过。

## 2026-05-23 01:52 +0800
- 进度：继续修复短视频全屏叠层混乱与退出后“有声音无画面”。根据用户截图确认，Dialog 全屏层只覆盖中间区域，竖屏短视频层仍在底下渲染，且两个 `PlayerView` 同时争用同一个 `ExoPlayer` surface。修法：撤掉 `ShortOverlayFullscreenHost` 的 Dialog 实现，改回同一 Compose 树内渲染；主页短视频全屏时采用二选一分支，只渲染 `ShortOverlayFullscreenHost`，不再同时渲染竖屏 `VerticalPager` 和竖屏 `PlayerView`。新增结构性红灯测试锁定“Host 不得使用 Dialog”和“主页短视频全屏必须隐藏竖屏 Pager”；`CONTEXT.md` 补充 PlayerView 独占渲染约束。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest` 先失败于 Dialog/叠层约束；实现后同一命令通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-23 01:36 +0800
- 进度：继续修复“首页短视频点击全屏后仍不像长视频真正全屏播放”。根因确认：此前只把播放器覆盖层提升为 Dialog，但首页短视频仍处于 `HomeScreen` 的头部内容 tab 与 `VideoHomeApp` 根 `Scaffold` 底部 tabbar 之间，外层壳没有进入全屏状态，所以视觉上仍残留头部和底部导航。新增红灯结构测试锁定 `VideoHomeApp` / `HomeScreen` 必须感知 `isShortFullscreen`；实现为 `ShortFeedScreen` 将全屏状态回传给 `HomeScreen` 和 `VideoHomeApp`，全屏期间隐藏首页头部 tab 与根底部 tabbar，退出/离开时回传 `false`。同时删除此前误加的 `DONE.md`，该任务待用户重新验收后再标记完成。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`CONTEXT.md`、`plan.md`、`tasks/2026-05-23-short-overlay-fullscreen-button/DONE.md`
- 验证：红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest` 先失败于 `short fullscreen state must hide the app shell chrome`；实现后同一命令通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-23 01:23 +0800
- 进度：沉淀 `tasks/` 完成标记约定。以后批量执行 `tasks/` 时，已包含 `DONE.md` 的任务目录默认视为完成并跳过；用户明确要求重开或复查时才重新处理。按用户确认测试完成的语义，为 `tasks/2026-05-23-short-overlay-fullscreen-button/` 新增 `DONE.md`，记录完成时间、关联提交和验证摘要。
- 影响文件：`AGENTS.md`、`CONTEXT.md`、`tasks/2026-05-23-short-overlay-fullscreen-button/DONE.md`、`plan.md`
- 验证：待执行 Markdown 乱码扫描与 diff 检查；文档规则变更无需构建。

## 2026-05-23 01:13 +0800
- 进度：完成手机端短视频浮层“全屏播放”任务收尾。共享 `ShortOverlayFullscreenHost` 已接入搜索、发现、主页短视频信息流和 `UnifiedPlayerScreen` 的短视频分支；`CONTEXT.md` 已补“短视频全屏播放”术语；`android-app/app/build.gradle.kts` 已按约定递增版本号。验证方面，手机端 `:app:testDebugUnitTest` 与 `:app:assembleDebug` 通过，TV 工程 `:tv-app:testDebugUnitTest` 也保持通过。ADB 已重新连接模拟器 `emulator-5554`，完成 `com.chee.videos` 安装与 `MainActivity` 启动确认，logcat 未见新的 `AndroidRuntime` 或 FATAL；由于当前设备侧不具备完整手测输入条件，本次仅记录到启动级现场校验。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`git diff --check` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-app/app/src/main/java android-app/app/src/test/java` 无输出；ADB `install -r` 与 `am start -n com.chee.videos/.MainActivity` 成功。

## 2026-05-23 00:56 +0800
- 进度：开始执行 `tasks/2026-05-23-short-overlay-fullscreen-button/`。已按沉淀规则先读 `prd.md`，确认目标是手机端四处短视频浮层新增“全屏播放”按钮，进入后强制横屏、隐藏系统栏、复用 `LongFormVideoPlayer`、临时强制 `REPEAT_MODE_ONE`，退出后恢复方向/系统栏/repeatMode；已读 `implement.md`，确认实现为共享 `ShortOverlayFullscreenHost` + `ShortOverlayFullscreenButton`，四处入口接入；已读 `review.md`，确认自动化验收包含 `:app:testDebugUnitTest`、`:app:assembleDebug`、`:tv-app:testDebugUnitTest`，真机手测项后续需说明是否已执行。下一步先按 TDD 新增 `ShortOverlayFullscreenSpecTest` 红灯测试，再实现共享 Host 和四处接入。
- 影响文件：预计 `android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`、四处短视频浮层、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md` 与新增单测。
- 验证：待红灯 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortOverlayFullscreenSpecTest`。

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

## 2026-05-22 23:39 +0800
- 进度：继续排查 TV 端登录后所有类型首页显示“加载失败”。ADB 已确认设备 `192.168.1.6:5555` 已登录并进入 `com.chee.videos.tv` 版本 `0.1.59` / `versionCode=60`；UIAutomator dump 显示右侧内容区为“加载失败 / TV 首页加载失败”，左侧 `电视剧`、`电影`、`18+` 等菜单正常。logcat 显示 `GET /api/v1/tv/home?kind=tv|movie|av&page=1&page_size=20` 全部返回 `200 OK`，说明不是登录态或服务端网络错误。Release R8 产物 `usage.txt` 明确列出 `TvHomePayload`、`TvHomeVideoDto`、`TvSectionDto`、`TvCatalogWallPayload`、`TvCatalogWallItemDto` 等 TV 首页/海报墙 DTO 被 shrink 处理；当前 `proguard-rules.pro` 只保留 `TvAuth*`，首页 DTO 未受保护。下一步先新增 R8 规则审计红灯测试，再把 Retrofit/Gson 反射模型保留规则扩展到 `core.model`，避免同类 Release-only DTO 裁剪再次发生。
- 影响文件：待修改 `android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：待红灯 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.network.TvAuthEnvelopeSpecTest`，实现后执行 TV 端单测与 Release 构建，并用 R8 `seeds.txt`/`usage.txt` 复核。

## 2026-05-23 00:00 +0800
- 进度：修复配对码生成 `ClassCastException`（版本 0.1.57 → 0.1.58，versionCode 58→59）。根因：Gson 泛型类型擦除问题——`ApiEnvelope<T>` 的 `data: T?` 字段在部分 Android TV 固件（老版 ART）上无法正确将类型参数 `T` 解析为具体类型 `TvAuthSessionCreatePayload` / `TvAuthSessionStatusPayload`，退化为 `LinkedTreeMap<String, Any>`；随后 Kotlin 编译器在 `requireEnvelope()` 返回值处插入的 CHECKCAST 字节码指令尝试将 `LinkedTreeMap` 强转为目标 payload 类，抛 `ClassCastException`，且部分老 ART 实现在这种情况下该异常的 `message` 为 null，触发上一次改动添加的"创建配对会话失败 (ExceptionClassName)"兜底串。修法：不改泛型 `ApiEnvelope<T>`（其他端点不受影响），仅针对两个实际消费 `data` 字段的 TV 认证端点，在 `ApiModels.kt` 新增两个具体非泛型包装类 `TvAuthCreateEnvelope`（`data: TvAuthSessionCreatePayload?`）与 `TvAuthStatusEnvelope`（`data: TvAuthSessionStatusPayload?`）；更新 `ApiService.kt` 的 `createTvAuthSession` / `getTvAuthSession` 返回类型由 `ApiEnvelope<TvAuthSessionCreatePayload>` / `ApiEnvelope<TvAuthSessionStatusPayload>` 改为对应的具体包装类；在 `TvAuthRepository.kt` 补充两个具体重载 `requireEnvelope(resp: TvAuthCreateEnvelope)` / `requireEnvelope(resp: TvAuthStatusEnvelope)`，Kotlin 在编译时静态选择正确重载，彻底消除运行时泛型推断。`callWithAuth`（approve / deny）保持 `ApiEnvelope<Map<String, Boolean>>` 不变，因该路径 `data` 结果被丢弃，不触发 CHECKCAST 问题。
- 影响文件：`core/model/ApiModels.kt`（新增 TvAuthCreateEnvelope / TvAuthStatusEnvelope）、`core/network/ApiService.kt`（两个 TV 认证端点返回类型）、`core/repository/TvAuthRepository.kt`（新增两个具体 requireEnvelope 重载 + 对应 import）、`build.gradle.kts`（版本号）、`plan.md`。
- 验证：待 `testDebugUnitTest` + `assembleDebug`。

## 2026-05-22 23:22 +0800
- 进度：确认并修复 TV 配对页 `ClassCastException` 的真实根因（版本 0.1.58 → 0.1.59，versionCode 59→60）。上次代码层把 TV 授权端点改成具体 envelope 后，debug 单测/源码字节码已正确，但用户安装的是 Release 形态 APK（设备拉回 `base.apk` 约 42MB，且本地 debug 包签名不匹配无法覆盖安装）。对比本地 Release R8 产物发现：未加规则时 `usage.txt` 将 `TvAuthCreateEnvelope` / `TvAuthStatusEnvelope` / `TvAuthSessionCreatePayload` / `TvAuthSessionStatusPayload` / `TvAuthSessionCreateRequest` 判定为可裁剪；这类模型只通过 Retrofit suspend 签名与 Gson 反射使用，Release R8 裁剪后设备运行时返回类型退化，最终仍触发 `ClassCastException`。修法：`proguard-rules.pro` 新增 `-keep class com.chee.videos.core.model.TvAuth* { *; }`，保留 TV 授权配对所有 envelope/payload/request 模型；`TvAuthEnvelopeSpecTest` 新增 R8 规则审计，防止回退；`CONTEXT.md` 更新“TV 配对会话响应包装”约定，明确 Release R8 keep 是这组模型契约的一部分。
- 影响文件：`android-tv-app/tv-app/proguard-rules.pro`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/network/TvAuthEnvelopeSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.network.TvAuthEnvelopeSpecTest` 失败于缺少 `-keep class com.chee.videos.core.model.TvAuth* { *; }`；实现后 `./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过，`./gradlew --no-daemon :tv-app:assembleRelease` 通过，且 `seeds.txt` 显示 `TvAuth*` 模型被 keep 规则选中、`mapping.txt` 显示类名和关键 getter/构造函数保留；`rg -n $'\uFFFD' ...` 无输出，未发现乱码替换字符。

## 2026-05-22 23:12 +0800
- 进度：继续排查 TV 配对页仍显示 `创建配对会话失败 (ClassCastException)`。ADB 已确认设备 `192.168.1.6:5555` 当前安装 `com.chee.videos.tv` 版本 `0.1.58` / `versionCode=59`，安装时间 `2026-05-22 23:08:02`。UIAutomator dump 证实配对页错误文案仍在；logcat 没有 Java FATAL，因为异常被 `TvAuthRepository.createSession` 捕获成 `Result.failure`。OkHttp 日志显示 `POST http://192.168.1.24:8080/api/v1/tv-auth/sessions` 返回 `200 OK`，curl 同请求返回完整 `{"code":0,"data":{...},"msg":""}`。本地 `javap` 反查未混淆 debug class，`createSession` 在接口返回后仍有 `checkcast TvAuthCreateEnvelope`，说明需要先拿到被吞掉的真实 `ClassCastException` 堆栈与运行时返回对象类型，再做正式修复。
- 影响文件：`plan.md`；下一步可能临时给 TV 认证错误路径加诊断日志，确认后再收敛成正式修复。
- 验证：已执行 `adb devices -l`、`adb shell dumpsys package com.chee.videos.tv`、`adb logcat -d`、`adb shell uiautomator dump`、`curl -X POST /api/v1/tv-auth/sessions`、`javap TvAuthRepository`；尚未修改生产代码。

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

## 2026-05-22 22:00 +0800
- 进度：C 批视觉/交互打磨（C1-C4），版本 0.1.54 → 0.1.55（versionCode 55→56）。C1：海报墙卡片标题 Box 补 `heightIn(min=74.dp)` + `contentAlignment=TopStart`，同行 1 行/2 行标题卡片底边对齐。C2：状态屏图标 28→36dp、加载圈 24→32dp、操作按钮 icon 18→20dp + 纵向 padding 10→12dp，TV 10-foot 可读性提升。C3：电视剧详情”主演：”文本补 `maxLines=1, overflow=Ellipsis`，防超长演员列表换行破坏布局。C4：播放器居中反馈 icon 显式 `size(22.dp)`，反馈文本补 `maxLines=1, overflow=Ellipsis`，防长文案跳变。
- 影响文件：`TvPosterWallScreen.kt`（import heightIn + 标题 Box 约束）、`TvStateFeedback.kt`（图标/圆圈/按钮尺寸）、`TvSeriesDetailScreen.kt`（主演行溢出保护）、`LongFormVideoPlayer.kt`（中心反馈 icon+文本约束）、`build.gradle.kts`（版本号）。
- 验证：`testDebugUnitTest` BUILD SUCCESSFUL 22s，`assembleDebug` BUILD SUCCESSFUL 11s（arm64-v8a + armeabi-v7a）。

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

## 2026-05-22 04:55 +0800
- 进度：用户实测 0.1.43 反馈三件事：(1) FATAL `FocusRequester is not initialized` 在 4K TV 上再次复现，栈走 `FocusRequester.findFocusTargetNode$ui_release` → `FocusOwnerImpl.focusSearch-ULY8qGw` → `AndroidComposeView$keyInputModifier$1.invoke-ZmokQxo` → `Activity.dispatchKeyEvent` → IMM/ViewRootImpl input pipeline → 主 Looper，最终栈底落到 `TvMainActivity.installMainLooperHoverExitGuard$lambda$0` 的 `Looper.loop()`；(2) 首页 hero 在 BringIntoView 钉位后「没有动画了」但顶部仍被视觉裁切；(3) 电影/`18+` 沉浸式详情页「仍有上滑动画 + 裁切」，4K TV 无可见状态栏。
- 修复（仅 FATAL 路径）：`TvMainActivity.kt` 的 `shouldSwallowTvComposeFocusRequesterCrash` 栈帧条件从原来的「className 必须等于 `androidx.compose.ui.focus.FocusRequester` 且方法名属于 `requestFocus` / `focus$ui_release` / `findFocusTargetNode$ui_release` 三者之一」放宽为「栈中任意一帧 `className.startsWith("androidx.compose.ui.focus.")`」。理由：Compose 1.7 在不同异步/同步路径下保留的栈帧组合不一致——`LaunchedEffect` 协程恢复路径保留 `FocusRequester` 帧，但 DPad keyInput 同步路径只保留 `FocusOwnerImpl` 帧 + `AndroidComposeView` 帧、`FocusRequester` 帧可能被剥离；旧 matcher 只贴方法名导致 keyInput 路径的同名异常逃出兜底。消息匹配 `contains("FocusRequester is not initialized")` 保留为主安全网，确保只吞这一类异常。同时新增两条单测：`swallows focus requester crash with only focus owner impl frame` 锁定 keyInput 路径正向用例（栈只有 `FocusOwnerImpl` + `AndroidComposeView$keyInputModifier$1`，无 `FocusRequester` 帧）；`does not swallow focus requester message when stack lacks compose focus package` 锁定误伤边界（消息含关键字但栈无 focus 包前缀帧时不吞）。`CONTEXT.md`：`TV 主 Looper FocusRequester 未初始化兜底` 词条改写为反映新的「消息 + focus 包前缀」两层判定，并把 keyInput 路径 (`AndroidComposeView$keyInputModifier$1.invoke-ZmokQxo` → `Activity.dispatchKeyEvent`) 写进文档案例。TV 版本 `0.1.43`→`0.1.44`，`versionCode` 44→45。本提交不动业务侧 `tryRequestFocus()`、不动详情页结构、不动 NavHost transition、不动 SharedTransitionLayout 链路。
- 未解决（等待用户复测信号）：
  - 首页 hero 顶部「无动画但仍裁切」：BringIntoView 钉位已生效（用户确认动画消失），但裁切仍在；表明 hero 的初始测量位置已偏移，而不是滚动行为造成。可能源头：(a) 4K TV ROM 在 `enableEdgeToEdge` 后报告异常 `WindowInsets` 把 `statusBarsPadding()` 推下非零像素；(b) Hero 内部 `Row(verticalAlignment = Alignment.Bottom)` 把内容贴底，背景图本身是渐变深色，被用户视觉上误判为「卡片被裁掉」。后续需要用户提供首页截图 / 实际 inset 数值（adb shell dumpsys SurfaceFlinger 或 `WindowCompat.getInsetsController` 读取）才能区分两类原因。
  - 详情页「仍有上滑动画 + 裁切」：详情页是 `Box(fillMaxSize)`，没有可滚动祖先，BringIntoView 不适用；hero `Surface` 没有 entrance 动画修饰器（grep 全无 `animatePlacement` / `slideIn` / `expandIn` / `Modifier.offset` / `Modifier.translationY`）。剩下最可能的源头：navigation-compose 2.7.7 默认 `enterTransition = fadeIn(tween(700))` 在 `SharedTransitionLayout` 包裹下退化为 `AnimatedContent` 的非纯 fade 行为，或 `TvLongFormDetailBackground` 内 `AsyncImage` 异步解码后第一帧填充时机与 `statusBarsPadding()` 应用时机错位（前一轮已把 `statusBarsPadding()` 从外层 Box 挪到返回按钮，但部分 4K TV 仍可能有 inset 延后到达），具体源头需要用户复测时附录屏才能定位。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.tv.TvMainActivityInputPolicyTest` 通过；待跑完整 `:tv-app:testDebugUnitTest :tv-app:assembleDebug` 锁住其它现有不变量（含 `TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvFocusSpecTest`、`TvListMotionSpecTest`、`TvSharedPosterTransitionSpecTest` 等）。

## 2026-05-22 04:05 +0800
- 进度：继续追查"进入 TV 首页与电影/`18+` 详情页时横幅大海报被上滑动画裁切"。上一轮（03:10）补齐 `statusBarsPadding()` 没解决——用户复测仍能看到「上滑动画 + 横幅大海报头部只剩一半」。本轮先把已经被排除的嫌疑写明：(1) `navigation-compose 2.7.7` `NavHost` 默认 `enterTransition = fadeIn(tween(700))` / `exitTransition = fadeOut(tween(700))`，纯 alpha 过渡无平移；(2) `tvFocusableGlow` / `tvFocusableScaleOnly` 只对 `scaleX/scaleY/shadowElevation` 走 `graphicsLayer` 弹簧动画，没有 `translationY`；(3) `tvStaggerEntry`（12dp 上移 260ms）只在 `TvPosterWallScreen` 调用，`TvCatalogScreen` / `TvLongFormDetailScreen` 内 grep 全无；(4) `SharedTransitionLayout`/`LookaheadScope` 只有 `sharedElement` / `animateBounds` / `animateContentSize` 接入才会驱动布局动画，两个目标页 grep 全无相关 modifier；(5) Activity 主题为 `Theme.Material3.DayNight`，未声明 `windowAnimationStyle`，也没 `installSplashScreen()`。剩下最可能的源头：Compose Foundation 1.7 `LazyColumn` 的 `BringIntoViewRequester`——`LaunchedTvInitialFocus` 通过 `featuredFocusRequester.tryRequestFocus()` 把焦点落在 324dp 高 `TvFeaturedHero` 底部的 `TvHeroActionButton`（Row `verticalAlignment = Alignment.Bottom` 内的播放按钮），LazyColumn 在初始测量 / inset 抵达时机不稳的瞬间把 hero 判为「未完全可见」，触发 spring 滚动让焦点节点对齐到「完全可见」位置，外观上就是上滑动画 + 横幅顶部被裁。
- 修复：`feature/tv/TvCatalogScreen.kt` 把主 `LazyColumn`（非搜索、非设置分支，行 ~221）的 `state` 从隐式改为显式 `val contentLazyListState = rememberLazyListState()`，并新增 `import androidx.compose.foundation.lazy.rememberLazyListState`、`import androidx.compose.runtime.mutableStateOf`、`import androidx.compose.runtime.setValue`；在焦点请求侧，`LaunchedTvInitialFocus { ... }` suspend block 内部 `tryRequestFocus()` 之后紧跟一段钉位逻辑：当 `var hasPinnedInitialScroll by remember { mutableStateOf(false) }` 为 false 时调用 `contentLazyListState.scrollToItem(0, 0)` 并把标志翻 true。该写法只在首屏「FEATURED」焦点路径副作用把 hero 向上滚动后做一次回钉，不影响后续 DPad 主动焦点移动时 BringIntoView 把后续 shelf 项滚入视口的正常 UX；也不影响 featured 内容延后到来导致 `initialFocusTarget` 第二次变化时（标志已 true 不会重复钉位，避免用户主动滚动被打回顶部）。`TvLongFormDetailScreen.kt` 上一轮已把 `statusBarsPadding()` 从外层 `Box` 移到返回按钮 `TvIconActionButton` 自己的 `Modifier.align(Alignment.TopStart).statusBarsPadding().padding(28.dp, 28.dp)` 上，保证 `TvLongFormDetailBackground` 的 `AsyncImage` 仍画到屏幕物理顶端、文字与可聚焦操作仍落在状态栏下方；本轮不再变动详情页结构（详情页是 `Box`，没有可滚动祖先，BringIntoView 不会改变位置——如果用户复测后详情页仍有相似体感，说明体感来源在 NavHost cross-fade + edge-to-edge inset 时序，需要额外信号才能定位）。`CONTEXT.md`：把 `TV 安全区域顶部留白` 词条改写为「非沉浸式页面外层叠 padding；沉浸式详情走分层规则——外层 Box 不叠，状态栏 inset 由前景元素自带」，与现实代码对齐；在 `TV 主 Looper FocusRequester 未初始化兜底` 之后新增 `TV 首页 hero BringIntoView 上滑钉位` 词条，固化 BringIntoView 根因、`rememberLazyListState()` + `scrollToItem(0, 0)` 钉位、`hasPinnedInitialScroll` 一次性标志的强约束。TV 版本 `0.1.42`→`0.1.43`，`versionCode` 43→44。本提交不动 hover-exit / FocusRequester 主 Looper 兜底、不动 NavHost 结构、不动 SharedTransitionLayout 链路、不引入自定义 `BringIntoViewSpec`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待跑 `./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` 确认现有 `TvCatalogFocusPolicyTest`（断言 `TvCatalogInitialFocusTarget.MENU -> menuFocusRequester.tryRequestFocus()` 字面仍在源代码内）、`TvInitialFocusSafeRequestTest`、`TvNoBareLaunchedEffectFocusRequestAuditTest` 等不变量继续绿；新增字符串没有破坏 `tryRequestFocus()` 调用点存在性断言。待手测：冷启动进入 TV 首页（含 featured、含 continue-watching、含 sections）应不再有「横幅大海报上滑、头部只剩一半」体感；进入电影/`18+` 详情页背景图应画到屏幕物理顶端、返回键不被状态栏遮挡；后续数据刷新（featured 晚到）后用户的主动滚动位置不被打回顶部。如果详情页体感仍存在，需要用户复测时附上设备型号、状态栏是否可见、是否仅在导航过渡瞬间出现的信号，才能进一步定位（可能涉及 NavHost cross-fade 与 edge-to-edge inset 抵达时序的二阶问题）。

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

## 2026-05-22 01:10 +0800
- 进度：落地 A 批第二项 A2——海报墙 → 电视剧详情 shared-element 过渡。helper 在 `core/ui/TvSharedPoster.kt` 暴露 `LocalTvSharedTransitionScope` / `LocalTvAnimatedContentScope` 两个 `CompositionLocal` 以及 `fun Modifier.tvSharedSeriesPoster(seriesId)`，scope 为 `null` 时直接 `return this` 保留非 SharedTransitionLayout / 单测环境安全；shared key 命名空间 `tv-series-poster-`。本轮把这套接口接到 `TvShellApp` 与两端页面：
  - `tv/TvShellApp.kt`：新增 `import androidx.compose.animation.SharedTransitionLayout` / `ExperimentalSharedTransitionApi`、`androidx.compose.runtime.CompositionLocalProvider`、`core/ui.LocalTvSharedTransitionScope` / `LocalTvAnimatedContentScope`；`TvAuthenticatedNav` 标注 `@OptIn(ExperimentalSharedTransitionApi::class)`；NavHost 外层包一层 `SharedTransitionLayout(modifier = Modifier.fillMaxSize())`，内部 `CompositionLocalProvider(LocalTvSharedTransitionScope provides this@SharedTransitionLayout)` 把 SharedTransitionScope 注入；只在 `composable(TvCatalogWallRoutePattern, ...) { ... }`、`composable(TvSeriesRoutePattern, ...) { ... }` 两个 destination 块体内额外 `CompositionLocalProvider(LocalTvAnimatedContentScope provides this@composable)` 注入 AnimatedContentScope，其他 destination（首页、长视频详情、长视频播放器、IPTV、剧集播放器）不注入，避免无关页面承担实验 API 成本。
  - `feature/tv/TvPosterWallScreen.kt`：新增 `import com.chee.videos.core.ui.tvSharedSeriesPoster`；`TvPosterWallCard` 的 9:16 海报 `Box` 在 `aspectRatio(9f/16f)` 之后、`.background(...)` 之前用 `.then(if (item.type == "tv") Modifier.tvSharedSeriesPoster(item.id) else Modifier)` 接入，保留电影/`18+` 走 LongFormDetail（A2b 待办）不参与本次 shared-element。
  - `feature/tv/TvSeriesDetailScreen.kt`：新增 `import com.chee.videos.core.ui.tvSharedSeriesPoster`；228dp / 2:3 / RoundedCornerShape(26.dp) 海报两个分支（AsyncImage + Box 无图回退）都在 `.aspectRatio(2f/3f)` 之后、`.clip(RoundedCornerShape(26.dp))` 之前叠加 `.tvSharedSeriesPoster(series.id)`；放在 `.clip(...)` 之前是为了避免裁剪与 shared-element 的 bounds resize 互相冲突造成裁切跳变。
  - `CONTEXT.md` 在 `TV 焦点请求安全调用 tryRequestFocus` 之后新增 `TV 电视剧海报 shared-element 过渡` 词条，固化 helper API、`@OptIn(ExperimentalSharedTransitionApi::class)` 约束、`tv-series-poster-` 命名空间、`Modifier.clip` 顺序约束以及"仅 TvCatalogWallRoutePattern / TvSeriesRoutePattern 注入 AnimatedContentScope"的强约束。
  - 不在本轮范围：A2b 长视频（电影/`18+`）shared-element（详情页没有可见小海报锚点，需要另设 UI 锚点）、A3 列表 stagger、A4 DPad 反馈、A5 动效 token 收口；B/C 批 ken-burns、glow 双层、`TvStateScreen` 等保持等待。TV 版本 `0.1.38`→`0.1.39`，`versionCode` 39→40。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含 `TvSharedPosterTransitionSpecTest` 3 条用例：helper 必须暴露 CompositionLocal + modifier + 命名空间 + `@OptIn` + scope 缺失回退；`TvShellApp` 必须 import `SharedTransitionLayout` 并出现 `LocalTvSharedTransitionScope provides`、`LocalTvAnimatedContentScope provides`；海报墙必须 import 并调用 `tvSharedSeriesPoster(...)` 且源文里存在 `"tv"` 字面量条件分支；详情页必须 import 并调用 `tvSharedSeriesPoster(...)`）。`./gradlew --no-daemon :tv-app:assembleDebug` 通过。审计测试 `TvNoBareLaunchedEffectFocusRequestAuditTest`、`TvFocusSpecTest`、`TvCatalogFocusPolicyTest`、`TvInitialFocusSafeRequestTest` 等历史不变量持续绿。待手测：从首页 / 海报墙打开任意电视剧（`item.type == "tv"`）应能看到海报放大变形过渡到详情页大海报；返回时海报应能逆向缩回；电影 / `18+` 走 LongFormDetail 不应触发该过渡；非 SharedTransitionLayout 环境（理论上仅单元测试） helper 应直接返回原 modifier 不崩。

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

## 2026-05-21 22:43 +0800
- 进度：登记 TV 第二阶段 UI 视觉与体感优化建议，本次只记录方案、不实施；建立在 `CONTEXT.md` 已落地的"TV 焦点视觉语言"、"沉浸式详情首屏"、"夜台玻璃面板"、"TV 状态反馈语言"、"TV 滚动内容底部安全留白"之上。建议按 A→B→C 三批推进，每批单独走"红→绿→assembleDebug→手测"流程并各自 bump `versionCode` / `versionName`。明确不在本轮优化范围内的项：播放内核 / 解码策略 / 播放历史上报 / IPTV 播放引擎选择 / 首页信息架构重排，延续"TV 第一阶段视觉覆盖范围"边界。
  - **A 批 · 流畅感（用户感知最强，改动收敛到共享 modifier 与 transition spec）**
    - A1 焦点放大改 spring 物理（damping≈0.8、stiffness≈380），替换 `tvFocusableGlow()` 内当前线性/easeOut 动效；先核对不与既有 focus state 冲突，必要时把动效参数下沉到共享 token。
    - A2 海报墙 → 长视频详情页接入 shared-element transition（海报缩略图位 → 沉浸式首屏小海报位）；电视剧详情走同一套语言。
    - A3 `LazyColumn` / `LazyVerticalGrid` 行/列入场 stagger 30–40ms / item，跨行使用 cubic 缓动而非 linear。
    - A4 DPad center 按下 100ms scale≈0.97 反馈，松开回弹；支持触觉的设备追加 `HapticFeedbackConstants.CONFIRM`。
    - A5 动效时长统一收口到 200–260ms（TV 端超过 300ms 即感觉迟滞），把 duration / easing token 落到共享文件复用。
    - 试水点位：A1 + A2 两项打头阵，体感对了再扩到全 app。
  - **B 批 · 精修视觉**
    - B1 10-foot 排版核对：主标题 ≥34sp、副标题 ≥22sp、辅助文字 ≥18sp（按 2m–2.5m 客厅距离反推），海报卡标题底色对比度提到 7:1。
    - B2 焦点 glow 从单层青蓝改双层（内层紧贴 0.6α、外层扩散 0.25α / 12dp），提升光感层级；继续遵守"不使用硬描边"约束。
    - B3 首页巨幅推荐 backdrop 加 120s 周期 ken-burns 缓慢运镜（1.03→1.08 scale + 微平移）降低静态感；尊重系统 reduced-motion 降级。
    - B4 沉浸式详情底部信息面板渐变改 16:9 短渐变 + 玻璃模糊，让横向背景信息透出而非整块全黑遮盖。
    - B5 海报墙、电视剧详情、IPTV 频道行的圆角统一收口到 16dp（当前 12–24dp 散落）。
  - **C 批 · 一致性收口**
    - C1 审计搜索 / 设置二级 / 播放器浮层 / 配对页等 Material 默认 `Button` / `IconButton` 残留点位，迁到共享 TV 焦点控件（沿用既有"TV 图标操作"组件）。
    - C2 加载 / 空态 / 错误统一走 `TvStateScreen` 类共享组件，错误态必须含可聚焦"重试"动作，符合"TV 状态反馈语言"。
    - C3 频道行台标、演员头像无图回退统一到圆形文字占位，颜色取主题 tertiary。
    - C4 滚动底部安全留白统一 64dp，回避手势栏；符合"TV 滚动内容底部安全留白"。
  - **依赖关系**：A 批的 spring token / 动效 token 共享文件落地后，B 批 ken-burns / glow 双层动画复用同一 token 体系；C 批 `TvStateScreen` 重试动作依赖 A 批焦点反馈到位才能保证遥控可达。
- 影响文件：`plan.md`
- 验证：本条目仅登记方案，未实施；下一步按 A→B→C 顺序拆 PR，每批落地时单独补"红→绿→assembleDebug→手测"和 TV 版本号 bump。

## 2026-05-21 21:36 +0800
- 进度：定位并修复 `LaunchedTvInitialFocus` 自身的 message 匹配缺陷。上一轮虽然把所有裸 `requestFocus` 都切到了 helper，但 helper 内 `runCatching` 过滤使用 `message.startsWith("FocusRequester is not initialized")`；崩溃日志显示 Compose 1.6 抛出的 `IllegalStateException` message 是 raw multiline 字符串字面量（`IllegalStateException:` 后跟换行和 3 空格缩进再到 `FocusRequester is not initialized...`），所以 `startsWith` 必然漏匹配，ISE 透出 helper、透出 hover-exit guard，直接 FATAL。修复：把字符串比较抽成 `internal fun isFocusRequesterNotInitialized(err: Throwable): Boolean`，匹配条件改成 `err is IllegalStateException && err.message?.contains("FocusRequester is not initialized") == true`；helper 内部改成调用该函数。新增 `TvInitialFocusRequesterMatcherTest` 用真实 Compose 1.6 raw multiline message 形态构造 ISE 验证 matcher 命中（红→绿），并补反向用例：紧凑形态、ACTION_HOVER_EXIT 不误吞、`null` message、`RuntimeException` 同消息不误吞。同步更新 `TvInitialFocusEffectShapeTest`：原有 `startsWith` 断言改为强制要求 `contains(FOCUS_REQUESTER_NOT_INITIALIZED_MARKER)`、并加注释解释 startsWith 漏匹配的成因；同时锁定 helper 必须复用 `isFocusRequesterNotInitialized(err)`。`CONTEXT.md` 把"以...开头"改为"包含...关键字"并写明"Compose 1.6 message 带前导换行+缩进必须 `contains` 不能 `startsWith`"作为强约束。TV 版本 `0.1.35`→`0.1.36`，`versionCode` 36→37。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusRequesterMatcherTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusEffectShapeTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：先红 `:tv-app:compileDebugUnitTestKotlin` 因为 `isFocusRequesterNotInitialized` 未定义编译失败；实现后 `:tv-app:testDebugUnitTest` 全绿，`:tv-app:assembleDebug` 通过。matcher 单测 4 条全部命中。审计测试和上轮 helper shape 测试仍绿。待手测：海报墙、电影/`18+`/电视剧详情页、IPTV、配对页、长视频播放器冷启动均不应再触发 `FocusRequester is not initialized` FATAL。

## 2026-05-21 20:58 +0800
- 进度：完成 TV `FocusRequester is not initialized` 同类裸用法全量清理。上一轮只覆盖 `TvCatalogScreen`，但 grep 出 7 处其他裸 `LaunchedEffect` + `requestFocus()` / 部分仅手工加了 `withFrameNanos { }` 但缺前缀过滤的同类风险点；本次按"全部切到 `LaunchedTvInitialFocus` 并删除冗余手动 `withFrameNanos`"统一处理：`TvPosterWallScreen.kt`（本次崩溃点，LazyVerticalGrid item 延迟组合）、`TvLongFormDetailScreen.kt`、`TvSeriesDetailScreen.kt`、`TvIptvScreen.kt`、`TvPairingScreen.kt`、`core/ui/LongFormVideoPlayer.kt`、`core/ui/SubtitlePicker.kt`。`LongFormVideoPlayer` 内 `pendingRootFocusRequest` / `pendingPlayPauseFocusRequest` 改用 `try { requestFocus() } finally { 清 pending }`，确保即使 helper 的 `runCatching` 兜下了 ISE，pending 标志也能被清，避免播放器状态机卡死。新增 `TvNoBareLaunchedEffectFocusRequestAuditTest` 审计测试：扫描 `src/main/java` 下所有 `.kt`，按行级括号深度追踪 `LaunchedEffect(...) {` 块体，若块体内出现 `.requestFocus(` 则 fail；同文件附 4 条 matcher 自测（命中、非命中、`LaunchedTvInitialFocus` 不误报、单行块体）。本提交不动 `CONTEXT.md`（"禁止业务 LaunchedEffect 内裸调"上轮已纳入约束）、不动 helper 自身逻辑、不动 hover-exit 兜底。TV 版本 `0.1.34`→`0.1.35`，`versionCode` 35→36。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvNoBareLaunchedEffectFocusRequestAuditTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：`./gradlew :tv-app:testDebugUnitTest` 通过（含新增审计测试和 matcher 自测）；`./gradlew :tv-app:assembleDebug` 通过。审计测试现已锁住"任何新加的 `LaunchedEffect` + `requestFocus()` 都会在 CI 红"的回归面，未来任何同类裸用法都会编译期外被挡住。待手测：进入海报墙、电影/`18+` / 电视剧详情页、IPTV 频道页、TV 配对页、长视频播放器都不应再触发 `FocusRequester is not initialized` 崩溃；播放器控制条焦点反复 toggle 后 pending 标志不应卡死。

## 2026-05-21 20:42 +0800
- 进度：完成 TV 首页冷启动 `FocusRequester is not initialized` 修复落地与验证。新建 `core/ui/TvInitialFocusEffect.kt` 暴露共享 helper `LaunchedTvInitialFocus(vararg keys, block)`，内部先 `withFrameNanos { }` 等过一帧再 `runCatching { block() }`，精确过滤 `IllegalStateException` 且 `message` 以 `FocusRequester is not initialized` 开头的异常并重抛 `CancellationException`，其他异常照常抛出。`TvCatalogScreen.kt` 移除裸 `LaunchedEffect` 焦点请求并切换到 helper（焦点选择策略 `resolveTvCatalogInitialFocusTarget` 不动）。补两类纯 Kotlin 结构性测试：`TvInitialFocusEffectShapeTest` 锁住 helper 的 `withFrameNanos`/`runCatching`/`CancellationException`/前缀匹配/重抛/`@Composable vararg keys` 这几个不变量；`TvCatalogFocusPolicyTest` 增 `sectionItemCounts = listOf(0, 0)` 的 MENU 兜底用例覆盖 sections 非空但全 0 的场景。`CONTEXT.md` 把原“TV 首页初始焦点”一条扩为“TV 初始焦点请求约束”，新增 LazyColumn 延迟组合 × `LaunchedEffect` 帧时序竞态的说明并要求统一走 helper。TV 版本 `0.1.33`→`0.1.34`，`versionCode` 34→35。本提交不动 compose-bom（升级到 2024.10+ 是后续单独 PR）、不动 `TvMainActivity` hover-exit 兜底、不动 `selectMenu(Settings)` 的 `loading=false` 行为（已被 helper 的 `runCatching` 兜住）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusEffectShapeTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew :tv-app:testDebugUnitTest` 通过；`./gradlew :tv-app:assembleDebug` 通过；helper shape 测试和 resolver MENU 用例均为绿。待手测：冷启动 TV 首页焦点落到巨幅推荐、Settings 菜单切回时不再触发 FocusRequester 崩溃、空内容首页落焦左侧菜单。

## 2026-05-21 20:32 +0800
- 进度：进入系统化排查 TV 首页冷启动 `IllegalStateException: FocusRequester is not initialized`。栈底落到 `installMainLooperHoverExitGuard$lambda$0` 只是上一轮 hover-exit 兜底在主 Looper 上接到的异常，不是因果链；真正抛出点位于 `TvCatalogScreen.kt:124` 即 `featuredFocusRequester.requestFocus()`，这是首屏 `LaunchedEffect(uiState.loading, isSearching, initialFocusTarget)` 体内的第一次焦点请求。当时 `featuredFocusRequester` 已经经过 `remember { FocusRequester() }` 创建，但承载该 `Modifier.focusRequester(...)` 的节点位于外层 `LazyColumn` 的 `item(key = "featured")` 里——`LazyColumn` 是延迟组合，`LaunchedEffect` 进入 RESUMED 时该 item 还没必然挂载，因此触发"未绑定节点"。同类历史已经在 `CONTEXT.md` 第 57 行有规则，但只覆盖"请求完全不存在的节点"，没有覆盖"会出现但当前还没挂载"的帧时序竞态。方案：抽 `LaunchedTvInitialFocus(vararg keys, block)` helper（先 `withFrameNanos { }` 等过一帧再 `runCatching` 精确过滤 `FocusRequester is not initialized` 前缀消息的 ISE，`CancellationException` 重抛，其他异常照常抛出），替换 `TvCatalogScreen` 那一处裸 `LaunchedEffect`；并扩展 `CONTEXT.md` 该条目，把"统一走 helper"和"禁止业务层裸调"写成强约束。本次不引入 Compose UI 测试 / Robolectric，结构性不变量测试足以锁回归面；不升级 compose-bom，不重排 LazyColumn 结构，不动 ViewModel `selectMenu(Settings)` 的 `loading=false` 行为（runCatching 兜底）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt`（新增）、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvInitialFocusEffectShapeTest.kt`（新增）、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待落地后跑 `:tv-app:testDebugUnitTest` + `:tv-app:assembleDebug`，再做冷启动手测。

## 2026-05-21 20:08 +0800
- 进度：完成海报墙发售时间排序 42P10 修复收尾验证；确认本次提交只纳入 `SearchVideosOrdered` SQL 重构、对应单测、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。后端无 Android 版本号需要 bump。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_search_test.go`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/repository/ -run 'TestSearchVideos' -count=1` 通过；`go test ./... -count=1` 通过；`go vet ./...` 无输出；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 20:02 +0800
- 进度：完成海报墙发售时间排序 42P10 红绿实现；把 `SearchVideosOrdered` 的 countSQL 和 selectSQL 抽成 `searchVideosCountSQL` 常量和 `searchVideosSelectSQL(orderClause)` helper，并用 `EXISTS (SELECT 1 FROM video_tags vt WHERE vt.video_id = v.id AND LOWER(COALESCE(vt.tag,'')) LIKE $2)` 替换原来的 `LEFT JOIN video_tags + SELECT DISTINCT`。结构上消除 DISTINCT 之后，发售时间排序使用的 `NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST` 不再被 Postgres 42P10 拦截。语义不变：标题/描述/任一标签匹配即命中，且不会再因为多标签 JOIN 出现重复行，因此 COUNT 由 `COUNT(DISTINCT v.id)` 改为 `COUNT(*)`。`CONTEXT.md` 的 `TV 海报墙排序` 词条补充 EXISTS 子查询与 42P10 约束说明。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_search_test.go`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/repository/ -run 'TestSearchVideos' -count=1` 因未定义 `searchVideosCountSQL` / `searchVideosSelectSQL` 编译失败；实现后同命令通过。待执行后端全量单测、`go vet`、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 19:56 +0800
- 进度：进入系统化排查 TV 海报墙按发售时间排序失败；后端报 `for select distinct, order by expressions must appear in select list (sqlstate 42p10)`。代码确认 `internal/repository/app_repository.go` 的 `SearchVideosOrdered` 同时使用 `SELECT DISTINCT` 和外部传入的 `ORDER BY ` 表达式，电影/`18+` 海报墙发售时间排序使用 `NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST, v.created_at DESC`，其中 `NULLIF(...)::date` 不在 SELECT 列表里，触发 Postgres `SELECT DISTINCT` 强制约束。DISTINCT 的存在原因是 `LEFT JOIN video_tags` 用于按 tag 模糊匹配，多标签匹配会产生重复行。电视剧路径走 `listTVSeriesSummariesOrdered` 的 `GROUP BY` 查询，`s.title` 在 GROUP BY 内、`MAX(v.created_at)` 是聚合函数，不受同类约束影响。推荐方向：用 `EXISTS` 子查询替换 `LEFT JOIN video_tags + SELECT DISTINCT`，结构上消除 DISTINCT 而非 SELECT 列表里硬塞排序表达式；同时保留 `SearchVideosOrdered` 的接口签名和外部 order clause 注入语义。
- 影响文件：`internal/repository/app_repository.go`、`internal/repository/app_repository_search_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待先补 `SearchVideosOrdered` SQL 形状红灯测试（覆盖无 DISTINCT、使用 EXISTS、order clause 原样嵌入），再实现并执行后端定向/全量验证。

## 2026-05-21 19:32 +0800
- 进度：完成 TV hover-exit 主 Looper 兜底收尾验证；确认本次提交只纳入 TV 主 Activity hover-exit 兜底扩展、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 通过；`./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`./gradlew --no-daemon :tv-app:assembleDebug` 通过；待执行乱码检查、diff 检查和提交范围检查。

## 2026-05-21 19:28 +0800
- 进度：完成 TV hover-exit 主 Looper 兜底红绿实现；`TvMainActivity.onCreate()` 安装主线程 `Handler.post { while(true) try { Looper.loop() } catch ... }` 外层异常拦截器，匹配到 Compose 平台层 hover-exit 异常后继续 loop，其它异常照常抛出。`shouldSwallowTvComposeHoverExitCrash` matcher 把方法名匹配放宽到 `sendHoverExitEvent` / `dispatchHoverEvent` 及其 `$lambda$` 合成方法名，覆盖 D8 生成的 lambda 调用帧。保留 `dispatchGenericMotionEvent` 同步兜底作为防御纵深。TV 版本更新为 `0.1.33` / `versionCode=34`，`CONTEXT.md` 在 `TV hover 输入兼容兜底` 词条补充主 Looper 调度路径与方法名匹配规则。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 因 matcher 不识别 `sendHoverExitEvent$lambda$5` / `dispatchHoverEvent$lambda$0` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 19:20 +0800
- 进度：进入系统化排查 TV App hover-exit 闪退复发；崩溃栈顶为 `AndroidComposeView.sendHoverExitEvent$lambda$5` 经 `Handler.handleCallback` 从 `Looper.loop` 抛出，调用链不再经过 `dispatchGenericMotionEvent`，因此上一轮在 Activity 输入边界 try/catch 的兜底接不到这条 `Handler.post` 路径；`shouldSwallowTvComposeHoverExitCrash` 的 matcher 也只匹配 `sendHoverExitEvent` / `dispatchHoverEvent` 精确方法名，无法识别 D8 合成的 `$lambda$` 帧。compose-bom 当前固定在 `2024.06.00`（compose-ui 1.6.x），属 Compose 平台层时序 bug，业务代码无法从源头规避。推荐方向 B：在 `TvMainActivity.onCreate()` 安装主线程外层 `Looper.loop()` try/catch 循环，并把方法名匹配放宽到 `$lambda$` 合成方法名，其他异常继续抛出；compose-bom 升级作为后续独立优化项。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 主 Looper hover-exit 兜底红灯测试（覆盖 `sendHoverExitEvent$lambda$5` 与 `dispatchHoverEvent$lambda$0`），再实现并执行 TV App 定向/全量验证。

## 2026-05-21 17:03 +0800
- 进度：完成 TV 工程编译边界瘦身收尾；确认本次提交只纳入 Gradle 编译排除边界、对应测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/*` 删除、openspec skill 未跟踪目录和未跟踪 `package-lock.json`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无输出；`git diff --check -- ...` 通过。

## 2026-05-21 17:02 +0800
- 进度：完成 TV 工程编译边界瘦身红绿实现；新增 Gradle 边界测试，红灯确认未声明排除清单。实现后通过 Kotlin sourceSets 排除手机端启动、手机首页/登录/Mine、短视频、图片合集、统一短视频播放器和相关测试源，保留 TV 主链路需要的连接页、详情 ViewModel、长视频播放器、网络模型和 IPTV。TV 版本更新到 `0.1.32` / `33`。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvApkPackagingConfigTest'` 通过；待执行 TV 端全量单测和 Debug 构建。

## 2026-05-21 16:56 +0800
- 进度：继续针对 TV App review 修复剩余结构问题；聚焦 TV 工程编译边界瘦身。代码确认 Manifest 只启动 `TvMainActivity`，但 `VideoHomeApp`、`MainActivity`、短视频、图片合集、Mine、手机端首页/登录等源文件仍参与 TV 编译；推荐先用 Gradle source exclude 明确排除手机端/短视频/图片合集主源与对应测试，保留 TV 主链路仍复用的 `DetailViewModel`、连接页、长视频播放器、网络 DTO 和 IPTV 依赖。本轮不物理删除源码，降低回滚成本。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、TV 编译边界测试、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 编译边界红灯测试，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-21 14:31 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第七轮；用户确认电影/`18+` 长视频详情页操作组件收尾，并要求如果还有明显可优化点一并处理。代码确认该页返回按钮仍是手写 `Surface + tvFocusableGlow + Icon`，播放/收藏已使用共享焦点语义；同时该页属于沉浸式详情首屏，不应套用滚动页底部安全留白。推荐本轮把返回操作接入共享 `TvIconActionButton`，并补测试锁住不使用默认 Material 操作控件和不误加滚动底部留白。`CONTEXT.md` 已记录 `TV 长视频详情页操作`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补长视频详情页操作和沉浸式边界红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 13:59 +0800
- 进度：完成 TV 电视剧详情页操作收尾验证；确认本次提交只纳入电视剧详情页返回图标操作统一、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 13:54 +0800
- 进度：完成 TV 电视剧详情页操作收尾红绿实现；新增电视剧详情页操作回归测试，红灯阶段确认返回操作尚未复用共享 `TvIconActionButton` 且仍残留默认 Material `IconButton` 导入；实现后返回按钮改为共享 TV 图标操作组件，播放、季选择和集选择继续使用 `tvFocusableGlow`，详情页布局、剧集选择逻辑和播放路由保持不变。TV 版本更新为 `0.1.29` / `versionCode=30`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailActionSpecTest'` 因未复用共享图标操作和默认 `IconButton` 导入失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 13:47 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第六轮；已确认聚焦“电视剧详情页的焦点/组件收尾”。推荐把电视剧详情页返回按钮接入共享 `TvIconActionButton`，保留播放、季选择和集选择的现有 `tvFocusableGlow` 语义；不改详情页布局、剧集选择逻辑或播放路由。`CONTEXT.md` 已记录 `TV 电视剧详情页操作`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补电视剧详情页图标操作红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 12:43 +0800
- 进度：完成 TV 图标类操作焦点统一收尾验证；确认本次提交只纳入共享 TV 图标操作组件、TV 首页搜索清空、海报墙返回、长视频播放器控制按钮替换、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvIconActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 12:40 +0800
- 进度：完成 TV 图标类操作焦点统一红绿实现；新增共享 `TvIconActionButton`，红灯阶段确认共享组件缺失且 TV 首页搜索清空、海报墙返回、长视频播放器控制按钮仍导入默认 Material `IconButton`；实现后这三类图标操作均改用共享 TV 图标操作组件，IPTV 根焦点容器继续保留用于接收遥控按键。TV 版本更新为 `0.1.28` / `versionCode=29`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvIconAction.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvIconActionSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvIconActionSpecTest'` 因缺少共享组件和目标文件仍导入默认 `IconButton` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 12:32 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第五轮；已确认聚焦“TV 图标类操作的焦点统一”。推荐新增共享 TV 图标操作组件，替换 `feature/tv` 与长视频播放器中仍依赖默认 Material `IconButton` 的主要图标操作；不替换 IPTV 根 `.focusable()`，不触碰短视频、图片合集和手机端遗留页面。`CONTEXT.md` 已记录 `TV 图标操作`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 图标操作红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 12:20 +0800
- 进度：完成 TV 连接服务器页优化收尾验证；确认本次提交只纳入连接页 TV 面板/焦点操作、连接页底部安全留白、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 12:18 +0800
- 进度：完成 TV 连接服务器页红绿实现；新增连接页体验回归测试，红灯阶段确认连接页仍使用默认 Material 区块和按钮；实现后自动嗅探、手动填写、历史地址区块改为 TV 深色 `Surface` 面板，重新扫描、测试并保存、使用/连接、删除改为共享 `tvFocusableGlow` 操作按钮；扫描 loading 保持小型行内状态，连接页滚动内容接入统一底部安全留白。TV 版本更新为 `0.1.27` / `versionCode=28`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest'` 因连接页缺少 TV 面板和共享焦点动作失败；底部留白红灯阶段 `--tests 'com.chee.videos.feature.tv.TvScrollableBottomPaddingTest'` 因连接页未使用统一底部留白失败；实现后两个定向测试通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 12:11 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第四轮；已确认聚焦“连接服务器页”。推荐把自动嗅探、手动填写、历史地址三个区块统一为 TV 深色面板风格，所有连接页操作接入共享 `tvFocusableGlow`，服务器扫描继续保持小型行内 loading；不改服务器发现逻辑、保存逻辑或接口协议。`CONTEXT.md` 已记录 `TV 服务器连接页`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、连接页相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补连接页焦点和面板风格红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 11:51 +0800
- 进度：完成 TV 配对/服务器连接与根启动体验优化收尾验证；确认本次提交只纳入配对页焦点按钮、根启动共享状态、配对连接体验测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvPairingConnectionExperienceTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。待执行暂存区复核与提交。

## 2026-05-21 11:50 +0800
- 进度：完成 TV 配对/服务器连接与根启动体验红绿实现；新增配对连接体验回归测试，红灯阶段确认配对页仍裸用 `.focusable()` 且根启动仍直接使用默认进度环；实现后配对页两个操作改为共享 `tvFocusableGlow` 焦点按钮，根启动改用 `TvPageLoadingState`，服务器自动嗅探保持小型行内 loading。TV 版本更新为 `0.1.26` / `versionCode=27`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvPairingConnectionExperienceTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvPairingConnectionExperienceTest'` 因配对页焦点和根启动 loading 约束失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-21 11:45 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第三轮；本轮聚焦“配对/服务器连接与根启动体验统一”。推荐把 TV 配对页操作接入共享焦点视觉，避免裸 `.focusable()` 形成重复或低质焦点；根启动 loading 接入共享页面级状态组件；服务器自动嗅探继续保持小型行内 loading，不改配对协议、认证接口或服务器发现逻辑。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、TV 相关测试、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补配对焦点与根启动状态反馈红灯测试，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-21 11:19 +0800
- 进度：继续 `$grill-with-docs` 做 TV App 整体优化第二轮；已确认聚焦“加载/空态/错误态统一”。推荐新增共享 TV 状态组件，页面级加载使用居中紧凑状态，列表分页加载使用行内状态，空态说明保持短句，错误态提供可聚焦重试动作；服务器自动嗅探保持表单内小型行内 loading，不改成页面级加载态。`CONTEXT.md` 已记录 `TV 状态反馈语言`。
- 影响文件：`CONTEXT.md`、`plan.md`；后续预计影响 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/*`、TV 单测与 `android-tv-app/tv-app/build.gradle.kts`
- 验证：待先补统一状态组件和重点页面使用红灯测试，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-21 10:50 +0800
- 进度：确认 TV 第一阶段视觉覆盖范围；覆盖首页左侧菜单、首页内容卡片、搜索、设置、海报墙、电影/`18+` 详情页、电视剧详情页、IPTV 频道列表、播放器音轨/字幕/返回提示等浮层控件；不覆盖播放内核、解码策略、播放历史上报、IPTV 播放引擎选择或首页信息架构重排。`CONTEXT.md` 已记录该范围。
- 影响文件：`CONTEXT.md`、`plan.md`；后续预计影响 TV 焦点修饰符、各 TV 页面可点击元素及相关单测。
- 验证：待确认实施顺序、红灯测试口径，再实现并执行 TV App 定向/全量验证。

## 2026-05-21 10:49 +0800
- 进度：确认 TV 第一阶段焦点视觉策略；海报/图片卡片只放大和轻阴影，按钮、菜单、筛选项、频道行统一柔和蓝青色光感或背景提亮，播放器浮层延续夜台玻璃面板，不再使用硬描边或粉红描边。`CONTEXT.md` 已记录 `TV 焦点视觉语言`。
- 影响文件：`CONTEXT.md`、`plan.md`；后续预计影响 TV 焦点修饰符、首页/海报墙/详情页/IPTV/设置页可点击元素及相关单测。
- 验证：待继续确认覆盖页面范围、红灯测试口径，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-20 22:59 +0800
- 进度：开始修复上传图片压缩失败；错误链路显示当前 ffmpeg 同时缺少 `libwebp` 与 `webp` 编码器，且系统未安装 `cwebp`，导致上传阶段强制转 WebP 失败。推荐策略是 WebP 编码能力不可用时保留原图作为已处理文件，上传不失败；图片变体仍按请求动态生成，后续可独立增强降级策略。
- 影响文件：`internal/services/image.go`、`internal/services/image_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`CONTEXT.md`、`plan.md`
- 验证：待先补 WebP 编码不可用时上传成功的红灯测试，再实现并执行 Go 定向/全量验证、乱码检查、diff 检查和提交范围检查。

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

## 2026-05-20 19:02 +0800
- 进度：进入 `$grill-with-docs` 设计 TV 海报墙排序；代码确认 TV 端海报墙当前只传 `kind/page/page_size`，后端 `/api/v1/tv/catalog` 也没有排序参数，电影/18+ 当前按 `videos.created_at DESC`，电视剧按固定 SQL。确认排序必须由服务端分页接口执行，客户端只排序当前页会破坏翻页全局顺序。推荐参数为 `sort_by=added|release`、`sort_order=asc|desc`；电影/18+ 添加时间用 `videos.created_at`、发售时间用 metadata `release_date`，电视剧发售时间用 `series.first_air_date`、添加时间用剧集关联视频最新 `created_at`，缺失日期排最后。
- 影响文件：`internal/handlers/tv.go`、`internal/services/tv_auth.go`、`internal/services/tv_catalog_wall_test.go`、`internal/repository/app_repository.go`、`internal/repository/tv_repository.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补后端排序语义和 TV ViewModel 红灯测试，再实现并执行后端/TV 定向与全量验证。

## 2026-05-20 18:46 +0800
- 进度：完成 TV hover-exit 闪退兜底收尾验证；确认本次提交只纳入 TV 主 Activity 输入异常兜底、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:43 +0800
- 进度：完成 TV hover-exit 闪退兜底红绿实现；`TvMainActivity.dispatchGenericMotionEvent()` 捕获 Compose 平台层 `The ACTION_HOVER_EXIT event was not cleared.` 异常，并通过异常消息与 `AndroidComposeView` 堆栈双重匹配后才吞掉，其他输入异常继续抛出。TV 版本更新为 `0.1.22` / `versionCode=23`，`CONTEXT.md` 记录 TV hover 输入兼容兜底。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvMainActivityInputPolicyTest'` 因缺少 hover-exit 兜底判断函数失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:36 +0800
- 进度：进入系统化排查 TV App hover 输入闪退；崩溃栈显示异常完全来自 `AndroidComposeView.dispatchHoverEvent` / `sendHoverExitEvent`，业务代码未出现在调用栈中。代码确认 `TvMainActivity` 当前没有统一 generic motion 兜底，仓库也没有自定义 hover 处理。推荐在 Activity 边界只吞掉 Compose `The ACTION_HOVER_EXIT event was not cleared.` 这一类平台输入异常，其他异常继续抛出。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvMainActivityInputPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 主 Activity hover-exit 闪退兜底红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 18:31 +0800
- 进度：完成 TV 服务器自动嗅探 loading 收尾验证；确认本次提交只纳入连接服务器页扫描 loading 尺寸、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:28 +0800
- 进度：完成 TV 服务器自动嗅探 loading 红绿实现；扫描状态改为 14dp 小型行内进度环并使用 2dp 线宽，避免只限制高度导致默认进度环视觉过大。TV 版本更新为 `0.1.21` / `versionCode=22`，`CONTEXT.md` 记录服务器自动嗅探状态应使用小型行内 loading。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.connection.ConnectionScreenLoadingSpecTest'` 因缺少小型行内 loading 规格失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:23 +0800
- 进度：进入 `$grill-with-docs` 检查 TV App 探测 IP 界面；代码确认目标页是 `ConnectionScreen` 的“自动嗅探”卡片，扫描态当前使用 `CircularProgressIndicator(modifier = Modifier.height(20.dp))`，只限制高度未限制宽度，可能保持默认进度环宽度而显得过大。推荐将其收敛为行内小进度环，不改扫描、连接和列表业务逻辑。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/connection/ConnectionScreenLoadingSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 服务器扫描 loading 尺寸红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 18:11 +0800
- 进度：完成 TV 海报墙 9:16 海报卡收尾验证；确认本次提交只纳入 TV 海报墙卡片视觉、无描边焦点修饰器、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 18:08 +0800
- 进度：完成 TV 海报墙 9:16 海报卡红绿实现；卡片图片区改为 9:16 且图片贴边显示，标题条紧贴图片底部并使用深色背景，卡片焦点改为仅放大/阴影的无描边焦点修饰器。AV 海报墙显示沿用后端 `title` 作为番号；TV 版本更新为 `0.1.20` / `versionCode=21`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvPosterWallCardContentTest' --tests 'com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest'` 因缺少 `showDescription` 失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 18:02 +0800
- 进度：进入 `$grill-with-docs` 优化 TV 海报墙视觉；代码确认当前海报墙卡片是 2:3 竖卡、图片有 12dp 内边距、标题和简介在普通卡片底部、焦点使用带粉色描边的 `tvFocusableGlow()`。用户紧急修正海报比例为 9:16，按最新需求更新为海报墙 9:16 竖向海报；API 当前没有独立番号字段，AV 海报墙标题沿用后端返回的 `title`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 海报墙 9:16 海报卡红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 17:23 +0800
- 进度：完成 TV 滚动内容底部安全留白收尾验证；确认本次提交只纳入 TV 可滚动内容底部留白、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLayoutSpec.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 17:22 +0800
- 进度：完成 TV 滚动内容底部安全留白红绿实现；新增共享 `TvLayoutSpec.scrollBottomSafePaddingDp=56`，TV 首页/搜索、海报墙、电视剧详情页、IPTV 频道列表、剧集选择底部抽屉统一使用该底部留白。播放器画面和沉浸式详情首屏保持不变。TV 版本更新为 `0.1.19` / `versionCode=20`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLayoutSpec.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvScrollableBottomPaddingTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvScrollableBottomPaddingTest'` 因缺少首页底部留白规格失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 17:17 +0800
- 进度：进入 `$grill-with-docs` 检查 TV App 页面底部留白；代码确认不是所有页面都缺少底部留白，但 TV 首页/搜索、海报墙、电视剧详情页、IPTV 频道列表、剧集选择底部抽屉等滚动内容底部留白只有 18-24dp 且不统一。已确认仅统一可滚动内容页的底部安全留白，不改播放器画面和沉浸式详情首屏。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvLayoutSpec.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 滚动内容底部安全留白红灯测试，再实现并执行 TV App 定向/全量验证。

## 2026-05-20 16:39 +0800
- 进度：完成 TV 播放器连按合并跳转收尾验证；确认本次提交只纳入 TV 播放器快进/快退 debounce、相关测试、TV 版本号、`CONTEXT.md` 和 `plan.md`，不纳入既有 `.codex/skills/av-scraper-optimization` 删除和 openspec skill 未跟踪目录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' ...` 无命中；`git diff --check -- ...` 通过。

## 2026-05-20 16:23 +0800
- 进度：完成 TV 播放器连按合并跳转红绿实现；新增 pending seek 纯逻辑，快进/快退按键每次都即时刷新累计目标和预览反馈，但实际 `seekTo` 延迟约 300ms 且只提交最后一次目标。TV 版本更新为 `0.1.18` / `versionCode=19`。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 因缺少 `TvPendingStepSeekUpdate` 和 `resolveTvPendingStepSeek` 编译失败；实现后同命令通过。待执行 TV App 全量单测、构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-20 16:16 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV 播放器快进/快退 debounce；代码确认当前 `performStepSeek` 每次按键都会立即 `player.seekTo`，已确认采用 `连按合并跳转`：连按期间只更新累计目标和预览，停止按键约 300ms 后只执行一次实际跳转，方向切换时以当前累计目标继续计算。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补连按合并跳转红灯测试，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-20 13:42 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV App 根页面退出确认；代码确认 `tv-home` 当前不拦截系统返回键，海报墙/详情页由壳层返回上一页，电影/电视剧播放器已有“再按一次返回”确认。本次推荐只在 `tv-home` 增加二次退出提示，播放页逻辑保持独立。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待先补 TV 壳层根退出确认红灯测试，再实现并执行 TV App 定向/全量验证。

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

## 2026-05-19 20:45 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV 电影详情页横向背景和 TV 音轨面板遥控确认/焦点视觉；代码确认 TV 电影详情当前只读 metadata 顶层 `backdrop_url`/`backdrop_path`，已确认需兼容 `metadata.tmdb.backdrop_url`/`metadata.tmdb.backdrop_path` 后再退回竖版海报兜底。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行 TV App 定向测试、文档乱码检查和 diff 检查。

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

## 2026-05-19 20:41 +0800
- 进度：确认影视演员信息刮削覆盖电影/电视剧自动上传刮削、管理端手动确认刮削、`SyncMovieMetadata` 和 `SyncTVEpisode` 等落库入口；候选预览阶段不下载演员头像，避免预览产生本地副作用。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:40 +0800
- 进度：确认影视演员合并策略：优先按 `source=scrape_tmdb + external_id=TMDB person id` 匹配，其次按姓名匹配；同名演员为 AV 或人工来源且已有头像/备注时不覆盖，只补空字段并绑定视频。头像仅在演员无头像时下载本地头像。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:39 +0800
- 进度：确认影视演员资料字段深度：电影/电视剧刮削补齐 TMDB person 的姓名、别名、性别、国家/地区、生日、简介、TMDB person id 与本地头像；本次不存 credits 角色名，避免扩展 `video_actors` 绑定语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:38 +0800
- 进度：进入 `$grill-with-docs` 讨论演员信息刮削；代码确认已有 `actors` / `video_actors` 表、AV 演员头像补全链路，以及电影/电视剧当前只按 TMDB credits 演员姓名绑定。已确认本次“演员信息刮削”先覆盖电影和电视剧 TMDB 演员，不调整 AV 演员策略。
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

## 2026-05-19 20:35 +0800
- 进度：确认本次不新增管理端刮削页自动回显模式，只修“查询预览”按钮行为；用户点击电影查询预览时强制在线重抓，页面打开仍等待用户主动点击。
- 影响文件：`plan.md`
- 验证：待实现阶段执行后端/管理端定向测试和文档检查。

## 2026-05-19 20:33 +0800
- 进度：确认通用 `/admin/scrape/preview` 增加显式 `bypass_cache` 语义；电影手动点击“查询预览”默认传 `bypass_cache=true`，后端据此绕过已有 metadata 复用和 `PreviewMovie` 短期候选缓存。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:31 +0800
- 进度：确认电影重新刮削绕过缓存规则本次只适用于 `type=movie`；电视剧和 AV 手动刮削流程暂不调整，避免扩大到季集绑定或 AV 站点缓存策略。
- 影响文件：`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 20:29 +0800
- 进度：进入 `$grill-with-docs` 讨论电影重新刮削缓存问题；代码确认当前管理端 `AdminScrapePreview` 会在视频已有 metadata 时返回 `from_cache=true`，`PreviewMovie` 也有 5 分钟候选缓存。已确认“电影重新刮削”必须实时查 TMDB，不复用已有 metadata 或短期预览缓存。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

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

## 2026-05-19 19:13 +0800
- 进度：确认为定位 TV 运行时切轨失败，允许在播放器选择音轨和 Tracks 变化时增加低噪声诊断日志；日志应包含选择的音轨 id、当前音轨列表、override 目标和切换后的 selected 轨，用于验证 Media3 是否实际切轨。
- 影响文件：`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:11 +0800
- 进度：确认已有电影重新确认刮削时允许覆盖旧电影横向背景；当前系统没有 artwork 锁定机制，先沿用“确认刮削即覆盖刮削来源 metadata”的语义。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:09 +0800
- 进度：确认自动电影刮削和管理端手动电影刮削都必须补齐电影横向背景，并尽量收敛到同一个后端确认/保存逻辑，避免入口差异导致 TV 展示素材不一致。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:07 +0800
- 进度：确认电影横向背景刮削结果应下载到本地并保存本地可访问路径到电影 metadata，不仅保存 TMDB 远程图片路径。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:05 +0800
- 进度：确认 TV 音轨列表项的信息层级：主标题显示语言或原始 label，声道作为弱化信息，编码仅在有值且简短时弱化展示；默认项文案统一为“自动选择”，表示跟随视频默认音轨。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:03 +0800
- 进度：确认 TV 字幕选择和音轨选择共用同一套“夜台玻璃面板”视觉语言，避免播放器相邻菜单出现割裂体验。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 19:00 +0800
- 进度：确认 TV 音轨选择界面采用“夜台玻璃面板”：居中深色半透明玻璃浮层，背景可感知但压暗，边缘细高光，焦点态使用柔和蓝青色光感；不采用底部抽屉。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 18:57 +0800
- 进度：确认 TV 电影音轨问题的验收口径为“运行时切轨”：选择音轨后应立即切换当前播放音频，不重启播放、不重新进入详情页，并且音轨列表应反映当前播放音轨；已补充到 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`、`plan.md`
- 验证：待完成讨论后执行文档乱码检查和 diff 检查。

## 2026-05-19 18:50 +0800
- 进度：进入 `$grill-with-docs` 讨论 TV App 电影音轨选择、音轨选择界面升级和电影横向背景刮削；已确认“电影横幅海报”统一称为“电影横向背景”，指 TMDB 的 16:9 `backdrop_path`/`backdrop_url`，不是竖版 `poster_path` 裁切。
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

## 2026-05-19 17:48 +0800
- 进度：开始清理不宜提交到 Git 的本地文件；计划检查已有忽略规则、已跟踪但应忽略的生成物、未跟踪本地缓存/构建产物，随后补充 `.gitignore`、从索引移除已跟踪生成物，并做忽略规则和状态验证。
- 影响文件：`.gitignore`、`CONTEXT.md`、`plan.md`，以及待确认需要从索引移除的生成物。
- 验证：待执行 `git status --short`、`git ls-files -ci --exclude-standard`、`git check-ignore -v ...`、乱码检查和 diff 检查。

## 2026-05-19 13:49 +0800
- 进度：完成 TV App 启动崩溃排查修复最终验证；确认本次提交只纳入 TV 首页空内容初始焦点兜底、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。当前 `adb devices` 无在线设备，因此未能直接抓取真机 `logcat` 或做安装启动实测。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusPolicyTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- ...` 通过；`adb devices` 无在线设备。

## 2026-05-19 13:45 +0800
- 进度：定位并修复 TV App 安装后启动崩溃的高概率根因；当前无 ADB 设备在线无法抓真实 `logcat`，静态排查发现 TV 首页默认内容为空时 `resolveTvCatalogInitialFocusTarget()` 返回 `SEARCH`，但搜索框仅在切到“搜索”菜单后才会组合，冷启动时 `searchFocusRequester.requestFocus()` 可能请求未绑定焦点节点并触发 Compose 启动崩溃。已新增 `MENU` 焦点兜底，空内容时请求已组合的左侧菜单焦点，并补测试约束。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusPolicyTest'` 因缺少 `TvCatalogInitialFocusTarget.MENU` 失败；实现后同命令通过。待执行 TV 全量单测、Debug 构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 13:38 +0800
- 进度：开始 review 并排查 TV App 安装后启动即崩溃；计划先抓取设备 `logcat` 和启动链路证据，结合最近 TV 改动检查 Manifest、Hilt、ABI split、Application/Activity 初始化和首屏 Compose，再补可回归测试或静态约束，修复根因并完成 TV 单测/构建验证。
- 影响文件：待确认，预计涉及 `android-tv-app/tv-app/src/main`、`android-tv-app/tv-app/src/test`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行崩溃复现/日志确认、TV 定向测试、TV 全量单测、TV Debug 构建、乱码检查和 diff 检查。

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

## 2026-05-19 11:54 +0800
- 进度：开始实现 TV App 电影/18+ 详情页沉浸式改版；计划先补 `TvLongFormDetailPresentationTest` 红灯测试约束信息行、背景兜底、演员头像/占位、收藏按钮和移除下方信息卡片，再改造 TV 长视频详情页首屏布局，最后更新 TV 版本号与技术沉淀。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 详情页红灯测试、TV 定向单测、TV 全量单测、TV Debug 构建、乱码检查和 diff 检查。

## 2026-05-19 11:28 +0800
- 进度：完成 TV 首页菜单确认键修复最终验证；确认本次提交只纳入 TV 首页侧边菜单单焦点修复、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/build.gradle.kts` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 11:27 +0800
- 进度：完成 TV 首页侧边菜单遥控确认键核心修复；根因是 `TvHomeSideMenuButton` 在 `tvFocusableGlow()` 已提供焦点目标后又叠加 `.focusable()`，造成菜单按钮重复焦点目标，遥控确认键可能第一次只落到内部焦点层、第二次才触发点击。已删除重复 `.focusable()`，TV 版本更新为 `0.1.11` / `versionCode=12`，`CONTEXT.md` 记录 TV 菜单按钮单焦点目标约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvHomeNavigationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 因侧边菜单按钮仍包含重复 `.focusable()` 失败；实现后同命令通过。待执行 TV 全量单测、Debug 构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 11:25 +0800
- 进度：开始排查 TV App 遥控菜单键需要按两次才生效的问题；计划先沿长视频播放器和 IPTV 播放页按键分发确认根因，再补纯逻辑红灯测试，修复菜单键一次触发目标菜单动作，并同步更新 TV 版本号、技术沉淀和验证记录。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行菜单键红灯测试、TV 定向单测、TV 全量单测、TV Debug 构建、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 10:45 +0800
- 进度：完成电影手动刮削入口最终验证；确认本次提交只纳入管理端电影手动刮削入口、路由 helper 单测、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`CONTEXT.md`、`plan.md`
- 验证：`cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 通过；`cd admin-web && npm run build` 通过（Vite 仅提示 chunk size 警告）；`go test ./internal/handlers -run 'TestAdminScrape|TestShouldEnqueueAdminScrapeConfirmTranscode' -count=1` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md admin-web/src` 无命中；`git diff --check -- CONTEXT.md plan.md admin-web/src/views/videoList.helpers.js admin-web/src/views/videoList.helpers.spec.js admin-web/src/views/VideoList.vue` 通过。

## 2026-05-19 10:44 +0800
- 进度：完成电影手动刮削入口核心实现；`VideoList.vue` 在电影详情抽屉播放预览操作区新增“电影手动刮削”按钮，点击关闭抽屉并跳转通用刮削页；`buildMovieManualScrapeRoute` 生成 `/scrape` 路由 query，并从 `metadata.release_date` 或 `metadata.tmdb.release_date` 解析年份；`CONTEXT.md` 补充手动刮削术语和复用接口约定。
- 影响文件：`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/VideoList.vue`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd admin-web && npm run test -- src/views/videoList.helpers.spec.js` 因 `buildMovieManualScrapeRoute` 尚不存在失败；实现后同命令通过。待执行管理端构建、后端刮削回归测试、乱码检查、diff 检查和提交范围检查。

## 2026-05-19 10:42 +0800
- 进度：开始实现管理端电影手动刮削入口；计划先补 `buildMovieManualScrapeRoute` 路由 helper 红灯测试，再在视频详情抽屉为所有 `type=movie` 的视频新增“电影手动刮削”按钮，复用通用刮削页并仅预填查询参数，不自动请求。
- 影响文件：`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/VideoList.vue`、`CONTEXT.md`、`plan.md`
- 验证：待执行前端 helper 红灯测试、管理端构建、后端刮削回归测试、乱码检查、diff 检查和提交范围检查。

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

## 2026-05-19 10:27 +0800
- 进度：开始实现 TV APK 按 ARM ABI 拆包瘦身；计划先补 Gradle 打包配置静态红灯测试，再启用 `armeabi-v7a`/`arm64-v8a` ABI split、关闭 universal APK、开启 Release R8 与资源瘦身，并保持 LibVLC IPTV 播放兼容性不变。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvApkPackagingConfigTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：待执行 TV 打包配置红灯测试、TV 定向/全量单测、Debug/Release 构建、APK ABI 内容检查、体积检查、乱码检查、diff 检查和提交范围检查。

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

## 2026-05-19 09:41 +0800
- 进度：开始实现 TV IPTV 频道列表与顶部提示优化；计划先补纯逻辑和静态红灯测试，再将顶部频道信息改为进入/切台后 3 秒临时提示，频道列表打开时隐藏，并让频道列表初次打开直接定位到当前频道附近。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 IPTV 定向红灯测试、TV IPTV 定向单测、TV 全量单测、TV Debug 构建、乱码检查和提交范围检查。

## 2026-05-19 09:11 +0800
- 进度：完成 AV 大背景最终验证；确认本次提交只纳入后端 TV DTO 映射、TV 详情展示、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVCatalogWallVideoItems|Test.*AV' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest' --tests 'com.chee.videos.feature.tv.TvCatalogFeaturedContentTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md internal/services android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv` 无命中；`git diff --check -- ...` 通过。

## 2026-05-19 09:09 +0800
- 进度：完成 AV 大背景核心实现；后端 TV 首页 AV DTO 的 `backdrop_url` 按原始横幅优先和固定 fallback 顺序解析，`poster_url` 保持 `thumbnail_path`；TV 详情页 `videoType=av` 时顶部背景改用详情 metadata 的原始海报，左侧小海报仍使用 `thumbnail_path`；TV 版本更新为 `0.1.8` / `versionCode=9`，`CONTEXT.md` 记录 AV 大背景与竖卡海报分工。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestBuildTVHomeVideoFromListItem' -count=1` 因 AV `backdrop_url` 仍为缩略图失败；红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest'` 因 AV 详情背景未使用原始海报失败；实现后上述两个命令通过。待执行计划内完整验证。

## 2026-05-19 09:05 +0800
- 进度：开始实现 TV 端 AV 大背景优先使用原始横幅海报；计划先补后端 TV 首页 DTO 与 TV 详情页展示层红灯测试，再实现后端统一 `backdrop_url` 映射、TV 详情页 AV 背景选择、TV 版本号和技术沉淀。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行后端红灯测试、TV 详情页红灯测试、计划内后端/TV 定向与全量验证、Debug 构建和乱码检查。

## 2026-05-18 21:56 +0800
- 进度：完成 TV IPTV 台标和频道列表滚动修复最终验证；确认本次提交只纳入 TV IPTV UI/交互、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptv*'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv` 无命中；`git diff --check -- ...` 通过。

## 2026-05-18 21:49 +0800
- 进度：完成 TV IPTV 台标和频道列表滚动核心实现；播放页顶部和频道列表行改为使用 Coil `AsyncImage` 渲染 `logoUrl`，缺失或加载失败时回退 TV 图标；频道列表新增按频道 id 解析 `LazyColumn` item index 的 helper，并在列表打开和焦点移动后自动滚动到对应频道。TV 版本更新为 `0.1.7` / `versionCode=8`，`CONTEXT.md` 记录台标和列表滚动约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少频道列表 item index helper 编译失败；实现初版后同命令因索引偏移断言失败；修正后同命令通过。待执行计划内全量 TV 单测、Debug 构建和乱码检查。

## 2026-05-18 21:48 +0800
- 进度：开始修复 TV App IPTV 台标显示和频道列表遥控滚动；计划在顶部当前频道和右侧频道列表中使用 `logoUrl` 加载台标并保留 TV 图标回退，同时让频道列表打开时定位当前频道、焦点上下移动时滚动到可见区域。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行 IPTV 定向红灯测试、TV IPTV 定向单测、TV 全量单测、TV Debug 构建和乱码检查。

## 2026-05-18 21:26 +0800
- 进度：根据 `TvIptv` 日志确认当前 IPTV 无画面根因是频道 URL 本身为音频专用 HLS：`videoTracks=0 audioTracks=2 videoTrack=-1`，且地址为 `/audio/cctv1_2.m3u8`，实际清单分片为 `cctv1_audio/*.ts`。计划在后端 M3U 解析阶段跳过明显音频源，并在 TV 端对既有旧频道数据做同样过滤，避免默认播放音频源。
- 影响文件：`internal/services/iptv.go`、`internal/services/iptv_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行后端红灯测试、TV 红灯测试、定向验证、TV Debug 构建、乱码检查和提交范围检查。

## 2026-05-18 21:27 +0800
- 进度：完成 IPTV 音频专用源过滤；后端解析 M3U 时跳过明显音频分组、音频命名、`/audio/`/`_audio/` 路径和音频文件后缀，TV 端对 API 返回的旧频道数据也执行同样过滤并只在可播放视频频道中切台。TV 版本更新为 `0.1.6` / `versionCode=7`，`CONTEXT.md` 记录音频源过滤规则。
- 影响文件：`internal/services/iptv.go`、`internal/services/iptv_test.go`、`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvNavigationPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvViewModelTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `go test ./internal/services -run 'TestParseM3UPlaylistSkipsAudioOnlyEntries' -count=1` 因音频源未跳过失败；红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvNavigationPolicyTest' --tests 'com.chee.videos.feature.tv.TvIptvViewModelTest'` 因缺少过滤 helper 编译失败；实现后 `go test ./internal/services -run 'TestParseM3UPlaylist|TestBuildIPTV|TestIPTVService' -count=1` 通过，TV 同一定向命令通过；`go test ./internal/services ./internal/handlers -run 'Test.*IPTV|TestRegisterIncludesIPTVRoutes' -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。待执行乱码检查和提交范围检查。

## 2026-05-18 21:07 +0800
- 进度：完成 IPTV LibVLC 输出层和诊断增强；LibVLC `attachViews` 改为 TextureView 输出，新增 `TvIptv` 日志记录 event、vout、视频轨/音频轨数量、当前视频轨 codec/分辨率，TV 版本更新为 `0.1.5` / `versionCode=6`，并在 `CONTEXT.md` 记录后续无画面排查依据。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少 TextureView 绑定和诊断日志失败；实现后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。并行复跑时曾触发 Kotlin/Kapt 增量缓存竞争，执行 `cd android-tv-app && ./gradlew --stop && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过。待执行乱码检查和提交范围检查。

## 2026-05-18 21:06 +0800
- 进度：继续排查 TV App IPTV 切到 LibVLC 后仍无画面；最新日志显示 LibVLC 已将流识别为 TS 且 AAC 音频 packetizer 正常，但仍未看到视频 packetizer/decoder/vout 证据。计划先补 VLC 播放事件和轨道诊断日志，并将 LibVLC `attachViews` 改为 TextureView，区分是 Compose/Surface 输出层问题，还是当前频道 URL 未解析出视频轨。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`
- 验证：待执行红灯静态测试、TV 定向单测、TV Debug 构建、乱码检查与提交范围检查。

## 2026-05-18 21:05 +0800
- 进度：开始修复 TV App IPTV 仍然只有声音没有画面；用户日志显示 API 返回正常、AAC 音频解码器已初始化，但没有视频解码器初始化，且出现多个 `VideoCapabilities` 不支持提示。前一次 `texture_view` 修复未生效后，根因判断从 Compose/Surface 渲染转为 IPTV 频道视频编码兼容性；计划将 IPTV 播放页单独切换到 LibVLC 播放，保留其他长视频 Media3 播放器不变，并补依赖/实现静态回归测试、TV 版本号和技术沉淀。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`CONTEXT.md`、`plan.md`
- 验证：待执行 IPTV 定向单测、TV 全量单测、TV Debug 构建、中文乱码检查与提交范围检查。

## 2026-05-18 20:55 +0800
- 进度：完成 IPTV 播放器兼容性实现；TV App IPTV 播放页从 Media3 `PlayerView` 单独切换为 LibVLC `VLCVideoLayout`，播放直播源时关闭硬解并配置网络缓存，避免设备硬解不支持视频轨时只出声音；其他长视频 Media3 播放路径保持不变。TV 版本更新为 `0.1.4` / `versionCode=5`，`CONTEXT.md` 更新 IPTV 播放器兼容性沉淀。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_iptv_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackDependencyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackDependencyTest' --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因缺少 `org.videolan.libvlc.LibVLC` 失败；实现后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。并行复跑时曾触发 Kotlin/Kapt 增量缓存竞争，执行 `cd android-tv-app && ./gradlew --stop && ./gradlew --no-daemon :tv-app:assembleDebug` 串行重跑通过。待执行乱码检查和提交范围检查。

## 2026-05-18 20:34 +0800
- 进度：完成 TV App IPTV 有声音无画面修复；新增 IPTV 专用 Media3 `PlayerView` XML 布局并指定 `surface_type="texture_view"`，播放页改为 inflate 该布局，TV 版本更新为 `0.1.3` / `versionCode=4`，并在 `CONTEXT.md` 记录 Compose + IPTV 播放页的 TextureView 约定。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/main/res/layout/tv_iptv_player_view.xml`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlayerViewLayoutTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlayerViewLayoutTest'` 因布局缺失失败；修复后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 串行重跑通过；目标文件 Python 乱码扫描无命中。并行跑单测和构建时曾遇到 Hilt 增量产物竞争，串行重跑已通过。

## 2026-05-18 20:29 +0800
- 进度：开始排查 TV App IPTV 有声音无画面；HLS 依赖已生效且播放有声音，问题从依赖崩溃转为视频渲染/编码层。代码对比发现 IPTV 播放页在 Compose `AndroidView` 中使用默认 `PlayerView`，Media3 默认 `surface_type=surface_view` 且只能通过 XML 指定；计划先将 IPTV 专用播放器切到 `texture_view`，排除 TV/Compose 下 `SurfaceView` 黑屏问题，并补布局属性回归测试。
- 影响文件：`android-tv-app/tv-app/src/main/res/layout/tv_iptv_player_view.xml`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`plan.md`
- 验证：待执行红灯布局测试、TV 单测、TV Debug 构建和提交范围检查。

## 2026-05-18 20:24 +0800
- 进度：完成 TV App IPTV 点击闪退修复；新增 `media3-exoplayer-hls` 依赖，补充 HLS 工厂编译期回归测试，TV 版本更新为 `0.1.2` / `versionCode=3`，并在 `CONTEXT.md` 记录 M3U8/HLS 播放依赖约定。确认本次不纳入既有 `.codex/skills/*` 无关变更。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvIptvPlaybackDependencyTest.kt`、`CONTEXT.md`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvIptvPlaybackDependencyTest'` 因缺少 `androidx.media3.exoplayer.hls` 编译失败；修复后同命令通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；目标文件 Python 乱码扫描无命中。

## 2026-05-18 20:20 +0800
- 进度：开始修复 TV App 点击 `IPTV` 后闪退；崩溃根因是 IPTV 频道常见 `.m3u8` HLS 流触发 Media3 `DefaultMediaSourceFactory` 反射加载 `androidx.media3.exoplayer.hls.HlsMediaSource$Factory`，但 TV App 未打包 `media3-exoplayer-hls` 模块。计划补 HLS 模块依赖、增加编译期回归测试、更新 TV 版本号与技术沉淀。
- 影响文件：`android-tv-app/tv-app/build.gradle.kts`、`android-tv-app/tv-app/src/test/java/...`、`CONTEXT.md`、`plan.md`
- 验证：待执行 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、源码乱码检查与提交范围检查。

## 2026-05-18 19:35 +0800
- 进度：完成 IPTV v1 最终验证与收尾；确认本次提交只纳入后端 IPTV、Admin Web IPTV 管理页、TV App IPTV 播放页、TV 版本号、`CONTEXT.md` 技术沉淀和 `plan.md` 记录，不纳入主工作区既有 `.codex/skills/*` 无关变更。
- 影响文件：`migrations/0020_iptv_playlist.*.sql`、`internal/models/iptv.go`、`internal/repository/iptv_repository.go`、`internal/services/iptv.go`、`internal/handlers/iptv.go`、`internal/handlers/router.go`、`admin-web/src/*`、`android-tv-app/tv-app/*`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./... -count=1` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过（仅既有 AGP compileSdk/native strip 警告）；Python 源码乱码扫描无命中；`git diff --check` 通过。

## 2026-05-18 19:31 +0800
- 进度：完成 IPTV v1 三端核心实现；后端新增单全局 M3U 播放列表迁移、宽松解析、Admin 管理接口和 TV 频道接口；Admin Web 新增 `IPTV 管理` 页面；TV App 新增 `IPTV` 一级菜单、全屏直连播放页、频道分组列表、上下键循环换台和右键/返回键策略，并将 TV 版本更新为 `0.1.1` / `versionCode=2`。已补充 `CONTEXT.md` IPTV 术语与接口约定。
- 影响文件：`migrations/0020_iptv_playlist.*.sql`、`internal/models/iptv.go`、`internal/repository/iptv_repository.go`、`internal/services/iptv.go`、`internal/handlers/iptv.go`、`internal/handlers/router.go`、`admin-web/src/*`、`android-tv-app/tv-app/*`、`CONTEXT.md`、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers ./internal/repository -run 'Test.*IPTV|TestRegisterIncludesIPTVRoutes|TestIPTVPlaylistMigration' -count=1` 通过；`cd admin-web && npm run build` 通过（仅 Vite chunk size 警告）；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过。待执行完整 Go 验证、TV Debug 构建和乱码检查。

## 2026-05-18 19:14 +0800
- 进度：开始实现 IPTV v1；计划在隔离分支 `feature/iptv` 中完成后端单个全局 M3U 播放列表、Admin Web `IPTV 管理` 页面、TV App `IPTV` 入口与直连播放页；并按约定更新 TV 版本号与 `CONTEXT.md` 技术沉淀。当前主工作区存在无关 `.codex/skills/*` 变更，本任务在 `.worktrees/iptv-feature` 内执行，避免纳入无关改动。
- 影响文件：预计涉及 `migrations/*`、`internal/models/*`、`internal/repository/*`、`internal/services/*`、`internal/handlers/*`、`admin-web/src/*`、`android-tv-app/tv-app/*`、`CONTEXT.md`、`plan.md`
- 验证：待执行后端 IPTV 定向测试、`go test ./internal/services ./internal/handlers ./internal/repository -run 'Test.*IPTV|TestRegister' -count=1`、`cd admin-web && npm run build`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`、`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`、乱码检查。

## 2026-05-18 18:40 +0800
- 进度：完成新增开发约定；根级 `AGENTS.md` 已要求 App 功能修改同步更新对应 App 版本号，并要求每次功能更新追加 `CONTEXT.md` 技术沉淀；手机端与 TV 端模块级 `AGENTS.md` 已写明各自版本文件和递增规则；`CONTEXT.md` 已新增技术沉淀约定。
- 影响文件：`AGENTS.md`、`android-app/AGENTS.md`、`android-tv-app/AGENTS.md`、`CONTEXT.md`、`plan.md`
- 验证：`rg -n $'\uFFFD' AGENTS.md android-app/AGENTS.md android-tv-app/AGENTS.md CONTEXT.md plan.md` 无命中；文档约定变更无需构建/单测；确认既有 `.codex/skills/*` 工作区变更不是本任务改动，不纳入提交。

## 2026-05-18 18:38 +0800
- 进度：开始新增仓库开发约定；计划把“App 功能修改必须同步更新版本号”和“每次功能更新必须进行技术沉淀”写入根级与模块级 `AGENTS.md`，并在 `CONTEXT.md` 固化技术沉淀入口。
- 影响文件：`AGENTS.md`、`android-app/AGENTS.md`、`android-tv-app/AGENTS.md`、`CONTEXT.md`、`plan.md`
- 验证：待执行中文乱码检查与工作区范围检查。

## 2026-05-18 12:44 +0800
- 进度：完成 TV App 左侧分类首页重设计最终验证；后端保留 `/api/v1/tv/home` 未传 `kind` 的旧字段兼容，新增类型化 `kind/featured/recent_watching/recent_updates`；TV 端完成左侧一级菜单、分类页、搜索页、设置面板、`18+` 文案统一和海报墙标题编码修正。确认既有 `.codex/skills/*` 工作区变更不是本任务改动，不纳入提交。
- 影响文件：`CONTEXT.md`、`internal/models/app.go`、`internal/services/tv.go`、`internal/services/tv_auth.go`、`internal/handlers/tv.go`、`internal/services/tv_service_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、相关测试、`plan.md`
- 验证：`go test ./internal/services ./internal/handlers -run 'Test.*TVHome|Test.*TVCatalog' -count=1` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过；`rg -n $'\uFFFD' CONTEXT.md plan.md android-tv-app internal` 无命中。

## 2026-05-18 12:38 +0800
- 进度：完成 TV 左侧分类首页核心实现；后端 `/api/v1/tv/home` 新增可选 `kind=tv|movie|av` 并返回 `kind`、`featured`、`recent_watching`、`recent_updates`，旧 payload 字段保留；TV 端新增左侧菜单模型、`18+ -> av` 请求映射、搜索/设置菜单页、类型化首页分区和左栏/内容焦点策略，Shell 右上设置菜单改为首页右侧设置面板。
- 影响文件：`internal/models/app.go`、`internal/services/tv.go`、`internal/services/tv_auth.go`、`internal/handlers/tv.go`、`internal/services/tv_service_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvHomeNavigation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、相关测试、`plan.md`
- 验证：红灯阶段 Go 因缺少 `buildTypedTVHomePayload` 失败，TV 因缺少菜单模型、`kind` 请求和分区 helper 失败；实现后 `go test ./internal/services -run 'TestBuildTypedTVHomePayload|TestBuildTVHomePayload' -count=1` 通过，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvHomeNavigationTest'` 通过。待执行计划内完整验证。

## 2026-05-18 12:24 +0800
- 进度：开始实现 TV App 左侧分类首页重设计；计划先补后端 `kind` 参数与类型化首页 payload 红灯测试，再补 TV 菜单模型、ViewModel 请求映射、分类分区、设置面板和焦点策略测试，随后实现后端与 TV UI 最小改动并更新 `CONTEXT.md` 术语。
- 影响文件：`internal/services/tv_auth.go`、`internal/handlers/tv.go`、`internal/models/app.go`、`internal/services/tv_service_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`CONTEXT.md`、`plan.md`
- 验证：待执行后端红灯测试、TV 红灯测试、计划内 Go/Gradle 验证与中文乱码检查。

## 2026-05-18 10:18 +0800
- 进度：完成仓库级开发流程 skill 生成；`.codex/skills/repo-dev-workflow` 已沉淀本仓库计划记录、TDD 红灯、模块化验证、中文/编码、收尾提交范围控制等 `plan.md` 历史经验，并保留既有无关 `.codex/skills/*` 工作区变更不纳入本次提交。
- 影响文件：`.codex/skills/repo-dev-workflow/SKILL.md`、`.codex/skills/repo-dev-workflow/agents/openai.yaml`、`plan.md`
- 验证：使用临时 venv 安装 PyYAML 后运行 `python3 /Users/chee/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/repo-dev-workflow` 通过；`rg -n $'\uFFFD' .codex/skills/repo-dev-workflow plan.md AGENTS.md` 无命中；`git status --short` 已确认本次只暂存新 skill 与 `plan.md`。

## 2026-05-18 10:14 +0800
- 进度：开始根据 `plan.md` 历史经验生成仓库级开发流程 skill；计划新增 `.codex/skills/repo-dev-workflow`，沉淀计划记录、TDD 红灯、定向/全量验证、中文编码、提交范围确认等流程经验。
- 影响文件：`.codex/skills/repo-dev-workflow/SKILL.md`、`.codex/skills/repo-dev-workflow/agents/openai.yaml`、`plan.md`
- 验证：待执行 skill `quick_validate.py`、中文乱码静态检查、工作区范围检查。

## 2026-05-17 21:25 +0800
- 进度：完成“保留长视频音轨并支持 TV 端音轨选择”的完整验证；确认本次提交只纳入后端转码/元数据、TV 音轨选择/偏好及对应测试、`plan.md`，不纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关测试文件、`plan.md`
- 验证：`go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProbe|TestResolveProbe|TestBuildTranscodePlan' -v` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 21:20 +0800
- 进度：完成后端和 TV 音轨核心实现；ffmpeg 转码显式映射主视频与全部音频，音频继续转 AAC 但不再强制双声道；ffprobe/转码元数据新增音轨数量。TV 端新增音轨偏好存储、Repository/ViewModel 读写、Media3 当前音轨解析、音轨选择参数应用，以及复用字幕居中弹窗的“音轨”选择入口。
- 影响文件：`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode.go`、`internal/services/transcode_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`、相关测试文件、`plan.md`
- 验证：红灯阶段后端因 `AudioTrackCount`/`-map` 期望缺失失败，TV 端因音轨偏好与音轨选择 API 缺失失败；实现后 `go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProbe|TestResolveProbe' -v` 通过，`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.data.AppPreferencesStoreTest' --tests 'com.chee.videos.core.ui.AudioTrackSelectionTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.detail.DetailViewModelTest'` 通过。待执行计划内完整验证。

## 2026-05-17 21:09 +0800
- 进度：开始实现“保留长视频音轨并支持 TV 端音轨选择”；按既有计划先补红灯测试，覆盖 ffmpeg 多音轨 `-map` 参数、移除 `-ac 2`、ffprobe 音轨数量元数据、TV 音轨偏好存储、音轨选择纯逻辑，以及电视剧/长视频 ViewModel 音轨偏好读写。
- 影响文件：`pkg/ffmpeg/ffmpeg_test.go`、`internal/services/transcode_test.go`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/data/AppPreferencesStoreTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/AudioTrackSelectionTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`plan.md`
- 验证：待执行红灯测试、后端定向单测、TV 全量单测与 Debug 构建。

## 2026-05-17 20:50 +0800
- 进度：完成 TV App 播放记录与断点续播修复的全量验证；确认本次只修改 `android-tv-app` 播放历史相关代码、测试和 `plan.md`，未纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*History*' --tests 'com.chee.videos.feature.detail.*DetailViewModel*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 20:49 +0800
- 进度：完成 TV 播放历史核心实现；新增播放历史策略 helper，电视剧播放器增加 15 秒定时上报与 `ON_PAUSE` 补报，电影/AV 长视频播放器增加详情进度续播、15 秒定时上报、`ON_PAUSE` 与销毁补报，`DetailViewModel` 增加历史上报入口并过滤无效输入。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`、`plan.md`
- 验证：红灯阶段定向测试因缺少 `TvPlaybackHistoryPolicy` 与 `DetailViewModel.reportHistory` 编译失败；实现后 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*History*' --tests 'com.chee.videos.feature.detail.*DetailViewModel*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` 通过。待执行 TV 全量单测与 Debug 构建。

## 2026-05-17 20:43 +0800
- 进度：开始修复 TV App 播放记录与断点续播；计划新增 TV 播放历史纯逻辑 helper，补齐电视剧播放中/暂停上报，补齐电影与 AV 长视频播放历史上报和详情 `user_state.watch_seconds` 续播。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicy.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlaybackHistoryPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/detail/DetailViewModelTest.kt`、`plan.md`
- 验证：待执行红灯测试、TV 播放历史定向单测、TV 全量单测与 Debug 构建。

## 2026-05-17 13:09 +0800
- 进度：完成 TV 首页设置按钮焦点边界修复；设置按钮仍只在 `tv-home` 显示，按左/下稳定回到首页搜索框，按右/上由 `FocusRequester.Cancel` 拦截越界焦点搜索，未改变海报墙、详情页、播放器页 Back 与焦点行为。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellSettingsFocusPolicyTest.kt`、`plan.md`
- 验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.*Settings*'` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 13:07 +0800
- 进度：开始修复 TV 首页右上角设置按钮焦点边界；新增 Shell 设置按钮焦点策略单测并完成红灯验证，随后为首页内容区接入共享 `FocusRequester`，设置按钮按左/下返回首页搜索框，按右/上使用边界拦截。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellSettingsFocusPolicyTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvShellSettingsFocusPolicyTest'` 因缺少焦点策略类型编译失败；实现后同命令通过。待执行计划内定向 Settings 单测、TV 全量单测与 Debug 构建。

## 2026-05-17 10:27 +0800
- 进度：完成 TV App AV 板块恢复的最终验证；确认本次提交只包含后端 TV 聚合、Android TV AV 展示/路由/文案和对应回归测试，未纳入既有 `.codex/skills/*` 工作区变更。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`internal/services/tv_catalog_wall_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`plan.md`
- 验证：`rg -n "ExcludesAV|ExcludeAV|filtersAv|WithoutAv|does not promote av|NormalizesAvToMovie|搜索剧名|相关剧集|部剧集" ...` 无命中；`go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVSearchPayload|TestBuildTVCatalogWallPayload|TestBuildTVCatalogWallVideoItems' -v` 通过；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-17 10:26 +0800
- 进度：完成 TV AV 链路恢复实现；后端 TV 首页/搜索/海报墙重新查询并返回 AV，TV 首页重新读取 AV shelf、保留 AV 搜索结果与继续观看，AV 海报墙和长视频路由保留 `videoType=av`，详情/精选/焦点/文案恢复 AV 分支。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`internal/services/tv_catalog_wall_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`、`plan.md`
- 验证：红灯阶段后端 `TestBuildTVSearchPayloadIncludesAVContent` 因 AV 未加入结果失败，TV 定向测试因缺少 AV 参数/焦点/路由分支编译失败；实现后 `go test ./internal/services -run 'TestBuildTVHomePayload|TestBuildTVSearchPayload|TestBuildTVCatalogWallPayload|TestBuildTVCatalogWallVideoItems' -v` 通过，TV feature 定向单测通过。待执行完整后端与 TV 验证。

## 2026-05-17 10:17 +0800
- 进度：开始恢复 TV App 的 AV 板块；计划按后端 TV 首页/搜索/海报墙、TV 首页 shelf/焦点/精选、长视频 AV 路由与详情文案、回归测试四部分执行。
- 影响文件：`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`internal/services/tv_catalog_wall_test.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/*`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/*`、`plan.md`
- 验证：待执行红灯测试、后端定向单测、TV 单测与构建。

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

## 2026-05-16 21:45 +0800
- 进度：开始实现手机端搜索播放页点赞/收藏能力；已按 TDD 补充搜索 ViewModel 红灯用例，新增详情预取、点赞、收藏、忙碌状态和认证失效处理，并在搜索播放覆盖层右侧动作栏接入喜欢/收藏按钮。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchViewModel.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/test/java/com/chee/videos/feature/shortsearch/ShortSearchViewModelStateTest.kt`、`plan.md`
- 验证：红灯阶段 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.feature.shortsearch.ShortSearchViewModelStateTest` 因缺少 `toggleLike`、`toggleFavorite`、`ensureDetailLoaded` 和详情状态字段失败；实现后同命令通过。待执行完整手机端单测与 Debug 构建。

## 2026-05-16 12:09 +0800
- 进度：完成 TV APP 电视剧目录去除 AV 内容的完整验证，确认 TV 专属 AV shelf/文案静态检查无命中，并准备提交。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`plan.md`
- 验证：`rg -n "AV 精选|全部 AV|av-shelf|AV_ITEM|继续播放 AV|\\\"AV\\\"" android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv internal/services/tv_auth.go` 无命中；`go test ./internal/services ./internal/handlers -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。

## 2026-05-16 12:07 +0800
- 进度：排查并修复 TV APP 电视剧目录混入 AV 内容；TV 端目录状态过滤 AV 首页/搜索/继续观看数据，移除 AV shelf、AV 焦点兜底和 AV 精选兜底，长视频路由将 AV 类型归一为电影；后端 TV 首页、TV 搜索、继续观看兜底和 TV 海报墙不再查询/返回 AV 内容。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`、`internal/services/tv_auth.go`、`internal/services/tv_auth_test.go`、`internal/services/tv_service_test.go`、`plan.md`
- 验证：已执行红灯验证，TV 端新增用例和后端 TV 搜索用例在修复前失败；修复后定向 TV feature 单测与 `go test ./internal/services -count=1` 通过。待执行完整 TV 单测与构建。

## 2026-05-16 11:47 +0800
- 进度：完成手机端 UI 圆角全局收敛后的静态检查、单测和 Debug 构建验证，准备提交本次手机端 UI 改动。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/AppChromeDensitySpecTest.kt`、`plan.md`
- 验证：`rg "RoundedCornerShape\\((1[3-9]|2[0-9])\\.dp|topStart = (1[3-9]|2[0-9])\\.dp" android-app/app/src/main/java/com/chee/videos` 无命中；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-16 11:45 +0800
- 进度：执行手机端 UI 圆角全局收敛，将 `android-app` 通用卡片/分区圆角收敛为 `8dp`，底部导航与短视频详情 sheet 顶部圆角收敛为 `12dp`，并同步收紧手机端卡片、封面、输入框、错误提示、播放器非圆形浮层等局部圆角。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`、`android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/AppChromeDensitySpecTest.kt`、`plan.md`
- 验证：`rg "RoundedCornerShape\\((1[3-9]|2[0-9])\\.dp|topStart = (1[3-9]|2[0-9])\\.dp" android-app/app/src/main/java/com/chee/videos` 无命中；待执行 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest`、`cd android-app && ./gradlew --no-daemon :app:assembleDebug`。

## 2026-05-16 11:09 +0800
- 进度：收尾检查演员页界面文案，移除空状态英文状态词并补齐作品类型中文标签；复跑完整后端与 Android 验证。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`plan.md`
- 验证：`go test ./internal/handlers ./internal/repository ./internal/services -count=1`、`go vet ./...`、`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest`、`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 均通过。

## 2026-05-16 11:06 +0800
- 进度：开始实现手机端演员页与作品瀑布流；后端新增 App 演员详情/作品分页 DTO、仓储查询、服务方法与 `GET /api/v1/actors/:id` 路由，手机端新增演员详情 DTO、API/repository 调用、actor 路由、演员页 ViewModel/页面，并将 AV 详情演员卡按稳定 actor id 接入跳转。
- 影响文件：`internal/models/app.go`、`internal/repository/actor_repository.go`、`internal/services/app.go`、`internal/handlers/actor.go`、`internal/handlers/router.go`、`internal/repository/actor_repository_test.go`、`internal/handlers/recommend_test.go`、`android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`、`android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`、`android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`、`android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`、`android-app/app/src/main/java/com/chee/videos/feature/detail/AvDetailPresentation.kt`、`android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorRoutes.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailViewModel.kt`、`android-app/app/src/main/java/com/chee/videos/feature/actor/ActorDetailScreen.kt`、`android-app/app/src/test/java/com/chee/videos/feature/detail/AvDetailPresentationTest.kt`、`android-app/app/src/test/java/com/chee/videos/feature/actor/ActorRoutesTest.kt`、`android-app/app/src/test/java/com/chee/videos/feature/actor/ActorDetailViewModelTest.kt`、`android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`plan.md`
- 验证：`go test ./internal/repository ./internal/handlers -run 'TestBuildActorWorks|TestRegisterIncludesImageCollectionRoutes' -count=1`、`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest' --tests 'com.chee.videos.feature.actor.ActorRoutesTest' --tests 'com.chee.videos.feature.actor.ActorDetailViewModelTest'`、`go test ./internal/handlers ./internal/repository ./internal/services -count=1`、`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest`、`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 均通过。

---

更早历史（2026-05-16 之前的所有条目，含旧 `### [YYYY-MM-DD HH:MM] Title` 格式与 `## Entry Template` 元数据段）已归档至 `plan.archive.md`，2026-05-25 拆分。
## 2026-05-25 18:08 +0800
- 进度：完成 `tasks/2026-05-25-tv-long-form-libvlc-migration` 的代码层收尾和验证收口。TV 长视频偏好已从 trackId 迁到 `language + type` 持久化，音轨/字幕恢复仍按当前 media 的临时 track id 运行；后端 ASS 原文落盘与内嵌抽取增加了安全清洗。`CONTEXT.md` 已同步长期约定，任务相关定向测试和全量验证已通过。
- 影响文件：`CONTEXT.md`、`plan.md`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/{data,model,repository,ui}/**`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/{detail,tv}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/{data,player,ui}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/{detail,tv}/**`、`android-tv-app/tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerLibVlcTest.kt`、`internal/services/subtitle.go`、`internal/services/subtitle_test.go`、`pkg/ffmpeg/ffmpeg.go`、`pkg/ffmpeg/ffmpeg_test.go`
- 验证：`go test ./internal/services ./pkg/ffmpeg -count=1` 通过；`go test ./... -count=1` 通过；`cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest` 通过；`cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过；`git diff --check` 通过；`rg -n $'\\uFFFD' ...` 无输出。

## 2026-05-25 17:42 +0800
- 进度：继续收尾 `tasks/2026-05-25-tv-long-form-libvlc-migration`。补齐 review 里尚未闭合的两项：TV 长视频音轨/字幕偏好从 trackId 语义迁到 `language + type` 持久化（运行时仍按当前 media track id 选轨），并给 ASS 原文落盘/抽取增加轻量安全清洗（去除本地路径 `\fn` 覆盖、钳制 `\fad` 参数）。同步扩展 `AppPreferencesStoreTest`、`AudioTrackSelectionTest`、`LongFormSubtitleSupportLibVlcTest`、`TvSeriesPlayerViewModelTest`、`DetailViewModelTest`、`internal/services/subtitle_test.go`，并在 `CONTEXT.md` 追加长期约定。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/{data,model,repository,ui}/**`、`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/{detail,tv}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/core/{data,ui}/**`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/{detail,tv}/**`、`internal/services/subtitle.go`、`internal/services/subtitle_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：后端定向测试 `go test ./internal/services -run 'TestSubtitle|TestSanitizeAss' -count=1` 通过；TV 定向测试 `cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.data.AppPreferencesStoreTest' --tests 'com.chee.videos.core.ui.AudioTrackSelectionTest' --tests 'com.chee.videos.core.ui.LongFormSubtitleSupportLibVlcTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.feature.detail.DetailViewModelTest'` 通过；全量验证待执行。

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

## 2026-05-26 00:55 +0800
- 进度：用户报告"TV 无法播放视频 + 另一台电脑预览卡顿"。诊断时绕路（先怀疑 Mac 内存、外接盘 I/O、自己的 commit、视频文件特定损坏），逐一回滚到 LibVLC 迁移基线 0.1.68 之后才用 curl 发现服务端对 `/api/v1/videos/:id/source?profile=compat` 返回 **HTTP 200 + body `{"code":401,"msg":"missing or invalid authorization header"}`**——LibVLC 走 libcurl 不带 Authorization Bearer header（[[response.Error 包装 401 为 HTTP 200]] 隐藏在 gin log），LibVLC 拿 70 字节 JSON 当视频字节流读 → 进入 `reading while paused` 半死态。**视频 / 音频 / 字幕在 0.1.68 LibVLC 迁移基线就完全没真正跑过**（迁移期 review.md §0 只有自动化 / spec audit / 静态扫描，没有"真机看到第一帧"的硬验收，DONE.md 误标"实测通过"）。
- Quick fix 落地：①服务端 `internal/middleware/auth.go` 的 `AuthMiddleware` 在 Authorization header 缺失时 fallback 从 `?access_token=` query 读 token；②TV 端 `applyLongFormMediaSource` 与字幕 addSlave 路径用新增的 `appendAccessTokenQuery(url, accessToken)` 把 token 拼进 URL；③LibVLC `MediaPlayer.addSlave` 改用 `Uri` 重载（String 重载把 `http://...` 当文件路径报 `cannot open file /http:/...`）；④audio LaunchedEffect 不再在 `resolvedSelection == null` 时写 `player.audioTrack = -1`（libvlc-all 3.x 的 -1 是关音频，不是自动）；⑤ `LongFormVideoPlayer` 根 Box 的 `onPreviewKeyEvent` 首行加 overlay 透传 guard（`if (currentFocusGuardInput.anyOverlayVisible()) return false`），修复续播卡 / picker 按钮 CENTER 被 player 路由吃掉的回归；⑥`TvLongFormPlayerScreen` 字幕选择链尾兜底加 `?: resolveInitialSubtitleTrackId(detail.subtitleTracks)`，无 preference 时也能默认挑 isDefault 字幕。
- 教训沉淀（写入 .claude memory）：[[Curl failing client URL first]]、[[response.Error 包装 401 为 HTTP 200]]、[[TV LibVLC 不经过 OkHttp interceptors]]、[[内核替换类 task 必须真机端到端验证]]、[[LibVLC MediaPlayer.addSlave 必须用 Uri 重载]]、[[Player 根 onPreviewKeyEvent 遇 overlay 必须 passthrough]] 共 6 条。CONTEXT.md 同步新增 5 条术语（TV LibVLC playback URL 携带 access_token / MediaPlayer.addSlave 用 Uri 重载 / Player onPreviewKeyEvent overlay 透传 / LibVLC audioTrack=-1 是关音频）。
- 影响文件：`internal/middleware/auth.go`、`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`、`com/chee/videos/core/ui/LongFormVideoPlayer.kt`、`com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`curl -is 'http://192.168.1.24:8080/.../source?profile=compat&access_token=invalid'` 返回 `{"code":401,"msg":"invalid access token"}`（fallback 路径有效）；TV 真机 0.1.71 实测视频可播 + 音频可听 + 字幕可见 + 续播卡按钮可交互。版本 `0.1.71 / 71`。

## 2026-06-19 11:36 +0800
- 进度：完成 TV 单片长视频播放器软重试语义审查并修正一处边界回退。已确认并修复 `BACK` 取消当前重试时不应把播放会话回退成未开始；现在取消只终止本次准备、清理错误并回写 `已取消重试` 轻提示，继续沿用已出现首帧后的承接播放态。同步把 TV 版本递增到 `0.1.121` / `versionCode=121`，并补了取消语义的源码断言。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormPlayerSoftRetrySpecTest.kt`、`android-tv-app/tv-app/build.gradle.kts`、`plan.md`
- 验证：受限于当前 sandbox，Gradle 先后卡在默认 `~/.gradle` 锁文件权限、wrapper 网络下载和 daemon 端口绑定；已改用本地缓存拷贝到 `/private/tmp/codex-gradle-home` 继续，但完整 `:tv-app:testDebugUnitTest` 仍未跑完。后续需在可绑定本地端口的环境里复跑 `cd android-tv-app && GRADLE_USER_HOME=/private/tmp/codex-gradle-home ./gradlew --no-daemon -Dkotlin.compiler.execution.strategy=in-process :tv-app:testDebugUnitTest`，再补 `assembleDebug`、`git diff --check` 和乱码扫描。

## 2026-06-25 15:25 +0800
- 进度：完成 TV 短视频单条播放失败非阻塞化改造。原行为：单条播放失败时弹阻塞卡片（重试/返回首页），根 `onPreviewKeyEvent` 顶部 `if (playbackErrorMessage != null) return false` 短路吞掉所有遥控键，导致上/下无法切条、用户被卡在失败视频上。改造后（仅单条播放失败路径，首屏加载失败/空态路径保留原阻塞卡片）：①删除顶部短路；②上/下照常 `movePrevious`/`moveNext`，切条后由既有 `LaunchedEffect(currentVideoId)`（`:238`）清 `playbackErrorMessage`；③OK/中键/PLAY_PAUSE 在失败态执行重试体（`playbackErrorMessage=null; renderedVideoId=null; hasEndedAtCurrentVideo=false; playbackRetryNonce+=1`，与旧重试按钮逐行一致，`return true` 防落暂停逻辑），非失败态保持原暂停逻辑；④左/右/REWIND/FF 在失败态 `return true` 静默消费不 seek；⑤BACK 退出不变；⑥把 `if (!playbackErrorMessage.isNullOrBlank()) { Box(0.56f scrim) { TvShortFeedProblemState } }` 替换为非可聚焦居中 chip「播放失败」（`AnimatedVisibility(visible = playbackErrorMessage != null)`，复用暂停指示器 `Surface(CircleShape, AppChrome.Surface alpha 0.82)` + `Row(Icon(Refresh)+Text)` 样式，去全屏遮罩，不夺焦不吞键）；⑦`onPlayerError` 置 `playbackErrorMessage` 时同步 `showCenterIndicator=false; centerIndicatorHideJob?.cancel()`，覆盖「先暂停（700ms 自动隐藏窗口内）随后失败」的叠显场景，保证失败提示干净接管中央位。不新增任何状态变量/定时器/Job/常量，不改任何 `LaunchedEffect`。`TvShortFeedProblemState`/`TvShortFeedStateButton` 保留（首屏失败路径仍用）。
- 影响文件：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`（版本 `0.1.123 / 123` → `0.1.124 / 124`）、`CONTEXT.md`（修订 `TV 短视频单条失败留在当前页` 为非阻塞模型 + `TV 短视频保留中央播放暂停提示` 加 OK 语义分态交叉引用）、`docs/superpowers/specs/2026-06-25-tv-short-feed-nonblocking-failure-design.md`（新设计规格）、`plan.md`。
- 验证：`cd android-tv-app && ./gradlew :tv-app:compileDebugKotlin :tv-app:testDebugUnitTest` BUILD SUCCESSFUL（exit 0），既有 `TvShortFeedScreenSpecTest` / `TvShortFeedViewModelTest` 全绿、无需改测试；子代理两轮独立评审（首轮发现 BLOCKER 版本号未升 + IMPORTANT 暂停/失败提示 700ms 叠显，均已修复并复审 CLEAN）；待真机/模拟器手测：失败态上/下切条、OK 重试、左/右静默、BACK 退出、首屏失败仍弹阻塞卡片、正常 OK 暂停样式不变。

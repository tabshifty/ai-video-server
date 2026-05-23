# plan.md

本文件用于增量记录”计划与修改”，不得覆盖历史记录，只能追加。

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

### [2026-05-16 09:23 CST] 修复 TV 遥控器返回键行为完成
- Type: `implementation`
- Summary:
  - 完成 TV Shell Back 策略与播放器二次返回确认，保持 `tv-home` 默认退出行为，播放页由自身确认逻辑处理。
  - 补充 Shell 路由策略与播放器确认窗口单测，覆盖二级页 pop、首页不拦截、播放页不走 Shell 策略、2 秒确认窗口。
  - 执行 TV 定向 Back 单测、TV 全量单测与 Debug 构建。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlayerBackConfirmTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.*Back*' --tests 'com.chee.videos.feature.tv.*Back*'` passed.
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` passed.
  - `cd android-tv-app && ./gradlew :tv-app:assembleDebug` passed.

### [2026-05-16 09:20 CST] 修复 TV 遥控器返回键行为实现中
- Type: `implementation`
- Summary:
  - 新增 TV Shell Back 策略测试与播放器返回确认 helper 测试，先确认缺少策略/helper 时定向测试失败。
  - `TvAuthenticatedNav` 接入 Shell 层 `BackHandler`：海报墙、长视频详情、电视剧详情执行 `popBackStack()`，首页与播放页不由 Shell 拦截。
  - 长视频播放页与电视剧播放页接入系统 Back 二次确认：首次提示“再按一次返回”，确认窗口内第二次返回才调用既有 `onBack()`。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlayerBackConfirmTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.*Back*' --tests 'com.chee.videos.feature.tv.*Back*'` passed.
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`。
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:assembleDebug`。

### [2026-05-16 09:16 CST] 修复 TV 遥控器返回键行为计划
- Type: `plan`
- Summary:
  - 仅修改 `android-tv-app` 独立 TV 工程，修复遥控器 Back 在二级页直接退出应用的问题。
  - TV Shell 增加普通二级页返回策略：首页不拦截，海报墙、长视频详情、电视剧详情执行 `popBackStack()`，播放页交给播放器自身处理。
  - 长视频播放页与电视剧播放页增加 2 秒内二次返回确认，首次返回只显示底部提示。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppBackPolicyTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPlayerBackConfirmTest.kt`
  - `plan.md`
- Verification:
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.*Back*' --tests 'com.chee.videos.feature.tv.*Back*'`。
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`。
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:assembleDebug`。

### [2026-05-16 00:41 CST] 手机端页面切换 iOS 动画完成
- Type: `implementation`
- Summary:
  - 手机端 `NavHost` 接入 iOS 风格左右滑动转场，底部 tab 按显示顺序判断左右方向，二级页面 push/pop 使用导航栈方向。
  - 新增内部导航转场方向与动效规格，保持路由结构、页面参数、ViewModel 和业务逻辑不变。
  - 补充转场方向规格测试，覆盖 tab 顺序、二级页面 push、pop 返回和默认方向。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppNavigationConfig.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppNavigationTransitionSpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.AppNavigationTransitionSpecTest` passed.
  - `cd android-app && ./gradlew :app:testDebugUnitTest` passed.
  - `cd android-app && ./gradlew :app:assembleDebug` passed.

### [2026-05-16 00:39 CST] 手机端页面切换 iOS 动画计划
- Type: `plan`
- Summary:
  - 仅调整手机端 `NavHost` 页面转场动画，不改 TV 工程、路由结构或业务逻辑。
  - 底部 tab 按 `首页 -> 搜索 -> 图集 -> 我的` 顺序左右切换，二级页面 push 从右进入，pop 返回向右退出。
  - 新增内部导航转场方向规格与单测，再接入 Compose Navigation 动画。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppNavigationConfig.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppNavigationTransitionSpecTest.kt`
  - `plan.md`
- Verification:
  - 待执行 `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.AppNavigationTransitionSpecTest`。
  - 待执行 `cd android-app && ./gradlew :app:testDebugUnitTest`。
  - 待执行 `cd android-app && ./gradlew :app:assembleDebug`。

### [2026-05-15 23:08 CST] 手机端界面轻收紧完成
- Type: `implementation`
- Summary:
  - 将手机端 `AppChrome` 通用卡片圆角从 `22dp` 收紧为 `18dp`，区块圆角从 `18dp` 收紧为 `14dp`，保留胶囊按钮形状不变。
  - 长视频播放器新增本地 chrome 规格，统一控制 seek 预览、中心反馈、顶部栏、底部栏的圆角、外边距、内边距和控制按钮尺寸。
  - 新增纯规格测试覆盖本次视觉密度目标，避免后续手机端圆角和播放器浮层参数漂移。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppChromeDensitySpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.AppChromeDensitySpecTest` passed.
  - `cd android-app && ./gradlew :app:testDebugUnitTest` passed.
  - `cd android-app && ./gradlew :app:assembleDebug` passed.

### [2026-05-15 23:06 CST] 手机端界面轻收紧计划
- Type: `plan`
- Summary:
  - 仅调整手机端通用形状与长视频播放器浮层密度，保持 TV 工程不变。
  - 将 `AppChrome` 卡片、区块圆角整体下调一档，保留胶囊形状语义。
  - 长视频播放器中心反馈、seek 预览、顶部栏、底部栏改用更紧凑的圆角、外边距、内边距和按钮尺寸，并用纯规格测试固定目标值。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppChromeDensitySpecTest.kt`
  - `plan.md`
- Verification:
  - 待执行 `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.AppChromeDensitySpecTest`。
  - 待执行 `cd android-app && ./gradlew :app:testDebugUnitTest`。
  - 待执行 `cd android-app && ./gradlew :app:assembleDebug`。

### [2026-05-15 20:47 CST] 增加手机端长视频全屏亮度音量手势
- Type: `implementation`
- Summary:
  - 为手机端长视频播放器增加全屏纵向轴向锁定手势：左半屏调节当前窗口亮度，右半屏调节系统媒体音量。
  - 将原横向快进/快退改为同一手势识别器内的横向锁定逻辑，保留 seek 预览、结束 seek、单击/双击/长按行为。
  - 新增长视频手势纯逻辑规格，覆盖全屏启用、左右区域、上下滑方向、亮度默认值、音量百分比和反馈文案 clamp。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormGestureSpec.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/LongFormGestureSpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.LongFormGestureSpecTest` passed.
  - `cd android-app && ./gradlew :app:testDebugUnitTest` passed.
  - `cd android-app && ./gradlew :app:assembleDebug` passed.

### [2026-05-15 16:54 CST] 修复 TV 海报焦点裁切验证完成
- Type: `implementation`
- Summary:
  - 将 `MainDispatcherRule` 提取到 `core/testing`，TV 测试继续通过 typealias 复用，避免 core/home 测试依赖 feature.tv 包。
  - HomeViewModel 单测绑定同一测试调度器与 DataStore scope，修复完整套件下后台协程调度不一致导致的偶发失败。
  - 重新执行完整 TV 单测与 Debug 构建。
- Changed Files:
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/testing/MainDispatcherRule.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` passed.
  - `cd android-tv-app && ./gradlew :tv-app:assembleDebug` passed.

### [2026-05-15 16:32 CST] 修复 TV 海报焦点裁切完成
- Type: `implementation`
- Summary:
  - 完成 TV 焦点安全规格、首页 AV/普通海报、TV 首页横向海报架、海报墙网格卡片的防裁切实现。
  - 补充焦点安全规格测试、首页布局规格测试、TV 目录横向架规格测试、海报墙网格规格测试。
  - 修复 TV 单测套件中 Main dispatcher 初始化不稳定的问题，确保新增测试和既有 Home/ViewModel/DataStore 测试可同套件运行。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeFocusLayoutSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/data/AppPreferencesStoreTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` passed.
  - `cd android-tv-app && ./gradlew :tv-app:assembleDebug` passed.

### [2026-05-15 16:11 CST] 修复 TV 海报焦点裁切实现中
- Type: `implementation`
- Summary:
  - 新增 TV 焦点安全规格与首页、TV 目录、海报墙布局规格测试，先确认缺少规格时测试失败，再补齐实现。
  - 首页 AV 海报和普通视频卡片接入 `tvFocusableGlow`，卡片外层预留焦点安全空间，内部图片区域单独裁圆角。
  - TV 首页横向海报架、查看更多卡片和海报墙卡片改为外层焦点容器，增加首尾/上下 padding 与 item spacing，避免缩放和描边被边界裁掉。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/TvFocusSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeFocusLayoutSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusLayoutSpecTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallFocusLayoutSpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvFocusSpecTest' --tests 'com.chee.videos.feature.home.HomeFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvCatalogFocusLayoutSpecTest' --tests 'com.chee.videos.feature.tv.TvPosterWallFocusLayoutSpecTest'` passed.
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`。
  - 待执行 `cd android-tv-app && ./gradlew :tv-app:assembleDebug`。

### [2026-05-15 12:29 CST] 调整手机端 AV 播放详情页
- Type: `implementation`
- Summary:
  - 手机端 AV 详情页在有播放地址时自动进入播放会话，保持非 AV 详情页现有行为不变。
  - 移除 AV 播放器区域内的标题/番号叠加文案，将内容顺序调整为播放器、互动操作、作品信息、演员、简介、标签。
  - 互动区改为小图标加数字展示播放、点赞、收藏和不喜欢；演员区改为横向可滑动列表并放大头像。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/AvDetailLayoutSpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.*'` passed.

### [2026-05-15 11:26 CST] 修复手机端短视频搜索播放保活
- Type: `implementation`
- Summary:
  - 为手机端短视频搜索播放浮层接入现有 `KeepScreenOnEffect`，按 `isPlayerActuallyPlaying` 在实际播放时保持屏幕常亮。
  - 新增播放器保活覆盖测试，扫描手机端 `ExoPlayer` 页面，防止后续遗漏保活接入。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/PlayerKeepScreenOnCoverageTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.PlayerKeepScreenOnCoverageTest` passed.

### [2026-05-14 18:43 +0800] 修复 AV 手动保存长标题完成
- Type: `implementation`
- Summary:
  - 新增 `0019` 迁移，将 `videos.title` 从 `VARCHAR(200)` 放宽为 `TEXT`，避免 AV 手动确认保存超长标题时报 PostgreSQL `SQLSTATE 22001`。
  - 回滚迁移使用 `LEFT(title, 200)`，确保 down migration 可执行且行为明确。
  - 通过现有 `ConfirmAV` 翻译链路继续将标题与简介写入中文字段，未改动接口或前端 payload。
  - 增加迁移保护测试，扫描 `0019` up/down SQL，防止以后把 `videos.title` 重新收窄。
- Changed Files:
  - `migrations/0019_video_title_text.up.sql`
  - `migrations/0019_video_title_text.down.sql`
  - `internal/repository/migrations_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/repository -count=1` passed.
  - `go test ./internal/services -run TestConfirmAVStoresLocalizedFieldsAndOriginalFields -count=1` passed.
  - `go test ./internal/services ./internal/handlers -count=1` passed.
  - `go vet ./...` passed.

### [2026-05-14 18:41 +0800] 修复 AV 手动保存长标题计划
- Type: `plan`
- Summary:
  - 新增迁移将 `videos.title` 从 `VARCHAR(200)` 放宽为 `TEXT`，避免 AV 手动确认保存长标题时报 `SQLSTATE 22001`。
  - 回滚迁移使用 `LEFT(title, 200)` 保证可执行，并明确回滚时会截断超长标题。
  - 保持现有 `ConfirmAV` 中文翻译落库链路不变，补充迁移保护测试并执行现有 AV 确认翻译测试。
- Changed Files:
  - `migrations/0019_video_title_text.up.sql`
  - `migrations/0019_video_title_text.down.sql`
  - `internal/repository/migrations_test.go`
  - `plan.md`
- Verification:
  - 待执行 `go test ./internal/repository -count=1`。
  - 待执行 `go test ./internal/services -run TestConfirmAVStoresLocalizedFieldsAndOriginalFields -count=1`。
  - 待执行 `go test ./internal/services ./internal/handlers -count=1`。
  - 待执行 `go vet ./...`。

### [2026-05-14 17:36 +0800] 修复 Jav321 横版海报与演员解析
- Type: `implementation`
- Summary:
  - 将 `jav321` 搜索页与详情页解析统一到专用 DOM parser，避免搜索页/确认页字段不一致。
  - 补齐 `<b>品番/番号</b>` 标签式番号解析，避免标题兜底把 `SSIS-001` 归一成 `SSIS-1`。
  - 补齐 `.img-responsive` 多图解析：首图作为 `thumb_url`，最后一张非首图作为横版 `poster_url`；同时从 `/star/`、`/heyzo_star/` 链接和标签文本兜底提取演员。
  - 新增搜索预览与 `ConfirmAV` 详情确认的 Jav321 回归断言。
- Changed Files:
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_av_mdcx_sites_test.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestMDCxMigratedSitesSearchCandidates|TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites/jav321' -count=1` passed.
  - `go test ./internal/services -count=1` passed.
  - `go vet ./...` passed.

### [2026-05-14 17:25 +0800] Jav321 横版海报与演员刮削修复计划
- Type: `plan`
- Summary:
  - 定位 `jav321` 搜索返回页解析缺字段的问题，优先对照 `references/mdcx` 的 Jav321 参考实现补齐横版大海报与演员解析。
  - 先补回归测试覆盖 `.img-responsive` 多图场景和 `/star/` 演员链接，再修改当前 Go 刮削器。
  - 保持改动限定在 Jav321 解析、相关单测和计划记录。
- Changed Files:
  - `plan.md`
- Verification:
  - 待执行 Jav321 定向测试。

### [2026-05-14 15:26 +0800] 忽略本地 storage 目录
- Type: `implementation`
- Summary:
  - 将 `storage/` 加入 `.gitignore`，避免本地存储目录或其软连接在 `git status` 中显示为未跟踪文件。
  - 保持真实视频存储目录不变，仅调整 Git 忽略规则。
- Changed Files:
  - `.gitignore`
  - `plan.md`
- Verification:
  - Documentation/config-only change; no build/test required.
  - `git status --short --ignored storage` showed `!! storage/`.

### [2026-05-14 15:11 +0800] AV 刮削封面转码覆盖修复验证完成
- Type: `verification`
- Summary:
  - 完成 AV 刮削封面转码覆盖修复后的定向与全量 Go 验证。
  - 确认 `internal/queue` 新增决策测试覆盖 AV 有封面、AV 无封面兜底、AV 空默认路径恢复、非 AV 维持旧行为四类场景。
  - 当前环境未安装 `golangci-lint`，未执行该项；已执行 `go vet ./...` 作为 Go 静态检查补充。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/queue -run 'TestResolveTranscodePersistence' -count=1` passed.
  - `go test ./internal/queue` passed.
  - `go test ./internal/services ./internal/queue ./internal/handlers ./internal/repository` passed.
  - `go vet ./...` passed.
  - `go test ./... -count=1` passed.

### [2026-05-14 15:10 +0800] 修复 AV 上传后刮削封面被转码覆盖
- Type: `implementation`
- Summary:
  - 确认根因为 AV 自动刮削先写入封面缩略图与 `poster_*` metadata，后续转码完成时 `UpdateTranscodeResult` 使用视频截图和转码 metadata 整体覆盖，导致默认封面与封面变体信息丢失。
  - 转码落库前新增内部决策：非 AV 保持使用转码截图与转码 metadata；AV 已有刮削封面 metadata 时保留现有封面路径并合并 metadata；AV 无封面 metadata 时使用转码截图兜底，同时保留 `scrape_error` 等诊断信息。
  - AV metadata 合并以旧 metadata 为基础补入转码播放字段，并在发现刮削封面时回写 `poster_*`、`thumb_url`、`poster`、`thumb` 等封面字段，避免同名转码字段覆盖封面决策。
- Changed Files:
  - `internal/queue/tasks.go`
  - `internal/queue/tasks_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/queue -run 'TestResolveTranscodePersistence' -count=1` passed.
  - `go test ./internal/queue` passed.
  - `go test ./internal/services ./internal/queue ./internal/handlers ./internal/repository` baseline passed before implementation.
  - Full planned verification pending after this entry.

### [2026-05-14 13:58 +0800] 电影电视剧长视频码率策略验证补全
- Type: `implementation`
- Summary:
  - 完成长视频码率策略调整后的汇总验证，确认 `movie`、`episode` 新封顶值与 `av` 旧阈值并存逻辑生效。
  - 确认 ffmpeg bitrate mode 的 `-maxrate 2x`、`-bufsize 4x` 参数变更已覆盖到服务层与 ffmpeg 层回归测试。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestDecideVideoBitrate|TestBuildTranscodePlan' -count=1` passed.
  - `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)' -count=1` passed.
  - `go test ./pkg/ffmpeg ./internal/services -count=1` passed.

### [2026-05-14 13:52 +0800] 电影电视剧长视频码率策略调整完成
- Type: `implementation`
- Summary:
  - 电影 `movie` 和剧集 `episode` 的 HEVC 长视频转码改为新的上限封顶策略：1080p 最高 `5000k`，4K 最高 `12000k`，低于上限时保持源码率，720p 等其他分辨率继续保持源码率。
  - 非长视频 `av`、`short` 继续沿用既有分辨率目标码率阈值，仅同步使用新的 ffmpeg bitrate mode 参数倍数。
  - ffmpeg bitrate mode 输出参数更新为 `-b:v <target>`、`-maxrate <2x>`、`-bufsize <4x>`，并补齐对应回归测试。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestDecideVideoBitrate|TestBuildTranscodePlan' -count=1` passed.
  - `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)' -count=1` passed.
  - 全量定向验证待继续执行。

### [2026-05-14 13:52 +0800] 电影电视剧长视频码率策略调整计划
- Type: `plan`
- Summary:
  - 仅调整 `movie` 和 `episode` 的长视频 HEVC 码率上限：1080p 封顶 `5000k`，4K 封顶 `12000k`，不主动抬高低于目标的源码率。
  - 保持长视频“有码率走 bitrate mode、无码率回退 `CRF 23`”和 HEVC 输出 profile 不变；`short`、`av` 的分辨率目标码率维持现状。
  - ffmpeg bitrate mode 参数统一改为 `-maxrate` 为目标码率 `2x`、`-bufsize` 为目标码率 `4x`，并同步调整服务层与 ffmpeg 层测试期望。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestDecideVideoBitrate|TestBuildTranscodePlan' -count=1`
  - `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)' -count=1`
  - `go test ./pkg/ffmpeg ./internal/services -count=1`

### [2026-05-13 23:19 +0800] 电影电视剧 HEVC 压缩恢复码率约束完成
- Type: `implementation`
- Summary:
  - 电影 `movie` 和电视剧分集 `episode` 仍然保持 HEVC 硬件转码输出，但当探测到源视频码率时，不再走无上限 `CRF` 路径。
  - 长视频转码计划改为复用现有分辨率码率策略：4K 最高 `8000k`，1080p 最高 `4000k`，更低分辨率保持源视频码率，避免转码后文件体积反向膨胀。
  - 保留源码率不可用时回退到 `CRF 23` 的兜底逻辑，并补齐电影/剧集 HEVC 码率模式回归测试。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestBuildTranscodePlan(WithoutSourceBitrateFallsBackToCRF|LongformUsesBitrateStrategyWhenSourceBitrateKnown|NonLongformKeepsBitrateStrategy)' -count=1` passed.
  - `go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscodePlan|TestChooseTranscodeOutputProfile|TestResolveProbeFields|TestBuildPlaybackMetadata|TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)|TestDecideVideoBitrate' -count=1` passed.

### [2026-05-13 23:19 +0800] 电影电视剧 HEVC 压缩恢复码率约束计划
- Type: `plan`
- Summary:
  - 保持电影和电视剧的 HEVC 输出不变，但移除“长视频有码率信息时仍强制走无上限 `CRF`”的策略。
  - 当源视频码率可用时，长视频改为复用现有码率控制逻辑，按分辨率对目标码率做上限约束，避免输出体积明显大于输入。
  - 对没有可用源码率的输入保留 `CRF 23` 兜底，测试覆盖电影、剧集和非长视频的分支差异。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestBuildTranscodePlan(WithoutSourceBitrateFallsBackToCRF|LongformUsesBitrateStrategyWhenSourceBitrateKnown|NonLongformKeepsBitrateStrategy)' -count=1`
  - `go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscodePlan|TestChooseTranscodeOutputProfile|TestResolveProbeFields|TestBuildPlaybackMetadata|TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)|TestDecideVideoBitrate' -count=1`

### [2026-05-13 18:00 +0800] 管理端视频详情手动状态修改完成
- Type: `implementation`
- Summary:
  - 视频详情弹窗新增“状态”下拉，管理员可在 `uploaded`、`scraping`、`tv_pending`、`ready`、`failed` 之间手动切换状态。
  - 当前状态为 `processing` 时，状态字段保持可见但禁用，并显示“处理中状态不支持手动修改”提示；标题、描述、封面等其他字段仍可继续编辑保存。
  - 状态候选、`processing` 禁改判定与保存时是否传 `status` 的分支统一收口到 `videoList.helpers.js`，并补上对应回归测试。
- Changed Files:
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.

### [2026-05-13 18:00 +0800] 管理端视频详情手动状态修改计划
- Type: `plan`
- Summary:
  - 在视频详情弹窗增加状态编辑控件，候选状态限定为 `uploaded`、`scraping`、`tv_pending`、`ready`、`failed`，不提供 `processing` 作为手动目标状态。
  - 当前状态为 `processing` 时仅禁用状态字段并提示“处理中状态不支持手动修改”，不影响标题、描述、封面等其他字段编辑保存。
  - 保持现有 `updateAdminVideo` 接口不变，通过前端 helper 统一状态选项、禁改判定和保存时的 `status` 传参规则。
- Changed Files:
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test`
  - `cd admin-web && npm run build`

### [2026-05-13 17:12 +0800] 管理端视频列表封面预览验证完成
- Type: `implementation`
- Summary:
  - 管理端视频列表已补上标题后的封面列，支持点击缩略图放大预览。
  - 非 `ready` 状态使用占位卡片，`ready` 状态优先尝试加载可访问的封面接口。
  - 前端封面地址统一走 `/api/v1/videos/:id/thumbnail`。
- Changed Files:
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.

### [2026-05-13 17:05 +0800] 管理端视频列表增加封面预览
- Type: `implementation`
- Summary:
  - 视频管理页的表格在标题后新增固定宽度封面列，`ready` 视频直接展示缩略图，点击缩略图可放大预览。
  - 非 `ready` 视频与加载失败场景显示固定尺寸占位卡片，避免表格行高抖动和破图。
  - 前端统一通过 `/api/v1/videos/:id/thumbnail` 生成封面访问地址，不再把列表返回的本地路径直接当作图片源。
- Changed Files:
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `plan.md`
- Verification:
  - pending

### [2026-05-13 15:05 +0800] TV 字幕选择居中对话框改造
- Type: `implementation`
- Summary:
  - 长视频播放器的字幕选择按 `tvMode` 分流：TV 模式渲染居中紧凑对话框，非 TV 模式继续使用原底部弹层。
  - 新增 TV 字幕对话框，包含半透明遮罩、标题“字幕”、“关闭字幕”和字幕轨道列表，当前项显示高亮与“已选中”状态。
  - 抽出字幕选项构造和弹层类型解析，保持 `selectedSubtitleTrackId`、`subtitleTracks`、`onSelectSubtitleTrack` 与 `subtitleTrackDisplayLabel` 数据流不变。
  - 新增单测覆盖 TV/非 TV 分流，以及空选择和当前字幕轨道的选中状态。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `plan.md`
- Verification:
  - Red: `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.SubtitleSelectionTest` failed before implementation because `SubtitlePickerSurface`、`resolveSubtitlePickerSurface`、`buildSubtitlePickerItems` did not exist.
  - Green: `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.SubtitleSelectionTest` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.core.ui.SubtitleSelectionTest --tests com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest --tests com.chee.videos.core.ui.LongFormVideoPlayerStyleTest` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` failed on existing `com.chee.videos.feature.home.HomeViewModelTest.loadCategory loads default av list`.
  - `adb devices` showed no connected devices, so `connectedDebugAndroidTest` was not run.
- Rollback:
  - `git revert <commit>`

### [2026-05-13 06:38 +0800] TV 海报墙与播放遥控体验优化
- Type: `implementation`
- Summary:
  - TV 海报墙卡片改为单一竖版 poster 卡片，只从 `posterUrl` 渲染主视觉；无海报时显示占位封面，标题/描述放入固定信息区。
  - 长视频播放器在 TV 隐藏控制栏状态下由根容器接收遥控：左右 10/30 秒快退快进，确认/播放键播放暂停，上下键唤出控制栏并聚焦播放按钮；隐藏时焦点回到根容器。
  - TV 全局右上设置菜单只在 `tv-home` 路由显示，离开首页自动关闭已展开菜单。
  - 新增纯函数回归测试覆盖海报墙卡片内容、隐藏控制栏按键动作和设置入口路由可见性。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallCardContentTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerTransportKeyTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/tv/TvShellAppRouteVisibilityTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvPosterWallCardContentTest --tests com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest --tests com.chee.videos.tv.TvShellAppRouteVisibilityTest --tests com.chee.videos.core.ui.LongFormVideoPlayerStyleTest --tests com.chee.videos.feature.tv.TvRoutesTest` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` failed on existing `com.chee.videos.feature.home.HomeViewModelTest.loadCategory loads default av list`; the same single test also fails in the original worktree before this branch's changes.
  - `adb devices` showed no connected devices, so `connectedDebugAndroidTest` was not run.
- Rollback:
  - `git revert <commit>`

### [2026-05-12 10:44 +0800] 电影电视剧硬件 HEVC 压缩策略
- Type: `implementation`
- Summary:
  - 电影 `movie` 和电视剧 `episode` 改用硬件 HEVC 长视频转码策略，输出 `video-hevc.mp4`，播放元数据记录 `playback_codec=hevc`。
  - 长视频固定使用 CRF 23，并启用 VideoToolbox 支持的 `spatial_aq` 近似暗部优化；不使用 CPU `libx265` 专属的 `preset` 或 `x265-params`。
  - 其他类型继续使用硬件 AVC 兼容策略，输出 `video-avc.mp4`，并保留现有码率上限/CRF 回退规则。
  - 所有转码显式设置 `-allow_sw 0`，避免硬件编码不可用时静默回退软件编码。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `plan.md`
- Verification:
  - `go test ./pkg/ffmpeg -run 'TestBuildTranscodeVideoArgsFor(HevcPrimary|AvcCompat)' -count=1` passed.
  - `go test ./internal/services -run 'TestBuildTranscodePlan|TestChooseTranscodeOutputProfile|TestBuildPlaybackMetadata|TestResolveProbeFields' -count=1` passed.
  - `go test ./pkg/ffmpeg ./internal/services ./internal/queue -count=1` passed.
  - `go test ./... -count=1` passed.
  - `go vet ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 19:37 +0800] AV 视频详情到手动刮削快捷入口
- Type: `implementation`
- Summary:
  - 视频管理的 AV 详情弹窗新增“去 AV 手动刮削”快捷按钮，仅在 AV 视频上显示。
  - 进入 `/av-scrape` 时自动携带 `video_id`、`external_id` 和用于搜索的 `title`，默认用番号作为搜索词。
  - 手动刮削页补齐 route query 预填，`video_id` 和 `external_id` 会自动写入表单，减少复制粘贴。
- Changed Files:
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/AVManualScrape.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `admin-web/src/views/avManualScrape.helpers.js`
  - `admin-web/src/views/avManualScrape.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test -- src/views/videoList.helpers.spec.js src/views/avManualScrape.helpers.spec.js` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 17:52 +0800] 统一 AV 站点海报策略完成
- Type: `implementation`
- Summary:
  - `poster_url` 统一表示主海报源图，确认落盘元数据同步记录 `poster_url`、`poster` 原图路径和 `thumb` 最终展示路径。
  - 裁剪开启时优先从主 poster 生成 cropped thumb；只有裁剪关闭或裁剪失败时才使用站点 portrait `thumb_url`。
  - ThePornDB 改用 `background.large/image` 作为主 poster，`posters.large/poster` 作为站点原始 thumb 候选；Mywife 保留 `topview.jpg` 为主 poster，并记录 `thumb.jpg` 为原始 thumb 候选。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVPrefersCroppedPosterOverSeparateThumbWhenCropEnabled|TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites|TestPreviewAVFallsBackToThePornDBWhenConfigured' -count=1` passed.
  - `go test ./internal/services -run 'Poster|ThePornDB|Mywife|DMM|MGStage|JavDB|MDCx' -count=1` passed.
  - `go test ./internal/services ./internal/queue ./internal/handlers -count=1` passed.
  - `go test ./... -count=1` passed.
  - `go vet ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 17:48 +0800] 统一 AV 站点海报策略实现进度
- Type: `implementation`
- Summary:
  - 先新增/调整回归测试，确认旧实现下 ThePornDB、Mywife 和通用 thumb 优先逻辑会失败。
  - 落盘层调整为裁剪开启时优先从主 poster 裁剪 thumb，裁剪不可用时再使用站点 portrait thumb。
  - ThePornDB 主海报改用 `background.large/image`，Mywife 主海报保留 `topview.jpg`，确认元数据补充 `poster_url`。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVPrefersCroppedPosterOverSeparateThumbWhenCropEnabled|TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites|TestPreviewAVFallsBackToThePornDBWhenConfigured' -count=1` passed.
  - 全量验证待执行。
- Rollback:
  - `git revert <commit>`

### [2026-05-10 11:22 +0800] 短视频转 AV 按标题选站修正
- Type: `implementation`
- Summary:
  - 短视频转 AV 自动刮削不再把 `filePath` 传入预览流程，改为只根据标题自动选站，避免文件名里的 FC2 / 编号干扰站点分类。
  - 默认 AV 站点顺序收口为：日系首选 `javdb`，欧美首选 `theporndb`，FC2 首选 FC2 站点；已保存的后台配置不迁移。
  - 新增回归测试锁定“标题驱动选站”和默认日系顺序，防止后续再被文件路径或默认顺序回退。
- Changed Files:
  - `internal/queue/scrape_tasks.go`
  - `internal/queue/scrape_tasks_test.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_strategy_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestDefaultAVScraperSiteConfigIncludesMDCxMigratedSites|TestResolveAVSearchPlanUsesDefaultFC2Sites|TestResolveAVSearchPlanUsesStoredWesternConfigOrder|TestResolveAVSearchPlanAllowsExplicitSiteSourceOverride' -count=1` passed.
  - `go test ./internal/queue -run 'TestAutoScrapeAVMarksReadyOnSuccess|TestAutoScrapeAVUsesTitleToSelectSite' -count=1` passed.
  - `go test ./internal/services ./internal/queue -count=1` passed.
  - `go vet ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 10:41 +0800] 开发脚本接入本地翻译模型
- Type: `implementation`
- Summary:
  - `dev-up.sh` 增加本地翻译模型启动与就绪检查，默认拉起 `llama-server` 监听 `127.0.0.1:8000`，对外提供 OpenAI-compatible `/v1` 接口。
  - 新增 `TRANSLATION_SERVER_BIN` 与 `TRANSLATION_MODEL_PATH` 运行时覆盖能力，默认使用本机已验证的 HY-MT1.5-1.8B GGUF Q4 模型文件。
  - `.env.example` 默认开启 `TRANSLATION_API_URL=http://127.0.0.1:8000/v1`，新建环境可直接走本地翻译服务。
- Changed Files:
  - `scripts/dev-up.sh`
  - `.env.example`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
  - 停掉既有翻译进程后执行 `bash scripts/dev-up.sh --frontend off` passed，脚本成功启动 translation，并识别既有 server/worker 进程。
  - `GET http://127.0.0.1:8000/v1/models` returned HY-MT1.5-1.8B.
  - `POST http://127.0.0.1:8000/v1/chat/completions` returned Chinese JSON translation.
  - `GET http://127.0.0.1:8080/healthz` returned `{"status":"ok"}`.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 09:02 +0800] DMM 刮削取消站点超时
- Type: `implementation`
- Summary:
  - DMM 刮削不再使用全局 AV HTTP client timeout；搜索页、详情页和 FANZA TV JSON 请求都会一直等待上游响应，除非外层 context 被主动取消。
  - 保持 DMM 迁移版顺序搜索策略不变，只调整 DMM HTTP 请求边界，避免 30 秒超时把 DMM 刮削中断。
  - 新增回归测试覆盖 DMM 上游响应慢于全局 AV timeout 时仍能继续等待并返回结果。
- Changed Files:
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_mdcx_sites_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestDMMSearchCandidatesWaitsPastConfiguredHTTPTimeout|TestDMMSearchCandidates' -count=1` passed.
  - `go test ./internal/services ./internal/handlers ./internal/queue -count=1` passed.
  - `go vet ./internal/services ./internal/handlers ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-10 08:38 +0800] AV 刮削迁移策略收口
- Type: `implementation`
- Summary:
  - 将 AV 站点执行顺序收口到迁移版配置逻辑，保留 `EnabledSites` 顺序作为主执行序列，避免再从类别回退到被禁用的站点。
  - DMM 刮削改为按迁移版顺序顺次请求搜索页，并只使用首个命中的结果，不再并发打满三条搜索 URL。
  - AV 上传刮削把原始 `filePath` 继续传入统一预览链路，ThePornDB 继续按迁移策略优先吃 `detailUrl > filePath > number`。
  - 管理端 AV 预览接口补充 `file_path/detail_url` 入参，统一走同一条迁移式预览链路。
- Changed Files:
  - `internal/handlers/admin_scrape.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_execution_test.go`
  - `internal/services/scraper_av_mdcx_sites_test.go`
  - `internal/services/scraper_av_strategy.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVSearchDoesNotFallBackToDisabledCategorySource|TestScrapeAVUploadPassesFilePathToThePornDBCrawler|TestDMMSearchCandidatesUsesFirstSearchURLOnly' -count=1` passed.
  - `go test ./internal/services ./internal/handlers ./internal/queue -count=1` passed.
  - `go vet ./internal/services ./internal/handlers ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-09 23:41 +0800] MDCx AV 迁移站点补齐实现
- Type: `implementation`
- Summary:
  - 补齐 MDCx 迁移站点默认配置和后台站点选项，新增 `dmm/jav321/mgstage` 并按 MDCx 优先级调整日系与 FC2 默认顺序。
  - 为 `dmm/mgstage/jav321/fc2ppvdb/fc2club/fc2hub` 补上编号搜索入口，避免站点已注册但自动预览空返回。
  - 修正 AV 图片本地保存选择：`poster` 保留横版原图，`thumb` 只采用竖版站点图；站点 thumb 非竖版时保留下载文件并从横版裁切竖版。
  - 保持确认流程旧 FC2 detail URL 兼容，同时在自动搜索路径按 MDCx 参考 URL 迁移。
- Changed Files:
  - `admin-web/src/views/avManualScrape.helpers.js`
  - `admin-web/src/views/avManualScrape.helpers.spec.js`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_mdcx_sites_test.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_strategy_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services` passed.
  - `go test ./internal/services ./internal/handlers ./internal/queue` passed.
  - `go vet ./internal/services ./internal/handlers ./internal/queue` passed.
  - `cd admin-web && npm test -- avManualScrape.helpers.spec.js` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-09 23:10 +0800] scraper-preview-go ThePornDB 欧美站点迁移计划
- Type: `plan`
- Summary:
  - 首批欧美站点只迁移 `theporndb`，保持 `/api/preview` 的 `number/sites` 兼容，并新增可选 `filePath/detailUrl`。
  - 新增服务端环境变量 `THEPORNDB_API_TOKEN`，token 缺失时只让 `theporndb` 返回站点级失败，不影响其他站点聚合。
  - 新增 ThePornDB crawler、Vue 表单入口与 README 中文说明；首版不读取本地媒体文件、不实现真实 hash 匹配。
- Changed Files:
  - `references/mdcx/scraper-preview-go/internal/config/config.go`
  - `references/mdcx/scraper-preview-go/internal/crawler/types.go`
  - `references/mdcx/scraper-preview-go/internal/preview/service.go`
  - `references/mdcx/scraper-preview-go/internal/sites/theporndb/theporndb.go`
  - `references/mdcx/scraper-preview-go/internal/sites/theporndb/theporndb_test.go`
  - `references/mdcx/scraper-preview-go/cmd/server/main.go`
  - `references/mdcx/scraper-preview-go/web/src/App.vue`
  - `references/mdcx/scraper-preview-go/README.md`
  - `plan.md`
- Verification:
  - `cd references/mdcx/scraper-preview-go && go test ./...` passed.
  - `cd references/mdcx/scraper-preview-go && go vet ./...` passed.
  - `cd references/mdcx/scraper-preview-go/web && npm run build` passed.
  - `THEPORNDB_API_TOKEN=*** PORT=18081 STATIC_DIR= REQUEST_TIMEOUT_SECONDS=20 go run ./cmd/server` live service started; `POST /api/preview` with `sites:["theporndb"]` and `filePath:"x-art.19.11.03.A.Kitten.For.Christmas.mp4"` returned HTTP 200 with title, actor, release, poster/thumb and API detail URL.
  - Same live service with public `detailUrl:"https://theporndb.net/scenes/xart-falling-in-love-with-alex-grey"` returned HTTP 200 and normalized `debug.detailUrl` to `https://api.theporndb.net/scenes/xart-falling-in-love-with-alex-grey`.
  - No-token service on `PORT=18082` returned HTTP 502 with `siteResults[0].site="theporndb"`, `ok=false`, and `error="theporndb: THEPORNDB_API_TOKEN is required"`.
- Rollback:
  - `git revert <commit>`

### [2026-05-09 22:15 +0800] AV 刮削共享配置与本地图片语义收口
- Type: `implementation`
- Summary:
  - AV 预览链路补齐 `filePath/detailUrl` 输入优先级，手动刮削、AV 上传自动刮削、短视频转 AV 重刮削共用同一份站点配置和站点选路。
  - 对齐 MDCx Go 字段语义：候选和最终 metadata 同时保留 `poster` 与 `thumb`，`poster` 优先为横版图，`thumb` 优先为站点提供的竖版图；没有独立竖图时从横版图裁切生成。
  - 刮削图片全部落盘，新增 `poster_thumb_file_path/poster_thumb_path` 兼容实际竖图，同时保留既有 `poster_original_*` 与 `poster_cropped_*` 字段；ThePornDB、JavDB、FC2 的图片字段按 MDCx 结构补齐。
  - 短视频转 AV 重刮削继续只产出候选和元数据，不改写视频状态；ThePornDB token 缺失保持站点级失败，不拖垮其他站点。
- Changed Files:
  - `internal/handlers/video_source.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/queue/scrape_tasks_test.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_javdb_mdcx_regression_test.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services ./internal/queue -run 'TestConfirmAVStores|TestConfirmAVFalls|TestResolveAVPosterAssets|TestPreviewAVJavDB|TestPreviewAVFallsBackToThePornDB|TestConfirmAVNormalizesThePornDB|TestAutoScrapeAVPreservesCurrentStatus|TestPreviewAVJavDBPrefersCoverAndDetailOverview' -count=1` passed.
  - `go test ./internal/services ./internal/handlers ./internal/queue -count=1` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-09 18:52 +0800] scraper-preview-go FC2 live 验证与 FC2Hub 样张修复
- Type: `implementation`
- Summary:
  - 启动 `scraper-preview-go` Go 预览服务并对 `fc2/fc2hub/fc2ppvdb/fc2club` 做真实 HTTP 预览验证。
  - 修复 `fc2hub` live 页面中样张选择器过宽的问题，避免把 LAXD、tag、seller 和推荐视频链接误并入 `extrafanart`。
  - 记录 live 验证现状：`fc2` 官方站可返回有效元数据，`fc2hub` 可返回有效元数据且样张已收敛为图片链接；`fc2ppvdb` 当前网络下被登录/Cloudflare Turnstile 页面拦截，`fc2club` 当前 live 页面未匹配既有标题结构。
- Changed Files:
  - `references/mdcx/scraper-preview-go/internal/sites/fc2hub/fc2hub.go`
  - `references/mdcx/scraper-preview-go/internal/sites/fc2hub/fc2hub_test.go`
  - `plan.md`
- Verification:
  - `cd references/mdcx/scraper-preview-go && go test ./...` passed.
  - `cd references/mdcx/scraper-preview-go && go vet ./...` passed.
  - `PORT=18080 STATIC_DIR= REQUEST_TIMEOUT_SECONDS=20 go run ./cmd/server` started the live preview service.
  - `POST /api/preview {"number":"FC2-1723984","sites":["fc2"]}` returned `ok: true` with title, poster/thumb, tags and 10 extrafanart items.
  - `POST /api/preview {"number":"FC2-1940476","sites":["fc2hub"]}` returned `ok: true` with title, thumb and 10 image-only extrafanart items after the selector fix.
- Rollback:
  - `git revert <commit>`

### [2026-05-09 18:46 +0800] scraper-preview-go FC2 专站迁移实现
- Type: `implementation`
- Summary:
  - 在 `references/mdcx/scraper-preview-go` 迁移 MDCx 的 FC2 官方站与 FC2Hub 两个专站，新增 `fc2` 与 `fc2hub` Go crawler，并保持 `/api/preview` 请求与响应结构不变。
  - 新增离线测试覆盖 FC2 编号归一化、URL 构造、搜索候选排序、详情字段解析、无码标签处理和失败路径。
  - 将两个新站点接入服务注册、Vue 站点选择和 README 返回格式说明；现有 `fc2ppvdb/fc2club` 逻辑未重写。
- Changed Files:
  - `references/mdcx/scraper-preview-go/internal/sites/fc2/fc2.go`
  - `references/mdcx/scraper-preview-go/internal/sites/fc2/fc2_test.go`
  - `references/mdcx/scraper-preview-go/internal/sites/fc2hub/fc2hub.go`
  - `references/mdcx/scraper-preview-go/internal/sites/fc2hub/fc2hub_test.go`
  - `references/mdcx/scraper-preview-go/cmd/server/main.go`
  - `references/mdcx/scraper-preview-go/web/src/App.vue`
  - `references/mdcx/scraper-preview-go/README.md`
  - `plan.md`
- Verification:
  - `cd references/mdcx/scraper-preview-go && go test ./...` passed.
  - `cd references/mdcx/scraper-preview-go && go vet ./...` passed.
  - `cd references/mdcx/scraper-preview-go/web && npm run build` passed after `npm ci` installed missing local frontend dependencies.
- Rollback:
  - `git revert <commit>`

## Entry Template
### [YYYY-MM-DD HH:MM] Title
- Type: `plan` | `implementation` | `docs`
- Summary:
- Changed Files:
  - `path/to/file`
- Verification:
  - `command/result`
- Rollback:
  - `git revert <commit>`

### [2026-05-08 19:12] TV 播放器焦点请求崩溃修复实现
- Type: `implementation`
- Summary:
  - 修复 `LongFormVideoPlayer` 在 TV 模式下的焦点请求时序：首次进入和控件隐藏后重新唤起时，不再在状态切换同一拍直接调用 `FocusRequester.requestFocus()`，改为先标记待请求并在下一帧、且控件可见时再申请焦点。
  - 新增 TV 播放器 Compose 仪器测试，覆盖“首次渲染时播放按钮应获取焦点且不崩溃”的回归场景。
  - 保持修复范围最小，仅调整播放器焦点调度，不改动播放交互与 UI 结构。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-tv-app/tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerFocusTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:assembleDebug :tv-app:compileDebugAndroidTestKotlin` passed.
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.LongFormVideoPlayerStyleTest'` passed.
  - `cd android-tv-app && ./gradlew :tv-app:connectedDebugAndroidTest -Pandroid.testInstrumentationRunnerArguments.class=com.chee.videos.core.ui.LongFormVideoPlayerFocusTest` blocked by device install error `INSTALL_FAILED_UPDATE_INCOMPATIBLE` because `com.chee.videos.tv` had a conflicting existing signature record on the connected BRAVIA device.
  - `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` failed on unrelated existing test `com.chee.videos.feature.home.HomeViewModelTest > loadCategory loads default av list`.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 19:08] TV 播放器焦点请求崩溃修复计划
- Type: `plan`
- Summary:
  - 排查 TV 端点击播放后 `LongFormVideoPlayer` 抛出的 `FocusRequester is not initialized`，重点检查 TV 模式首次进入和播放控件显隐切换时的焦点请求时序。
  - 通过最小代码改动把焦点申请延后到目标节点真正挂载后，再补一条播放器级回归测试锁住该崩溃。
  - 验证以 TV 工程构建、播放器相关测试和真机仪器测试为主，若设备环境阻塞则明确记录阻塞原因。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-tv-app/tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerFocusTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew :tv-app:connectedDebugAndroidTest -Pandroid.testInstrumentationRunnerArguments.class=com.chee.videos.core.ui.LongFormVideoPlayerFocusTest` first surfaced test setup issues, then was re-run after test compile fixes.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 17:54] AV 手动刮削站点选择失效修复实现
- Type: `implementation`
- Summary:
  - 收紧 AV 预览执行层的 crawler 解析逻辑：显式 `site_source` 仅执行目标站点，按配置收敛后的自动站点集合也不再偷偷追加其他 crawler，未知显式站点会直接返回错误。
  - 将 AV 预览缓存键改为包含 `site_category` 与归一化后的站点列表，并在缓存命中时回填 `recommended_source/used_source/enabled_sources`，避免跨站点复用旧结果。
  - 扩充日系默认站点链路，把 `theporndb/getchu` 纳入明确默认 fallback 清单；新增服务层回归测试覆盖显式站点、配置限站点、缓存隔离和未知站点报错。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_execution_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVSearchUsesExplicitSiteSourceOnly|TestPreviewAVSearchHonorsConfiguredEnabledSourcesOnly|TestPreviewAVSearchCacheKeyIncludesSiteSource|TestPreviewAVSearchRejectsUnknownExplicitSiteSource'` passed.
  - `go test -count=1 ./internal/services` passed.
  - `go test -count=1 ./internal/...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 17:51] AV 手动刮削站点选择失效修复计划
- Type: `plan`
- Summary:
  - 修复 AV 手动刮削中 `site_source` 与配置站点顺序未被执行层严格遵守的问题，避免显式指定其他站点时仍继续搜索并合并 `javdb`。
  - 调整 AV 预览缓存键，使不同 `site_source/site_category` 不再复用同一标题缓存结果。
  - 补充服务层回归测试，覆盖显式站点、配置限站点、缓存隔离与未知站点报错场景。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_execution_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVSearchUsesExplicitSiteSourceOnly|TestPreviewAVSearchHonorsConfiguredEnabledSourcesOnly|TestPreviewAVSearchCacheKeyIncludesSiteSource|TestPreviewAVSearchRejectsUnknownExplicitSiteSource'` 先红后绿。
- Rollback:
  - `git revert <commit>`

### [2026-05-08 09:56] Android TV 重设计最终验证收口
- Type: `implementation`
- Summary:
  - 补齐电视剧详情页的大海报展示所需 `baseUrl` 状态，确保 TV 详情首屏能正确解析并显示相对海报/背景地址。
  - 将 TV 焦点 modifier 简化为更稳定的实现，规避 Kotlin/KAPT 在当前本地环境下对该文件的编译器异常。
  - 最终验证过程中发现 TV 工程 Kotlin 增量缓存会放大 KAPT 内存占用，因此先清理 `.gradle/kotlin` 与 `tv-app/build` 后重新串行跑测试和构建，结果恢复稳定通过。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && rm -rf .gradle/kotlin tv-app/build` cleared Kotlin incremental caches and build outputs after KAPT `Java heap space` failures.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*' --tests 'com.chee.videos.tv.TvQrCodeEncoderTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerStyleTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 09:52] Android TV 影院化首页、独立详情页与全屏播放器重设计
- Type: `implementation`
- Summary:
  - 将独立 TV App 首页重做为“搜索栏 + 顶部推荐横幅 + 长视频货架”结构，新增推荐横幅内容选择逻辑，继续观看优先进入主视觉位；统一卡片和主按钮的焦点缩放、高亮边框和抬升反馈。
  - 新增 TV 专用电影/AV 详情页与独立全屏播放器路由，TV 端不再使用原来的详情内嵌播放器模式；电视剧播放器也改为默认全屏进入，选集通过 TV 式面板打开。
  - 改造长视频播放器控件以适配遥控器：支持可聚焦控制按钮、TV 模式下的快进快退、字幕选择、退出播放、进度条常显于操作期间，并让字幕列表支持遥控器确认键。
  - 电视剧详情页补齐大海报与横幅背景布局，并让详情主按钮默认落焦，统一电影/AV/电视剧三类长视频在 TV 上的视觉语言。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/TvFocus.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvLongFormDetailPresentationTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFeaturedContentTest' --tests 'com.chee.videos.feature.tv.TvLongFormDetailPresentationTest' --tests 'com.chee.videos.feature.tv.TvCatalogFocusPolicyTest' --tests 'com.chee.videos.feature.tv.TvRoutesTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*' --tests 'com.chee.videos.tv.TvQrCodeEncoderTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest' --tests 'com.chee.videos.feature.tv.TvSeriesDetailViewModelTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerStyleTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 21:50] Android TV 焦点链路补齐并修复遥控器不可操作
- Type: `implementation`
- Summary:
  - 为独立 TV App 首页补齐默认焦点策略，按“继续观看 -> 首页分区首卡 -> 电视剧 -> 电影 -> AV -> 搜索框”的顺序选择首个可聚焦目标，解决安卓 10 TV 设备首次进入页面后遥控器无法落焦的问题。
  - 为 TV 首页卡片、搜索结果、配对页主按钮/切换服务器按钮以及右上角账户菜单入口补齐 `focusable`/`FocusRequester`，确保 D-pad 确认键可以真实触发点击，而不是仅在触屏语义下可点。
  - 新增 TV 焦点策略单测，锁定首页初始焦点回退规则，避免后续再次引入“页面可见但按钮不可操作”的回归。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFocusPolicyTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogFocusPolicyTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest' --tests 'com.chee.videos.tv.TvQrCodeEncoderTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 07:59] Android 短视频播放器控件统一与非首页进度条间距调整
- Type: `implementation`
- Summary:
  - 将 Android 短视频播放器的播放模式切换按钮收敛到公共 UI helper，统一首页、发现页、搜索页和统一播放器的图标、激活态与无障碍文案；搜索页不再单独使用 `Cached` 图标分支。
  - 为非首页短视频播放器补齐公共进度条显示约定与底部间距约定：`ShortSearchScreen` 新增可拖拽进度条，`ShortDiscoverScreen` 与 `UnifiedPlayerScreen` 改为使用“底部安全区 + 额外 12.dp” 的统一抬高规则，首页 `ShortFeedScreen` 保持无进度条。
  - 新增 Android 单测，锁定播放模式文案、非首页进度条显示条件和额外底部间距常量，避免后续播放器入口再次分叉。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/ShortVideoPresentation.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/ShortVideoPlaybackChromeTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --stop` stopped Gradle daemons before rerun.
  - `cd android-app && rm -rf app/build .gradle/kotlin` cleared collided Kotlin/KAPT build caches after parallel verification invalidated them.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.ui.ShortVideoBottomProgressBarScrubMathTest' --tests 'com.chee.videos.core.ui.ShortVideoPlaybackChromeTest' --tests 'com.chee.videos.feature.shorts.ShortFeedProgressVisibilityTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-03 20:12] flick 迁移标题改为标签组合并回填历史
- Type: `implementation`
- Summary:
  - 停止当前迁移后，调整 `FlickImportService` 标题生成逻辑：新迁移视频标题改为标签组合形式 `#tag1 #tag2`，不再使用文件名或 md5。
  - 在 `cmd/import-flick` 增加 `--backfill-title-only` 模式，用于只回填 `migration_source=flick-server` 的历史数据标题，按 `video_tags` 聚合生成 `#标签` 标题并批量更新。
  - 已执行一次历史回填，确保既有 flick 迁移数据与后续迁移数据都采用统一标题规则。
- Changed Files:
  - `internal/services/flick_import.go`
  - `internal/services/flick_import_test.go`
  - `cmd/import-flick/main.go`
  - `plan.md`
- Verification:
  - `go test ./cmd/import-flick -count=1` passed.
  - `go test ./internal/services -run 'TestFlickImportServiceImportPlayableVideo' -count=1` passed.
  - `go run ./cmd/import-flick --postgres-dsn 'postgres://video:video@127.0.0.1:5432/video_server?sslmode=disable' --backfill-title-only` logged `updated=46389`.
  - `select count(*) from videos where metadata->>'migration_source'='flick-server' and title like '#%';` returned `46389`.
- Rollback:
  - `git revert <commit>`

### [2026-05-03 12:04] flick 迁移恢复运行并切回外部盘批处理
- Type: `implementation`
- Summary:
  - 恢复本机 MongoDB 源库后，重新启动 `cmd/import-flick`，迁移参数固定为 `--batch-size 1000` 和 `--storage-root /Volumes/large/video-server-storage`，避免继续写入内部盘。
  - 停掉了之前残留的 `--dry-run` 迁移进程和未指定外部盘根目录的旧批量迁移进程，避免它们继续占用资源或把数据迁回内部盘。
  - 当前批处理迁移已跑完前 2 批，每批 1000 条，累计处理 2000 条，当前都被判定为跳过，没有失败，迁移仍在继续推进。
- Changed Files:
  - `plan.md`
- Verification:
  - `brew services info mongodb-community` showed Mongo was installed but not running, then `mongod --config /opt/homebrew/etc/mongod.conf --fork` started the server successfully.
  - `lsof -nP -iTCP:27017 -sTCP:LISTEN` confirmed Mongo was listening on `127.0.0.1:27017`.
  - `go test ./cmd/import-flick -count=1` passed.
  - `go test ./internal/services -run 'TestFlickImportServiceImportPlayableVideo' -count=1` passed.
  - 迁移进程日志显示 `batch=1` 和 `batch=2` 均已完成，`batch_size=1000`，`batch_failed=0`。
- Rollback:
  - `git revert <commit>`

### [2026-05-02 22:38] flick-server 已转码可播放视频数据库迁移实现
- Type: `implementation`
- Summary:
  - 新增离线命令 `cmd/import-flick`，通过本机 `mongoexport` 读取 `flick-server` 的 `videos` 数据，只筛选 `canplay=true` 的已转码可播放视频，并支持 `tag`、`since`、`limit`、`dry-run` 等过滤参数。
  - 新增 `FlickImportService`，统一处理源视频解析、标签标准化、SHA256 去重、封面/视频复制、源时间保留与导入结果分类；导入目标统一落为当前项目的 `short + ready` 视频。
  - 仓储层新增 `CreateImportedReadyVideo` 原子写入入口，一次完成 `videos`、`video_tags`、`file_hashes` 落库；补齐命令与服务层单测，并生成 JSON 导入报告。
- Changed Files:
  - `cmd/import-flick/main.go`
  - `cmd/import-flick/main_test.go`
  - `internal/services/flick_import.go`
  - `internal/services/flick_import_test.go`
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `go test ./cmd/import-flick -count=1` passed.
  - `go test ./internal/services -run 'TestFlickImportServiceImportPlayableVideo' -count=1` passed.
  - `go test ./... -count=1` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-02 23:00] flick 迁移命令修正 mongoexport 参数兼容性
- Type: `implementation`
- Summary:
  - 修正 `cmd/import-flick` 对本机 `mongoexport 100.9.5` 的参数适配，移除不受支持的 `--batchSize`，避免 dry-run 启动即因命令行选项解析失败退出。
  - 实际排查确认本地 `flick` 库 `canplay=true` 的记录量为 `152647` 条，这也是本次 dry-run 扫描耗时较长的直接原因。
- Changed Files:
  - `cmd/import-flick/main.go`
  - `plan.md`
- Verification:
  - `mongoexport --help` confirmed no `--batchSize` option.
  - `go test ./cmd/import-flick -count=1` passed.
  - `mongosh 'mongodb://localhost:27017/flick' --quiet --eval 'db.videos.countDocuments({canplay:true})'` returned `152647`.
- Rollback:
  - `git revert <commit>`

### [2026-05-02 22:38] flick-server 已转码可播放视频数据库迁移计划
- Type: `plan`
- Summary:
  - 实现一个一次性离线导入命令，源数据来自 `flick-server` Mongo `videos` 集合，只处理 `canplay=true` 且已存在转码视频与封面的记录。
  - 导入时统一把视频写入当前项目 `short` 类型，并保留源 `createdAt/updatedAt`、源 `tags[]`、源 `md5` 与来源路径等迁移元数据。
  - 标签以视频级 `tags[]` 为准，不迁移源标签统计；媒体文件复制到当前项目 `storageRoot/videos/<uuid>/`，不直接依赖旧盘绝对路径。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./cmd/import-flick -count=1`
  - `go test ./internal/services -run 'TestFlickImportServiceImportPlayableVideo' -count=1`
  - `go test ./... -count=1`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 22:43] Android AV 详情页显示演员头像
- Type: `implementation`
- Summary:
  - Android AV 详情页新增独立演员区块，优先展示接口返回的 `detail.actors`，每个演员显示头像与姓名；无头像时显示姓名首字占位。
  - 演员头像地址接入现有 `baseUrl` 解析链路，支持相对路径 `/api/v1/actors/:id/avatar` 和绝对 URL，两种形式都能在详情页正常出图。
  - 保留 metadata `actors` 作为无 `detail.actors` 时的名字回退来源，但不再伪造头像，避免把演员展示和旧 metadata 混成单行文本。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/AvDetailPresentation.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/AvDetailPresentationTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest' --tests 'com.chee.videos.feature.detail.AvDetailLayoutSpecTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 22:42] Android AV 详情页显示演员头像计划
- Type: `plan`
- Summary:
  - 在已补齐后端演员头像抓取与本地访问的前提下，Android AV 详情页从“只显示演员名字”扩展为“显示头像 + 姓名”。
  - 详情页优先消费 `detail.actors[].avatar_url`，并沿用现有 `baseUrl` 拼接逻辑解析相对地址；metadata 仅用于演员名字回退。
  - 变更范围限定在 Android AV 详情页 presentation/UI 与相关单测，不扩展演员详情页、列表页或其他端。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest'`
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug`
- Rollback:
  - `git revert <commit>`

### [2026-05-08 14:11] Android TV 播放路由防崩与继续观看兜底
- Type: `implementation`
- Summary:
  - 修复独立 TV App 进入播放时的高风险路由问题：`seriesId/videoId` 统一按 URI segment 编码构建，避免 ID 中带 `/`、`?`、`#` 等字符时导航匹配失败或参数丢失。
  - TV 首页“继续观看”长视频新增统一播放目标解析，优先使用 `videoId`，缺失时回退到 `seriesId`，避免生成空播放路由导致点击进入播放器直接闪退。
  - 详情/播放器相关 ViewModel 读取路由参数时统一解码；`DetailViewModel` 对空 `videoId` 改为进入错误态而非主线程直接崩溃。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvCatalogFeaturedContentTest.kt`
  - `plan.md`
- Verification:
  - `cd android-tv-app && ./gradlew --stop` confirmed no active Gradle daemons before rerun.
  - `rm -rf android-tv-app/.gradle/kotlin android-tv-app/tv-app/build` cleared corrupted Kotlin incremental caches after local KAPT/IC cache exceptions.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvRoutesTest' --tests 'com.chee.videos.feature.tv.TvCatalogFeaturedContentTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*' --tests 'com.chee.videos.feature.detail.*'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug` failed on pre-existing `com.chee.videos.feature.home.HomeViewModelTest` assertions unrelated to本次 TV 播放修复。
- Rollback:
  - `git revert <commit>`

### [2026-05-08 10:26] 管理端统一分页跳转与视频预览销毁计划
- Type: `plan`
- Summary:
  - 为管理端所有列表页和图片合集抽屉分页新增统一的手动页码输入与跳转能力，避免各页面重复实现。
  - 修复视频管理详情弹窗关闭后预览视频继续播放的问题，要求关闭弹窗、切换详情和刷新播放链接时都显式销毁预览播放器资源。
  - 变更范围限定在 `admin-web` 的共享分页组件、视频预览 helper、相关页面接入与单测，不调整后端接口。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd admin-web && npm run test`
  - `cd admin-web && npm run build`
- Rollback:
  - `git revert <commit>`

### [2026-05-08 10:26] 管理端统一分页跳转并修复视频预览关闭未销毁
- Type: `implementation`
- Summary:
  - 新增 `AdminTablePagination` 共享组件和页码跳转 helper，为管理端所有表格分页统一提供“输入页码 + 回车/按钮跳转”能力，并对非法页码、越界页码和空数据场景做统一钳制。
  - 将 `视频管理、图片管理、图片合集管理、任务监控、电视剧管理、演员管理、用户管理、合集管理` 的分页全部切换到共享组件，包含图片合集抽屉内的二级分页。
  - 为 `VideoList` 新增预览播放器销毁逻辑和详情弹窗关闭回调，确保关闭详情、刷新播放链接、离开页面时都暂停播放、清空 `src` 并释放媒体资源。
- Changed Files:
  - `admin-web/src/components/AdminTablePagination.vue`
  - `admin-web/src/components/adminTablePagination.helpers.js`
  - `admin-web/src/components/adminTablePagination.helpers.spec.js`
  - `admin-web/src/views/ActorManage.vue`
  - `admin-web/src/views/CollectionManage.vue`
  - `admin-web/src/views/ImageCollectionManage.vue`
  - `admin-web/src/views/ImageManage.vue`
  - `admin-web/src/views/TaskMonitor.vue`
  - `admin-web/src/views/TvSeriesManage.vue`
  - `admin-web/src/views/UserManage.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm run test` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 12:52] Android 手机端 AV 海报墙瀑布流与分页刷新计划
- Type: `plan`
- Summary:
  - 将手机端 `AV` 标签页海报墙从普通网格改为瀑布流布局，支持下拉刷新和上滑自动加载更多。
  - 默认浏览态与搜索结果态统一支持分页、刷新和尾部补货失败反馈，不新增后端接口。
  - 变更范围限定在 `android-app` 的 Home AV 列表 UI、ViewModel 分页状态机、仓储分页返回值与单测。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.home.HomeViewModelTest'`
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug`
- Rollback:
  - `git revert <commit>`

### [2026-05-08 12:52] Android 手机端 AV 海报墙改为瀑布流并支持下拉刷新与上滑加载
- Type: `implementation`
- Summary:
  - `HomeScreen` 的 AV 海报墙改为双列 `LazyVerticalStaggeredGrid`，保留顶部搜索 Hero，并接入下拉刷新指示器、尾部加载中/失败/无更多状态。
  - `HomeViewModel` 新增 AV 浏览态与搜索态的统一分页状态机，支持默认列表分页、搜索分页、下拉刷新、清空搜索恢复浏览态和尾部补货错误保留已加载内容。
  - `VideoRepository.fetchCategory` 改为保留搜索接口返回的分页元数据，`HomeViewModelTest` 补齐浏览分页、搜索分页、下拉刷新和分页失败回归用例。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` passed after the parallel-run KAPT cache conflict was discarded.
  - `cd android-app && ./gradlew --stop && rm -rf app/build .gradle/kotlin && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.home.HomeViewModelTest'` passed after removing temporary debug output.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 22:10] AV 刮削联动演员头像补全并落地本地访问
- Type: `implementation`
- Summary:
  - AV 演员绑定后新增头像补全过程：仅对启用且 `avatar_url` 为空的演员执行补全，按 `JavDB -> TMDB` 顺序查找候选头像。
  - 命中头像后下载到 `storageRoot/actors/<actor_id>/avatar.*`，并把演员 `avatar_url` 更新为 `/api/v1/actors/:id/avatar`，同时回写来源与外部 ID。
  - 新增演员头像读取路由，优先返回本地头像文件；原有 AV 手动刮削状态保留、自动 AV 刮削入库与演员管理预览接口保持兼容。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_actor_avatar.go`
  - `internal/services/scraper_test.go`
  - `internal/repository/actor_repository.go`
  - `internal/handlers/router.go`
  - `internal/handlers/actor_avatar.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestSyncAVActorsCompletesMissingAvatarToLocalRoute|TestSyncAVActorsFallsBackToTMDBAvatarWhenJavDBHasNoImage|TestSyncAVActorsDoesNotOverrideExistingAvatar|TestScrapeAVUploadCodeFirstAndActorSync|TestConfirmAVPreservesExplicitReadyStatus|TestPreviewActorByNameTMDB|TestPreviewActorByNameJavDB' -count=1` passed.
  - `go test ./internal/handlers -run 'TestAdminActorScrapePreviewInvalidSource|TestAdminActorScrapePreviewTMDBSuccess' -count=1` passed.
  - `go test ./... -count=1` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 22:10] AV 刮削联动演员头像补全计划
- Type: `plan`
- Summary:
  - 为当前仓库补齐“AV 刮削后联动演员头像补全”，而不是修改 MDCX 单次刮削结果结构；头像来源默认固定为 `JavDB` 优先、`TMDB` 回退。
  - 头像补全默认只处理 `avatar_url` 为空的演员，避免覆盖人工维护；下载成功后优先改为本地访问地址，不新增新的演员头像字段。
  - 后端需要新增本地头像存储与读取路由，但管理端和 App 的 `avatar_url` 字段协议保持不变。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestSyncAVActorsCompletesMissingAvatarToLocalRoute|TestSyncAVActorsFallsBackToTMDBAvatarWhenJavDBHasNoImage|TestSyncAVActorsDoesNotOverrideExistingAvatar' -count=1`
  - `go test ./internal/handlers -run 'TestAdminActorScrapePreviewInvalidSource|TestAdminActorScrapePreviewTMDBSuccess' -count=1`
  - `go test ./... -count=1`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 21:31] AV 详情页播放器改为 16:9 并腾出顶部安全区
- Type: `implementation`
- Summary:
  - Android AV 详情页媒体区从固定 `440.dp` 高度改为按宽度使用 `16:9` 比例，避免播放器在详情页中占用过高竖向空间。
  - AV 详情页内容容器新增顶部状态栏安全区留白，由页面统一承担顶部 inset；海报态的返回按钮移除重复的 `statusBarsPadding()`，避免双重下推。
  - 新增 detail 布局 spec 单测，锁定“16:9 比例 + 顶部安全区留白”这两个布局约束。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/AvDetailLayoutSpecTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest' --tests 'com.chee.videos.feature.detail.AvDetailLayoutSpecTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 21:31] AV 详情页播放器改为 16:9 并腾出顶部安全区计划
- Type: `plan`
- Summary:
  - 将 Android AV 详情页媒体区从固定高度改成 `16:9` 比例，并让顶部状态栏安全区独立留白，避免播放器贴到屏幕最顶端。
  - 变更范围限定在 `DetailScreen` 的 AV 详情布局与对应单测，不改播放器内部手势/控制条逻辑，也不改非 AV 详情页。
  - 验证目标为 detail 相关单测与 `:app:assembleDebug`。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.detail.AvDetailLayoutSpecTest'`
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 21:08] Android AV 列表/详情页海报分流显示
- Type: `implementation`
- Summary:
  - Android AV 海报 helper 改为按场景分流：列表项继续优先使用 `poster_cropped_path`，详情页优先使用 `poster_original_path`，两端都保留现有 scraped poster 与 `thumbnailPath` 的回退链路。
  - 扩展 Android 单测覆盖：锁定“列表裁剪图优先、详情原图优先、缺图时互相回退”的行为，避免后续再次把两个场景收敛成同一张海报。
  - 全量 `:app:testDebugUnitTest` 仍被仓库既有失败 `HomeViewModelTest.loadCategory loads default av list` 阻断；已在未改动的 `master` 上复现，确认不是本次变更引入。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/AvPosterSupport.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/model/AvPosterSupportVariantTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.model.AvPosterSupportTest' --tests 'com.chee.videos.core.model.AvPosterSupportVariantTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.home.HomeViewModelTest.loadCategory loads default av list'` failed on branch and `master` with the same assertion, treated as pre-existing.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 21:07] Android AV 列表/详情页海报分流显示计划
- Type: `plan`
- Summary:
  - 将 Android AV 海报解析从“列表页与详情页共用同一优先级”改为“列表优先裁剪图、详情优先原图”，后端接口与 metadata 字段保持不变。
  - 实施范围限定在 Android `AvPosterSupport` helper 与其单测，不改服务端返回值、不改管理端海报落地逻辑。
  - 验证目标为相关 Android 单测与 `:app:assembleDebug`；若全量单测存在无关红测，则需先独立确认是否为既有问题。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.model.AvPosterSupportTest' --tests 'com.chee.videos.core.model.AvPosterSupportVariantTest'`
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 20:26] AV 裁剪模式补齐与手动刮削状态保留修正
- Type: `implementation`
- Summary:
  - AV 海报裁剪模式扩展为 `portrait_center`、`portrait_left`、`portrait_right` 三种枚举；后端在横向裁切时按模式决定左右锚点，空值或非法值统一回退到 `portrait_center`。
  - AV 手动刮削确认保存新增显式状态透传：管理端 handler 会把当前视频状态传给 `ConfirmAV`，从而让已 `ready` 的视频在手动保存后继续保持 `ready`；自动链路与 `ScrapeAVUpload` 仍维持原有 `uploaded` 行为。
  - 管理端 AV 配置页把 `poster_crop_mode` 从自由输入收敛为枚举下拉，前端 helper 同步只接受这 3 个值并在远端配置异常时回退到 `portrait_center`。

### [2026-05-07 12:41] 新增 TV 配对授权链路与独立 Android TV 模块
- Type: `implementation`
- Summary:
  - 后端新增 `tv-auth` 配对授权协议与存储：增加 `tv_devices` / `tv_auth_sessions` migration、配对会话创建/轮询/批准/拒绝 handler、repo 与 service；`/api/v1/tv/home` 扩展为固定长视频分区字段，新增 `/api/v1/tv/search` 聚合搜索电影、剧集与 AV。
  - Android 手机 App 新增 TV 授权协议模型、深链解析与确认授权页：`MainActivity` 支持 `cheevideos://tv-auth` 深链，登录后可进入确认页并调用批准/拒绝接口；补齐相关 repository、URL builder、API service 与解析单测。
  - 新增独立 `:tv-app` Android TV 模块，复用现有共享源码，提供独立 launcher / Leanback 入口、服务器选择、配对登录页与 TV 长视频导航壳，兼容当前 Android 10 测试机（API 29）运行基线。
- Changed Files:
  - `migrations/0018_tv_auth_support.up.sql`
  - `migrations/0018_tv_auth_support.down.sql`
  - `internal/models/app.go`
  - `internal/models/user.go`
  - `internal/repository/tv_auth_repository.go`
  - `internal/services/tv.go`
  - `internal/services/tv_auth.go`
  - `internal/services/tv_auth_test.go`
  - `internal/services/tv_service_test.go`
  - `internal/handlers/router.go`
  - `internal/handlers/tv.go`
  - `internal/handlers/tv_auth.go`
  - `android-app/app/src/main/AndroidManifest.xml`
  - `android-app/app/src/main/java/com/chee/videos/MainActivity.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/TvAuthRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthDeepLink.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthApprovalScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthApprovalViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tvauth/TvAuthDeepLinkParserTest.kt`
  - `android-app/settings.gradle.kts`
  - `android-app/tv-app/build.gradle.kts`
  - `android-app/tv-app/src/main/AndroidManifest.xml`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvMainActivity.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `plan.md`
- Verification:
  - `go test ./internal/services ./internal/handlers -count=1` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tvauth.TvAuthDeepLinkParserTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 13:34] TV 端消费四分区聚合首页与长视频聚合搜索
- Type: `implementation`
- Summary:
  - 将 Android `feature/tv` 从“仅电视剧专区”扩展为“TV 长视频入口”：`TvCatalogViewModel` 改为消费后端 `tv/home` 的固定分区字段 `tv_series`、`movies`、`av`，搜索改走独立 `tv/search` 聚合接口，不再复用旧的 `search_results` 剧集结果。
  - `TvCatalogScreen` 新增电影/AV 分区卡片流、混合类型搜索结果卡片与按类型分流的继续观看入口；电视剧继续进入剧集详情，电影/AV 直接进入现有长视频详情页。
  - `tv-app` 导航壳补齐电影/AV 详情路由，使独立 TV 应用现在可以从首页与搜索进入电视剧、电影、AV 三类长视频内容。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMockData.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --stop` stopped stale daemons before rerun.
  - `cd android-app && rm -rf app/build tv-app/build .gradle/kotlin` cleared corrupted Kotlin/KAPT incremental caches before final sequential verification.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 18:41] 手机扫码授权按深链服务器自动切换登录上下文
- Type: `implementation`
- Summary:
  - 手机 App 收到 `cheevideos://tv-auth` 深链且其中带有 `server` 参数时，`VideoHomeApp` 会在进入登录/授权流程前先触发服务器同步，不再要求用户手动切回对应服务。
  - `AppRootViewModel` 新增 TV 授权场景下的服务器应用入口：仅当深链服务器与当前活动服务器不同才切换，并在切换时清空旧 token，避免把 A 服务器扫码授权误发到 B 服务器登录态。
  - `ServerRepository.activateEndpoint` 补充可选 `clearTokens` 语义，供跨服务器扫码授权场景安全复用，现有普通连接页逻辑保持不变。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/ServerRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/viewmodel/AppRootViewModel.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 18:44] 手机 TV 授权确认页改为预加载服务端会话摘要
- Type: `implementation`
- Summary:
  - `TvAuthApprovalViewModel.bind()` 现在会在接收深链后主动请求 `tv-auth session` 状态，优先展示服务端返回的设备名、配对码、服务器地址和当前状态，而不是只依赖二维码里的静态参数。
  - 授权确认页新增会话状态展示，并把按钮可用性收敛到 `pending` 状态；对 `approved`、`expired`、`denied` 三种状态给出明确中文提示，避免用户在已失效会话上继续点击授权。
  - 手机端授权成功后会把本地状态同步成 `approved`，拒绝后同步成 `denied`，保证当前页与 TV 端轮询结果一致。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthApprovalScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthApprovalViewModel.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/services/scraper_test.go`
  - `internal/handlers/admin_scrape.go`
  - `admin-web/src/views/AVManualScrape.vue`
  - `admin-web/src/views/avManualScrape.helpers.js`
  - `admin-web/src/views/avManualScrape.helpers.spec.js`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestResolveAVPosterAssetsUsesConfiguredHorizontalCropAnchor|TestConfirmAVPreservesExplicitReadyStatus|TestConfirmAVStoresOriginalAndCroppedPosterAssets|TestConfirmAVFallsBackToOriginalPosterWhenCropFails|TestScrapeAVUploadCodeFirstAndActorSync' -count=1` passed.
  - `go test ./... -count=1` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 18:45] AV 手动刮削与站点配置重构落地
- Type: `implementation`
- Summary:
  - 后端新增 AV 站点配置与分类路由：自动/手动 AV 统一走“配置优先、标题分类兜底”的选站策略，支持 `fc2`、`western`、`japanese` 三类，手动 AV 预览新增专用接口并默认 `bypass_cache=true`，不再复用旧 metadata/from_cache 结果。
  - AV 海报落地改为“原图 + 裁剪图 + 展示变体”双产物：保存时写入 `poster_original_path`、`poster_cropped_path`、`poster_variant` 等 metadata，服务端 `/api/v1/videos/:id/thumbnail` 支持按变体回源本地文件，手动 AV 保存明确不再触发转码。
  - 管理端新增独立 `AV 手动刮削` 菜单页与配置面板；Android AV 海报解析优先裁剪图、再原图、最后回退旧缩略图；补齐 Go、Vitest 与 Android 单测回归。
- Changed Files:
  - `internal/models/admin.go`
  - `internal/repository/admin_settings_repository.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_strategy.go`
  - `internal/services/scraper_av_poster.go`
  - `internal/services/scraper_av_strategy_test.go`
  - `internal/services/scraper_av_poster_assets_test.go`
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/admin_scrape_test.go`
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/utils/video_url.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/router/index.js`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/AVManualScrape.vue`
  - `admin-web/src/views/avManualScrape.helpers.js`
  - `admin-web/src/views/avManualScrape.helpers.spec.js`
  - `android-app/app/src/main/java/com/chee/videos/core/model/AvPosterSupport.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/model/AvPosterSupportVariantTest.kt`
  - `migrations/0017_admin_settings.up.sql`
  - `migrations/0017_admin_settings.down.sql`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
  - `cd android-app && ./gradlew app:testDebugUnitTest --tests 'com.chee.videos.core.model.AvPosterSupportTest' --tests 'com.chee.videos.core.model.AvPosterSupportVariantTest'` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 11:43] Android AV 首页与详情页重设计落地
- Type: `implementation`
- Summary:
  - 首页 `AV` tab 改为独立海报墙布局，不再复用通用横向列表；顶部加入 AV 专属搜索头部、远程搜索状态文案和大海报竖版卡片。
  - `HomeViewModel` 新增 AV 搜索双态：默认浏览继续走 `fetchCategory("av")`，输入关键词后走新的 `VideoRepository.searchAv(...)` 远程搜索，清空查询可恢复默认浏览且不污染原列表。
  - AV 详情页重排为海报主导版式：首屏 Hero 强调海报、番号和播放入口，中段补作品信息/演员/简介/标签/操作胶囊；同时新增“番号优先、标题回退”的展示 helper 与对应单测。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeAvPresentation.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/AvDetailPresentation.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/home/HomeAvPresentationTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/AvDetailPresentationTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 03:08] 将 JavDB 切换到 MDCX 风格解析计划
- Type: `plan`
- Summary:
  - 用户确认 JavDB 不应继续使用仓库内自写正则解析，而应切换到 MDCX 迁移过来的搜索与详情解析思路。
  - 本轮修复范围限定在 `javdb` 站点：搜索页补封禁/Cloudflare 显式错误与 MDCX 风格匹配策略，详情页补 DOM 版标题、海报、LDJSON 简介与演员提取。
  - 现有 provider、候选合并、落库与下载策略保持不变，只替换 JavDB 内核解析，避免影响其他已迁移站点。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVJavDBUsesMDCXStyleDOMDetailParsing|TestPreviewAVJavDBReturnsExplicitErrorOnCloudflareBlock' -count=1` 先红后用于回归。
  - `go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestPreviewAVFallbackByTitle|TestPreviewAVJavDBPrefersCoverAndDetailOverview|TestPreviewAVMergeUsesBestFieldsAcrossSources|TestConfirmAVKeepsExistingThumbnailWhenOnlyFallbackPoster|TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail' -count=1` 用于兼容回归。
- Rollback:
  - `git revert <commit>`

### [2026-05-01 03:08] 将 JavDB AV 刮削切到 MDCX 风格搜索与 DOM 解析
- Type: `implementation`
- Summary:
  - JavDB 搜索页现在会识别 MDCX 参考实现中的封禁/版权限制/Cloudflare 拦截文本，并返回显式错误，不再静默退化成空结果或被其他站点 404 覆盖。
  - JavDB 详情页现在优先走 MDCX 风格 DOM 提取：`current-title`、`copy-to-clipboard`、`img.video-cover`、`application/ld+json` 简介、以及 `female/male` 演员结构；正则逻辑仅保留为兼容 fallback。
  - 搜索命中策略改为“精确匹配优先、清洗后模糊匹配次之、最后才回退到当前宽松候选收集”，兼容现有标题兜底行为，同时更接近 MDCX 原始 crawler。
- Changed Files:
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_javdb_mdcx_regression_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVJavDBUsesMDCXStyleDOMDetailParsing|TestPreviewAVJavDBReturnsExplicitErrorOnCloudflareBlock' -count=1` passed.
  - `go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestPreviewAVFallbackByTitle|TestPreviewAVJavDBPrefersCoverAndDetailOverview|TestPreviewAVMergeUsesBestFieldsAcrossSources|TestConfirmAVKeepsExistingThumbnailWhenOnlyFallbackPoster|TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail' -count=1` passed.
  - `go test ./... -count=1` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 02:56] 延长 AV 刮削运行超时计划
- Type: `plan`
- Summary:
  - 只调整运行配置，不改 Go 代码默认值，继续沿用现有 `AV_SCRAPER_TIMEOUT_SECONDS` 配置入口。
  - 当前代码默认超时仍保留 `10` 秒，但项目实际运行环境通过根目录 `.env` 显式覆盖到 `120` 秒。
  - 同步更新 `.env.example`，避免后续新环境仍按旧示例值 `10` 秒启动。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/config . -count=1` 作为配置链路回归验证。
  - `rg -n 'AV_SCRAPER_TIMEOUT_SECONDS' .env .env.example` 确认实际配置和示例值。
- Rollback:
  - `git revert <commit>`

### [2026-05-01 02:56] 将 AV 刮削运行超时提高到 120 秒
- Type: `implementation`
- Summary:
  - 在项目根目录 `.env` 中新增 `AV_SCRAPER_TIMEOUT_SECONDS=120`，让当前 server/worker 启动链路统一使用更长的 AV 刮削超时。
  - 在 `.env.example` 中把 AV 刮削超时示例值从 `10` 秒更新为 `120` 秒，和当前运行建议保持一致。
  - 本次不改 `internal/config/config.go` 中的默认值，仍保留代码 fallback 为 `10` 秒，仅通过运行配置实现延长。
- Changed Files:
  - `.env`
  - `.env.example`
  - `plan.md`
- Verification:
  - `go test ./internal/config . -count=1` passed.
  - `rg -n 'AV_SCRAPER_TIMEOUT_SECONDS' .env .env.example` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 01:41] MDCX AV 迁移收尾修复计划
- Type: `plan`
- Summary:
  - 针对最后 review 明确的 3 个缺口收尾：`love6` `external_id` 大小写往返不一致、MDCX 迁入站点的 URL 覆写配置仍停留在 4 个特例、以及 `resolveRelativeAVURL` 会把 query 错当 path 编码。
  - 方案固定为保留 `love6` 原始 path segment、把 `AV_SITE_URL_` 扩展成通用站点映射并保持 `AV_SCRAPER_BASE_URL` 仅为 `javdb` fallback、同时让相对 URL helper 直接复用现有浏览器语义解析。
  - 执行方式采用分治并行：分别补服务回归测试、配置/装配测试和 URL 解析回归，再在主线做交叉复查和全量 Go 回归。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/config . -run 'TestLoadBuildsAVSiteURLsFromPrefixedEnvVars|TestLoadIncludesAVSiteOverridesAndTokens|TestLoadFallsBackToBaseURLForJavDBSiteURL|TestConfigureAVScraperPassesThroughSharedSiteURLs' -count=1` passed.
  - `go test ./... -count=1` 作为最终回归目标。
- Rollback:
  - `git revert <commit>`

### [2026-05-01 01:41] 修复 MDCX AV 迁移收尾的 3 个逻辑缺口
- Type: `implementation`
- Summary:
  - `love6` crawler 现在保留详情页 URL 中原始 `external_id`，不再强制 lower-case；新增独立回归测试锁定 `ConfirmAV` 落库后的顶层与站点块 metadata 都保持原值。
  - MDCX 相对 URL 解析改为复用统一 `toAbsoluteURL(...)` 语义，修复 `avsex` 搜索结果详情链接和图片链接的 query/协议相对 URL 解析错误；新增 helper 与搜索回归测试。
  - 配置层新增通用 `AVSiteURLs`，扫描全部 `AV_SITE_URL_` 环境变量并通过共享装配逻辑传入 scraper；主线补齐 `MDTV`、`AIRAV_CC` 这类 key 的 canonical 归一化，避免未来直接读取 `cfg.AVSiteURLs` 时与服务层来源名规则不一致。
- Changed Files:
  - `internal/config/config.go`
  - `internal/config/config_test.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_av_mdcx_sites.go`
  - `internal/services/scraper_love6_regression_test.go`
  - `internal/services/scraper_av_relative_url_test.go`
  - `main.go`
  - `main_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/config . -run 'TestLoadBuildsAVSiteURLsFromPrefixedEnvVars|TestLoadIncludesAVSiteOverridesAndTokens|TestLoadFallsBackToBaseURLForJavDBSiteURL|TestConfigureAVScraperPassesThroughSharedSiteURLs' -count=1` passed.
  - `go test ./... -count=1` passed.
- Rollback:
  - `git revert <commit>`
 
### [2026-04-26 21:05] 电视继续播放与字幕记忆修复计划
- Type: `plan`
- Summary:
  - 用户反馈电视首页缺少海报/封面、继续播放入口不会直接恢复到上次剧集，且电视剧页没有稳定恢复上次播放时间与上次选择的字幕。
  - 本轮修复范围限定在 Android app：电视目录页补齐 `baseUrl` 与继续播放 artwork/直达播放器链路；电视剧播放器补齐 `watchSeconds` 恢复和本地字幕偏好记忆；后端接口与字幕 DTO 契约保持不变。
  - 执行策略遵循先补失败测试再改实现，验证目标包含定向电视单测，以及 `:app:testDebugUnitTest :app:assembleDebug` 全量回归。
- Changed Files:
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.core.data.AppPreferencesStoreTest'` 先红后用于回归。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 21:06] 修复电视继续播放封面、续播时间与字幕记忆
- Type: `implementation`
- Summary:
  - 电视首页现在会保留并使用服务端 `baseUrl`，继续播放卡片与电视剧列表/搜索卡片优先显示真实海报或横版 backdrop；继续播放点击路径改为直接进入 `buildTvPlayerRoute(seriesId, season, episode)`，不再绕回详情页。
  - 电视剧播放数据链路补齐 `watchSeconds` 与本地字幕偏好：`TvEpisodeUiModel` / `TvContinueWatchingUiModel` 增加续播与 artwork 字段，`AppPreferencesStore` 新增按 `videoId` 读写字幕选择，`TvSeriesPlayerViewModel` 会在切换分集时恢复上次字幕选择。
  - 电视剧播放器现在会在首次装载某个 `videoId` 时按历史 `watchSeconds` seek，并在退出页面时通过最新 `currentVideoId` 上报历史，避免继续播放时间丢失；字幕切换仍保持现有“只换字幕不丢进度”的重配路径。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/data/AppPreferencesStoreTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvRepositoryMappingTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest' --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest' --tests 'com.chee.videos.core.data.AppPreferencesStoreTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

---

### [2026-04-26 20:07] 去掉长视频字幕黑底并改为轻量描边
- Type: `implementation`
- Summary:
  - 长视频字幕黑底来自 app 端 `PlayerView` 的默认 `SubtitleView` 样式，而不是后端字幕内容；本次在共享长视频播放器里统一覆写字幕样式，改成“透明背景 + 细描边”，减少字幕底色对画面的遮挡。
  - `LongFormVideoPlayer` 现在会在创建和更新 `PlayerView` 时统一应用 app 字幕样式：白字、背景透明、窗口透明、轻量描边，并保持内嵌样式与字号支持开启，不影响已有字幕内容解析。
  - 新增样式回归测试，锁定共享长视频字幕样式的关键颜色与描边配置，避免后续又退回 Media3 默认黑底。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerStyleTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.ui.LongFormVideoPlayerStyleTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-26 20:06] 长视频字幕黑底样式修复计划
- Type: `plan`
- Summary:
  - 用户确认字幕已经正常显示，但黑色背景会遮挡画面，希望尽量去掉，优先在 app 端处理。
  - 本地检查确认：电视剧页、详情页和统一长视频页共用 `LongFormVideoPlayer`，当前 `PlayerView` 未对内部 `SubtitleView` 做任何样式覆写，因此直接沿用了 Media3 的默认字幕黑底。
  - 本轮按共享长视频播放器统一修复，目标样式锁定为“透明背景 + 细描边”；先补失败测试锁定样式，再在 `LongFormVideoPlayer` 统一应用。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本轮字幕样式修复相关文件。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 19:33] 修复长视频内嵌字幕手动选择后不自动启用
- Type: `implementation`
- Summary:
  - 这次“电视剧页选了内嵌字幕仍不显示”的直接根因不在 `baseUrl`，而在共享长视频字幕挂载逻辑：播放器重建 `MediaItem` 时只会挂载“当前选中的单条字幕轨”，但只有后端原始 `is_default=true` 的轨道才会被打上 `SELECTION_FLAG_DEFAULT`，导致用户手动选择的非默认内嵌字幕虽然被挂进 `MediaItem`，Media3 仍可能保持文本轨关闭。
  - `buildLongFormMediaItem` 现在改为把“当前唯一挂载的、可用的字幕轨”统一标记为默认选中轨，确保用户在电视剧页、详情页或统一长视频页显式切换到某条内嵌字幕后，播放器会实际启用这条字幕。
  - 新增回归测试锁定“手动选中的非默认内嵌字幕必须带 `SELECTION_FLAG_DEFAULT`”，避免后续再次回到“选中了但未启用”的状态。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.ui.SubtitleSelectionTest.resolveSubtitleSelectionFlags_marksExplicitlySelectedEmbeddedTrackAsDefault'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-26 19:32] 内嵌字幕不显示二次排查计划
- Type: `plan`
- Summary:
  - 用户反馈上一次修复后“仍然没有看到字幕”，且补充说明当前复现的是内嵌字幕，不是外挂字幕。
  - 重新核对后确认：电视剧页、电影详情页和统一长视频页共用同一套 `buildLongFormMediaItem(...)` 字幕挂载逻辑；本轮优先按共享长视频字幕链路排查，而不是继续只做电视剧页局部补丁。
  - 初步假设是“用户手动选择的非默认内嵌字幕未被标记为播放器默认选中轨”，验证目标是先补失败测试，再最小化修正 Media3 `SubtitleConfiguration` 的 selection flags。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本轮内嵌字幕修复相关文件。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 16:15] 修复电视剧播放器选择字幕后不显示字幕
- Type: `implementation`
- Summary:
  - 修复电视剧播放页字幕重配缺少 `baseUrl` 的问题：`TvSeriesPlayerViewModel` 现在会读取并透传当前服务端地址，`TvSeriesPlayerScreen` 在重建长视频 `MediaItem` 时改为使用真实 `baseUrl`，让相对路径字幕 URL 能被解析并挂入播放器。
  - 保持现有字幕切换重配策略不变，电视剧页仍沿用 `resolveLongFormPlayerUpdate(... preservePosition=true)` 的“仅换字幕时保留播放进度”行为，本次不改字幕轨优先级、不改后端字幕接口。
  - 补充两类回归测试：一条锁定相对字幕路径在存在 `baseUrl` 时必须解析成绝对地址；一条锁定电视剧播放器状态必须暴露 `baseUrl`，避免后续再次把字幕解析前提丢掉。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.ui.SubtitleSelectionTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 13:42] Android AV 海报数据源修正计划
- Type: `plan`
- Summary:
  - 按既有方案先修数据链路，不在本轮强行并入横竖版适配与安全区重构。
  - 后端 `VideoListItem` 补充 `metadata` 出口，Android AV 列表与详情统一改为读取同一套 AV 海报解析 helper，并保留 `thumbnail_path` 兜底。
  - AV fallback 海报保留策略改为“可用 fallback 海报可覆盖旧上传截帧封面”，同时继续保护无效 URL 场景。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/models ./internal/repository ./internal/services -run 'TestVideoListItemJSONIncludesMetadata|TestNormalizeVideoListItemMetadataDefaultsToJSONObject|TestNormalizeVideoListItemMetadataPreservesScrapedPosterFields|TestConfirmAVReplacesExistingThumbnailWhenFallbackPosterIsUsable|TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail|TestConfirmAVRejectsRelativePosterURLWithoutTMDBFallback' -count=1` 先红后用于回归。
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.model.AvPosterSupportTest' --tests 'com.chee.videos.feature.home.HomeAvPresentationTest' --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest'` 先红后用于回归。
- Rollback:
  - `git revert <commit>`

### [2026-05-01 13:42] 修复 Android AV 列表与详情页海报数据链路
- Type: `implementation`
- Summary:
  - 后端 `VideoListItem` 现在会返回 `metadata`，搜索/分类/用户视频列表查询统一补扫 `videos.metadata`，让 Android AV 列表能拿到 `scrape_source`、站点块 `poster_url`、`poster_decision` 等真实刮削海报信息。
  - Android 新增共享 AV 海报解析 helper，优先读取顶层或 `metadata.<scrape_source>` 下的真实刮削海报，遇到 `invalid_*` 决策时回退 `thumbnail_path`；首页 AV 海报墙与详情页统一复用这套解析逻辑，避免列表和详情取图分叉。
  - 后端 AV fallback 海报策略改为可用时覆盖旧上传截帧封面，`poster_decision` 新增 `fallback_replaced_existing` 回归语义；同时补充模型/仓储/Android helper 测试。为降低验证噪声，顺手把 `AppPreferencesStoreTest` 的 DataStore scope 绑定到 `backgroundScope`。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/AvPosterSupport.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/data/AppPreferencesStoreTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/model/AvPosterSupportTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `internal/models/app.go`
  - `internal/models/app_test.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/app_video_list_item_test.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/models ./internal/repository ./internal/services -run 'TestVideoListItemJSONIncludesMetadata|TestNormalizeVideoListItemMetadataDefaultsToJSONObject|TestNormalizeVideoListItemMetadataPreservesScrapedPosterFields|TestConfirmAVReplacesExistingThumbnailWhenFallbackPosterIsUsable|TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail|TestConfirmAVRejectsRelativePosterURLWithoutTMDBFallback' -count=1` passed.
  - `go test ./... -count=1` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.core.model.AvPosterSupportTest' --tests 'com.chee.videos.feature.home.HomeAvPresentationTest' --tests 'com.chee.videos.feature.detail.AvDetailPresentationTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` failed：现有 `HomeViewModelTest` 仍受 `DataStore`/`Dispatchers.Main` 清理时序影响，失败与本轮 AV 海报 helper 断言无直接对应关系。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 16:14] 电视剧播放器字幕不显示修复计划
- Type: `plan`
- Summary:
  - 已确认这次问题集中在电视剧播放器页面：字幕切换后重建长视频 `MediaItem` 时把 `baseUrl` 传成空字符串，导致相对路径字幕 URL 无法解析，播放器拿不到有效字幕配置。
  - 修复范围限制在电视剧播放链路：`TvRepository` / `TvSeriesPlayerViewModel` 补齐 `baseUrl` 透传，`TvSeriesPlayerScreen` 改为使用真实地址构造字幕资源；不改后端接口、不改字幕选择策略。
  - 验证目标是补一条字幕 URL 解析回归、一条电视剧播放器 `baseUrl` 状态回归，并执行 Android 单测和 `assembleDebug` 全量验证。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本次电视剧字幕修复相关文件。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 08:51] 长视频播放产物收敛为单 AVC
- Type: `implementation`
- Summary:
  - 长视频转码从“HEVC 主档 + AVC 兼容档”收敛为“单一 AVC 播放产物”：新上传与新重转码统一只生成 `video-avc.mp4`，停止写入新的双档 metadata。
  - 播放接口继续兼容 `profile=primary|compat`，但对新视频无论请求哪个 profile 都回落到同一份 `transcoded_path`；对历史双产物视频仍保留旧 `compat_transcoded_path` / `playback_profiles` 解析兼容。
  - Android 端不再按 HEVC 能力做分流，长视频 profile resolver 固定返回 `compat`，避免继续依赖 HEVC/模拟器能力探测。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_source_test.go`
  - `android-app/app/src/main/java/com/chee/videos/core/player/PlaybackProfileResolver.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/player/PlaybackProfileResolverTest.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.core.player.PlaybackProfileResolverTest'` failed: Gradle Wrapper 下载 `https://services.gradle.org/distributions/gradle-8.7-bin.zip` 连接超时，未能完成代码层验证。
- Rollback:
  - `git revert <commit>`

### [2026-04-26 08:49] 长视频播放产物收敛为单 AVC 计划
- Type: `plan`
- Summary:
  - 目标从“双档兼容”调整为“安卓稳定可播优先、存储占用优先”，因此新长视频播放产物统一收敛为单一 AVC/H.264。
  - 历史双产物数据本次不做批量清理，但播放接口需要兼容解析旧 metadata，避免旧视频立即失效。
  - Android 端本次同步取消 HEVC 能力分流，避免仍然保留“主档/兼容档”选择逻辑。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本次收敛改动。
- Rollback:
  - `git revert <commit>`

### [2026-04-25 15:58] 修复长视频重转码被 30 分钟任务超时杀死
- Type: `implementation`
- Summary:
  - 定位本次“`ffmpeg transcode failed: signal: killed`”的真实根因不是音频，而是 `asynq` 未显式设置任务超时，落回框架默认 `30 minutes`；对 58 分钟的 4K 电视剧重转码来说，worker 上下文到点取消后会由 `exec.CommandContext` 直接杀掉 ffmpeg。
  - 转码入队现在显式附带 `Timeout`，新增环境变量 `TRANSCODE_TASK_TIMEOUT_MINUTES`，默认 `360` 分钟，避免长视频重转码被默认超时提前终止。
  - 同步移除 `videotoolbox` 转码参数中的无效 `-preset`，清掉 `Codec AVOption preset ... has not been used` 的噪声日志；补充单测锁定任务 timeout 和 ffmpeg 参数构造。
- Changed Files:
  - `internal/config/config.go`
  - `internal/queue/tasks.go`
  - `internal/queue/tasks_test.go`
  - `main.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - 手工重转码未在本地复放：当前会话仅基于 `.run/worker.log` 与本地 `asynq` 源码确认默认 timeout=30min 的根因。
- Rollback:
  - `git revert <commit>`

### [2026-04-25 15:57] 长视频重转码超时杀进程修复计划
- Type: `plan`
- Summary:
  - 依据 `.run/worker.log`，5.1 音频已成功降混为 stereo，当前失败点转为 ffmpeg 运行中被 `SIGKILL`。
  - 本地 `github.com/hibiken/asynq@v0.25.1` 源码确认：未显式配置 `Timeout` 时默认超时为 `30 minutes`，与日志中 worker 启动到失败约 30 分钟完全一致。
  - 修复方案是为转码任务显式设置长超时并做成可配置，同时顺手移除 `videotoolbox` 下无效的 `-preset` 参数，减少误导性日志。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本次修复相关改动。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 19:47] AV 刮削切换到 mdcx Go 迁移内核计划
- Type: `plan`
- Summary:
  - 目标是在不改现有 `/api/v1/admin/scrape/*`、上传自动刮削和 `videos.metadata` 外部结构的前提下，把 AV 站点解析切到 `references/mdcx` 风格的 crawler/provider 内核。
  - 本轮优先补三类能力：结构化 AV 配置加载、mdcx 聚合补充站点 detail URL 回抓、以及至少一个新增站点的预览搜索回退链路。
  - 实施方式遵循 TDD：先补 `avsex` 预览/确认与配置加载红测，再最小化接入新 crawler，并跑 Go 全量回归。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/config -run 'TestLoad(IncludesAVSiteOverridesAndTokens|FallsBackToBaseURLForJavDBSiteURL)'` 预期先红。
  - `go test ./internal/services -run 'TestConfirmAVBuildsAVSexDetailURLFromSourceAndExternalID|TestPreviewAVFallsBackToAVSexWhenPrimarySitesHaveNoResult'` 预期先红。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 19:47] AV 刮削接入 mdcx 站点配置与聚合补充站点内核
- Type: `implementation`
- Summary:
  - `internal/config.Config` 新增 AV 站点 URL、JavDB/JavBus cookie、ThePornDB token/no-hash 字段；`main.go` 改为通过结构化 `AVScraperConfig` 注入 `ScraperService`，保留 `AV_SCRAPER_BASE_URL` 作为 JavDB 默认回退。
  - `ScraperService` 新增结构化 AV 配置入口与站点 URL 映射，`avSiteBaseURL` 现在优先读站点级配置；同时引入 mdcx 风格聚合补充站点 crawler，先接入 `airav_cc`、`avsex`、`cableav`、`hdouban`、`hscangku`、`iqqtv`、`7mmtv` 的 detail URL 识别与抓取。
  - `avsex` 额外补齐搜索回退链路：当前主站点无结果时，`PreviewAV` 可通过 `avsex` 搜索页拿到候选并复用 detail 抓取；`ConfirmAV` 也能按 `scrape_source=avsex + external_id` 自动构造详情 URL 完成落库与演员同步。
- Changed Files:
  - `internal/config/config.go`
  - `internal/config/config_test.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_sites.go`
  - `internal/services/scraper_test.go`
  - `main.go`
  - `plan.md`
- Verification:
  - `go test ./internal/config -run 'TestLoad(IncludesAVSiteOverridesAndTokens|FallsBackToBaseURLForJavDBSiteURL)'` passed.
  - `go test ./internal/services -run 'TestConfirmAVBuildsAVSexDetailURLFromSourceAndExternalID|TestPreviewAVFallsBackToAVSexWhenPrimarySitesHaveNoResult'` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/config ./internal/services ./internal/queue` passed.
  - `go test -race ./internal/config ./internal/services ./internal/queue ./internal/handlers` failed: `internal/handlers/admin_actor_scrape_test.go` 并行调用 `gin.SetMode()` 触发预存数据竞争，与本次 AV 迁移改动无直接关系。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:16] AV 刮削继续补入 mdcx 第二批站点计划
- Type: `plan`
- Summary:
  - 在首批 `avsex/airav_cc/...` 补充站点落地后，继续把参考实现里结构清晰、fixture 完整的站点接入当前 AV provider。
  - 本轮优先目标是 `dmm`、`mgstage`、`prestige`、`xcity`、`theporndb`，并通过红测锁定两条现有链路：`ConfirmAV` 的 detail URL 自动构造/回抓，以及 `PreviewAV` 在主站点无结果时对 `theporndb` 的搜索回退。
  - 实施策略保持最小替换：不改外部接口，不引入 mdcx 批处理，只在服务层内部新增 crawler 和站点识别规则。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForAdditionalSites|TestPreviewAVFallsBackToThePornDBWhenConfigured'` 预期先红。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:16] AV 刮削接入 DMM/MGStage/Prestige/XCity/ThePornDB crawler
- Type: `implementation`
- Summary:
  - `internal/services` 新增第二批 mdcx 风格 detail/API crawler：`dmm`、`mgstage`、`prestige`、`xcity`、`theporndb`，并补充本地 HTML/JSON 辅助解析函数，维持当前 `avScrapeCandidate -> PreviewAV/ConfirmAV` 的映射结构不变。
  - `scraper_av_framework.go` 扩展 provider 注册、detail URL 构造与站点识别规则；同时收紧 `cableav`、`javbus` 的宽路径匹配，避免在本地 host 或自定义站点 URL 下误判其他站点详情页。
  - `theporndb` 现在可在配置了 token 时参与 `PreviewAV` 搜索回退；`ConfirmAV` 现在可基于 `scrape_source + external_id` 自动回抓 `dmm/mgstage/prestige/xcity/theporndb` 详情并落库。
- Changed Files:
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForAdditionalSites|TestPreviewAVFallsBackToThePornDBWhenConfigured'` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-25 14:53] 修复长视频重转码 5.1 音频失败
- Type: `implementation`
- Summary:
  - 播放用转码产物的音频统一显式降混为 `AAC 2.0`：HEVC 主档与 AVC 兼容档的 ffmpeg 参数都新增 `-ac 2`，避免 5.1 输入在 `aac` 编码器阶段触发 `Unsupported channel layout "6 channels"` 一类失败。
  - 扩展 ffprobe 解析，补充音频流 `codec/channels` 探测；转码 metadata 新增 `audio_codec`、`audio_channels`、`audio_downmixed`，并基于输入音轨声道数判断是否发生降混。
  - 新增回归测试覆盖主档/兼容档音频参数、音频 probe 解析和转码 metadata，锁定后续播放产物持续为稳定的 `AAC 2.0`。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` failed: Gradle Wrapper 下载 `https://services.gradle.org/distributions/gradle-8.7-bin.zip` 超时，未能完成代码层验证。
  - 手工重转码验证未执行：仓库内未提供本次故障电视剧样本或可复用 5.1 测试素材。
- Rollback:
  - `git revert <commit>`

### [2026-04-25 14:52] 长视频重转码 5.1 音频失败修复计划
- Type: `plan`
- Summary:
  - 现有长视频重转码对 `AAC 5.1` 输入未显式做播放产物音频归一化，导致电视剧重转码可能在 `aac` 编码器打开阶段失败。
  - 计划把播放用 HEVC 主档与 AVC 兼容档的音频统一为 `AAC 2.0`，并在 metadata 中补充 `audio_codec`、`audio_channels`、`audio_downmixed`，优先恢复“可稳定重转码、可稳定播放”。
  - 验证目标包含 Go 单测、`go test ./...`、Android 单测与 assemble，以及有样本时的手工重转码回归。
- Changed Files:
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含本次计划相关改动。
- Rollback:
  - `git revert <commit>`

### [2026-04-25 14:21] HEVC 主档与 AVC 兼容档自动选路落地
- Type: `implementation`
- Summary:
  - 后端转码从“单一 HEVC 成品”扩展为“硬件 HEVC 主档 + 硬件 AVC 兼容档”双产物：主档继续走 H.265 小体积策略，兼容档落到 metadata 中的 `compat_transcoded_path` / `playback_profiles`，播放源接口新增 `profile` 选择。
  - app 端新增长视频播放档位解析器：模拟器或无 HEVC 解码能力设备默认走 `compat`，支持 HEVC 的设备继续走 `primary`；电视剧页、详情页和统一长视频播放页统一接入，并在编解码失败时展示明确中文错误提示。
  - ffmpeg 与后端补充了编码档位/播放源选择单测，Android 补充了播放档位解析和带 `profile` 的 source URL 单测，避免后续又回到“默认打 HEVC 导致首播失败”的行为。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_source_test.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `android-app/app/src/main/java/com/chee/videos/core/player/PlaybackProfileResolver.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/player/PlaybackErrorMessage.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/player/PlaybackProfileResolverTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/util/UrlBuilderSourceProfileTest.kt`
  - `docs/superpowers/plans/2026-04-25-hevc-hardware-compatible-playback.md`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.core.player.*' --tests 'com.chee.videos.core.util.UrlBuilderSourceProfileTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-25 12:18] HEVC 硬件转码与兼容播放方案
- Type: `plan`
- Summary:
  - 当前 app 端电视剧立即播放失败的首因不是页面生命周期，而是后端“转码成品”本身输出为 HEVC 10-bit/4K；Android 模拟器对该 profile 不支持，导致首播即 `MediaCodecVideoRenderer` 报错。
  - 按最新要求，后续实现不改成“纯 AVC”，而是保持“硬件加速优先、主产物优先 H.265、体积最小优先”，同时补一条硬件 AVC 兼容产物与播放档位选择链路，解决 Android 兼容性。
  - 计划文件已保存到 `docs/superpowers/plans/2026-04-25-hevc-hardware-compatible-playback.md`，后续按该方案分任务实施。
- Changed Files:
  - `docs/superpowers/plans/2026-04-25-hevc-hardware-compatible-playback.md`
  - `plan.md`
- Verification:
  - `git status --short` 预期仅包含计划文档改动。
- Rollback:
  - `git revert <commit>`

### [2026-04-25 10:29] Android 电视剧切集播放目标竞态修复
- Type: `implementation`
- Summary:
  - 修复电视剧播放页切换分集时的状态撕裂：此前 `selectedEpisodeNumber` 会先切到新分集，但 `currentVideoId/currentSourceUrl` 仍短暂保留旧分集，导致播放器在“新分集元数据 + 旧媒体源”组合下重复 reset，放大 `MediaCodec flush/release` 竞态。
  - 为电视剧播放器 ViewModel 增加播放目标请求版本控制；每次切集先清空旧播放目标，再异步解析新 `sourceUrl`，并丢弃过期请求返回，避免旧分集 URL 晚到后覆盖当前选集。
  - 新增延迟 `sourceUrl` 测试仓库与两条回归测试，覆盖“切集后在新 URL 就绪前必须脱离旧视频目标”和“上一集异步结果晚到时不得反写当前播放目标”。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-25 09:54] Android 长视频播放器重配编解码器竞态修复
- Type: `implementation`
- Summary:
  - 针对字幕切换、详情异步补齐和页面退出并发时的播放器重配路径，新增统一的长视频媒体更新决策，区分“清空播放器”“替换媒体源”“保留播放位置”三种行为，避免继续走高风险的 `stop() + clearMediaItems()`。
  - 详情页、统一播放页和电视剧播放页全部改为基于已准备 URL 与字幕轨状态的显式重配逻辑；其中统一播放页补上 `preparedUrl` 显式状态，不再从 `currentMediaItem` 反推播放器准备状态，降低 release 阶段竞态。
  - 新增单测覆盖“目标 URL 为空时清空”“仅字幕变化时保留播放位置替换源”“目标未变化时不重配”三类行为，锁定本次回归。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.SubtitleSelectionTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 18:32] Android 电视剧首页空列表空指针崩溃修复
- Type: `implementation`
- Summary:
  - 修复电视剧专区打开即崩溃问题：后端返回 `sections=null` 或 `search_results=null` 时，Gson 会把运行时空值灌入 Kotlin 非空列表字段，导致 `TvCatalogViewModel` 在映射时触发空指针。
  - 在电视剧映射层新增空列表兜底，统一对首页分组、搜索结果、详情标签、演员、季列表与分集列表做 `null -> emptyList()` 兼容，避免接口字段缺失或返回 `null` 时崩溃。
  - 新增回归测试，直接用 `Gson` 反序列化 `null` 数组字段，锁定首页与详情映射在异常返回下仍能稳定回退为空集合。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvRepositoryMappingTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*'` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew clean :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 17:42] Android 电视剧列表页标题搜索
- Type: `implementation`
- Summary:
  - 为电视剧列表页增加顶部搜索框，按剧名关键词实时过滤本地占位数据，清空搜索后恢复原有“继续追剧 + 分组推荐”的浏览流。
  - 搜索态切换为独立结果列表，补充结果标题、无结果空状态和可点击结果卡片，保持与现有深色影院风格一致。
  - 为电视剧目录 ViewModel 增加查询状态与过滤逻辑，并新增回归测试锁定“命中过滤”和“清空恢复”两类行为。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.tv.TvCatalogViewModelTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*'` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 16:55] Android 电视剧播放页顶部安全区修复
- Type: `implementation`
- Summary:
  - 修复电视剧详情进入播放页后，顶部返回栏与标题内容压到系统状态栏的问题。
  - 为电视剧播放页顶部标题栏补充 `statusBarsPadding()`，仅让头部区域进入安全区，不改动下方播放器占位区和操作区布局高度。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 16:44] Android 电视剧专区占位 UI 与独立播放流
- Type: `implementation`
- Summary:
  - 将首页原“电视剧”内容替换为独立的电视剧专区占位 UI，不再走现有分类接口加载；首页内提供继续观看、分组推荐和横向卡片浏览，整体采用深色影院风格。
  - 新增电视剧详情页与电视剧播放器页，使用本地 mock 数据驱动，覆盖剧集基础信息、分季选集、剧集播放占位区与底部选集抽屉，保持后续接真实接口时的路由结构稳定。
  - 为电视剧模块补充路由构建、mock 数据和 ViewModel 单测，修复实现阶段的 Compose 编译问题，包括字符串模板变量歧义、`RowScope.weight` 用法和 `ModalBottomSheet` 的 Material3 opt-in。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMockData.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvMockDataTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*'` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 12:49] Android 图集查看闪退修复（Offset Saveable）
- Type: `implementation`
- Summary:
  - 修复图集查看页进入即崩溃问题：`rememberSaveable` 直接保存 `Offset` 触发 `SaveableStateRegistry` 非 Bundle 类型异常。
  - 在图集 viewer 中为 `imageOffset` 引入自定义 `ImageViewerOffsetSaver`，把 `Offset` 序列化为可保存的 `List<Float>`，恢复时支持异常数据回退 `Offset.Zero`。
  - 新增 `saveImageViewerOffset/restoreImageViewerOffset` 内部函数，并补充单测覆盖“正常保存恢复”和“异常恢复回零”场景，确保后续不回归。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/imagecollections/ImageCollectionViewerZoomMathTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.imagecollections.ImageCollectionViewerZoomMathTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.imagecollections.*'` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 06:36] AV 海报精准刮削：主封面优先与降级覆盖策略
- Type: `implementation`
- Summary:
  - 对齐 references 规则重构 AV 海报质量分级：新增 `primary/fallback/invalid` 三档判定，并把 `noimage/logo/icon/banner` 等噪声 URL 统一判为无效，避免误抓站点标识图。
  - 重写 AV 合并海报选择逻辑：从“单一分数”改为“先 primary 后 fallback”的分层选择，输出 `poster_source/poster_quality/poster_decision` 到候选 metadata 与 scrape trace，增强可追踪性。
  - 调整 AV 落库封面策略（上传与确认共用）：`primary` 才覆盖；仅 `fallback` 时若已有封面则保留旧图、无旧图才下载使用；无效或相对路径海报不再触发 TMDB 前缀拼接。
  - 新增回归测试覆盖 `fallback` 保留旧封面、无旧封面使用 fallback、AV 相对路径不走 TMDB 下载，锁定本次行为变更。
- Changed Files:
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVKeepsExistingThumbnailWhenOnlyFallbackPoster|TestConfirmAVUsesFallbackPosterWhenNoExistingThumbnail|TestConfirmAVRejectsRelativePosterURLWithoutTMDBFallback' -v` passed.
  - `go test ./internal/services -run AV -v` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 06:17] 首页短视频切页后返回保持观看位置
- Type: `implementation`
- Summary:
  - 修复首页短视频在切换到其他页面再返回后回到第一个视频的问题：新增 pager 锚点记忆逻辑，持续记录当前页码与当前视频 ID，页面重建时优先恢复到上次观看位置。
  - 短视频数据重新加载时引入“锚点优先恢复页”策略：若锚点视频仍在列表中则恢复到该视频所在页；锚点缺失时回退到上一页码（并做边界钳制），避免无故回到首条。
  - 补货裁剪场景同步维护锚点视频 ID，确保窗口 trim 后 pager 重置仍对齐当前观看内容，保持回到首页时的连续观看体验。
  - 新增回归测试锁定 pager 恢复页解析规则，覆盖锚点命中、锚点缺失回退和页码越界钳制。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedPagerRestoreTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedPagerRestoreTest --tests com.chee.videos.feature.shorts.ShortFeedWindowManagerTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 06:10] App 图集预览支持按尺寸缩放与当前缩略图复位
- Type: `implementation`
- Summary:
  - 为 app 图集预览主图增加双指缩放与平移能力，最小缩放固定 `0.6x`，并按图片原始尺寸与容器尺寸动态计算最大缩放比例。
  - 对“小图”启用放大限制：当图片宽高都不超过当前可视容器时，最大缩放固定为 `1x`，仅支持缩小，不再允许放大。
  - 保留双击显隐 chrome 的既有行为，同时新增“点击当前缩略图恢复尺寸”交互：点击当前项时直接回到 `1x` 并居中；点击其他缩略图切页时也默认 `1x` 展示。
  - 新增图集 viewer 缩放数学回归测试，锁定小图禁放大、动态上限计算、低倍率回正与高倍率平移边界钳制，避免后续回退。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/imagecollections/ImageCollectionViewerZoomMathTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.imagecollections.ImageCollectionViewerZoomMathTest --tests com.chee.videos.feature.imagecollections.ImageCollectionViewerChromeTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-23 05:57] 短视频进度条改为增量拖动，不再按首次触点跳转
- Type: `implementation`
- Summary:
  - 重构 app 端短视频底部进度条手势语义：`onDragStart` 不再把进度映射到首次触点位置，仅进入拖动态；拖动过程中只按位移累计占比计算增量。
  - 首页短视频、瀑布流短视频和我的短视频统一播放器三处统一引入 `scrubAnchorMs` 锚点，拖动目标时间由 `anchor + duration * deltaFraction` 计算并做边界钳制，避免按下瞬间跳帧。
  - 新增短视频进度条数学回归测试，锁定“无拖动不跳转、正向/反向拖动按增量计算、越界钳制”的行为，防止后续回退到绝对触点跳转逻辑。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/ShortVideoBottomProgressBar.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/ShortVideoBottomProgressBarScrubMathTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortVideoBottomProgressBarScrubMathTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-22 16:27] 增加 Android 开发产物忽略规则
- Type: `docs`
- Summary:
  - 在仓库根 `.gitignore` 追加 Android 开发常见本地文件与构建产物忽略项，避免 `Gradle` 缓存、`IDE` 配置、`build` 目录、原生中间产物与本地签名文件被误提交。
  - 同时覆盖当前项目已使用的 `android-app/.gradle-local/` 目录，保持本地构建缓存与代码仓库隔离。
- Changed Files:
  - `.gitignore`
  - `plan.md`
- Verification:
  - `git status --short` passed (仅包含预期文件变更)。
- Rollback:
  - `git revert <commit>`

### [2026-04-22 11:32] 短视频详情接入相关图片合集预览
- Type: `implementation`
- Summary:
  - 扩展 app 端视频详情数据，在 `/api/v1/videos/:id` 中补充 `image_collection` 轻量字段，仅在图片合集启用且存在可浏览图片时返回，短视频详情可直接拿到关联图集入口。
  - Android 短视频详情层新增“相关图片”卡片，展示关联图片合集名称与封面缩略图；点击后直接复用现有 `image-collections/{collectionId}` 预览页，不再跳到图集瀑布流。
  - 调整短视频详情页到图片预览页的导航逻辑：打开图片合集时不关闭详情层，返回后恢复到原短视频详情状态，保持沉浸浏览链路连续。
  - 新增回归测试，锁定相关图片入口显隐、viewer 路由生成，以及 app 端视频图片合集封面 URL 的派生规则，防止后续又回到短视频发现流或丢失图集字段。
- Changed Files:
  - `internal/models/app.go`
  - `internal/repository/app_image_collection_repository.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/image_collection_repository_test.go`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedImageCollectionTest.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-25 00:38] 长视频字幕选择、后台上传与内嵌字幕抽取
- Type: `implementation`
- Summary:
  - 为 `movie`、`episode`、`av` 新增独立字幕资源模型与迁移，后端在视频详情和电视剧分集详情中返回 `subtitle_tracks`，并暴露字幕文件访问接口。
  - 新增字幕 service 与管理端接口，支持后台上传 `SRT/VTT` 外挂字幕、设默认轨、删除外挂字幕，以及在转码完成后自动探测并抽取内嵌字幕为可播放的 `VTT` 文件。
  - app 端长视频播放器新增字幕选择入口，电影详情、统一播放页和电视剧播放页都可按默认轨或手动选择字幕；管理端视频详情弹窗新增字幕管理区。
- Changed Files:
  - `migrations/0016_video_subtitles.up.sql`
  - `migrations/0016_video_subtitles.down.sql`
  - `internal/models/app.go`
  - `internal/models/admin.go`
  - `internal/models/models.go`
  - `internal/repository/video_subtitle_repository.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/tv_repository.go`
  - `internal/repository/admin_repository.go`
  - `internal/repository/video_repository.go`
  - `internal/services/subtitle.go`
  - `internal/handlers/router.go`
  - `internal/handlers/video_subtitle.go`
  - `internal/handlers/admin_video_subtitle.go`
  - `internal/handlers/admin.go`
  - `internal/queue/tasks.go`
  - `internal/utils/video_url.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/repository/video_delete_test.go`
  - `main.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/admin.spec.js`
  - `admin-web/src/views/VideoList.vue`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `cd admin-web && npm test && npm run build` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-25 09:22] Android 长视频字幕默认轨热切换崩溃修复
- Type: `implementation`
- Summary:
  - 修复长视频播放器在已开始播放后，因详情异步返回字幕轨而自动套用默认字幕、重建同一 `MediaItem` 导致的播放期热切换风险。
  - 新增字幕选择决策函数：仅在“尚未开始播放”时自动应用默认字幕；若视频已经开始播放且当前没有字幕选择，则保持关闭字幕，避免播放中因异步字幕到达触发播放器重配。
  - 详情页、统一播放页和电视剧播放页统一接入该策略，并补充回归测试锁定“播放前套默认、播放后不自动热切”的行为。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/SubtitleSelectionTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.SubtitleSelectionTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-24 10:48] 电视剧上传自动并入电视剧树与待绑定状态
- Type: `implementation`
- Summary:
  - 将 `type=episode` 自动刮削从“失败后静默回退普通上传视频”改为“高置信命中才自动并入电视剧树，失败则进入 `tv_pending` 待处理状态”，并在 metadata 中写入 `scrape_stage`、解析结果与候选摘要。
  - 重构 TMDB 电视剧同步逻辑：自动刮削、手工确认和旧 `SyncTVEpisode` 均改为同步整剧季/集骨架；命中分集绑定当前视频，其他分集仅补齐元数据且不覆盖已绑定 `video_id`，同时允许同一分集后续上传重新绑定到新视频。
  - 管理端视频列表新增 `tv_pending` 状态筛选与详情诊断面板，并提供跳转到刮削确认页、电视剧管理页的入口；手工确认成功后会自动补入转码队列。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `internal/services/scraper_episode_sync_test.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/queue/scrape_tasks_test.go`
  - `internal/repository/video_repository.go`
  - `internal/handlers/admin_scrape.go`
  - `migrations/0015_episode_tv_pending.up.sql`
  - `migrations/0015_episode_tv_pending.down.sql`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/videoList.helpers.js`
  - `admin-web/src/views/videoList.helpers.spec.js`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/TvSeriesManage.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-24 16:05] Android 图集查看页恢复左右滑动翻页
- Type: `implementation`
- Summary:
  - 修复图片查看页在未缩放状态下无法左右滑动切图的问题：根因是 `detectTransformGestures` 无条件挂在当前图片层上，单指横向拖动也会被图片手势先消费，`HorizontalPager` 收不到翻页手势。
  - 将图集查看页的缩放/平移手势切换为 Compose `transformable`，并用显式 helper 控制手势归属：`1x` 基础倍率时允许 pager 接管左右滑动；放大后才允许图片自身横向/纵向平移。
  - 新增回归测试锁定 `imageViewerPagerSwipeEnabled` 行为，避免后续再把基础倍率下的翻页手势吃掉。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/imagecollections/ImageCollectionViewerZoomMathTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.imagecollections.ImageCollectionViewerZoomMathTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.imagecollections.*'` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-22 10:28] App 图片合集瀑布流与沉浸式看图功能
- Type: `implementation`
- Summary:
  - 新增 app 端图片合集能力：后端提供 `/api/v1/image-collections`、`/api/v1/image-collections/:id` 和 `/api/v1/images/:id/view`，仅返回启用且可浏览的图片合集与图片，并为封面、缩略图、原图统一生成 app 端访问地址。
  - Android 首页短视频层增加“图片合集”浮层入口，进入后可浏览图集瀑布流卡片，展示合集标题与图片数量，并支持分页补货。
  - 新增沉浸式图集查看页：单张原图横向切换、底部缩略图条、顶部标题与页码、双击显隐上下 chrome，整体沿用当前 app 的深色影院视觉。
  - 新增回归测试，锁定后端路由注册、app 图片 URL 构建、viewer 双击显隐规则和封面 URL 生成，避免后续回退到管理端图片链路或丢失沉浸态交互。
- Changed Files:
  - `internal/handlers/app_image_collection.go`
  - `internal/handlers/recommend_test.go`
  - `internal/handlers/router.go`
  - `internal/models/app.go`
  - `internal/repository/app_image_collection_repository.go`
  - `internal/repository/image_collection_repository_test.go`
  - `internal/services/app.go`
  - `internal/services/image.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionViewerChrome.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/util/UrlBuilderImageCollectionsTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/imagecollections/ImageCollectionViewerChromeTest.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-21 23:24] 管理端图片合集查看关联图片并支持选图设封面
- Type: `implementation`
- Summary:
  - 为图片合集增加 `cover_image_id` 数据能力与迁移，合集输出同时保留兼容 `cover_url` 和原始 `manual_cover_url`，并优先把封面图片映射为管理端图片缩放预览地址。
  - 图片合集更新时新增封面图归属校验，只有当前合集已关联的图片才允许设为封面；删除封面图片后依赖外键自动清空 `cover_image_id`，回退到手填封面地址或空值。
  - 管理端图片合集列表新增“查看图片”抽屉，复用现有图片列表筛选与图片缩放接口按 blob 加载小图预览，并支持在抽屉内将任意关联图片设为合集封面。
  - 新增 Go/前端回归测试，锁定管理端图片预览 URL 生成、合集封面 URL 优先级、合集保存 payload 和 blob 预览 URL 释放逻辑。
- Changed Files:
  - `migrations/0013_image_collection_cover_image.up.sql`
  - `migrations/0013_image_collection_cover_image.down.sql`
  - `internal/models/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/repository/image_collection_repository.go`
  - `internal/repository/image_collection_repository_test.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `admin-web/src/views/ImageCollectionManage.vue`
  - `admin-web/src/views/imageCollectionManage.helpers.js`
  - `admin-web/src/views/imageCollectionManage.helpers.spec.js`
  - `plan.md`
- Verification:
  - `cd .worktrees/image-collection-cover && GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd .worktrees/image-collection-cover/admin-web && npm test` passed.
  - `cd .worktrees/image-collection-cover/admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-21 22:02] 管理端图片上传超时修复
- Type: `implementation`
- Summary:
  - 定位根因：管理端公共请求实例默认 `timeout: 30000`，而图片上传接口沿用了默认配置；当多图上传或服务端同步处理时间超过 30 秒时，会被前端 Axios 提前中断并表现为 timeout。
  - 为管理端图片上传接口显式关闭请求超时，保持与大文件/长耗时上传场景一致，避免上传仍在服务端处理时前端先失败。
  - 新增管理端 API 回归测试，锁定 `/admin/images/upload` 必须携带 `timeout: 0`，防止后续重构时再次回退到默认超时。
- Changed Files:
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/admin.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test -- src/api/admin.spec.js` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-21 21:18] Android 短视频独立展示模式与轻量化悬浮按钮
- Type: `implementation`
- Summary:
  - 首页短视频右侧交互按钮统一改为更浅的半透明悬浮底色，减少深色块对视频画面的遮挡，同时保留点赞、收藏、比例切换和详情入口。
  - 为瀑布流短视频播放页新增独立的比例切换按钮与持久化偏好，播放层视频与封面统一跟随 `FILL/FIT` 模式切换，且不影响首页短视频的既有模式。
  - 为“我的”里的统一播放器短视频新增独立比例切换偏好，仅在 `type == "short"` 时展示切换按钮，长视频与 AV 播放保持现有逻辑不变。
  - 新增共享短视频展示映射与 JVM 回归测试，锁定封面缩放、播放器 `resizeMode` 以及统一播放器短视频切换按钮的展示条件。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/ShortVideoPresentation.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerShortState.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/ShortVideoPresentationTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/player/UnifiedPlayerShortFitToggleTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd /Users/chee/Documents/workspace/ai-project/ai-video-server/.worktrees/short-fit-modes/android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.ShortVideoPresentationTest --tests com.chee.videos.feature.player.UnifiedPlayerShortFitToggleTest --tests com.chee.videos.feature.shorts.ShortFeedPosterScaleTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd /Users/chee/Documents/workspace/ai-project/ai-video-server/.worktrees/short-fit-modes/android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-21 20:41] Android 短视频详情与 AV 详情沉浸式改版
- Type: `implementation`
- Summary:
  - 短视频播放页新增右侧操作列显隐判定，详情层打开时自动隐藏点赞、收藏、比例切换和详情按钮，关闭后恢复，减少详情阅读时的画面干扰。
  - 重构短视频详情层视觉层级：改为标题头部卡、三列统计卡、简介卡、合集与标签胶囊区，并将“不喜欢”下沉为次级危险操作，整体统一为深色影院感。
  - 新增 AV 详情页动作规格模型，非全屏 AV 详情改为主视觉播放器 + 标题元信息卡 + 数据卡 + 简介卡 + 主次操作区的分层布局，保留原有播放、全屏和交互能力。
  - 新增 Android 单元测试锁定短视频操作列显隐规则与 AV 详情操作规格，防止后续回归到原来的平铺式交互结构。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedProgressState.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailActionState.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedActionRailVisibilityTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/DetailActionSpecTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedActionRailVisibilityTest --tests com.chee.videos.feature.shorts.ShortFeedProgressVisibilityTest --tests com.chee.videos.feature.detail.DetailActionSpecTest --tests com.chee.videos.feature.detail.LongFormPlaybackSessionTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-21 19:39] Android 首页导航压缩：移除说明文案并扩大短视频首屏
- Type: `implementation`
- Summary:
  - 首页顶部去掉“私人影库”和各分类解释文案，改为单行纯文字分类切换，保留 `短视频 / 电影 / 电视剧 / AV` 四类入口。
  - 底部导航去掉 icon 与解释文案，收缩为 `首页 / 我的` 两项纯文字按钮，把更多垂直空间让给首页短视频画面。
  - 新增可单测的导航配置 `AppNavigationConfig` 与 JVM 回归测试，锁定顶部四类和底部两项的精简结构，避免后续重新带回解释字段。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppNavigationConfig.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppNavigationConfigTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.core.ui.AppNavigationConfigTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-20 22:13] 管理端固定侧栏与全站视觉提质
- Type: `implementation`
- Summary:
  - 完成管理端共享壳层重构，固定侧栏导航在桌面端保持常驻，移动端切换为抽屉导航，统一信息结构与入口可达性。
  - 收敛全站主题变量与组件覆盖策略，修复历史样式冲突并提升页面层级、间距、卡片与表格可读性。
  - 将管理端核心页面统一迁移到新骨架布局，修复交互异常并保证页面在新视觉基线下的一致表现。
- Changed Files:
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/assets/theme.css`
  - `admin-web/src/views/ActorManage.vue`
  - `admin-web/src/views/CollectionManage.vue`
  - `admin-web/src/views/Dashboard.vue`
  - `admin-web/src/views/ImageCollectionManage.vue`
  - `admin-web/src/views/ImageManage.vue`
  - `admin-web/src/views/Login.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/SystemSettings.vue`
  - `admin-web/src/views/TaskMonitor.vue`
  - `admin-web/src/views/UserManage.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `cd admin-web && npm run build` passed.
  - 管理端视觉与交互人工验收未在纯终端执行，待浏览器人工验收。
- Rollback:
  - `git revert <commit>`

### [2026-04-19 12:26] 修复短视频尺寸模式切换时封面不跟随的问题
- Type: `implementation`
- Summary:
  - 定位根因：短视频尺寸模式切换只影响 `PlayerView.resizeMode`，封面图仍固定使用 `ContentScale.Crop`，导致视频切到完整显示时封面没有同步切换。
  - 在短视频页提取统一的封面缩放映射 `shortPosterContentScale(fitMode)`，使封面图与视频共用同一套 `VideoFitMode` 语义：`FILL -> Crop`，`FIT -> Fit`。
  - 当前页显示中的封面与非当前页预览封面统一接入该映射，避免只修一条分支后再次出现模式不一致。
  - 新增 Android 单元测试锁定短视频封面缩放映射，防止后续改动回归。

### [2026-04-22 15:28] App 图集入口迁移到底栏并支持标题搜索
- Type: `implementation`
- Summary:
  - 将 app 底部导航扩展为 `首页 / 图集 / 我的` 三项，图集列表页作为一级 tab 展示；首页短视频层移除原先占空间的图片合集悬浮入口，把更多首屏空间还给短视频。
  - 为 app 图集瀑布流增加标题搜索框，输入后通过防抖自动触发服务端标题检索，清空搜索可回到默认图集列表，并保持现有沉浸式图集 viewer 入口与行为不变。
  - 扩展 `/api/v1/image-collections` 支持 `q` 参数，后端仅按合集标题做大小写不敏感过滤，同时保留现有启用状态、可浏览图片和分页排序约束。
  - 新增回归测试，锁定底部导航三项结构、图集搜索状态重置规则，以及后端标题搜索模式生成，防止入口回退到首页浮层或搜索参数丢失。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppNavigationConfig.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/imagecollections/ImageCollectionsViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/core/ui/AppNavigationConfigTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/imagecollections/ImageCollectionsSearchStateTest.kt`
  - `internal/handlers/app_image_collection.go`
  - `internal/repository/app_image_collection_repository.go`
  - `internal/repository/app_image_collection_repository_search_test.go`
  - `internal/services/app.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedPosterScaleTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedPosterScaleTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 12:03] 修复短视频只能滑 20 条：尾部补货 + 滚动窗口 + 后端排重补货
- Type: `implementation`
- Summary:
  - 后端短视频随机接口新增 `exclude_ids` 解析与排重能力，优先排除客户端当前窗口中的短视频；当排除后不足以凑满批次时，自动回退补货，避免把短视频流刷到死路。
  - Android 短视频数据层接入 `exclude_ids` 参数，请求下一批短视频时携带当前窗口视频 ID，实现短距离内优先不重复。
  - 新增 `ShortFeedWindowManager`，将短视频流改为“首批加载 + 尾部预取 + 60 条滚动窗口”，接近尾部时自动补货，超过窗口上限时裁掉头部旧视频并同步清理详情/暂停/操作缓存。
  - `ShortFeedScreen` 的 `VerticalPager` 改为支持裁头后按当前视频重新定位，保证无限滑动过程中不会在窗口裁剪后错位到其他视频。
  - 新增后端与 Android 单元测试，覆盖 `exclude_ids` 解析、随机短视频补齐合并、短视频窗口裁剪与锚点保留逻辑。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/recommend.go`
  - `internal/handlers/recommend_test.go`
  - `internal/services/recommend.go`
  - `internal/repository/video_repository.go`
  - `internal/repository/video_repository_random_test.go`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedWindowManager.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedWindowManagerTest.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers ./internal/repository` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedWindowManagerTest` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 11:14] 修复长视频播放器暂停退回海报/空态，并收紧控制栏交互
- Type: `implementation`
- Summary:
  - 新增长视频播放状态机 `LongFormPlaybackSession`，将“已进入播放态”和“用户主动暂停”拆分，修复详情页暂停后误卸载播放器，避免小屏退回海报、全屏退回“暂无可播放视频”。
  - 详情页 `DetailScreen` 改为基于播放状态机驱动：首次播放后始终保留播放器挂载，生命周期恢复只在非用户暂停时自动继续播放。
  - 长视频播放器 `LongFormVideoPlayer` 新增双击播放/暂停、中间反馈浮层、长按 2.0x 倍速提示，并将顶部/底部控制栏改为更紧凑的暗黑悬浮条，降低对画面的遮挡。
  - 新增单元测试覆盖关键回归：开始播放后暂停仍应保留播放器挂载态，恢复播放后仍可自动续播。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/LongFormPlaybackSession.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/detail/LongFormPlaybackSessionTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.detail.LongFormPlaybackSessionTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 09:12] 清理长视频播放器重复导入并复核编译
- Type: `implementation`
- Summary:
  - 清理 `LongFormVideoPlayer.kt` 中重复 `import`，消除冗余引用，保持代码整洁。
  - 复跑 Android 构建确认播放器改版在清理后仍可正常编译。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 09:10] 长视频播放器改版（电影/剧集/AV）+ 历史类型分流（短视频保持旧逻辑）
- Type: `implementation`
- Summary:
  - 新增长视频播放器控件 `LongFormVideoPlayer`：自定义顶部/底部控制栏、单击显隐、5 秒自动隐藏、左右滑动快进快退（中间浮层实时提示）、长按 2.0x 倍速松手恢复、全屏/退出全屏按钮。
  - 详情页 `DetailScreen` 对 `movie/episode/av` 启用新长视频控件，保留 `short` 旧逻辑；全屏切换仅改方向与 UI，不重建播放器，播放进度无缝衔接。
  - 统一播放器 `UnifiedPlayerScreen` 按视频类型分流：`short` 继续旧短视频交互，`movie/episode/av` 使用新长视频控件；新增长视频全屏横屏沉浸与返回键退出全屏。
  - 后端 `history/continue` 返回项新增 `type` 字段；Android 端 `HistoryItemDto`/`MineViewModel`/`UnifiedPlayerViewModel` 全链路接入，确保历史列表可精确识别 short 并保持旧逻辑不变。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt`
  - `internal/models/app.go`
  - `internal/repository/app_repository.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 08:35] 修正 release 忽略规则位置：回退模块规则并在根 .gitignore 添加 `release`
- Type: `implementation`
- Summary:
  - 回退上一版误加到 `android-app/.gitignore` 的 `/app/release/` 规则，恢复到变更前状态。
  - 按要求在根 `.gitignore` 增加忽略规则 `release`，统一忽略名为 `release` 的目录/文件。
  - 使 `android-app/app/release/` 不再出现在 `git status` 未跟踪列表中。
- Changed Files:
  - `.gitignore`
  - `plan.md`
- Verification:
  - `git status --short` no longer shows `android-app/app/release/`.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 08:28] Android 视频播放防息屏：仅播放中保持亮屏，暂停/退出自动恢复
- Type: `implementation`
- Summary:
  - 新增通用 `KeepScreenOnEffect(enabled)`，统一管理 `FLAG_KEEP_SCREEN_ON`，支持进入播放时开启、暂停或退出时自动清除。
  - 短视频流接入真实播放态监听（`Player.onIsPlayingChanged`），仅当前视频实际播放中保持亮屏。
  - 统一播放器（历史/收藏/喜欢）接入真实播放态监听，滑动切片、暂停、返回页面时均按播放状态自动恢复息屏策略。
  - 详情页内嵌播放器接入真实播放态监听，含 AV 详情全屏/非全屏场景；播放时防息屏，暂停或离开详情页后恢复系统息屏。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/KeepScreenOnEffect.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 00:23] 修复 AV 详情点击全屏闪退回退：避免横竖屏切换触发 Activity 重建
- Type: `implementation`
- Summary:
  - 定位问题：详情页进入全屏时会强制切换横屏，默认 Activity 会因 `orientation/screenSize` 变化重建，表现为点击全屏后页面闪退回退。
  - 在 `MainActivity` 增加 `configChanges`，由当前 Activity 自行处理方向与屏幕配置变更，避免全屏切换导致重建丢失当前 UI 状态。
  - 详情页全屏方向锁定从 `SCREEN_ORIENTATION_SENSOR_LANDSCAPE` 调整为 `SCREEN_ORIENTATION_USER_LANDSCAPE`，减少传感器抖动触发的方向切换波动。
  - 保持原有沉浸式逻辑不变：全屏隐藏系统栏，退出后恢复系统栏与先前方向。
- Changed Files:
  - `android-app/app/src/main/AndroidManifest.xml`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 23:54] Android 全页面状态栏统一：留白暗黑背景 + 白色图标，保留全屏沉浸
- Type: `implementation`
- Summary:
  - 统一 Activity 系统栏样式为暗色背景上的白色状态栏/导航栏图标，避免页面切换后状态栏图标明暗不一致。
  - 为首页、登录页、连接页、我的页、统一播放器页补齐 `statusBarsPadding()`，保证非全屏页面顶部不遮挡状态栏并保留暗黑留白区域。
  - 连接页背景改为暗黑色，与全局暗黑风格一致，避免状态栏留白区域出现亮色断层。
  - 详情页 `Scaffold` 增加 `contentWindowInsets = WindowInsets(0, 0, 0, 0)`，与全局手动 inset 策略对齐，避免重复 inset。
  - 全屏播放逻辑保持不变：全屏时隐藏系统栏，退出全屏后恢复系统栏显示与统一样式。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/MainActivity.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/auth/LoginScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 22:01] Android AV 播放体验增强：时长格式化 + 详情内嵌播放 + 海报回退
- Type: `implementation`
- Summary:
  - AV 列表时长改为 `HH:MM:SS` 显示，不再以“秒”直出。
  - 详情页新增可点击“播放/暂停”按钮，并接入 Media3 内嵌播放器在详情页直接播放当前视频。
  - 详情页新增海报区域：优先 `metadata.poster_url`，其次 `metadata.poster_path`，最后回退 `thumbnail_path`（视频封面）。
  - 详情页播放地址与海报路径支持相对 URL 自动拼接 `baseUrl`，并复用登录 token 作为 `Authorization` 请求头。
  - 详情页状态新增 `baseUrl/accessToken`，用于播放与海报 URL 解析。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 21:33] 修复“我的”封面不显示：缩略图路径由本地绝对路径改为可访问 API URL
- Type: `implementation`
- Summary:
  - 定位根因：后端在历史/喜欢/收藏等接口中直接返回 `thumbnail_path` 的本地绝对路径（如 `/Volumes/.../thumb.jpg`），App 无法通过 HTTP 访问导致封面不显示。
  - 新增应用缩略图接口：`GET /api/v1/videos/:id/thumbnail`，按视频 ID 读取并返回缩略图文件。
  - 统一 App 返回字段：在详情、历史、搜索、喜欢/收藏等应用查询中，将 `thumbnail_path` 映射为 `/api/v1/videos/{id}/thumbnail`。
  - 补齐推荐/短视频返回：推荐视频模型新增 `thumbnail_path`，随机短视频/候选/热门查询返回并映射可访问封面 URL。
  - 前端兼容增强：首页分类列表支持相对缩略图路径拼接（`baseUrl + thumbnail_path`），避免仅 http/https 才渲染的问题。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/video_repository.go`
  - `internal/models/models.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 21:04] Android 首页全屏精简：移除顶部栏，Tab 标题改白
- Type: `implementation`
- Summary:
  - 首页移除顶部标题栏及其操作按钮（路由切换、退出登录），将控制入口统一收敛到“我的”页面。
  - 首页改为全屏内容布局，仅保留内容 Tab 与视频流区域，提升可视播放空间。
  - 内容 Tab（短视频/电影/电视剧/AV）标题改为白色风格，未选中态使用白色低透明度。
  - 导航层移除 `HomeScreen` 不再需要的 `onSwitchServer/onLogout` 参数传递。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 20:47] 修复“我的”三分区无数据：历史上报缺失 + 继续观看过滤过严
- Type: `implementation`
- Summary:
  - 定位根因 1：Android 端此前未在短视频/统一播放器中上报观看进度，导致后端 `user_video_actions(view)` 无新增数据，历史记录长期为空。
  - 定位根因 2：后端 `/history/continue` 查询逻辑只返回“未看完”视频，会过滤掉已看完视频，不符合“我的-历史记录”预期。
  - 修复历史写入：在 `ShortFeedScreen` 与 `UnifiedPlayerScreen` 中，切换视频与页面销毁时自动上报 `watch_seconds/completed`。
  - 修复历史查询：`ContinueWatching` 改为返回所有 `view` 记录（`watch_seconds>0`），并保留进度计算（上限 1）。
  - 修复分区刷新：切换“历史记录/收藏/喜欢”分区时强制刷新，避免首次空数据后长期缓存不更新。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineViewModel.kt`
  - `internal/repository/app_repository.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 19:45] Android “我的”三分区新增分页下拉加载（下拉刷新 + 滚动加载更多）
- Type: `implementation`
- Summary:
  - 为“历史记录 / 我的收藏 / 我的喜欢”三分区增加分页状态管理：`page / hasMore / loadingMore / refreshing`。
  - 新增加载策略：首次加载、下拉刷新（重置到第 1 页）、滚动到底自动加载下一页。
  - 列表底部新增分页状态反馈：加载中、无更多数据、加载失败点击重试。
  - 引入 `compose material` 依赖并接入 `pullrefresh`，实现“我的”页面下拉刷新指示器。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineScreen.kt`
  - `android-app/app/build.gradle.kts`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 17:48] Android App 新增“我的”面板 + 统一竖滑播放器（暗黑风格）
- Type: `implementation`
- Summary:
  - 扩展 Android 数据层：新增 `history/continue`、`user/liked-videos`、`user/favorited-videos`、`user/profile` 四类接口与仓库方法，复用现有鉴权刷新链路。
  - 重构已登录导航壳：新增底部主导航 `首页/我的`，保留详情页路由，并新增 `player/{source}/{videoId}` 统一播放器路由。
  - 新增“我的”模块：暗黑风格个人面板，包含 `历史记录/我的收藏/我的喜欢` 三个分区，支持资料展示、列表加载态/空态/错误态、点击条目直达播放。
  - 新增统一播放器模块：支持按来源（历史/收藏/喜欢）加载播放队列、竖向滑动切换、轻触暂停/继续、封面预加载及演员信息展示。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 15:20] AV 刮削增强：多站点接入 + 字段级置信合并（修复海报/简介/演员错抓）
- Type: `implementation`
- Summary:
  - 扩展 AV crawler provider，新增 `javbus/javlibrary/fc2`，形成 `javdb + javbus + javlibrary + fc2` 多站点检索与详情抓取能力。
  - 重构 AV 搜索结果聚合为“字段级置信合并”：按 `code/title` 分组后对 `title/code/poster/overview/release_date/actors` 分字段择优，输出融合候选并保留 `field_sources/merged_sources/raw_candidates` 追踪信息。
  - 修复 JavDB 字段误抓：海报优先 `video-cover` 并过滤 logo；简介优先详情字段/正文块再回退 meta；演员改为演员区域优先提取，过滤噪音链接。
  - 增强 AV 结果透传与确认链路：`PreviewAV` 返回 `detail_url/scrape_source`；`ScrapeAVUpload` 与 `ConfirmAV` 元数据改为动态来源存储，不再固定写死 `javdb`。
  - 增强自动刮削任务参数：队列在 AV 自动确认时补齐 `detail_url/scrape_source` hint，确保多来源候选可正确回放详情抓取。
  - 新增回归测试：覆盖 logo 海报误抓、SEO 简介误抓、演员噪音过滤，以及多站点冲突字段的择优合并。
- Changed Files:
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/handlers/admin_scrape.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 12:52] 修复重转码仍不执行（输入输出同路径 + 失败收口 SQL 类型错误）
- Type: `implementation`
- Summary:
  - 定位并修复重转码不执行根因一：当 `original_path` 缺失并回退到 `transcoded_path` 时，转码输入与输出同一路径，ffmpeg 无法原地覆盖导致立即失败。
  - 转码服务新增“同路径保护”：若输入路径与输出路径相同，先转码到同目录临时文件，完成后再替换目标输出，避免 `Output same as Input` 错误。
  - 定位并修复根因二：失败收口 SQL 参数类型推断错误（`MarkVideoFailed/FinishTranscodingJob`），导致任务失败后仍卡在 `running`、视频卡在 `processing`。
  - 为失败收口 SQL 增加显式类型转换（`::text`），确保失败时任务和视频状态能正确落库。
  - 新增单测覆盖同路径判定与临时输出路径生成。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestIsSameFilePath|TestBuildTranscodeOutputTempPath|TestDecideVideoBitrate|TestResolveProbeFieldsWithValidProbeParsesValues' -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 12:33] 修复重转码假运行与任务监控剩余时间异常
- Type: `implementation`
- Summary:
  - 修复“重转码无 ffmpeg 进程”的入口问题：`AdminRetranscodeVideo` 改为优先使用存在的 `original_path`，不存在时回退 `transcoded_path`，两者都不存在时直接返回业务错误，避免入队无效任务。
  - 修复“任务长期 running”风险：转码失败路径新增统一失败收口，`MarkVideoFailed` 与 `FinishTranscodingJob` 失败不再静默；当请求上下文已失效时自动切换短超时后台上下文继续收口并记录错误日志。
  - 新增卡住任务自动修复：`AdminTasks` 查询前执行 `HealStaleRunningTranscodingJobs`，将长时间无进度的 `running` 任务自动标记为 `failed`，并同步回写视频状态与错误信息。
  - 修复任务页剩余时间显示：前端 `TaskMonitor` 将 `null/undefined/''` 视为无值，`remaining_seconds` 为空时展示 `--`，不再误显示 `00:00`。
- Changed Files:
  - `internal/handlers/admin.go`
  - `internal/handlers/admin_retranscode_test.go`
  - `internal/repository/admin_repository.go`
  - `internal/queue/tasks.go`
  - `admin-web/src/views/TaskMonitor.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 11:49] 管理端转码任务新增实时剩余时间（ETA）展示
- Type: `implementation`
- Summary:
  - 为 `transcoding_jobs` 新增进度字段：`source_duration_seconds/processed_seconds/remaining_seconds/progress_percent/progress_updated_at`，用于承载转码实时进度与 ETA。
  - 转码执行保持 `hevc_videotoolbox` 不变，新增 ffmpeg 实时进度解析（`-progress pipe:1`），按媒体时间计算并回调上报进度；任务结束时自动收口为成功/失败终态。
  - 后端任务列表接口 `/api/v1/admin/tasks` 返回新增进度字段，失败重试时会重置旧进度，避免展示脏状态。
  - 管理端“任务监控”页面新增进度条、剩余时间、已耗时与进度更新时间列；剩余时间仅在 `running` 状态展示。
- Changed Files:
  - `migrations/0011_transcoding_job_progress.up.sql`
  - `migrations/0011_transcoding_job_progress.down.sql`
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/services/transcode.go`
  - `internal/queue/tasks.go`
  - `internal/models/admin.go`
  - `internal/repository/video_repository.go`
  - `internal/repository/admin_repository.go`
  - `admin-web/src/views/TaskMonitor.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 11:28] 调整转码策略（保持 Mac 硬件 H.265，加分辨率码率阈值）
- Type: `implementation`
- Summary:
  - 转码链路保持 `hevc_videotoolbox`（Mac 硬件加速）不变，将“仅 CRF”策略升级为“分辨率+源码率”目标码率策略。
  - 1080 档（宽>=1920 或 高>=1080，且非 4K）若源码率 > 4000k 则压到 4000k，否则保持源码率；4K 档（宽>=3840 或 高>=2160）若源码率 > 8000k 则压到 8000k，否则保持源码率。
  - 当输入视频无法探测到有效源码率时，自动回退到原有 CRF 策略，避免任务失败。
  - 扩展转码元数据：新增 `transcode_mode/resolution_tier/source_bitrate_kbps/target_bitrate_kbps/bitrate_capped`，并在 CRF 回退模式保留 `crf` 字段。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 09:43] 短视频支持改类为电影/剧集/AV并自动异步刮削
- Type: `implementation`
- Summary:
  - 管理端视频详情支持将 `short` 改为 `movie/episode/av`，并在改类后自动触发后台异步刮削。
  - 后端 `AdminUpdateVideo` 增加改类校验：仅允许 `short -> movie|episode|av`，其中改为 `episode` 时强制要求 `season_number/episode_number > 0`。
  - 改类为非短视频时自动清空视频合集关系；自动刮削任务会将状态置为 `scraping`，完成后置为 `ready`，失败时记录 `scrape_error` 并回退为 `uploaded`。
  - 新增队列任务 `video:scrape:retag`，采用“预览候选取第一个 + Confirm*”流程执行自动刮削，且不再触发转码任务。
  - 管理端详情弹窗新增“改为类型”和“季/集”输入（仅分集），并在触发改类时提示会后台自动刮削与清空合集。
- Changed Files:
  - `internal/handlers/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/queue/tasks.go`
  - `internal/queue/scrape_tasks.go`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 09:03] 修复大视频分片上传在合并阶段前端超时
- Type: `implementation`
- Summary:
  - 修复大视频上传时在 `upload/complete` 长时间等待后触发前端超时的问题。
  - 为分片上传关键接口设置独立超时策略：`uploadChunk` 与 `uploadComplete` 关闭 axios 30 秒默认超时（`timeout: 0`），避免大文件慢网络或后端合并耗时导致请求被客户端提前中断。
- Changed Files:
  - `admin-web/src/api/video.js`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 08:40] 修复短视频封面预加载未命中导致切换黑屏
- Type: `implementation`
- Summary:
  - 修复短视频切换时短暂黑屏问题：此前预加载请求使用了自定义 `memory/disk cache key`，但页面展示 `AsyncImage` 走 URL 默认 key，导致预加载缓存无法命中渲染请求。
  - 预加载策略改为统一使用 URL 作为缓存键，并扩大预加载范围为当前页前后各 2 条，提升连续滑动时封面命中率。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 08:32] 短视频封面预加载优化（前后相邻视频）
- Type: `implementation`
- Summary:
  - 在短视频流中新增相邻封面预加载：当前页变化时预热上一条与下一条视频 `thumbnail`，降低切换瞬间封面未命中导致的错帧感。
  - 首帧渲染判定改为依据 `sharedPlayer.currentMediaItem.mediaId`，避免异步事件竞争导致占位图过早隐藏。
  - 非当前页不再挂载 `PlayerView`，仅显示封面占位，减少播放器视图复用带来的串帧概率。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 08:25] 修复短视频切换后显示上一个视频封面的问题
- Type: `implementation`
- Summary:
  - 修复短视频切换后加载阶段显示上一条视频画面的视觉问题。
  - 切换视频源时先 `stop + clearMediaItems` 清空旧画面，并显式关闭 `PlayerView` 的 `keepContentOnPlayerReset`，避免复用播放器时残留上一帧。
  - 在当前活跃视频页面增加封面占位图（优先使用当前视频 `thumbnail_path`），待首帧渲染后自动隐藏，占位期间不再出现“上一条封面”错位。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 08:18] 修复短视频连续滑动触发 ExoPlayer OOM（单页多播放器并发）
- Type: `implementation`
- Summary:
  - 修复短视频页 `OutOfMemoryError`：原实现在 `VerticalPager` 每个页面都创建并 `prepare` 独立 ExoPlayer，连续滑动时会并发占用解码与缓冲内存。
  - 改为短视频页级别单实例 ExoPlayer 复用：仅当前页挂载 `PlayerView`，切页时切换 `MediaSource`，并释放页面级多实例。
  - 保留轻触暂停/播放、模式切换、右侧交互与详情弹层行为不变，同时在生命周期 `ON_PAUSE/ON_RESUME` 统一控制单播放器。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 00:55] 修复管理端批量上传在 HTTP 环境下哈希计算失败（digest undefined）
- Type: `implementation`
- Summary:
  - 修复管理端视频批量上传报错 `cannot read property of undefined('digest')`。
  - `sha256File` 新增无 `crypto.subtle` 场景的回退路径：当浏览器不支持 `WebCrypto SubtleCrypto`（常见于 `http + 局域网IP`）时，使用 `js-sha256` 进行分片 SHA-256 计算。
  - 保持原有进度回调与秒传逻辑不变，同时兼容图片上传复用哈希函数的场景。
- Changed Files:
  - `admin-web/package.json`
  - `admin-web/package-lock.json`
  - `admin-web/src/utils/hash.js`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 23:09] 修复短视频页演员解析空指针崩溃（actors/tags/collections 为空）
- Type: `implementation`
- Summary:
  - 修复短视频页 `extractActorNames` 在后端返回 `actors: null` 时触发的空指针崩溃。
  - Android 详情模型中 `tags/actors/collections/metadata` 改为可空类型，避免 Gson 将 `null` 写入 Kotlin 非空字段导致运行时 NPE。
  - 短视频详情弹层与旧详情页统一改为 `orEmpty()` 读取集合字段，兼容接口空值返回。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 22:50] Android 短视频页升级为抖音暗黑风：右侧互动、详情底部弹层与演员/合集展示
- Type: `implementation`
- Summary:
  - 短视频右侧操作区改为纯图标并重排为“喜欢/收藏/播放模式/详情”，移除文字提示；播放模式图标样式与详情图标统一为低干扰暗色风格。
  - 喜欢与收藏操作从详情页前移到短视频右侧交互，详情弹层只保留“不喜欢”动作。
  - 短视频详情改为页内底部弹层（约 3/4 屏），弹层打开时视频区域收缩为顶部约 1/4 并持续播放，不再跳转独立详情页。
  - 短视频底部文案仅保留标题；若存在关联演员则展示演员信息；详情弹层新增“关联合集”展示。
  - 后端视频详情接口补充 `actors` 字段，Android 详情模型同步支持 `actors/collections/metadata`，并在演员缺失时回退解析 `metadata.actors`。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt`
  - `internal/models/app.go`
  - `internal/repository/app_repository.go`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 22:01] Android 短视频页新增抖音风播放模式切换与轻触暂停
- Type: `implementation`
- Summary:
  - 在短视频播放页新增两种展示模式：`铺满（fill）` 与 `完整（fit）`，并通过 DataStore 持久化用户选择。
  - 新增轻触视频区域暂停/继续播放能力，按视频粒度记忆暂停状态，翻页时仅当前页自动播放。
  - 重构短视频单页 UI 为抖音风样式：右侧悬浮操作区（模式切换/详情）、底部渐变信息区、中心播放状态提示。
  - 右侧按钮事件改为 `clickable`，降低与全屏轻触手势冲突风险。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/model/VideoFitMode.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 19:50] 清理 Android 提交误纳入的 Gradle 本地缓存
- Type: `implementation`
- Summary:
  - 修复 Android 首版提交中误纳入 `.gradle-local` 构建缓存的问题，避免仓库体积膨胀。
  - 更新 `android-app/.gitignore`，新增 `/.gradle-local/` 忽略规则。
  - 将已追踪的 `.gradle-local` 文件从版本库移除，仅保留源码与构建脚本。
- Changed Files:
  - `android-app/.gitignore`
  - `plan.md`
- Verification:
  - `git rm -r --cached android-app/.gradle-local` executed.
  - `git status --short` no longer tracks `android-app/.gradle-local/**`.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 19:46] Android 客户端首版实现（服务发现+登录+分类+短视频竖滑）
- Type: `implementation`
- Summary:
  - 新增独立 Android 工程 `android-app`，采用 Kotlin + Jetpack Compose + MVVM + Hilt + Retrofit。
  - 实现首次启动连接流：局域网同网段嗅探 `/healthz`，失败后手动输入 IP/端口；支持历史地址保存、复用与删除。
  - 实现登录与会话：对接 `/api/v1/auth/login`、`/api/v1/auth/refresh`，本地持久化 token 并在鉴权失败时自动刷新。
  - 实现主页四分类展示：短视频、电影、电视剧（`episode`）、AV，按类型分流加载。
  - 实现短视频“抖音式”全屏竖滑播放：`VerticalPager + ExoPlayer`，视频流地址走 `/api/v1/videos/:id/source`。
  - 实现详情占位页与互动操作（点赞/收藏/不喜欢）基础能力。
- Changed Files:
  - `android-app/settings.gradle.kts`
  - `android-app/build.gradle.kts`
  - `android-app/gradle.properties`
  - `android-app/gradle/wrapper/*`
  - `android-app/gradlew`
  - `android-app/gradlew.bat`
  - `android-app/.gitignore`
  - `android-app/README.md`
  - `android-app/app/build.gradle.kts`
  - `android-app/app/src/main/AndroidManifest.xml`
  - `android-app/app/src/main/res/values/themes.xml`
  - `android-app/app/src/main/res/values/strings.xml`
  - `android-app/app/src/main/res/xml/network_security_config.xml`
  - `android-app/app/src/main/java/com/chee/videos/**`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME=\"$PWD/.gradle-local\" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 14:12] 管理端 token 失效统一处理（提示+跳转登录）
- Type: `implementation`
- Summary:
  - 修复管理端在 token 失效时未统一报错或跳转登录的问题。
  - 在前端请求响应拦截器中新增业务错误码鉴权判断：当返回 `code=401/403` 或消息包含 token/authorization/admin only 时，统一清理登录态并提示后跳转登录页。
  - 保留 HTTP 401 分支处理，并改为复用统一逻辑，避免重复代码和多次重定向。
- Changed Files:
  - `admin-web/src/api/request.js`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 14:03] dev-up 启动增加镜像预检与 descriptor 异常重试提示
- Type: `implementation`
- Summary:
  - 针对 `unable to fetch descriptor ... content size of zero` 启动失败场景，在 `dev-up.sh` 新增镜像预检流程。
  - 新增镜像拉取重试逻辑：仅在镜像本地不存在时执行 `docker pull`，并对 descriptor 异常给出可执行修复指引（镜像源配置问题排查）。
  - 新增 `compose_up_db_services`，在支持 `--pull` 时使用 `--pull missing`，减少重复启动时不必要拉取导致的失败概率。
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
  - `bash scripts/dev-up.sh --frontend off` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 12:50] AV 刮削框架化重构计划（参考 mdcx）与技能沉淀
- Type: `plan`
- Summary:
  - 参考 `references/mdcx` 的 crawler provider、模板流程与 parser 抽象，重构 Go AV 刮削内部架构。
  - 交付“技能+代码”：先沉淀可复用技能，再实现 JavDB 强化和多站点扩展预留。
  - 保持现有外部接口兼容，增强编号解析、trace 可观测性、演员自动关联稳定性。
- Changed Files:
  - `plan.md`
- Verification:
  - 计划项，无需构建/测试。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 12:50] 实现 AV crawler provider + 模板流程重构、trace 与编号匹配增强
- Type: `implementation`
- Summary:
  - 新增 AV 框架层：`avCrawlerProvider`、`avCrawler`、`avDetailParser`、`avScrapeRunContext`，并实现 JavDB crawler + parser + post-process。
  - `ScrapeAVUpload` 与 `ConfirmAV` 写入 `scrape_trace`，包含搜索词、URL、步骤、错误与匹配策略，用于排障与持续优化。
  - 升级番号提取逻辑，覆盖 `FC2PPV/FC2-PPV`、`HEYZO`、`259LUXU` 等变体，并改进候选排序规则。
  - JavDB 搜索默认附带 `locale=zh`，提升中文内容命中稳定性。
  - 新增技能 `.codex/skills/av-scraper-optimization`，沉淀 mdcx→Go 架构映射与编号策略实践。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_test.go`
  - `.codex/skills/av-scraper-optimization/SKILL.md`
  - `.codex/skills/av-scraper-optimization/references/architecture-map.md`
  - `.codex/skills/av-scraper-optimization/references/number-normalization.md`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestExtractAVCodeVariants'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 09:39] 自动刮削中文一致性修复计划（电影/剧集）
- Type: `plan`
- Summary:
  - 统一上传后自动刮削的 TMDB 语言策略为 `zh-CN` 优先，并在字段缺失时回退无语言详情补全。
  - 保持演员自动关联链路不变，仅修复自动刮削元数据语言不一致问题。
  - 通过服务层单测覆盖 `ScrapeMovieUpload` 与 `ScrapeEpisodeUpload` 的语言参数与回退行为。
- Changed Files:
  - `plan.md`
- Verification:
  - 计划项，无需构建/测试。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 09:39] 实现自动刮削 zh-CN 优先与回退补全（含单测）
- Type: `implementation`
- Summary:
  - `ScrapeMovieUpload` 与 `ScrapeEpisodeUpload` 改为通过 `getTMDBJSON(..., "zh-CN")` 请求搜索与详情。
  - 新增自动刮削详情回退流程：当中文字段缺失时，自动请求无语言详情并合并补全。
  - 新增季/集级别合并逻辑，补全 `season/episode` 的标题、简介、日期等空字段。
  - 新增两条服务层单测，验证自动电影/剧集刮削的语言参数与回退行为。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestScrape(Movie|Episode)UploadUsesChineseLanguageAndFallback'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-09 09:40] Add change-management conventions
- Type: `docs`
- Summary:
  - Added repository-level conventions: rollbackable changes, mandatory git commits, and mandatory incremental updates to `plan.md`.
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - Documentation-only change; no build/test required.
- Rollback:
  - Revert the commit that introduces this entry.

### [2026-04-09 09:33] Implement upload dedupe and instant-upload flow
- Type: `implementation`
- Summary:
  - Implemented hash verification, global dedupe (`file_hashes`), `/api/v1/upload/check`, updated upload handler/service/repository, queue payload/idempotency, and temp cleanup command.
- Changed Files:
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_check.go`
  - `internal/services/upload.go`
  - `internal/repository/video_repository.go`
  - `internal/queue/tasks.go`
  - `internal/hashutil/hash.go`
  - `migrations/0004_file_hashes.up.sql`
  - `migrations/0004_file_hashes.down.sql`
  - `cmd/cleanup-temp/main.go`
  - `main.go`
  - `internal/config/config.go`
  - `.env.example`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing the upload dedupe implementation.

### [2026-04-09 11:05] Add admin movie/episode upload with async scrape pipeline
- Type: `implementation`
- Summary:
  - Added role-based upload restrictions (`user` can only upload `short`, `admin` can upload `short/movie/episode`).
  - Added asynchronous scrape tasks for `movie/episode` uploads and chained transcode enqueue after scrape.
  - Added filename parser `ParseFilename` for movie and episode patterns and updated scrape service to write TMDB metadata, poster thumbnail, and `tmdb_id`.
  - Added migration for `videos.tmdb_id` and new `scraping` status.
- Changed Files:
  - `internal/handlers/upload.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/queue/tasks.go`
  - `internal/services/scraper.go`
  - `internal/services/upload.go`
  - `internal/repository/video_repository.go`
  - `internal/utils/filename_parser.go`
  - `internal/utils/filename.go`
  - `internal/utils/filename_test.go`
  - `internal/models/models.go`
  - `main.go`
  - `migrations/0005_video_scrape_fields.up.sql`
  - `migrations/0005_video_scrape_fields.down.sql`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 11:45] Add admin scrape preview and manual confirm APIs
- Type: `implementation`
- Summary:
  - Added admin-only scrape management endpoints: preview TMDB candidates and confirm manual metadata edits.
  - Added in-memory preview cache (5 minutes), configurable poster storage path, and poster download for confirm flow.
  - Extended scraper service with preview/confirm methods for movie and TV episode; confirm updates videos metadata and episode linkage.
- Changed Files:
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/router.go`
  - `internal/services/scraper.go`
  - `internal/config/config.go`
  - `main.go`
  - `.env.example`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 12:20] Add API docs generation and Swagger UI
- Type: `implementation`
- Summary:
  - Added API docs generation command `go run ./cmd/gen-openapi` and wired `go:generate` in `main.go`.
  - Added runtime Swagger docs exposure under `/swagger/index.html` and `/swagger/openapi.json` controlled by `ENABLE_SWAGGER`.
  - Added OpenAPI base metadata and handler documentation models/comments for key endpoints.
- Changed Files:
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/index.html`
  - `docs/swagger/openapi.json`
  - `internal/config/config.go`
  - `.env.example`
  - `main.go`
  - `internal/handlers/router.go`
  - `internal/handlers/swagger_models.go`
  - `internal/handlers/auth.go`
  - `internal/handlers/upload_check.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/admin_scrape.go`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
  - `go run ./cmd/gen-openapi` generated docs files successfully.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 12:55] Add one-command run scripts
- Type: `implementation`
- Summary:
  - Added Linux/macOS one-command startup and shutdown scripts.
  - `dev-up.sh` now starts postgres/redis, waits for DB readiness, runs all migrations, and starts server/worker with logs + PID files.
  - Added run guide documentation and `.run/` git ignore.
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `docs/run.md`
  - `.gitignore`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:10] Fix duplicate migration constraint on one-click startup
- Type: `implementation`
- Summary:
  - Fixed `0002_auth.up.sql` to make FK and column changes idempotent.
  - Updated `scripts/dev-up.sh` to use `schema_migrations` and skip already-applied migrations.
  - This prevents repeated startup from failing with `constraint ... already exists`.
- Changed Files:
  - `migrations/0002_auth.up.sql`
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:25] Add explicit ENV_FILE loading for startup
- Type: `implementation`
- Summary:
  - Added `ENV_FILE` support in `main.go` (`godotenv.Overload` when provided).
  - Updated `scripts/dev-up.sh` to load and export env vars from `ENV_FILE` (defaults to project `.env`).
  - Updated run docs to show custom env file usage.
- Changed Files:
  - `main.go`
  - `scripts/dev-up.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:35] Harden .env parsing in dev-up script
- Type: `implementation`
- Summary:
  - Replaced `source .env` with safe line-by-line env parsing in `scripts/dev-up.sh`.
  - Parser now handles BOM/CRLF, ignores blank/comment lines, skips invalid lines, and supports quoted values.
  - Updated run documentation accordingly.
- Changed Files:
  - `scripts/dev-up.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
  - `bash -n scripts/dev-up.sh` could not be run in this Windows shell (`bash` not installed).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:15] Implement admin web and admin management APIs integration
- Type: `implementation`
- Summary:
  - Added backend admin management APIs for stats/videos/users/tasks/system operations and chunked upload endpoints.
  - Added SSE event stream endpoint at `/api/v1/admin/events/ws` and static hosting of built admin frontend under `/admin`.
  - Generated full `admin-web` (Vue3 + Vite + Element Plus + Pinia + axios) with login, dashboard, video/upload/scrape/user/system/task pages and API integration.
  - Added module-level `admin-web/AGENTS.md` and included newly added skill assets/directories in workspace scope as requested.
- Changed Files:
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/admin_events.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/models/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `main.go`
  - `admin-web/*`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `npm --prefix admin-web install` succeeded.
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:27] Extend one-command scripts to include admin frontend
- Type: `implementation`
- Summary:
  - Extended `scripts/dev-up.sh` to support `--frontend dev|build|off` with default `dev`.
  - Added auto-install of frontend dependencies when `admin-web/node_modules` is missing.
  - Added frontend dev process management (`.run/frontend.pid`, `.run/frontend.log`) and build-mode output guidance.
  - Updated `scripts/dev-down.sh` to stop frontend process and refreshed run documentation.
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
  - `bash -n scripts/dev-down.sh` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:41] Polish admin-web UI with unified design system
- Type: `implementation`
- Summary:
  - Applied `ui-ux-pro-max` design direction to admin frontend with a unified visual system (fonts, colors, cards, tables, transitions, responsive behavior).
  - Refined shared layout shell (brand sidebar, icon navigation, sticky top bar) and upgraded all key views for stronger hierarchy and readability.
  - Enhanced chart styling, upload progress presentation, scrape candidate card interactions, and system log panel legibility.
- Changed Files:
  - `admin-web/src/assets/theme.css`
  - `admin-web/src/main.js`
  - `admin-web/src/App.vue`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/components/UploadProgress.vue`
  - `admin-web/src/views/Login.vue`
  - `admin-web/src/views/Dashboard.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/UserManage.vue`
  - `admin-web/src/views/SystemSettings.vue`
  - `admin-web/src/views/TaskMonitor.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:56] Add tag input to admin video upload form
- Type: `implementation`
- Summary:
  - Added missing upload form control for video tags using multi-select with create mode.
  - Added preset tag options and kept support for custom tag creation.
  - Added submit-time tag normalization (trim, empty filtering, lowercase dedupe) before calling `/upload/init`.
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:16] Fix dev-mode LAN access and Vite API proxy 404
- Type: `implementation`
- Summary:
  - Updated Vite dev server to listen on `0.0.0.0:5173` with strict port behavior for LAN access.
  - Switched proxy matching to `/api/v1` and made proxy backend target configurable via `VITE_API_PROXY_TARGET`.
  - Updated frontend and run documentation with LAN usage and `/api/v1/*` 404 troubleshooting steps.
- Changed Files:
  - `admin-web/vite.config.js`
  - `admin-web/.env.development`
  - `admin-web/README.md`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:28] Fix backend admin SPA route conflict causing startup panic
- Type: `implementation`
- Summary:
  - Fixed Gin route conflict between `/admin/assets/*filepath` and `/admin/*path` by replacing wildcard route with `NoRoute` fallback logic for `/admin/`.
  - Preserved static asset serving for `/admin/assets/*` and SPA index fallback for non-asset `/admin/*` paths.
  - This removes startup panic and restores normal admin API availability when server runs with latest code.
- Changed Files:
  - `internal/handlers/router.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:33] Harden dev-up startup checks to catch false-positive server start
- Type: `implementation`
- Summary:
  - Added background process startup helper in `scripts/dev-up.sh` with post-start liveness checks.
  - If `server`/`worker`/`frontend` exits immediately (for example, port conflict on `:8080`), script now fails fast and prints recent log tail.
  - Prevents misleading "starting server" success when actual process has already crashed.
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:47] Fix admin video detail NULL scan failure
- Type: `implementation`
- Summary:
  - Fixed `AdminVideoDetail` query scanning failures when nullable video columns contain `NULL`.
  - Added SQL-side `COALESCE` defaults for nullable detail fields (`duration_seconds`, `width`, `height`, path fields, metadata, and text fields) before scanning into Go struct fields.
  - Prevents `cannot scan NULL into *int` errors on video detail retrieval.
- Changed Files:
  - `internal/repository/admin_repository.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 13:52] 新增 Markdown 与前端文案中文约束
- Type: `implementation`
- Summary:
  - 在仓库全局规则中新增约束：所有 Markdown 文件及前端面向用户的界面文案默认使用中文。
  - 仅当任务明确要求其他语言时，允许例外。
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - 文档变更，无需构建或测试。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:15] 修复转码回退与失败源文件保留策略（计划）
- Type: `plan`
- Summary:
  - 检查上传后转码触发链路，确认 `short` 直转码、`movie/episode` 抓取后转码。
  - 调整转码元数据提取：`ffprobe` 失败或缺字段时写入默认值并保留可编辑元数据，不让任务因探测失败中断。
  - 调整临时清理保护条件，避免 `failed`/`scraping` 视频源文件被误删，保留给管理员决策。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/repository/video_repository.go`
- Verification:
  - 待执行：`go test ./...`
- Rollback:
  - Revert the commit containing this entry.
### [2026-04-10 14:20] 修复转码元数据回退与失败源文件保留
- Type: `implementation`
- Summary:
  - 保持现有上传转码链路：`short` 上传后直接转码，`movie/episode` 抓取后继续转码。
  - 调整 `TranscodeService`：`ffprobe` 失败不再导致转码任务失败，改为写入默认值（`duration/width/height=0`, `codec=unknown`）并记录 `probe_error`，便于管理员后续手动改写。
  - 调整临时文件清理保护查询，新增保护 `scraping` 与 `failed` 状态视频的 `original_path`，避免自动清理误删失败源文件。
  - 新增单测覆盖元数据回退与正常探测场景。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `go test ./...` failed（现有跨平台编译问题：`internal/handlers/admin.go` 使用 `syscall.Statfs_t/Statfs` 在当前 Windows 环境不可用）。
  - `go test ./internal/services ./internal/repository ./internal/queue` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 14:23] 强化中文无乱码约束并修复计划文档乱码
- Type: `implementation`
- Summary:
  - 更新仓库级 `AGENTS.md`：新增“中文内容禁止乱码”与“提交信息使用中文且无乱码”的强约束。
  - 修复 `plan.md` 顶部说明文字乱码。
  - 修复 `plan.md` 中 `[2026-04-10 13:52]` 条目的中文乱码内容。
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - 文档变更，人工检查文本显示正常，无乱码。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:28] 新增可播放视频源接口（计划）
- Type: `plan`
- Summary:
  - 新增鉴权播放源接口 `GET /api/v1/videos/:id/source`，用于直接返回可播放视频流。
  - 限定仅 `ready` 状态可播放，未就绪返回 `409`，不存在返回 `404`，未登录返回 `401`。
  - 在视频详情接口中新增 `play_url`（相对路径），减少客户端拼接逻辑。
  - 更新 OpenAPI 生成配置并补充最小单测。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_detail.go`
  - `internal/models/app.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
- Verification:
  - 待执行：`go test ./...`
- Rollback:
  - Revert the commit containing this entry.

### [2026-04-10 14:35] 实现可播放视频源接口与详情 play_url
- Type: `implementation`
- Summary:
  - 新增 `GET /api/v1/videos/:id/source`（鉴权）并返回视频流；仅 `ready` 可播放。
  - 新增错误返回策略：`401`（未登录）、`400`（非法 ID）、`404`（视频或源文件不存在）、`409`（视频未就绪）。
  - 视频详情响应新增 `play_url` 字段，固定为相对路径 `/api/v1/videos/{id}/source`。
  - 新增 `internal/utils/VideoPlayURL` 与对应单测，确保播放地址生成稳定。
  - 更新 `cmd/gen-openapi` 并重新生成 `docs/swagger/openapi.json`，加入新接口说明。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_detail.go`
  - `internal/models/app.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/services ./internal/repository ./internal/queue ./pkg/ffmpeg ./cmd/gen-openapi` passed。
  - `go test ./...` failed（既有 Windows 编译问题：`internal/handlers/admin.go` 中 `syscall.Statfs_t/Statfs` 不可用）。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:48] 管理端新增签名播放预览功能（计划）
- Type: `plan`
- Summary:
  - 为 admin-web 视频详情新增可播放预览能力，解决 `<video>` 无法携带 Bearer 头问题。
  - 后端新增管理员获取临时签名播放地址接口，并新增签名流播放接口。
  - 签名默认有效期 10 分钟，签名口径绑定 `video_id + exp`。
  - 同步更新 OpenAPI 与前端页面文案，确保中文无乱码。
- Changed Files:
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/video_source.go`
  - `internal/utils/play_url_sign.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoList.vue`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
- Verification:
  - 待执行：`go test ./...`
  - 待执行：`npm --prefix admin-web run build`
- Rollback:
  - Revert the commit containing this entry.

### [2026-04-10 14:59] 实现管理端签名播放预览与临时播放链接
- Type: `implementation`
- Summary:
  - 新增配置项 `PLAY_URL_SIGN_SECRET`（为空时回退 `JWT_SECRET`），用于签名播放地址。
  - 新增管理员接口 `GET /api/v1/admin/videos/:id/play-url`，返回 `signed_url` 与 `expires_at`。
  - 新增签名播放接口 `GET /api/v1/videos/:id/source/signed`，支持签名校验、过期校验与 `ready` 状态校验。
  - 保留原接口 `GET /api/v1/videos/:id/source`（Bearer 鉴权）不变，并复用统一可播放源检查逻辑。
  - 新增签名工具与单测：`SignVideoSource` / `VerifyVideoSourceSign`。
  - 管理端 `VideoList` 详情弹窗新增播放器预览、刷新播放链接按钮、到期时间展示与非 ready 状态提示。
  - 修复 `VideoList` 页面中文文案乱码，确保界面文案为中文且无乱码。
  - 更新 OpenAPI 生成器并重新生成 `docs/swagger/openapi.json`，补充两个新接口文档。
- Changed Files:
  - `.env.example`
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/video_source.go`
  - `internal/utils/play_url_sign.go`
  - `internal/utils/play_url_sign_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoList.vue`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
  - `main.go`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/services ./internal/repository ./internal/queue ./pkg/ffmpeg ./cmd/gen-openapi` passed。
  - `go test ./...` failed（既有 Windows 编译问题：`internal/handlers/admin.go` 中 `syscall.Statfs_t/Statfs` 不可用）。
  - `npm --prefix admin-web run build` failed（当前环境 Node 权限问题：`EPERM: operation not permitted, lstat 'C:\Users\CheeTsui'`）。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 17:44] 修复视频删除时 NULL 扫描导致失败
- Type: `implementation`
- Summary:
  - 修复 `AdminDeleteVideo` 依赖的 `GetVideoByID` 查询在历史数据存在 NULL 时扫描失败的问题（`duration_seconds/width/height` 等字段）。
  - 在 `GetVideoByID` 与 `GetVideoByOriginalPath` 中统一使用 SQL `COALESCE` 提供默认值，避免 `cannot scan NULL into *int/*string`。
  - 删除逻辑保持容错：即使磁盘文件不存在也不会阻断删除（`os.Remove` 错误已忽略）。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `go test ./internal/repository ./internal/services ./internal/queue ./internal/utils` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 17:53] 修复删除视频时 NULL 数值字段扫描报错
- Type: `implementation`
- Summary:
  - 修复 `GetVideoByID`/`GetVideoByOriginalPath` 在 `duration_seconds/width/height` 为 `NULL` 时扫描到 `int` 触发报错的问题。
  - 将上述字段改为先扫描 `sql.NullInt32`，再统一转换为 `int`，避免 `cannot scan NULL into *int`。
  - 新增回归测试覆盖 `NULL -> 0` 和有效值转换行为。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_repository_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 18:41] 修复分片上传后 worker 转码找不到源文件
- Type: `implementation`
- Summary:
  - 定位到 `UploadComplete` 会在入队后执行 `chunkUpload.Abort`，导致会话目录被删除，连同 `assembled-*.mp4` 一起被删。
  - 调整 `ChunkUploadService.Complete`：合并文件改为写入 `UPLOAD_TEMP_DIR/assembled/`，不再位于 `chunk-sessions/<session_id>/` 下。
  - 保持 `Abort(session_id)` 只清理分片会话目录，避免误删待转码输入文件。
  - 新增回归测试，覆盖“Complete 后 Abort，合并文件仍可读取”行为。
- Changed Files:
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run TestChunkUploadCompleteFileSurvivesAbort -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 19:40] 修复删除转码失败视频的外键约束报错
- Type: `implementation`
- Summary:
  - 修复删除视频时因 `transcoding_jobs.video_id` 与 `user_video_actions.video_id` 外键依赖导致的 `SQLSTATE 23503`。
  - 将 `DeleteVideoByID` 改为事务删除：先删 `transcoding_jobs`、再删 `user_video_actions`、最后删 `videos`。
  - 提取 `deleteVideoDependencies` 并补充顺序与错误分支回归测试。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_delete_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run TestDeleteVideoDependencies -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 21:28] 修复删除视频与转码任务并发导致的外键竞态
- Type: `implementation`
- Summary:
  - 针对“仍偶发 `transcoding_jobs_video_id_fkey`”问题，补充并发安全：删除事务先对目标视频行执行 `FOR UPDATE` 锁定。
  - 在锁定后再顺序删除 `transcoding_jobs`、`user_video_actions`、`videos`，避免 worker 并发插入同 `video_id` 任务造成竞态。
  - 更新回归测试，覆盖加锁后的 SQL 执行顺序与错误分支。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_delete_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run TestDeleteVideoDependencies -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 21:37] 通过外键级联删除消除视频删除阻塞
- Type: `implementation`
- Summary:
  - 新增迁移将 `transcoding_jobs.video_id` 与 `user_video_actions.video_id` 外键改为 `ON DELETE CASCADE`。
  - 解决删除视频时依赖记录并发/残留导致的 `SQLSTATE 23503` 阻塞问题。
  - 已在当前数据库执行迁移并核验 `confdeltype='c'`（cascade）。
- Changed Files:
  - `migrations/0006_video_fk_cascade.up.sql`
  - `migrations/0006_video_fk_cascade.down.sql`
  - `plan.md`
- Verification:
  - `docker exec -i video_server_postgres psql -U video -d video_server -v ON_ERROR_STOP=1 < migrations/0006_video_fk_cascade.up.sql` applied successfully。
  - `docker exec -i video_server_postgres psql -U video -d video_server -tA -c "SELECT conname, confdeltype FROM pg_constraint ..."` returned `c` for both constraints。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry, then执行 `migrations/0006_video_fk_cascade.down.sql`。

### [2026-04-16 12:26] 优化开发脚本进程管理与前端依赖安装策略
- Type: `implementation`
- Summary:
  - 优化 `dev-up/dev-down` 的 PID 管理：PID 文件升级为“PID+命令指纹”，避免 PID 复用导致误判或误杀。
  - `dev-down` 新增优雅停止等待与超时强制终止，避免仅发信号后立即返回导致残留进程占端口。
  - `dev-up` 新增启动失败自动清理机制：若后续步骤失败，会回收本次新启动的后台进程，减少半启动状态。
  - 前端依赖安装改为基于 `package-lock.json` 哈希判断，锁文件变化或缺依赖时执行 `npm ci`，提升环境一致性。
  - 迁移执行移除无必要管道（`cat | docker exec`），改为输入重定向，简化执行链路。
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh scripts/dev-down.sh` passed.
  - `bash scripts/dev-up.sh --help` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-16 13:52] 刮削管理补全详情展示与中文化返回，并增加加载/报错反馈
- Type: `implementation`
- Summary:
  - 后端 `ScraperService` 预览与确认流程接入 `language=zh-CN` 请求参数，优先返回中文字段。
  - 新增“中文缺字段时英文兜底”的合并逻辑，覆盖标题/简介/日期/类型/题材及部分嵌套字段，避免空白展示。
  - 新增 `scraper` 单测，验证预览接口会携带中文语言参数，并在中文缺失时执行英文兜底。
  - 重构 `admin-web` 刮削管理页：从仅概览升级为“候选列表 + 结构化详情 + 原始 metadata JSON 折叠查看”。
  - 前端新增查询与保存 `loading` 状态，以及失败场景的明确错误提示（优先显示后端返回消息）。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `admin-web/src/views/ScrapePreview.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreview(Movie|TV)' -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -count=1` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-16 18:52] 修复 dev-down 未能关闭 server/worker/frontend 残留进程
- Type: `implementation`
- Summary:
  - 定位根因：`dev-down` 对 PID 文件命令串做严格匹配，`go run` 启动后命令会漂移为临时 `main -mode ...`，导致误判并跳过停止。
  - 强化 `dev-down` 停止逻辑：新增进程树终止（先 `TERM` 后超时 `KILL`）、服务命令漂移容错匹配、无 PID 文件时的兜底清理（按模式/端口/worker 日志 PID）。
  - 同步增强 `dev-up` 启动失败清理逻辑，复用相同的进程树停止和命令漂移容错，避免半启动残留。
  - 修复脚本兼容性：去除 `mapfile`，改为 `bash 3.2` 兼容写法（macOS 默认 bash 可运行）。
- Changed Files:
  - `scripts/dev-down.sh`
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh scripts/dev-down.sh` passed.
  - 执行 `bash scripts/dev-down.sh` 后，`lsof -nP -iTCP:8080 -sTCP:LISTEN` 无输出。
  - 执行 `bash scripts/dev-down.sh` 后，`lsof -nP -iTCP:5173 -sTCP:LISTEN` 无输出。
  - 通过模拟命令漂移场景（PID 文件为 `go run ...`，进程命令变化）执行 `dev-down`，进程可被清理。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 19:45] 实现演员管理与上传/刮削演员关联（管理端全中文）
- Type: `implementation`
- Summary:
  - 新增数据库迁移 `0007_actor_support`，引入 `actors` 与 `video_actors`，支持演员主数据与视频多演员关联。
  - 后端新增演员仓储能力：演员列表/创建/更新、按姓名自动入库、视频演员替换与追加关联。
  - 扩展管理端视频详情：返回 `actors` 列表；视频更新接口支持 `actor_ids` + `actor_names`（可同时传已有演员与新建演员）。
  - 新增管理端演员接口：`GET /api/v1/admin/actors`、`POST /api/v1/admin/actors`、`PUT /api/v1/admin/actors/:id`。
  - 上传链路支持演员关联：`upload/init` 会话保存演员信息，`upload/complete` 落库并关联演员；普通上传同样支持 `actor_ids/actor_names`。
  - 刮削链路接入 TMDB cast：电影/分集刮削与确认后会自动补充演员关联（非阻断式，失败不影响主流程）。
  - 管理端前端新增“演员管理”页面，并在上传页、视频详情页接入演员搜索多选+可创建；本次新增/修改文案统一中文。
- Changed Files:
  - `migrations/0007_actor_support.up.sql`
  - `migrations/0007_actor_support.down.sql`
  - `internal/models/admin.go`
  - `internal/repository/actor_repository.go`
  - `internal/repository/actor_repository_test.go`
  - `internal/repository/admin_repository.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/handlers/upload_helpers_test.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/router/index.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/views/ActorManage.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestExtractCastNames|^$' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
  - 说明：`internal/services` 全量测试在当前沙箱环境存在 `httptest` 端口监听限制，故本次使用定向用例与编译校验。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 20:19] 新增演员支持按姓名刮削（TMDB + JavDB）并增强重名处理
- Type: `implementation`
- Summary:
  - 新增管理端演员刮削预览接口 `POST /api/v1/admin/actors/scrape/preview`，支持 `tmdb/javdb` 双来源按姓名查询。
  - `ScraperService` 新增演员刮削能力：TMDB 人物检索（`zh-CN` 优先 + 英文兜底）与 JavDB HTML 候选解析；统一返回候选结构。
  - 新增 AV 刮削配置项：`AV_SCRAPER_BASE_URL`、`AV_SCRAPER_TIMEOUT_SECONDS`、`AV_SCRAPER_USER_AGENT`，并在 server/worker 启动时注入。
  - 演员创建重名时返回已存在演员信息（`existing_actor_id / existing_actor / existing_actor_name`），用于前端直接跳转编辑。
  - 管理端“演员管理”弹窗新增“按姓名刮削”交互：来源切换、候选列表、loading、中文错误提示、一键回填字段；重名创建支持提示并跳转编辑。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_actor.go`
  - `internal/services/scraper_test.go`
  - `internal/handlers/admin_actor_scrape.go`
  - `internal/handlers/admin_actor_scrape_test.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/swagger_models.go`
  - `internal/repository/actor_repository.go`
  - `internal/config/config.go`
  - `main.go`
  - `.env.example`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/request.js`
  - `admin-web/src/views/ActorManage.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreviewActorByName(TMDB|JavDB)|TestPreview(Movie|TV)UsesChineseLanguageAnd(EnglishFallback|Fallback)|TestExtractCastNames' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestAdminActorScrapePreview' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run 'TestNormalizeActorName' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 21:20] 修复演员新增误报“同名已存在”
- Type: `implementation`
- Summary:
  - 定位误报根因：前端把 `code=1025` 一律视为重名；后端创建演员的非重名错误也会返回 `1025`，导致误判。
  - 后端修复：`AdminCreateActor` 仅在唯一约束冲突时返回 `1025`，并附带 `reason=duplicate_name`；其他创建失败改为 `1028`。
  - 前端修复：仅在明确重名（`reason=duplicate_name` 或 `code=1025` 且消息为“演员名称已存在”）时才弹“去编辑已有演员”提示。
- Changed Files:
  - `internal/handlers/admin.go`
  - `admin-web/src/views/ActorManage.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestAdminActorScrapePreview' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 21:43] 修复 dev-up 仅执行首条迁移导致 actors 表缺失
- Type: `implementation`
- Summary:
  - 线上报错 `create actor: relation "actors" does not exist (SQLSTATE 42P01)`，定位为数据库仅执行了 `0001_init.up.sql`。
  - 根因是 `scripts/dev-up.sh` 在迁移遍历中依赖 `find ... -print0 | sort -z`，导致在当前环境迁移流异常，出现只处理首条文件的问题。
  - 将迁移遍历改为 Bash 通配循环 `for file in "$ROOT_DIR"/migrations/*.up.sql`，按文件名顺序稳定执行并兼容当前环境。
  - 已补齐当前数据库缺失迁移 `0002~0007`，确认 `actors` 与 `video_actors` 表存在。
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed。
  - `bash scripts/dev-up.sh` 日志可见按顺序检查 `0001~0007` 迁移。
  - `SELECT version FROM schema_migrations ORDER BY version;` 返回 `0001~0007` 全部版本。
  - `SELECT to_regclass('public.actors'), to_regclass('public.video_actors');` 返回 `actors|video_actors`。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 22:09] 修复 TMDB 演员刮削 notes 被截断与乱码
- Type: `implementation`
- Summary:
  - 定位到演员刮削返回中 `notes` 不完整的根因：`previewActorsTMDB` 对 `biography` 执行了 `500` 字节硬截断。
  - 该截断会导致中文多字节字符被切断，出现 `\uFFFD` 替代字符，并且内容不完整。
  - 移除该截断逻辑，`notes` 直接返回完整 `biography`。
  - 新增回归测试：长中文简介场景下 `notes` 不被截断且不包含 `\uFFFD`。
- Changed Files:
  - `internal/services/scraper_actor.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreviewActorByNameTMDBNotesNotTruncated|TestPreviewActorByName(TMDB|JavDB)' -count=1` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 23:01] 上传中心支持“非电影批量上传”，电影保持单文件
- Type: `implementation`
- Summary:
  - 管理端上传页改造为队列式上传：`short/episode` 支持批量选择并串行分片上传，`movie` 继续仅支持单文件。
  - 类型切换为电影时，若已选择多个文件会自动仅保留第 1 个并提示，避免误提交。
  - 批量流程支持“失败不中断”：单文件失败后继续后续文件，最终展示成功/失败/取消汇总与逐文件结果。
  - 增加批次进度与当前文件进度展示，取消上传时终止当前会话并标记后续文件为未上传（已取消）。
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 00:39] 新增短视频合集功能（上传可选多选、后续可修改、删除仅解关联）
- Type: `implementation`
- Summary:
  - 新增数据库迁移 `0008_collections_support`，引入 `collections` 与 `video_collections`，支持视频与合集多对多关联。
  - 新增合集仓储能力：合集列表/创建/更新/删除、名称规范化唯一约束、视频合集关联替换与按视频批量查询。
  - 约束合集仅适用于 `short`：上传（普通/分片）与管理端视频编辑均支持 `collection_ids`，并在非短视频场景返回明确错误。
  - 管理端新增合集管理页面与路由（增删改查）；删除合集时仅删除合集并自动解除关联，不删除视频，并返回解除关联数量。
  - 管理端上传页新增“所属合集（可选、多选，仅短视频）”；视频详情页支持后续修改合集并保存。
  - 应用侧视频返回补充合集信息：详情、推荐与列表搜索结果可携带 `collections` 字段。
  - 补充合集相关单元测试：名称规范化、ID 去重、上传参数解析与类型约束校验。
- Changed Files:
  - `migrations/0008_collections_support.up.sql`
  - `migrations/0008_collections_support.down.sql`
  - `internal/models/models.go`
  - `internal/models/app.go`
  - `internal/models/admin.go`
  - `internal/repository/collection_repository.go`
  - `internal/repository/collection_repository_test.go`
  - `internal/repository/video_repository.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/admin_repository.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/router.go`
  - `internal/handlers/upload_collection_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/router/index.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/views/CollectionManage.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` failed（当前沙箱环境禁止 `httptest` 监听端口，`internal/handlers` 与 `internal/services` 的部分用例受限）。
  - `GOCACHE=$(pwd)/.gocache go test ./... -run '^$'` passed（全量编译校验通过）。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository ./internal/handlers ./internal/services -run 'TestNormalizeCollectionName|TestDedupeCollectionIDs|TestParseUploadCollectionIDs|TestCollectionTypeValidation|TestChunkUploadCompleteFileSurvivesAbort' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 00:49] 修正上传页合集远程搜索的选项保留
- Type: `implementation`
- Summary:
  - 优化管理端上传页 `searchCollections`：改为合并选项而非覆盖，避免远程搜索后已选合集标签丢失显示。
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 11:54] 新增 AV 视频分类与自动 AV 刮削（含演员自动关联）
- Type: `implementation`
- Summary:
  - 新增视频类型 `av`（管理员可上传），上传后与电影/剧集一样进入 `scraping` 流程，由 worker 自动执行 AV 刮削。
  - 扩展队列任务：新增 `video:scrape:av`，并在上传与分片上传完成后按类型路由到 AV 刮削任务。
  - 新增 AV 刮削能力（JavDB）：支持“番号优先、标题兜底”检索，抓取标题/简介/封面/发行日期/演员，写入视频元数据并自动关联演员（`scrape_av`）。
  - 扩展管理端刮削接口：`/admin/scrape/preview` 支持 `type=av`，`/admin/scrape/confirm` 对 AV 支持 `external_id` 确认；电影/剧集仍使用 `tmdb_id`。
  - 新增数据库迁移 `0009_av_type_support`，将 `videos.type` 约束扩展为 `short/movie/episode/av`。
  - 管理端界面新增 AV：上传类型、视频列表筛选、刮削页类型选择与 AV 候选保存；仪表盘新增 AV 数量统计卡片。
  - 后端统计扩展：`AdminStats` 新增 `av_videos`，`CountVideosByType` 统计新增 `av`。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `internal/queue/tasks.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/services/upload.go`
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/admin.go`
  - `internal/models/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/handlers/swagger_models.go`
  - `internal/handlers/upload_collection_test.go`
  - `migrations/0009_av_type_support.up.sql`
  - `migrations/0009_av_type_support.down.sql`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/Dashboard.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestPreviewAVFallbackByTitle|TestScrape(Movie|Episode)UploadUsesChineseLanguageAndFallback' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestCollectionTypeValidation|TestParseUploadCollectionIDs|TestParseUploadTags_JSON' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./... -run '^$'` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 11:58] AV 功能补充全量验证
- Type: `implementation`
- Summary:
  - 在完成 AV 分类与自动刮削改造后，补充执行后端全量单测回归，确认不影响既有模块。
- Changed Files:
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 16:33] 图片上传管理功能落地（批量上传/演员合集关联/缩放查看/WebP 转换后删除源文件）
- Type: `implementation`
- Summary:
  - 补齐后台图片管理后端接口：批量上传、列表筛选、详情编辑、删除、图片查看缩放接口，以及图片合集 CRUD。
  - 图片上传链路强化：非 GIF 上传统一转为 WebP 存储，并在转换成功后立即删除源文件；若删除源文件失败则回滚已生成的 WebP，避免脏数据残留。
  - 管理端新增“图片管理”和“图片合集”页面，支持批量上传、演员关联、合集关联、详情编辑、缩放参数预览。
  - 将图片服务注入 API 并注册 admin 路由，确保服务端与管理端链路完整可用。
- Changed Files:
  - `internal/handlers/admin_image.go`
  - `internal/handlers/router.go`
  - `internal/services/image.go`
  - `internal/repository/image_repository.go`
  - `internal/repository/image_collection_repository.go`
  - `internal/repository/actor_repository.go`
  - `internal/models/models.go`
  - `internal/models/admin.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `migrations/0010_image_management.up.sql`
  - `migrations/0010_image_management.down.sql`
  - `main.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/router/index.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/views/ImageManage.vue`
  - `admin-web/src/views/ImageCollectionManage.vue`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 16:41] 图片上传补齐秒传能力（对齐视频上传预检）
- Type: `implementation`
- Summary:
  - 新增后台图片秒传预检接口 `POST /api/v1/admin/images/check`，按 `hash + file_size` 查询是否已存在图片。
  - 管理端图片批量上传流程改为“先算 SHA-256 并逐个预检”：命中时直接返回“秒传命中”，未命中再走实际上传。
  - 上传结果汇总支持混合展示“秒传命中”和普通上传结果，确保批量上传在有重复文件时不会重复传输。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/admin_image.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/ImageManage.vue`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 16:46] 图片上传改为弹窗入口并优化列表界面视觉
- Type: `implementation`
- Summary:
  - 图片管理页将“批量上传”从页面内卡片改为独立弹窗，避免列表页被上传表单占用空间。
  - 在图片列表卡片头新增“新增图片”按钮，点击后弹出上传窗口。
  - 对图片管理页进行视觉优化：新增统计条、列表头信息区与按钮布局，提升信息层级与可读性。
- Changed Files:
  - `admin-web/src/views/ImageManage.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 17:11] 视频预览扩展：按当前播放时间截帧并替换封面
- Type: `implementation`
- Summary:
  - 新增管理员接口 `POST /api/v1/admin/videos/:id/thumbnail/capture`，支持传入 `time_seconds`，由 ffmpeg 按时间截取视频帧并更新为封面。
  - 截帧逻辑优先使用转码文件路径（空时回退原始文件），并支持封面文件原子替换，避免写入半成品。
  - 管理端视频详情页在“播放预览”区域新增“设为封面（当前帧）”按钮，读取播放器当前时间并调用新接口，成功后刷新当前详情封面路径。
- Changed Files:
  - `internal/handlers/admin_video_thumbnail.go`
  - `internal/handlers/router.go`
  - `internal/repository/admin_repository.go`
  - `pkg/ffmpeg/ffmpeg.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 17:26] 修复视频截帧封面失败（临时文件扩展名导致 ffmpeg 无法识别输出格式）
- Type: `implementation`
- Summary:
  - 修复 `AdminCaptureVideoThumbnail` 临时文件命名：改为保留目标封面的扩展名（例如 `.jpg`），避免生成 `thumb.jpg.tmp-uuid` 这种 ffmpeg 无法识别格式的路径。
  - 新增 `buildCaptureTempPath`，统一生成“同目录 + 隐藏临时名 + 正确扩展名”的临时文件路径。
  - 新增处理器层单测覆盖：验证临时文件路径扩展名保留与无扩展时回退 `.jpg`。
- Changed Files:
  - `internal/handlers/admin_video_thumbnail.go`
  - `internal/handlers/admin_video_thumbnail_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/handlers -run 'TestBuildCaptureTempPath' -count=1` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 22:17] Android AV 暗黑风格延续 + 横屏全屏播放
- Type: `implementation`
- Summary:
  - AV 分类列表延续暗黑视觉风格，卡片背景、文字与页面底色统一为深色主题，提升 AV 分区一致性。
  - 首页到详情导航补充 `videoType` 参数透传，详情页可基于类型切换 AV 暗黑展示。
  - AV 详情页新增“横屏全屏播放”按钮，点击后进入独立全屏播放器页面。
  - 新增全屏播放器页面：进入即锁定横屏、隐藏系统栏，返回后恢复原方向与系统栏状态。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/FullscreenVideoPlayerScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 22:52] Android 详情播放器重构：海报内播放 + 同页横屏全屏/退出
- Type: `implementation`
- Summary:
  - 将详情页播放流程改为“海报承载播放入口”：默认显示海报，点击海报中心播放按钮后原位切换为视频播放器。
  - 全屏能力改为同页状态切换，不再跳转独立全屏页面：播放器内提供“全屏”按钮，全屏后提供“退出全屏”按钮。
  - 全屏状态下锁定横屏并隐藏系统栏，退出全屏后恢复原方向与系统栏；系统返回键在全屏态优先执行“退出全屏”。
  - 保留 AV 暗黑视觉、海报回退逻辑与播放鉴权头，避免破坏现有播放链路。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/FullscreenVideoPlayerScreen.kt` (deleted)
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 23:30] 修复 AV 详情页顶部多余空白（去除外层 Scaffold 顶部 inset 叠加）
- Type: `implementation`
- Summary:
  - 定位到详情页顶部空白由嵌套 `Scaffold` 的系统窗口 inset 叠加导致：外层导航容器已添加顶部 inset，详情页内层继续处理后出现额外留白。
  - 在 `VideoHomeApp` 的认证后主 `Scaffold` 上将 `contentWindowInsets` 显式设置为 `WindowInsets(0, 0, 0, 0)`，由各页面自行控制系统栏适配。
  - 修复后 AV 详情页顶部不再额外空出一块，同时不影响现有详情页全屏/退出全屏逻辑。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-18 23:40] 调整 AV 详情页状态栏安全区（保留顶部留白，不遮挡系统状态栏）
- Type: `implementation`
- Summary:
  - 按视觉预期修正详情页顶部布局：`TopAppBar` 显式使用 `statusBarsPadding()`，确保内容不覆盖系统状态栏区域。
  - 为避免状态栏 inset 重复，`TopAppBar` 的 `windowInsets` 改为 `WindowInsets(0, 0, 0, 0)`，由 `statusBarsPadding()` 单独控制顶部留白。
  - 保持此前“同页全屏”逻辑不变：仅非全屏详情页保留状态栏留白，全屏时继续沉浸式。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 13:55] 视频上传新增图片图集单关联 + 热门标签 Top5 + 上传记录清空
- Type: `implementation`
- Summary:
  - 数据库新增 `videos.image_collection_id`（外键到 `collections_images`，删除图片图集时自动置空），用于“视频仅关联一个图片图集”的能力。
  - 后端上传链路（普通上传/分片上传）新增 `image_collection_id` 参数并落库；分片会话元数据新增该字段，完成合并后同样写入。
  - 后台视频详情编辑新增图片图集单选编辑能力，可修改或清空；视频详情接口返回 `image_collection_id` 与关联图片图集对象。
  - 新增管理端热门标签接口 `GET /api/v1/admin/video-tags/popular`，按使用次数统计并返回 TopN（默认 5，最大 20）。
  - 管理端上传页改造：标签推荐改为调用热门标签接口；新增图片图集单选；新增“清空上传记录”按钮，仅清理前端本地上传进度与结果，不影响后端文件和数据库。
  - 保留原有视频合集能力不变：`collection_ids` 仍仅用于短视频多选合集；新增图片图集能力与其并存。
  - 新增仓储层单元测试覆盖：图片图集单选归一化约束、热门标签排序与 limit 逻辑。
- Changed Files:
  - `migrations/0012_video_image_collection.up.sql`
  - `migrations/0012_video_image_collection.down.sql`
  - `internal/models/models.go`
  - `internal/models/admin.go`
  - `internal/repository/video_repository.go`
  - `internal/repository/admin_repository.go`
  - `internal/repository/image_collection_repository.go`
  - `internal/repository/image_collection_repository_test.go`
  - `internal/repository/video_repository_tags_test.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/router.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run 'TestNormalizeSingleImageCollectionID|TestNormalizePopularVideoTagsLimitAndSort'` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestParseUpload'` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 17:33] Android 短视频新增底部微型进度条（贴 Tab + 拖动 seek + 时分秒浮框）
- Type: `implementation`
- Summary:
  - 在短视频流页面新增底部常驻微型进度条，位置贴近底部 Tab，尽量减少对画面的遮挡。
  - 进度条接入当前激活视频的播放状态（时长/进度），并在详情弹层打开时自动隐藏，避免视觉冲突。
  - 支持在进度条上水平拖动调整播放进度，拖动期间进度条高度动画放大，松手后执行 seek。
  - 拖动中新增时间浮框，统一显示为 `HH:MM:SS / HH:MM:SS` 格式，提升定位精度。
  - 新增短视频时间格式化单元测试，覆盖 0 秒、分钟边界与小时边界。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedTimeFormatTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests com.chee.videos.feature.shorts.ShortFeedTimeFormatTest` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 19:02] 短视频详情美化 + 标签/合集可点击进入瀑布流并连续上下滑浏览
- Type: `implementation`
- Summary:
  - 新增后端精确筛选接口 `GET /api/v1/short/discover`（鉴权），支持 `mode=tag|collection`，按标签或合集精确匹配短视频并返回分页列表。
  - 扩展 App 服务与仓储：新增短视频发现查询链路，限制为 `ready + short`，按创建时间倒序，返回总数用于前端触底加载。
  - Android 数据层新增短视频发现 API 适配（`ApiService` / `UrlBuilder` / `VideoRepository`）。
  - 新增 Android `ShortDiscover` 模块：详情点击标签/合集进入瀑布流页，封面双列瀑布流展示，触底分页加载。
  - 瀑布流点击封面进入全屏连续浏览覆盖层：上下滑切换下一个视频，临近底部自动继续补货，可连续浏览到全部结果。
  - 美化短视频详情弹层（暗黑风格）：信息分区重排，统计卡片化，标签/合集芯片改为可点击并联动导航。
  - 首页路由补充短视频发现页面导航参数透传（`mode/value/title`）。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/short_discover.go`
  - `internal/services/app.go`
  - `internal/repository/app_repository.go`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 19:34] 瀑布流与“我的”短视频播放器新增底部进度条（安全区内）
- Type: `implementation`
- Summary:
  - 新增可复用组件 `ShortVideoBottomProgressBar`，保持短视频进度条样式与拖动交互一致（拖动放大 + 时分秒浮层 + 拖动 seek）。
  - 在“我的”入口对应的 `UnifiedPlayerScreen` 短视频模式接入进度状态同步与进度条显示。
  - 在瀑布流播放器 `ShortDiscoverPlayerOverlay` 接入同样的进度状态同步与进度条显示。
  - 两处进度条均使用底部安全区内边距（`navigationBarsPadding`），确保不压系统手势区/导航区。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/core/ui/ShortVideoBottomProgressBar.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 21:09] 修复短视频进度条与底部标题重叠（首页贴边 + 播放器标题上移）
- Type: `implementation`
- Summary:
  - 首页短视频进度条去除左右与底部多余间距，修复为贴满内容宽度并紧贴底部 Tab 顶边。
  - 修复“我的”短视频播放器（UnifiedPlayer）底部标题与进度条重叠：标题容器增加 `navigationBarsPadding` 与额外底部留白，始终在进度条上方。
  - 修复瀑布流播放器（ShortDiscover）同类重叠问题：标题与进度条层级分离，标题上移到进度条上方。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt`
  - `plan.md`
- Verification:
  - `cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 21:35] 修复图片转 WebP 失败（ffmpeg 无 libwebp 编码器时自动回退）
- Type: `implementation`
- Summary:
  - 修复图片上传转换 WebP 的兼容性：当 ffmpeg 返回 `Unknown encoder 'libwebp'` 时，自动从 `libwebp` 编码参数回退到 ffmpeg 内置 `webp` 编码器重试。
  - 同步修复图片缩放接口中 `format=webp` 的编码路径：同样支持 `libwebp` 缺失时自动回退，避免生成图片变体失败。
  - 新增编码器缺失错误识别单测，覆盖单引号/双引号与 `Encoder not found` 关键报错文本。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./pkg/ffmpeg -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-19 21:49] 修复图片转 WebP 失败（ffmpeg 无 webp 编码器时回退 cwebp）
- Type: `implementation`
- Summary:
  - 增强 `ConvertToWebP`：先尝试 `libwebp`，再尝试 `webp`，两者均不可用时自动回退到 `cwebp` 编码，避免因 ffmpeg 编译裁剪导致上传失败。
  - 增强 `ResizeImage(format=webp)`：同样采用三级回退；当 ffmpeg 两种 webp 编码器都缺失时，先用 ffmpeg 缩放到临时 PNG，再用 `cwebp` 编码生成目标 WebP。
  - 新增/补充编码器不可用识别测试，覆盖 `Unknown encoder 'webp'` 场景，确保回退逻辑触发条件准确。
- Changed Files:
  - `pkg/ffmpeg/ffmpeg.go`
  - `pkg/ffmpeg/ffmpeg_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./pkg/ffmpeg -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-20 19:01] 管理端上传中心新增“清空已选文件”按钮
- Type: `implementation`
- Summary:
  - 为管理端上传中心的 `el-upload` 增加实例引用，接入 Element Plus 的 `clearFiles()` 能力。
  - 新增 `canClearSelectedFiles` 状态与 `clearSelectedFiles()` 方法，在非上传中场景下同时清空组件内部文件列表和受控 `uploadFileList`。
  - 在上传操作区增加“清空已选文件”按钮，仅清除当前已选视频文件，不影响表单字段和上传记录，和现有“清空上传记录”职责分离。
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-20 19:36] 管理端固定侧栏与全站视觉提质设计稿
- Type: `docs`
- Summary:
  - 新增管理端设计稿，明确固定侧栏、移动端抽屉导航、全站视觉系统、重点页与标准页的改造边界。
  - 设计稿明确本次只调整后台壳层、全局主题和页面骨架，不改业务接口、路由结构和核心数据流。
  - 定义了验收标准、风险控制和关键交互细节，作为后续实现计划的直接输入。
- Changed Files:
  - `docs/superpowers/specs/2026-04-20-admin-web-layout-refresh-design.md`
  - `plan.md`
- Verification:
  - `文档变更，无需构建/测试。`
- Rollback:
  - `git revert <commit>`

### [2026-04-20 19:43] 管理端固定侧栏与全站视觉提质实施计划
- Type: `docs`
- Summary:
  - 新增实现计划文档，按共享壳层、全局主题、重点页、标准页和最终验收拆分任务。
  - 计划明确本次不新增前端测试框架，验证方式为 `npm run build` 配合桌面/平板/移动端手工验收。
  - 计划包含每阶段提交点与最终 `plan.md` 实施记录模板，便于后续按任务执行。
- Changed Files:
  - `docs/superpowers/plans/2026-04-20-admin-web-layout-refresh.md`
  - `plan.md`
- Verification:
  - `文档变更，无需构建/测试。`
- Rollback:
  - `git revert <commit>`

### [2026-04-21 06:07] 修复管理端登录页回归、后台切页动画与内容区限宽
- Type: `implementation`
- Summary:
  - 为管理端新增最小前端回归测试基建（Vitest），并补充路由过渡判定测试，锁定“登录页保留淡入淡出、后台页面禁用切页动画”的行为。
  - 调整 `App.vue` 的路由渲染逻辑，将全局过渡收敛为仅公开页面启用，修复左侧菜单切换时整页跟随淡入淡出的问题，并补充 reduced-motion 兼容。
  - 移除全局 `.page-shell` 的宽度限制，使后台内容区按壳层自然铺开；登录页改为独立的全屏双栏布局，修复上下居中、表单对齐与 `.hero-panel` 文案对比度不足的问题。
- Changed Files:
  - `admin-web/package.json`
  - `admin-web/package-lock.json`
  - `admin-web/src/App.vue`
  - `admin-web/src/assets/theme.css`
  - `admin-web/src/views/Login.vue`
  - `admin-web/src/router/transition.js`
  - `admin-web/src/router/transition.spec.js`
  - `plan.md`
- Verification:
  - `cd admin-web && npm test -- src/router/transition.spec.js` passed.
  - `cd admin-web && npm run build` passed.
  - 登录页与后台切页的最终视觉效果未在真实浏览器中做人工验收，待补充页面级人工检查。
- Rollback:
  - `git revert <commit>`

### [2026-04-21 13:01] Android 主流程影院感深色美化 + 首页短视频进度条贴底栏
- Type: `implementation`
- Summary:
  - 为 Android app 新增统一的影院感深色视觉 token，收敛首页、我的页、底部导航、登录页和选服页的颜色、表面层级与强调色，减少主流程页面之间的视觉割裂。
  - 重做首页主壳与内容结构：顶部改为影院感 segmented tabs，长内容列表统一为深色卡片排布；全局底部导航改为更稳定的沉浸式样式并增强选中态层级。
  - 首页短视频页接入共用 `ShortVideoBottomProgressBar`，去除本地重复实现；进度条改为锚定内容底边，使其视觉上贴住底部 tabs 顶边，同时在短视频上下滑动或 pager 未稳定时隐藏，稳定后再恢复。
  - 我的页、登录页、选服页同步做最小必要的视觉收口，保持现有业务流程不变。
  - 新增短视频进度条显示条件单元测试，锁定“滑动中隐藏、稳定后恢复”的行为。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/AppChrome.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/ui/ShortVideoBottomProgressBar.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/auth/LoginScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedProgressState.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/shorts/ShortFeedProgressVisibilityTest.kt`
  - `plan.md`
- Verification:
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest` passed.
  - `source ~/.zprofile >/dev/null 2>&1; cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:assembleDebug` passed.
  - 页面级视觉与手势体验未在真实 Android 设备上人工验收，待补充真机/模拟器检查。
- Rollback:
  - `git revert <commit>`

### [2026-04-23 20:50] 电视剧接口化与管理端补全
- Type: `implementation`
- Summary:
  - 后端新增 app 端电视剧专区与详情接口、管理端电视剧系列/季/集 CRUD 路由，并补充 `series.active` 与 `episodes.video_id` 删除置空迁移，避免删除底层视频后直接级联丢失分集。
  - 管理端新增独立“电视剧管理”页与菜单，支持系列筛选、系列详情编辑、季/集折叠编辑、分集绑定现有 `type=episode` 视频，以及对应 API 封装与 helper 单测。
  - Android 电视剧模块改为真实接口驱动：新增 TV DTO、Repository、映射与 Hilt 绑定，目录页/详情页/播放器页不再依赖 `TvMockData`，播放器按真实 `video_id` 播放并继续沿用历史上报。
- Changed Files:
  - `migrations/0014_tv_series_management.up.sql`
  - `migrations/0014_tv_series_management.down.sql`
  - `internal/models/app.go`
  - `internal/models/admin.go`
  - `internal/repository/tv_repository.go`
  - `internal/services/tv.go`
  - `internal/services/tv_service_test.go`
  - `internal/handlers/tv.go`
  - `internal/handlers/router.go`
  - `internal/handlers/recommend_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/admin.spec.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/router/index.js`
  - `admin-web/src/views/TvSeriesManage.vue`
  - `admin-web/src/views/tvSeriesManage.helpers.js`
  - `admin-web/src/views/tvSeriesManage.helpers.spec.js`
  - `android-app/app/build.gradle.kts`
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/di/TvRepositoryModule.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvCatalogViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMockData.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvCatalogViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvRepositoryMappingTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
  - `cd android-app && GRADLE_USER_HOME="/Users/chee/Documents/workspace/ai-project/ai-video-server/android-app/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-26 22:03] 电视剧海报落盘与最近观看优先修复计划
- Type: `plan`
- Summary:
  - 用户反馈电视剧仍存在三类问题：海报/封面缺失、退出后无法恢复上次播放时间，以及“继续播放”与详情/播放器默认落点没有优先回到最近观看的剧集与分集。
  - 本轮修复范围覆盖后端电视剧 artwork 与 Android 电视剧选择逻辑：后端为电视剧系列补稳定 artwork 路由，并在刮削同步时把海报/背景图下载到本地硬盘；Android 详情页与播放器在未显式指定季/集时，统一优先最近观看记录。
  - 延续已完成的继续播放直达播放器、续播 seek 与字幕偏好记忆，不改后端字幕接口；验证目标包含新增 Go/Android 回归测试，以及 `go test ./...`、`:app:testDebugUnitTest`、`:app:assembleDebug` 全量回归。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./...`
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug`
- Rollback:
  - `git revert <commit>`

### [2026-04-26 22:04] 修复电视剧海报缺失、最近观看默认落点与本地 artwork 落盘
- Type: `implementation`
- Summary:
  - 后端为电视剧系列补了稳定 artwork 访问链路：列表、详情和继续播放现在统一返回 `/api/v1/tv/series/:id/poster|backdrop`，新 handler 会优先服务本地落盘图片，并兼容旧数据里已有的本地绝对路径、远程绝对 URL 与 TMDB 相对路径。
  - 电视剧刮削同步现在会把系列 poster/backdrop 最佳努力下载到 `storageRoot/tv/series/<seriesID>/poster.jpg|backdrop.jpg`，满足“刮削图片落到硬盘而不是直接走网络图”的要求，同时不让 artwork 下载失败阻断剧集入库。
  - Android 电视剧详情页与播放器在没有显式季/集参数时，都会优先最近观看分集；同季内默认分集选择也会先比 `lastWatchedAt`、再比 `watchSeconds`，让继续播放与详情默认落点和历史记录保持一致。
- Changed Files:
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `internal/repository/tv_repository.go`
  - `internal/handlers/router.go`
  - `internal/handlers/tv_artwork.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_episode_sync_test.go`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesDetailViewModel.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesDetailViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `plan.md`
- Verification:
  - `go test ./...` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvSeriesDetailViewModelTest' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-27 11:15] 管理端上传页远程过滤提示计划
- Type: `plan`
- Summary:
  - 用户希望管理端上传时，标签、所属合集、图片图集都能根据输入内容做远程请求过滤提示，而不是只依赖静态候选或本地过滤。
  - 现状确认：所属合集与图片图集接口已经支持 `q` 查询；视频标签只有热门标签接口，缺少按关键字搜索的管理端入口。
  - 本轮修复范围限定在管理端上传页与其依赖 API：后端补标签搜索接口，前端统一三类选择器的远程搜索体验，标签仍允许自由输入新值；验证目标包含 Go 单测、admin-web 单测与构建。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./...`
  - `cd admin-web && npm test`
  - `cd admin-web && npm run build`
- Rollback:
  - `git revert <commit>`

### [2026-04-27 11:15] 实现管理端上传页标签与合集远程过滤提示
- Type: `implementation`
- Summary:
  - 后端新增 `GET /api/v1/admin/video-tags`，支持按 `q` 与 `limit` 返回标签候选；空查询时回退到热门标签，继续沿用现有标签使用次数排序和归一化逻辑。
  - 管理端 API 层补 `getAdminVideoTags`，上传页把“视频标签、所属合集、图片图集”统一切到远程搜索流：空输入显示默认推荐，输入后按关键字请求过滤候选。
  - 新增上传页远程搜索 helper，统一处理防抖、只采纳最后一次请求结果、以及“保留当前已选项但不保留旧关键字残留候选”；标签仍允许自由创建新值，不改现有提交时的归一化逻辑。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_repository_tags_test.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/router.go`
  - `internal/handlers/recommend_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/admin.spec.js`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/videoUpload.remote.js`
  - `admin-web/src/views/videoUpload.remote.spec.js`
  - `plan.md`
- Verification:
  - `go test ./internal/repository ./internal/handlers` passed.
  - `cd admin-web && npm test -- src/api/admin.spec.js src/views/videoUpload.remote.spec.js` passed.
  - `go test ./...` passed.
  - `cd admin-web && npm test` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:29] AV mdcx 第三批站点迁移计划
- Type: `plan`
- Summary:
  - 用户要求继续把 `references/mdcx` 的 Go 迁移代码补进当前 AV 刮削内核，本轮聚焦第三批站点：`fc2club`、`fc2hub`、`fc2ppvdb`、`airav`、`jav321`、`mywife`。
  - 范围限定在 `internal/services` 的 provider/crawler 与 `ConfirmAV` detail URL 构造，不改现有上传触发、预览/确认 API 和 `videos.metadata` 结构。
  - 执行策略保持 TDD：先补 `ConfirmAV` 红测锁定 detail URL 与最小元数据，再补站点实现，最后做 Go 回归和 race 校验。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites'` 预期先红后绿。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:30] 接入 mdcx 第三批 AV 站点 crawler
- Type: `implementation`
- Summary:
  - 新增第三批站点 crawler：`fc2club`、`fc2hub`、`fc2ppvdb`、`airav`、`jav321` 走最小 HTML detail 抓取，`mywife` 走专用 detail parser，现已支持 `ConfirmAV` 按 `scrape_source + external_id` 构造 URL 并回抓标题、番号、演员与概要等基础字段。
  - 扩展 AV provider、detail URL 构造与 URL 匹配规则，并收紧 `airav_cc` / `fc2` 的 detail URL 识别，避免新站点被旧宽匹配规则误判；同时补齐 `airav_cc` 的 source 规范化映射。
  - 新增第三批回归测试，覆盖上述 6 个站点的 `ConfirmAV` 路径，确保新接入站点在本地 override host 下也能正确命中 crawler。
- Changed Files:
  - `internal/services/scraper_test.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForThirdBatchSites'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:52] AV mdcx 第四批站点迁移计划
- Type: `plan`
- Summary:
  - 在第三批之后继续补 `references/mdcx` 的剩余最小 HTML 站点，优先选择结构最相近的一组：`avsox`、`freejavbt`、`madouqu`、`mdtv.com`、`cnmdb`。
  - 本轮除站点 crawler、detail URL 构造和匹配规则外，还需要处理 `mdtv.com` 与 `cnmdb` 都使用 `/video/<id>` 路径带来的本地 override host 歧义。
  - 执行策略保持 TDD：先补第四批 `ConfirmAV` 红测，再实现站点接入和 source-hint 优先选 crawler，最后回归验证并提交。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForFourthBatchSites'` 预期先红后绿。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 20:52] 接入 mdcx 第四批 AV 站点 crawler
- Type: `implementation`
- Summary:
  - 接入 `avsox`、`freejavbt`、`madouqu`、`mdtv.com`、`cnmdb` 五个最小 HTML 站点，复用现有 minimal detail parser 并补充对应的 detail URL 构造与 URL 匹配规则，`ConfirmAV` 现在可以按 `scrape_source + external_id` 直接回抓这批站点的标题与番号。
  - `ConfirmAV` 的 detail 抓取新增 source-hint 优先选 crawler 逻辑，避免 `mdtv.com` / `cnmdb` 这类同路径站点在本地 override host 下被 path-only 规则误判；同时 `ConfigureAVScraperConfig` 与 source 规范化补充了 `mdtv` 别名处理。
  - 新增第四批回归测试，覆盖上述五个站点和 `mdtv` 别名路径，锁定本轮新增规则及编号提取行为。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForFourthBatchSites'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-30 21:09] AV mdcx 第五批站点迁移计划
- Type: `plan`
- Summary:
  - 在第四批之后继续补 `references/mdcx` 的剩余日站点最小 HTML 组：`faleno`、`fantastica`、`giga`、`javday`、`kin8`、`love6`、`lulubar`。
  - 本轮除补站点 crawler、detail URL 构造和匹配规则外，还需要把 `giga`、`kin8`、`lulubar` 等站点的 site-specific external_id 规则带进当前 minimal HTML 内核，避免沿用通用 path 提取时取错值。
  - 执行策略保持 TDD：先补第五批 `ConfirmAV` 红测，再扩展现有 minimal HTML crawler 的 URL 与 external_id helper，最后做 Go 回归和 race 校验。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForFifthBatchSites'` 预期先红后绿。
- Rollback:
  - `git revert <commit>`

### [2026-04-30 21:09] 接入 mdcx 第五批 AV 站点 crawler
- Type: `implementation`
- Summary:
  - 接入 `faleno`、`fantastica`、`giga`、`javday`、`kin8`、`love6`、`lulubar` 七个日站点，复用现有 minimal HTML parser 并补充对应的 detail URL 构造、URL 匹配和 external_id 提取规则，`ConfirmAV` 现在可以按 `scrape_source + external_id` 直接回抓这批站点的标题与番号。
  - 扩展 minimal HTML crawler 的站点特化逻辑，覆盖 `giga.product_id`、`kin8.moviepages/<id>`、`love6` 路径片段和 `lulubar?id=` 等 external_id 规则；同时新增 `faleno`、`fantastica`、`giga`、`javday`、`kin8`、`love6`、`lulubar` 的 detail URL 匹配规则。
  - `love6` 的 detail URL 构造改为保留传入 external_id 原样，避免误改大小写导致 query 命中失败；第五批回归测试已覆盖这类 query/path 细节。
- Changed Files:
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_av_mdcx_third_batch_sites.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestConfirmAVBuildsMDCXDetailURLsForFifthBatchSites'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-30 21:15] AV mdcx 收尾检查计划
- Type: `plan`
- Summary:
  - 对照 `references/mdcx` 的 single scrape 站点与当前 provider 做收尾检查，确认站点清单本身已经全部对齐，重点转向搜索回退能力缺口。
  - 检查结果显示：当前范围内唯一明确未补齐的功能是 `getchu`，它已有 detail parser，但 `PreviewAV` 使用的 `SearchCandidates` 仍为空实现，与 reference 的搜索回退能力不一致。
  - 本轮按最小修复处理：补 `getchu` 的 `PreviewAV` 回归测试与搜索实现，完成后再做一次全量 Go 验证，作为这阶段迁移的收尾。
- Changed Files:
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVFallsBackTo(Getchu|ThePornDB)WhenConfigured|TestPreviewAVFallsBackToAVSexWhenPrimarySitesHaveNoResult'`
- Rollback:
  - `git revert <commit>`

### [2026-04-30 21:15] 补齐 getchu 搜索回退并完成收尾验证
- Type: `implementation`
- Summary:
  - 为 `getchu` 补上 `SearchCandidates`，现在 `PreviewAV` 可以按 reference 逻辑请求 `/php/search.phtml?genre=all&search_keyword=...&gc=gc`，解析搜索页里的 `a.blueb` 命中并回抓 detail。
  - 新增 `PreviewAV` 对 `getchu` 的回归测试，同时修正 `getchu` 搜索命中后的 detail URL 组装，避免相对链接里的 query 被误当成 path 导致 404。
  - 收尾检查结论：当前 `references/mdcx` 已进入本次迁移范围的站点和核心预览/确认链路已全部对齐，后续若继续扩展，重点会转向更细的搜索策略或字段丰富度，而不是站点缺口。
- Changed Files:
  - `internal/services/scraper_av_mdcx_detail_sites.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestPreviewAVFallsBackTo(Getchu|ThePornDB)WhenConfigured|TestPreviewAVFallsBackToAVSexWhenPrimarySitesHaveNoResult'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
  - `go vet ./...` passed.
  - `go test -race ./internal/services ./internal/queue` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 03:51] Android 电视剧播放页全屏与季数滚动修复计划
- Type: `plan`
- Summary:
  - 修复 `TvSeriesPlayerScreen` 的两个 Android 播放问题：播放器全屏按钮无效，以及选集抽屉里的季数标签超过 5 个后无法横向滚动。
  - 采用最小改动方案，在页面层直接接入现有 `LongFormVideoPlayer` 的全屏状态管理，并将抽屉季数条改成与电视剧详情页一致的横向滚动 `Row`。
  - 验证以 Android Kotlin 编译通过为准，并手工覆盖全屏进入/退出、系统返回键和多季滚动场景。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:compileDebugKotlin`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 03:51] 修复 Android 电视剧播放页全屏按钮与选集季数滚动
- Type: `implementation`
- Summary:
  - 在 `TvSeriesPlayerScreen` 接入 `isFullscreen`、`BackHandler` 和沉浸式横屏控制，非全屏状态下全屏按钮现在会切入独立黑底全屏播放器，退出全屏时恢复页面布局与系统栏。
  - 补上切集时退出全屏的状态复位，并抽出页面内共用的播放器错误提示 banner，保持普通模式和全屏模式下的错误提示表现一致。
  - 将选集抽屉中的季数选择条改为 `fillMaxWidth() + horizontalScroll(rememberScrollState())`，多季场景下可左右滑动，不再被压缩换行。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew :app:compileDebugKotlin` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-01 07:57] 管理端电视剧手动刮削季集支持计划
- Type: `plan`
- Summary:
  - 为管理端 `/scrape` 的 `type=tv` 模式补齐“剧名 + 季集”支持：后端预览接口要能从 `庆余年第一季第一集`、`Breaking Bad S01E01` 这类输入中拆出剧名与季集，前端要允许显式填写并在确认时透传 `season_number/episode_number`。
  - 这次实现同时修正手动电视剧确认默认覆写分集元数据的问题：候选选择后不再默认把整剧标题/简介/海报写回确认表单，而是让后端优先落目标分集的标题、简介、still 和播出日期。
  - 验证覆盖 Go handler/utils 单测、管理端 vitest 回归与 `admin-web` 生产构建，确保搜索清洗、payload 透传和页面语法一起稳定。
- Changed Files:
  - `internal/utils/filename_parser.go`
  - `internal/handlers/admin_scrape.go`
  - `admin-web/src/views/ScrapePreview.vue`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/handlers`
  - `cd admin-web && npm test -- src/views/scrapePreview.helpers.spec.js src/views/videoList.helpers.spec.js src/api/admin.spec.js`
  - `cd admin-web && npm run build`
- Rollback:
  - `git revert <commit>`

### [2026-05-01 07:57] 实现管理端电视剧手动刮削季集支持
- Type: `implementation`
- Summary:
  - 扩展文件名/标题解析器，新增中文季集模式支持：`第1季第2集`、`第一季第一集` 等标题现在可解析为清洗后的剧名和季集，`AdminScrapePreview` 在 `type=tv` 时会先用清洗后的剧名搜 TMDB，并把 `parsed_title/parsed_season_number/parsed_episode_number` 回传给前端。
  - 管理端刮削页新增电视剧季/集输入与回填逻辑，待绑定视频跳转 `/scrape` 时也会把解析出的季集带入；电视剧确认保存时会稳定透传 `season_number/episode_number`，同时默认留空标题/简介/海报/日期，让后端优先落目标分集元数据而不是整剧元数据。
  - 新增 Go handler 回归测试与前端 helper/vitest 回归，锁定“中文标题清洗后搜索 TMDB”、“电视剧 payload 带季集”和“候选解析状态回填”三条关键行为。
- Changed Files:
  - `internal/utils/filename_parser.go`
  - `internal/utils/filename_test.go`
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/admin_scrape_test.go`
  - `internal/handlers/swagger_models.go`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/scrapePreview.helpers.js`
  - `admin-web/src/views/scrapePreview.helpers.spec.js`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/handlers` passed.
  - `cd admin-web && npm test -- src/views/scrapePreview.helpers.spec.js src/views/videoList.helpers.spec.js src/api/admin.spec.js` passed.
  - `cd admin-web && npm run build` passed.
- Rollback:
  - `git revert <commit>`

## [2026-05-05 06:47 +0800] Android：短视频搜索 Tab + 短视频播放模式切换（增量实施）
- 摘要：完成底栏 `搜索` Tab、仅短视频搜索页（标题/标签搜索、瀑布流、点击进入上下滑沉浸播放）以及三处上下滑短视频入口统一播放模式切换（`循环单视频` / `自动播放下一个`，默认循环单视频并持久化）。
- 变更文件：
  - android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt
  - android-app/app/src/main/java/com/chee/videos/core/ui/AppNavigationConfig.kt
  - android-app/app/src/main/java/com/chee/videos/core/model/ShortPlaybackMode.kt
  - android-app/app/src/main/java/com/chee/videos/core/data/AppPreferencesStore.kt
  - android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchViewModel.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedViewModel.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverViewModel.kt
  - android-app/app/src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt
  - android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerViewModel.kt
  - android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt
  - android-app/app/src/test/java/com/chee/videos/core/data/AppPreferencesStoreTest.kt
  - android-app/app/src/test/java/com/chee/videos/core/ui/AppNavigationConfigTest.kt
  - android-app/app/src/test/java/com/chee/videos/core/model/ShortPlaybackModeTest.kt
  - android-app/app/src/test/java/com/chee/videos/feature/shortsearch/ShortSearchViewModelStateTest.kt
- 验证：
  - `cd android-app && ./gradlew testDebugUnitTest --tests "com.chee.videos.core.data.AppPreferencesStoreTest" --tests "com.chee.videos.core.ui.AppNavigationConfigTest" --tests "com.chee.videos.core.model.ShortPlaybackModeTest" --tests "com.chee.videos.feature.shortsearch.ShortSearchViewModelStateTest" --tests "com.chee.videos.feature.player.UnifiedPlayerShortFitToggleTest"`
  - 结果：通过（BUILD SUCCESSFUL）。
- 备注：未改动 `.codex/skills/*`；后续可继续补充 `ShortFeed/ShortDiscover/UnifiedPlayer` 播放模式行为断言的更细粒度测试。

### [2026-05-05 07:29] flick 迁移支持 checkpoint 断点续跑（不再每次全量重扫）
- Type: `implementation`
- Summary:
  - 为 `cmd/import-flick` 增加断点续跑能力，新增 `--checkpoint-path` 与 `--resume`（默认开启）。
  - 每批次完成后把最后一条源记录的 `createdAt + _id` 持久化到 checkpoint 文件。
  - 启动时自动读取 checkpoint 并构造 Mongo 条件：`createdAt > checkpoint.createdAt` 或 `createdAt == checkpoint.createdAt 且 _id > checkpoint._id`，避免从头扫描。
  - Mongo 导出排序改为 `createdAt asc, _id asc`，保证断点语义稳定可重复。
  - 报告中增加 checkpoint 相关元数据，便于排查是否启用续跑。
- Changed Files:
  - `cmd/import-flick/main.go`
  - `cmd/import-flick/main_test.go`
  - `plan.md`
- Verification:
  - `go test ./cmd/import-flick -count=1` passed.
  - 运行参数已切到新二进制并启用 checkpoint：`--checkpoint-path .run/reports/flick-import.checkpoint.json --resume=true`。
- Rollback:
  - `git revert <commit>`

### [2026-05-06 13:27] 管理端视频批量删除与空表宽度修复
- Type: `implementation`
- Summary:
  - 管理端视频列表新增勾选批量删除能力，前后端增加 `/admin/videos/batch-delete` 批量删除接口，按逐条执行并返回成功/失败汇总结果。
  - 视频管理页保留单条删除，同时新增批量删除按钮、选择列、二次确认、汇总提示与删除后分页回退处理。
  - 管理端公共表格容器样式补齐 `min-width` 与横向滚动约束，修复空查询结果时表格无限变宽问题，覆盖所有使用 `.table-wrap` 的页面。
- Changed Files:
  - `internal/handlers/admin.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin_batch_delete_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/admin.spec.js`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/assets/theme.css`
  - `plan.md`
- Verification:
  - `go test ./internal/handlers -run 'TestDeleteVideosIndividuallySummarizesPartialFailures|TestAdminBatchDeleteVideosRejectsEmptyVideoIDs|TestAdminBatchDeleteVideosRejectsInvalidUUID' -count=1` passed.
  - `cd admin-web && npm test -- --run admin.spec.js` passed.
  - `cd admin-web && npm run build` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 18:54 +0800] Android TV：补齐二维码渲染与账户/设备菜单
- Type: `implementation`
- Summary:
  - TV 配对页改为真实渲染二维码位图，继续保留配对码文案，适配手机扫码失败时的备用输入路径。
  - TV 配对页新增 `切换服务器` 入口，补齐“先选服务器，再生成配对会话”的回退链路。
  - TV 已登录态新增顶部 `账户与设备菜单`，提供 `重新配对`、`退出登录`、`切换服务器` 三个动作。
  - 新增 QR 编码器与菜单动作单测，保证新增 TV 闭环能力可回归验证。
- Changed Files:
  - `android-app/tv-app/build.gradle.kts`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvPairingScreen.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvQrCodeEncoder.kt`
  - `android-app/tv-app/src/main/java/com/chee/videos/tv/TvAccountMenuAction.kt`
  - `android-app/tv-app/src/test/java/com/chee/videos/tv/TvQrCodeEncoderTest.kt`
  - `android-app/tv-app/src/test/java/com/chee/videos/tv/TvAccountMenuActionTest.kt`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvQrCodeEncoderTest' --tests 'com.chee.videos.tv.TvAccountMenuActionTest'` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 19:15 +0800] Android：修复 TV 首页 continue_watching.type 为空导致的崩溃
- Type: `implementation`
- Summary:
  - 复现并修复 TV 首页 `continue_watching.type = null` 时映射层空指针，避免 `TvContinueWatchingUiModel` 构造时因非空参数收到 `null` 而崩溃。
  - TV 映射层新增视频类型归一化，统一兜底 `tv/movie/av` 三种合法值，脏数据或空串时按场景回退默认类型。
  - 补充回归测试，覆盖 `continue_watching.type = null` 的真实输入形态。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/tv/TvMappers.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tv/TvRepositoryMappingTest.kt`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvRepositoryMappingTest.mapContinueWatching_fallsBackWhenTypeIsNull' --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 19:31 +0800] Android：拆分独立 TV 工程目录并移除手机工程内 TV module
- Type: `implementation`
- Summary:
  - 新增顶级独立工程 `android-tv-app/`，复制并承接原 TV 运行所需源码，TV 工程可单独在 Android Studio 打开并独立构建。
  - 旧 `android-app/tv-app` 从手机工程中移除，`android-app/settings.gradle.kts` 不再包含 `:tv-app`，手机工程只保留 `:app`。
  - 新 TV 工程的 `tv-app` 取消对 `../app/src/*` 的 `sourceSets` 相对路径引用，真正改为目录隔离。
  - 补充 `android-tv-app/README.md`、`android-tv-app/AGENTS.md`，并更新根 `AGENTS.md` 与手机工程 `README.md` 说明新的目录边界。
- Changed Files:
  - `android-app/settings.gradle.kts`
  - `android-app/README.md`
  - `android-app/tv-app/*`（删除）
  - `android-tv-app/**`
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - `test -d android-tv-app` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.tv.TvQrCodeEncoderTest' --tests 'com.chee.videos.feature.tv.TvCatalogViewModelTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tvauth.TvAuthDeepLinkParserTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :tv-app:assembleDebug` failed as expected with `project 'tv-app' not found`。
- Rollback:
  - `git revert <commit>`

### [2026-05-07 19:34 +0800] Android TV：补充独立工程构建产物忽略规则
- Type: `implementation`
- Summary:
  - 为 `android-tv-app/.gitignore` 增加 `tv-app/build/` 忽略规则，避免独立 TV 工程模块内的 Gradle 产物再次进入版本库。
  - 清理已被误纳入版本库的 `android-tv-app/tv-app/build/**` 生成文件，保持独立工程只提交源码、资源和构建脚本。
- Changed Files:
  - `android-tv-app/.gitignore`
  - `plan.md`
- Verification:
  - `git status --short` 仅保留预期源码变更与未跟踪的 `storage/`。
- Rollback:
  - `git revert <commit>`

### [2026-05-07 20:35 +0800] Android：修复“我的”页来源短视频播放器工具按钮重合
- Type: `implementation`
- Summary:
  - 修复从 `历史记录 / 我的收藏 / 我的喜欢` 进入短视频播放器时，画面比例切换与播放模式切换两个覆盖按钮直接叠在同一 `Box` 中导致重合的问题。
  - 为短视频播放器新增可测试的工具动作栏布局规则：当两个工具按钮同时出现时，必须使用显式纵向排布并保留固定间距。
  - 在 `android-app/AGENTS.md` 增加约束：短视频覆盖层的多个按钮不得依赖默认叠放，必须有显式布局约束。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerShortState.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/player/UnifiedPlayerShortFitToggleTest.kt`
  - `android-app/AGENTS.md`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.player.UnifiedPlayerShortFitToggleTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-07 20:43 +0800] Android：补齐手机端 TV 扫码入口
- Type: `implementation`
- Summary:
  - 手机端新增应用内 TV 扫码入口，在“我的”页资料头部提供扫码按钮，直接拉起相机扫描 TV 登录二维码。
  - 扫码结果接入现有 TV 授权确认流：应用内扫码结果优先于外部 deep link，扫码无效时给出提示，不另起第二套确认逻辑。
  - 新增 `resolveTvAuthDeepLink` 合流逻辑与回归测试，补充扫码依赖、相机权限与工程约束，明确手机端 TV 授权不能只做 deep link 解析。
- Changed Files:
  - `android-app/app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/mine/MineScreen.kt`
  - `android-app/app/src/main/java/com/chee/videos/feature/tvauth/TvAuthDeepLink.kt`
  - `android-app/app/src/test/java/com/chee/videos/feature/tvauth/TvAuthDeepLinkParserTest.kt`
  - `android-app/app/src/main/AndroidManifest.xml`
  - `android-app/app/build.gradle.kts`
  - `android-app/AGENTS.md`
  - `plan.md`
- Verification:
  - `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests 'com.chee.videos.feature.tvauth.TvAuthDeepLinkParserTest'` passed.
  - `cd android-app && ./gradlew --no-daemon :app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`

### [2026-05-08 16:06] TV 首页货柜裁剪与分类海报墙
- Type: `implementation`
- Summary:
  - TV 首页货柜统一裁剪为前 `8` 张海报并追加“查看更多”尾卡，入口按 section / 电视剧 / 电影 / AV 分类进入独立墙页，不再跳混合页。
  - 新增 TV 分类海报墙路由、ViewModel 与页面，支持显式顶部刷新、底部自动续载下一页、返回与焦点高亮，后端补齐 `tv/catalog` 分页接口与墙页 DTO。
  - 共享 `HomeScreen` / `TvShellApp` / `VideoHomeApp` 的 TV 导航链同步接入新墙页，补齐测试 fake 与相关单测。
- Changed Files:
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/VideoHomeApp.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/model/ApiModels.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/network/ApiService.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvModels.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPosterWallViewModel.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvPresentation.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvRoutes.kt`
  - `android-tv-app/tv-app/src/main/java/com/chee/videos/tv/TvShellApp.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvPosterWallViewModelTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvRoutesTest.kt`
  - `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvTestSupport.kt`
  - `internal/handlers/router.go`
  - `internal/handlers/tv.go`
  - `internal/models/app.go`
  - `internal/repository/tv_repository.go`
  - `internal/services/tv_auth.go`
  - `internal/services/tv_catalog_wall_test.go`
  - `plan.md`
- Verification:
  - `go test ./... -count=1` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.TvRoutesTest' --tests 'com.chee.videos.feature.tv.TvPosterWallViewModelTest'` passed.
  - `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` passed.
- Rollback:
  - `git revert <commit>`
## 2026-05-09 19:20
- 进度：开始执行 AV 刮削重写方案，先补 ThePornDB 公网详情 URL 归一化回归测试，再调整短视频转 AV 的状态保持逻辑。
- 影响文件：`internal/services/scraper_test.go`、`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/queue/scrape_tasks.go`、`internal/services/scraper.go`、`internal/services/scraper_av_strategy.go`
- 验证：待执行

## 2026-05-10 00:00
- 进度：继续对齐 `references/mdcx/scraper-preview-go` 的已迁移 AV 站点策略；DMM 改为 reference 的多搜索 URL、区域封锁识别、FANZA TV GraphQL 与 digital 分支解析；ThePornDB 改为 `detailUrl > filePath > number` 链路下的 scenes/movies 搜索、MDCx 文件名关键词和配置 base 归一化；MGStage 会话 cookie + `adc=1` 回抓与 trailer URL 对齐；DMM 图片字段改为 `poster=横版 ps`、`thumb=竖版 pl`。
- 影响文件：`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_framework.go`、`internal/services/scraper_av_mdcx_sites_test.go`、`plan.md`
- 验证：`go test ./internal/services`、`go test ./internal/handlers ./internal/queue`、`go test ./internal/services -run 'TestPreviewAVFallsBackToThePornDBWhenConfigured|TestConfirmAVNormalizesThePornDBPublicDetailURL|TestMDCxMigratedSitesSearchCandidates' -count=1`、`go vet ./...` 均通过。

## 2026-05-10 00:20
- 进度：修复 DMM 搜索入口在单个搜索 URL 失败时直接返回的问题，改为按 reference 继续尝试后续搜索 URL；补充 DMM 回归测试，覆盖“首个搜索 URL 失败但第二个命中”的场景。
- 影响文件：`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_mdcx_sites_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestDMMSearchCandidatesReturnsRegionError|TestDMMSearchCandidatesSkipsFailedSearchURL|TestMDCxMigratedSitesSearchCandidates' -count=1` 通过。

## 2026-05-10 02:20
- 进度：进一步缓解 DMM 30 秒超时，把三个搜索 URL 改成并发探测，避免搜索失败按串行累积等待；保留区域封锁优先返回和失败 URL 跳过逻辑。
- 影响文件：`internal/services/scraper_av_mdcx_detail_sites.go`、`plan.md`
- 验证：`go test ./internal/services`、`go vet ./...` 通过。

## 2026-05-10 02:52
- 进度：修复 DMM 详情候选的失败回退逻辑，改为并发尝试当前搜索页里的多个候选，不再因首个候选超时或 504 直接中断；补回归测试覆盖“首个候选失败、第二个候选成功”。
- 影响文件：`internal/services/scraper_av_mdcx_detail_sites.go`、`internal/services/scraper_av_mdcx_sites_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestDMMSearchCandidatesReturnsRegionError|TestDMMSearchCandidatesSkipsFailedSearchURL|TestDMMSearchCandidatesSkipsFailedDetailURL|TestMDCxMigratedSitesSearchCandidates' -count=1`、`go test ./internal/services ./internal/handlers ./internal/queue`、`go vet ./...` 通过。

## 2026-05-10 08:20
- 进度：修复短视频转 AV 的状态收尾，成功后不再沿用 `scraping`，而是统一落回 `ready`；同步把自动重刮削回归测试改成验证成功后状态可放行给只认 `ready` 的资源接口。
- 影响文件：`internal/queue/scrape_tasks.go`、`internal/queue/scrape_tasks_test.go`、`plan.md`
- 验证：`go test ./internal/queue -run 'TestAutoScrapeAVMarksReadyOnSuccess|TestBuildScrapeFailureDecisionMarksEpisodeAsTVPending|TestBuildScrapeFailureDecisionKeepsMovieFallbackBehavior' -count=1`、`go test ./internal/services -run 'TestConfirmAVPreservesExplicitReadyStatus|TestScrapeAVUploadCodeFirstAndActorSync' -count=1`、`go test ./internal/services ./internal/handlers ./internal/queue`、`go vet ./...` 通过。

## 2026-05-23 01:17 +0800
- 进度：修复手机端短视频全屏没有真正覆盖首页头部和底部 tabbar 的问题。根因是 `ShortOverlayFullscreenHost` 之前只挂在 `ShortFeedScreen` 子树里，受外层 `VideoHomeApp` 的 `Scaffold(topBar/bottomBar)` 限制，无法跨出内容区。修法：将全屏层改为 `Dialog(usePlatformDefaultWidth = false, decorFitsSystemWindows = false)` 承载 `LongFormVideoPlayer`，让它提升到 window 级别覆盖；保留原有方向锁、系统栏隐藏、repeatMode 处理与返回键语义不变。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`、`android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`、`plan.md`
- 验证：新增结构性红灯测试 `shared fullscreen host uses fullscreen dialog to escape parent scaffold` 已先失败，确认抓到问题；随后实现后通过。手机端 `cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过，`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过。

## 2026-05-10 09:45
- 进度：新增刮削内容中文翻译接入，支持通过 OpenAI-compatible 本地翻译接口连接 `HY-MT1.5-1.8B` 等模型；刮削写库前会保存中文标题/简介到展示字段，并在 `metadata` 中同步写入 `title_original`、`description_original`、`title_zh`、`description_zh`，演员名保持原样。
- 影响文件：`main.go`、`.env.example`、`internal/config/config.go`、`internal/services/scraper.go`、`internal/services/translation.go`、`internal/config/config_test.go`、`internal/services/translation_test.go`、`internal/services/scraper_translation_test.go`、`plan.md`
- 验证：`go test ./internal/config ./internal/services -run 'TestLoadIncludesTranslationConfig|TestOpenAITextTranslatorTranslateScrapeContent|TestScrapeMovieUploadStoresLocalizedFields|TestScrapeAVUploadStoresLocalizedFieldsAndKeepsActors'`、`go test ./...`、`go vet ./...` 均通过。

## 2026-05-10 09:55
- 进度：补齐本机 `.env` 翻译配置，并修复 `dev-up.sh` 后台进程保活问题；脚本改为先构建 `.run/bin/video-server`，再用 detached 方式启动后端、worker 与前端，避免脚本退出后服务进程被带走。
- 影响文件：`.env`（本地忽略文件）、`scripts/dev-up.sh`、`plan.md`
- 验证：`bash -n scripts/dev-up.sh`、`bash scripts/dev-down.sh`、`bash scripts/dev-up.sh`、`curl -fsS http://127.0.0.1:8080/healthz`、`curl -I -fsS http://127.0.0.1:5173/` 均通过。

## 2026-05-10 15:18
- 进度：统一修正 AV 刮削翻译与 JavDB 海报策略；手动保存、上传刮削、短视频转 AV 三条链路都补齐中文字段落库回归，`metadata` 继续保留 `title_original` / `description_original` / `title_zh` / `description_zh`；JavDB 改为直接使用大横幅海报作为 `poster_url`，竖版海报由大图裁剪生成。
- 影响文件：`internal/services/scraper_av_framework.go`、`internal/services/scraper_av_poster_assets_test.go`、`internal/services/scraper_javdb_mdcx_regression_test.go`、`internal/services/scraper_test.go`、`internal/services/scraper_translation_test.go`、`internal/queue/scrape_tasks_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestScrapeAVUploadStoresLocalizedFieldsAndKeepsActors|TestConfirmAVStoresLocalizedFieldsAndOriginalFields|TestPreviewAVJavDBUsesMDCXStyleDOMDetailParsing|TestPreviewAVJavDBPrefersCoverAndDetailOverview|TestConfirmAVStoresOriginalAndCroppedPosterAssets|TestConfirmAVFallsBackToOriginalPosterWhenCropFails|TestConfirmAVStoresSeparateThumbWhenThumbURLIsAvailable' -count=1`、`go test ./internal/queue -run 'TestAutoScrapeAVMarksReadyOnSuccess|TestAutoScrapeAVStoresLocalizedFieldsAndMarksReady|TestAutoScrapeAVUsesTitleToSelectSite' -count=1`、`go test ./internal/services ./internal/queue ./internal/handlers -count=1`、`go test ./... -count=1`、`go vet ./...` 均通过。

## 2026-05-10 17:14
- 进度：修复真实 HY-MT 模型返回 `title` / `description` 时后端未读取的问题；翻译响应现在同时兼容 `title_zh` / `description_zh` 与 `title` / `description`，刮削落库会使用中文标题和中文简介，同时保留原文字段。
- 影响文件：`internal/services/translation.go`、`internal/services/translation_test.go`、`plan.md`
- 验证：`go test ./internal/services -run 'TestOpenAITextTranslatorTranslateScrapeContent|TestOpenAITextTranslatorAcceptsGenericTitleAndDescriptionFields' -count=1`、`go test ./internal/services -run 'Translation|Localized|ConfirmAVStoresLocalized|ScrapeAVUploadStoresLocalized' -count=1`、`go test ./internal/queue -run 'AutoScrapeAVStoresLocalized' -count=1`、`go test ./... -count=1`、`go vet ./...` 均通过。

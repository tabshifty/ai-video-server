# plan.md

本文件用于增量记录“计划与修改”，不得覆盖历史记录，只能追加。

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

# plan.md

本文件用于增量记录“计划与修改”，不得覆盖历史记录，只能追加。

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

---

### [2026-04-19 12:26] 修复短视频尺寸模式切换时封面不跟随的问题
- Type: `implementation`
- Summary:
  - 定位根因：短视频尺寸模式切换只影响 `PlayerView.resizeMode`，封面图仍固定使用 `ContentScale.Crop`，导致视频切到完整显示时封面没有同步切换。
  - 在短视频页提取统一的封面缩放映射 `shortPosterContentScale(fitMode)`，使封面图与视频共用同一套 `VideoFitMode` 语义：`FILL -> Crop`，`FIT -> Fit`。
  - 当前页显示中的封面与非当前页预览封面统一接入该映射，避免只修一条分支后再次出现模式不一致。
  - 新增 Android 单元测试锁定短视频封面缩放映射，防止后续改动回归。
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

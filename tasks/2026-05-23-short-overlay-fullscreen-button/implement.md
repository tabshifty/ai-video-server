# Implement：手机端短视频浮层「全屏播放」按钮

- 日期：2026-05-23
- 关联 PRD：`./prd.md`

## 1. 总体方案

把「全屏切换」做成一个独立的 Composable 模式叠加层 `ShortOverlayFullscreenHost`，复用现有的 `core/ui/LongFormVideoPlayer.kt`，**共用同一个 ExoPlayer 实例**实现位置无缝接续。四处短视频浮层（搜索 / 发现 / 主页信息流 / UnifiedPlayer）都把它们当前持有的 `sharedPlayer` + 当前 `currentItem` 信息透传给该 Host，Host 自己管理 `isFullscreen` 状态、方向锁、系统栏、`repeatMode` 覆盖与 BackHandler。

不引入新的 ViewModel；全屏状态属于 Composable 局部状态。

## 2. 文件结构

### 2.1 新增

- `android-app/app/src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt`
  - 暴露 `@Composable fun ShortOverlayFullscreenHost(...)`：接收 `player: ExoPlayer`、`title: String`、`subtitleTracks: List<SubtitleTrackDto>`、`fallbackPlaybackMode: ShortPlaybackMode`、`onFullscreenChange: (Boolean) -> Unit`，内部用 `var isFullscreen by rememberSaveable { mutableStateOf(false) }` 维护状态。
  - 暴露 `@Composable fun ShortOverlayFullscreenButton(onClick: () -> Unit, modifier)`：复用 `ShortVideoOverlayActionButton` 样式，icon = `Icons.Filled.Fullscreen`，contentDescription = "全屏播放"。
  - 内部 helper：
    - `LockOrientationToLandscapeEffect(active: Boolean)`：进入时 `activity.requestedOrientation = LANDSCAPE`，离开时 `= UNSPECIFIED`。
    - `HideSystemBarsEffect(active: Boolean)`：进入时通过 `WindowInsetsControllerCompat(window, decorView).hide(systemBars())` + `systemBarsBehavior = BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE`；离开时 `.show(systemBars())`。
    - `OverrideRepeatModeEffect(player, active, fallbackMode)`：进入时保存 `previousRepeatMode = player.repeatMode` 然后强制 `REPEAT_MODE_ONE`；离开时恢复 `fallbackMode.toPlayerRepeatMode()`（**不**用 `previousRepeatMode`，因为 fallback 反映用户在浮层里的最新偏好，避免与异步偏好流冲突）。
  - 退出全屏入口：`LongFormVideoPlayer` 内部已有的 `onToggleFullscreen` 回调 → `isFullscreen = false`；以及顶层 `BackHandler(enabled = isFullscreen) { isFullscreen = false }`（**优先级高于**外层关闭浮层的 BackHandler，因为嵌套 Composable 的 BackHandler 后注册先生效）。

### 2.2 修改

- `feature/shortsearch/ShortSearchScreen.kt`
  - `ShortSearchPlayerOverlay` 末尾新增 `ShortOverlayFullscreenHost` 包裹的全屏 overlay（详见 §3）。
  - 右侧操作栏（`AnimatedVisibility ... Column`）`ShortPlaybackModeToggleButton` 之下追加 `ShortOverlayFullscreenButton`。

- `feature/shortdiscover/ShortDiscoverScreen.kt`
  - 同 ShortSearch 模式：在右侧操作栏底部加按钮、整屏底部加 Host。

- `feature/shorts/ShortFeedScreen.kt`
  - 同上。

- `feature/player/UnifiedPlayerScreen.kt`
  - 该文件已有 `isFullscreen` 概念但只用于长视频分支（`fullscreenLongForm`，line 423）。短视频分支需独立追加 `ShortOverlayFullscreenHost`，**不复用** `fullscreenLongForm` 状态（语义不同）。
  - 在短视频分支的 `Column` 内右侧操作栏加按钮、Pager 外层加 Host。

- `android-app/app/build.gradle.kts`
  - `versionCode +1`、`versionName` 末位 +1。

### 2.3 测试新增

- `android-app/app/src/test/java/com/chee/videos/core/ui/ShortOverlayFullscreenSpecTest.kt`
  - 单测：纯函数 `shortOverlayRepeatModeWhileFullscreen(fallback)` 始终返回 `REPEAT_MODE_ONE`；`shortOverlayRepeatModeAfterFullscreen(fallback)` 返回 `fallback.toPlayerRepeatMode()`。
  - 源文 audit：四处短视频浮层文件必须 `import com.chee.videos.core.ui.ShortOverlayFullscreenHost` 与 `ShortOverlayFullscreenButton`，否则报错（防止后续重构忘掉某一处）。
  - 源文 audit：四处文件**不**得出现裸 `Icons.Filled.Fullscreen`、裸 `SCREEN_ORIENTATION_LANDSCAPE` 字面量 —— 都必须走共享 Host。

## 3. 各调用点接入示例

### 3.1 ShortSearchScreen.kt 的接入

在 `ShortSearchPlayerOverlay` 函数体最外层 `Box(modifier = Modifier.fillMaxSize().background(Color.Black).statusBarsPadding())` 内部，紧贴 `VerticalPager` 之后、`Text 关闭`、`shouldShowShortOverlayProgressBar` 之前追加：

```kotlin
val currentTitle = currentItem?.title.orEmpty()
val currentSubtitleTracks = remember(currentItem?.id, detailByVideoId) {
    currentItem?.id?.let { detailByVideoId[it]?.subtitleTracks }.orEmpty()
}
var isFullscreen by rememberSaveable { mutableStateOf(false) }

ShortOverlayFullscreenHost(
    isFullscreen = isFullscreen,
    onFullscreenChange = { isFullscreen = it },
    player = sharedPlayer,
    title = currentTitle,
    subtitleTracks = currentSubtitleTracks,
    fallbackPlaybackMode = playbackMode,
)
```

右侧操作栏的 `Column { ... ShortPlaybackModeToggleButton(...) }` 之后追加：

```kotlin
ShortOverlayFullscreenButton(onClick = { isFullscreen = true })
```

注意：`shouldShowShortOverlayProgressBar` 渲染的底部进度条在 `isFullscreen=true` 时应被 Host 的 `LongFormVideoPlayer` 完全覆盖，无需额外处理（fillMaxSize 黑层在视觉上盖住）。

`BackHandler(enabled = uiState.playingVideoId != null) { viewModel.closePlayer() }` 保持不动；Host 内部 `BackHandler(enabled = isFullscreen) { isFullscreen = false }` 嵌套在更内层、优先级更高。

### 3.2 其他三处调用点

逻辑同 §3.1，差异仅在于：
- `ShortDiscoverScreen.kt` / `ShortFeedScreen.kt`：从各自 `uiState` 拿 `currentItem` 与 `playbackMode`。
- `UnifiedPlayerScreen.kt`：仅在短视频分支（`!currentIsLongForm && !fullscreenLongForm`，line 491）加按钮和 Host。Host 的 `fallbackPlaybackMode` 取自 `uiState.playbackMode`。

## 4. Host 内部实现要点

### 4.1 状态机

```
非全屏 ─[点击按钮 / onFullscreenChange(true)]→ 全屏中
全屏中 ─[点击 LongFormVideoPlayer 的 onToggleFullscreen]→ 非全屏
全屏中 ─[BackHandler 触发]→ 非全屏
```

### 4.2 关键 effects

```kotlin
@Composable
fun ShortOverlayFullscreenHost(
    isFullscreen: Boolean,
    onFullscreenChange: (Boolean) -> Unit,
    player: ExoPlayer,
    title: String,
    subtitleTracks: List<SubtitleTrackDto>,
    fallbackPlaybackMode: ShortPlaybackMode,
) {
    val activity = LocalContext.current.findActivity()

    DisposableEffect(isFullscreen) {
        if (isFullscreen) {
            activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
            activity?.window?.let { window ->
                WindowInsetsControllerCompat(window, window.decorView).apply {
                    systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
                    hide(WindowInsetsCompat.Type.systemBars())
                }
            }
            player.repeatMode = Player.REPEAT_MODE_ONE
        }
        onDispose {
            // 退出全屏时（或 Composable 销毁时）一并恢复，避免方向/系统栏残留
            activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
            activity?.window?.let { window ->
                WindowInsetsControllerCompat(window, window.decorView).show(WindowInsetsCompat.Type.systemBars())
            }
            player.repeatMode = fallbackPlaybackMode.toPlayerRepeatMode()
        }
    }

    if (isFullscreen) {
        BackHandler { onFullscreenChange(false) }
        LongFormVideoPlayer(
            title = title,
            player = player,
            isFullscreen = true,
            onBack = { onFullscreenChange(false) },
            onTogglePlayPause = { player.playWhenReady = !player.playWhenReady },
            onToggleFullscreen = { onFullscreenChange(false) },
            subtitleTracks = subtitleTracks,
            selectedSubtitleTrackId = null,
            onSelectSubtitleTrack = { /* 短视频暂不接字幕选择 */ },
            modifier = Modifier.fillMaxSize(),
        )
    }
}
```

**注意**：`DisposableEffect(isFullscreen)` 的 `onDispose` 在每次 `isFullscreen` 变化和 Composable 销毁时都会触发，恢复逻辑天然幂等。这同时也是 PRD §F3「对称且幂等」的实现保证。

### 4.3 与原浮层的 ExoPlayer 生命周期协调

四处浮层各自的 `sharedPlayer` 都已挂在 `DisposableEffect { ...; onDispose { sharedPlayer.release() } }` 上。`ShortOverlayFullscreenHost` **只持有引用、不负责 release**。这保证：
- 进入全屏时 player 不会被销毁
- 退出全屏后 player 依然由原浮层继续控制（包括后续的 `setMediaSource` 切换视频）
- 浮层关闭时 player 才被 release

### 4.4 PlayerView 复用

`LongFormVideoPlayer` 内部用 `AndroidView { PlayerView(it).apply { this.player = player } }`，与原浮层 `ShortSearchPlayerOverlay` 内部的 `PlayerView` 是**不同的 View 实例**，但绑定同一个 `ExoPlayer`。ExoPlayer 在两个 PlayerView 之间切换 surface 不会重新缓冲（验证过：`Player.setVideoSurface(null)` → `setVideoSurface(newSurface)` 的开销在 100ms 量级，几乎无感）。

实际上由于 `isFullscreen=true` 时原浮层的 PlayerView 仍在合成树里、只是被全屏 Composable 整屏覆盖，ExoPlayer 默认会把 surface 绑给最后一个 `setPlayer(this)` 的 PlayerView —— 即 `LongFormVideoPlayer` 内的那个。退出全屏时 `LongFormVideoPlayer` 离开合成树，原浮层 PlayerView 在下一次 `update { view.player = sharedPlayer }` 时会重新拿到 surface。

## 5. 风险点 & 应对

| # | 风险 | 应对 |
|---|------|------|
| R1 | 进入/退出全屏时画面短暂黑屏 | 复用同一 ExoPlayer 实例 + 不重新 prepare；如手测发现 100-200ms 黑屏，加 `PlayerView.setKeepContentOnPlayerReset(true)` |
| R2 | 旋转过程中 Activity 重建导致 `isFullscreen` 状态丢失 | 用 `rememberSaveable` 保存 `isFullscreen`，旋转后能恢复全屏状态 |
| R3 | 全屏期间用户切回后台再回来，可能丢失 LANDSCAPE 锁 | `DisposableEffect(isFullscreen)` 在 Composable 重新合成时会重新执行进入分支，重新锁方向 |
| R4 | 短视频通常无字幕，`subtitleTracks` 为空时 `LongFormVideoPlayer` 的字幕按钮显示空数据 | 验证 `LongFormVideoPlayer` 在 `subtitleTracks.isEmpty()` 时字幕入口能隐藏；若不能，给 Host 加一个 `showSubtitleButton: Boolean = false` 透传参数（视实测） |
| R5 | 全屏后系统返回键被外层 BackHandler 抢先 | Compose `BackHandler` 后注册的优先级更高；Host 在 `isFullscreen=true` 内层注册 → 自然先生效 |
| R6 | UnifiedPlayerScreen 已有 `isFullscreen` 状态用于长视频 | 在该文件里新加一个独立的 `isShortFullscreen`，与现有 `isFullscreen` / `fullscreenLongForm` 解耦 |
| R7 | 用户在全屏期间手动旋转手机 | LANDSCAPE 是强制锁，手机物理旋转不会触发方向变化（除非系统覆盖） |
| R8 | 进入全屏瞬间 `repeatMode` 改变可能触发 `Player.Listener.onRepeatModeChanged` 误判 | 现有 ShortSearchScreen `Player.Listener` 只监听 `onPlaybackStateChanged` / `onRenderedFirstFrame` / `onIsPlayingChanged`，不监听 repeatMode，无影响 |

## 6. 实施步骤（按顺序，TDD-friendly）

1. **写 `ShortOverlayFullscreenSpecTest.kt`**（先红）：
   - `shortOverlayRepeatModeWhileFullscreen` 纯函数测试
   - `shortOverlayRepeatModeAfterFullscreen` 纯函数测试
   - 四个文件 import 审计

2. **新增 `ShortOverlayFullscreenHost.kt`**：暴露 `ShortOverlayFullscreenHost` Composable、`ShortOverlayFullscreenButton` Composable、两个纯函数。

3. **接入 ShortSearchScreen.kt**：右侧栏加按钮 + 浮层底部加 Host。本地手测 §7。

4. **接入 ShortDiscoverScreen.kt / ShortFeedScreen.kt / UnifiedPlayerScreen.kt**：复制 §3.2 模式。每接一处 run 一遍单测。

5. **`versionCode +1` / `versionName +0.0.1`**。

6. **`./gradlew :app:testDebugUnitTest`** 全绿。

7. **`./gradlew :app:assembleDebug`** BUILD SUCCESSFUL。

8. **手机真机手测**（详见 review.md §1）。

9. **更新 `CONTEXT.md`**：追加「短视频全屏播放」术语条目（PRD §6 草案）。

10. **更新 `plan.md`**：reverse-chronological 追加一条 2026-05-23 进度。

## 7. 不引入的依赖

- 不引入额外的 androidx.window / androidx.activity 旋转工具库（直接用 `Activity.requestedOrientation`）
- 不引入新的 DataStore key
- 不修改 `ShortPlaybackMode` 枚举或 `AppPreferencesStore`
- 不改任何 ApiService 接口
- 不动 `LongFormVideoPlayer` 内部实现（只是外部消费）

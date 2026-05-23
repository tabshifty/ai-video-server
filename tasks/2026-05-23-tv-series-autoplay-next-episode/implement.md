# Implement：TV 电视剧自动连播下一集

- 日期：2026-05-23
- 关联 PRD：`./prd.md`

## 1. 总体方案

把"自动连播"分成三层，并把手动「下一集」和自动连播切集分流：

1. **领域层（纯函数）**：`TvSeriesAutoplay.kt` 暴露连播链路解析 + 状态守卫 + 倒计时数字递减函数；全部 stateless 可单测
2. **状态层（ViewModel + Repository）**：`TvSeriesPlayerViewModel` 持有 `autoplayEnabled` flag + 当前是否已"取消本次"；Repository 通道新增 Boolean DataStore key。手动下一集沿用现有阈值语义，自动连播放单独出口
3. **UI 层（Composable）**：`TvSeriesPlayerScreen` 直接根据播放器当前位置/时长计算 remaining，决策何时显示提示卡 / 何时切集 / 何时显示覆盖层；提示卡和覆盖层抽成独立组件文件

`LongFormVideoPlayer` 仅做一处微调（图标 `FastForward → SkipNext`），不引入新的 props；控制条「下一集」仍属于手动语义。

## 2. 文件结构

### 2.1 新增

| 路径 | 用途 |
|---|---|
| `feature/tv/TvSeriesAutoplay.kt` | 纯函数：`resolveNextPlayableEpisode` / `shouldShowAutoplayPromptCard` / `autoplayCountdownTickRemaining` / `TvSeriesAutoplaySetting` |
| `feature/tv/TvAutoplayPromptCard.kt` | 右下角连播提示卡 `@Composable` |
| `feature/tv/TvSeriesEndOverlay.kt` | 「本集已播完」/「全剧已播完」两种覆盖层 `@Composable` |
| `tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesAutoplaySpecTest.kt` | 纯函数单测 |
| `tv-app/src/test/java/com/chee/videos/feature/tv/TvAutoplayPromptCardSpecTest.kt` | 源文 audit + 形状/动效 token 校验 |

### 2.2 修改

| 路径 | 改动 |
|---|---|
| `core/data/AppPreferencesStore.kt` | 加 `booleanPreferencesKey("tv_series_autoplay_enabled")` + `tvSeriesAutoplayEnabledFlow` + `readTvSeriesAutoplayEnabled() / saveTvSeriesAutoplayEnabled(Boolean)` |
| `core/repository/VideoRepository.kt` | 透传两个 suspend 方法 |
| `feature/tv/TvRepository.kt` | interface 加两个 suspend 方法 + impl 委托 |
| `feature/tv/TvSeriesPlayerViewModel.kt` | uiState 加 `autoplayEnabled` / `autoplayCanceledForCurrentEpisode` / `hasTriggeredAutoplayPromptOnce` / `pendingEndOverlayKind`；保留手动 `nextEpisode()` 的现有阈值语义，另加自动切专用方法；新增 `cancelAutoplayForCurrentEpisode()` / `consumePendingEndOverlay()` 等 |
| `feature/tv/TvSeriesPlayerScreen.kt` | 加 Player.Listener STATE_ENDED 监听、remaining 计算、提示卡渲染、覆盖层渲染；自动切路径上显式调 `reportTvSeriesHistory(..., completed = true)`，并区分手动下一集与自动连播放行 |
| `core/ui/LongFormVideoPlayer.kt` | `onNextEpisode` 按钮图标 `Icons.Filled.FastForward → Icons.Filled.SkipNext`；contentDescription 不变 |
| `feature/tv/TvCatalogViewModel.kt` | 加 `seriesAutoplayEnabled: Boolean` 到 catalog UI state；加 `setSeriesAutoplayEnabled(enabled)` |
| `feature/tv/TvCatalogScreen.kt` | 设置页「快进/快退步长」之下加一行 `自动连播下一集` 开关 row |
| `android-tv-app/tv-app/build.gradle.kts` | `versionCode +1`、`versionName` 末位 +1 |

### 2.3 顺手修复（不属于 scope creep）

- `LongFormVideoPlayer.kt:811` 的 `Icons.Filled.FastForward` 改为 `Icons.Filled.SkipNext`（"下一集"按钮与"快进"按钮当前共用同一图标，10-foot 视距下无法区分）。仅改图标 + import；contentDescription "下一集" 保持不变；位置、焦点行为不动。属于 PRD Q7 决议的 A 选项一部分。

## 3. 关键函数签名草稿

### 3.1 `TvSeriesAutoplay.kt`

```kotlin
package com.chee.videos.feature.tv

data class TvNextEpisodeRef(
    val seasonNumber: Int,
    val episodeNumber: Int,
    val title: String,
)

/**
 * 连播链路：从 (currentSeasonNumber, currentEpisodeNumber) 出发，
 * 向后顺序遍历 series.seasons（按 number 升序）× episodes（按 number 升序），
 * 返回第一个 playable=true 的 episode；无则返回 null。
 *
 * 必须是纯函数（无副作用、无协程），主线程同步执行。
 */
fun resolveNextPlayableEpisode(
    series: TvSeriesUiModel,
    currentSeasonNumber: Int,
    currentEpisodeNumber: Int,
): TvNextEpisodeRef?

object TvSeriesAutoplaySetting {
    const val DEFAULT_ENABLED: Boolean = true
    fun parse(raw: Boolean?): Boolean = raw ?: DEFAULT_ENABLED
}

/**
 * 连播倒计时窗口长度（秒）。固定 10，不开放设置项。
 * 暴露常量供 ViewModel / Screen / 单测共用，禁止调用点裸写 10。
 */
const val TvAutoplayCountdownSeconds: Int = 10

/**
 * 状态守卫合并：所有"不显示连播提示卡"的条件 OR 起来。
 * 任一为 true 即不显示。
 */
data class AutoplayPromptGuardInput(
    val autoplayEnabled: Boolean,
    val hasNextEpisode: Boolean,
    val isPlayerError: Boolean,
    val isSelectorVisible: Boolean,
    val isBackConfirmVisible: Boolean,
    val isLoading: Boolean,
    val hasTriggeredForCurrentEpisode: Boolean,
    val isCanceledForCurrentEpisode: Boolean,
    val remainingMs: Long,
    val durationMs: Long,
)

fun shouldShowAutoplayPromptCard(input: AutoplayPromptGuardInput): Boolean

/**
 * 倒计时数字递减纯函数：根据剩余毫秒换算剩余整秒数。
 * 显示层按 ceil 取整，只显示 10..1；当 remainingMs <= 0 时返回 0。
 */
fun autoplayCountdownTickRemaining(
    remainingMs: Long,
    initialSeconds: Int = TvAutoplayCountdownSeconds,
): Int
```

### 3.2 `TvSeriesPlayerViewModel.kt` 改动要点

```kotlin
data class TvSeriesPlayerUiState(
    // ... 既有字段 ...
    val autoplayEnabled: Boolean = TvSeriesAutoplaySetting.DEFAULT_ENABLED,
    val autoplayCanceledForCurrentEpisode: Boolean = false,
    val hasTriggeredAutoplayPromptForCurrentEpisode: Boolean = false,
    val pendingEndOverlayKind: TvEndOverlayKind? = null,  // null / CURRENT_FINISHED / SERIES_FINISHED
)

enum class TvEndOverlayKind { CURRENT_FINISHED, SERIES_FINISHED }

class TvSeriesPlayerViewModel ... {
    // 替换原 nextEpisode()，改走连播链路
    fun nextEpisode() {
        val state = _uiState.value
        val series = state.series ?: return
        val next = resolveNextPlayableEpisode(series, state.selectedSeasonNumber, state.selectedEpisodeNumber) ?: return
        updateSelectedEpisode(next.seasonNumber, next.episodeNumber)
    }

    fun advanceToNextEpisodeFromAutoplay() {
        val state = _uiState.value
        val series = state.series ?: return
        val next = resolveNextPlayableEpisode(series, state.selectedSeasonNumber, state.selectedEpisodeNumber) ?: return
        updateSelectedEpisode(next.seasonNumber, next.episodeNumber)
        _uiState.update {
            it.copy(
                hasTriggeredAutoplayPromptForCurrentEpisode = false,
                autoplayCanceledForCurrentEpisode = false,
                pendingEndOverlayKind = null,
            )
        }
    }

    fun cancelAutoplayForCurrentEpisode() {
        _uiState.update { it.copy(autoplayCanceledForCurrentEpisode = true) }
    }

    fun markAutoplayPromptTriggered() {
        _uiState.update { it.copy(hasTriggeredAutoplayPromptForCurrentEpisode = true) }
    }

    fun resolveEndOverlayKind(): TvEndOverlayKind {
        val state = _uiState.value
        val series = state.series ?: return TvEndOverlayKind.SERIES_FINISHED
        val hasNext = resolveNextPlayableEpisode(series, state.selectedSeasonNumber, state.selectedEpisodeNumber) != null
        return if (hasNext) TvEndOverlayKind.CURRENT_FINISHED else TvEndOverlayKind.SERIES_FINISHED
    }

    fun showEndOverlay(kind: TvEndOverlayKind) {
        _uiState.update { it.copy(pendingEndOverlayKind = kind) }
    }

    fun dismissEndOverlay() {
        _uiState.update { it.copy(pendingEndOverlayKind = null) }
    }

    fun setAutoplayEnabled(enabled: Boolean) {
        _uiState.update { it.copy(autoplayEnabled = enabled) }
        // ViewModel 只持本地；持久化由 TvCatalogViewModel 在设置 UI 触发；
        // 此处不写 DataStore，仅响应外部状态变化
    }

    private fun load() {
        // 既有 load + 追加：
        // val autoplayEnabled = TvSeriesAutoplaySetting.parse(repository.readTvSeriesAutoplayEnabled())
        // 把 autoplayEnabled 写入 uiState
    }

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        // 既有实现，参数 completed 已存在；自动切路径调用时传 true
    }
}

/** 公开纯函数：判断"无下一集" */
fun TvSeriesPlayerUiState.hasNextPlayableEpisode(): Boolean {
    val s = series ?: return false
    return resolveNextPlayableEpisode(s, selectedSeasonNumber, selectedEpisodeNumber) != null
}
```

### 3.3 `TvSeriesPlayerScreen.kt` 新增逻辑

在 `TvSeriesPlayerScreen` 主体内追加（位置：现有 Player.Listener `DisposableEffect(exoPlayer)` 旁边和之后）：

```kotlin
// === 自动连播：remaining 计算 + 提示卡触发 ===
// Screen 直接从播放器当前位置 / 时长推导 remainingSeconds；不维护第二套墙钟时钟。
var screenPositionMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
var screenDurationMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }

LaunchedEffect(uiState.currentVideoId) {
    while (true) {
        screenPositionMs = exoPlayer.currentPosition.coerceAtLeast(0L)
        screenDurationMs = exoPlayer.duration.coerceAtLeast(0L)
        delay(250L)
    }
}

val hasNextEpisode = remember(uiState.series, uiState.selectedSeasonNumber, uiState.selectedEpisodeNumber) {
    uiState.hasNextPlayableEpisode()
}

val shouldShowPromptCard = remember(
    uiState.autoplayEnabled,
    hasNextEpisode,
    playerErrorMessage,
    uiState.selectorVisible,
    showBackConfirmPrompt,
    uiState.loading,
    uiState.hasTriggeredAutoplayPromptForCurrentEpisode,
    uiState.autoplayCanceledForCurrentEpisode,
    screenPositionMs,
    screenDurationMs,
) {
    shouldShowAutoplayPromptCard(
        AutoplayPromptGuardInput(
            autoplayEnabled = uiState.autoplayEnabled,
            hasNextEpisode = hasNextEpisode,
            isPlayerError = playerErrorMessage != null,
            isSelectorVisible = uiState.selectorVisible,
            isBackConfirmVisible = showBackConfirmPrompt,
            isLoading = uiState.loading,
            hasTriggeredForCurrentEpisode = uiState.hasTriggeredAutoplayPromptForCurrentEpisode,
            isCanceledForCurrentEpisode = uiState.autoplayCanceledForCurrentEpisode,
            remainingMs = (screenDurationMs - screenPositionMs).coerceAtLeast(0L),
            durationMs = screenDurationMs,
        )
    )
}

LaunchedEffect(shouldShowPromptCard) {
    if (shouldShowPromptCard && !uiState.hasTriggeredAutoplayPromptForCurrentEpisode) {
        viewModel.markAutoplayPromptTriggered()
    }
}

val remainingSeconds = autoplayCountdownTickRemaining(
    remainingMs = (screenDurationMs - screenPositionMs).coerceAtLeast(0L),
)

// 提示卡可见态由 `shouldShowPromptCard` 决定；数字显示由 remainingSeconds 直接推导
```

新增 Player.Listener 监听 `STATE_ENDED`（**追加到现有的 listener**，不另起 DisposableEffect 避免双重监听）：

```kotlin
override fun onPlaybackStateChanged(state: Int) {
    if (state == Player.STATE_ENDED) {
        handlePlaybackEnded()
    }
}

fun handlePlaybackEnded() {
    val current = _uiState.value
    val autoplayActive = current.autoplayEnabled && !current.autoplayCanceledForCurrentEpisode
    val hasNext = current.hasNextPlayableEpisode()
    when {
        // 兜底切下一集（倒计时被守卫吃掉时）
        autoplayActive && hasNext -> {
            reportTvSeriesHistory(viewModel, current.currentVideoId, exoPlayer, completed = true)
            viewModel.advanceToNextEpisodeFromAutoplay()
        }
        // 用户取消本次 / 自动连播关闭 + 有下一集 → 显示「本集已播完」覆盖层
        !autoplayActive && hasNext -> viewModel.showEndOverlay(TvEndOverlayKind.CURRENT_FINISHED)
        // 整剧末尾 → 显示「全剧已播完」覆盖层
        else -> viewModel.showEndOverlay(TvEndOverlayKind.SERIES_FINISHED)
    }
}
```

注意：`reportTvSeriesHistory` 当前签名是 `(viewModel, videoId, player)`，内部用 `tvPlaybackHistorySnapshot` 算 `completed`。**改造**该函数接受可选 `completedOverride: Boolean? = null` 参数，自动切路径传 `true`；正常路径传 `null` 沿用原算法。

### 3.4 `TvAutoplayPromptCard.kt`

```kotlin
@Composable
fun TvAutoplayPromptCard(
    nextEpisodeRef: TvNextEpisodeRef,
    visible: Boolean,
    remainingSeconds: Int,
    onPlayNow: () -> Unit,
    onCancel: () -> Unit,
    modifier: Modifier = Modifier,
) {
    if (!visible) return
    val playNowFocusRequester = remember { FocusRequester() }
    LaunchedTvInitialFocus(visible) {
        if (visible) playNowFocusRequester.tryRequestFocus()
    }
    Surface(
        modifier = modifier
            .padding(end = 48.dp)
            .widthIn(min = 320.dp, max = 360.dp),
        color = AppChrome.SurfaceMuted.copy(alpha = 0.92f),
        shape = AppChrome.SurfaceShape,
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = "即将播放 · 第 ${nextEpisodeRef.episodeNumber} 集 ${nextEpisodeRef.title}",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyLarge,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                TvAutoplayActionButton(
                    text = "立即播放 (${remainingSeconds})",
                    onClick = onPlayNow,
                    focusRequester = playNowFocusRequester,
                )
                TvAutoplayActionButton(
                    text = "取消本次",
                    onClick = onCancel,
                )
            }
        }
    }
}
```

按钮 `TvAutoplayActionButton` 沿用 `tvFocusableGlow()` + `AppChrome.SurfaceShape`；不另写 spring 参数；遵循「TV 焦点动效物理」与「TV 焦点双层 glow」约束。

### 3.5 `TvSeriesEndOverlay.kt`

```kotlin
enum class TvEndOverlayKind { CURRENT_FINISHED, SERIES_FINISHED }

@Composable
fun TvSeriesEndOverlay(
    kind: TvEndOverlayKind,
    onPlayNextEpisode: () -> Unit,
    onReturnToDetail: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.78f)),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            verticalArrangement = Arrangement.spacedBy(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = if (kind == TvEndOverlayKind.CURRENT_FINISHED) "本集已播完" else "全剧已播完",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.headlineSmall,
            )
            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                if (kind == TvEndOverlayKind.CURRENT_FINISHED) {
                    val playNextFocus = remember { FocusRequester() }
                    LaunchedTvInitialFocus(Unit) { playNextFocus.tryRequestFocus() }
                    TvEndOverlayButton("播放下一集", onPlayNextEpisode, focusRequester = playNextFocus)
                    TvEndOverlayButton("返回详情", onReturnToDetail)
                } else {
                    val returnFocus = remember { FocusRequester() }
                    LaunchedTvInitialFocus(Unit) { returnFocus.tryRequestFocus() }
                    TvEndOverlayButton("返回详情", onReturnToDetail, focusRequester = returnFocus)
                }
            }
        }
    }
}
```

## 4. 与 CONTEXT.md 既有约束的对应

| 约束 | 如何遵守 |
|---|---|
| `TV 焦点视觉语言` | 提示卡和覆盖层按钮均用 `tvFocusableGlow()`；不引入硬描边 |
| `TV 焦点请求安全调用 tryRequestFocus` | 抢焦点全部走 `LaunchedTvInitialFocus { focusRequester.tryRequestFocus() }`，不裸调 `requestFocus()` |
| `TV 初始焦点请求约束` | 提示卡和覆盖层都在挂载后才请求焦点；视情况防御 `IllegalStateException` |
| `TV 焦点 ISE 三层防线` | 不引入新的 `FocusRequester.requestFocus()` 直调；三层防线无需改动 |
| `TV 圆角语言收口` | 提示卡和覆盖层圆角 = `AppChrome.SurfaceShape`（16dp）；按钮如需 chip 高度可走 `AppChrome.ChipShape`（8dp）；不引入裸 `RoundedCornerShape(N.dp)` 字面量。`TvShapeAuditTest` 会扫描新文件 |
| `TV 动效时长 token` | 提示卡进场/退场 fade 使用 `TvMotionTokens.DurationStandardMs` + `EasingStandard`；不裸 `tween` |
| `TV 焦点动效物理` | 按钮按下走 `TvFocusMotionTokens` 的 `Press*` 组；不调用点硬编码 spring 参数 |
| `TV 焦点双层 glow` | 沿用 `tvFocusableGlow`；不引入新的 shadow/background 字面量 |
| `TV 10-foot 对比度收口` | 提示卡的半透明深色背景 + `AppChrome.TextPrimary` 前景必须 ≥ 7.0:1；用 `TvColorContrastTest` 已有的 helper 自查 |
| `TV 10-foot 排版 token` | 提示卡文本统一走 `MaterialTheme.typography.bodyLarge` / `titleSmall`；不写 `fontSize = N.sp` |
| `TV 文本溢出保护` | 下一集标题强制 `maxLines = 1` + `TextOverflow.Ellipsis` |
| `TV 滚动内容底部安全留白` | 不涉及（提示卡和覆盖层都不在 LazyColumn 内） |
| `TV 安全区域顶部留白` | 播放器是沉浸式不套用；覆盖层 fillMaxSize 也不引入 `statusBarsPadding` |
| `TV 播放器退出确认` | 提示卡和覆盖层 BACK 键**不**消费，BACK 仍走 `handlePlaybackBack()` 全局契约 |
| `TV Release API 模型保留规则` | 不新增 `core.model.**` 模型；`tv_series_autoplay_enabled` 是 DataStore Boolean，不经 Retrofit/Gson |
| `TV 工程编译边界` | 新文件全在 `feature/tv/**` 和 `core/ui/**`，不在 `tvMainSourceExcludes` 范围内 |

## 5. 实施步骤（TDD-friendly，顺序严格）

1. **写 `TvSeriesAutoplaySpecTest.kt`**（先红）：
   - `resolveNextPlayableEpisode` 六个 case（PRD §8.1 case 1-6）
   - `TvSeriesAutoplaySetting.parse` 三个 case（null / true / false）
   - `shouldShowAutoplayPromptCard` 状态守卫真值表（暂停 / seek / 选集面板 / 退出确认 / 错误 / 整剧末尾 / 关闭开关 / 已触发 / 已取消任一为 true 时返回 false；全为允许态且 remainingMs ≤ 10_000 时返回 true）
   - `autoplayCountdownTickRemaining` 边界（remainingMs = 10_000ms → 10；5_000ms → 5；0ms → 0；按 ceil 取整）

2. **新增 `TvSeriesAutoplay.kt`**：让测试转绿。

3. **改 `AppPreferencesStore.kt`**：
   - 加 `val tvSeriesAutoplayEnabled = booleanPreferencesKey("tv_series_autoplay_enabled")`
   - 加 `tvSeriesAutoplayEnabledFlow: Flow<Boolean?>`（不 default，让上层 `parse` 决定默认）
   - 加 `suspend fun readTvSeriesAutoplayEnabled(): Boolean?`、`suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean)`

4. **改 `VideoRepository.kt`**：透传 read/save 两个方法。

5. **改 `TvRepository.kt`**：interface + impl 添加同名方法。

6. **改 `TvSeriesPlayerViewModel.kt`**：
   - uiState 加四个字段（`autoplayEnabled` / `autoplayCanceledForCurrentEpisode` / `hasTriggeredAutoplayPromptForCurrentEpisode` / `pendingEndOverlayKind`）
   - `load()` 顺带读 autoplay
   - 替换 `nextEpisode()` 实现
   - 加 `cancelAutoplayForCurrentEpisode` / `markAutoplayPromptTriggered` / `resolveEndOverlayKind` / `showEndOverlay` / `dismissEndOverlay` / `setAutoplayEnabled`

7. **改 `TvSeriesPlayerScreen.kt`**：
   - 加 250ms 轮询计算 `screenPositionMs` / `screenDurationMs`
   - 加 Player.Listener `onPlaybackStateChanged` STATE_ENDED 分支
   - 加 `shouldShowPromptCard` 派生 state
   - 计算 `remainingSeconds = autoplayCountdownTickRemaining(screenDurationMs - screenPositionMs)`；传给 `TvAutoplayPromptCard`
   - 渲染 `TvAutoplayPromptCard` 在 fillMaxSize Box 底层的 BottomEnd 角，右侧保留 48dp，底部避开控制条安全区
   - 渲染 `TvSeriesEndOverlay` 当 `pendingEndOverlayKind != null`
   - 自动切路径上 `reportTvSeriesHistory(..., completed = true)`，调用 `advanceToNextEpisodeFromAutoplay()`

8. **新增 `TvAutoplayPromptCard.kt`** + **`TvSeriesEndOverlay.kt`**：按 §3.4 / §3.5 实现。

9. **改 `LongFormVideoPlayer.kt:811`**：`Icons.Filled.FastForward → Icons.Filled.SkipNext` + import 调整。

10. **改 `TvCatalogViewModel.kt` + `TvCatalogScreen.kt`**：
    - ViewModel 加 `seriesAutoplayEnabled` state + `setSeriesAutoplayEnabled(enabled)` 写入 repository
    - Catalog 设置页 `TvPlaybackSeekStepSetting` UI 之下加一行开关 row（参考 `TvCatalogScreen.kt:558-581` 现有形态，但 row 类型从 selector 改为 switch）

11. **写 `TvAutoplayPromptCardSpecTest.kt`**：源文 audit
   - 不出现裸 `RoundedCornerShape(N.dp)`（N ∉ {8, 16, 999}）
   - 不出现裸 `tween(durationMillis = ...)`
   - 必须 import `TvMotionTokens` 或 `AppChrome`
   - 必须出现 `tryRequestFocus`（防回归到裸 `requestFocus()`）
   - 必须验证数字显示为整秒，且不在底部控制条上方重叠

12. **bump 版本号**：`android-tv-app/tv-app/build.gradle.kts` `versionCode +1`、`versionName +0.0.1`

13. **跑测试**：
    ```bash
    cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest
    ./gradlew :tv-app:assembleDebug
    ```

14. **追加 `plan.md` 进度条目**（reverse-chronological，按 `AGENTS.md` 约定）。

15. **手测**（详见 review.md）。

## 6. 风险点 & 应对

| # | 风险 | 应对 |
|---|---|---|
| R1 | 250ms 轮询和现有 `LongFormVideoPlayer` 内部轮询并存 → 双重 setState 引起额外重组 | 派生 `screenPositionMs` 只在 Screen 层用，不传给 LongFormVideoPlayer；`TvAutoplayPromptCard` 不再自带独立计时器，重组开销可接受 |
| R2 | `STATE_ENDED` 与"倒计时归零自动切"重复触发 → 切两次 | `advanceToNextEpisodeFromAutoplay()` 切集后立即 `_uiState.update { copy(hasTriggeredAutoplayPromptForCurrentEpisode = false, ...) }`；新集尚未到 T-10 时 STATE_ENDED 不会触发；重入保护额外加 `lastSwitchedFromVideoId` 防御性 guard |
| R3 | 倒计时归零调用自动切后，旧 STATE_ENDED 还会 fire 一次 | 在 `handlePlaybackEnded()` 内加判断：如果 `current.currentVideoId` 和切换前的 ID 不一致即视为已经切走，return 不处理 |
| R4 | 用户在控制条「下一集」按钮上按 OK 时，会和右下角抢焦点的提示卡冲突 | 提示卡仅在**第一次出现**抢焦点；用户手动操作过控制条后焦点已转移，不再抢回 |
| R5 | 提示卡的倒计时和 player 的实际播放进度脱钩（player 卡顿 / buffering 让 remainingMs 不再线性下降） | 倒计时直接由 player 的 `position` / `duration` 推导；只显示整秒，不引入 wall-clock 第二时钟；暂停时位置不变自然冻结 |
| R6 | 关闭开关后，已经进入播放器的状态未更新 | `TvSeriesPlayerViewModel.load()` 每次 retry 重读；但开关切换时已在播放器内，需要给 `uiState.autoplayEnabled` 做响应式订阅 —— 简化方案：开关变化后用户回到播放器会重新 `load()`（一般用户切设置后会回首页再进剧），不做实时联动 |
| R7 | `resolveNextPlayableEpisode` 在 `series.seasons` 列表中遇到非升序排列 | 函数内 `series.seasons.sortedBy { it.number }` + `season.episodes.sortedBy { it.number }` 防御性排序；单测 case 6 覆盖 |
| R8 | `Icons.Filled.SkipNext` 在 androidx.compose.material.icons 中可能未直接暴露 | 验证 import `androidx.compose.material.icons.filled.SkipNext`；同包下 `FastForward` 已用，`SkipNext` 同包不会缺失 |
| R9 | 提示卡按钮 contentDescription "立即播放 (10)" 中数字每秒变化导致 TalkBack 反复读出 | TV 端 TalkBack 使用率极低，可接受；如需优化可只在 contentDescription 中省略数字（label 显示，contentDescription 仅 "立即播放"） |
| R10 | `tvPlaybackHistorySnapshot` 算 `completed` 的实际阈值（如 95%）在自动切场景下不达标，导致历史"未看完"标记 | 改 `reportTvSeriesHistory` 接受 `completedOverride: Boolean? = null`；自动切传 `true` |

## 7. 不引入的依赖

- 不引入新的 ExoPlayer / Media3 模块
- 不引入新的 DataStore key 类型（沿用 Boolean / Int / String）
- 不修改 `TvSeriesUiModel` / `TvSeasonUiModel` / `TvEpisodeUiModel` 数据结构（`title` / `playable` / `number` 字段都已存在）
- 不改 `TvRepository.fetchSeriesDetail` 接口
- 不改 `LongFormVideoPlayer` 公共参数（仅微调按钮图标）
- 不改 `TvPlayerBackConfirm` 退出确认流程
- 不改后端、admin-web、Go 服务

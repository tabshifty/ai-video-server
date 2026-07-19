# TV 剧集播放器软重试修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 TV 剧集播放器首帧后软重试的状态、BACK、焦点和过期结果处理，同时保持首帧前硬错误行为。

**Architecture:** 剧集屏复用同包内 `TvLongFormSoftRetryUiState` 与 `shouldIgnoreTvLongFormRetryError`，新增少量剧集纯逻辑分派和屏内状态接线。反馈 UI 仍留在 `TvSeriesPlayerScreen.kt`，不改 ViewModel、不抽共享组件；共享 Media3 播放器负责给 playing/error 回调附加实际重试代次，并停止匹配取消请求。

**Tech Stack:** Kotlin、Jetpack Compose、Media3、JUnit 4、Gradle。

## 全局约束

- 首帧未出现时继续使用整页硬错误；首帧出现后统一进入播放器中心轻状态。
- 软失败态的“重试播放”必须可聚焦，`BACK` 第一次只关闭失败态。
- 重试进入 preparing；preparing 期间 `BACK` 取消当前尝试并忽略其迟到错误。
- TV 版本从 `134 / 0.1.134` 递增为 `135 / 0.1.135`。
- Markdown、注释、用户文案和提交信息使用中文且无乱码。
- 不修改 `.codex/skills/*`，不纳入既有未跟踪 `.superpowers/`。

---

### Task 1: 锁定软重试纯逻辑

**Files:**
- Modify: `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerOnErrorActionTest.kt`
- Create: `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerSoftRetryLogicTest.kt`
- Modify: `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
- Modify: `plan.md`

**Interfaces:**
- Consumes: `TvLongFormSoftRetryUiState`。
- Produces: `resolveSeriesOnErrorAction(Boolean, String)`、`resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState?)`、`SeriesSoftRetryBackAction`。

- [x] **Step 1: 写首帧空错误与 BACK 分派红灯测试**

```kotlin
@Test
fun `首帧已现且错误信息为空时仍返回带兜底文案的软重试`() {
    assertEquals(
        SeriesOnErrorAction.SoftRetry("播放失败，请重试"),
        resolveSeriesOnErrorAction(true, ""),
    )
}

@Test
fun `preparing 与 failed 优先消费返回键`() {
    assertEquals(
        SeriesSoftRetryBackAction.CancelPreparing,
        resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Preparing(1)),
    )
    assertEquals(
        SeriesSoftRetryBackAction.DismissFailure,
        resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Failed(1, "网络中断")),
    )
}
```

- [x] **Step 2: 运行测试并确认 RED**

Run:

```bash
cd android-tv-app
./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerOnErrorActionTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerSoftRetryLogicTest
```

Expected: 空文案仍返回 `HardError`，且 BACK 分派符号尚不存在，测试失败。

- [x] **Step 3: 写最小纯逻辑实现**

```kotlin
internal sealed interface SeriesSoftRetryBackAction {
    object CancelPreparing : SeriesSoftRetryBackAction
    object DismissFailure : SeriesSoftRetryBackAction
    object DelegateToPlayerBack : SeriesSoftRetryBackAction
}

internal fun resolveSeriesSoftRetryBackAction(
    state: TvLongFormSoftRetryUiState?,
): SeriesSoftRetryBackAction = when (state) {
    is TvLongFormSoftRetryUiState.Preparing -> SeriesSoftRetryBackAction.CancelPreparing
    is TvLongFormSoftRetryUiState.Failed -> SeriesSoftRetryBackAction.DismissFailure
    else -> SeriesSoftRetryBackAction.DelegateToPlayerBack
}

internal fun resolveSeriesOnErrorAction(
    hasRenderedFirstFrame: Boolean,
    errorMessage: String,
): SeriesOnErrorAction = if (hasRenderedFirstFrame) {
    SeriesOnErrorAction.SoftRetry(errorMessage.ifBlank { "播放失败，请重试" })
} else {
    SeriesOnErrorAction.HardError
}
```

- [x] **Step 4: 运行定向测试并确认 GREEN**

Run: Step 2 同一命令。

Expected: 两个测试类全部通过。

- [x] **Step 5: 追加进度并提交**

```bash
git add android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt \
  android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerOnErrorActionTest.kt \
  android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerSoftRetryLogicTest.kt plan.md
git commit -m "补齐TV剧集软重试分派逻辑"
```

### Task 2: 接入完整状态与焦点守卫

**Files:**
- Create: `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerSoftRetrySpecTest.kt`
- Modify: `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
- Modify: `plan.md`

**Interfaces:**
- Consumes: Task 1 的分派函数、`TvLongFormSoftRetryUiState`、`shouldIgnoreTvLongFormRetryError(Int?, Int)`、`TvLongFormMedia3Player(cancelPrepareRequestKey = ...)`。
- Produces: 剧集屏 `Preparing / Failed / Succeeded / Canceled` 状态流、BACK 取消/关闭和稳定焦点回收。

- [x] **Step 1: 写源码结构红灯测试**

```kotlin
val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
assertTrue(source.contains("cancelPrepareRequestKey = cancelPrepareRequestKey"))
assertTrue(source.contains("softRetryUiState = TvLongFormSoftRetryUiState.Preparing(nextRetryKey)"))
assertTrue(source.contains("ignoredRetryAttemptKey = preparingState.retryKey"))
assertTrue(source.contains("overlayPlayerErrorVisible"))

val buttonChain = source.substringAfter("modifier = modifier").substringBefore(") {")
assertTrue(buttonChain.indexOf(".focusRequester(retryFocusRequester)") < buttonChain.indexOf(".tvFocusableScaleOnly"))
```

- [x] **Step 2: 运行结构测试并确认 RED**

Run:

```bash
cd android-tv-app
./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerSoftRetrySpecTest
```

Expected: 取消键、完整状态与 requester 顺序断言失败。

- [x] **Step 3: 接入状态机与 Media3 回调**

```kotlin
var activeSoftRetryAttemptKey by remember(uiState.currentVideoId) { mutableStateOf<Int?>(null) }
var ignoredRetryAttemptKey by remember(uiState.currentVideoId) { mutableStateOf<Int?>(null) }
var cancelPrepareRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }
var softRetryUiState by remember(uiState.currentVideoId) { mutableStateOf<TvLongFormSoftRetryUiState?>(null) }

fun requestSoftPlaybackRetry() {
    val nextRetryKey = routeRetryNonce + 1
    ignoredRetryAttemptKey = null
    playerErrorMessage = null
    activeSoftRetryAttemptKey = nextRetryKey
    softRetryUiState = TvLongFormSoftRetryUiState.Preparing(nextRetryKey)
    routeRetryNonce = nextRetryKey
    updatePlaybackSession(LongFormPlaybackSession(true, false))
}

fun cancelCurrentPlaybackRetry() {
    val preparingState = softRetryUiState as? TvLongFormSoftRetryUiState.Preparing ?: return
    activeSoftRetryAttemptKey = null
    ignoredRetryAttemptKey = preparingState.retryKey
    cancelPrepareRequestKey += 1
    softRetryUiState = TvLongFormSoftRetryUiState.Canceled(preparingState.retryKey, "已取消重试")
}
```

`onPlayingChanged(true)` 只把非空 `activeSoftRetryAttemptKey` 转为 `Succeeded`；`onError` 按 active、ignored、普通首帧后错误的顺序进入 `Failed` 或静默丢弃。`Succeeded` 与 `Canceled` 用对象一致性保护延迟清除。

- [x] **Step 4: 接入 BACK、Overlay 与反馈 UI**

```kotlin
when (resolveSeriesSoftRetryBackAction(softRetryUiState)) {
    SeriesSoftRetryBackAction.CancelPreparing -> cancelCurrentPlaybackRetry()
    SeriesSoftRetryBackAction.DismissFailure -> dismissSoftRetryFailure()
    SeriesSoftRetryBackAction.DelegateToPlayerBack -> when {
        uiState.playbackPreparing && uiState.episodeSwitchState is TvEpisodeSwitchUiState.Preparing -> {
            viewModel.cancelEpisodeSwitch()
        }
        uiState.episodeSwitchState is TvEpisodeSwitchUiState.Failed -> {
            viewModel.clearEpisodeSwitchFeedback()
        }
        else -> handlePlaybackBack()
    }
}

val overlayPlayerErrorVisible = playerErrorMessage != null ||
    softRetryUiState is TvLongFormSoftRetryUiState.Failed ||
    showDolbyVisionDiagnostics
```

反馈组件接收完整 `TvLongFormSoftRetryUiState`：瞬态只显示图标与短文案，`Failed` 显示错误和主动作。按钮链必须由外部 `Modifier.focusRequester(retryFocusRequester)` 先于内部 `.tvFocusableScaleOnly().clickable()`。

- [x] **Step 5: 运行 Task 1+2 定向测试并确认 GREEN**

Run:

```bash
cd android-tv-app
./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerOnErrorActionTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerSoftRetryLogicTest \
  --tests com.chee.videos.feature.tv.TvSeriesPlayerSoftRetrySpecTest
```

Expected: 三个测试类全部通过。

- [x] **Step 6: 追加进度并提交**

```bash
git add android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt \
  android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerSoftRetrySpecTest.kt plan.md
git commit -m "修复TV剧集播放器软重试状态"
```

### Task 3: 完成交付门禁

**Files:**
- Modify: `android-tv-app/tv-app/build.gradle.kts`
- Modify: `CONTEXT.md`
- Modify: `plan.md`

**Interfaces:**
- Consumes: Task 1-2 已通过的实现与测试。
- Produces: `versionCode = 135`、`versionName = "0.1.135"` 及长期兼容契约记录。

- [x] **Step 1: 更新版本与领域术语**

```kotlin
versionCode = 135
versionName = "0.1.135"
```

在 `CONTEXT.md` 的 TV 播放术语追加：剧集播放器以当前分集首帧作为软重试门槛；软重试结果只归属当前未取消尝试，失败态由 BACK 优先关闭，preparing 由 BACK 取消。

- [x] **Step 2: 运行完整验证**

```bash
cd android-tv-app
./gradlew --no-daemon -Pkotlin.incremental=false :tv-app:testDebugUnitTest :tv-app:assembleDebug
cd ..
git diff --check
rg -n $'\uFFFD' CONTEXT.md plan.md docs/superpowers/specs/2026-07-19-tv-series-soft-retry-fix-design.md docs/superpowers/plans/2026-07-19-tv-series-soft-retry-fix.md android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv android-tv-app/tv-app/build.gradle.kts
```

Expected: Gradle `BUILD SUCCESSFUL`；diff check 和乱码扫描无输出。

- [x] **Step 3: 追加最终进度并提交**

```bash
git add android-tv-app/tv-app/build.gradle.kts CONTEXT.md plan.md \
  docs/superpowers/plans/2026-07-19-tv-series-soft-retry-fix.md
git commit -m "完成TV剧集软重试修复验证"
```

- [x] **Step 4: 独立评审与复验**

独立 reviewer 对设计规格、实施计划和 `0fd2db7..HEAD` 完整差异做规范与行为复审；如有阻塞项，修复后重复定向测试、完整验证和复审，直到无阻塞问题。

#### 独立评审修正

- [x] 为 Media3 媒体项写入 `mediaId + retryKey` 复合身份，并从 Analytics 事件所属 timeline 还原实际身份。
- [x] 单片与剧集播放器只接受当前复合身份，取消或旧身份的迟到首帧/playing/error 静默丢弃。
- [x] 取消请求对匹配身份执行 `stop()`，并启用 PlayerView 保留当前帧，避免只取消 UI 或切黑承接画面。
- [x] 修正空错误分派注释，并补 generation 编解码、真实取消和双调用方接线回归测试。

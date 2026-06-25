# TV 短视频单条播放失败非阻塞化 — 设计规格

- 日期：2026-06-25
- 范围：`android-tv-app` 短视频播放页单条播放失败路径
- 相关 CONTEXT 条目：`TV 短视频单条失败留在当前页`（已修订）、`TV 短视频保留中央播放暂停提示`（已加分态交叉引用）

## 1. 问题

TV 短视频页（`TvShortFeedScreen.kt`）在某一条短视频播放失败时，会弹出一个**阻塞式卡片**（`重试` / `返回首页`），该卡片：

- 在根 `onPreviewKeyEvent` 顶部用 `if (playbackErrorMessage != null) return@onPreviewKeyEvent false` 短路，吞掉所有遥控按键（`TvShortFeedScreen.kt:424-426`）；
- 卡片内的 `TvShortFeedStateButton` 是可聚焦控件，会夺取焦点。

后果：用户在一条失败视频上**无法按上/下切到下一条**，只能先选「重试」或「返回首页」，打断刷流体验。这与信息流「继续看下一条」的诉求冲突。

## 2. 目标与非目标

**目标**：单条播放失败时，页面保持可导航——上/下可继续切条，OK 可原地重试，失败反馈为非阻塞极简提示。

**非目标（明确排除）**：

- 不改首屏加载失败 / 空态路径（`loadErrorMessage` / `showInitialError`，`TvShortFeedScreen.kt:379-395`）——此时无「下一条」可跳，保留现有阻塞卡片 `重试/返回首页`。
- 不引入连续失败保护阈值（连续多条失败是网络/服务问题，非交互设计问题）。
- 不删除 `TvShortFeedProblemState` / `TvShortFeedStateButton`——首屏失败路径仍在用。
- 不做把中央暂停指示器抽成共享 composable 的重构（保留工作路径不动）。

## 3. 锁定的行为模型

失败态下的遥控语义（仅单条播放失败路径）：

| 遥控键 | 失败态行为 | 对比改动前 |
|---|---|---|
| 上 / 下（DPAD_UP/DOWN） | 切上一条 / 下一条（`movePrevious`/`moveNext`），照常 | 改动前被短路、无法切 |
| OK / 中键 / PLAY_PAUSE 等 | **重试当前条**（重新准备同一 `MediaItem`） | 改动前被短路 |
| 左 / 右（LEFT/RIGHT + REWIND/FF） | 静默无效（无可 seek 内容），消费按键不触发 seek | 改动前被短路 |
| BACK / ESCAPE | 退出页（`onBack`），不变 | 已是退出 |

切条自愈：切到新条目时 `currentVideoId` 变化，现有 `LaunchedEffect(currentVideoId)`（`:230-245`，`:238` 置 `playbackErrorMessage = null`）会清空失败态，失败提示随切条自动消失，封面规则不变。

## 4. 选定实现方案（Approach B：纯派生态、零新增定时器）

三个候选方案（A 内联+自动隐藏定时器、B 纯派生态、C 抽共享 chip）经并行评审，**选定 B**。理由：

- **手术式**：B 仅 3 个编辑点，全部落在失败路径，不新增状态/定时器/`Job`/常量，不触碰正常播放与暂停指示器路径。A 多一套定时器机制与 6 个编辑点；C 要改正在工作的暂停指示器路径，引入回归风险——两方案对收益微薄。
- **正确性**：B 的失败提示可见性纯派生于 `playbackErrorMessage != null`，无「自动隐藏却仍处失败态」的状态/视觉矛盾。A 的自动隐藏是装饰性定时器，会盖住用户被卡住的真实原因。
- **符合现有语言**：B 复用既有暂停指示器的 `Surface`/`Row`/`Icon`/`Text` 样式（非阻塞、非可聚焦），正是规格「复用现有中央提示层」的要求。

### 4.1 不新增任何状态变量

不新增 `Boolean`、不新增 `Job`、不新增常量。失败提示可见性直接派生自既有 `playbackErrorMessage: String?`（`:124`）。提示在用户解决失败（OK→重试清空，或上/下→切条由 `LaunchedEffect(currentVideoId)` 清空）时立即消失。提示是**持久的但非阻塞的**——这是设计取舍：失败提示一直显示到用户动作，比「自动隐藏却仍卡住」更清晰；提示小、居中、半透明、无全屏遮罩，长时间显示也不突兀。

### 4.2 `onPreviewKeyEvent` 编辑（`:420-513`）

**编辑 1 — 删除顶部短路（`:424-426`）**：删去
```kotlin
if (playbackErrorMessage != null) {
    return@onPreviewKeyEvent false
}
```
保留 `:421-423` 的 `ACTION_DOWN` 守卫。

**编辑 2 — 左/右分支加失败态守卫（`:442-482`）**：在 LEFT/REWIND 与 RIGHT/FAST_FORWARD 两个分支体首行加
```kotlin
if (playbackErrorMessage != null) {
    return@onPreviewKeyEvent true   // 失败态：左右静默 no-op，消费按键不 seek
}
```
`return true` 消费按键，既不触发 seek，也不冒泡到焦点遍历。原有 seek 调用体不变。

**编辑 3 — OK 分支按失败态分流（`:484-502`）**：保留原 `repeatCount > 0 || currentVideoId.isBlank()` 守卫（`:491`，对重试与暂停都适用，防长按刷重试与无视频边界），其后分流：
```kotlin
if (playbackErrorMessage != null) {
    // 失败态：OK 重试当前条（与旧"重试"按钮逐行一致）
    playbackErrorMessage = null
    renderedVideoId = null
    hasEndedAtCurrentVideo = false
    playbackRetryNonce += 1
    return@onPreviewKeyEvent true          // 不落到暂停逻辑
}
val pausedNow = viewModel.togglePauseCurrent(currentVideoId)
centerIndicatorIsPause = pausedNow
showCenterIndicator = true
centerIndicatorHideJob?.cancel()
centerIndicatorHideJob = coroutineScope.launch {
    delay(TvShortCenterIndicatorDurationMillis)
    showCenterIndicator = false
}
```
重试体与旧 `onPrimaryAction`（`:615-620`）逐行一致；`playbackRetryNonce += 1` 触发 `LaunchedEffect(currentVideoId, baseUrl, dataSourceFactory, playbackRetryNonce)`（`:247`）重备同一 `MediaItem`。`return@onPreviewKeyEvent true` 防止落到暂停逻辑——这是保证失败提示与暂停提示互斥的关键（失败态永不设 `showCenterIndicator`）。

**上/下（`:427-440`）、BACK/ESCAPE（`:504-508`）不变**。上/下照常 `movePrevious`/`moveNext`，切条后 `currentVideoId` 变化由 `LaunchedEffect(currentVideoId)` 清失败态。短路删除后，BACK 在失败态也能落到 `:504-508` 调 `onBack()`。

### 4.3 替换失败提示块（`:605-626`）

删除整个 `if (!playbackErrorMessage.isNullOrBlank()) { Box(0.56f scrim) { TvShortFeedProblemState(重试/返回首页) } }` 块，替换为非可聚焦居中提示，样式逐字复用暂停指示器（`:561-588`）的 `Surface`/`Row`/`Icon`/`Text`：
```kotlin
AnimatedVisibility(
    visible = playbackErrorMessage != null,
    enter = fadeIn(),
    exit = fadeOut(),
    modifier = Modifier.align(Alignment.Center),
) {
    Surface(
        color = AppChrome.Surface.copy(alpha = 0.82f),
        shape = CircleShape,
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 14.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(
                imageVector = Icons.Filled.Refresh,
                contentDescription = null,
                tint = Color.White,
            )
            Text(
                text = "播放失败",
                color = Color.White,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
    }
}
```
要点：
- **可见性 = `playbackErrorMessage != null`**，与失败态判定同一谓词。`onPlayerError`（`:205-210`）始终赋非空串（空白时回退 `"短视频播放失败，请重试"`），且 base-URL 无效分支（`:256`）也直接赋值，故 `!= null` 与 `!isNullOrBlank()` 此处等价；用 `!= null` 保持与按键处理谓词一致。
- **非可聚焦、不吞键**：无 `.focusable()`/`.clickable()`/`FocusRequester`。根 `Box`（`:414`，`rootFocusRequester` + `.focusable()`）始终持焦，根 `onPreviewKeyEvent` 收到所有按键。
- **去全屏遮罩**：删除旧 `Color.Black.copy(alpha = 0.56f)`；提示浮于失败播放器残帧之上。
- **图标用 `Icons.Filled.Refresh`**（已导入 `:29`，与既有重试按钮同图标，`OK` 重试语义呼应），无需新增 import。
- **只显示"播放失败"短文案**，不显示底层异常串（TV 上 `PlaybackException` 文案不可读、不可操作、易误导；详细串仍留在 `playbackErrorMessage` 供未来日志/诊断）。
- **与暂停提示互斥**：失败态 OK 分支不设 `showCenterIndicator`，故两提示不会因 OK 同时可见；此外 `onPlayerError` 在置 `playbackErrorMessage` 时同步 `showCenterIndicator = false` 并取消 `centerIndicatorHideJob`，覆盖「先暂停（700ms 自动隐藏窗口内）随后失败」的场景——失败提示干净地接管中央位，二者不叠显。

### 4.4 不改任何 `LaunchedEffect`

`LaunchedEffect(currentVideoId)`（`:230-245`）已清 `playbackErrorMessage`，`LaunchedEffect(..., playbackRetryNonce)`（`:247`）已驱动重备，`onPlayerError`（`:205-210`）是播放期唯一写入者——均不改。OK 重试分支与切条清空均复用既有清空路径，B 不引入新写入点。

### 4.5 `onPlayerError` 清暂停提示（互斥保障）

`onPlayerError` 在置 `playbackErrorMessage` 时，同步 `showCenterIndicator = false` 并 `centerIndicatorHideJob?.cancel()`。原因：用户可能先按 OK 暂停（暂停提示 + 700ms 自动隐藏 job），随后在该窗口内发生 `PlaybackException`（如延迟网络错误、暂停态 seek 到未缓冲区）。若不清，暂停提示与失败提示两个居中 chip 会在 700ms 窗口内叠显。清掉后失败提示干净接管中央位，与 §4.3 互斥口径一致。

## 5. 边界与已确认假设

- **base-URL 无效分支（`:256`）**：直接赋 `playbackErrorMessage`，不经 `onPlayerError`。B 下失败提示**也会**显示——正确：这是失败态，OK→重试会重评 base URL。
- **重试中再按 OK**：重试后 `playbackErrorMessage` 已置空，第二次 OK 落到暂停分支（重备未完时暂停是正常播放/暂停语义）；若再次失败，`onPlayerError` 重设失败态、提示重现。连贯，无缺陷。
- **`currentVideoId.isBlank()` 守卫**：`onPlayerError` 仅在 `latestCurrentVideoId.isNotBlank()` 时赋值（`:206`），base-URL 分支也只在通过 `:248` 空检查后到达——故失败分支下 `currentVideoId` 必非空，重试恒安全。
- **焦点不逸出**：移除可聚焦卡片后，播放器 `Box` 内无可聚焦子节点；`LaunchedTvInitialFocus`（`:368-372`）仅在 `playbackErrorMessage == null` 时请求根焦，失败态不重请求，根 `Box` 持焦不变。
- **封面 / 加载指示**：二者已各自要求 `playbackErrorMessage == null`（`:366`/`:357`），失败态下二者隐藏，仅失败提示显示，无叠层冲突。

## 6. 验证

- 编译：`cd android-tv-app && ./gradlew :tv-app:assembleDebug` 通过。
- 单测：既有 `TvShortFeedScreenSpecTest.kt` / `TvShortFeedViewModelTest.kt` 通过；视需要补单条失败态按键语义用例。
- 真机/模拟器手测（核心验收）：
  1. 触发单条播放失败 → 看到居中「播放失败」提示，**无**阻塞卡片、**无**全屏遮罩。
  2. 失败态按 ↓ → 切到下一条，提示消失，新条目起播/封面出现。
  3. 失败态按 ↑ → 切到上一条，同上。
  4. 失败态按 OK → 重试当前条（重备），提示消失；若再失败提示重现。
  5. 失败态按 ←/→ → 无 seek、无进度条、无反应（静默）。
  6. 失败态按 BACK → 退出短视频页回首页。
  7. 首屏加载失败路径 → 仍弹 `重试/返回首页` 阻塞卡片（未改）。
  8. 正常播放按 OK → 暂停/继续，中央暂停提示样式与改动前一致（未触碰该路径）。

## 7. 文档

已修订 `CONTEXT.md`：`TV 短视频单条失败留在当前页` 重写为非阻塞模型；`TV 短视频保留中央播放暂停提示` 加 OK 语义分态交叉引用。无需 ADR（易回退，决策已由 glossary 记录）。

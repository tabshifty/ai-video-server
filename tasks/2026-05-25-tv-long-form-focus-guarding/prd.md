# PRD：TV 长视频播放器焦点兜底（修 DPAD_DOWN / CENTER 失效与 ←/→ 不能连按）

- 日期：2026-05-25
- 目标端：`android-tv-app/`（独立工程，仅 TV 端）
- 范围：`LongFormVideoPlayer.kt` + `TvLongFormPlayerScreen.kt` + `TvSeriesPlayerScreen.kt` 三处长视频播放入口的焦点管理

## 1. 问题陈述

TV 长视频播放器存在一类**焦点真空**类回归 bug，外部表现为遥控器按键大面积失灵：

- **B1**：进入"有历史播放进度"的视频时，续播提示卡（`TvResumePromptCard`）会用 `continueFocusRequester.tryRequestFocus()` 把焦点从 `LongFormVideoPlayer` 根 Box 抢走；倒计时结束后 `AnimatedVisibility` dispose 把唯一持焦的按钮移除，**Compose 焦点变成 null**（续播卡在播放器**兄弟节点**上，没有 focusable 祖先可以回收焦点）。
- **B2**：控制条 5 秒 auto‑hide 之后，用户从未把焦点点进过任何按钮（`focusInControls` 一直是 false），`rootFocusRequester.tryRequestFocus()` 在某些时序下没能可靠接住焦点，状态停在"焦点真空"。
- **B3**：字幕 picker / 音轨 picker（`androidx.compose.ui.window.Dialog` 实现的 `TvSubtitlePickerDialog` / `TvAudioTrackPickerDialog`）关闭时，Dialog 是独立 window，焦点不会回归到主 window 的 player 根 Box，落入同一种焦点真空。
- **B4**：返回二次确认提示 (`TvPlayerBackConfirmPrompt`) 消失时同 B3。
- **B5**：用户按 DPAD_UP（`ExitFocus`）从控制条退出后又遇到 auto‑hide，同 B2。

**所有上述场景共享同一种失效面**：

- **DPAD_DOWN**：方向性 focus search 找不到下方 focusable（root Box 不在用户视觉"下方"），无焦点移动也无事件 dispatch → 无反应。**用户无法重新唤起控制条**。
- **CENTER（含 DPAD_CENTER / ENTER / 媒体键）**：需要已有焦点才能 dispatch，焦点真空下没有目标 → 无反应。**用户无法播放/暂停**。
- **DPAD_UP**：路由表 `focusInControls=false` 时返回 null（设计上不消费），等价于无反应。
- **←/→**：方向性 focus search 偶尔能让 root Box 临时拿到焦点并 dispatch 一次 Seek 事件，所以**单次能 seek 但不能连按**——一旦 Seek 副作用让 controls visible，多个 focusable 出现，下一次 ←/→ 的方向性搜索目标分流，root Box 不再可靠承接。

## 2. 用户故事

作为 TV 用户，我希望：

1. 进入有历史进度的视频，等续播卡自动消失之后，**按 DPAD_DOWN 能立即唤起控制条**并把焦点送到"播放/暂停"按钮上。
2. 长视频播放中任何时刻，**按 CENTER 能切换播放/暂停**——无论之前是否点过控制条按钮。
3. **←/→ 能连按快进/快退**：第一次能 seek，第二次紧接着也能 seek；不要求"先把控制条点亮"。
4. 关闭字幕/音轨 picker、关闭返回二次确认提示之后，立即可以按 DPAD_DOWN/CENTER 操作播放器，**不需要先按 ←/→ 把焦点"踢回去"**。
5. 上述任意路径下任何时刻进入"控制条没焦点"状态，都不应产生"按键全失灵"现象。

## 3. 作用域

### 3.1 覆盖

- `tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`：根 Box 的焦点 owner 重构、续播卡 slot 接入、overlay 状态聚合的 focus 兜底 `LaunchedEffect`
- `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`：原先把 `TvResumePromptCard` 平铺在外层 Box 的位置移到 LongFormVideoPlayer 的 `resumePromptSlot` 里
- `tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`：同上路径处理；电视剧的连播提示卡 `TvEpisodeAutoplayPromptCard`（如存在同类用法）一并审查
- 单测：`LongFormVideoPlayerFocusGuardTest.kt`（纯函数：overlay 聚合判定）+ `LongFormVideoPlayerSpecTest.kt` 扫描禁项
- 版本号：`android-tv-app/tv-app/build.gradle.kts` `versionCode = 69`、`versionName = "0.1.69"`

### 3.2 不覆盖

- 短视频全屏播放、IPTV、TV 设置、TV 首页等其他 TV 屏幕的焦点行为
- 手机端 `android-app/`（独立工程）
- `LongFormVideoPlayer` 的路由表 `TvLongFormRemoteKeyRouting.kt`——路由本身没有 bug，DPAD_DOWN 在 `focusInControls=false` 时已经正确返回 `EnterFocus`；问题在于 onPreviewKeyEvent 收不到事件
- 续播卡 / picker 的视觉样式、动画、文案
- 现有 5 秒 auto‑hide 的时长与触发逻辑

## 4. 非目标

- **N1**：不引入 Compose `focusGroup` / `focusRestorer` 等高阶抽象——本任务只补差缺，不重构焦点模型
- **N2**：不改变续播卡的 5 秒倒计时 / "继续观看" / "从头播放" 业务语义
- **N3**：不把 picker 从 Dialog 换成 Composition 内嵌 Surface——Dialog 的隔离窗口是有意行为
- **N4**：不在路由表里改 DPAD_UP `focusInControls=false` 时的"返回 null"——这是设计选择
- **N5**：不收集焦点变化的统计/埋点

## 5. 术语对应（实施完成时 sync 到 CONTEXT.md）

| 术语 | 一句话定义 |
|------|-----------|
| `TV 长视频焦点真空` | TV 长视频播放器在 [[续播提示卡]] / 字幕 picker / 音轨 picker / 返回二次确认提示等叠加层关闭后留下的 Compose 焦点状态：没有任何 focusable 持焦，导致 root Box 的 `onPreviewKeyEvent` 收不到 DPAD_DOWN / CENTER / DPAD_UP 事件，只有 ←/→ 还能借方向性 focus search 偶尔触发。属于 LibVLC 迁移之后暴露的 [[兄弟节点焦点抢占]] 副作用。 |
| `LongFormVideoPlayer focus 兜底` | 在 [[TV 长视频 LibVLC 内核]] 的 root Box 上挂的一道 LaunchedEffect：聚合 `controlsVisible` / `subtitleSheetVisible` / `audioTrackSheetVisible` / `resumePromptVisible` / `showBackConfirmPrompt` / `playerErrorVisible` 六个 overlay 可见性，任一从 true→false 跃迁且没有其他 overlay 仍在显示时，显式 `rootFocusRequester.tryRequestFocus()`，把焦点回收到 player 根 Box。 |
| `续播提示卡内嵌位置` | 续播提示卡（`TvResumePromptCard`）从 `TvLongFormPlayerScreen` / `TvSeriesPlayerScreen` 外层 Box 的兄弟位置改到 `LongFormVideoPlayer` 内部子树，通过 `resumePromptSlot: @Composable () -> Unit` 槽位提供。配合 [[LongFormVideoPlayer focus 兜底]]，dispose 时 Compose 自然把焦点回收到 ancestor 的 root Box，而不是清空成 null。 |

## 6. 决策表（grill 过程的结论）

| # | 决策点 | 选项 | 结论 | 理由 |
|---|--------|------|------|------|
| Q1 | 修复粒度 | A 只挪续播卡 / A' 三件套 | **A'** | 续播卡只是触发口之一；picker / back prompt / auto‑hide 同类问题都在 |
| Q2 | 续播卡放哪 | 兄弟位置加 callback / 内嵌进 player | **内嵌** | 让 Compose 自然焦点回收，少一道协调代码 |
| Q3 | overlay 状态怎么观察 | 各自挂 onDismiss callback / 集中 LaunchedEffect | **集中 LaunchedEffect** | 单点维护；新增 overlay 时只在聚合处加一个 state |
| Q4 | root Box 是不是该自挂 onFocusChanged | 是 / 否 | **是（兜底用）** | Dialog dismiss 路径不一定经过 LaunchedEffect；onFocusChanged 是最后一道关 |
| Q5 | 路由表要不要联动改 | 要 / 不要 | **不要** | 路由本身正确；改了反而引入回归 |

## 7. 验收 (acceptance criteria)

`review.md §1` 全部场景在 TV 模拟器或真机跑通，并经用户确认。摘要：

1. **R1（B1 重现路径）**：开有历史进度的视频 → 等续播卡倒计时完 → 按 DPAD_DOWN → 控制条出现且焦点在播放/暂停按钮上
2. **R2（B2 路径）**：新视频开播 → 等 5 秒 auto‑hide → 按 DPAD_DOWN → 同 R1
3. **R3（B3 路径）**：播放中 → 按"字幕"或"音轨"按钮打开 picker → 关闭 picker → 立即按 DPAD_DOWN → 同 R1
4. **R4（B4 路径）**：播放中 → 按返回 → 出现二次确认提示 → 提示自动消失 → 按 DPAD_DOWN → 同 R1
5. **R5（B5 路径）**：从控制条按 DPAD_UP 退出 → 等 5 秒 → 按 DPAD_DOWN → 同 R1
6. **R6（←/→ 连按）**：进入新视频 → 等 5 秒 auto‑hide → 连续按 ← 三次 → 每次都触发快退反馈、累加 30 秒
7. **R7（CENTER 任意时刻）**：上述 R1~R5 任一时刻按 CENTER → 触发播放/暂停反馈
8. **R8（无回归）**：从控制条按钮已经持焦的状态下按 ←/→/CENTER → 行为不变（PassThrough / 按钮自身 onClick）
9. **R9（电视剧路径）**：R1~R7 在 `TvSeriesPlayerScreen` 上同样跑通
10. **R10（无副作用）**：长按 2x 倍速 / 触屏 tap / drag seek / 续播卡按"继续观看" / 续播卡按"从头播放" 等既有交互无回归

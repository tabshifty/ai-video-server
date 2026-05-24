# Review：TV 长视频播放器操作 UI 跟随遥控器互动

- 日期：2026-05-24
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`

## 0. Review 准入条件

- [ ] PRD §8.1 全部纯函数单测已写、已转绿（`TvLongFormRemoteKeyRoutingTest` / `TvLongFormTitleOverlayDataTest` / `TvLongFormControlsAutoHideTest`）
- [ ] PRD §8.3 源文 audit 测试已写、已转绿（`TvLongFormTitleOverlaySpecTest`）
- [ ] `./gradlew :tv-app:testDebugUnitTest` BUILD SUCCESSFUL
- [ ] `./gradlew :tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] `./gradlew :tv-app:connectedDebugAndroidTest` BUILD SUCCESSFUL（焦点环绕集成测试）
- [ ] `git diff` 仅触及 implement.md §2 清单内文件
- [ ] `CONTEXT.md` 已追加 implement.md §4 草稿的 7 条新术语 + 修订 line 130 `TV 播放器退出确认` 条目
- [ ] `android-tv-app/tv-app/build.gradle.kts` `versionCode +1`、`versionName` 末位 +1
- [ ] `plan.md` 已追加 reverse-chronological 进度条目
- [ ] 旧测试 `LongFormVideoPlayerTransportKeyTest.kt` 已删除（如内部 helper `resolveTvHiddenTransportKeyAction` 也已被移除）

未满足任一项 → **回到 implement 阶段**，不进入 review。

## 1. 手测脚本（按 PRD §6 验收标准映射）

### 1.1 任意遥控器互动唤起操作 UI — PRD §6.1

进入一部 TV 长视频（建议先用一部电视剧 S1E7 测，覆盖左上信息层主行 + 副行）。等首次进入的 5 秒自动隐藏完成（[[TV 操作 UI 层]] 完全淡出）。

1. **A1**：按 ← 一次
   - [ ] 视频回退到 10 秒前（或当前 [[快进/快退步长]] 步长前）
   - [ ] 底部控制条玻璃条出现 + 左上信息层出现（主行剧名 + 副行季集）
   - [ ] 焦点光感**不**出现在任何控制按钮上（焦点停在播放器根）
   - [ ] 5 秒后操作 UI 完整淡出（240ms tween，不卡顿、不闪烁）
2. **A2**：按 → 一次
   - [ ] 视频前进 10 秒
   - [ ] 操作 UI 出现，焦点不进入
3. **A3**：按 ↓ 一次
   - [ ] 操作 UI 出现
   - [ ] 焦点蓝青色光感落在「播放/暂停」按钮（控制条最左侧）
4. **A4**：按 OK 一次（焦点在播放器根）
   - [ ] 视频暂停 + 中央 toast「已暂停」出现 ~900ms 后淡出
   - [ ] 操作 UI **不**亮（无底部控制条 / 无左上信息层）

### 1.2 控制条焦点入口与左右键焦点环绕 — PRD §6.2

接 A3 完成（焦点在「播放/暂停」，UI 可见）。

5. **B1**：按 → 一次
   - [ ] 焦点移到下一个按钮（电视剧 TV 模式下应为「快退」）
   - [ ] 5 秒计时重置（UI 继续可见）
6. **B2**：连续按 → 直到走完所有可见按钮
   - [ ] 焦点穿过：播放/暂停 → 快退 → 快进 → 选集（仅电视剧） → 下一集（仅电视剧且有下一集时） → 字幕 → 音轨 → 返回详情 → 退出播放（如有）
   - [ ] 焦点**不**进入 Slider（Slider 不获得光感、不开始 scrub）
7. **B3**：在最右侧按钮上再按 → 一次
   - [ ] 焦点**绕回**最左侧「播放/暂停」按钮
8. **B4**：在「播放/暂停」按钮上按 ← 一次
   - [ ] 焦点**绕回**最右侧按钮
9. **B5**：在 B3 状态下按 OK
   - [ ] 该按钮原有 onClick 触发（如「退出播放」会进入 [[TV 播放器退出确认]] 第一次提示）

### 1.3 5 秒计时重置 — PRD §6.3

接 B1 完成（焦点已进入 controls）。

10. **C1**：以 4 秒一次的节奏按 →（在按钮间环绕），持续 ~30 秒
    - [ ] 操作 UI 始终可见，不被自动隐藏打断
11. **C2**：停止操作 5 秒
    - [ ] 操作 UI 完整淡出
    - [ ] 焦点光感**不**残留（焦点在控制条消失时自然失去 visual focus）
12. **C3**：UI 完全隐藏后，4 秒内连续按 ← 三次
    - [ ] seek 走 [[连按合并跳转]] 300ms 防抖合并（一次实际 seekTo）
    - [ ] 操作 UI 全程可见
    - [ ] 每次按键重置 5 秒计时

### 1.4 左上信息层 — PRD §6.4

13. **D1**：电影播放器（任一电影）
    - [ ] UI 可见时左上角主行 = 电影标题，白色 SemiBold，带阴影
    - [ ] **无**副行（季集信息不出现）
14. **D2**：`18+` 长视频播放器
    - [ ] 同电影：主行 = AV 标题，无副行
15. **D3**：电视剧 S1E7 播放器
    - [ ] 主行 = 剧名（如「明朝那些事」）
    - [ ] 副行 = `第 1 季 · 第 7 集 朱元璋登基`，字号偏小、alpha 约 0.72
    - [ ] 主行 + 副行均有阴影；阴影方向一致
16. **D4**：D3 状态下，触发自动连播或手动切到下一集
    - [ ] 副行的集号与单集标题自动更新（无 stale state）
17. **D5**：电视剧某集 `episode.title` 为空字符串或 null
    - [ ] 副行仅显示「第 X 季 · 第 Y 集」，**无**尾部空格 / 无单独的点
18. **D6**：在亮场景画面（如纯白雪景、白色 UI 截图、强光镜头）暂停
    - [ ] 文字仍可读（阴影提供对比度），不出现"白底白字消失"

### 1.5 BACK 优先收 UI — PRD §6.5

19. **E1**：[[TV 操作 UI 层]] 可见时按 BACK 一次
    - [ ] 操作 UI 立即开始收起动画（240ms fade）
    - [ ] **不**出现 [[TV 播放器退出确认]] 的第一次提示文字
    - [ ] 焦点（若在控制条内）回到播放器根，无焦点光感残留
20. **E2**：UI 已收起后按 BACK 一次
    - [ ] 出现 [[TV 播放器退出确认]] 第一次提示
21. **E3**：E2 状态下 5 秒内再按 BACK 一次
    - [ ] 真正退出播放器
22. **E4**：UI 可见 + 焦点在「退出播放」按钮上时按 OK
    - [ ] **进入** [[TV 播放器退出确认]] 第一次提示（按钮点击不被 BACK 优先收 UI 影响）

### 1.6 首次进入视频的初始状态 — PRD §6.6

23. **F1**：冷启从详情页进入任一 TV 长视频播放器
    - [ ] 进入瞬间操作 UI 自动亮起（5 秒）
    - [ ] 焦点光感**不**出现在任何控制按钮上（焦点在播放器根）
24. **F2**：F1 状态不操作，等 5 秒
    - [ ] 操作 UI 完整淡出
25. **F3**：F1 状态在 5 秒内按 ↓
    - [ ] 焦点进入「播放/暂停」按钮
26. **F4**：F1 状态在 5 秒内按 →
    - [ ] seek 前进 + UI 继续可见 + 5 秒计时重置（焦点仍在播放器根）

### 1.7 与既有特性不冲突 — PRD §6.7

27. **G1**：进入电视剧 S1E1，有续播位置 → [[续播提示卡]] 出现
    - [ ] 左上信息层正常显示（不被续播卡遮挡 / 不重叠）
    - [ ] 续播卡的 OK 「继续观看」/「从头播放」按钮正常工作（焦点契约不被本任务破坏）
28. **G2**：电视剧片尾 T-10 → [[连播提示卡]] 出现
    - [ ] 左上信息层正常显示
    - [ ] 连播卡焦点抢占到「立即播放」按钮（焦点契约不被本任务破坏）
29. **G3**：UI 可见 + 焦点在「字幕」按钮上按 OK 打开 [[字幕选择夜台玻璃面板]]
    - [ ] 面板出现期间左上信息层继续可见（让用户记得在哪一集选轨）
    - [ ] 关闭面板（BACK）后 UI 仍可见 5 秒后自动隐藏（沿用既有 sheet 关闭 → controls 5s 计时）
30. **G4**：[[运行时切轨]] 在播放中切音轨
    - [ ] 切轨正常工作，左上信息层不影响
31. **G5**：[[连按合并跳转]] 长按 → 3 秒
    - [ ] seek 累积 ~30 × 3 = 90 秒（按 `repeatCount > 0` 走 3 倍步长，沿用既有）
    - [ ] UI 全程可见、计时重置
32. **G6**：UI 可见 + 「字幕」按钮上按 OK 打开字幕面板，再按 BACK
    - [ ] 先关闭字幕面板（既有 sheet 行为）
    - [ ] **此时控制条仍可见**（sheet 关闭不直接收 controls）
    - [ ] 再按 BACK 一次 → 收 controls（[[BACK 优先收 UI]] 触发）
    - [ ] 再按 BACK 一次 → [[TV 播放器退出确认]] 第一次提示

## 2. 代码审计（diff 检查）

### 2.1 必须出现

- [ ] `core/ui/TvLongFormRemoteKeyRouting.kt` 新增，含 `TvRemoteKeyAction` 密封类 + `resolveTvRemoteKeyAction` 纯函数 + `shouldResetAutoHideTimer` 纯函数 + `buildTvLongFormTitleOverlayData` 纯函数
- [ ] `core/ui/TvLongFormTitleOverlay.kt` 新增，含 `TvLongFormTitleOverlayTokens` object + `TvLongFormTitleOverlayData` data class + `TvLongFormTitleOverlay` Composable
- [ ] `core/ui/LongFormVideoPlayer.kt` 改动：
  - [ ] `fun LongFormVideoPlayer(...)` 签名新增 `seasonNumber: Int? = null` / `episodeNumber: Int? = null` / `episodeTitle: String? = null` / `seriesTitleForOverlay: String? = null`
  - [ ] `if (tvMode)` 分支调用 `TvLongFormTitleOverlay`
  - [ ] `else` 分支保留现有顶部玻璃栏（line 670–718 内容在 else 分支内完整保留）
  - [ ] `onPreviewKeyEvent` 内仅一处调用 `resolveTvRemoteKeyAction`，分支处理 6 种 action
  - [ ] 控制条按钮上挂 `Modifier.focusProperties { left = ...; right = ... }`
  - [ ] Slider 挂 `Modifier.focusProperties { canFocus = false }`
  - [ ] `focusInControls` 状态有 setter（任一按钮 onFocusChanged）
  - [ ] 旧 `handleTvTransportKey` / `resolveTvHiddenTransportKeyAction` / `TvHiddenTransportKeyAction` 已**移除**
- [ ] `feature/tv/TvSeriesPlayerScreen.kt` 改动：
  - [ ] `LongFormVideoPlayer(...)` 调用点 `title = ...` 改为剧名兜底逻辑
  - [ ] 新增 `seriesTitleForOverlay = series.title` / `seasonNumber = uiState.selectedSeasonNumber` / `episodeNumber = uiState.selectedEpisodeNumber` / `episodeTitle = currentEpisode?.title`
- [ ] `CONTEXT.md` 改动：
  - [ ] 「TV 播放术语」区追加 7 条新术语（[[TV 操作 UI 层]] / [[左上信息层]] / [[操作 UI 互动唤起]] / [[controls 焦点入口]] / [[controls 焦点环绕]] / [[BACK 优先收 UI]] / [[controls 焦点退出键]]）
  - [ ] line 130 [[TV 播放器退出确认]] 条目末尾追加 [[BACK 优先收 UI]] 前置例外说明
- [ ] `android-tv-app/tv-app/build.gradle.kts`：`versionCode` 数字 +1，`versionName` 末位 +1
- [ ] `plan.md`：顶部新增本任务的 reverse-chronological 条目（日期、摘要、影响文件、验证状态）

### 2.2 禁止出现

- [ ] `LongFormVideoPlayer.kt` 中 `fadeIn(`、`fadeOut(` 后**必带** `tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)`；不出现裸 `fadeIn()` / `fadeOut()` / `fadeIn(tween(NNN))` 字面量
- [ ] `TvLongFormTitleOverlay.kt` 调用点不出现裸 `22.sp` / `16.sp` / `Color(0xCC000000)` / `Offset(0f, 2f)` / `4f`（所有数值经 `TvLongFormTitleOverlayTokens`）
- [ ] `TvLongFormTitleOverlay.kt` 使用 `TextStyle(shadow = Shadow(...))`，不使用 `Modifier.shadow(...)`
- [ ] `LongFormVideoPlayer.kt` 中 `onPreviewKeyEvent` 内**不**出现按 keyCode 的 `when` 块（已统一收口到 `resolveTvRemoteKeyAction`）
- [ ] `LongFormVideoPlayer.kt` 中**不**保留 `handleTvTransportKey` / `resolveTvHiddenTransportKeyAction` / `TvHiddenTransportKeyAction` 任何引用
- [ ] 调用方代码（`TvSeriesPlayerScreen` / `TvLongFormPlayerScreen`）**不**直接 import `TvLongFormTitleOverlay` 或 `TvRemoteKeyAction`（内部细节由 `LongFormVideoPlayer` 隔离）
- [ ] `LongFormVideoPlayer.kt` 中**不**出现 `focusable(false)` 反向覆盖 Slider 焦点

### 2.3 不变量保护（既有约束）

- [ ] `tvFocusableGlow` / `tvFocusableScaleOnly` 调用方不变（焦点视觉沿用 [[TV 焦点视觉]] 现有 token）
- [ ] `AppChrome.SurfaceShape` 仍为底部 controls 玻璃条圆角来源，未引入新 `RoundedCornerShape(N.dp)` 字面量（`TvShapeAuditTest` 应继续通过）
- [ ] `TvMotionTokens.DurationStandardMs` 仍为唯一 fade duration token（`TvMotionTokensTest` 应继续通过）
- [ ] [[TV 文本溢出保护]] 在 `TvLongFormTitleOverlay` 主行与副行均 `maxLines = 1 + ellipsis`
- [ ] [[字幕处理约定]] / [[运行时切轨]] / [[续播提示卡]] / [[连播提示卡]] 行为零回归

## 3. 性能与边界

- [ ] 控制条按钮焦点环绕在按钮可见性变化（电视剧切到无下一集 / 进入退出确认后控制条按钮可见性变化）时**不崩**——FocusRequester 列表动态重建后焦点不丢
- [ ] 长按 → 持续 10 秒（连按 ~50 次）后释放，UI 全程可见、5 秒计时随每次 keyDown 重置
- [ ] 进入 → 立即按 ↓ 把焦点送到「播放/暂停」→ 5 秒后焦点退回根（自动 hide），再按 ← 重新进入循环——多次轮回不漏焦点 / 不卡 controls
- [ ] 视频加载中（player STATE_BUFFERING）按 ← / →：seek 走 `player.seekTo`，position 状态由现有 `LaunchedEffect` 250ms 轮询同步，UI 表现一致

## 4. 回归扫描区

执行 review 时务必单独验证以下既有交互未受影响：

- [ ] **手机短视频全屏**（`UnifiedPlayerScreen` 中 `LongFormVideoPlayer` 非 TV 模式）：tap 切换 controls / 拖拽 seek / 长按倍速 / 双击切换播放暂停——全部沿用原行为
- [ ] **TV 长视频续播提示**（[[续播提示卡]]）：抢焦点行为、倒计时驱动、永久 dismiss 信号
- [ ] **TV 电视剧自动连播**（[[连播提示卡]] + [[连播覆盖层]]）：T-10 倒计时、抢焦点、取消本次、整剧末尾覆盖层
- [ ] **TV 长视频字幕 / 音轨选择**（[[夜台玻璃面板]]）：触发、关闭、运行时切轨

## 5. 不在 review 范围

- 后端 `internal/services/scraper_*` 改动（本任务零后端改动）
- admin-web 改动（本任务零 admin-web 改动）
- 短视频 `feature/shorts/` 改动（本任务零短视频改动）
- IPTV 播放（独立 LibVLC 路径，本任务不动）
- `pkg/ffmpeg` 改动（无后端改动）

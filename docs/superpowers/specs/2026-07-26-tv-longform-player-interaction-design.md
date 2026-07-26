# TV 长视频播放器交互模型重构设计

- 日期：2026-07-26
- 范围：`android-tv-app` 长视频播放器（单片电影 / `18+` 与电视剧分集）
- 目标：把 TV 长视频播放器的交互从「步进式盲 seek」重做为主流电视播放器的 scrub 模型，补齐缓冲反馈、控制条分层与 MediaSession，使其在遥控器上的行为符合电视应用形态

## 1. 背景与问题

当前 TV 长视频播放器已经使用 Media3/ExoPlayer 内核与共享控制层 `TvSeriesCorePlaybackOverlay`，具备三层焦点（Root / Controls / EpisodeRail）、字幕音轨选择、续播提示、自动连播、软重试等能力，工程基础扎实。但与主流电视播放器相比存在结构性缺口：

| 缺口 | 现状 |
| --- | --- |
| 快进快退是盲操作 | 左右键按固定步长跳转，用户看不到目标点在时间轴上的位置，长片里定位需要反复试探 |
| 缓冲态无任何反馈 | `STATE_BUFFERING` 不驱动任何 UI，卡顿时画面直接停住，用户无法区分「卡了」与「坏了」 |
| 控制条单行挤压 | 播放键、当前时间、进度条、总时长、字幕、音轨全部塞进一个 `Row`，不是电视端标准的分层布局 |
| 无 MediaSession | 遥控器媒体键、蓝牙耳机控制、系统播放状态、Android TV 首屏「继续观看」频道全部不可用 |
| `DPAD_UP` 在 Root 层空转 | 电视播放器惯例中 UP 用于唤起信息层 |

其中「盲 seek」是最核心的违和来源：它让播放器像一个带快进按钮的视频控件，而不像电视播放器。

## 2. 范围

**纳入**：单片长视频播放器与电视剧分集播放器的共享控制层与按键路由层。

**不纳入**：
- IPTV 直播播放器（LibVLC 路径，独立维护）
- TV 本地短视频页与手机投放页（竖屏流交互本质不同）
- 手机端播放器（`CONTEXT.md` 已锁定「TV 长视频 ExoPlayer 端隔离」）
- 播放内核、解码策略、Dolby Vision 门控、播放兼容决策
- 服务端接口与转码产物
- 电视剧自动连播的业务规则（10s 倒计时窗口、取消本次、连播覆盖层）——其触发条件与文案不变；本次只保证 scrub 期间不显示连播提示卡，且 scrub 提交后若仍在 T-10 窗口内则正常恢复显示

**本次不做缩略图预览**：scrub 只呈现时间轴信息（目标时间、偏移量、ghost 游标位置），不显示画面缩略图。缩略图需要服务端新增雪碧图生成与接口，属于独立的全栈特性，不在本次范围内。

## 3. 交互模型

### 3.1 状态机

播放器 UI 收敛为四个互斥模式：

```
Hidden ──[←/→]───────────→ Scrubbing        （直接进入，不经 Chrome）
Hidden ──[↓]────────────→ Chrome（焦点进入控制条）
Hidden ──[↑ / OK / 其它]──→ Chrome（仅显示，焦点留在 Root）
Chrome ──[←/→]───────────→ Scrubbing
Chrome ──[↓]────────────→ Chrome（焦点进入控制条）
Chrome ──[↓↓]───────────→ EpisodeRail       （仅电视剧）
Chrome ──[5s 无操作]──────→ Hidden
Scrubbing ──[←/→]────────→ Scrubbing（累加，长按加速）
Scrubbing ──[OK]─────────→ 提交跳转 → Chrome
Scrubbing ──[BACK]───────→ 放弃，回锚点 → Chrome
Scrubbing ──[1.5s 无操作]─→ 提交跳转 → Chrome
```

按键优先级：左右键的 scrub 语义优先于「任意键唤起 Chrome」——从 Hidden 按左右键直接进入 Scrubbing，不经过 Chrome 中间态。DOWN 在 Hidden 与 Chrome 下语义一致（焦点进入控制条），差别只在 Chrome 是否已经可见。OK 在 Root 层（Hidden 或 Chrome 但焦点未入控制条）执行播放暂停切换并显示中心反馈，不夺取焦点。

### 3.2 关键交互决定

- **scrub 期间视频继续播放，不暂停。** 对齐 Netflix / Disney+；暂停会让「顺手快进一下」变成一次卡顿。
- **scrub 由 Root 层驱动，进度轨不进入焦点链。** 保留既有「seek 只属于播放器根层」的合理内核，只把反馈做厚。控制条聚焦时左右键仍然是切换按钮，不触发 seek。
- **长按分三档加速。** `repeatCount` 0–2 → 1× 步长，3–7 → 3×，8+ → 6×。现状是简单的 `repeatCount > 0 → 3×`，在长片里跨越一小时需要按住过久。
- **`DPAD_UP` 在 Root 层唤起 Chrome**（只显示，不夺取焦点），替代当前的空转。
- **BACK 语义分层**：Scrubbing 时取消 scrub → Chrome 可见时收起 UI → Hidden 时才进入退出确认。退出确认本身沿用既有的双击契约（首次按显示提示，窗口期内再次按才退出），本次只是在它前面插入两级拦截。

### 3.3 被替换的既有契约

以下 `CONTEXT.md` 条目在本次重构后失效，需显式标注为已替换：

| 条目 | 处理 |
| --- | --- |
| `seek 进度显示防抖` | 废止。新模型下实际位置与目标位置分别渲染，不存在需要防抖掩盖的回跳 |
| `连按合并跳转`（300ms 防抖提交） | 废止。由 scrub 的 1.5s 空闲提交与 OK 显式提交取代 |
| `电视剧进度条只展示不交互` | 修订。进度轨成为 scrub 的可视化载体，但仍不进入焦点链，不承担拖动 |
| `controls 焦点入口`（DOWN 分层） | 保留 DOWN 分层语义，新增 Root 层 UP 分支 |

`操作 UI 互动唤起`、`controls 焦点环绕`、`controls 左右键切焦点`、`controls 持焦横向导航` 继续有效。

## 4. 架构与模块边界

```
feature/tv/TvLongFormPlayerScreen.kt   ← 单片宿主（现存，瘦身）
feature/tv/TvSeriesPlayerScreen.kt     ← 剧集宿主（现存，瘦身）
        │  两者都渲染 ↓
core/ui/player/TvPlayerChrome.kt       ← 新，替代 TvSeriesCorePlaybackOverlay
        ├─ TvPlayerTopInfo.kt          ← 顶部信息层（片名 / 季集信息）
        ├─ TvPlayerScrubTrack.kt       ← 进度轨 + ghost 游标 + 时间
        ├─ TvPlayerActionRow.kt        ← 功能按钮行（播放 / 字幕 / 音轨 / 选集）
        ├─ TvPlayerCenterFeedback.kt   ← 中心瞬时反馈
        └─ TvPlayerBufferingIndicator.kt
core/ui/player/TvPlayerInteraction.kt  ← 纯 Kotlin reducer，零 Compose 依赖
core/ui/player/TvPlayerScrubMath.kt    ← 纯 Kotlin：锚点 / 累加 / 加速档 / 钳位
feature/tv/TvLongFormMediaSession.kt   ← 新，MediaSession 独立 effect
```

### 4.1 三条边界规则

**规则一：reducer 不知道 Compose 存在。**
`TvPlayerInteraction.kt` 输入 `(当前模式, 按键, 播放状态, 是否有选集, 经过时间)`，输出 `(下一模式, 副作用列表)`。副作用是 data class（`CommitSeek(ms)`、`RequestFocus(target)`、`TogglePlayPause` 等），由 Chrome 层翻译为实际调用。跨状态约束（scrub 期间冻结自动隐藏、缓冲不打断 scrub、overlay 可见时透传按键）成为 reducer 的显式分支，可在 JVM 单测中穷举。

这不是新发明的形态——项目已有的 `resolveTvRemoteKeyAction`、`resolveTvPendingStepSeek` 就是这个模式，本次把它推到完整状态机。

**规则二：Chrome 不知道播放内核存在。**
Chrome 接收 `TvPlayerUiState`（位置 / 时长 / 是否播放 / 是否缓冲 / 轨道列表）与一组回调，维持既有「TV 播放内核适配」契约，也让 Chrome 可在无 ExoPlayer 环境下测试。

**规则三：MediaSession 是宿主层的 effect，不进 Chrome。**
`DisposableEffect` 中创建 `MediaSession` 绑定 ExoPlayer 实例，`onDispose` 释放。与「TV 长视频播放器导航即时退出」契约相容——session 随播放器页销毁而释放。需新增依赖 `androidx.media3:media3-session:1.4.1`。

### 4.2 顺带的边界清理

`core/ui/LongFormVideoPlayer.kt`（1664 行）的主 composable `LongFormVideoPlayer` 在 TV 编译图中是死代码——它只被 `feature/detail/DetailScreen.kt` 与 `feature/player/UnifiedPlayerScreen.kt` 调用，两者都在 `tvMainSourceExcludes` 中。TV 侧实际只使用该文件的 7 个工具：`TvSeriesControlsPage`、`TvEpisodeRail`、`CompactPlayerControlButton`、`formatPlaybackTime`、`TvPlaybackProgressBar`、`resolveTvPendingStepSeek`、`normalizeTvSeekStepSeconds`。

处理：把 TV 仍需要的工具移入 `core/ui/player/`。手机端文件保留原样不动（不在 TV 编译图内，改动无收益也无验证手段）。

### 4.3 替换与保留

- `TvSeriesCorePlaybackOverlay.kt` 及其协作件 `TvSeriesControlsPage`、`TvPlaybackProgressBar` 在 TV 侧被新组件取代后删除
- `TvEpisodeRail` 保留复用，选集轨交互本次不变
- 续播提示卡的内嵌槽位模式（`resumePromptSlot`）保留，`CONTEXT.md`「续播提示卡内嵌位置」契约继续有效

## 5. 状态与数据流

```
ExoPlayer ──4Hz 快照──→ 宿主 Screen ──TvPlayerUiState──→ TvPlayerChrome
                            ↑                                  │
                            │                              遥控按键
                       副作用执行                                │
                            │                                  ↓
                       PlayerEffect ←──────────── TvPlayerInteraction (reducer)
```

### 5.1 scrub 的两个进度真相

scrub 期间存在两个位置：`player.currentPosition`（实际播放，仍在前进）与 `scrubTargetMs`（用户意图）。进度轨同时渲染两者——实心条走实际位置，ghost 游标停在目标位置。

这消除了现状「`pendingStepSeek` 覆盖显示位置」导致的进度回跳抖动；`CONTEXT.md`「seek 进度显示防抖」正是为掩盖该问题而存在，新模型下不再必要。

### 5.2 scrub 锚点只在进入时取一次

进入 Scrubbing 时快照 `anchorMs = player.currentPosition`，后续累加基于 `scrubTargetMs` 而非实时位置——否则视频边播边累加会让目标点持续漂移。

BACK 取消时不执行 seek（因为从未 seek 过），只丢弃 `scrubTargetMs`。

### 5.3 缓冲态抖动抑制

`STATE_BUFFERING` 不直接驱动 UI：进入缓冲后延迟 350ms 才显示指示器，退出缓冲立即隐藏。seek 后必然经过的短暂缓冲不会造成闪烁。缓冲指示器与 scrub 可以共存（scrub 提交后进入缓冲是正常序列）。

### 5.4 自动隐藏计时的冻结条件

Chrome 的 5s 计时在以下情况暂停：
- Scrubbing 模式
- EpisodeRail 可见
- 任一 overlay 可见（续播卡 / 字幕 / 音轨 / 返回确认 / 连播提示）
- **播放已暂停**（新增——暂停时 UI 常驻是电视播放器惯例）

### 5.5 沿用不变的契约

- `PlayerFocusGuardInput.anyOverlayVisible()` 的按键透传规则
- 续播提示卡内嵌槽位
- `tryRequestFocus` 三层焦点防线
- 播放历史上报三条路径（周期上报 / dispose 上报 / 生命周期 pause 上报）；scrub 提交后的位置变化由 4Hz 快照自然捕获

新增的是：reducer 在任一 overlay 可见时直接返回 `PassThrough`，把「不抢键」从散落的守卫上收成一条规则。

## 6. 错误处理

### 6.1 沿用现有分层

软重试体系（`TvLongFormSoftRetryUiState` 四态：Preparing / Succeeded / Canceled / Failed）、首帧前后的错误分流（未出首帧 → 整页错误态，已出首帧 → 播放器内轻量反馈）、`DV 分集失败不退出剧集会话`、15s 启动超时——全部保留不动。这是「TV 长视频播放器软准备」契约的实现，重构不重造。

### 6.2 新增：缓冲长时间不恢复

现状只有「启动超时」（`preparedSourceKey` 存在但从未进入 READY，15s）。播放中途卡在 `STATE_BUFFERING` 且无恢复迹象时，当前无限转圈无反馈。

新增降级路径：
- 缓冲 8s 后，指示器追加「网络较慢，正在缓冲」文案
- 缓冲 20s 后，转入既有软重试失败态（复用 `Failed`，走同一个重试按钮）

不新建错误通道。

### 6.3 scrub 边界处理

- 目标位置钳位到 `[0, duration]`
- `duration <= 0`（未知时长）时禁止进入 Scrubbing，左右键退化为无操作并给一次中心反馈
- 提交时若播放器已 release 或已换源，副作用被宿主层丢弃（沿用现有 `eventIdentity` 身份校验模式）

## 7. 测试策略

三层，全部 JVM 单测，无需真机。

**第一层：reducer 穷举测试**（新增，最重要）
`TvPlayerInteractionTest`。模式 × 按键的完整矩阵，覆盖全部守卫分支：
- scrub 期间 5s 计时冻结
- overlay 可见时 PassThrough
- BACK 三层语义
- `duration <= 0` 时拒绝进入 scrub
- 长按加速档切换
- 暂停时 Chrome 常驻

**第二层：scrub 数学测试**（新增）
`TvPlayerScrubMathTest`。锚点固定、累加、方向切换、加速档、钳位、边界（0 位置向左、末尾向右）。

**第三层：源码形态审计测试**（沿用本项目既有习惯）
锁定「进度轨 `canFocus = false`」「Chrome 不 import ExoPlayer」「reducer 文件不 import androidx.compose」「MediaSession 在 `DisposableEffect` 中 release」。

### 7.1 既有测试的处理

`TvLongFormRemoteKeyRoutingTest`、`TvSeriesCorePlaybackOverlaySpecTest`、`LongFormVideoPlayerTransportKeyTest`、`LongFormVideoPlayerControlsFocusPolicyTest` 中断言旧 seek 语义的用例会失败。

处理原则是**迁移而非删除**：
- 仍然成立的约束（焦点环绕、选集轨、overlay 透传）迁移到新测试文件
- 被交互模型替换的（300ms 防抖、`pendingStepSeek` 覆盖显示）显式删除，并在 `plan.md` 记录原因

手机端相关测试文件不动（不在 TV 编译图内）。

### 7.2 验证命令

```bash
cd android-tv-app
./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug
```

### 7.3 真机验收项（单列，不阻塞交付）

- scrub 手感：1.5s 空闲提交是否偏长
- 长按加速档跨度是否合适
- 缓冲指示 350ms 阈值是否足够抑制闪烁
- MediaSession 在实际遥控器媒体键与系统「继续观看」频道上的响应

## 8. 版本与提交

按 `AGENTS.md` 约定，TV 端功能变更需递增版本：`versionCode` 141 → 142、`versionName` 0.1.141 → 0.1.142。若拆分多批提交则逐批递增。

每批完成后运行 `:tv-app:testDebugUnitTest` 与 `:tv-app:assembleDebug`，并按仓库约定向 `plan.md` 追加反向时间序条目、向 `CONTEXT.md` 追加长期技术契约（含第 3.3 节被替换契约的显式标注）。

未跟踪的 `.superpowers/` 目录保持原样，不读取、不修改、不纳入提交。

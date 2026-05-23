# Review：TV 电视剧自动连播下一集

- 日期：2026-05-23
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`

## 0. Review 准入条件

- [ ] 所有 PRD §8.1 列出的纯函数单测已写、已转绿
- [ ] `./gradlew :tv-app:testDebugUnitTest` BUILD SUCCESSFUL
- [ ] `./gradlew :tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] `git diff` 仅触及 implement.md §2 清单内文件
- [ ] `CONTEXT.md` 已追加 `电视剧自动连播` / `连播链路` / `连播倒计时窗口` / `连播提示卡` / `取消本次连播` / `连播覆盖层` / `连播自动切上报` / `手动下一集按钮` / `TV 下一集按钮图标` 九条术语
- [ ] `android-tv-app/tv-app/build.gradle.kts` `versionCode +1`、`versionName` 末位 +1
- [ ] `plan.md` 已追加 reverse-chronological 进度条目

未满足任一项，**回到 implement 阶段**而不是继续 review。

## 1. 手测脚本（按 PRD §6 验收标准映射）

### 1.1 自动连播开启路径（默认状态）— PRD §6.1

进入一部至少 2 集且 S1E2 `playable=true` 的电视剧 S1E1 播放器。

1. **A1**：seek 到剩余 11 秒处等待
   - ✓ 右下角约剩余 10 秒时出现连播提示卡，右侧约 48dp，底部避开控制条安全区
   - ✓ 上行显示「即将播放 · 第 2 集 [E2 真实标题]」
   - ✓ 下行显示「立即播放 (10)」+「取消本次」
   - ✓ 焦点蓝青色光感落在「立即播放」按钮
   - ✓ 数字每秒递减 10 → 9 → ... → 1
2. **A2**：不操作让倒计时归零
   - ✓ 立即切到 S1E2，从头开始播放（00:00 起点）
   - ✓ 后端 history 接口接收 S1E1 `completed=true` 上报（看后端日志或 admin-web 历史页）
3. **A3**：再 seek 一部新剧 S1E1 到剩余 5 秒处，提示卡出现后按 OK
   - ✓ 立即切到 S1E2（不等数字归零）
   - ✓ 历史上报 S1E1 `completed=true`
4. **A4**：再来一次 S1E1，提示出现后 DPad LEFT 切焦点到「取消本次」，按 OK
   - ✓ 提示卡消失
   - ✓ 本集继续播完
   - ✓ STATE_ENDED 时显示「本集已播完」覆盖层（黑色半透明 + 居中文案 + 两按钮）
   - ✓ 焦点自动落在「播放下一集」按钮
   - ✓ 点「播放下一集」→ 切到 S1E2 从头
   - ✓ S1E2 再播到 T-10 时再次出现提示卡（无"已取消"记忆）

### 1.2 跨季顺接 — PRD §6.2

需要测试库里有一部至少 2 季的剧（S1 + S2，且 S2E1 `playable=true`）。

5. **B1**：在 S1 最后一集（设为 S1E10）seek 到 T-11，等待提示出现
   - ✓ 提示显示「即将播放 · 第 1 集 [S2E1 标题]」
   - ✓ 倒计时归零自动切到 S2E1（季号从 1 切到 2，集号回到 1）
6. **B2**：再进 S1E10，控制条上点「下一集」按钮（用 `Icons.Filled.SkipNext` 图标识别）
   - ✓ 立即切到 S2E1，不等倒计时

### 1.3 跳过不可播放集 — PRD §6.3

需构造 / 选择一部剧，S1E5 后面 S1E6 `playable=false`，S1E7 `playable=true`。

7. **C1**：S1E5 seek 到 T-11
   - ✓ 提示显示「即将播放 · 第 7 集 [S1E7 标题]」（跳过 E6）
   - ✓ 倒计时归零切到 S1E7
8. **C2**：在 S1E5 控制条点「下一集」
   - ✓ 立即切到 S1E7
9. **C3**：打开选集面板（控制条「选集」按钮）
   - ✓ S1E6 仍可见，副标题"待绑定 / 未就绪"（现状不变）
   - ✓ S1E6 整行点击无反应（现状不变）

### 1.4 整剧末尾 — PRD §6.4

10. **D1**：在整剧最后一集（无下季、本季最后）seek 到 T-11
    - ✓ 不出现提示卡
    - ✓ 控制条上「下一集」按钮**不渲染**（识别：右键扫到「选集」按钮就跳到「字幕」，没有 SkipNext 图标）
11. **D2**：让 D1 状态继续播到 STATE_ENDED
    - ✓ 显示「全剧已播完」覆盖层
    - ✓ 仅一个「返回详情」按钮，焦点自动落在它
    - ✓ 点击 → 退回详情页
12. **D3**：D2 状态下按 BACK 键
    - ✓ 先显示退出确认提示「再按一次返回」（不被覆盖层吞掉）
    - ✓ 再按 BACK → 退出播放器

### 1.5 状态守卫 — PRD §6.5

13. **E1**：S1E1 seek 到 T-7（在 10 秒倒计时窗口内）后按暂停
    - ✓ 提示卡如已出现，倒计时数字**停在当前值**，不再递减
    - ✓ 继续播放后倒计时从停止处继续往下走
14. **E2**：S1E1 提示出现倒计时到 7 时，DPad LEFT 焦点到「取消本次」，但**不**点；改 seek 回 T-30 之前
    - ✓ 提示卡立即消失，倒计时复位
    - ✓ 再次让媒体自然播到 T-10，提示卡重新出现并从 10 开始
15. **E3**：S1E1 提示出现倒计时到 7 时，按控制条「选集」打开选集面板
    - ✓ 提示卡消失
    - ✓ 关闭选集面板后，若仍在 T-10 区间，提示卡重新出现
16. **E4**：S1E1 提示出现倒计时到 7 时，按 BACK 键
    - ✓ 退出确认提示出现，连播提示卡隐藏
    - ✓ 5 秒内不再按 BACK，退出确认消失；若仍在 T-10 区间，连播提示卡重新出现
17. **E5**：模拟播放错误（停止后端 / 切到一个 URL 失效的视频）
    - ✓ 错误 banner 出现期间连播提示卡不出现

### 1.6 全局开关 — PRD §6.6

18. **F1**：进入 TV 设置 → 播放分组
    - ✓ 看到「自动连播下一集」开关 row，紧邻「快进/快退步长」
    - ✓ 默认开（首次安装时）
    - ✓ 焦点蓝青色光感正确
19. **F2**：关闭开关，回到首页，重新进入电视剧 S1E1 播放，seek 到 T-5
    - ✓ 不出现提示卡
20. **F3**：F2 状态让 S1E1 STATE_ENDED
    - ✓ 显示「本集已播完」覆盖层（因为有下一集），含「播放下一集」+「返回详情」两按钮
21. **F4**：F2 状态下点控制条「下一集」
    - ✓ 立即切到 S1E2（开关关闭**不**影响手动按钮）
22. **F5**：杀掉 App 进程，重新启动并打开播放器
    - ✓ 开关状态保持关闭（持久化生效）
23. **F6**：在 F5 状态重新进入设置打开开关，再回播放器
    - ✓ 提示卡逻辑恢复（让 S1E1 播到 T-10 出现提示卡）

### 1.7 手动按钮 + 图标 — PRD §6.7

24. **G1**：控制条上「下一集」按钮的图标视觉
    - ✓ 是 `SkipNext`（有竖线 + 三角箭头），不是 `FastForward`（双三角箭头）
    - ✓ 与左侧「快进」按钮（`FastForward`）图标明显不同
    - ✓ 保留手动阈值语义，不强制 `completed=true`
25. **G2**：整剧最后一集时
    - ✓ 「下一集」按钮**不渲染**（D1 已覆盖）
26. **G3**：还有 `playable=true` 后续集时
    - ✓ 「下一集」按钮正常渲染并可聚焦
27. **G4**：自动连播开关关闭时
    - ✓ 「下一集」按钮仍渲染（F4 已覆盖）

### 1.8 自动切的历史上报 — PRD §6.8

28. **H1**：A2、A3 路径上检查 admin-web 历史页或后端日志
    - ✓ S1E1 `completed=true`
29. **H2**：A4 路径上让本集 STATE_ENDED 后再切到 S1E2
    - ✓ S1E1 `completed=true`（取消本次也算看完）
30. **H3**：直接点控制条「下一集」按钮手动切（不经过倒计时）
    - ✓ 沿用 `tvPlaybackHistorySnapshot` 位置阈值；如 seek 到 T-30 直接点下一集，可能 `completed=false`（行为不变）
    - ✓ 不和连播提示卡的自动切上报混淆

## 2. 回归手测（确保没破坏现有功能）

| # | 项 | 期望 |
|---|---|---|
| GR1 | 字幕选择 | 控制条「字幕」按钮 → 夜台玻璃面板正常弹出、可选择、运行时切轨正常 |
| GR2 | 音轨选择 | 控制条「音轨」按钮 → 同上 |
| GR3 | 选集面板 | 控制条「选集」按钮 → ModalBottomSheet 正常 |
| GR4 | 快进/快退步长 | 全局设置不受影响；播放器左右键步长按用户设置 |
| GR5 | 连按合并跳转 | 连续按右键的合并 seekTo 行为不变 |
| GR6 | TV 播放器退出确认 | 一次提示、二次退出仍正常 |
| GR7 | 长视频电影 / `18+` 详情页全屏播放 | 不受影响（不走电视剧路径） |
| GR8 | 手机端短视频全屏 | 不受影响（手机工程隔离） |
| GR9 | IPTV 频道播放 | 不受影响 |
| GR10 | TV 首页焦点 / 海报墙 / 详情页 | 不受影响 |
| GR11 | TV 配对页、连接服务器页 | 不受影响 |
| GR12 | `LongFormVideoPlayer` 其他按钮（暂停、字幕、音轨、快进 FastForward） | 视觉和行为不变 |
| GR13 | 自动连播开关默认值 | 全新安装的 App 首次进入设置默认开 |
| GR14 | DataStore 兼容性 | 老用户升级时无 `tv_series_autoplay_enabled` key，`parse(null) = true`，行为是开启 |

## 3. 自动化验证

```bash
# 单元测试
cd android-tv-app
./gradlew :tv-app:testDebugUnitTest
# 期望 BUILD SUCCESSFUL，新增 TvSeriesAutoplaySpecTest / TvAutoplayPromptCardSpecTest 全绿
# 期望 TvShapeAuditTest 仍绿（新文件不破规则）
# 期望 TvTypographySpecTest / TvFocusSpecTest / TvFeaturedHeroMotionSpecTest 仍绿

# Debug 打包
./gradlew :tv-app:assembleDebug
# 期望 BUILD SUCCESSFUL

# Release 打包（验证 R8 不会裁掉新 DataStore key 相关类型）
./gradlew :tv-app:assembleRelease
# 期望 BUILD SUCCESSFUL；Boolean DataStore key 不经过 Retrofit/Gson 不需要 keep 规则
```

## 4. 边界 case 自查

提交前对照下列易遗漏点逐一检查：

1. **整剧只有一集的剧**：`resolveNextPlayableEpisode` 应返回 null；提示卡不出现；STATE_ENDED 显示「全剧已播完」覆盖层
2. **整剧每集 `playable=false`，仅 S1E1 可播**：进入 S1E1 播放，提示卡不出现（链路解析为 null）；STATE_ENDED 显示「全剧已播完」覆盖层
3. **video duration 未就绪（duration ≤ 0）**：守卫条件 `durationMs > 0` 应使提示卡不出现，避免拿不到 duration 时算错 remaining
4. **快速反复 seek 跨过 T-10**：seek 进入 → 提示出现 → seek 出去 → 提示消失；连续多次操作不应出现 ghost 提示卡或多次切集
5. **倒计时归零的瞬间用户按 BACK**：BACK 应触发退出确认；不应在退出确认弹出后还把切集动作 fire 出去（实现层用 `lastSwitchedFromVideoId` guard）
6. **DataStore 写入异步延迟**：开关关闭后立即返回播放器，刚好播到 T-10 → uiState.autoplayEnabled 是否已更新？应在 setAutoplayEnabled 后同步更新 ViewModel state，不只依赖 read 流（也可接受用户切设置后回首页再进剧的延迟模型——这是 implement.md R6 的简化）
7. **设置开关 UI 焦点**：Switch row 的 DPad 上下导航不应卡在开关上无法继续；左右键切换状态正常
8. **横屏模拟器 vs 真机 TV**：在 1080p / 2160p 不同分辨率下提示卡位置距边 48dp 不应被裁掉
9. **同一集 STATE_ENDED 重入**：极端边界 — 倒计时归零切集后 ExoPlayer 还 fire 一次旧集的 STATE_ENDED，应被 `lastSwitchedFromVideoId` 防御吃掉
10. **从历史"继续观看"进入一部剧并直接定位到 T-10 之后**：自然播到 T-10 是从历史 watchSeconds 倒推的位置之后吗？看现有 `currentEpisode?.watchSeconds * 1000L` 是初始 resume 点；若历史已超 T-10 则进入时即触发提示卡——这可能让用户一进剧就被弹卡。守卫里**不**额外抑制（用户视角"我接着上次看完后就该自动连播"是合理的），但提示卡的"第一次出现抢焦点"行为可能短暂打扰，可接受

## 5. 验收清单（提 PR 前 checkbox）

视觉与交互：
- [ ] 连播提示卡在 T-10 自动出现，右侧约 48dp、底部避开控制条安全区，与底部控制条不重叠
- [ ] 提示卡上「立即播放 (N)」按钮的数字每秒递减，文字、按钮、布局符合 PRD §6.1 描述
- [ ] 提示卡第一次出现时焦点抢占到「立即播放」，DPad LEFT 切到「取消本次」
- [ ] BACK 键不被提示卡 / 覆盖层吞掉，仍走「TV 播放器退出确认」
- [ ] 倒计时归零 / 立即播放 → 切到下一集从头播放
- [ ] 取消本次 → 提示卡消失，STATE_ENDED 时显示「本集已播完」覆盖层
- [ ] 整剧末尾不出现提示卡，控制条「下一集」按钮不渲染，STATE_ENDED 显示「全剧已播完」覆盖层
- [ ] 控制条「下一集」按钮使用 `SkipNext` 图标，且保留手动阈值语义
- [ ] 跨季顺接：S1 最后一集 → S2E1 自动切
- [ ] 跳过 `playable=false` 集：手动 + 自动同语义

设置：
- [ ] TV 设置 → 播放 分组下有「自动连播下一集」开关 row
- [ ] 默认开（全新安装）
- [ ] 持久化（杀进程后保持）
- [ ] 关闭后提示卡和自动切都不再触发，手动按钮不受影响

技术：
- [ ] 纯函数 `resolveNextPlayableEpisode` / `shouldShowAutoplayPromptCard` / `autoplayCountdownTickRemaining` / `TvSeriesAutoplaySetting.parse` 全部有单测覆盖
- [ ] 自动切路径上 `reportTvSeriesHistory` 显式传 `completed = true`
- [ ] 提示卡和覆盖层焦点抢占使用 `LaunchedTvInitialFocus + tryRequestFocus`，不裸调 `requestFocus()`
- [ ] 提示卡和覆盖层圆角 = `AppChrome.SurfaceShape`（16dp）；按钮如使用 chip 高度走 `AppChrome.ChipShape`（8dp）
- [ ] 提示卡进场/退场 fade 使用 `TvMotionTokens.DurationStandardMs` + `EasingStandard`
- [ ] 文本前景对比度 ≥ 7.0:1（按 WCAG AAA）
- [ ] `TvShapeAuditTest` / `TvFocusSpecTest` / `TvTypographySpecTest` 通过

工程：
- [ ] `:tv-app:testDebugUnitTest` 全绿
- [ ] `:tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] `:tv-app:assembleRelease` BUILD SUCCESSFUL
- [ ] `versionCode +1` / `versionName +0.0.1`
- [ ] `CONTEXT.md` 已沉淀八条新术语
- [ ] `plan.md` 已追加 reverse-chronological 进度
- [ ] git 提交信息使用中文，无乱码

## 6. 上线后观察点

- 用户反馈"自动连播太快 / 太慢"——10s 是否合适，是否要做设置项
- 用户反馈"自动连播提示影响观影"——是否需要更轻量的样式（如顶部小条而非右下角抢焦点卡）
- 用户反馈"连播跨季会不会跳过我没看过的特别篇 / 番外"——`playable=true` 但用户视角属于"非正片"的集如何处理（未来可能加 `episode_type` 字段，本期不解决）
- 关闭开关的用户占比——指导未来是否要默认关 / 引入 per-series 覆盖
- 后端历史接口接收 `completed=true` 比例是否显著上升——验证 H1 / H2 路径
- 整剧最后一集"全剧已播完"覆盖层的"返回详情"点击率——是否需要扩展功能（如推荐相似剧）

## 7. 上线开关 / 回退预案

- **无 feature flag**：改动作用域仅 TV 客户端 UI 与 DataStore，不依赖后端兼容
- **回退方式**：revert 提交即可；老用户保留的 `tv_series_autoplay_enabled` DataStore key 不会造成问题（key 闲置无害）
- **不需要分阶段灰度**：TV App 通过 APK 分发，无 OTA 灰度通道；如有严重问题，发布修复版本覆盖

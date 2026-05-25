# PRD：TV 长视频播放器从 Media3 迁移到 LibVLC（ASS 完整渲染）

- 日期：2026-05-25
- 目标端：`android-tv-app/`（TV App，独立工程）+ 后端 `internal/services/subtitle.go` 与 `pkg/ffmpeg` 字幕处理链
- 范围：TV 长视频（电影 / `18+` / 电视剧）的播放器内核与字幕渲染管线；IPTV 路径不动

## 1. 用户故事

作为 TV 用户，我希望：

1. 看番剧 OP/ED 时，**ASS 字幕的卡拉 OK 时间高亮**能正确显示（`{\k}`/`{\kf}` 逐字着色随音乐推进）。
2. 看字幕组特效番时，**ASS 字幕的动态特效**能正确显示（`{\fad}` 淡入淡出 / `{\move}` 移动 / `{\pos}` 精确定位 / `{\t}` 时间渐变 / `{\frx}` 旋转）。
3. 看 ASS 装饰场景时，**ASS 字幕的矢量绘图与多字体回退**能正确显示（`{\p1}` 绘图原语、`Style:` 段定义的多语种字体）。
4. 同时保留现有 TV 长视频播放器的**所有交互契约**：[[TV 操作 UI 层]] / [[BACK 优先收 UI]] / [[controls 焦点环绕]] / [[续播提示卡]] / [[连播提示卡]] / [[手动下一集按钮]] / 字幕选择 / 音轨选择 / 5/10/15/20/30 秒 [[快进/快退步长]] / [[连按合并跳转]] / 长按 2x 倍速 / 触屏 tap/drag/double-tap / [[TV 播放器退出确认]] —— 零退步。

## 2. 作用域

- **覆盖**：
  - TV `LongFormVideoPlayer.kt`（电影 / `18+` / 电视剧共用 Composable）的内核从 `androidx.media3.exoplayer.ExoPlayer` 切到 `org.videolan.libvlc.MediaPlayer`
  - `LongFormSubtitleSupport.kt` 字幕注入路径从 `MediaItem.SubtitleConfiguration` 改为 LibVLC `addSlave(SUBTITLE, uri, true)`
  - `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` viewmodel 中 `ExoPlayer.Builder(...)` 改为 LibVLC `MediaPlayer(libVLC)` 构造
  - 后端 `internal/services/subtitle.go` 取消 ASS/SSA → WebVTT 强转分支
  - 后端 `pkg/ffmpeg` 抽取内嵌字幕时保留原始 ASS（仅 `mov_text` 等无 Style 格式才转 VTT）
- **不覆盖**：
  - IPTV (`TvIptvScreen.kt`) — 已经在 LibVLC，且播放模型（无 seek / 无续播 / 无字幕轨切换）不同
  - 手机端 `android-app/` — 独立 gradle 工程，继续 Media3；手机端拿到 ASS 后 `SsaParser` 仍只解析对话文本（与今天等价）
  - admin-web 字幕上传 UI — 后端接受的格式不变（仍允许 .ass/.ssa），只是落库格式变
  - DRM / 加密视频 — 项目无此需求

## 3. 非目标（明确不做）

- **N1**：客户端为 SRT/VTT 字幕注入默认 ASS Style — 接受 libass 默认外观，用户反馈"字小/描边细"再独立子任务补救
- **N2**：手机端同步切到 LibVLC — 独立工程独立决策；本任务不动
- **N3**：双 player 共存 + feature flag fallback — Q4 决议 X+Y：不抽象、不留 Media3 fallback。回滚靠 git revert
- **N4**：Player 抽象接口 — Q4 决议不引入
- **N5**：分阶段上线（按播放屏拆 / 按 client-server 拆） — Q8 决议一次性合
- **N6**：用户偏好"硬解/软解开关" UI — 本任务先用代码常量；上线后真出问题再补 settings
- **N7**：管理端字幕样式自定义 UI — libass 接管渲染后描边/字号偏好不再受 Compose 控制，相关偏好 UI 移除
- **N8**：DataStore 中既有 `selectedSubtitleTrackId` / `selectedAudioTrackId` 反向迁移 — 升级时清空，用户重新选一次
- **N9**：服务端反向迁移历史 vtt 文件回 ASS — 老视频字幕维持现状；用户想要新效果重新上传
- **N10**：调整 IPTV 现有 LibVLC 配置（软解 + TextureView）— IPTV 路径完全不动
- **N11**：APK 体积优化（LibVLC native lib 减体积、按需加载 .so 等） — 本任务不做；接受 Media3 删除后约 -3~5MB 净减体积

## 4. 术语对应（待沉淀到 CONTEXT.md）

本任务涉及如下术语，**实施完成时**才 sync 到 CONTEXT.md（避免未实施先沉淀导致漂移）：

| 术语 | 一句话定义 |
|------|----------|
| `TV 长视频 LibVLC 内核` | TV 长视频播放器（电影 / `18+` / 电视剧共用的 `LongFormVideoPlayer.kt`）的播放引擎；从 Media3 ExoPlayer 切换而来。视频解码、字幕渲染、音轨切换均走 LibVLC。与 [[IPTV LibVLC 路径]] 共用同一份 `libvlc-all` 依赖与同一个 LibVLC 实例。 |
| `libass 自渲染字幕` | TV 长视频 LibVLC 内核的字幕渲染路径：所有外挂与内嵌字幕都通过 LibVLC `addSlave(SUBTITLE, uri, true)` 注入或自动加载，由 LibVLC 内置的 libass 直接画到视频 Surface 上。ASS 文件的 `[V4+ Styles]` 段、卡拉 OK、动态特效、矢量绘图、多字体回退全部按文件意图渲染；SRT/VTT 等无样式格式接受 libass 默认外观。客户端不再用 `CaptionStyleCompat` / `SubtitleView` 控制字幕样式。 |
| `TV 长视频 TextureView 硬解默认` | TV 长视频 LibVLC 内核的 Surface 与解码策略：视频输出走 `VLCVideoLayout` + TextureView（与 [[IPTV LibVLC 路径]] 同 Surface 类型保证 Compose `AndroidView` 协同），解码偏好开启 mediacodec 硬解（`Media.setHWDecoderEnabled(true, true)`）。与 IPTV 的"TextureView + 软解"区分：长视频源经后端 `pkg/ffmpeg` 主动转码，codec 边界确定（H.264 + HEVC），硬解兼容性远好于 IPTV 第三方流。 |
| `ASS 字幕原文存储策略` | 后端字幕上传与抽取的新策略：上传的 .ass/.ssa 文件**不再**用 ffmpeg 转 WebVTT，直接以 ASS 原文落库，`format=ass` / `mime_type=text/x-ssa`。视频内嵌字幕抽取时，若内嵌格式为 ASS（典型 mkv 场景）也保留 ASS 原文；`mov_text` 等不带 Style 段的内嵌格式继续转 VTT。替代旧的「外挂 ASS/SSA 字幕上传策略」（强转 VTT）。本策略**不**反向迁移历史 vtt 文件——老视频维持现状，新上传走新路径。 |
| `字幕样式 libass 让位` | TV 长视频从 Media3 切到 LibVLC 后，字幕外观控制权从 Compose / `CaptionStyleCompat` 完全移交给 libass：ASS 文件按其 `[V4+ Styles]` 段画，SRT/VTT 等无样式格式接受 libass 默认外观（白字默认描边）。原 `applyLongFormSubtitleStyle()` 代码删除；客户端不再有"统一描边" UI 偏好。 |
| `LibVLC track id 不稳定` | LibVLC `MediaPlayer.getAudioTracks()` / `getSpuTracks()` 返回的 track id 在 Media 重新加载时不保证稳定，禁止把 track id 写入 DataStore 或后端。客户端字幕/音轨偏好持久化必须改为 language code + 偏好类型（"default" / "forced"），下次播放时按 language by-name match 选轨。 |

## 5. 关键决策表（grill-with-docs 八问的结果）

| # | 决策点 | 选项 | 理由 |
|---|--------|------|------|
| Q1 | ASS 渲染能力下限 | **完整支持**（基础排版 + 卡拉 OK + 动态特效 + 矢量绘图 + 多字体） | 用户全选；任何非完整方案都覆盖不到番剧 OP/ED 卡拉 OK 与字幕组特效番场景 |
| Q2 | 重写覆盖范围 | **电影 + `18+` + 电视剧 一起切**；IPTV 不动 | 三类共享 `LongFormVideoPlayer`；分阶段需双实现，与 Q4 矛盾 |
| Q3 | 字幕渲染路径 | **A. LibVLC 自渲染（libass）** | 工程量最小；自接 libass JNI 性价比极低；libass 是工业唯一答案 |
| Q4 | Player 抽象 + Media3 fallback | **X+Y：不抽象、不留 fallback** | 项目无 Player 切换历史，引入接口 YAGNI；fallback 需要双实现成本爆炸；回滚靠 git revert |
| Q5 | 底层 Surface + 解码 | **TextureView + 硬解默认** | TextureView 与 IPTV 一致保证 Compose 协同；长视频源 codec 可控，硬解兼容性 OK |
| Q6 | SRT/VTT 字幕样式 | **接受 libass 默认外观** | 用户敏感再独立子任务补救 |
| Q7 | 服务端 ASS→VTT 转换 | **P. 不再转，只存 ASS 原文** | 手机端能力今天/明天等价；不反向迁移老数据 |
| Q8 | 任务拆解 | **一次性合 + 内部 PoC 前置** | 中间态无价值；PoC 不单独成任务 |

## 6. 验收标准（QA scenarios）

### 6.1 ASS 复杂样式渲染（PRD §1.1 / §1.2 / §1.3）

准备样本 ASS 字幕文件（典型番剧字幕组制作，含 `[V4+ Styles]` 段 + 卡拉 OK + 特效）。

- **A1**：电视剧某集挂上含 `{\k}` 卡拉 OK 标记的 ASS 字幕，播放 OP
  - ✓ 歌词逐字着色，高亮颜色与时间跟音乐节拍匹配
  - ✓ 主色 / 副色 / 边色按 `[V4+ Styles]` 段定义渲染
- **A2**：电影挂上含 `{\fad(500,500)}` + `{\pos(960,100)}` 标题特效的 ASS 字幕
  - ✓ 标题在指定坐标位置淡入 500ms、停留、淡出 500ms
- **A3**：番剧挂上含 `{\move(0,0,1920,0,0,2000)}` 横向滚动 + `{\frz45}` 旋转 45° 的特效行
  - ✓ 文字按指定路径在 2s 内匀速移动；旋转角度准确
- **A4**：番剧挂上含 `{\p1}m 0 0 l 100 0 100 100 0 100{\p0}` 矢量绘图行
  - ✓ 渲染出 100×100 矩形图形（不被当作纯文本输出）
- **A5**：番剧挂上含 `Style: Default,微软雅黑,...` 段并指定 `{\fn思源黑体}` 行内字体覆盖
  - ✓ Default 行用微软雅黑；显式覆盖行用思源黑体；缺字体时 libass 回退到系统等宽字体而非显示"豆腐块"

### 6.2 现有 TV 长视频契约零回归（PRD §1.4）

按 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote/prd.md` §6.1–§6.7（A1–G6）全场景重跑：

- **B1**：[[TV 操作 UI 层]] / [[BACK 优先收 UI]] / [[controls 焦点入口]] / [[controls 焦点环绕]] / [[controls 焦点退出键]] 全部沿用 — 这些是纯 Compose 层，零回归
- **B2**：[[续播提示卡]] 抢焦点 / 倒计时 / 永久 dismiss 信号沿用
- **B3**：[[连播提示卡]] / [[连播覆盖层]] / [[电视剧自动连播]] T-10 倒计时沿用 — `STATE_ENDED` 替换为 LibVLC `MediaPlayer.Event.EndReached`
- **B4**：[[快进/快退步长]] 5/10/15/20/30 秒选项沿用；[[连按合并跳转]] 300ms 防抖沿用
- **B5**：长按 2x 倍速 — LibVLC `setRate(2.0f)` 替换 `PlaybackParameters(2f)`
- **B6**：触屏 tap 切换 controls / drag scrub / double-tap 切换播放暂停沿用
- **B7**：[[TV 播放器退出确认]] 沿用

### 6.3 字幕轨与音轨切换（PRD §1.4）

- **C1**：电视剧某集有两条内嵌音轨（日语 / 中文配音）+ 一条外挂 ASS 字幕
  - ✓ 字幕菜单显示 ASS 字幕条目，选中后 libass 渲染
  - ✓ 音轨菜单显示两条音轨，切换响应 < 300ms
- **C2**：进入下一集（[[电视剧自动连播]] 触发）
  - ✓ 上一集的 track id 不持久化；新一集按 language code + 偏好类型 by-name match 选轨
  - ✓ 老 DataStore 中的 trackId 已清空，不影响新会话
- **C3**：外挂字幕从 ASS / SRT / VTT 三种格式同时存在的视频，切换字幕
  - ✓ ASS 切到后看到完整样式；SRT/VTT 切到后看到 libass 默认外观（白字默认描边）

### 6.4 续播 / 自动连播 / EndReached（PRD §1.4）

- **D1**：上次看到 47:23 退出，重新进入
  - ✓ [[续播提示卡]] 显示 47:23 位置；点"继续观看"从该位置 seek
- **D2**：电视剧某集还剩 10 秒，[[连播提示卡]] 应显示
  - ✓ T-10 倒计时基于 LibVLC `time-changed` 事件 250ms 轮询同步
- **D3**：视频播放到结束
  - ✓ LibVLC `EndReached` 事件触发 → [[连播覆盖层]] 或自动跳下一集（按 [[电视剧自动连播]] 偏好）

### 6.5 性能与稳定性

- **E1**：4K HEVC mp4 在中端 Android TV 设备（Mi Box S 或同级）播放
  - ✓ 硬解开启，CPU 占用 < 60%
  - ✓ ASS 字幕渲染不掉帧
- **E2**：连续切换 10 集电视剧（不退出播放器）
  - ✓ 无内存泄露（LibVLC MediaPlayer 正确 release）
  - ✓ 无 ANR

### 6.6 服务端 ASS 原文存储（PRD §2 后端范围）

- **F1**：上传 .ass 文件
  - ✓ DB `video_subtitles.format = ass` / `mime_type = text/x-ssa`
  - ✓ 文件系统保留原 .ass 文件，**不**生成 .vtt
- **F2**：上传 .ssa 文件
  - ✓ 同上，仅扩展名不同
- **F3**：上传 .srt / .vtt 文件
  - ✓ 维持原 format（`srt` / `vtt`），无任何转换
- **F4**：上传含 mkv ASS 内嵌字幕的视频
  - ✓ 抽取出的字幕落 `format = ass`，不转 VTT
- **F5**：上传含 mp4 `mov_text` 内嵌字幕的视频
  - ✓ 抽取出的字幕落 `format = vtt`（`mov_text` 无 Style 段，转 VTT 不损失信息）

## 7. 非功能要求

- **I1**：`media3-exoplayer:1.4.1` / `media3-exoplayer-hls:1.4.1` / `media3-ui:1.4.1` 三条依赖**完全删除**（仅保留 [[IPTV LibVLC 路径]] 与本任务用到的 `libvlc-all:3.6.0`）
- **I2**：DataStore key `selectedSubtitleTrackId` / `selectedAudioTrackId` 升级时清空；新模型用 language code + 偏好类型（"default" / "forced" / "commentary"）持久化
- **I3**：LibVLC 实例（`org.videolan.libvlc.LibVLC`）全应用单例；长视频 `MediaPlayer` 与 IPTV `MediaPlayer` 共享同一个 `LibVLC` 实例
- **I4**：长视频 `MediaPlayer` 构造参数与 IPTV `MediaPlayer` **不共用 helper**——长视频走硬解，IPTV 走软解，配置代码分别落 `core/player/TvLongFormVlcPlayer.kt` 与现有 `feature/tv/TvIptvScreen.kt`
- **I5**：上传 ASS 字幕需要后端安全校验：sanitize `\fn` 字段不指向本地路径、`\\fad` 等参数范围合法；防 libass 历史 CVE
- **I6**：版本号：`android-tv-app/tv-app/build.gradle.kts` `versionCode` +1（67 → 68）、`versionName` 末位 +1（0.1.67 → 0.1.68）
- **I7**：服务端 `internal/services/subtitle.go` 改动不影响既有 vtt 文件读取路径；老视频字幕维持现状

## 8. 测试策略

### 8.1 PoC 阶段（实施前置，不单独成任务）

在分支 `chore/libvlc-poc` 跑通：

1. LibVLC `MediaPlayer` + `VLCVideoLayout` + TextureView 在 Compose `AndroidView` 中正常显示 4K HEVC mp4
2. `addSlave(SUBTITLE, ass_url, true)` 后 libass 自渲染 ASS 字幕（含卡拉 OK / `\fad` / `\pos` 验证片段）
3. `setAudioTrack(id)` / `setSpuTrack(id)` 切轨能力通
4. 硬解开启时 CPU < 80%（中端 Android TV 设备）

任一不通过则回到 grill 重选方案。**PoC 不进主线、不留代码**。

### 8.2 单测（必须）

- `tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcConfigTest.kt`：纯函数构造 `MediaPlayer` 启动参数、`Media.addOption` 列表、硬解开关、字幕同步偏移
- `tv-app/src/test/java/com/chee/videos/core/ui/LongFormSubtitleSupportLibVlcTest.kt`：字幕 by-language 选轨逻辑、language code normalize、preferred-type fallback
- 后端 `internal/services/subtitle_test.go`：扩展现有测试，断言 .ass/.ssa 上传后 `format=ass`，不生成 .vtt 副产物；mkv 内嵌 ASS 抽取后保留 ass

### 8.3 源文 audit（必须）

`tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcSpecTest.kt`：

- `LongFormVideoPlayer.kt` 不出现 `androidx.media3.*` import
- `LongFormVideoPlayer.kt` 不出现 `ExoPlayer` / `PlayerView` / `MediaItem.SubtitleConfiguration` / `CaptionStyleCompat`
- `build.gradle.kts` 不出现 `media3-` 字样
- `LongFormVideoPlayer.kt` 必须 import `org.videolan.libvlc.MediaPlayer` 与 `VLCVideoLayout`
- 调用点必须 `setHWDecoderEnabled(true, true)`（硬解默认）
- 字幕注入必须经过 `addSlave(IMedia.Slave.Type.Subtitle, ..., true)`

### 8.4 androidTest（编写，不强求执行）

`tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerLibVlcTest.kt`：

- 基础冷启 + 字幕注入 + seek 三个场景
- 接受"本机无设备时仅编译通过"（沿用 `connectedDebugAndroidTest` 既有约定）

### 8.5 手测（review.md 详写，必须真机过一遍）

在 Android TV 真机或模拟器上覆盖 §6 全部 A–F 场景。本任务**必须**有一次真机回归记录在 DONE.md 才能合 PR。

## 9. 不引入的依赖 / 改动

- 不引入 mpv / Vitamio / VLC4J 等其他播放器栈
- 不引入新的 native lib（除 `libvlc-all` 已含）
- 不修改 `pkg/ffmpeg` 转码 / `internal/services/transcode.go` 等视频转码流程
- 不修改 admin-web 字幕上传 UI / 视频管理列表
- 不修改 `android-app/`（手机端）任何代码

## 10. 关联 ADR

- `docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md` — 决策记录：为什么 TV 长视频从 Media3 切到 LibVLC

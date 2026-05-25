# Review：TV 长视频 LibVLC 迁移

- 日期：2026-05-25
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`
- 关联 ADR：`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`

## 0. 准入条件

未满足任一项 → **回到 implement 阶段**：

- [ ] PoC 阶段 7 条全部跑通（implement §5）；任一不通过此任务 STOP
- [ ] 后端单测 `go test ./internal/services -run TestSubtitle` 全绿
- [ ] TV 单测 `cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest` 全绿（含新增 `TvLongFormVlcConfigTest` / `TvLongFormTrackSelectionTest` / `TvLongFormVlcSpecTest` / `LongFormSubtitleSupportLibVlcTest`）
- [ ] TV 构建 `cd android-tv-app && ./gradlew :tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] androidTest 编译通过（设备执行非阻塞）：`cd android-tv-app && ./gradlew :tv-app:assembleDebugAndroidTest`
- [ ] `build.gradle.kts` **不**包含 `media3-` 字样（grep 验证）
- [ ] `LongFormVideoPlayer.kt` **不**包含 `androidx.media3.` import（grep 验证）
- [ ] `git diff` 范围仅触及 implement.md §3 文件清单
- [ ] `CONTEXT.md` 已 sync 实施完成后的 6 条新术语 + 推翻 line 9 / line 170 旧约定
- [ ] `docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md` 已落地
- [ ] `android-tv-app/tv-app/build.gradle.kts` `versionCode = 68`、`versionName = "0.1.68"`
- [ ] `plan.md` 追加进度条目
- [ ] **真机回归**：至少在一台 Android TV 真机或模拟器跑完 §1 全部场景，结果写入 DONE.md

## 1. 手测脚本（按 PRD §6 验收映射）

### 1.1 ASS 复杂样式渲染 — PRD §6.1

准备 ASS 测试素材包（含卡拉 OK / `\fad` / `\pos` / `\move` / `\frz` / `\p1` / 多字体）。

1. **A1**：电视剧某集挂卡拉 OK ASS 字幕，播放 OP
   - [ ] 逐字高亮颜色按 `\k` 时间推进
   - [ ] 高亮色 / 副色 / 边色按 `[V4+ Styles]` 渲染
2. **A2**：电影挂 `{\fad(500,500)}{\pos(960,100)}` 标题特效
   - [ ] 文字在指定坐标淡入 500ms / 停留 / 淡出 500ms
3. **A3**：番剧挂 `{\move(0,0,1920,0,0,2000)}{\frz45}` 滚动旋转
   - [ ] 文字 2s 内匀速移动并旋转 45°
4. **A4**：番剧挂 `{\p1}` 矢量绘图行
   - [ ] 渲染为图形而非纯文本
5. **A5**：番剧挂指定 `Style:` 字体 + `{\fn}` 行内覆盖
   - [ ] 各行按指定字体渲染；缺字体时回退到系统字体（无"豆腐块"）

### 1.2 现有 TV 长视频契约零回归 — PRD §6.2

按 `tasks/2026-05-24-tv-long-form-operation-ui-on-remote/review.md` 的 A1–G6 全部场景重跑一遍。

- [ ] PRD §6.1 A1–A4 全部成立
- [ ] PRD §6.2 B1–B5 全部成立
- [ ] PRD §6.3 C1–C3 全部成立
- [ ] PRD §6.4 D1–D6 全部成立
- [ ] PRD §6.5 E1–E4 全部成立
- [ ] PRD §6.6 F1–F4 全部成立
- [ ] PRD §6.7 G1–G6 全部成立

### 1.3 字幕轨与音轨切换 — PRD §6.3

1. **C1**：电视剧某集（两条内嵌音轨 + 一条外挂 ASS）
   - [ ] 字幕菜单显示 ASS 条目，选中后 libass 完整渲染
   - [ ] 音轨菜单显示两条；切换响应 < 300ms
2. **C2**：自动连播跳下一集
   - [ ] 新一集按 language code by-name 选轨；老 trackId 已清空不干扰
3. **C3**：同时存在 ASS / SRT / VTT 字幕
   - [ ] ASS：完整样式
   - [ ] SRT / VTT：libass 默认外观（白字默认描边）

### 1.4 续播 / EndReached / 自动连播 — PRD §6.4

1. **D1**：47:23 退出后重进
   - [ ] [[续播提示卡]] 显示位置；从该位置 seek
2. **D2**：电视剧某集 T-10
   - [ ] [[连播提示卡]] 显示，倒计时基于 LibVLC `time-changed` 同步
3. **D3**：播放到结束
   - [ ] LibVLC `EndReached` 触发 → [[连播覆盖层]] 或自动跳

### 1.5 性能 — PRD §6.5

1. **E1**：4K HEVC mp4 在中端 Android TV 设备
   - [ ] 硬解开启，CPU < 60%
   - [ ] ASS 字幕渲染不掉帧
2. **E2**：连续切换 10 集电视剧
   - [ ] Android Profiler 看 LibVLC native 内存稳定
   - [ ] 无 ANR

### 1.6 服务端 ASS 原文存储 — PRD §6.6

1. **F1**：上传 .ass 文件
   - [ ] DB `format=ass` / `mime_type=text/x-ssa`
   - [ ] 文件系统保留 .ass，**不**生成 .vtt
2. **F2**：上传 .ssa 文件
   - [ ] 同 F1，扩展名差异
3. **F3**：上传 .srt / .vtt
   - [ ] 维持原 format
4. **F4**：上传含 mkv ASS 内嵌字幕的视频
   - [ ] 抽取后落 `format=ass`
5. **F5**：上传含 mp4 mov_text 内嵌字幕的视频
   - [ ] 抽取后落 `format=vtt`

## 2. 代码审计（diff 检查）

### 2.1 必须出现

- [ ] `core/player/TvVlcLibrary.kt` 新增，单例 `LibVLC`
- [ ] `core/player/TvLongFormVlcPlayer.kt` 新增，`buildLongFormMedia` + `newLongFormMediaPlayer` 纯函数 / 工厂
- [ ] `core/player/TvLongFormVlcSurface.kt` 新增，`VlcLongFormSurface` Composable + `attachViews(useTextureView=true)` + `setHWDecoderEnabled(true, true)`
- [ ] `core/player/TvLongFormTrackSelection.kt` 新增，by-language 选轨纯函数
- [ ] `LongFormVideoPlayer.kt`：
  - [ ] 签名 `player: MediaPlayer`（LibVLC，不是 Media3）
  - [ ] 使用 `VlcLongFormSurface` 而非 `AndroidView(PlayerView)`
  - [ ] 事件监听是 `MediaPlayer.EventListener`（不是 `Player.Listener`）
  - [ ] `seekTo` / `time` / `length` 全部经 LibVLC API
  - [ ] `setRate` 替换 `PlaybackParameters`
- [ ] `LongFormSubtitleSupport.kt`：
  - [ ] `applyLongFormMediaSource(libVLC, mediaPlayer, sourceUrl, baseUrl, selectedSubtitleTrack)` 新增
  - [ ] 字幕注入走 `addSlave(IMedia.Slave.Type.Subtitle, url, true)`
- [ ] `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt`：
  - [ ] `ExoPlayer.Builder` → `MediaPlayer(TvVlcLibrary.shared(...))`
- [ ] `TvLongFormPlayerViewModel.kt` / `TvSeriesPlayerViewModel.kt`：
  - [ ] State 字段 `selectedSubtitleLanguage` / `selectedSubtitleType` / `selectedAudioLanguage` / `selectedAudioType` 替换原 trackId 字段
- [ ] DataStore key 重命名 + 升级清除老 key 的逻辑
- [ ] `internal/services/subtitle.go`：
  - [ ] `case "ass"` / `case "ssa"` 分支：`StoredFormat=ass`、`NeedsWebVTT=false`
  - [ ] 内嵌字幕抽取按 codec switch
- [ ] `pkg/ffmpeg/subtitle.go`：`ExtractSubtitleToAss` 新增
- [ ] `internal/services/subtitle_test.go` 扩展用例：.ass/.ssa 上传 + mkv ASS 内嵌抽取
- [ ] `tv-app/build.gradle.kts`：
  - [ ] **删除** `media3-exoplayer` / `media3-exoplayer-hls` / `media3-ui` 三条
  - [ ] `versionCode = 68`、`versionName = "0.1.68"`
- [ ] `CONTEXT.md`：
  - [ ] line 9 旧条目替换为 [[ASS 字幕原文存储策略]]
  - [ ] line 170 旧条目改写，删除 "其他长视频播放器继续使用 Media3"
  - [ ] 追加 6 条新术语（PRD §4）
- [ ] `docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md` 落地
- [ ] `plan.md` 追加进度条目

### 2.2 禁止出现

- [ ] `LongFormVideoPlayer.kt` **不**包含 `import androidx.media3.` 任意子包
- [ ] `LongFormVideoPlayer.kt` **不**包含 `ExoPlayer` / `PlayerView` / `MediaItem` / `CaptionStyleCompat` / `Player.Listener` / `PlaybackParameters`
- [ ] `build.gradle.kts` **不**包含 `media3-` 字符串
- [ ] **不**新增 `LongFormPlayer` 抽象接口（Q4 决议）
- [ ] **不**保留 ExoPlayer 实现作 fallback（Q4 决议）
- [ ] **不**新增 `IPTV LibVLC 配置`复用（IPTV 用软解 / 长视频用硬解，配置不共享 helper）
- [ ] **不**为 SRT/VTT 字幕注入 ASS 默认 Style（N1 决议）

### 2.3 不变量保护

- [ ] [[TV 操作 UI 层]] / [[BACK 优先收 UI]] / [[controls 焦点环绕]] / [[controls 焦点入口]] / [[controls 焦点退出键]] 视觉与契约零回归
- [ ] [[续播提示卡]] / [[连播提示卡]] / [[连播覆盖层]] / [[电视剧自动连播]] 行为零回归
- [ ] [[TV 焦点 ISE 三层防线]] 沿用（不被 LibVLC surface 改动破坏）
- [ ] [[TV 主 Looper FocusRequester 未初始化兜底]] 沿用
- [ ] [[IPTV LibVLC 路径]] 配置零变化（继续 TextureView + 软解）
- [ ] [[TV 文本溢出保护]] 在 [[左上信息层]] 等位置继续生效

## 3. 性能与边界

- [ ] 4K HEVC mp4 硬解：中端 Android TV 设备 CPU < 60% / 不掉帧
- [ ] 进入 / 退出播放屏 20 次：LibVLC native 内存稳定（Profiler 可观察）
- [ ] 连续切换 10 集（电视剧自动连播）：内存稳定 / 无 ANR
- [ ] 字幕注入后切换字幕轨：响应 < 300ms
- [ ] 音轨切换：响应 < 300ms
- [ ] 长按 ←/→ 持续 10 秒：seek 防抖正确 / UI 全程可见 / 5 秒计时重置每次按键

## 4. 回归扫描区

- [ ] **IPTV** (`TvIptvScreen`)：依然是软解 + TextureView，频道切换 / 播放 / 黑屏诊断契约零变化
- [ ] **手机端** (`android-app/`)：零改动（独立工程）；如有共享后端契约（字幕上传），手机端拿到 ASS 后用 Media3 SsaParser 显示纯文本（与今天等价）
- [ ] **admin-web 字幕上传**：上传 .ass/.ssa 流程零变化；提示文案需要确认是否要更新（当前文案 "ASS 会转为 VTT，复杂样式丢失"——若有此类文案应更新为"ASS 原样保留，TV 端能完整渲染")
- [ ] **管理端视频管理**：字幕列表显示 `format` 列若有，从 vtt 变为 ass，UI 是否需要调整？检查

## 5. ADR 完整性

- [ ] ADR 标题清晰、决策可读
- [ ] 列出"考虑过的替代方案"段（至少 3 个：保留 Media3 + libass overlay；服务端预渲染 ASS→PGS；继续 ASS→VTT 转换）
- [ ] 列出"关联"段（PRD 路径、CONTEXT.md 相关术语、影响的代码文件）

## 6. 不在 review 范围

- 手机端 (`android-app/`) 字幕渲染体验提升 — 独立工程
- LibVLC 自定义 native 编译减体积 — 工程化优化，独立任务
- ASS 字幕样式定制 UI（用户偏好字号 / 颜色）— libass 让位决议下不做
- 服务端 ASS 反向迁移老 vtt 数据 — N9 决议不做
- 引入硬解/软解用户偏好开关 — N6 决议不做

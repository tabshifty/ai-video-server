# TV 长视频播放内核从 Media3 切换到 LibVLC，由 libass 自渲染 ASS 字幕

> 状态：已被 `0013-tv-long-form-exoplayer-unification.md` 取代。该 ADR 记录迁移前决策背景；当前 TV 长视频目标是统一 ExoPlayer 内核，ASS/SSA 不再承诺 libass 级保真，LibVLC 仅保留给 IPTV。

TV 长视频（电影 / `18+` / 电视剧共用的 `LongFormVideoPlayer.kt`）原本基于 `androidx.media3:1.4.1`（ExoPlayer + PlayerView）。Media3 内置的 `SsaParser` 只解析 ASS/SSA 的对话文本行，**完整丢弃**样式、特效、卡拉 OK、矢量绘图、字体覆盖；ExoPlayer `SubtitleView` 也无法承载 ASS 的复杂渲染能力。为支撑用户场景里的卡拉 OK 番剧、字幕组特效番、ASS 装饰场景，决定**把 TV 长视频内核换成 `org.videolan.android:libvlc-all:3.6.0`**，并让字幕渲染完全由 LibVLC 内置的 libass 接管。

服务端同步取消既有的 `ASS/SSA → WebVTT` 强转链路，DB 落 `format=ass` / `mime_type=text/x-ssa` 原文存储；视频内嵌字幕抽取时 `ass`/`ssa` codec 保留原 ASS、`mov_text` 等无 Style 格式继续 VTT。这一服务端改动是本决策的对偶项——只有 TV 客户端能渲染 ASS、服务端不再多此一举降级，整个用户故事才闭环。

本决策**只**覆盖 TV 长视频。IPTV (`TvIptvScreen`) 早已在 LibVLC 上（软解 + TextureView），保持现状；手机端 (`android-app/`) 是独立 gradle 工程，继续 Media3，手机端拿到 ASS 原文后用 Media3 `SsaParser` 仍只解析文本（与今天 `ASS → VTT` 后的能力等价）。

替换策略**强制**不引入 Player 抽象层、不保留 Media3 fallback——`media3-exoplayer` / `media3-exoplayer-hls` / `media3-ui` 三条依赖全部删除。回滚靠 `git revert` 单 PR。理由：项目从未发生过 Player 切换的工程需求，抽象层是 YAGNI；双实现 fallback 等于把 PRD §6 的所有契约（控制条、续播、连播、字幕选择、音轨选择、退出确认、焦点环绕等）在两个 player 上 cross-check，维护负担巨大，回报为零。

## 考虑过的替代方案

- **Media3 + libass overlay 自渲染层**：保留 ExoPlayer 解视频；自接 libass JNI，把 ASS 帧画到 Compose Canvas 上叠加。理论上"最稳"，但 Android 上没有维护良好的 Kotlin libass 绑定（mpv-android 私有），自己写 JNI 等于多一个长期维护负担。卡拉 OK / `\t` / `\move` 必须跟视频时钟精确同步（错一帧就抖动），需要订阅 ExoPlayer 解码时间戳 + 自己插值 — 工程量与 LibVLC 整体替换相当，回报反而少（少了 LibVLC 自带的 demuxer / 多 codec 兼容）。
- **服务端预渲染 ASS → PGS / Bitmap 字幕轨道**：用 ffmpeg/MKVToolNix 把 ASS 烧成位图字幕嵌入到 mkv，App 端解 PGS。优点：客户端零改造（Media3 支持 PGS）。缺点：每个视频转码时间 +5~10 分钟；位图字幕只能"开/关"，多语种字幕用户切换体验退步；ASS 渲染参数（字体、缩放）烧死无法调整；用户期望"libass 准确度"也得不到（ffmpeg ASS→PGS 用的是 libass，但烧 bitmap 后失去了"动态字号 / 用户偏好"能力）。
- **继续 `ASS → VTT` 转换 + 接受能力上限**：成本最低（不改任何代码）。但用户场景里的卡拉 OK / `\fad` / 矢量绘图全部无法满足，相当于"接受永远不支持复杂 ASS"。与用户需求直接冲突。
- **保留 Media3 作 fallback，feature flag 灰度**：保留 ExoPlayer 实现，新增 LibVLC 实现，运行时按设备/偏好切换。优点：紧急情况可关 LibVLC 回退。缺点：双实现 = 双套字幕注入 / 双套切轨 / 双套续播同步 / 双套退出确认；任一边出 bug 都算回归；维护成本爆炸。回退应该靠 `git revert`，不是 feature flag 长期共存。

## 关联

- `tasks/2026-05-25-tv-long-form-libvlc-migration/prd.md` — 完整需求与验收
- `tasks/2026-05-25-tv-long-form-libvlc-migration/implement.md` — 实施步骤、PoC 阶段、已知陷阱
- `tasks/2026-05-25-tv-long-form-libvlc-migration/review.md` — 准入条件、回归扫描
- `CONTEXT.md` 新术语：[[TV 长视频 LibVLC 内核]] / [[libass 自渲染字幕]] / [[TV 长视频 TextureView 硬解默认]] / [[ASS 字幕原文存储策略]] / [[字幕样式 libass 让位]] / [[LibVLC track id 不稳定]]
- `CONTEXT.md` 推翻：line 9（旧 [[外挂 ASS/SSA 字幕上传策略]]）/ line 170（旧 "其他长视频播放器继续使用 Media3"）
- 影响的客户端代码：`android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` / `LongFormSubtitleSupport.kt` / `feature/tv/TvLongFormPlayerScreen.kt` / `feature/tv/TvSeriesPlayerScreen.kt` / 对应 ViewModel
- 影响的后端代码：`internal/services/subtitle.go` / `pkg/ffmpeg/subtitle.go` / `internal/services/subtitle_test.go`
- 移除的依赖：`androidx.media3:media3-exoplayer:1.4.1` / `media3-exoplayer-hls:1.4.1` / `media3-ui:1.4.1`
- 保留的依赖：`org.videolan.android:libvlc-all:3.6.0`（已为 IPTV 服务，本决策共用单例）

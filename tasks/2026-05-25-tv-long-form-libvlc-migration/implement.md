# Implement：TV 长视频 LibVLC 迁移

- 日期：2026-05-25
- 关联 PRD：`./prd.md`
- 关联 ADR：`docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md`

## 1. 总体方案

把 `LongFormVideoPlayer` 的播放内核从 `ExoPlayer` 替换为 LibVLC `MediaPlayer`，字幕渲染让位给 libass，服务端字幕上传链同步取消 ASS→VTT 转换。**不引入抽象层**、**不保留 Media3 fallback**。Compose UI 契约（[[TV 操作 UI 层]] / [[controls 焦点环绕]] / [[BACK 优先收 UI]] 等）全部保留——这些是纯 Compose 状态机，跟 player 内核解耦。

## 2. 实施顺序（强制）

```
0. PoC 验证 (chore/libvlc-poc 分支)
   └→ 通过则继续；不通过 STOP，回到 grill 重选方案
1. 后端字幕策略变更 + 单测
2. 客户端 LibVLC 内核接入 + 字幕注入
3. 客户端 Compose UI 层契约对齐
4. 删除 Media3 依赖与残留代码
5. 源文 audit + 单测 + androidTest 编写
6. 真机回归 (PRD §6 全场景)
7. CONTEXT.md 沉淀 + 版本号 + 提交
```

## 3. 文件结构

### 3.1 新增

| 路径 | 用途 |
|---|---|
| `tv-app/src/main/java/com/chee/videos/core/player/TvLongFormVlcPlayer.kt` | LibVLC `MediaPlayer` 工厂、配置（硬解开启、字幕字符集、缓存大小）、生命周期管理 |
| `tv-app/src/main/java/com/chee/videos/core/player/TvVlcLibrary.kt` | 全应用单例 `LibVLC` 实例 + 长视频用启动参数（共享 IPTV 单例池但配置不同） |
| `tv-app/src/main/java/com/chee/videos/core/player/TvLongFormVlcSurface.kt` | `@Composable VlcLongFormSurface(mediaPlayer, modifier)` 包裹 `VLCVideoLayout` + Compose `AndroidView` + 生命周期 attach/detach |
| `tv-app/src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt` | 字幕/音轨 by-language 选轨：`resolveSpuTrackByLanguage(tracks, prefLang, prefType)` / `resolveAudioTrackByLanguage(...)` 纯函数 |
| `tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcConfigTest.kt` | 构造参数 / 启动参数纯函数单测 |
| `tv-app/src/test/java/com/chee/videos/core/player/TvLongFormTrackSelectionTest.kt` | by-language 选轨纯函数单测 |
| `tv-app/src/test/java/com/chee/videos/core/player/TvLongFormVlcSpecTest.kt` | 源文 audit：禁出现 media3 / ExoPlayer / PlayerView / CaptionStyleCompat；必须 import 与硬解开启 |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormSubtitleSupportLibVlcTest.kt` | 字幕 by-name 选轨 + language code normalize |
| `tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerLibVlcTest.kt` | 冷启 + 字幕注入 + seek（需要设备执行） |
| `docs/adr/0004-tv-long-form-libvlc-for-ass-rendering.md` | 决策记录 |

### 3.2 修改

| 路径 | 改动 |
|---|---|
| `tv-app/build.gradle.kts` | 删除 `media3-exoplayer:1.4.1` / `media3-exoplayer-hls:1.4.1` / `media3-ui:1.4.1` 三条依赖；`versionCode = 68`、`versionName = "0.1.68"` |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` | 签名 `player: ExoPlayer` → `player: MediaPlayer`（LibVLC）；`AndroidView(PlayerView)` → `VlcLongFormSurface`；删除 `applyLongFormSubtitleStyle` / `CaptionStyleCompat` / `MediaItem.SubtitleConfiguration` 等所有 Media3 依赖；事件监听从 `Player.Listener.onIsPlayingChanged/onEvents` 改为 `MediaPlayer.EventListener` (`Playing / Paused / TimeChanged / EndReached / EncounteredError / VoutCount`)；`player.seekTo(ms)` 改 `mediaPlayer.time = ms`；`player.duration` → `mediaPlayer.length`；`PlaybackParameters(2f)` → `setRate(2f)`；音轨抽取 `buildLongFormAudioTracks(player.currentTracks)` 改为 `buildLongFormAudioTracksFromVlc(mediaPlayer.audioTracks)` |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt` | 删除 `buildLongFormMediaItem`；新增 `applyLongFormMediaSource(mediaPlayer, sourceUrl, selectedSubtitleTrack, baseUrl)`：构造 `Media` → 设置 `addOption(":file-caching=1500")` 等 → `addSlave(IMedia.Slave.Type.Subtitle, resolvedUrl, true)` → `mediaPlayer.media = media` |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt` | `ExoPlayer.Builder(context).build()` → `MediaPlayer(TvVlcLibrary.shared())`；状态轮询 `LaunchedEffect` 250ms 改为 LibVLC `time` polling + EventListener 双通道；DataStore 读 `selectedSubtitleTrackId` 改为读 `selectedSubtitleLanguage` (新 key) |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt` | 同上；自动连播 `STATE_ENDED` 检测改为 `MediaPlayer.Event.EndReached` |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerViewModel.kt` / `TvSeriesPlayerViewModel.kt` | `selectedSubtitleTrackId: String` 替换为 `selectedSubtitleLanguage: String?` + `selectedSubtitleType: String?`（"default" / "forced" / "commentary"）；`selectedAudioTrackId` 同模式重命名 |
| `tv-app/src/main/java/com/chee/videos/core/datastore/*.kt`（具体文件待 grep） | DataStore key 重命名；升级时旧 key 删除，不迁移 |
| `tv-app/proguard-rules.pro` | 添加 `-keep class org.videolan.libvlc.** { *; }` 保险（IPTV 已经在跑应该不需要重加，但本任务把长视频也接入后确认下规则覆盖） |
| `internal/services/subtitle.go` | `subtitleUploadPlan` 的 `.ass` / `.ssa` 分支：`StoredFormat=ass` / `StoredMIMEType=text/x-ssa` / `NeedsWebVTT=false`；`ExtractSubtitleToWebVTT` 增加 ASS 分支：如果 ffprobe 报告内嵌 codec 是 `ass`/`ssa`，调用新 `pkg/ffmpeg.ExtractSubtitleToAss` 保留原 ASS；其他（`mov_text` 等）继续转 VTT |
| `pkg/ffmpeg/subtitle.go` | 新增 `ExtractSubtitleToAss(ctx, inputPath, streamIndex, outputPath)`：`ffmpeg -i in -map 0:s:N -c:s copy out.ass`，参数化校验 |
| `internal/services/subtitle_test.go` | 扩展现有用例：.ass/.ssa 上传后断言 plan.StoredFormat == "ass"、不调用 ConvertSubtitleToWebVTT；mkv ASS 内嵌抽取后断言落 .ass |
| `internal/services/subtitle.go` | 新增 `sanitizeAssContent(raw string)` —— 处理上传 ASS 文本时 strip `\fn` 指向绝对路径、参数范围超界等防 libass CVE 的边界 |
| `migrations/NNNN_subtitle_format_ass_mime.up.sql` / `.down.sql` | （可选）若 DB schema 限制了 `mime_type` 枚举，加 `text/x-ssa` 到允许集；多数情况是 text 字段无需 |
| `CONTEXT.md` | line 9 重写 `外挂 ASS/SSA 字幕上传策略` 为 [[ASS 字幕原文存储策略]]；line 170 删除 "其他长视频播放器继续使用 Media3"，改为 "TV 长视频与 IPTV 均使用 LibVLC，但配置分轨（见 [[TV 长视频 TextureView 硬解默认]] / [[IPTV LibVLC 路径]]）"；新增 PRD §4 列出的 6 条术语 |
| `plan.md` | 追加 reverse-chronological 进度条目 |

### 3.3 删除

| 路径 | 删除原因 |
|---|---|
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleStyle.kt`（如存在） | `applyLongFormSubtitleStyle()` 整体删除，libass 接管 |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrack.kt` 中 `buildLongFormAudioTracks(Player.Tracks)` 函数 | 函数 by-name + Media3 specific，替换为 `buildLongFormAudioTracksFromVlc(MediaPlayer.TrackDescription[])` |
| 现有 ExoPlayer 单测中只测 Media3 行为的（如 `LongFormPlayerListenerTest`，如存在） | 行为已不存在 |

## 4. 关键函数 / 类型签名草稿

### 4.1 `TvVlcLibrary.kt`

```kotlin
package com.chee.videos.core.player

import android.content.Context
import org.videolan.libvlc.LibVLC

object TvVlcLibrary {
    @Volatile private var instance: LibVLC? = null

    fun shared(context: Context): LibVLC {
        return instance ?: synchronized(this) {
            instance ?: LibVLC(context.applicationContext, buildSharedArgs()).also { instance = it }
        }
    }

    private fun buildSharedArgs(): ArrayList<String> = arrayListOf(
        "--no-drop-late-frames",
        "--no-skip-frames",
        "--rtsp-tcp",
        "-vvv",
    )
}
```

### 4.2 `TvLongFormVlcPlayer.kt`

```kotlin
package com.chee.videos.core.player

import android.content.Context
import android.net.Uri
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.Media
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.interfaces.IMedia

internal data class TvLongFormVlcMediaSpec(
    val sourceUrl: String,
    val subtitleUrl: String?,
)

internal fun buildLongFormMedia(
    libVLC: LibVLC,
    spec: TvLongFormVlcMediaSpec,
): Media {
    val media = Media(libVLC, Uri.parse(spec.sourceUrl))
    // 长视频 caching：源已 ffmpeg 转码，文件级 1500ms 即可
    media.addOption(":file-caching=1500")
    media.addOption(":network-caching=2000")
    media.setHWDecoderEnabled(true, true)  // 硬解默认
    spec.subtitleUrl?.let { url ->
        media.addSlave(Media.Slave(IMedia.Slave.Type.Subtitle, 0, url))
    }
    return media
}

internal fun newLongFormMediaPlayer(libVLC: LibVLC): MediaPlayer {
    return MediaPlayer(libVLC)
}
```

### 4.3 `TvLongFormVlcSurface.kt`

```kotlin
@Composable
internal fun VlcLongFormSurface(
    mediaPlayer: MediaPlayer,
    modifier: Modifier = Modifier,
) {
    val view = remember { VLCVideoLayout(/* context */) }
    AndroidView(
        factory = { view },
        modifier = modifier,
    )
    DisposableEffect(mediaPlayer, view) {
        // useTextureView=true 强制 TextureView，跟 IPTV 保持一致
        mediaPlayer.attachViews(view, null, /* subtitlesEnabled */ true, /* useTextureView */ true)
        onDispose {
            mediaPlayer.detachViews()
        }
    }
}
```

### 4.4 `TvLongFormTrackSelection.kt`

```kotlin
package com.chee.videos.core.player

import org.videolan.libvlc.MediaPlayer

internal data class TvLongFormLanguagePreference(
    val language: String?,       // BCP-47 e.g. "zh-CN" / "ja"
    val type: String?,           // "default" / "forced" / "commentary"
)

internal fun resolveSpuTrackByLanguage(
    tracks: Array<MediaPlayer.TrackDescription>?,
    preference: TvLongFormLanguagePreference,
): Int {
    val normalized = preference.language?.trim()?.lowercase().orEmpty()
    if (normalized.isBlank() || tracks == null) return -1
    val match = tracks.firstOrNull {
        // LibVLC TrackDescription.name 通常含 language code 子串
        it.name?.lowercase()?.contains(normalized) == true
    }
    return match?.id ?: -1
}

internal fun resolveAudioTrackByLanguage(
    tracks: Array<MediaPlayer.TrackDescription>?,
    preference: TvLongFormLanguagePreference,
): Int = /* 同 SPU，逻辑可复用 */
```

### 4.5 `LongFormSubtitleSupport.kt` 重写

```kotlin
internal fun applyLongFormMediaSource(
    libVLC: LibVLC,
    mediaPlayer: MediaPlayer,
    sourceUrl: String,
    baseUrl: String,
    selectedSubtitleTrack: SubtitleTrackDto?,
) {
    val subtitleUrl = selectedSubtitleTrack
        ?.let { resolvePlaybackAssetUrl(baseUrl, it.url) }
    val spec = TvLongFormVlcMediaSpec(sourceUrl, subtitleUrl)
    val media = buildLongFormMedia(libVLC, spec)
    mediaPlayer.media = media
    media.release()  // mediaPlayer 持有 ref，本地 release 不影响播放
}
```

### 4.6 后端 `subtitle.go` 改动

```go
case "ass":
    return subtitleUploadPlan{UploadedFormat: "ass", StoredFormat: "ass", StoredMIMEType: "text/x-ssa", NeedsWebVTT: false}, nil
case "ssa":
    return subtitleUploadPlan{UploadedFormat: "ssa", StoredFormat: "ass", StoredMIMEType: "text/x-ssa", NeedsWebVTT: false}, nil
```

抽取内嵌字幕：

```go
codec := probe.CodecName
switch codec {
case "ass", "ssa":
    if err := ffmpeg.ExtractSubtitleToAss(ctx, inputPath, probe.Index, outputPath); err != nil { ... }
case "mov_text", "subrip", "webvtt":
    if err := ffmpeg.ExtractSubtitleToWebVTT(ctx, inputPath, probe.Index, outputPath); err != nil { ... }
}
```

## 5. PoC 验证清单（实施第一步）

在分支 `chore/libvlc-poc` 上跑通以下场景，**不进主线**：

1. **冷启 + 视频**：随便一段 1080p H.264 mp4，能播能 seek
2. **冷启 + 4K HEVC**：硬解模式下 Mi Box S 或同级设备 CPU < 80%、不掉帧
3. **ASS 字幕渲染**：示例 ASS（含 `{\\k}` 卡拉 OK / `{\\fad}` / `{\\pos}` / `\\p` 矢量绘图测试串）能正确显示
4. **字幕切轨**：在两个内嵌 + 一个外挂 ASS 之间通过 `setSpuTrack` 切换、`addSlave` 注入后能选中
5. **音轨切轨**：双语视频切换 `setAudioTrack`、响应时间 < 300ms
6. **EndReached 事件**：跑完一段短视频，`MediaPlayer.Event.EndReached` 能触发回调
7. **生命周期**：进入 / 退出播放屏 20 次，无内存泄露（Android Profiler 看 LibVLC native 内存稳定）

通过则继续步骤 1；不通过则 STOP 回到 grill 评估替代方案。

## 6. 步骤详细

1. **PoC**（上述）
2. **后端**：
   - 改 `subtitleUploadPlan` 的 ass/ssa 分支
   - 加 `pkg/ffmpeg.ExtractSubtitleToAss`
   - 改 `subtitle.go::ExtractSubtitleToWebVTT` 加 codec switch
   - 扩展 `subtitle_test.go`
   - `go test ./internal/services -run TestSubtitle` 转绿
3. **客户端 LibVLC 接入**：
   - 写 `TvVlcLibrary` / `TvLongFormVlcPlayer` / `TvLongFormVlcSurface` / `TvLongFormTrackSelection`
   - 写对应单测（红 → 绿）
4. **`LongFormVideoPlayer` 重写**：
   - 删除所有 `androidx.media3.*` import
   - 改造签名、状态、`AndroidView`、事件监听
   - 改造音轨/字幕选择逻辑（by-language）
5. **调用方对齐**：
   - 两个 PlayerScreen 改 `ExoPlayer.Builder` → `MediaPlayer(TvVlcLibrary.shared(context))`
   - DataStore 改名 + 旧 key 清除（一次性升级清空，PRD N8）
6. **依赖清理**：
   - `build.gradle.kts` 删 `media3-*` 三条
   - 编译应 SUCCESSFUL（如果有遗漏 import 编译会报，逐一清）
7. **源文 audit 测试**：
   - `TvLongFormVlcSpecTest` 写完跑绿
8. **真机回归**：
   - PRD §6.1 ~ §6.6 全跑
   - DONE.md 记录每场景结果
9. **CONTEXT.md 沉淀**：
   - line 9 / 170 改写
   - 追加 6 条新术语
10. **ADR + 版本号 + plan.md + commit**

## 7. 已知陷阱 / 风险

- **LibVLC native 加载时序**：`LibVLC(context, args)` 必须在主线程或后台单次初始化；多次构造会浪费 native 资源。`TvVlcLibrary` 用 double-checked locking 单例
- **`MediaPlayer.attachViews/detachViews` 与 Compose 重组**：`DisposableEffect` 的 dispose 时机要早于 `MediaPlayer.release()`，否则 native crash。生命周期严格 attach → detach → release
- **`Media` 对象的 ref count**：`mediaPlayer.media = media` 之后**必须**调 `media.release()`，否则 native 泄露；MediaPlayer 内部已持有 ref，本地 release 不影响播放
- **track id 不稳定**：每次 `setMedia` 后 audioTracks / spuTracks 列表的 id 可能变化；UI 必须订阅 `MediaPlayerESAdded` / `ESDeleted` 事件刷新菜单
- **字幕 SPU 内置 id 0 = "Disable"**：LibVLC 把"无字幕"映射到 id 0 / id -1。客户端"关闭字幕"选项映射到 `setSpuTrack(-1)`
- **mediacodec 黑屏机型**：早期 Android TV / 部分国产 ROM 的 mediacodec H.265 黑屏。PoC 中观察到则在 `Media.addOption` 加 `:codec=avcodec,all`（强制软解 H.265）；否则保持硬解
- **HLS 直播流 seek 行为**：项目长视频源以 mp4 为主，HLS 较少；LibVLC HLS 的 seek 必须 chunk 边界对齐，业务层 seek 后 LibVLC 自己处理，无需特殊代码
- **APK 体积**：删除 Media3 后约净减 3-5MB（debug build）；release 经 R8 后净变化更小
- **proguard 规则**：IPTV 已经在跑说明现有 `proguard-rules.pro` 覆盖 LibVLC native 类。本任务无需新增，但需要确认长视频额外用到的类（如 `MediaPlayer.TrackDescription`、`Media.Slave`）都被 `-keep` 覆盖

## 8. 回滚计划

如果上线后发现致命问题（设备兼容性、性能、字幕错位等），**回滚靠 git revert**：

```bash
git revert <merge-commit>  # 单 PR 整体 revert
```

不保留 Media3 fallback 代码路径（PRD N3 已锁）。

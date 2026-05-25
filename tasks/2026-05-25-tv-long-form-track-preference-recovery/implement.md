# Implement：TV 长视频字幕/音轨偏好恢复

- 日期：2026-05-25
- 关联 PRD：`./prd.md`

## 1. 总体方案

三条独立 surface（F1 VLC Playing gate / F2 type-only fallback / F3 audio 状态回灌）合并到同一个任务，因为它们组合起来才覆盖用户感受到的"完全不记"。**单做任一一条都不能让用户验收通过**：

- 只 F1：preference 真的被应用了，但 type-only preference 仍永远找不到 track；audio picker 仍撒谎
- 只 F2：edge case 救回来了，但主线（"英语 aac"）仍因 VLC 时序丢；audio picker 仍撒谎
- 只 F3：picker 不撒谎了，但本身代表的 selection 还是错的（VLC 没切到 preference 对应的轨道）

## 2. 实施顺序（强制）

```
1. F2 — 纯函数 type-only fallback（最小风险，先做）
2. F1 — VLC Playing gate（涉及播放器状态机，谨慎做、单测托底）
3. F3 — audio 状态回灌（建立在 F1 之上验证才有意义）
4. 单测全跑
5. 真机/模拟器手测 review.md §1 全场景
6. CONTEXT.md sync + 版本号 + 提交
```

## 3. 文件结构

### 3.1 新增

| 路径 | 用途 |
|---|---|
| `tv-app/src/test/java/com/chee/videos/core/player/TvLongFormTrackSelectionFallbackTest.kt` | `resolveLongFormTrackByLanguage` 的 type-only fallback 全分支单测 |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormSubtitlePreferenceFallbackTest.kt` | `resolveSelectedSubtitleTrackByPreference` 的 type-only fallback 单测 |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormAudioPreferenceFallbackTest.kt` | `resolveAudioSelectionOnTrackLoad` 各分支（含 type-only fallback）单测 |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormVlcPlayingGateSpecTest.kt` | 源文 audit：`LongFormVideoPlayer.kt` audio LaunchedEffect 起首必须出现 `isVlcPlaying` 判断；`TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 字幕 reload 路径同样 |

### 3.2 修改

| 路径 | 改动 |
|---|---|
| `tv-app/src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt` | `resolveLongFormTrackByLanguage` 在 `language.isBlank() && type.isNotBlank()` 分支按 type 匹配第一条同 type 的 track 返回；不能再直接 `return null` |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt` | `resolveSelectedSubtitleTrackByPreference` 同样在 type-only 时按 `subtitlePreferenceType(track) == typePreference` 匹配第一条 |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt` | `resolveAudioSelectionOnTrackLoad`（双参与三参版本）顺着 `resolveLongFormTrackByLanguage` 更新自动获得 type-only fallback 支持；新增防御：`storedPreference` 是 type-only 时也算"非 blank" |
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` | 新增 `var isVlcPlaying by remember { mutableStateOf(player.isPlaying) }`；EventListener 中 `Playing → true` / `Stopped/EncounteredError/EndReached → false`；audio LaunchedEffect 起首 `if (!isVlcPlaying || audioTracks.isEmpty()) return`；末尾增加：`if (resolvedSelection != null && resolvedSelection != selectedAudioTrackId) { onSelectAudioTrack(resolvedSelection, audioTracks.firstOrNull { it.id == resolvedSelection }?.let(::buildAudioTrackPreference)) }` |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt` | 字幕 reload `LaunchedEffect(playUrl, selectedSubtitleTrackId, hasStartedPlayback)` 加 gate：将 `applyLongFormMediaSource` + `mediaPlayer.play()` 的 subtitle slave 注入改为等到收到 `Playing` 事件之后再绑定 — 改造方式：把 subtitle 注入从 `applyLongFormMediaSource` 拆出来，初次 `applyLongFormMediaSource` 不带 subtitleUrl 仅 setMedia → play → 收到 Playing 事件后通过 `mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, url, true)` 注入。同时调用 `onSelectAudioTrack` 时禁止额外 DataStore save（state 回灌不算用户手动选择） |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt` | 同上路径 |
| `android-tv-app/tv-app/build.gradle.kts` | `versionCode = 70`、`versionName = "0.1.70"` |

### 3.3 父级"onSelectAudioTrack 是手动选 vs 状态回灌"的区分

PRD §5 F3 要求父级在收到 audio LaunchedEffect 的回灌时**不**再写 DataStore（避免 save → state change → LaunchedEffect 再 fire → 再回灌的循环）。两种做法择一：

- **方案 a**：把 `onSelectAudioTrack` 拆成两个 callback：`onUserSelectAudioTrack(trackId, preference)`（写 DataStore）与 `onResolveAudioTrack(trackId, preference)`（只更新 in-memory state）。LongFormVideoPlayer 内部按调用源头区分使用。
- **方案 b**：保留单 callback，第三个 boolean 参数 `isUserAction: Boolean`：

```kotlin
onSelectAudioTrack: (trackId: String?, preference: TvTrackPreference?, isUserAction: Boolean) -> Unit
```

picker 调用时传 `isUserAction = true`；audio LaunchedEffect 回灌时传 `false`。父级仅 `isUserAction = true` 才写 DataStore。

**采用方案 b**：只改一个签名比拆两个 API 简单。子幕侧不需要类似回灌（因为 `selectedSubtitleTrackId` 直接由父级的 LaunchedEffect 写出 resolved 值，路径已经正确）。

## 4. 关键代码片段

### 4.1 type-only fallback —— `TvLongFormTrackSelection.kt`

```kotlin
internal fun resolveLongFormTrackByLanguage(
    tracks: List<TvLongFormVlcTrack>,
    preference: TvLongFormLanguagePreference,
): TvLongFormVlcTrack? {
    val language = normalizeLongFormLanguageCode(preference.language)
    val type = preference.type?.trim()?.lowercase(Locale.ROOT).orEmpty()

    // F2：language 空但 type 不空时按 type 直接匹配
    if (language.isBlank()) {
        if (type.isBlank()) return null
        return tracks.firstOrNull { it.type?.trim()?.lowercase(Locale.ROOT) == type }
            ?: tracks.firstOrNull { it.name.lowercase(Locale.ROOT).contains(type) }
    }

    // 既有 language 匹配路径
    val languageMatches = tracks.filter { track ->
        val trackLanguage = normalizeLongFormLanguageCode(track.language)
        trackLanguage == language ||
            trackLanguage.startsWith("$language-") ||
            language.startsWith("$trackLanguage-") ||
            track.name.lowercase(Locale.ROOT).contains(language)
    }
    if (languageMatches.isEmpty()) return null
    if (type.isNotBlank()) {
        languageMatches.firstOrNull { it.type?.trim()?.lowercase(Locale.ROOT) == type }?.let { return it }
        languageMatches.firstOrNull { it.name.lowercase(Locale.ROOT).contains(type) }?.let { return it }
    }
    return languageMatches.first()
}
```

### 4.2 type-only fallback —— `LongFormSubtitleSupport.kt`

```kotlin
fun resolveSelectedSubtitleTrackByPreference(
    tracks: List<SubtitleTrackDto>,
    preference: TvTrackPreference?,
): SubtitleTrackDto? {
    val safePreference = preference ?: return null
    val normalizedPreference = normalizeLongFormLanguageCode(safePreference.language)
    val typePreference = safePreference.type.trim().lowercase()

    // F2：language 空时按 type 直接匹配
    if (normalizedPreference.isBlank()) {
        if (typePreference.isBlank()) return null
        return tracks.firstOrNull { it.available && subtitlePreferenceType(it) == typePreference }
    }

    // 既有 language 匹配路径
    val languageMatches = tracks.filter { track ->
        val language = normalizeLongFormLanguageCode(track.languageCode)
        track.available && (
            language == normalizedPreference ||
                language.startsWith("$normalizedPreference-") ||
                normalizedPreference.startsWith("$language-")
        )
    }
    if (languageMatches.isEmpty()) return null
    if (typePreference.isNotBlank()) {
        languageMatches.firstOrNull { subtitlePreferenceType(it) == typePreference }?.let { return it }
    }
    return languageMatches.first()
}
```

### 4.3 VLC Playing gate + audio 回灌 —— `LongFormVideoPlayer.kt`

```kotlin
var isVlcPlaying by remember { mutableStateOf(player.isPlaying) }

DisposableEffect(player) {
    val listener = object : MediaPlayer.EventListener {
        override fun onEvent(event: MediaPlayer.Event) {
            latestOnVlcEvent(event)
            when (event.type) {
                MediaPlayer.Event.Playing -> {
                    isPlaying = true
                    isVlcPlaying = true
                }
                MediaPlayer.Event.Paused -> {
                    isPlaying = false
                    // 仍处于 Playing 状态机里，不复位 isVlcPlaying
                }
                MediaPlayer.Event.Stopped,
                MediaPlayer.Event.EndReached,
                MediaPlayer.Event.EncounteredError -> {
                    isPlaying = false
                    isVlcPlaying = false
                }
                // ... 既有其它事件
            }
            audioTracks = buildLongFormAudioTracksFromVlc(player.audioTracks, player.audioTrack)
        }
    }
    // ...
}

LaunchedEffect(player, audioTracks, selectedAudioTrackId, selectedAudioPreference, isVlcPlaying) {
    if (!isVlcPlaying || audioTracks.isEmpty()) return@LaunchedEffect

    val resolvedSelection = resolveAudioSelectionOnTrackLoad(
        currentSelection = selectedAudioTrackId,
        storedPreference = selectedAudioPreference,
        tracks = audioTracks,
    )
    if (selectedAudioTrackId?.isNotBlank() == true && resolvedSelection == null) {
        onSelectAudioTrack(null, null, /* isUserAction = */ false)
    }
    val selected = audioTracks.firstOrNull { it.id == resolvedSelection }
    player.audioTrack = selected?.vlcTrackId ?: -1

    // F3：把 resolvedSelection 回灌给父级 state，避免 picker 撒谎
    if (resolvedSelection != null && resolvedSelection != selectedAudioTrackId) {
        val preferenceForTrack = buildAudioTrackPreference(selected)
        onSelectAudioTrack(resolvedSelection, preferenceForTrack, /* isUserAction = */ false)
    }
}
```

### 4.4 subtitle slave 注入解耦 —— `TvLongFormPlayerScreen.kt`

```kotlin
// applyLongFormMediaSource 拆改为仅 setMedia，不带 subtitle
applyLongFormMediaSource(
    libVLC = libVLC,
    mediaPlayer = mediaPlayer,
    sourceUrl = playUrl,
    baseUrl = uiState.baseUrl,
    selectedSubtitleTrack = null,  // 此处不带 subtitle，等 Playing 之后再 addSlave
)
mediaPlayer.play()

// 新增：等 Playing 之后注入 subtitle slave
var pendingSubtitleSlave by remember(detail.id, uiState.accessToken) {
    mutableStateOf<String?>(null)
}
LaunchedEffect(selectedSubtitleTrackId, detail.subtitleTracks, uiState.baseUrl) {
    val track = resolveSelectedSubtitleTrack(detail.subtitleTracks, selectedSubtitleTrackId)
    pendingSubtitleSlave = track
        ?.takeIf { it.available && it.url.isNotBlank() }
        ?.let { resolvePlaybackAssetUrl(uiState.baseUrl, it.url) }
}
LaunchedEffect(pendingSubtitleSlave, isVlcPlaying) {
    val url = pendingSubtitleSlave ?: return@LaunchedEffect
    if (!isVlcPlaying) return@LaunchedEffect
    mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, url, true)
}
```

注意 `isVlcPlaying` 这个 state 在 `TvLongFormPlayerScreen` 里需要从 `LongFormVideoPlayer.onVlcEvent` 反向通信，或直接在 screen 自己挂 EventListener。实施时取最简单且无信号丢失的方式（建议在 screen 里维护一个独立的 `isVlcPlaying` 跟着 `MediaPlayer.Event.Playing/Stopped/EncounteredError` 走，与 LongFormVideoPlayer 内部那份不互相依赖）。

`resolvePlaybackAssetUrl` 已在 `LongFormSubtitleSupport.kt` 是 `private`；本任务需要把它改成 `internal` 让 screen 也能调，或者直接复制等价路径解析逻辑过来。

### 4.5 `applyLongFormMediaSource` 签名收缩

既然 subtitle slave 不再在 setMedia 时绑定，`applyLongFormMediaSource` 的 `selectedSubtitleTrack` 参数实质废弃，删除或保留为 `null` 默认值。建议**删除**，逼调用方走新的 slave 注入路径（避免半旧半新）。

### 4.6 `resolveLongFormPlayerUpdate` 调整

`resolveLongFormPlayerUpdate` 之前在 `preparedSubtitleTrackId != nextSubtitleTrackId` 时返回 `shouldReplaceSource=true, preservePosition=true`，目的是为了通过 setMedia 重新带 subtitle 入媒体。现在 subtitle 走 `addSlave`，不需要重新 setMedia。改为：

```kotlin
if (preparedSubtitleTrackId != nextSubtitleTrackId) {
    return LongFormPlayerUpdateDecision(
        shouldClear = false,
        shouldReplaceSource = false,  // 改为 false：不再 reload 媒体
        preservePosition = false,
    )
}
```

字幕变化由独立的 `LaunchedEffect(pendingSubtitleSlave, isVlcPlaying)` 处理。`preparedSubtitleTrackId` 这个 state 改名为 `appliedSubtitleSlaveId` 或干脆删除（看后续是否需要 idempotency 保护 addSlave 重复调用 — LibVLC `addSlave` 是允许重复 + 带 select=true 切换的）。

## 5. 单测

### 5.1 `TvLongFormTrackSelectionFallbackTest.kt`

- L1：preference = `("en", "")`，tracks 中有 `language="en"` 一条 → 返回该 track
- L2：preference = `("", "default")`，tracks 中有 `type="default"` 一条但 language 空 → **返回该 track**（type-only fallback）
- L3：preference = `("", "forced")`，tracks 中无任何 `type="forced"` 但有 name 包含 "forced" 一条 → 返回该 track
- L4：preference = `("", "")`，blank preference → 返回 null
- L5：preference = `("zh", "default")`，tracks 中既有 `language="zh"` 也有 `type="default"` → 优先 language 匹配后用 type 二次筛
- L6：preference = `("en", "")`，tracks 中无 "en"、但有 name 含 "english" → 兜底 name contains 匹配
- L7：preference = `("zh-cn", "")`，tracks 中有 `language="zh"` → 前缀匹配返回

### 5.2 `LongFormSubtitlePreferenceFallbackTest.kt`

- S1~S7：与 §5.1 对称（基于 `SubtitleTrackDto` 与 `subtitlePreferenceType`）

### 5.3 `LongFormAudioPreferenceFallbackTest.kt`

- A1：`currentSelection` 与某 track.id 完全匹配 → 直接返回 currentSelection（既有快路径）
- A2：`currentSelection` 不匹配任何 track + storedPreference 命中 language → 返回 preference 解析结果
- A3：`currentSelection` 不匹配 + preference type-only 命中 → 返回 fallback 结果
- A4：双空 → 返回 null

### 5.4 `LongFormVlcPlayingGateSpecTest.kt`

源文 audit：

- G1：`LongFormVideoPlayer.kt` 必须包含 `isVlcPlaying` 标识符出现 ≥ 3 处（state 声明 + Event 处理 + LaunchedEffect gate）
- G2：`LongFormVideoPlayer.kt` 的 audio LaunchedEffect 块（grep `LaunchedEffect(player, audioTracks`）内必须包含 `isVlcPlaying` 关键字
- G3：`TvLongFormPlayerScreen.kt` 必须包含 `addSlave(IMedia.Slave.Type.Subtitle` 字样且**不**通过 `applyLongFormMediaSource` 传 subtitle
- G4：`TvSeriesPlayerScreen.kt` 同 G3
- G5：`LongFormVideoPlayer.kt` 的 `onSelectAudioTrack` 调用必须包含 `isUserAction = false` 与 `isUserAction = true` 两种调用（说明回灌路径与用户路径分开）

## 6. 验证脚本

### 6.1 静态

```bash
cd android-tv-app
./gradlew --no-daemon :tv-app:compileDebugKotlin
./gradlew --no-daemon :tv-app:testDebugUnitTest
./gradlew --no-daemon :tv-app:assembleDebug
./gradlew --no-daemon :tv-app:assembleDebugAndroidTest
```

repo 根：

```bash
rg -n 'addSlave\(IMedia\.Slave\.Type\.Subtitle' android-tv-app/tv-app/src/main
rg -n 'selectedSubtitleTrack =' android-tv-app/tv-app/src/main  # 期望 applyLongFormMediaSource 不再带此参数
rg -n 'versionCode = 70' android-tv-app/tv-app/build.gradle.kts
rg -n 'isUserAction' android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt
git diff --check
rg -n $'�' android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt
```

### 6.2 手测

见 `review.md §1`。**未跑通真机/模拟器手测之前不允许 mark DONE.md**。

## 7. CONTEXT.md sync

实施完成时插入 PRD §6 的三条术语。`[[LibVLC track id 不稳定]]` 在 LibVLC 迁移任务里已有，引用即可。

## 8. 提交

```
修复TV长视频字幕音轨偏好不记忆

- VLC Playing gate：setAudioTrack / subtitle addSlave 等到收到 MediaPlayer.Event.Playing 之后再发
- type-only preference fallback：language 空但 type 不空时按 type 匹配第一条同 type 的 track
- audio LaunchedEffect 状态回灌：resolved track id 通过 onSelectAudioTrack 回到父级，picker 不再撒谎
- subtitle 注入路径改造：从 setMedia 时绑定改为 addSlave 解耦，避免 reload media 抖动
- 版本号 0.1.70 / 70
```

## 9. 与 focus-guarding 任务的协调

`tasks/2026-05-25-tv-long-form-focus-guarding/` 任务同样修改 `LongFormVideoPlayer.kt` / `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt`。两个任务可以并行推进，但合并时必须按下列规则：

- focus-guarding 改的是 root Box 的 modifier 与 LaunchedEffect、`resumePromptSlot` 参数
- track-preference-recovery 改的是 EventListener 状态机、audio LaunchedEffect 内容、subtitle slave 注入路径
- 没有同一行/同一函数的修改冲突；任一先合都不会阻塞另一个

合并顺序建议：focus-guarding 先（影响面纯 UI，更容易快速验证），track-preference-recovery 后（涉及 LibVLC 状态机调试时间更长）。

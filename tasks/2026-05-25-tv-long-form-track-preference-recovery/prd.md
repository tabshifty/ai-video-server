# PRD：TV 长视频字幕/音轨偏好恢复（修复"下次播放显示默认"）

- 日期：2026-05-25
- 目标端：`android-tv-app/`（独立工程，仅 TV 端）
- 范围：长视频字幕/音轨持久化与重进时恢复链路

## 1. 问题陈述

LibVLC 迁移之后（参见 `tasks/2026-05-25-tv-long-form-libvlc-migration/`），长视频字幕与音轨的 **DataStore 持久化设计已经按 language code 改写**（`tv_subtitle_language_preferences` / `tv_audio_language_preferences`），但用户报告**重进同一视频时偏好不生效**：

- 字幕 picker 高亮的是"关闭字幕"（首项），屏幕上**实际无字幕**
- 音轨 picker 高亮的是"自动选择"（首项），**实际播放也不是上次选的语种**
- 同会话内（不杀 App）退回详情页 → 再进播放器，仍然丢
- 同会话内**不退出**播放器、直接重开 picker：picker 显示正确（in-memory 状态没问题）
- 选完轨道**等 30 秒再退出**仍然丢失（排除 save coroutine 被取消）

通过 grill-with-docs 排查，三条独立故障面共同导致用户感受到的"完全不记":

- **F1 / VLC 时序**：`player.audioTrack = vlcTrackId` 与 `applyLongFormMediaSource` 中的 `addSlave(SUBTITLE, ...)` 在 `MediaPlayer.Event.Playing` 之前被调用时 LibVLC 会丢弃。重进时 preference 异步 read 与 audioTracks 异步上报到达的时刻常常在 `Playing` 之前。
- **F2 / preference 单向死路**：`buildSubtitleTrackPreference` / `buildAudioTrackPreference` 允许写入"只有 type、无 language"的 preference（典型场景：track `isDefault=true` 但 server 没给 `languageCode`），但 `resolveSelectedSubtitleTrackByPreference` / `resolveLongFormTrackByLanguage` 在 `if (normalizedPreference.isBlank()) return null` 直接放弃。**preference 能存进 DataStore，但永远 resolve 不回任何 track。**
- **F3 / picker 状态没回灌**：`LongFormVideoPlayer` 的音轨 LaunchedEffect 把 preference 解析出 `resolvedSelection` 并写到 `player.audioTrack`，**但没有把 `resolvedSelection` 回调给父级的 `selectedAudioTrackId`**。结果即使 VLC 实际切了轨，picker 显示的仍是"自动选择"。F1 修好后，没有 F3 的话 picker 会撒谎；F3 修好后，没有 F1 的话 VLC 实际没切。两个必须一起。

## 2. 用户故事

作为 TV 用户，我希望：

1. **R1（字幕）**：上次播放选了"中文字幕" → 这次重进同一视频 → 屏幕上自动出现"中文字幕"且 picker 显示选中状态。
2. **R2（音轨）**：上次播放选了"英语 aac" → 这次重进 → 实际听到的就是英语轨，且 picker 显示选中"英语 aac"。
3. **R3（isDefault 型字幕）**：服务端给的字幕没有 languageCode 只有 isDefault=true，我选了它 → 下次重进依然能恢复（即使是 type-only preference）。
4. **R4（同会话内退回→再进）**：不杀 App，从详情页进入 → 选轨 → 退出 → 同一视频再进，记忆生效。
5. **R5（跨 App 重启）**：杀 App 重开 → 进入同一视频，记忆生效。
6. **R6（语言代号映射）**：上次保存的是 "en"，VLC 报的 track name 含 "english" / "英语" / "英文" / "eng" 都能匹配上。
7. **R7（不破坏既有逻辑）**：自动连播下一集时，新一集仍按用户的语种偏好挑轨道。

## 3. 作用域

### 3.1 覆盖

- `tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt`：audio LaunchedEffect 加 VLC `Playing` gate + 回灌 `selectedAudioTrackId`
- `tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`：`applyLongFormMediaSource` 改为 "set media → wait Playing → addSlave" 或采用 LibVLC `Media.addSlave(...)`（Media 级注入，不依赖 player 状态）
- `tv-app/src/main/java/com/chee/videos/core/ui/LongFormAudioTrackSupport.kt`：`resolveAudioSelectionOnTrackLoad` 加 type-only fallback
- `tv-app/src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt`：`resolveSelectedSubtitleTrackByPreference` 加 type-only fallback
- `tv-app/src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt`：`resolveLongFormTrackByLanguage` 加 type-only fallback
- `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`：字幕 LaunchedEffect 同样加 VLC `Playing` gate（媒体重 load 时机调整）
- `tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`：同上路径处理
- 单测：纯函数路径全覆盖（preference resolve + audio selection on track load + media slave 接入时机）
- 版本号：`android-tv-app/tv-app/build.gradle.kts` `versionCode = 70`、`versionName = "0.1.70"`

### 3.2 不覆盖

- 后端字幕/音轨数据生产（`internal/services/subtitle.go` 等）— 由 LibVLC 迁移任务已经定型
- 用户能不能手动**编辑** preference（设置页提供"始终默认英语"开关）— 独立功能
- IPTV / 短视频 / 手机端
- 持久化的 schema/key 迁移：本任务沿用现有 `tv_subtitle_language_preferences` / `tv_audio_language_preferences`

## 4. 非目标

- **N1**：不再回归用 vlcTrackId 直接持久化的方案 — LibVLC 迁移已确认 [[LibVLC track id 不稳定]]
- **N2**：不引入"播放时不自动应用 preference，只在用户点击 picker 时记忆"模式 — 与既有行为契约相悖
- **N3**：不做 type-only preference 的 UI 表达（picker 里"isDefault 标记"图标）— 仅修复 resolve 路径
- **N4**：不调整 5/10/15/20/30 秒快进步长偏好的持久化（在另一个 key 上，无关）
- **N5**：不为 audio 增加"无 language 信息时按 track index 持久化"的兜底 — 跨 media 重新加载时 index 不稳定，先按 type-only fallback 兜底已经覆盖绝大多数实际场景；如果未来出现 type 也无的情况再单独立任务

## 5. 三条故障面的技术对应

| F | 故障 | 修复方式 |
|---|------|---------|
| **F1** | LibVLC 在 `Playing` 之前 `setAudioTrack` / 子幕注入被丢弃 | 在 `LongFormVideoPlayer` 加 `var isVlcPlaying by remember { mutableStateOf(false) }`，EventListener 收到 `Playing` 时置 true、`Stopped` / `EncounteredError` 时置 false；audio LaunchedEffect 顶部 `if (!isVlcPlaying || audioTracks.isEmpty()) return`。同时 `TvLongFormPlayerScreen` 的 subtitle 媒体 reload LaunchedEffect 也加同样 gate（或改用 `Media.addSlave` 而非 `MediaPlayer.addSlave`，避免依赖 player 状态） |
| **F2** | `("","default")` 形 preference 写得进去但 resolve 端在 `if (language.isBlank()) return null` 死掉 | `resolveSelectedSubtitleTrackByPreference` 与 `resolveLongFormTrackByLanguage` 在 language 空但 type 不空时改为按 type 匹配：扫描所有 track，找第一条 `subtitlePreferenceType(it) == type` 或（audio）`inferLongFormTrackPreferenceType(name) == type` 的 track 返回 |
| **F3** | audio LaunchedEffect 写 `player.audioTrack` 但不告诉父级 | 在 audio LaunchedEffect 末尾，`if (resolvedSelection != null && resolvedSelection != selectedAudioTrackId)` 时调用 `onSelectAudioTrack(resolvedSelection, preferenceForTrack)`；调用方在父级把它存到 in-memory `selectedAudioTrackId` 且**不**重复触发 DataStore save（要避免 save-loop） |

## 6. 术语对应（实施完成时 sync 到 CONTEXT.md）

| 术语 | 一句话定义 |
|------|----------|
| `VLC Playing gate` | TV 长视频 LibVLC 内核为 `setAudioTrack` 与字幕 slave 注入加的等待门：所有 track 切换性调用必须在收到 `MediaPlayer.Event.Playing` 之后执行，否则 LibVLC 会丢弃。由 `var isVlcPlaying` 状态门把守，Event.Stopped/Error 时复位。是 LibVLC 集成的已知约束，不是某个 bug 的临时绕路。 |
| `Type-only preference fallback` | TV 长视频字幕/音轨偏好恢复的 fallback 规则：当 DataStore 存的 `TvTrackPreference` 只有 `type`（如 "default" / "forced" / "commentary"）而 `language` 为空时，resolve 端按 type 字段直接在当前 track 列表里找第一条同 type 的 track 返回，而不是直接放弃。修复 LibVLC 迁移之后 isDefault-but-no-language 类字幕/音轨被永久丢失的 design hole。 |
| `Audio LaunchedEffect 状态回灌` | TV 长视频音轨偏好恢复要求：`LongFormVideoPlayer` 内的 audio LaunchedEffect 不仅要把 resolved track 推给 LibVLC（`player.audioTrack = vlcTrackId`），还必须把 `resolvedSelection` 通过 `onSelectAudioTrack(...)` 回灌给父级的 `selectedAudioTrackId` state，保证 audio picker UI 不撒谎。回灌时不重复触发 DataStore save。 |

## 7. 验收 (acceptance criteria)

详见 `review.md §1`。摘要：

1. **A1**：上次选了"英语 aac" → 退出 → 重进 → 听到的是英语 + picker 高亮"英语 aac"
2. **A2**：上次选了"中文字幕" → 退出 → 重进 → 屏幕显示中文字幕 + picker 高亮"中文字幕"
3. **A3**：上次选了一条 isDefault=true 但无 languageCode 的字幕（typically 内嵌强制中文）→ 退出 → 重进 → 字幕恢复
4. **A4**：同会话内退回详情页再进，A1/A2/A3 同等通过
5. **A5**：杀 App 后重开，A1/A2/A3 同等通过
6. **A6**：电视剧自动连播下一集，新一集开始时 audio/subtitle 仍按用户语种偏好
7. **A7**：无回归：手动切换音轨/字幕、关闭字幕、5 秒快进步长设置、字幕/音轨 picker 显示状态等

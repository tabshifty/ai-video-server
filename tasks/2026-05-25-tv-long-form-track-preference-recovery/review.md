# Review：TV 长视频字幕/音轨偏好恢复

- 日期：2026-05-25
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`

## 0. 准入条件

未满足任一项 → **回到 implement 阶段**：

- [ ] TV 单测 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含新增 `TvLongFormTrackSelectionFallbackTest` / `LongFormSubtitlePreferenceFallbackTest` / `LongFormAudioPreferenceFallbackTest` / `LongFormVlcPlayingGateSpecTest`）
- [ ] TV 构建 `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] androidTest 编译通过：`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebugAndroidTest`
- [ ] `git diff --check` 通过、乱码扫描无输出
- [ ] `rg -n 'addSlave\(IMedia\.Slave\.Type\.Subtitle' android-tv-app/tv-app/src/main` 至少出现在 `TvLongFormPlayerScreen.kt` 与 `TvSeriesPlayerScreen.kt` 两处
- [ ] `rg -n 'isUserAction' android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` 同时出现 `true` 与 `false` 两种调用
- [ ] `android-tv-app/tv-app/build.gradle.kts` `versionCode = 70`、`versionName = "0.1.70"`
- [ ] `CONTEXT.md` 已 sync 三条新术语
- [ ] `plan.md` 追加进度条目
- [ ] **真机或模拟器回归**：§1 所有 A1~A7 跑通，结果写入 `DONE.md`

## 1. 手测脚本

每条场景按下列模板执行；任一未通过 → fix → 回到 §0 重跑准入。

### 1.1 A1 — 音轨恢复（主线）

**前置素材**：一个**有多音轨**（至少含英语 + 默认非英语）的长视频，已观看历史 < MinResumeMs 或为 0（避免续播卡干扰）。

1. 进入该视频播放器
2. 等控制条 5 秒 auto‑hide 之后，按 DPAD_DOWN 唤起 UI，按 ←/→ 移到"音轨"按钮，按 OK 打开音轨 picker
3. 选中"英语 aac"（或其它包含 "english" / "英语" 的音轨），按 OK 选中
4. 验证当下：屏幕上**听到的音频**变成了英语
5. 等播放至少 5 秒（让 DataStore save 完成）
6. 按 BACK → 返回二次确认 → 再按 BACK 退出播放器
7. 立即从详情页**再次进入**播放器
8. 进入后**前 10 秒**专注听：

**预期**：
- [ ] 第 7 步之后，播放开始的 1~3 秒内，音频自动切到英语（VLC `Playing` 之后 audio LaunchedEffect 把 setAudioTrack 应用上）
- [ ] 在第 8 步等 10 秒内**任意时刻**按 DPAD_DOWN → 音轨 picker → 检查"英语 aac"被 ✓ 标记选中

### 1.2 A2 — 字幕恢复（主线）

**前置素材**：一个**有多字幕轨**且至少有一条**中文字幕**的长视频。

1. 进入播放器
2. 等 5 秒，按 DPAD_DOWN → ←/→ 移到"字幕"按钮 → OK 打开字幕 picker
3. 选中"中文字幕"，按 OK
4. 验证当下：屏幕上**出现中文字幕**
5. 等播放至少 5 秒
6. 退出 → 再进同一视频
7. 进入后前 10 秒看屏幕：

**预期**：
- [ ] 第 6 步进入后 1~5 秒内，屏幕自动出现中文字幕（VLC `Playing` 之后 `addSlave(SUBTITLE, ..., true)` 应用上）
- [ ] 第 7 步内按 DPAD_DOWN → 字幕 picker → "中文字幕"被 ✓ 标记选中（picker 不再撒谎）

### 1.3 A3 — type-only preference（isDefault 兜底）

**前置素材**：一个长视频，其某条字幕在服务端 metadata 里 `isDefault=true` 但 `languageCode=""`（典型场景：上传时未指定语种但勾选了"默认字幕"）。

- 如果手头没有这种素材，可在 admin-web 上传一个 .ass 字幕但不填语种、勾上"默认"
- 服务端确认 `detail.subtitleTracks[i].isDefault=true && languageCode==""`

1. 进入播放器
2. 等 5 秒，按 DPAD_DOWN → ←/→ → 字幕 picker
3. 选中那条 isDefault 字幕（label 应该有"（默认）"或类似标记）
4. 验证当下：屏幕显示该字幕
5. 等 5 秒
6. 退出 → 再进

**预期**：
- [ ] 进入后字幕自动恢复，picker 显示该条选中（type-only fallback 触发）
- [ ] DataStore inspect（如能拿到）应看到 `tv_subtitle_language_preferences` 里该 videoId 的值是 `{"language":"","type":"default"}`，不是 null（save 没有把它擦掉）

### 1.4 A4 — 同会话内退回→再进

把 A1/A2/A3 在同一次 App 启动内做完，不杀 App。每次的"退出再进"都走详情页 → 播放器。

**预期**：
- [ ] A1/A2/A3 在每一次"退出再进"循环里都通过
- [ ] picker 显示状态在每一次进入后 5~10 秒内最终稳定到正确选中

### 1.5 A5 — 跨 App 重启

1. 完成 A1（选好"英语 aac"，等 30 秒）
2. 完全杀 App（包括从最近任务列表 swipe 掉）
3. 重新打开 App，进入同一视频

**预期**：
- [ ] A1 的预期都成立（DataStore 持久化跨进程仍有效）

### 1.6 A6 — 电视剧自动连播

**前置素材**：一部至少有两集且至少有英语 + 中文字幕轨的剧。

1. 进入第 1 集
2. 选好"英语 aac" + "中文字幕"，等 5 秒
3. 拖动播放进度到接近片尾（让自动连播触发）
4. 第 2 集自动开始播放

**预期**：
- [ ] 第 2 集开始播放后 5 秒内，音轨是英语、字幕是中文
- [ ] picker 在第 2 集里显示这两项被选中

如果第 2 集的字幕轨里没有"中文字幕"那条（语种缺失），降级表现接受为：picker 显示"关闭字幕"且无字幕；不算回归（preference resolve 找不到 track 是合理的）。

### 1.7 A7 — 无回归

依次操作以下既有交互，确保未被本任务改坏：

- [ ] 手动从 picker 切换音轨 / 字幕，立即生效，picker 显示更新；30 秒后退出再进入仍能恢复
- [ ] 在 picker 里选"关闭字幕"/"自动选择" → 退出再进 → 字幕真的关 / 音轨真的自动
- [ ] 5/10/15/20/30 秒快进步长设置仍能从 TV 设置页改、播放器内 ←/→ 实际步长按此变化
- [ ] 自动连播提示卡 / 续播提示卡的显示与消失行为不变
- [ ] **重要**：手动切换音轨后**不**应该触发"回灌"循环，导致 picker 抖动或 DataStore 反复写。手动切之后 logcat（如有）里 audio LaunchedEffect 应只 fire 一次新写入，**不**出现连续两次 `onSelectAudioTrack(...)` 调用
- [ ] 切换字幕轨之后，**不**应该发生媒体被重新 setMedia + 跳回开头（subtitle 走 `addSlave` 不应导致 reload）

## 2. 自动化产物

提交里必须包含：

- [ ] `TvLongFormTrackSelectionFallbackTest.kt` ≥ 7 个 case（implement §5.1 L1~L7）
- [ ] `LongFormSubtitlePreferenceFallbackTest.kt` ≥ 7 个 case
- [ ] `LongFormAudioPreferenceFallbackTest.kt` ≥ 4 个 case
- [ ] `LongFormVlcPlayingGateSpecTest.kt`（源文 audit，G1~G5）
- [ ] `TvLongFormTrackSelection.kt` / `LongFormSubtitleSupport.kt` / `LongFormAudioTrackSupport.kt` 代码变更
- [ ] `LongFormVideoPlayer.kt` / `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 代码变更
- [ ] `applyLongFormMediaSource` 签名变更（或删除 `selectedSubtitleTrack` 参数）+ 所有调用方更新
- [ ] `android-tv-app/tv-app/build.gradle.kts` 版本号 +1
- [ ] `CONTEXT.md` 新增三条术语
- [ ] `plan.md` 追加条目

## 3. 文档与提交

- [ ] 提交信息为 implement §8 给出的中文格式
- [ ] 不修改 `.codex/skills/*`
- [ ] 用户在真机/模拟器跑完 §1 所有 A1~A7 并确认 → 创建 `DONE.md` 记录验证范围、自动化命令输出摘要、人工验证场景的真实结果

## 4. 可能的回归风险与缓解

| 风险 | 缓解 |
|---|---|
| `applyLongFormMediaSource` 删除 subtitle 参数后，某些调用方仍 setMedia 时带字幕 → 行为不一致 | implement §5 G3/G4 用源文 audit 强制检查；review §2 自动化产物校验 |
| `isVlcPlaying` gate 让初次播放的 audio 切换延迟到用户能感知（500~1500ms）| 接受：相比"完全不切换"是绝对改善；如果用户反馈延迟太长，独立任务做"在 Opening 事件就 setAudioTrack 然后 Playing 时再校验"的优化 |
| 回灌循环导致 audio picker 抖动或 DataStore 反复写 | implement §3.3 方案 b 的 `isUserAction: Boolean` 参数 + review §1.7 单独检查 |
| `addSlave` 在快速切换字幕轨时 LibVLC 不接受新 slave | `addSlave` 第三个参数 `select=true` 强制选中；如果实测仍有问题，回退为 setMedia 重 load 但保留 `addSlave` 时机 + Playing gate（保留 preservePosition 行为） |
| Subtitle reload `LaunchedEffect` 改造之后 [[续播提示卡]] 流程出问题（续播位置 + 初次媒体加载 + 字幕加载三者顺序） | A1/A2/A3 用"有历史进度的视频"重测一遍：观察播放从历史进度开始、音轨/字幕也按 preference 恢复，三者无干扰 |
| `LibVLC track id 不稳定` 在 type-only fallback 路径下放大（同 type 多条 track 之间顺序不稳定） | 接受首条匹配；后续可加"prefer track index 上次保存"的二级偏好，但本任务不做（N4 已声明） |
| Compose `LaunchedEffect` 受 `audioTracks` 列表新引用每次 fire 影响，gate 加上后还是会 retry 多次 | 接受 retry 多次但每次都是同一次 `setAudioTrack(同一 vlcTrackId)` 调用，LibVLC 应当幂等。如果 logcat 显示连续切换造成卡顿，下一步用 derivedStateOf 缩减依赖 |

## 5. 与 focus-guarding 任务的合并约束

如果同时合并 `tasks/2026-05-25-tv-long-form-focus-guarding/` 与本任务：

- [ ] `LongFormVideoPlayer.kt` 的 modifier chain（focus-guarding 改）与 EventListener / audio LaunchedEffect（本任务改）合并后逐处 diff 确认无重复 import、无 state 覆盖
- [ ] `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 的 `resumePromptSlot` 接入（focus-guarding）与 subtitle slave 注入路径（本任务）合并后两处都生效
- [ ] 两任务的版本号 +1 累加：本任务先合则 `versionCode = 70`、`versionName = "0.1.70"`；focus-guarding 再合时再 +1 到 71
- [ ] 两任务的 CONTEXT.md 术语都加入（互不冲突）
- [ ] 真机回归时把两个任务的 review §1 全部场景串起来跑一遍，确认无叠加 bug

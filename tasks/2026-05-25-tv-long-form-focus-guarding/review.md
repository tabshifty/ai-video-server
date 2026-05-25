# Review：TV 长视频播放器焦点兜底

- 日期：2026-05-25
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`

## 0. 准入条件

未满足任一项 → **回到 implement 阶段**：

- [ ] TV 单测 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 全绿（含新增 `LongFormPlayerFocusGuardTest` / `LongFormVideoPlayerSpecTest`）
- [ ] TV 构建 `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` BUILD SUCCESSFUL
- [ ] androidTest 编译通过：`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebugAndroidTest`
- [ ] `git diff --check` 通过、乱码扫描无输出
- [ ] `rg -n 'TvResumePromptCard\(' android-tv-app/tv-app/src/main` 仅出现在 `TvResumePromptCard.kt` 自身和三处 `resumePromptSlot = {` 块内
- [ ] `android-tv-app/tv-app/build.gradle.kts` `versionCode = 69`、`versionName = "0.1.69"`
- [ ] `CONTEXT.md` 已 sync 三条新术语（`TV 长视频焦点真空` / `LongFormVideoPlayer focus 兜底` / `续播提示卡内嵌位置`）
- [ ] `plan.md` 追加进度条目
- [ ] **真机或模拟器回归**：§1 所有 R1~R10 跑通，结果写入 `DONE.md`

## 1. 手测脚本

每条场景按下列模板执行；任一未通过 → fix → 回到 §0 重跑准入。

### 1.1 R1 — 续播卡 dispose 后 DPAD_DOWN 必须能唤起 UI

**前置**：选一个**有播放历史**的长视频（电影或 `18+`，watchSeconds > MinResumeMs）。

1. 从 TV 首页进入该视频详情页 → 按 OK 进入播放器
2. 此时屏幕底部应出现"上次播放至 …  继续观看 (5) / 从头播放"卡片
3. **不按任何遥控器键**，等待续播卡倒计时跑完（约 5 秒）→ 卡片淡出
4. 立即按 **DPAD_DOWN**

**预期**：
- [ ] 控制条出现（5 秒可见 + 自动隐藏）
- [ ] "播放/暂停"按钮被蓝色 glow 高亮
- [ ] 按 ←/→ 在按钮间移动焦点；按 OK 触发按钮 onClick

### 1.2 R2 — 5 秒 auto‑hide 之后 DPAD_DOWN 必须能唤起 UI

**前置**：选一个**首次观看（无历史进度）**的视频。

1. 进入播放器，控制条初始显示
2. **不按任何键**，等 5 秒控制条自动隐藏
3. 按 **DPAD_DOWN**

**预期**：同 R1

### 1.3 R3 — 字幕 / 音轨 picker 关闭后焦点自动回归

**前置**：选一个**有字幕轨**且**有多音轨**的视频。

A. **字幕路径**：

1. 进入播放器，按 DPAD_DOWN 唤起控制条
2. 按 ←/→ 把焦点移到"字幕"图标按钮，按 OK 打开字幕 picker dialog
3. 按 BACK 关闭 dialog（或按 OK 选中一条字幕）
4. 等 2 秒（避免被 picker 关闭瞬间的余焦掩盖）
5. 按 **DPAD_DOWN**

**预期**：
- [ ] 控制条出现，焦点在播放/暂停按钮上

B. **音轨路径**：同 A，把"字幕"换成"音轨"按钮

### 1.4 R4 — 返回二次确认提示消失后焦点自动回归

1. 进入播放器播放
2. 按 BACK → 屏幕底部出现"再按一次返回退出"二次确认提示
3. **不按任何键**，等提示自动淡出（约 2 秒）
4. 按 **DPAD_DOWN**

**预期**：
- [ ] 控制条出现，焦点在播放/暂停按钮上

### 1.5 R5 — DPAD_UP 退焦 + auto‑hide 之后焦点必须能再次进入

1. 进入播放器，按 DPAD_DOWN 唤起控制条，焦点落在播放/暂停按钮
2. 按 **DPAD_UP**（路由表的 ExitFocus，应该把 `focusInControls=false` 并把焦点回收到 root）
3. **不按任何键**，等 5 秒 auto‑hide
4. 按 **DPAD_DOWN**

**预期**：同 R1

### 1.6 R6 — ←/→ 连按快进/快退（修复后的关键回归）

1. 进入新视频，等 5 秒 auto‑hide（重现"焦点真空"）
2. 按 **← 三次**（每次间隔 < 300 ms）

**预期**：
- [ ] 第 1 次：触发快退反馈（屏幕中央显示"快退 -10 秒"或当前 seek 步长）
- [ ] 第 2 次：累加快退（"快退 -30 秒"，因为连按合并跳转）
- [ ] 第 3 次：再累加
- [ ] 控制条可见，但焦点不在按钮上（focusInControls=false，符合路由表 Seek 副作用）

按 → 同样测一遍。

### 1.7 R7 — CENTER 在任意焦点真空时刻都能切播放暂停

依次在 R1~R5 步骤的"按 DPAD_DOWN 之前"那一刻，**改按 CENTER（OK 键）**：

**预期**：
- [ ] 屏幕中央出现 Pause / Play 图标 + "已暂停" / "继续播放" 文案 800ms
- [ ] 视频实际播放状态切换

### 1.8 R8 — 控制条已聚焦时按 ←/→/CENTER 无回归

1. 进入新视频，按 DPAD_DOWN 唤起 UI
2. 焦点在播放/暂停按钮上时按 ←
3. 焦点在播放/暂停按钮上时按 →
4. 焦点在播放/暂停按钮上时按 CENTER

**预期**：
- [ ] ← 把焦点移到上一个按钮（按 controls focus 环绕规则）
- [ ] → 把焦点移到下一个按钮
- [ ] CENTER 触发当前焦点按钮的 onClick（播放/暂停）
- [ ] 全程不触发 Seek、不触发额外的播放暂停反馈
- [ ] 这是 PassThrough 路径的回归测试，确保新加的 onFocusChanged 兜底没破坏既有 focusInControls 状态转换

### 1.9 R9 — 电视剧路径同等通过

把 R1~R7 在**电视剧详情页**（`TvSeriesPlayerScreen`）重测一遍，使用一个有多集且至少有一集有历史进度的剧。

**预期**：
- [ ] 全部场景行为与电影/`18+` 路径一致
- [ ] 连播提示卡（如有）显示与消失不影响焦点行为

### 1.10 R10 — 既有交互回归

- [ ] 触屏 tap 切播放暂停正常
- [ ] 触屏 long press 触发 2x 倍速、松手恢复 1x
- [ ] 触屏 horizontal drag 进入拖动 seek 模式，松手 commit position
- [ ] 续播卡按 OK"继续观看"：卡片消失，焦点不再被卡片持有，后续 DPAD_DOWN 可用
- [ ] 续播卡按 ←/→ 切到"从头播放"再按 OK：mediaPlayer.time = 0，继续播放
- [ ] 字幕 picker 内 ←/→/↑/↓/OK/BACK 行为符合 [[TV 操作 UI 层]] 既有规范（不在本任务范围内修改）
- [ ] 音轨 picker 同上
- [ ] 长视频 5/10/15/20/30 秒快进步长设置仍生效
- [ ] BACK 双击退出（[[TV 播放器退出确认]]）行为不变

## 2. 自动化产物

提交里必须包含：

- [ ] `LongFormPlayerFocusGuard.kt`（生产）
- [ ] `LongFormPlayerFocusGuardTest.kt`（≥ 10 个 case，覆盖 implement §4.1 T1~T10）
- [ ] `LongFormVideoPlayerSpecTest.kt`（源文 audit，覆盖 implement §4.2 S1~S5）
- [ ] `LongFormVideoPlayer.kt` / `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 代码变更
- [ ] `android-tv-app/tv-app/build.gradle.kts` 版本号 +1
- [ ] `CONTEXT.md` 新增三条术语
- [ ] `plan.md` 追加条目

## 3. 文档与提交

- [ ] 提交信息为 implement §7 给出的中文格式
- [ ] 不修改 `.codex/skills/*`
- [ ] 用户在真机/模拟器跑完 §1 所有场景并确认 → 创建 `DONE.md` 记录验证范围、自动化命令输出摘要、人工验证场景的真实结果

## 4. 可能的回归风险与缓解

| 风险 | 缓解 |
|---|---|
| 续播卡内嵌后视觉位置漂移（padding 不一致） | implement §3.2 已规定按既有 `TvResumePromptTokens.HorizontalPaddingDp` / `BottomPaddingDp` 设置，R1 视觉确认 |
| `onFocusChanged` 兜底与 LaunchedEffect 双触发导致 `tryRequestFocus` 抢自己 | `pendingRootFocusRequest` 已是 one-shot；多触发不会引发副作用（每次都被 LaunchedTvInitialFocus 消费一次） |
| `backConfirmPromptVisible` / `playerErrorVisible` 透传后改变 `LongFormVideoPlayer` 公开签名 | 默认值 `false`，调用方不强制传；既有手机端/其他调用方无破坏 |
| Compose `focusable` 与 `onFocusChanged` 的事件顺序在不同版本表现差异 | 由 R3/R4 dialog dismiss 场景兜底验证；如果某 Compose 版本下兜底不触发，需在 implement.md §3.4 加描述并补加 LaunchedEffect 通道 |
| 电视剧屏的连播提示卡（如有 + 是 sibling 位置）依然在 LongFormVideoPlayer 外，焦点兜底覆盖不到 | implement §2.2 已在 `TvSeriesPlayerScreen` 改动里要求审查；如果存在独立连播提示卡，**必须**一并迁入 slot 或加入聚合可见性 state |

# Review：TV 播放续播提醒

- 日期：2026-05-23
- 关联 PRD：`prd.md`
- 关联 Implement：`implement.md`
- 任务三段执行流：当前阶段 = Review（CONTEXT.md「tasks 任务三段执行流」定义）

## 1. 自动化必跑（red → green）

```bash
cd android-tv-app
./gradlew :tv-app:testDebugUnitTest
./gradlew :tv-app:assembleDebug
```

必须全部通过；任一红则视为 review 失败，回 Implement 阶段修复，不得 skip 或注释掉用例。

### 1.1 单测覆盖核查

| 测试文件 | 必须包含的用例组 |
|---------|------------------|
| `TvResumePromptTest` | ① `shouldShowResumePromptCard` 守卫真值表（每个守卫翻转一次确认能 false 化 + 全部正向 + remainingMs > 0 → true；remainingMs ≤ 0 → false；`isBackConfirmVisible` 只影响可见性，不触发永久 dismiss）<br>② `shouldTickResumePromptCountdown` 守卫真值表（全部正向时 remainingMs = 0 仍 true，由 loop 设置永久 dismiss；任一临时隐藏 / 永久 dismiss 守卫翻转则 false）<br>③ `resumePromptCountdownTickRemaining` 边界：`5_000 → 5`、`4_999 → 5`、`4_001 → 5`、`4_000 → 4`、`1_001 → 2`、`1_000 → 1`、`1 → 1`、`0 → 0`、`-100 → 0`<br>④ `shouldTriggerResumePrompt`：`9_999 → false`、`10_000 → true`、`10_001 → true`、`Long.MAX_VALUE → true`、`0 → false`、`-1 → false`<br>⑤ `formatResumePromptTimestamp`：`0 → "0:00"`、`754_000 → "12:34"`、`5_025_000 → "1:23:45"`、`-100 → "0:00"`、`3_599_000 → "59:59"`、`3_600_000 → "1:00:00"` |
| `TvResumePromptCardSpecTest` | 见第 2 节源文 audit 列表 |

### 1.2 既有测试不得回归

- `TvSeriesAutoplaySpecTest`（连播卡）必须仍全绿
- `TvShapeAuditTest` 必须仍全绿（`RoundedCornerShape(N.dp)` 白名单 `{8, 16, 999}` 不变）
- `TvTypographySpecTest` / `TvFocusSpecTest` / `TvDetailPanelTokensTest` 等既有 TV 视觉 audit 必须仍全绿

## 2. 源文 audit 断言点（`TvResumePromptCardSpecTest` 锁定）

| # | 断言（出现 = 通过） | 文件 |
|---|---------------------|------|
| 1 | 含 `TvResumePromptCard(` | `feature/tv/TvLongFormPlayerScreen.kt` |
| 2 | 含 `TvResumePromptCard(` | `feature/tv/TvSeriesPlayerScreen.kt` |
| 3 | 含 `Alignment.BottomStart` | 两个 player 文件（调用点） |
| 4 | 含 `TvResumePromptTokens.HorizontalPaddingDp` 与 `TvResumePromptTokens.BottomPaddingDp` | 两个 player 文件（调用点） |
| 5 | 含 `LaunchedTvInitialFocus(` 与 `.tryRequestFocus()` | `feature/tv/TvResumePromptCard.kt` |
| 6 | 含 `AppChrome.SurfaceShape` 与 `AppChrome.ChipShape` | `feature/tv/TvResumePromptCard.kt` |
| 7 | 含 `tvFocusableGlow(` | `feature/tv/TvResumePromptCard.kt` |
| 8 | 含 `TvMotionTokens.DurationStandardMs` 与 `TvMotionTokens.EasingStandard` | `feature/tv/TvResumePromptCard.kt` |
| 9 | 含 `resumedFromHistoryVideoId == detail.id`（节流复用证据） | `feature/tv/TvLongFormPlayerScreen.kt` |
| 10 | 含 `resumedFromHistoryVideoId == uiState.currentVideoId`（节流复用证据） | `feature/tv/TvSeriesPlayerScreen.kt` |
| 11 | 含 `shouldTriggerResumePrompt(` | 两个 player 文件 |
| 12 | 含 `shouldShowResumePromptCard(` | 两个 player 文件 |
| 13 | **不**含裸字面量 `5_000L` / `5000L`（必须从 `TvResumePromptTokens.CountdownDurationMs` 取） | 两个 player 文件、`TvResumePromptCard.kt` |
| 14 | **不**含裸字面量 `10_000L` / `10000L`（必须从 `TvResumePromptTokens.MinResumeMs` 取） | 两个 player 文件、`TvResumePromptCard.kt` |
| 15 | **不**含裸 `padding(start = 48.dp` / `bottom = 156.dp` 字面量（必须从 `TvResumePromptTokens` 取） | 两个 player 文件 |
| 16 | 倒计时递减 `LaunchedEffect` 依赖 `shouldTickResumePromptCountdown`，不直接复用 `shouldShowResumePromptCard` | 两个 player 文件 |
| 17 | 不含用于「从头播放」按钮 focus state 冻结倒计时的状态或回调；焦点移动不冻结倒计时 | `TvResumePromptCard.kt`、两个 player 文件 |

## 3. 手测脚本（PRD 6.1–6.3 全部 18 条）

**预条件**：
- 准备一部电影、`18+` 视频、电视剧（含 ≥2 集），分别 seed `watchSeconds` 值满足各用例
- 设备登录、已配对、连得上后端
- TV 设置中「电视剧自动连播」保持默认（开启），「快进/快退步长」保持默认

**执行顺序**：按 H1–H18 顺序，任一不符合预期视为 review 失败。

**结果记录格式**（每条手测一行）：
```
H{N}：[通过 / 失败] - {简短观察 / 失败原因}
```

按 `prd.md` 第 6 节用例期望逐条核对，**不**允许只核对部分预期。

## 4. Done definition

通过本 review 必须**同时**满足：

- [ ] 第 1.1 节自动化命令全绿
- [ ] 第 1.1 节单测覆盖核查全部命中
- [ ] 第 1.2 节既有测试无回归
- [ ] 第 2 节源文 audit 全部命中
- [ ] 第 3 节 18 条手测全部通过
- [ ] `android-tv-app/tv-app/build.gradle.kts` 版本号已更新（`versionCode 62 → 63` / `versionName 0.1.61 → 0.1.62`）
- [ ] `CONTEXT.md` 已追加 6 条去实现细节的术语（语义对齐 PRD 第 4 节，未改动既有任何条目）
- [ ] `plan.md` 顶部已追加本次条目（implement.md 第 1.2 节模板）
- [ ] 实现改动落在**一个** git commit（中文 subject + 中文 body）；审查阶段的任务文档 / `CONTEXT.md` / `plan.md` 提交可独立存在

## 5. review 后流程

review 通过后：

1. **用户验收** → 在 `tasks/2026-05-23-tv-resume-from-history-prompt/` 下新增 `DONE.md`，记录：
   - 完成日期
   - 关联 commit hash
   - 验证摘要（自动化全绿 + 18 条手测全通）
2. **若用户提出反馈** → 新增 `feedback.md` 记录反馈内容；按 CONTEXT.md「tasks 任务三段执行流」规则，已 `DONE.md` 的任务**不**自动重开，只作为归档材料；用户明确要求重开或复查时才重新进入三段流程
3. **不**允许只完成实现跳过 review 直接宣称任务完成（CONTEXT.md 明文要求）

## 6. review 失败处理

- **自动化红**：回 Implement 阶段修复，不得 skip 或注释掉用例；修复后回到第 1 节重新跑全套
- **源文 audit 红**：检查是 spec 错（回 PRD 校准并写入 `plan.md`）还是接入代码错（回 Implement）
- **手测异常**：判断是 spec 错（回 PRD 校准并写入 `plan.md`）还是实现错（回 Implement），并在 `plan.md` 追加调整条目；不允许偷偷下调验收门槛

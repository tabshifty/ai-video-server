---
name: repo-dev-workflow
description: Use when implementing, fixing, verifying, or preparing a commit in this repository, especially when work touches Go backend, Android phone app, Android TV app, admin web, plan.md, AGENTS.md, Chinese UI text, or repo-local skills.
---

# 仓库开发流程

## 核心原则

把 `plan.md` 当作开发账本：先记录计划，再做红灯测试和实现，最后记录验证与提交范围。所有 Markdown、提交信息和面向用户的界面文案默认使用中文，并确认没有乱码。`CONTEXT.md` 只沉淀长期有效的术语、架构决策、接口约定、兼容策略和踩坑经验，不写一次性进度。

## 开始前

- 先读取当前目录适用的 `AGENTS.md`；遇到更深层 `AGENTS.md` 时，以更深层规则为准。
- 运行 `git status --short`，识别已有用户改动；只处理本任务文件，不回退无关变更。
- 判断影响模块：后端 Go、`android-app`、`android-tv-app`、`admin-web`、部署/脚本、文档或 skill。
- 不修改 `.codex/skills/*`，除非任务明确要求创建或维护 skill。
- Android 构建依赖 Java 17；本机优先使用 `~/.gradle/gradle.properties` 的 `org.gradle.java.home=/Users/cuiqi/.jdks/jdk-17.0.19+10/Contents/Home`。若 Gradle 报 Java 8/variant mismatch，先查 `java -version` 与 `org.gradle.java.home`，不要改项目源码绕过。
- 若涉及家用部署机、部署脚本或 migration，先读 `CONTEXT.md` 的部署术语和 `docs/家用部署机.md`；部署默认是 push 到家用部署机后 hard restart Go server/worker，数据层独立生命周期。

## plan.md 记录

- 每个新任务开始前，向 `plan.md` 顶部追加一条反向时间顺序记录，不覆盖历史。
- 记录至少包含：时间、进度摘要、影响文件、待执行验证。
- 实现中出现关键阶段时继续追加：红灯结果、核心实现、定向验证、全量验证。
- 完成前追加最终记录，写明只纳入哪些文件，并说明无关工作区变更未纳入。
- 如果 `plan.md` 开始时已有用户未提交内容，继续在工作区顶部追加本任务记录；提交时只把本任务新增记录精确暂存，保留其它工作区差异。

建议格式：

```markdown
## YYYY-MM-DD HH:MM +0800
- 进度：...
- 影响文件：`path/to/file`、`plan.md`
- 验证：待执行 ...；或 `command` 通过。
```

## 实施节奏

- 优先 TDD：先补能失败的定向测试或静态检查，再实现最小改动。
- 对业务规则新增纯逻辑 helper 时，同步补纯逻辑单测，降低 Compose/UI/网络测试成本。
- 对 UI 改动，锁定可回归的策略、文案、尺寸或路由 helper；必要时再补构建验证。
- 对跨端或跨层改动，先分别做模块内定向验证，再做受影响模块全量验证。
- 保持改动最小，避免顺手重构、格式化无关文件或纳入既有 skill 工作区变更。
- App 功能修改必须同步更新对应版本号：手机端 `android-app/app/build.gradle.kts`，TV 端 `android-tv-app/tv-app/build.gradle.kts`；默认 `versionCode +1`、`versionName` patch `+1`。
- 功能更新若改变长期术语、接口、部署、兼容或踩坑经验，追加 `CONTEXT.md`；仅临时进度写 `plan.md`，不要把流水账塞进 `CONTEXT.md`。
- 新增 migration 必须满足“上一版本二进制仍能在新 schema 上运行”的前向兼容契约：优先加表/可空列/默认值/索引；避免 rename；删列先停用再删。

## 任务目录三段流

当用户要求完成 `tasks/` 下任务时：

- 跳过已有 `DONE.md` 的任务，除非用户明确要求重开。
- 按 `prd.md` → `implement.md` → `review.md` 顺序读取和执行；缺文件时在 `plan.md` 记录缺口，只有能从上下文重建时才继续。
- 实现不能跳过 review 阶段；review 清单里的人工验证无法执行时要写明原因。
- 用户确认验收后，新增 `tasks/<任务名>/DONE.md`，写完成日期、关联提交、验证摘要，并单独提交该 marker。

## 常用验证选择

- Go 后端：优先 `go test ./internal/...` 的定向包和 `-run`；收尾常用 `go test ./... -count=1`、`go vet ./...`。
- Go 转码/ffmpeg：常用 `go test ./pkg/ffmpeg ./internal/services -run 'TestBuildTranscode|TestParseProbe|TestResolveProbe|TestBuildTranscodePlan' -v`。
- 手机端：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest`，需要构建时加 `:app:assembleDebug`；若只改纯 helper，可用 `--tests` 跑定向类。
- TV 端：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`，需要构建时加 `:tv-app:assembleDebug`；定向常用 `--tests com.chee.videos...`。LibVLC/IPTV/播放器改动优先跑对应源文测试和纯逻辑测试，再视风险补 `compileDebugKotlin` 或 assemble。
- 管理端：按 `admin-web` 现有脚本执行，通常先 `npm run build`；不要新增测试框架来满足普通 UI 调整。
- 部署脚本：优先做 shell 语法/静态检查和 dry-run；涉及家用部署机时核对 `docs/家用部署机.md` 的 push 分桶、fail-open 和绝对路径契约。
- 文档或 skill：运行相关静态校验即可，例如 skill-creator 的 `quick_validate.py`、`git diff --check` 和乱码搜索。

## 收尾与提交

- 复查 `git status --short` 和 `git diff --stat`，确认无关改动未暂存。
- 追加 `plan.md` 完成记录，写明通过的命令；不能运行的验证必须说明原因。
- 运行 `rg -n $'\uFFFD'` 检查本次 Markdown、中文文案或 skill 文件没有乱码。
- 每个完成变更都要提交；提交信息使用中文且无乱码。
- 提交时精确暂存本任务文件，避免把用户已有改动或不相关 skill 变化纳入；若文件同时含用户旧改动和本任务改动，使用 patch/索引级暂存，只提交本任务片段。
- 提交后再次 `git status --short`，明确剩余未提交项是否为用户既有改动。

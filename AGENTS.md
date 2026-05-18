# AGENTS.md

## Scope
- This file applies to the entire repository: `D:\workspace\git\App\video-test`.
- If a deeper `AGENTS.md` is added later, deeper scope overrides parent scope for files under that directory.

## Repository Snapshot
- This repo currently contains:
  - `android-app`（Android 手机端工程）
  - `android-tv-app`（Android TV 独立工程）
  - `.codex/skills/agents-md` (skill definition for generating AGENTS docs)
  - `.codex/skills/ui-ux-pro-max` (UI/UX skill assets/scripts)

## Global Rules
- Keep changes minimal and scoped to the requested task.
- Do not modify skill files under `.codex/skills/*` unless the task is explicitly about skill authoring.
- Prefer fast search/read commands (`rg`, targeted file reads) over broad scans.
- Before claiming completion, run only the verifications that actually exist in this repo.
- 所有 Markdown 文件以及前端面向用户的界面文案默认必须使用中文；仅在任务明确要求其他语言时可例外。
- 中文内容必须保证编码与显示正确，禁止出现乱码（包括 Markdown、注释、界面文案）。
- Git 提交信息默认使用中文，且必须无乱码。
- App 功能修改必须同步更新受影响 App 的版本号：手机端为 `android-app/app/build.gradle.kts`，TV 端为 `android-tv-app/tv-app/build.gradle.kts`；默认 `versionCode +1`，`versionName` 的 patch 位 `+1`。
- 每次功能更新必须在根级 `CONTEXT.md` 追加技术沉淀，记录长期有效的术语、架构决策、接口约定、兼容策略或踩坑经验，避免只写临时进度流水账。

## Change Management Rules
- Every change must be rollbackable:
  - Prefer small, scoped commits.
  - Keep migration and code changes reversible where practical.
- Every completed change must be committed to git; do not leave accepted work only in the working tree.
- `plan.md` is mandatory and append-only for progress tracking:
  - Every new plan must be appended as an incremental entry.
  - Every implementation update must append what changed (not overwrite old entries).
  - Maintain reverse-chronological entries with date/time, summary, affected files, and verification status.

## Module Map (Current)
| Module | Path | Purpose | Notes |
|---|---|---|---|
| Root workspace | `.` | Project root and coordination point | Put cross-module rules here |
| Android app | `android-app` | 手机端 Android 工程 | 已有模块级 AGENTS.md |
| Android TV app | `android-tv-app` | TV 独立 Android 工程 | 已与手机端源码目录隔离 |
| Agents skill | `.codex/skills/agents-md` | AGENTS.md generation guidance | Treat as managed skill content |
| UI/UX skill | `.codex/skills/ui-ux-pro-max` | UI/UX reference assets and scripts | Read-only unless requested |

## Cross-Domain Workflow
- Current state: `android-app` 与 `android-tv-app` 是相互隔离的独立 App 工程，暂无共享运行时代码依赖。
- When adding a real module (for example `frontend/` or `backend/`):
  - Create a nested `AGENTS.md` in that module root.
  - Document run/test/build commands owned by that module.
  - Document contracts between modules (API schema, shared types, env vars).

## Verification Guidance
- For documentation-only edits (like this file), no build/test is required.
- For future code modules, include in that module's `AGENTS.md`:
  - local run command
  - test command
  - build/lint command
  - required environment variables

## Maintenance
- Keep this file concise; move long details to module-level `AGENTS.md`.
- Update the module map whenever top-level directories or ownership boundaries change.

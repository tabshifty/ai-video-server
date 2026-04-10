# AGENTS.md

## Scope
- This file applies to the entire repository: `D:\workspace\git\App\video-test`.
- If a deeper `AGENTS.md` is added later, deeper scope overrides parent scope for files under that directory.

## Repository Snapshot
- This repo currently contains:
  - `.codex/skills/agents-md` (skill definition for generating AGENTS docs)
  - `.codex/skills/ui-ux-pro-max` (UI/UX skill assets/scripts)
- No application runtime modules (`frontend/`, `backend/`, `infra/`) are present yet.

## Global Rules
- Keep changes minimal and scoped to the requested task.
- Do not modify skill files under `.codex/skills/*` unless the task is explicitly about skill authoring.
- Prefer fast search/read commands (`rg`, targeted file reads) over broad scans.
- Before claiming completion, run only the verifications that actually exist in this repo.
- 所有 Markdown 文件以及前端面向用户的界面文案默认必须使用中文；仅在任务明确要求其他语言时可例外。

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
| Agents skill | `.codex/skills/agents-md` | AGENTS.md generation guidance | Treat as managed skill content |
| UI/UX skill | `.codex/skills/ui-ux-pro-max` | UI/UX reference assets and scripts | Read-only unless requested |

## Cross-Domain Workflow
- Current state: no cross-domain runtime workflow exists (no app modules yet).
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

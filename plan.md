# plan.md

本文件用于增量记录“计划与修改”，不得覆盖历史记录，只能追加。

## Entry Template
### [YYYY-MM-DD HH:MM] Title
- Type: `plan` | `implementation` | `docs`
- Summary:
- Changed Files:
  - `path/to/file`
- Verification:
  - `command/result`
- Rollback:
  - `git revert <commit>`

---

### [2026-04-09 09:40] Add change-management conventions
- Type: `docs`
- Summary:
  - Added repository-level conventions: rollbackable changes, mandatory git commits, and mandatory incremental updates to `plan.md`.
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - Documentation-only change; no build/test required.
- Rollback:
  - Revert the commit that introduces this entry.

### [2026-04-09 09:33] Implement upload dedupe and instant-upload flow
- Type: `implementation`
- Summary:
  - Implemented hash verification, global dedupe (`file_hashes`), `/api/v1/upload/check`, updated upload handler/service/repository, queue payload/idempotency, and temp cleanup command.
- Changed Files:
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_check.go`
  - `internal/services/upload.go`
  - `internal/repository/video_repository.go`
  - `internal/queue/tasks.go`
  - `internal/hashutil/hash.go`
  - `migrations/0004_file_hashes.up.sql`
  - `migrations/0004_file_hashes.down.sql`
  - `cmd/cleanup-temp/main.go`
  - `main.go`
  - `internal/config/config.go`
  - `.env.example`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing the upload dedupe implementation.

### [2026-04-09 11:05] Add admin movie/episode upload with async scrape pipeline
- Type: `implementation`
- Summary:
  - Added role-based upload restrictions (`user` can only upload `short`, `admin` can upload `short/movie/episode`).
  - Added asynchronous scrape tasks for `movie/episode` uploads and chained transcode enqueue after scrape.
  - Added filename parser `ParseFilename` for movie and episode patterns and updated scrape service to write TMDB metadata, poster thumbnail, and `tmdb_id`.
  - Added migration for `videos.tmdb_id` and new `scraping` status.
- Changed Files:
  - `internal/handlers/upload.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/queue/tasks.go`
  - `internal/services/scraper.go`
  - `internal/services/upload.go`
  - `internal/repository/video_repository.go`
  - `internal/utils/filename_parser.go`
  - `internal/utils/filename.go`
  - `internal/utils/filename_test.go`
  - `internal/models/models.go`
  - `main.go`
  - `migrations/0005_video_scrape_fields.up.sql`
  - `migrations/0005_video_scrape_fields.down.sql`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 11:45] Add admin scrape preview and manual confirm APIs
- Type: `implementation`
- Summary:
  - Added admin-only scrape management endpoints: preview TMDB candidates and confirm manual metadata edits.
  - Added in-memory preview cache (5 minutes), configurable poster storage path, and poster download for confirm flow.
  - Extended scraper service with preview/confirm methods for movie and TV episode; confirm updates videos metadata and episode linkage.
- Changed Files:
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/router.go`
  - `internal/services/scraper.go`
  - `internal/config/config.go`
  - `main.go`
  - `.env.example`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 12:20] Add API docs generation and Swagger UI
- Type: `implementation`
- Summary:
  - Added API docs generation command `go run ./cmd/gen-openapi` and wired `go:generate` in `main.go`.
  - Added runtime Swagger docs exposure under `/swagger/index.html` and `/swagger/openapi.json` controlled by `ENABLE_SWAGGER`.
  - Added OpenAPI base metadata and handler documentation models/comments for key endpoints.
- Changed Files:
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/index.html`
  - `docs/swagger/openapi.json`
  - `internal/config/config.go`
  - `.env.example`
  - `main.go`
  - `internal/handlers/router.go`
  - `internal/handlers/swagger_models.go`
  - `internal/handlers/auth.go`
  - `internal/handlers/upload_check.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/admin_scrape.go`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
  - `go run ./cmd/gen-openapi` generated docs files successfully.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 12:55] Add one-command run scripts
- Type: `implementation`
- Summary:
  - Added Linux/macOS one-command startup and shutdown scripts.
  - `dev-up.sh` now starts postgres/redis, waits for DB readiness, runs all migrations, and starts server/worker with logs + PID files.
  - Added run guide documentation and `.run/` git ignore.
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `docs/run.md`
  - `.gitignore`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:10] Fix duplicate migration constraint on one-click startup
- Type: `implementation`
- Summary:
  - Fixed `0002_auth.up.sql` to make FK and column changes idempotent.
  - Updated `scripts/dev-up.sh` to use `schema_migrations` and skip already-applied migrations.
  - This prevents repeated startup from failing with `constraint ... already exists`.
- Changed Files:
  - `migrations/0002_auth.up.sql`
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:25] Add explicit ENV_FILE loading for startup
- Type: `implementation`
- Summary:
  - Added `ENV_FILE` support in `main.go` (`godotenv.Overload` when provided).
  - Updated `scripts/dev-up.sh` to load and export env vars from `ENV_FILE` (defaults to project `.env`).
  - Updated run docs to show custom env file usage.
- Changed Files:
  - `main.go`
  - `scripts/dev-up.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 13:35] Harden .env parsing in dev-up script
- Type: `implementation`
- Summary:
  - Replaced `source .env` with safe line-by-line env parsing in `scripts/dev-up.sh`.
  - Parser now handles BOM/CRLF, ignores blank/comment lines, skips invalid lines, and supports quoted values.
  - Updated run documentation accordingly.
- Changed Files:
  - `scripts/dev-up.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `go test ./...` passed (with local `GOCACHE`).
  - `bash -n scripts/dev-up.sh` could not be run in this Windows shell (`bash` not installed).
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:15] Implement admin web and admin management APIs integration
- Type: `implementation`
- Summary:
  - Added backend admin management APIs for stats/videos/users/tasks/system operations and chunked upload endpoints.
  - Added SSE event stream endpoint at `/api/v1/admin/events/ws` and static hosting of built admin frontend under `/admin`.
  - Generated full `admin-web` (Vue3 + Vite + Element Plus + Pinia + axios) with login, dashboard, video/upload/scrape/user/system/task pages and API integration.
  - Added module-level `admin-web/AGENTS.md` and included newly added skill assets/directories in workspace scope as requested.
- Changed Files:
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/admin_events.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/models/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `main.go`
  - `admin-web/*`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
  - `npm --prefix admin-web install` succeeded.
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:27] Extend one-command scripts to include admin frontend
- Type: `implementation`
- Summary:
  - Extended `scripts/dev-up.sh` to support `--frontend dev|build|off` with default `dev`.
  - Added auto-install of frontend dependencies when `admin-web/node_modules` is missing.
  - Added frontend dev process management (`.run/frontend.pid`, `.run/frontend.log`) and build-mode output guidance.
  - Updated `scripts/dev-down.sh` to stop frontend process and refreshed run documentation.
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
  - `bash -n scripts/dev-down.sh` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:41] Polish admin-web UI with unified design system
- Type: `implementation`
- Summary:
  - Applied `ui-ux-pro-max` design direction to admin frontend with a unified visual system (fonts, colors, cards, tables, transitions, responsive behavior).
  - Refined shared layout shell (brand sidebar, icon navigation, sticky top bar) and upgraded all key views for stronger hierarchy and readability.
  - Enhanced chart styling, upload progress presentation, scrape candidate card interactions, and system log panel legibility.
- Changed Files:
  - `admin-web/src/assets/theme.css`
  - `admin-web/src/main.js`
  - `admin-web/src/App.vue`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/components/UploadProgress.vue`
  - `admin-web/src/views/Login.vue`
  - `admin-web/src/views/Dashboard.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/UserManage.vue`
  - `admin-web/src/views/SystemSettings.vue`
  - `admin-web/src/views/TaskMonitor.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 17:56] Add tag input to admin video upload form
- Type: `implementation`
- Summary:
  - Added missing upload form control for video tags using multi-select with create mode.
  - Added preset tag options and kept support for custom tag creation.
  - Added submit-time tag normalization (trim, empty filtering, lowercase dedupe) before calling `/upload/init`.
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:16] Fix dev-mode LAN access and Vite API proxy 404
- Type: `implementation`
- Summary:
  - Updated Vite dev server to listen on `0.0.0.0:5173` with strict port behavior for LAN access.
  - Switched proxy matching to `/api/v1` and made proxy backend target configurable via `VITE_API_PROXY_TARGET`.
  - Updated frontend and run documentation with LAN usage and `/api/v1/*` 404 troubleshooting steps.
- Changed Files:
  - `admin-web/vite.config.js`
  - `admin-web/.env.development`
  - `admin-web/README.md`
  - `docs/run.md`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` succeeded.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:28] Fix backend admin SPA route conflict causing startup panic
- Type: `implementation`
- Summary:
  - Fixed Gin route conflict between `/admin/assets/*filepath` and `/admin/*path` by replacing wildcard route with `NoRoute` fallback logic for `/admin/`.
  - Preserved static asset serving for `/admin/assets/*` and SPA index fallback for non-asset `/admin/*` paths.
  - This removes startup panic and restores normal admin API availability when server runs with latest code.
- Changed Files:
  - `internal/handlers/router.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:33] Harden dev-up startup checks to catch false-positive server start
- Type: `implementation`
- Summary:
  - Added background process startup helper in `scripts/dev-up.sh` with post-start liveness checks.
  - If `server`/`worker`/`frontend` exits immediately (for example, port conflict on `:8080`), script now fails fast and prints recent log tail.
  - Prevents misleading "starting server" success when actual process has already crashed.
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-09 19:47] Fix admin video detail NULL scan failure
- Type: `implementation`
- Summary:
  - Fixed `AdminVideoDetail` query scanning failures when nullable video columns contain `NULL`.
  - Added SQL-side `COALESCE` defaults for nullable detail fields (`duration_seconds`, `width`, `height`, path fields, metadata, and text fields) before scanning into Go struct fields.
  - Prevents `cannot scan NULL into *int` errors on video detail retrieval.
- Changed Files:
  - `internal/repository/admin_repository.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed.
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 13:52] 新增 Markdown 与前端文案中文约束
- Type: `implementation`
- Summary:
  - 在仓库全局规则中新增约束：所有 Markdown 文件及前端面向用户的界面文案默认使用中文。
  - 仅当任务明确要求其他语言时，允许例外。
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - 文档变更，无需构建或测试。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:15] 修复转码回退与失败源文件保留策略（计划）
- Type: `plan`
- Summary:
  - 检查上传后转码触发链路，确认 `short` 直转码、`movie/episode` 抓取后转码。
  - 调整转码元数据提取：`ffprobe` 失败或缺字段时写入默认值并保留可编辑元数据，不让任务因探测失败中断。
  - 调整临时清理保护条件，避免 `failed`/`scraping` 视频源文件被误删，保留给管理员决策。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/repository/video_repository.go`
- Verification:
  - 待执行：`go test ./...`
- Rollback:
  - Revert the commit containing this entry.
### [2026-04-10 14:20] 修复转码元数据回退与失败源文件保留
- Type: `implementation`
- Summary:
  - 保持现有上传转码链路：`short` 上传后直接转码，`movie/episode` 抓取后继续转码。
  - 调整 `TranscodeService`：`ffprobe` 失败不再导致转码任务失败，改为写入默认值（`duration/width/height=0`, `codec=unknown`）并记录 `probe_error`，便于管理员后续手动改写。
  - 调整临时文件清理保护查询，新增保护 `scraping` 与 `failed` 状态视频的 `original_path`，避免自动清理误删失败源文件。
  - 新增单测覆盖元数据回退与正常探测场景。
- Changed Files:
  - `internal/services/transcode.go`
  - `internal/services/transcode_test.go`
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `go test ./...` failed（现有跨平台编译问题：`internal/handlers/admin.go` 使用 `syscall.Statfs_t/Statfs` 在当前 Windows 环境不可用）。
  - `go test ./internal/services ./internal/repository ./internal/queue` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 14:23] 强化中文无乱码约束并修复计划文档乱码
- Type: `implementation`
- Summary:
  - 更新仓库级 `AGENTS.md`：新增“中文内容禁止乱码”与“提交信息使用中文且无乱码”的强约束。
  - 修复 `plan.md` 顶部说明文字乱码。
  - 修复 `plan.md` 中 `[2026-04-10 13:52]` 条目的中文乱码内容。
- Changed Files:
  - `AGENTS.md`
  - `plan.md`
- Verification:
  - 文档变更，人工检查文本显示正常，无乱码。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:28] 新增可播放视频源接口（计划）
- Type: `plan`
- Summary:
  - 新增鉴权播放源接口 `GET /api/v1/videos/:id/source`，用于直接返回可播放视频流。
  - 限定仅 `ready` 状态可播放，未就绪返回 `409`，不存在返回 `404`，未登录返回 `401`。
  - 在视频详情接口中新增 `play_url`（相对路径），减少客户端拼接逻辑。
  - 更新 OpenAPI 生成配置并补充最小单测。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_detail.go`
  - `internal/models/app.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
- Verification:
  - 待执行：`go test ./...`
- Rollback:
  - Revert the commit containing this entry.

### [2026-04-10 14:35] 实现可播放视频源接口与详情 play_url
- Type: `implementation`
- Summary:
  - 新增 `GET /api/v1/videos/:id/source`（鉴权）并返回视频流；仅 `ready` 可播放。
  - 新增错误返回策略：`401`（未登录）、`400`（非法 ID）、`404`（视频或源文件不存在）、`409`（视频未就绪）。
  - 视频详情响应新增 `play_url` 字段，固定为相对路径 `/api/v1/videos/{id}/source`。
  - 新增 `internal/utils/VideoPlayURL` 与对应单测，确保播放地址生成稳定。
  - 更新 `cmd/gen-openapi` 并重新生成 `docs/swagger/openapi.json`，加入新接口说明。
- Changed Files:
  - `internal/handlers/router.go`
  - `internal/handlers/video_source.go`
  - `internal/handlers/video_detail.go`
  - `internal/models/app.go`
  - `internal/utils/video_url.go`
  - `internal/utils/video_url_test.go`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/services ./internal/repository ./internal/queue ./pkg/ffmpeg ./cmd/gen-openapi` passed。
  - `go test ./...` failed（既有 Windows 编译问题：`internal/handlers/admin.go` 中 `syscall.Statfs_t/Statfs` 不可用）。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 14:48] 管理端新增签名播放预览功能（计划）
- Type: `plan`
- Summary:
  - 为 admin-web 视频详情新增可播放预览能力，解决 `<video>` 无法携带 Bearer 头问题。
  - 后端新增管理员获取临时签名播放地址接口，并新增签名流播放接口。
  - 签名默认有效期 10 分钟，签名口径绑定 `video_id + exp`。
  - 同步更新 OpenAPI 与前端页面文案，确保中文无乱码。
- Changed Files:
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/video_source.go`
  - `internal/utils/play_url_sign.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoList.vue`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
- Verification:
  - 待执行：`go test ./...`
  - 待执行：`npm --prefix admin-web run build`
- Rollback:
  - Revert the commit containing this entry.

### [2026-04-10 14:59] 实现管理端签名播放预览与临时播放链接
- Type: `implementation`
- Summary:
  - 新增配置项 `PLAY_URL_SIGN_SECRET`（为空时回退 `JWT_SECRET`），用于签名播放地址。
  - 新增管理员接口 `GET /api/v1/admin/videos/:id/play-url`，返回 `signed_url` 与 `expires_at`。
  - 新增签名播放接口 `GET /api/v1/videos/:id/source/signed`，支持签名校验、过期校验与 `ready` 状态校验。
  - 保留原接口 `GET /api/v1/videos/:id/source`（Bearer 鉴权）不变，并复用统一可播放源检查逻辑。
  - 新增签名工具与单测：`SignVideoSource` / `VerifyVideoSourceSign`。
  - 管理端 `VideoList` 详情弹窗新增播放器预览、刷新播放链接按钮、到期时间展示与非 ready 状态提示。
  - 修复 `VideoList` 页面中文文案乱码，确保界面文案为中文且无乱码。
  - 更新 OpenAPI 生成器并重新生成 `docs/swagger/openapi.json`，补充两个新接口文档。
- Changed Files:
  - `.env.example`
  - `internal/config/config.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/video_source.go`
  - `internal/utils/play_url_sign.go`
  - `internal/utils/play_url_sign_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/views/VideoList.vue`
  - `cmd/gen-openapi/main.go`
  - `docs/swagger/openapi.json`
  - `main.go`
  - `plan.md`
- Verification:
  - `go test ./internal/utils ./internal/services ./internal/repository ./internal/queue ./pkg/ffmpeg ./cmd/gen-openapi` passed。
  - `go test ./...` failed（既有 Windows 编译问题：`internal/handlers/admin.go` 中 `syscall.Statfs_t/Statfs` 不可用）。
  - `npm --prefix admin-web run build` failed（当前环境 Node 权限问题：`EPERM: operation not permitted, lstat 'C:\Users\CheeTsui'`）。
- Rollback:
  - Revert the commit containing this feature entry.
### [2026-04-10 17:44] 修复视频删除时 NULL 扫描导致失败
- Type: `implementation`
- Summary:
  - 修复 `AdminDeleteVideo` 依赖的 `GetVideoByID` 查询在历史数据存在 NULL 时扫描失败的问题（`duration_seconds/width/height` 等字段）。
  - 在 `GetVideoByID` 与 `GetVideoByOriginalPath` 中统一使用 SQL `COALESCE` 提供默认值，避免 `cannot scan NULL into *int/*string`。
  - 删除逻辑保持容错：即使磁盘文件不存在也不会阻断删除（`os.Remove` 错误已忽略）。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `plan.md`
- Verification:
  - `go test ./internal/repository ./internal/services ./internal/queue ./internal/utils` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 17:53] 修复删除视频时 NULL 数值字段扫描报错
- Type: `implementation`
- Summary:
  - 修复 `GetVideoByID`/`GetVideoByOriginalPath` 在 `duration_seconds/width/height` 为 `NULL` 时扫描到 `int` 触发报错的问题。
  - 将上述字段改为先扫描 `sql.NullInt32`，再统一转换为 `int`，避免 `cannot scan NULL into *int`。
  - 新增回归测试覆盖 `NULL -> 0` 和有效值转换行为。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_repository_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 18:41] 修复分片上传后 worker 转码找不到源文件
- Type: `implementation`
- Summary:
  - 定位到 `UploadComplete` 会在入队后执行 `chunkUpload.Abort`，导致会话目录被删除，连同 `assembled-*.mp4` 一起被删。
  - 调整 `ChunkUploadService.Complete`：合并文件改为写入 `UPLOAD_TEMP_DIR/assembled/`，不再位于 `chunk-sessions/<session_id>/` 下。
  - 保持 `Abort(session_id)` 只清理分片会话目录，避免误删待转码输入文件。
  - 新增回归测试，覆盖“Complete 后 Abort，合并文件仍可读取”行为。
- Changed Files:
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run TestChunkUploadCompleteFileSurvivesAbort -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 19:40] 修复删除转码失败视频的外键约束报错
- Type: `implementation`
- Summary:
  - 修复删除视频时因 `transcoding_jobs.video_id` 与 `user_video_actions.video_id` 外键依赖导致的 `SQLSTATE 23503`。
  - 将 `DeleteVideoByID` 改为事务删除：先删 `transcoding_jobs`、再删 `user_video_actions`、最后删 `videos`。
  - 提取 `deleteVideoDependencies` 并补充顺序与错误分支回归测试。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_delete_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run TestDeleteVideoDependencies -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 21:28] 修复删除视频与转码任务并发导致的外键竞态
- Type: `implementation`
- Summary:
  - 针对“仍偶发 `transcoding_jobs_video_id_fkey`”问题，补充并发安全：删除事务先对目标视频行执行 `FOR UPDATE` 锁定。
  - 在锁定后再顺序删除 `transcoding_jobs`、`user_video_actions`、`videos`，避免 worker 并发插入同 `video_id` 任务造成竞态。
  - 更新回归测试，覆盖加锁后的 SQL 执行顺序与错误分支。
- Changed Files:
  - `internal/repository/video_repository.go`
  - `internal/repository/video_delete_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run TestDeleteVideoDependencies -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry.

### [2026-04-10 21:37] 通过外键级联删除消除视频删除阻塞
- Type: `implementation`
- Summary:
  - 新增迁移将 `transcoding_jobs.video_id` 与 `user_video_actions.video_id` 外键改为 `ON DELETE CASCADE`。
  - 解决删除视频时依赖记录并发/残留导致的 `SQLSTATE 23503` 阻塞问题。
  - 已在当前数据库执行迁移并核验 `confdeltype='c'`（cascade）。
- Changed Files:
  - `migrations/0006_video_fk_cascade.up.sql`
  - `migrations/0006_video_fk_cascade.down.sql`
  - `plan.md`
- Verification:
  - `docker exec -i video_server_postgres psql -U video -d video_server -v ON_ERROR_STOP=1 < migrations/0006_video_fk_cascade.up.sql` applied successfully。
  - `docker exec -i video_server_postgres psql -U video -d video_server -tA -c "SELECT conname, confdeltype FROM pg_constraint ..."` returned `c` for both constraints。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository` passed。
- Rollback:
  - Revert the commit containing this feature entry, then执行 `migrations/0006_video_fk_cascade.down.sql`。

### [2026-04-16 12:26] 优化开发脚本进程管理与前端依赖安装策略
- Type: `implementation`
- Summary:
  - 优化 `dev-up/dev-down` 的 PID 管理：PID 文件升级为“PID+命令指纹”，避免 PID 复用导致误判或误杀。
  - `dev-down` 新增优雅停止等待与超时强制终止，避免仅发信号后立即返回导致残留进程占端口。
  - `dev-up` 新增启动失败自动清理机制：若后续步骤失败，会回收本次新启动的后台进程，减少半启动状态。
  - 前端依赖安装改为基于 `package-lock.json` 哈希判断，锁文件变化或缺依赖时执行 `npm ci`，提升环境一致性。
  - 迁移执行移除无必要管道（`cat | docker exec`），改为输入重定向，简化执行链路。
- Changed Files:
  - `scripts/dev-up.sh`
  - `scripts/dev-down.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh scripts/dev-down.sh` passed.
  - `bash scripts/dev-up.sh --help` passed.
- Rollback:
  - `git revert <commit>`

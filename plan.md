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

### [2026-04-17 14:12] 管理端 token 失效统一处理（提示+跳转登录）
- Type: `implementation`
- Summary:
  - 修复管理端在 token 失效时未统一报错或跳转登录的问题。
  - 在前端请求响应拦截器中新增业务错误码鉴权判断：当返回 `code=401/403` 或消息包含 token/authorization/admin only 时，统一清理登录态并提示后跳转登录页。
  - 保留 HTTP 401 分支处理，并改为复用统一逻辑，避免重复代码和多次重定向。
- Changed Files:
  - `admin-web/src/api/request.js`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 14:03] dev-up 启动增加镜像预检与 descriptor 异常重试提示
- Type: `implementation`
- Summary:
  - 针对 `unable to fetch descriptor ... content size of zero` 启动失败场景，在 `dev-up.sh` 新增镜像预检流程。
  - 新增镜像拉取重试逻辑：仅在镜像本地不存在时执行 `docker pull`，并对 descriptor 异常给出可执行修复指引（镜像源配置问题排查）。
  - 新增 `compose_up_db_services`，在支持 `--pull` 时使用 `--pull missing`，减少重复启动时不必要拉取导致的失败概率。
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed.
  - `bash scripts/dev-up.sh --frontend off` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 12:50] AV 刮削框架化重构计划（参考 mdcx）与技能沉淀
- Type: `plan`
- Summary:
  - 参考 `references/mdcx` 的 crawler provider、模板流程与 parser 抽象，重构 Go AV 刮削内部架构。
  - 交付“技能+代码”：先沉淀可复用技能，再实现 JavDB 强化和多站点扩展预留。
  - 保持现有外部接口兼容，增强编号解析、trace 可观测性、演员自动关联稳定性。
- Changed Files:
  - `plan.md`
- Verification:
  - 计划项，无需构建/测试。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 12:50] 实现 AV crawler provider + 模板流程重构、trace 与编号匹配增强
- Type: `implementation`
- Summary:
  - 新增 AV 框架层：`avCrawlerProvider`、`avCrawler`、`avDetailParser`、`avScrapeRunContext`，并实现 JavDB crawler + parser + post-process。
  - `ScrapeAVUpload` 与 `ConfirmAV` 写入 `scrape_trace`，包含搜索词、URL、步骤、错误与匹配策略，用于排障与持续优化。
  - 升级番号提取逻辑，覆盖 `FC2PPV/FC2-PPV`、`HEYZO`、`259LUXU` 等变体，并改进候选排序规则。
  - JavDB 搜索默认附带 `locale=zh`，提升中文内容命中稳定性。
  - 新增技能 `.codex/skills/av-scraper-optimization`，沉淀 mdcx→Go 架构映射与编号策略实践。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_av_framework.go`
  - `internal/services/scraper_test.go`
  - `.codex/skills/av-scraper-optimization/SKILL.md`
  - `.codex/skills/av-scraper-optimization/references/architecture-map.md`
  - `.codex/skills/av-scraper-optimization/references/number-normalization.md`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestExtractAVCodeVariants'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 09:39] 自动刮削中文一致性修复计划（电影/剧集）
- Type: `plan`
- Summary:
  - 统一上传后自动刮削的 TMDB 语言策略为 `zh-CN` 优先，并在字段缺失时回退无语言详情补全。
  - 保持演员自动关联链路不变，仅修复自动刮削元数据语言不一致问题。
  - 通过服务层单测覆盖 `ScrapeMovieUpload` 与 `ScrapeEpisodeUpload` 的语言参数与回退行为。
- Changed Files:
  - `plan.md`
- Verification:
  - 计划项，无需构建/测试。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 09:39] 实现自动刮削 zh-CN 优先与回退补全（含单测）
- Type: `implementation`
- Summary:
  - `ScrapeMovieUpload` 与 `ScrapeEpisodeUpload` 改为通过 `getTMDBJSON(..., "zh-CN")` 请求搜索与详情。
  - 新增自动刮削详情回退流程：当中文字段缺失时，自动请求无语言详情并合并补全。
  - 新增季/集级别合并逻辑，补全 `season/episode` 的标题、简介、日期等空字段。
  - 新增两条服务层单测，验证自动电影/剧集刮削的语言参数与回退行为。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `go test ./internal/services -run 'TestScrape(Movie|Episode)UploadUsesChineseLanguageAndFallback'` passed.
  - `go test ./internal/services` passed.
  - `go test ./...` passed.
- Rollback:
  - `git revert <commit>`

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

### [2026-04-16 13:52] 刮削管理补全详情展示与中文化返回，并增加加载/报错反馈
- Type: `implementation`
- Summary:
  - 后端 `ScraperService` 预览与确认流程接入 `language=zh-CN` 请求参数，优先返回中文字段。
  - 新增“中文缺字段时英文兜底”的合并逻辑，覆盖标题/简介/日期/类型/题材及部分嵌套字段，避免空白展示。
  - 新增 `scraper` 单测，验证预览接口会携带中文语言参数，并在中文缺失时执行英文兜底。
  - 重构 `admin-web` 刮削管理页：从仅概览升级为“候选列表 + 结构化详情 + 原始 metadata JSON 折叠查看”。
  - 前端新增查询与保存 `loading` 状态，以及失败场景的明确错误提示（优先显示后端返回消息）。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `admin-web/src/views/ScrapePreview.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreview(Movie|TV)' -count=1` passed.
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -count=1` passed.
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-16 18:52] 修复 dev-down 未能关闭 server/worker/frontend 残留进程
- Type: `implementation`
- Summary:
  - 定位根因：`dev-down` 对 PID 文件命令串做严格匹配，`go run` 启动后命令会漂移为临时 `main -mode ...`，导致误判并跳过停止。
  - 强化 `dev-down` 停止逻辑：新增进程树终止（先 `TERM` 后超时 `KILL`）、服务命令漂移容错匹配、无 PID 文件时的兜底清理（按模式/端口/worker 日志 PID）。
  - 同步增强 `dev-up` 启动失败清理逻辑，复用相同的进程树停止和命令漂移容错，避免半启动残留。
  - 修复脚本兼容性：去除 `mapfile`，改为 `bash 3.2` 兼容写法（macOS 默认 bash 可运行）。
- Changed Files:
  - `scripts/dev-down.sh`
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh scripts/dev-down.sh` passed.
  - 执行 `bash scripts/dev-down.sh` 后，`lsof -nP -iTCP:8080 -sTCP:LISTEN` 无输出。
  - 执行 `bash scripts/dev-down.sh` 后，`lsof -nP -iTCP:5173 -sTCP:LISTEN` 无输出。
  - 通过模拟命令漂移场景（PID 文件为 `go run ...`，进程命令变化）执行 `dev-down`，进程可被清理。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 19:45] 实现演员管理与上传/刮削演员关联（管理端全中文）
- Type: `implementation`
- Summary:
  - 新增数据库迁移 `0007_actor_support`，引入 `actors` 与 `video_actors`，支持演员主数据与视频多演员关联。
  - 后端新增演员仓储能力：演员列表/创建/更新、按姓名自动入库、视频演员替换与追加关联。
  - 扩展管理端视频详情：返回 `actors` 列表；视频更新接口支持 `actor_ids` + `actor_names`（可同时传已有演员与新建演员）。
  - 新增管理端演员接口：`GET /api/v1/admin/actors`、`POST /api/v1/admin/actors`、`PUT /api/v1/admin/actors/:id`。
  - 上传链路支持演员关联：`upload/init` 会话保存演员信息，`upload/complete` 落库并关联演员；普通上传同样支持 `actor_ids/actor_names`。
  - 刮削链路接入 TMDB cast：电影/分集刮削与确认后会自动补充演员关联（非阻断式，失败不影响主流程）。
  - 管理端前端新增“演员管理”页面，并在上传页、视频详情页接入演员搜索多选+可创建；本次新增/修改文案统一中文。
- Changed Files:
  - `migrations/0007_actor_support.up.sql`
  - `migrations/0007_actor_support.down.sql`
  - `internal/models/admin.go`
  - `internal/repository/actor_repository.go`
  - `internal/repository/actor_repository_test.go`
  - `internal/repository/admin_repository.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/handlers/upload_helpers_test.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/router/index.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/views/ActorManage.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestExtractCastNames|^$' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
  - 说明：`internal/services` 全量测试在当前沙箱环境存在 `httptest` 端口监听限制，故本次使用定向用例与编译校验。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 20:19] 新增演员支持按姓名刮削（TMDB + JavDB）并增强重名处理
- Type: `implementation`
- Summary:
  - 新增管理端演员刮削预览接口 `POST /api/v1/admin/actors/scrape/preview`，支持 `tmdb/javdb` 双来源按姓名查询。
  - `ScraperService` 新增演员刮削能力：TMDB 人物检索（`zh-CN` 优先 + 英文兜底）与 JavDB HTML 候选解析；统一返回候选结构。
  - 新增 AV 刮削配置项：`AV_SCRAPER_BASE_URL`、`AV_SCRAPER_TIMEOUT_SECONDS`、`AV_SCRAPER_USER_AGENT`，并在 server/worker 启动时注入。
  - 演员创建重名时返回已存在演员信息（`existing_actor_id / existing_actor / existing_actor_name`），用于前端直接跳转编辑。
  - 管理端“演员管理”弹窗新增“按姓名刮削”交互：来源切换、候选列表、loading、中文错误提示、一键回填字段；重名创建支持提示并跳转编辑。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_actor.go`
  - `internal/services/scraper_test.go`
  - `internal/handlers/admin_actor_scrape.go`
  - `internal/handlers/admin_actor_scrape_test.go`
  - `internal/handlers/router.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/swagger_models.go`
  - `internal/repository/actor_repository.go`
  - `internal/config/config.go`
  - `main.go`
  - `.env.example`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/api/request.js`
  - `admin-web/src/views/ActorManage.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreviewActorByName(TMDB|JavDB)|TestPreview(Movie|TV)UsesChineseLanguageAnd(EnglishFallback|Fallback)|TestExtractCastNames' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestAdminActorScrapePreview' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository -run 'TestNormalizeActorName' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 21:20] 修复演员新增误报“同名已存在”
- Type: `implementation`
- Summary:
  - 定位误报根因：前端把 `code=1025` 一律视为重名；后端创建演员的非重名错误也会返回 `1025`，导致误判。
  - 后端修复：`AdminCreateActor` 仅在唯一约束冲突时返回 `1025`，并附带 `reason=duplicate_name`；其他创建失败改为 `1028`。
  - 前端修复：仅在明确重名（`reason=duplicate_name` 或 `code=1025` 且消息为“演员名称已存在”）时才弹“去编辑已有演员”提示。
- Changed Files:
  - `internal/handlers/admin.go`
  - `admin-web/src/views/ActorManage.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestAdminActorScrapePreview' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 21:43] 修复 dev-up 仅执行首条迁移导致 actors 表缺失
- Type: `implementation`
- Summary:
  - 线上报错 `create actor: relation "actors" does not exist (SQLSTATE 42P01)`，定位为数据库仅执行了 `0001_init.up.sql`。
  - 根因是 `scripts/dev-up.sh` 在迁移遍历中依赖 `find ... -print0 | sort -z`，导致在当前环境迁移流异常，出现只处理首条文件的问题。
  - 将迁移遍历改为 Bash 通配循环 `for file in "$ROOT_DIR"/migrations/*.up.sql`，按文件名顺序稳定执行并兼容当前环境。
  - 已补齐当前数据库缺失迁移 `0002~0007`，确认 `actors` 与 `video_actors` 表存在。
- Changed Files:
  - `scripts/dev-up.sh`
  - `plan.md`
- Verification:
  - `bash -n scripts/dev-up.sh` passed。
  - `bash scripts/dev-up.sh` 日志可见按顺序检查 `0001~0007` 迁移。
  - `SELECT version FROM schema_migrations ORDER BY version;` 返回 `0001~0007` 全部版本。
  - `SELECT to_regclass('public.actors'), to_regclass('public.video_actors');` 返回 `actors|video_actors`。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 22:09] 修复 TMDB 演员刮削 notes 被截断与乱码
- Type: `implementation`
- Summary:
  - 定位到演员刮削返回中 `notes` 不完整的根因：`previewActorsTMDB` 对 `biography` 执行了 `500` 字节硬截断。
  - 该截断会导致中文多字节字符被切断，出现 `\uFFFD` 替代字符，并且内容不完整。
  - 移除该截断逻辑，`notes` 直接返回完整 `biography`。
  - 新增回归测试：长中文简介场景下 `notes` 不被截断且不包含 `\uFFFD`。
- Changed Files:
  - `internal/services/scraper_actor.go`
  - `internal/services/scraper_test.go`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestPreviewActorByNameTMDBNotesNotTruncated|TestPreviewActorByName(TMDB|JavDB)' -count=1` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-16 23:01] 上传中心支持“非电影批量上传”，电影保持单文件
- Type: `implementation`
- Summary:
  - 管理端上传页改造为队列式上传：`short/episode` 支持批量选择并串行分片上传，`movie` 继续仅支持单文件。
  - 类型切换为电影时，若已选择多个文件会自动仅保留第 1 个并提示，避免误提交。
  - 批量流程支持“失败不中断”：单文件失败后继续后续文件，最终展示成功/失败/取消汇总与逐文件结果。
  - 增加批次进度与当前文件进度展示，取消上传时终止当前会话并标记后续文件为未上传（已取消）。
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed.
- Rollback:
  - `git revert <commit>`

### [2026-04-17 00:39] 新增短视频合集功能（上传可选多选、后续可修改、删除仅解关联）
- Type: `implementation`
- Summary:
  - 新增数据库迁移 `0008_collections_support`，引入 `collections` 与 `video_collections`，支持视频与合集多对多关联。
  - 新增合集仓储能力：合集列表/创建/更新/删除、名称规范化唯一约束、视频合集关联替换与按视频批量查询。
  - 约束合集仅适用于 `short`：上传（普通/分片）与管理端视频编辑均支持 `collection_ids`，并在非短视频场景返回明确错误。
  - 管理端新增合集管理页面与路由（增删改查）；删除合集时仅删除合集并自动解除关联，不删除视频，并返回解除关联数量。
  - 管理端上传页新增“所属合集（可选、多选，仅短视频）”；视频详情页支持后续修改合集并保存。
  - 应用侧视频返回补充合集信息：详情、推荐与列表搜索结果可携带 `collections` 字段。
  - 补充合集相关单元测试：名称规范化、ID 去重、上传参数解析与类型约束校验。
- Changed Files:
  - `migrations/0008_collections_support.up.sql`
  - `migrations/0008_collections_support.down.sql`
  - `internal/models/models.go`
  - `internal/models/app.go`
  - `internal/models/admin.go`
  - `internal/repository/collection_repository.go`
  - `internal/repository/collection_repository_test.go`
  - `internal/repository/video_repository.go`
  - `internal/repository/app_repository.go`
  - `internal/repository/admin_repository.go`
  - `internal/services/upload.go`
  - `internal/services/chunk_upload.go`
  - `internal/services/chunk_upload_test.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/handlers/admin.go`
  - `internal/handlers/router.go`
  - `internal/handlers/upload_collection_test.go`
  - `admin-web/src/api/admin.js`
  - `admin-web/src/router/index.js`
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/views/CollectionManage.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` failed（当前沙箱环境禁止 `httptest` 监听端口，`internal/handlers` 与 `internal/services` 的部分用例受限）。
  - `GOCACHE=$(pwd)/.gocache go test ./... -run '^$'` passed（全量编译校验通过）。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/repository ./internal/handlers ./internal/services -run 'TestNormalizeCollectionName|TestDedupeCollectionIDs|TestParseUploadCollectionIDs|TestCollectionTypeValidation|TestChunkUploadCompleteFileSurvivesAbort' -count=1` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 00:49] 修正上传页合集远程搜索的选项保留
- Type: `implementation`
- Summary:
  - 优化管理端上传页 `searchCollections`：改为合并选项而非覆盖，避免远程搜索后已选合集标签丢失显示。
- Changed Files:
  - `admin-web/src/views/VideoUpload.vue`
  - `plan.md`
- Verification:
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 11:54] 新增 AV 视频分类与自动 AV 刮削（含演员自动关联）
- Type: `implementation`
- Summary:
  - 新增视频类型 `av`（管理员可上传），上传后与电影/剧集一样进入 `scraping` 流程，由 worker 自动执行 AV 刮削。
  - 扩展队列任务：新增 `video:scrape:av`，并在上传与分片上传完成后按类型路由到 AV 刮削任务。
  - 新增 AV 刮削能力（JavDB）：支持“番号优先、标题兜底”检索，抓取标题/简介/封面/发行日期/演员，写入视频元数据并自动关联演员（`scrape_av`）。
  - 扩展管理端刮削接口：`/admin/scrape/preview` 支持 `type=av`，`/admin/scrape/confirm` 对 AV 支持 `external_id` 确认；电影/剧集仍使用 `tmdb_id`。
  - 新增数据库迁移 `0009_av_type_support`，将 `videos.type` 约束扩展为 `short/movie/episode/av`。
  - 管理端界面新增 AV：上传类型、视频列表筛选、刮削页类型选择与 AV 候选保存；仪表盘新增 AV 数量统计卡片。
  - 后端统计扩展：`AdminStats` 新增 `av_videos`，`CountVideosByType` 统计新增 `av`。
- Changed Files:
  - `internal/services/scraper.go`
  - `internal/services/scraper_test.go`
  - `internal/queue/tasks.go`
  - `internal/queue/scrape_tasks.go`
  - `internal/handlers/upload.go`
  - `internal/handlers/upload_chunk.go`
  - `internal/services/upload.go`
  - `internal/handlers/admin_scrape.go`
  - `internal/handlers/admin.go`
  - `internal/models/admin.go`
  - `internal/repository/admin_repository.go`
  - `internal/handlers/swagger_models.go`
  - `internal/handlers/upload_collection_test.go`
  - `migrations/0009_av_type_support.up.sql`
  - `migrations/0009_av_type_support.down.sql`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/Dashboard.vue`
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./internal/services -run 'TestScrapeAVUploadCodeFirstAndActorSync|TestPreviewAVFallbackByTitle|TestScrape(Movie|Episode)UploadUsesChineseLanguageAndFallback' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./internal/handlers -run 'TestCollectionTypeValidation|TestParseUploadCollectionIDs|TestParseUploadTags_JSON' -count=1` passed。
  - `GOCACHE=$(pwd)/.gocache go test ./... -run '^$'` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

### [2026-04-17 11:58] AV 功能补充全量验证
- Type: `implementation`
- Summary:
  - 在完成 AV 分类与自动刮削改造后，补充执行后端全量单测回归，确认不影响既有模块。
- Changed Files:
  - `plan.md`
- Verification:
  - `GOCACHE=$(pwd)/.gocache go test ./...` passed。
  - `npm --prefix admin-web run build` passed。
- Rollback:
  - `git revert <commit>`

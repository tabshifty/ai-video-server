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

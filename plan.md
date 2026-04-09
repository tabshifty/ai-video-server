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

# 欧美 AV oshash 与刮削确认门控 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为欧美 AV 上传链路增加 `os_hash`、`av_scrape_pending` 门控、弃刮分流与 ThePornDB hash 优先刮削，同时让管理端把待确认门控页和已 ready 的手动重刮页分开。

**Architecture:** 数据层新增 `videos.os_hash` 和 `av_scrape_pending`，上传与队列处理器负责在欧美 AV 上传时同步计算与持久化 oshash。刮削服务在自动上传、手动重刮和 retag 三条路径复用同一 hash 候选逻辑，但 `av_scrape_pending` 只走确认/弃刮，`ready` 只走手动刮削，不共享状态语义。ThePornDB 侧只做局部修复，不扩展到其它站点。

**Tech Stack:** Go、PostgreSQL、Asynq、Vue 3、Element Plus、Vite、Kotlin/Android 仅回归验证不改动

---

## File Structure

- `migrations/0021_western_av_oshash_gate.up.sql`
- `migrations/0021_western_av_oshash_gate.down.sql`
  - 负责 `os_hash` 列、索引和 `videos.status` 约束。
- `internal/models/video.go`
- `internal/repository/video_repository.go`
  - 负责 `Video.OSHash` 持久化读写。
- `internal/services/upload.go`
- `internal/services/chunk_upload.go`
- `internal/handlers/upload.go`
- `internal/handlers/upload_chunk.go`
  - 负责上传与分片上传的 `site_category` 透传和 oshash 落库。
- `internal/services/scraper_av_mdcx_detail_sites.go`
- `internal/services/scraper.go`
- `internal/queue/scrape_tasks.go`
- `internal/handlers/admin_scrape.go`
- `internal/handlers/router.go`
  - 负责 hash 候选、ThePornDB 修复、门控状态分流、skip/confirm 端点。
- `admin-web/src/views/VideoUpload.vue`
- `admin-web/src/views/VideoList.vue`
- `admin-web/src/views/ScrapePreview.vue`
- `admin-web/src/views/AVManualScrape.vue`
- `admin-web/src/views/videoList.helpers.js`
- `admin-web/src/views/scrapePreview.helpers.js`
- `admin-web/src/api/admin.js`
  - 负责上传表单、状态筛选、门控页与手动刮削页的分离。
- `pkg/oshash/oshash.go`
- `pkg/oshash/oshash_test.go`
  - 负责 OpenSubtitles 风格 oshash 计算。
- `plan.md`
  - 记录实施过程与验证结果。

## Task 1: 数据层与迁移

**Files:**
- Create: `migrations/0021_western_av_oshash_gate.up.sql`
- Create: `migrations/0021_western_av_oshash_gate.down.sql`
- Modify: `internal/models/video.go`
- Modify: `internal/repository/video_repository.go`
- Test: `internal/repository/*_test.go`

- [ ] **Step 1: 写迁移测试或先跑 schema 失败检查**

```bash
bash scripts/dev-down.sh
bash scripts/dev-up.sh
```

Expected: 现在还没有 `os_hash` 列和 `av_scrape_pending` 约束，后续迁移测试需要覆盖它们。

- [ ] **Step 2: 实现 migration 和 repository 读写**

```sql
ALTER TABLE videos ADD COLUMN os_hash CHAR(16);
CREATE INDEX IF NOT EXISTS idx_videos_os_hash ON videos(os_hash) WHERE os_hash IS NOT NULL;
ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
  CHECK (status IN ('uploaded','scraping','tv_pending','av_scrape_pending','processing','ready','failed'));
```

```go
func (r *VideoRepository) UpdateVideoOSHash(ctx context.Context, videoID uuid.UUID, hash string) error
func (r *VideoRepository) GetVideoOSHash(ctx context.Context, videoID uuid.UUID) (string, error)
```

- [ ] **Step 3: 验证迁移幂等与 repo 单测**

```bash
go test ./internal/repository -run 'OSHash|Video' -v
bash scripts/dev-down.sh && bash scripts/dev-up.sh
```

- [ ] **Step 4: 提交**

```bash
git add migrations/0021_western_av_oshash_gate.* internal/models/video.go internal/repository/video_repository.go
git commit -m "增加欧美AV oshash 数据层"
```

## Task 2: oshash 计算与上传写入

**Files:**
- Create: `pkg/oshash/oshash.go`
- Create: `pkg/oshash/oshash_test.go`
- Modify: `internal/services/upload.go`
- Modify: `internal/services/chunk_upload.go`
- Modify: `internal/handlers/upload.go`
- Modify: `internal/handlers/upload_chunk.go`
- Test: `internal/services/*_test.go`

- [ ] **Step 1: 写 oshash 失败测试**

```go
func TestCompute_ReturnsErrFileTooSmall(t *testing.T)
func TestCompute_ReturnsDeterministicHex(t *testing.T)
```

- [ ] **Step 2: 运行测试确认当前失败**

```bash
go test ./pkg/oshash -v
```

- [ ] **Step 3: 实现最小代码**

```go
if typ == "av" && siteCategory == "western" {
    if h, err := oshash.Compute(in.FilePath); err == nil {
        result.OSHash = h
    } else {
        s.logger.Warn("oshash compute failed", "video_id", videoID, "error", err)
    }
}
```

上传与分片上传都要把 `site_category` 透传到 `SaveUpload` / `SaveUploadedFile`。

- [ ] **Step 4: 验证上传路径**

```bash
go test ./internal/services -run 'SaveUpload|SaveUploadedFile|OSHash' -v
```

- [ ] **Step 5: 提交**

```bash
git add pkg/oshash internal/services/upload.go internal/services/chunk_upload.go internal/handlers/upload.go internal/handlers/upload_chunk.go
git commit -m "实现欧美AV oshash 计算与落库"
```

## Task 3: ThePornDB 修复与 hash 候选复用

**Files:**
- Modify: `internal/services/scraper_av_mdcx_detail_sites.go`
- Modify: `internal/services/scraper.go`
- Modify: `internal/queue/scrape_tasks.go`
- Test: `internal/services/*theporndb*_test.go`
- Test: `internal/queue/scrape_tasks_test.go`

- [ ] **Step 1: 写失败测试**

覆盖：
- `/movies` 走 `q=`
- 429 重试、401 不重试
- `thePornDBBestMatchURL` 低于 0.4 返回空
- `datedNumberPattern` 的 P1/P2/P3 三种文件名
- `collectThePornDBHashCandidate` 命中时首位候选带 `match_source=oshash`

- [ ] **Step 2: 运行测试确认红灯**

```bash
go test ./internal/services -run 'ThePornDB|AVScrape|OSHash' -v
```

- [ ] **Step 3: 实现最小代码**

```go
func thePornDBSearchURL(baseURL, category, keyword string) string
func (s *ScraperService) fetchThePornDBJSON(...)
func (s *ScraperService) collectThePornDBHashCandidate(...)
```

hash 候选在自动刮削、手动刮削、retag 三路复用，但只提供候选，不直接自动确认。

- [ ] **Step 4: 验证回归**

```bash
go test ./internal/services -run 'ThePornDB|AVScrape|ConfirmAV|OSHash' -v
go test ./internal/queue -run 'ScrapeAV|ScrapeRetag' -v
```

- [ ] **Step 5: 提交**

```bash
git add internal/services/scraper_av_mdcx_detail_sites.go internal/services/scraper.go internal/queue/scrape_tasks.go
git commit -m "修复欧美AV刮削与ThePornDB hash逻辑"
```

## Task 4: 门控状态机与 admin-web 分流

**Files:**
- Modify: `internal/handlers/admin_scrape.go`
- Modify: `internal/handlers/router.go`
- Modify: `admin-web/src/views/VideoUpload.vue`
- Modify: `admin-web/src/views/VideoList.vue`
- Modify: `admin-web/src/views/ScrapePreview.vue`
- Modify: `admin-web/src/views/AVManualScrape.vue`
- Modify: `admin-web/src/views/videoList.helpers.js`
- Modify: `admin-web/src/views/scrapePreview.helpers.js`
- Modify: `admin-web/src/api/admin.js`
- Test: `admin-web/src/views/*.spec.js`

- [ ] **Step 1: 写失败测试/前端 helper 测试**

覆盖：
- `av_scrape_pending` 状态 meta
- `VideoList` 能筛到“欧美 AV 待确认”
- `ScrapePreview` 处理待确认门控
- `AVManualScrape` 只处理 ready 的手动重刮

- [ ] **Step 2: 运行测试确认失败**

```bash
cd admin-web && npm test
```

- [ ] **Step 3: 实现门控分流**

`ScrapePreview.vue` 只用于 `av_scrape_pending` 的确认/弃刮。  
`AVManualScrape.vue` 只用于已 `ready` 视频的手动重刮。  
`PUT /admin/scrape/skip` 只允许 `av_scrape_pending`，并转 `processing` 后入队转码。

- [ ] **Step 4: 验证 UI 与接口**

```bash
cd admin-web && npm run build
go test ./internal/handlers -run 'ScrapeSkip|Upload|ScrapeConfirm' -v
```

- [ ] **Step 5: 提交**

```bash
git add internal/handlers/admin_scrape.go internal/handlers/router.go admin-web/src/views/*.vue admin-web/src/views/*.js admin-web/src/api/admin.js
git commit -m "拆分欧美AV门控页与手动刮削页"
```

## Task 5: 收尾验证与文档

**Files:**
- Modify: `CONTEXT.md`
- Modify: `plan.md`
- Verify: task docs and Git history

- [ ] **Step 1: 跑定向验证**

```bash
go test ./pkg/oshash/...
go test ./internal/services -run 'ThePornDB|AVScrape|ConfirmAV|SkipScrape|OSHash' -v
go test ./internal/queue -run 'ScrapeAV|ScrapeRetag' -v
go test ./internal/handlers -run 'ScrapeSkip|Upload' -v
cd admin-web && npm test && npm run build
```

- [ ] **Step 2: 跑全量验证**

```bash
go test ./...
go vet ./...
```

- [ ] **Step 3: 回写计划与任务状态**

把验证结果追加到 `plan.md`，用户确认后再补 `tasks/2026-05-23-western-av-oshash-confirm-gate/DONE.md`。

- [ ] **Step 4: 提交收尾**

```bash
git add CONTEXT.md plan.md
git commit -m "完成欧美AV oshash 门控收尾"
```

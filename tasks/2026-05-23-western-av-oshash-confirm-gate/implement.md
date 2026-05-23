# Implement：欧美 AV oshash 与刮削确认门控

按阶段顺序推进。每个阶段结束前跑该阶段的最小验证，再进下一阶段；中途若与 PRD 冲突，回到 PRD 校准后再继续（不允许在 implement 阶段单方面偏离 PRD）。

---

## 阶段 0：基线确认

**目的**：固定本任务开始前的"绿色"基线，便于事后回退判定回归。

- 在干净的 master 上跑：
  - `go test ./...` → 记录通过基线
  - `go vet ./...` → 无 issue 基线
  - `cd admin-web && npm test` → 通过基线
- 启动一次 `bash scripts/dev-up.sh` 确认后端 + 前端 + DB + Redis 全绿
- 把当前 `git rev-parse HEAD` 记录到 implement 的 commit log 里作 baseline 锚点

---

## 阶段 1：数据层

### 1.1 Migration `0021_western_av_oshash_gate`

`migrations/0021_western_av_oshash_gate.up.sql`：

```sql
ALTER TABLE videos ADD COLUMN os_hash CHAR(16);

CREATE INDEX IF NOT EXISTS idx_videos_os_hash
    ON videos(os_hash)
    WHERE os_hash IS NOT NULL;

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','av_scrape_pending','processing','ready','failed'));
```

`migrations/0021_western_av_oshash_gate.down.sql`：

```sql
-- 先把残留的 av_scrape_pending 视频回退到 uploaded,避免新约束装回时违反
UPDATE videos SET status = 'uploaded' WHERE status = 'av_scrape_pending';

ALTER TABLE videos DROP CONSTRAINT IF EXISTS videos_status_check;
ALTER TABLE videos ADD CONSTRAINT videos_status_check
    CHECK (status IN ('uploaded','scraping','tv_pending','processing','ready','failed'));

DROP INDEX IF EXISTS idx_videos_os_hash;
ALTER TABLE videos DROP COLUMN IF EXISTS os_hash;
```

**验证**：
- `bash scripts/dev-down.sh && bash scripts/dev-up.sh` 跑通完整 up
- 手工 `\d videos` 检查 `os_hash` 列存在、index 存在、CHECK 约束包含 `av_scrape_pending`
- 跑 down → up 验证幂等

### 1.2 Repository 扩展

`internal/repository/video_repository.go`：

- `UpdateVideoOSHash(ctx, videoID uuid.UUID, hash string) error`：单列更新；hash 为空字符串时写 NULL
- `GetVideoOSHash(ctx, videoID uuid.UUID) (string, error)`：单值查询，nil 行返回 `("", nil)`
- 现有 `GetVideoByID` 已返回完整 row，把 `os_hash` 加入 `models.Video` 结构体（`internal/models/video.go` 或类似位置——找到现有 Video struct 定义后扩字段）

**验证**：repo 单测覆盖 set / get / 空值场景。

---

## 阶段 2：oshash 算法实现

### 2.1 `pkg/oshash/oshash.go`

```go
package oshash

import (
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "os"
)

const ChunkSize = 64 * 1024 // 64 KiB

var ErrFileTooSmall = errors.New("file too small for oshash (need >= 128 KiB)")

// Compute 返回 path 文件的 OpenSubtitles 风格 oshash，16 位小写 hex。
// 算法：file_size + sum(first 64KB as uint64 LE) + sum(last 64KB as uint64 LE)，mod 2^64。
func Compute(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    info, err := f.Stat()
    if err != nil {
        return "", err
    }
    size := info.Size()
    if size < 2*ChunkSize {
        return "", ErrFileTooSmall
    }

    hash := uint64(size)
    if err := accumulate(f, 0, &hash); err != nil {
        return "", err
    }
    if err := accumulate(f, size-ChunkSize, &hash); err != nil {
        return "", err
    }
    return fmt.Sprintf("%016x", hash), nil
}

func accumulate(f *os.File, offset int64, hash *uint64) error {
    if _, err := f.Seek(offset, io.SeekStart); err != nil {
        return err
    }
    buf := make([]byte, ChunkSize)
    if _, err := io.ReadFull(f, buf); err != nil {
        return err
    }
    for i := 0; i+8 <= len(buf); i += 8 {
        *hash += binary.LittleEndian.Uint64(buf[i : i+8])
    }
    return nil
}
```

### 2.2 `pkg/oshash/oshash_test.go`

- **Fixture 测试**：用一个固定 256KB 内容（例如全 0x42 填充）算出来跟 Python `oshash.oshash()` 跑一遍记录的值对齐（在 PR 描述里附 Python 重现脚本，确保后人能重算）
- 文件 < 128KB 返回 `ErrFileTooSmall`
- 文件不存在返回 `os.ErrNotExist`

### 2.3 接入 `UploadService.SaveUpload`

`internal/services/upload.go`：

- `UploadResult` 结构里加 `OSHash string` 字段（可空）
- 在 SHA256 算完落盘后、`InsertFileHash` 之前插入：

```go
// 仅欧美 AV 算 oshash
if typ == "av" && siteCategory == "western" {
    if h, err := oshash.Compute(in.FilePath); err == nil {
        result.OSHash = h
    } else {
        // 算失败不阻塞上传,落 NULL + log warn
        s.logger.Warn("oshash compute failed", "video_id", videoID, "error", err)
    }
}
```

- `SaveUpload` 入参增加 `siteCategory string`（透传给内部 `SaveUploadInput`）
- `InsertVideo` 调用点把 `os_hash` 一起塞进 INSERT（或 INSERT 完后单独 `UpdateVideoOSHash`，选实现简单的）

### 2.4 分片上传补 oshash

`internal/services/chunk_upload.go`：merge 完输出最终 file path 时，同样按 `type=av && siteCategory=western` 触发 oshash 算 + 落库。共用 `pkg/oshash.Compute`。

**验证**：
- 单测覆盖 `SaveUpload` 对欧美 AV 算 hash、非欧美档不算
- 集成测：手动上传一个欧美 AV 样本，DB 看到 `os_hash` 非 NULL

---

## 阶段 3：scrape 流水线核心改造

### 3.1 共享 helper `collectThePornDBHashCandidate`

`internal/services/scraper_av_mdcx_detail_sites.go`（或新文件 `scraper_av_theporndb_hash.go`）：

```go
// collectThePornDBHashCandidate 用 oshash 命中 ThePornDB /scenes/hash/{hash}。
// 命中返回 1 个候选，404 / 网络错误返回 (nil, nil) 不抛——hash 路径是优化不是必需。
func (s *ScraperService) collectThePornDBHashCandidate(
    ctx context.Context,
    osHash string,
) ([]avScrapeCandidate, error) {
    osHash = strings.TrimSpace(osHash)
    if osHash == "" {
        return nil, nil
    }
    token := strings.TrimSpace(s.avThePornDBAPIToken)
    if token == "" {
        return nil, nil // 没 token 也不抛,跟 hash miss 等价
    }
    base := strings.TrimRight(s.avSiteBaseURL("theporndb", "https://api.theporndb.net"), "/")
    url := fmt.Sprintf("%s/scenes/hash/%s", base, osHash)

    var payload thePornDBItemResponse
    err := s.fetchThePornDBJSON(ctx, url, map[string]string{
        "Authorization": "Bearer " + token,
        "Accept":        "application/json",
    }, &payload)
    if err != nil {
        // 401 / 404 / 网络错误都吞掉,只记 log
        s.logger.Info("theporndb hash lookup miss", "hash", osHash, "error", err)
        return nil, nil
    }
    if payload.Data.Slug == "" {
        return nil, nil
    }
    candidate := thePornDBMetadataFromItem(payload.Data, "scenes")
    if candidate.Raw == nil {
        candidate.Raw = map[string]any{}
    }
    candidate.Raw["match_source"] = "oshash"
    return []avScrapeCandidate{candidate}, nil
}
```

### 3.2 `HandleScrapeAV` 西方分支改造

`internal/queue/scrape_tasks.go`：

- `handleScrape` 在 `expectedType == "av"` 时，先读 `video.Metadata.SiteCategory`（或调 `normalizeAVSiteCategory(asString(video.Metadata["site_category"]))`）
- 若 `siteCategory == "western"`，调用新方法 `p.scrape.ScrapeAVWesternUpload(ctx, videoID, payload.FilePath, payload.Filename)`，行为：
  1. 读 `video.os_hash`，非空 → 调 `collectThePornDBHashCandidate`
  2. 跑既有 keyword 搜索（scenes + movies 两路），收集候选
  3. 去重 + 排序 + 截断到 6 条
  4. 写入 `videos.metadata.scrape_preview`、`scrape_attempt`
  5. `UpdateVideoStatus(ctx, videoID, "av_scrape_pending")`
  6. **不**入队 transcode、**不**调 ConfirmAV
- 否则（非欧美档）走现行 `ScrapeAVUpload` 路径

`scrape_tasks.go:208-213` 的 `EnqueueTranscode` 这段保留，但在 western 分支提前 return 前不触发。

### 3.3 `ConfirmAV` 按状态决定是否入队 transcode

`internal/services/scraper.go:674` 附近的 `ConfirmAV`：

- 函数入口先读 `video, _ := s.repo.GetVideoByID(ctx, in.VideoID)` 拿当前状态
- 写完 metadata / actors / poster 之后判断：
  ```go
  if video.Status == "av_scrape_pending" {
      if err := s.repo.UpdateVideoStatus(ctx, in.VideoID, "processing"); err != nil {
          return err
      }
      return s.enqueuer.EnqueueTranscode(queue.TranscodePayload{
          VideoID:      in.VideoID.String(),
          InputPath:    video.OriginalPath,
          TargetFormat: "mp4",
      })
  }
  // 其它来源（手动刮削已 ready 视频等）保持现有行为
  return s.repo.UpdateVideoStatus(ctx, in.VideoID, "ready")
  ```
- 注意 `ScraperService` 是否已经持有 `enqueuer` 依赖——若没有需要在 service 构造器加注入。查 `main.go` 的 `configureAVScraper` 看现有依赖图。

### 3.4 `HandleScrapeRetag` 补算 oshash

`internal/queue/scrape_tasks.go:95` 的 `HandleScrapeRetag`：

- 在 `targetType == "av"` 分支里，retag 完成后判断目标 video 的 `site_category`：
  ```go
  if normalizedCategory == "western" && video.OSHash == "" {
      if _, err := os.Stat(video.OriginalPath); err == nil {
          if hash, hashErr := oshash.Compute(video.OriginalPath); hashErr == nil {
              _ = p.repo.UpdateVideoOSHash(ctx, videoID, hash)
          }
      }
  }
  ```
- 原文件不存在 / hash 算失败都不报错，落 NULL 即可

### 3.5 `PreviewAVSearch` 注入 hash 路径

`internal/services/scraper.go:452` 的 `PreviewAVSearch`：

- `AVPreviewOptions` 加字段 `VideoID uuid.UUID`（可选，零值表示纯关键字模式）
- 函数内部：若 `opts.VideoID` 非零，去 DB 读 `os_hash`，非空时先调 `collectThePornDBHashCandidate` 拿候选，附加到返回列表首位（带 `match_source=oshash`）
- 现有调用方（不传 VideoID 的）行为完全不变

**验证**：
- 单测覆盖三条入口（auto / manual / retag）都能拿到 hash 候选
- 集成测：注入一个 mock `os_hash`，自动刮削后看到候选首位 `match_source=oshash`

---

## 阶段 4：ThePornDB 五项修复（次级目标）

### 4.1 `/movies` 端点参数修复

`internal/services/scraper_av_mdcx_detail_sites.go:1683-1687` 的 `thePornDBSearchURL`：

```go
func thePornDBSearchURL(baseURL, category, keyword string) string {
    values := url.Values{}
    switch category {
    case "movies":
        values.Set("q", keyword)
    default: // "scenes"
        values.Set("parse", keyword)
    }
    values.Set("per_page", "100")
    return strings.TrimRight(baseURL, "/") + "/" + category + "?" + values.Encode()
}
```

附 unit test：`TestThePornDBSearchURL_MoviesUsesQ` / `TestThePornDBSearchURL_ScenesUsesParse`。

### 4.2 `OUMEI_NAME` 字典移植

阶段 9（独立工具）一次性生成 `internal/services/theporndb_oumei_name.go`，里面是：

```go
package services

// 由 cmd/gen-theporndb-oumei 从 references/mdcx/mdcx/manual.py 生成,请勿手工编辑。
var thePornDBOumeiName = map[string]string{
    "wgp":   "WhenGirlsPlay",
    "18og":  "18OnlyGirls",
    // ... 200+ 条
}
```

修改 `thePornDBLongName` 实现：

```go
func thePornDBLongName(shortName string) string {
    if v, ok := thePornDBOumeiName[shortName]; ok {
        return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(v), "-", ""), ".", "")
    }
    return shortName
}
```

注意：原版只有一条 `clubseventeen → clubsweethearts` 必须被新字典覆盖（MDCX 的 `OUMEI_NAME` 应该包含；如果没有则手工保留这一条 fallback）。

### 4.3 401 / 429 / 5xx 重试包装

`internal/services/scraper_av_mdcx_detail_sites.go`：

```go
func (s *ScraperService) fetchThePornDBJSON(
    ctx context.Context,
    endpoint string,
    headers map[string]string,
    out any,
) error {
    backoffs := []time.Duration{200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}
    var lastErr error
    for attempt := 0; attempt <= len(backoffs); attempt++ {
        err := s.fetchAVJSON(ctx, endpoint, headers, out)
        if err == nil {
            return nil
        }
        lastErr = err
        if !isRetryableTPDBError(err) {
            return err
        }
        if attempt < len(backoffs) {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(backoffs[attempt]):
            }
        }
    }
    return lastErr
}

func isRetryableTPDBError(err error) bool {
    // fetchAVJSON 的错误形态是 "AV JSON 请求失败，状态码=XXX，响应=..."
    // 解析出状态码,429 / 5xx 重试,401 / 404 不重试
    msg := err.Error()
    switch {
    case strings.Contains(msg, "状态码=429"):
        return true
    case strings.Contains(msg, "状态码=5"): // 500-599
        return true
    case errors.Is(err, context.DeadlineExceeded):
        return true
    case strings.Contains(msg, "请求失败") && !strings.Contains(msg, "状态码="):
        return true // 网络错误,没有状态码
    }
    return false
}
```

- 将 `thePornDBAVCrawler` 内**所有**对 `s.svc.fetchAVJSON` 的调用改成 `s.svc.fetchThePornDBJSON`（搜 `scraper_av_mdcx_detail_sites.go:437/586/622` 三处）
- 其它站点 crawler 继续直接用 `fetchAVJSON`，零影响

### 4.4 相似度阈值

`scraper_av_mdcx_detail_sites.go:1690` 的 `thePornDBBestMatchURL`：

- 新增包级常量 `const thePornDBMinSimilarity = 0.4`
- 在 `for _, candidates := range [][]candidate{...}` 末尾 sort 之后取 `top := candidates[0]`，若 `top.score < thePornDBMinSimilarity` 则该桶视同没匹配，continue 到下一桶
- 三桶都低于阈值最终返回空

附 unit test：`TestThePornDBBestMatchURL_BelowThreshold_ReturnsEmpty`。

### 4.5 关键字正则扩展

`scraper_av_mdcx_detail_sites.go:46` 附近的 `datedNumberPattern`：

```go
var (
    datedNumberPatternP1 = regexp.MustCompile(`(?i)(([A-Z0-9.-]{2,})[-_. ]2?0?(\d{2}[-.]\d{2}[-.]\d{2}))`)
    datedNumberPatternP2 = regexp.MustCompile(`(?i)(([A-Z0-9]{3,})[-_. ]+(\d{4}-\d{2}-\d{2}))`)
    datedNumberPatternP3 = regexp.MustCompile(`(?i)\[(([A-Z0-9.-]{2,}))\][\s_.-]+(\d{2}[-.]\d{2}[-.]\d{2})`)

    datedNumberPatterns = []*regexp.Regexp{datedNumberPatternP1, datedNumberPatternP2, datedNumberPatternP3}
)
```

`thePornDBSearchKeywords` 改成按优先级试三个正则，第一个命中即用其 series + date 输出，三个都不命中走原 fallback 路径。

附 unit test：覆盖三种文件名命中三个正则。

**阶段验证**：
- 全部 5 个修复对应单测通过
- 跑 `go test ./internal/services/... -run ThePornDB` 全绿
- 现有 ThePornDB 相关回归测试（`scraper_av_mdcx_sites_test.go` 等）不破

---

## 阶段 5：admin 端点

### 5.1 `PUT /admin/scrape/skip` 新增

`internal/handlers/admin_scrape.go`：

```go
type ScrapeSkipRequest struct {
    VideoID string `json:"video_id"`
    Reason  string `json:"reason,omitempty"`
}

func (a *API) PutAdminScrapeSkip(c *gin.Context) {
    var req ScrapeSkipRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, 1, err.Error())
        return
    }
    videoID, err := uuid.Parse(strings.TrimSpace(req.VideoID))
    if err != nil {
        response.Error(c, 1, "invalid video_id")
        return
    }
    if err := a.scrapeSvc.SkipScrape(c.Request.Context(), videoID, req.Reason); err != nil {
        response.Error(c, 2, err.Error())
        return
    }
    response.OK(c, gin.H{"video_id": videoID})
}
```

`internal/services/scraper.go` 新增 `SkipScrape(ctx, videoID, reason)`：
- 读 video, 校验 `status == "av_scrape_pending"`，否则报错 "video not in scrape pending state"
- `MergeVideoMetadata` 写入 `scrape_skipped=true` / `scrape_skip_reason` / `scrape_skipped_at`
- `UpdateVideoStatus(ctx, videoID, "processing")`
- `EnqueueTranscode(...)` 入队

`internal/handlers/router.go`：注册 `admin.PUT("/scrape/skip", a.PutAdminScrapeSkip)`，复用现有 admin 中间件。

### 5.2 `/upload` 与 `/upload-chunk` 接收 `site_category`

`internal/handlers/upload.go` 与 `internal/handlers/upload_chunk.go`：

- `c.PostForm("site_category")` 读字段，`type=av` 时允许为空；为空时按 `japanese` 默认，非空时必须能被 `normalizeAVSiteCategory(...)` 识别
- 传给 `UploadService.SaveUpload`
- 上传响应 payload 不增加新字段，体感不变

**阶段验证**：
- `curl -X PUT /admin/scrape/skip` 在三个状态下（av_scrape_pending / ready / scraping）行为符合预期
- 上传带 `site_category=western` 时 DB 看到 metadata 写入；不带 site_category 时 av 落 `japanese` 默认

---

## 阶段 6：admin-web 主要 UI（VideoUpload / VideoList / ScrapePreview）

### 6.1 `admin-web/src/views/VideoUpload.vue`

- `form` 加字段 `site_category: 'japanese'`
- 在 `type` select 下面新增条件 radio：

```vue
<el-form-item v-if="form.type === 'av'" label="AV 地区">
  <el-radio-group v-model="form.site_category">
    <el-radio-button value="japanese">日本</el-radio-button>
    <el-radio-button value="western">欧美</el-radio-button>
    <el-radio-button value="fc2">FC2</el-radio-button>
  </el-radio-group>
</el-form-item>
```

- `submit` 把 `form.site_category` 加入 FormData（仅 type=av 时）
- 分片上传也透传

### 6.2 `admin-web/src/views/VideoList.vue`

- 状态筛选 `statusOptions` 数组加 `{ value: 'av_scrape_pending', label: '欧美 AV 待确认' }`
- `getVideoStatusMeta`（共享 utils，找到现有定义文件后扩展）加：
  ```js
  av_scrape_pending: { label: '欧美 AV 待确认', tagType: 'warning' },
  ```

### 6.3 `admin-web/src/views/ScrapePreview.vue`

- 这里只处理 `av_scrape_pending` 的确认/弃刮门控，不承载已 `ready` 视频的手动重刮
- 进入页面时如果 `video.status === 'av_scrape_pending'` 且 `video.metadata.scrape_preview` 是非空数组，直接渲染该数组（**不**调用 `POST /admin/scrape/preview`）
- 候选卡片右上角加徽章 component：根据 `candidate.match_source` 显示 `hash 命中` / `场景搜索` / `电影搜索`
- `match_source === 'oshash'` 的候选默认选中 + scroll into view
- 保存按钮文案改为"确认刮削"（仅在 `status === 'av_scrape_pending'` 时）
- 新增按钮"弃刮"：放在保存按钮旁，红色 outline，点击弹 `el-message-box` 确认 + 输入 reason，确认后调用 `PUT /admin/scrape/skip`
- 新增按钮"重新搜索"：弹 dialog 让 admin 输入 keyword 或 detail URL，提交后调 `POST /admin/scrape/preview` 带 `bypass_cache=true` 重新写 `scrape_preview`
- 新增折叠面板"刮削诊断"：展开后渲染 `video.metadata.scrape_attempt`（hash_used / keywords_tried / errors）

**阶段验证**：
- `cd admin-web && npm test` 通过
- 手动开 `npm run dev`，上传一个欧美 AV → VideoList 里筛选 → 进入 ScrapePreview → 确认 / 弃刮 / 重搜三个路径都顺
- 上传日本 AV：UI 没有"欧美 AV 待确认"残留

---

## 阶段 7：admin-web `AVManualScrape.vue` 集成 hash 红利

### 7.1 背景

`AVManualScrape.vue` 服务于 [`手动刮削`] 术语：对已 ready 的 AV 视频事后重新走刮削预览 + 选候选 + 覆盖 metadata。本期阶段 3.5 让 `PreviewAVSearch` 透传 `video_id` 时自动跑 hash 路径——前端需要确认这条透传是否已经在调用上送了 `video_id`，并补齐缺失部分。它与 6.3 的 `av_scrape_pending` 门控页分离，候选可复用，状态语义不复用。

### 7.2 实施步骤

1. 打开 `admin-web/src/views/AVManualScrape.vue`，找到调用 `POST /admin/scrape/av/preview`（或等价端点）的位置
2. 检查 request body 是否带 `video_id`：
   - 若已带 → 后端阶段 3.5 改完即可生效，前端不动
   - 若没带 → 加上 `video_id: video.value.id`
3. 候选列表渲染部分加 `hash 命中` 徽章逻辑：复用阶段 6.3 ScrapePreview 用的徽章 component（建议把徽章抽成 `src/components/AVMatchSourceTag.vue` 共享）
4. 确认 confirm 按钮**不**改文案（仍为"保存" / "覆盖"——这是手动刮削语义，不是门控确认）

### 7.3 后端校验

确认阶段 3.3 改造后的 `ConfirmAV`：
- 当 video.Status 不是 `av_scrape_pending`（即手动刮削场景，status 已是 `ready`）→ 不入队 transcode、不改 status
- 这是 PRD 故事 D 与回归点的关键，必须有显式 unit test 覆盖：`TestConfirmAV_FromReadyStatus_DoesNotEnqueueTranscode`

### 7.4 验证

- 用一个已 ready 的欧美 AV（手工 SQL 给它 set 一个 `os_hash`）走 AVManualScrape：
  - 候选列表里看到带 `hash 命中` 徽章的候选
  - 点保存后视频仍是 ready 状态、transcode 没被重新触发
- 用一个 `os_hash` 为 NULL 的日本 AV 走同一页面：
  - 候选列表纯关键字结果，无 hash 徽章
  - 行为跟改动前完全一致

---

## 阶段 8：CONTEXT.md / plan.md

### 8.1 CONTEXT.md
- 本任务在 grill-with-docs 阶段已经同步落地 `AV 上传分类术语` 章节和 5 个新术语 + 1 个原术语补强。
- implement 阶段若与术语定义产生偏离（例如 N7 改成"零候选自动弃刮"），必须**回头修订 CONTEXT.md + PRD**，不允许只改代码。

### 8.2 plan.md
- 追加两条逆时序条目：
  - 一条本任务的**计划**条目（写在 grill-with-docs 完成时间，列出 PRD 链接 + 关键设计决策摘要）
  - 一条实施完成的**实施更新**条目（含 commit hash、影响文件列表、自动化测试与手动验收结果）

---

## 阶段 9：`cmd/gen-theporndb-oumei` 工具（可选 / 独立阶段）

### 9.1 目的

把 `references/mdcx/mdcx/manual.py:57` 起的 `OUMEI_NAME` Python 字典自动转成 Go map 字面量，避免 200+ 条手工录入出错；未来 MDCX 上游更新字典只需重跑脚本即可同步。

### 9.2 实现

`cmd/gen-theporndb-oumei/main.go`：

- 用 `os.ReadFile` 读 `references/mdcx/mdcx/manual.py`
- 简单状态机扫描：跳到 `OUMEI_NAME = {` 那行后，按行读 `"key": "value",` 直到匹配的 `}` 结束
- 写入 `internal/services/theporndb_oumei_name.go`，附 `// 由 cmd/gen-theporndb-oumei 生成，请勿手工编辑` 的 do-not-edit 注释
- 失败时报错退出（不假装生成成功）

### 9.3 不进 `go:generate`

不在 `main.go` 的 `//go:generate` 指令里挂这个工具——它依赖 `references/` 路径，CI 上不稳定。改在 `cmd/gen-theporndb-oumei/README.md` 写一行使用说明：

```
# 当 MDCX 上游更新 OUMEI_NAME 字典时手动重跑：
go run ./cmd/gen-theporndb-oumei
```

### 9.4 顺序

阶段 9 可以在阶段 4.2 之前先跑一遍把 `theporndb_oumei_name.go` 生成出来，再让阶段 4.2 引用它；或者阶段 4.2 先手工搬运一份 fixture（20 条左右）让代码先编译过，再回头跑阶段 9 替换全集。**推荐先跑阶段 9** 一次成型，避免双份字典。

---

## 提交策略

按阶段 1 / 阶段 2 / 阶段 3+5 / 阶段 4 / 阶段 6 / 阶段 7 / 阶段 8 / 阶段 9 切 commit；每个 commit 编译通过 + 该阶段最小验证通过。中文 commit message。

阶段 4 五个修复可以再细切五个 commit 也可以打成一个——视代码量决定，但每个修复都要有对应的单测在该 commit 里同步落。

实施完成（所有阶段验收过）才进入 review.md 跑全量验证。

# Review 反馈：欧美 AV 刮削确认门控（commit `ac7766f2`）

- 日期：2026-05-23
- 审查范围：commit `ac7766f2caa7992af95d484f575f7f04e106228c` "完成欧美 AV 刮削确认门控"，31 个文件 +1409/-75。
- 同行审查者：Claude（review skill）
- PRD/Implement 锚点：`tasks/2026-05-23-western-av-oshash-confirm-gate/prd.md`、`tasks/2026-05-23-western-av-oshash-confirm-gate/implement.md`

## 概览

PR 落地了 PRD 中规划的核心欧美 AV 门控流程：新增 oshash 计算与持久化、`av_scrape_pending` 状态、`/admin/scrape/skip` 端点、admin-web 待确认 UI、ThePornDB 多项修复，以及对应的迁移 + repo + 模型扩展。整体落地紧贴 PRD，状态机/数据流/UI 都对齐了术语表，回归点也做了显式分支。下面按严重程度列出审查发现。

---

## 🔴 必须修复（潜在功能回归）

### 1. ThePornDB 2-digit 年份日期丢失 "20" 世纪前缀

**位置**：`internal/services/scraper_av_mdcx_detail_sites.go:1746-1752`

```go
matches := datedNumberPattern.FindStringSubmatch(value)
if len(matches) == 3 {
    fullNumber := matches[0]
    series := thePornDBLongName(...)
    date := strings.ReplaceAll(matches[2], ".", "-")
```

旧实现明确给 2 位年份补 `"20" +`，新正则同时覆盖 2 位与 4 位年份，但提取逻辑**统一去掉了补 "20" 的步骤**。对最常见的 OUMEI 命名 `Series.20.05.12.xxx.mp4`，date 会变成 `20-05-12` 而不是 `2020-05-12`，对 ThePornDB 的日期相似度比较直接失效。

**建议修复**：在拿到 `matches[2]` 之后根据长度判断是否补 `20-`：

```go
date := strings.ReplaceAll(strings.ReplaceAll(matches[2], ".", "-"), "_", "-")
if len(date) == 8 { // YY-MM-DD
    date = "20" + date
}
```

需要新增单测覆盖 `Series.20.05.12.xxx.mp4` / `Series_2020-05-12_Title.mp4` / `[Series] 20.05.12.Title.mp4` 三种命名，回到 PRD §4.2 验收项。

---

### 2. `/scenes` 端点参数被一起改成 `q=`，违背 implement.md §4.1

**位置**：`internal/services/scraper_av_mdcx_detail_sites.go:1778`

```go
func thePornDBSearchURL(baseURL, category, keyword string) string {
    values := url.Values{}
    values.Set("q", keyword)   // ← 不区分 category
    ...
```

implement.md §4.1 明确写：

```go
case "movies": values.Set("q", keyword)
default:       values.Set("parse", keyword)   // scenes
```

现在 `/scenes` 也用 `q=`，可能破坏既有 scenes 搜索（旧逻辑相当多就靠 `parse=`）。

**建议**：如果决意统一为 `q=`，需要：
1. PR 描述里写明原因
2. 跑一组真实 cassette/fixture 验证 scenes 命中率没退化
3. 同步 PRD/implement.md

否则按计划用 `switch category` 区分。

---

### 3. `pkg/oshash` 单测没有与 MDCX/Python 黄金值对齐

**位置**：`pkg/oshash/oshash_test.go`

当前仅测了 (a) 文件 < 128 KiB 报错 (b) 两次调用结果相等 + 长度 16。PRD §2.2 与 implement.md §2.2 都明确要求**与 Python `oshash.oshash()` 在固定 fixture 上字节级一致**——只验自洽是没法捕捉端序/分块偏移这种致命算法 bug 的（hash 错了 ThePornDB 永远无命中且没有任何报错）。

**建议**：补一条 fixture 测试：

```go
// 固定 fixture：256 KiB 全 0x42，期望值由 references/mdcx 同 fixture 跑出来后写死。
const fixture256KiBOf0x42Expected = "....." // PR 描述里附 Python 重现脚本
```

不挡 merge，但**强烈建议合入前补上**——这是整条 hash 通道的唯一外部约束。

---

## 🟡 需要注意

### 4. `ConfirmAV` → handler 转码入队依赖**陈旧的** `video.Status`（TOCTOU）

**位置**：`internal/handlers/admin_scrape.go:300` 与 `:405-420`

```go
video, err := a.repo.GetVideoByID(...)   // 读到 status=av_scrape_pending
...
a.scrapeSvc.ConfirmAV(...)               // 内部把 status 改成 processing
if shouldEnqueueAdminScrapeConfirmTranscode(video) { ... } // 用陈旧 video 判断
```

当前能 work，因为 `shouldEnqueueAdminScrapeConfirmTranscode` 检查的是 "is it av_scrape_pending"，刚好需要旧状态。但很脆弱：今后任何 "完成后重新读 video" 的好心 refactor 都会让这条门控失效，且 `ConfirmAV` 自己（`scraper.go:788-791`）就重写过 status。

**建议**（任选一种）：

- (a) 在 `ConfirmAV` 内部根据 `video.Status == "av_scrape_pending"` 自己入队转码（按 implement.md §3.3 的原始设计），handler 不再判 status
- (b) 用一个显式的返回值（如 `ConfirmAVResult{ShouldTranscode: bool}`）把信号往外传，不再依赖 handler 自己保存旧值
- (c) 至少加一行注释 `// shouldEnqueueAdminScrapeConfirmTranscode intentionally reads pre-confirm status`

---

### 5. site_category 推断逻辑出现三份近似实现

**位置**：
- `internal/queue/scrape_tasks.go:345 queueAVSiteCategory`
- `internal/handlers/admin_scrape.go:504 avSiteCategoryFromAdminVideo`
- `internal/services/scraper.go:1561 avSiteCategoryFromVideo`

三段都做 "先看 metadata.site_category；空 + 有 os_hash → western" 的同一件事。下次需求改 fallback（比如加 IPTV 来源）三个地方很可能漂移。

**建议**：合并为一个 `services.InferAVSiteCategory(video models.Video)` 给 queue / handler / service 共用。

---

### 6. retag-to-western 在零候选时直接失败 + Asynq 重试 3 次

**位置**：`internal/queue/scrape_tasks.go:140-147` + `autoScrapeAV:308-310`

```go
case "av":
    scrapeErr = p.autoScrapeAV(...) // 0 candidates → fmt.Errorf("no av candidate ...")
...
if scrapeErr != nil { if targetType == "av" { return scrapeErr } } // Asynq 重试 3 次
```

上传链路同样零候选时落 `av_scrape_pending` 等管理员（PRD §5 网络分区/§N7），retag 却走 fail + retry，行为不一致。如果 retag 走来的视频自己也已经持有 os_hash，让它走和上传同款 "落 av_scrape_pending + 暴露 preview = []" 会更对称。

PRD 回归点 #2 只声明了 retag-to-`av(非western)` 行为不变，**没有锁死 retag-to-western 的行为**，所以建议主动对齐 upload 路径，免得管理员遇到 retag 后视频静默失踪。

---

### 7. 阶段 4.2 OUMEI_NAME 走 inline 字典 + 没有 codegen 工具

**位置**：`internal/services/scraper_av_mdcx_detail_sites.go:2026-`

直接手写了 ~200 条 `thePornDBLongNameMap`。`cmd/gen-theporndb-oumei` 没建。功能上没问题，但 PRD §9 的核心论点是 "MDCX 上游更新就需重跑脚本"——手抄一份脱离上游后将来同步靠人工 diff。

**建议**：
- 至少加一段 `// 来源：references/mdcx/mdcx/manual.py:OUMEI_NAME，最后同步日期 2026-05-23` 注释
- 若要省略 codegen，PRD/implement 应同步修订（implement.md §3 显式约束 "实施阶段不允许单方面偏离 PRD"）

---

### 8. `scrapeAVWesternUpload` 把 ThePornDB 401 / token 缺失也吞成 "零候选 pending"

**位置**：`internal/services/scraper.go:1451-1490` & `scraper_av_mdcx_detail_sites.go:914`（401 不重试直接返回 error）

链路：
1. `fetchThePornDBJSON` 遇 401 immediately return error
2. `collectThePornDBHashCandidate` 把任何错误吞掉返回 `nil, nil`（含 401）
3. `scrapeAVWesternUpload` 把 `scrapeErr` 写进 trace，但不阻塞，落 `av_scrape_pending` + `scrape_preview=[]`

效果：token 配错时管理员看到的是 "自动刮削无命中" 而不是 "鉴权失败"——需要主动 inspect `metadata.scrape_attempt.error` 才知道。PRD §N7 选择了这种宽容兜底（zero candidate 不自动 skip），但 admin 端 UI 应在 "待确认" panel 显式渲染 `scrape_attempt.error`（VideoList.vue 已经做了 `attempt.error || '-'`，OK），同时 `ScrapePreview.vue`/`AVManualScrape.vue` 也建议把 401/403 这类与无候选区分出来的提示文案做一个区分。

属于 UX 提升，不挡 merge。

---

### 9. Migration down 把 `av_scrape_pending` 回退到 `uploaded`

**位置**：`migrations/0021_western_av_oshash_gate.down.sql:1`

```sql
UPDATE videos SET status = 'uploaded' WHERE status = 'av_scrape_pending';
```

down 在生产基本不会跑，但如果跑了：那些视频此刻**没有 transcoded_path**，且回到 `uploaded` 状态不会自动重新入队 scrape / transcode（admin 路径上 `uploaded` 是手工状态）。

**建议**：如果真要回滚，更安全的可能是 `UPDATE ... SET status='processing'` + 单独写一个 backfill 脚本入队转码。PRD 没要求，可不动，但写一行注释提示 down 之后需要管理员处理这批视频会比较稳。

---

## 🟢 做得好

- **CHECK 约束改写**：drop & recreate 干净，down 顺序也对（先 update 再换约束）。
- **`videos.os_hash` 用 `CHAR(16) + partial index ... WHERE os_hash IS NOT NULL`**：和 PRD §3.1 完全一致；不为非欧美 AV 浪费索引。
- **oshash 计算失败不阻塞上传**：`s.logger.Warn` 后落 NULL，符合 §5 出错边界。
- **`queueAVSiteCategory` 用 os_hash 兜底推断**：retag 场景下即使 metadata 丢了也能走对分支。
- **chunk_upload merge 路径**：通过共享 `SaveUploadedFile` 自动复用 oshash 逻辑——比在 `chunk_upload.go` 里重新算一遍清爽。
- **handler 层 transcode 入队的 `Force: true`**：避免 ConfirmAV 重写 status 后转码 worker 跳过的副作用。
- **回归路径覆盖**：日本/FC2 AV 上传零变化、movie/episode/short 路径不动、AVManualScrape 不改语义只加 `hash 命中` 徽章——都和 PRD §4.3 对齐。
- **scrape_attempt / scrape_preview 写入 metadata 而非新列**：保持 schema 简洁、UI 可直接渲染。
- **`scrape_skip` 审计三件套**（`scrape_skipped` / `scrape_skip_reason` / `scrape_skipped_at`）落得完整，方便事后排查。
- **重试包装把 401 与 5xx/429 分开处理**：401 不重试避免无效流量，5xx/429/网络错误 200/400/800ms 退避——符合 PRD §3.5 第 3 条。
- **`thePornDBMinSimilarity = 0.4` 常量阈值**：实现简单清晰，三桶都低于阈值时主动返回空（line 1847），交给上层进下一个 keyword。
- **admin-web 待确认 panel**：在 `VideoList` 详情页直接展示候选列表 + hash 信息 + 错误，无需先跳到 `ScrapePreview`，体验比 PRD 设计的还顺手。

---

## 合入前最低门槛

| 项目 | 严重度 | 工作量 |
| --- | --- | --- |
| #1 date "20" 前缀回归 + 单测 | 🔴 阻塞 | 半小时 |
| #2 `/scenes` 是否真要从 `parse=` 切到 `q=`，否则恢复 switch + 加 fixture 测试 | 🔴 阻塞 | 半小时 |
| #3 oshash 黄金值 fixture 测试 | 🔴 强烈建议 | 半小时（含 Python 跑一次） |
| #4 ConfirmAV TOCTOU：要么加注释要么重构 | 🟡 选一 | 5 分钟（注释） |
| #5-#9 | 🟡 后续单独 PR 可 | 视改动 |

#1 / #2 没修就 merge，欧美 AV 自动命中率反而**比改动前更差**（旧 2-digit 年份案例彻底变垃圾），违反 PRD 故事 E（"刮不到比例从 ~80% 降到 ~30%"）的次级目标承诺。其他都不挡 merge。

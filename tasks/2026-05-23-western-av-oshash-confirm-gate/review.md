# Review：欧美 AV oshash 与刮削确认门控

按"自动化 → 手动 → 回归 → DONE"顺序走。任何一档卡住都不算完成，回到 implement 阶段补完再回来重跑该档。

---

## A. 自动化验收

### A.1 Go 后端测试

```bash
go test ./pkg/oshash/...
go test ./internal/services/... -run "ThePornDB|AVScrape|ConfirmAV|SkipScrape|OSHash"
go test ./internal/queue/... -run "ScrapeAV|ScrapeRetag"
go test ./internal/handlers/... -run "ScrapeSkip|Upload"
go test ./internal/repository/... -run "OSHash"
go test ./...   # 全量,确保零回归
go vet ./...
```

通过标准：全绿。任何 fail 都不算可接受——即使是"看似无关"的回归，也要排查是否本任务引入。

### A.2 admin-web 测试

```bash
cd admin-web
npm test
npm run build   # 也跑一遍 build,确保新 radio / 徽章 / 按钮 在 production bundle 里没编译错误
```

### A.3 Migration 幂等

```bash
bash scripts/dev-down.sh
bash scripts/dev-up.sh                            # up 一次
psql ... -c "\d videos"                            # 校验 os_hash 列 + index + CHECK 约束
# 手工跑 down → up 一次校验可回退
```

---

## B. 手动验收脚本

### 前置准备

- 准备 3 个真实欧美 AV 样本文件（不要小于 128KB；项目 .gitignore 中 `references/` 已忽略,放这里临时用即可）：
  - **样本 1**：原版未压缩、已知能在 ThePornDB hash 库命中（找一个 OUMEI_NAME 字典里常见站点 + 近期发行）
  - **样本 2**：原版但 ThePornDB hash 库没有索引（小众发行；keyword 路径仍能搜到）
  - **样本 3**：完全找不到的稀有发行（hash 没有、keyword 也搜不到）
- 准备 1 个日本 AV 样本作为对照
- 准备 1 个 movie 样本作对照
- `bash scripts/dev-up.sh` 全栈起好

### B.1 故事 A & B 主链路（hash 命中）

1. 浏览器开 admin-web 上传页，选样本 1，类型 `AV`，地区 `欧美`，提交
2. 上传响应应包含 `status: "scraping"`
3. 等待自动刮削任务完成后，DB 查询：
   ```sql
   SELECT id, status, os_hash,
          metadata->'scrape_preview' AS preview,
          metadata->>'site_category' AS cat
   FROM videos WHERE id = '<上传返回的 id>';
   ```
   预期：
   - `status = 'av_scrape_pending'`
   - `os_hash` 是 16 位 hex 非空
   - `cat = 'western'`
   - `preview` 是非空 JSON 数组，首个元素 `match_source = 'oshash'`
   - **不**存在 `transcoded_path`，Asynq UI 上看不到该 video_id 的 transcode task
4. admin-web 进 `VideoList`，选状态 `欧美 AV 待确认`，看到样本 1
5. 点进去 `ScrapePreview`，首个候选有 `hash 命中` 徽章 + 默认预选中
6. 点"确认刮削"，DB 复查：
   - `status` → `processing` 或 `ready`（视 transcode 速度）
   - `transcoded_path` 非空
   - `metadata.title` / `actors` / `poster_url` 等写入

### B.2 故事 B 补充（hash miss + keyword 命中）

1. 上传样本 2（同上配置）
2. 等待落 `av_scrape_pending`
3. `metadata.scrape_preview` 仍非空，但**没有** `match_source=oshash` 候选；至少有一条 `keyword:scenes` 或 `keyword:movies`
4. admin 选一条 confirm，链路完成同 B.1

### B.3 故事 C 主链路（弃刮）

1. 上传样本 3
2. 等待落 `av_scrape_pending`，DB 看到 `scrape_preview` 是空数组 `[]`
3. admin-web `ScrapePreview` 显示"自动刮削未命中候选，请手动输入关键字重试或弃刮"
4. 点"弃刮"按钮，输入 reason `"测试稀有发行"`
5. DB 复查：
   - `metadata.scrape_skipped = true`
   - `metadata.scrape_skip_reason = "测试稀有发行"`
   - `metadata.scrape_skipped_at` 非空
   - `status = 'processing'`（最终 ready）
   - `transcoded_path` 非空
6. TV / phone 客户端可播该视频（元数据空但播放链路无障碍）

### B.4 故事 C 补充（"重新搜索"逃生）

1. 在 B.3 弃刮之前，先点"重新搜索"
2. dialog 中输入自定义 keyword（例如人工识别出的真实番号）
3. 点搜索，候选列表刷新；如果命中则可走 confirm；如果仍不命中再走弃刮

### B.5 故事 D（手动刮削复用 hash）

1. 拿 B.1 已经 ready 的视频
2. admin 进 `AVManualScrape` 页（或对该视频走"重新刮削"入口）
3. 候选列表首位仍带 `hash 命中` 徽章
4. 点保存：
   - DB 看到 metadata 被覆盖
   - `status` 仍为 `ready`，**不**重新入队 transcode
   - `transcoded_path` 不变

### B.6 次级目标·`/movies` 修复

1. 抓包工具（mitmproxy / Charles）拦截 ThePornDB 出站请求
2. 触发一个会走 movies fallback 的样本
3. 校验 URL：`https://api.theporndb.net/movies?q=...&per_page=100`（不是 `parse=`）

### B.7 次级目标·OUMEI_NAME 字典

1. 准备文件名 `Wgp.20.05.12.SceneName.mp4`（或字典里任意其它短名）
2. 上传为欧美 AV
3. 刮削诊断面板（`scrape_attempt`）显示 keyword 之一是 `whengirlsplay 2020-05-12`（series 已被展开）

### B.8 次级目标·重试

1. 临时把 `AV_SITE_URL_THEPORNDB` env 指向一个 mock server，按"前两次返回 429、第三次返回 200"配置
2. 上传样本，最终成功（不报错 / 不卡死）
3. Mock server 改成"返回 401"，上传后看到 `scrape_attempt.errors` 里捕获了 401，**没有**反复重试

### B.9 次级目标·阈值

1. 上传一个文件名跟 ThePornDB 候选全部相似度低的样本
2. `scrape_preview` 为空（不是"硬选一个低分候选"）

### B.10 次级目标·正则

1. 上传文件名 `[BlackedRaw] 20.05.12 SceneName.mp4`（P3 正则形态）
2. 刮削诊断面板看到 series + date 正确抽出

---

## C. 回归验收

### C.1 日本 AV
1. 上传一个日本 AV 样本，地区 `日本`
2. DB 校验：
   - `videos.os_hash IS NULL`
   - `videos.metadata.site_category = 'japanese'`
   - `videos.metadata.scrape_preview` 不存在
   - 流转过程 `uploaded → scraping → ready`（或 `uploaded`，跟现行一致）
   - transcode 自动跑完
3. 行为完全跟改动前一致

### C.2 FC2 AV
- 同 C.1，地区 `FC2`

### C.3 movie / episode / 短视频
1. 每种类型上传一个样本
2. 状态机 / transcode / metadata 完全跟改动前一致
3. 不应出现 `os_hash` 计算调用（log 里看不到 `oshash compute` 字样）

### C.4 现有手动刮削（非欧美档）
1. 对一个已 ready 的日本 AV 走 `AVManualScrape`
2. 候选列表纯关键字结果，无 hash 徽章
3. confirm 后 status 不变、不入队 transcode

### C.5 现有 retag 路径
1. 把一个 short 视频 retag 成 `av`（在 admin VideoList 编辑），地区选 `日本`
2. 应走 `EnqueueScrapeRetag`，落 ready（不接入门控）
3. 把同一个 video retag 成 `av` 地区 `欧美`：
   - 若原文件还在 → `os_hash` 被补算
   - 落 ready（**不**回到 `av_scrape_pending`——retag 不是新上传）

### C.6 现有 ThePornDB 测试套
跑 `go test ./internal/services/... -run "ThePornDB"` 全绿。重点关注 `scraper_av_mdcx_sites_test.go` / `scraper_av_strategy_test.go` 等既有用例。

---

## D. 验收脚本（半自动化）

可选：把 B.1 / B.2 / B.3 / C.1 / C.3 五条核心路径写成一个 shell 脚本 `scripts/verify-western-av-gate.sh`，调用 admin-web 的 REST API 直接跑端到端校验，无需开浏览器。脚本本身不进 PR（避免污染主仓），放在 task 目录下作为 review artifact 即可。

---

## E. 文档与发布

- [ ] CONTEXT.md：5 个新术语 + 1 个补强已在 grill-with-docs 阶段落地；implement 阶段无偏离则保持
- [ ] plan.md：追加本任务实施完成条目（commit hash、影响文件、验证摘要）
- [ ] PR 描述：列出主目标 + 5 个次级目标，提供 B.1 / B.3 截图，附 hash 命中前后体验对比
- [ ] Android：不需要 versionCode 升

---

## F. DONE 标准

全部满足才可在该任务目录新增 `DONE.md`：

1. **自动化验收 A 全过**：所有 `go test` + `npm test` + migration 幂等通过
2. **手动验收 B 全过**：B.1 / B.2 / B.3 / B.4 / B.5 五个主链路 + B.6–B.10 五个次级目标全部走通
3. **回归 C 全过**：日本 / FC2 / movie / episode / 短视频 / 既有 ThePornDB 测试套零回归
4. **CONTEXT.md / plan.md 同步**：术语、计划条目、实施更新条目齐全
5. **用户验收**：把 B.1 / B.3 在用户面前演示一次，用户确认后由**用户**在本目录新增 `DONE.md`，记录完成日期、关联 commit hash、用户验收备注

`DONE.md` 不允许由实施 agent 自行创建——这是 CONTEXT.md 中 `tasks 任务三段执行流` 明确要求的人机分工边界。

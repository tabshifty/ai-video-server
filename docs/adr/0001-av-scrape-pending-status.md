# 新增 `av_scrape_pending` 视频状态承载欧美 AV 刮削确认门控

[`欧美 AV`] 上传后要求管理员手动确认刮削结果之后才允许 transcode，这中间需要一个"自动刮削已跑完、等人确认"的中间态。决定新增独立 status `av_scrape_pending`（与 `tv_pending` 对称），加入 `videos.status` 的 CHECK 约束，由状态机本身阻断 transcode 入队——而不是复用 `uploaded` + metadata flag，因为后者会让所有按 status 过滤的查询（admin 列表、统计、导出）都要变成"status + flag"复合判断，长期心智负担远高于一次性 migration 成本。同时拒绝引入通用 `pending_confirm` 状态以避免过早抽象，本期范围明确只覆盖欧美 AV。

## 考虑过的替代方案

- **复用 `uploaded` + `metadata.scrape_preview_ready=true`**：零迁移，但 `uploaded` 同时承载"刚上传啥也没干"和"等确认"两种语义，所有 status 筛选点都要补 flag 判断
- **新增通用 `pending_confirm` 状态**：覆盖未来可能扩展到 movie / episode 的场景，但本期明确只做欧美 AV（PRD N1），过早抽象给后续维护者制造"这个状态到底覆盖哪些类型"的猜测空间

## 关联

- `migrations/0021_western_av_oshash_gate.up.sql`
- `CONTEXT.md` 中 [`刮削确认门控`] 术语
- `tasks/2026-05-23-western-av-oshash-confirm-gate/` 全套

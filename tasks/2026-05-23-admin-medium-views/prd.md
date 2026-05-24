# PRD：admin 重设计 Phase 3 · 中等视图重排

- 日期：2026-05-23
- 阶段：Phase 3 / 4（中等视图重排）
- 前置任务：`tasks/2026-05-23-admin-shell-redesign/` + `tasks/2026-05-23-admin-simple-views/` 必须先标记 DONE
- 范围：4 个中等复杂度视图按 Phase 1 设计系统重排

## 1. 用户故事

作为管理员，希望：

1. ScrapePreview / AVManualScrape 这类「预览候选 + 确认」流程视觉清晰、卡片节奏统一
2. TvSeriesManage 的「剧集列表 + 季管理」分区清晰，SectionCard 折叠次要信息
3. ImageCollectionManage 的合集卡片网格视觉跟 Phase 4 的 ImageManage 一致（提前用 SectionCard + 网格预演）

## 2. 作用域

本期触碰的 4 个视图（按文件大小升序）：

| 视图 | 当前行数 | 主要改造点 |
|------|---------|-----------|
| `views/AVManualScrape.vue` | 549 | PageHeader + Toolbar + SectionCard 分区（搜索 / 候选预览 / 已选）+ EmptyState |
| `views/ScrapePreview.vue` | 651 | 同 AVManualScrape；候选卡片网格用 SectionCard 承载 |
| `views/ImageCollectionManage.vue` | 685 | PageHeader + Toolbar + 合集网格 SectionCard + 上传 dialog 改 drawer |
| `views/TvSeriesManage.vue` | 716 | PageHeader + Toolbar + 剧 / 季 / 集三层 SectionCard 嵌套；编辑 dialog 改 drawer |

## 3. 非目标

- N1：本期**不动**三巨头（Phase 4）
- N2：本期**不改** Phase 2 已重排过的 7 个视图
- N3：本期**不改** API 接口、Pinia store 形状、路由 URL
- N4：本期**不改** ScrapePreview / AVManualScrape 的刮削流程业务逻辑——视觉重排不动 CONTEXT.md「手动刮削」/ 「刮削确认门控」语义
- N5：本期 ImageCollectionManage 的网格视图是 Phase 4 ImageManage 网格视图的**预演**——两个视图都使用相同的 SectionCard + grid 模式

## 4. 引用的前置阶段决策

| 来源 | 引用点 |
|------|--------|
| Phase 1 PRD Q3 / Q6 / Q8 / Q9 | 设计语言 / 字体 / 色板 / 侧栏底色 |
| Phase 1 Implement 第 2.6 节 | 7 个 base 组件接口 |
| Phase 2 实际落地形态 | SectionCard / Toolbar 在简单视图上的真实使用模式 |

## 5. 验收用例骨架（待 Phase 1 + Phase 2 验收后细化）

| 用例 | 路径 | 期望 |
|------|------|------|
| H31 | 进入 4 个 view | PageHeader 统一 + 无残留旧自定义样式 |
| H32 | TvSeriesManage 编辑剧 / 季 / 集 | 编辑入口改 drawer（与 Phase 4 三巨头一致）；底部固定 Toolbar 含取消 / 保存 |
| H33 | ScrapePreview / AVManualScrape 候选卡片网格 | 卡片用 SectionCard 承载；hover / focus / 选中状态使用 token |
| H34 | ImageCollectionManage 合集网格 | 网格视图（不再是表格）；选中浮起 BulkActionBar |
| H35 | 各 view 上传 / 编辑 dialog 改 drawer 后 | drawer 宽 560px；底部 Toolbar 固定；dirty 关闭弹确认 |

## 6. 实施前提清单

- Phase 1 + Phase 2 三件套全部 Done
- 4 个视图的「编辑 / 上传 dialog → drawer」改造原则与 Phase 4 三巨头对齐（共享 drawer + SectionCard + 底部 Toolbar 模式）

## 7. 自动化必跑

```bash
cd admin-web
npm run build
npm test
```

不新增 spec；依赖 Phase 1 spec 保持绿。

## 8. 待 Phase 2 验收后细化的内容

- Implement.md（4 个视图详细 diff）
- review.md（H31–H35 详细期望 + 截图归档清单）
- drawer 替换 dialog 的具体 API 兼容性 check

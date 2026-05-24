# PRD：admin 重设计 Phase 4 · 三巨头 IA 改造

- 日期：2026-05-23
- 阶段：Phase 4 / 4（三巨头 IA 改造）
- 前置任务：Phase 1 / 2 / 3 全部标记 DONE
- 范围：VideoList / ImageManage / VideoUpload 三个最大视图，IA 改造而非仅视觉重排

## 1. 用户故事

作为管理员，希望：

1. **VideoList**：筛选条不再常驻挤占屏幕，列可自选，编辑视频改为右侧抽屉一边看一边改，多选时底部浮起批量操作条
2. **ImageManage**：默认网格视图（图片产品的直觉视图），切回表格视图自由选择；批量选中浮起 BulkActionBar；上传从大 dialog 改 drawer
3. **VideoUpload**：单页大表分 3 步 wizard（选文件 → 基础信息 → 关联与上传）；上一步保留已填字段；上传执行后保留进度与历史不被 step 切换覆盖

## 2. 作用域

| 视图 | 当前行数 | IA 改造 |
|------|---------|---------|
| `views/VideoList.vue` | 1421 | 筛选条折叠为 chip + 「更多筛选」抽屉；列设置（localStorage 持久化）；编辑 dialog 改 drawer（560px 含 SectionCard 折叠区块）；BulkActionBar |
| `views/ImageManage.vue` | 1045 | 视图切换（网格默认 / 列表可选 + localStorage 持久化）；筛选条折叠；上传 dialog 改 drawer；BulkActionBar |
| `views/VideoUpload.vue` | 719 | 拆 3 步 wizard：`views/VideoUpload/StepFile.vue` / `StepBasic.vue` / `StepRelate.vue`；主 `VideoUpload.vue` 只做 step 容器 + 共享 form state |

## 3. 非目标

- N1：本期**不改** API 接口、Pinia store 形状、路由 URL
- N2：本期**不引入** vue-grid-layout / interact.js / 拖拽布局
- N3：本期**不接入**「模板保存」（VideoUpload 重复上传场景的快捷模板）
- N4：本期**不接入**「图片标签 AI 自动识别」
- N5：本期**不做"分屏视图"**（左列右详情）—— drawer 已够用
- N6：本期**不改** VideoList 编辑 drawer 内部的业务字段（episode/av/short 各自的 conditional 字段保持现有 schema）

## 4. 引用的前置阶段决策

| 来源 | 引用点 |
|------|--------|
| Phase 1 PRD Q10 | 三巨头 IA 改造方案详表 |
| Phase 1 Implement 第 2.6 节 | base 组件接口（SectionCard / Toolbar / BulkActionBar） |
| Phase 1 Implement 第 2.2 节 | element-overrides 的 drawer 样式 |
| Phase 3 实际落地形态 | drawer 替换 dialog 的成熟模式 |

## 5. 新增 / 拟引入的 `CONTEXT.md` 术语

| 术语 | 一句话定义 |
|------|----------|
| `admin 编辑入口 Drawer` | admin 端 VideoList / ImageManage / VideoUpload 三巨头与 Phase 3 视图的"编辑/上传"操作统一改用右侧 drawer 而非居中 dialog；宽度 560px（< 1024px 改全屏）；底部固定 Toolbar 含取消/保存；dirty 关闭弹确认；drawer 内不嵌套 dialog |
| `admin 表格列设置` | VideoList 顶部「列设置」按钮，用户可勾选显示/隐藏列；偏好写入 `localStorage` key `admin-videolist-columns`；窗口 < 1280px 自动隐藏次要列但仍尊重用户既有显式偏好 |
| `admin 视图模式切换` | ImageManage 顶部「视图切换」开关（网格/列表）；默认网格；偏好写入 `localStorage` key `admin-imagemanage-view`；切换状态跨会话持久化 |
| `admin 上传向导三步` | VideoUpload 拆 3 步：选文件 / 基础信息 / 关联与上传；上一步保留字段；第 3 步上传进度与历史不被 step 切换覆盖；类型选择驱动条件字段（AV 地区分类 / 短视频所属合集） |
| `admin 批量操作浮条` | 表格/网格选中 ≥ 1 项时顶部浮起 BulkActionBar；含选中数量 + 主操作（批量删除/批量改状态/批量打标）；取消选择即消失 |

## 6. 验收用例骨架（待 Phase 3 验收后细化）

| 用例 | 路径 | 期望 |
|------|------|------|
| H41 | VideoList 进入 | 筛选条收成 chip 行 + 「更多筛选」按钮；表格仅显示用户配置或默认列 |
| H42 | VideoList 点击「列设置」 | popper / drawer 显示列勾选；勾选状态写 localStorage；刷新保持 |
| H43 | VideoList 选中 2 行 | 顶部浮起 BulkActionBar 显示「已选 2 项」+ 主操作；取消选择消失 |
| H44 | VideoList 行操作「编辑」 | 右侧 drawer 滑出 560px；内含 SectionCard 折叠区块；播放预览置顶；底部固定 Toolbar |
| H45 | ImageManage 进入 | 默认网格视图；切换按钮可切到列表；切换状态刷新后保持 |
| H46 | ImageManage 网格 hover | 卡片显示快捷操作；左上角 selection checkbox |
| H47 | ImageManage 上传 | 上传按钮唤起右侧 drawer，分「批量选图」+「默认元数据」+「上传队列」SectionCard |
| H48 | VideoUpload 进入 | 顶部 el-steps 显示 3 步；step 1 选文件；点「下一步」进 step 2 |
| H49 | VideoUpload step 2 → step 1 | 已填基础信息保留 |
| H50 | VideoUpload step 3 「开始上传」 | 上传进度显示；切换回 step 1/2 后再回 step 3 时进度与历史结果保留 |

## 7. 实施前提清单

- Phase 1 / 2 / 3 三件套全部 Done
- 三巨头的接口/store 接受 drawer 替代 dialog 调用方式不变
- Phase 3 的 TvSeriesManage / ImageCollectionManage 已经验证 drawer 模式可用

## 8. 自动化必跑（Phase 4 阶段）

```bash
cd admin-web
npm run build
npm test
```

Phase 4 拟新增的 spec（待详化）：

- `videoList.helpers.spec.js` 扩展：列设置序列化 / 反序列化
- `videoUpload.wizard.spec.js`：3 步 wizard 的 step 切换 + 已填字段保留逻辑
- `imageManage.helpers.spec.js` 扩展：视图模式 localStorage 序列化

## 9. 待 Phase 3 验收后细化的内容

- Implement.md（三巨头详细文件改动 + Step 子组件骨架 + drawer 内 SectionCard 划分）
- review.md（H41–H50 详细期望 + 截图归档清单）
- 三巨头的 store / API 适配点（如果有）

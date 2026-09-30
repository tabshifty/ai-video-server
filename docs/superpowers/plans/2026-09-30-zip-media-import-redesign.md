# 压缩包视频图片导入重构实施计划

**目标：** 在现有管理端压缩包导入工作台上实现按目录分组、视频图片关联、合集多对多关系和确认后逐项入库。

**架构：** 保留现有压缩包批次、文件清单、解包和去重流程；新增正式媒体关系表与导入暂存关系表。导入组由目录自动生成，组内候选关系先保存到暂存表，确认入库时将暂存关系和合集继承解析为正式关系。

**技术栈：** Go、Gin、pgx/PostgreSQL、Vue 3、Element Plus、现有管理端 API 和 Vitest/Go 测试。

## 全局约束

- 所有 Markdown 和管理端用户文案使用中文并保持 UTF-8。
- 普通上传和历史压缩包接口保持兼容。
- 每项关系写入使用唯一键和幂等插入，重复确认不得创建重复媒体或关系。
- 代码变更完成后更新 `CONTEXT.md` 与 `plan.md`，并提交到 Git。

### 任务一：建立正式关系与导入暂存表

**文件：**
- 新增：`migrations/0040_archive_media_relations.up.sql`
- 新增：`migrations/0040_archive_media_relations.down.sql`
- 修改：`internal/repository/migrations_test.go`

**接口：**
- 提供 `video_image_collections`、`video_collection_image_collections`、`video_images` 和 `archive_import_file_images` 四张关系表。
- 迁移已有 `videos.image_collection_id` 到 `video_image_collections`。

- [ ] 写迁移 SQL 结构测试，锁定外键、唯一键、级联删除和旧字段回填。
- [ ] 运行 `go test ./internal/repository -run Migration -count=1` 确认测试先失败。
- [ ] 实现 up/down 迁移。
- [ ] 运行迁移测试和 `git diff --check`。
- [ ] 提交数据库结构变更。

### 任务二：扩展模型与视频仓储关系 API

**文件：**
- 修改：`internal/models/models.go`、`internal/models/admin.go`
- 修改：`internal/repository/video_repository.go`
- 修改：`internal/repository/image_collection_repository.go`
- 新增：`internal/repository/media_relation_repository.go`
- 测试：对应仓储测试文件

**接口：**
- `ReplaceVideoImageCollections(ctx, videoID, collectionIDs)`。
- `ReplaceVideoCollectionImageCollections(ctx, videoCollectionID, imageCollectionIDs)`。
- `ReplaceVideoImages(ctx, videoID, imageIDs, source)`。
- `ListVideoImageCollections(ctx, videoID)` 与 `ListVideoImages(ctx, videoID)`。

- [ ] 先写多对多去重、清空和重复插入测试。
- [ ] 实现关系仓储，所有写入使用事务、`ON CONFLICT DO NOTHING` 和外键校验。
- [ ] 普通视频详情读取优先新关系，旧字段为空时保持空结果。
- [ ] 运行 `go test ./internal/repository ./internal/models -count=1`。
- [ ] 提交关系仓储变更。

### 任务三：按目录自动生成导入组和视频图片候选关系

**文件：**
- 修改：`internal/services/archive_import.go`
- 修改：`internal/services/archive_import_groups.go`
- 修改：`internal/services/archive_import_test.go`
- 修改：`internal/models/models.go`

**接口：**
- `archiveImportGroupKey(relativePath)` 返回媒体所在规范化目录。
- `buildArchiveImportRelations(files)` 为同目录每个视频生成全部图片候选关系。
- `ListBatch` 返回组内视频、图片和候选关系。

- [ ] 写根目录、多层目录、同组多视频多图片和父子目录隔离测试。
- [ ] 扫描完成后自动创建混合媒体导入组，并把文件的 `group_id` 写入组。
- [ ] 扫描完成后写入 `archive_import_file_images`，重复重扫先清理当前批次暂存关系。
- [ ] 放宽组的 `media_kind` 限制为混合组，保留组级视频字段只对视频生效。
- [ ] 运行归档相关 Go 测试。
- [ ] 提交自动分组和候选关系变更。

### 任务四：实现导入关系编辑、合集继承和逐项入库

**文件：**
- 修改：`internal/services/archive_import.go`
- 修改：`internal/services/archive_import_batch_update.go`
- 修改：`internal/handlers/admin_archive_import.go`
- 修改：`internal/handlers/router.go`
- 新增或修改：服务端测试

**接口：**
- `UpdateArchiveImportFileImages(ctx, videoFileID, imageFileIDs)`。
- `ResolveArchiveVideoImageCollections(ctx, videoCollectionIDs, directImageCollectionIDs)`。
- `ConfirmArchiveImportBatch(ctx, batchID)`。
- `RetryArchiveImportFiles(ctx, batchID, fileIDs)`。

- [ ] 先写跨批次引用、非图片引用、视频合集继承和确认幂等测试。
- [ ] 扩展文件更新请求，区分视频展示图片合集与图片自身入库合集，视频字段支持多选。
- [ ] 入库前锁定批次；视频和图片写入后，以事务保存合集关系和暂存媒体关系。
- [ ] 处理队列入队失败时保留已创建媒体 ID，并把条目置为可重试失败。
- [ ] 增加保存关系、确认批次和重试失败项的管理员路由。
- [ ] 运行 `go test ./internal/services ./internal/handlers -count=1`。
- [ ] 提交服务端导入流程变更。

### 任务五：管理端工作台支持文件系统树、多选和视频图片关系编辑

**文件：**
- 修改：`admin-web/src/api/admin.js`
- 修改：`admin-web/src/views/ToolboxArchiveImport.vue`
- 修改：`admin-web/src/views/ToolboxArchiveImport.spec.js`
- 修改：`admin-web/src/api/admin.spec.js`

**接口：**
- 新增保存视频图片候选关系的 API 调用。
- 新增视频合集与图片合集关联管理入口。

- [ ] 先写 API 调用断言和页面源码行为测试。
- [ ] 将现有平面导入组列表转换为目录树数据，按路径段构造父子节点，使用连接线和叶子图标展示视频、图片及状态。
- [ ] 将视频关联图片合集选择器改为多选，图片入库合集保持独立多选。
- [ ] 在目录节点展示视频/图片计数，选中视频叶子后展示同目录图片并提供全选、取消和单张编辑。
- [ ] 确认入库前显示待处理统计；确认后锁定编辑并显示逐项失败和重试。
- [ ] 运行 `npm test -- --run src/api/admin.spec.js src/views/ToolboxArchiveImport.spec.js`。
- [ ] 运行 `npm run build`。
- [ ] 提交管理端变更。

### 任务六：兼容验证、文档沉淀和最终检查

**文件：**
- 修改：`CONTEXT.md`
- 修改：`plan.md`
- 可能修改：迁移测试、服务端和管理端回归测试

- [ ] 运行 `go test ./...`，记录与本次无关的既有失败。
- [ ] 运行管理端完整测试和构建。
- [ ] 执行 `git diff --check`、Markdown UTF-8 乱码扫描和工作树审计。
- [ ] 在 `CONTEXT.md` 记录导入组、关系继承和兼容字段的长期约定。
- [ ] 在 `plan.md` 追加实现和验证结果。
- [ ] 提交最终变更。

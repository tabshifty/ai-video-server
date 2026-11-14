# 已存在视频的合集关系追加

## 背景

管理端有四个会“带合集入库”的入口：上传页的分片上传与秒传、压缩包经典批次、压缩包工作台、115 云下载导入。它们在 SHA-256 与文件大小命中已有视频时都会跳过创建，但合集关系是否补写并不一致：

- 上传页 `VideoUpload.vue:350-353` 在 `/upload/check` 命中后直接返回「秒传命中」，不再调用任何写接口，合集关系完全没有机会落库；`/upload` 与 `/upload/complete` 的命中分支（`internal/services/upload.go:143-174`）虽然会追加，但因为入口层只允许 short 携带合集，实际覆盖面很窄。
- 上传并发竞态分支（`internal/services/upload.go:255-284`）在 `file_hashes` 唯一键冲突时会删掉新建视频（连带刚写入的合集），返回既有视频时不补写。
- 压缩包经典批次与工作台主链路会追加（`archive_import.go:871`、`upload.go:154`、`applyArchiveGroupRelations`），但文件状态一旦写成 `existing`，`processArchiveFile` 在 `archive_import.go:724-726` 直接早退，事后修改合集配置或重跑“入库”都不会补到视频上；工作台还要求 `GroupRelations=true`（`layout_version>=2`）才写关系，v1 工作区即使已链接视频也不写。
- 115 云导入命中后走 `finishSkippedExisting`（`cloud_download_import_service.go:620-626`），只写 `skipped_existing` 状态，没有任何合集写入。

结果是管理员明明在入库时选了合集，界面却只回「已存在/复用已有」，需要事后逐个进视频详情手工补，且秒传路径连事后补的提示都没有。领域侧 `CONTEXT.md` 早有 `压缩包去重保留现有资产` 一条，但只覆盖压缩包，未覆盖秒传与 115。

## 决策

- **命中即追加，不覆盖**：命中已有视频时把本次合集关系并入（`INSERT ... ON CONFLICT DO NOTHING`），保留其原有归属；不删除、不替换任何既有关系。重复入库天然幂等。
- **只补关系，不改元数据**：不同步本次填写的标题、说明、标签、演员，沿用 `CONTEXT.md` 中“复用已有媒体并保留元数据与历史关系”的既有语义，避免覆盖复用资产的呈现。
- **图片合集不覆盖**：已存在视频已有图片合集时不改写，没有时补上；视频合集与图片合集的配对关系继续沿用压缩包工作台的既有规则。
- **类型边界不动**：视频合集仍等于短视频合集。上传页与分片初始化继续 short-only 校验；压缩包对电影/电视剧/AV 的既有 any-type 追加能力保留，公共追加方法不带类型守卫，类型约束留在入口层。本次既不放开也不回收，避免让在用的压缩包批次突然失败。
- **类型冲突仍然报错**：哈希命中但本次 `type` 与既有视频不同，继续返回 409 `hash exists with another video type`，不做关联；它说明同一份文件被登记成不同资产语义，静默补关系会掩盖数据问题。
- **待删除视频不追加**：`status='pending_delete'` 的视频不接收新关系（`video_collections` 随视频级联删除，写了也会消失），按「已存在（该视频待删除），未关联合集」反馈；转码 `failed` 等其它状态照常追加。
- **秒传走独立写接口**：`/api/v1/upload/check` 保持只读，命中后的追加走新增的 `POST /api/v1/upload/attach`，要求管理员身份并用 `hash + file_size` 重新反查视频，不信任客户端传入的视频 ID。
- **压缩包对已存在文件做关系对齐**：`existing` 状态不再整体冻结；`processArchiveFile` 早退前先按当前文件/分组配置幂等对齐一次关系，且不再依赖 `GroupRelations` 或 layout 版本。
- **失败可见、不回滚已写入关系**：追加失败使该文件或 115 导入记录进入失败态并携带中文原因；已经写入的关系不回滚，因为回滚可能误删其它入口写入的同一批关系，而重试本身幂等。

## 考虑过的替代方案

- **复用 `PUT /admin/videos/:id` 补合集**：该接口是整体替换语义且只允许 short，用它补关系会把视频原有的其它合集抹掉，还会被批量编辑的“整批都是 short”前置条件挡住；不采用。
- **扩展 `/upload/check` 让它在命中时写库**：省一次请求，但把只读探针变成写接口，后来的维护者会把它当成越权写入口；不采用。
- **取消秒传、让文件走完整分片上传**：`/upload/complete` 的命中分支已经会追加，改动最小，但等于牺牲秒传对大文件的核心价值；不采用。
- **命中时同步本次元数据**：能让复用资产带上“最新一次”的信息，但会覆盖历史资产标题与人工整理结果，且与既有“复用已有媒体并保留元数据”的约定冲突；不采用。
- **放开所有视频类型都能关联视频合集**：App 端合集目录与可见性查询硬编码 `v.type='short' AND v.status='ready'`（`internal/repository/collection_repository.go:382-390`、`:444-453`），放开写入也不会出现在手机端合集里，只是让数据更不一致；不采用。
- **新增“补充关联”独立按钮**：可以避开改动入库主流程，但把一次入库拆成两个动作，管理员心智与批量操作都会变复杂；不采用。

## 后果与风险

- 「秒传命中无任何副作用」的旧假设不再成立：命中后可能写 `video_collections`。只读性由接口边界保证（`check` 只读、`attach` 显式写入），不靠调用顺序约定。
- 电影/电视剧/AV 的 `video_collections` 仍然存在且在管理端视频详情可见，但 App 端合集不展示——这是本次之前就有的边界，本次既不扩大也不清理，记录在此避免后来者误判为新缺陷。
- 失败不回滚意味着“部分成功”是允许态，界面必须如实显示新增数，不能一律报成功。
- 追加需要覆盖并发竞态分支（`InsertFileHash` 冲突）与压缩包 `existing` 早退两条旁路，实现上必须收敛到同一个服务方法，否则会再次分叉。
- 验证需要真实 Postgres 覆盖幂等与并发：纯内存单测最容易在这里假绿。

## 相关

- `CONTEXT.md` 中 `已存在视频合集关系追加` 术语与接口约定
- `internal/services/upload.go`、`internal/services/archive_import.go`、`internal/services/archive_workspace.go`、`internal/services/cloud_download_import_service.go`
- `internal/repository/collection_repository.go`、`internal/repository/media_relation_repository.go`
- `admin-web/src/views/VideoUpload.vue`、`admin-web/src/views/ToolboxArchiveImport.vue`
- `docs/adr/0033-archive-group-multiple-memberships.md`

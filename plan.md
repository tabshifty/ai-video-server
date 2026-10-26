# plan.md

## 2026-10-26 13:44 +0800 访谈第四轮记录：确认主分组术语与成功成员保护
- 进度：已写入主分组等同上传配置、成功成员不可换组或排除、含成功成员分组树不可删除，以及新入库仅通过图片合集关联的确认结果；第四轮建议明确标记为待确认，未将术语澄清扩大为根节点和成员规则已获确认。仅修改文档，未进入功能实现。
- 影响文件：仅四份本任务文档；既有未跟踪 `docs/examples/` 保留。
- 验证：`git diff --check`、四份文档严格 UTF-8 与替换字符检查通过；纯文档未运行构建或业务测试。

## 2026-10-26 13:43 +0800 访谈第四轮计划：主分组定义与最终执行边界
- 进度：用户明确主分组就是上传时填写的内容，并确认 Q22 成功文件归属冻结、含成功成员的分组及祖先禁止删除，Q23 新流程仅通过图片合集关联并保留历史直接图片关系。记录已确认定义，继续确认主分组结构、根目录与未分组回退、缺省配置及执行和编辑界面边界。
- 影响文件：`docs/design/2026-10-26-archive-import-workspace.md`、`docs/adr/0032-archive-import-virtual-workspace.md`、`CONTEXT.md`、`plan.md`；仅文档。
- 验证：待执行文档差异、UTF-8 和替换字符检查；既有 `docs/examples/` 不纳入。

## 2026-10-26 11:51 +0800 访谈第三轮记录：第二轮十一项决策已固化
- 进度：已记录 Q10～Q20 确认结果，补充目录建组、集合覆盖、虚拟文件名标题、成功冻结、重复媒体复用、现场新建合集与草稿冲突术语，更新 ADR 中草稿与正式媒体边界。下一轮只确认主分组、含成功成员操作和直接图片关联；依赖主分组的细节留到定义明确后再定，尚未进入实现。
- 影响文件：仅四份本任务文档；既有未跟踪 `docs/examples/` 不纳入提交。
- 验证：`git diff --check`、四份文档严格 UTF-8 与替换字符检查通过；纯文档未执行构建或业务测试。

## 2026-10-26 11:49 +0800 访谈第三轮计划：固化第二轮规则并收口主分组
- 进度：用户确认第二轮 Q10～Q20 全部按推荐，记录目录建组与递归操作、删组与继承、标题、新建合集、执行冻结与重试、重复媒体追加关联、保存冲突、旧批次和 App 展示范围。继续确认主分组定义、包含已入库成员的操作边界及直接视频图片关联；依赖主分组的根目录与顶层默认值暂不预设。
- 影响文件：`docs/design/2026-10-26-archive-import-workspace.md`、`docs/adr/0032-archive-import-virtual-workspace.md`、`CONTEXT.md`、`plan.md`；仅文档。
- 验证：待执行差异、UTF-8 和替换字符检查；既有 `docs/examples/` 不纳入。

## 2026-10-26 11:41 +0800 访谈第二轮记录：确认领域边界并核实关联消费
- 进度：已将首轮九项决策记入需求文档与领域语言，建立 ADR 0032 说明虚拟目录、分组、物理文件和显式入库的边界；只读核实 App 仍展示单个图片合集、重复视频关联可能全量替换、重解压后端保护缺口及图片字段差异，供第二轮确认范围和行为。访谈尚未收口，业务代码未修改。
- 影响文件：仅本轮四份文档；既有未跟踪 `docs/examples/` 保留。
- 验证：`git diff --check`、四份文档严格 UTF-8 解码和替换字符检查通过；纯文档未运行构建或业务测试。

## 2026-10-26 11:38 +0800 访谈第二轮计划：记录首轮确认并展开边界
- 进度：用户以“按推荐”确认首轮九项建议；同步记录独立页面、虚拟文件操作、单组归属与混合媒体、两棵独立树、逐字段继承、合集多选、标题模式、显式入库和批次自动保存。继续梳理按目录建组、目录拖拽、排除恢复、删组、合集创建和提交恢复的具体边界，主分组仍待用户后续说明。
- 影响文件：`docs/design/2026-10-26-archive-import-workspace.md`、`docs/adr/0032-archive-import-virtual-workspace.md`、`CONTEXT.md`、`plan.md`；仅文档。
- 验证：待执行差异、UTF-8 和替换字符检查；已有 `docs/examples/` 不纳入。

## 2026-10-26 11:20 +0800 访谈首轮记录：需求边界与现状差异
- 进度：记录用户已明确要求的全屏、虚拟文件整理与入库分组术语，列出首轮九项待决问题。核实当前分组平级、按目录自动建组、混合媒体与图片合集多选存在不同路径行为不一致；尚未完成需求收口，不修改业务代码，不将建议写成已接受的决策。
- 影响文件：仅提交 `docs/design/2026-10-26-archive-import-workspace.md`、`CONTEXT.md` 和本文件的访谈记录；既有未跟踪 `docs/examples/` 保留。
- 验证：`git diff --check`、三份文档严格 UTF-8 解码与替换字符检查通过；纯文档未运行构建或业务测试。

## 2026-10-26 11:18 +0800 需求访谈：压缩包全屏入库工作台
- 进度：使用用户指定的 grill-with-docs，结合 grilling 与 domain-modeling 梳理现有流程；先记录用户明确要求的领域术语与待决问题，分轮确认关系、继承、合集和执行边界。当前处于需求访谈阶段，尚未确认实施方案。
- 影响文件：`docs/design/2026-10-26-archive-import-workspace.md`、`CONTEXT.md`、`plan.md`；业务代码暂不修改。
- 验证：待执行文档差异、UTF-8 与乱码检查；纯文档无需构建。已有未跟踪 `docs/examples/` 不纳入本任务。

## 2026-10-23 09:30 +0800 完成：压缩包解压兼容与失败恢复
- 进度：完成 RAR 头部错误回退、原生工具命令修正、解压失败清理、密码和取消边界及中文诊断；复查差异确认未改 API、数据库或 App。仅提交本任务 5 个文件，保留既有未跟踪 `docs/examples/`。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_extract_test.go`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`。
- 验证：14 组工具链恢复场景及头部诊断、文件名误判、取消用例通过；`go test ./internal/services -run 'Test.*Archive|TestExtract.*|TestZip.*' -race -count=1`、`go test ./internal/handlers -run 'Test.*Archive' -race -parallel 1 -count=1`、`go test ./... -skip '^TestParseTVAPKMetadataParsesReleaseAPK$' -count=1`、`go vet ./...`、`go build ./...`、`git diff --check` 及本次 5 文件严格 UTF-8 和替换字符检查均通过。原生 unrar 对普通和加密中文文本归档实测解压并逐字节比对通过。
- 验证边界：未跳过时全量测试存在既有 APK 版本断言失败（144 vs 121）；handlers 默认并行 race 存在既有 Gin.SetMode 全局变量竞争，串行复验通过；未安装 golangci-lint，仓库未提供对应配置，未执行额外 lint。未部署、未复验原始问题归档，部署后需要对失败批次手动重试，实际恢复取决于原归档完整性与兼容工具是否可用。

## 2026-10-23 09:28 +0800 实施：压缩包解压兼容与失败恢复
- 进度：新增 RAR 头部错误兼容回退，修正 unrar 参数并补齐无交互密码参数；兼容工具逐次清理半成品，最终失败保留合并诊断且删除输出；区分 7z、CRC、密码和取消，头部错误按诊断行识别而非文件名匹配。原始问题归档不在开发机，不执行现场内容操作。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_extract_test.go`、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`。
- 验证：新用例先在旧实现失败；首轮定向测试、services 解压 race 检查、build/vet 通过；RAR 官方 7.30 beta 1 工具生成普通和加密的中文文件名文本归档，两种 unrar 命令实际解压并逐字节比对通过。全量测试仅遇既有 TV APK fixture 版本断言（144 vs 121）；扩展 handlers race 检查发现既有并行测试反复写 Gin 全局模式，准备串行复验并记录范围，避免修改无关模块。最终代码的补充密码和诊断边界测试与复验运行中。

## 2026-10-23 计划：压缩包解压兼容与失败恢复
- 进度：确认 RAR 头部错误缺少兼容回退、unrar 参数顺序错误、备用工具间半成品未隔离；计划对 RAR 的 Headers Error/Unsupported Method 尝试原生 unrar 与 unar，保留密码/CRC/取消边界，修正无交互命令参数并清理解压失败残留，收敛头部错误中文诊断。
- 影响文件：`internal/services/archive_import.go`、解压回归测试、`docs/家用部署机.md`、`CONTEXT.md`、`plan.md`。
- 验证：先补工具链模拟回归测试，之后运行定向测试（含 race）、后端测试、构建与 vet；当前开发机不存在报错归档，不能据日志断定真实文件损坏或宣称真实归档已恢复。

## 2026-10-23 00:35 +0800 修复：115 云下载扫码登录取码
- 进度：登录二维码改为复用 115Browser 默认常开的 `https://115.com/` 标签页，不再打开已返回 404 的 `/login`；从 `.qrcode-login .pic img`（`#js_login_qrcode_img`）读取 data URL，二维码过期（“二维码失效/请点击刷新”）时先点击 `.ql-refresh` 再取图；非 data URL 时按二维码元素 rect 调 `Page.captureScreenshot` 裁剪，不再用整页截图兜底。`loginTarget` 会等待并优先选择无 query 的 115 首页，只有完全不存在 115 页面时才新开标签页。
- 影响文件：`internal/services/cloud_download_service.go`、`internal/services/cloud_download_service_test.go`、`internal/cloud/browser.go`、`CONTEXT.md`、`plan.md`。
- 验证：`go build ./...`；`go test ./internal/services -skip '^TestParseTVAPKMetadataParsesReleaseAPK$'`、`go test ./internal/cloud ./internal/handlers` 通过；真实 115Browser（9333）实测：复用默认 `https://115.com/` tab，约 0.41s 返回 334×334 GIF data URL，前后目标列表一致（未新增 tab）；新增 `preferredLoginTarget`/`loginPageScore` 单测 4 项通过。历史 APK fixture 失败（version_code 144 vs 121）与本次无关。

## 2026-10-22 17:27 +0800 计划：云下载状态列统一 Tag 展示
- 进度：状态列将任务状态、需处理原因和 115 回收站状态统一以 Tag 呈现，并用间距与自动换行区分多个状态。
- 影响文件：`admin-web/src/views/CloudDownloadList.vue`、`admin-web/src/views/cloudDownloadList.spec.js`、`plan.md`。
- 验证：新增断言先失败后通过；定向 Vitest 14 项通过；`npm run build` 通过（存在既有 chunk 大小提示）；`git diff --check` 通过；本次文件未发现乱码。

## 2026-10-22 00:05 +0800 实施完成：115 云盘源文件手动移入回收站
- 进度：新增 migration `0046_cloud_download_file_deletions`（task 增加 cloud_file_id/cloud_parent_id/cloud_user_id/cloud_file_category，清理记录 task_id 唯一，状态 deleting/deleted/failed）；新增 `cloud.DeleteFile` 使用已登录 115 页面的 `POST https://webapi.115.com/rb/delete`，严格校验 state；云下载任务 ready 时保存 `delete_file_id`/`wp_path_id`/账号/文件类型；新增同步管理接口 `POST /admin/cloud-downloads/:id/cloud-file/delete`（必须 confirm=true），独立 service/记录摘要；历史任务按 info_hash 有界分页唯一匹配、文件名/大小/普通文件校验；本地源或导入终态放行、活动导入/缺失本地拒绝；目录/非普通文件拒绝；账号不一致拒绝；并发 claim 409；5 分钟 stale 可重试；终态写库用 WithoutCancel + attempt CAS；前端“更多 → 移入 115 回收站”确认框、状态标签、导入终态/超时重试、登录错误打开二维码。
- 评审采纳：只使用 delete_file_id 作 fid[0]、file_id 不当删除目标；新增 cloud_user_id/file_category 守卫；Start 在浏览器前落 failed 记录；Mark 返回 rowsAffected 并重读权威状态；115 文件不存在按幂等 deleted；发现 after-import 本地源已删后允许以导入终态放行。
- 影响文件：`migrations/0046_*`、`internal/cloud/protocol.go`、`internal/models/cloud_download*.go`、`internal/repository/cloud_download*repository.go`、`internal/services/cloud_file_deletion_service.go`、`internal/services/cloud_download_service.go`、`internal/handlers/admin_cloud_file_deletion.go`、`router.go`、`main.go`、`admin-web/src/{api/admin.js,views/CloudDownloadList.vue,views/cloudDownloadFileDeletion.helpers.js}` 及测试、`docs/adr/0031-manual-115-cloud-source-cleanup.md`、`CONTEXT.md`、`docs/家用部署机.md`。
- 验证：`go build ./...`；`go test ./... -skip '^TestParseTVAPKMetadataParsesReleaseAPK$'` 全绿（含新增 deletion service/handler/migration tests）；前端 52 文件 780 测试通过；`npm run build` 通过；受控测试任务 10 已真实调用 115 `/rb/delete` 返回 state=true 并移入回收站，本地/任务记录未改。
- 残余风险：115 删除是私有 web 接口，未覆盖全部 errno/账号切换/大量请求；历史任务无创建时 115 账号快照，保留有界懒解析的理论跨账号风险；无独立回收站只读核对，ambiguous 超时按 failed 允许手动重试；未部署、未做真实端到端 UI 点击。

## 2026-10-18 15:15 +0800 修复：压缩包批次批量操作条不显示
- 进度：`BulkActionBar` 原先放在批次详情抽屉正文末尾，长文件清单下虽为 sticky 但初始在视口外，勾选后看不到批量入口。现改为 `el-drawer` 的 `footer` 插槽，选中 ≥1 项时固定在抽屉底部立即显示；删除失效的 `.archive-drawer :deep(.bulk-action-bar)` 移动端规则，保留共享组件自身的响应式布局。更新抽屉/批量条测试断言。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`ToolboxArchiveImport.spec.js`、`CONTEXT.md`、`plan.md`。
- 验证：`npm test -- --run` 51 文件 773 测试通过；`npm run build` 通过；待提交并推送 `deploy`（仅前端重建）。

## 2026-10-18 15:00 +0800 前端修复：压缩包导入弹窗支持新建合集
- 进度：压缩包入媒体库弹窗的“默认视频合集/默认图片合集”增加“新建视频合集/新建图片合集”按钮；复用管理端 collection/image-collection 创建 API，弹出名称+简介小对话框，创建成功后把新合集合并进候选并自动加入当前默认合集数组，随导入表单提交。视频弹窗未改。
- 影响文件：`admin-web/src/components/CloudDownloadImportDialog.vue`、`CloudDownloadImportDialog.spec.js`、`CONTEXT.md`、`plan.md`。
- 验证：`npm test -- --run` 51 文件 773 测试通过；`npm run build` 通过；`git diff --check` 干净。
- 收尾：提交并推送 `deploy`，触发前端重建（Go 不在变更桶，不重启 server/worker）。

## 2026-10-18 01:21 +0800 实施完成/评审修复：云下载任务显式导入媒体库
- 进度：后端、前端、迁移、文档已完成，独立评审的阻塞/重要问题已修复并复验。新增 `cloud:ingest`（`cloud` 队列，2h/2 retry，TaskID 含 import id + attempt）与 `cloud_download_imports`（task_id 唯一，状态 queued/hashing/copying/creating/imported/skipped_existing/failed，UUID FK）；服务端 preview/import，任务列表 import 摘要；视频 SHA-256/size 查重、hash 命中（含 pending_delete）删下载源不改已有视频、未命中复制到 `STORAGE_ROOT/videos/<videoID>/source-original`、建 videos 入现有队列、转码成功删导入源（DV `source-dv` 播放产物保留）；压缩包复制到 `archive-imports/<batchID>/original`、异步解包、工具箱处理，批次删除不触碰 DOWNLOAD_DIR；修 `RetryExtract` encoding_mode。
- 评审修复：`Process`/`MarkFailed` 校验 payload attempt，旧 asynq 重试不再覆盖新尝试；活动导入 10 分钟后 UI 允许“重新导入”，后端用同一 attempt/TaskID 幂等补投（队列丢失可自救，表单需重填）；视频 `SaveImportedFile` 成功但 `video_id` 未回写时按 metadata `source=cloud_download`+`cloud_task_id` 走 resume，不误判 skipped；压缩包重试只有 uploaded/failed 才 RetryExtract，已入队/已解包批次只关联完成；补 copy 父目录、ready 孤儿源清理、target 缺失时非 ready/processing 失败保源、sha256 持久化。
- 影响文件：`migrations/0045_cloud_download_imports.{up,down}.sql`、`internal/models/cloud_download_import.go`、`internal/repository/cloud_download_import_repository.go`、`internal/services/cloud_download_import_service.go`、`internal/queue/cloud_ingest.go`、`internal/handlers/admin_cloud_download_import.go`、`internal/handlers/router.go`、`internal/queue/tasks.go`、`internal/services/{upload,archive_import,cloud_download_service}.go`、`main.go`、`admin-web/src/{api/admin.js,views/CloudDownloadList.vue,views/cloudDownloadImport.helpers.js,components/CloudDownloadImportDialog.vue}` 及测试、`CONTEXT.md`、`docs/adr/0028`、`docs/adr/0030`、`docs/家用部署机.md`。
- 验证：`go build ./...`；`go test ./... -skip 'TestParseTVAPKMetadataParsesReleaseAPK'`（全绿；历史 APK 测试仍失败，已知）；前端 `npm test -- --run` 51 文件 772 测试通过，`npm run build` 成功；`gofmt`/`git diff --check` 干净。未跑真实 115/ffprobe 端到端，未部署。
- 残余风险：无独立 import reconciler（活动记录靠 10 分钟重新提交自救）；`SetArchiveBatch` 失败时批次已创建但未关联，重试可能产生第二个批次；弹窗未加“新建合集/图片合集”入口；本次收尾提交包含以上改动，未部署。

## 2026-10-18 00:55 +0800 计划/实施：云下载任务显式导入媒体库
- 进度：grilling 已收口，开始实现。入口只在 CloudDownloadList 的 completed 单文件；点击“入媒体库”后端探测实际文件，视频/压缩包分流；异步 `cloud:ingest` + `cloud_download_imports`（task_id 唯一，状态 queued/hashing/copying/creating/imported/skipped_existing/failed，2h/2 retry）。视频先 SHA-256 + size，hash 命中不判断类型、关联已有视频并直接删除下载原文件（含 pending_delete）；未命中复制到 `STORAGE_ROOT/videos/<videoID>/source-original`，建 videos 并入现有转码/刮削，转码成功删除导入源文件（DV 的 `source-dv` 是播放产物，保留；original_path 若已 move 成播放路径则跳过），失败保留。压缩包新增本地文件入口，复制到 `STORAGE_ROOT/archive-imports/<batchID>/original`，异步解包，工具箱继续处理，永不把 DOWNLOAD_DIR 写进批次 original_path。一次性/失败重试；不自动、不等待转码/解压；不做多文件/目录/ISO；顺手修 RetryExtract encoding_mode。
- 影响文件：迁移 0045、models/repository/services/queue/handlers/main、admin-web CloudDownloadList/两个弹窗/admin.js/tests、CONTEXT/ADR 0028/0030/部署文档。
- 验证：待后端/前端完成，执行 go build/定向测试、admin tests/build、独立评审。
- 备注：部署机尚未更新。

## 2026-10-17 23:58 +0800 评审修复：无效结果映射、确认框无效计数与并发重试
- 进度：第二轮独立评审提出两个阻塞项并已修复。无效项只按 info_hash/link 单键索引，导致前端用 hash 查不到后端 invalid 结果、不显示“无效”也不保持勾选；现 `indexCloudBatchItems` 同时登记 info_hash、原始 link 和 link 解析出的 hash。确认框缺少“无效”且无 hash 项被计为“新提交”；现 `summarizeEd2kBatchSelection` 分离无 hash 的 invalid，确认框显示无效/重复/将提交/已存在/重试，并列出被忽略的非 ED2K 行，invalid 不再计入新提交。
- 重要项：`BumpRetry` 改为 `WHERE status IN ('failed','blocked')` 并返回 RowsAffected；并发重试已被抢占时 `Submit`/`Retry` 不再二次入队。部分失败/无效的结果提示从 success 改为 warning。`CLOUD_DOWNLOAD_QUEUE` 增加 asynq 队列名字符校验。修正 handler 路由注释。补 50 条接受/503、50 条 service、状态查询失败、source 不覆盖、并发抢占、invalid 映射与确认计数测试。
- 影响文件：`internal/repository/cloud_download_repository.go`、`internal/services/cloud_download_service{,_test}.go`、`internal/handlers/admin_cloud_download{,_test}.go`、`internal/config/config{,_cloud_download_test}.go`、`admin-web/src/views/ToolboxEd2k{,.spec}.vue`、`admin-web/src/views/toolbox.helpers{,.spec}.js`、`plan.md`。
- 验证：`go build ./...`、`go vet` 目标包、`go test ./... -skip 'TestParseTVAPKMetadataParsesReleaseAPK'` 全部通过；admin-web 739 项测试与 `npm run build` 通过；`git diff --check` 通过。

## 2026-10-17 23:53 +0800 评审修复：入队失败语义与超 50 全选
- 进度：独立评审确认无阻塞，但指出两个重要项并已修复。其一，`Submit` 在 enqueue 失败时行仍为 `pending`，reconciler 会继续跑，而单条/批量却报失败；现与 `Retry` 一致，enqueue 失败调用 `s.fail` 落为 `failed` 并返回错误。其二，工具箱有效链接超过 50 时“全选→发送”必然超限；现全选只选前 50 条，达到 50 后其余行复选框禁用并提示“取消一条后可改选其他链接”。
- 影响文件：`internal/services/cloud_download_service.go`、`internal/services/cloud_download_service_test.go`、`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/ToolboxEd2k.spec.js`、`plan.md`。
- 验证：service 定向测试（含新增 `TestSubmitMarksFailedWhenEnqueueFails`）通过；admin-web 737 项测试与 `npm run build` 通过。

## 2026-10-17 23:51 +0800 完成：工具箱批量 115 提交与论坛按钮样式
- 进度：批量接口、source、cloud 队列、工具箱勾选/全选/确认/逐条状态、论坛按钮样式、任务来源列、ADR/CONTEXT/部署文档均已完成。论坛右侧操作列按用户要求延期；未执行部署。
- 评审修复：独立方案子代理在实现前后审阅 WIP，指出并已修复——handler 先过滤空项导致响应不 1:1；批量忽略 `created` 导致并发创建误报 submitted；source 仓库兜底 manual 与 DB 默认不一致；队列名未 trim；关闭功能时路由 404 而非 503；invalid 项未保留勾选；输入变化后选中集残留；窄屏操作目标不足 44px。
- 影响文件：`migrations/0044_cloud_download_source.{up,down}.sql`、`internal/models/cloud_download.go`、`internal/repository/cloud_download_repository.go`、`internal/services/cloud_download_service{,_test}.go`、`internal/handlers/admin_cloud_download{,_test}.go`、`internal/handlers/router.go`、`internal/queue/{tasks,cloud_download,cloud_download_test}.go`、`internal/config/{config,config_test,config_cloud_download_test}.go`、`main.go`、`cmd/telegram-ingestor/main.go`、`admin-web/src/api/{admin,admin.spec}.js`、`admin-web/src/views/{ToolboxEd2k,ToolboxEd2k.spec,toolbox.helpers,toolbox.helpers.spec,ForumPostList,ForumPostList.spec,CloudDownloadList,cloudDownloadList.spec,precisionOpsRollout.spec}.{vue,js}`、`CONTEXT.md`、`docs/adr/0028-...md`、`docs/adr/0029-...md`、`docs/家用部署机.md`、`.env.example`、`plan.md`。
- 验证：`go build ./...`、`go vet` 目标包、`go test ./... -skip 'TestParseTVAPKMetadataParsesReleaseAPK' -count=1` 全部通过；admin-web `npm test -- --run` 737 项通过、`npm run build` 通过；`git diff --check` 与乱码扫描通过。历史失败 `TestParseTVAPKMetadataParsesReleaseAPK`（version_code 144 vs 121）未触碰。
- 备注：论坛操作列/`>2 条弹层` 留待后续；部署机尚未更新到本提交。

## 2026-10-17 23:48 +0800 实施：工具箱批量 115 提交与 cloud 队列
- 进度：已落地 migration 0044（`cloud_download_tasks.source`，默认 `forum_post`，check 三值）、任务来源读写与“来源”列、批量接口 `POST /admin/cloud-downloads/batch`（handler/service 双重 50 条上限、空项保留为 invalid、逐条结果、completed/进行中跳过、failed/blocked 重试、`created`/状态竞态分类）、独立 `cloud` 队列（权重 1，server/worker 配置一致，关闭功能时 worker 不消费）、工具箱勾选/全选/确认/逐条状态/失败保留勾选、论坛行内按钮改为普通 `plain` 按钮（不加操作列）、ADR 0029/CONTEXT/ADR 0028/部署文档/.env.example。
- 影响文件：见本轮 plan 计划条目；另新增 `internal/handlers/admin_cloud_download_test.go`、`migrations/0044_cloud_download_source.{up,down}.sql`、`docs/adr/0029-explicit-batch-cloud-download-from-ed2k-toolbox.md`。
- 验证：`go build ./...` 通过；`go vet` 目标包通过；`go test ./internal/handlers ./internal/config ./internal/queue ./internal/repository ./internal/cloud -count=1` 通过；service 定向测试（批量/Source/Process）通过；admin-web `npm test -- --run` 736 项通过，`npm run build` 通过；`git diff --check` 与乱码扫描通过。已知 `TestParseTVAPKMetadataParsesReleaseAPK` 为历史失败（version_code 144 vs 121），未在本轮触碰。
- 评审：独立子代理评审进行中；发现问题后修复并复审。

## 2026-10-17 23:34 +0800 计划：工具箱批量 115 提交与论坛按钮样式
- 进度：按 grilling 结论实施。工具箱 ED2K 生成器增加勾选/全选、单批最多 50 条、确认后显式批量提交；新增 `POST /admin/cloud-downloads/batch` 逐条返回结果；云下载任务改走独立 `cloud` 队列（权重 1），避免批量任务占满转码槽位；任务新增 `source`（`forum_post`/`toolbox`/`manual`）。论坛仍逐条，仅把行内 `el-button link` 改成普通 `size="small"` 按钮，重试保留原位；右侧操作列改造延后。新增 ADR 0029，更新 CONTEXT/部署文档。
- 影响文件：`migrations/0044_cloud_download_source.{up,down}.sql`、`internal/models/cloud_download.go`、`internal/repository/cloud_download_repository.go`、`internal/services/cloud_download_service.go`、`internal/handlers/admin_cloud_download.go`、`internal/queue/cloud_download.go`、`internal/queue/tasks.go`、`internal/config/config.go`、`main.go`、`admin-web/src/api/admin.js`、`admin-web/src/views/ToolboxEd2k.vue`、`admin-web/src/views/toolbox.helpers.js`、`admin-web/src/views/ForumPostList.vue`、`admin-web/src/views/CloudDownloadList.vue`、对应测试、`CONTEXT.md`、`docs/adr/0029-explicit-batch-cloud-download-from-ed2k-toolbox.md`、`docs/adr/0028-...md`、`docs/家用部署机.md`、`.env.example`、`plan.md`。
- 验证：待执行 `GOPROXY=off GOSUMDB=off go build ./...`、`go vet` 目标包、Go 定向测试、`admin-web npm test -- --run` 与 `npm run build`，并做 `git diff --check` 与乱码扫描。

## 2026-10-17 23:07 +0800 修复：cloud_download_tasks.post_id 改为 UUID
- 进度：部署 hook 在 `0042_cloud_download_tasks` 失败：`collected_forum_posts.id` 是 UUID，原 `post_id BIGINT` 无法建立外键（`incompatible types: bigint and uuid`）。已把 migration 0042 的 `post_id` 改为 `UUID`，`CloudDownloadTask.PostID`、service `Submit`、handler 请求体统一为 `*uuid.UUID`；新增 `TestCloudDownloadTasksMigration` 锁定列类型，新增 `TestSubmitKeepsForumPostUUID` 锁定提交链路。部署机确认未创建该表、未写入 0042/0043 版本记录，修正后可直接重放迁移。
- 影响文件：`migrations/0042_cloud_download_tasks.up.sql`、`internal/models/cloud_download.go`、`internal/services/cloud_download_service.go`、`internal/services/cloud_download_service_test.go`、`internal/handlers/admin_cloud_download.go`、`internal/repository/migrations_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：`GOPROXY=off GOSUMDB=off go build ./...` 通过；`go vet` 目标包通过；`go test ./internal/cloud ./internal/config ./internal/handlers ./internal/queue ./internal/repository -count=1` 通过；service 定向测试（16 项，含 UUID 提交）通过；admin-web 730 项测试与 `npm run build` 通过。本机 Go 默认 `GOPROXY` 会因访问 `proxy.golang.org` 卡住，验证统一加 `GOPROXY=off`。

## 2026-10-17 20:02 +0800 补充：重试后的 pending 任务也先查 115
- 进度：`pending` 且 `retry_count > 0` 的任务（手动重试或失败重置）在提交前同样调用 `FindTask`；115 侧已有该任务时直接进入等待，避免对已完成的旧任务重复提交。新增 `TestProcessSkipsSubmitWhenRetriedPendingTaskExistsOn115`。
- 影响文件：`internal/services/cloud_download_service.go`、`internal/services/cloud_download_service_test.go`、`plan.md`。
- 验证：`go test ./internal/services -run 'TestProcessResubmits|TestProcessSkipsSubmit|TestProcessWritesBlocked|TestProcessBlocksWhenQuota|TestProcessDoesNotBlock|TestProcessSkipsTrigger|TestRetry|TestSubmitResets|TestReconcile|TestMarkFailed|TestFindDownloaded|TestExpectedDownload' -count=1` 通过（15 项云下载状态机/落盘测试）；`go build ./...` 通过。

## 2026-10-17 19:58 +0800 完成：115 云下载失败恢复与幂等修复
- 进度：已按计划完成。① `submitting` 先查 115 是否已有该 info_hash，查到则直接等待、查不到才重新提交，消除一次 CDP/网络错误后的永久挂起；② `block`/`fail`/`MarkFailed` 使用 `context.WithoutCancel` + 10 秒写终态，asynq 最后一次重试失败时由 `HandleCloudDownload` 调用 `MarkFailed`；worker 启动及每 2 分钟对账非终态任务；③ `blocked` 在云下载页和论坛资源页均可直接重试；④ asynq 任务 ID 为 `cloud:download:<id>:<retry_count>`，重复入队返回 `ErrTaskIDConflict` 视为成功，新增 `download_triggered_at` 一次性触发标记与 migration 0043，手动重试清空标记；⑤ 落盘校验优先使用 115 返回大小，要求文件大小/mtime 稳定一个轮询周期，跳过常见临时后缀；⑥ 页面展开行展示错误信息/本地路径/pick_code/重试次数；⑦ 开启功能时校验 CDP URL 与端口，补齐 handler 404/409、docs/ADR/CONTEXT。
- 影响文件：`internal/cloud/browser.go`、`internal/cloud/protocol.go`、`internal/services/cloud_download_service.go`、`internal/services/cloud_download_service_test.go`、`internal/repository/cloud_download_repository.go`、`internal/models/cloud_download.go`、`internal/queue/cloud_download.go`、`internal/queue/cloud_download_test.go`、`internal/queue/tasks.go`、`internal/handlers/admin_cloud_download.go`、`internal/config/config.go`、`internal/config/config_test.go`、`internal/config/config_cloud_download_test.go`、`main.go`、`migrations/0043_cloud_download_triggered.{up,down}.sql`、`admin-web/src/views/CloudDownloadList.vue`、`admin-web/src/views/cloudDownloadList.spec.js`、`admin-web/src/views/ForumPostList.vue`、`admin-web/src/views/ForumPostList.spec.js`、`CONTEXT.md`、`docs/adr/0028-explicit-cloud-download-from-forum-resources.md`、`docs/家用部署机.md`、`.gitignore`、`go.mod`、`plan.md`。
- 验证：`GOCACHE=/tmp/ai-video-go-cache go build ./...` 通过；`go vet` 目标包通过；`go test ./internal/cloud ./internal/config ./internal/handlers ./internal/queue ./internal/repository -count=1` 通过；`go test ./internal/services -run 'TestProcessResubmits|TestProcessSkipsSubmit|TestProcessWritesBlocked|TestProcessBlocksWhenQuota|TestProcessDoesNotBlock|TestProcessSkipsTrigger|TestRetry|TestSubmitResets|TestReconcile|TestMarkFailed|TestFindDownloaded|TestExpectedDownload' -count=1` 通过（14 项云下载状态机/落盘测试）；`admin-web npm test -- --run` 730 项通过；`npm run build` 通过；`git diff --check` 与乱码扫描通过。未执行 `go test ./...`（既有 `TestParseTVAPKMetadataParsesReleaseAPK` 夹具版本不一致，与本次无关）；本机 9333 无 CDP 监听，未做 115 真机端到端。

## 2026-10-17 19:23 +0800 计划：115 云下载失败恢复与幂等修复
- 进度：复核工作区中未提交的 115 浏览器云下载功能后，确认三类阻塞问题：① `submitting` 状态在提交失败后不可重入；② asynq 超时/重启时终态写库使用已取消 ctx，且没有非终态任务对账；③ `blocked` 任务在管理端没有恢复入口。另确认重复入队/重复触发下载、落盘校验使用链接声明大小、失败原因不可见等问题。本轮按 P0→P1 修复：重提交幂等、终态写库脱离取消的 ctx、worker 启动与周期对账、asynq TaskID 去重、`download_triggered_at` 触发标记、使用 115 返回大小并做稳定性校验、管理端 blocked 重试与详情、配置校验及文档收口。
- 影响文件：`internal/cloud/browser.go`、`internal/services/cloud_download_service.go`、`internal/repository/cloud_download_repository.go`、`internal/models/cloud_download.go`、`internal/queue/cloud_download.go`、`internal/queue/tasks.go`、`internal/handlers/admin_cloud_download.go`、`internal/config/config.go`、`main.go`、`migrations/0043_cloud_download_triggered.{up,down}.sql`、`admin-web/src/views/CloudDownloadList.vue`、`admin-web/src/views/ForumPostList.vue`、对应测试、`CONTEXT.md`、`docs/adr/0028-explicit-cloud-download-from-forum-resources.md`、`docs/家用部署机.md`、`plan.md`。
- 验证：待实现后执行 `GOCACHE=/tmp/ai-video-go-cache go build ./...`、`go vet` 目标包、`go test ./internal/cloud ./internal/services ./internal/config ./internal/handlers ./internal/queue ./internal/repository -count=1`、`admin-web npm test -- --run` 与 `npm run build`，并做 `git diff --check` 与乱码扫描。

## 2026-10-14 RAR 新版压缩算法不兼容解压诊断优化
- 进度：收到部署机真实报错 `7zz 26.03 ERROR: Unsupported Method`（RAR5 + `Method = v6:32M:m0:m1:m3` + `QuickOpen` + `Blocks = 55`），确认不是密码误判，而是 7zz 不支持该 RAR 压缩方法；现有实现把 7zz 全部输出原样写入 `last_error`，既缺中文归因也缺可执行建议。计划新增“压缩算法不兼容”错误分类、精简错误文本与工具回退建议，并按 TDD 补纯逻辑单测。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：待执行 `go test ./internal/services -run 'TestArchive'` 定向测试；回退链是否引入 unar/unrar 依赖需先确认部署机可用工具。

## 2026-10-14 视频管理列表刷新按钮完成
- 进度：已在视频列表底部左侧增加带 Refresh 图标、加载状态和可访问名称的“刷新”按钮，复用 `load()` 保留当前查询条件；分页区域继续在同一行右侧对齐，并允许窄屏换行。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoListPage.spec.js`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：`npm test -- --run src/views/videoListPage.spec.js` 通过（19/19）；`npm run build` 通过；`git diff --check` 和本次文件乱码检查通过。待提交本次任务文件。

## 2026-10-14 视频管理列表增加刷新按钮
- 进度：确认视频列表底部现有分页仅靠右显示；计划在左侧增加带 Refresh 图标和加载状态的“刷新”按钮，复用当前查询条件下的 `load()`。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoListPage.spec.js`、`plan.md`。
- 验证：先补充底部刷新按钮与当前查询复用契约测试，再运行视频列表定向测试和管理端构建。

## 2026-10-13 压缩包解压兼容与误判修复完成
- 进度：收紧密码错误识别为明确的密码/passphrase 诊断；统一保留 7zz、7z、bsdtar、unzip 的原始错误文本；ZIP 元数据检测加密条目；Go ZIP 遇到不支持的压缩算法时清理半成品并自动回退 7zz/7z。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：归档相关定向测试通过；`git diff --check` 与乱码扫描通过。`go test ./internal/services -count=1` 受沙箱禁止 `httptest` 监听端口影响，既有 `TestMDCxMigratedSitesSearchCandidates` 失败，非本次改动。

## 2026-10-13 压缩包密码误判优化
- 进度：确认解压器错误被按模糊关键词归类为 `needs_password`，导致损坏、CRC、格式兼容性错误被误报为密码错误；计划收紧密码错误识别并保留外部解压器诊断文本。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`plan.md`。
- 验证：先补充密码误判与错误诊断回归测试并确认红灯，再运行服务端归档定向测试。

## 2026-10-10 11:38 +0800
- 进度：排查压缩包上传后 7zz 报 `ERRORS: Headers Error`。确认 7zz 调用参数正确；补充上传落盘字节数与 multipart 声明大小校验，并完整合并 7zz 标准输出/错误输出，避免截断归档进入解包阶段且保留完整头部诊断。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`plan.md`。
- 验证：归档定向测试通过；`git diff --check` 与乱码扫描通过；`go test ./internal/services -count=1` 仅失败于既有 `TestParseTVAPKMetadataParsesReleaseAPK`（测试期望版本 121，当前 release APK 为 144）。

## 2026-09-30 开放环境复验
- 验证：开放本机端口后运行 `GOCACHE=/tmp/ai-video-go-cache go test ./... -count=1`；网络监听型测试已通过。仅 `internal/services/TestParseTVAPKMetadataParsesReleaseAPK` 失败，仓库中的测试期望版本 `121 / 0.1.121`，当前 release APK 元数据为 `144`，TV 工程 Gradle 为 `147 / 0.1.147`，属于既有测试夹具与构建产物版本不一致，与本次压缩包媒体导入改动无关。

## 2026-09-30 收尾：补齐 HEIC/HEIF 识别与多图片合集配置
- 进度：归档扫描按图片扩展名返回稳定 MIME，HEIC/HEIF 在内容探测器无法识别时仍能进入图片流程；工作台移除视频关联图片合集的单选截断，保持视频级、分组级和批量编辑的多合集契约。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`admin-web/src/views/ToolboxArchiveImport.vue`、`plan.md`。
- 验证：`GOCACHE=/tmp/ai-video-go-cache go test ./... -run '^$' -count=1` 通过；HEIC/HEIF、TS 和图片 MIME 定向测试通过；管理端定向 60 项测试和 `npm run build` 通过；`git diff --check` 与乱码扫描通过。完整包测试中的 `httptest` 监听型既有测试受沙箱禁止本机端口影响，未能运行。
- 提交范围：仅纳入本次压缩包导入、媒体关系、格式识别和相关文档/测试文件；保留工作区既有未跟踪目录 `.reasonix/`、`docs/examples/`。

## 2026-09-30 实施：压缩包目录树、媒体关系与格式扩展
- 进度：已实现第一阶段。新增 `video_image_collections`、`video_collection_image_collections`、`video_images`、`archive_import_file_images` 关系表及旧 `videos.image_collection_id` 回填；归档扫描按目录生成导入组并默认建立同目录视频到全部图片的候选关系；新增网页 API 可查询/替换视频条目的图片文件关系；管理端增加文件系统层级树、目录统计、视频图片关系多选编辑和视频关联图片合集多选。
- 媒体识别：归档视频新增 `.ts`、`.mts`、`.m2ts`、`.mpeg`、`.mpg`、`.vob`、`.3gp`、`.ogv`；图片新增 `.heic`、`.heif`，图片服务支持按扩展名回退识别 HEIC/HEIF。
- 影响文件：`migrations/0040_archive_media_relations.{up,down}.sql`、`internal/repository/media_relation_repository.go`、`internal/services/archive_import_relations.go`、归档服务/模型/路由/管理端工作台及对应测试、`CONTEXT.md`。
- 验证：`GOCACHE=/tmp/ai-video-go-cache go test ./... -run '^$' -count=1` 通过；归档/关系定向 Go 测试和迁移测试通过；管理端定向 60 项测试通过；`npm run build` 通过（仅有既有 bundle 体积提示）；`git diff --check` 与 UTF-8 乱码扫描通过。完整 Go 测试未执行网络监听型测试，之前定向完整包测试受沙箱禁止 `httptest` 监听端口影响。

## 2026-09-30 设计：压缩包视频图片导入重构
- 进度：需求已确认。上传 ZIP 后递归扫描并按媒体所在目录生成导入组；同组每个视频默认关联同组全部图片；网页支持批次默认、导入组覆盖和单文件修正；视频可加入多个视频合集，图片可加入多个图片合集，视频合集可关联多个图片合集并向视频继承；确认入库后逐项处理，成功项保留、失败项可重试。
- 影响文档：`docs/superpowers/specs/2026-09-30-zip-media-import-redesign-design.md`、`plan.md`。本阶段未修改运行时代码、数据库或 App 版本。
- 验证：设计规格待用户评审；提交前执行 Markdown 差异检查、占位符扫描和 UTF-8 乱码扫描。

## 2026-09-30 需求补充：压缩包工作台使用文件系统层级展示
- 进度：工作台改为左侧目录树、右侧详情编辑；目录按路径段构造父子节点，用连接线显示层级，视频和图片作为叶子节点，目录节点提供统计和批量选择，视频节点提供同目录图片关系编辑。
- 影响文档：`docs/superpowers/specs/2026-09-30-zip-media-import-redesign-design.md`、`docs/superpowers/plans/2026-09-30-zip-media-import-redesign.md`、`plan.md`。尚未修改运行时代码。
- 验证：待实现后执行管理端源码测试、构建和差异检查。

## 2026-09-29 09:13 +0800 完成：手机端搜索改为点击后提交
- 进度：`updateAvQuery` 现在只更新输入并在清空时恢复默认浏览；新增 `submitAvQuery`，由放大镜按钮和键盘 Search 动作触发远程搜索。新增 `submittedQuery`，确保编辑待提交关键词时分页加载/下拉刷新不会隐式搜索新词；手机端版本升级为 `versionCode 12`、`versionName 0.1.11`。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/home/HomeViewModel.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest --tests com.chee.videos.feature.home.HomeViewModelTest` 通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`git diff --check` 通过；本次文件乱码扫描无结果；已提交 `3b3f479`。

## 2026-09-29 计划：手机端搜索改为点击后提交
- 进度：已定位 `HomeViewModel.updateAvQuery` 的防抖实时搜索，以及首页搜索框缺少显式提交动作的问题；计划改为输入仅更新状态，放大镜按钮/键盘搜索动作提交请求。
- 影响文件：`android-app/app/src/main/java/com/chee/videos/feature/home/HomeViewModel.kt`、`android-app/app/src/main/java/com/chee/videos/feature/home/HomeScreen.kt`、`android-app/app/src/test/java/com/chee/videos/feature/home/HomeViewModelTest.kt`、`android-app/app/build.gradle.kts`、`CONTEXT.md`、`plan.md`。
- 验证：待补充输入阶段无请求、提交后请求的单测，并运行手机端定向单测与差异检查。

## 2026-09-28 11:47 +0800 部署：启用 7zip 解包依赖
- 进度：部署机通过 Homebrew 安装 `sevenzip` 26.03，`/opt/homebrew/bin/7zz` 已存在；launchd server/worker 的 `PATH` 均包含 `/opt/homebrew/bin`，因此新解包逻辑会优先命中 7zz。提交 `b066f17` 已推送到 `deploy/master`，远端 Go 构建、签名、迁移、重启全部成功。
- 验证：部署 hook 报告 `/healthz OK — deploy succeeded`；部署机 server 与 worker 均为 `running`，HTTP `/healthz` 返回 200，7-Zip 输出版本 `26.03`。尚未使用用户原始归档做端到端导入复现，若安装 7zip 后仍报 CRC，应重新上传或校验归档完整性。

## 2026-09-28 11:45 +0800 验证：修复 RAR/7z 压缩包 CRC 导入失败
- 进度：定位到无密码 RAR/7z 强制走旧版 `bsdtar`，导致兼容格式被报为 `Header CRC error`。已让两种格式无论是否有密码都优先调用 `7zz`/`7z`；只有命令不存在时才回退 `bsdtar`，7zip 的真实 CRC、密码或格式错误直接返回。新增回归测试锁定无密码 RAR 的 7zip 优先级。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `.reasonix/`、`docs/examples/`。
- 验证：新增测试及归档相关测试通过；`go test ./internal/services -count=1` 仅失败于既有 `TestParseTVAPKMetadataParsesReleaseAPK`（部署样本 version_code=144、测试期望 121），与本次改动无关；`git diff --check` 通过。部署机当前未安装 `7zz`/`7z`，生产环境仍会回退 `bsdtar`，需安装 7zip 后再验证真实归档。

## 2026-09-25 15:01 +0800 验证：管理端视频列表按时长排序
- 进度：实现与复核完成。默认查询仍按上传时间倒序；时长排序与关键词/类型/状态筛选和服务端分页共用接口；新列在旧列设置中自动迁移为默认显示，之后可以隐藏；旧保存视图加载为默认排序。
- 影响文件：`CONTEXT.md`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/videoListPage.spec.js`、`internal/models/admin.go`、`internal/handlers/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`plan.md`。未纳入既有未跟踪目录 `.reasonix/`、`docs/examples/`。
- 验证：`go test ./internal/repository ./internal/handlers ./internal/models -count=1` 通过；`npm test -- --run src/views/videoList.helpers.spec.js src/views/videoListPage.spec.js` 的 47 项测试通过；`npm run build` 通过（仅报告 bundle 大小提示）；`git diff --check` 通过；乱码扫描无结果；Vite 开发服务 `/` 返回 HTTP 200。

## 2026-09-25 计划：管理端视频列表按时长排序
- 进度：需求经 `grill-with-docs` 访谈确认。仅修改管理端 `/videos`；新增默认/时长升序/时长降序排序，服务端分页与现有筛选组合；未知时长（NULL/0）始终置后；同长度按上传时间倒序、视频 ID 升序稳定排序；保存视图包含排序；时长列默认显示且可通过列设置隐藏。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、对应前端测试、`internal/models/admin.go`、`internal/handlers/admin.go`、`internal/repository/admin_repository.go`、对应 Go 测试、`CONTEXT.md`、`plan.md`。不涉及 Android App 版本。
- 验证：先运行新增定向测试确认红灯，再执行 Go handler/repository 测试、管理端定向测试与构建；完成后追加验证记录。

## 2026-09-25 15:00 +0800 实施：管理端视频列表按时长排序
- 进度：后端列表新增 `duration_seconds`、受控时长升降序和未知时长置后；时长并列时按上传时间倒序、ID 升序稳定排序；默认排序保持原 `created_at DESC`。前端新增排序下拉和时长列，排序随筛选及分页请求，保存视图记住排序，并把已有列设置迁移为默认显示时长列。
- 影响文件：`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/videoListPage.spec.js`、`internal/models/admin.go`、`internal/handlers/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：`go test ./internal/repository ./internal/handlers ./internal/models -count=1`、前端定向 47 项测试及 `npm run build` 已通过；待最终复跑和差异检查。

## 2026-09-20 11:20 +0800
- 进度：核查凌晨迁移后的上传暂存与记录关联，并修复发现的缺口。① `uploads/assembled/` 迁移后 33 个文件中 28 个有记录引用，5 个为孤儿（清单归档 `repair-backups/orphan-files-20260920-111616.txt`；其中 `fb4ff09a-…ASMR Puffin 熟客限定` 经确认是重复上传，同一视频另有 session `9f65e51f` 的记录已 `ready`），在 `archive_import_files` 引用校验为 0 后删除，释放约 6.4 GB；② 未转码记录共 25 条（全部 `short` 类型、9/19 21:11–22:18 创建），源文件 25/25 全部存在，不存在"记录缺文件"；③ 其中 23 条为 `uploaded` 且 `transcoding_jobs` 计数为 0（对照：近 3 天 82 条 `ready` 记录 82/82 都有 job 行），属"记录已落库但转码任务未入队"的缺口；另 2 条 `failed` 的 job 错误信息为「转码任务长时间无进度，系统已自动标记失败」。
- 进度：用管理端 `POST /api/v1/admin/videos/:id/retranscode` 逐条补齐这 25 条。触发前先校验"记录存在且源文件存在"（25/25 通过、缺失 0，否则整体中止），全部返回 `200` 且 `enqueued:true`、`input_source:original`。
- 关键约束：上传的"写记录"与"入队转码"是两步、非同一事务，入队失败会留下 `status=uploaded` 且 `transcoding_jobs` 为 0 的哑记录，只能靠人工 `retranscode` 恢复；该缺口恰好发生在 `server.log` 轮转空档期内，具体入队报错已丢失。已作为 `上传转码入队缺口` 写入 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`（新增 `上传转码入队缺口`）、`plan.md`；部署机侧仅删除 5 个孤儿暂存文件并触发 25 条重转码，未修改应用代码、配置或 Android 版本。
- 验证：触发后 `asynq:{transcode}` 为 `pending=23`、`active=2`，两个 `ffmpeg` 进程以 `tmp/uploads/assembled/...` 相对路径读取源文件；`transcoding_jobs` 出现新的 `running` 行、对应视频状态转为 `processing`；`videos` 近 3 天状态分布由 `ready 82 / uploaded 23 / failed 2` 变为 `ready 82 / uploaded 23 / processing 2`。25 条全部转码完成需数小时（worker 并发 2），本轮未等待完成。

## 2026-09-20 00:47 +0800
- 进度：清理家用部署机内盘的构建/包管理缓存，共释放约 8 GB（`/System/Volumes/Data` 已用 124 GB → 115 GB，可用 71 GB → 79 GB）：`~/.npm` 2.4 GB → 6.3 MB、`~/Library/Caches/Homebrew` 2.3 GB → 797 MB、`~/Library/Caches/Google/Chrome` 1.6 GB、`~/Library/Caches/go-build` 598 MB → 8 KB、`~/Library/Caches/115Browser` 605 MB、`~/.cache/huggingface/hub`（翻译模型下载缓存）1.1 GB、`~/.docker/modules` 176 MB。按约定未触碰 Android 开发环境（`~/Library/Android/sdk` 11 GB、`~/.android/avd` 6.7 GB、`~/.gradle` 5.1 GB、`Android Studio.app` 3 GB）与 Chrome 用户数据（`~/Library/Application Support/Google` 7.5 GB），也保留了 `~/Library/Caches/Google/AndroidStudio2025.3.3`（810 MB）。
- 关键约束：翻译服务的 `llama-server` 实际由 `~/.docker/bin/inference/llama-server`（Docker Desktop 卸载后的残留二进制）提供，因此 `~/.docker/bin` 不能跟着残留一起清理；它用绝对路径 `-m` 加载 `~/Library/Caches/llama.cpp/` 下的 gguf，所以 `~/.cache/huggingface/hub` 的同源下载缓存可以安全删除。已作为 `部署机翻译服务运行时依赖` 写入 `CONTEXT.md`。
- 影响文件：`CONTEXT.md`（新增 `部署机翻译服务运行时依赖`）、`plan.md`；本次只做部署机缓存清理，未修改应用代码、配置或 Android 版本。
- 验证：清理后 `curl 127.0.0.1:8000/health` 返回 200（模型 gguf 1,133,080,512 字节仍在、`llama-server` 进程存活）、`curl 127.0.0.1:8080/healthz` 返回 200；各缓存目录清理后大小逐一复核，`df -h /System/Volumes/Data` 确认释放约 8 GB。

## 2026-09-20 00:47 +0800
- 进度：修正上一则条目的遗留 ①。`0d3ca1be-f41c-4cfb-9c5c-b96aa8b49e4b` 并未停留在 `failed`：symlink 生效后 asynq 重试队列重新取回该任务，worker 已再次开始转码，因此无需人工执行「重新转码」；遗留 ②（`server.log` 轮转空档）与 ③（Postgres 绑定 `0.0.0.0:5432`）仍然成立。
- 验证：`pgrep -fl ffmpeg` 同时可见 `689edf6e`（pid 45721）与 `0d3ca1be`（pid 45860）两个转码进程，两者命令行都以 `tmp/uploads/assembled/...` 相对路径读取外盘源文件；`/Volumes/large/videos/videos/689edf6e-.../video-avc.mp4` 已写入 466 MB 且持续增长；`asynq:{transcode}:active` = 2、`pending` = 0。

## 2026-09-20 00:41 +0800
- 进度：恢复家用部署机服务并完成上传暂存目录外盘化的收尾。部署机上 `com.aivideo.server` / `com.aivideo.worker` 两个 launchd 服务当时已不存在（plist 仍在、非 disabled 状态），8080 无监听，因此客户端无法访问；机器本身连续在线 116 天，Postgres（OrbStack）/Redis/翻译端点均正常，DB 数据完整（171397 条 `videos`、schema `0039_telegram_management`）。根因是 2026-09-20 00:19～00:21 的容器运行时迁移（卸载 Docker Desktop、改用 OrbStack）期间停了服务但没有恢复：worker 于 00:19:40 收到 SIGTERM 并 graceful shutdown，Postgres 容器 00:21:20 重新启动。处置：用 `launchctl bootstrap` 重新加载两个 plist，服务恢复，`/healthz` 在本机与 LAN 均返回 200。
- 进度：完成 `~/deploy/ai-video-server/current/tmp/uploads` 到 `/Volumes/large/ai-video-server/tmp/uploads` 的迁移收尾。rsync 复制此前已完成，核对两边 1286 个文件、73,718,611,048 字节以及「相对路径 + 大小」清单完全一致后删除内置盘副本，根盘可用空间由 2.4 GB 恢复到 71 GB（占用 192 GB → 124 GB）；同时清理外盘 `chunk-sessions` 中 1251 个全部超过 30 天的废弃分片（4.9 GB），清单留档 `repair-backups/chunk-sessions-list-20260920-003746.txt`。
- 关键约束：`UPLOAD_TEMP_DIR` 必须维持相对路径 `./tmp/uploads`，靠 `current/tmp/uploads -> 外盘` 的 symlink 生效。`videos.original_path` 与 asynq 任务参数里持久化的是 `tmp/uploads/assembled/...` 这类相对 launchd 工作目录（`current/`）的路径，直接改成外盘绝对路径会让历史记录与队列任务全部报 `ffmpeg ... No such file or directory`（本次已在删除内置盘副本的窗口内实测到该失败）。
- 影响文件：`CONTEXT.md`（改写 `admin 上传暂存目录`，新增 `launchd 日志轮转空档`）、`docs/家用部署机.md`（新增「上传暂存目录：实体在外盘，配置保持相对路径」小节，更新 `.env` 示例注释与「只做服务器时的权限收口方案」第 1 条）、`plan.md`；部署机侧改动 `~/deploy/ai-video-server/.env`（备份 `.env.bak-20260920003742`，最终值与迁移前一致）与 `current/tmp/uploads` symlink。未修改应用代码，未改 Android 版本。
- 遗留：① `0d3ca1be-f41c-4cfb-9c5c-b96aa8b49e4b` 在删除窗口内被 worker 重试到 `MaxRetry(3)` 上限并置为 `failed`（它原本就是卡住的 stale `processing`），需在管理端对该视频执行「重新转码」（`POST /api/v1/admin/videos/:id/retranscode`，源文件现已可通过 symlink 解析）；② `server.log` 自 2026-09-14 10:30 newsyslog 轮转后连续 6 天零输出，轮转方案需另行处理；③ Postgres 容器端口当前绑定 `0.0.0.0:5432`，与文档约定的 `127.0.0.1:5432` 不一致。
- 验证：部署机 `curl 127.0.0.1:8080/healthz` 与开发机 `curl 192.168.1.24:8080/healthz` 均返回 `{"status":"ok"}`；删除内置盘副本后 worker 经 symlink 正常转码（`ffmpeg -i tmp/uploads/assembled/3083dce6-...mp4 ... /Volumes/large/videos/videos/689edf6e-.../video-avc.mp4` 实测读取成功）；`df -h /System/Volumes/Data` 确认释放 69 GB；`lsof` 确认无进程仍占用被删目录；`launchctl print` 确认两个服务 running。

## 2026-09-12 计划：删除 53 条失效视频记录
- 进度：已完成。最终核对 53/53 个 ID 仍存在且全部为 `ready`、`short`、标题 `#迁移`；在部署机 PostgreSQL 单事务内锁定目标后删除 53 条 `videos` 记录，事务保护条件未触发异常。
- 影响范围：部署机 PostgreSQL 中明确列出的 53 个视频 ID、数据库关联数据及这些记录自身引用的媒体文件；不修改应用代码，不触碰工作树已有 `.reasonix/`、`docs/examples/`。
- 安全措施：删除前备份目录为 `/Users/chee/deploy/ai-video-server/repair-backups/delete-invalid-videos-20260912-1558`，包含 53 行视频快照、107 行关联快照、53 行路径来源、267 行路径状态、删除 SQL/结果及全表数据备份；清单 `SHA256SUMS.ALL` 的 SHA-256 为 `03b3b07d2b25227236e0eae4b2808c7100ee6046dc7aca634afaba1c27db646a`，全表备份 SHA-256 为 `50a0d792c67a5f6b9549b792ecd704088daf51726c859846558c19dcc7b894c`。
- 删除语义：显式清理 `transcoding_jobs`、`user_video_actions`、`video_subtitles` 后删除 `videos`；数据库外键级联清理 `video_tags`、`file_hashes`、`video_actors`、`video_collections`，`episodes`/`telegram_media` 的设置为空约束保持一致。目标记录的主视频路径均已缺失，因此没有应用层媒体文件可删除；特殊记录的 `/Volumes/large/videos/videos/1e450f13-ea2e-4740-9166-7d9592af97ed/source-original.mp4` 不在记录路径字段中，保留为待后续人工确认的孤立文件（1,484,461 字节，MD5 `3dad85b48a2281b5213c7609332c539a`）。
- 验证：删除后上述 10 个关联表按目标 ID 查询均为 0 行；删除结果 UUID 共 106 行、去重后 53 个，快照 53 行；目标旧/新视频目录均不存在（特殊 `source-original` 目录除外），4 类视频路径缺失 212 项、53 个 metadata 封面仍存在；`GET /api/v1/videos/1e450f13-ea2e-4740-9166-7d9592af97ed/thumbnail` 返回 `404 video not found`，部署健康检查 `/healthz` 返回 200。

## 2026-09-12 15:20 +0800
- 进度：完成部署机媒体路径修复。事务内更新 4,107 条 `ready` 短视频记录的 `original_path`、`transcoded_path`、`thumbnail_path`，将 `storage/videos/...` 统一归一化为 `/Volumes/large/video-server-storage/videos/...`；未搬动媒体文件，未改状态、创建时间或数据库结构。
- 影响文件：`plan.md`；部署机 PostgreSQL `videos` 表 4,107 条记录；远程备份 `/Users/chee/deploy/ai-video-server/repair-backups/relative-video-paths-20260912-1510.csv`（4,108 行含表头，SHA-256 `c12d260f6888dbe74f41b0017717f57a4fcdb9d10f694e514581fd6d3df5e0a6`）。未纳入仓库无关未跟踪目录 `.reasonix/`、`docs/examples/`。
- 验证：数据库 `ready` 相对路径剩余 0；备份中的 4,107 个 ID 全部找到且仍为 `ready`，路径逐条匹配、4,107 × 3 个文件均为现存非空普通文件；修复记录与现用存储记录的 signed source 抽样均返回 `206 video/mp4`。

## 2026-09-12 15:10 +0800
- 进度：确认部署机 4,107 条 `ready` 短视频记录的三类媒体路径均为 `storage/videos/...` 相对路径，实际文件全部存在于 `/Volumes/large/video-server-storage/videos/...`；计划在部署机导出原路径快照后，将 `original_path`、`transcoded_path`、`thumbnail_path` 统一归一化为该旧存储根下的绝对路径，不搬动媒体文件。
- 影响文件：`plan.md`；部署机 PostgreSQL `videos` 表中的上述 4,107 条记录及远程路径快照文件。未修改应用代码、数据库结构或 Android 版本。
- 验证：已完成全量文件元数据核对，三类路径均为现存非空普通文件；待执行快照校验、事务更新、更新后计数与播放接口抽样验证。

## 2026-09-12 10:53 +0800
- 进度：管理端「视频管理」列表新增「大小」列。列表接口 `AdminListVideos` 在 SQL 内一次带出大小，口径为「已转码优先 `videos.transcoded_file_size`，缺失回落 `file_hashes.file_size`（上传原始大小）」，都缺失返回 0、前端显示 `--`，不再需要为看大小逐行打开详情；列设置新增 `size` 项并默认可见，窄屏不隐藏。用户已确认该口径。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`admin-web/src/views/VideoList.vue`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/videoListPage.spec.js`、`CONTEXT.md`、`plan.md`。未修改运行时代码以外的数据库结构或 Android 版本，未纳入既有未跟踪 `docs/examples/`。
- 验证：`go test ./internal/repository/ ./internal/handlers/ ./internal/models/` 通过（含新增源码守卫 `TestAdminListVideosFileSizeSourceGuard`）；`npx vitest run src/views/videoListPage.spec.js src/views/videoList.helpers.spec.js` 44 项通过；`npm test` 715/716 通过，唯一失败为既有 `precisionOpsRollout.spec.js` 的 views 清单断言（写死 26 个 `.vue`，目录实际 27 个），与本次改动无关。未连接数据库，未执行列表接口的真实数据回放。

## 2026-09-08 10:29 +0800
- 进度：完成日本 AV 元数据目录第三轮设计。ThePornDB 官方 OpenAPI 能力与政策缺口已写入来源台账；新增可实施的数据与同步契约，覆盖审批证据、准入/运行双状态、目录发行与来源快照、人物/分类/图片、人工覆盖、本地视频关联、运行游标、适配接口、页事务、增量重叠水位、限速重试、自动暂停和管理 API。来源仍未获准，未调用目录数据接口。
- 影响文件：`tasks/2026-09-08-japanese-av-catalog/{prd.md,implement.md,review.md,source-audit.md,domain-model.md,data-and-sync-contract.md}`、`plan.md`；未修改运行时代码、数据库或 Android 版本，未纳入既有未跟踪 `docs/examples/`。
- 验证：官方 OpenAPI 实时回放确认版本 `v3.24.747`、Bearer 认证、`/jav`、`/changes/{type}`、`jav` 变化类型和每页上限 100，同时确认 `termsOfService` 与 `license` 均为空；`git diff --check` 通过；相关 Markdown 均存在且非空，U+FFFD 扫描无结果。

## 2026-09-08 10:24 +0800
- 进度：进入日本 AV 元数据目录第三轮设计。ThePornDB 官方 OpenAPI 只读审计确认其提供 `/jav` 列表、`/changes/jav` 增量变化、Bearer Token、最大每页 100 条，以及标题、简介、日期、时长、演员、标签、导演、站点、图片和更新时间字段；但规范未声明 Terms、License、Rate Limit、缓存或再展示授权，站点也未提供可访问的条款页，因此继续保持“待证据”。本轮将新增数据与同步契约，细化核心 schema、来源适配接口、事务水位、失败暂停和管理 API。
- 影响文件：预计新增 `tasks/2026-09-08-japanese-av-catalog/data-and-sync-contract.md`，更新 `source-audit.md`、`prd.md`、`implement.md`、`review.md` 与 `plan.md`；不修改运行时代码、数据库、Android 版本或既有未跟踪 `docs/examples/`。
- 验证：待执行文档差异/乱码检查，并逐项核对 OpenAPI 版本、路径、分页上限、认证和缺失政策声明；来源未批准前不得调用目录数据接口或实施生产调度。

## 2026-09-08 10:19 +0800
- 进度：完成日本 AV 元数据目录第二轮设计沉淀。新增来源准入状态与证据门禁、现有 35 个抓取器盘点、五个候选来源的只读 `robots.txt` 检查、目录领域模型，以及“目录发行条目与本地视频分离”的提议中 ADR；明确当前没有已批准来源，禁止启动全量或周期抓取。
- 影响文件：`tasks/2026-09-08-japanese-av-catalog/{prd.md,implement.md,review.md,source-audit.md,domain-model.md}`、`docs/adr/0027-av-catalog-separate-from-local-video.md`、`plan.md`；未修改运行时代码、数据库或 Android 版本，未纳入既有未跟踪 `docs/examples/`。
- 验证：代码计数确认 `newAVCrawlerProvider` 注册 35 个抓取器、`defaultAVEnabledSites` 启用 15 个；`git diff --check` 通过；相关 Markdown U+FFFD 扫描无结果；ADR 和领域模型均保持“提议中”。

## 2026-09-08 10:18 +0800
- 进度：继续日本 AV 元数据目录需求收口。代码审计确认现有 35 个 AV 抓取器只支持按查询搜索和详情读取，统一候选模型缺少分类、片商、系列、导演、时长及目录发现/同步状态；首批五个候选来源的 `robots.txt` 只读检查未产生可直接上线的授权证据。将新增来源准入台账、领域模型草案和“外部目录与本地视频分离”的提议中 ADR，不进入运行时代码。
- 影响文件：预计新增 `tasks/2026-09-08-japanese-av-catalog/source-audit.md`、`tasks/2026-09-08-japanese-av-catalog/domain-model.md`、`docs/adr/0027-av-catalog-separate-from-local-video.md`，更新任务三段文档与 `plan.md`；保留既有未跟踪 `docs/examples/`。
- 验证：待执行 Markdown 差异检查、U+FFFD 乱码扫描、来源数量与默认启用来源对照检查；ADR 保持“提议中”，用户未确认前不写入 `CONTEXT.md` 作为已接受术语。

## 2026-09-08 10:12 +0800
- 进度：完成“日本 AV 元数据目录”长期任务骨架，建立访谈中的 PRD、分阶段实现草案和评审门禁。文档明确复用现有单片刮削能力，但在产品范围、使用范围、来源许可、采集规模、业务主键、字段冲突、图片版权、语言、删除策略和本地视频关联十项决策完成前不进入实现。
- 影响文件：`tasks/2026-09-08-japanese-av-catalog/prd.md`、`tasks/2026-09-08-japanese-av-catalog/implement.md`、`tasks/2026-09-08-japanese-av-catalog/review.md`、`plan.md`；未修改运行时代码，未纳入既有未跟踪 `docs/examples/`。
- 验证：`git diff --check` 通过；任务三段文档均存在且非空；`plan.md` 与任务目录 U+FFFD 扫描无结果。长期目标仍处于需求访谈阶段，不标记功能完成。

## 2026-09-08 10:10 +0800
- 进度：启动“日本 AV 元数据目录”长期任务并进入需求访谈。现有系统已经具备本地视频刮削、候选确认、演员关联、封面落地、文本翻译及多站点适配；本任务需要先确认是仅补全本地媒体，还是建立独立于本地视频的外部完整目录，并收口来源授权、采集规模、更新策略、数据溯源和内容安全边界后再实施。
- 影响文件：新增 `tasks/2026-09-08-japanese-av-catalog/{prd.md,implement.md,review.md}`，更新 `plan.md`；不修改运行时代码、数据库、Android 版本或既有未跟踪 `docs/examples/`。
- 验证：待执行 Markdown 差异检查、U+FFFD 乱码扫描和任务三段文档完整性检查；需求关键决策未完成，当前不得进入实现或宣称任务完成。

## 2026-09-07 16:50 +0800
- 进度：完成暖金深色导航精修。侧栏固定文字与图标、加载期间常驻；内容返回先回当前侧栏，浏览操作取消退出确认；侧栏右键及当前项重复确认进入内容，首目标未挂载时回退可见方向寻焦；搜索保留原生输入焦点与 IME 搜索动作，加载改行内指示；首分区请求器只绑定首个非空分区；推荐区去浮卡阴影并裁剪背景动画。
- 影响文件：`TvCatalogScreen.kt`、`TvShellApp.kt`、导航/焦点回归测试、TV 版本文件（147 / 0.1.147）、`CONTEXT.md`、`plan.md`。排除既有 `docs/examples/`。
- 验证：TV 全量 630 项单测与 debug 构建通过；独立评审发现重复确认在首目标卸载时缺少回退，已修复并新增覆盖，最终导航/焦点定向单测和 debug 构建通过。`git diff --check` 与 U+FFFD 扫描通过；无连接的 ADB 设备，本机只有手机模拟器配置，未执行真实 TV 截图、IME 返回或遥控焦点验收，不能以源码约束单测替代实机结果。

## 2026-09-07 16:42 +0800
- 进度：用户确认沿用暖金深色精修，希望靠近主流 TV UI 与操作习惯。第二批聚焦导航可识别性与焦点连续性：侧栏固定宽度图标配中文名称，分类加载不卸载侧栏，内容返回先回当前侧栏，侧栏再返回走退出确认；检查搜索输入的原生焦点与键盘行为，避免嵌套焦点妨碍输入。
- 影响文件：预计修改 TV 目录、导航壳层及对应测试、TV 版本、`CONTEXT.md`、`plan.md`。保留既有 `docs/examples/`。
- 验证：TV 全量单测与 debug 构建、独立评审、差异和编码检查；已询问常用设备与可用 ADB 地址，实机验证依赖设备可用性。

## 2026-09-07 16:09 +0800
- 进度：完成 TV 整体优化的第一批确定性改进：设置页可滚动、连播整行单一开关与步长单选语义、账户动作区分图标；共享状态页收口主次按钮并支持换行；海报墙保存影片 ID 并恢复跨页焦点；重复确认当前菜单不重载或清空搜索。视觉方向与返回键偏好尚未收到答复，本批沿用原视觉和退出语义，整体重设计及实机验收仍待继续。
- 影响文件：TV 设置/目录与 ViewModel、海报墙及焦点策略、共享状态组件、对应回归测试、`android-tv-app/tv-app/build.gradle.kts`（146 / 0.1.146）、`CONTEXT.md`、`plan.md`。既有 `docs/examples/` 不纳入。
- 验证：搜索重复确认回归测试先失败后修复；完整 `:tv-app:testDebugUnitTest :tv-app:assembleDebug` 通过。独立评审发现返回期间刷新结束不触发焦点恢复，已补 `refreshing` effect key 并保留一次性抢焦保护；修复后 `:tv-app:testDebugUnitTest --tests '*TvPosterWall*' :tv-app:assembleDebug` 通过。差异与中文 U+FFFD 检查通过。ADB 无连接设备，尚未执行真实电视截图与遥控焦点验收。

## 2026-09-07 16:00 +0800
- 进度：启动 TV App 整体 UI 与遥控操作优化。已读取实际 TV 入口、共享视觉组件、目录/海报墙、设置与历史约定；确认设置页不可滚动、连播开关重复焦点、状态按钮独立样式等问题。先收口共享组件与设置可达性，再处理浏览恢复与导航；视觉方向和返回语义已向用户征询。
- 影响文件：预计涉及 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/`、`feature/tv/`、`tv/` 及对应测试、TV 版本文件、`CONTEXT.md`、`plan.md`；排除既有未跟踪 `docs/examples/`。
- 验证：按风险执行 TV 定向单测、全量单测与 debug 构建，收尾检查差异和中文编码；当前 `adb devices -l` 无连接设备，实机焦点与视觉验收仍需设备。

## 2026-09-04 17:41 +0800
- 进度：完成 Telegram 管理台收尾。部署手册契约已改为校验管理页面、手机号/二维码授权、来源预解析确认、失败恢复和重新回填，并明确移除旧的交互式登录、直接来源 API 和手工媒体 SQL 流程；环境示例测试同步覆盖控制令牌。
- 影响文件：`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`、`.env.example`、Telegram 管理 API/服务/采集器/管理端页面及其测试、`CONTEXT.md`；用户既有未跟踪目录 `docs/examples/` 明确不纳入。
- 验证：Telegram 定向 Go 测试、精确 `-race` 测试、部署契约、`go vet ./...`、`go build ./...`、管理端 55 项定向测试、`npm run build`、`git diff --check` 和 U+FFFD 乱码扫描通过。全量 `go test ./... -count=1` 仅因既有 `internal/services/tv_apk_test.go` 固定期望 `versionCode=121` 而实际为 `144` 失败；该问题与本次改动无关。待精确暂存并提交。

## 2026-09-04 16:56 +0800
- 进度：继续 Telegram 管理台页面实现。将在 `admin-web` 增加受保护的 `/telegram` 工作区、控制 API 封装和定向测试，覆盖手机号/二维码授权、验证码与二次验证、来源预解析确认、来源运行控制、进度与审计；授权秘密、确认票据和私密邀请引用仅保留当前页面内存。轮询按授权中 2 秒、普通状态 5 秒执行，禁止请求重入，页面隐藏或离开时停止。
- 影响文件：预计修改 `admin-web/src/api/admin.js`、`admin-web/src/api/admin.spec.js`、`admin-web/src/router/index.js`、导航配置及新增 Telegram 管理视图/帮助函数和测试，后续更新 Telegram 运维文档、ADR、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：先运行新增 API/轮询 helper 的 Vitest 定向测试，再运行 `npm run build`；收尾复跑 Telegram Go 定向测试、race、vet、差异检查与乱码扫描。

## 2026-09-04 16:36 +0800
- 进度：继续 Telegram 管理 API 接线。确认新增 handler 草稿当前红灯仅因局部 `ok` 布尔变量遮蔽统一响应函数，且 `API` 尚未注入管理服务；下一步将用管理服务替换旧的原始 `chat_ref` 直写路由，补管理端权限/票据/API 覆盖，并保持私密邀请链接不落库、不入审计。
- 影响文件：`internal/handlers/admin_telegram_management.go`、`internal/handlers/admin_telegram_test.go`、`internal/handlers/router.go`、`main.go`、`plan.md`；后续继续处理来源恢复、Vue 管理页、文档与长期约定；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/handlers ./internal/services ./internal/telegram -run 'Test(AdminTelegram|TelegramManagement|TelegramConfirmation|TelegramSource|ControlClient)' -count=1` 当前预期红灯，报错为 handler 内 `ok` 遮蔽；服务与控制客户端相关测试已通过。

## 2026-09-04 16:14 +0800
- 进度：进入 Telegram 管理 API 阶段。将新增 API 到采集器的受控 HTTP 客户端、管理员密码确认的一次性短期票据，以及预览确认后按 canonical chat ID 创建来源的服务边界；随后接入现有 admin 路由、审计、失败恢复和 Vue 工作区。旧的直接 `chat_ref` 新增路径不再作为管理台正常流程。
- 影响文件：预计新增 `internal/telegram/control_client.go`、`internal/services/telegram_management.go` 及测试，修改 `internal/handlers/`、`internal/services/telegram_source.go`、`main.go`、`CONTEXT.md`、`plan.md`；后续修改 `admin-web/src/`；继续排除用户既有 `docs/examples/`。
- 验证：先建立控制客户端、票据消费和已确认来源创建的无网络红灯测试；随后运行 Telegram/API 相关 Go 包测试、race、vet 和管理端构建。

## 2026-09-04 16:11 +0800
- 进度：完成采集器常驻控制面接线。`-mode run` 通过运行时门控重建已授权 MTProto 客户端及其采集/订阅/Asynq 实例，不再创建 stdin 登录输入器；私有控制 HTTP、立即及 30 秒心跳独立常驻，未授权 session 时页面仍可查询并启动手机号或二维码授权。补充真实监听器集成测试，覆盖控制状态响应、初始心跳和取消后的优雅关闭。
- 影响文件：`cmd/telegram-ingestor/main.go`、新增 `cmd/telegram-ingestor/control_service.go`、`cmd/telegram-ingestor/heartbeat.go`、`cmd/telegram-ingestor/control_service_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./cmd/telegram-ingestor -count=1`、`GOPROXY=off go test -race ./cmd/telegram-ingestor -count=1`、`GOPROXY=off go vet ./cmd/telegram-ingestor`、`GOPROXY=off go build ./cmd/telegram-ingestor`、`git diff --check` 通过；`golangci-lint` 未安装。构建产生的临时根目录二进制已移入系统废纸篓，未纳入工作区或提交。

## 2026-09-04 16:05 +0800
- 进度：续接 Telegram 管理控制面。新增采集器内进程控制服务与心跳的定向测试已通过，下一步把授权服务、来源预览、运行时门控和仅内部可达的控制 HTTP 接入 `-mode run`；采集连接恢复后将重建消费服务与订阅，未授权状态仍保留页面管理能力。
- 影响文件：新增 `cmd/telegram-ingestor/control_service.go`、`cmd/telegram-ingestor/heartbeat.go`、`cmd/telegram-ingestor/control_service_test.go`；预计修改 `cmd/telegram-ingestor/main.go`、后续 API/管理端文件、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./cmd/telegram-ingestor -run 'Test(IngestorControlService|WriteTelegramHeartbeat)' -count=1` 通过；待继续运行采集器包测试、race、vet 和构建。

## 2026-09-04 15:56 +0800
- 进度：完成来源预解析确认票据层。控制协议的预览响应新增加入/审批标记且不再包含原始 `chat_ref`；`SourcePreviewService` 仅保存操作者、过期时间、Telegram 返回的展示结果和输入 SHA-256 指纹，确认时校验操作者/引用/类型/canonical ID，并且只有返回可保存 chat ID 后才消费票据。
- 影响文件：`internal/telegram/control.go`、`internal/telegram/source_preview.go`、`internal/telegram/source_preview_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、对应 `-race`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；待精确提交后将服务接入采集器控制 HTTP。

## 2026-09-04 15:52 +0800
- 进度：完成控制通道配置与采集器运行时门控子步骤。配置新增地址、API 访问 URL 和必填共享令牌，Compose 与示例环境同步注入；新运行时状态机在未授权后停驻等待页面授权恢复、在授权开始时取消并等待旧连接释放 session、恢复后以新客户端重连，控制操作仅在已授权运行态可用。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`cmd/telegram-ingestor/runtime.go`、`cmd/telegram-ingestor/runtime_test.go`、`.env.example`、`deploy/docker-compose.telegram.yml`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/config ./cmd/telegram-ingestor -count=1`、对应 `-race`、`GOPROXY=off go vet ./internal/config ./cmd/telegram-ingestor` 通过；待检查 Compose 解析、精确提交后实现控制服务与心跳入口接线。

## 2026-09-04 15:42 +0800
- 进度：提交 `8436e85` 后进入采集器常驻控制阶段。将增加 `TELEGRAM_CONTROL_ADDR`、`TELEGRAM_CONTROL_URL`、`TELEGRAM_CONTROL_TOKEN`，在独立运行时管理器中把未授权空闲、授权暂停、session 重载和已授权采集连接收敛为可测试状态；随后启动仅内部可达的控制 HTTP 并按 30 秒写入心跳。
- 影响文件：预计修改 `internal/config/`、`cmd/telegram-ingestor/`、`internal/telegram/control.go` 及测试，后续同步 Compose、`.env.example`、运维文档、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：先为配置与运行时状态机建立无网络红灯测试；阶段性运行 `GOPROXY=off go test ./internal/config ./cmd/telegram-ingestor -count=1`、对应 `-race`、`go vet` 与 `go build ./cmd/telegram-ingestor`。

## 2026-09-04 15:38 +0800
- 进度：完成第二阶段采集器运行时基础的客户端子步骤。`GotdClient.RunAuthorized` 仅验证现有 persistent session 的授权状态，未授权时返回可识别错误且不进入 stdin 登录；来源解析已拆为不入群的 `PreviewChat` 与确认后才可能执行邀请导入的 `ConfirmChat`，保留 `ResolveChat` 的既有兼容语义。私密邀请的审批加入请求不会自动入群或生成来源。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/gotd_client.go`、`internal/telegram/client_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、`GOPROXY=off go test -race ./internal/telegram -count=1`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；待精确提交后进入常驻控制 HTTP、授权维护 gate 与心跳接线。

## 2026-09-04 15:32 +0800
- 进度：授权服务层已提交 `1cb50d0`，开始第二阶段采集器运行时基础：为 gotd 客户端增加仅使用现有 session 的无交互运行入口，并将来源引用拆成不入群的预解析和确认后入群两步；随后把该能力接入常驻控制 HTTP、授权维护 gate 和心跳。旧 `-mode run` 不再依赖 stdin，页面化授权成为正常路径。
- 影响文件：预计修改 `internal/telegram/client.go`、`internal/telegram/gotd_client.go` 及测试，后续修改 `cmd/telegram-ingestor/`、`internal/config/`、`.env.example`、Compose、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：先建立无网络的客户端状态/引用解析测试，阶段性运行 `GOPROXY=off go test ./internal/telegram`、`go vet` 和采集器构建；真实 session、私密邀请和在线运行仍在部署环境验证。

## 2026-09-04 15:26 +0800
- 进度：完成页面化 Telegram 授权服务的首个可提交层：单例授权服务将账号元数据、审计、管理员所有权和取消/过期清理与短期 session 分离；首次绑定仅允许手机号，二维码仅允许已有绑定且成功后校验 Telegram 用户 ID。新增 gotd 临时 session adapter，验证码、二次验证密码、二维码 token 均不进入仓储、审计或日志；临时文件在同目录经原子替换提升为正式 session。二维码图片过期独立于授权会话总时限，避免 token 刷新导致授权提前过期。
- 影响文件：新增 `internal/telegram/authorization_service.go`、`internal/telegram/authorization_service_test.go`、`internal/telegram/authorization_session_gotd.go`、`internal/telegram/authorization_session_gotd_test.go`；修改 `internal/telegram/auth_state.go`、`internal/telegram/auth_state_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、`GOPROXY=off go test -race ./internal/telegram -count=1`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；二维码登录、手机号验证码和 session 切换仍需在具备真实 Telegram 凭据的部署环境验收。待提交本层，继续采集器常驻控制端口与运行时 gate。

## 2026-09-04 14:54 +0800
- 进度：已提交 Telegram 管理状态与控制协议基线 `ec96d66`。继续实现页面化授权运行时：采集器常驻内部控制端口、手机号验证码/二次验证、二维码短期图像、临时 session 校验和原子替换，以及 30 秒心跳；之后接入管理 API、来源预解析确认、失败恢复和 Vue 工作区。
- 影响文件：预计新增/修改 `internal/telegram/` 授权服务与 gotd adapter、`cmd/telegram-ingestor/`、`internal/config/`、`internal/handlers/`、`internal/services/`、`admin-web/src/`、`deploy/docker-compose.telegram.yml`、`.env.example`、ADR、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：每层先补定向单测，阶段性运行 Telegram 相关 Go 测试与 `go vet`；前端阶段运行 `npm run build`；收尾运行差异检查、乱码扫描。真实 Telegram 授权、二维码扫描、私密邀请和 Compose 端到端仍需具备凭据的部署环境验收。

## 2026-09-04 继续 +0800
- 进度：接续 Telegram 管理台控制平面实现。已确认上轮红灯测试进程退出，开始实现 `0039` 管理状态 schema、账号/心跳/授权/审计模型与仓储；随后接入授权状态机、内部控制 HTTP 和管理 API/页面。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`internal/models/telegram_management.go`、`internal/repository/telegram_management_repository.go`、`internal/telegram/auth_state.go`、`internal/telegram/control.go` 及后续管理服务、页面和部署配置；保留用户既有 `docs/examples/` 不纳入。
- 验证：已确认 `TestTelegramManagementMigration` 与 `TestAuthorizationStateMachine|TestControlHandler` 按预期红灯；实现后分阶段运行 Go 定向测试、`go vet`、管理端构建、差异检查和乱码扫描。真实 Telegram 凭据与 Docker 端到端验收仍待部署环境。

## 2026-09-03 13:05 +0800
- 进度：开始实现 Telegram 管理台控制平面。第一阶段锁定 `0039` 账号状态、采集器心跳、短期授权元数据和追加式审计表，并建立内部控制协议、常量时间 token 校验、敏感字段脱敏和单例授权状态机的红灯测试；不改变既有 `0038` 来源/媒体表及采集队列语义。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`internal/models/telegram_management.go`、`internal/repository/telegram_management_repository.go`、`internal/telegram/control.go`、`internal/telegram/auth_state.go` 及对应测试；保留用户既有 `docs/examples/` 不纳入。
- 验证：先运行新增定向测试确认因实现缺失而失败；随后按阶段运行 Go 定向测试、`go vet`、管理端构建和差异/乱码检查。真实 Telegram 手机号、二维码、私密邀请和重启恢复继续留待具备凭据与 Docker 的部署机验收。

## 2026-09-03 12:20 +0800
- 进度：用户确认采用 Telegram 管理台全部推荐方案，开始进入实现。新增能力将以 `telegram-ingestor` 独占 session、API 仅经 Docker 内网控制通道转发为边界；引入前向兼容的账号状态/采集器心跳/审计 migration，保留现有来源 API 与队列语义。首次以手机号验证码绑定，二维码仅用于已绑定账号重新授权；高风险操作使用当前管理员密码确认与一次性短期凭据；来源采用预解析后确认、canonical chat ID 去重和来源级失败恢复。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`docs/adr/0026-telegram-ingestor-control-plane.md`、Telegram 控制/管理 service 与测试；修改 `internal/telegram/`、`cmd/telegram-ingestor/`、`internal/handlers/`、`internal/repository/`、`internal/config/`、`main.go`、`deploy/docker-compose.telegram.yml`、`admin-web/src/`、`.env.example`、`CONTEXT.md`、`plan.md`。保留用户既有 `docs/examples/` 不纳入。
- 验证：先建立 migration、控制协议、权限/脱敏和页面轮询的红灯测试；实现后运行相关 Go 测试、`go vet`、管理端 `npm run build` 与差异/乱码检查。真实手机号、二维码、私密群组、回填及重启恢复仅能在具备 Telegram 凭据与 Docker 的部署环境验收。

## 2026-09-03 11:57 +0800
- 进度：`grill-with-docs` 第五轮全部采用推荐方案。沿用现有 `admin` 权限；首次绑定、重新授权和私密邀请解析需要管理员密码确认及页面二次确认；授权会话全局单例且发起者独占秘密提交；二维码只返回短期图像；重新授权有限排空现有下载/导入；账号、采集器、来源状态分层；首期限定群组/超级群组/频道、预解析短期有效、无自定义别名；审计保留一年；普通状态 5 秒、授权状态 2 秒轮询且不重入。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅记录领域术语和设计共识，未修改运行时代码、接口或管理端页面。
- 验证：待最后一轮确认页面字段、接口响应、危险操作文案和端到端验收清单；用户确认共享理解前不进入实现。

> 2026-08-11 压缩整理版：按用户要求删除纯部署、推送、重启、健康检查和镜像同步流水；将同一事项的访谈、开始、红灯、实现、复核、待提交等过程记录合并为最终有效结论。历史精确差异与验证细节以 Git 提交、`CONTEXT.md`、ADR 和 `tasks/*/DONE.md` 为准。后续仍按反向时间顺序在顶部追加计划与进度。

## 2026-09-03 11:36 +0800
- 进度：`grill-with-docs` Telegram 管理台设计访谈完成前三轮。用户确认采用完整页面化管理、单个系统 Telegram 采集账号、固定导入归属用户、添加来源前确认并自动历史回填、暂停/恢复/重新回填、失败恢复和轮询状态；新增二维码授权，但首次绑定仍要求手机号验证码，二维码只用于重新授权或恢复。页面不编辑部署级 API ID/hash、手机号或服务生命周期；验证码和二次验证密码属于短期授权输入，不持久化。二维码重新授权必须校验首次绑定的 Telegram 用户 ID；新 session 仅在校验成功后替换旧 session。来源先预解析再确认加入、按 canonical chat ID 去重，采集器运行状态通过心跳观测，页面不启动容器。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅记录领域术语与设计共识，未修改运行时代码、接口或管理端页面。
- 验证：已核对现有 Vue 管理端、管理员认证、Telegram gotd 客户端和二维码 API 能力；待继续完成登录状态机、来源生命周期、错误恢复和页面验收边界访谈，用户确认共享理解前不进入实现。

## 2026-09-03 11:52 +0800
- 进度：`grill-with-docs` 第四轮完成。用户确认 API 通过仅内部可达的采集器控制通道转发授权交互；新增单例 Telegram 账号状态元数据与追加式管理审计；私密邀请 token 解析后丢弃；采集器以 30 秒心跳、90 秒超时提供结构化运行状态；管理台 PC 优先并保证窄视口无溢出，不展示原始采集器日志。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅补充领域语言和设计进度，未修改运行时代码、接口或管理端页面。
- 验证：已核对现有 Compose session 挂载边界、Telegram 客户端 QR 能力和管理端路由/日志现状；待继续确认控制通道认证、管理员权限、审计保留和异常恢复细节，用户确认共享理解前不进入实现。

## 2026-09-02 21:47 +0800
- 进度：Task 7 自动实现与仓库内验收完成。部署契约覆盖统一镜像、三服务共享路径、session 隔离、非 root 运行、容器服务名连接、固定容器端口、敏感构建上下文排除和手册必备操作；最终包级测试还发现并修复 Task 1 测试行扫描器缺少 Go `int` 目标支持的问题，生产仓储逻辑未改。
- 影响文件：`.dockerignore`、`.gitignore`、`.env.example`、`deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`、`CONTEXT.md`、`internal/repository/telegram_repository_test.go`、`plan.md`；提交时明确排除用户既有 `docs/examples/`。
- 验证：`go test ./deploy ./internal/repository -count=1`、计划包集与全仓测试在跳过既有 `TestParseTVAPKMetadataParsesReleaseAPK` 后通过；`go vet ./...`、两个 `CGO_ENABLED=0 GOOS=linux go build`、`git diff --check`、U+FFFD 扫描通过。未跳过时计划包集和 `go test ./... -count=1` 仅因既有 TV APK 实际 version code `144`、断言仍为 `121` 失败。当前机器没有 `docker` 命令，故 `docker compose ... config`、镜像实际构建和测试群组登录/回填/重启/实时采集需在 OrbStack 部署机按运行手册执行，未冒充已验证。

## 2026-09-02 18:41 +0800
- 进度：Task 7 核心实现完成。新增 Go 1.22 多阶段镜像和三服务 Compose，统一共享媒体/暂存路径并隔离个人 session；Compose 同步透传既有 API/worker 的外部服务配置，避免容器部署丢失原有能力；补齐中文运维手册和 Telegram 长期架构约定。Dockerfile 只复制 Go 构建必需源码，`.dockerignore` 进一步阻止本地 `.env`、session 和媒体目录进入构建上下文。
- 影响文件：`.dockerignore`、`.gitignore`、`deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`、`.env.example`、`CONTEXT.md`、`internal/repository/telegram_repository_test.go`、`plan.md`；保留用户既有 `docs/examples/` 未跟踪内容。
- 验证：`go test ./deploy -count=1`、`git diff --check` 通过，目标文件 U+FFFD 扫描无输出；本机缺少 `docker` 命令，Compose CLI 解析和真实 Telegram 测试待最终验收时如实记录。

## 2026-09-02 18:27 +0800
- 进度：Task 7 部署契约红灯已建立，覆盖三服务统一镜像/共享挂载、session 仅由非 root 采集器挂载、`ffmpeg`/`ffprobe`、采集器 `-mode run` 和 Telegram 环境变量完整性。
- 影响文件：`deploy/telegram_deploy_test.go`、`plan.md`。
- 验证：`go test ./deploy -count=1` 按预期失败，原因仅为 `deploy/docker-compose.telegram.yml`、`deploy/video-server.Dockerfile` 和三个宿主挂载变量尚未实现。

## 2026-09-02 18:24 +0800
- 进度：进入 Task 7。开始建立 OrbStack 部署契约测试，并补齐统一应用镜像、三服务 Compose、Telegram 一次性登录与日常运维手册；三个应用服务共享媒体和上传暂存容器路径，个人 session 仅由非 root 采集器持有。
- 影响文件：预计新增 `deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`，修改 `.env.example`、`CONTEXT.md`、`plan.md`；保留用户既有 `docs/examples/`，不修改 Android 工程。
- 验证：先运行 `go test ./deploy -count=1` 确认部署契约红灯；实现后执行 Task 7 定向测试、全量 Go 测试、`go vet ./...`、Compose 配置解析、差异及乱码检查。

## 2026-09-02 18:14 +0800
- 进度：Task 6 完成。管理员接口已覆盖来源列表、新增、暂停、恢复、重新回填和进度查询；新增来源统一保存规范化 `chat_ref`，恢复保留游标，重新回填清零游标，控制任务异步投递。
- 影响文件：`internal/services/telegram_source.go`、`internal/services/telegram_source_test.go`、`internal/handlers/admin_telegram.go`、`internal/handlers/admin_telegram_test.go`、`internal/handlers/router.go`、`main.go`、`plan.md`；未纳入用户既有 `docs/examples/`。
- 验证：Task 6 定向测试、handler 全量测试、`go vet` 和主程序构建通过；待提交 `增加 Telegram 群组管理接口`。

## 2026-09-02 18:02 +0800
- 进度：进入 Task 6。按已批准计划实现 Telegram 来源管理服务和管理员 API；HTTP 请求只更新 PostgreSQL 状态并投递 control 任务，群组解析与历史扫描继续由独立采集器执行。
- 影响文件：`internal/services/telegram_source.go`、`internal/services/telegram_source_test.go`、`internal/handlers/admin_telegram.go`、`internal/handlers/admin_telegram_test.go`、`internal/handlers/router.go`、`main.go`、`plan.md`；保留用户既有 `docs/examples/`。
- 验证：`go test ./internal/services ./internal/handlers -run 'TestAdminTelegram|TestTelegramSource' -count=1` 已按预期红灯，待实现后复跑定向测试、`go vet` 和主入口构建。

## 2026-09-02 17:51 +0800
- 进度：Task 5 完成。补齐暂停来源的 claim 前检查、超时下载/导入记录恢复、下载并发限制、转码与下载双重 reconcile，并新增独立 `telegram-ingestor` 的登录/运行入口；实时与回填队列按 `10:1` 权重消费。
- 影响文件：`internal/services/telegram_ingestion.go`、`internal/services/telegram_ingestion_test.go`、`internal/services/upload.go`、`internal/repository/telegram_repository.go`、`internal/queue/telegram_processor.go`、`internal/queue/telegram_processor_test.go`、`cmd/telegram-ingestor/main.go`、`cmd/telegram-ingestor/main_test.go`、`plan.md`；保留用户既有 `docs/examples/` 未跟踪内容。
- 验证：`go test ./internal/services ./internal/queue -run 'TestTelegram(Ingestion|Processor)' -count=1`、`go test ./cmd/telegram-ingestor ./internal/services ./internal/queue ./internal/telegram -run 'TestTelegram' -count=1`、`go vet ./internal/services ./internal/queue ./internal/telegram ./cmd/telegram-ingestor`、`go build ./cmd/telegram-ingestor` 通过；Task 5 包级测试仅因既有 `TestParseTVAPKMetadataParsesReleaseAPK` 断言 `121` 与当前 APK 实际版本 `144` 不一致而失败。
- 提交：待提交 `实现 Telegram 视频采集与回填`，随后进入 Task 6 管理 API。

## 2026-09-02 16:31 +0800
- 进度：进入 Task 5。开始建立 Telegram 来源同步、下载处理、三层幂等和转码补偿的服务/队列处理器边界；独立采集器入口继续保持与现有 `server`/`worker` 入口隔离。
- 影响文件：预计新增 `internal/services/telegram_ingestion.go`、`internal/services/telegram_ingestion_test.go`、`internal/queue/telegram_processor.go`、`internal/queue/telegram_processor_test.go`、`cmd/telegram-ingestor/main.go`，必要时最小修改 `main.go` 与本记录；保留用户既有 `docs/examples/`，不修改 Android 工程。
- 验证：先运行 `go test ./internal/services ./internal/queue -run 'TestTelegram(Ingestion|Processor)' -count=1` 验证红灯，再按 Task 5 完成定向测试、`go vet` 和入口构建。

## 2026-09-02 16:18 +0800
- 进度：Task 4 完成。固定 `github.com/gotd/td v0.99.2`，封装个人账号 session、chat 解析、历史分页、实时订阅、消息刷新、document 流式下载和 FloodWait 错误；元数据构造规则已覆盖 caption 标题与文件名回退。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/client_test.go`、`internal/telegram/gotd_client.go`、`internal/services/telegram_metadata.go`、`internal/services/telegram_metadata_test.go`、`go.mod`、`go.sum`、`plan.md`；未纳入用户既有 `docs/examples/`。
- 验证：`go test ./internal/telegram ./internal/services -count=1`、`go vet ./internal/telegram ./internal/services`、`gofmt -d`、`git diff --check` 通过；准备精确提交 Task 4。

## 2026-09-02 16:00 +0800
- 进度：进入 Task 4。接口、FloodWait 包装类型和 Telegram 元数据测试已建立；待固定 `github.com/gotd/td` 稳定版本并实现个人账号 session、群组解析、历史分页、实时消息订阅、消息刷新和 document 流式下载适配器。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/client_test.go`、`internal/telegram/gotd_client.go`、`internal/services/telegram_metadata.go`、`internal/services/telegram_metadata_test.go`、`go.mod`、`go.sum`、`plan.md`；保留 `docs/examples/` 未跟踪内容，不修改 Android 工程。
- 验证：现有 `go test ./internal/telegram ./internal/services -run 'Test(Client|BuildTelegramVideoMetadata)' -count=1` 待适配器完成后复跑；随后执行 `go vet`、竞态定向测试和提交前差异/乱码检查。

## 2026-09-02 15:47 +0800
- 进度：Task 3 完成。抽取 `UploadService` 的本地文件保存核心；手动上传保留重复时刷新可替换原片路径的既有语义，`SaveImportedFile` 在重复内容时不写入临时导入路径，并覆盖 hash 竞态分支。
- 影响文件：`internal/services/upload.go`、`internal/services/upload_test.go`、`internal/services/import_file.go`、`internal/services/import_file_test.go`、`plan.md`。
- 验证：`go test ./internal/services -run 'TestSaveImportedFile|TestShouldUpdateExistingVideoOriginalPath|TestShouldRefreshExistingVideoOriginalPath|TestPredictedUploadStatus' -count=1`、`go vet ./internal/services`、`git diff --check` 通过；完整 `go test ./internal/services -count=1` 仅受既有 TV APK 测试版本断言 `144 != 121` 失败，未涉及本次代码。

## 2026-09-02 15:41 +0800
- 进度：Task 2 完成。新增 Telegram API ID/hash、电话、session、导入归属用户、实时/历史/控制队列及下载限流配置；`Load` 保持 Telegram 配置可选，`ValidateTelegramIngestor` 单独校验采集器凭据、UUID 和并发边界；新增仅携带 UUID 的 Asynq 采集任务和默认队列。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`internal/queue/telegram_tasks.go`、`internal/queue/telegram_tasks_test.go`、`.env.example`、`plan.md`。
- 验证：`go test ./internal/config ./internal/queue -run 'Test(LoadIncludesTelegram|LoadDefaultsTelegram|ValidateTelegram|TelegramTask|NewTelegram)' -count=1`、`go test ./internal/config ./internal/queue -count=1`、`go vet ./internal/config ./internal/queue`、`git diff --check` 均通过；待提交 `增加 Telegram 配置与采集队列`。

## 2026-09-02 15:24 +0800
- 进度：根据 `docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md` 开始实现 Telegram 指定群组视频采集、历史回填和管理 API。当前进入 Task 1，先建立可回滚 migration、来源/媒体模型和仓储边界；保持个人 MTProto session 独占、短视频导入、三层幂等及队列/数据库状态分工。
- 影响文件：预计涉及 `migrations/0038_*`、`internal/models/telegram.go`、`internal/repository/telegram_*`、`internal/config`、`internal/queue`、`internal/services`、`internal/telegram`、`internal/handlers`、`cmd/telegram-ingestor`、`deploy`、`docs/telegram-video-ingestion.md`、`CONTEXT.md`、`plan.md`；保留既有 `docs/examples/` 和其它工作区差异，不修改 Android 工程。
- 验证：待按 Task 1 至 Task 7 逐项执行定向测试、`go vet`、构建、迁移/Compose 静态检查及最终全量验证；每个任务完成后独立使用中文提交。

## 2026-09-02 15:32 +0800
- 进度：Task 1 完成。新增 `0038` Telegram 来源/媒体 migration 及可回滚 down migration，建立来源、消息、document ID、SHA-256、导入/转码状态、重试字段和索引；新增模型与 VideoRepository 仓储边界，支持消息幂等登记、同事务游标推进、行锁 claim、部分更新、转码补偿查询和状态统计。
- 影响文件：`migrations/0038_telegram_video_ingestion.up.sql`、`migrations/0038_telegram_video_ingestion.down.sql`、`internal/models/telegram.go`、`internal/repository/telegram_repository.go`、`internal/repository/telegram_repository_test.go`、`internal/repository/migrations_test.go`、`plan.md`。
- 验证：`go test ./internal/repository -run 'TestTelegram|Test.*Migration' -count=1`、`go test ./internal/repository -count=1`、`go vet ./internal/repository`、`git diff --check` 均通过；待提交 `新增 Telegram 采集状态模型`。

## 2026-09-02
- 进度：完成 Telegram 指定群组视频采集与全部历史回填的设计确认，生成实现计划 `docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md`。方案固定多群组数据库来源、实时/历史独立下载队列、`(source_id,message_id)` + Telegram document ID + `SHA-256/文件大小` 三层幂等、短视频导入及转码投递补偿；尚未修改业务代码。
- 影响文件：`docs/superpowers/specs/2026-09-02-telegram-video-ingestion-design.md`、`docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md`、`plan.md`；不修改 Android/TV 版本号。设计提交为 `b187a8e`，实现计划待确认执行方式。
- 验证：设计文档自检通过；计划完成占位词、接口一致性和 `git diff --check` 检查，待用户选择按任务执行方式后进入实现。

## 2026-08-14 09:25 +0800
- 进度：论坛资源改动完成提交前复核，代码、测试、接口契约、ADR 与长期上下文一致；提交仅纳入本任务文件及本任务 `plan.md` 记录，保留既有历史重抓记录和 `docs/examples/` 工作区内容。
- 影响文件：本任务文件同 09:24 记录；不修改 Android/TV 版本号。
- 验证：`git diff --check`、Go 格式检查及本任务中文文件 U+FFFD 扫描通过，暂存后继续核对 staged diff 并提交。

## 2026-08-14 09:24 +0800
- 进度：论坛资源标题搜索、受限帖子展示及保留规则实现完成。管理员列表接口新增最多 200 字符的标题字面子串查询，投影纳入并标记 `restricted + included`；有效 Hermes `discover` 触发的 30 天清理现在保护当前 `pending` 与 `restricted`，其余无资源旧帖规则不变。管理端采用独立搜索草稿和显式提交，刷新/翻页保留已提交查询，清空恢复完整列表；长期契约已同步至 `CONTEXT.md` 与 ADR-0025。
- 影响文件：`internal/models/hermes_forum.go`、`internal/handlers/hermes_forum.go`、`internal/handlers/hermes_forum_test.go`、`internal/services/hermes_forum.go`、`internal/services/hermes_forum_test.go`、`internal/repository/hermes_forum_repository.go`、`internal/repository/hermes_forum_repository_test.go`、`internal/repository/hermes_forum_repository_integration_test.go`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList.vue`、`admin-web/src/views/ForumPostList.spec.js`、`CONTEXT.md`、`docs/adr/0021-admin-forum-resource-list.md`、`docs/adr/0023-hermes-resource-aware-retention.md`、`docs/adr/0025-admin-forum-restricted-search-retention.md`、`plan.md`；不修改 Android/TV 版本号，不纳入既有 `docs/examples/` 与无关 `plan.md` 记录。
- 验证：论坛定向 Go 测试与对应 `-race` 通过，`go vet ./...` 通过；管理端定向 36 项、全量 705 项测试及 `npm run build` 通过；浏览器桌面与 390px 视口验证搜索提交、空态、清空恢复、受限标记、容器滚动及控制台均通过。PostgreSQL 集成测试因未设置 `HERMES_TEST_DATABASE_URL` 按既有机制跳过，`golangci-lint` 本机未安装；`go test ./... -count=1` 的无关 TV APK 固件版本 144/旧断言 121 失败，整包 `-race` 的无关并行 Gin `SetMode` 基线竞态仍存在。待执行提交前差异、乱码与暂存范围检查。

## 2026-08-14 09:12 +0800
- 进度：论坛资源标题搜索与受限帖子展示/保留红灯已建立。Go 定向测试因旧代码缺少搜索参数、受限状态字段、查询构造及超长错误类型而按预期编译失败；管理端定向测试 36 项中 4 项按预期失败，覆盖搜索控件与提交状态、受限标记和搜索空态，API 请求包装既有测试保持通过。
- 影响文件：`internal/handlers/hermes_forum_test.go`、`internal/services/hermes_forum_test.go`、`internal/repository/hermes_forum_repository_test.go`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList.spec.js`、`plan.md`。
- 验证：红灯命令为 `go test ./internal/handlers ./internal/services ./internal/repository -run 'TestAdminForumPosts|TestHermesForumService(ListsAdminReadModel|RejectsLongAdminSearchQuery)|TestBuildAdminForumPostListSQL|TestDiscoverForumPostsCleanup' -count=1` 与 `npm test -- --run src/api/admin.spec.js src/views/ForumPostList.spec.js`；待实现后复跑转绿。

## 2026-08-14 09:08 +0800
- 进度：用户确认 PC 管理端论坛资源完整方案，进入实现。列表新增服务端标题字面子串搜索并展示 `restricted + included`，标题旁标记“受限”；当前为 `pending` 或 `restricted` 的记录不再参与 30 天无资源清理，豁免仅跟随当前状态，不恢复已删除历史。先补 Handler、Service、Repository 与 Vue 页面红灯测试，再做最小实现。
- 影响文件：预计涉及 `internal/models/hermes_forum.go`、`internal/handlers/hermes_forum*`、`internal/services/hermes_forum*`、`internal/repository/hermes_forum*`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList*`、`CONTEXT.md`、`docs/adr/0021-admin-forum-resource-list.md`、`docs/adr/0023-hermes-resource-aware-retention.md`、`docs/adr/0025-admin-forum-restricted-search-retention.md`、`plan.md`；不修改 Android/TV 版本号。
- 验证：待执行后端定向红灯与回归、`go test ./... -count=1`、`go test -race` 定向包、`go vet ./...`、管理端定向与全量 Vitest、`npm run build`、页面桌面/窄视口检查、`git diff --check` 与中文乱码扫描。

## 2026-08-13 17:24 +0800
- 进度：中国行政区划地图生产解析故障修复完成。Go 管理端静态服务现在于 SPA 回退前原样提供 `/admin/china-map/**` 下的 `.json` 与 `.geojson`，设置正确 MIME 和每次复验缓存策略；缺失、越界或非允许扩展统一 404，不再返回 `index.html`。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`CONTEXT.md`、`plan.md`；无关 Hermes 记录与 `docs/examples/115LocalNatManager/` 不纳入提交。
- 验证：静态路由红灯已复现并转绿；`go test ./internal/handlers -count=1`、定向 `-race`、`go vet ./...`、`go build ./...`、管理端 46 个文件 704/704、`npm run build`、`npm run china-map:validate`、差异/乱码/控制字符检查通过。`go test ./... -count=1` 仅有既有 TV release APK 元数据漂移失败：产物 `versionCode=144`，测试仍期望 `121`，未修改无关 TV 产物或测试。待提交、推送部署机并按生产 JSON/GeoJSON 响应和浏览器地图加载复验。

## 2026-08-13 17:19 +0800
- 进度：修复中国行政区划地图部署后数据解析失败。生产取证确认 `/admin/china-map/catalog.json` 与 `manifest.json` 被 Go 管理端静态服务错误回退为 `index.html`；先补后端红灯测试，再为地图 JSON/GeoJSON 增加受路径边界保护的静态文件服务，缺失文件保持 404，不得伪装成 SPA 页面。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：待执行后端定向测试、全量 Go 测试与 `go vet ./...`、管理端构建/地图校验、生产响应 Content-Type/正文及浏览器地图加载复验；完成后提交并推送部署机。

## 2026-08-13 15:32 +0800
- 进度：中国行政区划地图完成提交前收口。启动数据明确为目录与清单必需、南海诸岛附图可选，附图加载失败不阻塞主地图；地图信息弹层补齐打开后聚焦、Tab 焦点约束、Esc 关闭和触发按钮焦点恢复。无关 Hermes 记录与 `docs/examples/115LocalNatManager/` 保留在工作区，不纳入本次提交。
- 影响文件：`admin-web/src/views/ToolboxChinaMap.vue`、`admin-web/src/views/chinaMap.data.js`、对应测试、地图功能既有代码与快照、`CONTEXT.md`、`plan.md`。
- 验证：定向回归 19/19、全量 Vitest 46 个文件 704/704、`npm run build`、`npm run china-map:validate` 和浏览器控制台检查通过；1024×768 实测画布与视口同为 1024×768、画布像素非空、悬浮控件无重叠，弹层焦点闭环正常；模拟 `CN-63` 图层首次网络失败时保留全国地图，点击重试后恢复“中国 / 青海省”。构建仅有项目既有的大包体积警告。

## 2026-08-13 15:22 +0800
- 进度：中国行政区划地图完成代码、数据和浏览器收尾。新页面作为 `/toolbox/china-map` 管理员认证独立工作区接入工具箱，画布覆盖完整视口；支持全国、省、地、县可变层级钻取、末级高亮、全局搜索、URL/历史恢复、缩放复位、错误重试、数据说明及不可交互南海诸岛附图。正式提交固定 OSM 时间点的 34 个省级范围与港澳台分层快照，并补齐生成、完整性校验、语义色门禁、首屏返回状态和窄屏按钮可访问性。
- 影响文件：`admin-web/package*.json`、`admin-web/src/App.vue`、`admin-web/src/assets/theme.css`、`admin-web/src/router/*`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxChinaMap.vue`、地图 helper/loader 与测试、`admin-web/scripts/china-map/*`、`admin-web/public/china-map/*`、`docs/adr/0024-admin-china-administrative-map-data.md`、`CONTEXT.md`、`plan.md`。
- 验证：定向 Precision Ops 门禁 227/227、地图与工具箱回归 20/20、全量 Vitest 702/702、`npm run build`、`npm run china-map:validate`、`git diff --check`、U+FFFD 与控制字符扫描均通过；数据门禁为 3631 个区域、34 个省级范围、3245 个末级区域、386 个图层、3630 个面要素。Chrome 验收覆盖 1440×900 与 375×812：画布非空且等于视口，控件无重叠，搜索“乌丘”可定位并恢复完整路径，URL/浏览器后退正常，控制台无报错；无关 Hermes 记录与 `docs/examples/` 不纳入本次提交。最终全量复验与独立审阅待提交前完成。

## 2026-08-13 14:57 +0800
- 进度：中国行政区划地图主体与正式数据已完成，收尾复现全量测试剩余的 5 个 Precision Ops 门禁失败：新视图尚未登记到独立页面边界、工具箱数量仍锁定为 5，以及地图视图存在直接颜色、3 处常驻面板阴影和 2 处直接百分比圆角。按现有门禁修复页面和接入清单，不放宽审计规则。
- 影响文件：`admin-web/src/assets/theme.css`、`admin-web/src/views/ToolboxChinaMap.vue`、`admin-web/src/views/precisionOpsRollout.spec.js`、`plan.md`。
- 验证：定向 `npm test -- --run src/views/precisionOpsAudit.spec.js src/views/precisionOpsRollout.spec.js` 已按预期出现 5 项红灯；待重跑定向与全量测试、管理端构建、地图数据校验和浏览器多视口验收。

## 2026-08-13 10:18 +0800
- 进度：用户确认中国行政区划地图完整方案，进入管理端实现。采用现有 Vue 3、Vue Router、ECharts 与 Element Plus；新增 `/toolbox/china-map` 认证独立路由和工具箱新标签页入口，页面保持无 shell 且地图占满视口。实现按静态索引与父级拆分 GeoJSON 契约分离层级、搜索、URL 恢复、缓存和错误处理，并先补红灯测试；正式数据只接受许可与来源可追溯的快照，不把临时无许可证样本纳入仓库。
- 影响文件：预计涉及 `admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxChinaMap.vue`、地图 helper/测试、`admin-web/src/router/*`、`admin-web/public/china-map/*`、数据构建脚本与说明、`CONTEXT.md`、ADR-0024、`plan.md`。
- 验证：待执行管理端定向 Vitest、全量 `npm test`、`npm run build`、数据完整性校验、`git diff --check`、中文乱码扫描，以及 1024/1440/窄视口浏览器截图核验。

## 2026-08-13 09:42 +0800
- 进度：开始处理 Hermes 权限误判修复后的历史重抓。目标为 `source=sehuatang` 且 `inspection_status=restricted` 的最终帖子；先只读统计并按旧权限误判、已有资源和真实登录/验证受限分类，再备份数据库行、资源行、Hermes 任务定义和相关状态，暂停新帖与历史补漏调度后逐条执行 `discover -> cdp_fetch -> enrich_thread -> inspection`。仅新发现资源的帖子允许产生通知，避免重复投递。
- 影响文件：`plan.md`；运行时目标为远程 Hermes Forum API、PostgreSQL、`~/.hermes/cron/jobs.json` 与 `~/.hermes/state/`，不修改 Android/TV 版本号。
- 验证：待完成 restricted 统计、备份完整性、重抓结果、通知队列清理及调度恢复核对。

## 2026-08-13 09:50 +0800
- 进度：按用户确认改用状态机恢复方案：将 `source=sehuatang` 的 150 条 `restricted` 终态重置为 `pending`，保留帖子身份和观察时间，清空旧检查结论与指纹；恢复两个论坛 cron，让其按既有 `discover -> inspection` 流程重抓。由于历史补漏游标可能已越过旧页，重置后核对并在必要时把明确候选加入持久队列。
- 影响文件：`plan.md`；运行时目标为远程 `collected_forum_posts` / `collected_forum_post_resources`、Hermes 任务定义和补漏队列。
- 验证：待执行事务行数与约束核对、cron 入队/检查进度、资源结果、重复通知防护及任务恢复状态检查。

## 2026-08-13 09:31 +0800
- 进度：修复 Hermes 将附件提示“阅读权限: 10”误判为整页权限门禁的问题。外部活动脚本 `~/.hermes/scripts/watch_sehuatang_forum95.py` 仅移除 `is_blocked_or_permission_text()` 中的裸“阅读权限”词项，保留明确的“需要阅读权限”“阅读权限高于”“本帖隐藏的内容需要”及登录/验证门禁；已生成可回滚备份 `~/.hermes/scripts/watch_sehuatang_forum95.py.pre-attachment-permission-fix-20260813-0920`。
- 影响文件：外部 Hermes 脚本及其回滚备份；仓库仅追加 `CONTEXT.md`、`docs/adr/0020-hermes-forum-post-ingestion.md`、`plan.md`，既有 `docs/adr/0024-admin-china-administrative-map-data.md` 与 `docs/examples/` 不纳入。
- 验证：临时回归 `/tmp/hermes_attachment_regression_test.py` 修改前按预期红灯，修改后通过；Hermes venv `py_compile` 通过；真实 CDP 页面和 `/tmp/hermes-2526602-cdp.json` 均确认正文含“阅读权限: 10”时 `cdp_fetch()` 为 `ok=true`，`extract_attachments()` 返回 `mod=attachment` 链接；09:21 新帖任务保持 `last_status=ok`。只读数据库核对显示 `tid=2526602` 当前历史终态仍为 `restricted/included`、附件数 0；根据检查结果不可覆盖约束，本次不直接改写记录、不补发通知，待明确授权后另行走可回滚修复流程。

## 2026-08-11 16:42 +0800
- 进度：`grill-with-docs` 数据治理轮全部采用推荐方案。行政边界以固定版本 OpenStreetMap 快照为几何来源，结合可追溯公开中文行政区目录离线规范化并生成分层 GeoJSON；生成数据遵守 ODbL 归属和相同方式共享要求，页面运行时不访问第三方地图服务。区域统一使用仓库稳定 `region_code` 作为“区划标识”，来源行政编码另存且允许为空。数据构建对完整省级范围与港澳台、父子闭合、标识唯一、中文名、有效几何和末级可达实行失败即停止门禁；快照只允许人工触发、审核差异后随管理端版本发布。新增 ADR-0024 固化该难以逆转的数据选择。
- 影响文件：`CONTEXT.md`、`docs/adr/0024-admin-china-administrative-map-data.md`、`plan.md`；生产代码仍未修改。
- 验证：待用户确认完整共同理解后执行文档差异、乱码与术语一致性检查并提交；确认前不进入实现。

## 2026-08-11 15:53 +0800
- 进度：`grill-with-docs` 第三轮全部采用推荐方案。静态行政区划按父级编码拆分并逐层懒加载；切层时保留旧图和轻量加载反馈，数据失败保留上一层并提供重试/返回；县级浮层只显示名称、完整路径和编码；搜索支持即时候选、键盘选择、回车定位和无结果反馈；全国视图包含不参与钻取的南海诸岛附图；层级动画为 200–260ms 且尊重减弱动效；数据来源、版本和许可放入按需地图信息浮层；缓存仅使用静态资源 HTTP 缓存与会话内存。
- 影响文件：`CONTEXT.md`、`plan.md`；生产代码仍未修改。
- 验证：数据源核查确认现成数据均有关键缺口：无许可证、离线再分发条款不明确，或虽开放许可但中国县级数据代表年份为 2017、英文名称且无行政区编码。需要最后确认采用可追溯数据构建方案及编码口径。

## 2026-08-11 15:37 +0800
- 进度：`grill-with-docs` 第二轮全部采用推荐方案。县级点击仅高亮并展示完整行政路径；直辖市、港澳台等按真实父子关系使用可变层级；面包屑可跳祖先、地图返回逐级回退、浏览器历史与 URL 行政区编码同步，退出工具箱独立；全局搜索覆盖省/地/县并以完整路径消歧；标签只显示当前层级且碰撞时隐藏、悬停可查；地图沿用管理端中性语义色与暖金选中态，支持滚轮缩放、拖拽、按钮缩放和复位；正式支持不低于 1024px 的 PC，窄窗口只保证基本可用。
- 影响文件：`CONTEXT.md`、`plan.md`；仍未修改管理端生产代码、依赖、路由或 App 版本。
- 验证：行政区划数据来源事实核查重新执行中；待完成加载、异常、数据版本和验收边界的最后一轮访谈。

## 2026-08-11 09:58 +0800
- 进度：`grill-with-docs` 第一轮确认地图定位。页面是只读的完整中国行政区划浏览工具，包含台湾，不承载业务统计、热力、选择、绑定或编辑；按全国省级轮廓 → 地级层级 → 县级行政区逐级进入，“县级行政区”包含市辖区、县级市、自治县、旗等同级单位。行政区划 GeoJSON 随前端版本静态打包，入口位于工具箱并以新标签页打开，地图浮层提供返回、当前位置、搜索、缩放和复位。
- 影响文件：`CONTEXT.md`、`plan.md`；尚未修改 `admin-web` 代码、路由、依赖或 App 版本。
- 验证：已核对管理端现有 ECharts 5.6 依赖与无 shell 工具页路由模式；行政区划数据来源、许可和体积正在核查，待完成后继续访谈与文档静态检查。

## 2026-08-11 09:51 +0800
- 进度：开始规划 PC 管理端中国行政区划地图页面。已确认页面采用 `admin 全屏地图工作区`：无后台 shell，地图画布占满浏览器视口，返回、搜索和层级操作只能以地图浮层呈现；尚未决定地图用途、县级交互、数据新鲜度和入口位置。
- 影响文件：`CONTEXT.md`、`plan.md`；尚未修改 `admin-web` 代码、路由或依赖。
- 验证：待完成需求访谈后执行文档差异、乱码和结构检查。

## 2026-08-11 09:18 +0800
- 进度：完成 `plan.md` 压缩清理，并明确今后纯部署等非开发操作无需写入开发账本。1133 条、5707 行历史流水按事项和月份归并；纯部署记录删除，已完成事项只保留最终结论，未发现仍待验收的任务目录。
- 影响文件：`AGENTS.md`、`plan.md`；既有未跟踪 `docs/examples/` 不纳入。
- 验证：`git diff --check -- AGENTS.md plan.md` 通过；U+FFFD 与 C0/DEL 控制字符扫描无输出；确认 `tasks/` 下 13 个任务目录均有 `DONE.md`；`plan.md` 压缩为 88 行，工作区范围仅含 `AGENTS.md`、`plan.md`，既有未跟踪 `docs/examples/` 未纳入。

## 当前状态

- 当前没有进行中的仓库实现任务；`tasks/` 下现有 13 个任务目录均已有 `DONE.md`。
- 最近一次功能提交为 `3262790`（修复 Hermes 编辑后资源漏检）。
- 工作区既有未跟踪 `docs/examples/` 不属于本次清理范围。

## 2026-08-09：Hermes 论坛采集可靠性

- 完成编辑后补资源的延迟定稿：标题含 `115` 或 `ed2k` 且首次无资源时保持 `pending`，按 30 分钟、1 小时、2 小时、之后每 6 小时复查；普通空帖立即定稿，历史最终态不追溯重开。7 项隔离回归、脚本编译和真实调度验证通过（`3262790`）。
- 补录漏通知帖子 `tid=3680933`，重新检查后以 `included` 入库 1 个 115 附件并完成单次通知，任务和队列恢复正常（`6fcf4c2`）。
- 修复历史补漏游标停滞：公开列表页的通用权限文案不再误判为访问受限，真实抓取失败不推进游标，空页或无待处理项正常推进；第 1991 页 30 条候选完成处理（`4646a37`）。
- 修复新帖监控遗漏：候选、检查和通知改为可恢复队列，只有明确投递成功才出队；抓取或投递失败非零退出并保留重试，故障恢复和幂等回归通过（`417a6cf`）。

## 2026-08-08：Hermes 队列与保留策略

- 历史补漏改为持久化稳定候选队列，检查成功后才移除，队列清空后才翻页；既有 102 条 `pending` 完成恢复处理，真实调度验证队列可跨批次续跑（`fce95b0`）。
- 论坛记录采用资源感知保留策略：超过 30 天且无附件、无 ED2K 才清理；30 天内记录和有资源的旧记录保留，管理端读取使用同一边界。仓储单测、可选 PostgreSQL 集成测试和 `go vet` 通过（`15e3cc0`）。

## 2026-08-07：TV App 精简

- 新版 TV App 完整移除 IPTV 客户端入口、路由、页面、状态、客户端 API/Repository 和专属资源；后端、管理端、频道数据、旧 APK、通用 LibVLC 与双 ABI 分发继续保留。TV 版本升至 `0.1.145 (145)`，617 项单测和 Debug 双 ABI 构建通过（`87ff990`）。

## 2026-08-06：手机 App 精简

- 手机端移除电影、电视剧入口与专属实现，只允许 `short`、`av` 内容进入列表、详情和播放器；混合来源支持跨页补读，异常旧入口提示不可用并回退。TV 扫码授权、短视频投屏和遥控链路保留。手机版本升至 `0.1.10 (11)`，165 项单测和 Debug 构建通过（`509ba89`）。

## 2026-08-04：论坛资源管理

- 新增管理端论坛资源列表与详情能力，统一展示帖子、附件和 ED2K 资源，并补齐后端读取模型、前端交互、ADR 和领域术语；管理端构建与 Go 定向验证通过（`09875fb`）。
- 完成 Hermes 两阶段论坛采集接口与脚本切换：以 `source + tid` 去重，区分 `dedupe_only`、`pending` 和最终检查态，支持多附件/多 ED2K 与机器 Token 鉴权；历史基线成功导入且重复导入幂等（`0663a8d`、`e07924b`）。

## 2026-08-02：文件大小与管理端标签

- 转码完成后记录主播放文件大小，新增历史回填命令并保持未知值为 `NULL`；服务层、仓储和命令测试通过（`c40e134`、`a82e3f5`）。
- 修复管理端视频标签远程候选，统一已有标签选择与新标签输入语义，管理端构建通过（`5a122cf`）。

## 2026-08-01：转码与资源链路修复

- 修复转码封面流误选、异常媒体流探测、进度假失败、AVC 硬编超长边和源文件兼容等问题；补充 ffmpeg 计划测试与长期兼容约束。
- 收口视频、图片、演员、合集等管理端资源链路的状态反馈、分页、预览与危险操作保护；定向测试和管理端构建通过。

## 2026-07：管理端、TV 与 115 能力

- 完成管理端 Precision Ops 分阶段改造：统一 compact 页面壳、加载/错误/空态、Drawer、分页、状态指示器、危险确认和键盘可达性；修复旧响应覆盖、查询身份错配、窄屏溢出、长文本裁剪等问题。各阶段均以定向测试、全量测试和生产构建收口。
- 完成 TV 首页、海报墙、短视频、长视频详情/播放器、焦点、字幕音轨、软刷新和状态反馈的多轮优化；相关 App 更新均同步递增版本并完成 TV 单测或构建验证。
- 完成 ED2K 下载工作台与 aMule 退役，保留浏览器端 ED2K 链接生成器和已入库成品；数据库历史表保持前向兼容，不再保留运行时代码与部署资产（`3c8fdd6`）。
- 完成 115 开放平台 OAuth/PKCE、令牌刷新、云下载与本地暂存链路设计及实现边界，保持“云下载完成”与“本地落盘/媒体入库”语义分离。
- 完成压缩包导入的编码纠偏、多分组、批量编辑/处理、稳定源路径、删除和去重兼容等能力；后端、管理端和相关文档验证通过。

## 2026-06：播放器、压缩包导入与管理端工具

- TV 长视频播放器完成软准备、软刷新、重试取消、焦点守卫、LibVLC 迁移和音轨/字幕偏好恢复；对应任务均已通过自动化与用户验收并写入 `DONE.md`。
- TV 短视频迁入独立全屏播放页，后续补齐单条失败非阻塞切换、播放提示、返回焦点和导航语义；相关 TV 版本、测试和构建同步更新。
- 压缩包导入完成从上传、扫描、批次详情、分组、编码纠偏到视频/图片处理的完整工作流；错误文本统一为合法 UTF-8，ZIP 编码模式与 RAR/7z 外部解压器边界明确。
- 管理端工具箱完成自适应布局和密码管理简单 CRUD；不引入乐观锁、导入导出、密码生成或额外验证。
- 修复 TV release APK 元数据测试随固件版本漂移、视频稳定源路径、WebP/压缩包处理和管理端页面交互等回归。

## 2026-05：Android/TV 主功能与任务闭环

- 建立独立 Android TV 工程，完成扫码配对、首页分类、搜索、海报墙、长视频/电视剧/IPTV 播放、字幕音轨、断点续播、焦点与遥控交互，并持续按功能更新版本。
- 手机端完成短视频、AV、电视剧、图集、演员详情、搜索、喜欢/收藏、统一播放器和 TV 授权协作等主链路；后续精简边界见 2026-08 记录。
- 后端完成电影/电视剧/AV 转码码率、字幕抽取、播放源、历史记录、刮削与媒体关联等多轮修复。
- 管理端完成演员、合集、图片、视频状态、手动刮削、APK 分发和工具页等能力，并统一分页、反馈与中文文案。
- `tasks/2026-05-23-*`、`tasks/2026-05-24-*`、`tasks/2026-05-25-*`、`tasks/2026-05-26-*` 均已通过对应 `review.md` 和用户验收，完成标记见各目录 `DONE.md`。

## 2026-04：项目基础能力

- 建立 Go 服务、PostgreSQL migration、管理端 Web、上传/分片/秒传、异步刮削、转码任务、签名播放与 Swagger/OpenAPI 基础链路。
- 建立视频、图片、演员、标签、合集、电视剧和 AV 领域模型及管理能力；补齐删除级联、并发竞态、空值扫描、源文件保留和失败恢复。
- 创建手机 Android 首版，完成服务发现、登录、分类、短视频竖滑播放、AV 详情、图集和统一长视频播放器。
- 统一中文 Markdown/界面文案、UTF-8 无乱码、开发脚本、环境文件加载和本地启动约定。

## 历史追溯规则

- 精确实现差异：`git log -- plan.md` 定位对应提交，再用 `git show <commit>` 查看。
- 长期术语与兼容约束：查阅 `CONTEXT.md`。
- 架构取舍：查阅 `docs/adr/`。
- 有 PRD/实现/评审流程的任务：查阅 `tasks/<任务>/prd.md`、`implement.md`、`review.md` 和 `DONE.md`。
- 本文件不再记录纯部署流水；只有部署暴露新的长期兼容约束或未解决阻塞时，才记录结论。
## 2026-09-29 计划：手机端短视频合集支持显式名称搜索
- 进度：需求访谈已确认。仅改手机端短视频合集目录及其后端接口；新增可选 `q` 参数，按名称做去首尾空格、大小写不敏感的字面子串匹配，最多 200 个字符；空词恢复完整目录，搜索提交后从第一页开始，翻页/刷新沿用已提交词；取消置顶能力，不新增字段、管理配置、特殊排序或标注。手机端只保留搜索草稿，显式提交后请求；旧请求响应不得覆盖新查询状态；搜索失败按已有列表/无列表分别反馈。手机端版本按规则升级。
- 影响文件：`internal/handlers/app_short_collection.go`、`internal/services/app.go`、`internal/repository/collection_repository.go`、对应 Go 测试；`android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`core/repository/VideoRepository.kt`、`feature/shortcollections/ShortCollectionsViewModel.kt`、`ShortCollectionsScreen.kt`、对应测试；`android-app/app/build.gradle.kts`、`CONTEXT.md`。
- 验证：先运行后端与手机端新增定向测试确认红灯，再完成实现；收尾运行 Go 定向测试、`go vet ./...`、手机端 `:app:testDebugUnitTest` 与 `:app:assembleDebug`，并检查乱码、差异和提交范围。

## 2026-09-29 实施：手机端短视频合集支持显式名称搜索
- 进度：后端 `GET /api/v1/short/collections` 接受可选 `q`，服务层按 200 Unicode 字符限制并去首尾空格，仓储层对合集名称执行大小写不敏感字面子串匹配并转义 `%`、`_`、反斜杠；缺省 `q` 保持旧客户端完整目录行为。手机端新增搜索草稿/已提交词、显式按钮与键盘 Search 提交、空词恢复、分页/刷新沿用已提交词、请求代次 latest-wins、搜索空态和错误态；未实现置顶能力。
- 影响文件：`internal/handlers/app_short_collection.go`、`internal/services/app.go`、`internal/repository/collection_repository.go`、`internal/services/app_test.go`、`internal/repository/collection_search_test.go`；`android-app/app/src/main/java/com/chee/videos/core/network/ApiService.kt`、`core/repository/VideoRepository.kt`、`feature/shortcollections/ShortCollectionsViewModel.kt`、`ShortCollectionsScreen.kt`、测试替身与 `ShortCollectionsSearchStateTest.kt`；`android-app/app/build.gradle.kts`、`CONTEXT.md`。
- 验证：新增 Go/Kotlin 定向测试已按 TDD 先红后绿；手机端 `:app:testDebugUnitTest` 通过。Go 受影响包测试中既有 `TestParseTVAPKMetadataParsesReleaseAPK` 仍因仓库当前 APK versionCode 144 与旧断言 121 失败，其余 handlers/repository 测试通过；待执行 `go vet ./...`、Android assemble、乱码与提交前差异检查。

## 2026-09-29 验证：手机端短视频合集搜索
- 进度：实现完成，未增加置顶能力。手机端版本已升级为 `versionCode 13`、`versionName 0.1.12`。
- 验证：`go test ./internal/handlers ./internal/repository -count=1` 通过；`go vet ./...` 通过；`cd android-app && ./gradlew --no-daemon :app:testDebugUnitTest` 通过；`cd android-app && ./gradlew --no-daemon :app:assembleDebug` 通过；本次涉及文件乱码扫描通过，`git diff --check` 待提交前复核。`go test ./internal/services ./internal/repository ./internal/handlers -count=1` 中 services 仍有既有 `TestParseTVAPKMetadataParsesReleaseAPK` 失败（当前 APK versionCode 144、旧断言 121），与本任务无关。

## 2026-10-14 RAR 新版压缩算法回退与错误诊断
- 进度：基于真实 18GB RAR5 报错确认 `7zz 26.03` 不支持 `Method = v6:32M...`；新增 Unsupported Method 错误分类与中文收敛诊断，避免 49 条逐项错误、7zz 横幅或密码误判。RAR/7z 在 7zz 返回该分类时清理半成品并尝试 `unar`/`unrar` 回退，失败保留原始压缩算法不兼容诊断。
- 影响文件：`internal/services/archive_import.go`、`internal/services/archive_import_test.go`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：`go test ./internal/services -count=1 -run 'TestArchiveCommandError|TestExtractArchive'` 通过；`go test ./internal/services -count=1 -run 'TestArchive|TestExtractArchive|TestScanArchive'` 通过；`git diff --check` 通过。部署机需安装 `unar` 或 `unrar`（`brew install unar`）后重试该批次。

## 2026-10-14 修复失败压缩包无法重试解包
- 进度：定位到前端 `canRetryArchiveExtract` 仅允许 `needs_password` / `needs_encoding`，导致 `failed` 批次既不显示重试按钮也不显示补密码卡片，但后端 `/retry-extract` 本身支持任意非处理中批次。计划将该判定扩展到 `failed`，并让失败批次的说明文案/密码占位按是否可补密区分；补前端契约测试锁定失败批次可重试。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`。
- 验证：`npm test -- --run src/views/ToolboxArchiveImport.spec.js` 通过（31/31）；`npm run build` 通过；`git diff --check` 通过。后端 `/retry-extract` 无需改动；部署机可直接重新解包失败批次。

## 2026-10-14 排查大压缩包上传接口 ERR_EMPTY_RESPONSE
- 进度：部署机日志显示 `/admin/archive-import/upload` 实际返回 HTTP 200，耗时 7m48s/7m44s；`ERR_EMPTY_RESPONSE` 来自浏览器端连接中断，而不是服务端崩溃。上传本身已完成 18GB 落盘、复制与批次入库；卡在同步解压，期间批次状态停在 `uploaded`，无子进程、无磁盘活动、无 CPU，疑似 RAR5 QuickOpen/压缩算法路径在外部解压器启动或等待处卡死。已关闭部署机低功耗模式复测观察。
- 影响文件：暂无代码变更；记录 `plan.md`。
- 验证：检查 server/worker 日志、PostgreSQL 批次状态、进程表、lsof 与 iostat；确认 `/healthz` 正常。后续需把上传后的同步解压改为可观测状态或异步任务，避免浏览器与代理等待长请求。

## 2026-10-14 新增解包中状态约束
- 进度：修复异步解压入队时 `extracting` 违反既有 status check 的问题，新增 migration 0041 扩展 `archive_import_batches_status_check`。部署机迁移、重启和健康检查通过。
- 影响文件：`migrations/0041_archive_import_extracting.up.sql`、`migrations/0041_archive_import_extracting.down.sql`、`plan.md`。
- 验证：`git push deploy master` 触发构建、迁移与 `/healthz OK`。

## 2026-10-14 允许 uploaded 批次重试解压
- 进度：异步解压上线后，历史卡住的批次状态是 `uploaded`，但前端重试入口此前只覆盖 `failed / needs_password / needs_encoding`，导致没有按钮。已把 `uploaded` 纳入 `canRetryArchiveExtract`，并更新契约测试。
- 影响文件：`admin-web/src/views/ToolboxArchiveImport.vue`、`admin-web/src/views/ToolboxArchiveImport.spec.js`、`plan.md`。
- 验证：`npm test -- --run src/views/ToolboxArchiveImport.spec.js` 通过（31/31）。

## 2026-10-14 压缩包导入异步解压
- 进度：按用户确认，把上传与重试后的同步解压改为 asynq `archive:extract` 任务。上传/重试只落盘、清理旧文件并立即返回 `extracting`，worker 执行 `ProcessExtractTask` 完成解包、扫描、统计与失败归档；队列任务 12 小时超时且不自动重试，避免同一 18GB 包反复重解。前端补充 `extracting=解包中` 状态显示。已知全量测试仍仅失败于既有 TV APK 版本夹具 144/121。
- 影响文件：`main.go`、`internal/queue/tasks.go`、`internal/queue/archive_extract.go`、`internal/services/archive_import.go`、`admin-web/src/views/ToolboxArchiveImport.vue`、`plan.md`。
- 验证：`go test ./... -count=1` 除既有 `TestParseTVAPKMetadataParsesReleaseAPK` 外通过；`git diff --check` 通过。待推送部署机后真实验证。
## 2026-10-15 09:25 压缩包解压队列与详情状态

- 变更：压缩包解压任务改用独立 `archive` 队列，worker 增加 `ARCHIVE_EXTRACT_WORKERS=1` 专用并发；批次详情在 `extracting` 状态每 5 秒自动刷新并显示后台解包提示。
- 影响文件：`internal/config/config.go`、`internal/queue/archive_extract.go`、`main.go`、`admin-web/src/views/ToolboxArchiveImport.vue`。
- 验证：`go build ./...`、`go test ./internal/config ./internal/queue`、`admin-web npm run build` 均通过。


## 2026-10-16 09:43 +0800
- 进度：管理端视频列表新增「原始标题」列。原始标题没有独立数据库列，只能从 `videos.metadata` JSONB 派生：刮削链路统一写入 `metadata.title_original`（`withLocalizedScrapeMetadata`，刮削确认/同步时落库），上传/Telegram/压缩包链路不写该键，仅有 `metadata.original_filename`。计划在列表 SQL 用 `metadata->>'title_original'` 带出，无该键时回退显示当前标题（而非空），避免把已等同的标题渲染成空列；同时新增 `original_title` 列设置项、窄屏归为 secondary。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoListPage.spec.js`、`CONTEXT.md`、`plan.md`。
- 验证：待执行 `go test ./internal/repository -run 'TestAdminListVideos' -count=1`、`go vet ./internal/repository ./internal/models`、`npm test -- --run src/views/videoListPage.spec.js`、`npm run build`、`git diff --check`、乱码扫描。

## 2026-10-16 09:55 +0800
- 进度：管理端视频列表「原始标题」列完成。后端 `AdminVideoListItem` 新增 `OriginalTitle`，`AdminListVideos` 在既有列表 SQL 内以 `COALESCE(v.metadata->>'title_original','')` 带出，无需新增列或迁移；前端新增 `original_title` 列（默认可见、窄屏归为 secondary、可经列设置隐藏），取值 helper `videoOriginalTitle` 按 `original_title` → `metadata.title_original` → 当前标题回退，上传/Telegram/压缩包等未写原始标题的视频退化为当前标题而非空值。历史用户列设置的 localStorage 版本号从 `2` 提到 `3`，升级时补齐 `duration` 与 `original_title`。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/videoListPage.spec.js`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：`go build ./...`、`go vet ./internal/repository ./internal/models`、`go test ./internal/repository ./internal/models -count=1` 通过；`admin-web` `npx vitest run src/views/videoListPage.spec.js`（19/19）、`npx vitest run src/views/videoList.helpers.spec.js`（30/30）、`npm run build` 通过；`npm test` 全量仅 `src/views/precisionOpsRollout.spec.js` 视图清单用例失败，`git stash` 后复现为既有失败，与本任务无关。待提交。

## 2026-10-16 19:36 +0800
- 进度：按需求移除管理端视频列表「原始标题」列，整体回退提交 `62d4ebe` 的功能改动。7 个代码文件按字节校验已回到该提交父版本，列表不再展示原始标题，`AdminListVideos` 列表 SQL 与 `AdminVideoListItem` 不再带出 `original_title`（避免留下无消费方的死字段），列设置 localStorage 版本号随之回到 `2`（旧用户存储中的 `original_title` 会被 `ALL_COLUMNS` 白名单过滤丢弃，无需额外迁移）。
- 影响文件：`internal/models/admin.go`、`internal/repository/admin_repository.go`、`internal/repository/admin_repository_test.go`、`admin-web/src/views/VideoList.vue`、`admin-web/src/views/videoList.helpers.js`、`admin-web/src/views/videoList.helpers.spec.js`、`admin-web/src/views/videoListPage.spec.js`、`CONTEXT.md`、`plan.md`。未纳入既有未跟踪目录 `docs/examples/`。
- 验证：`go build ./...`、`go vet ./internal/repository ./internal/models`、`go test ./internal/repository ./internal/models -count=1` 通过；`admin-web` `npx vitest run src/views/videoListPage.spec.js src/views/videoList.helpers.spec.js` 通过（19+29=48 用例，helper 用例数由 30 回到 29，与新列移除一致）、`npm run build` 通过；7 个代码文件与 `62d4ebe^` 逐字节 diff 一致；`git diff --check` 通过，改动文件 UTF-8 校验通过、无替换字符。

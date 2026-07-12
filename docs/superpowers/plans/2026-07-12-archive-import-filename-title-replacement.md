# 压缩包视频按文件名替换标题实施计划

> **供代理执行：** 必须使用 `subagent-driven-development`（推荐）或 `executing-plans` 逐任务实施；所有行为变更严格执行测试先行，每个任务完成后独立复核。

**目标：** 在管理端压缩包导入页为 `pending` / `failed` 视频提供单文件和批量“按各自文件名替换标题”能力，并把原标题原子迁移到描述。

**架构：** 管理端用纯函数生成草稿与前 5 条预览；保存时向新增的语义化批量接口提交目标 ID、`updated_at` 与其它字段开关。服务端锁定完整目标集，从数据库中的 `relative_path` 重新派生标题、合并描述并在单个 PostgreSQL 事务中更新；`ProcessFile` 标记 `processing` 后重新读取元数据以封住并发旧快照。

**技术栈：** Go、Gin、pgx v5、PostgreSQL、Vue 3 `<script setup>`、Element Plus、Vitest。

## 全局约束

- 权威规格：`docs/superpowers/specs/2026-07-12-archive-import-filename-title-replacement-design.md`。
- 广告文本按大小写不敏感的完整字面值全局删除：前端 `/www\.98T\.la@/gi`，Go `(?i)www\.98T\.la@`。
- 标题只取 `relative_path` 最后一段，移除最后扩展名，去首尾空白并把连续空白折叠为一个空格；其它符号保留。
- 原标题非空且与新标题精确不同时，描述变为 `原标题 + "\n" + 原描述`；原标题为空或相同时描述不变。
- 替换只允许 `entry_type=file`、`media_kind=video`、状态为 `pending` 或 `failed` 的记录。
- 任一目标无效、过期或派生为空时整批回滚；文件名模式不能同时统一覆盖说明。
- “处理当前文件”“处理所选”不得隐式改名。
- 不新增 migration，不修改 Android 工程，也不递增 Android 版本号。
- Markdown、界面文案和提交信息使用中文并检查乱码。

---

### Task 1: Go 标题派生、描述合并和错误模型

**Files:**
- Create: `internal/services/archive_import_batch_update.go`
- Create: `internal/services/archive_import_batch_update_test.go`

**Interfaces:**
- Produces: `ArchiveImportBatchUpdateTarget`、`ArchiveImportBatchUpdateInput`、`ArchiveImportBatchUpdateIssue`、`ArchiveImportBatchUpdateError`。
- Produces: `deriveArchiveFilenameTitle(relativePath string) string`。
- Produces: `mergeArchiveFilenameTitleDescription(oldTitle, newTitle, description string) string`。
- Produces: `canReplaceArchiveFilenameTitle(file models.ArchiveImportFileListItem) bool`。

- [ ] **Step 1: 写标题派生红灯测试**

在 `internal/services/archive_import_batch_update_test.go` 写表驱动测试：

```go
func TestDeriveArchiveFilenameTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "removes ad text", path: "目录/www.98T.la@ABC-123.mp4", want: "ABC-123"},
		{name: "matches case insensitively and globally", path: "WWW.98t.LA@ A www.98T.la@ B.mkv", want: "A B"},
		{name: "keeps punctuation", path: "目录/www.98T.la@-ABC_[01].mp4", want: "-ABC_[01]"},
		{name: "keeps near match", path: "目录/98T.la@ABC.mp4", want: "98T.la@ABC"},
		{name: "collapses whitespace", path: "目录/www.98T.la@  A  B .mp4", want: "A B"},
		{name: "returns empty after cleaning", path: "目录/www.98T.la@.mp4", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveArchiveFilenameTitle(tt.path); got != tt.want {
				t.Fatalf("deriveArchiveFilenameTitle(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `go test ./internal/services -run TestDeriveArchiveFilenameTitle -count=1`

Expected: FAIL，原因是 `deriveArchiveFilenameTitle` 尚未定义。

- [ ] **Step 3: 实现最小标题派生函数**

在新文件中定义：

```go
var archiveFilenameAdPattern = regexp.MustCompile(`(?i)www\.98T\.la@`)

func deriveArchiveFilenameTitle(relativePath string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(relativePath), `\`, "/")
	if normalized == "" {
		return ""
	}
	base := path.Base(normalized)
	if base == "." || base == "/" {
		return ""
	}
	base = strings.TrimSuffix(base, path.Ext(base))
	base = archiveFilenameAdPattern.ReplaceAllString(base, "")
	return strings.Join(strings.Fields(base), " ")
}
```

- [ ] **Step 4: 运行派生测试并确认转绿**

Run: `go test ./internal/services -run TestDeriveArchiveFilenameTitle -count=1`

Expected: PASS。

- [ ] **Step 5: 写描述合并、资格和错误模型红灯测试**

```go
func TestMergeArchiveFilenameTitleDescription(t *testing.T) {
	t.Parallel()
	tests := []struct{ oldTitle, newTitle, description, want string }{
		{oldTitle: "旧标题", newTitle: "新标题", description: "原说明", want: "旧标题\n原说明"},
		{oldTitle: "旧标题", newTitle: "新标题", description: "", want: "旧标题"},
		{oldTitle: "", newTitle: "新标题", description: "原说明", want: "原说明"},
		{oldTitle: "同名", newTitle: "同名", description: "原说明", want: "原说明"},
		{oldTitle: "Title", newTitle: "title", description: "原说明", want: "Title\n原说明"},
	}
	for _, tt := range tests {
		if got := mergeArchiveFilenameTitleDescription(tt.oldTitle, tt.newTitle, tt.description); got != tt.want {
			t.Fatalf("merge description = %q, want %q", got, tt.want)
		}
	}
}

func TestCanReplaceArchiveFilenameTitle(t *testing.T) {
	t.Parallel()
	eligible := models.ArchiveImportFileListItem{EntryType: "file", MediaKind: "video", Status: "pending"}
	if !canReplaceArchiveFilenameTitle(eligible) { t.Fatal("pending video should be eligible") }
	for _, status := range []string{"processing", "ready", "existing", "skipped"} {
		file := eligible
		file.Status = status
		if canReplaceArchiveFilenameTitle(file) { t.Fatalf("status %s should be ineligible", status) }
	}
}
```

- [ ] **Step 6: 运行新测试并确认按预期失败**

Run: `go test ./internal/services -run 'TestMergeArchiveFilenameTitleDescription|TestCanReplaceArchiveFilenameTitle' -count=1`

Expected: FAIL，原因是两个函数尚未定义。

- [ ] **Step 7: 实现类型、描述规则和稳定错误原因**

```go
const ArchiveImportTitleModeFilename = "filename"

const (
	ArchiveImportBatchReasonInvalidSelection = "invalid_selection"
	ArchiveImportBatchReasonIneligibleTarget = "ineligible_target"
	ArchiveImportBatchReasonStaleTarget = "stale_target"
	ArchiveImportBatchReasonEmptyTitle = "empty_derived_title"
	ArchiveImportBatchReasonInvalidPatch = "invalid_patch"
	ArchiveImportBatchReasonUpdateFailed = "update_failed"
)

type ArchiveImportBatchUpdateTarget struct {
	ID uuid.UUID
	UpdatedAt time.Time
}

type ArchiveImportBatchUpdateInput struct {
	Targets                  []ArchiveImportBatchUpdateTarget
	TitleMode                string
	UpdateTags               bool
	Tags                     []string
	UpdateVideoType          bool
	VideoType                string
	UpdateVideoCollectionIDs bool
	VideoCollectionIDs       []uuid.UUID
	UpdateImageCollectionIDs bool
	ImageCollectionIDs       []uuid.UUID
}

type ArchiveImportBatchUpdateIssue struct {
	ID           uuid.UUID `json:"id"`
	RelativePath string    `json:"relative_path"`
	Message      string    `json:"message"`
}

type ArchiveImportBatchUpdateError struct {
	Reason string                         `json:"reason"`
	Issues []ArchiveImportBatchUpdateIssue `json:"issues"`
	Err    error                          `json:"-"`
}

func (e *ArchiveImportBatchUpdateError) Error() string {
	if e.Err != nil { return e.Err.Error() }
	return e.Reason
}

func (e *ArchiveImportBatchUpdateError) Unwrap() error { return e.Err }

func mergeArchiveFilenameTitleDescription(oldTitle, newTitle, description string) string {
	oldTitle = strings.TrimSpace(oldTitle)
	newTitle = strings.TrimSpace(newTitle)
	description = strings.TrimSpace(description)
	if oldTitle == "" || oldTitle == newTitle { return description }
	if description == "" { return oldTitle }
	return oldTitle + "\n" + description
}

func canReplaceArchiveFilenameTitle(file models.ArchiveImportFileListItem) bool {
	if file.EntryType != "file" || file.MediaKind != "video" { return false }
	return file.Status == "pending" || file.Status == "failed"
}
```

- [ ] **Step 8: 运行 Task 1 全部测试并提交**

Run: `gofmt -w internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go`

Run: `go test ./internal/services -run 'TestDeriveArchiveFilenameTitle|TestMergeArchiveFilenameTitleDescription|TestCanReplaceArchiveFilenameTitle' -count=1`

Expected: PASS。

Commit: `git add internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go && git commit -m "实现压缩包文件名标题派生规则"`

---

### Task 2: 服务端批量事务与处理快照竞态

**Files:**
- Modify: `internal/services/archive_import_batch_update.go`
- Modify: `internal/services/archive_import_batch_update_test.go`
- Modify: `internal/services/archive_import.go:524-555`
- Test: `internal/services/archive_import_test.go`

**Interfaces:**
- Consumes: Task 1 的输入、错误和纯规则函数。
- Produces: `func (s *ArchiveImportService) BatchUpdateFiles(context.Context, ArchiveImportBatchUpdateInput) ([]models.ArchiveImportFileListItem, error)`。

- [ ] **Step 1: 写完整选择集规划红灯测试**

构造同批次文件和目标时间，覆盖成功、重复 ID、跨批次、图片、`processing`、过期时间、空派生标题和超过 200 字符标题。成功断言每个文件得到各自标题与描述，且其它启用字段被应用；失败断言 `ArchiveImportBatchUpdateError.Reason` 和问题文件：

```go
func TestPlanArchiveFilenameBatchUpdateRejectsStaleTarget(t *testing.T) {
	now := time.Date(2026, 7, 12, 5, 0, 0, 0, time.UTC)
	file := models.ArchiveImportFileListItem{
		ID: uuid.New(), BatchID: uuid.New(), EntryType: "file", MediaKind: "video",
		Status: "pending", RelativePath: "www.98T.la@ABC.mp4", Title: "旧标题", UpdatedAt: now,
	}
	_, err := planArchiveFilenameBatchUpdate(
		[]models.ArchiveImportFileListItem{file},
		ArchiveImportBatchUpdateInput{Targets: []ArchiveImportBatchUpdateTarget{{ID: file.ID, UpdatedAt: now.Add(-time.Second)}}, TitleMode: ArchiveImportTitleModeFilename},
	)
	var batchErr *ArchiveImportBatchUpdateError
	if !errors.As(err, &batchErr) || batchErr.Reason != ArchiveImportBatchReasonStaleTarget {
		t.Fatalf("error = %#v, want stale target", err)
	}
}
```

- [ ] **Step 2: 运行规划测试并确认按预期失败**

Run: `go test ./internal/services -run TestPlanArchiveFilenameBatchUpdate -count=1`

Expected: FAIL，原因是 `planArchiveFilenameBatchUpdate` 尚未定义。

- [ ] **Step 3: 实现先全量校验、后生成写入计划的纯函数**

定义内部结构 `archiveImportBatchPlannedFile`，先建立 target map 并收集全部 issue；没有 issue 时才生成变更：

```go
type archiveImportBatchPlannedFile struct {
	File models.ArchiveImportFileListItem
}

func planArchiveFilenameBatchUpdate(files []models.ArchiveImportFileListItem, in ArchiveImportBatchUpdateInput) ([]archiveImportBatchPlannedFile, error) {
	if in.TitleMode != ArchiveImportTitleModeFilename {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: fmt.Errorf("title_mode 仅支持 filename")}
	}
	if len(in.Targets) == 0 {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidSelection, Err: fmt.Errorf("至少选择一个文件")}
	}
	targets := make(map[uuid.UUID]ArchiveImportBatchUpdateTarget, len(in.Targets))
	for _, target := range in.Targets {
		if target.ID == uuid.Nil || target.UpdatedAt.IsZero() {
			return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidSelection, Err: fmt.Errorf("目标 ID 和更新时间不能为空")}
		}
		if _, exists := targets[target.ID]; exists {
			return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidSelection, Err: fmt.Errorf("文件 ID 重复")}
		}
		targets[target.ID] = target
	}
	byID := make(map[uuid.UUID]models.ArchiveImportFileListItem, len(files))
	for _, file := range files { byID[file.ID] = file }
	if len(byID) != len(targets) {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidSelection, Err: fmt.Errorf("部分文件不存在")}
	}
	var batchID uuid.UUID
	issues := make([]ArchiveImportBatchUpdateIssue, 0)
	for _, target := range in.Targets {
		file := byID[target.ID]
		if batchID == uuid.Nil { batchID = file.BatchID }
		if file.BatchID != batchID {
			issues = append(issues, batchUpdateIssueForFile(file, "所选文件不属于同一批次"))
		}
	}
	if len(issues) > 0 { return nil, batchUpdateError(ArchiveImportBatchReasonInvalidSelection, issues) }
	for _, target := range in.Targets {
		file := byID[target.ID]
		if !canReplaceArchiveFilenameTitle(file) {
			issues = append(issues, batchUpdateIssueForFile(file, "仅待处理或失败的视频可以替换标题"))
		}
	}
	if len(issues) > 0 { return nil, batchUpdateError(ArchiveImportBatchReasonIneligibleTarget, issues) }
	for _, target := range in.Targets {
		file := byID[target.ID]
		if !file.UpdatedAt.Equal(target.UpdatedAt) {
			issues = append(issues, batchUpdateIssueForFile(file, "文件信息已变化，请刷新后重试"))
		}
	}
	if len(issues) > 0 { return nil, batchUpdateError(ArchiveImportBatchReasonStaleTarget, issues) }
	for _, target := range in.Targets {
		file := byID[target.ID]
		title := deriveArchiveFilenameTitle(file.RelativePath)
		if title == "" || utf8.RuneCountInString(title) > 200 {
			issues = append(issues, batchUpdateIssueForFile(file, "文件名无法生成有效标题"))
		}
	}
	if len(issues) > 0 { return nil, batchUpdateError(ArchiveImportBatchReasonEmptyTitle, issues) }
	plans := make([]archiveImportBatchPlannedFile, 0, len(in.Targets))
	for _, target := range in.Targets {
		file := byID[target.ID]
		newTitle := deriveArchiveFilenameTitle(file.RelativePath)
		file.Description = mergeArchiveFilenameTitleDescription(file.Title, newTitle, file.Description)
		file.Title = newTitle
		if in.UpdateTags { file.Tags = normalizeArchiveTags(in.Tags) }
		if in.UpdateVideoType { file.VideoType = strings.ToLower(strings.TrimSpace(in.VideoType)) }
		if in.UpdateVideoCollectionIDs { file.VideoCollectionIDs = dedupeArchiveUUIDs(in.VideoCollectionIDs) }
		if in.UpdateImageCollectionIDs { file.ImageCollectionIDs = dedupeArchiveUUIDs(in.ImageCollectionIDs) }
		plans = append(plans, archiveImportBatchPlannedFile{File: file})
	}
	return plans, nil
}
```

同时实现 `batchUpdateIssueForFile(file, message)` 与 `batchUpdateError(reason, issues)`，用文件 ID、相对路径和中文原因构造完整问题清单。

实现时必须在第一次文件字段赋值前完成所有目标校验，保证任何错误都不会留下部分内存计划。

- [ ] **Step 4: 运行规划测试并确认转绿**

Run: `go test ./internal/services -run TestPlanArchiveFilenameBatchUpdate -count=1`

Expected: PASS。

- [ ] **Step 5: 写事务应用失败红灯测试**

用嵌入 `pgx.Tx` 的测试 fake 覆盖 `Exec`，在第二次写入返回错误；断言 helper 返回错误且不会执行第三次写入：

```go
type failingArchiveImportTx struct {
	pgx.Tx
	calls int
	failAt int
}

func (tx *failingArchiveImportTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	tx.calls++
	if tx.calls == tx.failAt { return pgconn.CommandTag{}, errors.New("forced update failure") }
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
```

- [ ] **Step 6: 实现锁定、读取默认值、事务写入和提交**

新增 `applyArchiveImportBatchPlanTx(ctx, tx, plans, batch, groups, in)`；事务失败测试直接调用它。`BatchUpdateFiles` 按以下固定顺序实现：

```go
func (s *ArchiveImportService) BatchUpdateFiles(ctx context.Context, in ArchiveImportBatchUpdateInput) ([]models.ArchiveImportFileListItem, error) {
	if in.TitleMode != ArchiveImportTitleModeFilename {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: fmt.Errorf("title_mode 仅支持 filename")}
	}
	if in.UpdateVideoType && !isArchiveImportVideoType(in.VideoType) {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: fmt.Errorf("视频类型无效")}
	}
	var err error
	if in.UpdateVideoCollectionIDs {
		in.VideoCollectionIDs, err = s.resolveArchiveImportVideoCollectionIDs(ctx, in.VideoCollectionIDs)
		if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: err} }
	}
	if in.UpdateImageCollectionIDs {
		in.ImageCollectionIDs, err = s.resolveArchiveImportVideoImageCollectionIDs(ctx, in.ImageCollectionIDs)
		if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonInvalidPatch, Err: err} }
	}
	ids := sortedArchiveImportBatchTargetIDs(in.Targets)
	tx, err := s.db.Begin(ctx)
	if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err} }
	defer tx.Rollback(ctx)
	files, err := lockArchiveImportFilesTx(ctx, tx, ids)
	if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err} }
	plans, err := planArchiveFilenameBatchUpdate(files, in)
	if err != nil { return nil, err }
	batch, err := getArchiveImportBatchDefaultsTx(ctx, tx, plans[0].File.BatchID)
	if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err} }
	groups, err := getArchiveImportGroupsTx(ctx, tx, plans)
	if err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err} }
	if err := applyArchiveImportBatchPlanTx(ctx, tx, plans, batch, groups, in); err != nil {
		return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err}
	}
	if err := tx.Commit(ctx); err != nil { return nil, &ArchiveImportBatchUpdateError{Reason: ArchiveImportBatchReasonUpdateFailed, Err: err} }
	return s.listArchiveFilesByIDs(ctx, plans[0].File.BatchID, ids)
}
```

返回列表沿用 `listArchiveFilesByIDs` 的相对路径顺序，前端成功后会刷新批次详情，不依赖响应顺序。`applyArchiveImportBatchPlanTx` 对每个计划调用 `resolveArchiveImportFileDefaults`，只重算本次显式修改字段的 override，然后调用 `updateArchiveImportFileStateTx`。新增 `sortedArchiveImportBatchTargetIDs`、`lockArchiveImportFilesTx`、`getArchiveImportBatchDefaultsTx` 和 `getArchiveImportGroupsTx` 小函数；查询必须复用 `scanArchiveImportFileRecord` / `scanArchiveImportGroupRecord`，不得用第二条 pool 连接绕开当前事务。

- [ ] **Step 7: 写并通过 `ProcessFile` 重读顺序回归测试**

在 `archive_import_test.go` 增加源码顺序测试：截取 `func (s *ArchiveImportService) ProcessFile`，断言 `markArchiveFileProcessing` 之后、`archiveFileTitleForProcessing` 之前再次出现 `s.getArchiveFile(ctx, fileID)`。先运行并观察失败，再在 `markArchiveFileProcessing` 成功后加入：

```go
file, err = s.getArchiveFile(ctx, fileID)
if err != nil {
	return models.ArchiveImportFileListItem{}, err
}
```

Run: `go test ./internal/services -run 'TestPlanArchiveFilenameBatchUpdate|TestApplyArchiveImportBatchPlan|TestProcessFileReloadsMetadataAfterMarkingProcessing' -count=1`

Expected: PASS。

- [ ] **Step 8: 运行服务层回归并提交**

Run: `gofmt -w internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go internal/services/archive_import.go internal/services/archive_import_test.go`

Run: `go test ./internal/services -run 'ArchiveImport|ProcessFileReloadsMetadata' -count=1`

Expected: PASS。

Commit: `git add internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go internal/services/archive_import.go internal/services/archive_import_test.go && git commit -m "实现压缩包标题替换事务"`

---

### Task 3: Gin 接口、路由和结构化错误

**Files:**
- Modify: `internal/handlers/router.go:38-58,300-312`
- Modify: `internal/handlers/admin_archive_import.go`
- Modify: `internal/handlers/admin_archive_import_test.go`

**Interfaces:**
- Consumes: `ArchiveImportService.BatchUpdateFiles`。
- Produces: `PUT /api/v1/admin/archive-import/files/batch-update`。

- [ ] **Step 1: 扩展 stub 并写 Handler 红灯测试**

给 `archiveImportServiceStub` 增加 `batchUpdateInput`、`batchUpdateFiles`、`batchUpdateErr` 和调用次数，实现新接口方法。测试有效请求把目标时间、模式和字段开关传给 service：

```go
func TestAdminBatchUpdateArchiveImportFilesPassesSemanticInput(t *testing.T) {
	payload := `{"targets":[{"id":"11111111-1111-4111-8111-111111111111","updated_at":"2026-07-12T05:00:00Z"}],"title_mode":"filename","update_tags":true,"tags":["标签"]}`
	stub := &archiveImportServiceStub{batchUpdateFiles: []models.ArchiveImportFileListItem{{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111")}}}
	api := &API{archiveImportSvc: stub}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/archive-import/files/batch-update", bytes.NewBufferString(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	api.AdminBatchUpdateArchiveImportFiles(ctx)
	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil { t.Fatal(err) }
	if resp.Code != 0 || stub.batchUpdateCalls != 1 { t.Fatalf("body=%s calls=%d", rec.Body.String(), stub.batchUpdateCalls) }
	if stub.batchUpdateInput.TitleMode != services.ArchiveImportTitleModeFilename { t.Fatalf("title mode=%q", stub.batchUpdateInput.TitleMode) }
	if len(stub.batchUpdateInput.Targets) != 1 || len(stub.batchUpdateInput.Tags) != 1 { t.Fatalf("input=%#v", stub.batchUpdateInput) }
}
```

再写两个测试：非法 UUID 返回 `code=1` 且 service 未调用；service 返回 `ArchiveImportBatchUpdateError{Reason: stale_target}` 时响应 `code=1078`，`data.reason` 和 `data.issues` 保留。

- [ ] **Step 2: 运行 Handler 测试并确认按预期失败**

Run: `go test ./internal/handlers -run TestAdminBatchUpdateArchiveImportFiles -count=1`

Expected: FAIL，原因是 Handler 和 interface 方法尚未定义。

- [ ] **Step 3: 实现请求解析和错误映射**

请求结构使用字符串 ID 与 `time.Time`：

```go
type adminArchiveImportBatchTargetRequest struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type adminArchiveImportBatchUpdateRequest struct {
	Targets                  []adminArchiveImportBatchTargetRequest `json:"targets"`
	TitleMode                string                                  `json:"title_mode"`
	UpdateTags               bool                                    `json:"update_tags"`
	Tags                     []string                                `json:"tags"`
	UpdateVideoType          bool                                    `json:"update_video_type"`
	VideoType                string                                  `json:"video_type"`
	UpdateVideoCollectionIDs bool                                    `json:"update_video_collection_ids"`
	VideoCollectionIDs       []string                                `json:"video_collection_ids"`
	UpdateImageCollectionIDs bool                                    `json:"update_image_collection_ids"`
	ImageCollectionIDs       []string                                `json:"image_collection_ids"`
}
```

Handler 校验目标非空、UUID、非零时间和合集 UUID；成功返回 `updated_count` 与 `items`。业务错误映射：

```go
var batchErr *services.ArchiveImportBatchUpdateError
if errors.As(err, &batchErr) {
	response.JSON(c, 1078, "压缩包文件批量更新失败", gin.H{
		"reason": batchErr.Reason,
		"issues": batchErr.Issues,
	})
	return
}
```

- [ ] **Step 4: 注册接口并更新 service interface**

在 `archiveImportService` 添加 `BatchUpdateFiles`，在单文件 `PUT` 路由之前注册：

```go
admin.PUT("/archive-import/files/batch-update", a.AdminBatchUpdateArchiveImportFiles)
```

- [ ] **Step 5: 运行 Handler 和路由相关测试并提交**

Run: `gofmt -w internal/handlers/router.go internal/handlers/admin_archive_import.go internal/handlers/admin_archive_import_test.go`

Run: `go test ./internal/handlers -run 'TestAdminBatchUpdateArchiveImportFiles|TestAdmin.*ArchiveImport' -count=1`

Expected: PASS。

Commit: `git add internal/handlers/router.go internal/handlers/admin_archive_import.go internal/handlers/admin_archive_import_test.go && git commit -m "接入压缩包标题批量更新接口"`

---

### Task 4: 管理端纯函数和 API 客户端

**Files:**
- Create: `admin-web/src/views/toolboxArchiveImport.helpers.js`
- Create: `admin-web/src/views/toolboxArchiveImport.helpers.spec.js`
- Modify: `admin-web/src/api/admin.js:141-180`
- Modify: `admin-web/src/api/admin.spec.js`

**Interfaces:**
- Produces: `deriveArchiveFilenameTitle`、`canReplaceArchiveFilenameTitle`、`buildArchiveFilenameTitleDraft`、`buildArchiveFilenameBatchPreview`、`buildArchiveFilenameBatchTargets`、`buildArchiveFilenameBatchPayload`。
- Produces: `batchUpdateAdminArchiveImportFiles(payload)`。

- [ ] **Step 1: 写前端纯函数红灯测试**

测试数据与 Task 1 保持逐项一致，并增加草稿和前 5 条预览：

```js
it('derives cleaned titles and preserves old titles in descriptions', () => {
  expect(buildArchiveFilenameTitleDraft({
    relative_path: '目录/WWW.98t.LA@  ABC  123.mp4',
    title: '旧标题',
    description: '原说明',
    entry_type: 'file',
    media_kind: 'video',
    status: 'pending'
  })).toMatchObject({ ok: true, title: 'ABC 123', description: '旧标题\n原说明' })
})

it('blocks the whole preview when one target is empty', () => {
  const preview = buildArchiveFilenameBatchPreview([
    pendingVideo('www.98T.la@ABC.mp4'),
    pendingVideo('www.98T.la@.mp4')
  ], 5)
  expect(preview.ok).toBe(false)
  expect(preview.issues).toHaveLength(1)
  expect(preview.items).toHaveLength(1)
})
```

- [ ] **Step 2: 运行 helper 测试并确认按预期失败**

Run: `cd admin-web && npm test -- toolboxArchiveImport.helpers.spec.js`

Expected: FAIL，原因是 helper 模块尚不存在。

- [ ] **Step 3: 实现最小纯函数模块**

```js
const archiveFilenameAdPattern = /www\.98T\.la@/gi

export function deriveArchiveFilenameTitle(relativePath) {
  const normalized = String(relativePath || '').trim().replace(/\\/g, '/')
  const filename = normalized.split('/').pop() || ''
  const dotIndex = filename.lastIndexOf('.')
  const basename = dotIndex >= 0 ? filename.slice(0, dotIndex) : filename
  return basename.replace(archiveFilenameAdPattern, '').trim().replace(/\s+/g, ' ')
}

export function canReplaceArchiveFilenameTitle(file) {
  return file?.entry_type === 'file'
    && file?.media_kind === 'video'
    && (file?.status === 'pending' || file?.status === 'failed')
}

export function buildArchiveFilenameTitleDraft(file) {
  const originalTitle = String(file?.title || '').trim()
  const originalDescription = String(file?.description || '').trim()
  if (!canReplaceArchiveFilenameTitle(file)) {
    return { ok: false, title: originalTitle, description: originalDescription, changed: false, issue: { id: String(file?.id || ''), relative_path: String(file?.relative_path || ''), message: '仅待处理或失败的视频可以替换标题' } }
  }
  const title = deriveArchiveFilenameTitle(file?.relative_path)
  if (!title || Array.from(title).length > 200) {
    return { ok: false, title: originalTitle, description: originalDescription, changed: false, issue: { id: String(file?.id || ''), relative_path: String(file?.relative_path || ''), message: '文件名无法生成有效标题' } }
  }
  const description = originalTitle && originalTitle !== title
    ? [originalTitle, originalDescription].filter(Boolean).join('\n')
    : originalDescription
  return { ok: true, title, description, changed: title !== originalTitle || description !== originalDescription, issue: null }
}

export function buildArchiveFilenameBatchPreview(files, limit = 5) {
  const drafts = (Array.isArray(files) ? files : []).map((file) => ({ file, draft: buildArchiveFilenameTitleDraft(file) }))
  const issues = drafts.filter(({ draft }) => !draft.ok).map(({ draft }) => draft.issue)
  const validItems = drafts.filter(({ draft }) => draft.ok).map(({ file, draft }) => ({
    id: String(file.id || ''),
    relative_path: String(file.relative_path || ''),
    old_title: String(file.title || '').trim(),
    new_title: draft.title
  }))
  return {
    ok: issues.length === 0 && drafts.length > 0,
    total: drafts.length,
    items: validItems.slice(0, limit),
    remaining: Math.max(validItems.length - limit, 0),
    issues
  }
}

export function buildArchiveFilenameBatchTargets(files) {
  return files.map((file) => ({ id: String(file.id || ''), updated_at: String(file.updated_at || '') }))
}

export function buildArchiveFilenameBatchPayload(files, patch = {}) {
  return { targets: buildArchiveFilenameBatchTargets(files), title_mode: 'filename', ...patch }
}
```

- [ ] **Step 4: 写 API 红灯测试并实现请求函数**

在 `admin.spec.js` 导入并断言：

```js
const payload = { targets: [{ id: 'file-1', updated_at: '2026-07-12T05:00:00Z' }], title_mode: 'filename' }
await batchUpdateAdminArchiveImportFiles(payload)
expect(put).toHaveBeenCalledWith('/admin/archive-import/files/batch-update', payload)
```

实现：

```js
export const batchUpdateAdminArchiveImportFiles = (payload) =>
  request.put('/admin/archive-import/files/batch-update', payload)
```

- [ ] **Step 5: 运行定向测试并提交**

Run: `cd admin-web && npm test -- toolboxArchiveImport.helpers.spec.js admin.spec.js`

Expected: PASS。

Commit: `git add admin-web/src/views/toolboxArchiveImport.helpers.js admin-web/src/views/toolboxArchiveImport.helpers.spec.js admin-web/src/api/admin.js admin-web/src/api/admin.spec.js && git commit -m "增加压缩包标题替换前端规则"`

---

### Task 5: Vue 单文件草稿与批量三态交互

**Files:**
- Modify: `admin-web/src/views/ToolboxArchiveImport.vue`
- Modify: `admin-web/src/views/ToolboxArchiveImport.spec.js`
- Test: `admin-web/src/views/toolboxArchiveImport.helpers.spec.js`

**Interfaces:**
- Consumes: Task 4 的 helper 与 API。
- Preserves: 未选择文件名模式时继续调用既有 `updateAdminArchiveImportFile` 循环。

- [ ] **Step 1: 写页面结构和行为红灯测试**

在源码规格测试中断言：

```js
it('adds explicit filename title modes without changing process actions', () => {
  expect(source).toContain("{ label: '不修改', value: 'none' }")
  expect(source).toContain("{ label: '统一标题', value: 'uniform' }")
  expect(source).toContain("{ label: '按各自文件名替换', value: 'filename' }")
  expect(source).toContain('applySelectedFilenameTitleDraft')
  expect(source).toContain('saveArchiveFilenameBatchUpdate')
  expect(source).toContain('使用文件名')
  expect(source).toContain('原标题')
  expect(source).not.toContain('processSelectedArchiveFiles()\n  applySelectedFilenameTitleDraft')
})
```

补 helper 测试锁定前 5 条预览与剩余数量；页面源码测试锁定标题/描述输入的 `@input` 会退出文件名模式。

- [ ] **Step 2: 运行页面测试并确认按预期失败**

Run: `cd admin-web && npm test -- ToolboxArchiveImport.spec.js toolboxArchiveImport.helpers.spec.js`

Expected: FAIL，原因是模式、动作和保存函数尚不存在。

- [ ] **Step 3: 接入单文件草稿状态**

新增 `selectedFilenameDraftSnapshot`，在打开/切换/关闭文件时清空。按钮动作调用 helper 后写入 `selectedFile.title/description` 并保存生成快照；标题或说明输入的 `@input` 调用 `deactivateSelectedFilenameMode`。保存时：

```js
if (selectedFilenameDraftSnapshot.value) {
  await batchUpdateAdminArchiveImportFiles(buildArchiveFilenameBatchPayload([selectedFile.value], {
    update_tags: true,
    tags: normalizeTagSelection(selectedFile.value.tags),
    update_video_type: true,
    video_type: selectedFile.value.video_type || 'short',
    update_video_collection_ids: true,
    video_collection_ids: normalizeUUIDSelection(selectedFile.value.video_collection_ids),
    update_image_collection_ids: true,
    image_collection_ids: normalizeUUIDSelection([selectedVideoImageCollectionID.value])
  }))
} else {
  await updateAdminArchiveImportFile(selectedFile.value.id, payload)
}
```

单文件按钮使用 Element Plus 图标加文字，且只在 `canReplaceArchiveFilenameTitle(selectedFile)` 时出现。

- [ ] **Step 4: 把批量标题改成三态模式**

`batchEditForm` 使用 `title_mode: 'none'`。`uniform` 显示现有标题输入；`filename` 自动关闭并禁用 `description_enabled`，显示 `batchFilenamePreview.items` 前 5 条和 `remaining`。重置、序列化、脏数据检测都包含 `title_mode`。

```js
const batchTitleModeOptions = [
  { label: '不修改', value: 'none' },
  { label: '统一标题', value: 'uniform' },
  { label: '按各自文件名替换', value: 'filename' }
]

const batchFilenamePreview = computed(() =>
  batchEditForm.title_mode === 'filename'
    ? buildArchiveFilenameBatchPreview(selectedBatchFilesForActions.value, 5)
    : { ok: false, total: 0, items: [], remaining: 0, issues: [] }
)

function onBatchTitleModeChange(mode) {
  if (mode !== 'filename') return
  batchEditForm.description_enabled = false
  batchEditForm.description = ''
}
```

- [ ] **Step 5: 实现文件名模式的单次事务保存**

新增 `saveArchiveFilenameBatchUpdate`：

- 本地 preview 不通过时显示问题路径且不发请求。
- 请求只包含 `targets`、`title_mode=filename` 和已勾选的其它字段。
- `error.data.reason` 为 `stale_target` / `ineligible_target` 时刷新批次详情并保留批量表单字段。
- 其它错误保留弹窗和草稿。
- 成功刷新详情、关闭弹窗并提示 `已更新 N 个视频`。

`saveBatchEdit` 在 `title_mode === 'filename'` 时只调用该函数；其它模式继续沿用逐文件逻辑，其中统一标题判定改为 `title_mode === 'uniform'`。

- [ ] **Step 6: 增加紧凑预览样式并运行测试**

预览使用普通列表，不嵌套卡片；每行固定两列并允许长标题换行，移动端退化为单列。新增类名：`archive-title-mode`、`archive-title-preview`、`archive-title-preview__row`、`archive-title-preview__issues`。

Run: `cd admin-web && npm test -- ToolboxArchiveImport.spec.js toolboxArchiveImport.helpers.spec.js admin.spec.js`

Expected: PASS。

- [ ] **Step 7: 运行管理端全量测试和构建并提交**

Run: `cd admin-web && npm test`

Run: `cd admin-web && npm run build`

Expected: 两条命令均退出 0；允许报告仓库既有 Vite chunk size warning，但不得新增编译错误。

Commit: `git add admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js admin-web/src/views/toolboxArchiveImport.helpers.js admin-web/src/views/toolboxArchiveImport.helpers.spec.js admin-web/src/api/admin.js admin-web/src/api/admin.spec.js && git commit -m "实现压缩包文件名标题替换交互"`

---

### Task 6: 集成验证、文档账本和独立复审

**Files:**
- Modify: `plan.md`
- Modify only if implementation reveals a durable correction: `CONTEXT.md`
- Review: all files changed since implementation-plan commit

**Interfaces:**
- Verifies: design spec sections 2 through 10 and every global constraint above。

- [ ] **Step 1: 运行前后端定向与全量验证**

Run:

```bash
go test ./internal/services -run 'ArchiveImport|ProcessFileReloadsMetadata' -count=1
go test ./internal/handlers -run 'Admin.*ArchiveImport' -count=1
go test ./internal/services ./internal/handlers -count=1
go test ./... -count=1
cd admin-web && npm test -- ToolboxArchiveImport.spec.js toolboxArchiveImport.helpers.spec.js admin.spec.js
cd admin-web && npm test
cd admin-web && npm run build
git diff --check
```

Expected: 全部退出 0；若 `go test ./...` 命中已知无关 fixture 失败，记录完整测试名和定向测试通过证据，不得把全量失败写成通过。

- [ ] **Step 2: 扫描乱码、接口残留和范围**

Run:

```bash
rg -n $'\uFFFD' internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go internal/handlers/admin_archive_import.go internal/handlers/admin_archive_import_test.go admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/toolboxArchiveImport.helpers.js admin-web/src/views/toolboxArchiveImport.helpers.spec.js admin-web/src/views/ToolboxArchiveImport.spec.js admin-web/src/api/admin.js admin-web/src/api/admin.spec.js CONTEXT.md plan.md
rg -n "www\\.98T\\.la@|title_mode|batch-update" internal admin-web/src
git status --short
git diff --stat
```

Expected: 乱码扫描无输出；规则和接口只出现在预期文件；无无关改动。

- [ ] **Step 3: 让独立子代理做规格与代码双轴评审**

评审输入固定为设计提交 `2135ae8` 到当前 HEAD。要求分别检查：

- 规格轴：清洗、描述迁移、状态门禁、单/批量草稿、说明互斥、原子性、竞态和反馈是否全部满足。
- 工程轴：事务锁顺序、`updated_at` 比较、`field_overrides`、错误数据、Vue 脏状态、并发重试和测试是否存在阻塞问题。

发现 blocker/high/medium 后先写失败测试，再修复并重新评审；直到没有阻塞问题。

- [ ] **Step 4: 追加 `plan.md` 最终记录并提交**

记录实际改动、红绿测试证据、全量验证结果、评审结论和未执行的人工验收。精确暂存本任务文件后提交：

```bash
git add -- internal/services/archive_import_batch_update.go internal/services/archive_import_batch_update_test.go internal/services/archive_import.go internal/services/archive_import_test.go internal/handlers/router.go internal/handlers/admin_archive_import.go internal/handlers/admin_archive_import_test.go admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js admin-web/src/views/toolboxArchiveImport.helpers.js admin-web/src/views/toolboxArchiveImport.helpers.spec.js admin-web/src/api/admin.js admin-web/src/api/admin.spec.js CONTEXT.md plan.md
git diff --cached --check
git commit -m "实现压缩包文件名标题替换"
```

- [ ] **Step 5: 启动管理端开发服务器供用户验收**

若默认端口已占用，选择下一个空闲端口：

Run: `cd admin-web && npm run dev -- --host 127.0.0.1`

向用户提供实际 URL，并明确仍需人工验证设计规格第 10 节的 10 个场景；在用户确认前不创建 `DONE.md`，因为本任务不属于现有 `tasks/` 三段目录。

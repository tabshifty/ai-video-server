package services

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestArchiveWorkspaceInheritance(t *testing.T) {
	id := uuid.New()
	groupID := uuid.NewString()
	draft := newArchiveWorkspace(models.ArchiveImportBatch{Title: "批次", DefaultTitlePrefix: "上传标题", DefaultTags: []string{"父标签"}}, []models.ArchiveImportFileListItem{{ID: id, RelativePath: "目录/片段.mp4", EntryType: "file", MediaKind: "video", Status: "pending"}}, nil, true)
	draft.Groups = append(draft.Groups, ArchiveWorkspaceGroup{ID: groupID, ParentID: archiveWorkspaceMain, Name: "子组"})
	f := draft.Files[id.String()]
	f.GroupID = groupID
	draft.Files[id.String()] = f
	got, err := resolveWorkspaceFile(draft, id.String(), "video")
	if err != nil || got.Title != "上传标题" || len(got.Tags) != 1 {
		t.Fatalf("继承失败: %+v %v", got, err)
	}
	empty := []string{}
	draft.Groups[len(draft.Groups)-1].Config.Tags = &empty
	got, err = resolveWorkspaceFile(draft, id.String(), "video")
	if err != nil || len(got.Tags) != 0 {
		t.Fatalf("显式清空失败: %+v %v", got, err)
	}
	mode := "filename"
	draft.Groups[len(draft.Groups)-1].Config.TitleMode = &mode
	f = draft.Files[id.String()]
	f.Name = "www.98T.la@新名称.mp4"
	draft.Files[id.String()] = f
	got, err = resolveWorkspaceFile(draft, id.String(), "video")
	if err != nil || got.Title != "新名称" {
		t.Fatalf("虚拟名称派生失败: %+v %v", got, err)
	}
}

func TestArchiveWorkspaceTreeAndFrozenProtection(t *testing.T) {
	id := uuid.New()
	files := []models.ArchiveImportFileListItem{{ID: id, RelativePath: "a.mp4", EntryType: "file", MediaKind: "video", Status: "ready"}}
	base := newArchiveWorkspace(models.ArchiveImportBatch{Title: "标题"}, files, nil, true)
	clone := func() ArchiveWorkspaceDraft {
		d := base
		d.Files = map[string]ArchiveWorkspaceFile{}
		for k, v := range base.Files {
			d.Files[k] = v
		}
		d.Groups = append([]ArchiveWorkspaceGroup{}, base.Groups...)
		return d
	}
	t.Run("禁止移动成功文件", func(t *testing.T) {
		d := clone()
		f := d.Files[id.String()]
		f.GroupID = ""
		d.Files[id.String()] = f
		if validateWorkspace(d, base, files, nil) == nil {
			t.Fatal("应拒绝")
		}
	})
	t.Run("禁止删除主组", func(t *testing.T) {
		d := clone()
		d.Groups = nil
		if validateWorkspace(d, base, files, nil) == nil {
			t.Fatal("应拒绝")
		}
	})
	t.Run("拒绝分组循环", func(t *testing.T) {
		d := clone()
		a, b := uuid.NewString(), uuid.NewString()
		d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: a, ParentID: b, Name: "甲"}, ArchiveWorkspaceGroup{ID: b, ParentID: a, Name: "乙"})
		if validateWorkspace(d, base, files, nil) == nil {
			t.Fatal("应拒绝")
		}
	})
	t.Run("活动成员不受父配置更改影响", func(t *testing.T) {
		d := clone()
		title := "新默认值"
		d.Groups[0].Config.Title = &title
		if err := validateWorkspace(d, base, files, map[string]bool{id.String(): true}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestArchiveWorkspaceInitialization(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	oldGroup := uuid.New()
	files := []models.ArchiveImportFileListItem{{ID: a, RelativePath: "根.mp4", EntryType: "file", MediaKind: "video", Status: "pending", Title: "旧文件标题", GroupID: &oldGroup}, {ID: b, RelativePath: "父/子/图.jpg", EntryType: "file", MediaKind: "image", Status: "pending", Title: "图片"}}
	batch := models.ArchiveImportBatch{Title: "批次", DefaultTitlePrefix: "上传标题"}
	fresh := newArchiveWorkspace(batch, files, nil, true)
	if len(fresh.Groups) != 3 || fresh.Files[a.String()].GroupID != archiveWorkspaceMain {
		t.Fatalf("目录层级或根归属错误: %+v", fresh)
	}
	got, err := resolveWorkspaceFile(fresh, a.String(), "video")
	if err != nil || got.Title != "上传标题" {
		t.Fatalf("不应以目录覆盖标题: %+v %v", got, err)
	}
	old := newArchiveWorkspace(batch, files, []models.ArchiveImportGroup{{ID: oldGroup, Name: "旧分组"}}, false)
	got, err = resolveWorkspaceFile(old, a.String(), "video")
	if err != nil || got.Title != "旧文件标题" || old.Files[a.String()].GroupID != oldGroup.String() {
		t.Fatalf("旧配置丢失: %+v %v", got, err)
	}
}

func TestArchiveWorkspaceOldNestedGroupAndImageTitle(t *testing.T) {
	id, groupID := uuid.New(), uuid.New()
	title := "原图组标题"
	files := []models.ArchiveImportFileListItem{{ID: id, GroupID: &groupID, RelativePath: "父/子/图.jpg", MediaKind: "image", EntryType: "file", Status: "pending", Title: title, FieldOverrides: map[string]bool{}}}
	d := newArchiveWorkspace(models.ArchiveImportBatch{Title: "批次"}, files, []models.ArchiveImportGroup{{ID: groupID, Name: "父/子", Title: &title}}, false)
	if err := validateWorkspace(d, d, files, nil); err != nil {
		t.Fatalf("旧路径分组名应可保存: %v", err)
	}
	got, err := resolveWorkspaceFile(d, id.String(), "image")
	if err != nil || got.Title != title {
		t.Fatalf("旧图片标题未保留: %+v %v", got, err)
	}
}

func TestArchiveWorkspaceActiveAncestorCannotMove(t *testing.T) {
	id := uuid.New()
	files := []models.ArchiveImportFileListItem{{ID: id, RelativePath: "父/子/片.mp4", MediaKind: "video", EntryType: "file", Status: "pending"}}
	old := newArchiveWorkspace(models.ArchiveImportBatch{Title: "批次"}, files, nil, true)
	raw, _ := json.Marshal(old)
	var d ArchiveWorkspaceDraft
	_ = json.Unmarshal(raw, &d)
	d.Directories[0].Name = "改名"
	if err := validateWorkspace(d, old, files, map[string]bool{id.String(): true}); err == nil {
		t.Fatal("活动文件祖先不能改名")
	}
}

func TestArchiveWorkspaceFileTitleOverridesFilenameMode(t *testing.T) {
	id := uuid.New()
	files := []models.ArchiveImportFileListItem{{ID: id, RelativePath: "片.mp4", MediaKind: "video", EntryType: "file", Status: "pending"}}
	d := newArchiveWorkspace(models.ArchiveImportBatch{Title: "批次"}, files, nil, true)
	mode, title := "filename", "单独标题"
	d.Groups[0].Config.TitleMode = &mode
	f := d.Files[id.String()]
	f.Config.Title = &title
	d.Files[id.String()] = f
	got, err := resolveWorkspaceFile(d, id.String(), "video")
	if err != nil || got.Title != title {
		t.Fatalf("文件标题优先级错误: %+v %v", got, err)
	}
}

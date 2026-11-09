package services

import (
	"encoding/json"
	"github.com/google/uuid"
	"testing"
	"video-server/internal/models"
)

func cloneDesktopDraft(t *testing.T, d ArchiveWorkspaceDraft) ArchiveWorkspaceDraft {
	t.Helper()
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var out ArchiveWorkspaceDraft
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestWorkspaceDesktopSingleMembership(t *testing.T) {
	id := uuid.New()
	files := []models.ArchiveImportFileListItem{{ID: id, RelativePath: "视频.mp4", EntryType: "file", MediaKind: "video", Status: "ready"}}
	old := newArchiveWorkspace(models.ArchiveImportBatch{Title: "上传"}, files, nil, true)
	for _, tc := range []struct {
		name      string
		active    bool
		wantError bool
	}{{"成功文件可以移出及删组", false, false}, {"执行中文件不能移出", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			d := cloneDesktopDraft(t, old)
			f := d.Files[id.String()]
			f.GroupID = ""
			d.Files[id.String()] = f
			d.Groups = nil
			active := map[string]bool{id.String(): tc.active}
			err := validateWorkspace(d, old, files, active)
			if (err != nil) != tc.wantError {
				t.Fatalf("验证结果: %v", err)
			}
		})
	}
}

func TestWorkspaceDesktopInitialization(t *testing.T) {
	files := []models.ArchiveImportFileListItem{{ID: uuid.New(), RelativePath: "甲/子/片.mp4", EntryType: "file", MediaKind: "video"}, {ID: uuid.New(), RelativePath: "乙/子/图.jpg", EntryType: "file", MediaKind: "image"}}
	d := newArchiveWorkspace(models.ArchiveImportBatch{Title: "上传"}, files, nil, true)
	if len(d.Groups) != 2 {
		t.Fatalf("只为实际含媒体目录建平级组: %+v", d.Groups)
	}
	for _, g := range d.Groups {
		if g.ParentID != "" {
			t.Fatal("不能保留父子继承")
		}
	}
	if d.Groups[0].Name == d.Groups[1].Name {
		t.Fatal("重名目录需要路径区分")
	}
}

func TestWorkspaceDesktopLegacyFlatten(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	vc1, vc2, ic := uuid.New(), uuid.New(), uuid.New()
	title, mode := "父标题", "uniform"
	vcs := []uuid.UUID{vc1}
	own := []uuid.UUID{vc2}
	ics := []uuid.UUID{ic}
	old := ArchiveWorkspaceDraft{Groups: []ArchiveWorkspaceGroup{{ID: "main", Name: "主组", Config: ArchiveWorkspaceConfig{Title: &title, TitleMode: &mode, VideoCollectionIDs: &vcs, ImageCollectionIDs: &ics}}, {ID: "g", ParentID: "main", Name: "子组"}}, Files: map[string]ArchiveWorkspaceFile{a.String(): {Name: "甲.mp4", GroupID: "g", Config: ArchiveWorkspaceConfig{VideoCollectionIDs: &own}}, b.String(): {Name: "乙.mp4", GroupID: "g"}}}
	files := []models.ArchiveImportFileListItem{{ID: a, EntryType: "file", MediaKind: "video"}, {ID: b, EntryType: "file", MediaKind: "video"}}
	d := flattenWorkspace(old, files)
	for _, source := range files {
		before, e := resolveWorkspaceFile(old, source.ID.String(), source.MediaKind)
		if e != nil {
			t.Fatal(e)
		}
		after, e := resolveWorkspaceFile(d, source.ID.String(), source.MediaKind)
		if e != nil || before.Title != after.Title || !workspaceSameIDs(before.VideoCollectionIDs, after.VideoCollectionIDs) || !workspaceSameIDs(before.ImageCollectionIDs, after.ImageCollectionIDs) {
			t.Fatalf("转换丢失配置: %+v %+v %v", before, after, e)
		}
	}
	if d.Files[a.String()].GroupID == d.Files[b.String()].GroupID {
		t.Fatal("文件目标差异未拆组")
	}
	for _, g := range d.Groups {
		if g.ParentID != "" {
			t.Fatal("旧继承未固化")
		}
	}
	if err := validateWorkspace(d, d, files, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceDesktopAppendTargets(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name       string
		have, want ArchiveWorkspaceValues
		applied    bool
	}{
		{"旧任务必须补配对", ArchiveWorkspaceValues{VideoCollectionIDs: []uuid.UUID{a}}, ArchiveWorkspaceValues{VideoCollectionIDs: []uuid.UUID{a}}, false},
		{"新增合集需补关系", ArchiveWorkspaceValues{GroupRelations: true, VideoCollectionIDs: []uuid.UUID{a}}, ArchiveWorkspaceValues{VideoCollectionIDs: []uuid.UUID{a, b}}, false},
		{"取消选择不撤销关系", ArchiveWorkspaceValues{GroupRelations: true, VideoCollectionIDs: []uuid.UUID{a, b}}, ArchiveWorkspaceValues{VideoCollectionIDs: []uuid.UUID{a}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if workspaceRelationsApplied(tc.have, tc.want) != tc.applied {
				t.Fatal("关系完成判定错误")
			}
		})
	}
}

func TestWorkspaceDesktopTargetValidation(t *testing.T) {
	id := uuid.New()
	f := models.ArchiveImportFileListItem{ID: id, EntryType: "file", MediaKind: "video"}
	d := ArchiveWorkspaceDraft{LayoutVersion: 2, Files: map[string]ArchiveWorkspaceFile{id.String(): {GroupID: "g"}}}
	if validateWorkspaceImportTarget(d, f, ArchiveWorkspaceValues{}) == nil {
		t.Fatal("不能绕过视频合集要求")
	}
	if err := validateWorkspaceImportTarget(d, f, ArchiveWorkspaceValues{VideoCollectionIDs: []uuid.UUID{uuid.New()}}); err != nil {
		t.Fatal(err)
	}
	f.MediaKind = "image"
	if validateWorkspaceImportTarget(d, f, ArchiveWorkspaceValues{}) == nil {
		t.Fatal("不能绕过图片合集要求")
	}
	d.Files[id.String()] = ArchiveWorkspaceFile{}
	if validateWorkspaceImportTarget(d, f, ArchiveWorkspaceValues{ImageCollectionIDs: []uuid.UUID{uuid.New()}}) == nil {
		t.Fatal("未分组不可直接入库")
	}
}

func TestWorkspaceDesktopLegacyUngrouped(t *testing.T) {
	for _, test := range []string{"继承主组", "显式相同合集", "独立合集"} {
		t.Run(test, func(t *testing.T) {
			id := uuid.New()
			vc := []uuid.UUID{uuid.New()}
			other := []uuid.UUID{uuid.New()}
			title, description := "上传标题", "原说明"
			tags := []string{"原标签"}
			config := ArchiveWorkspaceConfig{}
			if test == "显式相同合集" {
				config.VideoCollectionIDs = &vc
			}
			if test == "独立合集" {
				config.VideoCollectionIDs = &other
			}
			old := ArchiveWorkspaceDraft{Groups: []ArchiveWorkspaceGroup{{ID: "main", Name: "主组", Config: ArchiveWorkspaceConfig{Title: &title, Description: &description, Tags: &tags, VideoCollectionIDs: &vc}}}, Files: map[string]ArchiveWorkspaceFile{id.String(): {Name: "视频.mp4", Config: config}}}
			files := []models.ArchiveImportFileListItem{{ID: id, EntryType: "file", MediaKind: "video"}}
			before, err := resolveWorkspaceFile(old, id.String(), "video")
			if err != nil {
				t.Fatal(err)
			}
			d := flattenWorkspace(old, files)
			after, err := resolveWorkspaceFile(d, id.String(), "video")
			if err != nil || d.Files[id.String()].GroupID == "" || before.Title != after.Title || before.Description != after.Description || len(after.Tags) != 1 || !workspaceSameIDs(before.VideoCollectionIDs, after.VideoCollectionIDs) {
				t.Fatalf("原未分组配置丢失: %+v %+v %v", before, after, err)
			}
		})
	}
}

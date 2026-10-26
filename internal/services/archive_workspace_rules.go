package services

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"video-server/internal/models"
)

const archiveWorkspaceMain = "main"

// ArchiveWorkspaceConfig uses nil for inheritance and non-nil empty values for explicit clearing.
type ArchiveWorkspaceConfig struct {
	TitleMode          *string      `json:"title_mode,omitempty"`
	Title              *string      `json:"title,omitempty"`
	Description        *string      `json:"description,omitempty"`
	Tags               *[]string    `json:"tags,omitempty"`
	VideoType          *string      `json:"video_type,omitempty"`
	VideoCollectionIDs *[]uuid.UUID `json:"video_collection_ids,omitempty"`
	ImageCollectionIDs *[]uuid.UUID `json:"image_collection_ids,omitempty"`
}

// ArchiveWorkspaceDirectory describes a logical directory, never a disk path.
type ArchiveWorkspaceDirectory struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
}

// ArchiveWorkspaceGroup describes one node in the independently maintained import group tree.
type ArchiveWorkspaceGroup struct {
	ID          string                 `json:"id"`
	ParentID    string                 `json:"parent_id"`
	Name        string                 `json:"name"`
	DirectoryID string                 `json:"directory_id,omitempty"`
	Config      ArchiveWorkspaceConfig `json:"config"`
}

// ArchiveWorkspaceFile holds editable metadata for a stable source-file identifier.
type ArchiveWorkspaceFile struct {
	DirectoryID string                 `json:"directory_id"`
	Name        string                 `json:"name"`
	GroupID     string                 `json:"group_id"`
	Excluded    bool                   `json:"excluded"`
	Config      ArchiveWorkspaceConfig `json:"config"`
}

// ArchiveWorkspaceDraft is persisted independently of extracted file locations.
type ArchiveWorkspaceDraft struct {
	Directories []ArchiveWorkspaceDirectory     `json:"directories"`
	Groups      []ArchiveWorkspaceGroup         `json:"groups"`
	Files       map[string]ArchiveWorkspaceFile `json:"files"`
}

// ArchiveWorkspaceValues is the resolved, immutable configuration of a submitted file.
type ArchiveWorkspaceValues struct {
	TitleMode          string      `json:"title_mode"`
	Title              string      `json:"title"`
	Description        string      `json:"description"`
	Tags               []string    `json:"tags"`
	VideoType          string      `json:"video_type"`
	VideoCollectionIDs []uuid.UUID `json:"video_collection_ids"`
	ImageCollectionIDs []uuid.UUID `json:"image_collection_ids"`
}

func applyWorkspaceConfig(v *ArchiveWorkspaceValues, c ArchiveWorkspaceConfig) {
	if c.TitleMode != nil {
		v.TitleMode = *c.TitleMode
	}
	if c.Title != nil {
		v.Title = *c.Title
	}
	if c.Description != nil {
		v.Description = *c.Description
	}
	if c.Tags != nil {
		v.Tags = normalizeArchiveTags(*c.Tags)
	}
	if c.VideoType != nil {
		v.VideoType = *c.VideoType
	}
	if c.VideoCollectionIDs != nil {
		v.VideoCollectionIDs = dedupeArchiveUUIDs(*c.VideoCollectionIDs)
	}
	if c.ImageCollectionIDs != nil {
		v.ImageCollectionIDs = dedupeArchiveUUIDs(*c.ImageCollectionIDs)
	}
}

func resolveWorkspaceFile(d ArchiveWorkspaceDraft, id, kind string) (ArchiveWorkspaceValues, error) {
	f, ok := d.Files[id]
	if !ok {
		return ArchiveWorkspaceValues{}, fmt.Errorf("文件不在当前工作区")
	}
	groups := make(map[string]ArchiveWorkspaceGroup, len(d.Groups))
	for _, g := range d.Groups {
		groups[g.ID] = g
	}
	groupID := f.GroupID
	if groupID == "" {
		groupID = archiveWorkspaceMain
	}
	chain := []ArchiveWorkspaceGroup{}
	seen := map[string]bool{}
	for groupID != "" {
		g, ok := groups[groupID]
		if !ok || seen[groupID] {
			return ArchiveWorkspaceValues{}, fmt.Errorf("分组层级无效")
		}
		seen[groupID] = true
		chain = append(chain, g)
		groupID = g.ParentID
	}
	v := ArchiveWorkspaceValues{TitleMode: "uniform", VideoType: "short", Tags: []string{}, VideoCollectionIDs: []uuid.UUID{}, ImageCollectionIDs: []uuid.UUID{}}
	for i := len(chain) - 1; i >= 0; i-- {
		applyWorkspaceConfig(&v, chain[i].Config)
	}
	if kind == "image" {
		v.TitleMode = "filename"
	}
	applyWorkspaceConfig(&v, f.Config)
	if f.Config.Title != nil && f.Config.TitleMode == nil {
		v.TitleMode = "uniform"
	}
	if v.TitleMode == "filename" {
		v.Title = deriveArchiveFilenameTitle(f.Name)
	}
	v.Title = strings.TrimSpace(v.Title)
	return v, nil
}

func newArchiveWorkspace(b models.ArchiveImportBatch, files []models.ArchiveImportFileListItem, oldGroups []models.ArchiveImportGroup, fresh bool) ArchiveWorkspaceDraft {
	title := strings.TrimSpace(b.DefaultTitlePrefix)
	if title == "" {
		title = b.Title
	}
	mode, videoType := "uniform", "short"
	tags := append([]string{}, b.DefaultTags...)
	vc := append([]uuid.UUID{}, b.DefaultVideoCollectionIDs...)
	ic := append([]uuid.UUID{}, b.DefaultImageCollectionIDs...)
	d := ArchiveWorkspaceDraft{Directories: []ArchiveWorkspaceDirectory{}, Groups: []ArchiveWorkspaceGroup{{ID: archiveWorkspaceMain, Name: "主分组", Config: ArchiveWorkspaceConfig{TitleMode: &mode, Title: &title, Description: &b.DefaultDescription, Tags: &tags, VideoType: &videoType, VideoCollectionIDs: &vc, ImageCollectionIDs: &ic}}}, Files: map[string]ArchiveWorkspaceFile{}}
	dirs := map[string]string{"": ""}
	var ensureDir func(string) string
	ensureDir = func(p string) string {
		if p == "." || p == "" {
			return ""
		}
		if id, ok := dirs[p]; ok {
			return id
		}
		parent := ensureDir(path.Dir(p))
		id := uuid.NewString()
		dirs[p] = id
		d.Directories = append(d.Directories, ArchiveWorkspaceDirectory{ID: id, ParentID: parent, Name: path.Base(p)})
		return id
	}
	for _, f := range files {
		if f.EntryType == "directory" {
			ensureDir(f.RelativePath)
			continue
		}
		dir := ensureDir(path.Dir(f.RelativePath))
		node := ArchiveWorkspaceFile{DirectoryID: dir, Name: path.Base(f.RelativePath)}
		if !fresh {
			var group *models.ArchiveImportGroup
			for i := range oldGroups {
				if f.GroupID != nil && oldGroups[i].ID == *f.GroupID {
					group = &oldGroups[i]
					node.GroupID = f.GroupID.String()
					break
				}
			}
			o := archiveImportFieldOverridesForFile(f, b, group)
			title, desc, typ := f.Title, f.Description, f.VideoType
			if o.Title || f.MediaKind == "image" {
				node.Config.Title = &title
			}
			if o.Description {
				node.Config.Description = &desc
			}
			if o.Tags {
				node.Config.Tags = workspaceSlicePointer(append([]string{}, f.Tags...))
			}
			if o.VideoType && f.MediaKind == "video" {
				node.Config.VideoType = &typ
			}
			if o.VideoCollectionIDs {
				node.Config.VideoCollectionIDs = workspaceSlicePointer(append([]uuid.UUID{}, f.VideoCollectionIDs...))
			}
			if o.ImageCollectionIDs {
				node.Config.ImageCollectionIDs = workspaceSlicePointer(append([]uuid.UUID{}, f.ImageCollectionIDs...))
			}
		}
		d.Files[f.ID.String()] = node
	}
	if fresh {
		generateWorkspaceDirectoryGroups(&d, files, true)
	} else {
		for _, g := range oldGroups {
			d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: g.ID.String(), ParentID: archiveWorkspaceMain, Name: g.Name, Config: ArchiveWorkspaceConfig{Title: g.Title, Description: g.Description, Tags: workspaceSlicePointer(g.Tags), VideoType: g.VideoType, VideoCollectionIDs: workspaceSlicePointer(g.VideoCollectionIDs), ImageCollectionIDs: workspaceSlicePointer(g.ImageCollectionIDs)}})
		}
	}
	return d
}

func workspaceSlicePointer[T any](v []T) *[]T {
	if v == nil {
		return nil
	}
	out := append([]T{}, v...)
	return &out
}

func generateWorkspaceDirectoryGroups(d *ArchiveWorkspaceDraft, files []models.ArchiveImportFileListItem, initial bool) {
	dirMap := map[string]ArchiveWorkspaceDirectory{}
	for _, dir := range d.Directories {
		dirMap[dir.ID] = dir
	}
	groupForDir := map[string]string{"": archiveWorkspaceMain}
	for _, g := range d.Groups {
		if g.DirectoryID != "" {
			groupForDir[g.DirectoryID] = g.ID
		}
	}
	var ensureGroup func(string) string
	ensureGroup = func(id string) string {
		if g, ok := groupForDir[id]; ok {
			return g
		}
		dir, ok := dirMap[id]
		if !ok {
			return archiveWorkspaceMain
		}
		parent := ensureGroup(dir.ParentID)
		name := dir.Name
		for n := 2; ; n++ {
			found := false
			for _, g := range d.Groups {
				if g.ParentID == parent && strings.EqualFold(g.Name, name) {
					found = true
					break
				}
			}
			if !found {
				break
			}
			name = fmt.Sprintf("%s（%d）", dir.Name, n)
		}
		gid := uuid.NewString()
		d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: gid, ParentID: parent, Name: name, DirectoryID: id})
		groupForDir[id] = gid
		return gid
	}
	for _, file := range files {
		if !workspaceMedia(file) {
			continue
		}
		f, ok := d.Files[file.ID.String()]
		if !ok {
			continue
		}
		g := ensureGroup(f.DirectoryID)
		if initial {
			f.GroupID = g
			d.Files[file.ID.String()] = f
		}
	}
}

func workspaceMedia(f models.ArchiveImportFileListItem) bool {
	return f.EntryType == "file" && (f.MediaKind == "video" || f.MediaKind == "image")
}
func workspaceSuccess(f models.ArchiveImportFileListItem) bool {
	return f.Status == "ready" || f.Status == "existing"
}
func workspaceValidName(s string) bool {
	return s != "" && s == strings.TrimSpace(s) && s != "." && s != ".." && !strings.ContainsAny(s, "/\\\x00\r\n") && utf8.RuneCountInString(s) <= 200
}

func validateWorkspaceConfig(c ArchiveWorkspaceConfig) error {
	if c.TitleMode != nil && *c.TitleMode != "uniform" && *c.TitleMode != "filename" {
		return fmt.Errorf("标题模式无效")
	}
	if c.VideoType != nil && *c.VideoType != "short" && *c.VideoType != "movie" && *c.VideoType != "episode" && *c.VideoType != "av" {
		return fmt.Errorf("视频类型无效")
	}
	if c.Title != nil && utf8.RuneCountInString(*c.Title) > 200 {
		return fmt.Errorf("标题不能超过200字")
	}
	for _, ids := range []*[]uuid.UUID{c.VideoCollectionIDs, c.ImageCollectionIDs} {
		if ids != nil {
			for _, id := range *ids {
				if id == uuid.Nil {
					return fmt.Errorf("合集标识无效")
				}
			}
		}
	}
	return nil
}

func validateWorkspace(d, old ArchiveWorkspaceDraft, files []models.ArchiveImportFileListItem, active map[string]bool) error {
	if len(d.Groups) == 0 || len(d.Groups) > 10000 || len(d.Directories) > 100000 {
		return fmt.Errorf("工作区层级数量无效")
	}
	groups := map[string]ArchiveWorkspaceGroup{}
	names := map[string]bool{}
	for _, g := range d.Groups {
		if g.ID == "" || !workspaceValidGroupName(g.Name) {
			return fmt.Errorf("分组名称或标识无效")
		}
		if _, ok := groups[g.ID]; ok {
			return fmt.Errorf("分组标识重复")
		}
		key := g.ParentID + "/" + strings.ToLower(g.Name)
		if names[key] {
			return fmt.Errorf("同级分组不能重名")
		}
		names[key] = true
		groups[g.ID] = g
		if err := validateWorkspaceConfig(g.Config); err != nil {
			return err
		}
	}
	main, ok := groups[archiveWorkspaceMain]
	if !ok || main.ParentID != "" {
		return fmt.Errorf("主分组不可删除或移动")
	}
	for id := range groups {
		seen := map[string]bool{}
		cur := id
		for cur != "" {
			g, ok := groups[cur]
			if !ok || seen[cur] {
				return fmt.Errorf("分组层级无效或存在循环")
			}
			seen[cur] = true
			if cur != archiveWorkspaceMain && g.ParentID == "" {
				return fmt.Errorf("普通分组必须继承主分组")
			}
			cur = g.ParentID
		}
	}
	dirs := map[string]ArchiveWorkspaceDirectory{}
	names = map[string]bool{}
	for _, dir := range d.Directories {
		if dir.ID == "" || !workspaceValidName(dir.Name) {
			return fmt.Errorf("目录名称或标识无效")
		}
		if _, ok := dirs[dir.ID]; ok {
			return fmt.Errorf("目录标识重复")
		}
		key := dir.ParentID + "/" + strings.ToLower(dir.Name)
		if names[key] {
			return fmt.Errorf("同目录名称重复")
		}
		names[key] = true
		dirs[dir.ID] = dir
	}
	for id := range dirs {
		seen := map[string]bool{}
		for id != "" {
			dir, ok := dirs[id]
			if !ok || seen[id] {
				return fmt.Errorf("目录层级无效或存在循环")
			}
			seen[id] = true
			id = dir.ParentID
		}
	}
	expected := 0
	for _, file := range files {
		if file.EntryType == "directory" {
			continue
		}
		expected++
		id := file.ID.String()
		f, ok := d.Files[id]
		if !ok || !workspaceValidName(f.Name) {
			return fmt.Errorf("源文件不能删除且必须有有效名称")
		}
		if f.DirectoryID != "" {
			if _, ok := dirs[f.DirectoryID]; !ok {
				return fmt.Errorf("文件目录不存在")
			}
		}
		if f.GroupID != "" {
			if _, ok := groups[f.GroupID]; !ok {
				return fmt.Errorf("文件分组不存在")
			}
		}
		key := f.DirectoryID + "/" + strings.ToLower(f.Name)
		if names[key] {
			return fmt.Errorf("同目录下文件或目录不能重名")
		}
		names[key] = true
		if err := validateWorkspaceConfig(f.Config); err != nil {
			return err
		}
		if workspaceSuccess(file) || active[id] || file.Status == "processing" {
			prev, exists := old.Files[id]
			if !exists || prev.GroupID != f.GroupID || prev.Excluded != f.Excluded || !reflect.DeepEqual(prev.Config, f.Config) {
				return fmt.Errorf("已入库或处理中文件不能修改归属、资格和配置")
			}
			if active[id] || file.Status == "processing" {
				if prev.Name != f.Name || prev.DirectoryID != f.DirectoryID || workspaceLogicalPath(old, id) != workspaceLogicalPath(d, id) {
					return fmt.Errorf("处理中文件不能改名或移动")
				}
			}
			oldGroups := map[string]ArchiveWorkspaceGroup{}
			for _, g := range old.Groups {
				oldGroups[g.ID] = g
			}
			cur := prev.GroupID
			for cur != "" {
				g, ok := oldGroups[cur]
				if !ok {
					break
				}
				if _, exists := groups[cur]; !exists {
					return fmt.Errorf("包含已入库或处理中文件的分组及祖先不能删除")
				}
				cur = g.ParentID
			}
		}
	}
	if len(d.Files) != expected {
		return fmt.Errorf("工作区文件清单与源文件不一致")
	}
	return nil
}

func workspaceValidGroupName(s string) bool {
	return s != "" && s == strings.TrimSpace(s) && !strings.ContainsAny(s, "\x00\r\n") && utf8.RuneCountInString(s) <= 2000
}
func workspaceLogicalPath(d ArchiveWorkspaceDraft, id string) string {
	f := d.Files[id]
	parts := []string{f.Name}
	dirs := map[string]ArchiveWorkspaceDirectory{}
	for _, dir := range d.Directories {
		dirs[dir.ID] = dir
	}
	seen := map[string]bool{}
	for cur := f.DirectoryID; cur != "" && !seen[cur]; {
		seen[cur] = true
		dir := dirs[cur]
		parts = append(parts, dir.Name)
		cur = dir.ParentID
	}
	return strings.Join(parts, "/")
}

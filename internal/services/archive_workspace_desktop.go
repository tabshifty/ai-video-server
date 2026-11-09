package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"video-server/internal/models"
)

func workspaceConfigValues(v ArchiveWorkspaceValues) ArchiveWorkspaceConfig {
	return ArchiveWorkspaceConfig{TitleMode: &v.TitleMode, Title: &v.Title, Description: &v.Description, Tags: workspaceSlicePointer(v.Tags), VideoType: &v.VideoType, VideoCollectionIDs: workspaceSlicePointer(v.VideoCollectionIDs), ImageCollectionIDs: workspaceSlicePointer(v.ImageCollectionIDs)}
}

func freshDesktopWorkspace(old ArchiveWorkspaceDraft, files []models.ArchiveImportFileListItem) ArchiveWorkspaceDraft {
	d := ArchiveWorkspaceDraft{LayoutVersion: 2, Template: old.Groups[0].Config, Directories: old.Directories, Files: old.Files, Groups: []ArchiveWorkspaceGroup{}}
	generateDesktopGroups(&d, files, true)
	return d
}

func uniqueWorkspaceGroupName(base string, groups []ArchiveWorkspaceGroup) string {
	name := base
	for n := 2; ; n++ {
		found := false
		for _, g := range groups {
			if strings.EqualFold(g.Name, name) {
				found = true
				break
			}
		}
		if !found {
			return name
		}
		name = fmt.Sprintf("%s（%d）", base, n)
	}
}

func generateDesktopGroups(d *ArchiveWorkspaceDraft, files []models.ArchiveImportFileListItem, initial bool) {
	dirs := map[string]ArchiveWorkspaceDirectory{}
	counts := map[string]int{}
	for _, dir := range d.Directories {
		dirs[dir.ID] = dir
		counts[strings.ToLower(dir.Name)]++
	}
	existing := map[string]string{}
	for _, g := range d.Groups {
		existing[g.DirectoryID] = g.ID
	}
	for _, source := range files {
		if !workspaceMedia(source) {
			continue
		}
		f, ok := d.Files[source.ID.String()]
		if !ok {
			continue
		}
		gid := existing[f.DirectoryID]
		if gid == "" {
			name := "根目录"
			if d.Template.Title != nil && strings.TrimSpace(*d.Template.Title) != "" {
				name = *d.Template.Title
			}
			if dir, ok := dirs[f.DirectoryID]; ok {
				name = dir.Name
				if counts[strings.ToLower(name)] > 1 {
					parts := []string{name}
					seen := map[string]bool{}
					for id := dir.ParentID; id != "" && !seen[id]; {
						seen[id] = true
						p := dirs[id]
						parts = append([]string{p.Name}, parts...)
						id = p.ParentID
					}
					name = strings.Join(parts, " / ")
				}
			}
			gid = uuid.NewString()
			d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: gid, Name: uniqueWorkspaceGroupName(name, d.Groups), DirectoryID: f.DirectoryID, Config: d.Template})
			existing[f.DirectoryID] = gid
		}
		if initial {
			f.GroupID = gid
			d.Files[source.ID.String()] = f
		}
	}
}

// flattenWorkspace captures legacy inheritance before removing parent relationships.
func flattenWorkspace(old ArchiveWorkspaceDraft, files []models.ArchiveImportFileListItem) ArchiveWorkspaceDraft {
	if old.LayoutVersion >= 2 {
		return old
	}
	d := ArchiveWorkspaceDraft{LayoutVersion: 2, Directories: old.Directories, Files: map[string]ArchiveWorkspaceFile{}, Groups: []ArchiveWorkspaceGroup{}}
	byID := map[string]ArchiveWorkspaceGroup{}
	for _, g := range old.Groups {
		byID[g.ID] = g
		if g.ID == archiveWorkspaceMain {
			d.Template = g.Config
		}
	}
	for _, g := range old.Groups {
		temp := old
		temp.Files = map[string]ArchiveWorkspaceFile{"config": {GroupID: g.ID}}
		v, err := resolveWorkspaceFile(temp, "config", "video")
		if err != nil {
			continue
		}
		parts := []string{g.Name}
		seen := map[string]bool{g.ID: true}
		for id := g.ParentID; id != "" && id != archiveWorkspaceMain && !seen[id]; {
			seen[id] = true
			p := byID[id]
			parts = append([]string{p.Name}, parts...)
			id = p.ParentID
		}
		g.Name = uniqueWorkspaceGroupName(strings.Join(parts, " / "), d.Groups)
		g.ParentID = ""
		g.Config = workspaceConfigValues(v)
		// Resolve the raw group title separately: a synthetic file has no filename.
		titleValues := ArchiveWorkspaceValues{}
		chain := []ArchiveWorkspaceGroup{}
		visited := map[string]bool{}
		for id := g.ID; id != "" && !visited[id]; {
			visited[id] = true
			ancestor := byID[id]
			chain = append(chain, ancestor)
			id = ancestor.ParentID
		}
		for i := len(chain) - 1; i >= 0; i-- {
			applyWorkspaceConfig(&titleValues, chain[i].Config)
		}
		g.Config.Title = &titleValues.Title
		d.Groups = append(d.Groups, g)
	}
	split := map[string]string{}
	ungrouped := ""
	for _, source := range files {
		id := source.ID.String()
		f, ok := old.Files[id]
		if !ok {
			continue
		}
		if f.GroupID == "" {
			if ungrouped == "" {
				for _, g := range d.Groups {
					if g.ID == archiveWorkspaceMain {
						ungrouped = uuid.NewString()
						d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: ungrouped, Name: uniqueWorkspaceGroupName("原未分组", d.Groups), Config: g.Config})
						break
					}
				}
			}
			f.GroupID = ungrouped
		}
		if f.Config.VideoCollectionIDs != nil || f.Config.ImageCollectionIDs != nil {
			v, err := resolveWorkspaceFile(old, id, source.MediaKind)
			if err == nil {
				var base ArchiveWorkspaceGroup
				for _, g := range d.Groups {
					if g.ID == f.GroupID || (f.GroupID == "" && g.ID == archiveWorkspaceMain) {
						base = g
						break
					}
				}
				b := ArchiveWorkspaceValues{}
				applyWorkspaceConfig(&b, base.Config)
				if !workspaceSameIDs(b.VideoCollectionIDs, v.VideoCollectionIDs) || !workspaceSameIDs(b.ImageCollectionIDs, v.ImageCollectionIDs) {
					raw, _ := json.Marshal([][]uuid.UUID{v.VideoCollectionIDs, v.ImageCollectionIDs})
					key := f.GroupID + string(raw)
					gid := split[key]
					if gid == "" {
						gid = uuid.NewString()
						c := base.Config
						c.VideoCollectionIDs = workspaceSlicePointer(v.VideoCollectionIDs)
						c.ImageCollectionIDs = workspaceSlicePointer(v.ImageCollectionIDs)
						name := base.Name
						if name == "" {
							name = "原文件配置"
						}
						d.Groups = append(d.Groups, ArchiveWorkspaceGroup{ID: gid, Name: uniqueWorkspaceGroupName(name+" · 独立合集", d.Groups), Config: c})
						split[key] = gid
					}
					f.GroupID = gid
				}
			}
		}
		f.Config.VideoCollectionIDs = nil
		f.Config.ImageCollectionIDs = nil
		d.Files[id] = f
	}
	return d
}

func workspaceSameIDs(a, b []uuid.UUID) bool {
	return workspaceContainsIDs(a, b) && workspaceContainsIDs(b, a)
}
func workspaceContainsIDs(have, want []uuid.UUID) bool {
	set := map[uuid.UUID]bool{}
	for _, id := range have {
		set[id] = true
	}
	for _, id := range want {
		if !set[id] {
			return false
		}
	}
	return true
}

func workspaceRelationsApplied(have, want ArchiveWorkspaceValues) bool {
	return have.GroupRelations && workspaceContainsIDs(have.VideoCollectionIDs, want.VideoCollectionIDs) && workspaceContainsIDs(have.ImageCollectionIDs, want.ImageCollectionIDs)
}

func validateWorkspaceImportTarget(d ArchiveWorkspaceDraft, f models.ArchiveImportFileListItem, v ArchiveWorkspaceValues) error {
	if d.LayoutVersion < 2 {
		return nil
	}
	if d.Files[f.ID.String()].GroupID == "" {
		return fmt.Errorf("请先将文件加入分组")
	}
	if f.MediaKind == "video" && len(v.VideoCollectionIDs) == 0 {
		return fmt.Errorf("视频所在分组必须选择视频合集")
	}
	if f.MediaKind == "image" && len(v.ImageCollectionIDs) == 0 {
		return fmt.Errorf("图片所在分组必须选择图片合集")
	}
	return nil
}

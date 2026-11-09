// Package forumdownload provides the authenticated Telegram forum download boundary.
package forumdownload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"video-server/internal/cloud"
)

// Resource is a persisted ED2K resource and its current download projection.
type Resource struct {
	ID         uuid.UUID `json:"id"`
	Link       string    `json:"-"`
	Name       string    `json:"name"`
	Size       *int64    `json:"size"`
	Status     string    `json:"status"`
	TaskID     *int64    `json:"task_id"`
	RetryCount int       `json:"retry_count"`
}

// Post contains the complete server-owned ED2K resource set.
type Post struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Fingerprint string     `json:"fingerprint"`
	Resources   []Resource `json:"resources"`
}

var errInvalid = errors.New("请求无效")
var errConflict = errors.New("资源或操作已变化，请重新选择并确认")
var errNotFound = errors.New("帖子或资源已不存在")

func selectResources(resources []Resource, ids []uuid.UUID, all bool) ([]Resource, error) {
	if all && len(ids) > 0 {
		return nil, errInvalid
	}
	if all {
		if len(resources) == 0 || len(resources) > 50 {
			return nil, fmt.Errorf("%w：每次选择 1 至 50 条资源", errInvalid)
		}
		return resources, nil
	}
	if len(ids) == 0 || len(ids) > 50 {
		return nil, fmt.Errorf("%w：每次选择 1 至 50 条资源", errInvalid)
	}
	lookup := make(map[uuid.UUID]Resource, len(resources))
	for _, r := range resources {
		lookup[r.ID] = r
	}
	selected := make([]Resource, 0, len(ids))
	seen := map[uuid.UUID]bool{}
	for _, id := range ids {
		r, ok := lookup[id]
		if !ok || seen[id] {
			return nil, errInvalid
		}
		seen[id] = true
		selected = append(selected, r)
	}
	return selected, nil
}

func fingerprint(resources []Resource) string {
	h := sha256.New()
	for _, r := range resources {
		b, _ := json.Marshal([]string{r.ID.String(), r.Link})
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func parseResource(r Resource) Resource {
	l, err := cloud.ParseLink(r.Link)
	if err != nil || l.Kind != cloud.KindED2K {
		r.Name = "无效 ED2K 链接"
		r.Status = "invalid"
		return r
	}
	r.Name = l.Name
	r.Size = l.SizeBytes
	return r
}

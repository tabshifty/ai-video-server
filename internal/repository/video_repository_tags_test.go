package repository

import (
	"reflect"
	"testing"

	"video-server/internal/models"
)

func TestNormalizePopularVideoTagsLimitAndSort(t *testing.T) {
	stats := []models.VideoTagStat{
		{Tag: "剧情", UsedCount: 2},
		{Tag: "动作", UsedCount: 5},
		{Tag: "爱情", UsedCount: 5},
		{Tag: "科幻", UsedCount: 3},
	}

	got := normalizePopularVideoTags(stats, 3)
	want := []models.VideoTagStat{
		{Tag: "动作", UsedCount: 5},
		{Tag: "爱情", UsedCount: 5},
		{Tag: "科幻", UsedCount: 3},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected popular tag order: got %#v want %#v", got, want)
	}
}

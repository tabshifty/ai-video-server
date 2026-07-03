package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppListQueriesHidePendingDeleteVideos(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(".", "app_repository.go"))
	if err != nil {
		t.Fatalf("read app repository source: %v", err)
	}
	source := string(raw)

	tests := []struct {
		name      string
		signature string
		wants     []string
	}{
		{
			name:      "continue watching",
			signature: "func (r *VideoRepository) ContinueWatching(",
			wants: []string{
				"JOIN videos v ON v.id = a.video_id",
				"AND v.status='ready'",
			},
		},
		{
			name:      "uploaded videos",
			signature: "func (r *VideoRepository) GetUploadedVideos(",
			wants: []string{
				"WHERE user_id=$1 AND status<>'pending_delete'",
			},
		},
		{
			name:      "liked and favorited videos",
			signature: "func (r *VideoRepository) GetActionVideos(",
			wants: []string{
				"JOIN videos v ON v.id = a.video_id",
				"AND v.status='ready'",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := functionSourceForTest(t, source, tt.signature)
			for _, want := range tt.wants {
				if !strings.Contains(body, want) {
					t.Fatalf("%s should contain %q:\n%s", tt.signature, want, body)
				}
			}
		})
	}
}

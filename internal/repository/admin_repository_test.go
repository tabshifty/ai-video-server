package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestBuildAdminListTranscodingTasksSQLWithoutStatus(t *testing.T) {
	t.Parallel()

	countSQL, countArgs, listSQL, listArgs := buildAdminListTranscodingTasksSQL("", 1, 20)

	for _, want := range []string{"from transcoding_jobs t where 1=1", "select count(*)"} {
		if !strings.Contains(strings.ToLower(countSQL), want) {
			t.Fatalf("count sql missing %q: %s", want, countSQL)
		}
	}
	if len(countArgs) != 0 {
		t.Fatalf("expected no count args, got %#v", countArgs)
	}
	if !strings.Contains(strings.ToLower(listSQL), "limit $1 offset $2") {
		t.Fatalf("list sql should use limit/offset placeholders 1/2: %s", listSQL)
	}
	if !strings.Contains(strings.ToLower(listSQL), "left join videos") || !strings.Contains(strings.ToLower(listSQL), "video_title") {
		t.Fatalf("list sql should include video title for task identification: %s", listSQL)
	}
	if len(listArgs) != 2 || listArgs[0] != 20 || listArgs[1] != 0 {
		t.Fatalf("unexpected list args: %#v", listArgs)
	}
}

func TestBuildAdminListTranscodingTasksSQLWithStatus(t *testing.T) {
	t.Parallel()

	countSQL, countArgs, listSQL, listArgs := buildAdminListTranscodingTasksSQL("FAILED", 3, 15)

	if !strings.Contains(strings.ToLower(countSQL), "t.status = $1") {
		t.Fatalf("count sql should filter by status: %s", countSQL)
	}
	if len(countArgs) != 1 || countArgs[0] != "failed" {
		t.Fatalf("unexpected count args: %#v", countArgs)
	}
	if !strings.Contains(strings.ToLower(listSQL), "where 1=1 and t.status = $1") {
		t.Fatalf("list sql should keep status placeholder stable: %s", listSQL)
	}
	if !strings.Contains(strings.ToLower(listSQL), "limit $2 offset $3") {
		t.Fatalf("list sql should shift limit/offset placeholders: %s", listSQL)
	}
	if len(listArgs) != 3 || listArgs[0] != "failed" || listArgs[1] != 15 || listArgs[2] != 30 {
		t.Fatalf("unexpected list args: %#v", listArgs)
	}
}

func TestValidateAdminVideoStatusEditRejectsPendingDeleteManualChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		video  models.Video
		status string
	}{
		{
			name:   "cannot manually enter pending delete",
			video:  models.Video{Type: "short", Status: "ready"},
			status: "pending_delete",
		},
		{
			name:   "cannot manually restore pending delete short",
			video:  models.Video{Type: "short", Status: "pending_delete"},
			status: "ready",
		},
		{
			name:   "cannot manually edit malformed pending delete row",
			video:  models.Video{Type: "movie", Status: "pending_delete"},
			status: "ready",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateAdminVideoStatusEdit(tt.video, tt.status); !errors.Is(err, ErrAdminPendingDeleteStatusEdit) {
				t.Fatalf("expected ErrAdminPendingDeleteStatusEdit, got %v", err)
			}
		})
	}
}

func TestValidateAdminVideoStatusEditAllowsNormalStatus(t *testing.T) {
	t.Parallel()

	err := validateAdminVideoStatusEdit(models.Video{Type: "movie", Status: "processing"}, "ready")
	if err != nil {
		t.Fatalf("expected normal status edit to pass, got %v", err)
	}
}

func TestUpdateVideoStatusRejectsPendingDeleteWorkflowBypass(t *testing.T) {
	t.Parallel()

	repo := &VideoRepository{}
	err := repo.UpdateVideoStatus(context.Background(), uuid.New(), " pending_delete ")
	if !errors.Is(err, ErrAdminPendingDeleteStatusEdit) {
		t.Fatalf("expected ErrAdminPendingDeleteStatusEdit, got %v", err)
	}
}

func TestAdminPendingDeleteMutationsUseRowLocks(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(".", "admin_repository.go"))
	if err != nil {
		t.Fatalf("read admin repository source: %v", err)
	}
	source := string(raw)

	adminUpdate := functionSourceForTest(t, source, "func (r *VideoRepository) AdminUpdateVideo(")
	if !strings.Contains(adminUpdate, "lockVideoForAdminUpdate(ctx, tx, videoID)") {
		t.Fatalf("AdminUpdateVideo should lock the video row before changing type/status:\n%s", adminUpdate)
	}
	if !strings.Contains(adminUpdate, "validateAdminVideoStatusEdit(video, requestedStatus)") {
		t.Fatalf("AdminUpdateVideo should validate requested status before field updates:\n%s", adminUpdate)
	}
	if !strings.Contains(adminUpdate, "UPDATE videos SET status=$2, pending_delete_at=NULL") {
		t.Fatalf("AdminUpdateVideo should update manual status inside the same transaction:\n%s", adminUpdate)
	}

	statusUpdate := functionSourceForTest(t, source, "func (r *VideoRepository) AdminUpdateVideoStatus(")
	if !strings.Contains(statusUpdate, "lockVideoForAdminUpdate(ctx, tx, videoID)") {
		t.Fatalf("AdminUpdateVideoStatus should lock the video row before changing status:\n%s", statusUpdate)
	}
	if !strings.Contains(statusUpdate, "validateAdminVideoStatusEdit(video, status)") {
		t.Fatalf("AdminUpdateVideoStatus should validate the locked row:\n%s", statusUpdate)
	}

	lockHelper := functionSourceForTest(t, source, "func lockVideoForAdminUpdate(")
	if !strings.Contains(lockHelper, "FOR UPDATE") {
		t.Fatalf("lockVideoForAdminUpdate should use SELECT FOR UPDATE:\n%s", lockHelper)
	}
}

func TestShortPendingDeleteWorkflowSourceGuards(t *testing.T) {
	t.Parallel()

	adminRaw, err := os.ReadFile(filepath.Join(".", "admin_repository.go"))
	if err != nil {
		t.Fatalf("read admin repository source: %v", err)
	}
	videoRaw, err := os.ReadFile(filepath.Join(".", "video_repository.go"))
	if err != nil {
		t.Fatalf("read video repository source: %v", err)
	}
	adminSource := string(adminRaw)
	videoSource := string(videoRaw)

	listPending := functionSourceForTest(t, adminSource, "func (r *VideoRepository) AdminListPendingDeleteShortVideos(")
	if !strings.Contains(listPending, "WHERE v.type='short' AND v.status='pending_delete'") {
		t.Fatalf("pending delete list should only expose pending short videos:\n%s", listPending)
	}
	if !strings.Contains(listPending, "ORDER BY v.pending_delete_at DESC NULLS LAST") {
		t.Fatalf("pending delete list should use internal pending_delete_at ordering:\n%s", listPending)
	}

	markPending := functionSourceForTest(t, videoSource, "func (r *VideoRepository) MarkShortVideoPendingDelete(")
	for _, want := range []string{
		"FOR UPDATE",
		`strings.ToLower(strings.TrimSpace(video.Type)) != "short"`,
		`video.Status == "pending_delete"`,
		`video.Status != "ready"`,
		"pending_delete_at=NOW()",
	} {
		if !strings.Contains(markPending, want) {
			t.Fatalf("MarkShortVideoPendingDelete should contain %q:\n%s", want, markPending)
		}
	}
	alreadyPendingIndex := strings.Index(markPending, `video.Status == "pending_delete"`)
	requiresReadyIndex := strings.Index(markPending, `video.Status != "ready"`)
	if alreadyPendingIndex < 0 || requiresReadyIndex < 0 || alreadyPendingIndex > requiresReadyIndex {
		t.Fatalf("idempotent pending_delete branch should run before ready-state rejection:\n%s", markPending)
	}

	keepPending := functionSourceForTest(t, videoSource, "func (r *VideoRepository) KeepPendingDeleteShortVideo(")
	for _, want := range []string{
		"FOR UPDATE",
		`video.Status != "pending_delete"`,
		"SET status='ready'",
		"pending_delete_at=NULL",
	} {
		if !strings.Contains(keepPending, want) {
			t.Fatalf("KeepPendingDeleteShortVideo should contain %q:\n%s", want, keepPending)
		}
	}
}

func functionSourceForTest(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("missing function %s", signature)
	}
	rest := source[start+len(signature):]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len(signature)+next]
}

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

type hermesForumRepositoryStub struct {
	adminListPage     int
	adminListPageSize int
	adminListItems    []models.AdminForumPostListItem
	adminListTotal    int
	adminListErr      error
	discoverInput     models.ForumPostDiscoverInput
	discoverResults   []models.ForumPostDiscoverResult
	discoverErr       error
	inspectionPostID  uuid.UUID
	inspectionInput   models.ForumPostInspectionInput
	inspectionHash    string
	inspectionResult  models.ForumPostInspectionResult
	inspectionErr     error
}

func (s *hermesForumRepositoryStub) ListAdminForumPosts(_ context.Context, page, pageSize int) ([]models.AdminForumPostListItem, int, error) {
	s.adminListPage = page
	s.adminListPageSize = pageSize
	return s.adminListItems, s.adminListTotal, s.adminListErr
}

func (s *hermesForumRepositoryStub) DiscoverForumPosts(_ context.Context, input models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error) {
	s.discoverInput = input
	return s.discoverResults, s.discoverErr
}

func (s *hermesForumRepositoryStub) CompleteForumPostInspection(_ context.Context, postID uuid.UUID, input models.ForumPostInspectionInput, fingerprint string) (models.ForumPostInspectionResult, error) {
	s.inspectionPostID = postID
	s.inspectionInput = input
	s.inspectionHash = fingerprint
	return s.inspectionResult, s.inspectionErr
}

func TestHermesForumServiceListsAdminReadModel(t *testing.T) {
	t.Parallel()

	wantItems := []models.AdminForumPostListItem{{
		ID:          uuid.New(),
		Title:       "帖子标题",
		URL:         "https://sehuatang.org/thread-1",
		Attachments: []string{"https://sehuatang.org/attachment.php?aid=1"},
		ED2KLinks:   []string{"ed2k://|file|a.zip|1|0123456789ABCDEF0123456789ABCDEF|/"},
	}}
	repo := &hermesForumRepositoryStub{adminListItems: wantItems, adminListTotal: 21}

	items, total, err := NewHermesForumService(repo).ListAdmin(context.Background(), 2, 20)
	if err != nil {
		t.Fatalf("ListAdmin returned error: %v", err)
	}
	if repo.adminListPage != 2 || repo.adminListPageSize != 20 {
		t.Fatalf("repository pagination=(%d,%d), want (2,20)", repo.adminListPage, repo.adminListPageSize)
	}
	if len(items) != 1 || items[0].ID != wantItems[0].ID || total != 21 {
		t.Fatalf("items=%#v total=%d, want=%#v total=21", items, total, wantItems)
	}
}

func TestHermesForumServiceDiscoverValidatesWholeBatch(t *testing.T) {
	t.Parallel()

	observedAt := time.Date(2026, 8, 4, 10, 30, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	valid := models.ForumPostDiscoverInput{
		Mode:       models.ForumPostDiscoverModeNormal,
		Source:     "sehuatang",
		BoardKey:   "95",
		ObservedAt: &observedAt,
		Posts: []models.ForumPostCandidate{{
			TID:   "3670055",
			Title: "帖子标题",
			URL:   "https://sehuatang.org/forum.php?mod=viewthread&tid=3670055",
		}},
	}

	tests := []struct {
		name   string
		mutate func(*models.ForumPostDiscoverInput)
	}{
		{name: "empty posts", mutate: func(in *models.ForumPostDiscoverInput) { in.Posts = nil }},
		{name: "too many posts", mutate: func(in *models.ForumPostDiscoverInput) { in.Posts = make([]models.ForumPostCandidate, 101) }},
		{name: "invalid mode", mutate: func(in *models.ForumPostDiscoverInput) { in.Mode = "other" }},
		{name: "missing source", mutate: func(in *models.ForumPostDiscoverInput) { in.Source = "" }},
		{name: "normal missing observed time", mutate: func(in *models.ForumPostDiscoverInput) { in.ObservedAt = nil }},
		{name: "normal missing title", mutate: func(in *models.ForumPostDiscoverInput) { in.Posts[0].Title = "" }},
		{name: "invalid url", mutate: func(in *models.ForumPostDiscoverInput) { in.Posts[0].URL = "relative/path" }},
		{name: "duplicate tid in batch", mutate: func(in *models.ForumPostDiscoverInput) { in.Posts = append(in.Posts, in.Posts[0]) }},
		{name: "dedupe only observed time", mutate: func(in *models.ForumPostDiscoverInput) {
			in.Mode = models.ForumPostDiscoverModeDedupeOnly
			in.Posts[0].Title = ""
			in.Posts[0].URL = ""
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			input.Posts = append([]models.ForumPostCandidate(nil), valid.Posts...)
			tt.mutate(&input)
			repo := &hermesForumRepositoryStub{}
			_, err := NewHermesForumService(repo).Discover(context.Background(), input)
			if !errors.Is(err, ErrHermesForumInvalidInput) {
				t.Fatalf("error=%v, want ErrHermesForumInvalidInput", err)
			}
			if len(repo.discoverInput.Posts) != 0 {
				t.Fatal("repository must not be called for invalid batch")
			}
		})
	}
}

func TestHermesForumServiceDiscoverNormalizesAndDelegates(t *testing.T) {
	t.Parallel()

	observedAt := time.Now()
	wantResult := []models.ForumPostDiscoverResult{{TID: "3670055", ID: uuid.New(), Disposition: models.ForumPostDispositionCreated, InspectionStatus: models.ForumPostStatusPending}}
	repo := &hermesForumRepositoryStub{discoverResults: wantResult}
	svc := NewHermesForumService(repo)

	got, err := svc.Discover(context.Background(), models.ForumPostDiscoverInput{
		Mode:       models.ForumPostDiscoverModeNormal,
		Source:     " sehuatang ",
		BoardKey:   " 95 ",
		ObservedAt: &observedAt,
		Posts: []models.ForumPostCandidate{{
			TID:   " 3670055 ",
			Title: " 帖子标题 ",
			URL:   " https://sehuatang.org/forum.php?mod=viewthread&tid=3670055 ",
		}},
	})
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(got) != 1 || got[0] != wantResult[0] {
		t.Fatalf("result=%v, want=%v", got, wantResult)
	}
	if repo.discoverInput.Source != "sehuatang" || repo.discoverInput.BoardKey != "95" {
		t.Fatalf("unexpected normalized source/board: %#v", repo.discoverInput)
	}
	post := repo.discoverInput.Posts[0]
	if post.TID != "3670055" || post.Title != "帖子标题" || post.URL != "https://sehuatang.org/forum.php?mod=viewthread&tid=3670055" {
		t.Fatalf("unexpected normalized post: %#v", post)
	}
}

func TestHermesForumServiceAllowsDedupeOnlyTID(t *testing.T) {
	t.Parallel()

	repo := &hermesForumRepositoryStub{}
	_, err := NewHermesForumService(repo).Discover(context.Background(), models.ForumPostDiscoverInput{
		Mode:     models.ForumPostDiscoverModeDedupeOnly,
		Source:   "sehuatang",
		BoardKey: "95",
		Posts:    []models.ForumPostCandidate{{TID: "3670055"}},
	})
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
}

func TestHermesForumServiceCompleteInspectionPreservesResourcesAndBuildsStableFingerprint(t *testing.T) {
	t.Parallel()

	postID := uuid.New()
	input := models.ForumPostInspectionInput{
		Status:         models.ForumPostStatusRestricted,
		FilterDecision: models.ForumPostFilterIncluded,
		FilterReasons:  []string{" attachment ", "restricted"},
		FetchMethod:    " chrome-cdp ",
		ErrorSummary:   " permission required ",
		Attachments:    []string{" https://sehuatang.org/attachment.php?aid=1 ", "https://sehuatang.org/attachment.php?aid=2"},
		ED2KLinks:      []string{" ed2k://|file|a.zip|1|0123456789ABCDEF0123456789ABCDEF|/ ", "ed2k://|file|b.zip|2|FEDCBA9876543210FEDCBA9876543210|/"},
	}
	repo := &hermesForumRepositoryStub{inspectionResult: models.ForumPostInspectionResult{Status: models.ForumPostStatusRestricted}}
	svc := NewHermesForumService(repo)

	if _, err := svc.CompleteInspection(context.Background(), postID, input); err != nil {
		t.Fatalf("CompleteInspection returned error: %v", err)
	}
	firstHash := repo.inspectionHash
	if len(firstHash) != 64 {
		t.Fatalf("fingerprint length=%d, want=64", len(firstHash))
	}
	if repo.inspectionInput.Attachments[0] != "https://sehuatang.org/attachment.php?aid=1" || repo.inspectionInput.ED2KLinks[1] != input.ED2KLinks[1] {
		t.Fatalf("resources were not normalized/preserved: %#v", repo.inspectionInput)
	}

	if _, err := svc.CompleteInspection(context.Background(), postID, input); err != nil {
		t.Fatalf("second CompleteInspection returned error: %v", err)
	}
	if repo.inspectionHash != firstHash {
		t.Fatalf("fingerprint changed: first=%s second=%s", firstHash, repo.inspectionHash)
	}
}

func TestHermesForumServiceCompleteInspectionRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	valid := models.ForumPostInspectionInput{
		Status:         models.ForumPostStatusInspected,
		FilterDecision: models.ForumPostFilterIncluded,
		FilterReasons:  []string{"ed2k"},
		ED2KLinks:      []string{"ed2k://|file|a.zip|1|0123456789ABCDEF0123456789ABCDEF|/"},
	}
	tests := []struct {
		name   string
		mutate func(*models.ForumPostInspectionInput)
	}{
		{name: "nil post id", mutate: func(*models.ForumPostInspectionInput) {}},
		{name: "invalid status", mutate: func(in *models.ForumPostInspectionInput) { in.Status = models.ForumPostStatusPending }},
		{name: "invalid decision", mutate: func(in *models.ForumPostInspectionInput) { in.FilterDecision = "maybe" }},
		{name: "invalid attachment", mutate: func(in *models.ForumPostInspectionInput) { in.Attachments = []string{"file:///tmp/a"} }},
		{name: "invalid ed2k", mutate: func(in *models.ForumPostInspectionInput) { in.ED2KLinks = []string{"magnet:?xt=1"} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			input.FilterReasons = append([]string(nil), valid.FilterReasons...)
			input.Attachments = append([]string(nil), valid.Attachments...)
			input.ED2KLinks = append([]string(nil), valid.ED2KLinks...)
			postID := uuid.New()
			if tt.name == "nil post id" {
				postID = uuid.Nil
			}
			tt.mutate(&input)
			repo := &hermesForumRepositoryStub{}
			_, err := NewHermesForumService(repo).CompleteInspection(context.Background(), postID, input)
			if !errors.Is(err, ErrHermesForumInvalidInput) {
				t.Fatalf("error=%v, want ErrHermesForumInvalidInput", err)
			}
			if repo.inspectionPostID != uuid.Nil {
				t.Fatal("repository must not be called for invalid input")
			}
		})
	}
}

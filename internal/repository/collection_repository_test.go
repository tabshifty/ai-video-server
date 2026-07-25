package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCollectionName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":                "",
		"   ":             "",
		" 热 门 合 集 ":       "热 门 合 集",
		"  Mix\t\nName  ": "mix name",
		"  A   B   C  ":   "a b c",
		"\n\t中文  合集  名  ": "中文 合集 名",
	}

	for raw, want := range cases {
		got := normalizeCollectionName(raw)
		if got != want {
			t.Fatalf("normalizeCollectionName(%q)=%q want=%q", raw, got, want)
		}
	}
}

func TestDedupeCollectionIDs(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	got := dedupeCollectionIDs([]uuid.UUID{id1, id2, id1, id3, id2, id3})
	if len(got) != 3 {
		t.Fatalf("len=%d want=3", len(got))
	}
	if got[0] != id1 || got[1] != id2 || got[2] != id3 {
		t.Fatalf("unexpected dedupe order: %v", got)
	}
}

func TestShortVideosByCollectionVisibilityClauseLocksShortReadyAndActive(t *testing.T) {
	t.Parallel()
	upper := strings.ToUpper(shortVideosByCollectionVisibilityClause)
	if strings.Contains(upper, "DISTINCT") {
		t.Fatalf("visibility clause must not use DISTINCT, got: %s", shortVideosByCollectionVisibilityClause)
	}
	for _, want := range []string{"c.active = TRUE", "v.type = 'short'", "v.status = 'ready'", "EXISTS ("} {
		if !strings.Contains(shortVideosByCollectionVisibilityClause, want) {
			t.Fatalf("visibility clause must contain %q, got: %s", want, shortVideosByCollectionVisibilityClause)
		}
	}
}

func TestResolveAppShortCollectionCoverURLPrefersAdminCover(t *testing.T) {
	t.Parallel()
	if got := resolveAppShortCollectionCoverURL("https://cdn/x.jpg", uuid.New()); got != "https://cdn/x.jpg" {
		t.Fatalf("admin cover_url must win, got %q", got)
	}
	if got := resolveAppShortCollectionCoverURL("  ", uuid.Nil); got != "" {
		t.Fatalf("empty cover and nil video must yield empty, got %q", got)
	}
}

func TestResolveAppShortCollectionCoverURLFallsBackToLatestPlayableVideoThumbnail(t *testing.T) {
	t.Parallel()
	videoID := uuid.New()
	got := resolveAppShortCollectionCoverURL("", videoID)
	if !strings.HasSuffix(got, videoID.String()+"/thumbnail") {
		t.Fatalf("fallback cover must point to latest playable video thumbnail, got %q", got)
	}
}

// TestDiscoverShortVideosAndSearchVideosShareCollectionQueryPath 守护 ADR-0016 不变量：
// DiscoverShortVideos(mode=collection) 与 SearchVideosOrdered(collectionID 非空) 必须共用
// queryShortVideosByCollection 这一份 SQL，不能各自维护分叉的查询。DiscoverShortVideos 的
// collection 分支直接 return 共享函数，这里用零值 pool 触发共享函数的 nil-id 校验分支，
// 断言 DiscoverShortVideos(collection, nil) 与 queryShortVideosByCollection(nil) 返回完全
// 一致的错误信息，证明前者委托共享实现而未在分支里另写 SQL。
func TestDiscoverShortVideosAndSearchVideosShareCollectionQueryPath(t *testing.T) {
	t.Parallel()
	repo := &VideoRepository{}

	_, _, errShared := repo.queryShortVideosByCollection(context.Background(), (*uuid.UUID)(nil), 10, 0)
	if errShared == nil || !strings.Contains(errShared.Error(), "collection_id is required") {
		t.Fatalf("queryShortVideosByCollection(nil) must surface nil-id guard error, got %v", errShared)
	}

	_, _, errDiscover := repo.DiscoverShortVideos(context.Background(), "collection", "", nil, 10, 0)
	if errDiscover == nil || errDiscover.Error() != errShared.Error() {
		t.Fatalf("DiscoverShortVideos(collection,nil) must delegate to shared query (same error), got %v want %v", errDiscover, errShared)
	}
}

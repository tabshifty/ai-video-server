package repository

import "testing"

func TestNormalizeAdminVideoTags(t *testing.T) {
	t.Parallel()

	got := normalizeAdminVideoTags([]string{" 热门 ", "HOT", "hot", "", "动作", "动作 "})
	want := []string{"热门", "hot", "动作"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want=%d values=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d got=%q want=%q all=%v", i, got[i], want[i], got)
		}
	}
}

func TestMergeAdminVideoTagsAppend(t *testing.T) {
	t.Parallel()

	got, err := mergeAdminVideoTags([]string{"动作", "剧情"}, []string{"剧情", "悬疑"}, AdminVideoTagsModeAppend)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []string{"动作", "剧情", "悬疑"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want=%d values=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d got=%q want=%q all=%v", i, got[i], want[i], got)
		}
	}
}

func TestMergeAdminVideoTagsRemove(t *testing.T) {
	t.Parallel()

	got, err := mergeAdminVideoTags([]string{"动作", "剧情", "悬疑"}, []string{"剧情", "不存在"}, AdminVideoTagsModeRemove)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []string{"动作", "悬疑"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want=%d values=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d got=%q want=%q all=%v", i, got[i], want[i], got)
		}
	}
}

func TestMergeAdminVideoTagsInvalidMode(t *testing.T) {
	t.Parallel()

	if _, err := mergeAdminVideoTags([]string{"动作"}, []string{"剧情"}, AdminVideoTagsMode("invalid")); err != ErrAdminVideoTagsModeInvalid {
		t.Fatalf("expected invalid mode error, got %v", err)
	}
}

func TestMergeAdminVideoTagsRemoveEmptyTargetKeepsExisting(t *testing.T) {
	t.Parallel()

	got, err := mergeAdminVideoTags([]string{"动作", "剧情"}, nil, AdminVideoTagsModeRemove)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []string{"动作", "剧情"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want=%d values=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d got=%q want=%q all=%v", i, got[i], want[i], got)
		}
	}
}

func TestNormalizeAdminVideoTagsEmptyInput(t *testing.T) {
	t.Parallel()

	got := normalizeAdminVideoTags([]string{"", "   "})
	if len(got) != 0 {
		t.Fatalf("expected empty normalized tags, got %v", got)
	}
}

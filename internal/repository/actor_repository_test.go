package repository

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeActorName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trim and lowercase",
			in:   "  Alice  Bob  ",
			want: "alice bob",
		},
		{
			name: "collapse spaces",
			in:   "张三   李四",
			want: "张三 李四",
		},
		{
			name: "empty input",
			in:   "   ",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeActorName(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeActorName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildActorWorksQueryFiltersReadyVideosByActorAndOrdersNewestFirst(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	query, args := buildActorWorksQuery(actorID, 24, 48)
	normalized := strings.ToLower(strings.Join(strings.Fields(query), " "))

	for _, want := range []string{
		"join video_actors va on va.video_id = v.id",
		"where va.actor_id=$1",
		"and v.status='ready'",
		"order by v.created_at desc",
		"limit $2 offset $3",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("expected actor works query to contain %q, got: %s", want, normalized)
		}
	}
	if len(args) != 3 || args[0] != actorID || args[1] != 24 || args[2] != 48 {
		t.Fatalf("unexpected actor works args: %#v", args)
	}
}

func TestBuildActorWorksCountQueryCountsOnlyReadyVideosForActor(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	query, args := buildActorWorksCountQuery(actorID)
	normalized := strings.ToLower(strings.Join(strings.Fields(query), " "))

	for _, want := range []string{
		"select count(*)",
		"from video_actors va",
		"join videos v on v.id = va.video_id",
		"where va.actor_id=$1",
		"and v.status='ready'",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("expected actor works count query to contain %q, got: %s", want, normalized)
		}
	}
	if len(args) != 1 || args[0] != actorID {
		t.Fatalf("unexpected actor works count args: %#v", args)
	}
}

func TestBuildUpsertScrapedActorProfileSQLPrefersExternalIDAndPreservesCuratedFields(t *testing.T) {
	t.Parallel()

	query := buildUpsertScrapedActorProfileSQL()
	normalized := strings.ToLower(strings.Join(strings.Fields(query), " "))

	for _, want := range []string{
		"where ($10 <> '' and source=$9 and external_id=$10) or normalized_name=$3",
		"case when $10 <> '' and source=$9 and external_id=$10 then 0 else 1 end",
		"avatar_url = case when coalesce(a.avatar_url,'') = '' then $8 else a.avatar_url end",
		"notes = case when coalesce(a.notes,'') = '' then $11 else a.notes end",
		"on conflict(normalized_name) do update",
		"external_id = case when coalesce(actors.external_id,'') = '' then excluded.external_id else actors.external_id end",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("expected scraped actor upsert query to contain %q, got: %s", want, normalized)
		}
	}
}

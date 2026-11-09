package forumdownload

import (
	"testing"

	"github.com/google/uuid"
)

func TestSelectResources(t *testing.T) {
	resources := []Resource{{ID: uuid.New(), Link: "ed2k://|file|甲.zip|123|0123456789abcdef0123456789abcdef|/"}, {ID: uuid.New(), Link: "invalid"}}
	for _, tt := range []struct {
		name string
		ids  []uuid.UUID
		all  bool
		want int
		bad  bool
	}{
		{"all retains invalid for feedback", nil, true, 2, false},
		{"explicit selection", []uuid.UUID{resources[0].ID}, false, 1, false},
		{"foreign resource rejected", []uuid.UUID{uuid.New()}, false, 0, true},
		{"duplicate selection rejected", []uuid.UUID{resources[0].ID, resources[0].ID}, false, 0, true},
		{"empty rejected", nil, false, 0, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectResources(resources, tt.ids, tt.all)
			if (err != nil) != tt.bad || len(got) != tt.want {
				t.Fatalf("got %v, %v", got, err)
			}
		})
	}
	many := make([]Resource, 51)
	if _, err := selectResources(many, nil, true); err == nil {
		t.Fatal("expected batch limit")
	}
}

func TestResourceFingerprintIncludesIdentityAndLink(t *testing.T) {
	r := []Resource{{ID: uuid.New(), Link: "first"}}
	original := fingerprint(r)
	r[0].Link = "second"
	if fingerprint(r) == original {
		t.Fatal("link change was not detected")
	}
	r[0].Link = "first"
	r[0].ID = uuid.New()
	if fingerprint(r) == original {
		t.Fatal("identity change was not detected")
	}
}

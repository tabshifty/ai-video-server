package services

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func TestNewTVAuthSessionBuildsPendingAndroidTVSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC)
	session, err := newTVAuthSession("living-room-01", "客厅电视", now)
	if err != nil {
		t.Fatalf("newTVAuthSession returned err=%v", err)
	}
	if session.Status != "pending" {
		t.Fatalf("expected pending status, got=%s", session.Status)
	}
	if session.Platform != "android_tv" {
		t.Fatalf("expected android_tv platform, got=%s", session.Platform)
	}
	if session.ExpiresAt.Sub(now) != tvAuthSessionTTL {
		t.Fatalf("unexpected ttl: got=%s want=%s", session.ExpiresAt.Sub(now), tvAuthSessionTTL)
	}
	if len(session.PairCode) != 6 {
		t.Fatalf("expected 6 char pair code, got=%q", session.PairCode)
	}
}

func TestBuildTVQRContentCarriesServerAndSession(t *testing.T) {
	t.Parallel()

	session := models.TvAuthSession{
		ID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		PairCode:   "ABC234",
		DeviceName: "客厅电视",
	}

	raw := buildTVQRContent("http://192.168.1.8:8080", session)
	if !strings.HasPrefix(raw, "cheevideos://tv-auth?") {
		t.Fatalf("unexpected scheme: %s", raw)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse qr content: %v", err)
	}
	if got := parsed.Query().Get("server"); got != "http://192.168.1.8:8080" {
		t.Fatalf("unexpected server in qr content: %s", got)
	}
	if got := parsed.Query().Get("session_id"); got != session.ID.String() {
		t.Fatalf("unexpected session id in qr content: %s", got)
	}
}

func TestBuildTVSearchPayloadExcludesAVContent(t *testing.T) {
	t.Parallel()

	payload := buildTVSearchPayload(
		1,
		20,
		[]models.TvSeriesSummaryDto{{ID: 7, Title: "雾城档案"}},
		[]models.VideoListItem{{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Title: "午夜列车", Type: "movie", ThumbnailPath: "/movie.jpg"}},
		[]models.VideoListItem{{ID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), Title: "SNIS-001", Type: "av", ThumbnailPath: "/av.jpg"}},
		3,
	)

	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 items, got=%d", len(payload.Items))
	}
	gotTypes := []string{payload.Items[0].Type, payload.Items[1].Type}
	if strings.Join(gotTypes, ",") != "movie,tv" {
		t.Fatalf("unexpected types ordering/content: %v", gotTypes)
	}
}

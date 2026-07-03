package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func TestStartTVRemoteSessionRejectsOfflineDevice(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 3, 14, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-(tvRemoteDeviceSeenTTL + time.Second))
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := &fakeTVRemoteRepository{
		device: models.TvDeviceRecord{
			DeviceID:   "living-room",
			DeviceName: "客厅电视",
			Platform:   tvRemotePlatform,
			UserID:     &userID,
			LastSeenAt: &lastSeenAt,
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	_, err := svc.StartTVRemoteSessionAt(context.Background(), userID, "living-room", []models.TvRemoteSessionItem{
		{VideoID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Title: "条目 1"},
	}, 0, now)
	if !errors.Is(err, ErrTVRemoteDeviceOffline) {
		t.Fatalf("expected ErrTVRemoteDeviceOffline, got=%v", err)
	}
}

func TestStartTVRemoteSessionCreatesSessionWithCurrentItem(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 3, 14, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-5 * time.Second)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	videoID1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	videoID2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	repo := &fakeTVRemoteRepository{
		device: models.TvDeviceRecord{
			DeviceID:   "living-room",
			DeviceName: "客厅电视",
			Platform:   tvRemotePlatform,
			UserID:     &userID,
			LastSeenAt: &lastSeenAt,
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	session, err := svc.StartTVRemoteSessionAt(context.Background(), userID, " living-room ", []models.TvRemoteSessionItem{
		{VideoID: videoID1, Title: " 第一条 ", ThumbnailPath: " /thumb-1.jpg ", Duration: 11, Type: "short"},
		{VideoID: videoID2, Title: "第二条", ThumbnailPath: "/thumb-2.jpg", Duration: 12, Type: "short"},
	}, 1, now)
	if err != nil {
		t.Fatalf("StartTVRemoteSessionAt returned err=%v", err)
	}
	if repo.created == nil {
		t.Fatal("expected repository CreateOrReplaceTVRemoteSession to be called")
	}
	if session.DeviceID != "living-room" || session.DeviceName != "客厅电视" {
		t.Fatalf("unexpected device info: %#v", session)
	}
	if session.CurrentVideoID == nil || *session.CurrentVideoID != videoID2 {
		t.Fatalf("unexpected current video id: %#v", session.CurrentVideoID)
	}
	if session.CurrentItem == nil || session.CurrentItem.VideoID != videoID2 {
		t.Fatalf("unexpected current item: %#v", session.CurrentItem)
	}
	if !session.HasPrevious || session.HasNext {
		t.Fatalf("unexpected navigation flags: prev=%v next=%v", session.HasPrevious, session.HasNext)
	}
	if got := session.Items[0].Title; got != "第一条" {
		t.Fatalf("expected trimmed title, got=%q", got)
	}
	if got := session.Items[0].ThumbnailPath; got != "/thumb-1.jpg" {
		t.Fatalf("expected trimmed thumbnail, got=%q", got)
	}
}

func TestStartTVRemoteSessionKeepsCurrentIndexBeyondHundredthItem(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 1, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-5 * time.Second)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	items := make([]models.TvRemoteSessionItem, 0, 120)
	for i := 0; i < 120; i++ {
		items = append(items, models.TvRemoteSessionItem{
			VideoID: uuid.New(),
			Title:   "条目",
			Type:    "short",
		})
	}

	repo := &fakeTVRemoteRepository{
		device: models.TvDeviceRecord{
			DeviceID:   "living-room",
			DeviceName: "客厅电视",
			Platform:   tvRemotePlatform,
			UserID:     &userID,
			LastSeenAt: &lastSeenAt,
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	session, err := svc.StartTVRemoteSessionAt(context.Background(), userID, "living-room", items, 119, now)
	if err != nil {
		t.Fatalf("StartTVRemoteSessionAt returned err=%v", err)
	}
	if session.CurrentIndex != 119 {
		t.Fatalf("expected current index 119, got=%d", session.CurrentIndex)
	}
	if len(session.Items) != 120 {
		t.Fatalf("expected all snapshot items to be kept, got=%d", len(session.Items))
	}
	if session.CurrentItem == nil || session.CurrentItem.VideoID != items[119].VideoID {
		t.Fatalf("expected current item to remain the 120th entry, got=%#v", session.CurrentItem)
	}
}

func TestGetCurrentTVRemoteSessionForDeviceTouchesDeviceSeen(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 3, 14, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-5 * time.Second)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := &fakeTVRemoteRepository{
		device: models.TvDeviceRecord{
			DeviceID:   "living-room",
			DeviceName: "客厅电视",
			Platform:   tvRemotePlatform,
			UserID:     &userID,
			LastSeenAt: &lastSeenAt,
		},
		getActiveErr: pgx.ErrNoRows,
	}
	svc := &AppService{tvRemoteRepo: repo}

	session, err := svc.GetCurrentTVRemoteSessionForDeviceAt(context.Background(), userID, "living-room", "", now)
	if err != nil {
		t.Fatalf("GetCurrentTVRemoteSessionForDeviceAt returned err=%v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session, got=%#v", session)
	}
	if repo.touchedDeviceID != "living-room" {
		t.Fatalf("expected touch to record device id, got=%q", repo.touchedDeviceID)
	}
}

func TestGetCurrentTVRemoteSessionForDeviceFallsBackToLegacyDeviceID(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 1, 30, 0, 0, time.UTC)
	lastSeenAt := now.Add(-5 * time.Second)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	videoID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	repo := &fakeTVRemoteRepository{
		device: models.TvDeviceRecord{
			DeviceID:   "legacy-device",
			DeviceName: "客厅电视",
			Platform:   tvRemotePlatform,
			UserID:     &userID,
			LastSeenAt: &lastSeenAt,
		},
		session: models.TvRemoteSession{
			ID:           sessionID,
			UserID:       userID,
			DeviceID:     "legacy-device",
			DeviceName:   "客厅电视",
			Platform:     tvRemotePlatform,
			Status:       tvRemoteSessionStatusActive,
			Items:        []models.TvRemoteSessionItem{{VideoID: videoID, Title: "条目 1"}},
			CurrentIndex: 0,
			CurrentVideoID: func() *uuid.UUID {
				value := videoID
				return &value
			}(),
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	session, err := svc.GetCurrentTVRemoteSessionForDeviceAt(context.Background(), userID, "new-device", "legacy-device", now)
	if err != nil {
		t.Fatalf("GetCurrentTVRemoteSessionForDeviceAt returned err=%v", err)
	}
	if session == nil || session.DeviceID != "legacy-device" {
		t.Fatalf("expected legacy device session, got=%#v", session)
	}
	if repo.touchedDeviceID != "legacy-device" {
		t.Fatalf("expected legacy device id to be touched, got=%q", repo.touchedDeviceID)
	}
}

func TestGetCurrentTVRemoteSessionForDeviceChecksLegacyAfterCurrentDeviceHasNoActiveSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 1, 40, 0, 0, time.UTC)
	lastSeenAt := now.Add(-5 * time.Second)
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	videoID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	repo := &fakeTVRemoteRepository{
		devicesByID: map[string]models.TvDeviceRecord{
			"new-device": {
				DeviceID:   "new-device",
				DeviceName: "客厅电视",
				Platform:   tvRemotePlatform,
				UserID:     &userID,
				LastSeenAt: &lastSeenAt,
			},
			"legacy-device": {
				DeviceID:   "legacy-device",
				DeviceName: "客厅电视",
				Platform:   tvRemotePlatform,
				UserID:     &userID,
				LastSeenAt: &lastSeenAt,
			},
		},
		sessionsByDevice: map[string]models.TvRemoteSession{
			"legacy-device": {
				ID:           sessionID,
				UserID:       userID,
				DeviceID:     "legacy-device",
				DeviceName:   "客厅电视",
				Platform:     tvRemotePlatform,
				Status:       tvRemoteSessionStatusActive,
				Items:        []models.TvRemoteSessionItem{{VideoID: videoID, Title: "条目 1"}},
				CurrentIndex: 0,
				CurrentVideoID: func() *uuid.UUID {
					value := videoID
					return &value
				}(),
			},
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	session, err := svc.GetCurrentTVRemoteSessionForDeviceAt(context.Background(), userID, "new-device", "legacy-device", now)
	if err != nil {
		t.Fatalf("GetCurrentTVRemoteSessionForDeviceAt returned err=%v", err)
	}
	if session == nil || session.DeviceID != "legacy-device" {
		t.Fatalf("expected legacy device session, got=%#v", session)
	}
	if got := repo.touchedDeviceIDs; len(got) != 2 || got[0] != "new-device" || got[1] != "legacy-device" {
		t.Fatalf("expected touch order [new-device legacy-device], got=%v", got)
	}
}

func TestUpdateTVRemoteSessionCurrentIndexRejectsOutOfRange(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	videoID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	repo := &fakeTVRemoteRepository{
		session: models.TvRemoteSession{
			ID:           sessionID,
			UserID:       userID,
			DeviceID:     "living-room",
			DeviceName:   "客厅电视",
			Platform:     tvRemotePlatform,
			Status:       tvRemoteSessionStatusActive,
			Items:        []models.TvRemoteSessionItem{{VideoID: videoID, Title: "条目 1"}},
			CurrentIndex: 0,
			CurrentVideoID: func() *uuid.UUID {
				value := videoID
				return &value
			}(),
		},
	}
	svc := &AppService{tvRemoteRepo: repo}

	_, err := svc.UpdateTVRemoteSessionCurrentIndex(context.Background(), userID, sessionID, 3)
	if !errors.Is(err, ErrTVRemoteIndexOutOfRange) {
		t.Fatalf("expected ErrTVRemoteIndexOutOfRange, got=%v", err)
	}
}

func TestStepTVRemoteSessionIndexStopsAtBoundaries(t *testing.T) {
	t.Parallel()

	if got := stepTVRemoteSessionIndex(0, -1, 3); got != 0 {
		t.Fatalf("expected lower boundary no-op, got=%d", got)
	}
	if got := stepTVRemoteSessionIndex(2, 1, 3); got != 2 {
		t.Fatalf("expected upper boundary no-op, got=%d", got)
	}
	if got := stepTVRemoteSessionIndex(1, 1, 3); got != 2 {
		t.Fatalf("expected step forward to 2, got=%d", got)
	}
}

type fakeTVRemoteRepository struct {
	device           models.TvDeviceRecord
	session          models.TvRemoteSession
	created          *models.TvRemoteSession
	touchedDeviceID  string
	touchedDeviceIDs []string
	devicesByID      map[string]models.TvDeviceRecord
	sessionsByDevice map[string]models.TvRemoteSession
	getActiveErr     error
}

func (r *fakeTVRemoteRepository) ListUserTVDevices(context.Context, uuid.UUID, string) ([]models.TvDeviceRecord, error) {
	if len(r.devicesByID) > 0 {
		items := make([]models.TvDeviceRecord, 0, len(r.devicesByID))
		for _, item := range r.devicesByID {
			items = append(items, item)
		}
		return items, nil
	}
	return []models.TvDeviceRecord{r.device}, nil
}

func (r *fakeTVRemoteRepository) GetTVDeviceByDeviceIDAndUser(_ context.Context, _ uuid.UUID, deviceID string, _ string) (models.TvDeviceRecord, error) {
	if len(r.devicesByID) > 0 {
		item, ok := r.devicesByID[deviceID]
		if !ok {
			return models.TvDeviceRecord{}, pgx.ErrNoRows
		}
		return item, nil
	}
	if r.device.DeviceID == "" {
		return models.TvDeviceRecord{}, pgx.ErrNoRows
	}
	return r.device, nil
}

func (r *fakeTVRemoteRepository) TouchTVDeviceSeen(_ context.Context, _ uuid.UUID, deviceID string, _ string, seenAt time.Time) error {
	if len(r.devicesByID) > 0 {
		item, ok := r.devicesByID[deviceID]
		if !ok {
			return pgx.ErrNoRows
		}
		item.LastSeenAt = &seenAt
		r.devicesByID[deviceID] = item
		r.touchedDeviceID = deviceID
		r.touchedDeviceIDs = append(r.touchedDeviceIDs, deviceID)
		return nil
	}
	if r.device.DeviceID != "" && r.device.DeviceID != deviceID {
		return pgx.ErrNoRows
	}
	r.touchedDeviceID = deviceID
	r.touchedDeviceIDs = append(r.touchedDeviceIDs, deviceID)
	r.device.LastSeenAt = &seenAt
	return nil
}

func (r *fakeTVRemoteRepository) CreateOrReplaceTVRemoteSession(_ context.Context, session models.TvRemoteSession, _ time.Time) error {
	cloned := session
	r.created = &cloned
	r.session = cloned
	r.session.DeviceName = r.device.DeviceName
	return nil
}

func (r *fakeTVRemoteRepository) GetTVRemoteSessionByID(_ context.Context, sessionID uuid.UUID) (models.TvRemoteSession, error) {
	if r.session.ID != sessionID {
		return models.TvRemoteSession{}, pgx.ErrNoRows
	}
	return r.session, nil
}

func (r *fakeTVRemoteRepository) GetActiveTVRemoteSessionForDevice(_ context.Context, _ uuid.UUID, deviceID string, _ string) (models.TvRemoteSession, error) {
	if r.getActiveErr != nil {
		return models.TvRemoteSession{}, r.getActiveErr
	}
	if len(r.sessionsByDevice) > 0 {
		session, ok := r.sessionsByDevice[deviceID]
		if !ok || session.ID == uuid.Nil {
			return models.TvRemoteSession{}, pgx.ErrNoRows
		}
		return session, nil
	}
	if r.session.ID == uuid.Nil {
		return models.TvRemoteSession{}, pgx.ErrNoRows
	}
	return r.session, nil
}

func (r *fakeTVRemoteRepository) UpdateTVRemoteSessionCurrentIndex(_ context.Context, sessionID, _ uuid.UUID, currentIndex int, currentVideoID uuid.UUID, _ time.Time) error {
	if r.session.ID != sessionID {
		return pgx.ErrNoRows
	}
	r.session.CurrentIndex = currentIndex
	r.session.CurrentVideoID = &currentVideoID
	return nil
}

func (r *fakeTVRemoteRepository) EndTVRemoteSession(_ context.Context, sessionID, _ uuid.UUID, endedReason string, endedAt time.Time) error {
	if r.session.ID != sessionID {
		return pgx.ErrNoRows
	}
	r.session.Status = tvRemoteSessionStatusEnded
	r.session.EndedReason = endedReason
	r.session.EndedAt = &endedAt
	return nil
}

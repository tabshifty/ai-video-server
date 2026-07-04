package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/repository"
)

const (
	tvRemotePlatform            = "android_tv"
	tvRemoteSessionStatusActive = "active"
	tvRemoteSessionStatusEnded  = "ended"
	tvRemoteDeviceSeenTTL       = 30 * time.Second
)

var (
	ErrTVRemoteItemsRequired   = errors.New("投放列表不能为空")
	ErrTVRemoteIndexOutOfRange = errors.New("当前视频位置无效")
	ErrTVRemoteDeviceOffline   = errors.New("目标电视当前不可接收投放")
	ErrTVRemoteSessionInactive = errors.New("投放会话已结束")
)

type tvRemoteRepository interface {
	ListUserTVDevices(ctx context.Context, userID uuid.UUID, platform string) ([]models.TvDeviceRecord, error)
	GetTVDeviceByDeviceIDAndUser(ctx context.Context, userID uuid.UUID, deviceID string, platform string) (models.TvDeviceRecord, error)
	TouchTVDeviceSeen(ctx context.Context, userID uuid.UUID, deviceID string, platform string, seenAt time.Time) error
	CreateOrReplaceTVRemoteSession(ctx context.Context, session models.TvRemoteSession, now time.Time) error
	GetTVRemoteSessionByID(ctx context.Context, sessionID uuid.UUID) (models.TvRemoteSession, error)
	GetActiveTVRemoteSessionForDevice(ctx context.Context, userID uuid.UUID, deviceID string, platform string) (models.TvRemoteSession, error)
	UpdateTVRemoteSessionState(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, items []models.TvRemoteSessionItem, currentIndex int, currentVideoID uuid.UUID, searchContext *models.TvRemoteSearchContext, now time.Time) error
	EndTVRemoteSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, endedReason string, endedAt time.Time) error
	SearchVideos(ctx context.Context, q, typ string, limit, offset int) ([]models.VideoListItem, int, error)
}

func (s *AppService) ListTVDevices(ctx context.Context, userID uuid.UUID) (models.TvDeviceListPayload, error) {
	items, err := s.tvRemoteRepo.ListUserTVDevices(ctx, userID, tvRemotePlatform)
	if err != nil {
		return models.TvDeviceListPayload{}, err
	}
	now := time.Now().UTC()
	for i := range items {
		items[i].IsOnline = isTVDeviceReachable(items[i].LastSeenAt, now)
	}
	return models.TvDeviceListPayload{Items: items}, nil
}

func (s *AppService) StartTVRemoteSession(
	ctx context.Context,
	userID uuid.UUID,
	deviceID string,
	items []models.TvRemoteSessionItem,
	currentIndex int,
	searchContext *models.TvRemoteSearchContext,
) (models.TvRemoteSession, error) {
	return s.StartTVRemoteSessionAt(ctx, userID, deviceID, items, currentIndex, searchContext, time.Now().UTC())
}

func (s *AppService) StartTVRemoteSessionAt(
	ctx context.Context,
	userID uuid.UUID,
	deviceID string,
	items []models.TvRemoteSessionItem,
	currentIndex int,
	searchContext *models.TvRemoteSearchContext,
	now time.Time,
) (models.TvRemoteSession, error) {
	if s.tvRemoteRepo == nil {
		return models.TvRemoteSession{}, fmt.Errorf("tv remote repository is not configured")
	}
	normalizedDeviceID := strings.TrimSpace(deviceID)
	normalizedItems, err := normalizeTVRemoteSessionItems(items)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	if currentIndex < 0 || currentIndex >= len(normalizedItems) {
		return models.TvRemoteSession{}, ErrTVRemoteIndexOutOfRange
	}
	device, err := s.tvRemoteRepo.GetTVDeviceByDeviceIDAndUser(ctx, userID, normalizedDeviceID, tvRemotePlatform)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	if !isTVDeviceReachable(device.LastSeenAt, now) {
		return models.TvRemoteSession{}, ErrTVRemoteDeviceOffline
	}
	currentVideoID := normalizedItems[currentIndex].VideoID
	session := models.TvRemoteSession{
		ID:             uuid.New(),
		UserID:         userID,
		DeviceID:       normalizedDeviceID,
		Platform:       tvRemotePlatform,
		Status:         tvRemoteSessionStatusActive,
		Items:          normalizedItems,
		SearchContext:  normalizeTVRemoteSearchContext(searchContext, len(normalizedItems)),
		CurrentIndex:   currentIndex,
		CurrentVideoID: &currentVideoID,
	}
	if err := s.tvRemoteRepo.CreateOrReplaceTVRemoteSession(ctx, session, now); err != nil {
		return models.TvRemoteSession{}, err
	}
	created, err := s.tvRemoteRepo.GetTVRemoteSessionByID(ctx, session.ID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	created.DeviceName = device.DeviceName
	return decorateTVRemoteSession(created), nil
}

func (s *AppService) GetTVRemoteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (models.TvRemoteSession, error) {
	session, err := s.getTVRemoteSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	return decorateTVRemoteSession(session), nil
}

func (s *AppService) StepTVRemoteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, delta int) (models.TvRemoteSession, error) {
	session, err := s.getTVRemoteSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	if session.Status != tvRemoteSessionStatusActive {
		return decorateTVRemoteSession(session), ErrTVRemoteSessionInactive
	}
	if delta > 0 {
		session, err = s.ensureTVRemoteSessionHasNextItem(ctx, userID, session)
		if err != nil {
			return models.TvRemoteSession{}, err
		}
	}
	nextIndex := stepTVRemoteSessionIndex(session.CurrentIndex, delta, len(session.Items))
	if nextIndex == session.CurrentIndex {
		return decorateTVRemoteSession(session), nil
	}
	return s.UpdateTVRemoteSessionCurrentIndex(ctx, userID, sessionID, nextIndex)
}

func (s *AppService) GetCurrentTVRemoteSessionForDevice(
	ctx context.Context,
	userID uuid.UUID,
	deviceID string,
	legacyDeviceID string,
) (*models.TvRemoteSession, error) {
	return s.GetCurrentTVRemoteSessionForDeviceAt(ctx, userID, deviceID, legacyDeviceID, time.Now().UTC())
}

func (s *AppService) GetCurrentTVRemoteSessionForDeviceAt(
	ctx context.Context,
	userID uuid.UUID,
	deviceID string,
	legacyDeviceID string,
	now time.Time,
) (*models.TvRemoteSession, error) {
	if s.tvRemoteRepo == nil {
		return nil, fmt.Errorf("tv remote repository is not configured")
	}
	candidateIDs := buildTVRemoteDeviceCandidates(deviceID, legacyDeviceID)
	if len(candidateIDs) == 0 {
		return nil, fmt.Errorf("device id is required")
	}
	touchedAny := false
	for _, candidateID := range candidateIDs {
		if err := s.tvRemoteRepo.TouchTVDeviceSeen(ctx, userID, candidateID, tvRemotePlatform, now); err != nil {
			if repository.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		touchedAny = true
		session, err := s.tvRemoteRepo.GetActiveTVRemoteSessionForDevice(ctx, userID, candidateID, tvRemotePlatform)
		if err != nil {
			if repository.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		decorated := decorateTVRemoteSession(session)
		return &decorated, nil
	}
	if touchedAny {
		return nil, nil
	}
	return nil, pgx.ErrNoRows
}

func (s *AppService) UpdateTVRemoteSessionCurrentIndex(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	currentIndex int,
) (models.TvRemoteSession, error) {
	session, err := s.getTVRemoteSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	if session.Status != tvRemoteSessionStatusActive {
		return decorateTVRemoteSession(session), ErrTVRemoteSessionInactive
	}
	if currentIndex < 0 || currentIndex >= len(session.Items) {
		return models.TvRemoteSession{}, ErrTVRemoteIndexOutOfRange
	}
	currentVideoID := session.Items[currentIndex].VideoID
	if currentIndex == session.CurrentIndex && session.CurrentVideoID != nil && *session.CurrentVideoID == currentVideoID {
		return decorateTVRemoteSession(session), nil
	}
	if err := s.tvRemoteRepo.UpdateTVRemoteSessionState(
		ctx,
		sessionID,
		userID,
		session.Items,
		currentIndex,
		currentVideoID,
		session.SearchContext,
		time.Now().UTC(),
	); err != nil {
		return models.TvRemoteSession{}, err
	}
	updated, err := s.getTVRemoteSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	return decorateTVRemoteSession(updated), nil
}

func (s *AppService) EndTVRemoteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, endedReason string) error {
	session, err := s.getTVRemoteSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return err
	}
	if session.Status != tvRemoteSessionStatusActive {
		return nil
	}
	return s.tvRemoteRepo.EndTVRemoteSession(ctx, sessionID, userID, normalizeTVRemoteEndedReason(endedReason), time.Now().UTC())
}

func (s *AppService) getTVRemoteSessionForUser(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (models.TvRemoteSession, error) {
	if s.tvRemoteRepo == nil {
		return models.TvRemoteSession{}, fmt.Errorf("tv remote repository is not configured")
	}
	session, err := s.tvRemoteRepo.GetTVRemoteSessionByID(ctx, sessionID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	if session.UserID != userID {
		return models.TvRemoteSession{}, pgx.ErrNoRows
	}
	return session, nil
}

func normalizeTVRemoteSessionItems(items []models.TvRemoteSessionItem) ([]models.TvRemoteSessionItem, error) {
	normalized := make([]models.TvRemoteSessionItem, 0, len(items))
	for _, item := range items {
		if item.VideoID == uuid.Nil {
			return nil, ErrTVRemoteItemsRequired
		}
		normalized = append(normalized, models.TvRemoteSessionItem{
			VideoID:       item.VideoID,
			Title:         strings.TrimSpace(item.Title),
			ThumbnailPath: strings.TrimSpace(item.ThumbnailPath),
			Duration:      item.Duration,
			Type:          strings.TrimSpace(item.Type),
		})
	}
	if len(normalized) == 0 {
		return nil, ErrTVRemoteItemsRequired
	}
	return normalized, nil
}

func normalizeTVRemoteSearchContext(searchContext *models.TvRemoteSearchContext, loadedItemCount int) *models.TvRemoteSearchContext {
	if searchContext == nil {
		return nil
	}
	query := strings.TrimSpace(searchContext.Query)
	if query == "" {
		return nil
	}
	typ := strings.ToLower(strings.TrimSpace(searchContext.Type))
	if typ == "" {
		typ = "short"
	}
	pageSize := searchContext.PageSize
	if pageSize <= 0 {
		pageSize = loadedItemCount
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	page := searchContext.Page
	if page < 1 {
		page = 1
		if loadedItemCount > 0 {
			page = int(math.Ceil(float64(loadedItemCount) / float64(pageSize)))
		}
	}
	totalCount := searchContext.TotalCount
	if totalCount < loadedItemCount {
		totalCount = loadedItemCount
	}
	return &models.TvRemoteSearchContext{
		Query:      query,
		Type:       typ,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
}

func decorateTVRemoteSession(session models.TvRemoteSession) models.TvRemoteSession {
	session.CurrentItem = currentTVRemoteSessionItem(session)
	session.HasPrevious = session.CurrentIndex > 0
	session.HasNext = (session.CurrentIndex >= 0 && session.CurrentIndex < len(session.Items)-1) || tvRemoteSessionSearchHasMore(session)
	return session
}

func currentTVRemoteSessionItem(session models.TvRemoteSession) *models.TvRemoteSessionItem {
	if session.CurrentIndex < 0 || session.CurrentIndex >= len(session.Items) {
		return nil
	}
	item := session.Items[session.CurrentIndex]
	return &item
}

func stepTVRemoteSessionIndex(currentIndex int, delta int, length int) int {
	if length <= 0 {
		return 0
	}
	next := currentIndex + delta
	if next < 0 {
		return 0
	}
	if next >= length {
		return length - 1
	}
	return next
}

func tvRemoteSessionSearchHasMore(session models.TvRemoteSession) bool {
	if session.SearchContext == nil {
		return false
	}
	if session.SearchContext.TotalCount <= 0 {
		return false
	}
	return len(session.Items) < session.SearchContext.TotalCount
}

func buildTVRemoteSessionItemsFromSearchResults(items []models.VideoListItem) []models.TvRemoteSessionItem {
	next := make([]models.TvRemoteSessionItem, 0, len(items))
	for _, item := range items {
		if item.ID == uuid.Nil {
			continue
		}
		next = append(next, models.TvRemoteSessionItem{
			VideoID:       item.ID,
			Title:         strings.TrimSpace(item.Title),
			ThumbnailPath: strings.TrimSpace(item.ThumbnailPath),
			Duration:      item.Duration,
			Type:          strings.TrimSpace(item.Type),
		})
	}
	return next
}

func mergeTVRemoteSessionItems(existing []models.TvRemoteSessionItem, incoming []models.TvRemoteSessionItem) []models.TvRemoteSessionItem {
	merged := make([]models.TvRemoteSessionItem, 0, len(existing)+len(incoming))
	seen := make(map[uuid.UUID]struct{}, len(existing)+len(incoming))
	for _, item := range existing {
		if item.VideoID == uuid.Nil {
			continue
		}
		if _, ok := seen[item.VideoID]; ok {
			continue
		}
		seen[item.VideoID] = struct{}{}
		merged = append(merged, item)
	}
	for _, item := range incoming {
		if item.VideoID == uuid.Nil {
			continue
		}
		if _, ok := seen[item.VideoID]; ok {
			continue
		}
		seen[item.VideoID] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}

func (s *AppService) ensureTVRemoteSessionHasNextItem(
	ctx context.Context,
	userID uuid.UUID,
	session models.TvRemoteSession,
) (models.TvRemoteSession, error) {
	if session.CurrentIndex < len(session.Items)-1 || !tvRemoteSessionSearchHasMore(session) || session.SearchContext == nil {
		return session, nil
	}
	searchContext := *session.SearchContext
	prevLen := len(session.Items)
	nextPage := searchContext.Page + 1
	if nextPage < 1 {
		nextPage = 1
	}
	offset := (nextPage - 1) * searchContext.PageSize
	items, totalCount, err := s.tvRemoteRepo.SearchVideos(ctx, searchContext.Query, searchContext.Type, searchContext.PageSize, offset)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	searchContext.Page = nextPage
	if totalCount > 0 {
		searchContext.TotalCount = totalCount
	}
	mergedItems := mergeTVRemoteSessionItems(session.Items, buildTVRemoteSessionItemsFromSearchResults(items))
	session.Items = mergedItems
	session.SearchContext = &searchContext
	currentVideoID := uuid.Nil
	if currentItem := currentTVRemoteSessionItem(session); currentItem != nil {
		currentVideoID = currentItem.VideoID
	}
	if currentVideoID == uuid.Nil {
		return session, nil
	}
	if err := s.tvRemoteRepo.UpdateTVRemoteSessionState(
		ctx,
		session.ID,
		userID,
		session.Items,
		session.CurrentIndex,
		currentVideoID,
		session.SearchContext,
		time.Now().UTC(),
	); err != nil {
		return models.TvRemoteSession{}, err
	}
	if len(mergedItems) == prevLen && !tvRemoteSessionSearchHasMore(session) {
		return decorateTVRemoteSession(session), nil
	}
	updated, err := s.getTVRemoteSessionForUser(ctx, userID, session.ID)
	if err != nil {
		return models.TvRemoteSession{}, err
	}
	return updated, nil
}

func isTVDeviceReachable(lastSeenAt *time.Time, now time.Time) bool {
	if lastSeenAt == nil {
		return false
	}
	return now.Sub(lastSeenAt.UTC()) <= tvRemoteDeviceSeenTTL
}

func normalizeTVRemoteEndedReason(reason string) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return "tv_back"
	}
	if len(trimmed) > 32 {
		return trimmed[:32]
	}
	return trimmed
}

func buildTVRemoteDeviceCandidates(deviceID string, legacyDeviceID string) []string {
	seen := make(map[string]struct{}, 2)
	items := make([]string, 0, 2)
	for _, raw := range []string{deviceID, legacyDeviceID} {
		normalized := strings.TrimSpace(raw)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

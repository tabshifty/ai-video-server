package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

const (
	tvAuthSessionTTL         = 5 * time.Minute
	tvAuthPollIntervalSecond = 5
)

type tvAuthRepository interface {
	CreateTVAuthSession(ctx context.Context, session models.TvAuthSession) error
	GetTVAuthSession(ctx context.Context, sessionID uuid.UUID) (models.TvAuthSession, error)
	UpdateTVAuthSessionExpired(ctx context.Context, sessionID uuid.UUID) error
	ApproveTVAuthSession(ctx context.Context, sessionID uuid.UUID, user models.User, tokens models.AuthTokens, approvedAt time.Time) error
	DenyTVAuthSession(ctx context.Context, sessionID uuid.UUID) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error)
	ListActiveTVSeriesSummaries(ctx context.Context, limit, offset int) ([]models.TvSeriesSummaryDto, int, error)
	SearchTVSeriesSummaries(ctx context.Context, q string, limit, offset int) ([]models.TvSeriesSummaryDto, int, error)
	SearchVideos(ctx context.Context, q, typ string, limit, offset int) ([]models.VideoListItem, int, error)
	GetTVContinueWatching(ctx context.Context, userID uuid.UUID) (*models.TvContinueWatchingDto, error)
	ContinueWatching(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.HistoryItem, int, error)
}

func newTVAuthSession(deviceID, deviceName string, now time.Time) (models.TvAuthSession, error) {
	pairCode, err := generateTVPairCode()
	if err != nil {
		return models.TvAuthSession{}, err
	}
	return models.TvAuthSession{
		ID:         uuid.New(),
		PairCode:   pairCode,
		DeviceID:   strings.TrimSpace(deviceID),
		DeviceName: strings.TrimSpace(deviceName),
		Platform:   "android_tv",
		Status:     "pending",
		ExpiresAt:  now.Add(tvAuthSessionTTL),
	}, nil
}

func generateTVPairCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate tv pair code: %w", err)
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}

func buildTVQRContent(serverBaseURL string, session models.TvAuthSession) string {
	values := url.Values{}
	values.Set("session_id", session.ID.String())
	values.Set("pair_code", session.PairCode)
	values.Set("device_name", session.DeviceName)
	if strings.TrimSpace(serverBaseURL) != "" {
		values.Set("server", strings.TrimSpace(serverBaseURL))
	}
	return "cheevideos://tv-auth?" + values.Encode()
}

func (s *AppService) CreateTVAuthSession(
	ctx context.Context,
	deviceID string,
	deviceName string,
	serverBaseURL string,
) (models.TvAuthSessionCreateResult, error) {
	session, err := newTVAuthSession(deviceID, deviceName, time.Now())
	if err != nil {
		return models.TvAuthSessionCreateResult{}, err
	}
	if session.DeviceID == "" || session.DeviceName == "" {
		return models.TvAuthSessionCreateResult{}, fmt.Errorf("device id and name are required")
	}
	if err := s.repo.CreateTVAuthSession(ctx, session); err != nil {
		return models.TvAuthSessionCreateResult{}, err
	}
	return models.TvAuthSessionCreateResult{
		SessionID:           session.ID,
		PairCode:            session.PairCode,
		QRContent:           buildTVQRContent(serverBaseURL, session),
		ExpiresAt:           session.ExpiresAt,
		PollIntervalSeconds: tvAuthPollIntervalSecond,
	}, nil
}

func (s *AppService) PollTVAuthSession(ctx context.Context, sessionID uuid.UUID) (models.TvAuthSessionPollResult, error) {
	session, err := s.repo.GetTVAuthSession(ctx, sessionID)
	if err != nil {
		return models.TvAuthSessionPollResult{}, err
	}
	if session.Status == "pending" && time.Now().After(session.ExpiresAt) {
		if err := s.repo.UpdateTVAuthSessionExpired(ctx, sessionID); err != nil {
			return models.TvAuthSessionPollResult{}, err
		}
		session.Status = "expired"
	}
	result := models.TvAuthSessionPollResult{
		SessionID:  session.ID,
		Status:     session.Status,
		ExpiresAt:  session.ExpiresAt,
		DeviceName: session.DeviceName,
		PairCode:   session.PairCode,
	}
	if session.Status == "approved" && session.UserID != nil {
		result.AccessToken = session.AccessToken
		result.RefreshToken = session.RefreshToken
		result.User = &models.TvAuthUserBrief{
			UserID:   *session.UserID,
			Username: session.ApprovedUsername,
			Role:     session.ApprovedRole,
		}
	}
	return result, nil
}

func (s *AppService) ApproveTVAuthSession(
	ctx context.Context,
	sessionID uuid.UUID,
	user models.User,
	tokens models.AuthTokens,
) error {
	return s.repo.ApproveTVAuthSession(ctx, sessionID, user, tokens, time.Now())
}

func (s *AppService) DenyTVAuthSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.DenyTVAuthSession(ctx, sessionID)
}

func buildTVHomeVideoFromSeries(dto models.TvSeriesSummaryDto) models.TvHomeVideoDto {
	return models.TvHomeVideoDto{
		ID:          fmt.Sprintf("%d", dto.ID),
		Type:        "tv",
		Title:       dto.Title,
		Overview:    dto.Overview,
		PosterURL:   dto.PosterURL,
		BackdropURL: dto.BackdropURL,
	}
}

func buildTVHomeVideoFromListItem(item models.VideoListItem) models.TvHomeVideoDto {
	overview := ""
	if len(item.Metadata) > 0 {
		overview = ""
	}
	return models.TvHomeVideoDto{
		ID:          item.ID.String(),
		Type:        normalizeTVSearchType(item.Type),
		Title:       item.Title,
		Overview:    overview,
		PosterURL:   item.ThumbnailPath,
		BackdropURL: item.ThumbnailPath,
		VideoID:     item.ID.String(),
	}
}

func buildTVCatalogWallSeriesItems(items []models.TvSeriesSummaryDto) []models.TvCatalogWallItemDto {
	out := make([]models.TvCatalogWallItemDto, 0, len(items))
	for _, item := range items {
		out = append(out, models.TvCatalogWallItemDto{
			ID:          fmt.Sprintf("%d", item.ID),
			Type:        "tv",
			Title:       item.Title,
			Overview:    item.Overview,
			PosterURL:   item.PosterURL,
			BackdropURL: item.BackdropURL,
		})
	}
	return out
}

func buildTVCatalogWallVideoItems(items []models.VideoListItem, videoType string) []models.TvCatalogWallItemDto {
	out := make([]models.TvCatalogWallItemDto, 0, len(items))
	for _, item := range items {
		normalizedType := normalizeTVSearchType(item.Type)
		if normalizedType == "" {
			normalizedType = normalizeTVSearchType(videoType)
		}
		out = append(out, models.TvCatalogWallItemDto{
			ID:          item.ID.String(),
			Type:        normalizedType,
			Title:       item.Title,
			Overview:    "",
			PosterURL:   item.ThumbnailPath,
			BackdropURL: item.ThumbnailPath,
		})
	}
	return out
}

func buildTVCatalogWallPayload(
	page int,
	pageSize int,
	totalCount int,
	items []models.TvCatalogWallItemDto,
) models.PageResult[models.TvCatalogWallItemDto] {
	return models.PageResult[models.TvCatalogWallItemDto]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}
}

func normalizeTVSearchType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "episode":
		return "tv"
	case "movie":
		return "movie"
	case "av":
		return "av"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func buildTVSearchPayload(
	page int,
	pageSize int,
	series []models.TvSeriesSummaryDto,
	movies []models.VideoListItem,
	avs []models.VideoListItem,
	total int,
) models.TvSearchPayload {
	items := make([]models.TvSearchResultDto, 0, len(series)+len(movies))
	for _, item := range series {
		items = append(items, models.TvSearchResultDto{
			ID:          fmt.Sprintf("%d", item.ID),
			Type:        "tv",
			Title:       item.Title,
			Overview:    item.Overview,
			PosterURL:   item.PosterURL,
			BackdropURL: item.BackdropURL,
		})
	}
	for _, item := range movies {
		items = append(items, models.TvSearchResultDto{
			ID:          item.ID.String(),
			Type:        "movie",
			Title:       item.Title,
			PosterURL:   item.ThumbnailPath,
			BackdropURL: item.ThumbnailPath,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].Title < items[j].Title
		}
		return items[i].Type < items[j].Type
	})
	return models.TvSearchPayload{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}
}

func (s *AppService) TVSearch(ctx context.Context, q string, page, pageSize int) (models.TvSearchPayload, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := strings.TrimSpace(q)
	if query == "" {
		return models.TvSearchPayload{Items: []models.TvSearchResultDto{}, Page: page, PageSize: pageSize}, nil
	}

	series, totalSeries, err := s.repo.SearchTVSeriesSummaries(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return models.TvSearchPayload{}, err
	}
	movies, totalMovies, err := s.repo.SearchVideos(ctx, query, "movie", pageSize, (page-1)*pageSize)
	if err != nil {
		return models.TvSearchPayload{}, err
	}
	return buildTVSearchPayload(page, pageSize, series, movies, nil, totalSeries+totalMovies), nil
}

func (s *AppService) buildTVHomeContinueWatching(ctx context.Context, userID uuid.UUID) (*models.TvContinueWatchingDto, error) {
	continueWatching, err := s.repo.GetTVContinueWatching(ctx, userID)
	if err != nil {
		return nil, err
	}
	if continueWatching != nil {
		continueWatching.Type = "tv"
		return continueWatching, nil
	}

	items, _, err := s.repo.ContinueWatching(ctx, userID, 10, 0)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		switch normalizeTVSearchType(item.Type) {
		case "movie":
			return &models.TvContinueWatchingDto{
				Type:            normalizeTVSearchType(item.Type),
				SeriesTitle:     item.Title,
				VideoID:         item.VideoID.String(),
				PosterURL:       item.ThumbnailPath,
				BackdropURL:     item.ThumbnailPath,
				WatchSeconds:    item.WatchSeconds,
				DurationSeconds: item.Duration,
				ProgressPercent: int(item.Progress * 100),
			}, nil
		}
	}
	return nil, nil
}

func (s *AppService) TVHome(ctx context.Context, userID uuid.UUID, q string, page, pageSize int) (models.TvHomePayload, error) {
	if strings.TrimSpace(q) != "" {
		legacy, err := s.tvHomeLegacy(ctx, userID, q, page, pageSize)
		if err != nil {
			return models.TvHomePayload{}, err
		}
		return legacy, nil
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	series, _, err := s.repo.ListActiveTVSeriesSummaries(ctx, min(pageSize, 12), 0)
	if err != nil {
		return models.TvHomePayload{}, err
	}
	movies, _, err := s.repo.SearchVideos(ctx, "", "movie", min(pageSize, 12), 0)
	if err != nil {
		return models.TvHomePayload{}, err
	}
	continueWatching, err := s.buildTVHomeContinueWatching(ctx, userID)
	if err != nil {
		return models.TvHomePayload{}, err
	}

	payload := models.TvHomePayload{
		ContinueWatching: continueWatching,
		Sections:         buildTVSections(series, time.Now()),
		SearchResults:    []models.TvSeriesSummaryDto{},
		TvSeries:         make([]models.TvHomeVideoDto, 0, len(series)),
		Movies:           make([]models.TvHomeVideoDto, 0, len(movies)),
		AV:               []models.TvHomeVideoDto{},
		Page:             page,
		PageSize:         pageSize,
	}
	for _, item := range series {
		payload.TvSeries = append(payload.TvSeries, buildTVHomeVideoFromSeries(item))
	}
	for _, item := range movies {
		payload.Movies = append(payload.Movies, buildTVHomeVideoFromListItem(item))
	}
	return payload, nil
}

func (s *AppService) TVCatalogWall(
	ctx context.Context,
	kind string,
	page,
	pageSize int,
) (models.PageResult[models.TvCatalogWallItemDto], error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "recent", "tv":
		items, total, err := s.repo.ListActiveTVSeriesSummaries(ctx, pageSize, offset)
		if err != nil {
			return models.PageResult[models.TvCatalogWallItemDto]{}, err
		}
		return buildTVCatalogWallPayload(page, pageSize, total, buildTVCatalogWallSeriesItems(items)), nil
	case "binge":
		items, total, err := s.repo.ListBingeTVSeriesSummaries(ctx, pageSize, offset)
		if err != nil {
			return models.PageResult[models.TvCatalogWallItemDto]{}, err
		}
		return buildTVCatalogWallPayload(page, pageSize, total, buildTVCatalogWallSeriesItems(items)), nil
	case "classic":
		items, total, err := s.repo.ListClassicTVSeriesSummaries(ctx, pageSize, offset)
		if err != nil {
			return models.PageResult[models.TvCatalogWallItemDto]{}, err
		}
		return buildTVCatalogWallPayload(page, pageSize, total, buildTVCatalogWallSeriesItems(items)), nil
	case "movie":
		items, total, err := s.repo.SearchVideos(ctx, "", kind, pageSize, offset)
		if err != nil {
			return models.PageResult[models.TvCatalogWallItemDto]{}, err
		}
		return buildTVCatalogWallPayload(page, pageSize, total, buildTVCatalogWallVideoItems(items, kind)), nil
	default:
		return models.PageResult[models.TvCatalogWallItemDto]{}, fmt.Errorf("unsupported tv wall kind: %s", kind)
	}
}

func (s *AppService) tvHomeLegacy(ctx context.Context, userID uuid.UUID, q string, page, pageSize int) (models.TvHomePayload, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	items, _, err := s.repo.SearchTVSeriesSummaries(ctx, q, pageSize, offset)
	if err != nil {
		return models.TvHomePayload{}, err
	}
	return buildTVHomePayload(q, nil, items, page, pageSize), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var _ tvAuthRepository = (*repository.VideoRepository)(nil)

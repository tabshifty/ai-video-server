package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

// AppService handles app-facing APIs for detail/history/interaction/search/user.
type AppService struct {
	repo *repository.VideoRepository
}

func NewAppService(repo *repository.VideoRepository) *AppService {
	return &AppService{repo: repo}
}

func (s *AppService) VideoDetail(ctx context.Context, userID, videoID uuid.UUID) (models.VideoDetail, error) {
	return s.repo.GetVideoDetail(ctx, videoID, userID)
}

func (s *AppService) RecordHistory(ctx context.Context, userID, videoID uuid.UUID, watchSeconds int, completed bool) error {
	if watchSeconds < 0 {
		watchSeconds = 0
	}
	return s.repo.UpsertViewHistory(ctx, userID, videoID, watchSeconds, completed)
}

func (s *AppService) ContinueWatching(ctx context.Context, userID uuid.UUID, page, pageSize int) (models.PageResult[models.HistoryItem], error) {
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
	items, total, err := s.repo.ContinueWatching(ctx, userID, pageSize, offset)
	if err != nil {
		return models.PageResult[models.HistoryItem]{}, err
	}
	return models.PageResult[models.HistoryItem]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *AppService) DeleteHistory(ctx context.Context, userID, videoID uuid.UUID) error {
	return s.repo.DeleteHistory(ctx, userID, videoID)
}

func (s *AppService) ToggleInteraction(ctx context.Context, userID, videoID uuid.UUID, action string) (bool, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	return s.repo.ToggleAction(ctx, userID, videoID, action)
}

func (s *AppService) Search(ctx context.Context, q, typ string, page, pageSize int) (models.PageResult[models.VideoListItem], error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	typ = strings.ToLower(strings.TrimSpace(typ))
	if typ == "" {
		typ = "all"
	}
	offset := (page - 1) * pageSize
	items, total, err := s.repo.SearchVideos(ctx, q, typ, pageSize, offset)
	if err != nil {
		return models.PageResult[models.VideoListItem]{}, err
	}
	return models.PageResult[models.VideoListItem]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *AppService) Profile(ctx context.Context, userID uuid.UUID) (models.UserProfileView, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return models.UserProfileView{}, err
	}
	return models.UserProfileView{
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *AppService) UpdateProfile(ctx context.Context, userID uuid.UUID, oldPassword, newEmail, newPassword string) error {
	return s.repo.UpdateUserProfile(ctx, userID, oldPassword, newEmail, newPassword)
}

func (s *AppService) UploadedVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (models.PageResult[models.VideoListItem], error) {
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
	items, total, err := s.repo.GetUploadedVideos(ctx, userID, pageSize, offset)
	if err != nil {
		return models.PageResult[models.VideoListItem]{}, err
	}
	return models.PageResult[models.VideoListItem]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *AppService) LikedVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (models.PageResult[models.VideoListItem], error) {
	return s.actionVideos(ctx, userID, "like", page, pageSize)
}

func (s *AppService) FavoritedVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (models.PageResult[models.VideoListItem], error) {
	return s.actionVideos(ctx, userID, "favorite", page, pageSize)
}

func (s *AppService) actionVideos(ctx context.Context, userID uuid.UUID, action string, page, pageSize int) (models.PageResult[models.VideoListItem], error) {
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
	items, total, err := s.repo.GetActionVideos(ctx, userID, action, pageSize, offset)
	if err != nil {
		return models.PageResult[models.VideoListItem]{}, err
	}
	return models.PageResult[models.VideoListItem]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

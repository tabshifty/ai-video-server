package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

var (
	ErrHermesForumInvalidInput       = errors.New("invalid Hermes forum input")
	ErrHermesForumPostNotFound       = repository.ErrHermesForumPostNotFound
	ErrHermesForumInspectionConflict = repository.ErrHermesForumInspectionConflict
	ErrAdminForumSearchQueryTooLong  = errors.New("标题搜索词不能超过 200 个字符")
)

const adminForumSearchQueryMaxRunes = 200

type hermesForumRepository interface {
	ListAdminForumPosts(context.Context, int, int, string) ([]models.AdminForumPostListItem, int, error)
	DiscoverForumPosts(context.Context, models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error)
	CompleteForumPostInspection(context.Context, uuid.UUID, models.ForumPostInspectionInput, string) (models.ForumPostInspectionResult, error)
}

type HermesForumService struct {
	repo hermesForumRepository
}

func NewHermesForumService(repo hermesForumRepository) *HermesForumService {
	return &HermesForumService{repo: repo}
}

// ListAdmin returns the paginated read model used by the administrator forum resource page.
func (s *HermesForumService) ListAdmin(ctx context.Context, page, pageSize int, query string) ([]models.AdminForumPostListItem, int, error) {
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) > adminForumSearchQueryMaxRunes {
		return nil, 0, ErrAdminForumSearchQueryTooLong
	}
	return s.repo.ListAdminForumPosts(ctx, page, pageSize, query)
}

func (s *HermesForumService) Discover(ctx context.Context, input models.ForumPostDiscoverInput) ([]models.ForumPostDiscoverResult, error) {
	normalized, err := normalizeForumDiscoverInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.DiscoverForumPosts(ctx, normalized)
}

func normalizeForumDiscoverInput(input models.ForumPostDiscoverInput) (models.ForumPostDiscoverInput, error) {
	input.Mode = strings.TrimSpace(input.Mode)
	input.Source = strings.TrimSpace(input.Source)
	input.BoardKey = strings.TrimSpace(input.BoardKey)
	if input.Mode != models.ForumPostDiscoverModeNormal && input.Mode != models.ForumPostDiscoverModeDedupeOnly {
		return models.ForumPostDiscoverInput{}, invalidHermesForumInput("mode 非法")
	}
	if input.Source == "" || len(input.Source) > 64 || input.BoardKey == "" || len(input.BoardKey) > 128 {
		return models.ForumPostDiscoverInput{}, invalidHermesForumInput("source 或 board_key 非法")
	}
	if len(input.Posts) == 0 || len(input.Posts) > 100 {
		return models.ForumPostDiscoverInput{}, invalidHermesForumInput("posts 数量必须为 1 到 100")
	}
	if input.Mode == models.ForumPostDiscoverModeNormal && input.ObservedAt == nil {
		return models.ForumPostDiscoverInput{}, invalidHermesForumInput("normal 模式缺少 observed_at")
	}
	if input.Mode == models.ForumPostDiscoverModeDedupeOnly && input.ObservedAt != nil {
		return models.ForumPostDiscoverInput{}, invalidHermesForumInput("dedupe_only 模式不接受 observed_at")
	}

	seen := make(map[string]struct{}, len(input.Posts))
	posts := make([]models.ForumPostCandidate, len(input.Posts))
	for i, post := range input.Posts {
		post.TID = strings.TrimSpace(post.TID)
		post.Title = strings.TrimSpace(post.Title)
		post.URL = strings.TrimSpace(post.URL)
		if post.TID == "" || len(post.TID) > 128 {
			return models.ForumPostDiscoverInput{}, invalidHermesForumInput("posts[%d].tid 非法", i)
		}
		if _, exists := seen[post.TID]; exists {
			return models.ForumPostDiscoverInput{}, invalidHermesForumInput("批次内 tid 重复")
		}
		seen[post.TID] = struct{}{}
		if input.Mode == models.ForumPostDiscoverModeNormal {
			if post.Title == "" || len(post.Title) > 4096 || !isHTTPURL(post.URL) || len(post.URL) > 8192 {
				return models.ForumPostDiscoverInput{}, invalidHermesForumInput("posts[%d] 的标题或 URL 非法", i)
			}
		} else if post.Title != "" || post.URL != "" {
			return models.ForumPostDiscoverInput{}, invalidHermesForumInput("dedupe_only 模式只接受 tid")
		}
		posts[i] = post
	}
	input.Posts = posts
	return input, nil
}

func (s *HermesForumService) CompleteInspection(ctx context.Context, postID uuid.UUID, input models.ForumPostInspectionInput) (models.ForumPostInspectionResult, error) {
	normalized, err := normalizeForumInspectionInput(postID, input)
	if err != nil {
		return models.ForumPostInspectionResult{}, err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return models.ForumPostInspectionResult{}, fmt.Errorf("序列化检查结果: %w", err)
	}
	digest := sha256.Sum256(payload)
	return s.repo.CompleteForumPostInspection(ctx, postID, normalized, hex.EncodeToString(digest[:]))
}

func normalizeForumInspectionInput(postID uuid.UUID, input models.ForumPostInspectionInput) (models.ForumPostInspectionInput, error) {
	if postID == uuid.Nil {
		return models.ForumPostInspectionInput{}, invalidHermesForumInput("帖子 ID 非法")
	}
	input.Status = strings.TrimSpace(input.Status)
	input.FilterDecision = strings.TrimSpace(input.FilterDecision)
	input.FetchMethod = strings.TrimSpace(input.FetchMethod)
	input.ErrorSummary = strings.TrimSpace(input.ErrorSummary)
	if input.Status != models.ForumPostStatusInspected && input.Status != models.ForumPostStatusRestricted && input.Status != models.ForumPostStatusFailed {
		return models.ForumPostInspectionInput{}, invalidHermesForumInput("status 非法")
	}
	if input.FilterDecision != models.ForumPostFilterIncluded && input.FilterDecision != models.ForumPostFilterExcluded {
		return models.ForumPostInspectionInput{}, invalidHermesForumInput("filter_decision 非法")
	}
	if len(input.FetchMethod) > 64 || len(input.ErrorSummary) > 4096 || len(input.FilterReasons) > 100 || len(input.Attachments) > 100 || len(input.ED2KLinks) > 100 {
		return models.ForumPostInspectionInput{}, invalidHermesForumInput("检查结果字段过长或资源过多")
	}

	input.FilterReasons = normalizeStringSlice(input.FilterReasons)
	for i, reason := range input.FilterReasons {
		if reason == "" || len(reason) > 256 {
			return models.ForumPostInspectionInput{}, invalidHermesForumInput("filter_reasons[%d] 非法", i)
		}
	}
	input.Attachments = normalizeStringSlice(input.Attachments)
	for i, item := range input.Attachments {
		if len(item) > 8192 || !isHTTPURL(item) {
			return models.ForumPostInspectionInput{}, invalidHermesForumInput("attachments[%d] 非法", i)
		}
	}
	input.ED2KLinks = normalizeStringSlice(input.ED2KLinks)
	for i, item := range input.ED2KLinks {
		if len(item) > 8192 || !strings.HasPrefix(strings.ToLower(item), "ed2k://") {
			return models.ForumPostInspectionInput{}, invalidHermesForumInput("ed2k_links[%d] 非法", i)
		}
	}
	return input, nil
}

func normalizeStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = strings.TrimSpace(value)
	}
	return out
}

func isHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func invalidHermesForumInput(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrHermesForumInvalidInput, fmt.Sprintf(format, args...))
}

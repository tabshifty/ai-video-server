package services

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-server/internal/models"
)

func actorAvatarRoute(actorID uuid.UUID) string {
	return "/api/v1/actors/" + actorID.String() + "/avatar"
}

func actorAvatarStorageDir(storageRoot string, actorID uuid.UUID) string {
	return filepath.Join(storageRoot, "actors", actorID.String())
}

func (s *ScraperService) completeAVActorAvatars(ctx context.Context, videoID uuid.UUID) error {
	actors, err := s.repo.ListVideoActors(ctx, videoID)
	if err != nil {
		return err
	}
	for _, actor := range actors {
		if !shouldCompleteActorAvatar(actor) {
			continue
		}
		if err := s.completeActorAvatar(ctx, actor.ID, actor.Name); err != nil {
			continue
		}
	}
	return nil
}

func shouldCompleteActorAvatar(actor models.AdminVideoActor) bool {
	if actor.ID == uuid.Nil || !actor.Active {
		return false
	}
	return strings.TrimSpace(actor.AvatarURL) == ""
}

func (s *ScraperService) completeActorAvatar(ctx context.Context, actorID uuid.UUID, actorName string) error {
	candidate, source, err := s.findActorAvatarCandidate(ctx, actorName)
	if err != nil {
		return err
	}
	if strings.TrimSpace(candidate.AvatarURL) == "" {
		return nil
	}
	localURL, err := s.downloadActorAvatar(ctx, actorID, candidate.AvatarURL)
	if err != nil {
		return err
	}
	return s.repo.UpdateActorAvatar(ctx, actorID, localURL, source, strings.TrimSpace(candidate.ExternalID))
}

func (s *ScraperService) findActorAvatarCandidate(ctx context.Context, actorName string) (ActorScrapeCandidate, string, error) {
	if candidate, ok := s.previewActorAvatarCandidate(ctx, actorName, "javdb"); ok {
		return candidate, "scrape_av", nil
	}
	if candidate, ok := s.previewActorAvatarCandidate(ctx, actorName, "tmdb"); ok {
		return candidate, "scrape_tmdb", nil
	}
	return ActorScrapeCandidate{}, "", nil
}

func (s *ScraperService) previewActorAvatarCandidate(ctx context.Context, actorName, source string) (ActorScrapeCandidate, bool) {
	items, err := s.PreviewActorByName(ctx, actorName, source, actorPreviewLimitDefault)
	if err != nil {
		return ActorScrapeCandidate{}, false
	}
	return chooseActorAvatarCandidate(actorName, items)
}

func chooseActorAvatarCandidate(actorName string, items []ActorScrapeCandidate) (ActorScrapeCandidate, bool) {
	normalizedName := normalizeActorAvatarName(actorName)
	for _, item := range items {
		if strings.TrimSpace(item.AvatarURL) == "" {
			continue
		}
		if normalizeActorAvatarName(item.Name) == normalizedName {
			return item, true
		}
	}
	for _, item := range items {
		if strings.TrimSpace(item.AvatarURL) == "" {
			continue
		}
		return item, true
	}
	return ActorScrapeCandidate{}, false
}

func normalizeActorAvatarName(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func (s *ScraperService) downloadActorAvatar(ctx context.Context, actorID uuid.UUID, avatarURL string) (string, error) {
	avatarURL = strings.TrimSpace(avatarURL)
	if actorID == uuid.Nil || avatarURL == "" {
		return "", nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("create actor avatar request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download actor avatar failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("actor avatar status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("actor avatar content type must be image, got=%s", contentType)
	}

	outputDir := actorAvatarStorageDir(s.storageRoot, actorID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create actor avatar dir: %w", err)
	}
	if matches, globErr := filepath.Glob(filepath.Join(outputDir, "avatar.*")); globErr == nil {
		for _, match := range matches {
			_ = os.Remove(match)
		}
	}
	outputPath := filepath.Join(outputDir, "avatar"+actorAvatarExt(contentType, avatarURL))
	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create actor avatar file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("write actor avatar file: %w", err)
	}
	return actorAvatarRoute(actorID), nil
}

func actorAvatarExt(contentType, rawURL string) string {
	switch {
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err == nil {
		ext := strings.ToLower(filepath.Ext(parsed.Path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
			if ext == ".jpeg" {
				return ".jpg"
			}
			return ext
		}
	}
	exts, _ := mime.ExtensionsByType(contentType)
	for _, ext := range exts {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == ".jpeg" {
			return ".jpg"
		}
		if ext == ".jpg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
			return ext
		}
	}
	return ".jpg"
}

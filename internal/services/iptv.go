package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
)

var ErrIPTVSourceURLRequired = errors.New("iptv source_url required")

type IPTVRepository interface {
	GetIPTVPlaylist(ctx context.Context) (models.IPTVPlaylistMeta, []models.IPTVChannel, error)
	SaveIPTVSourceURL(ctx context.Context, sourceURL string) (models.IPTVPlaylistMeta, []models.IPTVChannel, error)
	ReplaceIPTVPlaylist(ctx context.Context, meta models.IPTVPlaylistMeta, channels []models.IPTVChannel) (models.IPTVPlaylistMeta, []models.IPTVChannel, error)
}

type IPTVService struct {
	repo   IPTVRepository
	client *http.Client
}

func NewIPTVService(repo IPTVRepository, client *http.Client) *IPTVService {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &IPTVService{repo: repo, client: client}
}

func (s *IPTVService) Status(ctx context.Context) (models.IPTVPlaylistStatus, error) {
	meta, channels, err := s.repo.GetIPTVPlaylist(ctx)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	return BuildIPTVPlaylistStatus(meta, channels), nil
}

func (s *IPTVService) Upload(ctx context.Context, filename string, r io.Reader) (models.IPTVPlaylistStatus, error) {
	name := strings.ToLower(strings.TrimSpace(filename))
	if !strings.HasSuffix(name, ".m3u") && !strings.HasSuffix(name, ".m3u8") {
		return models.IPTVPlaylistStatus{}, fmt.Errorf("仅支持 .m3u 或 .m3u8 文件")
	}
	channels, skipped := ParseM3UPlaylist(r)
	meta := models.IPTVPlaylistMeta{
		SkippedCount: skipped,
		UpdatedAt:    timePtr(time.Now().UTC()),
	}
	savedMeta, savedChannels, err := s.repo.ReplaceIPTVPlaylist(ctx, meta, channels)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	return BuildIPTVPlaylistStatus(savedMeta, savedChannels), nil
}

func (s *IPTVService) SaveSourceURL(ctx context.Context, sourceURL string) (models.IPTVPlaylistStatus, error) {
	normalized, err := normalizeHTTPURL(sourceURL)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	meta, channels, err := s.repo.SaveIPTVSourceURL(ctx, normalized)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	return BuildIPTVPlaylistStatus(meta, channels), nil
}

func (s *IPTVService) Refresh(ctx context.Context) (models.IPTVPlaylistStatus, error) {
	meta, _, err := s.repo.GetIPTVPlaylist(ctx)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	sourceURL := strings.TrimSpace(meta.SourceURL)
	if sourceURL == "" {
		return models.IPTVPlaylistStatus{}, ErrIPTVSourceURLRequired
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return models.IPTVPlaylistStatus{}, fmt.Errorf("创建 IPTV 刷新请求失败: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return models.IPTVPlaylistStatus{}, fmt.Errorf("拉取 IPTV 播放列表失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return models.IPTVPlaylistStatus{}, fmt.Errorf("拉取 IPTV 播放列表失败: HTTP %d", resp.StatusCode)
	}
	channels, skipped := ParseM3UPlaylist(resp.Body)
	meta.SkippedCount = skipped
	meta.UpdatedAt = timePtr(time.Now().UTC())
	savedMeta, savedChannels, err := s.repo.ReplaceIPTVPlaylist(ctx, meta, channels)
	if err != nil {
		return models.IPTVPlaylistStatus{}, err
	}
	return BuildIPTVPlaylistStatus(savedMeta, savedChannels), nil
}

func BuildIPTVPlaylistStatus(meta models.IPTVPlaylistMeta, channels []models.IPTVChannel) models.IPTVPlaylistStatus {
	if channels == nil {
		channels = []models.IPTVChannel{}
	}
	groups := make([]string, 0)
	seen := map[string]struct{}{}
	for _, channel := range channels {
		group := strings.TrimSpace(channel.Group)
		if group == "" {
			group = "未分组"
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		groups = append(groups, group)
	}
	return models.IPTVPlaylistStatus{
		SourceURL:    meta.SourceURL,
		UpdatedAt:    meta.UpdatedAt,
		ChannelCount: len(channels),
		SkippedCount: meta.SkippedCount,
		Groups:       groups,
		Channels:     channels,
	}
}

type pendingM3UEntry struct {
	name    string
	group   string
	logoURL string
	tvgID   string
}

var m3uAttrPattern = regexp.MustCompile(`([A-Za-z0-9_-]+)="([^"]*)"`)

func ParseM3UPlaylist(r io.Reader) ([]models.IPTVChannel, int) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	channels := make([]models.IPTVChannel, 0)
	skipped := 0
	var pending *pendingM3UEntry

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || line == "#EXTM3U" {
			continue
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			if pending != nil {
				skipped++
			}
			entry := parseEXTINF(line)
			pending = &entry
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if pending == nil {
			continue
		}
		if validHTTPURL(line) && strings.TrimSpace(pending.name) != "" && playableIPTVVideoSource(*pending, line) {
			sortOrder := len(channels)
			channels = append(channels, models.IPTVChannel{
				ID:        uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("%d:%s:%s", sortOrder, pending.name, line))).String(),
				Name:      pending.name,
				URL:       line,
				Group:     pending.group,
				LogoURL:   pending.logoURL,
				TVGID:     pending.tvgID,
				SortOrder: sortOrder,
			})
		} else {
			skipped++
		}
		pending = nil
	}
	if pending != nil {
		skipped++
	}
	return channels, skipped
}

func playableIPTVVideoSource(entry pendingM3UEntry, rawURL string) bool {
	group := strings.ToLower(strings.TrimSpace(entry.group))
	name := strings.ToLower(strings.TrimSpace(entry.name))
	if group == "audio" || group == "音频" || strings.Contains(group, "audio only") {
		return false
	}
	if strings.Contains(name, "音频") || strings.Contains(name, "audio only") {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return true
	}
	path := strings.ToLower(parsed.EscapedPath())
	if strings.Contains(path, "/audio/") || strings.Contains(path, "_audio/") {
		return false
	}
	for _, suffix := range []string{".mp3", ".aac", ".m4a", ".flac", ".wav", ".ogg", ".opus"} {
		if strings.HasSuffix(path, suffix) {
			return false
		}
	}
	return true
}

func parseEXTINF(line string) pendingM3UEntry {
	body := strings.TrimPrefix(line, "#EXTINF:")
	beforeName, name := splitM3UName(body)
	attrs := parseM3UAttrs(beforeName)
	return pendingM3UEntry{
		name:    strings.TrimSpace(name),
		group:   attrs["group-title"],
		logoURL: attrs["tvg-logo"],
		tvgID:   attrs["tvg-id"],
	}
}

func splitM3UName(s string) (string, string) {
	inQuote := false
	for i, r := range s {
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if r == ',' && !inQuote {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func parseM3UAttrs(s string) map[string]string {
	out := map[string]string{}
	matches := m3uAttrPattern.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		out[strings.ToLower(match[1])] = strings.TrimSpace(match[2])
	}
	return out
}

func normalizeHTTPURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if !validHTTPURL(value) {
		return "", fmt.Errorf("source_url 必须是 http 或 https 地址")
	}
	return value, nil
}

func validHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func timePtr(t time.Time) *time.Time {
	return &t
}

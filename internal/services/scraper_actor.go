package services

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	actorPreviewLimitDefault = 10
	actorPreviewLimitMax     = 20
)

var (
	javDBActorAnchorRe = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*/actors/[^"']+)["'][^>]*>(.*?)</a>`)
	htmlTagRe          = regexp.MustCompile(`(?is)<[^>]+>`)
	imgSrcRe           = regexp.MustCompile(`(?is)<img[^>]*src=["']([^"']+)["']`)
	titleAttrRe        = regexp.MustCompile(`(?is)\btitle=["']([^"']+)["']`)
)

type ActorScrapeCandidate struct {
	Source     string         `json:"source"`
	ExternalID string         `json:"external_id"`
	Name       string         `json:"name"`
	Aliases    []string       `json:"aliases"`
	Gender     string         `json:"gender"`
	Country    string         `json:"country"`
	BirthDate  string         `json:"birth_date"`
	AvatarURL  string         `json:"avatar_url"`
	Notes      string         `json:"notes"`
	Raw        map[string]any `json:"raw"`
}

func (s *ScraperService) PreviewActorByName(ctx context.Context, name, source string, limit int) ([]ActorScrapeCandidate, error) {
	keyword := strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
	if keyword == "" {
		return nil, fmt.Errorf("演员姓名不能为空")
	}

	src := strings.ToLower(strings.TrimSpace(source))
	if src == "" {
		src = "tmdb"
	}
	if src != "tmdb" && src != "javdb" {
		return nil, fmt.Errorf("source 仅支持 tmdb 或 javdb")
	}

	if limit <= 0 {
		limit = actorPreviewLimitDefault
	}
	if limit > actorPreviewLimitMax {
		limit = actorPreviewLimitMax
	}

	if src == "javdb" {
		return s.previewActorsJavDB(ctx, keyword, limit)
	}
	return s.previewActorsTMDB(ctx, keyword, limit)
}

func (s *ScraperService) previewActorsTMDB(ctx context.Context, keyword string, limit int) ([]ActorScrapeCandidate, error) {
	if strings.TrimSpace(s.apiKey) == "" {
		return nil, fmt.Errorf("TMDB_API_KEY 未配置")
	}

	query := url.Values{}
	query.Set("query", keyword)
	raw, err := s.getTMDBJSON(ctx, "/search/person", query, tmdbLangChinese)
	if err != nil {
		return nil, fmt.Errorf("TMDB 演员查询失败: %w", err)
	}

	rows, _ := raw["results"].([]any)
	out := make([]ActorScrapeCandidate, 0, limit)
	seen := map[int]struct{}{}
	for _, row := range rows {
		if len(out) >= limit {
			break
		}
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		personID := asInt(item["id"])
		if personID <= 0 {
			continue
		}
		if _, exists := seen[personID]; exists {
			continue
		}
		seen[personID] = struct{}{}

		detail, detailErr := s.getTMDBJSON(ctx, fmt.Sprintf("/person/%d", personID), nil, tmdbLangChinese)
		if detailErr != nil {
			detail = item
		} else if needsPersonLocalizedFallback(detail) {
			fallback, fallbackErr := s.getTMDBJSON(ctx, fmt.Sprintf("/person/%d", personID), nil, "")
			if fallbackErr == nil {
				detail = mergeLocalizedPersonDetail(detail, fallback)
			}
		}

		actorName := strings.TrimSpace(asString(detail["name"]))
		if actorName == "" {
			actorName = strings.TrimSpace(asString(item["name"]))
		}
		if actorName == "" {
			continue
		}

		avatarURL := ""
		if profilePath := strings.TrimSpace(asString(detail["profile_path"])); profilePath != "" {
			avatarURL = "https://image.tmdb.org/t/p/w500" + profilePath
		}
		notes := strings.TrimSpace(asString(detail["biography"]))
		if len(notes) > 500 {
			notes = notes[:500]
		}

		out = append(out, ActorScrapeCandidate{
			Source:     "tmdb",
			ExternalID: strconv.Itoa(personID),
			Name:       actorName,
			Aliases:    extractStringSlice(detail["also_known_as"]),
			Gender:     tmdbGenderText(asInt(detail["gender"])),
			Country:    normalizeTMDBPlace(asString(detail["place_of_birth"])),
			BirthDate:  strings.TrimSpace(asString(detail["birthday"])),
			AvatarURL:  avatarURL,
			Notes:      notes,
			Raw:        detail,
		})
	}

	return out, nil
}

func needsPersonLocalizedFallback(detail map[string]any) bool {
	if isBlankAnyString(detail["name"]) || isBlankAnyString(detail["biography"]) {
		return true
	}
	return len(extractStringSlice(detail["also_known_as"])) == 0
}

func mergeLocalizedPersonDetail(primary, fallback map[string]any) map[string]any {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}

	fillBlankStringField(primary, fallback, "name")
	fillBlankStringField(primary, fallback, "biography")
	fillBlankStringField(primary, fallback, "place_of_birth")
	fillBlankStringField(primary, fallback, "birthday")
	fillBlankStringField(primary, fallback, "profile_path")
	if len(extractStringSlice(primary["also_known_as"])) == 0 {
		if aliases, ok := fallback["also_known_as"].([]any); ok && len(aliases) > 0 {
			primary["also_known_as"] = aliases
		}
	}
	return primary
}

func tmdbGenderText(v int) string {
	switch v {
	case 1:
		return "女"
	case 2:
		return "男"
	case 3:
		return "非二元"
	default:
		return ""
	}
}

func normalizeTMDBPlace(v string) string {
	place := strings.Join(strings.Fields(strings.TrimSpace(v)), " ")
	if place == "" {
		return ""
	}
	parts := strings.Split(place, ",")
	if len(parts) == 0 {
		return place
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func extractStringSlice(v any) []string {
	rows, ok := v.([]any)
	if !ok || len(rows) == 0 {
		return nil
	}
	out := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		name := strings.Join(strings.Fields(strings.TrimSpace(asString(row))), " ")
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
}

func (s *ScraperService) previewActorsJavDB(ctx context.Context, keyword string, limit int) ([]ActorScrapeCandidate, error) {
	baseURL := strings.TrimSuffix(strings.TrimSpace(s.avBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://javdb.com"
	}

	query := url.Values{}
	query.Set("q", keyword)
	query.Set("f", "actor")
	searchURL := fmt.Sprintf("%s/search?%s", baseURL, query.Encode())
	content, err := s.fetchAVHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	items := parseJavDBActorCandidates(content, baseURL, limit)
	return items, nil
}

func (s *ScraperService) fetchAVHTML(ctx context.Context, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("创建 AV 刮削请求失败: %w", err)
	}
	if strings.TrimSpace(s.avUserAgent) != "" {
		req.Header.Set("User-Agent", s.avUserAgent)
	}

	client := s.avHTTPClient
	if client == nil {
		client = s.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AV 演员查询失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("AV 站点请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("读取 AV 响应失败: %w", err)
	}
	return string(body), nil
}

func parseJavDBActorCandidates(docHTML, baseURL string, limit int) []ActorScrapeCandidate {
	matches := javDBActorAnchorRe.FindAllStringSubmatch(docHTML, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]ActorScrapeCandidate, 0, limit)
	seen := map[string]struct{}{}
	for _, m := range matches {
		if len(out) >= limit {
			break
		}
		if len(m) < 3 {
			continue
		}

		href := strings.TrimSpace(m[1])
		anchorInner := m[2]
		fullAnchor := m[0]
		actorID := extractJavDBActorID(href)
		name := stripHTMLText(anchorInner)
		if name == "" {
			if mm := titleAttrRe.FindStringSubmatch(fullAnchor); len(mm) > 1 {
				name = stripHTMLText(mm[1])
			}
		}
		if name == "" {
			continue
		}

		key := actorID
		if key == "" {
			key = strings.ToLower(name)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		avatarURL := ""
		if mm := imgSrcRe.FindStringSubmatch(fullAnchor); len(mm) > 1 {
			avatarURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		}
		out = append(out, ActorScrapeCandidate{
			Source:     "javdb",
			ExternalID: actorID,
			Name:       name,
			AvatarURL:  avatarURL,
			Raw: map[string]any{
				"href": toAbsoluteURL(baseURL, href),
			},
		})
	}
	return out
}

func extractJavDBActorID(href string) string {
	parsed, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return ""
	}
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return ""
	}
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return ""
	}
	if segments[len(segments)-2] != "actors" {
		return ""
	}
	return strings.TrimSpace(segments[len(segments)-1])
}

func stripHTMLText(raw string) string {
	plain := htmlTagRe.ReplaceAllString(raw, " ")
	plain = html.UnescapeString(plain)
	return strings.Join(strings.Fields(strings.TrimSpace(plain)), " ")
}

func toAbsoluteURL(baseURL, raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return v
	}
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return v
	}
	ref, err := url.Parse(v)
	if err != nil {
		return v
	}
	return base.ResolveReference(ref).String()
}

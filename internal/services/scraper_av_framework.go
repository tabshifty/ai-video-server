package services

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type avCrawlerProvider interface {
	Crawlers() []avCrawler
	Default() avCrawler
}

type avCrawler interface {
	Name() string
	SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error)
	FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error)
}

type staticAVCrawlerProvider struct {
	crawlers []avCrawler
}

func newAVCrawlerProvider(svc *ScraperService) avCrawlerProvider {
	return &staticAVCrawlerProvider{
		crawlers: []avCrawler{newJavDBAVCrawler(svc)},
	}
}

func (p *staticAVCrawlerProvider) Crawlers() []avCrawler {
	if len(p.crawlers) == 0 {
		return nil
	}
	out := make([]avCrawler, len(p.crawlers))
	copy(out, p.crawlers)
	return out
}

func (p *staticAVCrawlerProvider) Default() avCrawler {
	if len(p.crawlers) == 0 {
		return nil
	}
	return p.crawlers[0]
}

type avScrapeTrace struct {
	Source        string
	SearchKeyword string
	TargetCode    string
	SearchQueries []string
	SearchURLs    []string
	DetailURLs    []string
	Steps         []string
	Errors        []string
	MatchedBy     string
	MatchedCode   string
}

type avScrapeRunContext struct {
	trace avScrapeTrace
}

func newAVScrapeRunContext(keyword, targetCode string) *avScrapeRunContext {
	return &avScrapeRunContext{
		trace: avScrapeTrace{
			SearchKeyword: strings.TrimSpace(keyword),
			TargetCode:    strings.TrimSpace(targetCode),
		},
	}
}

func (r *avScrapeRunContext) setSource(source string) {
	source = strings.TrimSpace(source)
	if source == "" {
		return
	}
	r.trace.Source = source
}

func (r *avScrapeRunContext) addSearchQuery(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	r.trace.SearchQueries = appendUniqueString(r.trace.SearchQueries, query)
}

func (r *avScrapeRunContext) addSearchURL(searchURL string) {
	searchURL = strings.TrimSpace(searchURL)
	if searchURL == "" {
		return
	}
	r.trace.SearchURLs = appendUniqueString(r.trace.SearchURLs, searchURL)
}

func (r *avScrapeRunContext) addDetailURL(detailURL string) {
	detailURL = strings.TrimSpace(detailURL)
	if detailURL == "" {
		return
	}
	r.trace.DetailURLs = appendUniqueString(r.trace.DetailURLs, detailURL)
}

func (r *avScrapeRunContext) addStep(format string, args ...any) {
	step := strings.TrimSpace(fmt.Sprintf(format, args...))
	if step == "" {
		return
	}
	r.trace.Steps = append(r.trace.Steps, step)
}

func (r *avScrapeRunContext) addError(err error) {
	if err == nil {
		return
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return
	}
	r.trace.Errors = append(r.trace.Errors, msg)
}

func (r *avScrapeRunContext) setMatch(reason, code string) {
	r.trace.MatchedBy = strings.TrimSpace(reason)
	r.trace.MatchedCode = strings.TrimSpace(code)
}

func (r *avScrapeRunContext) toMap() map[string]any {
	trace := map[string]any{}
	if r.trace.Source != "" {
		trace["source"] = r.trace.Source
	}
	if r.trace.SearchKeyword != "" {
		trace["search_keyword"] = r.trace.SearchKeyword
	}
	if r.trace.TargetCode != "" {
		trace["target_code"] = r.trace.TargetCode
	}
	if len(r.trace.SearchQueries) > 0 {
		queries := make([]string, len(r.trace.SearchQueries))
		copy(queries, r.trace.SearchQueries)
		trace["search_queries"] = queries
	}
	if len(r.trace.SearchURLs) > 0 {
		searchURLs := make([]string, len(r.trace.SearchURLs))
		copy(searchURLs, r.trace.SearchURLs)
		trace["search_urls"] = searchURLs
	}
	if len(r.trace.DetailURLs) > 0 {
		detailURLs := make([]string, len(r.trace.DetailURLs))
		copy(detailURLs, r.trace.DetailURLs)
		trace["detail_urls"] = detailURLs
	}
	if len(r.trace.Steps) > 0 {
		steps := make([]string, len(r.trace.Steps))
		copy(steps, r.trace.Steps)
		trace["steps"] = steps
	}
	if len(r.trace.Errors) > 0 {
		errors := make([]string, len(r.trace.Errors))
		copy(errors, r.trace.Errors)
		trace["errors"] = errors
	}
	if r.trace.MatchedBy != "" {
		trace["matched_by"] = r.trace.MatchedBy
	}
	if r.trace.MatchedCode != "" {
		trace["matched_code"] = r.trace.MatchedCode
	}
	return trace
}

func appendUniqueString(dst []string, item string) []string {
	for _, existing := range dst {
		if existing == item {
			return dst
		}
	}
	return append(dst, item)
}

func (s *ScraperService) defaultAVCrawler() avCrawler {
	if s.avProvider == nil {
		s.avProvider = newAVCrawlerProvider(s)
	}
	crawler := s.avProvider.Default()
	if crawler == nil {
		crawler = newJavDBAVCrawler(s)
	}
	return crawler
}

func (s *ScraperService) searchAVCandidatesWithTrace(ctx context.Context, keyword string, limit int) ([]avScrapeCandidate, map[string]any, error) {
	normalizedKeyword := normalizeAVKeyword(keyword)
	if normalizedKeyword == "" {
		return nil, nil, fmt.Errorf("av keyword is required")
	}
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}

	targetCode := extractAVCode(normalizedKeyword)
	run := newAVScrapeRunContext(normalizedKeyword, targetCode)
	queries := buildAVSearchQueries(normalizedKeyword, targetCode)
	if len(queries) == 0 {
		return nil, run.toMap(), fmt.Errorf("av keyword is required")
	}
	for _, query := range queries {
		run.addSearchQuery(query)
	}

	if s.avProvider == nil {
		s.avProvider = newAVCrawlerProvider(s)
	}
	crawlers := s.avProvider.Crawlers()
	if len(crawlers) == 0 {
		crawler := newJavDBAVCrawler(s)
		crawlers = []avCrawler{crawler}
	}

	seen := map[string]struct{}{}
	out := make([]avScrapeCandidate, 0, limit)
	var firstErr error

	for _, crawler := range crawlers {
		run.setSource(crawler.Name())
		for _, query := range queries {
			hits, err := crawler.SearchCandidates(ctx, run, query, limit)
			if err != nil {
				run.addError(err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			for _, hit := range hits {
				key := strings.TrimSpace(hit.ExternalID)
				if key == "" {
					key = strings.ToLower(strings.TrimSpace(hit.DetailURL))
				}
				if key == "" {
					continue
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, hit)
				if len(out) >= limit {
					out = prioritizeAVCandidatesByCode(out, targetCode)
					if len(out) > 0 {
						_, reason := scoreAVCandidate(out[0], targetCode, normalizedKeyword)
						run.setMatch(reason, out[0].Code)
					}
					return out, run.toMap(), nil
				}
			}
		}
		if len(out) > 0 {
			break
		}
	}

	out = prioritizeAVCandidatesByCode(out, targetCode)
	if len(out) > 0 {
		_, reason := scoreAVCandidate(out[0], targetCode, normalizedKeyword)
		run.setMatch(reason, out[0].Code)
		return out, run.toMap(), nil
	}
	if firstErr != nil {
		return nil, run.toMap(), firstErr
	}
	return nil, run.toMap(), nil
}

func (s *ScraperService) fetchAVCandidateByDetailURLWithTrace(ctx context.Context, detailURL string) (avScrapeCandidate, map[string]any, error) {
	run := newAVScrapeRunContext("", extractAVCode(detailURL))
	crawler := s.defaultAVCrawler()
	run.setSource(crawler.Name())
	candidate, err := crawler.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		run.addError(err)
		return avScrapeCandidate{}, run.toMap(), err
	}
	run.setMatch("detail_url", candidate.Code)
	return candidate, run.toMap(), nil
}

func buildAVSearchQueries(keyword, code string) []string {
	queries := make([]string, 0, 3)
	code = strings.TrimSpace(code)
	keyword = strings.TrimSpace(keyword)
	if code != "" {
		queries = append(queries, code)
		if compact := normalizeAVCodeForCompare(code); compact != "" {
			queries = appendUniqueString(queries, compact)
		}
	}
	if keyword != "" {
		queries = appendUniqueString(queries, keyword)
	}
	if code != "" {
		if fallback := strings.ReplaceAll(code, "-", " "); fallback != "" {
			queries = appendUniqueString(queries, fallback)
		}
	}
	return queries
}

func scoreAVCandidate(candidate avScrapeCandidate, targetCode, keyword string) (int, string) {
	targetCode = strings.TrimSpace(targetCode)
	candidateCode := strings.TrimSpace(candidate.Code)
	if candidateCode == "" {
		candidateCode = extractAVCode(candidate.Title)
	}
	if candidateCode == "" {
		candidateCode = extractAVCode(candidate.DetailURL)
	}

	if targetCode != "" {
		if strings.EqualFold(candidateCode, targetCode) {
			return 400, "exact_code"
		}
		if normalizeAVCodeForCompare(candidateCode) == normalizeAVCodeForCompare(targetCode) {
			return 320, "normalized_code"
		}
		if strings.Contains(normalizeAVCodeForCompare(candidate.Title), normalizeAVCodeForCompare(targetCode)) {
			return 260, "title_contains_code"
		}
	}

	targetKeyword := normalizeAVCodeForCompare(keyword)
	candidateKeyword := normalizeAVCodeForCompare(candidate.Title)
	if targetKeyword != "" {
		if targetKeyword == candidateKeyword {
			return 220, "exact_keyword"
		}
		if strings.Contains(candidateKeyword, targetKeyword) || strings.Contains(targetKeyword, candidateKeyword) {
			return 180, "fuzzy_keyword"
		}
	}

	if strings.TrimSpace(candidate.ReleaseDate) != "" {
		return 80, "default_with_date"
	}
	return 20, "default"
}

type javDBAVCrawler struct {
	svc    *ScraperService
	parser avDetailParser
}

func newJavDBAVCrawler(svc *ScraperService) *javDBAVCrawler {
	return &javDBAVCrawler{
		svc:    svc,
		parser: &javDBDetailParser{},
	}
}

func (c *javDBAVCrawler) Name() string {
	return "javdb"
}

func (c *javDBAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}

	searchURL := c.buildSearchURL(query)
	run.addSearchURL(searchURL)
	run.addStep("search %s", searchURL)

	content, err := c.svc.fetchAVHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	hits := c.parseSearchHits(content, limit)
	if len(hits) == 0 {
		run.addStep("no search result for query=%s", query)
		return nil, nil
	}

	out := make([]avScrapeCandidate, 0, len(hits))
	seen := map[string]struct{}{}
	for _, hit := range hits {
		if len(out) >= limit {
			break
		}
		candidate, err := c.FetchByDetailURL(ctx, run, hit.DetailURL)
		if err != nil {
			run.addError(err)
			continue
		}
		if candidate.ExternalID == "" {
			candidate.ExternalID = strings.TrimSpace(hit.ExternalID)
		}
		if candidate.ExternalID == "" {
			candidate.ExternalID = extractJavDBVideoID(candidate.DetailURL)
		}
		key := strings.TrimSpace(candidate.ExternalID)
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(candidate.DetailURL))
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, candidate)
	}
	return out, nil
}

func (c *javDBAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	detailURL = strings.TrimSpace(detailURL)
	if detailURL == "" {
		return avScrapeCandidate{}, fmt.Errorf("detail url is required")
	}
	run.addDetailURL(detailURL)
	run.addStep("fetch detail %s", detailURL)

	content, err := c.svc.fetchAVHTML(ctx, detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	baseURL := c.baseURL()
	candidate, fieldState := c.parser.Parse(content, detailURL, baseURL)
	c.postProcess(&candidate)
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["field_state"] = fieldState
	return candidate, nil
}

func (c *javDBAVCrawler) buildSearchURL(query string) string {
	baseURL := c.baseURL()
	q := url.Values{}
	q.Set("q", strings.TrimSpace(query))
	q.Set("locale", "zh")
	return fmt.Sprintf("%s/search?%s", baseURL, q.Encode())
}

func (c *javDBAVCrawler) parseSearchHits(content string, limit int) []avSearchHit {
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}
	matches := javDBVideoAnchorRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	baseURL := c.baseURL()
	out := make([]avSearchHit, 0, limit)
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(out) >= limit {
			break
		}
		if len(match) < 2 {
			continue
		}
		detailURL := toAbsoluteURL(baseURL, strings.TrimSpace(match[1]))
		externalID := extractJavDBVideoID(match[1])
		if externalID == "" {
			externalID = extractJavDBVideoID(detailURL)
		}
		key := strings.TrimSpace(externalID)
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(detailURL))
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, avSearchHit{DetailURL: detailURL, ExternalID: externalID})
	}
	return out
}

func (c *javDBAVCrawler) postProcess(candidate *avScrapeCandidate) {
	if candidate == nil {
		return
	}
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	if posterURL := strings.TrimSpace(candidate.PosterURL); posterURL != "" {
		if strings.Contains(posterURL, "/covers/") {
			candidate.Raw["poster_thumb_url"] = strings.Replace(posterURL, "/covers/", "/thumbs/", 1)
		}
	}
	title := strings.ToLower(strings.TrimSpace(candidate.Title))
	if strings.Contains(title, "無碼") || strings.Contains(title, "无码") || strings.Contains(title, "uncensored") {
		candidate.Raw["mosaic"] = "无码"
	} else {
		candidate.Raw["mosaic"] = "有码"
	}
}

func (c *javDBAVCrawler) baseURL() string {
	baseURL := strings.TrimSuffix(strings.TrimSpace(c.svc.avBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://javdb.com"
	}
	return baseURL
}

type avSearchHit struct {
	DetailURL  string
	ExternalID string
}

type avDetailParser interface {
	Parse(content, detailURL, baseURL string) (avScrapeCandidate, map[string]string)
}

type javDBDetailParser struct{}

func (p *javDBDetailParser) Parse(content, detailURL, baseURL string) (avScrapeCandidate, map[string]string) {
	fieldState := map[string]string{
		"title":        "empty",
		"code":         "empty",
		"poster_url":   "empty",
		"overview":     "empty",
		"release_date": "empty",
		"actors":       "empty",
		"trailer":      "not_supported",
	}

	title := extractJavDBTitle(content)
	if title != "" {
		fieldState["title"] = "filled"
	}

	code := extractAVCode(title)
	if code == "" {
		if mm := javDBVideoCodeFieldRe.FindStringSubmatch(content); len(mm) > 1 {
			code = extractAVCode(stripHTMLText(mm[1]))
		}
	}
	if code == "" {
		code = extractAVCode(detailURL)
	}
	if code != "" {
		fieldState["code"] = "filled"
	}

	posterURL := ""
	if mm := javDBOGImageRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
	}
	if posterURL == "" {
		if mm := imgSrcRe.FindStringSubmatch(content); len(mm) > 1 {
			posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		}
	}
	if posterURL != "" {
		fieldState["poster_url"] = "filled"
	}

	releaseDate := ""
	if mm := javDBReleaseDateFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		releaseDate = normalizeAVDate(stripHTMLText(mm[1]))
	}
	if releaseDate == "" {
		if mm := javDBGenericDateInlineRe.FindStringSubmatch(content); len(mm) > 0 {
			releaseDate = normalizeAVDate(mm[0])
		}
	}
	if releaseDate != "" {
		fieldState["release_date"] = "filled"
	}

	overview := ""
	if mm := javDBDescriptionMetaRe.FindStringSubmatch(content); len(mm) > 1 {
		overview = stripHTMLText(mm[1])
	}
	if overview != "" {
		fieldState["overview"] = "filled"
	}

	actors := parseJavDBDetailActors(content)
	if len(actors) > 0 {
		fieldState["actors"] = "filled"
	}

	externalID := extractJavDBVideoID(detailURL)
	candidate := avScrapeCandidate{
		ExternalID:  externalID,
		Code:        code,
		Title:       title,
		Overview:    overview,
		PosterURL:   posterURL,
		ReleaseDate: releaseDate,
		Actors:      actors,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"detail_url":   detailURL,
			"external_id":  externalID,
			"code":         code,
			"title":        title,
			"overview":     overview,
			"poster_url":   posterURL,
			"release_date": releaseDate,
			"actors":       actors,
		},
	}
	return candidate, fieldState
}

func prioritizeAVCandidatesByCode(candidates []avScrapeCandidate, code string) []avScrapeCandidate {
	if len(candidates) <= 1 {
		return candidates
	}
	targetCode := strings.TrimSpace(code)
	base := append([]avScrapeCandidate(nil), candidates...)
	sort.SliceStable(base, func(i, j int) bool {
		scoreI, _ := scoreAVCandidate(base[i], targetCode, "")
		scoreJ, _ := scoreAVCandidate(base[j], targetCode, "")
		if scoreI == scoreJ {
			return i < j
		}
		return scoreI > scoreJ
	})
	return base
}

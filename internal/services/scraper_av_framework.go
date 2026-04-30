package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

var (
	javDBVideoCoverImgRe     = regexp.MustCompile(`(?is)<img[^>]*class=["'][^"']*video-cover[^"']*["'][^>]*src=["']([^"']+)["']`)
	javDBCodeClipboardRe     = regexp.MustCompile(`(?is)copy-to-clipboard[^>]*data-clipboard-text=["']([^"']+)["']`)
	javDBOverviewFieldRe     = regexp.MustCompile(`(?is)(?:簡介|简介|介紹|Description)[^<]{0,20}</strong>\s*<span[^>]*>(.*?)</span>`)
	javDBOverviewBlockRe     = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*(?:synopsis|introduction|overview|video-detail|content|description)[^"']*["'][^>]*>(.*?)</div>`)
	javDBActorFieldBlockRe   = regexp.MustCompile(`(?is)<span[^>]*>\s*<strong[^>]*>(?:演員|演员|Actress|Actor)[^<]*</strong>(.*?)</span>`)
	javDBInlineActorNameRe   = regexp.MustCompile(`(?is)<a[^>]*href=["'][^"']*/actors/[^"']+["'][^>]*>(.*?)</a>`)
	javDBFemaleActorNameRe   = regexp.MustCompile(`(?is)<strong[^>]*class=["'][^"']*(?:female|male)[^"']*["'][^>]*>.*?</strong>(.*?)</span>`)
	javBusMovieBoxHrefRe     = regexp.MustCompile(`(?is)<a[^>]*class=["'][^"']*movie-box[^"']*["'][^>]*href=["']([^"']+)["']`)
	javBusTitleRe            = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	javBusCodeFieldRe        = regexp.MustCompile(`(?is)(?:識別碼|识别码|識別码|番号|ID)[^<]{0,12}</span>\s*([^<\s]+)`)
	javBusCoverHrefRe        = regexp.MustCompile(`(?is)<a[^>]*class=["'][^"']*bigImage[^"']*["'][^>]*href=["']([^"']+)["']`)
	javBusReleaseFieldRe     = regexp.MustCompile(`(?is)(?:發行日期|发行日期|発売日|Release Date)[^<]{0,20}</span>\s*([^<]+)`)
	javBusActorNameRe        = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*star-name[^"']*["'][^>]*>\s*<a[^>]*>(.*?)</a>`)
	javBusExternalIDPathRe   = regexp.MustCompile(`(?i)/([a-z0-9][a-z0-9._-]*)/?$`)
	javLibrarySearchLinkRe   = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*\?v=jav[^"']+)["'][^>]*`)
	javLibraryTitleRe        = regexp.MustCompile(`(?is)<div[^>]*id=["']video_title["'][^>]*>.*?<h3[^>]*>\s*<a[^>]*>(.*?)</a>\s*</h3>`)
	javLibraryCodeFieldRe    = regexp.MustCompile(`(?is)<div[^>]*id=["']video_id["'][^>]*>.*?<td[^>]*class=["']text["'][^>]*>(.*?)</td>`)
	javLibraryCoverImgRe     = regexp.MustCompile(`(?is)<img[^>]*id=["']video_jacket_img["'][^>]*src=["']([^"']+)["']`)
	javLibraryReleaseFieldRe = regexp.MustCompile(`(?is)<div[^>]*id=["']video_date["'][^>]*>.*?<td[^>]*class=["']text["'][^>]*>(.*?)</td>`)
	javLibraryActorNameRe    = regexp.MustCompile(`(?is)<div[^>]*id=["']video_cast["'][^>]*>.*?<a[^>]*>(.*?)</a>`)
	fc2TitleRe               = regexp.MustCompile(`(?is)<div[^>]*data-section=["']userInfo["'][^>]*>.*?<h3[^>]*>(.*?)</h3>`)
	fc2SampleCoverRe         = regexp.MustCompile(`(?is)<ul[^>]*class=["'][^"']*items_article_SampleImagesArea[^"']*["'][^>]*>.*?<a[^>]*href=["']([^"']+)["']`)
	fc2MainThumbRe           = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*items_article_MainitemThumb[^"']*["'][^>]*>.*?<img[^>]*src=["']([^"']+)["']`)
	fc2ReleaseFieldRe        = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*items_article_Releasedate[^"']*["'][^>]*>.*?<p[^>]*>(.*?)</p>`)
	fc2OverviewMetaRe        = regexp.MustCompile(`(?is)<meta[^>]+name=["']description["'][^>]*content=["']([^"']+)["']`)
	fc2ExternalIDPathRe      = regexp.MustCompile(`(?i)/article/(\d{5,8})/?`)
	fc2NumberRe              = regexp.MustCompile(`(?i)\b(?:FC2(?:-?PPV)?)?[-_ ]?(\d{5,8})\b`)
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
		crawlers: []avCrawler{
			newJavDBAVCrawler(svc),
			newAirAVCCAVCrawler(svc),
			newAVSexAVCrawler(svc),
			newCableAVAVCrawler(svc),
			newHDOUBANAVCrawler(svc),
			newHSCangkuAVCrawler(svc),
			newIQQTVAVCrawler(svc),
			newSevenMMTVAVCrawler(svc),
			newDMMAVCrawler(svc),
			newMGStageAVCrawler(svc),
			newPrestigeAVCrawler(svc),
			newXCityAVCrawler(svc),
			newGetchuAVCrawler(svc),
			newThePornDBAVCrawler(svc),
			newJavBusAVCrawler(svc),
			newJavLibraryAVCrawler(svc),
			newFC2AVCrawler(svc),
			newFC2ClubAVCrawler(svc),
			newFC2HubAVCrawler(svc),
			newFC2PPVDBAVCrawler(svc),
			newAirAVAVCrawler(svc),
			newJav321AVCrawler(svc),
			newMywifeAVCrawler(svc),
			newAVSOXAVCrawler(svc),
			newFreeJAVBTAVCrawler(svc),
			newMadouquAVCrawler(svc),
			newMDTVAVCrawler(svc),
			newCNMDBAVCrawler(svc),
			newFalenoAVCrawler(svc),
			newFantasticaAVCrawler(svc),
			newGigaAVCrawler(svc),
			newJavdayAVCrawler(svc),
			newKin8AVCrawler(svc),
			newLove6AVCrawler(svc),
			newLulubarAVCrawler(svc),
		},
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

func (s *ScraperService) avSiteBaseURL(site, fallback string) string {
	site = strings.ToLower(strings.TrimSpace(site))
	if s.avSiteURLs != nil {
		if base := strings.TrimSuffix(strings.TrimSpace(s.avSiteURLs[site]), "/"); base != "" {
			return base
		}
	}
	configured := strings.TrimSuffix(strings.TrimSpace(s.avBaseURL), "/")
	if configured == "" {
		return fallback
	}
	if site == "javdb" {
		return configured
	}
	host := ""
	if parsed, err := url.Parse(configured); err == nil {
		host = strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	}
	if isLocalHost(host) {
		return configured
	}
	return fallback
}

func isLocalHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	if strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") {
		return true
	}
	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".lan") {
		return true
	}
	return false
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
		crawlers = []avCrawler{newJavDBAVCrawler(s)}
	}

	rawLimit := limit * 4
	if rawLimit < 12 {
		rawLimit = 12
	}

	seen := map[string]struct{}{}
	raw := make([]avScrapeCandidate, 0, rawLimit)
	var firstErr error

	for _, crawler := range crawlers {
		run.setSource(crawler.Name())
		for _, query := range queries {
			if len(raw) >= rawLimit {
				break
			}
			hits, err := crawler.SearchCandidates(ctx, run, query, limit)
			if err != nil {
				run.addError(err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			for _, hit := range hits {
				if len(raw) >= rawLimit {
					break
				}
				hit.Source = normalizeAVSourceName(chooseStr(hit.Source, crawler.Name()))
				key := avCandidateIdentityKey(hit)
				if key == "" {
					continue
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				raw = append(raw, hit)
			}
		}
	}

	raw = prioritizeAVCandidatesByCode(raw, targetCode)
	if len(raw) == 0 {
		if firstErr != nil {
			return nil, run.toMap(), firstErr
		}
		return nil, run.toMap(), nil
	}

	final := append([]avScrapeCandidate(nil), raw...)
	if merged, ok := mergeAVCandidatesByField(raw, targetCode, normalizedKeyword); ok {
		mergedKey := avCandidateIdentityKey(merged)
		replaced := false
		for i := range final {
			if mergedKey != "" && avCandidateIdentityKey(final[i]) == mergedKey {
				final[i] = merged
				replaced = true
				break
			}
		}
		if !replaced {
			final = append([]avScrapeCandidate{merged}, final...)
		}
	}

	if len(final) > limit {
		final = final[:limit]
	}
	if len(final) > 0 {
		run.setSource(final[0].Source)
		_, reason := scoreAVCandidate(final[0], targetCode, normalizedKeyword)
		run.setMatch(reason, final[0].Code)
	}
	trace := run.toMap()
	trace["sources"] = collectAVCandidateSources(final)
	trace["raw_candidate_count"] = len(raw)
	if len(final) > 0 && final[0].Raw != nil {
		if v, ok := final[0].Raw["field_sources"]; ok {
			trace["field_sources"] = v
		}
		if v, ok := final[0].Raw["merged_sources"]; ok {
			trace["merged_sources"] = v
		}
		if v, ok := final[0].Raw["poster_source"]; ok {
			trace["poster_source"] = v
		}
		if v, ok := final[0].Raw["poster_quality"]; ok {
			trace["poster_quality"] = v
		}
		if v, ok := final[0].Raw["poster_decision"]; ok {
			trace["poster_decision"] = v
		}
	}
	return final, trace, nil
}

func (s *ScraperService) fetchAVCandidateByDetailURLWithTrace(ctx context.Context, detailURL string) (avScrapeCandidate, map[string]any, error) {
	return s.fetchAVCandidateBySourceAndDetailURLWithTrace(ctx, "", detailURL)
}

func (s *ScraperService) fetchAVCandidateBySourceAndDetailURLWithTrace(ctx context.Context, sourceHint, detailURL string) (avScrapeCandidate, map[string]any, error) {
	run := newAVScrapeRunContext("", extractAVCode(detailURL))
	crawler := s.resolveAVCrawlerBySourceOrDetailURL(sourceHint, detailURL)
	run.setSource(crawler.Name())
	candidate, err := crawler.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		run.addError(err)
		return avScrapeCandidate{}, run.toMap(), err
	}
	if candidate.Source == "" {
		candidate.Source = crawler.Name()
	}
	run.setSource(candidate.Source)
	run.setMatch("detail_url", candidate.Code)
	trace := run.toMap()
	trace["resolved_source"] = candidate.Source
	return candidate, trace, nil
}

func (s *ScraperService) resolveAVCrawlerBySourceOrDetailURL(sourceHint, detailURL string) avCrawler {
	if crawler := s.resolveAVCrawlerBySource(sourceHint); crawler != nil {
		return crawler
	}
	return s.resolveAVCrawlerByDetailURL(detailURL)
}

func (s *ScraperService) resolveAVCrawlerBySource(source string) avCrawler {
	source = normalizeAVSourceName(source)
	if source == "" {
		return nil
	}
	if s.avProvider == nil {
		s.avProvider = newAVCrawlerProvider(s)
	}
	for _, crawler := range s.avProvider.Crawlers() {
		if normalizeAVSourceName(crawler.Name()) == source {
			return crawler
		}
	}
	return nil
}

func (s *ScraperService) resolveAVCrawlerByDetailURL(detailURL string) avCrawler {
	if s.avProvider == nil {
		s.avProvider = newAVCrawlerProvider(s)
	}
	crawlers := s.avProvider.Crawlers()
	if len(crawlers) == 0 {
		return s.defaultAVCrawler()
	}

	host := ""
	path := ""
	query := url.Values{}
	if parsed, err := url.Parse(strings.TrimSpace(detailURL)); err == nil {
		host = strings.ToLower(strings.TrimSpace(parsed.Hostname()))
		path = strings.ToLower(strings.TrimSpace(parsed.Path))
		query = parsed.Query()
	}

	for _, crawler := range crawlers {
		if matchAVCrawlerDetailURL(crawler.Name(), host, path, query) {
			return crawler
		}
	}
	for _, crawler := range crawlers {
		if crawler.Name() == "javdb" {
			return crawler
		}
	}
	return crawlers[0]
}

func (s *ScraperService) buildAVDetailURLBySource(source, externalID string) string {
	source = normalizeAVSourceName(source)
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return ""
	}
	switch source {
	case "airav_cc":
		return toAbsoluteURL(s.avSiteBaseURL("airav_cc", "https://airav.io"), "/jp/playon.aspx?hid="+url.QueryEscape(externalID))
	case "avsex":
		return toAbsoluteURL(s.avSiteBaseURL("avsex", "https://gg5.co"), "/cn/video/detail/"+url.PathEscape(externalID))
	case "cableav":
		return toAbsoluteURL(s.avSiteBaseURL("cableav", "https://cableav.tv"), "/"+url.PathEscape(externalID)+"/")
	case "hdouban":
		return toAbsoluteURL(s.avSiteBaseURL("hdouban", "https://ormtgu.com"), "/moviedetail/"+url.PathEscape(externalID))
	case "iqqtv":
		return toAbsoluteURL(s.avSiteBaseURL("iqqtv", "https://iqq5.xyz"), "/jp/player/"+url.PathEscape(externalID))
	case "7mmtv":
		return toAbsoluteURL(s.avSiteBaseURL("7mmtv", "https://www.7mmtv.sx"), "/zh/view_video/"+url.PathEscape(externalID)+".html")
	case "dmm":
		return toAbsoluteURL(s.avSiteBaseURL("dmm", "https://www.dmm.co.jp"), "/mono/dvd/-/detail/=/cid="+url.PathEscape(strings.ToLower(externalID))+"/")
	case "mgstage":
		return toAbsoluteURL(s.avSiteBaseURL("mgstage", "https://www.mgstage.com"), "/product/product_detail/"+url.PathEscape(externalID)+"/")
	case "prestige":
		return toAbsoluteURL(s.avSiteBaseURL("prestige", "https://www.prestige-av.com"), "/api/product/"+url.PathEscape(externalID))
	case "xcity":
		return toAbsoluteURL(s.avSiteBaseURL("xcity", "https://xcity.jp"), "/avod/detail/?id="+url.QueryEscape(externalID))
	case "getchu":
		return toAbsoluteURL(s.avSiteBaseURL("getchu", "https://www.getchu.com"), "/soft.phtml?id="+url.QueryEscape(externalID)+"&gc=gc")
	case "theporndb":
		return toAbsoluteURL(s.avSiteBaseURL("theporndb", "https://api.theporndb.net"), "/scenes/"+url.PathEscape(externalID))
	case "javbus":
		return toAbsoluteURL(s.avSiteBaseURL("javbus", "https://www.javbus.com"), "/"+externalID)
	case "javlibrary":
		base := s.avSiteBaseURL("javlibrary", "https://www.javlibrary.com")
		q := url.Values{}
		q.Set("v", externalID)
		return fmt.Sprintf("%s/cn/?%s", strings.TrimSuffix(base, "/"), q.Encode())
	case "fc2":
		number := normalizeFC2NumericID(externalID)
		if number == "" {
			number = normalizeFC2NumericID(extractAVCode(externalID))
		}
		if number == "" {
			number = normalizeFC2NumericID(externalID)
		}
		if number == "" {
			return ""
		}
		return toAbsoluteURL(s.avSiteBaseURL("fc2", "https://adult.contents.fc2.com"), "/article/"+number+"/")
	case "fc2club":
		if code := buildFC2PPVPathCode(externalID); code != "" {
			return toAbsoluteURL(s.avSiteBaseURL("fc2club", "https://fc2club.com"), "/html/"+url.PathEscape(code)+".html")
		}
		return ""
	case "fc2hub":
		if number := normalizeFC2NumericID(externalID); number != "" {
			return toAbsoluteURL(s.avSiteBaseURL("fc2hub", "https://fc2hub.com"), "/detail/"+url.PathEscape(number))
		}
		return ""
	case "fc2ppvdb":
		if code := buildFC2PPVPathCode(externalID); code != "" {
			return toAbsoluteURL(s.avSiteBaseURL("fc2ppvdb", "https://fc2ppvdb.com"), "/articles/"+url.PathEscape(code))
		}
		return ""
	case "airav":
		if code := buildFC2PPVPathCode(externalID); code != "" {
			return toAbsoluteURL(s.avSiteBaseURL("airav", "https://www.airav.wiki"), "/video/"+url.PathEscape(code))
		}
		return ""
	case "jav321":
		return toAbsoluteURL(s.avSiteBaseURL("jav321", "https://www.jav321.com"), "/video/"+url.PathEscape(strings.ToLower(externalID)))
	case "mywife":
		return toAbsoluteURL(s.avSiteBaseURL("mywife", "https://www.mywife.cc"), "/teigaku/model/no/"+url.PathEscape(externalID))
	case "avsox":
		return toAbsoluteURL(s.avSiteBaseURL("avsox", "https://avsox.host"), "/cn/movie/"+url.PathEscape(strings.ToLower(externalID)))
	case "freejavbt":
		if code := buildFC2PPVPathCode(externalID); code != "" {
			return toAbsoluteURL(s.avSiteBaseURL("freejavbt", "https://freejavbt.com"), "/detail/"+url.PathEscape(code))
		}
		return ""
	case "madouqu":
		return toAbsoluteURL(s.avSiteBaseURL("madouqu", "https://madouqu.com"), "/archives/"+url.PathEscape(externalID))
	case "mdtv.com":
		return toAbsoluteURL(s.avSiteBaseURL("mdtv.com", "https://mdtv.com.cn"), "/video/"+url.PathEscape(externalID))
	case "cnmdb":
		return toAbsoluteURL(s.avSiteBaseURL("cnmdb", "https://cnmdb.com"), "/video/"+url.PathEscape(externalID))
	case "faleno":
		return toAbsoluteURL(s.avSiteBaseURL("faleno", "https://faleno.jp"), "/top/works/"+url.PathEscape(strings.ToLower(externalID))+"/")
	case "fantastica":
		return toAbsoluteURL(s.avSiteBaseURL("fantastica", "https://fantastica-vr.com"), "/items/detail/"+url.PathEscape(strings.ToUpper(externalID)))
	case "giga":
		return toAbsoluteURL(s.avSiteBaseURL("giga", "https://www.giga-web.jp"), "/product/index.php?product_id="+url.QueryEscape(externalID))
	case "javday":
		return toAbsoluteURL(s.avSiteBaseURL("javday", "https://javday.tv"), "/videos/"+url.PathEscape(strings.ToLower(externalID))+"/")
	case "kin8":
		return toAbsoluteURL(s.avSiteBaseURL("kin8", "https://www.kin8tengoku.com"), "/moviepages/"+url.PathEscape(externalID)+"/index.html")
	case "love6":
		return toAbsoluteURL(s.avSiteBaseURL("love6", "https://love6.tv"), "/albums/view/"+url.PathEscape(externalID))
	case "lulubar":
		return toAbsoluteURL(s.avSiteBaseURL("lulubar", "https://lulubar.co"), "/video/detail?id="+url.QueryEscape(externalID))
	default:
		return toAbsoluteURL(s.avSiteBaseURL("javdb", "https://javdb.com"), "/v/"+externalID)
	}
}

func matchAVCrawlerDetailURL(name, host, path string, query url.Values) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	path = strings.ToLower(strings.TrimSpace(path))
	switch normalizeAVSourceName(name) {
	case "airav_cc":
		return strings.Contains(path, "playon.aspx") || query.Get("hid") != ""
	case "avsex":
		if strings.Contains(host, "gg5.") || strings.Contains(host, "9sex.") || strings.Contains(host, "paycalling") || strings.Contains(host, "avsex.") {
			return true
		}
		return strings.Contains(path, "/video/detail/")
	case "cableav":
		if strings.Contains(host, "cableav") {
			return true
		}
		trimmed := strings.Trim(path, "/")
		return trimmed != "" && strings.Count(trimmed, "/") == 0 && !strings.Contains(trimmed, ".")
	case "hdouban":
		if strings.Contains(host, "ormtgu") || strings.Contains(host, "byym21") || strings.Contains(host, "huangdb2") {
			return true
		}
		return strings.Contains(path, "/moviedetail/")
	case "iqqtv":
		if strings.Contains(host, "iqq") {
			return true
		}
		return strings.Contains(path, "/player/")
	case "7mmtv":
		if strings.Contains(host, "7mmtv") {
			return true
		}
		return strings.Contains(path, "/view_video/")
	case "dmm":
		if strings.Contains(host, "dmm.") {
			return true
		}
		return strings.Contains(path, "/mono/dvd/-/detail/")
	case "mgstage":
		if strings.Contains(host, "mgstage") {
			return true
		}
		return strings.Contains(path, "/product/product_detail/")
	case "prestige":
		if strings.Contains(host, "prestige-av") {
			return true
		}
		return strings.Contains(path, "/api/product/")
	case "xcity":
		if strings.Contains(host, "xcity") {
			return true
		}
		return strings.Contains(path, "/avod/detail/")
	case "getchu":
		if strings.Contains(host, "getchu") {
			return true
		}
		return strings.Contains(path, "/soft.phtml")
	case "theporndb":
		if strings.Contains(host, "theporndb") {
			return true
		}
		return strings.Contains(path, "/scenes/")
	case "fc2club":
		return strings.HasPrefix(path, "/html/") && strings.HasSuffix(path, ".html")
	case "fc2hub":
		return fc2HubDetailPathRe.MatchString(path)
	case "fc2ppvdb":
		return fc2PPVDBDetailPathRe.MatchString(path)
	case "airav":
		return airAVDetailPathRe.MatchString(path)
	case "jav321":
		return jav321DetailPathRe.MatchString(path) && !strings.Contains(path, "fc2-ppv-")
	case "mywife":
		return mywifeDetailPathRe.MatchString(path)
	case "avsox":
		if strings.Contains(host, "avsox") {
			return true
		}
		return avsoxDetailPathRe.MatchString(path)
	case "freejavbt":
		if strings.Contains(host, "freejavbt") {
			return true
		}
		return freeJAVBTDetailPathRe.MatchString(path)
	case "madouqu":
		if strings.Contains(host, "madouqu") {
			return true
		}
		return madouquDetailPathRe.MatchString(path)
	case "mdtv.com":
		if strings.Contains(host, "mdtv") {
			return true
		}
		return false
	case "cnmdb":
		if strings.Contains(host, "cnmdb") {
			return true
		}
		return false
	case "faleno":
		if strings.Contains(host, "faleno") || strings.Contains(host, "falenogroup") {
			return true
		}
		return falenoDetailPathRe.MatchString(path)
	case "fantastica":
		if strings.Contains(host, "fantastica") {
			return true
		}
		return fantasticaDetailPathRe.MatchString(path)
	case "giga":
		if strings.Contains(host, "giga-web") {
			return true
		}
		return gigaDetailPathRe.MatchString(path) && query.Get("product_id") != ""
	case "javday":
		if strings.Contains(host, "javday") {
			return true
		}
		return javdayDetailPathRe.MatchString(path)
	case "kin8":
		if strings.Contains(host, "kin8") {
			return true
		}
		return kin8DetailPathRe.MatchString(path)
	case "love6":
		if strings.Contains(host, "love6") {
			return true
		}
		return love6DetailPathRe.MatchString(path)
	case "lulubar":
		if strings.Contains(host, "lulubar") {
			return true
		}
		return lulubarDetailPathRe.MatchString(path) && query.Get("id") != ""
	case "javdb":
		if strings.Contains(host, "javdb") {
			return true
		}
		return strings.HasPrefix(path, "/v/")
	case "javbus":
		if strings.Contains(host, "javbus") {
			return true
		}
		trimmed := strings.Trim(path, "/")
		return trimmed != "" && strings.Count(trimmed, "/") == 0 && !strings.Contains(trimmed, ".") && !strings.Contains(path, "vl_searchbyid") && !strings.Contains(path, "/article/") && !strings.HasPrefix(path, "/v/")
	case "javlibrary":
		if strings.Contains(host, "javlibrary") {
			return true
		}
		return strings.Contains(path, "vl_searchbyid") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(query.Get("v"))), "jav")
	case "fc2":
		if strings.Contains(host, "adult.contents.fc2.com") {
			return true
		}
		return strings.Contains(path, "/article/")
	default:
		return false
	}
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
	if err := validateJavDBSearchPage(content, searchURL); err != nil {
		return nil, err
	}

	hits := c.parseSearchHits(content, query, limit)
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
		candidate.Source = c.Name()
		key := avCandidateIdentityKey(candidate)
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
	candidate.Source = c.Name()
	candidate.DetailURL = chooseStr(candidate.DetailURL, detailURL)
	if candidate.ExternalID == "" {
		candidate.ExternalID = extractJavDBVideoID(candidate.DetailURL)
	}
	c.postProcess(&candidate)
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["field_state"] = fieldState
	candidate.Raw["source"] = c.Name()
	enrichAVCandidatePoster(&candidate)
	return candidate, nil
}

func (c *javDBAVCrawler) buildSearchURL(query string) string {
	baseURL := c.baseURL()
	q := url.Values{}
	q.Set("q", strings.TrimSpace(query))
	q.Set("locale", "zh")
	return fmt.Sprintf("%s/search?%s", baseURL, q.Encode())
}

func (c *javDBAVCrawler) parseSearchHits(content string, query string, limit int) []avSearchHit {
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}
	matches := javDBVideoAnchorRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	baseURL := c.baseURL()
	out := make([]avSearchHit, 0, limit)
	exactMatches := make([]avSearchHit, 0, limit)
	fuzzyMatches := make([]avSearchHit, 0, limit)
	seen := map[string]struct{}{}
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	cleanQuery := normalizeJavDBSearchCompare(query)
	for _, match := range matches {
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
		hit := avSearchHit{DetailURL: detailURL, ExternalID: externalID}
		out = append(out, hit)

		text := ""
		if len(match) > 2 {
			text = strings.ToUpper(normalizeWhitespace(stripHTMLText(match[2])))
		}
		switch {
		case upperQuery != "" && strings.Contains(text, upperQuery):
			exactMatches = append(exactMatches, hit)
		case cleanQuery != "" && strings.Contains(normalizeJavDBSearchCompare(text), cleanQuery):
			fuzzyMatches = append(fuzzyMatches, hit)
		}
	}
	if len(exactMatches) > 0 {
		if len(exactMatches) > limit {
			return exactMatches[:limit]
		}
		return exactMatches
	}
	if len(fuzzyMatches) > 0 {
		if len(fuzzyMatches) > limit {
			return fuzzyMatches[:limit]
		}
		return fuzzyMatches
	}
	if len(out) > limit {
		return out[:limit]
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
	return c.svc.avSiteBaseURL("javdb", "https://javdb.com")
}

type javBusAVCrawler struct {
	svc    *ScraperService
	parser avDetailParser
}

func newJavBusAVCrawler(svc *ScraperService) *javBusAVCrawler {
	return &javBusAVCrawler{
		svc:    svc,
		parser: &javBusDetailParser{},
	}
}

func (c *javBusAVCrawler) Name() string {
	return "javbus"
}

func (c *javBusAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
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
		candidate.Source = c.Name()
		key := avCandidateIdentityKey(candidate)
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

func (c *javBusAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
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
	candidate.Source = c.Name()
	candidate.DetailURL = chooseStr(candidate.DetailURL, detailURL)
	if candidate.ExternalID == "" {
		candidate.ExternalID = extractJavBusVideoID(candidate.DetailURL)
	}
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["field_state"] = fieldState
	candidate.Raw["source"] = c.Name()
	enrichAVCandidatePoster(&candidate)
	return candidate, nil
}

func (c *javBusAVCrawler) buildSearchURL(query string) string {
	baseURL := strings.TrimSuffix(c.baseURL(), "/")
	q := url.PathEscape(strings.TrimSpace(query))
	return fmt.Sprintf("%s/search/%s", baseURL, q)
}

func (c *javBusAVCrawler) parseSearchHits(content string, limit int) []avSearchHit {
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}
	matches := javBusMovieBoxHrefRe.FindAllStringSubmatch(content, -1)
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
		externalID := extractJavBusVideoID(detailURL)
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

func (c *javBusAVCrawler) baseURL() string {
	return c.svc.avSiteBaseURL("javbus", "https://www.javbus.com")
}

type javLibraryAVCrawler struct {
	svc    *ScraperService
	parser avDetailParser
}

func newJavLibraryAVCrawler(svc *ScraperService) *javLibraryAVCrawler {
	return &javLibraryAVCrawler{
		svc:    svc,
		parser: &javLibraryDetailParser{},
	}
}

func (c *javLibraryAVCrawler) Name() string {
	return "javlibrary"
}

func (c *javLibraryAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
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
		candidate.Source = c.Name()
		key := avCandidateIdentityKey(candidate)
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

func (c *javLibraryAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
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
	candidate.Source = c.Name()
	candidate.DetailURL = chooseStr(candidate.DetailURL, detailURL)
	if candidate.ExternalID == "" {
		candidate.ExternalID = extractJavLibraryVideoID(candidate.DetailURL)
	}
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["field_state"] = fieldState
	candidate.Raw["source"] = c.Name()
	enrichAVCandidatePoster(&candidate)
	return candidate, nil
}

func (c *javLibraryAVCrawler) buildSearchURL(query string) string {
	baseURL := strings.TrimSuffix(c.baseURL(), "/")
	q := url.Values{}
	q.Set("keyword", strings.TrimSpace(query))
	return fmt.Sprintf("%s/cn/vl_searchbyid.php?%s", baseURL, q.Encode())
}

func (c *javLibraryAVCrawler) parseSearchHits(content string, limit int) []avSearchHit {
	if limit <= 0 {
		limit = avPreviewLimitDefault
	}
	matches := javLibrarySearchLinkRe.FindAllStringSubmatch(content, -1)
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
		externalID := extractJavLibraryVideoID(detailURL)
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

func (c *javLibraryAVCrawler) baseURL() string {
	return c.svc.avSiteBaseURL("javlibrary", "https://www.javlibrary.com")
}

type fc2AVCrawler struct {
	svc    *ScraperService
	parser avDetailParser
}

func newFC2AVCrawler(svc *ScraperService) *fc2AVCrawler {
	return &fc2AVCrawler{
		svc:    svc,
		parser: &fc2DetailParser{},
	}
}

func (c *fc2AVCrawler) Name() string {
	return "fc2"
}

func (c *fc2AVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	number := normalizeFC2NumericID(query)
	if number == "" {
		code := extractAVCode(query)
		number = normalizeFC2NumericID(code)
	}
	if number == "" {
		return nil, nil
	}
	detailURL := toAbsoluteURL(c.baseURL(), "/article/"+number+"/")
	run.addSearchURL(detailURL)
	run.addStep("fc2 direct detail %s", detailURL)
	candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.Title) == "" {
		return nil, nil
	}
	candidate.Source = c.Name()
	if candidate.ExternalID == "" {
		candidate.ExternalID = number
	}
	return []avScrapeCandidate{candidate}, nil
}

func (c *fc2AVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
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
	candidate.Source = c.Name()
	candidate.DetailURL = chooseStr(candidate.DetailURL, detailURL)
	if candidate.ExternalID == "" {
		candidate.ExternalID = extractFC2VideoID(candidate.DetailURL)
	}
	if candidate.ExternalID != "" && strings.TrimSpace(candidate.Code) == "" {
		candidate.Code = "FC2-" + normalizeFC2NumericID(candidate.ExternalID)
	}
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["field_state"] = fieldState
	candidate.Raw["source"] = c.Name()
	enrichAVCandidatePoster(&candidate)
	return candidate, nil
}

func (c *fc2AVCrawler) baseURL() string {
	return c.svc.avSiteBaseURL("fc2", "https://adult.contents.fc2.com")
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
	fieldState := defaultAVFieldState()
	root, _ := html.Parse(strings.NewReader(content))

	title := extractJavDBTitleFromDOM(root)
	if title == "" {
		title = extractJavDBTitle(content)
	}
	if title != "" {
		fieldState["title"] = "filled"
	}

	code := extractJavDBCodeFromDOM(root)
	if code == "" {
		code = extractAVCode(title)
	}
	if code == "" {
		if mm := javDBCodeClipboardRe.FindStringSubmatch(content); len(mm) > 1 {
			code = extractAVCode(stripHTMLText(mm[1]))
		}
	}
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

	posterURL, posterSource := extractJavDBPosterURLFromDOM(root, detailURL, baseURL)
	if posterURL == "" {
		posterURL, posterSource = extractJavDBPosterURL(content, baseURL)
	}
	if posterURL != "" {
		fieldState["poster_url"] = "filled"
	}

	releaseDate := extractJavDBReleaseDateFromDOM(root)
	if mm := javDBReleaseDateFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		if releaseDate == "" {
			releaseDate = normalizeAVDate(stripHTMLText(mm[1]))
		}
	}
	if releaseDate == "" {
		if mm := javDBGenericDateInlineRe.FindStringSubmatch(content); len(mm) > 0 {
			releaseDate = normalizeAVDate(mm[0])
		}
	}
	if releaseDate != "" {
		fieldState["release_date"] = "filled"
	}

	overview, overviewSource := extractJavDBOverviewFromDOM(root)
	if overview == "" {
		overview, overviewSource = extractJavDBOverview(content)
	}
	if overview != "" {
		fieldState["overview"] = "filled"
	}

	actors := extractJavDBActorsFromDOM(root)
	if len(actors) == 0 {
		actors = parseJavDBDetailActorsStrict(content)
	}
	if len(actors) > 0 {
		fieldState["actors"] = "filled"
	}

	externalID := extractJavDBVideoID(detailURL)
	candidate := avScrapeCandidate{
		Source:      "javdb",
		ExternalID:  externalID,
		Code:        code,
		Title:       title,
		Overview:    overview,
		PosterURL:   posterURL,
		ReleaseDate: releaseDate,
		Actors:      actors,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"source":          "javdb",
			"detail_url":      detailURL,
			"external_id":     externalID,
			"code":            code,
			"title":           title,
			"overview":        overview,
			"overview_source": overviewSource,
			"poster_url":      posterURL,
			"poster_source":   posterSource,
			"release_date":    releaseDate,
			"actors":          actors,
		},
	}
	return candidate, fieldState
}

func validateJavDBSearchPage(content string, searchURL string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	switch {
	case strings.Contains(content, "The owner of this website has banned your access based on your browser's behaving"):
		return fmt.Errorf("javdb 搜索被封禁：%s", searchURL)
	case strings.Contains(content, "Due to copyright restrictions"):
		return fmt.Errorf("javdb 搜索因版权限制被拦截：%s", searchURL)
	case strings.Contains(strings.ToLower(content), "ray-id"):
		return fmt.Errorf("javdb 搜索被 Cloudflare 拦截")
	default:
		return nil
	}
}

func normalizeJavDBSearchCompare(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	return strings.NewReplacer(".", "", "-", "", " ", "").Replace(value)
}

func extractJavDBTitleFromDOM(root *html.Node) string {
	title := nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "strong" && hasClass(n, "current-title")
	}))
	if title != "" {
		return title
	}
	return ""
}

func extractJavDBCodeFromDOM(root *html.Node) string {
	code := strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "copy-to-clipboard")
	}, "data-clipboard-text"))
	return extractAVCode(code)
}

func extractJavDBPosterURLFromDOM(root *html.Node, detailURL, baseURL string) (string, string) {
	posterURL := strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img" && hasClass(n, "video-cover")
	}, "src"))
	if posterURL == "" {
		return "", ""
	}
	posterURL = toAbsoluteURL(chooseStr(detailURL, baseURL), posterURL)
	if isLikelyLogoURL(posterURL) {
		return "", ""
	}
	return posterURL, "video_cover"
}

func extractJavDBReleaseDateFromDOM(root *html.Node) string {
	return normalizeAVDate(extractJavDBPanelValue(root, "Released Date:", "日期:"))
}

func extractJavDBOverviewFromDOM(root *html.Node) (string, string) {
	for _, script := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "script" && strings.EqualFold(attrValue(n, "type"), "application/ld+json")
	}) {
		payload := strings.TrimSpace(rawNodeText(script))
		if payload == "" {
			continue
		}
		var object map[string]any
		if err := json.Unmarshal([]byte(payload), &object); err == nil {
			if description := normalizeWhitespace(asString(object["description"])); description != "" && !isLikelyGenericOverview(description) {
				return description, "ld_json"
			}
		}
		var objects []map[string]any
		if err := json.Unmarshal([]byte(payload), &objects); err == nil {
			for _, object := range objects {
				if description := normalizeWhitespace(asString(object["description"])); description != "" && !isLikelyGenericOverview(description) {
					return description, "ld_json"
				}
			}
		}
	}
	return "", ""
}

func extractJavDBActorsFromDOM(root *html.Node) []string {
	if root == nil {
		return nil
	}
	actors := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, span := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "span" }) {
		if findFirst(span, func(n *html.Node) bool {
			return n.Type == html.ElementNode && (hasClass(n, "female") || hasClass(n, "male"))
		}) == nil {
			continue
		}
		for _, actorNode := range findAll(span, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "a" }) {
			name := normalizeWhitespace(nodeText(actorNode))
			if !isLikelyActorName(name) {
				continue
			}
			key := strings.ToLower(name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			actors = append(actors, name)
		}
	}
	if len(actors) == 0 {
		return nil
	}
	return actors
}

func extractJavDBPanelValue(root *html.Node, labels ...string) string {
	for _, strong := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "strong"
	}) {
		text := normalizeWhitespace(nodeText(strong))
		for _, label := range labels {
			if strings.EqualFold(text, strings.TrimSpace(label)) {
				parent := strong.Parent
				if parent == nil {
					continue
				}
				for child := parent.FirstChild; child != nil; child = child.NextSibling {
					if child == strong {
						continue
					}
					if value := normalizeWhitespace(nodeText(child)); value != "" {
						return value
					}
				}
			}
		}
	}
	return ""
}

func rawNodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			parts = append(parts, n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(parts, "")
}

type javBusDetailParser struct{}

func (p *javBusDetailParser) Parse(content, detailURL, baseURL string) (avScrapeCandidate, map[string]string) {
	fieldState := defaultAVFieldState()

	title := ""
	if mm := javBusTitleRe.FindStringSubmatch(content); len(mm) > 1 {
		title = stripHTMLText(mm[1])
	}
	if title != "" {
		fieldState["title"] = "filled"
	}

	code := ""
	if mm := javBusCodeFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		code = extractAVCode(stripHTMLText(mm[1]))
	}
	if code == "" {
		code = extractAVCode(title)
	}
	if code == "" {
		code = extractAVCode(detailURL)
	}
	if code != "" {
		fieldState["code"] = "filled"
	}

	posterURL := ""
	posterSource := ""
	if mm := javBusCoverHrefRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		posterSource = "big_image"
	}
	if posterURL == "" {
		if mm := javLibraryCoverImgRe.FindStringSubmatch(content); len(mm) > 1 {
			posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
			posterSource = "cover_img"
		}
	}
	if isLikelyLogoURL(posterURL) {
		posterURL = ""
		posterSource = ""
	}
	if posterURL != "" {
		fieldState["poster_url"] = "filled"
	}

	releaseDate := ""
	if mm := javBusReleaseFieldRe.FindStringSubmatch(content); len(mm) > 1 {
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
	overviewSource := ""
	if mm := javDBDescriptionMetaRe.FindStringSubmatch(content); len(mm) > 1 {
		overview = stripHTMLText(mm[1])
		overviewSource = "meta_description"
	}
	if isLikelyGenericOverview(overview) {
		overview = ""
		overviewSource = ""
	}
	if overview != "" {
		fieldState["overview"] = "filled"
	}

	actors := parseActorNamesByRegex(content, javBusActorNameRe)
	if len(actors) > 0 {
		fieldState["actors"] = "filled"
	}

	externalID := extractJavBusVideoID(detailURL)
	candidate := avScrapeCandidate{
		Source:      "javbus",
		ExternalID:  externalID,
		Code:        code,
		Title:       title,
		Overview:    overview,
		PosterURL:   posterURL,
		ReleaseDate: releaseDate,
		Actors:      actors,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"source":          "javbus",
			"detail_url":      detailURL,
			"external_id":     externalID,
			"code":            code,
			"title":           title,
			"overview":        overview,
			"overview_source": overviewSource,
			"poster_url":      posterURL,
			"poster_source":   posterSource,
			"release_date":    releaseDate,
			"actors":          actors,
		},
	}
	return candidate, fieldState
}

type javLibraryDetailParser struct{}

func (p *javLibraryDetailParser) Parse(content, detailURL, baseURL string) (avScrapeCandidate, map[string]string) {
	fieldState := defaultAVFieldState()

	title := ""
	if mm := javLibraryTitleRe.FindStringSubmatch(content); len(mm) > 1 {
		title = stripHTMLText(mm[1])
	}
	if title != "" {
		fieldState["title"] = "filled"
	}

	code := ""
	if mm := javLibraryCodeFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		code = extractAVCode(stripHTMLText(mm[1]))
	}
	if code == "" {
		code = extractAVCode(title)
	}
	if code == "" {
		code = extractAVCode(detailURL)
	}
	if code != "" {
		fieldState["code"] = "filled"
	}

	posterURL := ""
	posterSource := ""
	if mm := javLibraryCoverImgRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		posterSource = "video_jacket"
	}
	if isLikelyLogoURL(posterURL) {
		posterURL = ""
		posterSource = ""
	}
	if posterURL != "" {
		fieldState["poster_url"] = "filled"
	}

	releaseDate := ""
	if mm := javLibraryReleaseFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		releaseDate = normalizeAVDate(stripHTMLText(mm[1]))
	}
	if releaseDate != "" {
		fieldState["release_date"] = "filled"
	}

	overview := ""
	overviewSource := ""
	if mm := javDBDescriptionMetaRe.FindStringSubmatch(content); len(mm) > 1 {
		overview = stripHTMLText(mm[1])
		overviewSource = "meta_description"
	}
	if isLikelyGenericOverview(overview) {
		overview = ""
		overviewSource = ""
	}
	if overview != "" {
		fieldState["overview"] = "filled"
	}

	actors := parseActorNamesByRegex(content, javLibraryActorNameRe)
	if len(actors) > 0 {
		fieldState["actors"] = "filled"
	}

	externalID := extractJavLibraryVideoID(detailURL)
	candidate := avScrapeCandidate{
		Source:      "javlibrary",
		ExternalID:  externalID,
		Code:        code,
		Title:       title,
		Overview:    overview,
		PosterURL:   posterURL,
		ReleaseDate: releaseDate,
		Actors:      actors,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"source":          "javlibrary",
			"detail_url":      detailURL,
			"external_id":     externalID,
			"code":            code,
			"title":           title,
			"overview":        overview,
			"overview_source": overviewSource,
			"poster_url":      posterURL,
			"poster_source":   posterSource,
			"release_date":    releaseDate,
			"actors":          actors,
		},
	}
	return candidate, fieldState
}

type fc2DetailParser struct{}

func (p *fc2DetailParser) Parse(content, detailURL, baseURL string) (avScrapeCandidate, map[string]string) {
	fieldState := defaultAVFieldState()

	title := ""
	if mm := fc2TitleRe.FindStringSubmatch(content); len(mm) > 1 {
		title = stripHTMLText(mm[1])
	}
	if title == "" {
		if mm := javDBPageTitleRe.FindStringSubmatch(content); len(mm) > 1 {
			title = strings.TrimSpace(strings.TrimSuffix(stripHTMLText(mm[1]), "- FC2"))
		}
	}
	if title != "" {
		fieldState["title"] = "filled"
	}

	externalID := extractFC2VideoID(detailURL)
	code := ""
	if externalID != "" {
		code = "FC2-" + normalizeFC2NumericID(externalID)
	}
	if code == "" {
		code = extractAVCode(title)
	}
	if code != "" {
		fieldState["code"] = "filled"
	}

	posterURL := ""
	posterSource := ""
	if mm := fc2SampleCoverRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		posterSource = "sample_image"
	}
	if posterURL == "" {
		if mm := fc2MainThumbRe.FindStringSubmatch(content); len(mm) > 1 {
			posterURL = toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
			posterSource = "main_thumb"
		}
	}
	if isLikelyLogoURL(posterURL) {
		posterURL = ""
		posterSource = ""
	}
	if posterURL != "" {
		fieldState["poster_url"] = "filled"
	}

	releaseDate := ""
	if mm := fc2ReleaseFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		releaseDate = normalizeAVDate(stripHTMLText(mm[1]))
	}
	if releaseDate != "" {
		fieldState["release_date"] = "filled"
	}

	overview := ""
	overviewSource := ""
	if mm := fc2OverviewMetaRe.FindStringSubmatch(content); len(mm) > 1 {
		overview = stripHTMLText(mm[1])
		overviewSource = "meta_description"
	}
	if isLikelyGenericOverview(overview) {
		overview = ""
		overviewSource = ""
	}
	if overview != "" {
		fieldState["overview"] = "filled"
	}

	candidate := avScrapeCandidate{
		Source:      "fc2",
		ExternalID:  externalID,
		Code:        code,
		Title:       title,
		Overview:    overview,
		PosterURL:   posterURL,
		ReleaseDate: releaseDate,
		Actors:      nil,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"source":          "fc2",
			"detail_url":      detailURL,
			"external_id":     externalID,
			"code":            code,
			"title":           title,
			"overview":        overview,
			"overview_source": overviewSource,
			"poster_url":      posterURL,
			"poster_source":   posterSource,
			"release_date":    releaseDate,
			"actors":          []string{},
		},
	}
	return candidate, fieldState
}

func defaultAVFieldState() map[string]string {
	return map[string]string{
		"title":        "empty",
		"code":         "empty",
		"poster_url":   "empty",
		"overview":     "empty",
		"release_date": "empty",
		"actors":       "empty",
		"trailer":      "not_supported",
	}
}

func parseJavDBDetailActorsStrict(content string) []string {
	out := make([]string, 0, 8)
	seen := map[string]struct{}{}
	appendName := func(raw string) {
		name := strings.Join(strings.Fields(stripHTMLText(raw)), " ")
		if !isLikelyActorName(name) {
			return
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}

	if mm := javDBActorFieldBlockRe.FindStringSubmatch(content); len(mm) > 1 {
		for _, actorMM := range javDBInlineActorNameRe.FindAllStringSubmatch(mm[1], -1) {
			if len(actorMM) > 1 {
				appendName(actorMM[1])
			}
		}
	}
	if len(out) > 0 {
		return out
	}

	if mm := javDBFemaleActorNameRe.FindStringSubmatch(content); len(mm) > 1 {
		for _, actorMM := range javDBInlineActorNameRe.FindAllStringSubmatch(mm[1], -1) {
			if len(actorMM) > 1 {
				appendName(actorMM[1])
			}
		}
	}
	if len(out) > 0 {
		return out
	}

	matches := javDBActorAnchorRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		appendName(match[2])
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseActorNamesByRegex(content string, re *regexp.Regexp) []string {
	if re == nil {
		return nil
	}
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := strings.Join(strings.Fields(stripHTMLText(match[1])), " ")
		if !isLikelyActorName(name) {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isLikelyActorName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	if len([]rune(name)) > 48 {
		return false
	}
	low := strings.ToLower(name)
	noise := []string{"more", "更多", "全部", "查看", "profile", "详情", "演員", "演员", "actress", "actor"}
	for _, n := range noise {
		if low == n {
			return false
		}
	}
	return true
}

func extractJavDBPosterURL(content, baseURL string) (string, string) {
	if mm := javDBVideoCoverImgRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL := toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		if !isLikelyLogoURL(posterURL) {
			return posterURL, "video_cover"
		}
	}
	if mm := javDBOGImageRe.FindStringSubmatch(content); len(mm) > 1 {
		posterURL := toAbsoluteURL(baseURL, strings.TrimSpace(mm[1]))
		if !isLikelyLogoURL(posterURL) {
			return posterURL, "og_image"
		}
	}
	return "", ""
}

func extractJavDBOverview(content string) (string, string) {
	if mm := javDBOverviewFieldRe.FindStringSubmatch(content); len(mm) > 1 {
		overview := stripHTMLText(mm[1])
		if !isLikelyGenericOverview(overview) {
			return overview, "detail_field"
		}
	}
	if mm := javDBOverviewBlockRe.FindStringSubmatch(content); len(mm) > 1 {
		overview := stripHTMLText(mm[1])
		if !isLikelyGenericOverview(overview) {
			return overview, "detail_block"
		}
	}
	if mm := javDBDescriptionMetaRe.FindStringSubmatch(content); len(mm) > 1 {
		overview := stripHTMLText(mm[1])
		if !isLikelyGenericOverview(overview) {
			return overview, "meta_description"
		}
	}
	return "", ""
}

func isLikelyLogoURL(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return false
	}
	logoTokens := []string{"logo", "favicon", "icon", "navbar", "brand", "avatar", "banner", "noimage", "placeholder"}
	for _, token := range logoTokens {
		if strings.Contains(v, token) {
			return true
		}
	}
	return false
}

const (
	avPosterQualityPrimary  = "primary"
	avPosterQualityFallback = "fallback"
	avPosterQualityInvalid  = "invalid"
)

func avPosterSourceFromCandidate(candidate avScrapeCandidate) string {
	if candidate.Raw == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(asString(candidate.Raw["poster_source"])))
}

func classifyAVPosterURL(posterURL, posterSource string) string {
	posterURL = strings.TrimSpace(posterURL)
	if posterURL == "" {
		return avPosterQualityInvalid
	}
	if isLikelyLogoURL(posterURL) {
		return avPosterQualityInvalid
	}

	source := strings.ToLower(strings.TrimSpace(posterSource))
	switch source {
	case "video_cover", "big_image", "video_jacket", "sample_image":
		return avPosterQualityPrimary
	case "og_image", "meta_og", "main_thumb", "cover_img":
		return avPosterQualityFallback
	}

	lower := strings.ToLower(posterURL)
	if strings.Contains(lower, "/thumb") || strings.Contains(lower, "sample") || strings.Contains(lower, "preview") {
		return avPosterQualityFallback
	}
	if strings.Contains(lower, "/cover") || strings.Contains(lower, "video-cover") || strings.Contains(lower, "jacket") {
		return avPosterQualityPrimary
	}
	return avPosterQualityPrimary
}

func candidatePosterQuality(candidate avScrapeCandidate) string {
	if candidate.Raw != nil {
		if quality := strings.ToLower(strings.TrimSpace(asString(candidate.Raw["poster_quality"]))); quality != "" {
			switch quality {
			case avPosterQualityPrimary, avPosterQualityFallback, avPosterQualityInvalid:
				return quality
			}
		}
	}
	return classifyAVPosterURL(candidate.PosterURL, avPosterSourceFromCandidate(candidate))
}

func enrichAVCandidatePoster(candidate *avScrapeCandidate) {
	if candidate == nil {
		return
	}
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	source := avPosterSourceFromCandidate(*candidate)
	quality := classifyAVPosterURL(candidate.PosterURL, source)
	candidate.Raw["poster_source"] = source
	candidate.Raw["poster_quality"] = quality
	if quality == avPosterQualityInvalid {
		candidate.PosterURL = ""
		candidate.Raw["poster_url"] = ""
		return
	}
	candidate.Raw["poster_url"] = strings.TrimSpace(candidate.PosterURL)
}

func isLikelyGenericOverview(raw string) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return false
	}
	if len([]rune(text)) < 4 {
		return true
	}
	genericTokens := []string{"javdb", "javbus", "javlibrary", "adult.contents.fc2", "在线看", "在线观看", "線上看", "番号", "magnet", "torrent", "下载", "下載"}
	count := 0
	for _, token := range genericTokens {
		if strings.Contains(text, token) {
			count++
		}
	}
	return count >= 2
}

func extractJavBusVideoID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(parsed.Path)
	if path == "" || path == "/" {
		return ""
	}
	if mm := javBusExternalIDPathRe.FindStringSubmatch(path); len(mm) > 1 {
		return strings.TrimSpace(mm[1])
	}
	return ""
}

func extractJavLibraryVideoID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	if id := strings.TrimSpace(parsed.Query().Get("v")); id != "" {
		return id
	}
	return ""
}

func normalizeFC2NumericID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if mm := fc2NumberRe.FindStringSubmatch(raw); len(mm) > 1 {
		num := strings.TrimLeft(strings.TrimSpace(mm[1]), "0")
		if num == "" {
			num = "0"
		}
		return num
	}
	if mm := fc2ExternalIDPathRe.FindStringSubmatch(raw); len(mm) > 1 {
		num := strings.TrimLeft(strings.TrimSpace(mm[1]), "0")
		if num == "" {
			num = "0"
		}
		return num
	}
	return ""
}

func extractFC2VideoID(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	if mm := fc2ExternalIDPathRe.FindStringSubmatch(rawURL); len(mm) > 1 {
		return normalizeFC2NumericID(mm[1])
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return normalizeFC2NumericID(rawURL)
	}
	if mm := fc2ExternalIDPathRe.FindStringSubmatch(parsed.Path); len(mm) > 1 {
		return normalizeFC2NumericID(mm[1])
	}
	return normalizeFC2NumericID(rawURL)
}

func normalizeAVSourceName(source string) string {
	source = strings.ToLower(strings.TrimSpace(source))
	source = strings.ReplaceAll(source, "-", "")
	source = strings.ReplaceAll(source, "_", "")
	switch source {
	case "javdb":
		return "javdb"
	case "airavcc":
		return "airav_cc"
	case "mdtv", "mdtv.com", "mdtvcom":
		return "mdtv.com"
	case "javbus":
		return "javbus"
	case "javlibrary", "javlib":
		return "javlibrary"
	case "fc2", "fc2ppv", "adultfc2":
		return "fc2"
	case "theporndb", "porndb", "tpdb":
		return "theporndb"
	default:
		return source
	}
}

func avCandidateIdentityKey(candidate avScrapeCandidate) string {
	source := normalizeAVSourceName(candidate.Source)
	externalID := strings.ToLower(strings.TrimSpace(candidate.ExternalID))
	detailURL := strings.ToLower(strings.TrimSpace(candidate.DetailURL))
	if source == "" {
		source = "unknown"
	}
	if externalID != "" {
		return source + "|id|" + externalID
	}
	if detailURL != "" {
		return source + "|url|" + detailURL
	}
	if code := normalizeAVCodeForCompare(candidate.Code); code != "" {
		return source + "|code|" + code
	}
	return ""
}

func collectAVCandidateSources(candidates []avScrapeCandidate) []string {
	if len(candidates) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		source := normalizeAVSourceName(candidate.Source)
		if source == "" {
			continue
		}
		if _, ok := seen[source]; ok {
			continue
		}
		seen[source] = struct{}{}
		out = append(out, source)
	}
	sort.Strings(out)
	return out
}

func mergeAVCandidatesByField(candidates []avScrapeCandidate, targetCode, keyword string) (avScrapeCandidate, bool) {
	if len(candidates) == 0 {
		return avScrapeCandidate{}, false
	}

	groups := map[string][]avScrapeCandidate{}
	groupOrder := make([]string, 0, len(candidates))
	for i, candidate := range candidates {
		candidate.Source = normalizeAVSourceName(candidate.Source)
		key := avCandidateGroupKey(candidate)
		if key == "" {
			key = fmt.Sprintf("fallback:%d", i)
		}
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], candidate)
	}
	if len(groupOrder) == 0 {
		return avScrapeCandidate{}, false
	}

	bestKey := ""
	bestGroupScore := -1
	for _, key := range groupOrder {
		score := scoreAVCandidateGroup(groups[key], targetCode, keyword)
		if score > bestGroupScore {
			bestGroupScore = score
			bestKey = key
		}
	}
	selected := groups[bestKey]
	if len(selected) == 0 {
		return avScrapeCandidate{}, false
	}

	primary := pickPrimaryAVCandidate(selected, targetCode, keyword)
	merged := primary
	fieldSources := map[string]string{}

	title, titleSource := selectBestTitle(selected, targetCode, keyword)
	if title != "" {
		merged.Title = title
		fieldSources["title"] = titleSource
	}

	code, codeSource := selectBestCode(selected, targetCode)
	if code != "" {
		merged.Code = code
		fieldSources["code"] = codeSource
	}
	if merged.Code == "" {
		merged.Code = extractAVCode(chooseStr(merged.Title, merged.DetailURL))
	}

	posterURL, posterSource, posterRawSource, posterQuality := selectBestPosterURL(selected, targetCode, keyword)
	posterDecision := ""
	if posterURL != "" {
		merged.PosterURL = posterURL
		fieldSources["poster_url"] = posterSource
		switch posterQuality {
		case avPosterQualityPrimary:
			posterDecision = "primary_selected"
		case avPosterQualityFallback:
			posterDecision = "fallback_selected"
		default:
			posterDecision = "invalid_selected"
		}
	}

	overview, overviewSource := selectBestOverview(selected)
	if overview != "" {
		merged.Overview = overview
		fieldSources["overview"] = overviewSource
	}

	releaseDate, releaseSource := selectBestReleaseDate(selected, targetCode, keyword)
	if releaseDate != "" {
		merged.ReleaseDate = releaseDate
		fieldSources["release_date"] = releaseSource
	}

	actors, actorSource := mergeBestActors(selected)
	if len(actors) > 0 {
		merged.Actors = actors
		fieldSources["actors"] = actorSource
	}

	if strings.TrimSpace(merged.ExternalID) == "" {
		for _, candidate := range selected {
			if strings.TrimSpace(candidate.ExternalID) != "" {
				merged.ExternalID = strings.TrimSpace(candidate.ExternalID)
				break
			}
		}
	}
	if strings.TrimSpace(merged.DetailURL) == "" {
		for _, candidate := range selected {
			if strings.TrimSpace(candidate.DetailURL) != "" {
				merged.DetailURL = strings.TrimSpace(candidate.DetailURL)
				break
			}
		}
	}
	if strings.TrimSpace(merged.Source) == "" {
		merged.Source = "javdb"
	}

	rawCandidates := make([]map[string]any, 0, len(selected))
	for _, candidate := range selected {
		rawCandidates = append(rawCandidates, map[string]any{
			"source":       candidate.Source,
			"external_id":  candidate.ExternalID,
			"detail_url":   candidate.DetailURL,
			"code":         candidate.Code,
			"title":        candidate.Title,
			"overview":     candidate.Overview,
			"poster_url":   candidate.PosterURL,
			"release_date": candidate.ReleaseDate,
			"actors":       append([]string(nil), candidate.Actors...),
			"metadata":     candidate.Raw,
		})
	}

	merged.Raw = map[string]any{
		"source":          merged.Source,
		"detail_url":      merged.DetailURL,
		"external_id":     merged.ExternalID,
		"code":            merged.Code,
		"title":           merged.Title,
		"overview":        merged.Overview,
		"poster_url":      merged.PosterURL,
		"release_date":    merged.ReleaseDate,
		"actors":          append([]string(nil), merged.Actors...),
		"field_sources":   fieldSources,
		"merged":          true,
		"merged_sources":  collectAVCandidateSources(selected),
		"poster_source":   posterRawSource,
		"poster_quality":  posterQuality,
		"poster_decision": posterDecision,
		"raw_candidates":  rawCandidates,
	}
	return merged, true
}

func avCandidateGroupKey(candidate avScrapeCandidate) string {
	if code := normalizeAVCodeForCompare(candidate.Code); code != "" {
		return "code:" + code
	}
	if code := normalizeAVCodeForCompare(extractAVCode(candidate.Title)); code != "" {
		return "code:" + code
	}
	if code := normalizeAVCodeForCompare(extractAVCode(candidate.DetailURL)); code != "" {
		return "code:" + code
	}
	title := normalizeAVCodeForCompare(candidate.Title)
	if title != "" {
		return "title:" + title
	}
	if detail := strings.ToLower(strings.TrimSpace(candidate.DetailURL)); detail != "" {
		return "url:" + detail
	}
	return ""
}

func scoreAVCandidateGroup(group []avScrapeCandidate, targetCode, keyword string) int {
	if len(group) == 0 {
		return -1
	}
	best := -1
	total := 0
	hasPoster := false
	hasActors := false
	for _, candidate := range group {
		score, _ := scoreAVCandidate(candidate, targetCode, keyword)
		if score > best {
			best = score
		}
		total += score
		if strings.TrimSpace(candidate.PosterURL) != "" && candidatePosterQuality(candidate) != avPosterQualityInvalid {
			hasPoster = true
		}
		if len(candidate.Actors) > 0 {
			hasActors = true
		}
	}
	groupScore := best*3 + total + len(group)*80
	if hasPoster {
		groupScore += 80
	}
	if hasActors {
		groupScore += 40
	}
	return groupScore
}

func pickPrimaryAVCandidate(group []avScrapeCandidate, targetCode, keyword string) avScrapeCandidate {
	primary := group[0]
	best := -1
	for _, candidate := range group {
		score, _ := scoreAVCandidate(candidate, targetCode, keyword)
		if strings.TrimSpace(candidate.DetailURL) != "" {
			score += 20
		}
		if strings.TrimSpace(candidate.ExternalID) != "" {
			score += 15
		}
		if score > best {
			best = score
			primary = candidate
		}
	}
	return primary
}

func selectBestTitle(candidates []avScrapeCandidate, targetCode, keyword string) (string, string) {
	bestValue := ""
	bestSource := ""
	bestScore := -1
	for _, candidate := range candidates {
		title := strings.TrimSpace(candidate.Title)
		if title == "" {
			continue
		}
		score, _ := scoreAVCandidate(candidate, targetCode, keyword)
		score += minInt(len([]rune(title)), 80)
		if strings.TrimSpace(candidate.Code) != "" && strings.Contains(normalizeAVCodeForCompare(title), normalizeAVCodeForCompare(candidate.Code)) {
			score += 35
		}
		if score > bestScore {
			bestScore = score
			bestValue = title
			bestSource = candidate.Source
		}
	}
	return bestValue, bestSource
}

func selectBestCode(candidates []avScrapeCandidate, targetCode string) (string, string) {
	bestValue := ""
	bestSource := ""
	bestScore := -1
	targetNorm := normalizeAVCodeForCompare(targetCode)
	for _, candidate := range candidates {
		code := strings.TrimSpace(candidate.Code)
		if code == "" {
			code = extractAVCode(candidate.Title)
		}
		if code == "" {
			continue
		}
		score := 100
		codeNorm := normalizeAVCodeForCompare(code)
		if targetNorm != "" {
			if codeNorm == targetNorm {
				score += 220
			}
			if strings.Contains(codeNorm, targetNorm) || strings.Contains(targetNorm, codeNorm) {
				score += 120
			}
		}
		if strings.TrimSpace(candidate.Title) != "" && strings.Contains(normalizeAVCodeForCompare(candidate.Title), codeNorm) {
			score += 40
		}
		if score > bestScore {
			bestScore = score
			bestValue = code
			bestSource = candidate.Source
		}
	}
	return bestValue, bestSource
}

func selectBestPosterURL(candidates []avScrapeCandidate, targetCode, keyword string) (string, string, string, string) {
	bestPrimaryValue := ""
	bestPrimarySource := ""
	bestPrimaryRawSource := ""
	bestPrimaryScore := -1
	bestFallbackValue := ""
	bestFallbackSource := ""
	bestFallbackRawSource := ""
	bestFallbackScore := -1
	for _, candidate := range candidates {
		posterURL := strings.TrimSpace(candidate.PosterURL)
		if posterURL == "" {
			continue
		}
		quality := candidatePosterQuality(candidate)
		if quality == avPosterQualityInvalid {
			continue
		}
		score, _ := scoreAVCandidate(candidate, targetCode, keyword)
		if strings.TrimSpace(candidate.DetailURL) != "" {
			score += 20
		}
		rawSource := avPosterSourceFromCandidate(candidate)
		if quality == avPosterQualityPrimary {
			score += 120
			if score > bestPrimaryScore {
				bestPrimaryScore = score
				bestPrimaryValue = posterURL
				bestPrimarySource = candidate.Source
				bestPrimaryRawSource = rawSource
			}
			continue
		}
		score += 60
		if score > bestFallbackScore {
			bestFallbackScore = score
			bestFallbackValue = posterURL
			bestFallbackSource = candidate.Source
			bestFallbackRawSource = rawSource
		}
	}
	if bestPrimaryValue != "" {
		return bestPrimaryValue, bestPrimarySource, bestPrimaryRawSource, avPosterQualityPrimary
	}
	if bestFallbackValue != "" {
		return bestFallbackValue, bestFallbackSource, bestFallbackRawSource, avPosterQualityFallback
	}
	return "", "", "", avPosterQualityInvalid
}

func selectBestOverview(candidates []avScrapeCandidate) (string, string) {
	bestValue := ""
	bestSource := ""
	bestScore := -1
	for _, candidate := range candidates {
		overview := strings.TrimSpace(candidate.Overview)
		if overview == "" || isLikelyGenericOverview(overview) {
			continue
		}
		score := minInt(len([]rune(overview)), 220)
		if rawOverviewSource, ok := candidate.Raw["overview_source"].(string); ok {
			rawOverviewSource = strings.ToLower(strings.TrimSpace(rawOverviewSource))
			switch rawOverviewSource {
			case "detail_field", "detail_block":
				score += 120
			case "meta_description":
				score += 20
			}
		}
		if score > bestScore {
			bestScore = score
			bestValue = overview
			bestSource = candidate.Source
		}
	}
	return bestValue, bestSource
}

func selectBestReleaseDate(candidates []avScrapeCandidate, targetCode, keyword string) (string, string) {
	bestValue := ""
	bestSource := ""
	bestScore := -1
	for _, candidate := range candidates {
		releaseDate := normalizeAVDate(candidate.ReleaseDate)
		if releaseDate == "" {
			continue
		}
		score, _ := scoreAVCandidate(candidate, targetCode, keyword)
		score += 50
		if score > bestScore {
			bestScore = score
			bestValue = releaseDate
			bestSource = candidate.Source
		}
	}
	return bestValue, bestSource
}

func mergeBestActors(candidates []avScrapeCandidate) ([]string, string) {
	base := []string{}
	baseSource := ""
	baseScore := -1
	for _, candidate := range candidates {
		filtered := filterActorNames(candidate.Actors)
		score := len(filtered)
		if score > baseScore {
			baseScore = score
			base = filtered
			baseSource = candidate.Source
		}
	}
	if len(base) == 0 {
		return nil, ""
	}
	merged := append([]string(nil), base...)
	seen := map[string]struct{}{}
	for _, actor := range merged {
		seen[strings.ToLower(actor)] = struct{}{}
	}
	for _, candidate := range candidates {
		if candidate.Source == baseSource {
			continue
		}
		for _, actor := range filterActorNames(candidate.Actors) {
			key := strings.ToLower(actor)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, actor)
		}
	}
	return merged, baseSource
}

func filterActorNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, raw := range names {
		name := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
		if !isLikelyActorName(name) {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

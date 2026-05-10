package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

type dmmAVCrawler struct {
	svc *ScraperService
}

type mgstageAVCrawler struct {
	svc *ScraperService
}

type prestigeAVCrawler struct {
	svc *ScraperService
}

type xcityAVCrawler struct {
	svc *ScraperService
}

type getchuAVCrawler struct {
	svc *ScraperService
}

type thePornDBAVCrawler struct {
	svc *ScraperService
}

var (
	dmmNumberPartsPattern = regexp.MustCompile(`(?i)(\d*[a-z]+)-?(\d+)`)
	dmmDetailURLPattern   = regexp.MustCompile(`(?is)detailUrl\\":\\"(.*?)\\"`)
	datedNumberPattern    = regexp.MustCompile(`(?i)(([A-Z0-9.-]{2,})[-_. ]2?0?(\d{2}[-.]\d{2}[-.]\d{2}))`)
)

const (
	dmmCategoryFanzaTV = "fanza_tv"
	dmmCategoryDMMTV   = "dmm_tv"
	dmmCategoryDigital = "digital"
	dmmCategoryMono    = "mono"
	dmmCategoryRental  = "rental"
	dmmCategoryOther   = "other"
)

func newDMMAVCrawler(svc *ScraperService) avCrawler       { return &dmmAVCrawler{svc: svc} }
func newMGStageAVCrawler(svc *ScraperService) avCrawler   { return &mgstageAVCrawler{svc: svc} }
func newPrestigeAVCrawler(svc *ScraperService) avCrawler  { return &prestigeAVCrawler{svc: svc} }
func newXCityAVCrawler(svc *ScraperService) avCrawler     { return &xcityAVCrawler{svc: svc} }
func newGetchuAVCrawler(svc *ScraperService) avCrawler    { return &getchuAVCrawler{svc: svc} }
func newThePornDBAVCrawler(svc *ScraperService) avCrawler { return &thePornDBAVCrawler{svc: svc} }

func (c *dmmAVCrawler) Name() string      { return "dmm" }
func (c *mgstageAVCrawler) Name() string  { return "mgstage" }
func (c *prestigeAVCrawler) Name() string { return "prestige" }
func (c *xcityAVCrawler) Name() string    { return "xcity" }
func (c *getchuAVCrawler) Name() string   { return "getchu" }
func (c *thePornDBAVCrawler) Name() string {
	return "theporndb"
}

func (c *dmmAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	base := strings.TrimRight(c.svc.avSiteBaseURL("dmm", "https://www.dmm.co.jp"), "/")
	searchURLs := dmmSearchURLs(base, query)
	if len(searchURLs) == 0 {
		return nil, nil
	}
	var regionErr error
	var firstErr error
	for _, searchURL := range searchURLs {
		if run != nil {
			run.addSearchURL(searchURL)
		}
		content, err := c.fetchDMMHTMLDocumentText(ctx, searchURL, map[string]string{
			"Cookie": "age_check_done=1",
		})
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if dmmIsRegionBlocked(content) {
			regionErr = fmt.Errorf("dmm: content is not available in this region")
			continue
		}
		hits := dmmParseSearchCandidates(content, query, base)
		if len(hits) == 0 {
			continue
		}
		for _, hit := range hits {
			candidate, err := c.FetchByDetailURL(ctx, run, hit.URL)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				if strings.Contains(err.Error(), "content is not available in this region") {
					regionErr = err
				}
				continue
			}
			if candidate.DetailURL == "" {
				candidate.DetailURL = hit.URL
			}
			if candidate.Raw == nil {
				candidate.Raw = map[string]any{}
			}
			if candidate.Raw["category"] == nil {
				candidate.Raw["category"] = hit.Category
			}
			return []avScrapeCandidate{candidate}, nil
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if regionErr != nil {
		return nil, regionErr
	}
	return nil, fmt.Errorf("dmm: no matched detail page for %s", query)
}

func (c *mgstageAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	base := strings.TrimRight(c.svc.avSiteBaseURL("mgstage", "https://www.mgstage.com"), "/")
	detailURL := base + "/product/product_detail/" + url.PathEscape(strings.ToUpper(query)) + "/"
	if run != nil {
		run.addSearchURL(detailURL)
	}
	candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		return nil, err
	}
	return []avScrapeCandidate{candidate}, nil
}

func (c *mgstageAVCrawler) sessionCookies(ctx context.Context, detailURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/122 Safari/537.36")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8,zh-CN;q=0.7")
	client := c.svc.avHTTPClient
	if client == nil {
		client = c.svc.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", detailURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch %s: status %d", detailURL, resp.StatusCode)
	}
	cookies := make([]string, 0, len(resp.Cookies()))
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "" && cookie.Value != "" {
			cookies = append(cookies, cookie.Name+"="+cookie.Value)
		}
	}
	return cookies, nil
}

func (c *prestigeAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	return nil, nil
}

func (c *xcityAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	return nil, nil
}

func (c *getchuAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	base := strings.TrimRight(c.svc.avSiteBaseURL("getchu", "https://www.getchu.com"), "/")
	searchURL := fmt.Sprintf("%s/php/search.phtml?genre=all&search_keyword=%s&gc=gc", base, url.QueryEscape(query))
	if run != nil {
		run.addSearchURL(searchURL)
	}
	content, err := c.svc.fetchAVHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse getchu search html: %w", err)
	}
	link := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "blueb")
	})
	href := strings.TrimSpace(attrValue(link, "href"))
	if href == "" {
		return nil, nil
	}
	detailURL := toAbsoluteURL(base+"/", strings.Replace(href, "../", "/", 1))
	if detailURL == "" {
		return nil, nil
	}
	if !strings.Contains(detailURL, "gc=gc") {
		if strings.Contains(detailURL, "?") {
			detailURL += "&gc=gc"
		} else {
			detailURL += "?gc=gc"
		}
	}
	candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		return nil, err
	}
	return []avScrapeCandidate{candidate}, nil
}

func (c *dmmAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	detailURL = strings.TrimSpace(detailURL)
	if detailURL == "" {
		return avScrapeCandidate{}, fmt.Errorf("detail url is required")
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	category := dmmParseCategory(detailURL)
	if category == dmmCategoryFanzaTV || category == dmmCategoryDMMTV {
		return c.fetchDMMFanzaTV(ctx, detailURL)
	}
	root, body, err := c.fetchDMMHTMLDocument(ctx, detailURL, map[string]string{
		"Cookie": "age_check_done=1",
	})
	if err != nil {
		return avScrapeCandidate{}, err
	}
	metadata, err := dmmParseDigitalDetail(root, body, detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	metadata.Source = c.Name()
	metadata.DetailURL = detailURL
	return metadata, nil
}

func (c *dmmAVCrawler) fetchDMMHTMLDocument(ctx context.Context, endpoint string, extraHeaders map[string]string) (*html.Node, string, error) {
	body, err := c.fetchDMMHTMLDocumentText(ctx, endpoint, extraHeaders)
	if err != nil {
		return nil, "", err
	}
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("解析 DMM 响应失败: %w", err)
	}
	return root, body, nil
}

func (c *dmmAVCrawler) fetchDMMHTMLDocumentText(ctx context.Context, endpoint string, extraHeaders map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("创建 DMM 刮削请求失败: %w", err)
	}
	if strings.TrimSpace(c.svc.avUserAgent) != "" {
		req.Header.Set("User-Agent", c.svc.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	client := c.dmmHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("DMM 站点请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("DMM 站点请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("读取 DMM 响应失败: %w", err)
	}
	return string(body), nil
}

func (c *dmmAVCrawler) dmmHTTPClient() *http.Client {
	client := c.svc.avHTTPClient
	if client == nil {
		client = c.svc.httpClient
	}
	if client == nil {
		return http.DefaultClient
	}
	cloned := *client
	cloned.Timeout = 0
	return &cloned
}

func (c *dmmAVCrawler) postDMMJSON(ctx context.Context, endpoint string, body any, extraHeaders map[string]string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("编码 DMM JSON 请求失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("创建 DMM JSON POST 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.svc.avUserAgent) != "" {
		req.Header.Set("User-Agent", c.svc.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := c.dmmHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("DMM JSON POST 请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("DMM JSON POST 请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析 DMM JSON POST 响应失败: %w", err)
	}
	return nil
}

func (c *mgstageAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	detailURL = strings.TrimSpace(detailURL)
	if detailURL == "" {
		return avScrapeCandidate{}, fmt.Errorf("detail url is required")
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	cookies, err := c.sessionCookies(ctx, detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	root, _, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, map[string]string{
		"Cookie": cookieHeader(cookies, "adc=1"),
	})
	if err != nil {
		return avScrapeCandidate{}, err
	}
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1"
	})))
	if title == "" {
		title = strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && attrValue(n, "id") == "center_column"
		})))
	}
	code := strings.TrimSpace(avTableValueAny(root, "Number", "品番"))
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	poster := strings.TrimSpace(findAttrByID(root, "a", "EnlargeImage", "href"))
	poster = toAbsoluteURL(detailURL, poster)
	thumb := poster
	trailerURL := ""
	if review := strings.TrimSpace(findAttrByID(root, "a", "review-btn", "href")); review != "" {
		trailerURL = toAbsoluteURL(detailURL, strings.Replace(review, "/mypage/review.php", "/sampleplayer/sampleRespons.php", 1))
	}
	actors := avTableLinksAny(root, "出演", "演员", "Actress", "Actor")
	tags := avTableLinksAny(root, "ジャンル")
	releaseDate := normalizeSlashDate(avTableValueAny(root, "Release", "配信開始日", "発売日"))
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  strings.ToLower(extractSingleSitePathID(detailURL)),
		Code:        code,
		Title:       title,
		Overview:    strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && attrValue(n, "id") == "introduction" }))),
		PosterURL:   strings.Replace(poster, "/pb_", "/pf_", 1),
		ThumbURL:    thumb,
		ReleaseDate: releaseDate,
		Actors:      actors,
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": strings.ToLower(extractSingleSitePathID(detailURL)),
			"av_code":     code,
			"title":       title,
			"poster_url":  strings.Replace(poster, "/pb_", "/pf_", 1),
			"thumb_url":   thumb,
			"trailer_url": trailerURL,
			"tags":        tags,
		},
	}
	if trailerURL != "" {
		candidate.Raw["trailer_url"] = trailerURL
	}
	return candidate, nil
}

func cookieHeader(values []string, extra ...string) string {
	all := make([]string, 0, len(values)+len(extra))
	all = append(all, values...)
	all = append(all, extra...)
	return strings.Join(all, "; ")
}

func (c *prestigeAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	apiURL, externalID, err := prestigeNormalizeURL(detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(apiURL)
	}
	var payload prestigeAVPayload
	if err := c.svc.fetchAVJSON(ctx, apiURL, map[string]string{
		"Accept": "application/json",
	}, &payload); err != nil {
		return avScrapeCandidate{}, err
	}
	title := normalizeWhitespace(payload.Title)
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  externalID,
		Code:        prestigeExtractSKU(detailURL),
		Title:       title,
		Overview:    normalizeWhitespace(payload.Body),
		PosterURL:   prestigeMediaURL(payload.Thumbnail.Path),
		ReleaseDate: normalizeSlashDate(payload.salesStartAt()),
		Actors:      payload.actorNames(),
		DetailURL:   apiURL,
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  apiURL,
			"external_id": externalID,
			"av_code":     prestigeExtractSKU(detailURL),
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *xcityAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	root, _, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, nil)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	code := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && attrValue(n, "id") == "hinban"
	})))
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && attrValue(n, "id") == "program_detail_title"
	})))
	title = strings.TrimSpace(strings.TrimSuffix(title, code))
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	poster := strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img" && hasClass(n, "packageThumb")
	}, "src"))
	if strings.HasPrefix(poster, "//") {
		poster = "https:" + poster
	}
	poster = strings.Replace(poster, "/package/medium/poster/", "/poster/", 1)
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  xcityQueryID(detailURL),
		Code:        code,
		Title:       title,
		Overview:    strings.TrimSpace(findNodeTextOnClass(root, "p", "lead")),
		PosterURL:   poster,
		ReleaseDate: normalizeSlashDate(xcityListValue(root, "Release")),
		Actors:      xcityTextsFromContainer(root, "li", "credit-links", "a"),
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": xcityQueryID(detailURL),
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *getchuAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	root, body, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, map[string]string{
		"Referer": "http://www.getchu.com/top.html",
	})
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1" && attrValue(n, "id") == "soft-title"
	})))
	if title == "" {
		title = strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h1"
		})))
	}
	if title == "" {
		title = stripHTMLText(extractRegexGroup(body, `(?is)<h1[^>]*id=["']soft-title["'][^>]*>(.*?)</h1>`))
	}
	if title == "" {
		title = stripHTMLText(extractRegexGroup(body, `(?is)<h1[^>]*>(.*?)</h1>`))
	}
	code := strings.TrimSpace(getchuTableValue(root, "Item Code"))
	if code == "" {
		code = stripHTMLText(extractRegexGroup(body, `(?is)Item Code</td>\s*<td[^>]*>(.*?)</td>`))
	}
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("getchu empty scrape result for url %q", detailURL)
	}
	poster := toAbsoluteURL(detailURL, findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && strings.EqualFold(attrValue(n, "property"), "og:image")
	}, "content"))
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  normalizeWhitespace(getchuQueryValue(detailURL, "id")),
		Code:        code,
		Title:       title,
		Overview:    getchuOutline(root),
		PosterURL:   poster,
		ReleaseDate: normalizeSlashDate(getchuTableValue(root, "Release Date")),
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": normalizeWhitespace(getchuQueryValue(detailURL, "id")),
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *thePornDBAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	if strings.TrimSpace(c.svc.avThePornDBAPIToken) == "" {
		return nil, fmt.Errorf("theporndb: THEPORNDB_API_TOKEN is required")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	base := strings.TrimRight(c.svc.avSiteBaseURL("theporndb", "https://api.theporndb.net"), "/")
	keywords, series, date := thePornDBSearchKeywords(query)
	searchURLs := make([]string, 0, len(keywords)*2)
	for _, category := range []string{"scenes", "movies"} {
		for _, keyword := range keywords {
			searchURL := thePornDBSearchURL(base, category, keyword)
			searchURLs = append(searchURLs, searchURL)
			if run != nil {
				run.addSearchURL(searchURL)
			}
			var response thePornDBSearchListResponse
			if err := c.svc.fetchAVJSON(ctx, searchURL, map[string]string{
				"Authorization": "Bearer " + c.svc.avThePornDBAPIToken,
				"Accept":        "application/json",
			}, &response); err != nil {
				return nil, err
			}
			detailURL := thePornDBBestMatchURL(response.Data, query, series, date, category, base)
			if detailURL == "" {
				continue
			}
			candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
			if err != nil {
				return nil, err
			}
			if candidate.Title == "" {
				continue
			}
			return []avScrapeCandidate{candidate}, nil
		}
	}
	return nil, fmt.Errorf("theporndb: no matched detail page for %s", query)
}

func (c *thePornDBAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	if strings.TrimSpace(c.svc.avThePornDBAPIToken) == "" {
		return avScrapeCandidate{}, fmt.Errorf("theporndb: THEPORNDB_API_TOKEN is required")
	}
	normalized, err := normalizeThePornDBDetailURL(detailURL, c.svc.avSiteBaseURL("theporndb", "https://api.theporndb.net"))
	if err != nil {
		return avScrapeCandidate{}, err
	}
	detailURL = normalized
	if run != nil {
		run.addDetailURL(detailURL)
	}
	var payload thePornDBItemResponse
	if err := c.svc.fetchAVJSON(ctx, detailURL, map[string]string{
		"Authorization": "Bearer " + c.svc.avThePornDBAPIToken,
		"Accept":        "application/json",
	}, &payload); err != nil {
		return avScrapeCandidate{}, err
	}
	candidate := thePornDBMetadataFromItem(payload.Data, categoryFromThePornDBDetailURL(detailURL))
	if candidate.Title == "" {
		return avScrapeCandidate{}, fmt.Errorf("theporndb: empty detail data")
	}
	candidate.Source = c.Name()
	candidate.DetailURL = detailURL
	if candidate.Raw == nil {
		candidate.Raw = map[string]any{}
	}
	candidate.Raw["site"] = c.Name()
	candidate.Raw["detail_url"] = detailURL
	candidate.Raw["external_id"] = candidate.ExternalID
	return candidate, nil
}

func normalizeThePornDBDetailURL(rawURL string, fallbackBase ...string) (string, error) {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return "", fmt.Errorf("theporndb: detailUrl is empty")
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("theporndb: parse detailUrl: %w", err)
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || (parts[0] != "scenes" && parts[0] != "movies") || parts[1] == "" {
		return "", fmt.Errorf("theporndb: detailUrl must point to scenes/<slug> or movies/<slug>")
	}
	if host == "theporndb.net" || host == "api.theporndb.net" {
		base := "https://api.theporndb.net"
		if len(fallbackBase) > 0 && strings.TrimSpace(fallbackBase[0]) != "" {
			base = strings.TrimRight(strings.TrimSpace(fallbackBase[0]), "/")
		}
		baseParsed, err := url.Parse(base)
		if err == nil && baseParsed.Scheme != "" && baseParsed.Host != "" {
			parsed.Scheme = baseParsed.Scheme
			parsed.Host = baseParsed.Host
		} else {
			parsed.Scheme = "https"
			parsed.Host = "api.theporndb.net"
		}
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = "/" + parts[0] + "/" + parts[1]
	return parsed.String(), nil
}

type prestigeAVPayload struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	PlayTime int    `json:"playTime"`
	Actress  []struct {
		Name string `json:"name"`
	} `json:"actress"`
	Genre []struct {
		Name string `json:"name"`
	} `json:"genre"`
	SKU []struct {
		SalesStartAt string `json:"salesStartAt"`
	} `json:"sku"`
	Maker struct {
		Name string `json:"name"`
	} `json:"maker"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
	Thumbnail struct {
		Path string `json:"path"`
	} `json:"thumbnail"`
	PackageImage struct {
		Path string `json:"path"`
	} `json:"packageImage"`
	Movie struct {
		Path string `json:"path"`
	} `json:"movie"`
}

func (p prestigeAVPayload) salesStartAt() string {
	if len(p.SKU) == 0 {
		return ""
	}
	return strings.Split(strings.TrimSpace(p.SKU[0].SalesStartAt), "T")[0]
}

func (p prestigeAVPayload) actorNames() []string {
	names := make([]string, 0, len(p.Actress))
	for _, item := range p.Actress {
		if name := normalizeWhitespace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return dedupeStrings(names)
}

type thePornDBSceneDetailResponse struct {
	Data thePornDBSceneDetail `json:"data"`
}

type thePornDBSearchResponse struct {
	Data []struct {
		Slug string `json:"slug"`
	} `json:"data"`
}

type thePornDBSceneDetail struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Image       string `json:"image"`
	Poster      string `json:"poster"`
	Background  struct {
		Large string `json:"large"`
	} `json:"background"`
	Posters struct {
		Large string `json:"large"`
	} `json:"posters"`
	Performers []struct {
		Name   string `json:"name"`
		Parent struct {
			Extras struct {
				Gender string `json:"gender"`
			} `json:"extras"`
		} `json:"parent"`
	} `json:"performers"`
}

func (s *ScraperService) fetchAVHTMLDocument(ctx context.Context, endpoint string, extraHeaders map[string]string) (*html.Node, string, error) {
	body, err := s.fetchAVHTMLDocumentText(ctx, endpoint, extraHeaders)
	if err != nil {
		return nil, "", err
	}
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("解析 AV 响应失败: %w", err)
	}
	return root, body, nil
}

func (s *ScraperService) fetchAVHTMLDocumentText(ctx context.Context, endpoint string, extraHeaders map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("创建 AV 刮削请求失败: %w", err)
	}
	if strings.TrimSpace(s.avUserAgent) != "" {
		req.Header.Set("User-Agent", s.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	client := s.avHTTPClient
	if client == nil {
		client = s.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AV 站点请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("AV 站点请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("读取 AV 响应失败: %w", err)
	}
	return string(body), nil
}

func (s *ScraperService) postAVFormText(ctx context.Context, endpoint string, form url.Values, extraHeaders map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建 AV 表单请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if strings.TrimSpace(s.avUserAgent) != "" {
		req.Header.Set("User-Agent", s.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	client := s.avHTTPClient
	if client == nil {
		client = s.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AV 表单请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("AV 表单请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("读取 AV 表单响应失败: %w", err)
	}
	return string(body), nil
}

func (s *ScraperService) fetchAVJSON(ctx context.Context, endpoint string, extraHeaders map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("创建 AV JSON 请求失败: %w", err)
	}
	if strings.TrimSpace(s.avUserAgent) != "" {
		req.Header.Set("User-Agent", s.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	client := s.avHTTPClient
	if client == nil {
		client = s.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("AV JSON 请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("AV JSON 请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析 AV JSON 响应失败: %w", err)
	}
	return nil
}

func (s *ScraperService) postAVJSON(ctx context.Context, endpoint string, body any, extraHeaders map[string]string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("编码 AV JSON 请求失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("创建 AV JSON POST 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(s.avUserAgent) != "" {
		req.Header.Set("User-Agent", s.avUserAgent)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	client := s.avHTTPClient
	if client == nil {
		client = s.httpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("AV JSON POST 请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("AV JSON POST 请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析 AV JSON POST 响应失败: %w", err)
	}
	return nil
}

func avTableValue(root *html.Node, label string) string {
	for _, row := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "tr" }) {
		cells := findAll(row, func(n *html.Node) bool { return n.Type == html.ElementNode && (n.Data == "th" || n.Data == "td") })
		if len(cells) < 2 {
			continue
		}
		if strings.Contains(strings.ToLower(nodeText(cells[0])), strings.ToLower(label)) {
			return strings.TrimSpace(nodeText(cells[1]))
		}
	}
	return ""
}

func avTableValueAny(root *html.Node, labels ...string) string {
	for _, label := range labels {
		if value := avTableValue(root, label); strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func avTableLinks(root *html.Node, label string) []string {
	for _, row := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "tr" }) {
		cells := findAll(row, func(n *html.Node) bool { return n.Type == html.ElementNode && (n.Data == "th" || n.Data == "td") })
		if len(cells) < 2 {
			continue
		}
		if strings.Contains(strings.ToLower(nodeText(cells[0])), strings.ToLower(label)) {
			values := make([]string, 0)
			for _, link := range findAll(cells[1], func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "a" }) {
				if text := strings.TrimSpace(nodeText(link)); text != "" {
					values = append(values, text)
				}
			}
			return dedupeStrings(values)
		}
	}
	return nil
}

func avTableLinksAny(root *html.Node, labels ...string) []string {
	for _, label := range labels {
		if values := avTableLinks(root, label); len(values) > 0 {
			return values
		}
	}
	return nil
}

type dmmSearchHit struct {
	URL      string
	Category string
}

func dmmSearchURLs(baseURL, input string) []string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://www.dmm.co.jp"
	}
	number := strings.ToLower(strings.TrimSpace(input))
	if match := dmmNumberPartsPattern.FindStringSubmatch(number); len(match) > 2 {
		digits := match[2]
		switch {
		case len(digits) >= 5 && strings.HasPrefix(digits, "00"):
			number = strings.Replace(number, digits, digits[2:], 1)
		case len(digits) == 4:
			number = strings.ReplaceAll(number, "-", "0")
		}
	}
	number00 := strings.ReplaceAll(number, "-", "00")
	numberNo00 := strings.ReplaceAll(number, "-", "")
	return []string{
		baseURL + "/search/=/searchstr=" + number00 + "/sort=ranking/",
		baseURL + "/search/=/searchstr=" + numberNo00 + "/sort=ranking/",
		strings.Replace(baseURL, "www.dmm.co.jp", "www.dmm.com", 1) + "/search/=/searchstr=" + numberNo00 + "/sort=ranking/",
	}
}

func dmmIsRegionBlocked(body string) bool {
	return strings.Contains(body, "not available in your region") ||
		strings.Contains(body, "お住まいの地域から") ||
		strings.Contains(body, "content is not available in this region")
}

func dmmParseSearchCandidates(content, input, baseURL string) []dmmSearchHit {
	hits := dmmParseAnchorCandidates(content, input, baseURL)
	if len(hits) > 0 {
		return hits
	}
	matches := dmmDetailURLPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	parts := dmmNumberPartsPattern.FindStringSubmatch(strings.ToLower(input))
	if len(parts) < 3 {
		return nil
	}
	prefix := parts[1]
	digits := parts[2]
	n1 := prefix + leftPad(digits, 5)
	n2 := prefix + digits
	result := make([]dmmSearchHit, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		raw := strings.ReplaceAll(match[1], `\u0026`, "&")
		raw = strings.Split(raw, "?")[0]
		if !matchesDMMID(raw, n1, n2) {
			continue
		}
		detailURL := toAbsoluteURL(baseURL+"/", raw)
		if detailURL == "" {
			continue
		}
		key := strings.ToLower(detailURL)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, dmmSearchHit{URL: detailURL, Category: dmmParseCategory(detailURL)})
	}
	return result
}

func dmmParseAnchorCandidates(content, input, baseURL string) []dmmSearchHit {
	matches := regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']*(?:cid=|content=)[^"']+)["'][^>]*>(.*?)</a>`).FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	parts := dmmNumberPartsPattern.FindStringSubmatch(strings.ToLower(input))
	if len(parts) < 3 {
		return nil
	}
	targetCode := normalizeAVCodeForCompare(input)
	n1 := parts[1] + leftPad(parts[2], 5)
	n2 := parts[1] + parts[2]
	result := make([]dmmSearchHit, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		href := strings.TrimSpace(html.UnescapeString(match[1]))
		if href == "" {
			continue
		}
		if targetCode != "" {
			candidateCode := normalizeAVCodeForCompare(stripHTMLText(match[2]))
			if candidateCode == "" {
				candidateCode = normalizeAVCodeForCompare(href)
			}
			if candidateCode != "" && candidateCode != targetCode {
				continue
			}
		}
		if !matchesDMMID(href, n1, n2) {
			continue
		}
		detailURL := toAbsoluteURL(baseURL+"/", href)
		if detailURL == "" {
			continue
		}
		key := strings.ToLower(detailURL)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, dmmSearchHit{URL: detailURL, Category: dmmParseCategory(detailURL)})
	}
	return result
}

func dmmParseCategory(rawURL string) string {
	switch {
	case strings.Contains(rawURL, "tv.dmm.co.jp"):
		return dmmCategoryFanzaTV
	case strings.Contains(rawURL, "tv.dmm.com"):
		return dmmCategoryDMMTV
	case strings.Contains(rawURL, "/digital/") || strings.Contains(rawURL, "video.dmm.co.jp"):
		return dmmCategoryDigital
	case strings.Contains(rawURL, "/mono/"):
		return dmmCategoryMono
	case strings.Contains(rawURL, "/rental/"):
		return dmmCategoryRental
	default:
		return dmmCategoryOther
	}
}

func dmmParseDigitalDetail(root *html.Node, body, detailURL string) (avScrapeCandidate, error) {
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1"
	})))
	if title == "" {
		title = strings.TrimSpace(extractJSONLDString(body, "name"))
	}
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	code := strings.TrimSpace(avTableValue(root, "Number"))
	poster := strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && strings.EqualFold(attrValue(n, "property"), "og:image")
	}, "content"))
	thumb := strings.Replace(poster, "ps.jpg", "pl.jpg", 1)
	metadata := avScrapeCandidate{
		ExternalID:  detailURL,
		Code:        code,
		Title:       title,
		Overview:    strings.TrimSpace(extractJSONLDString(body, "description")),
		PosterURL:   poster,
		ThumbURL:    thumb,
		ReleaseDate: normalizeSlashDate(avTableValue(root, "Release Date")),
		Actors:      extractJSONLDActorNames(body),
		DetailURL:   detailURL,
		Raw: map[string]any{
			"site":            "dmm",
			"detail_url":      detailURL,
			"external_id":     detailURL,
			"av_code":         code,
			"title":           title,
			"poster_url":      poster,
			"thumb_url":       thumb,
			"release_date":    normalizeSlashDate(avTableValue(root, "Release Date")),
			"actors":          extractJSONLDActorNames(body),
			"overview":        strings.TrimSpace(extractJSONLDString(body, "description")),
			"original_plot":   strings.TrimSpace(extractJSONLDString(body, "description")),
			"mosaic":          "有码",
			"mosaic_reason":   "dmm",
			"external_url":    detailURL,
			"search_category": dmmParseCategory(detailURL),
		},
	}
	if metadata.ThumbURL == "" && metadata.PosterURL != "" {
		metadata.ThumbURL = strings.Replace(metadata.PosterURL, "ps.jpg", "pl.jpg", 1)
	}
	return metadata, nil
}

func (c *dmmAVCrawler) fetchDMMFanzaTV(ctx context.Context, detailURL string) (avScrapeCandidate, error) {
	contentID := extractDMMContentID(detailURL)
	if contentID == "" {
		return avScrapeCandidate{}, fmt.Errorf("dmm: cannot extract FANZA TV content id from %s", detailURL)
	}
	var resp dmmFanzaTVResponse
	if err := c.postDMMJSON(ctx, "https://api.tv.dmm.co.jp/graphql", dmmFanzaTVPayload(contentID), map[string]string{
		"Accept": "application/json",
	}, &resp); err != nil {
		return avScrapeCandidate{}, err
	}
	candidate, err := dmmParseFanzaTVResponse(resp, detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	return candidate, nil
}

func dmmParseFanzaTVResponse(resp dmmFanzaTVResponse, detailURL string) (avScrapeCandidate, error) {
	content := resp.Data.FanzaTVPlus.Content
	if strings.TrimSpace(content.Title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("dmm: empty FANZA TV content response")
	}
	extrafanart := make([]string, 0, len(content.SamplePictures))
	for _, sample := range content.SamplePictures {
		if sample.ImageLarge != "" {
			extrafanart = append(extrafanart, sample.ImageLarge)
		}
	}
	trailer := dmmFanzaTrailer(content.SampleMovie.URL)
	actors := dmmItemNames(content.Actresses)
	directors := dmmItemNames(content.Directors)
	tags := dmmItemNames(content.Genres)
	release := content.StartDeliveryAt
	if len(release) >= 10 {
		release = release[:10]
	}
	thumb := content.PackageLargeImage
	poster := content.PackageImage
	candidate := avScrapeCandidate{
		ExternalID:  detailURL,
		Code:        extractDMMContentID(detailURL),
		Title:       content.Title,
		Overview:    content.Description,
		PosterURL:   poster,
		ThumbURL:    thumb,
		ReleaseDate: release,
		Actors:      actors,
		DetailURL:   detailURL,
		Raw: map[string]any{
			"site":           "dmm",
			"detail_url":     detailURL,
			"external_id":    detailURL,
			"av_code":        extractDMMContentID(detailURL),
			"title":          content.Title,
			"original_title": content.Title,
			"outline":        content.Description,
			"original_plot":  content.Description,
			"poster_url":     poster,
			"thumb_url":      thumb,
			"trailer_url":    trailer,
			"release_date":   release,
			"actors":         actors,
			"directors":      directors,
			"tags":           tags,
			"extrafanart":    extrafanart,
			"mosaic":         "有码",
		},
	}
	return candidate, nil
}

func dmmFanzaTVPayload(cid string) map[string]any {
	return map[string]any{
		"operationName": "FetchFanzaTvPlusContent",
		"variables": map[string]any{
			"id":         cid,
			"device":     "BROWSER",
			"isForeign":  false,
			"withResume": false,
		},
		"query": `query FetchFanzaTvPlusContent($id: ID!, $device: Device!, $withResume: Boolean!, $isForeign: Boolean) {
  fanzaTvPlus(device: $device) {
    content(id: $id, isForeign: $isForeign) {
      title
      description(format: HTML)
      packageImage
      packageLargeImage
      startDeliveryAt
      sampleMovie { url }
      samplePictures { imageLarge }
      actresses { name }
      directors { name }
      series { name }
      maker { name }
      label { name }
      genres { name }
      reviewSummary { averagePoint }
      playInfo(withResume: $withResume, device: $device) { duration }
    }
  }
}`,
	}
}

type dmmFanzaTVResponse struct {
	Data struct {
		FanzaTVPlus struct {
			Content dmmFanzaTVContent `json:"content"`
		} `json:"fanzaTvPlus"`
	} `json:"data"`
}

type dmmFanzaTVContent struct {
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	PackageImage      string              `json:"packageImage"`
	PackageLargeImage string              `json:"packageLargeImage"`
	StartDeliveryAt   string              `json:"startDeliveryAt"`
	SampleMovie       dmmFanzaTVMovie     `json:"sampleMovie"`
	SamplePictures    []dmmFanzaTVPicture `json:"samplePictures"`
	Actresses         []dmmFanzaTVItem    `json:"actresses"`
	Directors         []dmmFanzaTVItem    `json:"directors"`
	Series            dmmFanzaTVItem      `json:"series"`
	Maker             dmmFanzaTVItem      `json:"maker"`
	Label             dmmFanzaTVItem      `json:"label"`
	Genres            []dmmFanzaTVItem    `json:"genres"`
	ReviewSummary     struct {
		AveragePoint float64 `json:"averagePoint"`
	} `json:"reviewSummary"`
	PlayInfo struct {
		Duration int `json:"duration"`
	} `json:"playInfo"`
}

type dmmFanzaTVMovie struct {
	URL string `json:"url"`
}

type dmmFanzaTVPicture struct {
	ImageLarge string `json:"imageLarge"`
}

type dmmFanzaTVItem struct {
	Name string `json:"name"`
}

func dmmItemNames(items []dmmFanzaTVItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if name := normalizeWhitespace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return dedupeStrings(names)
}

func dmmFanzaTrailer(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	trailerURL := strings.Replace(rawURL, "hlsvideo", "litevideo", 1)
	match := regexp.MustCompile(`/([^/]+)/playlist\.m3u8`).FindStringSubmatch(trailerURL)
	if match == nil {
		return ""
	}
	return strings.Replace(trailerURL, "playlist.m3u8", match[1]+"_sm_w.mp4", 1)
}

func leftPad(value string, size int) string {
	for len(value) < size {
		value = "0" + value
	}
	return value
}

func extractDMMContentID(rawURL string) string {
	matches := regexp.MustCompile(`(?:content|cid)=([a-zA-Z0-9_]+)`).FindStringSubmatch(rawURL)
	if len(matches) == 2 {
		return strings.ToLower(matches[1])
	}
	return extractSingleSitePathID(rawURL)
}

func matchesDMMID(rawURL string, numbers ...string) bool {
	lower := strings.ToLower(rawURL)
	for _, number := range numbers {
		if number == "" {
			continue
		}
		if strings.Contains(lower, "cid="+number) ||
			strings.Contains(lower, "content="+number) ||
			strings.Contains(lower, "/"+number+"/") {
			return true
		}
	}
	return false
}

func parseDMMDetailURLs(content, query, baseURL string) []string {
	matches := regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']*cid=[^"']+)["'][^>]*>(.*?)</a>`).FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	targetCode := normalizeAVCodeForCompare(query)
	out := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		href := strings.TrimSpace(html.UnescapeString(match[1]))
		anchorText := stripHTMLText(match[2])
		if targetCode != "" {
			candidateCode := normalizeAVCodeForCompare(anchorText)
			if candidateCode == "" {
				candidateCode = normalizeAVCodeForCompare(href)
			}
			if candidateCode != "" && candidateCode != targetCode {
				continue
			}
		}
		detailURL := toAbsoluteURL(baseURL+"/", href)
		if detailURL == "" {
			continue
		}
		key := strings.ToLower(detailURL)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, detailURL)
	}
	return out
}

func findAll(root *html.Node, predicate func(*html.Node) bool) []*html.Node {
	if root == nil {
		return nil
	}
	out := make([]*html.Node, 0)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if predicate(n) {
			out = append(out, n)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return out
}

func findFirstAttr(root *html.Node, predicate func(*html.Node) bool, attr string) string {
	node := findFirst(root, predicate)
	if node == nil {
		return ""
	}
	return attrValue(node, attr)
}

func findAttrs(root *html.Node, predicate func(*html.Node) bool, attr string) []string {
	nodes := findAll(root, predicate)
	if len(nodes) == 0 {
		return nil
	}
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if value := strings.TrimSpace(attrValue(node, attr)); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func findNodeTextOnClass(root *html.Node, tag string, className string) string {
	return strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag && hasClass(n, className)
	})))
}

func findAttrOnClass(root *html.Node, tag string, className string, attr string) string {
	return strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag && hasClass(n, className)
	}, attr))
}

func findAttrByID(root *html.Node, tag string, id string, attr string) string {
	return strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag && attrValue(n, "id") == id
	}, attr))
}

func resolveURLs(base string, values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if resolved := toAbsoluteURL(base, value); strings.TrimSpace(resolved) != "" {
			out = append(out, resolved)
		}
	}
	return dedupeStrings(out)
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeSlashDate(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "/", "-")
	return regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}`).FindString(value)
}

func dmmNormalizeImageURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Replace(value, "ps.jpg", "pl.jpg", 1)
}

func extractDMMCID(rawURL string) string {
	matches := regexp.MustCompile(`cid=([a-zA-Z0-9_]+)`).FindStringSubmatch(rawURL)
	if len(matches) == 2 {
		return strings.ToLower(matches[1])
	}
	return extractSingleSitePathID(rawURL)
}

func extractJSONLDString(raw, key string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?is)"%s"\s*:\s*"([^"]+)"`, regexp.QuoteMeta(key)))
	matches := pattern.FindStringSubmatch(raw)
	if len(matches) < 2 {
		return ""
	}
	return normalizeWhitespace(matches[1])
}

func extractJSONLDActorNames(raw string) []string {
	matches := regexp.MustCompile(`(?is)"actor"\s*:\s*\[(.*?)\]`).FindStringSubmatch(raw)
	if len(matches) < 2 {
		return nil
	}
	nameMatches := regexp.MustCompile(`(?is)"name"\s*:\s*"([^"]+)"`).FindAllStringSubmatch(matches[1], -1)
	names := make([]string, 0, len(nameMatches))
	for _, match := range nameMatches {
		if len(match) < 2 {
			continue
		}
		if name := normalizeWhitespace(match[1]); name != "" {
			names = append(names, name)
		}
	}
	return dedupeStrings(names)
}

func extractRegexGroup(raw, pattern string) string {
	matches := regexp.MustCompile(pattern).FindStringSubmatch(raw)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func prestigeNormalizeURL(rawURL string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("invalid prestige url %q", rawURL)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid prestige url %q", rawURL)
	}
	externalID := parts[len(parts)-1]
	if !strings.Contains(parsed.Path, "/api/product/") {
		parsed.Path = strings.Replace(parsed.Path, "/goods/", "/api/product/", 1)
	}
	parsed.RawQuery = ""
	return parsed.String(), externalID, nil
}

func prestigeExtractSKU(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(parsed.Query().Get("skuId")))
}

func prestigeMediaURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if parsed, err := url.Parse(path); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return parsed.String()
	}
	return "https://www.prestige-av.com/api/media/" + strings.TrimLeft(path, "/")
}

func xcityListValue(root *html.Node, label string) string {
	for _, item := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "li" }) {
		text := nodeText(item)
		if !strings.Contains(strings.ToLower(text), strings.ToLower(label)) {
			continue
		}
		return strings.TrimSpace(strings.ReplaceAll(text, label, ""))
	}
	return ""
}

func xcityTextsFromContainer(root *html.Node, tag string, className string, childTag string) []string {
	container := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag && hasClass(n, className)
	})
	if container == nil {
		return nil
	}
	values := make([]string, 0)
	for _, node := range findAll(container, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == childTag }) {
		if text := strings.TrimSpace(nodeText(node)); text != "" {
			values = append(values, text)
		}
	}
	return dedupeStrings(values)
}

func xcityQueryID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return normalizeWhitespace(parsed.Query().Get("id"))
}

func getchuOutline(root *html.Node) string {
	values := make([]string, 0)
	for _, node := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "tablebody")
	}) {
		if text := strings.TrimSpace(nodeText(node)); text != "" {
			values = append(values, text)
		}
	}
	return normalizeWhitespace(strings.Join(values, " "))
}

func getchuTableValue(root *html.Node, label string) string {
	for _, row := range findAll(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "tr" }) {
		cells := findAll(row, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "td" })
		if len(cells) < 2 {
			continue
		}
		if strings.Contains(strings.TrimSpace(nodeText(cells[0])), label) {
			return strings.TrimSpace(nodeText(cells[1]))
		}
	}
	return ""
}

func getchuQueryValue(rawURL string, key string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return parsed.Query().Get(key)
}

func buildThePornDBQuery(number string) string {
	replacer := strings.NewReplacer(".", " ", "-", " ", "_", " ", ",", " ")
	return normalizeWhitespace(strings.ToUpper(replacer.Replace(number)))
}

func thePornDBSearchKeywords(filePath string) ([]string, string, string) {
	value := strings.TrimSpace(filePath)
	fileName := strings.ReplaceAll(value, "\\", "/")
	fileName = path.Base(fileName)
	fileName = strings.ReplaceAll(fileName, ",", ".")
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))

	matches := datedNumberPattern.FindStringSubmatch(value)
	if len(matches) == 4 {
		fullNumber := matches[1]
		series := thePornDBLongName(strings.NewReplacer("-", "", ".", "").Replace(strings.ToLower(matches[2])))
		date := "20" + strings.ReplaceAll(matches[3], ".", "-")
		date = strings.ReplaceAll(date, "_", "-")
		title := strings.TrimSpace(strings.Replace(fileName, fullNumber, "", 1))
		title = regexp.MustCompile(`[-_&.]`).ReplaceAllString(title, " ")
		titleParts := thePornDBMeaningfulParts(title, series)
		keywords := []string{series + " " + date}
		if len(titleParts) > 0 {
			keywords = append(keywords, series+" "+strings.Join(titleParts[:min(2, len(titleParts))], " "))
		}
		return thePornDBUniqueStrings(keywords), series, date
	}

	plain := strings.ReplaceAll(fileName, "-", " ")
	plain = strings.ReplaceAll(plain, "_", " ")
	parts := strings.Split(plain, ".")
	limit := min(2, len(parts))
	if limit == 0 {
		return []string{}, "", ""
	}
	keyword := strings.Join(parts[:limit], " ")
	keyword = strings.Join(strings.Fields(keyword), " ")
	if keyword == "" {
		return []string{}, "", ""
	}
	return []string{keyword}, "", ""
}

func thePornDBSearchURL(baseURL, category, keyword string) string {
	values := url.Values{}
	values.Set("parse", keyword)
	values.Set("per_page", "100")
	return strings.TrimRight(baseURL, "/") + "/" + category + "?" + values.Encode()
}

func thePornDBBestMatchURL(items []thePornDBItem, filePath, series, date, category, baseURL string) string {
	if len(items) == 0 {
		return ""
	}
	fileName := strings.ToLower(path.Base(strings.ReplaceAll(filePath, "\\", "/")))
	dateSuffix := regexp.MustCompile(`[\.-_]\d{2}\.\d{2}\.\d{2}(.+)`).FindStringSubmatch(fileName)
	if len(dateSuffix) > 1 {
		fileName = dateSuffix[1]
	}
	actorNumber := len(strings.Split(strings.ReplaceAll(fileName, ".and.", "&"), "&"))
	fileSpace := thePornDBCleanWords(filePath)
	fileNoSpace := strings.ReplaceAll(fileSpace, " ", "")

	type candidate struct {
		url   string
		text  string
		score float64
	}
	var dateMatches, titleMatches, actorMatches []candidate
	for _, item := range items {
		detailURL := thePornDBItemDetailURL(item, category, baseURL)
		if detailURL == "" {
			continue
		}
		itemSeries := strings.ToLower(item.Site.ShortName)
		itemURL := strings.ReplaceAll(strings.ToLower(item.Site.URL), "-", "")
		titleSpace := thePornDBCleanWords(item.Title)
		titleNoSpace := strings.ReplaceAll(titleSpace, " ", "")
		actorSpaces, actorNoSpaces := thePornDBPerformerWords(item.Performers)
		actorTitle := strings.Join(append(actorSpaces, titleSpace), " ")
		next := candidate{url: detailURL, text: actorTitle, score: thePornDBSimilarity(actorTitle, fileSpace)}

		if series != "" {
			if series == itemSeries || strings.Contains(itemURL, series) {
				switch {
				case date != "" && item.Date == date:
					dateMatches = append(dateMatches, next)
				case titleNoSpace != "" && strings.Contains(fileNoSpace, titleNoSpace):
					titleMatches = append(titleMatches, next)
				case len(actorNoSpaces) >= actorNumber && thePornDBAllContained(actorNoSpaces, fileNoSpace):
					actorMatches = append(actorMatches, next)
				}
			} else if date != "" && item.Date == date && titleNoSpace != "" && strings.Contains(fileNoSpace, titleNoSpace) {
				titleMatches = append(titleMatches, next)
			}
			continue
		}
		if titleNoSpace == "" || strings.Contains(fileNoSpace, titleNoSpace) || len(items) == 1 {
			titleMatches = append(titleMatches, next)
		}
	}
	for _, candidates := range [][]candidate{dateMatches, titleMatches, actorMatches} {
		if len(candidates) > 0 {
			sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
			return candidates[0].url
		}
	}
	return ""
}

func thePornDBMetadataFromItem(item thePornDBItem, category string) avScrapeCandidate {
	poster := firstNonEmptyString(item.Background.Large, item.Image, item.Posters.Large, item.Poster)
	thumb := firstNonEmptyString(item.Posters.Large, item.Poster)
	actors, allActors := thePornDBActors(item.Performers)
	release := item.Date
	candidate := avScrapeCandidate{
		ExternalID:  item.Slug,
		Title:       item.Title,
		Overview:    thePornDBCleanDescription(item.Description),
		PosterURL:   poster,
		ThumbURL:    thumb,
		ReleaseDate: release,
		Actors:      actors,
		DetailURL:   thePornDBItemDetailURL(item, category, "https://api.theporndb.net"),
		Raw: map[string]any{
			"site":           "theporndb",
			"external_id":    item.Slug,
			"title":          item.Title,
			"original_title": item.Title,
			"outline":        thePornDBCleanDescription(item.Description),
			"original_plot":  thePornDBCleanDescription(item.Description),
			"poster_url":     poster,
			"thumb_url":      thumb,
			"release_date":   release,
			"actors":         actors,
			"all_actors":     allActors,
			"tags":           thePornDBNames(item.Tags),
			"mosaic":         "有码",
		},
	}
	if len(release) >= 4 {
		candidate.Raw["year"] = release[:4]
	}
	if item.Trailer != "" {
		candidate.Raw["trailer_url"] = item.Trailer
	}
	if item.Director.Name != "" {
		candidate.Raw["directors"] = []string{item.Director.Name}
	}
	if item.Site.Name != "" {
		candidate.Raw["series"] = item.Site.Name
	}
	if item.Site.Network.Name != "" {
		candidate.Raw["studio"] = item.Site.Network.Name
		candidate.Raw["publisher"] = item.Site.Network.Name
	}
	return candidate
}

func thePornDBSearchListKeywords(input string) string {
	return buildThePornDBQuery(input)
}

func thePornDBActorNames(items []struct {
	Name   string `json:"name"`
	Parent struct {
		Extras struct {
			Gender string `json:"gender"`
		} `json:"extras"`
	} `json:"parent"`
}) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(normalizeWhitespace(item.Parent.Extras.Gender), "male") {
			continue
		}
		if name := normalizeWhitespace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return dedupeStrings(names)
}

func thePornDBActors(items []thePornDBPerformer) ([]string, []string) {
	allActors := make([]string, 0, len(items))
	actors := make([]string, 0, len(items))
	for _, performer := range items {
		name := strings.TrimSpace(performer.Name)
		if name == "" {
			continue
		}
		allActors = append(allActors, name)
		if !strings.EqualFold(performer.Parent.Extras.Gender, "Male") {
			actors = append(actors, name)
		}
	}
	return thePornDBUniqueStrings(actors), thePornDBUniqueStrings(allActors)
}

func thePornDBNames(values []thePornDBNameRef) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value.Name) != "" {
			result = append(result, strings.TrimSpace(value.Name))
		}
	}
	return thePornDBUniqueStrings(result)
}

func thePornDBPerformerWords(values []thePornDBPerformer) ([]string, []string) {
	spaces := make([]string, 0, len(values))
	noSpaces := make([]string, 0, len(values))
	for _, value := range values {
		cleaned := thePornDBCleanWords(value.Name)
		if cleaned == "" {
			continue
		}
		spaces = append(spaces, cleaned)
		noSpaces = append(noSpaces, strings.ReplaceAll(cleaned, " ", ""))
	}
	return spaces, noSpaces
}

func thePornDBAllContained(values []string, haystack string) bool {
	for _, value := range values {
		if !strings.Contains(haystack, value) {
			return false
		}
	}
	return len(values) > 0
}

func thePornDBCleanWords(value string) string {
	cleaned := regexp.MustCompile(`[\W_]+`).ReplaceAllString(strings.ToLower(value), " ")
	return strings.Join(strings.Fields(cleaned), " ")
}

func thePornDBMeaningfulParts(value string, series string) []string {
	fields := strings.Fields(value)
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.EqualFold(field, series) {
			continue
		}
		result = append(result, field)
	}
	return result
}

func thePornDBCleanDescription(value string) string {
	return strings.NewReplacer("＜p＞", "", "＜/p＞", "", "<p>", "", "</p>", "").Replace(value)
}

func thePornDBSimilarity(a string, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	ar := []rune(a)
	br := []rune(b)
	previous := make([]int, len(br)+1)
	current := make([]int, len(br)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		current[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, min(current[j-1]+1, previous[j-1]+cost))
		}
		previous, current = current, previous
	}
	distance := previous[len(br)]
	maxLen := math.Max(float64(len(ar)), float64(len(br)))
	return 1 - float64(distance)/maxLen
}

func thePornDBLongName(shortName string) string {
	names := map[string]string{
		"clubseventeen": "clubsweethearts",
	}
	if value, ok := names[shortName]; ok {
		return strings.ReplaceAll(strings.ReplaceAll(value, "-", ""), ".", "")
	}
	return shortName
}

func thePornDBUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func categoryFromThePornDBDetailURL(detailURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(detailURL))
	if err != nil {
		return "scenes"
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) > 0 && parts[0] == "movies" {
		return "movies"
	}
	return "scenes"
}

func thePornDBItemDetailURL(item thePornDBItem, category, baseURL string) string {
	if strings.TrimSpace(item.Slug) == "" {
		return ""
	}
	switch category {
	case "movies":
	case "scenes":
	default:
		category = "scenes"
	}
	return strings.TrimRight(baseURL, "/") + "/" + category + "/" + item.Slug
}

type thePornDBSearchListResponse struct {
	Data []thePornDBItem `json:"data"`
}

type thePornDBItemResponse struct {
	Data thePornDBItem `json:"data"`
}

type thePornDBItem struct {
	Slug        string               `json:"slug"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Date        string               `json:"date"`
	Duration    int                  `json:"duration"`
	Trailer     string               `json:"trailer"`
	Image       string               `json:"image"`
	Poster      string               `json:"poster"`
	Background  thePornDBImageSet    `json:"background"`
	Posters     thePornDBImageSet    `json:"posters"`
	Director    thePornDBNameRef     `json:"director"`
	Site        thePornDBSiteRef     `json:"site"`
	Tags        []thePornDBNameRef   `json:"tags"`
	Performers  []thePornDBPerformer `json:"performers"`
}

type thePornDBImageSet struct {
	Large string `json:"large"`
}

type thePornDBNameRef struct {
	Name string `json:"name"`
}

type thePornDBSiteRef struct {
	Name      string           `json:"name"`
	ShortName string           `json:"short_name"`
	URL       string           `json:"url"`
	Network   thePornDBNameRef `json:"network"`
}

type thePornDBPerformer struct {
	Name   string                   `json:"name"`
	Parent thePornDBPerformerParent `json:"parent"`
}

type thePornDBPerformerParent struct {
	Extras thePornDBPerformerExtras `json:"extras"`
}

type thePornDBPerformerExtras struct {
	Gender string `json:"gender"`
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func minPreviewLimit(a, b int) int {
	if a <= 0 {
		return b
	}
	if b < a {
		return b
	}
	return a
}

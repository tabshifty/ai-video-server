package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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

func (c *dmmAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	return nil, nil
}

func (c *mgstageAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	return nil, nil
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
	root, body, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, map[string]string{
		"Cookie": "age_check_done=1",
	})
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1"
	})))
	code := strings.TrimSpace(avTableValue(root, "Number"))
	poster := strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && strings.EqualFold(attrValue(n, "property"), "og:image")
	}, "content"))
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSpace(extractJSONLDString(body, "name"))
	}
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  extractDMMCID(detailURL),
		Code:        code,
		Title:       title,
		Overview:    strings.TrimSpace(extractJSONLDString(body, "description")),
		PosterURL:   dmmNormalizeImageURL(poster),
		ReleaseDate: normalizeSlashDate(avTableValue(root, "Release Date")),
		Actors:      extractJSONLDActorNames(body),
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": extractDMMCID(detailURL),
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *mgstageAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	root, _, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, map[string]string{
		"Cookie": "adc=1",
	})
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}
	title := strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && attrValue(n, "id") == "center_column"
	})))
	code := strings.TrimSpace(avTableValue(root, "Number"))
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	poster := strings.TrimSpace(findAttrByID(root, "a", "EnlargeImage", "href"))
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  strings.ToLower(extractSingleSitePathID(detailURL)),
		Code:        code,
		Title:       title,
		Overview:    strings.TrimSpace(nodeText(findFirst(root, func(n *html.Node) bool { return n.Type == html.ElementNode && attrValue(n, "id") == "introduction" }))),
		PosterURL:   strings.Replace(poster, "/pb_", "/pf_", 1),
		ReleaseDate: normalizeSlashDate(avTableValue(root, "Release")),
		Actors:      avTableLinks(root, "Actor"),
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": strings.ToLower(extractSingleSitePathID(detailURL)),
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
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
		return nil, nil
	}
	query = buildThePornDBQuery(query)
	if query == "" {
		return nil, nil
	}
	base := strings.TrimRight(c.svc.avSiteBaseURL("theporndb", "https://api.theporndb.net"), "/")
	searchURL := fmt.Sprintf("%s/scenes?parse=%s", base, url.QueryEscape(query))
	if run != nil {
		run.addSearchURL(searchURL)
	}
	var payload thePornDBSearchResponse
	if err := c.svc.fetchAVJSON(ctx, searchURL, map[string]string{
		"Authorization": "Bearer " + c.svc.avThePornDBAPIToken,
		"Accept":        "application/json",
	}, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data) == 0 || strings.TrimSpace(payload.Data[0].Slug) == "" {
		return nil, nil
	}
	candidates := make([]avScrapeCandidate, 0, minPreviewLimit(limit, len(payload.Data)))
	for _, item := range payload.Data {
		if len(candidates) >= limit && limit > 0 {
			break
		}
		detailURL := base + "/scenes/" + strings.TrimSpace(item.Slug)
		candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (c *thePornDBAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	if strings.TrimSpace(c.svc.avThePornDBAPIToken) == "" {
		return avScrapeCandidate{}, fmt.Errorf("theporndb api token is required")
	}
	detailURL = normalizeThePornDBDetailURL(detailURL, c.svc.avSiteBaseURL("theporndb", "https://api.theporndb.net"))
	if run != nil {
		run.addDetailURL(detailURL)
	}
	var payload thePornDBSceneDetailResponse
	if err := c.svc.fetchAVJSON(ctx, detailURL, map[string]string{
		"Authorization": "Bearer " + c.svc.avThePornDBAPIToken,
		"Accept":        "application/json",
	}, &payload); err != nil {
		return avScrapeCandidate{}, err
	}
	title := normalizeWhitespace(payload.Data.Title)
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	candidate := avScrapeCandidate{
		Source:      c.Name(),
		ExternalID:  normalizeWhitespace(payload.Data.Slug),
		Title:       title,
		Overview:    normalizeWhitespace(payload.Data.Description),
		PosterURL:   firstNonEmptyString(payload.Data.Posters.Large, payload.Data.Poster),
		ThumbURL:    firstNonEmptyString(payload.Data.Background.Large, payload.Data.Image),
		ReleaseDate: normalizeWhitespace(payload.Data.Date),
		Actors:      thePornDBActorNames(payload.Data.Performers),
		DetailURL:   strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": normalizeWhitespace(payload.Data.Slug),
			"title":       title,
		},
	}
	return candidate, nil
}

func normalizeThePornDBDetailURL(rawURL, fallbackBase string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if strings.Contains(host, "theporndb") && !strings.Contains(host, "api.theporndb") {
		return toAbsoluteURL(strings.TrimSpace(fallbackBase), parsed.RequestURI())
	}
	if strings.Contains(host, "api.theporndb") {
		return rawURL
	}
	return rawURL
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建 AV 刮削请求失败: %w", err)
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
		return nil, "", fmt.Errorf("AV 站点请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, "", fmt.Errorf("AV 站点请求失败，状态码=%d，响应=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, "", fmt.Errorf("读取 AV 响应失败: %w", err)
	}
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, "", fmt.Errorf("解析 AV 响应失败: %w", err)
	}
	return root, string(body), nil
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

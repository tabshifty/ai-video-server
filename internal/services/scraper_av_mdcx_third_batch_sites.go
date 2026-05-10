package services

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	fc2HubDetailPathRe     = regexp.MustCompile(`^/detail/\d+/?$`)
	fc2PPVDBDetailPathRe   = regexp.MustCompile(`^/articles/(?:fc2-ppv-)?\d+/?$`)
	airAVDetailPathRe      = regexp.MustCompile(`^/video/fc2-ppv-\d+/?$`)
	jav321DetailPathRe     = regexp.MustCompile(`^/video/[a-z0-9]+/?$`)
	mywifeDetailPathRe     = regexp.MustCompile(`^/teigaku/model/no/\d+/?$`)
	avsoxDetailPathRe      = regexp.MustCompile(`^/cn/movie/[a-z0-9]+/?$`)
	freeJAVBTDetailPathRe  = regexp.MustCompile(`^/detail/fc2-ppv-\d+/?$`)
	madouquDetailPathRe    = regexp.MustCompile(`^/archives/\d+/?$`)
	falenoDetailPathRe     = regexp.MustCompile(`^/top/works/[a-z0-9-]+/?$`)
	fantasticaDetailPathRe = regexp.MustCompile(`^/items/detail/[a-z0-9-]+/?$`)
	gigaDetailPathRe       = regexp.MustCompile(`^/product/index\.php$`)
	javdayDetailPathRe     = regexp.MustCompile(`^/videos/[a-z0-9-]+/?$`)
	kin8DetailPathRe       = regexp.MustCompile(`^/moviepages/\d+/index\.html$`)
	love6DetailPathRe      = regexp.MustCompile(`^/albums/view/[a-z0-9=]+/?$`)
	lulubarDetailPathRe    = regexp.MustCompile(`^/video/detail/?$`)
)

type minimalHTMLAVCrawler struct {
	svc  *ScraperService
	site string
}

type mywifeAVCrawler struct {
	svc *ScraperService
}

func newFC2ClubAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "fc2club"}
}

func newFC2HubAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "fc2hub"}
}

func newFC2PPVDBAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "fc2ppvdb"}
}

func newAirAVAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "airav"}
}

func newJav321AVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "jav321"}
}

func newAVSOXAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "avsox"}
}

func newFreeJAVBTAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "freejavbt"}
}

func newMadouquAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "madouqu"}
}

func newMDTVAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "mdtv.com"}
}

func newCNMDBAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "cnmdb"}
}

func newFalenoAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "faleno"}
}

func newFantasticaAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "fantastica"}
}

func newGigaAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "giga"}
}

func newJavdayAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "javday"}
}

func newKin8AVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "kin8"}
}

func newLove6AVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "love6"}
}

func newLulubarAVCrawler(svc *ScraperService) avCrawler {
	return &minimalHTMLAVCrawler{svc: svc, site: "lulubar"}
}

func newMywifeAVCrawler(svc *ScraperService) avCrawler {
	return &mywifeAVCrawler{svc: svc}
}

func (c *minimalHTMLAVCrawler) Name() string {
	return c.site
}

func (c *minimalHTMLAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	switch c.site {
	case "fc2club":
		if number := normalizeFC2NumericID(query); number != "" {
			detailURL := toAbsoluteURL(c.svc.avSiteBaseURL(c.site, "https://fc2club.top"), "/html/FC2-"+number+".html")
			if run != nil {
				run.addSearchURL(detailURL)
			}
			candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
			if err != nil {
				return nil, err
			}
			return []avScrapeCandidate{candidate}, nil
		}
	case "fc2ppvdb":
		if number := normalizeFC2NumericID(query); number != "" {
			detailURL := toAbsoluteURL(c.svc.avSiteBaseURL(c.site, "https://fc2ppvdb.com"), "/articles/"+number)
			if run != nil {
				run.addSearchURL(detailURL)
			}
			candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
			if err != nil {
				return nil, err
			}
			return []avScrapeCandidate{candidate}, nil
		}
	case "fc2hub":
		if number := normalizeFC2NumericID(query); number != "" {
			base := c.svc.avSiteBaseURL(c.site, "https://javten.com")
			searchURL := toAbsoluteURL(base, "/search?kw="+url.QueryEscape(number))
			if run != nil {
				run.addSearchURL(searchURL)
			}
			content, err := c.svc.fetchAVHTML(ctx, searchURL)
			if err != nil {
				return nil, err
			}
			detailURL := parseFC2HubDetailURL(content, number, base)
			if detailURL == "" {
				return nil, nil
			}
			candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
			if err != nil {
				return nil, err
			}
			return []avScrapeCandidate{candidate}, nil
		}
	case "jav321":
		base := c.svc.avSiteBaseURL(c.site, "https://www.jav321.com")
		searchURL := toAbsoluteURL(base, "/search")
		if run != nil {
			run.addSearchURL(searchURL)
		}
		content, err := c.svc.postAVFormText(ctx, searchURL, url.Values{"sn": {query}}, nil)
		if err != nil {
			return nil, err
		}
		candidate, err := parseJav321SearchCandidate(content, query, base)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(candidate.Title) == "" {
			return nil, nil
		}
		return []avScrapeCandidate{candidate}, nil
	}
	return nil, nil
}

func (c *minimalHTMLAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	if run != nil {
		run.addDetailURL(detailURL)
	}
	content, err := c.svc.fetchAVHTML(ctx, detailURL)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return avScrapeCandidate{}, fmt.Errorf("parse %s detail html: %w", c.site, err)
	}

	title := extractMinimalHTMLHeadingTitle(root)
	if title == "" {
		title = extractMinimalHTMLTitle(root)
	}
	code := c.extractCode(root, detailURL)
	title = trimTitleByNumber(title, code)
	if c.site == "fc2club" {
		title = strings.TrimSpace(regexp.MustCompile(`(?i)^FC2[-_ ]?`+regexp.QuoteMeta(normalizeFC2NumericID(code))+`\s*`).ReplaceAllString(title, ""))
	}
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}

	externalID := minimalHTMLExternalID(detailURL, c.site)
	posterURL := c.extractPosterURL(root, detailURL)
	candidate := avScrapeCandidate{
		Source:     c.site,
		ExternalID: externalID,
		Code:       code,
		Title:      title,
		PosterURL:  posterURL,
		ThumbURL:   posterURL,
		DetailURL:  strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.site,
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": externalID,
			"av_code":     code,
			"title":       title,
			"poster_url":  posterURL,
			"thumb_url":   posterURL,
		},
	}
	return candidate, nil
}

func (c *minimalHTMLAVCrawler) extractCode(root *html.Node, detailURL string) string {
	content := nodeText(root)
	switch c.site {
	case "fc2club", "fc2hub", "fc2ppvdb", "airav", "freejavbt":
		if number := normalizeFC2NumericIDFromDetailPath(detailURL); number != "" {
			return "FC2-PPV-" + number
		}
		if number := normalizeFC2NumericID(content); number != "" {
			return "FC2-PPV-" + number
		}
	case "jav321", "avsox", "madouqu", "mdtv.com", "cnmdb", "faleno", "fantastica", "giga", "javday", "kin8", "love6", "lulubar":
		switch c.site {
		case "madouqu":
			if code := regexp.MustCompile(`(?i)\bMD-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		case "mdtv.com":
			if code := regexp.MustCompile(`(?i)\bMDTV-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		case "cnmdb":
			if code := regexp.MustCompile(`(?i)\bCNMDB-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		case "kin8":
			if code := regexp.MustCompile(`(?i)\bKIN8-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		case "love6":
			if code := regexp.MustCompile(`(?i)\bLOVE6-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		case "lulubar":
			if code := regexp.MustCompile(`(?i)\bLULU-\d+\b`).FindString(content); code != "" {
				return strings.ToUpper(strings.TrimSpace(code))
			}
		}
		if code := extractAggregateSupplementNumber(content); code != "" {
			return code
		}
		if code := extractAVCode(content); code != "" {
			return code
		}
		if code := extractAVCode(detailURL); code != "" {
			return code
		}
	}
	return extractAggregateSupplementNumber(content)
}

func normalizeFC2NumericIDFromDetailPath(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(parsed.Path)
	if path == "" {
		return ""
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)/articles/(?:fc2-ppv-)?(\d{5,8})/?$`),
		regexp.MustCompile(`(?i)/html/fc2-(\d{5,8})\.html$`),
		regexp.MustCompile(`(?i)/detail/(\d{5,8})/?$`),
		regexp.MustCompile(`(?i)/video/fc2-ppv-(\d{5,8})/?$`),
	}
	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(path); len(matches) > 1 {
			number := strings.TrimLeft(matches[1], "0")
			if number == "" {
				return "0"
			}
			return number
		}
	}
	return ""
}

func (c *minimalHTMLAVCrawler) extractPosterURL(root *html.Node, detailURL string) string {
	if poster := findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "meta" && strings.EqualFold(attrValue(n, "property"), "og:image")
	}, "content"); strings.TrimSpace(poster) != "" {
		return toAbsoluteURL(detailURL, poster)
	}
	for _, image := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img"
	}) {
		src := strings.TrimSpace(attrValue(image, "src"))
		if src == "" || isLikelyLogoURL(src) {
			continue
		}
		return toAbsoluteURL(detailURL, src)
	}
	return ""
}

func parseFC2HubDetailURL(content, number, baseURL string) string {
	number = normalizeFC2NumericID(number)
	if number == "" {
		return ""
	}
	matches := regexp.MustCompile(`(?is)<(?:a|link)[^>]+href=["']([^"']+)["']`).FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		href := strings.TrimSpace(match[1])
		if !strings.Contains(href, number) {
			continue
		}
		if detailURL := toAbsoluteURL(baseURL, href); detailURL != "" {
			return detailURL
		}
	}
	return ""
}

func parseJav321SearchCandidate(content, query, baseURL string) (avScrapeCandidate, error) {
	if strings.Contains(content, "AVが見つかりませんでした") {
		return avScrapeCandidate{}, nil
	}
	title := stripHTMLText(regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`).FindString(content))
	title = strings.TrimSpace(regexp.MustCompile(`\s+sample\s*$`).ReplaceAllString(title, ""))
	if title == "" {
		title = stripHTMLText(extractRegexGroup(content, `(?is)<title[^>]*>(.*?)</title>`))
	}
	if title == "" {
		return avScrapeCandidate{}, nil
	}
	detailURL := toAbsoluteURL(baseURL, extractRegexGroup(content, `(?is)<a[^>]+href=["']([^"']*/video/[^"']+)["']`))
	code := firstNonEmptyString(
		extractRegexGroup(content, `(?is)(?:品番|番号|品號)\s*[:：]\s*([A-Za-z0-9_-]+)`),
		extractAVCode(title),
		extractAVCode(query),
	)
	poster := toAbsoluteURL(baseURL, extractRegexGroup(content, `(?is)<img[^>]+src=["']([^"']+)["']`))
	actors := splitDelimitedNames(extractRegexGroup(content, `(?is)(?:出演者|出演)\s*[:：]\s*([^<]+)`))
	candidate := avScrapeCandidate{
		Source:     "jav321",
		ExternalID: strings.ToLower(extractSingleSitePathID(detailURL)),
		Code:       strings.ToUpper(strings.TrimSpace(code)),
		Title:      title,
		PosterURL:  poster,
		ThumbURL:   poster,
		Actors:     actors,
		DetailURL:  detailURL,
		Raw: map[string]any{
			"site":        "jav321",
			"detail_url":  detailURL,
			"external_id": strings.ToLower(extractSingleSitePathID(detailURL)),
			"av_code":     strings.ToUpper(strings.TrimSpace(code)),
			"title":       title,
			"poster_url":  poster,
			"thumb_url":   poster,
			"actors":      actors,
		},
	}
	return candidate, nil
}

func splitDelimitedNames(raw string) []string {
	raw = strings.TrimSpace(stripHTMLText(raw))
	if raw == "" {
		return nil
	}
	parts := regexp.MustCompile(`[,，、/\s]+`).Split(raw, -1)
	return dedupeStrings(parts)
}

func minimalHTMLExternalID(rawURL string, site string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	switch site {
	case "giga":
		return strings.ToLower(strings.TrimSpace(parsed.Query().Get("product_id")))
	case "lulubar":
		return strings.ToLower(strings.TrimSpace(parsed.Query().Get("id")))
	case "kin8":
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 2 && parts[0] == "moviepages" {
			return strings.ToLower(strings.TrimSpace(parts[1]))
		}
	case "love6":
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}
	return strings.ToLower(extractSingleSitePathID(rawURL))
}

func (c *mywifeAVCrawler) Name() string {
	return "mywife"
}

func (c *mywifeAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
	return nil, nil
}

func (c *mywifeAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
	root, _, err := c.svc.fetchAVHTMLDocument(ctx, detailURL, nil)
	if err != nil {
		return avScrapeCandidate{}, err
	}
	if run != nil {
		run.addDetailURL(detailURL)
	}

	number, title := mywifeTitleParts(root)
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}

	code := ""
	if number != "" {
		code = "Mywife " + number
	}
	externalID := strings.ToLower(extractSingleSitePathID(detailURL))
	if code == "" && externalID != "" {
		code = "Mywife No." + externalID
	}
	poster := strings.TrimSpace(mywifeVideoAttr(root, "poster"))
	thumb := strings.Replace(poster, "topview.jpg", "thumb.jpg", 1)
	candidate := avScrapeCandidate{
		Source:     c.Name(),
		ExternalID: externalID,
		Code:       code,
		Title:      title,
		Overview:   mywifeOutline(root),
		PosterURL:  poster,
		ThumbURL:   thumb,
		Actors:     mywifeActors(root),
		DetailURL:  strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": externalID,
			"av_code":     code,
			"title":       title,
			"poster_url":  poster,
			"thumb_url":   thumb,
		},
	}
	return candidate, nil
}

func buildFC2PPVPathCode(externalID string) string {
	if number := normalizeFC2NumericID(externalID); number != "" {
		return "FC2-PPV-" + number
	}
	if number := normalizeFC2NumericID(extractAVCode(externalID)); number != "" {
		return "FC2-PPV-" + number
	}
	return ""
}

func extractMinimalHTMLHeadingTitle(root *html.Node) string {
	for _, tag := range []string{"h1", "h2", "h3"} {
		title := nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == tag
		}))
		if title != "" {
			return title
		}
	}
	return ""
}

func mywifeTitleParts(root *html.Node) (string, string) {
	title := nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "title"
	}))
	matches := regexp.MustCompile(`(?i)(No\.\d+)\s*(.+)`).FindStringSubmatch(title)
	if len(matches) != 3 {
		return "", normalizeWhitespace(title)
	}
	return normalizeWhitespace(matches[1]), normalizeWhitespace(matches[2])
}

func mywifeOutline(root *html.Node) string {
	return nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "modelsamplephototop")
	}))
}

func mywifeActors(root *html.Node) []string {
	values := make([]string, 0, 2)
	for _, node := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img" && n.Parent != nil && hasClass(n.Parent, "modelwaku0")
	}) {
		if name := normalizeWhitespace(attrValue(node, "alt")); name != "" {
			values = append(values, name)
		}
	}
	return dedupeStrings(values)
}

func mywifeVideoAttr(root *html.Node, key string) string {
	return strings.TrimSpace(findFirstAttr(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "video" && attrValue(n, "id") == "video"
	}, key))
}

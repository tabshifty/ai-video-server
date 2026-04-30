package services

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var aggregateSupplementStandardNumberRe = regexp.MustCompile(`(?i)\b[A-Z0-9]{2,10}-\d+\b`)

type aggregateSupplementAVCrawler struct {
	svc   *ScraperService
	site  string
	hosts []string
}

func newAggregateSupplementAVCrawler(svc *ScraperService, site string, hosts ...string) avCrawler {
	return &aggregateSupplementAVCrawler{
		svc:   svc,
		site:  site,
		hosts: append([]string(nil), hosts...),
	}
}

func newAirAVCCAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(svc, "airav_cc", "airav.io", "www.airav.io")
}

func newAVSexAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(
		svc,
		"avsex",
		"gg5.co",
		"www.gg5.co",
		"paycalling.com",
		"www.paycalling.com",
		"9sex.tv",
		"www.9sex.tv",
		"avsex.cc",
		"www.avsex.cc",
		"avsex.club",
		"www.avsex.club",
	)
}

func newCableAVAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(svc, "cableav", "cableav.tv", "www.cableav.tv")
}

func newHDOUBANAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(
		svc,
		"hdouban",
		"ormtgu.com",
		"www.ormtgu.com",
		"byym21.com",
		"www.byym21.com",
		"huangdb2.com",
		"www.huangdb2.com",
	)
}

func newHSCangkuAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(svc, "hscangku", "hscangku.net", "www.hscangku.net", "hsck.net", "www.hsck.net")
}

func newIQQTVAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(svc, "iqqtv", "iqq5.xyz", "www.iqq5.xyz", "iqqtv.cloud", "www.iqqtv.cloud")
}

func newSevenMMTVAVCrawler(svc *ScraperService) avCrawler {
	return newAggregateSupplementAVCrawler(svc, "7mmtv", "7mmtv.sx", "www.7mmtv.sx", "7mmtv.tv", "www.7mmtv.tv")
}

func (c *aggregateSupplementAVCrawler) Name() string {
	return c.site
}

func (c *aggregateSupplementAVCrawler) FetchByDetailURL(ctx context.Context, run *avScrapeRunContext, detailURL string) (avScrapeCandidate, error) {
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

	code := c.extractNumber(root)
	title := c.extractTitle(root, code)
	if strings.TrimSpace(title) == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}
	if strings.TrimSpace(code) == "" {
		code = extractAVCode(nodeText(root))
	}

	candidate := avScrapeCandidate{
		Source:     c.site,
		ExternalID: extractAggregateSupplementExternalID(c.site, detailURL),
		Code:       code,
		Title:      title,
		Actors:     c.extractActors(root),
		DetailURL:  strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.site,
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": extractAggregateSupplementExternalID(c.site, detailURL),
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *aggregateSupplementAVCrawler) SearchCandidates(ctx context.Context, run *avScrapeRunContext, query string, limit int) ([]avScrapeCandidate, error) {
	if c.site != "avsex" {
		return nil, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	base := c.svc.avSiteBaseURL("avsex", "https://gg5.co")
	searchURL := strings.TrimRight(base, "/") + "/tw/search?query=" + url.QueryEscape(strings.ToLower(query))
	if run != nil {
		run.addSearchURL(searchURL)
	}
	content, err := c.svc.fetchAVHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse avsex search html: %w", err)
	}
	link := findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && strings.Contains(attrValue(n, "href"), "/video/detail/")
	})
	if link == nil {
		return nil, nil
	}
	detailURL := resolveRelativeAVURL(base, attrValue(link, "href"))
	if detailURL == "" {
		return nil, nil
	}
	candidate, err := c.FetchByDetailURL(ctx, run, detailURL)
	if err != nil {
		return nil, err
	}
	if candidate.DetailURL == "" {
		candidate.DetailURL = detailURL
	}
	if candidate.PosterURL == "" {
		if posterNode := findFirst(link, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "img"
		}); posterNode != nil {
			candidate.PosterURL = resolveRelativeAVURL(base, attrValue(posterNode, "src"))
		}
	}
	return []avScrapeCandidate{candidate}, nil
}

func (c *aggregateSupplementAVCrawler) extractTitle(root *html.Node, number string) string {
	var title string
	switch c.site {
	case "airav_cc":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "video-title")
		}))
	case "avsex":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			if n.Type != html.ElementNode || n.Data != "span" {
				return false
			}
			classValue := attrValue(n, "class")
			return strings.Contains(classValue, "truncate") && strings.Contains(classValue, "font-bold")
		}))
	case "cableav":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "p" && n.Parent != nil && hasClass(n.Parent, "entry-content")
		}))
	case "hdouban":
		title = extractMinimalHTMLTitle(root)
	case "hscangku":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h3" && hasClass(n, "title")
		}))
	case "iqqtv":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h1" && hasClass(n, "h4") && hasClass(n, "b")
		}))
	case "7mmtv":
		title = nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h1" && hasClass(n, "fullvideo-title")
		}))
	}
	if title == "" {
		title = extractMinimalHTMLTitle(root)
	}
	return trimTitleByNumber(title, number)
}

func (c *aggregateSupplementAVCrawler) extractNumber(root *html.Node) string {
	switch c.site {
	case "7mmtv":
		if value := extractAggregateSupplementNumber(nodeText(findFirst(root, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "d-flex") && hasClass(n, "mb-4")
		}))); value != "" {
			return value
		}
	}
	return extractAggregateSupplementNumber(nodeText(root))
}

func (c *aggregateSupplementAVCrawler) extractActors(root *html.Node) []string {
	actors := make([]string, 0, 4)
	seen := map[string]struct{}{}
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			href := strings.ToLower(strings.TrimSpace(attrValue(n, "href")))
			if strings.Contains(href, "actor") || strings.Contains(href, "star") {
				name := normalizeWhitespace(nodeText(n))
				if name != "" {
					if _, ok := seen[name]; !ok {
						seen[name] = struct{}{}
						actors = append(actors, name)
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return actors
}

func extractAggregateSupplementExternalID(site string, rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	switch site {
	case "airav_cc":
		if hid := strings.TrimSpace(parsed.Query().Get("hid")); hid != "" {
			return strings.ToLower(hid)
		}
	case "avsex":
		if id := strings.TrimSpace(parsed.Query().Get("id")); id != "" {
			return strings.ToLower(id)
		}
	}
	return extractSingleSitePathID(rawURL)
}

func extractAggregateSupplementNumber(value string) string {
	value = normalizeWhitespace(value)
	if value == "" {
		return ""
	}

	if match := fc2NumberRe.FindString(value); match != "" {
		return normalizeAggregateSupplementNumber(match)
	}
	if match := aggregateSupplementStandardNumberRe.FindString(value); match != "" {
		return strings.ToUpper(strings.TrimSpace(match))
	}
	return ""
}

func normalizeAggregateSupplementNumber(value string) string {
	value = strings.ToUpper(normalizeWhitespace(value))
	if strings.HasPrefix(value, "FC2") {
		digitMatches := regexp.MustCompile(`\d+`).FindAllString(value, -1)
		if len(digitMatches) > 0 {
			return "FC2-PPV-" + digitMatches[len(digitMatches)-1]
		}
	}
	return strings.ReplaceAll(value, " ", "-")
}

func trimTitleByNumber(title string, number string) string {
	title = normalizeWhitespace(title)
	number = strings.TrimSpace(number)
	if title == "" || number == "" {
		return title
	}

	variants := []string{
		number,
		strings.ReplaceAll(number, "-PPV-", "-PPV "),
		strings.ReplaceAll(number, "-", " "),
		"[" + number + "]",
		"[" + strings.ReplaceAll(number, "-PPV-", "-PPV ") + "]",
	}
	for _, variant := range variants {
		variant = strings.TrimSpace(variant)
		if variant == "" {
			continue
		}
		title = strings.TrimSpace(strings.TrimPrefix(title, variant))
		title = strings.TrimSpace(strings.TrimSuffix(title, variant))
	}
	return normalizeWhitespace(title)
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func findFirst(root *html.Node, predicate func(n *html.Node) bool) *html.Node {
	if root == nil {
		return nil
	}
	if predicate(root) {
		return root
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := findFirst(child, predicate); found != nil {
			return found
		}
	}
	return nil
}

func hasClass(node *html.Node, className string) bool {
	className = strings.TrimSpace(className)
	if node == nil || className == "" {
		return false
	}
	for _, token := range strings.Fields(attrValue(node, "class")) {
		if token == className {
			return true
		}
	}
	return false
}

func attrValue(node *html.Node, key string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func nodeText(node *html.Node) string {
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
			if text := normalizeWhitespace(n.Data); text != "" {
				parts = append(parts, text)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return normalizeWhitespace(strings.Join(parts, " "))
}

func extractMinimalHTMLTitle(root *html.Node) string {
	title := nodeText(findFirst(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "title"
	}))
	title = strings.TrimSpace(strings.TrimSuffix(title, "- JavDB"))
	return normalizeWhitespace(title)
}

func extractSingleSitePathID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	trimmed := strings.Trim(parsed.Path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	last := strings.TrimSpace(parts[len(parts)-1])
	last = strings.TrimSuffix(last, ".html")
	if last != "" {
		return strings.ToLower(last)
	}
	return ""
}

func resolveRelativeAVURL(baseURL, rawPath string) string {
	return toAbsoluteURL(baseURL, rawPath)
}

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
	fc2PPVDBDetailPathRe   = regexp.MustCompile(`^/articles/fc2-ppv-\d+/?$`)
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

func (c *minimalHTMLAVCrawler) SearchCandidates(context.Context, *avScrapeRunContext, string, int) ([]avScrapeCandidate, error) {
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
	if title == "" {
		return avScrapeCandidate{}, fmt.Errorf("empty scrape result for url %q", detailURL)
	}

	externalID := minimalHTMLExternalID(detailURL, c.site)
	candidate := avScrapeCandidate{
		Source:     c.site,
		ExternalID: externalID,
		Code:       code,
		Title:      title,
		DetailURL:  strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.site,
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": externalID,
			"av_code":     code,
			"title":       title,
		},
	}
	return candidate, nil
}

func (c *minimalHTMLAVCrawler) extractCode(root *html.Node, detailURL string) string {
	content := nodeText(root)
	switch c.site {
	case "fc2club", "fc2hub", "fc2ppvdb", "airav", "freejavbt":
		if number := normalizeFC2NumericID(content); number != "" {
			return "FC2-PPV-" + number
		}
		if number := normalizeFC2NumericID(detailURL); number != "" {
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
			return strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))
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
	poster := strings.Replace(mywifeVideoAttr(root, "poster"), "topview.jpg", "thumb.jpg", 1)
	candidate := avScrapeCandidate{
		Source:     c.Name(),
		ExternalID: externalID,
		Code:       code,
		Title:      title,
		Overview:   mywifeOutline(root),
		PosterURL:  poster,
		Actors:     mywifeActors(root),
		DetailURL:  strings.TrimSpace(detailURL),
		Raw: map[string]any{
			"site":        c.Name(),
			"detail_url":  strings.TrimSpace(detailURL),
			"external_id": externalID,
			"av_code":     code,
			"title":       title,
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

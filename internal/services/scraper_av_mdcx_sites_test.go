package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMDCxMigratedSitesSearchCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch {
		case strings.Contains(r.URL.Path, "/dmm-search"):
			_, _ = w.Write([]byte(`<html><body><a href="/mono/dvd/-/detail/=/cid=ssis001/">SSIS-001</a></body></html>`))
		case r.URL.Path == "/mono/dvd/-/detail/=/cid=ssis001/":
			_, _ = w.Write([]byte(`<html><head><meta property="og:image" content="` + serverURLFromRequest(r) + `/dmm/ssis001ps.jpg"></head><body><h1>DMM Title</h1><table><tr><th>Number</th><td>SSIS-001</td></tr><tr><th>Release Date</th><td>2024/01/02</td></tr></table></body></html>`))
		case r.URL.Path == "/product/product_detail/300MIUM-382/":
			_, _ = w.Write([]byte(`<html><body><div id="center_column"><h1>MGStage Title</h1><table><tr><th>品番</th><td>300MIUM-382</td></tr><tr><th>配信開始日</th><td>2024/02/03</td></tr><tr><th>出演</th><td><a>Actor A</a></td></tr></table><a id="EnlargeImage" href="` + serverURLFromRequest(r) + `/mgstage/pb_300mium382.jpg">cover</a><div id="introduction"><dd><p>MG outline</p></dd></div></div></body></html>`))
		case r.URL.Path == "/search" && r.Method == http.MethodPost:
			if err := r.ParseForm(); err != nil || strings.TrimSpace(r.PostForm.Get("sn")) != "SSIS-001" {
				http.Error(w, "unexpected jav321 search form", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`<html><body><h3>SSIS-001 Jav321 Title <small>sample</small></h3><a href="/video/abc123">detail</a><p>品番: SSIS-001</p><p>出演者: Actor B</p><img class="img-responsive" src="` + serverURLFromRequest(r) + `/jav321/thumb.jpg"></body></html>`))
		case r.URL.Path == "/articles/FC2-PPV-3259498" || r.URL.Path == "/articles/3259498":
			_, _ = w.Write([]byte(`<html><body><h2><a>FC2PPVDB Title</a></h2><img src="` + serverURLFromRequest(r) + `/fc2ppvdb/3259498.jpg"><p>販売日：2024-03-04</p></body></html>`))
		case r.URL.Path == "/html/FC2-743423.html":
			_, _ = w.Write([]byte(`<html><body><h3>FC2-743423 FC2Club Title</h3><div class="items_article_MainitemThumb"><img src="` + serverURLFromRequest(r) + `/fc2club/743423.jpg"></div><div class="col des">FC2Club outline</div></body></html>`))
		case r.URL.Path == "/search":
			query := r.URL.Query().Get("kw")
			_, _ = w.Write([]byte(`<html><head><link href="` + serverURLFromRequest(r) + `/detail/` + url.PathEscape(query) + `/"></head><body><a href="/detail/` + query + `/">FC2Hub result</a></body></html>`))
		case r.URL.Path == "/detail/1940476/":
			_, _ = w.Write([]byte(`<html><body><h1>FC2Hub Title</h1><img src="` + serverURLFromRequest(r) + `/fc2hub/1940476.jpg"><p>FC2-1940476</p></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "", "", "", time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"dmm":      server.URL + "/dmm-search",
			"mgstage":  server.URL,
			"jav321":   server.URL,
			"fc2ppvdb": server.URL,
			"fc2club":  server.URL,
			"fc2hub":   server.URL,
		},
		Timeout: time.Second,
	})

	cases := []struct {
		name       string
		crawler    avCrawler
		query      string
		wantTitle  string
		wantCode   string
		wantPoster string
	}{
		{name: "dmm", crawler: newDMMAVCrawler(svc), query: "SSIS-001", wantTitle: "DMM Title", wantCode: "SSIS-001", wantPoster: server.URL + "/dmm/ssis001ps.jpg"},
		{name: "mgstage", crawler: newMGStageAVCrawler(svc), query: "300MIUM-382", wantTitle: "MGStage Title", wantCode: "300MIUM-382", wantPoster: server.URL + "/mgstage/pf_300mium382.jpg"},
		{name: "jav321", crawler: newJav321AVCrawler(svc), query: "SSIS-001", wantTitle: "SSIS-001 Jav321 Title", wantCode: "SSIS-001", wantPoster: server.URL + "/jav321/thumb.jpg"},
		{name: "fc2ppvdb", crawler: newFC2PPVDBAVCrawler(svc), query: "FC2-PPV-3259498", wantTitle: "FC2PPVDB Title", wantCode: "FC2-PPV-3259498", wantPoster: server.URL + "/fc2ppvdb/3259498.jpg"},
		{name: "fc2club", crawler: newFC2ClubAVCrawler(svc), query: "FC2PPV-743423", wantTitle: "FC2Club Title", wantCode: "FC2-PPV-743423", wantPoster: server.URL + "/fc2club/743423.jpg"},
		{name: "fc2hub", crawler: newFC2HubAVCrawler(svc), query: "FC2-1940476", wantTitle: "FC2Hub Title", wantCode: "FC2-PPV-1940476", wantPoster: server.URL + "/fc2hub/1940476.jpg"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hits, err := tc.crawler.SearchCandidates(context.Background(), newAVScrapeRunContext(tc.query, tc.wantCode), tc.query, 3)
			if err != nil {
				t.Fatalf("SearchCandidates returned error: %v", err)
			}
			if len(hits) != 1 {
				t.Fatalf("expected one hit, got=%d hits=%v", len(hits), hits)
			}
			got := hits[0]
			if got.Title != tc.wantTitle {
				t.Fatalf("expected title %q, got %q", tc.wantTitle, got.Title)
			}
			if got.Code != tc.wantCode {
				t.Fatalf("expected code %q, got %q", tc.wantCode, got.Code)
			}
			if got.PosterURL != tc.wantPoster {
				t.Fatalf("expected poster %q, got %q", tc.wantPoster, got.PosterURL)
			}
		})
	}
}

func TestDMMSearchCandidatesReturnsRegionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><p>Sorry! This content is not available in your region.</p></body></html>`))
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "", "", "", time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"dmm": server.URL,
		},
		Timeout: time.Second,
	})

	_, err := newDMMAVCrawler(svc).SearchCandidates(context.Background(), newAVScrapeRunContext("SSIS-001", "SSIS-001"), "SSIS-001", 1)
	if err == nil || !strings.Contains(err.Error(), "content is not available in this region") {
		t.Fatalf("expected region error, got=%v", err)
	}
}

func TestDMMSearchCandidatesSkipsFailedSearchURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/search/=/searchstr=ssis00001/sort=ranking/":
			http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
		case "/search/=/searchstr=ssis001/sort=ranking/":
			_, _ = w.Write([]byte(`<html><body><a href="/digital/videoa/-/detail/=/cid=ssis00001/">SSIS-001</a></body></html>`))
		case "/digital/videoa/-/detail/=/cid=ssis00001/":
			_, _ = w.Write([]byte(`<html><head><meta property="og:image" content="` + serverURLFromRequest(r) + `/dmm/ssis001ps.jpg"></head><body><h1>DMM Fallback Title</h1><table><tr><th>Number</th><td>SSIS-001</td></tr></table></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewScraperService(nil, "", "", "", "", time.Second)
	svc.ConfigureAVScraperConfig(AVScraperConfig{
		BaseURL: server.URL,
		SiteURLs: map[string]string{
			"dmm": server.URL,
		},
		Timeout: time.Second,
	})

	hits, err := newDMMAVCrawler(svc).SearchCandidates(context.Background(), newAVScrapeRunContext("SSIS-001", "SSIS-001"), "SSIS-001", 1)
	if err != nil {
		t.Fatalf("SearchCandidates returned error: %v", err)
	}
	if len(hits) != 1 || hits[0].Title != "DMM Fallback Title" {
		t.Fatalf("expected fallback hit, got=%v", hits)
	}
}

package services

import (
	"net/url"
	"testing"
)

func TestThePornDBSearchKeywordsExpandsSupportedDateFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filePath   string
		wantSeries string
		wantDate   string
	}{
		{
			name:       "dot separated two digit year",
			filePath:   "Wgp.20.05.12.Title.mp4",
			wantSeries: "whengirlsplay",
			wantDate:   "2020-05-12",
		},
		{
			name:       "underscore separated four digit year",
			filePath:   "Wgp_2020-05-12_Title.mp4",
			wantSeries: "whengirlsplay",
			wantDate:   "2020-05-12",
		},
		{
			name:       "bracketed series with two digit year",
			filePath:   "[Wgp] 20.05.12.Title.mp4",
			wantSeries: "whengirlsplay",
			wantDate:   "2020-05-12",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keywords, series, date := thePornDBSearchKeywords(tt.filePath)

			if series != tt.wantSeries {
				t.Fatalf("series = %q, want %q", series, tt.wantSeries)
			}
			if date != tt.wantDate {
				t.Fatalf("date = %q, want %q", date, tt.wantDate)
			}
			if len(keywords) == 0 || keywords[0] != tt.wantSeries+" "+tt.wantDate {
				t.Fatalf("keywords = %#v, want first %q", keywords, tt.wantSeries+" "+tt.wantDate)
			}
		})
	}
}

func TestThePornDBSearchURLUsesEndpointSpecificQueryParameter(t *testing.T) {
	t.Parallel()

	scenesURL := thePornDBSearchURL("https://api.example.test", "scenes", "WhenGirlsPlay 2020-05-12")
	moviesURL := thePornDBSearchURL("https://api.example.test", "movies", "WhenGirlsPlay 2020-05-12")

	scenesValues := queryValues(t, scenesURL)
	moviesValues := queryValues(t, moviesURL)

	if got := scenesValues.Get("parse"); got != "WhenGirlsPlay 2020-05-12" {
		t.Fatalf("scenes parse = %q", got)
	}
	if got := scenesValues.Get("q"); got != "" {
		t.Fatalf("scenes q = %q, want empty", got)
	}
	if got := moviesValues.Get("q"); got != "WhenGirlsPlay 2020-05-12" {
		t.Fatalf("movies q = %q", got)
	}
	if got := moviesValues.Get("parse"); got != "" {
		t.Fatalf("movies parse = %q, want empty", got)
	}
	if scenesValues.Get("per_page") != "100" || moviesValues.Get("per_page") != "100" {
		t.Fatalf("per_page not preserved: scenes=%q movies=%q", scenesValues.Get("per_page"), moviesValues.Get("per_page"))
	}
}

func queryValues(t *testing.T, rawURL string) url.Values {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse URL %q: %v", rawURL, err)
	}
	return parsed.Query()
}

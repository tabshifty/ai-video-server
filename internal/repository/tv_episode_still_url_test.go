package repository

import "testing"

func TestResolveTVEpisodeStillURL(t *testing.T) {
	t.Parallel()

	got := resolveTVEpisodeStillURL(42, 1, 2, "/remote/still.jpg")
	want := "/api/v1/tv/series/42/seasons/1/episodes/2/still"
	if got != want {
		t.Fatalf("unexpected still url: got=%s want=%s", got, want)
	}
}

func TestResolveTVEpisodeStillURLRejectsMissingInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		seriesID      int64
		seasonNumber  int
		episodeNumber int
		raw           string
	}{
		{name: "blank raw", seriesID: 42, seasonNumber: 1, episodeNumber: 2, raw: " "},
		{name: "zero series", seriesID: 0, seasonNumber: 1, episodeNumber: 2, raw: "/still.jpg"},
		{name: "zero season", seriesID: 42, seasonNumber: 0, episodeNumber: 2, raw: "/still.jpg"},
		{name: "zero episode", seriesID: 42, seasonNumber: 1, episodeNumber: 0, raw: "/still.jpg"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveTVEpisodeStillURL(tc.seriesID, tc.seasonNumber, tc.episodeNumber, tc.raw)
			if got != "" {
				t.Fatalf("expected blank url, got=%s", got)
			}
		})
	}
}

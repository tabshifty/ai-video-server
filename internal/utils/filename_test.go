package utils

import "testing"

func TestExtractTitleYear(t *testing.T) {
	title, year, ok := ExtractTitleYear("Dune (2021).mkv")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if title != "Dune" {
		t.Fatalf("unexpected title: %s", title)
	}
	if year != 2021 {
		t.Fatalf("unexpected year: %d", year)
	}
}

func TestParseSeriesEpisode(t *testing.T) {
	name, season, episode, ok := ParseSeriesEpisode("Severance S02E03.mp4")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if name != "Severance" || season != 2 || episode != 3 {
		t.Fatalf("unexpected parsed result: %s S%02dE%02d", name, season, episode)
	}
}

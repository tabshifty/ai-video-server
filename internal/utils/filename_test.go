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

func TestParseFilename_Movie(t *testing.T) {
	title, year, season, episode, ok := ParseFilename("Inception (2010).mp4")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if title != "Inception" || year != 2010 || season != 0 || episode != 0 {
		t.Fatalf("unexpected parse result: title=%q year=%d season=%d episode=%d", title, year, season, episode)
	}
}

func TestParseFilename_SeriesWithDots(t *testing.T) {
	title, year, season, episode, ok := ParseFilename("Game.of.Thrones.S01E01.mp4")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if title != "Game of Thrones" || year != 0 || season != 1 || episode != 1 {
		t.Fatalf("unexpected parse result: title=%q year=%d season=%d episode=%d", title, year, season, episode)
	}
}

func TestParseFilename_SeriesWithSpace(t *testing.T) {
	title, year, season, episode, ok := ParseFilename("剧名 S01E02.mkv")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if title != "剧名" || year != 0 || season != 1 || episode != 2 {
		t.Fatalf("unexpected parse result: title=%q year=%d season=%d episode=%d", title, year, season, episode)
	}
}

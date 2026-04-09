package utils

import (
	"strings"
)

// ExtractTitleYear parses `Title (2024)` format from file name.
func ExtractTitleYear(name string) (string, int, bool) {
	title, year, season, episode, ok := ParseFilename(name)
	if !ok || season > 0 || episode > 0 || year <= 0 {
		return "", 0, false
	}
	return strings.TrimSpace(title), year, true
}

// ParseSeriesEpisode parses `Show Name S01E02` format from file name.
func ParseSeriesEpisode(name string) (string, int, int, bool) {
	title, _, season, episode, ok := ParseFilename(name)
	if !ok || season <= 0 || episode <= 0 {
		return "", 0, 0, false
	}
	return strings.TrimSpace(title), season, episode, true
}

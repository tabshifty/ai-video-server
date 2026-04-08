package utils

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	titleYearPattern = regexp.MustCompile(`(?i)^(.+?)\s*\((\d{4})\)$`)
	seriesPattern    = regexp.MustCompile(`(?i)^(.+?)\s+[s](\d{1,2})[e](\d{1,3})$`)
)

// ExtractTitleYear parses `Title (2024)` format from file name.
func ExtractTitleYear(name string) (string, int, bool) {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	match := titleYearPattern.FindStringSubmatch(strings.TrimSpace(base))
	if len(match) != 3 {
		return "", 0, false
	}
	year, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, false
	}
	return strings.TrimSpace(match[1]), year, true
}

// ParseSeriesEpisode parses `Show Name S01E02` format from file name.
func ParseSeriesEpisode(name string) (string, int, int, bool) {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	match := seriesPattern.FindStringSubmatch(strings.TrimSpace(base))
	if len(match) != 4 {
		return "", 0, 0, false
	}
	season, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, 0, false
	}
	episode, err := strconv.Atoi(match[3])
	if err != nil {
		return "", 0, 0, false
	}
	return strings.TrimSpace(match[1]), season, episode, true
}

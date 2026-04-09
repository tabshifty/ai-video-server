package utils

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	filenameTitleYearPattern = regexp.MustCompile(`(?i)^(.+?)\s*\((\d{4})\)$`)
	filenameSeriesPattern    = regexp.MustCompile(`(?i)^(.+?)[\s._-]*s(\d{1,2})e(\d{1,3})(?:[\s._-].*)?$`)
)

// ParseFilename parses movie and TV episode names from filename.
func ParseFilename(filename string) (title string, year int, season int, episode int, ok bool) {
	base := strings.TrimSpace(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	if base == "" {
		return "", 0, 0, 0, false
	}

	if m := filenameSeriesPattern.FindStringSubmatch(base); len(m) == 4 {
		s, sErr := strconv.Atoi(m[2])
		e, eErr := strconv.Atoi(m[3])
		if sErr == nil && eErr == nil {
			return normalizeTitle(m[1]), 0, s, e, true
		}
	}

	normalized := normalizeTitle(base)
	if m := filenameTitleYearPattern.FindStringSubmatch(normalized); len(m) == 3 {
		y, err := strconv.Atoi(m[2])
		if err == nil {
			return normalizeTitle(m[1]), y, 0, 0, true
		}
	}

	return "", 0, 0, 0, false
}

func normalizeTitle(raw string) string {
	replaced := strings.NewReplacer(".", " ", "_", " ").Replace(strings.TrimSpace(raw))
	return strings.Join(strings.Fields(replaced), " ")
}

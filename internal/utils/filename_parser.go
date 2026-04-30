package utils

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	filenameTitleYearPattern  = regexp.MustCompile(`(?i)^(.+?)\s*\((\d{4})\)$`)
	filenameSeriesPattern     = regexp.MustCompile(`(?i)^(.+?)[\s._-]*s(\d{1,2})e(\d{1,3})(?:[\s._-].*)?$`)
	filenameSeriesCNPattern   = regexp.MustCompile(`(?i)^(.+?)[\s._-]*(?:第?\s*(\d{1,2})\s*季)\s*(?:第?\s*(\d{1,3})\s*[集话話])(?:[\s._-].*)?$`)
	filenameSeriesCNCNPattern = regexp.MustCompile(`(?i)^(.+?)[\s._-]*(?:第?\s*([零〇一二两三四五六七八九十百千]+)\s*季)\s*(?:第?\s*([零〇一二两三四五六七八九十百千]+)\s*[集话話])(?:[\s._-].*)?$`)
)

// ParseFilename parses movie and TV episode names from filename.
func ParseFilename(filename string) (title string, year int, season int, episode int, ok bool) {
	base := strings.TrimSpace(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	if base == "" {
		return "", 0, 0, 0, false
	}

	if title, season, episode, ok := parseSeriesEpisodeBase(base); ok {
		return title, 0, season, episode, true
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

func parseSeriesEpisodeBase(base string) (title string, season int, episode int, ok bool) {
	if m := filenameSeriesPattern.FindStringSubmatch(base); len(m) == 4 {
		s, sErr := strconv.Atoi(m[2])
		e, eErr := strconv.Atoi(m[3])
		if sErr == nil && eErr == nil {
			return normalizeTitle(m[1]), s, e, true
		}
	}

	if m := filenameSeriesCNPattern.FindStringSubmatch(base); len(m) == 4 {
		s, sErr := strconv.Atoi(m[2])
		e, eErr := strconv.Atoi(m[3])
		if sErr == nil && eErr == nil {
			return normalizeTitle(m[1]), s, e, true
		}
	}

	if m := filenameSeriesCNCNPattern.FindStringSubmatch(base); len(m) == 4 {
		s, sOK := parseChineseEpisodeNumber(m[2])
		e, eOK := parseChineseEpisodeNumber(m[3])
		if sOK && eOK {
			return normalizeTitle(m[1]), s, e, true
		}
	}

	return "", 0, 0, false
}

func parseChineseEpisodeNumber(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}

	if value, err := strconv.Atoi(raw); err == nil {
		if value > 0 {
			return value, true
		}
		return 0, false
	}

	values := map[rune]int{
		'零': 0,
		'〇': 0,
		'一': 1,
		'二': 2,
		'两': 2,
		'三': 3,
		'四': 4,
		'五': 5,
		'六': 6,
		'七': 7,
		'八': 8,
		'九': 9,
	}

	total := 0
	current := 0
	for _, r := range raw {
		switch r {
		case '十':
			if current == 0 {
				current = 1
			}
			total += current * 10
			current = 0
		case '百':
			if current == 0 {
				current = 1
			}
			total += current * 100
			current = 0
		case '千':
			if current == 0 {
				current = 1
			}
			total += current * 1000
			current = 0
		default:
			value, ok := values[r]
			if !ok {
				return 0, false
			}
			current += value
		}
	}

	total += current
	if total <= 0 {
		return 0, false
	}
	return total, true
}

func normalizeTitle(raw string) string {
	replaced := strings.NewReplacer(".", " ", "_", " ").Replace(strings.TrimSpace(raw))
	return strings.Join(strings.Fields(replaced), " ")
}

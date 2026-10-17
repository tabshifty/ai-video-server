// Package cloud drives the locally installed 115 browser to submit ed2k/magnet
// links to 115 cloud download (offline download) and to pull the finished file
// down to the local download directory.
//
// It talks to the browser through the Chrome DevTools Protocol only; no
// credential is read or stored here. The browser session owns authentication.
package cloud

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Kind identifies a supported remote-link protocol.
type Kind string

const (
	KindED2K   Kind = "ed2k"
	KindMagnet Kind = "magnet"
)

// Link is one parsed remote download link.
type Link struct {
	Kind      Kind
	InfoHash  string // lowercase hex: 32 chars for ed2k (MD4), 40 chars for magnet (BTIH)
	Name      string
	SizeBytes *int64 // nil when the link does not declare a size
	Raw       string
}

var (
	// ErrEmptyLink means the input had no content.
	ErrEmptyLink = errors.New("cloud: empty link")
	// ErrUnsupportedLink means the scheme is not ed2k/magnet.
	ErrUnsupportedLink = errors.New("cloud: unsupported link scheme")
	// ErrMalformedLink means the scheme matched but the payload is invalid.
	ErrMalformedLink = errors.New("cloud: malformed link")
)

var (
	ed2kHashRe = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	btihHexRe  = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	btihB32Re  = regexp.MustCompile(`^[A-Za-z2-7]{32}$`)
)

// ParseLink parses one ed2k:// or magnet:? link.
func ParseLink(raw string) (Link, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Link{}, ErrEmptyLink
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "ed2k://"):
		return parseED2K(trimmed)
	case strings.HasPrefix(lower, "magnet:?"):
		return parseMagnet(trimmed)
	default:
		return Link{}, ErrUnsupportedLink
	}
}

// ParseLinks parses a batch, returning parsed links and per-line errors.
// Invalid lines are reported rather than silently dropped.
func ParseLinks(raws []string) ([]Link, []error) {
	links := make([]Link, 0, len(raws))
	errs := make([]error, 0)
	for i, raw := range raws {
		link, err := ParseLink(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", i+1, err))
			continue
		}
		links = append(links, link)
	}
	return links, errs
}

// parseED2K handles ed2k://|file|NAME|SIZE|HASH|/ and the variant without a
// trailing slash. NAME may contain literal spaces and %XX escapes.
func parseED2K(raw string) (Link, error) {
	body := raw[len("ed2k://"):]
	parts := strings.Split(body, "|")
	// parts[0] is empty for the canonical form; tolerate a missing leading pipe.
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) < 4 {
		return Link{}, fmt.Errorf("%w: expected at least file|name|size|hash", ErrMalformedLink)
	}
	if !strings.EqualFold(parts[0], "file") {
		return Link{}, fmt.Errorf("%w: unsupported ed2k kind %q", ErrMalformedLink, parts[0])
	}
	name, err := url.PathUnescape(parts[1])
	if err != nil {
		name = parts[1]
	}
	if !ed2kHashRe.MatchString(parts[3]) {
		return Link{}, fmt.Errorf("%w: invalid ed2k hash", ErrMalformedLink)
	}
	link := Link{
		Kind:     KindED2K,
		InfoHash: strings.ToLower(parts[3]),
		Name:     strings.TrimSpace(name),
		Raw:      raw,
	}
	if size, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err == nil && size > 0 {
		link.SizeBytes = &size
	}
	return link, nil
}

// parseMagnet handles magnet:?xt=urn:btih:<hash>&dn=<name>&xl=<size>.
// The BTIH may be 40-char hex or 32-char base32.
func parseMagnet(raw string) (Link, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return Link{}, fmt.Errorf("%w: %v", ErrMalformedLink, err)
	}
	query := parsed.Query()
	var infoHash string
	for _, xt := range query["xt"] {
		const prefix = "urn:btih:"
		if len(xt) < len(prefix) || !strings.EqualFold(xt[:len(prefix)], prefix) {
			continue
		}
		infoHash = normalizeBTIH(xt[len(prefix):])
		if infoHash != "" {
			break
		}
	}
	if infoHash == "" {
		return Link{}, fmt.Errorf("%w: missing urn:btih", ErrMalformedLink)
	}
	link := Link{
		Kind:     KindMagnet,
		InfoHash: infoHash,
		Name:     strings.TrimSpace(query.Get("dn")),
		Raw:      raw,
	}
	if size, err := strconv.ParseInt(strings.TrimSpace(query.Get("xl")), 10, 64); err == nil && size > 0 {
		link.SizeBytes = &size
	}
	return link, nil
}

// normalizeBTIH returns lowercase 40-char hex for either hex or base32 input.
func normalizeBTIH(value string) string {
	value = strings.TrimSpace(value)
	switch {
	case btihHexRe.MatchString(value):
		return strings.ToLower(value)
	case btihB32Re.MatchString(value):
		decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(value))
		if err != nil || len(decoded) != 20 {
			return ""
		}
		return fmt.Sprintf("%x", decoded)
	default:
		return ""
	}
}

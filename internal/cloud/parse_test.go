package cloud

import (
	"errors"
	"strings"
	"testing"
)

func TestParseLinkED2K(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		wantHash string
		wantName string
		wantSize int64
		wantErr  error
	}{
		{
			name:     "canonical with trailing slash",
			raw:      "ed2k://|file|demo.zip|1718424557|8D045AE604E1B799B95D4BCC7CA6C167|/",
			wantHash: "8d045ae604e1b799b95d4bcc7ca6c167",
			wantName: "demo.zip",
			wantSize: 1718424557,
		},
		{
			name:     "no trailing slash",
			raw:      "ed2k://|file|demo.zip|123|0123456789abcdef0123456789abcdef|",
			wantHash: "0123456789abcdef0123456789abcdef",
			wantName: "demo.zip",
			wantSize: 123,
		},
		{
			name:     "literal spaces and percent escapes in name",
			raw:      "ed2k://|file|www.98T.la@ ST-985 %E7%BB%9D%E7%BE%8E.zip|5354568012|56E72BBE67AE75D99541808D6366BBDF|",
			wantHash: "56e72bbe67ae75d99541808d6366bbdf",
			wantName: "www.98T.la@ ST-985 绝美.zip",
			wantSize: 5354568012,
		},
		{
			name:    "invalid hash",
			raw:     "ed2k://|file|demo.zip|123|zzzz|/",
			wantErr: ErrMalformedLink,
		},
		{
			name:    "missing fields",
			raw:     "ed2k://|file|demo.zip|",
			wantErr: ErrMalformedLink,
		},
		{
			name:    "not a file",
			raw:     "ed2k://|server|1.2.3.4|4242|/",
			wantErr: ErrMalformedLink,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			link, err := ParseLink(tc.raw)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if link.Kind != KindED2K {
				t.Fatalf("kind = %q, want ed2k", link.Kind)
			}
			if link.InfoHash != tc.wantHash {
				t.Fatalf("hash = %q, want %q", link.InfoHash, tc.wantHash)
			}
			if tc.wantName != "" && link.Name != tc.wantName {
				t.Fatalf("name = %q, want %q", link.Name, tc.wantName)
			}
			if link.SizeBytes == nil || *link.SizeBytes != tc.wantSize {
				t.Fatalf("size = %v, want %d", link.SizeBytes, tc.wantSize)
			}
		})
	}
}

func TestParseLinkMagnet(t *testing.T) {
	const hex = "0123456789abcdef0123456789abcdef01234567"
	link, err := ParseLink("magnet:?xt=urn:btih:" + strings.ToUpper(hex) + "&dn=Example&xl=1024")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Kind != KindMagnet {
		t.Fatalf("kind = %q", link.Kind)
	}
	if link.InfoHash != hex {
		t.Fatalf("hash = %q, want %q", link.InfoHash, hex)
	}
	if link.Name != "Example" {
		t.Fatalf("name = %q", link.Name)
	}
	if link.SizeBytes == nil || *link.SizeBytes != 1024 {
		t.Fatalf("size = %v", link.SizeBytes)
	}

	// base32 form of the same 20 bytes must normalize to the same hex hash.
	b32 := "AERUKZ4JVPG66AJDIVTYTK6N54ASGRLH"
	decoded, err := ParseLink("magnet:?xt=urn:btih:" + b32)
	if err != nil {
		t.Fatalf("base32: unexpected error: %v", err)
	}
	if decoded.InfoHash != hex {
		t.Fatalf("base32 hash = %q, want %q", decoded.InfoHash, hex)
	}
}

func TestParseLinkRejects(t *testing.T) {
	for _, raw := range []string{"", "   ", "http://example.com/a.zip", "magnet:?dn=no-hash"} {
		if _, err := ParseLink(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestParseLinksReportsLineNumbers(t *testing.T) {
	links, errs := ParseLinks([]string{
		"ed2k://|file|a.zip|1|0123456789abcdef0123456789abcdef|/",
		"nonsense",
		"ed2k://|file|b.zip|2|0123456789abcdef0123456789abcdef|/",
	})
	if len(links) != 2 {
		t.Fatalf("links = %d, want 2", len(links))
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "line 2") {
		t.Fatalf("errs = %v, want one error for line 2", errs)
	}
}

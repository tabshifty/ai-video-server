package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestParseSeenStateOnlyAcceptsExplicitTIDKeys(t *testing.T) {
	t.Parallel()

	tids, ignored, err := parseSeenState(strings.NewReader(`{
  "updated_at": 1,
  "seen_threads": [
    "tid:123", "https://sehuatang.org/thread-456-1-1.html|hash", "tid:123", "tid:789"
  ]
}`))
	if err != nil {
		t.Fatalf("parseSeenState returned error: %v", err)
	}
	if strings.Join(tids, ",") != "123,789" || ignored != 2 {
		t.Fatalf("tids=%v ignored=%d", tids, ignored)
	}
}

func TestParseSeenStateRejectsMalformedTIDKey(t *testing.T) {
	t.Parallel()

	_, _, err := parseSeenState(strings.NewReader(`{"seen_threads":["tid:not-number"]}`))
	if err == nil {
		t.Fatal("parseSeenState expected malformed tid error")
	}
}

func TestRunImportUsesBatchesAndCountsDispositions(t *testing.T) {
	t.Parallel()

	tids := make([]string, 0, 101)
	for i := 1; i <= 101; i++ {
		tids = append(tids, string(rune(1000+i)))
	}
	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Header.Get("Authorization") != "Bearer machine-token" {
			t.Fatalf("unexpected Authorization header: %q", req.Header.Get("Authorization"))
		}
		var requestBody struct {
			Posts []struct {
				TID string `json:"tid"`
			} `json:"posts"`
		}
		if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		data := make([]map[string]string, len(requestBody.Posts))
		for i, post := range requestBody.Posts {
			disposition := "created"
			if requests == 2 {
				disposition = "duplicate"
			}
			data[i] = map[string]string{"tid": post.TID, "id": "00000000-0000-0000-0000-000000000001", "disposition": disposition, "inspection_status": "dedupe_only"}
		}
		encoded, _ := json.Marshal(map[string]any{"code": 0, "msg": "ok", "data": data})
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(encoded))), Header: make(http.Header)}, nil
	})}

	report, err := runImport(context.Background(), client, importOptions{
		BaseURL: "http://127.0.0.1:8080", Token: "machine-token", Source: "sehuatang", BoardKey: "95", BatchSize: 100,
	}, tids)
	if err != nil {
		t.Fatalf("runImport returned error: %v", err)
	}
	if requests != 2 || report.Submitted != 101 || report.Created != 100 || report.Duplicate != 1 {
		t.Fatalf("requests=%d report=%+v", requests, report)
	}
}

func TestRunImportRejectsMismatchedResponseTID(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		body := `{"code":0,"msg":"ok","data":[{"tid":"other","id":"00000000-0000-0000-0000-000000000001","disposition":"created","inspection_status":"dedupe_only"}]}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	_, err := runImport(context.Background(), client, importOptions{
		BaseURL: "http://127.0.0.1:8080", Token: "machine-token", Source: "sehuatang", BoardKey: "95", BatchSize: 100,
	}, []string{"123"})
	if err == nil {
		t.Fatal("runImport expected mismatched tid error")
	}
}

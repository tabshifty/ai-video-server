package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAITextTranslatorTranslateScrapeContent(t *testing.T) {
	t.Parallel()

	var gotRequest openAIChatCompletionRequest
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(openAIChatCompletionResponse{
			Choices: []openAIChatCompletionChoice{{
				Message: openAIChatMessage{Content: `{"title_zh":"花与108位女孩","description_zh":"一部日本电影简介"}`},
			}},
		})
	}))
	defer server.Close()

	client := NewOpenAITextTranslator(TranslationConfig{
		APIURL:  server.URL + "/v1",
		APIKey:  "token-123",
		Model:   "HY-MT1.5-1.8B",
		Timeout: time.Second,
	})

	got, err := client.TranslateScrapeContent(context.Background(), "A Flower and 108 Girls", "Japanese synopsis")
	if err != nil {
		t.Fatalf("TranslateScrapeContent returned error: %v", err)
	}

	if got.Title != "花与108位女孩" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
	if got.Description != "一部日本电影简介" {
		t.Fatalf("unexpected description: %s", got.Description)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if gotRequest.Model != "HY-MT1.5-1.8B" {
		t.Fatalf("unexpected model: %s", gotRequest.Model)
	}
	if len(gotRequest.Messages) != 2 {
		t.Fatalf("unexpected messages length: %d", len(gotRequest.Messages))
	}
	if gotRequest.Messages[0].Role != "system" {
		t.Fatalf("unexpected first message role: %s", gotRequest.Messages[0].Role)
	}
	if gotRequest.Messages[1].Role != "user" {
		t.Fatalf("unexpected second message role: %s", gotRequest.Messages[1].Role)
	}
	if !strings.Contains(gotRequest.Messages[1].Content, `"title":"A Flower and 108 Girls"`) {
		t.Fatalf("request payload did not include title JSON: %s", gotRequest.Messages[1].Content)
	}
}

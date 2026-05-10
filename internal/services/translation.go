package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TranslationConfig struct {
	APIURL  string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type TranslationResult struct {
	Title       string
	Description string
}

type contentTranslator interface {
	TranslateScrapeContent(ctx context.Context, title, description string) (TranslationResult, error)
}

type OpenAITextTranslator struct {
	chatURL string
	apiKey  string
	model   string
	client  *http.Client
}

type openAIChatMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content"`
}

type openAIChatCompletionRequest struct {
	Model          string              `json:"model"`
	Messages       []openAIChatMessage `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat map[string]string   `json:"response_format,omitempty"`
}

type openAIChatCompletionResponse struct {
	Choices []openAIChatCompletionChoice `json:"choices"`
}

type openAIChatCompletionChoice struct {
	Message openAIChatMessage `json:"message"`
}

type scrapeTranslationPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type scrapeTranslationResponse struct {
	TitleZH       string `json:"title_zh"`
	DescriptionZH string `json:"description_zh"`
}

func NewOpenAITextTranslator(cfg TranslationConfig) *OpenAITextTranslator {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "HY-MT1.5-1.8B"
	}
	return &OpenAITextTranslator{
		chatURL: normalizeOpenAIChatURL(cfg.APIURL),
		apiKey:  strings.TrimSpace(cfg.APIKey),
		model:   model,
		client:  &http.Client{Timeout: timeout},
	}
}

func (t *OpenAITextTranslator) TranslateScrapeContent(ctx context.Context, title, description string) (TranslationResult, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" && description == "" {
		return TranslationResult{}, nil
	}
	if t == nil || strings.TrimSpace(t.chatURL) == "" {
		return TranslationResult{}, fmt.Errorf("translation api url is empty")
	}

	payload, err := json.Marshal(scrapeTranslationPayload{
		Title:       title,
		Description: description,
	})
	if err != nil {
		return TranslationResult{}, fmt.Errorf("marshal translation payload: %w", err)
	}
	body, err := json.Marshal(openAIChatCompletionRequest{
		Model: t.model,
		Messages: []openAIChatMessage{
			{
				Role:    "system",
				Content: "你是中文翻译引擎。将输入中的非中文内容翻译成简体中文；演员名、编号、代码、网址、文件名保持原样。只输出 JSON 对象，字段固定为 title_zh 和 description_zh，不要输出其它文字。",
			},
			{
				Role:    "user",
				Content: string(payload),
			},
		},
		Temperature: 0,
	})
	if err != nil {
		return TranslationResult{}, fmt.Errorf("marshal translation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.chatURL, bytes.NewReader(body))
	if err != nil {
		return TranslationResult{}, fmt.Errorf("create translation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return TranslationResult{}, fmt.Errorf("translation request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return TranslationResult{}, fmt.Errorf("translation status=%d body=%s", resp.StatusCode, string(raw))
	}

	var out openAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return TranslationResult{}, fmt.Errorf("decode translation response: %w", err)
	}
	if len(out.Choices) == 0 {
		return TranslationResult{}, fmt.Errorf("translation response has no choices")
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	if content == "" {
		return TranslationResult{}, fmt.Errorf("translation response content is empty")
	}

	var translated scrapeTranslationResponse
	if err := json.Unmarshal([]byte(extractJSONObject(content)), &translated); err != nil {
		return TranslationResult{}, fmt.Errorf("decode translation content: %w", err)
	}
	if strings.TrimSpace(translated.TitleZH) == "" {
		translated.TitleZH = title
	}
	if strings.TrimSpace(translated.DescriptionZH) == "" {
		translated.DescriptionZH = description
	}
	return TranslationResult{
		Title:       strings.TrimSpace(translated.TitleZH),
		Description: strings.TrimSpace(translated.DescriptionZH),
	}, nil
}

func normalizeOpenAIChatURL(raw string) string {
	base := strings.TrimRight(strings.TrimSpace(raw), "/")
	if base == "" {
		return ""
	}
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

func extractJSONObject(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(strings.TrimSpace(content), "```")
		content = strings.TrimSpace(content)
	}
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end >= start {
		return content[start : end+1]
	}
	return content
}

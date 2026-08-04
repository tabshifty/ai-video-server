package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"video-server/internal/models"
)

var explicitTIDKey = regexp.MustCompile(`^tid:([0-9]+)$`)

type importOptions struct {
	BaseURL   string
	Token     string
	Source    string
	BoardKey  string
	BatchSize int
}

type importReport struct {
	InputTIDs int `json:"input_tids"`
	Ignored   int `json:"ignored_entries"`
	Submitted int `json:"submitted"`
	Created   int `json:"created"`
	Pending   int `json:"pending"`
	Duplicate int `json:"duplicate"`
}

func main() {
	loadImportEnvironment()
	home, _ := os.UserHomeDir()
	defaultStatePath := filepath.Join(home, ".hermes", "state", "sehuatang_forum95_seen.json")
	var statePath, baseURL, token, source, boardKey string
	var batchSize int
	flag.StringVar(&statePath, "state", defaultStatePath, "Hermes seen.json 路径")
	flag.StringVar(&baseURL, "base-url", "http://127.0.0.1:8080", "项目 API 根地址")
	flag.StringVar(&token, "token", os.Getenv("HERMES_API_TOKEN"), "Hermes API Token，默认读取 HERMES_API_TOKEN")
	flag.StringVar(&source, "source", "sehuatang", "论坛来源代码")
	flag.StringVar(&boardKey, "board-key", "95", "论坛版块代码")
	flag.IntVar(&batchSize, "batch-size", 100, "每批导入数量，最大 100")
	flag.Parse()

	if strings.TrimSpace(token) == "" || batchSize <= 0 || batchSize > 100 {
		slog.Error("token 不能为空，batch-size 必须为 1 到 100")
		os.Exit(2)
	}
	file, err := os.Open(statePath)
	if err != nil {
		slog.Error("打开 Hermes 状态文件失败", "error", err)
		os.Exit(1)
	}
	tids, ignored, err := parseSeenState(file)
	_ = file.Close()
	if err != nil {
		slog.Error("解析 Hermes 状态文件失败", "error", err)
		os.Exit(1)
	}
	if len(tids) == 0 {
		slog.Error("状态文件中没有可导入的 tid:<数字> 键")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	report, err := runImport(ctx, &http.Client{Timeout: 30 * time.Second}, importOptions{
		BaseURL: baseURL, Token: token, Source: source, BoardKey: boardKey, BatchSize: batchSize,
	}, tids)
	report.InputTIDs = len(tids)
	report.Ignored = ignored
	if encodeErr := json.NewEncoder(os.Stdout).Encode(report); encodeErr != nil {
		slog.Error("输出导入报告失败", "error", encodeErr)
		os.Exit(1)
	}
	if err != nil {
		slog.Error("导入失败", "error", err)
		os.Exit(1)
	}
}

func loadImportEnvironment() {
	if envFile := strings.TrimSpace(os.Getenv("ENV_FILE")); envFile != "" {
		_ = godotenv.Overload(envFile)
		return
	}
	_ = godotenv.Load()
}

func parseSeenState(reader io.Reader) ([]string, int, error) {
	var state struct {
		SeenThreads []string `json:"seen_threads"`
		Seen        []string `json:"seen"`
	}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&state); err != nil {
		return nil, 0, err
	}
	entries := state.SeenThreads
	if entries == nil {
		entries = state.Seen
	}
	tids := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	ignored := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry, "tid:") && !explicitTIDKey.MatchString(entry) {
			return nil, ignored, fmt.Errorf("非法 tid 键 %q", entry)
		}
		matches := explicitTIDKey.FindStringSubmatch(entry)
		if len(matches) != 2 {
			ignored++
			continue
		}
		if _, exists := seen[matches[1]]; exists {
			ignored++
			continue
		}
		seen[matches[1]] = struct{}{}
		tids = append(tids, matches[1])
	}
	return tids, ignored, nil
}

func runImport(ctx context.Context, client *http.Client, options importOptions, tids []string) (importReport, error) {
	var report importReport
	if client == nil || strings.TrimSpace(options.BaseURL) == "" || strings.TrimSpace(options.Token) == "" ||
		strings.TrimSpace(options.Source) == "" || strings.TrimSpace(options.BoardKey) == "" || options.BatchSize <= 0 || options.BatchSize > 100 {
		return report, fmt.Errorf("导入参数非法")
	}
	endpoint := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/") + "/api/v1/integrations/hermes/forum-posts/discover"
	for start := 0; start < len(tids); start += options.BatchSize {
		end := start + options.BatchSize
		if end > len(tids) {
			end = len(tids)
		}
		posts := make([]models.ForumPostCandidate, end-start)
		for i, tid := range tids[start:end] {
			posts[i] = models.ForumPostCandidate{TID: tid}
		}
		payload, err := json.Marshal(models.ForumPostDiscoverInput{
			Mode: models.ForumPostDiscoverModeDedupeOnly, Source: strings.TrimSpace(options.Source),
			BoardKey: strings.TrimSpace(options.BoardKey), Posts: posts,
		})
		if err != nil {
			return report, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return report, err
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(options.Token))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return report, fmt.Errorf("提交第 %d 批: %w", start/options.BatchSize+1, err)
		}
		var envelope struct {
			Code int                              `json:"code"`
			Msg  string                           `json:"msg"`
			Data []models.ForumPostDiscoverResult `json:"data"`
		}
		decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&envelope)
		_ = resp.Body.Close()
		if decodeErr != nil {
			return report, fmt.Errorf("解析第 %d 批响应: %w", start/options.BatchSize+1, decodeErr)
		}
		if resp.StatusCode != http.StatusOK || envelope.Code != 0 {
			return report, fmt.Errorf("第 %d 批失败: HTTP %d, code=%d, msg=%s", start/options.BatchSize+1, resp.StatusCode, envelope.Code, envelope.Msg)
		}
		if len(envelope.Data) != len(posts) {
			return report, fmt.Errorf("第 %d 批响应数量不匹配: got=%d want=%d", start/options.BatchSize+1, len(envelope.Data), len(posts))
		}
		report.Submitted += len(posts)
		for i, result := range envelope.Data {
			if result.TID != posts[i].TID {
				return report, fmt.Errorf("第 %d 批第 %d 项 tid 不匹配: got=%q want=%q", start/options.BatchSize+1, i+1, result.TID, posts[i].TID)
			}
			switch result.Disposition {
			case models.ForumPostDispositionCreated:
				report.Created++
			case models.ForumPostDispositionPending:
				report.Pending++
			case models.ForumPostDispositionDuplicate:
				report.Duplicate++
			default:
				return report, fmt.Errorf("第 %d 批返回未知 disposition %q", start/options.BatchSize+1, result.Disposition)
			}
		}
	}
	return report, nil
}

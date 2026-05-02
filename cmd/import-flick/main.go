package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"video-server/internal/database"
	"video-server/internal/repository"
	"video-server/internal/services"
)

type importConfig struct {
	MongoURI       string
	MongoDB        string
	Collection     string
	PostgresDSN    string
	SourceVideoDir string
	SourceCoverDir string
	StorageRoot    string
	Tag            string
	Since          time.Time
	Limit          int64
	BatchSize      int
	DryRun         bool
	ReportPath     string
}

type mongoExportVideo struct {
	ID        mongoObjectID `json:"_id"`
	MD5       string        `json:"md5"`
	Tags      []string      `json:"tags"`
	Desc      string        `json:"desc"`
	CanPlay   bool          `json:"canplay"`
	CreatedAt mongoTime     `json:"createdAt"`
	UpdatedAt mongoTime     `json:"updatedAt"`
}

type mongoObjectID string

func (m *mongoObjectID) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch value := raw.(type) {
	case string:
		*m = mongoObjectID(strings.TrimSpace(value))
		return nil
	case map[string]any:
		if oid, ok := value["$oid"].(string); ok {
			*m = mongoObjectID(strings.TrimSpace(oid))
			return nil
		}
	}
	return fmt.Errorf("unsupported object id payload: %s", string(data))
}

type mongoTime struct {
	time.Time
}

func (m *mongoTime) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	parsed, err := decodeMongoTime(raw)
	if err != nil {
		return err
	}
	m.Time = parsed
	return nil
}

func decodeMongoTime(raw any) (time.Time, error) {
	switch value := raw.(type) {
	case string:
		return parseSinceFlag(value)
	case map[string]any:
		dateValue, ok := value["$date"]
		if !ok {
			return time.Time{}, fmt.Errorf("missing $date value")
		}
		switch typed := dateValue.(type) {
		case string:
			return parseSinceFlag(typed)
		case map[string]any:
			if millis, ok := typed["$numberLong"].(string); ok {
				parsedMillis, err := parseEpochMillis(millis)
				if err != nil {
					return time.Time{}, err
				}
				return parsedMillis, nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("unsupported mongo time payload")
}

func parseEpochMillis(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty epoch millis")
	}
	var millis int64
	if _, err := fmt.Sscan(value, &millis); err != nil {
		return time.Time{}, fmt.Errorf("parse epoch millis: %w", err)
	}
	return time.UnixMilli(millis).UTC(), nil
}

type importReport struct {
	StartedAt      time.Time         `json:"started_at"`
	FinishedAt     time.Time         `json:"finished_at"`
	DryRun         bool              `json:"dry_run"`
	Processed      int               `json:"processed"`
	Imported       int               `json:"imported"`
	WouldImport    int               `json:"would_import"`
	Skipped        int               `json:"skipped"`
	Failed         int               `json:"failed"`
	SkipReasons    map[string]int    `json:"skip_reasons"`
	Failures       []importFailure   `json:"failures"`
	MongoURI       string            `json:"mongo_uri"`
	MongoDB        string            `json:"mongo_db"`
	Collection     string            `json:"collection"`
	PostgresDSN    string            `json:"postgres_dsn"`
	SourceVideoDir string            `json:"source_video_dir"`
	SourceCoverDir string            `json:"source_cover_dir"`
	StorageRoot    string            `json:"storage_root"`
	Filters        map[string]string `json:"filters"`
}

type importFailure struct {
	SourceID string `json:"source_id"`
	MD5      string `json:"md5"`
	Error    string `json:"error"`
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		slog.Error("parse flags failed", "error", err)
		os.Exit(1)
	}

	report := importReport{
		StartedAt:      time.Now().UTC(),
		DryRun:         cfg.DryRun,
		SkipReasons:    map[string]int{},
		MongoURI:       cfg.MongoURI,
		MongoDB:        cfg.MongoDB,
		Collection:     cfg.Collection,
		PostgresDSN:    redactDSN(cfg.PostgresDSN),
		SourceVideoDir: cfg.SourceVideoDir,
		SourceCoverDir: cfg.SourceCoverDir,
		StorageRoot:    cfg.StorageRoot,
		Filters: map[string]string{
			"tag":   strings.TrimSpace(cfg.Tag),
			"since": cfg.Since.Format(time.RFC3339),
		},
	}
	defer func() {
		report.FinishedAt = time.Now().UTC()
		if writeErr := writeReport(cfg.ReportPath, report); writeErr != nil {
			slog.Error("write report failed", "error", writeErr, "report_path", cfg.ReportPath)
		}
	}()

	ctx := context.Background()
	pool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("connect postgres failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewVideoRepository(pool)
	svc := services.NewFlickImportService(repo)

	err = streamMongoExport(ctx, cfg, func(doc mongoExportVideo) error {
		report.Processed++
		outcome, err := svc.ImportPlayableVideo(ctx, services.FlickSourceVideo{
			SourceID:    string(doc.ID),
			MD5:         strings.TrimSpace(doc.MD5),
			Tags:        doc.Tags,
			Description: doc.Desc,
			CanPlay:     doc.CanPlay,
			CreatedAt:   doc.CreatedAt.Time,
			UpdatedAt:   doc.UpdatedAt.Time,
		}, services.FlickImportOptions{
			SourceVideoDir: cfg.SourceVideoDir,
			SourceCoverDir: cfg.SourceCoverDir,
			StorageRoot:    cfg.StorageRoot,
			DryRun:         cfg.DryRun,
		})
		if err != nil {
			report.Failed++
			report.Failures = append(report.Failures, importFailure{
				SourceID: string(doc.ID),
				MD5:      doc.MD5,
				Error:    err.Error(),
			})
			return nil
		}
		switch outcome.Status {
		case services.FlickImportStatusImported:
			report.Imported++
		case services.FlickImportStatusDryRun:
			report.WouldImport++
		case services.FlickImportStatusSkipped:
			report.Skipped++
			report.SkipReasons[outcome.Reason]++
		}
		return nil
	})
	if err != nil {
		slog.Error("stream mongoexport failed", "error", err)
		os.Exit(1)
	}

	slog.Info("flick import finished",
		"processed", report.Processed,
		"imported", report.Imported,
		"would_import", report.WouldImport,
		"skipped", report.Skipped,
		"failed", report.Failed,
		"report_path", cfg.ReportPath,
	)
}

func parseFlags() (importConfig, error) {
	var sinceRaw string
	cfg := importConfig{}
	flag.StringVar(&cfg.MongoURI, "mongo-dsn", firstNonEmpty(os.Getenv("FLICK_MONGO_DSN"), "mongodb://localhost:27017/flick"), "Mongo connection string")
	flag.StringVar(&cfg.MongoDB, "mongo-db", firstNonEmpty(os.Getenv("FLICK_MONGO_DB"), "flick"), "Mongo database name")
	flag.StringVar(&cfg.Collection, "collection", firstNonEmpty(os.Getenv("FLICK_MONGO_COLLECTION"), "videos"), "Mongo collection name")
	flag.StringVar(&cfg.PostgresDSN, "postgres-dsn", strings.TrimSpace(os.Getenv("POSTGRES_DSN")), "Postgres DSN")
	flag.StringVar(&cfg.SourceVideoDir, "source-video-dir", firstNonEmpty(os.Getenv("FLICK_SOURCE_VIDEO_DIR"), "/Volumes/large/dest/video"), "Source playable video directory")
	flag.StringVar(&cfg.SourceCoverDir, "source-cover-dir", firstNonEmpty(os.Getenv("FLICK_SOURCE_COVER_DIR"), "/Volumes/large/dest/cover"), "Source cover directory")
	flag.StringVar(&cfg.StorageRoot, "storage-root", firstNonEmpty(os.Getenv("STORAGE_ROOT"), "./storage"), "Target storage root")
	flag.StringVar(&cfg.Tag, "tag", "", "Optional tag filter")
	flag.StringVar(&sinceRaw, "since", "", "Optional createdAt lower bound, supports YYYY-MM-DD or RFC3339")
	flag.Int64Var(&cfg.Limit, "limit", 0, "Optional max documents to scan")
	flag.IntVar(&cfg.BatchSize, "batch-size", 100, "mongoexport batch size")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Only evaluate candidates, do not copy or insert")
	flag.StringVar(&cfg.ReportPath, "report-path", defaultReportPath(), "JSON report path")
	flag.Parse()

	if strings.TrimSpace(cfg.PostgresDSN) == "" {
		return importConfig{}, fmt.Errorf("postgres-dsn is required")
	}
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return importConfig{}, fmt.Errorf("mongo-dsn is required")
	}
	if sinceRaw != "" {
		since, err := parseSinceFlag(sinceRaw)
		if err != nil {
			return importConfig{}, err
		}
		cfg.Since = since
	}
	return cfg, nil
}

func buildMongoFilter(tag string, since time.Time) map[string]any {
	filter := map[string]any{
		"canplay": true,
	}
	normalizedTag := strings.TrimSpace(tag)
	if normalizedTag != "" {
		filter["tags"] = map[string]any{
			"$regex":   "^" + regexp.QuoteMeta(normalizedTag) + "$",
			"$options": "i",
		}
	}
	if !since.IsZero() {
		filter["createdAt"] = map[string]any{
			"$gte": since.UTC(),
		}
	}
	return filter
}

func parseSinceFlag(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid since value: %s", raw)
}

func streamMongoExport(ctx context.Context, cfg importConfig, handle func(mongoExportVideo) error) error {
	queryJSON, err := json.Marshal(toMongoExtendedJSON(buildMongoFilter(cfg.Tag, cfg.Since)))
	if err != nil {
		return fmt.Errorf("marshal mongo filter: %w", err)
	}
	args := []string{
		"--uri=" + cfg.MongoURI,
		"--db=" + cfg.MongoDB,
		"--collection=" + cfg.Collection,
		"--type=json",
		"--jsonFormat=canonical",
		"--sort={\"createdAt\":1}",
		"--fields=_id,md5,tags,desc,canplay,createdAt,updatedAt",
		"--query=" + string(queryJSON),
		"--batchSize=" + fmt.Sprintf("%d", cfg.BatchSize),
	}
	if cfg.Limit > 0 {
		args = append(args, "--limit="+fmt.Sprintf("%d", cfg.Limit))
	}
	cmd := exec.CommandContext(ctx, "mongoexport", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open mongoexport stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mongoexport: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 256*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var doc mongoExportVideo
		if err := json.Unmarshal([]byte(line), &doc); err != nil {
			return fmt.Errorf("decode mongoexport line: %w", err)
		}
		if err := handle(doc); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan mongoexport output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("mongoexport failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func toMongoExtendedJSON(value any) any {
	switch typed := value.(type) {
	case time.Time:
		return map[string]any{"$date": typed.UTC().Format(time.RFC3339Nano)}
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, entry := range typed {
			out[key] = toMongoExtendedJSON(entry)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, entry := range typed {
			out = append(out, toMongoExtendedJSON(entry))
		}
		return out
	default:
		return value
	}
}

func writeReport(path string, report importReport) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir report dir: %w", err)
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write report file: %w", err)
	}
	return nil
}

func defaultReportPath() string {
	name := "flick-import-" + time.Now().UTC().Format("20060102-150405") + ".json"
	return filepath.Join(".run", "reports", name)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func redactDSN(dsn string) string {
	if strings.TrimSpace(dsn) == "" {
		return ""
	}
	re := regexp.MustCompile(`:[^:@/]+@`)
	return re.ReplaceAllString(dsn, ":***@")
}

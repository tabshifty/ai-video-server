package queue

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"video-server/internal/models"
)

func TestCloudIngestTaskIDIncludesAttempt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		importID int64
		attempt  int
		want     string
	}{
		{importID: 9, attempt: 0, want: "cloud:ingest:9:0"},
		{importID: 9, attempt: 3, want: "cloud:ingest:9:3"},
	}
	for _, tc := range cases {
		if got := CloudIngestTaskID(tc.importID, tc.attempt); got != tc.want {
			t.Fatalf("CloudIngestTaskID(%d, %d) = %q, want %q", tc.importID, tc.attempt, got, tc.want)
		}
	}
}

func TestBuildCloudIngestTaskOptionsUsesGivenQueue(t *testing.T) {
	t.Parallel()

	opts := buildCloudIngestTaskOptions("cloud-ingest", "cloud:ingest:9:1")

	assertOption(t, opts, asynq.QueueOpt, "cloud-ingest")
	assertOption(t, opts, asynq.TaskIDOpt, "cloud:ingest:9:1")
	assertOption(t, opts, asynq.MaxRetryOpt, 2)
	assertOption(t, opts, asynq.TimeoutOpt, 2*time.Hour)
}

type fakeCloudIngestProcessor struct {
	processID      int64
	processAttempt int
	processInput   models.CloudDownloadImportInput
	processErr     error
	markFailedID   int64
	markAttempt    int
	markMessage    string
	markCalls      int
}

func (f *fakeCloudIngestProcessor) Process(_ context.Context, importID int64, attempt int, input models.CloudDownloadImportInput) error {
	f.processID = importID
	f.processAttempt = attempt
	f.processInput = input
	return f.processErr
}

func (f *fakeCloudIngestProcessor) MarkFailed(_ context.Context, importID int64, attempt int, message string) error {
	f.markFailedID = importID
	f.markAttempt = attempt
	f.markMessage = message
	f.markCalls++
	return nil
}

func TestHandleCloudIngestDelegatesPayloadInput(t *testing.T) {
	t.Parallel()

	fake := &fakeCloudIngestProcessor{}
	processor := &Processor{cloudImport: fake, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	task := asynq.NewTask(TypeCloudIngest, mustMarshalCloudIngestPayload(t, CloudIngestPayload{
		ImportID: 42,
		Attempt:  1,
		Input: models.CloudDownloadImportInput{
			Kind:     models.CloudImportKindArchive,
			Title:    "合集",
			Password: "secret",
		},
	}))

	if err := processor.HandleCloudIngest(context.Background(), task); err != nil {
		t.Fatalf("HandleCloudIngest() error = %v", err)
	}
	if fake.processID != 42 {
		t.Fatalf("Process import id = %d, want 42", fake.processID)
	}
	if fake.processAttempt != 1 {
		t.Fatalf("Process attempt = %d, want 1", fake.processAttempt)
	}
	if fake.processInput.Kind != models.CloudImportKindArchive || fake.processInput.Password != "secret" {
		t.Fatalf("Process input = %+v", fake.processInput)
	}
}

func TestHandleCloudIngestRequiresService(t *testing.T) {
	t.Parallel()

	processor := &Processor{}
	task := asynq.NewTask(TypeCloudIngest, mustMarshalCloudIngestPayload(t, CloudIngestPayload{ImportID: 1}))
	if err := processor.HandleCloudIngest(context.Background(), task); err == nil {
		t.Fatal("expected error when import service is not configured")
	}
}

func TestHandleCloudIngestRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	processor := &Processor{cloudImport: &fakeCloudIngestProcessor{}}
	task := asynq.NewTask(TypeCloudIngest, []byte("{"))
	if err := processor.HandleCloudIngest(context.Background(), task); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func mustMarshalCloudIngestPayload(t *testing.T, payload CloudIngestPayload) []byte {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}

// 回归护栏：导入曾经错误地复用下载队列，TaskID 与卡住的旧任务冲突后被
// 静默吞掉，自动入库永远不执行。这里直接检查调用点，避免再次写错字段。
func TestEnqueueCloudIngestUsesIngestQueue(t *testing.T) {
	source, err := os.ReadFile("cloud_ingest.go")
	if err != nil {
		t.Fatalf("read cloud_ingest.go: %v", err)
	}
	if strings.Contains(string(source), "buildCloudIngestTaskOptions(e.cloudQueue") {
		t.Fatal("EnqueueCloudIngest must not enqueue on the download queue")
	}
	if !strings.Contains(string(source), "buildCloudIngestTaskOptions(e.cloudIngestQueue") {
		t.Fatal("EnqueueCloudIngest must enqueue on the ingest queue")
	}
}

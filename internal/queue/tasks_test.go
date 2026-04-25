package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestBuildTranscodeTaskOptionsIncludesExplicitTimeout(t *testing.T) {
	t.Parallel()

	opts := buildTranscodeTaskOptions("transcode", 6*time.Hour)

	assertOption(t, opts, asynq.QueueOpt, "transcode")
	assertOption(t, opts, asynq.ProcessInOpt, 2*time.Second)
	assertOption(t, opts, asynq.TimeoutOpt, 6*time.Hour)
	assertOption(t, opts, asynq.MaxRetryOpt, 3)
}

func assertOption(t *testing.T, opts []asynq.Option, typ asynq.OptionType, want any) {
	t.Helper()
	for _, opt := range opts {
		if opt.Type() != typ {
			continue
		}
		if got := opt.Value(); got != want {
			t.Fatalf("option %v value=%v want=%v", typ, got, want)
		}
		return
	}
	t.Fatalf("expected option type %v in %v", typ, opts)
}

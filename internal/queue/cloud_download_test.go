package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestBuildCloudDownloadTaskOptionsUsesDedicatedQueue(t *testing.T) {
	opts := buildCloudDownloadTaskOptions("cloud", "cloud:download:7:1")

	assertOption(t, opts, asynq.QueueOpt, "cloud")
	assertOption(t, opts, asynq.TaskIDOpt, "cloud:download:7:1")
	assertOption(t, opts, asynq.MaxRetryOpt, 2)
	assertOption(t, opts, asynq.TimeoutOpt, 6*time.Hour)
}

func TestCloudDownloadTaskIDIncludesRetryCount(t *testing.T) {
	cases := []struct {
		taskID int64
		retry  int
		want   string
	}{
		{taskID: 7, retry: 0, want: "cloud:download:7:0"},
		{taskID: 7, retry: 2, want: "cloud:download:7:2"},
	}
	for _, tc := range cases {
		if got := CloudDownloadTaskID(tc.taskID, tc.retry); got != tc.want {
			t.Fatalf("CloudDownloadTaskID(%d, %d) = %q, want %q", tc.taskID, tc.retry, got, tc.want)
		}
	}
}

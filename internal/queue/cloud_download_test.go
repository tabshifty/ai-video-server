package queue

import "testing"

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

package repository

import (
	"strings"
	"testing"
)

func TestBuildAdminListTranscodingTasksSQLWithoutStatus(t *testing.T) {
	t.Parallel()

	countSQL, countArgs, listSQL, listArgs := buildAdminListTranscodingTasksSQL("", 1, 20)

	for _, want := range []string{"from transcoding_jobs where 1=1", "select count(*)"} {
		if !strings.Contains(strings.ToLower(countSQL), want) {
			t.Fatalf("count sql missing %q: %s", want, countSQL)
		}
	}
	if len(countArgs) != 0 {
		t.Fatalf("expected no count args, got %#v", countArgs)
	}
	if !strings.Contains(strings.ToLower(listSQL), "limit $1 offset $2") {
		t.Fatalf("list sql should use limit/offset placeholders 1/2: %s", listSQL)
	}
	if len(listArgs) != 2 || listArgs[0] != 20 || listArgs[1] != 0 {
		t.Fatalf("unexpected list args: %#v", listArgs)
	}
}

func TestBuildAdminListTranscodingTasksSQLWithStatus(t *testing.T) {
	t.Parallel()

	countSQL, countArgs, listSQL, listArgs := buildAdminListTranscodingTasksSQL("FAILED", 3, 15)

	if !strings.Contains(strings.ToLower(countSQL), "status = $1") {
		t.Fatalf("count sql should filter by status: %s", countSQL)
	}
	if len(countArgs) != 1 || countArgs[0] != "failed" {
		t.Fatalf("unexpected count args: %#v", countArgs)
	}
	if !strings.Contains(strings.ToLower(listSQL), "where 1=1 and status = $1") {
		t.Fatalf("list sql should keep status placeholder stable: %s", listSQL)
	}
	if !strings.Contains(strings.ToLower(listSQL), "limit $2 offset $3") {
		t.Fatalf("list sql should shift limit/offset placeholders: %s", listSQL)
	}
	if len(listArgs) != 3 || listArgs[0] != "failed" || listArgs[1] != 15 || listArgs[2] != 30 {
		t.Fatalf("unexpected list args: %#v", listArgs)
	}
}

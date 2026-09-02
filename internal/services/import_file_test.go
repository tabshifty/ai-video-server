package services

import "testing"

func TestSaveImportedFileDoesNotRefreshExistingOriginalPath(t *testing.T) {
	t.Parallel()

	if shouldUpdateExistingVideoOriginalPath(false, "uploaded") {
		t.Fatal("imported duplicate must not refresh the existing original path")
	}
}

package hashutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	got, err := SHA256(path)
	if err != nil {
		t.Fatalf("compute sha256: %v", err)
	}

	const want = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if got != want {
		t.Fatalf("unexpected hash: got=%s want=%s", got, want)
	}
}

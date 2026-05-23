package oshash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeReturnsErrFileTooSmall(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "small.bin")
	if err := os.WriteFile(path, make([]byte, ChunkSize*2-1), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if _, err := Compute(path); err != ErrFileTooSmall {
		t.Fatalf("expected ErrFileTooSmall, got %v", err)
	}
}

func TestComputeReturnsDeterministicHex(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.bin")
	if err := os.WriteFile(path, bytesOf('B', ChunkSize*4), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	first, err := Compute(path)
	if err != nil {
		t.Fatalf("Compute() error = %v", err)
	}
	second, err := Compute(path)
	if err != nil {
		t.Fatalf("Compute() second error = %v", err)
	}
	if first != second {
		t.Fatalf("expected stable hash, got %q and %q", first, second)
	}
	if len(first) != 16 {
		t.Fatalf("expected 16 hex chars, got %q", first)
	}
}

func TestComputeMatchesPythonOshashGoldenFixture(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "fixture_256k_0x42.bin")
	if err := os.WriteFile(path, bytesOf(0x42, 256*1024), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	got, err := Compute(path)
	if err != nil {
		t.Fatalf("Compute() error = %v", err)
	}
	const want = "9090909090948000"
	if got != want {
		t.Fatalf("Compute() = %q, want Python oshash golden value %q", got, want)
	}
}

func bytesOf(value byte, n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = value
	}
	return data
}

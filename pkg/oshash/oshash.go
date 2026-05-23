package oshash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const ChunkSize = 64 * 1024

var ErrFileTooSmall = errors.New("file too small for oshash (need >= 128 KiB)")

func Compute(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	size := info.Size()
	if size < 2*ChunkSize {
		return "", ErrFileTooSmall
	}

	hash := uint64(size)
	if err := accumulate(f, 0, &hash); err != nil {
		return "", err
	}
	if err := accumulate(f, size-ChunkSize, &hash); err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x", hash), nil
}

func accumulate(f *os.File, offset int64, hash *uint64) error {
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	buf := make([]byte, ChunkSize)
	if _, err := io.ReadFull(f, buf); err != nil {
		return err
	}
	for i := 0; i+8 <= len(buf); i += 8 {
		*hash += binary.LittleEndian.Uint64(buf[i : i+8])
	}
	return nil
}

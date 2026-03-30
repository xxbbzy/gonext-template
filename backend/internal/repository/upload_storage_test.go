package repository

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalFileStorageRepositoryRejectsExistingFilename(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")

	if err := storage.SaveFile(context.Background(), "existing.txt", strings.NewReader("first")); err != nil {
		t.Fatalf("first SaveFile() error = %v", err)
	}
	if err := storage.SaveFile(context.Background(), "existing.txt", strings.NewReader("second")); err == nil {
		t.Fatal("second SaveFile() error = nil, want error")
	}

	content, err := os.ReadFile(filepath.Join(tempDir, "existing.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "first" {
		t.Fatalf("content = %q, want %q", string(content), "first")
	}
}

func TestLocalFileStorageRepositoryRemovesPartialFileOnReaderError(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewLocalFileStorageRepository(tempDir, "http://localhost:8080")

	storedName := "partial.txt"
	err := storage.SaveFile(context.Background(), storedName, &failingReader{
		firstChunk: []byte("partial data"),
		err:        errors.New("injected read failure"),
	})
	if err == nil {
		t.Fatal("SaveFile() error = nil, want error")
	}

	_, statErr := os.Stat(filepath.Join(tempDir, storedName))
	if !os.IsNotExist(statErr) {
		t.Fatalf("os.Stat() error = %v, want not-exist error", statErr)
	}
}

type failingReader struct {
	firstChunk []byte
	err        error
	readOnce   bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if !r.readOnce {
		r.readOnce = true
		n := copy(p, r.firstChunk)
		return n, nil
	}
	return 0, r.err
}

var _ io.Reader = (*failingReader)(nil)

package storage

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorage_Save(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &LocalStorage{tmpDir}

	ctx := context.Background()
	filename := "test.txt"
	content := []byte("Hello, world!")

	err := storage.Save(ctx, filename, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	path := filepath.Join(tmpDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if !bytes.Equal(data, content) {
		t.Errorf("file content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestLocalStorage_InvalidDirectory(t *testing.T) {
	storage := &LocalStorage{"/nonexistent_folder"}

	ctx := context.Background()
	filename := "test.txt"
	content := []byte("Hello, world!")

	err := storage.Save(ctx, filename, bytes.NewReader(content))
	if err == nil {
		t.Fatal("expected error when saving to nonexistent folder, got nil")
	}
}

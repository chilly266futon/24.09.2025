package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage сохраняет файлы на локальную файловую систему
type LocalStorage struct {
	Dir string
}

func NewLocalStorage(dir string) *LocalStorage {
	return &LocalStorage{
		Dir: dir,
	}
}

func (l *LocalStorage) Save(ctx context.Context, filename string, r io.Reader) error {
	if err := os.MkdirAll(l.Dir, os.ModePerm); err != nil {
		return err
	}

	path := filepath.Join(l.Dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

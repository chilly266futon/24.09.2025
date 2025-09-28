package domain

import (
	"context"
	"io"
)

// Storage абстрактное хранилище файлов
type Storage interface {
	// Save сохраняет данные по указанному имени файла
	Save(ctx context.Context, filename string, r io.Reader) error
}

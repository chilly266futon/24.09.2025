package workerpool_test

import (
	"context"
	"file-downloader-service/internal/domain"
	"file-downloader-service/internal/infrastructure/workerpool"
	"file-downloader-service/internal/usecase"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockStorage реализует domain.Storage для теста
type mockStorage struct{}

func (s *mockStorage) Save(ctx context.Context, filename string, r io.Reader) error {
	_, err := io.ReadAll(r)
	return err
}

func TestWorkerPoolProcessTask_Success(t *testing.T) {
	// тестовый HTTP-сервер с мгновенным ответом
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mockdata"))
	}))
	defer testServer.Close()

	svc := usecase.NewService()
	storage := &mockStorage{}
	pool := workerpool.NewWorkerPool(svc, storage, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task1 := svc.Create(ctx, []string{testServer.URL})
	pool.AddTask(task1)

	// Закрываем канал, чтобы воркеры завершились после обработки
	close(pool.TasksCh)

	// Ждём завершения всех воркеров
	pool.Wait()

	got, _ := svc.Get(ctx, task1.ID)
	if got.Status != domain.StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", got.Status)
	}
}

func TestWorkerPoolProcessTask_Failure(t *testing.T) {
	svc := usecase.NewService()
	storage := &mockStorage{}
	pool := workerpool.NewWorkerPool(svc, storage, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task1 := svc.Create(ctx, []string{"http://badhost.local"})
	pool.AddTask(task1)

	// Закрываем канал после добавления
	close(pool.TasksCh)
	pool.Wait()

	got, _ := svc.Get(ctx, task1.ID)
	if got.Status != domain.StatusFailed {
		t.Errorf("expected StatusFailed, got %v", got.Status)
	}
}

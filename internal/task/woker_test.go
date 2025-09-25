package task

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockStorage struct{}

func (s *mockStorage) Save(ctx context.Context, filename string, r io.Reader) error {
	_, err := io.ReadAll(r)
	return err
}

func TestWorkerPoolProcessTask(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mockdata"))
	}))
	defer testServer.Close()

	svc := NewService()
	storage := &mockStorage{}
	pool := NewWorkerPool(svc, storage, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task1 := svc.Create(ctx, []string{testServer.URL})
	pool.AddTask(task1)

	time.Sleep(100 * time.Millisecond)

	got, _ := svc.Get(ctx, task1.ID)
	if got.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %v", got.Status)
	}

	pool.Wait()
}

package usecase

import (
	"context"
	"file-downloader-service/internal/domain"
	"os"
	"testing"
)

func TestServiceCreateGetAll(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	urls := []string{"https://go.dev/dl/go1.25.1.darwin-arm64.pkg"}
	task := svc.Create(ctx, urls)

	if task.Status != domain.StatusPending {
		t.Errorf("expected StatusPending, got %v", task.Status)
	}

	got, ok := svc.Get(ctx, task.ID)
	if !ok || got.ID != task.ID {
		t.Errorf("failed to get task by ID")
	}

	list := svc.GetAll(ctx)
	if len(list) != len(urls) {
		t.Errorf("expected %v task in list, got %v", len(urls), len(list))
	}
}

func TestServiceSaveLoadTasks(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	task := svc.Create(ctx, []string{"url1"})

	filename := "test_tasks.json"
	defer os.Remove(filename)

	if err := svc.SaveTasks(filename); err != nil {
		t.Errorf("failed to save tasks: %v", err)
	}

	svc2 := NewService()
	if err := svc2.LoadTasks(filename); err != nil {
		t.Errorf("failed to load tasks: %v", err)
	}

	got, ok := svc2.Get(ctx, task.ID)
	if !ok || got.ID != task.ID {
		t.Errorf("loaded task not found or incorrect, got: %+v", got)
	}

	if got.Status != task.Status {
		t.Errorf("expected status %v, got: %v", task.Status, got.Status)
	}
}

func TestServiceLoadTasks_WithRunningReset(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	task := svc.Create(ctx, []string{"https://go.dev/dl/go1.25.1.darwin-arm64.pkg"})
	task.Status = domain.StatusRunning

	filename := "test_tasks_running.json"
	defer os.Remove(filename)

	if err := svc.SaveTasks(filename); err != nil {
		t.Errorf("failed to save tasks: %v", err)
	}

	svc2 := NewService()
	if err := svc2.LoadTasks(filename); err != nil {
		t.Errorf("failed to load tasks: %v", err)
	}

	got, ok := svc2.Get(ctx, task.ID)
	if !ok {
		t.Fatalf("expected task %s to be loaded", task.ID)
	}

	if got.Status != domain.StatusPending {
		t.Errorf("expected StatusPending after reload, got: %v", got.Status)
	}
}

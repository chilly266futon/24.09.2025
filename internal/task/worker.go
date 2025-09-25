package task

import (
	"context"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// WorkerPool управляет пулом воркеров для обработки задач
type WorkerPool struct {
	svc       *Service
	storage   Storage
	workerNum int
	wg        sync.WaitGroup
	tasksCh   chan *Task
}

func NewWorkerPool(svc *Service, storage Storage, workerNum int) *WorkerPool {
	return &WorkerPool{
		svc:       svc,
		storage:   storage,
		workerNum: workerNum,
		tasksCh:   make(chan *Task, 100),
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workerNum; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

// AddTask добавляет задачу в очередь в обработку
func (p *WorkerPool) AddTask(t *Task) {
	p.tasksCh <- t
}

// Wait ждет завершения всех воркеров
func (p *WorkerPool) Wait() {
	close(p.tasksCh)
	p.wg.Wait()
}

// worker один воркер, который берет задачи из канала и обрабатывает их
func (p *WorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()
	for t := range p.tasksCh {
		select {
		case <-ctx.Done():
			return
		default:
			p.processTask(ctx, t)
		}
	}
}

// processTask скачивает все файлы задачи
func (p *WorkerPool) processTask(ctx context.Context, t *Task) {
	t.Status = StatusRunning
	t.UpdatedAt = time.Now()

	for _, url := range t.URLs {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			t.Status = StatusFailed
			t.Error = err.Error()
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Status = StatusFailed
			t.Error = err.Error()
			return
		}

		err = p.storage.Save(ctx, filepath.Base(url), resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Status = StatusFailed
			t.Error = err.Error()
			return
		}
	}

	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

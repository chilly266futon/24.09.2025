package task

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// таймаут на скачивание одного файла
const fileDownloadTimeout = 3 * time.Minute

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
	if t.Done == nil {
		t.Done = make(chan struct{})
	}
	p.tasksCh <- t
}

// Wait ждет завершения всех воркеров
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// worker один воркер, который берет задачи из канала и обрабатывает их
func (p *WorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-p.tasksCh:
			if t == nil {
				continue
			}
			p.processTask(ctx, t)
		}
	}
}

// processTask скачивает все файлы задачи
func (p *WorkerPool) processTask(ctx context.Context, t *Task) {
	defer close(t.Done) // сигнал о завершении задачи
	t.Status = StatusRunning
	t.UpdatedAt = time.Now()

	for _, url := range t.URLs {
		fileCtx, cancel := context.WithTimeout(ctx, fileDownloadTimeout)
		err := p.downloadWithStorage(fileCtx, url)
		cancel()

		if err != nil {
			t.Status = StatusFailed
			t.Error = err.Error()
			log.Printf("failed to download url %s: %v", url, err)
			return
		}
	}

	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

// downloadWithStorage скачивает и сохраняет один файл
func (p *WorkerPool) downloadWithStorage(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return p.storage.Save(ctx, filepath.Base(url), resp.Body)
}

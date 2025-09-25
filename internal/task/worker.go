package task

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WorkerPool управляет пулом воркеров для обработки задач
type WorkerPool struct {
	svc       *Service
	workerNum int
	wg        sync.WaitGroup
	tasksCh   chan *Task
}

func NewWorkerPool(svc *Service, workerNum int) *WorkerPool {
	return &WorkerPool{
		svc:       svc,
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
		if err := downloadFile(ctx, url, "downloads"); err != nil {
			t.Status = StatusFailed
			t.Error = err
			log.Printf("failed to download url %s: %v", url, err)
			return
		}
	}

	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

// downloadFile скачивает файл по URL в указанную папку
func downloadFile(ctx context.Context, url string, dir string) interface{} {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	fileName := filepath.Join(dir, filepath.Base(url))
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

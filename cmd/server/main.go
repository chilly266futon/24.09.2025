package main

import (
	"context"
	"errors"
	"file-downloader-service/internal/infrastructure/config"
	"log"
	"net/http"
	"os"
	"os/signal"

	"file-downloader-service/internal/domain"
	"file-downloader-service/internal/infrastructure/handlers"
	"file-downloader-service/internal/infrastructure/storage"
	"file-downloader-service/internal/infrastructure/workerpool"
	"file-downloader-service/internal/usecase"
)

const tasksFile = "tasks.json"

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := usecase.NewService()
	// загружаем задачи при старте
	if err := svc.LoadTasks(cfg.Tasks.File); err != nil {
		log.Fatalf("failed to load tasks: %v", err)
	}

	localStorage := storage.NewLocalStorage(cfg.Storage.DownloadsDir)
	pool := workerpool.NewWorkerPool(svc, localStorage, cfg.WorkerPool.NumWorkers)
	pool.Start(ctx)

	// добавляем ранее незавершенные задачи в пул
	for _, t := range svc.GetAll(ctx) {
		if t.Status == domain.StatusPending || t.Status == domain.StatusRunning {
			pool.AddTask(t)
		}
	}

	r := handlers.NewHandlersWithPool(svc, pool)
	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		log.Println("shutting down...")
		cancel()
		server.Shutdown(context.Background())

		if err := svc.SaveTasks(tasksFile); err != nil {
			log.Fatalf("failed to save tasks: %v", err)
		}
	}()

	log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// ждем завершения всех воркеров
	pool.Wait()
}

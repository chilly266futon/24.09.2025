package main

import (
	"context"
	"errors"
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
	// контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := usecase.NewService()
	// загружаем задачи при старте
	if err := svc.LoadTasks(tasksFile); err != nil {
		log.Fatalf("failed to load tasks: %v", err)
	}

	localStorage := storage.NewLocalStorage("downloads")
	pool := workerpool.NewWorkerPool(svc, localStorage, 4)
	pool.Start(ctx)

	// добавляем ранее незавершенные задачи в пул
	for _, t := range svc.GetAll(ctx) {
		if t.Status == domain.StatusPending || t.Status == domain.StatusRunning {
			pool.AddTask(t)
		}
	}

	r := handlers.NewHandlersWithPool(svc, pool)
	server := &http.Server{
		Addr:    ":8080",
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

	log.Println("Starting server on port :8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// ждем завершения всех воркеров
	pool.Wait()
}

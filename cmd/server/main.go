package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"file-downloader-service/internal/task"
)

func main() {
	// контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := task.NewService()
	storage := task.NewLocalStorage("downloads")
	pool := task.NewWorkerPool(svc, storage, 4)
	pool.Start(ctx)

	r := task.NewHandlersWithPool(svc, pool)
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		log.Println("shutting down")
		cancel()
		server.Shutdown(context.Background())
	}()

	log.Println("Starting server on port :8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// ждем завершения всех воркеров
	pool.Wait()
}

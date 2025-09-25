package main

import (
	"log"
	"net/http"

	"file-downloader-service/internal/task"
)

func main() {
	svc := task.NewService()

	server := &http.Server{
		Addr:    ":8080",
		Handler: task.NewHandlers(svc),
	}

	log.Println("Starting server on port :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

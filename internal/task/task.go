package task

import (
	"time"
)

// TaskStatus определяет состояние задачи
type TaskStatus uint

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

// Task описывает задачу по скачиванию файлов
type Task struct {
	ID        string     `json:"id"`
	URLs      []string   `json:"urls"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Error     string     `json:"error,omitempty"`
}

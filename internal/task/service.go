package task

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service управляет задачами в памяти
type Service struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewService() *Service {
	return &Service{
		tasks: make(map[string]*Task),
	}
}

// Create создает новую задачу
func (s *Service) Create(ctx context.Context, urls []string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := &Task{
		ID:        uuid.NewString(),
		URLs:      urls,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.tasks[t.ID] = t
	return t
}

// Get возвращает задачу по ID
func (s *Service) Get(ctx context.Context, id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// GetAll возвращает все задачи
func (s *Service) GetAll(ctx context.Context) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

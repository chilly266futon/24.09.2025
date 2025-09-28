package usecase

import (
	"context"
	"encoding/json"
	"file-downloader-service/internal/domain"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service управляет задачами в памяти
type Service struct {
	mu    sync.RWMutex
	tasks map[string]*domain.Task
}

func NewService() *Service {
	return &Service{
		tasks: make(map[string]*domain.Task),
	}
}

// Create создает новую задачу
func (s *Service) Create(ctx context.Context, urls []string) *domain.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := &domain.Task{
		ID:        uuid.NewString(),
		URLs:      urls,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Done:      make(chan struct{}),
	}
	s.tasks[t.ID] = t
	return t
}

// Get возвращает задачу по ID
func (s *Service) Get(ctx context.Context, id string) (*domain.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// GetAll возвращает все задачи
func (s *Service) GetAll(ctx context.Context) []*domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*domain.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

// SaveTasks сохраняет все задачи в файл
func (s *Service) SaveTasks(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]*domain.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		clone := *t
		if clone.Status == domain.StatusRunning {
			clone.Status = domain.StatusPending
			clone.Error = ""
		}
		tasks = append(tasks, &clone)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadTasks загружает задачи из файла
func (s *Service) LoadTasks(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		s.tasks = make(map[string]*domain.Task)
		return nil
	}

	var tasks []*domain.Task
	if err = json.Unmarshal(data, &tasks); err != nil {
		return err
	}

	s.tasks = make(map[string]*domain.Task)
	for _, t := range tasks {
		s.tasks[t.ID] = t
	}
	return nil
}

package task

import (
	"context"
	"encoding/json"
	"os"
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

// SaveTasks сохраняет все задачи в файл
func (s *Service) SaveTasks(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		clone := *t
		if clone.Status == StatusPending {
			clone.Status = StatusPending
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
		s.tasks = make(map[string]*Task)
		return nil
	}

	var tasks []*Task
	if err = json.Unmarshal(data, &tasks); err != nil {
		return err
	}

	s.tasks = make(map[string]*Task)
	for _, t := range tasks {
		s.tasks[t.ID] = t
	}
	return nil
}

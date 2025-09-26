package tasks

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

// TaskStorage определяет интерфейс для работы с задачами
type TaskStorage interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id int) (*Task, error)
	GetAll(ctx context.Context) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id int) error

	// ВОПРОС: Зачем в каждом методе передавать context.Context, по заданию нужно было, но не понял зачем тут
}

// InMemoryStorage реализация TaskStorage в памяти
type InMemoryStorage struct {
	tasks  map[int]*Task
	nextID int
	mu     sync.RWMutex
}

// NewInMemoryStorage создает новое хранилище в памяти
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

func (s *InMemoryStorage) Create(ctx context.Context, task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.ID = s.nextID
	s.tasks[s.nextID] = task
	s.nextID++
	return nil
}

func (s *InMemoryStorage) GetByID(ctx context.Context, id int) (*Task, error) {
	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *InMemoryStorage) GetAll(ctx context.Context) ([]*Task, error) {
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *InMemoryStorage) Update(ctx context.Context, task *Task) error {
	if _, exists := s.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *InMemoryStorage) Delete(ctx context.Context, id int) error {
	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}
	delete(s.tasks, id)
	return nil
}

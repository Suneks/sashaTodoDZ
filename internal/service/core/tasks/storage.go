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
}

// InMemoryStorage реализация TaskStorage в памяти
type InMemoryStorage struct {
	tasks  map[int]*Task
	nextID int
	mu     sync.RWMutex // Мьютекс для синхронизации доступа
}

// NewInMemoryStorage создает новое хранилище в памяти
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

// Create создает новую задачу
func (s *InMemoryStorage) Create(ctx context.Context, task *Task) error {
	// Блокируем на запись - только одна горутина может писать
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, не отменили ли контекст
	select {
	case <-ctx.Done():
		return ctx.Err() // Возвращаем ошибку отмены/таймаута
	default:
	}

	// Присваиваем ID и сохраняем задачу
	task.ID = s.nextID
	s.tasks[s.nextID] = task
	s.nextID++

	return nil
}

// GetByID получает задачу по ID
func (s *InMemoryStorage) GetByID(ctx context.Context, id int) (*Task, error) {
	// Блокируем на чтение - если много горутин могут читать одновременно
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем контекст
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Ищем задачу
	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// GetAll получает все задачи
func (s *InMemoryStorage) GetAll(ctx context.Context) ([]*Task, error) {
	// Блокируем на чтение
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Проверяем контекст
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Создаем копию списка задач
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		// Проверяем контекст на каждой итерации для долгих операций
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update обновляет задачу
func (s *InMemoryStorage) Update(ctx context.Context, task *Task) error {
	// Блокируем на запись
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем контекст
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Проверяем, существует ли задача
	if _, exists := s.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}

	// Обновляем задачу
	s.tasks[task.ID] = task

	return nil
}

// Delete удаляет задачу
func (s *InMemoryStorage) Delete(ctx context.Context, id int) error {
	// Блокируем на запись
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем контекст
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)

	return nil
}

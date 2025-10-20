package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"toDo/internal/service/core/tasks"
)

// запуск всех тестов go test -run "TaskHandler" -v

// MockStorage - мок для тестирования
type MockStorage struct {
	tasks  map[int]*tasks.Task
	nextID int
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		tasks:  make(map[int]*tasks.Task),
		nextID: 1,
	}
}

func (m *MockStorage) Create(ctx context.Context, task *tasks.Task) error {
	task.ID = m.nextID
	m.tasks[m.nextID] = task
	m.nextID++
	return nil
}

func (m *MockStorage) GetByID(ctx context.Context, id int) (*tasks.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, tasks.ErrTaskNotFound
	}
	return task, nil
}

func (m *MockStorage) GetAll(ctx context.Context) ([]*tasks.Task, error) {
	tasksList := make([]*tasks.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasksList = append(tasksList, task)
	}
	return tasksList, nil
}

func (m *MockStorage) Update(ctx context.Context, task *tasks.Task) error {
	if _, exists := m.tasks[task.ID]; !exists {
		return tasks.ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, id int) error {
	if _, exists := m.tasks[id]; !exists {
		return tasks.ErrTaskNotFound
	}
	delete(m.tasks, id)
	return nil
}

// TestTaskHandler_CreateTask_TableTests табличные тесты для CreateTask
func TestTaskHandler_CreateTask_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:           "успешное создание задачи",
			requestBody:    `{"title":"Новая задача","description":"Описание задачи"}`,
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "создание задачи без заголовка",
			requestBody:    `{"description":"Только описание"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "некорректный JSON",
			requestBody:    `{"title":`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "пустое тело запроса",
			requestBody:    ``,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := NewMockStorage()
			handler := NewTaskHandler(storage).(*TaskHandler)

			req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Действие
			handler.CreateTask(rr, req)

			// Проверка
			if rr.Code != tc.expectedStatus {
				t.Errorf("ожидался статус %d, но получен %d", tc.expectedStatus, rr.Code)
			}

			// Если ожидается ошибка, проверяем, что в ответе есть error
			if tc.expectedError {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON ошибки: %v", err)
				}
				if response.Error == "" {
					t.Error("ожидалось сообщение об ошибке, но его нет")
				}
			} else if !tc.expectedError {
				// Если ожидается успех, проверяем, что задача создалась
				var response tasks.Task
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON задачи: %v", err)
				}
				if response.Title == "" {
					t.Error("ожидалось, что задача будет создана")
				}
			}
		})
	}
}

// TestTaskHandler_GetTaskByID_TableTests табличные тесты для GetTaskByID
func TestTaskHandler_GetTaskByID_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *MockStorage
		taskID         string
		expectedStatus int
		expectedError  bool
	}

	testCases := []testCase{
		{
			name: "успешное получение существующей задачи",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Существующая задача", Description: "Описание"}
				// Вручную добавляем задачу с ID=1
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "1",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "попытка получить несуществующую задачу",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Существующая задача", Description: "Описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "некорректный ID задачи",
			setupStorage: func() *MockStorage {
				return NewMockStorage()
			},
			taskID:         "abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := tc.setupStorage()
			handler := NewTaskHandler(storage).(*TaskHandler)

			// Создаём запрос с маршрутизацией
			req := httptest.NewRequest("GET", "/tasks/"+tc.taskID, nil)
			rr := httptest.NewRecorder()

			// Используем gorilla/mux для парсинга URL
			router := mux.NewRouter()
			router.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
				handler.GetTaskByID(w, r)
			}).Methods("GET")

			// Выполняем запрос через роутер
			router.ServeHTTP(rr, req)

			// Проверка
			if rr.Code != tc.expectedStatus {
				t.Errorf("ожидался статус %d, но получен %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedError {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON ошибки: %v", err)
				}
				if response.Error == "" {
					t.Error("ожидалось сообщение об ошибке, но его нет")
				}
			} else if !tc.expectedError {
				var response tasks.Task
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON задачи: %v", err)
				}
				if response.ID != 1 {
					t.Errorf("ожидался ID 1, но получен %d", response.ID)
				}
			}
		})
	}
}

// TestTaskHandler_UpdateTask_TableTests табличные тесты для UpdateTask
func TestTaskHandler_UpdateTask_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *MockStorage
		taskID         string
		requestBody    string
		expectedStatus int
		expectedError  bool
	}

	testCases := []testCase{
		{
			name: "успешное обновление задачи",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Старый заголовок", Description: "Старое описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "1",
			requestBody:    `{"title":"Новый заголовок","description":"Новое описание"}`,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "попытка обновить несуществующую задачу",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Существующая задача", Description: "Описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "999",
			requestBody:    `{"title":"Новый заголовок","description":"Новое описание"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "некорректный ID задачи",
			setupStorage: func() *MockStorage {
				return NewMockStorage()
			},
			taskID:         "abc",
			requestBody:    `{"title":"Заголовок","description":"Описание"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "обновление без заголовка",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Старый заголовок", Description: "Старое описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "1",
			requestBody:    `{"description":"Только описание"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := tc.setupStorage()
			handler := NewTaskHandler(storage).(*TaskHandler)

			// Создаём запрос с маршрутизацией
			req := httptest.NewRequest("PUT", "/tasks/"+tc.taskID, bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Используем gorilla/mux для парсинга URL
			router := mux.NewRouter()
			router.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
				handler.UpdateTask(w, r)
			}).Methods("PUT")

			// Выполняем запрос через роутер
			router.ServeHTTP(rr, req)

			// Проверка
			if rr.Code != tc.expectedStatus {
				t.Errorf("ожидался статус %d, но получен %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedError {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON ошибки: %v", err)
				}
				if response.Error == "" {
					t.Error("ожидалось сообщение об ошибке, но его нет")
				}
			} else if !tc.expectedError {
				var response tasks.Task
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON задачи: %v", err)
				}
				if response.Title == "" {
					t.Error("ожидалось, что задача будет обновлена")
				}
			}
		})
	}
}

// TestTaskHandler_DeleteTask_TableTests табличные тесты для DeleteTask
func TestTaskHandler_DeleteTask_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *MockStorage
		taskID         string
		expectedStatus int
		expectedError  bool
	}

	testCases := []testCase{
		{
			name: "успешное удаление задачи",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Задача для удаления", Description: "Описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "1",
			expectedStatus: http.StatusNoContent,
			expectedError:  false,
		},
		{
			name: "попытка удалить несуществующую задачу",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task := &tasks.Task{ID: 1, Title: "Существующая задача", Description: "Описание"}
				storage.tasks[1] = task
				storage.nextID = 2
				return storage
			},
			taskID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name: "некорректный ID задачи",
			setupStorage: func() *MockStorage {
				return NewMockStorage()
			},
			taskID:         "abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := tc.setupStorage()
			handler := NewTaskHandler(storage).(*TaskHandler)

			// Создаём запрос с маршрутизацией
			req := httptest.NewRequest("DELETE", "/tasks/"+tc.taskID, nil)
			rr := httptest.NewRecorder()

			// Используем gorilla/mux для парсинга URL
			router := mux.NewRouter()
			router.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
				handler.DeleteTask(w, r)
			}).Methods("DELETE")

			// Выполняем запрос через роутер
			router.ServeHTTP(rr, req)

			// Проверка
			if rr.Code != tc.expectedStatus {
				t.Errorf("ожидался статус %d, но получен %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedError {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("не смогли распарсить JSON ошибки: %v", err)
				}
				if response.Error == "" {
					t.Error("ожидалось сообщение об ошибке, но его нет")
				}
			}
		})
	}
}

// TestTaskHandler_GetAllTasks_TableTests табличные тесты для GetAllTasks
func TestTaskHandler_GetAllTasks_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *MockStorage
		expectedStatus int
		expectedCount  int
	}

	testCases := []testCase{
		{
			name: "получение всех задач из пустого хранилища",
			setupStorage: func() *MockStorage {
				return NewMockStorage()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "получение всех задач из непустого хранилища",
			setupStorage: func() *MockStorage {
				storage := NewMockStorage()
				task1 := &tasks.Task{ID: 1, Title: "Задача 1", Description: "Описание 1"}
				task2 := &tasks.Task{ID: 2, Title: "Задача 2", Description: "Описание 2"}
				storage.tasks[1] = task1
				storage.tasks[2] = task2
				storage.nextID = 3
				return storage
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := tc.setupStorage()
			handler := NewTaskHandler(storage).(*TaskHandler)

			req := httptest.NewRequest("GET", "/tasks", nil)
			rr := httptest.NewRecorder()

			// Действие
			handler.GetAllTasks(rr, req)

			// Проверка
			if rr.Code != tc.expectedStatus {
				t.Errorf("ожидался статус %d, но получен %d", tc.expectedStatus, rr.Code)
			}

			var response []tasks.Task
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("не смогли распарсить JSON списка задач: %v", err)
			}

			if len(response) != tc.expectedCount {
				t.Errorf("ожидалось %d задач, но получено %d", tc.expectedCount, len(response))
			}
		})
	}
}

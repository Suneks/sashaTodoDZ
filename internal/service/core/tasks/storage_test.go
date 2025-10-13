package tasks

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestInMemoryStorage_Create_TableTests табличные тесты для Create
func TestInMemoryStorage_Create_TableTests(t *testing.T) {
	// Определяем структуру тест-кейса
	type testCase struct {
		name        string
		inputTask   *Task
		expectError bool
	}

	// Таблица
	testCases := []testCase{
		{
			name:        "успешное создание задачи",
			inputTask:   &Task{Title: "Test Task", Description: "Test Description"},
			expectError: false,
		},
		{
			name:        "создание задачи и описания на русском",
			inputTask:   &Task{Title: "Первая задача", Description: "Описание первой задачи"},
			expectError: false,
		},
		{
			name:        "создание задачи без описания",
			inputTask:   &Task{Title: "Task Without Description", Description: ""},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := NewInMemoryStorage()

			// Действие
			err := storage.Create(context.Background(), tc.inputTask)

			// Проверка
			if tc.expectError && err == nil {
				t.Errorf("ожидалась ошибка, но её не было")
			}
			if !tc.expectError && err != nil {
				t.Errorf("ожидалась успешная операция, но получена ошибка: %v", err)
			}

			// Если ошибка не ожидалась, проверяем, что задача создалась
			if !tc.expectError {
				if tc.inputTask.ID == 0 {
					t.Error("ожидался ID задачи, но он равен 0")
				}
			}
		})
	}
}

// TestInMemoryStorage_GetByID_TableTests табличные тесты для GetByID
func TestInMemoryStorage_GetByID_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *InMemoryStorage // Функция для подготовки storage
		testID         int                     // ID для поиска
		expectError    bool                    // Ожидаем ли ошибку
		expectNotFound bool                    // Ожидаем ли ErrTaskNotFound
	}

	testCases := []testCase{
		{
			name: "успешное получение существующей задачи",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Existing Task", Description: "Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			testID:         1,
			expectError:    false,
			expectNotFound: false,
		},
		{
			name: "попытка получить несуществующую задачу",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Existing Task", Description: "Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			testID:         999,
			expectError:    true,
			expectNotFound: true,
		},
		{
			name: "попытка получить задачу из пустого хранилища",
			setupStorage: func() *InMemoryStorage {
				return NewInMemoryStorage()
			},
			testID:         1,
			expectError:    true,
			expectNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка
			storage := tc.setupStorage()

			// Действие
			task, err := storage.GetByID(context.Background(), tc.testID)

			// Проверка ошибки
			if tc.expectError && err == nil {
				t.Errorf("ожидалась ошибка, но её не было")
			}
			if !tc.expectError && err != nil {
				t.Errorf("ожидалась успешная операция, но получена ошибка: %v", err)
			}

			// Проверка конкретной ошибки
			if tc.expectNotFound && err != ErrTaskNotFound {
				t.Errorf("ожидалась ошибка ErrTaskNotFound, но получена: %v", err)
			}

			// Проверка результата
			if !tc.expectError && task == nil {
				t.Error("ожидалась задача, но получили nil")
			}
		})
	}
}

// TestInMemoryStorage_Update_TableTests табличные тесты для Update
func TestInMemoryStorage_Update_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *InMemoryStorage
		taskToUpdate   *Task
		expectError    bool
		expectNotFound bool
	}

	testCases := []testCase{
		{
			name: "успешное обновление существующей задачи",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Original Title", Description: "Original Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			taskToUpdate: &Task{
				ID:          1,
				Title:       "Updated Title",
				Description: "Updated Desc",
			},
			expectError:    false,
			expectNotFound: false,
		},
		{
			name: "попытка обновить несуществующую задачу",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Existing Task", Description: "Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			taskToUpdate: &Task{
				ID:          999,
				Title:       "New Title",
				Description: "New Desc",
			},
			expectError:    true,
			expectNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := tc.setupStorage()

			err := storage.Update(context.Background(), tc.taskToUpdate)

			if tc.expectError && err == nil {
				t.Errorf("ожидалась ошибка, но её не было")
			}
			if !tc.expectError && err != nil {
				t.Errorf("ожидалась успешная операция, но получена ошибка: %v", err)
			}

			if tc.expectNotFound && err != ErrTaskNotFound {
				t.Errorf("ожидалась ошибка ErrTaskNotFound, но получена: %v", err)
			}

			// Проверяем, что задача обновилась (если обновление прошло успешно)
			if !tc.expectError {
				updatedTask, getErr := storage.GetByID(context.Background(), tc.taskToUpdate.ID)
				if getErr != nil {
					t.Errorf("не смогли получить обновлённую задачу: %v", getErr)
				} else if updatedTask.Title != tc.taskToUpdate.Title {
					t.Errorf("ожидалось название '%s', но получено '%s'",
						tc.taskToUpdate.Title, updatedTask.Title)
				}
			}
		})
	}
}

// TestInMemoryStorage_Delete_TableTests табличные тесты для Delete
func TestInMemoryStorage_Delete_TableTests(t *testing.T) {
	type testCase struct {
		name           string
		setupStorage   func() *InMemoryStorage
		deleteID       int
		expectError    bool
		expectNotFound bool
	}

	testCases := []testCase{
		{
			name: "успешное удаление существующей задачи",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Task to Delete", Description: "Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			deleteID:       1,
			expectError:    false,
			expectNotFound: false,
		},
		{
			name: "попытка удалить несуществующую задачу",
			setupStorage: func() *InMemoryStorage {
				storage := NewInMemoryStorage()
				task := &Task{Title: "Existing Task", Description: "Desc"}
				storage.Create(context.Background(), task)
				return storage
			},
			deleteID:       999,
			expectError:    true,
			expectNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := tc.setupStorage()

			err := storage.Delete(context.Background(), tc.deleteID)

			if tc.expectError && err == nil {
				t.Errorf("ожидалась ошибка, но её не было")
			}
			if !tc.expectError && err != nil {
				t.Errorf("ожидалась успешная операция, но получена ошибка: %v", err)
			}

			if tc.expectNotFound && err != ErrTaskNotFound {
				t.Errorf("ожидалась ошибка ErrTaskNotFound, но получена: %v", err)
			}

			// Проверяем, что задача действительно удалена (если удаление прошло успешно)
			if !tc.expectError {
				_, getErr := storage.GetByID(context.Background(), tc.deleteID)
				if getErr != ErrTaskNotFound {
					t.Error("ожидалось, что задача удалена, но она всё ещё существует")
				}
			}
		})
	}
}

// TestInMemoryStorage_ConcurrentAccess проверяет конкурентный доступ к хранилищу
func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	storage := NewInMemoryStorage()

	// Запускаем 100 горутин, которые одновременно создают задачи
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Создаём задачи конкурентно
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			task := &Task{
				Title:       "Task " + string(rune(id+'0')),
				Description: "Description " + string(rune(id+'0')),
			}

			// Создаём задачу
			err := storage.Create(context.Background(), task)
			if err != nil {
				t.Errorf("ошибка при создании задачи: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Проверяем, что все задачи были созданы
	tasks, err := storage.GetAll(context.Background())
	if err != nil {
		t.Fatalf("ошибка при получении всех задач: %v", err)
	}

	if len(tasks) != numGoroutines {
		t.Errorf("ожидалось %d задач, но получено %d", numGoroutines, len(tasks))
	}
}

// TestInMemoryStorage_ConcurrentReadWrite проверяет одновременное чтение и запись
func TestInMemoryStorage_ConcurrentReadWrite(t *testing.T) {
	storage := NewInMemoryStorage()

	// Сначала создадим несколько задач
	initialTasks := []*Task{
		{Title: "Task 1", Description: "Desc 1"},
		{Title: "Task 2", Description: "Desc 2"},
		{Title: "Task 3", Description: "Desc 3"},
	}

	for _, task := range initialTasks {
		err := storage.Create(context.Background(), task)
		if err != nil {
			t.Fatalf("ошибка при создании начальных задач: %v", err)
		}
	}

	var wg sync.WaitGroup

	// Запускаем конкурентное чтение
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_, err := storage.GetByID(context.Background(), 1)
			if err != nil && err != ErrTaskNotFound {
				t.Errorf("ошибка при чтении задачи: %v", err)
			}
			time.Sleep(time.Microsecond) // Немного задержки для лучшего чередования
		}
	}()

	// Запускаем конкурентное создание задач
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			task := &Task{
				Title:       "Concurrent Task " + string(rune(i+'0')),
				Description: "Concurrent Desc " + string(rune(i+'0')),
			}
			err := storage.Create(context.Background(), task)
			if err != nil {
				t.Errorf("ошибка при создании задачи: %v", err)
			}
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
}

// TestInMemoryStorage_ConcurrentUpdate проверяет конкурентное обновление
func TestInMemoryStorage_ConcurrentUpdate(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаём задачу для обновления
	initialTask := &Task{Title: "Initial Title", Description: "Initial Description"}
	err := storage.Create(context.Background(), initialTask)
	if err != nil {
		t.Fatalf("ошибка при создании задачи: %v", err)
	}

	const numUpdates = 50
	var wg sync.WaitGroup

	// Запускаем конкурентные обновления одной и той же задачи
	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(updateID int) {
			defer wg.Done()

			task := &Task{
				ID:          1,
				Title:       "Updated Title " + string(rune(updateID+'0')),
				Description: "Updated Description " + string(rune(updateID+'0')),
			}

			err := storage.Update(context.Background(), task)
			if err != nil {
				t.Errorf("ошибка при обновлении задачи: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Проверяем, что задача всё ещё существует (даже если обновлена последним)
	finalTask, err := storage.GetByID(context.Background(), 1)
	if err != nil {
		t.Errorf("ошибка при получении обновлённой задачи: %v", err)
	} else if finalTask == nil {
		t.Error("задача должна существовать после обновления")
	}
}

// TestInMemoryStorage_ConcurrentDelete проверяет конкурентное удаление
func TestInMemoryStorage_ConcurrentDelete(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаём несколько задач для удаления
	for i := 1; i <= 10; i++ {
		task := &Task{
			Title:       "Task " + string(rune(i+'0')),
			Description: "Description " + string(rune(i+'0')),
		}
		err := storage.Create(context.Background(), task)
		if err != nil {
			t.Fatalf("ошибка при создании задачи: %v", err)
		}
	}

	var wg sync.WaitGroup

	// Запускаем конкурентные попытки удаления
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			_ = storage.Delete(context.Background(), id)
			// Не проверяем ошибку - может быть ErrTaskNotFound если уже удалена
		}(i)
	}

	wg.Wait()

	// Проверяем, что все задачи удалены
	allTasks, err := storage.GetAll(context.Background())
	if err != nil {
		t.Errorf("ошибка при получении всех задач: %v", err)
	}

	if len(allTasks) != 0 {
		t.Errorf("ожидалось 0 задач после удаления, но получено %d", len(allTasks))
	}
}

// TestInMemoryStorage_ParallelOperations использует t.Parallel() для проверки конкурентности
func TestInMemoryStorage_ParallelOperations(t *testing.T) {
	// Запускаем несколько тестов параллельно
	t.Run("parallel create operations", func(t *testing.T) {
		t.Parallel()

		storage := NewInMemoryStorage()
		const numTasks = 30

		var wg sync.WaitGroup
		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				task := &Task{
					Title:       "Parallel Task " + string(rune(id+'0')),
					Description: "Parallel Desc " + string(rune(id+'0')),
				}

				err := storage.Create(context.Background(), task)
				if err != nil {
					t.Errorf("ошибка при создании задачи: %v", err)
				}
			}(i)
		}

		wg.Wait()

		tasks, err := storage.GetAll(context.Background())
		if err != nil {
			t.Errorf("ошибка при получении задач: %v", err)
		}

		if len(tasks) != numTasks {
			t.Errorf("ожидалось %d задач, но получено %d", numTasks, len(tasks))
		}
	})

	t.Run("parallel read operations", func(t *testing.T) {
		t.Parallel()

		storage := NewInMemoryStorage()

		// Создаём задачу
		task := &Task{Title: "Read Task", Description: "Read Description"}
		err := storage.Create(context.Background(), task)
		if err != nil {
			t.Fatalf("ошибка при создании задачи: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := storage.GetByID(context.Background(), 1)
				if err != nil && err != ErrTaskNotFound {
					t.Errorf("ошибка при чтении задачи: %v", err)
				}
			}()
		}

		wg.Wait()
	})

	t.Run("parallel update operations", func(t *testing.T) {
		t.Parallel()

		storage := NewInMemoryStorage()

		// Создаём задачу
		task := &Task{Title: "Update Task", Description: "Update Description"}
		err := storage.Create(context.Background(), task)
		if err != nil {
			t.Fatalf("ошибка при создании задачи: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(updateID int) {
				defer wg.Done()

				updatedTask := &Task{
					ID:          1,
					Title:       "Updated Title " + string(rune(updateID+'0')),
					Description: "Updated Description " + string(rune(updateID+'0')),
				}

				err := storage.Update(context.Background(), updatedTask)
				if err != nil && err != ErrTaskNotFound {
					t.Errorf("ошибка при обновлении задачи: %v", err)
				}
			}(i)
		}

		wg.Wait()
	})
}

// TestInMemoryStorage_RaceCondition проверяет, что нет гонок данных
// Для запуска используй: go test -race
func TestInMemoryStorage_RaceCondition(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаём задачу
	task := &Task{Title: "Race Task", Description: "Race Description"}
	err := storage.Create(context.Background(), task)
	if err != nil {
		t.Fatalf("ошибка при создании задачи: %v", err)
	}

	// Запускаем много конкурентных операций
	var wg sync.WaitGroup
	const numOperations = 100

	// Создание задач
	for i := 0; i < numOperations/3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			newTask := &Task{
				Title:       "New Task " + string(rune(id+'0')),
				Description: "New Description " + string(rune(id+'0')),
			}
			storage.Create(context.Background(), newTask)
		}(i)
	}

	// Чтение задач
	for i := 0; i < numOperations/3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storage.GetByID(context.Background(), 1)
		}()
	}

	// Получение всех задач
	for i := 0; i < numOperations/3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storage.GetAll(context.Background())
		}()
	}

	wg.Wait()
}

// TestInMemoryStorage_ConcurrentGetAll проверяет конкурентное получение всех задач
func TestInMemoryStorage_ConcurrentGetAll(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаём несколько задач
	const numInitialTasks = 10
	for i := 1; i <= numInitialTasks; i++ {
		task := &Task{
			Title:       "Initial Task " + string(rune(i+'0')),
			Description: "Initial Description " + string(rune(i+'0')),
		}
		err := storage.Create(context.Background(), task)
		if err != nil {
			t.Fatalf("ошибка при создании начальной задачи: %v", err)
		}
	}

	var wg sync.WaitGroup
	const numConcurrentGetAll = 20

	// Запускаем конкурентные вызовы GetAll
	for i := 0; i < numConcurrentGetAll; i++ {
		wg.Add(1)
		go func(getAllID int) {
			defer wg.Done()

			tasks, err := storage.GetAll(context.Background())
			if err != nil {
				t.Errorf("ошибка при получении всех задач в горутине %d: %v", getAllID, err)
			}

			// Проверяем, что количество задач не меньше начального
			// (может быть больше, если другие горутины создали задачи)
			if len(tasks) < numInitialTasks {
				t.Errorf("в горутине %d ожидалось не менее %d задач, но получено %d",
					getAllID, numInitialTasks, len(tasks))
			}
		}(i)
	}

	wg.Wait()

	// Проверяем, что в хранилище всё ещё есть начальные задачи
	finalTasks, err := storage.GetAll(context.Background())
	if err != nil {
		t.Fatalf("ошибка при получении финальных задач: %v", err)
	}

	if len(finalTasks) < numInitialTasks {
		t.Errorf("ожидалось не менее %d задач в конце, но получено %d",
			numInitialTasks, len(finalTasks))
	}
}

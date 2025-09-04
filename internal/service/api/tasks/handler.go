package tasks

import (
	_ "context"
	"encoding/json"
	"net/http"
	"strconv"
	"toDo/internal/service/core/tasks"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	storage tasks.TaskStorage
}

func NewTaskHandler(storage tasks.TaskStorage) *TaskHandler {
	return &TaskHandler{
		storage: storage,
<<<<<<< HEAD

	// ВОПРОС: Зачем возвращать указатель, а не значение?
	// ВОПРОС: Почему storage публичное поле, а не приватное?
=======
>>>>>>> 76e41d7 (добавил мьютексы)
	}
}

// ErrorResponse структура для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

// TaskRequest структура для входящих данных
type TaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

<<<<<<< HEAD

// CreateTask создает новую задачу
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

// ВОПРОС: Почему лучще использовать отдельный TaskRequest, а не использовать Task напрямую?
	
	var req TaskRequest

=======
// CreateTask создает новую задачу
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
>>>>>>> 76e41d7 (добавил мьютексы)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		h.writeError(w, "title is required", http.StatusBadRequest)
		return
	}

	task := &tasks.Task{
		Title:       req.Title,
		Description: req.Description,
	}

<<<<<<< HEAD
	// ВОПРОС. Использовал контекст, но не понял зачем он тут
=======
>>>>>>> 76e41d7 (добавил мьютексы)
	ctx := r.Context()
	if err := h.storage.Create(ctx, task); err != nil {
		h.writeError(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetAllTasks получает все задачи
func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskList, err := h.storage.GetAll(ctx)
	if err != nil {
		h.writeError(w, "failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskList)
}

// GetTaskByID получает задачу по ID
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	task, err := h.storage.GetByID(ctx, id)
	if err != nil {
		if err == tasks.ErrTaskNotFound {
			h.writeError(w, "task not found", http.StatusNotFound)
			return
		}
		h.writeError(w, "failed to get task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask обновляет задачу
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		h.writeError(w, "title is required", http.StatusBadRequest)
		return
	}

	task := &tasks.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
	}

	ctx := r.Context()
	if err := h.storage.Update(ctx, task); err != nil {
		if err == tasks.ErrTaskNotFound {
			h.writeError(w, "task not found", http.StatusNotFound)
			return
		}
		h.writeError(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// DeleteTask удаляет задачу
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.storage.Delete(ctx, id); err != nil {
		if err == tasks.ErrTaskNotFound {
			h.writeError(w, "task not found", http.StatusNotFound)
			return
		}
		h.writeError(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeError вспомогательный метод для записи ошибок
func (h *TaskHandler) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

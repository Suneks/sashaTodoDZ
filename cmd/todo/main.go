package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	apiTasks "toDo/internal/service/api/tasks"
	coreTasks "toDo/internal/service/core/tasks"
)

func main() {
	// Создаем хранилище
	storage := coreTasks.NewInMemoryStorage()

	// Создаем хендлер
	taskHandler := apiTasks.NewTaskHandler(storage)

	// Создаем роутер
	r := mux.NewRouter()

	// Настраиваем маршруты
	r.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	r.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
	r.HandleFunc("/tasks/{id}", taskHandler.GetTaskByID).Methods("GET")
	r.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")

	// Запускаем сервер
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

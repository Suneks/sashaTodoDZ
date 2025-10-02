package main

import (
	"fmt"
	"strconv"
	"sync"
)

//func main() {
//	// Создаем хранилище
//	storage := coreTasks.NewInMemoryStorage()
//
//	// Создаем хендлер
//	taskHandler := apiTasks.NewTaskHandler(storage)
//
//	// Создаем роутер
//	r := mux.NewRouter()
//
//	// Настраиваем маршруты
//	r.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
//	r.HandleFunc("/tasks", taskHandler.GetAllTasks).Methods("GET")
//	r.HandleFunc("/tasks/{id}", taskHandler.GetTaskByID).Methods("GET")
//	r.HandleFunc("/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
//	r.HandleFunc("/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
//
//	// Запускаем сервер
//	log.Println("Server starting on :8080")
//	log.Fatal(http.ListenAndServe(":8080", r))
//}

// пример 1
func main() {
	var wg sync.WaitGroup
	c := make(chan string, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(с chan<- string, i int, group *sync.WaitGroup) {
			defer wg.Done()
			с <- fmt.Sprintf("Goroutine %s", strconv.Itoa(i))
		}(c, i, &wg)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for {
		select {
		case v := <-c:
			fmt.Println(v)
		}
	}
}

// пример 2. Вывод 1 0
//func main() {
//	a := 0
//	defer fmt.Println(a)
//	defer func() {
//		a++
//		fmt.Println(a)
//	}()
//}

// пример 3. Вывод hello world panic
//func main() {
//	defer fmt.Println("world")
//	fmt.Println("hello")
//	panic("error")
//}

// пример 4. Изначально паника из-за того что записываем в nil мапу, Инициализировал через make
//func main() {
//	var m map[string]int
//	m = make(map[string]int)
//
//	fmt.Println(m["foo"])
//
//	m["foo"] = 42
//
//	fmt.Println(m["foo"])
//}

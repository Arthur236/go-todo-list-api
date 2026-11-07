package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Todo represents a single task item.
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// In-memory database with concurrency support.
var (
	todos   = []Todo{}
	nextID  = 1
	todosMu sync.Mutex
)

func main() {
	// Register route handlers
	http.HandleFunc("/todos", handleTodos)
	http.HandleFunc("/todos/", handleTodoByID)

	fmt.Println("Server running on :8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleTodos handles routing for /todos (GET all, POST create)
func handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		todosMu.Lock()
		json.NewEncoder(w).Encode(todos)
		todosMu.Unlock()

	case http.MethodPost:
		var newTodo Todo

		if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if newTodo.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		todosMu.Lock()
		newTodo.ID = nextID
		nextID++
		todos = append(todos, newTodo)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTodo)
		todosMu.Unlock()

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleTodoByID handles routing for /todos/{id} (GET, PUT update, DELETE)
func handleTodoByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL path (e.g., "/todos/5" -> "5")
	idStr := r.URL.Path[len("/todos/"):]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	todosMu.Lock()
	defer todosMu.Unlock()

	// Find the item index in our slice
	index := -1
	for i, t := range todos {
		if t.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, "Todo item not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(todos[index])

	case http.MethodPut:
		var updatedTodo struct {
			Title     string `json:"title"`
			Completed *bool  `json:"completed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if updatedTodo.Title != "" {
			todos[index].Title = updatedTodo.Title
		}

		if updatedTodo.Completed != nil {
			todos[index].Completed = *updatedTodo.Completed
		}

		json.NewEncoder(w).Encode(todos[index])

	case http.MethodDelete:
		// Remove element from slice
		todos = append(todos[:index], todos[index+1:]...)
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

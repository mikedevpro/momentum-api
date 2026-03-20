package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"momentum-api/internal/handlers"
)

func RegisterRoutes(database *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/tasks", tasksHandler(database))
	mux.HandleFunc("/tasks/", taskByIDHandler(database))

	return mux
}

func tasksHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTasks(database, w, r)
		case http.MethodPost:
			handlers.CreateTask(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func taskByIDHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTaskByID(database, w, r)
		case http.MethodPut:
			handlers.UpdateTask(database, w, r)
		case http.MethodDelete:
			handlers.DeleteTask(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status":  "ok",
		"message": "Momentum API is running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
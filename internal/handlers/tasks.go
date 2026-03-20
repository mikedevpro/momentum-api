package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"momentum-api/internal/models"
)

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

func GetTasks(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	pageParam := r.URL.Query().Get("page")
	limitParam := r.URL.Query().Get("limit")

	if pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	rows, err := database.Query(`
		SELECT id, title, description, completed, created_at
		FROM tasks
		ORDER BY datetime(created_at) DESC, id DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to parse tasks", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	var total int
	err = database.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		http.Error(w, "Failed to count tasks", http.StatusInternalServerError)
		return
	}

	hasNext := (offset + len(tasks)) < total

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Page", strconv.Itoa(page))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Has-Next", strconv.FormatBool(hasNext))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func CreateTask(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	result, err := database.Exec(
		"INSERT INTO tasks (title, description, completed) VALUES (?, ?, ?)",
		req.Title,
		req.Description,
		false,
	)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve ID", http.StatusInternalServerError)
		return
	}

	var newTask models.Task
	err = database.QueryRow(
		`SELECT id, title, description, completed, created_at FROM tasks WHERE id = ?`,
		id,
	).Scan(
		&newTask.ID,
		&newTask.Title,
		&newTask.Description,
		&newTask.Completed,
		&newTask.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Failed to fetch created task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func GetTaskByID(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task

	err = database.QueryRow(
		"SELECT id, title, description, completed, created_at FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req updateTaskRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	result, err := database.Exec(
		"UPDATE tasks SET title = ?, description = ?, completed = ? WHERE id = ?",
		req.Title,
		req.Description,
		req.Completed,
		id,
	)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updatedTask models.Task
	err = database.QueryRow(
		"SELECT id, title, description, completed, created_at FROM tasks WHERE id = ?",
		id,
	).Scan(
		&updatedTask.ID,
		&updatedTask.Title,
		&updatedTask.Description,
		&updatedTask.Completed,
		&updatedTask.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

func DeleteTask(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	result, err := database.Exec(
		"DELETE FROM tasks WHERE id = ?",
		id,
	)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getTaskIDFromPath(path string) (int, error) {
	idStr := strings.TrimPrefix(path, "/tasks/")
	return strconv.Atoi(idStr)
}

package task

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// NewHandlers возвращает Router с привязанными handler'ами
func NewHandlers(svc *Service) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/tasks", createTaskHandler(svc)).Methods("POST")
	r.HandleFunc("/tasks", getAllTasksHandler(svc)).Methods("GET")
	r.HandleFunc("/tasks/{id}", getTaskHandler(svc)).Methods("GET")
	return r
}

// createTaskRequest описывает тело запроса на создание задачи
type createTaskRequest struct {
	URLs []string `json:"urls"`
}

// createTaskHandler создает новую задачу
func createTaskHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		t := svc.Create(r.Context(), req.URLs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	}
}

// getAllTasksHandler возвращает все задачи
func getAllTasksHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks := svc.GetAll(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

// getTaskHandler возвращает задачу по ID
func getTaskHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		t, ok := svc.Get(r.Context(), id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	}
}

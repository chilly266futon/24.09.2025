package task

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// NewHandlersWithPool возвращает Router с привязанными handler'ами и воркер-пулом
func NewHandlersWithPool(svc *Service, pool *WorkerPool) http.Handler {
	r := mux.NewRouter()
	r.Use(RecoveryMW)
	r.Use(LoggingMW)

	r.HandleFunc("/tasks", createTaskHandler(svc, pool)).Methods("POST")
	r.HandleFunc("/tasks", getAllTasksHandler(svc)).Methods("GET")
	r.HandleFunc("/tasks/{id}", getTaskHandler(svc)).Methods("GET")
	return r
}

// createTaskRequest описывает тело запроса на создание задачи
type createTaskRequest struct {
	URLs []string `json:"urls"`
}

// createTaskHandler создает новую задачу и добавляет её в воркер-пул
func createTaskHandler(svc *Service, pool *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		t := svc.Create(r.Context(), req.URLs)
		pool.AddTask(t) // добавляем в очередь воркер-пула

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	}
}

// getAllTasksHandler возвращает все задачи
func getAllTasksHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks := svc.GetAll(r.Context())

		out := make([]map[string]interface{}, 0, len(tasks))
		for _, t := range tasks {
			out = append(out, map[string]interface{}{
				"id":         t.ID,
				"urls":       t.URLs,
				"status":     t.Status.String(),
				"error":      t.Error,
				"created_at": t.CreatedAt,
				"updated_at": t.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

// getTaskHandler возвращает задачу по ID
func getTaskHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		t, ok := svc.Get(r.Context(), id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		out := map[string]interface{}{
			"id":         t.ID,
			"urls":       t.URLs,
			"status":     t.Status.String(),
			"error":      t.Error,
			"created_at": t.CreatedAt,
			"updated_at": t.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

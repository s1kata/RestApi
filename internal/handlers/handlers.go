package handlers

import (
	"encoding/json"
	"net/http"
	database "restapi-tasks/internal/db"
	"restapi-tasks/internal/models"
	"strconv"
	"strings"
)

type Handler struct {
	// store - зависимость, через которую обработчики обращаются к базе.
	store *database.TaskStore
}

// NewHandler создаёт HTTP-обработчик и получает хранилище извне.
func NewHandler(store *database.TaskStore) *Handler {
	return &Handler{
		store: store,
	}
}

func parseTaskIDFromPath(rawPath string) (int, bool) {
	path := strings.TrimPrefix(rawPath, "/tasks")
	path = strings.Trim(path, "/")
	if path == "" {
		return 0, false
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		return 0, false
	}

	return id, true
}

// respondWithJSON формирует единый JSON-ответ для всех endpoint'ов.
func respondWithJSON(w http.ResponseWriter, statuscode int, payload interface{}) {
	// Сообщаем клиенту, что тело ответа содержит JSON.
	w.Header().Set("Content-Type", "application/json")
	// Статус нужно записать до тела ответа.
	w.WriteHeader(statuscode)
	// Encoder преобразует Go-значение в JSON и записывает его в ResponseWriter.
	_ = json.NewEncoder(w).Encode(payload)
}

// respondWithError возвращает ошибку в одинаковом формате {"error": "..."}.
func respondWithError(w http.ResponseWriter, statuscode int, message string) {
	respondWithJSON(w, statuscode, map[string]string{"error": message})
}

// GetAllTasks обрабатывает запрос на получение списка задач.
func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	// Handler делегирует работу с базой объекту store.
	q := r.URL.Query().Get("completed")
	var tasks []models.Task
	var err error

	switch q{
	case "":
		tasks, err = h.store.GetAllFiltered(nil)
	case "true":
		b := true
		tasks, err = h.store.GetAllFiltered(&b)
	case "false":
		b := false
		tasks, err = h.store.GetAllFiltered(&b)
	default:
		respondWithError(w, http.StatusBadRequest,"Некоректный параметр comleted")
		return
	}
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения задач ")
		return
	}
	respondWithJSON(w,http.StatusOK, tasks)
	
}

// GetTask обрабатывает запрос на получение одной задачи по ID из URL.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskIDFromPath(r.URL.Path)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Некорректный id задачи")
		return
	}

	task, err := h.store.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// CreateTask создаёт новую задачу по JSON-данным.
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input models.CreateTaskInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "некореттные отправленные данные")
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Название задачи не может быть пустым")
		return
	}

	task, err := database.CreateTask(h.store, input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, task)
}
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	pathPass := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks"), "/")
	idStr := pathPass[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некоректный id")
		return
	}
	var input models.UpdateTaskInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некоректные данные")
		return
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Заголовок обязателен")
		return
	}
	task, err := database.UpdateTask(h.store, id, input)

	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, task)
}
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskIDFromPath(r.URL.Path)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Некорректный id")
		return
	}

	if err := database.DELETEtask(h.store, id); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

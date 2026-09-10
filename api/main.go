package main

import (
	"log"
	"net/http"
	"os"

	database "restapi-tasks/internal/db"
	"restapi-tasks/internal/handlers"
)

func main() {
    databaseURL := os.Getenv("Database_URL")
    if databaseURL == "" {
        databaseURL = "postgres://taskuser:taskpass@localhost:5432/tasksdb?sslmode=disable"
        log.Println("Используется локальная база данных")
    }

    serverPort := os.Getenv("SERVER_PORT")
    if serverPort == "" {
        serverPort = "8080"
    }

    db, err := database.Connect(databaseURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    log.Println("Успешно подключено к БД")

    taskStore := database.NewTaskStore(db)
    handler := handlers.NewHandler(taskStore)

    mux := http.NewServeMux()
    mux.HandleFunc("/tasks", tasksHandler(handler))
    mux.HandleFunc("/tasks/", taskIDHandler(handler))

    log.Printf("Сервер запущен на порту %s", serverPort)

    if err := http.ListenAndServe(":"+serverPort, loggingMiddleware(mux)); err != nil {
        log.Fatal(err)
    }
}

func tasksHandler(handler *handlers.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            handler.GetAllTasks(w, r)
        case http.MethodPost:
            handler.CreateTask(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

func taskIDHandler(handler *handlers.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            handler.GetTask(w, r)
        case http.MethodPut:
            handler.UpdateTask(w, r)
        case http.MethodDelete:
            handler.DeleteTask(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
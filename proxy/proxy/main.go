package main

import (
	"net/http"	
	handlers "restapi-tasks/handler"
)

func main() {
	BaseURL := "http://localhost:8080"
	handler := handlers.NewHandler(BaseURL)
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodGet:
			handler.GetTask(w,r)
		case http.MethodPost:
			handler.PostTask(w,r)
		}
	

	})
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
			case http.MethodPut:
			handler.PutTask(w,r)
			case http.MethodDelete:
			handler.DeleteTasks(w,r)
			case http.MethodGet:
				handler.GetTaskByID(w,r)
		default:
			http.Error(w, "405", http.StatusMethodNotAllowed)
		}
	})

	
	http.ListenAndServe(":8090",mux)
}

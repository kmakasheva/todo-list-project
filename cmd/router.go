package main

import (
	"github.com/gorilla/mux"
	"github.com/kmakasheva/todo-list-project/handlers"
	"net/http"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/task", handlers.PostTaskHandler).Methods("POST")
	r.HandleFunc("/api/task", handlers.GetTaskHandler).Methods("GET")
	r.HandleFunc("/api/task", handlers.UpdateTaskHandler).Methods("PUT")
	r.HandleFunc("/api/task", handlers.DeleteTaskHandler).Methods("DELETE")
	r.HandleFunc("/api/tasks", handlers.GetTasksHandler)
	r.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	r.HandleFunc("/api/task/done", handlers.DoneTaskHandler)

	webDir := "./web/"
	fileServer := http.FileServer(http.Dir(webDir))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fileServer)) // google why I should do this strip and prefix and why put at last line

	return r
}

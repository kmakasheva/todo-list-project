package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kmakasheva/todo-list-project/handlers"

	//"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	d "github.com/kmakasheva/todo-list-project/db"
	"net/http"
	"os"
)

func main() {
	d.CreateDB()

	r := mux.NewRouter()

	//r.Handle("/", http.FileServer(http.Dir(webDir)))

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

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	port := os.Getenv("TODO_PORT")
	fmt.Printf("starting server on port %s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		fmt.Println(err)
		return
	}
}

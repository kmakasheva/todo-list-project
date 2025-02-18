package main

import (
	"fmt"
	//"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	d "github.com/kmakasheva/todo-list-project/db"
	"net/http"
	"os"
)

func main() {
	d.CreateDB()

	r := SetupRouter()

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

package main

import (
	"fmt"
	"github.com/kmakasheva/todo-list-project/db"
	"github.com/kmakasheva/todo-list-project/handlers"
	"github.com/kmakasheva/todo-list-project/logger"
	"log"

	//"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

func main() {
	db := db.CreateDB()
	defer db.Close()

	handlers.InitDB(db)

	r := SetupRouter()

	//TODO:  передать конфиг а не просто локал
	logger.InitLogger("local")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file %v", err)
	}
	port := os.Getenv("TODO_PORT")
	fmt.Printf("starting server on port %s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Error listening the port %v", err)
	}
}

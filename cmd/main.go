package main

import (
	"fmt"
	"github.com/kmakasheva/todo-list-project/db"
	"github.com/kmakasheva/todo-list-project/handlers"
	"github.com/kmakasheva/todo-list-project/internal/config"
	"github.com/kmakasheva/todo-list-project/logger"
	"net/http"

	//"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	cfg := config.MustLoad()

	Log := logger.SetupLogger(cfg.Env)

	db := db.CreateDB()
	defer db.Close()

	handlers.InitDB(db)

	r := SetupRouter()

	err := godotenv.Load()
	if err != nil {
		Log.Error("Error loading .env file %v", err)
		os.Exit(1)
	}
	port := os.Getenv("TODO_PORT")
	fmt.Printf("starting server on port %s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		Log.Error("Error listening the port %v", err)
		os.Exit(1)
	}
}

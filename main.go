package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	d "github.com/kmakasheva/todo-list-project/db"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	CreateDB()

	http.HandleFunc("/api/nextdate", func(w http.ResponseWriter, r *http.Request) {
		now := r.URL.Query().Get("now")
		date := r.URL.Query().Get("date")
		repeat := r.URL.Query().Get("repeat")
		if now == "" || date == "" || repeat == "" {
			http.Error(w, "Missing required query parametres", http.StatusBadRequest)
			return
		}

		nowtime, err := time.Parse("20060102", now)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		nextDate, err := NextDate(nowtime, date, repeat)
		if err != nil {
			http.Error(w, "Error while calculating next data:", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(nextDate))
	})

	webDir := "./web/"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	port := os.Getenv("TODO_PORT")
	fmt.Printf("starting server on port %s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func CreateDB() {
	appPath := os.Getenv("TODO_DBFILE")

	if appPath == "" {
		nowPath, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory")
			return
		}
		appPath = nowPath
	}

	dbFile := filepath.Join(appPath, "scheduler.db")
	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		if os.IsNotExist(err) {
			install = true
		} else {
			fmt.Println(err)
			return
		}
	}

	_, err = os.Create(dbFile)
	if err != nil {
		fmt.Println("Error creating database file")
		return
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	defer db.Close()

	if install {
		_, err = db.Exec(d.CreateTableQuery)
		if err != nil {
			fmt.Println("Error creating table:", err)
			return
		}
		_, err = db.Exec(d.CreateIndex)
		if err != nil {
			fmt.Println("Error creating index by date:", err)
			return
		}
	}
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	layout := "20060102"
	datetime, err := time.Parse(layout, date)
	var nextdate time.Time
	if err != nil {
		return "", err
	}
	if repeat == "" || now.Format(layout) == "" || date == "" {
		return "", errors.New("Make sure you have a valid data")
	}
	if repeat[0] != 'd' && repeat[0] != 'y' && repeat[0] != 'w' && repeat[0] != 'm' {
		return "", errors.New("The repeat enter is incorrect")
	}
	if repeat == "y" {
		nextdate = datetime.AddDate(1, 0, 0)
		for nextdate.Format(layout) < now.Format(layout) {
			nextdate = nextdate.AddDate(1, 0, 0)
		}
		return nextdate.Format(layout), nil
	}
	value := strings.Split(repeat, " ")
	if value[0] == "d" && len(value) == 2 {
		dnumber, err := strconv.Atoi(value[1])
		if err != nil {
			return "", err
		}
		if dnumber <= 0 || dnumber > 366 {
			fmt.Println("превышен максимально допустимый интервал")
			return "", err
		}
		nextdate = datetime.AddDate(0, 0, dnumber)
		nextdatestr := nextdate.Format(layout)
		for nextdatestr < now.Format(layout) {
			datetime = datetime.AddDate(0, 0, dnumber)
			nextdatestr = datetime.Format(layout)
		}
		return nextdatestr, nil
	} else if value[0] == "d" && len(value) != 2 {
		return "", errors.New("Enter should be in format 'd' days_number")
	}
	return "", nil
}

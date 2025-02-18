package db

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (string, error) {
	appPath := os.Getenv("TODO_DBFILE")

	if appPath == "" {
		nowPath, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory")
			return "", err
		}
		appPath = nowPath
	}

	dbFile := path.Join(appPath, "scheduler.db")
	return dbFile, nil
}

func OpenDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateDB() {
	dbFile, err := InitDB()
	if err != nil {
		//TODO: LOGGER ERROR
		return
	}
	_, err = os.Stat(dbFile)

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

	db, err := OpenDB(dbFile)
	if err != nil {
		// TODO: LOGGER ERROR
	}
	defer db.Close()

	_, err = db.Exec(CreateTableQuery)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
	if install {
		_, err = db.Exec(CreateIndex)
		if err != nil {
			fmt.Println("Error creating index by date:", err)
			return
		}
	}
}

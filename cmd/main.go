package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	//"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	d "github.com/kmakasheva/todo-list-project/db"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id,omitempty" `
	Date    string `json:"date" `
	Title   string `json:"title" `
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat" `
}

func main() {
	d.CreateDB()

	r := mux.NewRouter()

	//r.Handle("/", http.FileServer(http.Dir(webDir)))

	r.HandleFunc("/api/task", PostTaskHandler).Methods("POST")
	r.HandleFunc("/api/task", GetTaskHandler).Methods("GET")
	r.HandleFunc("/api/task", UpdateTaskHandler).Methods("PUT")
	r.HandleFunc("/api/task", DeleteTaskHandler).Methods("DELETE")
	r.HandleFunc("/api/tasks", GetTasksHandler)
	r.HandleFunc("/api/nextdate", NextDateHandler)
	r.HandleFunc("/api/task/done", DoneTaskHandler)

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

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodPost {
		dbFile, err := d.InitDB()
		if err != nil {
			http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
			return
		}

		var task Task
		var nextDate string

		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		//validate := validator.New()
		//if err := validate.Struct(&task); err != nil {
		//		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		//		return
		//	}
		if task.Date == "" {
			task.Date = time.Now().AddDate(0, 0, 0).Format("20060102")
		}

		if _, err := time.Parse("20060102", task.Date); err != nil {
			http.Error(w, `{"error":"Date format seems incorrect"}`, http.StatusBadRequest)
			return
		}

		if task.Date < time.Now().Format("20060102") {
			if task.Repeat == "" {
				task.Date = time.Now().AddDate(0, 0, 1).Format("20060102")
			} else {
				nextDate, err := NextDate(time.Now(), time.Now().AddDate(0, 0, 1).Format("20060102"), task.Repeat)
				if err != nil {
					http.Error(w, `{"error":"something went wrong"}`, http.StatusBadRequest)
					return
				}
				task.Date = nextDate
			}
			w.Write([]byte(nextDate))
		}

		if task.Title == "" {
			http.Error(w, `{"error": "title is empty"}`, http.StatusBadRequest)
			return
		}

		db, err := d.OpenDB(dbFile)
		if err != nil {
			http.Error(w, `{"error" : "error opening db"}`, http.StatusBadRequest)
			return
		}
		defer db.Close()

		res, err := db.Exec(d.InsertData, task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Mistakes while data processing"}`, http.StatusBadRequest)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, `{"error":"error while taking id"}`, http.StatusBadRequest)
			return
		}
		_, err = db.Exec(d.UpdateData, task.Date, id)
		if err != nil {
			http.Error(w, `{"error":"error while updating data in db"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":"%d"}`, id)
	}
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	queryParam := r.URL.Query()
	search := queryParam.Get("search")

	taski := make(map[string][]map[string]string, 0)
	tasks := make([]map[string]string, 0)

	dbFile, err := d.InitDB()
	if err != nil {
		http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
		return
	}

	db, err := d.OpenDB(dbFile)
	if err != nil {
		http.Error(w, `{"error": "error while opening db"}`, http.StatusBadGateway)
		return
	}
	defer db.Close()

	var rows *sql.Rows

	if search == "" {
		EmptySearchRows, err := db.Query(d.GetTasks, 10)
		if err != nil {
			http.Error(w, `{"error":"error while getting sorted tasks"}`, http.StatusInternalServerError)
			return
		}
		rows = EmptySearchRows
		defer EmptySearchRows.Close()
	} else if len(search) == 10 && search[2] == '.' && search[5] == '.' {
		dateTime, err := time.Parse(`02.01.2006`, search)
		if err != nil {
			http.Error(w, `{"error":"error while converting seacrh time from URL"}`, http.StatusBadRequest)
			return
		}
		DateRows, err := db.Query(d.GetTasksByDate, sql.Named("date", dateTime.Format(`20060102`)),
			sql.Named("limit", 10))
		if err != nil {
			http.Error(w, `{"error":"error while getting tasks by date"}`, http.StatusBadRequest)
			return
		}
		rows = DateRows
		defer DateRows.Close()
	} else {
		search = "%" + search + "%"
		WordsRows, err := db.Query(d.GetTasksByWords, sql.Named("word", search), sql.Named("limit", 10))
		if err != nil {
			http.Error(w, `{"error":"error while selecting tasks by words"}`, http.StatusBadRequest)
			return
		}
		rows = WordsRows
		defer WordsRows.Close()
	}

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			http.Error(w, `{"error":"error while scanning rows"}`, http.StatusBadGateway)
			return
		}
		m := map[string]string{
			"id":      task.ID,
			"date":    task.Date,
			"title":   task.Title,
			"comment": task.Comment,
			"repeat":  task.Repeat,
		}
		tasks = append(tasks, m)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"error while iteration"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	taski["tasks"] = tasks

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(taski)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var task Task
		taskMap := make(map[string]string, 0)
		idString := r.URL.Query().Get("id")
		if idString == "" {
			http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
			taskMap["error"] = "Не указан идентификатор"
			return
		}
		id, err := strconv.Atoi(idString)
		if err != nil {
			http.Error(w, `{"error":"error while converting id to int"}`, http.StatusBadRequest)
			return
		}
		dbFile, err := d.InitDB()
		if err != nil {
			http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
			return
		}

		db, err := d.OpenDB(dbFile)
		if err != nil {
			http.Error(w, `{"error": "error while opening db"}`, http.StatusBadGateway)
			return
		}
		defer db.Close()

		row := db.QueryRow(d.GetTaskByID, id)
		if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
			} else {
				http.Error(w, `{"error":"err while selecting task"}`, http.StatusInternalServerError)
			}
			return
		}

		taskMap["id"] = task.ID
		taskMap["date"] = task.Date
		taskMap["title"] = task.Title
		taskMap["comment"] = task.Comment
		taskMap["repeat"] = task.Repeat

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(taskMap)
	}
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"methods except PUT are not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var updatedTask Task

	err := json.NewDecoder(r.Body).Decode(&updatedTask)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "error while decoding body"})
		return
	}

	defer r.Body.Close()

	if err := validateTask(updatedTask); err != nil {
		http.Error(w, `{"error": "incorrect rows in task"}`, http.StatusBadRequest)
		return
	}

	dbFile, err := d.InitDB()
	if err != nil {
		http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
		return
	}
	db, err := d.OpenDB(dbFile)
	if err != nil {
		http.Error(w, `{"error": "error while opening db"}`, http.StatusBadGateway)
		return
	}
	defer db.Close()

	_, err = db.Exec(d.GetTaskByID, sql.Named("id", updatedTask.ID))
	if err != nil {
		http.Error(w, `{"error":"Задача не существует"}`, http.StatusNotFound)
		return
	}

	result, err := db.Exec(d.UpdateTask,
		sql.Named("id", updatedTask.ID),
		sql.Named("date", updatedTask.Date),
		sql.Named("title", updatedTask.Title),
		sql.Named("comment", updatedTask.Comment),
		sql.Named("repeat", updatedTask.Repeat))

	if err != nil {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	if now == "" || date == "" || repeat == "" {
		http.Error(w, "Missing required query parametres", http.StatusBadRequest)
		return
	}

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, "Error while calculating next data:", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(nextDate))
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	layout := "20060102"
	now = time.Date(2024, time.January, 26, 0, 0, 0, 0, time.UTC)
	dateTime, err := time.Parse(layout, date)
	var nextDate time.Time
	if err != nil {
		return "", err
	}

	if repeat == "" || now.Format(layout) == "" || date == "" {
		return "", errors.New("make sure you have a valid data")
	}

	if repeat[0] != 'd' && repeat[0] != 'y' && repeat[0] != 'w' && repeat[0] != 'm' {
		return "", errors.New("the repeat enter is incorrect")
	}

	if repeat == "y" {
		nextDate = dateTime.AddDate(1, 0, 0)
		for nextDate.Format(layout) < now.Format(layout) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format(layout), nil
	}

	value := strings.Split(repeat, " ")

	if value[0] == "d" {
		if len(value) == 2 {
			dNumber, err := strconv.Atoi(value[1])
			if err != nil {
				return "", err
			}
			if dNumber <= 0 || dNumber > 366 {
				fmt.Println("превышен максимально допустимый интервал")
				return "", err
			}
			nextDate = dateTime.AddDate(0, 0, dNumber)
			nextDateStr := nextDate.Format(layout)
			for nextDateStr <= now.Format(layout) {
				dateTime = dateTime.AddDate(0, 0, dNumber)
				nextDateStr = dateTime.Format(layout)
			}
			return nextDateStr, nil
		} else {
			return "", errors.New("enter should be in format 'd' days_number")
		}
	}

	if value[0] == "w" {
		if len(value) == 2 {
			var res []time.Time
			var min time.Time
			//dayOfWeek, err := strconv.Atoi(value[1])
			//fmt.Printf("mya sagan %d\n", dayOfWeek)
			//if err != nil {
			//	return "", err
			//}
			//if dayOfWeek >= 1 && dayOfWeek <= 7 {
			nextDate = dateTime
			daysOfWeek := strings.Split(value[1], ",")
			for i := 0; i < len(daysOfWeek); i++ {
				dayOfWeekInt, err := strconv.Atoi(daysOfWeek[i])
				if err != nil {
					return "", err
				}
				if dayOfWeekInt >= 1 && dayOfWeekInt <= 7 {
					if now.Format(layout) > dateTime.Format(layout) {
						for now.Format(layout) > nextDate.Format(layout) {
							nextDate = nextDate.AddDate(0, 0, 1)
							//time.Sleep(5 * time.Second)
							//fmt.Printf("AAAnextDate: %v, int: %d dayOfweek: %d\n", nextDate, int(nextDate.Weekday())+1, dayOfWeekInt)
						}
						for int(nextDate.Weekday())+1 != dayOfWeekInt {
							nextDate = nextDate.AddDate(0, 0, 1)
							//time.Sleep(5 * time.Second)
							//fmt.Printf("BBBnextDate: %v, int: %d dayOfweek: %d\n", nextDate, int(nextDate.Weekday())+1, dayOfWeekInt)
						}
						if int(nextDate.Weekday())+1 == dayOfWeekInt {
							res = append(res, nextDate)
						}
					} else {
						for int(nextDate.Weekday())+1 != dayOfWeekInt {
							nextDate = nextDate.AddDate(0, 0, 1)
							//time.Sleep(5 * time.Second)
							//fmt.Printf("nextDate: %v, int: %d dayOfweek: %d\n", nextDate, int(nextDate.Weekday())+1, dayOfWeekInt)
						}
						if int(nextDate.Weekday())+1 == dayOfWeekInt {
							res = append(res, nextDate)
							nextDate = dateTime
						}
					}
					min = res[0]
					for _, v := range res {
						if v.Format(layout) < min.Format(layout) {
							min = v
						}
					}
				} else {
					return "", errors.New("make sure you have a valid date for week")
				}
			}
			nextDate = min.AddDate(0, 0, 1)
			return nextDate.Format(layout), nil
			//} else {
			//	return "", errors.New("Invalid weekday format")
			//}

		} else {
			return "", errors.New("enter should be in format 'w' weeks number")
		}
	}

	if value[0] == "m" {
		if dateTime.Format(layout) < now.Format(layout) {
			dateTime = now
		}
		var res []time.Time
		var min time.Time
		if len(value) == 2 {
			nextDate = dateTime.AddDate(0, 0, 1)
			daysOfMonth := strings.Split(value[1], ",")
			for _, dayOfMonth := range daysOfMonth {
				dayOfMonthInt, err := strconv.Atoi(dayOfMonth)
				if err != nil || dayOfMonthInt < -2 || dayOfMonthInt > 31 {
					return "", errors.New("enter should be in format 'm' days_number")
				}
				if dayOfMonthInt == -1 {
					nextDate = nextDate.AddDate(0, 1, 0)
					nextDate = nextDate.AddDate(0, 0, -nextDate.Day())
					dayOfMonthInt = nextDate.Day()
				} else if dayOfMonthInt == -2 {
					nextDate = nextDate.AddDate(0, 1, 0)
					nextDate = nextDate.AddDate(0, 0, -nextDate.Day()-1)
					dayOfMonthInt = nextDate.Day()
				}
				for nextDate.Day() != dayOfMonthInt {
					nextDate = nextDate.AddDate(0, 0, 1)
				}
				if nextDate.Day() == dayOfMonthInt {
					res = append(res, nextDate)
				}
				nextDate = dateTime.AddDate(0, 0, 1)
			}
			min = res[0]
			for _, v := range res {
				if min.Format(layout) > v.Format(layout) {
					min = v
				}
			}
			return min.Format(layout), nil
		}
		if len(value) == 3 {
			nextDate = dateTime.AddDate(0, 0, 1)
			daysOfMonth := strings.Split(value[1], ",")
			months := strings.Split(value[2], ",")
			for _, month := range months {
				for _, dayOfMonth := range daysOfMonth {
					monthInt, err := strconv.Atoi(month)
					if err != nil || monthInt <= 0 || monthInt > 12 {
						return "", errors.New("make sure you have entered correct months")
					}
					dayOfMonthInt, err := strconv.Atoi(dayOfMonth)
					if err != nil {
						return "", errors.New("make sure you have entered correct days of months")
					}
					for int(nextDate.Month()) != monthInt {
						nextDate = nextDate.AddDate(0, 1, 0)
					}
					if int(nextDate.Month()) == monthInt {
						for nextDate.Day() != dayOfMonthInt {
							nextDate = nextDate.AddDate(0, 0, 1)
						}
						if nextDate.Day() == dayOfMonthInt && int(nextDate.Month()) == monthInt && nextDate.Format(layout) > now.Format(layout) {
							res = append(res, nextDate)
							nextDate = dateTime.AddDate(0, 0, 1)
						} else if nextDate.Day() == dayOfMonthInt && int(nextDate.Month())-1 == monthInt &&
							nextDate.AddDate(0, -1, 0).Format(layout) > now.Format(layout) {
							res = append(res, nextDate.AddDate(0, -1, 0))
							nextDate = dateTime.AddDate(0, 0, 1)
						}
					}
				}
			}
			min = res[0]
			for _, v := range res {
				if v.Format(layout) < min.Format(layout) {
					min = v
				}
			}
			return min.Format(layout), nil
		} else if len(value) != 2 && len(value) != 3 {
			return "", errors.New("enter should be in format 'm' days_number month_number")
		}
	}

	return "", nil
}

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, `{"error":"error while converting id from str to int"}`, http.StatusBadRequest)
		return
	}

	dbFile, err := d.InitDB()
	if err != nil {
		http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
		return
	}

	db, err := d.OpenDB(dbFile)
	if err != nil {
		http.Error(w, `{"error": "error while opening db"}`, http.StatusBadGateway)
		return
	}
	defer db.Close()

	row := db.QueryRow(d.GetTaskByID, sql.Named("id", id))

	var task Task

	if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		http.Error(w, `{"error":"error while getting task by id"}`, http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		updDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"error while calculating a new date"}`, http.StatusBadRequest)
			return
		}

		_, err = db.Exec(d.UpdateTask,
			sql.Named("id", id),
			sql.Named("date", updDate),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat))

		if err != nil {
			http.Error(w, `{"error":"error while updating date to new"}`, http.StatusInternalServerError)
			return
		}
	} else {
		_, err := db.Exec(d.DeleteTask, sql.Named("id", id))
		if err != nil {
			http.Error(w, `{"error":"error while deleting task"}`, http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Cotnent-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})

}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed only delete is allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	idString := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, `{"error":"error while converting id from str to int"}`, http.StatusBadRequest)
		return
	}

	dbFile, err := d.InitDB()
	if err != nil {
		http.Error(w, `{"error": "error finding db"}`, http.StatusNotFound)
		return
	}

	db, err := d.OpenDB(dbFile)
	if err != nil {
		http.Error(w, `{"error": "error while opening db"}`, http.StatusBadGateway)
		return
	}
	defer db.Close()

	_, err = db.Exec(d.DeleteTask, sql.Named("id", id))
	if err != nil {
		http.Error(w, `{"error":"error while deleting not actual task"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})
}

func validateTask(t Task) error {
	id, err := strconv.Atoi(t.ID)
	if err != nil {
		return errors.New("id seems problematic")
	}
	if id < 0 {
		return errors.New("ID is not valid")
	}
	if _, err := time.Parse("20060102", t.Date); err != nil {
		return errors.New("date is not valid")
	}
	if t.Title == "" {
		return errors.New("title cannot be empty")
	}
	if t.Repeat == "" {
		return errors.New("repeat is empty")
	}
	if t.Repeat[0] != 'y' && t.Repeat[0] != 'm' && t.Repeat[0] != 'w' && t.Repeat[0] != 'd' {
		return errors.New("invalid repeat format")
	}
	row := strings.Split(t.Repeat, " ")
	if row[0] == "y" && len(row) != 1 {
		return errors.New("only one year repeat is available")
	}
	if row[0] == "m" && len(row) > 3 {
		return errors.New("incorrect repeat in months")
	}
	if row[0] == "w" && len(row) != 2 {
		return errors.New("repeat in weeks is not valid")
	}
	if row[0] == "d" && len(row) != 2 {
		return errors.New("repeat in days is not valid")
	}
	return nil
}

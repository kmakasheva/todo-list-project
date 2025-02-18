package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	d "github.com/kmakasheva/todo-list-project/db"
	"github.com/kmakasheva/todo-list-project/domain"
	"github.com/kmakasheva/todo-list-project/services"
	"net/http"
	"strconv"
	"time"
)

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

		var task domain.Task
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
				nextDate, err := services.NextDate(time.Now(), time.Now().AddDate(0, 0, 1).Format("20060102"), task.Repeat)
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

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var task domain.Task
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
		var task domain.Task
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

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"methods except PUT are not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var updatedTask domain.Task

	err := json.NewDecoder(r.Body).Decode(&updatedTask)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "error while decoding body"})
		return
	}

	defer r.Body.Close()

	if err := services.ValidateTask(updatedTask); err != nil {
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

	nextDate, err := services.NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, "Error while calculating next data:", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(nextDate))
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

	var task domain.Task

	if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		http.Error(w, `{"error":"error while getting task by id"}`, http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		updDate, err := services.NextDate(time.Now(), task.Date, task.Repeat)
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

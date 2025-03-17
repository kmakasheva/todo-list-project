package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/kmakasheva/todo-list-project/auth"
	d "github.com/kmakasheva/todo-list-project/db"
	"github.com/kmakasheva/todo-list-project/domain"
	"github.com/kmakasheva/todo-list-project/logger"
	"github.com/kmakasheva/todo-list-project/services"
	"net/http"
	"os"
	"strconv"
	"time"
)

var db *sql.DB
var MYTOKEN string

func InitDB(sqlDB *sql.DB) {
	db = sqlDB
}

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task domain.Task
	var nextDate string

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		logger.Log.Error("Invalid JSON")
		return
	}
	defer r.Body.Close()

	if task.Date == "" {
		task.Date = time.Now().AddDate(0, 0, 0).Format("20060102")
	}

	if _, err := time.Parse("20060102", task.Date); err != nil {
		http.Error(w, `{"error":"Date format seems incorrect"}`, http.StatusBadRequest)
		logger.Log.Error("Date format seems incorrect")
		return
	}

	if task.Date < time.Now().Format("20060102") {
		if task.Repeat == "" {
			task.Date = time.Now().AddDate(0, 0, 1).Format("20060102")
		} else {
			nextDate, err := services.NextDate(time.Now(), time.Now().AddDate(0, 0, 1).Format("20060102"), task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"something went wrong"}`, http.StatusBadRequest)
				logger.Log.Error("something went wrong")
				return
			}
			task.Date = nextDate
		}
		w.Write([]byte(nextDate))
	}

	if task.Title == "" {
		http.Error(w, `{"error": "title is empty"}`, http.StatusBadRequest)
		logger.Log.Error("title is empty")
		return
	}

	res, err := db.Exec(d.InsertData, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Mistakes while data processing"}`, http.StatusBadRequest)
		logger.Log.Error("Mistakes while data processing")
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"error while taking id"}`, http.StatusBadRequest)
		logger.Log.Error("error while taking id")
		return
	}
	_, err = db.Exec(d.UpdateData, task.Date, id)
	if err != nil {
		http.Error(w, `{"error":"error while updating data in db"}`, http.StatusBadRequest)
		logger.Log.Error("error while updating data in db")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":"%d"}`, id)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task domain.Task
	//taskMap := make(map[string]string, 0)
	idString := r.URL.Query().Get("id")
	if idString == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		//taskMap["error"] = "Не указан идентификатор"
		logger.Log.Error("Не указан идентификатор")
		return
	}
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, `{"error":"error while converting id to int"}`, http.StatusBadRequest)
		logger.Log.Error("error while converting id to int")
		return
	}

	row := db.QueryRow(d.GetTaskByID, id)
	if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
			logger.Log.Error("Задача не найдена")
		} else {
			http.Error(w, `{"error":"err while selecting task"}`, http.StatusInternalServerError)
			logger.Log.Error("err while selecting task")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	queryParam := r.URL.Query()
	search := queryParam.Get("search")

	tasks := make([]domain.Task, 0)
	taski := map[string]interface{}{"tasks": tasks}

	var rows *sql.Rows

	if search == "" {
		EmptySearchRows, err := db.Query(d.GetTasks, 10)
		if err != nil {
			http.Error(w, `{"error":"error while getting sorted tasks"}`, http.StatusInternalServerError)
			logger.Log.Error("error while getting sorted tasks")
			return
		}
		rows = EmptySearchRows
		defer EmptySearchRows.Close()
	} else if len(search) == 10 && search[2] == '.' && search[5] == '.' {
		dateTime, err := time.Parse(`02.01.2006`, search)
		if err != nil {
			http.Error(w, `{"error":"error while converting seacrh time from URL"}`, http.StatusBadRequest)
			logger.Log.Error("error while converting seacrh time from URL")
			return
		}
		DateRows, err := db.Query(d.GetTasksByDate, sql.Named("date", dateTime.Format(`20060102`)),
			sql.Named("limit", 10))
		if err != nil {
			http.Error(w, `{"error":"error while getting tasks by date"}`, http.StatusBadRequest)
			logger.Log.Error("error while getting tasks by date")
			return
		}
		rows = DateRows
		defer DateRows.Close()
	} else {
		search = "%" + search + "%"
		WordsRows, err := db.Query(d.GetTasksByWords, sql.Named("word", search), sql.Named("limit", 10))
		if err != nil {
			http.Error(w, `{"error":"error while selecting tasks by words"}`, http.StatusBadRequest)
			logger.Log.Error("error while selecting tasks by words")
			return
		}
		rows = WordsRows
		defer WordsRows.Close()
	}

	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			http.Error(w, `{"error":"error while scanning rows"}`, http.StatusBadGateway)
			logger.Log.Error("error while scanning rows")
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"error while iteration"}`, http.StatusInternalServerError)
		logger.Log.Error("error while iteration")
		return
	}
	defer rows.Close()

	taski["tasks"] = tasks

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(taski)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {

	var updatedTask domain.Task

	err := json.NewDecoder(r.Body).Decode(&updatedTask)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "error while decoding body"})
		logger.Log.Error("error while decoding body")
		return
	}

	defer r.Body.Close()

	if err := services.ValidateTask(updatedTask); err != nil {
		http.Error(w, `{"error": "incorrect rows in task"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Exec(d.GetTaskByID, sql.Named("id", updatedTask.ID))
	if err != nil {
		http.Error(w, `{"error":"Задача не существует"}`, http.StatusNotFound)
		logger.Log.Error("Задача не существует")
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
		logger.Log.Error("Задача не найдена")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		logger.Log.Error("Задача не найдена")
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
		logger.Log.Error("Invalid date format")
		return
	}

	nextDate, err := services.NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, "Error while calculating next data:", http.StatusInternalServerError)
		logger.Log.Error("Error while calculating next data")
		return
	}
	w.Write([]byte(nextDate))
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {

	idString := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, `{"error":"error while converting id from str to int"}`, http.StatusBadRequest)
		logger.Log.Error("error while converting id from str to int")
		return
	}

	_, err = db.Exec(d.DeleteTask, sql.Named("id", id))
	if err != nil {
		http.Error(w, `{"error":"error while deleting not actual task"}`, http.StatusBadRequest)
		logger.Log.Error("error while deleting not actual task")
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
		logger.Log.Error("error while converting id from str to int")
		return
	}

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
			logger.Log.Error("error while calculating a new date")
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
			logger.Log.Error("error while updating date to new")
			return
		}
	} else {
		_, err := db.Exec(d.DeleteTask, sql.Named("id", id))
		if err != nil {
			http.Error(w, `{"error":"error while deleting task"}`, http.StatusBadRequest)
			logger.Log.Error("error while deleting task")
			return
		}
	}
	w.Header().Set("Cotnent-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})

}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("TOD0_PASSWORDD") == "" {
		http.Error(w, `{"error":"Пароль не установлен в переменных окружения"}`, http.StatusInternalServerError)
		logger.Log.Error("Пароль не установлен в переменных окружения")
		return
	}
	var request auth.PasswordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, `{"error":"error while decoding inputted password"}`, http.StatusBadRequest)
		logger.Log.Error("error while decoding inputted password")
		return
	}
	if request.Password != os.Getenv("TOD0_PASSWORDD") {
		http.Error(w, `{"error":"Неверный пароль"}`, http.StatusUnauthorized)
		logger.Log.Error("error Неверный пароль")
		return
	}

	token, err := auth.CreateJWT()
	if err != nil {
		http.Error(w, `{"error":"error getting cookie token"}`, http.StatusUnauthorized)
		logger.Log.Error("error getting cookie token")
		return
	}

	if token == "" {
		http.Error(w, `{"error":"Ошибка при создании токена"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   2 * 60 * 60,
		Secure:   false,
	})
	logger.Log.Info("Token установлен, пользователь вошел в систему")

	MYTOKEN = token
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Успешный вход", "token":"` + token + `"}`))

}

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var db *sql.DB

type Task struct {
	ID      int64  `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Неверный формат времени: %v", err), http.StatusBadRequest)
	}
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	resp, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, resp)
}

func NewTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var task Task
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &task)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка обработки запроса: %v", err), http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		http.Error(w, "Поле тайтл не дожно быть пустым", http.StatusBadRequest)
		return
	}
	_, err = time.Parse("20060102", task.Date)
	if err != nil {
		task.Date = time.Now().Format("20060102")
	}
	if task.Repeat != "" {
		var nextDataErr error
		task.Date, nextDataErr = NextDate(time.Now(), task.Date, task.Repeat)
		if nextDataErr != nil {
			http.Error(w, fmt.Sprint(nextDataErr), http.StatusBadRequest)
			return
		}
		res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES ($1, $2, $3, $4)", task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка записи в базу данных: %v", err), http.StatusInternalServerError)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка на стороне сервера"), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		jsID, err := json.Marshal(map[string]int64{"id": id})
		if err != nil {
			http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(jsID)
	}
}

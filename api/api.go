package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/paxaf/go_final_project/database"
)

type task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Неверный формат времени: %v", err), http.StatusBadRequest)
		return
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

func AddTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var task task
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &task)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Ошибка обработки запроса: %v", err), http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		respondWithError(w, "Поле 'title' не может быть пустым", http.StatusBadRequest)
		return
	}

	_, err = time.Parse("20060102", task.Date)
	if err != nil {
		respondWithError(w, "Ошибка распознования времени", http.StatusBadRequest)
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	if task.Repeat != "" {
		task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, fmt.Sprintf("Ошибка вычисления следующей даты: %v", err), http.StatusBadRequest)
			return
		}
	}

	db, err := database.Dbinit()
	if err != nil {
		respondWithError(w, fmt.Sprintf("Ошибка инициализации базы данных: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES ($1, $2, $3, $4);", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Ошибка записи в базу данных: %v", err), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		respondWithError(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// respondWithError отправляет ошибку в формате JSON
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// respondWithJSON отправляет JSON-ответ
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
		return
	}

	log.Printf("Тело запроса: %s", body) // Логирование для отладки

	var task task
	if err := json.Unmarshal(body, &task); err != nil {
		respondWithError(w, "Ошибка обработки запроса", http.StatusBadRequest)
		return
	}

	if strings.ReplaceAll(task.Title, " ", "") == "" {
		respondWithError(w, "Поле 'title' не может быть пустым", http.StatusBadRequest)
		return
	}

	if strings.ReplaceAll(task.Date, " ", "") == "" {
		task.Date = time.Now().Format("20060102")
	} else {
		_, err = time.Parse("20060102", task.Date)
		if err != nil {
			respondWithError(w, "Ошибка распознавания времени", http.StatusBadRequest)
			return
		}
	}
	task.Repeat = strings.TrimSpace(task.Repeat)
	if task.Repeat != "" {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, "Ошибка вычисления следующей даты", http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else {
		currentDate := time.Now().Format("20060102")
		if task.Date < currentDate {
			respondWithError(w, "Дата не может быть меньше сегодняшней", http.StatusBadRequest)
			return
		}
	}
	db := database.DB
	if err != nil {
		respondWithError(w, "Ошибка инициализации базы данных", http.StatusInternalServerError)
		return
	}
	var id int64

	err = db.QueryRow(
		`INSERT INTO scheduler (date, title, comment, repeat) 
         VALUES ($1, $2, $3, $4) RETURNING id`,
		task.Date, task.Title, task.Comment, task.Repeat,
	).Scan(&id)
	if err != nil {
		log.Printf("Ошибка записи в БД: %v", err) // Добавьте это
		respondWithError(w, "Ошибка записи в базу данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
	}
}
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
	}
}

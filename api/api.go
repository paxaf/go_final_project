package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/paxaf/go_final_project/database"
)

type loginRequest struct {
	Password string `json:"password"`
}

type task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
type tasksResponse struct {
	Tasks []task `json:"tasks"`
}

var userDate time.Time

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
		userDate, err = time.Parse("20060102", task.Date)
		if err != nil {
			respondWithError(w, "Ошибка распознавания времени", http.StatusBadRequest)
			return
		}
	}
	task.Repeat = strings.TrimSpace(task.Repeat)
	dateRep, err := time.Parse("20060102", task.Date)
	if err != nil {
		dateRep = time.Now()
	}
	if task.Repeat != "" && dateRep.Before(time.Now().Truncate(24*time.Hour)) {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, "Ошибка вычисления следующей даты", http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else {
		if userDate.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}
	db := database.DB
	var id int64

	err = db.QueryRow(
		`INSERT INTO scheduler (date, title, comment, repeat) 
         VALUES ($1, $2, $3, $4) RETURNING id`,
		task.Date, task.Title, task.Comment, task.Repeat,
	).Scan(&id)
	if err != nil {
		log.Printf("Ошибка записи в БД: %v", err)
		respondWithError(w, "Ошибка записи в базу данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
func Tasks(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	searchTime, err := time.Parse("02.01.2006", search)
	var tasks []task
	var rows *sql.Rows
	db := database.DB
	switch {
	case err == nil:
		search = searchTime.Format("20060102")
		rows, err = db.Query("SELECT CAST(id AS TEXT), date, title, comment, repeat FROM scheduler WHERE date = :search_date ORDER BY date ASC;", sql.Named("search_date", search))
		if err != nil {
			respondWithError(w, ("Ошибка на стороне сервера"), http.StatusInternalServerError)
			return
		}
	case search != "" && err != nil:
		search = "%" + search + "%"
		rows, err = db.Query("SELECT CAST(id AS TEXT), date, title, comment, repeat FROM scheduler WHERE title LIKE :search_text OR comment LIKE :search_text ORDER BY date ASC;", sql.Named("search_text", search))
		if err != nil {
			respondWithError(w, ("Ошибка на стороне сервера"), http.StatusInternalServerError)
			return
		}
	default:
		rows, err = db.Query("SELECT CAST(id AS TEXT), date, title, comment, repeat FROM scheduler ORDER BY date ASC;")
		if err != nil {
			respondWithError(w, ("Ошибка на стороне сервера"), http.StatusInternalServerError)
			return
		}
	}

	for rows.Next() {
		var task task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			respondWithError(w, ("Ошибка преобразования из базы данных"), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		respondWithError(w, "Ошибка при обработке данных", http.StatusInternalServerError)
		return
	}
	if len(tasks) == 0 {
		respondWithJSON(w, http.StatusOK, tasksResponse{Tasks: []task{}})
	} else {
		respondWithJSON(w, http.StatusOK, tasksResponse{Tasks: tasks})
	}
}
func Task(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, ("Некорректный номер задачи"), http.StatusBadRequest)
		return
	}
	var task task
	db := database.DB
	err = db.QueryRow("SELECT CAST(id AS TEXT), date, title, comment, repeat FROM scheduler WHERE id = :id;", sql.Named("id", idInt)).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		respondWithError(w, ("Такой id не найден"), http.StatusBadRequest)
		return
	}
	respondWithJSON(w, http.StatusOK, task)
}
func EditTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
		return
	}
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
		userDate, err = time.Parse("20060102", task.Date)
		if err != nil {
			respondWithError(w, "Ошибка распознавания времени", http.StatusBadRequest)
			return
		}
	}
	task.Repeat = strings.TrimSpace(task.Repeat)
	dateRep, err := time.Parse("20060102", task.Date)
	if err != nil {
		dateRep = time.Now()
	}
	if task.Repeat != "" && dateRep.Before(time.Now().Truncate(24*time.Hour)) {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, "Ошибка вычисления следующей даты", http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else {
		if userDate.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}
	db := database.DB
	idInt, err := strconv.Atoi(task.Id)
	if err != nil {
		respondWithError(w, ("Задача не найдена"), http.StatusBadRequest)
		return
	}
	res, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id", sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat), sql.Named("id", idInt))
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, ("Задача не найдена"), http.StatusBadRequest)
		return
	}
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func Done(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	db := database.DB
	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, ("Некорректный id задачи"), http.StatusBadRequest)
		return
	}
	var date string
	var repeat string
	err = db.QueryRow("SELECT date, repeat FROM scheduler WHERE id = :id", sql.Named("id", idInt)).Scan(&date, &repeat)
	if err != nil {
		respondWithError(w, ("Задача не найдена"), http.StatusNotFound)
		return
	}
	if repeat == "" {
		res, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", idInt))
		if err != nil {
			respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := res.RowsAffected()
		if rowsAffected == 0 {
			respondWithError(w, ("Задача не найдена"), http.StatusBadRequest)
			return
		}
		if err != nil {
			respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}
	date, err = NextDate(time.Now(), date, repeat)
	if err != nil {
		respondWithError(w, fmt.Sprintf("ошибка :%v", err), http.StatusInternalServerError)
		return
	}
	res, err := db.Exec("UPDATE scheduler SET date = :date WHERE id = :id", sql.Named("date", date), sql.Named("id", idInt))
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, ("Задача не найдена"), http.StatusBadRequest)
		return
	}
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func DelTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	db := database.DB
	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, ("Некорректный id задачи"), http.StatusBadRequest)
		return
	}
	res, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", idInt))
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, ("Задача не найдена"), http.StatusBadRequest)
		return
	}
	if err != nil {
		respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, ("Ошибка запроса"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	envPass := os.Getenv("TODO_PASSWORD")
	if req.Password != envPass {
		respondWithError(w, ("Ошибка авторизации"), http.StatusUnauthorized)
		return
	}
	hash := sha256.Sum256([]byte(envPass))
	hashString := hex.EncodeToString(hash[:])
	claims := jwt.RegisteredClaims{
		Subject: hashString,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		respondWithError(w, ("ошибка подписи токена"), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusAccepted, map[string]string{"token": tokenString})
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

package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/paxaf/go_final_project/internal/models"
	"github.com/paxaf/go_final_project/internal/repository"
	"github.com/paxaf/go_final_project/internal/service"
)

const FormatTime string = "20060102"

type loginRequest struct {
	Password string `json:"password"`
}

var userDate time.Time

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse(FormatTime, r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Неверный формат времени: %v", err), http.StatusBadRequest)
		return
	}
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	resp, err := service.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, resp)
}

func AddTask(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
			return
		}

		var task models.Task
		if err := json.Unmarshal(body, &task); err != nil {
			respondWithError(w, "Ошибка обработки запроса", http.StatusBadRequest)
			return
		}

		err = service.Validate(&task)
		if err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := repo.Create(task)
		if err != nil {
			log.Printf("Ошибка записи в БД: %v", err)
			respondWithError(w, "Ошибка записи в базу данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
	}
}
func Tasks(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		tasks, err := repo.SearchTasks(search)
		if err != nil {
			respondWithError(w, fmt.Sprintf("ошибка: %v", err), http.StatusBadRequest)
		}
		if len(tasks) == 0 {
			respondWithJSON(w, http.StatusOK, models.TasksResponse{Tasks: []models.Task{}})
		} else {
			respondWithJSON(w, http.StatusOK, models.TasksResponse{Tasks: tasks})
		}
	}

}
func Task(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		task, err := repo.GetByID(id)
		if err != nil {
			respondWithError(w, ("Такой id не найден"), http.StatusBadRequest)
			return
		}
		respondWithJSON(w, http.StatusOK, task)
	}

}
func EditTask(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
			return
		}
		var task models.Task
		if err := json.Unmarshal(body, &task); err != nil {
			respondWithError(w, "Ошибка обработки запроса", http.StatusBadRequest)
			return
		}
		err = service.Validate(&task)
		if err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = repo.Update(task)
		if err != nil {
			respondWithError(w, ("Ошибка обращения к базе данных"), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{})
	}
}
func Done(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		err := repo.Done(id)
		if err != nil {
			respondWithError(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{})
	}
}
func DelTask(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		err := repo.Delete(id)
		if err != nil {
			respondWithError(w, fmt.Sprintf("Ошибка обращения к базе данных: %v", err), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{})
	}
}
func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, "Ошибка запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	envPass := os.Getenv("TODO_PASSWORD")
	if req.Password != envPass {
		respondWithError(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}
	hash := sha256.Sum256([]byte(envPass))
	hashString := hex.EncodeToString(hash[:])
	claims := jwt.RegisteredClaims{
		Subject: hashString,
	}
	secretkey := os.Getenv("TODO_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretkey))
	if err != nil {
		respondWithError(w, "ошибка подписи токена", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusAccepted, map[string]string{"token": tokenString})
}
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie("token")
		if err != nil {
			respondWithError(w, "Ошибка авторизации", http.StatusUnauthorized)
			return
		}
		jwtToken := cookie.Value
		secret := os.Getenv("TODO_SECRET")
		parsedToken, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !parsedToken.Valid {
			respondWithError(w, "Неверный токен", http.StatusUnauthorized)
			return
		}
		claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
		if !ok {
			respondWithError(w, "Ошибка чтения токена", http.StatusUnauthorized)
			return
		}
		hash := sha256.Sum256([]byte(pass))
		expectedHash := hex.EncodeToString(hash[:])
		if claims.Subject != expectedHash {
			respondWithError(w, "Неверные учетные данные", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/paxaf/go_final_project/internal/models"
	"github.com/paxaf/go_final_project/internal/repository"
	"github.com/paxaf/go_final_project/internal/service"
)

const FormatTime string = "20060102"

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

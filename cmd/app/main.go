package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/paxaf/go_final_project/internal/handlers"
	"github.com/paxaf/go_final_project/internal/repository"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		//	log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	repo, err := repository.Dbinit()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных")
	}
	defer repo.DB.Close()
	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Group(func(r chi.Router) {
		r.Mount("/", fileServer)
		r.Post("/api/signin", handlers.Login)
		r.Get("/api/nextdate", handlers.NextDateHandler)
	})
	pass := os.Getenv("TODO_PASSWORD")
	secret := os.Getenv("TODO_SECRET")
	r.Group(func(r chi.Router) {
		r.Use(handlers.Auth(pass, secret))
		r.Get("/api/tasks", handlers.Tasks(repo))
		r.Get("/api/task", handlers.Task(repo))
		r.Post("/api/task", handlers.AddTask(repo))
		r.Put("/api/task", handlers.EditTask(repo))
		r.Post("/api/task/done", handlers.Done(repo))
		r.Delete("/api/task", handlers.DelTask(repo))
	})

	port := os.Getenv("TODO_PORT")

	if len(port) < 1 {
		port = "7540"
	}
	log.Printf("Запуск на порте: %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %s", err.Error())
	}
}

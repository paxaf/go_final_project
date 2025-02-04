package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/paxaf/go_final_project/api"
	"github.com/paxaf/go_final_project/database"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		//	log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	err = database.Dbinit()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных")
	}
	defer database.DB.Close()
	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Group(func(r chi.Router) {
		r.Mount("/", fileServer)
		r.Post("/api/signin", api.Login)
		r.Get("/api/nextdate", api.NextDateHandler)
	})
	r.Group(func(r chi.Router) {
		r.Use(api.Auth)
		r.Get("/api/tasks", api.Tasks)
		r.Get("/api/task", api.Task)
		r.Post("/api/task", api.AddTask)
		r.Put("/api/task", api.EditTask)
		r.Post("/api/task/done", api.Done)
		r.Delete("/api/task", api.DelTask)
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

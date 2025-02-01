package main

import (
	"fmt"
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
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	err = database.Dbinit()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных")
	}
	defer database.DB.Close()
	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Mount("/", fileServer)
	r.Get("/api/tasks*", api.Tasks)
	r.Get("/api/nextdate", api.NextDateHandler)
	r.Post("/api/task", api.AddTask)
	port := os.Getenv("TODO_PORT")

	if len(port) < 1 {
		port = "7540"
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

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
	_, err = database.Dbinit()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных")
	}
	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Mount("/", fileServer)
	r.Get("/api/nextdate", api.NextDateHandler)
	r.Post("/api/tasks", api.AddTask)
	port := os.Getenv("TODO_PORT")

	if len(port) < 1 {
		port = "7540"
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

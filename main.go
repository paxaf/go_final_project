package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/paxaf/go_final_project/api"
	_ "modernc.org/sqlite"
)

func checkDBConnection(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	db, err := dbinit.DbInit()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()
	checkDBConnection(db)
	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Mount("/", fileServer)
	r.Get("/api/nextdate", api.NextDateHandler)
	r.Post("/api/task", api.NewTask)
	port := os.Getenv("TODO_PORT")

	if len(port) < 1 {
		port = "7540"
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

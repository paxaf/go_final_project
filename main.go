package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Не удалось получить текущую рабочую директорию: %v", err)
	}
	database := os.Getenv("TODO_DBFILE")
	if len(database) < 1 {
		database = "scheduler.db"
	}

	dbFile := filepath.Join(workDir, database)
	_, err = os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных: ", err)
	}
	defer db.Close()
	if install {
		sqlBytesFile, err := os.ReadFile("scheduler.sql")
		if err != nil {
			log.Fatal("Ошибка чтения sql файла", err)
		}
		sqlReadFile := string(sqlBytesFile)
		_, err = db.Exec(sqlReadFile)
		if err != nil {
			log.Fatal("Ошибка выполнения SQL запросов: ", err)
		}
	}

	webDir := "./web"
	r := chi.NewRouter()
	fileServer := http.FileServer(http.Dir(webDir))
	r.Mount("/", fileServer)
	port := os.Getenv("TODO_PORT")

	if len(port) < 1 {
		port = "7540"
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

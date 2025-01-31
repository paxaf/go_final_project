package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func Dbinit() (*sql.DB, error) {
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
		sqlBytesFile, err := os.ReadFile("database/scheduler.sql")
		if err != nil {
			log.Fatal("Ошибка чтения sql файла", err)
		}
		sqlReadFile := string(sqlBytesFile)
		_, err = db.Exec(sqlReadFile)
		if err != nil {
			log.Fatal("Ошибка выполнения SQL запросов: ", err)
		}
	}
	return db, nil
}

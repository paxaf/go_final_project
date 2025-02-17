package repository

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func Dbinit() (*TaskRepository, error) {
	repo := &TaskRepository{}
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Не удалось получить текущую рабочую директорию: %v", err)
	}
	database := os.Getenv("TODO_DBFILE")

	if len(database) < 1 {
		database = "scheduler.db"
	}

	dbFile := filepath.Join(workDir, database)
	if err != nil {
		log.Fatalf("Ошибка при создании пути к базе данных: %v", err)
	}
	var install bool
	_, err = os.Stat(dbFile)
	if err != nil {
		install = true
	}
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных: ", err)
	}

	if install {
		filePath := "migration/scheduler.sql"
		sqlFile := filepath.Join(workDir, filePath)

		sqlBytesFile, err := os.ReadFile(sqlFile)
		if err != nil {
			log.Fatal("Ошибка чтения sql файла", err)
		}
		sqlReadFile := string(sqlBytesFile)
		_, err = db.Exec(sqlReadFile)
		if err != nil {
			log.Fatal("Ошибка выполнения SQL запросов: ", err)
		}
	}
	repo.DB = db
	return repo, nil
}

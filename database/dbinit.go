package dbinit

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func DbInit() (*sql.DB, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить текущую рабочую директорию: %w", err)
	}

	database := os.Getenv("TODO_DBFILE")
	if len(database) < 1 {
		database = "scheduler.db"
	}

	dbFile := filepath.Join(workDir, database)
	_, err = os.Stat(dbFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("Ошибка доступа к файлу базы данных: %w", err)
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при подключении к базе данных: %w", err)
	}

	if os.IsNotExist(err) {
		sqlBytesFile, err := os.ReadFile("database/scheduler.sql")
		if err != nil {
			return nil, fmt.Errorf("Ошибка чтения sql файла: %w", err)
		}

		sqlReadFile := string(sqlBytesFile)
		_, err = db.Exec(sqlReadFile)
		if err != nil {
			return nil, fmt.Errorf("Ошибка выполнения SQL запросов: %w", err)
		}
	}

	return db, nil
}

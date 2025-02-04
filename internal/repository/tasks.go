package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/paxaf/go_final_project/internal/models"
)

type TaskRepository struct {
	DB *sql.DB
}

func (r *TaskRepository) Create(task models.Task) (int64, error) {
	var id int64
	err := r.DB.QueryRow(
		`INSERT INTO scheduler (date, title, comment, repeat) 
         VALUES ($1, $2, $3, $4) RETURNING id`,
		task.Date, task.Title, task.Comment, task.Repeat,
	).Scan(&id)
	return id, err
}

func (r *TaskRepository) FindTasks(search string) ([]models.Task, error) {
	var tasks []models.Task
	var rows *sql.Rows

	searchTime, err := time.Parse("02.01.2006", search)
	switch {
	case err == nil:
		searchDate := searchTime.Format(models.DateFormatYYYYMMDD)
		rows, err = r.DB.Query(
			`SELECT id, date, title, comment, repeat 
             FROM scheduler 
             WHERE date = $1 
             ORDER BY date ASC`,
			searchDate)

	case search != "" && err != nil:
		searchText := "%" + search + "%"
		rows, err = r.DB.Query(
			`SELECT id, date, title, comment, repeat 
             FROM scheduler 
             WHERE title LIKE $1 OR comment LIKE $1 
             ORDER BY date ASC`,
			searchText)

	default:
		rows, err = r.DB.Query(
			`SELECT id, date, title, comment, repeat 
             FROM scheduler 
             ORDER BY date ASC`)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка с базой данных: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		var id int

		err := rows.Scan(
			&id,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка инициализации строки: %w", err)
		}

		task.ID = strconv.Itoa(id)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк: %w", err)
	}

	return tasks, nil
}

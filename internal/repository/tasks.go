package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/paxaf/go_final_project/internal/models"
	"github.com/paxaf/go_final_project/internal/service"
)

type TaskRepository struct {
	DB *sql.DB
}

func (r *TaskRepository) Create(task models.Task) (int64, error) {
	var id int64
	err := r.DB.QueryRow(
		`INSERT INTO scheduler (date, title, comment, repeat) 
         VALUES (:date, :title, :comment, :repeat) RETURNING id`,
		sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка вставки: %v", err)
	}
	return id, nil
}

func (r *TaskRepository) SearchTasks(search string) ([]models.Task, error) {
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

func (r *TaskRepository) GetByID(id string) (models.Task, error) {
	idInt, err := strconv.Atoi(id)
	var task models.Task
	var idTask int
	if err != nil {
		return task, err
	}
	err = r.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id;", sql.Named("id", idInt)).Scan(&idTask, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	task.ID = strconv.Itoa(idTask)
	return task, err
}

func (r *TaskRepository) Update(task models.Task) error {
	idInt, err := strconv.Atoi(task.ID)
	if err != nil {
		return err
	}
	res, err := r.DB.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id", sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat), sql.Named("id", idInt))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ни одна строка не была изменена")
	}
	if err != nil {
		return err
	}
	return err
}

func (r *TaskRepository) Done(id string) error {
	var date string
	var repeat string
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	err = r.DB.QueryRow("SELECT date, repeat FROM scheduler WHERE id = :id", sql.Named("id", idInt)).Scan(&date, &repeat)
	if err != nil {
		return err
	}
	if repeat == "" {
		err = r.Delete(id)
		return err
	}
	date, err = service.NextDate(time.Now(), date, repeat)
	if err != nil {
		return err
	}
	res, err := r.DB.Exec("UPDATE scheduler SET date = :date WHERE id = :id", sql.Named("date", date), sql.Named("id", idInt))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ни одна строка не была изменена")
	}
	if err != nil {
		return err
	}
	return err
}

func (r *TaskRepository) Delete(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	res, err := r.DB.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", idInt))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ни одна строка не была изменена")
	}
	return err
}

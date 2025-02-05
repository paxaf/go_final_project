package service

import (
	"errors"
	"strings"
	"time"

	"github.com/paxaf/go_final_project/internal/models"
)

const FormatTime = "20060102"

func Validate(task *models.Task) error {
	if strings.ReplaceAll(task.Title, " ", "") == "" {
		return errors.New("поле 'title' не может быть пустым")
	}

	if strings.ReplaceAll(task.Date, " ", "") == "" {
		task.Date = time.Now().Format(FormatTime)
	} else {
		_, err := time.Parse(FormatTime, task.Date)
		if err != nil {
			return errors.New("ошибка формата времени")
		}
	}
	task.Repeat = strings.TrimSpace(task.Repeat)
	dateRep, err := time.Parse(FormatTime, task.Date)
	if err != nil {
		dateRep = time.Now()
	}
	if task.Repeat != "" && dateRep.Before(time.Now().Truncate(24*time.Hour)) {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return errors.New("ошибка вычисления следующей даты")
		}
		task.Date = nextDate
	} else {
		if dateRep.Before(time.Now()) {
			task.Date = time.Now().Format(FormatTime)
		}
	}
	return nil
}

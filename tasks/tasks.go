package tasks

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      int
	Date    string
	Title   string
	Comment string
	Repeat  string
}

func NextDate(now time.Time, date, repeat string) (string, error) {
	if len(repeat) < 1 {
		//return удаляем из БД
	}
	switch repeat[0] {
	case 'd':
		repeatSplit := strings.Split(repeat, " ")
		if len(repeatSplit) != 2 {
			return "", fmt.Errorf("Параметр задан не правильно")
		}
		days, err := strconv.Atoi(repeatSplit[1])
		if err != nil || days > 400 {
			return "", fmt.Errorf("Недопустимое количество дней: %v", err)
		}
		dateTime, err := time.Parse("20060102", date)
		if err != nil {
			return "", fmt.Errorf("Ошибка преобразования даты: %v", err)
		}
		times := int(now.Sub(dateTime).Hours()/24)/days + 1
		dateTime = dateTime.AddDate(0, 0, days*times)
		return dateTime.Format("20060102"), nil
	case 'y':
		dateTime, err := time.Parse("20060102", date)
		if err != nil {
			return "", fmt.Errorf("Ошибка преобразования даты: %v", err)
		}
		dateTime = dateTime.AddDate(1, 0, 0)
		for dateTime.Before(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
		}
		return dateTime.Format("20060102"), nil
	case 'w':
		weekDay := strings.Split(repeat[2:], ",")
		if len(weekDay) < 1 {
			return "", fmt.Errorf("Аргументы отсутствуют")
		}
		nameDays := map[string]string{
			"1": "Monday",
			"2": "Tuesday",
			"3": "Wednesday",
			"4": "Thursday",
			"5": "Friday",
			"6": "Saturday",
			"7": "Sunday",
		}
		for _, i := range weekDay {
			if _, ok := nameDays[i]; !ok {
				return "", fmt.Errorf("Не допустимые значения в аргументе")
			}
		}
		dateTime, err := time.Parse("20060102", date)
		if err != nil {
			return "", fmt.Errorf("Ошибка преобразования даты: %v", err)
		}
		if dateTime.Before(now) {
			dateTime = now
		}

		currentWeekDay := int(dateTime.Weekday())
		if currentWeekDay == 0 {
			currentWeekDay = 7
		}

		daysDifference := 7
		for _, value := range weekDay {
			targetDay, err := strconv.Atoi(value)
			if err != nil {
				return "", fmt.Errorf("Ошибка преобразования дня недели")
			}
			difference := (targetDay - currentWeekDay + 7) % 7
			if difference == 0 {
				difference = 7
			}
			if difference < daysDifference {
				daysDifference = difference
			}
		}
		return dateTime.AddDate(0, 0, daysDifference).Format("20060102"), nil

	default:
		return "", fmt.Errorf("Неизвестный тип")
	}
}

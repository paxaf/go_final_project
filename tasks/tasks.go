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
	default:
		return "", fmt.Errorf("Неизвестный тип")
	}
}

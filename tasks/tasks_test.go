package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type nextDate struct {
	now    time.Time
	date   string
	repeat string
	want   string
	err    bool // Ожидается ли ошибка
}

func TestNextDate(t *testing.T) {
	now := time.Date(2024, 1, 26, 0, 0, 0, 0, time.UTC) // Текущая дата (26 января 2024)

	// Таблица тестов
	tests := []nextDate{
		{
			now:    now,
			date:   "20240113", // Начальная дата — 23 января 2025
			repeat: "d 7",      // Повтор через каждые 7 дней
			want:   "20240127", // Следующая дата должна быть 06 января 2025
			err:    false,
		},
		{
			now:    now,
			date:   "20240113", // Начальная дата — 23 января 2025
			repeat: "d 10",     // Повтор через каждые 7 дней
			want:   "20240202", // Следующая дата должна быть 06 января 2025
			err:    false,
		},
		{
			now:    now,
			date:   "16890220",
			repeat: "y",
			want:   "20240220",
			err:    false,
		},
		{
			now:    now,
			date:   "20250701",
			repeat: "y",
			want:   "20260701",
			err:    false,
		},
		{
			now:    now,
			date:   "220240126",
			repeat: "k 34",
			want:   "",
			err:    true,
		},
		{
			now:    now,
			date:   "20240125",
			repeat: "w 1,2,3",
			want:   "20240129",
			err:    false,
		},
		{
			now:    now,
			date:   "20240126",
			repeat: "w 7",
			want:   "20240128",
			err:    false,
		},
		{
			now:    now,
			date:   "20240126",
			repeat: "w 8,4,5",
			want:   "",
			err:    true,
		},
	}

	// Выполняем каждый тест
	for _, tt := range tests {
		t.Run(tt.repeat, func(t *testing.T) {
			got, err := NextDate(tt.now, tt.date, tt.repeat)

			// Проверяем возникновение ошибок
			if tt.err {
				assert.Error(t, err, "ожидалась ошибка, но её не было")
			} else {
				assert.NoError(t, err, "ошибка не ожидалась, но она произошла")
			}

			// Сравниваем фактический результат с ожидаемым
			assert.Equal(t, tt.want, got, "результат функции NextDate() некорректен")
		})
	}
}

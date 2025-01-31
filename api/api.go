package api

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/paxaf/go_final_project/api"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Неверный формат времени: %v", err), http.StatusBadRequest)
		return
	}
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	resp, err := api.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, resp)
}

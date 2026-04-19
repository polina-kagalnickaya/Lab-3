package handler

import (
    "encoding/json"
    "net/http"
    "time"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    now := time.Now()
    nextYear := now.Year() + 1
    newYear := time.Date(nextYear, time.January, 1, 0, 0, 0, 0, now.Location())
    daysLeft := int(newYear.Sub(now).Hours() / 24)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{
        "days_before_new_year": daysLeft,
    })
}
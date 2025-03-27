package main

import (
	"encoding/json"
	"net/http"
)

type Match struct {
	ID        int    `json:"id"`
	TeamA     string `json:"team_a"`
	TeamB     string `json:"team_b"`
	ScoreA    int    `json:"score_a"`
	ScoreB    int    `json:"score_b"`
	MatchDate string `json:"match_date"`
}

func main() {
	http.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		matches := []Match{
			{ID: 1, TeamA: "Barça", TeamB: "Madrid", ScoreA: 3, ScoreB: 2, MatchDate: "2024-04-12"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(matches)
	})

	http.ListenAndServe(":8080", nil)
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Match struct {
    ID          int    `json:"id"`
    HomeTeam    string `json:"homeTeam"`
    AwayTeam    string `json:"awayTeam"`
    MatchDate   string `json:"matchDate"`
    Goals       int    `json:"goals"`
    YellowCards int    `json:"yellowCards"`
    RedCards    int    `json:"redCards"`
    ExtraTime   string `json:"extraTime"`
}

var db *sql.DB

func main() {
	var err error
	connStr := fmt.Sprintf("host=db port=5432 user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		fmt.Println("⏳ Esperando a que la base de datos esté lista...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("❌ No se pudo conectar a la base de datos:", err)
	}

	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS matches (
		id SERIAL PRIMARY KEY,
		team_a VARCHAR(100),
		team_b VARCHAR(100),
		match_date DATE,
		goals INT DEFAULT 0,
		yellow_cards INT DEFAULT 0,
		red_cards INT DEFAULT 0,
		extra_time VARCHAR(20)
	)`)

	http.HandleFunc("/api/matches", corsMiddleware(handleMatches))

	http.HandleFunc("/api/matches/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/goals"):
			handleUpdateGoals(w, r)
		case strings.HasSuffix(r.URL.Path, "/yellowcards"):
			handleUpdateYellowCards(w, r)
		case strings.HasSuffix(r.URL.Path, "/redcards"):
			handleUpdateRedCards(w, r)
		case strings.HasSuffix(r.URL.Path, "/extratime"):
			handleUpdateExtraTime(w, r)
		default:
			handleMatchByID(w, r)
		}
	}))

	fmt.Println("✅ Backend corriendo en :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func handleMatches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.Query(`SELECT id, team_a, team_b, match_date, 
			goals, yellow_cards, red_cards, extra_time FROM matches`)
		if err != nil {
			http.Error(w, "DB error", 500)
			return
		}
		defer rows.Close()

		var matches []Match
		for rows.Next() {
			var m Match
			err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate,
				&m.Goals, &m.YellowCards, &m.RedCards, &m.ExtraTime)
			if err != nil {
				http.Error(w, "Error al leer datos", 500)
				return
			}
			matches = append(matches, m)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(matches)

	case "POST":
		var m Match
		json.NewDecoder(r.Body).Decode(&m)
		err := db.QueryRow(`INSERT INTO matches (team_a, team_b, match_date)
			VALUES ($1, $2, $3) RETURNING id`,
			m.HomeTeam, m.AwayTeam, m.MatchDate).Scan(&m.ID)
		if err != nil {
			http.Error(w, "Insert failed", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func handleMatchByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "PUT" && r.Method != "DELETE" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/matches/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", 400)
		return
	}

	switch r.Method {
	case "GET":
		var m Match
		err := db.QueryRow(`SELECT id, team_a, team_b, match_date, 
			goals, yellow_cards, red_cards, extra_time 
			FROM matches WHERE id = $1`, id).
			Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate,
				&m.Goals, &m.YellowCards, &m.RedCards, &m.ExtraTime)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No encontrado", http.StatusNotFound)
			} else {
				http.Error(w, "Error en DB", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)

	case "PUT":
		var m Match
		json.NewDecoder(r.Body).Decode(&m)
		_, err := db.Exec(`UPDATE matches SET team_a=$1, team_b=$2, match_date=$3 WHERE id=$4`,
			m.HomeTeam, m.AwayTeam, m.MatchDate, id)
		if err != nil {
			http.Error(w, "No se pudo actualizar", 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case "DELETE":
		_, err := db.Exec("DELETE FROM matches WHERE id=$1", id)
		if err != nil {
			http.Error(w, "No se pudo eliminar", 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// PATCH /api/matches/:id/goals
func handleUpdateGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id := extractID(r.URL.Path, "/api/matches/", "/goals")
	if id == -1 {
		http.Error(w, "ID inválido", 400)
		return
	}
	_, err := db.Exec(`UPDATE matches SET goals = goals + 1 WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Error al actualizar goles", 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/matches/:id/yellowcards
func handleUpdateYellowCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id := extractID(r.URL.Path, "/api/matches/", "/yellowcards")
	if id == -1 {
		http.Error(w, "ID inválido", 400)
		return
	}
	_, err := db.Exec(`UPDATE matches SET yellow_cards = yellow_cards + 1 WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Error al actualizar tarjetas amarillas", 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/matches/:id/redcards
func handleUpdateRedCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id := extractID(r.URL.Path, "/api/matches/", "/redcards")
	if id == -1 {
		http.Error(w, "ID inválido", 400)
		return
	}
	_, err := db.Exec(`UPDATE matches SET red_cards = red_cards + 1 WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Error al actualizar tarjetas rojas", 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/matches/:id/extratime
func handleUpdateExtraTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id := extractID(r.URL.Path, "/api/matches/", "/extratime")
	if id == -1 {
		http.Error(w, "ID inválido", 400)
		return
	}
	var payload struct {
		ExtraTime string `json:"extraTime"`
	}
	json.NewDecoder(r.Body).Decode(&payload)

	_, err := db.Exec(`UPDATE matches SET extra_time = $1 WHERE id = $2`, payload.ExtraTime, id)
	if err != nil {
		http.Error(w, "Error al actualizar tiempo extra", 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Función auxiliar para extraer ID entre dos rutas
func extractID(path, prefix, suffix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, suffix)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1
	}
	return id
}
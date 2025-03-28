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
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
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

	// Espera hasta que la base de datos esté lista
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

	// Crea tabla si no existe
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS matches (
		id SERIAL PRIMARY KEY,
		team_a VARCHAR(100),
		team_b VARCHAR(100),
		match_date DATE
	)`)

	// Endpoints
	http.HandleFunc("/api/matches", corsMiddleware(handleMatches))
	http.HandleFunc("/api/matches/", corsMiddleware(handleMatchByID))

	fmt.Println("✅ Backend corriendo en :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

// Middleware para manejar CORS
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// GET /api/matches y POST /api/matches
func handleMatches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, team_a, team_b, match_date FROM matches")
		if err != nil {
			http.Error(w, "DB error", 500)
			return
		}
		defer rows.Close()

		var matches []Match
		for rows.Next() {
			var m Match
			_ = rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
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

// GET, PUT, DELETE /api/matches/:id
func handleMatchByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/matches/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", 400)
		return
	}

	switch r.Method {
	case "GET":
		var m Match
		err := db.QueryRow("SELECT id, team_a, team_b, match_date FROM matches WHERE id=$1", id).
			Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
		if err != nil {
			http.Error(w, "No encontrado", 404)
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
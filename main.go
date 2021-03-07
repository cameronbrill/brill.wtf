package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var (
	DB_HOST     string
	DB_PORT     int
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string
	API_PORT    int
	DB          *sql.DB
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func initEnvVars() {
	DB_HOST = getEnv("DB_HOST", "localhost")
	DB_USER = getEnv("DB_USER", "postgres")
	DB_PASSWORD = getEnv("DB_PASSWORD", "postgres")
	DB_NAME = getEnv("DB_NAME", "postgres")
	DB_PORT, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", DB_PORT)
	}
	API_PORT, err := strconv.Atoi(getEnv("API_PORT", "8080"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", API_PORT)
	}
}

func setupDB() (DB *sql.DB) {
	// Connect to postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	DB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("connection to database failed: %v\n", err)
	}

	// Testing postgres connection
	err = DB.Ping()
	if err != nil {
		log.Fatalf("cannot contact database: %v\n", err)
	}

	return DB
}

type ShortURLRequest struct {
	OriginalURL string `json:"original_url"`
	TinyURL     string `json:"tiny_url,omitempty"`
}

type ShortURLResponse struct {
	NormalizedOriginalURL string `json:"normalized_original_url"`
	TinyURL               string `json:"tiny_url"`
}

type ShortURL struct {
	ID                    int       `json:"id"`
	NormalizedOriginalURL string    `json:"normalized_original_url"`
	TinyURL               string    `json:"tiny_url"`
	CreatedAt             time.Time `json:"created_at"`
	LastAccessed          time.Time `json:"last_accessed"`
	TimesAccessed         int       `json:"times_accessed"`
}

func createShortLink(w http.ResponseWriter, r *http.Request) {
	var shortURLReq ShortURLRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&shortURLReq); err != nil {
		log.Fatalf("invalid request: %+v\nerr:%v\n", r, err)
		return
	}
	defer r.Body.Close()

	var id string

	err := DB.QueryRow(
		"INSERT INTO links(url, shorturl) VALUES(?, ?) RETURNING id",
		shortURLReq.OriginalURL, shortURLReq.TinyURL).Scan(&id)

	if err != nil {
		log.Fatalf("failed to upload link to database: %v\n", err)
	}
}

func getURLGivenShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["short_url"]
	var shortURL ShortURLResponse
	err := DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE short_url = ? RETURNING (url, short_url);",
		shortLink).Scan(&shortURL)

	if err != nil {
		log.Fatalf("failed to get url given {short_url: %s} from database: %v\n", shortLink, err)
	}
}

func getShortURLGivenURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := vars["url"]
	var shortURL ShortURLResponse
	err := DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE url = ? RETURNING (url, short_url);",
		url).Scan(&shortURL)

	if err != nil {
		log.Fatalf("failed to get short_url given {url: %s} from database: %v\n", url, err)
	}
}

func main() {
	// Initialize environment variables
	initEnvVars()

	// Setup database
	DB = setupDB()

	// Setup routing
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/short_url", createShortLink).Methods("POST")
	router.HandleFunc("/{short_url}", getURLGivenShortURL).Methods("GET")
	router.HandleFunc("/get_short_url/{url}", getShortURLGivenURL).Methods("GET")

	// Listen on port for API calls
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", API_PORT), nil))
}

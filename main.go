package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
	var err error
	DB_HOST = getEnv("DB_HOST", "localhost")
	DB_USER = getEnv("DB_USER", "postgres")
	DB_PASSWORD = getEnv("DB_PASSWORD", "postgres")
	DB_NAME = getEnv("DB_NAME", "postgres")
	DB_PORT, err = strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", DB_PORT)
	}
	API_PORT, err = strconv.Atoi(getEnv("API_PORT", "8080"))
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
	URL      string `json:"url"`
	ShortURL string `json:"short_url,omitempty"`
}

type ShortURLResponse struct {
	URL      string `json:"url"`
	ShortURL string `json:"short_url"`
}

/*
type ShortURL struct {
	ID                    int       `json:"id"`
	NormalizedOriginalURL string    `json:"normalized_original_url"`
	TinyURL               string    `json:"tiny_url"`
	CreatedAt             time.Time `json:"created_at"`
	LastAccessed          time.Time `json:"last_accessed"`
	TimesAccessed         int       `json:"times_accessed"`
}
*/

func createShortLink(w http.ResponseWriter, r *http.Request) {
	var shortURLReq ShortURLRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&shortURLReq); err != nil {
		log.Fatalf("invalid request: %+v\nerr:%+v\n", r, err)
		return
	}
	log.Infof("body: %+v", r.Body)
	log.Infof("creating short_url: {url: %v}{short_url: %v}", shortURLReq.URL, shortURLReq.ShortURL)
	defer r.Body.Close()

	var id string

	err := DB.QueryRow(
		"INSERT INTO links(url, short_url) VALUES($1, $2) RETURNING id",
		shortURLReq.URL, shortURLReq.ShortURL).Scan(&id)

	if err != nil {
		log.Fatalf("failed to upload link to database: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getURLGivenShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["short_url"]

	log.Infof("getting url given short_url: {short_url: %s}", shortLink)
	var shortURL ShortURLResponse

	// query db
	err := DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE short_url = $1 RETURNING url, short_url;",
		shortLink).Scan(&shortURL.URL, &shortURL.ShortURL)
	if err != nil {
		log.Fatalf("failed to get url given {short_url: %s} from database: %v\n", shortLink, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// build response
	log.Infof("marshalling shortURL into json: %+v", &shortURL)
	response, err := json.Marshal(&shortURL)
	if err != nil {
		log.Fatalf("failed to marshal database response into json: shortURL: %+v\nresponse:%+v\nerr:%v", shortURL, response, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status, err := w.Write(response)
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func getShortURLGivenURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	log.Infof("getting short_url given url: {url: %s}", url)
	var shortURL ShortURLResponse

	err := DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE url = $1 RETURNING url, short_url;",
		url).Scan(&shortURL.URL, &shortURL.ShortURL)
	if err != nil {
		log.Fatalf("failed to get short_url given {url: %s} from database: %v\n", url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// build response
	log.Infof("marshalling shortURL into json: %+v", &shortURL)
	response, err := json.Marshal(&shortURL)
	if err != nil {
		log.Fatalf("failed to marshal database response into json: shortURL: %+v\nresponse:%+v\nerr:%v", shortURL, response, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status, err := w.Write(response)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

}

func main() {
	// Initialize environment variables
	log.Infoln("initializing environment variables...")
	initEnvVars()

	// Setup database
	log.Infoln("setting up database...")
	DB = setupDB()

	// Setup routing
	log.Infoln("setting up routes...")
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/short_url", createShortLink).Methods("POST")
	router.HandleFunc("/get_short_url", getShortURLGivenURL).Methods("GET").Queries("url", "{url}")
	router.HandleFunc("/{short_url}", getURLGivenShortURL).Methods("GET")

	http.Handle("/", router)

	// Listen on port for API calls
	log.Infof("starting api server on port:%d...", API_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", API_PORT), router))
}

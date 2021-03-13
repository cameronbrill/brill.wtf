package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	RDB    *redis.Client
}

var (
	DB_HOST        string
	DB_PORT        int
	DB_USER        string
	DB_PASSWORD    string
	DB_NAME        string
	REDIS_PORT     int
	REDIS_HOST     string
	REDIS_PASSWORD string
	PORT           int
	SSL_MODE       string
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (a *App) initEnvVars() {
	log.Infof("initializing environment variables")
	var err error
	DB_HOST = getEnv("DB_HOST", "127.0.0.1")
	DB_USER = getEnv("DB_USER", "postgres")
	DB_PASSWORD = getEnv("DB_PASSWORD", "postgres")
	DB_NAME = getEnv("DB_NAME", "postgres")
	DB_PORT, err = strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", DB_PORT)
	}
	REDIS_HOST = getEnv("REDIS_HOST", "localhost")
	REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	REDIS_PORT, err = strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", DB_PORT)
	}
	PORT, err = strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		log.Fatalf("port %v cannot be parsed\n", PORT)
	}
	if getEnv("IS_HEROKU", "") != "" {
		SSL_MODE = "require"
	} else {
		SSL_MODE = "disable"
	}
}

func (a *App) setupDB() {
	log.Infof("setting up database")
	var err error
	// Connect to postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=%s", DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, SSL_MODE)
	a.DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("connection to database failed: %v\n", err)
	}

	// Testing postgres connection
	err = a.DB.Ping()
	if err != nil {
		log.Fatalf("cannot contact database: %v\n", err)
	}
}

func (a *App) setupRedisClient() {
	log.Infof("connecting to redis")
	a.RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", REDIS_HOST, REDIS_PORT),
		Password: REDIS_PASSWORD, // no password set
		DB:       0,              // use default DB
	})
}

func (a *App) setupRouter() {
	log.Infof("setting up router")
	a.Router = mux.NewRouter().StrictSlash(true)

	a.Router.HandleFunc("/", a.baseHandler)
	a.Router.HandleFunc("/short_url", a.createShortLink).Methods("POST")
	a.Router.HandleFunc("/get_short_url", a.getShortURLGivenURL).Methods("GET").Queries("url", "{url}")
	a.Router.HandleFunc("/url_info/{short_url}", a.getURLInfoGivenShortURL).Methods("GET")
	a.Router.HandleFunc("/{short_url}", a.getURLGivenShortURL).Methods("GET")

	http.Handle("/", a.Router)
}

func (a *App) SetupApp() {
	a.initEnvVars()
	a.setupDB()
	a.setupRouter()
	a.setupRedisClient()
}

func (a *App) Run() {
	// Listen on port for API calls
	log.Infof("starting api server on port:%d...", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), a.Router))
}

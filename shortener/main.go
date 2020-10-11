package main

import (
	"database/sql"
	"fmt"
	"net/http"

	handler "github.com/cameronbrill/brill.wtf/handler"
	util "github.com/cameronbrill/brill.wtf/internal/util"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to database.
	dbHost := util.GetEnvVar("DB_HOST")
	dbPort := util.GetEnvVar("DB_PORT")
	dbUser := util.GetEnvVar("DB_USER")
	dbPassword := util.GetEnvVar("DB_PASSWORD")
	dbname := util.GetEnvVar("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}

	// handle any url path
	http.Handle("/", handler.MapHandler(pathsToUrls, mux))

	port := util.GetEnvVar("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Starting the server on :" + port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

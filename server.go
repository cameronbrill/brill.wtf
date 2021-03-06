package main

import (
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	host	 string
	port     string
	user     string
	password string
	dbname   string
)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func initEnvVars() {
	host = getEnv("host", "localhost")
	user = getEnv("user", "postgres")
	password = getEnv("password", "postgres")
	dbname = getEnv("links", "postgres")
	port, err := strconv.Atoi(getEnv("port", "5432"))
	if err != nil {
		log.Fatalf("Port %v cannot be parsed", port)
	}
}


func main() {

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

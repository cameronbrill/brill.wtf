package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var shortURLReq ShortURL
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&shortURLReq); err != nil {
		log.Fatalf("invalid request: %+v\nerr:%+v\n", r, err)
		return
	}
	log.Infof("body: %+v", r.Body)
	log.Infof("creating short_url: {url: %v}{short_url: %v}", shortURLReq.URL, shortURLReq.ShortURL)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Fatalf("issue closing request body in createShortLink: %v", err)
		}
	}()

	// insert into postgres
	var id string
	err := a.DB.QueryRow(
		"INSERT INTO links(url, short_url) VALUES($1, $2) RETURNING id",
		shortURLReq.URL, shortURLReq.ShortURL).Scan(&id)

	if err != nil {
		log.Fatalf("failed to upload link to database: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// insert into redis
	a.RDB.Set(a.ctx, shortURLReq.URL, shortURLReq.ShortURL, 0)
}

func (a *App) getURLGivenShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["short_url"]

	// search redis for short_url
	log.Infof("searching redis for short_url: {short_url: %s}", shortLink)
	url := a.RDB.Get(a.ctx, shortLink)
	if url.Err() == nil {
		log.Infof("url found in redis for short_url, redirecting: {url: %s}{short_url: %s}", url.Val(), shortLink)
		http.Redirect(w, r, url.Val(), http.StatusMovedPermanently)
	}

	log.Infof("getting url given short_url: {short_url: %s}", shortLink)
	var shortURL ShortURL

	// query db
	err := a.DB.QueryRow(
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

	// set short_url in redis
	err = a.RDB.Set(a.ctx, shortURL.ShortURL, shortURL.URL, 0).Err()
	if err != nil {
		log.Fatalf("failed to set short_url:url in redis: {short_url: %s} {url: %s}", shortURL.ShortURL, shortURL.URL)
	}

	// redirect user to url
	log.Infof("url found in postgres for short_url, redirecting: {url: %s}{short_url: %s}", shortURL.URL, shortURL.ShortURL)
	http.Redirect(w, r, shortURL.URL, http.StatusMovedPermanently)
}

func (a *App) getURLInfoGivenShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["short_url"]

	log.Infof("getting url given short_url: {short_url: %s}", shortLink)
	var shortURL ShortURL

	// query db
	err := a.DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE short_url = $1 RETURNING url, short_url, created_at, last_accessed, unique_visits;",
		shortLink).Scan(&shortURL.URL, &shortURL.ShortURL, &shortURL.CreatedAt, &shortURL.LastAccessed, &shortURL.UniqueVisits)
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

func (a *App) getShortURLGivenURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	log.Infof("getting short_url given url: {url: %s}", url)
	var shortURL ShortURL

	err := a.DB.QueryRow(
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

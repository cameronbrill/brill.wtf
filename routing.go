package main

import (
	"encoding/json"
	"net/http"

	"github.com/PuerkitoBio/purell"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	log.Infof("entering createShortLink handler...")
	var shortURLReq ShortURL
	var err error
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&shortURLReq); err != nil {
		log.Fatalf("invalid request: %+v\nerr:%+v\n", r, err)
		return
	}
	log.Infof("creating short_url: {url: %v}{short_url: %v}", shortURLReq.URL, shortURLReq.ShortURL)
	//defer func() {
	//	err = r.Body.Close()
	//	if err != nil {
	//		log.Fatalf("issue closing request body in createShortLink: %v", err)
	//	}
	//}()

	// normalize url for insert
	shortURLReq.URL, err = purell.NormalizeURLString(shortURLReq.URL, purell.FlagsUsuallySafeNonGreedy)
	if err != nil {
		log.Fatalf("error normalizing url: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// insert into postgres
	var id string
	log.Infof("inserting url with short_url into postgres: {url: %s}{short_url: %s}", shortURLReq.URL, shortURLReq.ShortURL)
	err = a.DB.QueryRow(
		"INSERT INTO links(url, short_url) VALUES($1, $2) RETURNING id",
		shortURLReq.URL, shortURLReq.ShortURL).Scan(&id)
	if err != nil {
		log.Fatalf("failed to upload link to database: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// insert into redis
	ctx := r.Context()
	log.Infof("setting short_url to url in redis: {short_url: %s} {url: %s}", shortURLReq.ShortURL, shortURLReq.URL)
	err = a.RDB.Set(ctx, shortURLReq.ShortURL, shortURLReq.URL, 0).Err()
	if err != nil {
		log.Fatalf("failed to set short_url:url in redis: {short_url: %s} {url: %s} {err: %v}", shortURLReq.ShortURL, shortURLReq.URL, err)
	}

	// send back success response
	response, err := json.Marshal(shortURLReq)
	if err != nil {
		log.Fatalf("error marshalling response object: {err: %v}", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status, err := w.Write(response)
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func (a *App) getURLGivenShortURL(w http.ResponseWriter, r *http.Request) {
	log.Infof("entering getURLGivenShortURL handler...")
	vars := mux.Vars(r)
	shortLink := vars["short_url"]

	// search redis for short_url
	log.Infof("searching redis for short_url: {short_url: %s}", shortLink)
	ctx := r.Context()
	val, err := a.RDB.Get(ctx, shortLink).Result()
	if err == redis.Nil {
		log.Infof("short_url not found in redis: {short_url: %s}", shortLink)
	} else if err != nil {
		log.Infof("error searching for short_url in redis: {short_url: %s} {error: %v}", shortLink, err)
	} else {
		log.Infof("url found in redis for short_url, redirecting: {url: %s}{short_url: %s}", val, shortLink)
		http.Redirect(w, r, val, http.StatusMovedPermanently)
		return
	}

	var shortURL ShortURL

	// query db
	log.Infof("querying url given short_url from postgres: {short_url: %s}", shortLink)
	err = a.DB.QueryRow(
		"UPDATE links SET unique_visits = unique_visits + 1 WHERE short_url = $1 RETURNING url, short_url;",
		shortLink).Scan(&shortURL.URL, &shortURL.ShortURL)
	if err != nil {
		log.Fatalf("failed to get url given {short_url: %s} from database: %v\n", shortLink, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// build response
	log.Infof("marshalling shortURL into json: %+v", &shortURL)
	response, err := json.Marshal(&shortURL)
	if err != nil {
		log.Fatalf("failed to marshal database response into json: shortURL: %+v\nresponse:%+v\nerr:%v", shortURL, response, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set short_url in redis
	log.Infof("setting short_url to url in redis: {short_url: %s} {url: %s}", shortURL.ShortURL, shortURL.URL)
	err = a.RDB.Set(ctx, shortURL.ShortURL, shortURL.URL, 0).Err()
	if err != nil {
		log.Fatalf("failed to set short_url:url in redis: {short_url: %s} {url: %s} {err: %v}", shortURL.ShortURL, shortURL.URL, err)
	}

	// redirect user to url
	log.Infof("url found in postgres for short_url, redirecting: {url: %s}{short_url: %s}", shortURL.URL, shortURL.ShortURL)
	http.Redirect(w, r, shortURL.URL, http.StatusMovedPermanently)
}

func (a *App) getURLInfoGivenShortURL(w http.ResponseWriter, r *http.Request) {
	log.Infof("entering getURLInfoGivenShortURL handler...")
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
	log.Infof("entering getShortURLGivenURL handler...")
	var err error
	url := r.URL.Query().Get("url")
	// normalize url for insert
	url, err = purell.NormalizeURLString(url, purell.FlagsUsuallySafeNonGreedy)
	if err != nil {
		log.Fatalf("error normalizing url: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	log.Infof("getting short_url given url: {url: %s}", url)
	var shortURL ShortURL

	err = a.DB.QueryRow(
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

func (a *App) baseHandler(w http.ResponseWriter, r *http.Request) {
	// redirect user to my site
	log.Infof("user hit root route, redirecting to https://cameronbrill.me")
	http.Redirect(w, r, "https://cameronbrill.me", http.StatusMovedPermanently)
}

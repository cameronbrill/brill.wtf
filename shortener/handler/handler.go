package shortener

import (
	"net/http"
)

func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.Handler {
	handleFunc := func(res http.ResponseWriter, req *http.Request) {
		url := pathsToUrls[req.URL.Path]
		if url != "" {
			http.Redirect(res, req, url, http.StatusPermanentRedirect)
		} else {
			fallback.ServeHTTP(res, req)
		}
	}
	return http.HandlerFunc(handleFunc)
}

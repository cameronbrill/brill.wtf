package shortener

import (
	"net/http"
)

//type pathToURL struct {
//	Path string
//	URL  string
//}

func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		url := pathsToUrls[req.URL.Path]
		if url != "" {
			http.Redirect(res, req, url, http.StatusPermanentRedirect)
		} else {
			fallback.ServeHTTP(res, req)
		}
	})
}

//func buildMap(pathsToURLs []pathToURL) (builtMap map[string]string) {
//	builtMap = make(map[string]string)
//	for _, ptu := range pathsToURLs {
//		builtMap[ptu.Path] = ptu.URL
//	}
//	return
//}

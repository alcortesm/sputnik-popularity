package web

import "net/http"

func StyleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")
		w.Write([]byte(css))
	})
}

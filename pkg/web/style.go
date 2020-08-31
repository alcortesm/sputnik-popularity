package web

import "net/http"

func StyleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")

		body := []byte(`h1 {
color: #36a8e1;
}`)
		w.Write(body)
	})
}

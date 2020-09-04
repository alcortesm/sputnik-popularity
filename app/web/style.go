package web

import "net/http"

func StyleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")

		body := []byte(`
h1 {
  color: #36a8e1;
  text-align: center;
}

.container {
  width:70vw;
  margin:auto;
  display:block;
}

canvas {
  width:100%;
  height:auto;
}`)

		w.Write(body)
	})
}

package web

import "net/http"

func Popularity() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := []byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Hello world!</h1>
	<p>Foo.</p>

</body>
</html>`)

		w.Header().Set("Content-type", "text/html")
		w.Write(body)
	})
}

func Style() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")

		body := []byte(`h1 {
color: #36a8e1;
}`)
		w.Write(body)
	})
}

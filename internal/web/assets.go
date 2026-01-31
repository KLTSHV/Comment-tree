package web

import (
	"embed"
	"net/http"
)

//go:embed static/index.html static/app.js
var content embed.FS

func UIHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		b, err := content.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "ui not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b)
	})

	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		b, err := content.ReadFile("static/app.js")
		if err != nil {
			http.Error(w, "app.js not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write(b)
	})

	return mux
}

package main

import (
	"log"
	"net/http"

	"github.com/KLTSHV/Comment-tree/internal/httpapi"
	"github.com/KLTSHV/Comment-tree/internal/store"
	"github.com/KLTSHV/Comment-tree/internal/web"
)

func main() {
	repo := store.NewMemoryStore()

	mux := http.NewServeMux()

	// Web UI: /
	mux.Handle("/", web.UIHandler())

	// API: /comments, /comments/{id}, /comments/search
	httpapi.RegisterRoutes(mux, repo)

	addr := ":8080"
	log.Printf("listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, httpapi.WithCORS(mux)))
}

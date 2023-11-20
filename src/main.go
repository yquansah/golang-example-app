package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	mux := chi.NewRouter()

	mux.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Yoofi"))
	})

	server := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

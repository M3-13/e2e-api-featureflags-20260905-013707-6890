package main

import (
	"log"
	"net/http"
	"os"
)

func storeHandler(s *Store, h func(http.ResponseWriter, *http.Request, *Store)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(w, r, s)
	})
}

func main() {
	s := NewStore()

	mux := http.NewServeMux()

	mux.Handle("POST /flags", Logging(storeHandler(s, handleCreateFlag)))
	mux.Handle("GET /flags", Logging(storeHandler(s, handleListFlags)))
	mux.Handle("GET /flags/{key}", Logging(storeHandler(s, handleGetFlag)))
	mux.Handle("PUT /flags/{key}", Logging(storeHandler(s, handleUpdateFlag)))
	mux.Handle("DELETE /flags/{key}", Logging(storeHandler(s, handleDeleteFlag)))
	mux.Handle("GET /flags/{key}/evaluate", Logging(storeHandler(s, handleEvaluateFlag)))
	mux.Handle("GET /healthz", Logging(http.HandlerFunc(handleHealth)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

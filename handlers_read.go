package main

import (
	"encoding/json"
	"net/http"
)

func handleListFlags(w http.ResponseWriter, r *http.Request, s *Store) {
	flags := s.List()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flags)
}

func handleGetFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	key := r.PathValue("key")
	flag, ok := s.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flag)
}

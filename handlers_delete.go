package main

import "net/http"

func handleDeleteFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

package main

import "net/http"

func handleEvaluateFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

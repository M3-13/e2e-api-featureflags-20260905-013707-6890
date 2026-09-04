package main

import "net/http"

func handleDeleteFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	key := r.PathValue("key")
	if !s.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

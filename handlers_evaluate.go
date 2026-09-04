package main

import (
	"encoding/json"
	"net/http"
)

func handleEvaluateFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	key := r.PathValue("key")

	user := r.URL.Query().Get("user")
	if user == "" {
		writeError(w, http.StatusBadRequest, "user query parameter is required")
		return
	}

	flag, ok := s.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	decision := RolloutHash(key, user)%100 < uint64(flag.RolloutPercent)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"decision": decision})
}

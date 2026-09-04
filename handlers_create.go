package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func handleCreateFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	var body Flag
	if err := decodeJSONBody(w, r, &body); err != nil {
		return
	}

	if !keyPattern.MatchString(body.Key) {
		writeError(w, http.StatusBadRequest, "key must match ^[A-Za-z0-9._-]{1,128}$")
		return
	}

	if body.RolloutPercent < 0 || body.RolloutPercent > 100 {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	if err := s.Create(body); err != nil {
		if errors.Is(err, ErrFlagExists) {
			writeError(w, http.StatusConflict, "flag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create flag")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(body)
}

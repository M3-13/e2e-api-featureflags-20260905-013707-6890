package main

import (
	"encoding/json"
	"net/http"
)

type flagUpdate struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

func handleUpdateFlag(w http.ResponseWriter, r *http.Request, s *Store) {
	key := r.PathValue("key")

	var update flagUpdate
	if err := decodeJSONBody(w, r, &update); err != nil {
		return
	}

	existing, ok := s.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	if update.Enabled != nil {
		existing.Enabled = *update.Enabled
	}
	if update.Description != nil {
		existing.Description = *update.Description
	}
	if update.RolloutPercent != nil {
		if *update.RolloutPercent < 0 || *update.RolloutPercent > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		existing.RolloutPercent = *update.RolloutPercent
	}

	updated, _ := s.Update(key, existing)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updated)
}

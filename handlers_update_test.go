package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newUpdateServer(t *testing.T) (*Store, *http.ServeMux) {
	t.Helper()
	s := NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /flags/{key}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateFlag(w, r, s)
	})
	return s, mux
}

func doUpdate(t *testing.T, mux *http.ServeMux, key, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestUpdateFlagPartialPreservesOtherFields(t *testing.T) {
	s, mux := newUpdateServer(t)
	_ = s.Create(Flag{Key: "feature1", Enabled: true, Description: "original", RolloutPercent: 40})

	rec := doUpdate(t, mux, "feature1", `{"description":"new desc"}`, "application/json")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Key != "feature1" {
		t.Fatalf("expected key feature1, got %q", got.Key)
	}
	if !got.Enabled {
		t.Fatal("expected enabled to remain true")
	}
	if got.Description != "new desc" {
		t.Fatalf("expected description updated to %q, got %q", "new desc", got.Description)
	}
	if got.RolloutPercent != 40 {
		t.Fatalf("expected rollout_percent to remain 40, got %d", got.RolloutPercent)
	}
}

func TestUpdateFlagUnknownKeyReturns404(t *testing.T) {
	_, mux := newUpdateServer(t)

	rec := doUpdate(t, mux, "nope", `{"enabled":true}`, "application/json")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateFlagRolloutPercentOutOfRangeReturns400(t *testing.T) {
	s, mux := newUpdateServer(t)
	_ = s.Create(Flag{Key: "feature1", Enabled: true, Description: "d", RolloutPercent: 40})

	for _, v := range []string{"-1", "101"} {
		rec := doUpdate(t, mux, "feature1", `{"rollout_percent":`+v+`}`, "application/json")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent=%s: expected 400, got %d", v, rec.Code)
		}
	}

	got, _ := s.Get("feature1")
	if got.RolloutPercent != 40 {
		t.Fatalf("expected rollout_percent to remain 40 after invalid update, got %d", got.RolloutPercent)
	}
}

func TestUpdateFlagUnknownKeyWrongContentTypeReturns415(t *testing.T) {
	_, mux := newUpdateServer(t)

	rec := doUpdate(t, mux, "nope", `{"enabled":true}`, "text/plain")

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for unknown key with wrong content type, got %d", rec.Code)
	}
}

func TestUpdateFlagUnknownKeyLargeBodyReturns413(t *testing.T) {
	_, mux := newUpdateServer(t)

	large := `{"enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes+1024) + `"}`
	rec := doUpdate(t, mux, "nope", large, "application/json")

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for unknown key with large body, got %d", rec.Code)
	}
}

func TestUpdateFlagWrongContentTypeReturns415(t *testing.T) {
	s, mux := newUpdateServer(t)
	_ = s.Create(Flag{Key: "feature1"})

	rec := doUpdate(t, mux, "feature1", `{"enabled":true}`, "text/plain")

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestUpdateFlagMissingContentTypeReturns415(t *testing.T) {
	s, mux := newUpdateServer(t)
	_ = s.Create(Flag{Key: "feature1"})

	rec := doUpdate(t, mux, "feature1", `{"enabled":true}`, "")

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestUpdateFlagLargeBodyReturns413(t *testing.T) {
	s, mux := newUpdateServer(t)
	_ = s.Create(Flag{Key: "feature1"})

	large := `{"enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes+1024) + `"}`
	rec := doUpdate(t, mux, "feature1", large, "application/json")

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

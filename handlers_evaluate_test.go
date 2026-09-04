package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func evaluateRequest(s *Store, key, user string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/"+key+"/evaluate", nil)
	if user != "" {
		q := req.URL.Query()
		q.Set("user", user)
		req.URL.RawQuery = q.Encode()
	}
	req.SetPathValue("key", key)
	handleEvaluateFlag(rec, req, s)
	return rec
}

func TestEvaluateSameKeyUserSameResult(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature-x", Enabled: true, RolloutPercent: 50})

	var first bool
	for i := 0; i < 20; i++ {
		rec := evaluateRequest(s, "feature-x", "user-42")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var body struct {
			Decision bool `json:"decision"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if i == 0 {
			first = body.Decision
		} else if body.Decision != first {
			t.Fatalf("same key/user returned different decisions")
		}
	}
}

func TestEvaluateRolloutZeroAlwaysFalse(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature-zero", Enabled: true, RolloutPercent: 0})

	rec := evaluateRequest(s, "feature-zero", "any-user")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		Decision bool `json:"decision"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Decision {
		t.Fatal("expected decision false for rollout_percent 0")
	}
}

func TestEvaluateRolloutHundredAlwaysTrue(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature-hundred", Enabled: false, RolloutPercent: 100})

	rec := evaluateRequest(s, "feature-hundred", "any-user")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		Decision bool `json:"decision"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Decision {
		t.Fatal("expected decision true for rollout_percent 100")
	}
}

func TestEvaluateMissingUserReturns400(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature-x", RolloutPercent: 50})

	rec := evaluateRequest(s, "feature-x", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestEvaluateUnknownKeyReturns404(t *testing.T) {
	s := NewStore()

	rec := evaluateRequest(s, "does-not-exist", "user-42")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

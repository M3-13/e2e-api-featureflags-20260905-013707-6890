package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func doCreate(t *testing.T, s *Store, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	handleCreateFlag(rec, req, s)
	return rec
}

func TestCreateFlagValid(t *testing.T) {
	s := NewStore()
	rec := doCreate(t, s, `{"key":"feature1","enabled":true,"description":"desc","rollout_percent":50}`, "application/json")
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if flag.Key != "feature1" || !flag.Enabled || flag.Description != "desc" || flag.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", flag)
	}
	if _, ok := s.Get("feature1"); !ok {
		t.Fatal("expected flag to be stored")
	}
}

func TestCreateFlagDuplicate(t *testing.T) {
	s := NewStore()
	_ = doCreate(t, s, `{"key":"feature1","enabled":true}`, "application/json")
	rec := doCreate(t, s, `{"key":"feature1","enabled":false}`, "application/json")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("expected error field in response")
	}
}

func TestCreateFlagEmptyKey(t *testing.T) {
	s := NewStore()
	rec := doCreate(t, s, `{"key":"","enabled":true}`, "application/json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFlagInvalidKeyFormat(t *testing.T) {
	s := NewStore()
	for _, key := range []string{"has space", "bad/slash", "unicodeä"} {
		rec := doCreate(t, s, `{"key":"`+key+`","enabled":true}`, "application/json")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for key %q, got %d", key, rec.Code)
		}
	}
}

func TestCreateFlagRolloutPercentOutOfRange(t *testing.T) {
	s := NewStore()
	for _, p := range []int{-1, 101} {
		rec := doCreate(t, s, `{"key":"feature1","enabled":true,"rollout_percent":`+strconv.Itoa(p)+`}`, "application/json")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for rollout_percent %d, got %d", p, rec.Code)
		}
	}
}

func TestCreateFlagWrongContentType(t *testing.T) {
	s := NewStore()
	rec := doCreate(t, s, `{"key":"feature1","enabled":true}`, "text/plain")
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestCreateFlagMissingContentType(t *testing.T) {
	s := NewStore()
	rec := doCreate(t, s, `{"key":"feature1","enabled":true}`, "")
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	s := NewStore()
	big := strings.Repeat("x", 2<<20)
	rec := doCreate(t, s, `{"key":"`+big+`","enabled":true}`, "application/json")
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

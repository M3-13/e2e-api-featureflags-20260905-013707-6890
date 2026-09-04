package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleListFlagsEmpty(t *testing.T) {
	s := NewStore()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	handleListFlags(rec, req, s)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "[]\n" {
		t.Fatalf("expected empty JSON array, got %q", body)
	}
}

func TestHandleListFlagsWithFlags(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature1", Enabled: true, Description: "d", RolloutPercent: 50})
	_ = s.Create(Flag{Key: "feature2", Enabled: false, Description: "", RolloutPercent: 0})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	handleListFlags(rec, req, s)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var flags []Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flags); err != nil {
		t.Fatalf("response is not a valid JSON array: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	seen := map[string]Flag{}
	for _, f := range flags {
		seen[f.Key] = f
	}
	if f, ok := seen["feature1"]; !ok || !f.Enabled || f.RolloutPercent != 50 {
		t.Fatalf("feature1 not returned correctly: %+v", f)
	}
	if f, ok := seen["feature2"]; !ok || f.Enabled {
		t.Fatalf("feature2 not returned correctly: %+v", f)
	}
}

func TestHandleGetFlagFound(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature1", Enabled: true, Description: "d", RolloutPercent: 50})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature1", nil)
	req.SetPathValue("key", "feature1")
	handleGetFlag(rec, req, s)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var flag Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &flag); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if flag.Key != "feature1" || !flag.Enabled || flag.Description != "d" || flag.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestHandleGetFlagNotFound(t *testing.T) {
	s := NewStore()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/unknown", nil)
	req.SetPathValue("key", "unknown")
	handleGetFlag(rec, req, s)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error object with message")
	}
}

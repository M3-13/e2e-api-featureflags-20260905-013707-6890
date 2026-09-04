package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "something went wrong")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body["error"] != "something went wrong" {
		t.Fatalf("expected error message, got %q", body["error"])
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
}

func TestDecodeJSONBodyRejectsWrongContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a"}`))
	req.Header.Set("Content-Type", "text/plain")
	var dst map[string]any
	if err := decodeJSONBody(rec, req, &dst); err == nil {
		t.Fatal("expected error")
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestDecodeJSONBodyRejectsMissingContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a"}`))
	var dst map[string]any
	if err := decodeJSONBody(rec, req, &dst); err == nil {
		t.Fatal("expected error")
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestDecodeJSONBodyAcceptsJSONWithCharset(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	var dst map[string]any
	if err := decodeJSONBody(rec, req, &dst); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDecodeJSONBodyRejectsTooLarge(t *testing.T) {
	rec := httptest.NewRecorder()
	big := strings.Repeat("x", 2<<20)
	body := `{"key":"` + big + `"}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	var dst map[string]any
	if err := decodeJSONBody(rec, req, &dst); err == nil {
		t.Fatal("expected error")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestDecodeJSONBodyValid(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"a","enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	var dst map[string]any
	if err := decodeJSONBody(rec, req, &dst); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if dst["key"] != "a" {
		t.Fatalf("expected key a, got %v", dst["key"])
	}
}

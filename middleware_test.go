package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingPassesRequestThroughUnchanged(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method changed: got %q, want POST", r.Method)
		}
		if r.URL.Path != "/flags/my.flag/evaluate" {
			t.Errorf("path changed: got %q", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "alice" {
			t.Errorf("query lost: user=%q", r.URL.Query().Get("user"))
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("POST", "/flags/my.flag/evaluate?user=alice", nil)
	rec := httptest.NewRecorder()

	Logging(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status not passed through: got %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestLoggingRecordsStatusCode(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest("GET", "/flags/missing", nil)
	rec := httptest.NewRecorder()

	Logging(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("unexpected status: got %d, want %d", rec.Code, http.StatusNotFound)
	}

	out := buf.String()
	if !strings.Contains(out, "404") {
		t.Errorf("log does not contain status code 404: %q", out)
	}
}

func TestLoggingDefaultsToStatusOK(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/flags", nil)
	rec := httptest.NewRecorder()

	Logging(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status: got %d, want %d", rec.Code, http.StatusOK)
	}

	out := buf.String()
	if !strings.Contains(out, "200") {
		t.Errorf("log does not contain default status 200: %q", out)
	}
}

func TestLoggingDoesNotLogQueryStringOrUser(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/flags/my.flag/evaluate?user=secretuser", nil)
	rec := httptest.NewRecorder()

	Logging(inner).ServeHTTP(rec, req)

	out := buf.String()
	if strings.Contains(out, "secretuser") {
		t.Errorf("log leaks user value: %q", out)
	}
	if strings.Contains(out, "user=") {
		t.Errorf("log leaks query string: %q", out)
	}
	if strings.Contains(out, "evaluate?") {
		t.Errorf("log leaks query string: %q", out)
	}
}

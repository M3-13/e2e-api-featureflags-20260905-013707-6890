package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleDeleteFlagRemovesFlag(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature1", Enabled: true})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/flags/feature1", nil)
	req.SetPathValue("key", "feature1")

	handleDeleteFlag(rec, req, s)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
	if _, ok := s.Get("feature1"); ok {
		t.Fatal("expected flag to be deleted")
	}
}

func TestHandleDeleteFlagThenLookupGone(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "feature1", Enabled: true})

	del := httptest.NewRecorder()
	delReq := httptest.NewRequest(http.MethodDelete, "/flags/feature1", nil)
	delReq.SetPathValue("key", "feature1")
	handleDeleteFlag(del, delReq, s)
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", del.Code)
	}

	_, ok := s.Get("feature1")
	if ok {
		t.Fatal("expected flag to be gone so a subsequent GET answers 404")
	}
}

func TestHandleDeleteFlagUnknownKey(t *testing.T) {
	s := NewStore()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/flags/nope", nil)
	req.SetPathValue("key", "nope")

	handleDeleteFlag(rec, req, s)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

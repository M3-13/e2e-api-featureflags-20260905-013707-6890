package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newIntegrationServer() (*Store, *http.ServeMux) {
	s := NewStore()
	mux := http.NewServeMux()

	mux.Handle("POST /flags", Logging(storeHandler(s, handleCreateFlag)))
	mux.Handle("GET /flags", Logging(storeHandler(s, handleListFlags)))
	mux.Handle("GET /flags/{key}", Logging(storeHandler(s, handleGetFlag)))
	mux.Handle("PUT /flags/{key}", Logging(storeHandler(s, handleUpdateFlag)))
	mux.Handle("DELETE /flags/{key}", Logging(storeHandler(s, handleDeleteFlag)))
	mux.Handle("GET /flags/{key}/evaluate", Logging(storeHandler(s, handleEvaluateFlag)))
	mux.Handle("GET /healthz", Logging(http.HandlerFunc(handleHealth)))

	return s, mux
}

func TestRoutesAreReachable(t *testing.T) {
	s, mux := newIntegrationServer()
	_ = s.Create(Flag{Key: "some.key", Enabled: true, RolloutPercent: 100})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"create_flag", http.MethodPost, "/flags"},
		{"list_flags", http.MethodGet, "/flags"},
		{"get_flag", http.MethodGet, "/flags/some.key"},
		{"update_flag", http.MethodPut, "/flags/some.key"},
		{"evaluate_flag", http.MethodGet, "/flags/some.key/evaluate?user=u1"},
		{"delete_flag", http.MethodDelete, "/flags/some.key"},
		{"health", http.MethodGet, "/healthz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code == http.StatusNotFound {
				t.Fatalf("route %s %s returned 404: not registered", tt.method, tt.path)
			}
		})
	}
}

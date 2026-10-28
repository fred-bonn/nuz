package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexPage(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   []string
	}{
		{name: "get index", method: http.MethodGet, path: "/", wantStatus: http.StatusOK,
			wantBody: []string{`id="player"`, `id="opponent"`, `type="checkbox"`, `<select`, `Hail`, `Send`}},
		{name: "head index", method: http.MethodHead, path: "/", wantStatus: http.StatusOK},
		{name: "post index", method: http.MethodPost, path: "/", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/nope", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newServer().handler().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			for _, s := range tt.wantBody {
				if !strings.Contains(rec.Body.String(), s) {
					t.Errorf("body missing %q", s)
				}
			}
		})
	}
}

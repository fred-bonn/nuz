package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBattlePage(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
		wantNot    string
	}{
		{name: "with id", method: http.MethodGet, path: "/battle?id=abc123", wantStatus: http.StatusOK, wantBody: "initialized successfully"},
		{name: "id is escaped", method: http.MethodGet, path: "/battle?id=%3Cscript%3E", wantStatus: http.StatusOK, wantBody: "&lt;script&gt;", wantNot: "<script>"},
		{name: "missing id", method: http.MethodGet, path: "/battle", wantStatus: http.StatusSeeOther},
		{name: "wrong method", method: http.MethodPost, path: "/battle?id=abc", wantStatus: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newServer().handler().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q", tt.wantBody)
			}
			if tt.wantNot != "" && strings.Contains(rec.Body.String(), tt.wantNot) {
				t.Errorf("body contains %q", tt.wantNot)
			}
		})
	}
}

func TestIndexPage(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   []string
	}{
		{name: "get index", method: http.MethodGet, path: "/", wantStatus: http.StatusOK,
			wantBody: []string{`id="player"`, `id="opponent"`, `type="checkbox"`, `<select`, `Hail`, `Send`, `id="error"`, `fetch("/battles"`}},
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

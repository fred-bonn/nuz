package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCreateBattle(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		initErr    error
		wantStatus int
		wantType   int
		wantInit   bool
	}{
		{name: "single battle", method: http.MethodPost, body: `{"player":"a","opponent":"b"}`, wantStatus: http.StatusCreated, wantInit: true},
		{name: "double with weather", method: http.MethodPost, body: `{"player":"a","opponent":"b","double":true,"weather":3}`, wantStatus: http.StatusCreated, wantType: 1, wantInit: true},
		{name: "wrong method", method: http.MethodGet, wantStatus: http.StatusMethodNotAllowed},
		{name: "invalid json", method: http.MethodPost, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", method: http.MethodPost, body: `{"player":"a","opponent":"b","x":1}`, wantStatus: http.StatusBadRequest},
		{name: "missing opponent", method: http.MethodPost, body: `{"player":"a"}`, wantStatus: http.StatusBadRequest},
		{name: "blank player", method: http.MethodPost, body: `{"player":"  ","opponent":"b"}`, wantStatus: http.StatusBadRequest},
		{name: "weather too high", method: http.MethodPost, body: `{"player":"a","opponent":"b","weather":5}`, wantStatus: http.StatusBadRequest},
		{name: "negative weather", method: http.MethodPost, body: `{"player":"a","opponent":"b","weather":-1}`, wantStatus: http.StatusBadRequest},
		{name: "engine error", method: http.MethodPost, body: `{"player":"a","opponent":"b"}`, initErr: errors.New("bad party"), wantStatus: http.StatusBadRequest, wantInit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer()
			called, gotType := false, -1
			s.init = func(bt int, p, o string, w int) (battle, error) {
				called, gotType = true, bt
				return &fakeBattle{}, tt.initErr
			}
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, httptest.NewRequest(tt.method, "/battles", strings.NewReader(tt.body)))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if called != tt.wantInit {
				t.Fatalf("init called = %v, want %v", called, tt.wantInit)
			}
			if called && gotType != tt.wantType {
				t.Fatalf("init type = %d, want %d", gotType, tt.wantType)
			}
			if tt.wantStatus != http.StatusCreated {
				if len(s.battles) != 0 {
					t.Fatalf("battle stored on failure")
				}
				return
			}
			var resp createBattleResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if _, ok := s.battles[resp.ID]; !ok || resp.ID == "" {
				t.Fatalf("battle %q not stored", resp.ID)
			}
			if got := rec.Header().Get(battleIDHeader); got != resp.ID {
				t.Fatalf("%s header = %q, want %q", battleIDHeader, got, resp.ID)
			}
		})
	}
}

func TestCreateBattleWithDemoFiles(t *testing.T) {
	t.Chdir("../..")
	player, err := os.ReadFile("showdown_demo_files/player.txt")
	if err != nil {
		t.Fatal(err)
	}
	opponent, err := os.ReadFile("showdown_demo_files/opponent.txt")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		player     string
		opponent   string
		double     bool
		weather    int
		wantStatus int
	}{
		{name: "single", player: string(player), opponent: string(opponent), wantStatus: http.StatusCreated},
		{name: "double with rain", player: string(player), opponent: string(opponent), double: true, weather: 1, wantStatus: http.StatusCreated},
		{name: "invalid player party", player: "not a showdown file", opponent: string(opponent), wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer()
			body, err := json.Marshal(createBattleRequest{Player: tt.player, Opponent: tt.opponent, Double: tt.double, Weather: tt.weather})
			if err != nil {
				t.Fatal(err)
			}
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/battles", bytes.NewReader(body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.wantStatus == http.StatusCreated && len(s.battles) != 1 {
				t.Fatalf("stored battles = %d, want 1", len(s.battles))
			}
		})
	}
}

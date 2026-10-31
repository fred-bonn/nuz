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

// createBattleRequest mirrors the JSON body accepted by POST /battles.
type createBattleRequest struct {
	Player   string `json:"player"`
	Opponent string `json:"opponent"`
	Double   bool   `json:"double"`
	Weather  int    `json:"weather"`
}

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
			if len(resp.State.Player.Active) != 1 || resp.State.Player.Active[0].Name != "pika" || resp.State.Player.Active[0].MaxHP != 45 {
				t.Fatalf("unexpected state in response: %+v", resp.State)
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
		{name: "path traversal in pokemon name", player: "Pikachu/../../x\n- Thunderbolt\n", opponent: string(opponent), wantStatus: http.StatusBadRequest},
		{name: "illegal move name", player: "Pikachu\n- Thunder~bolt\n", opponent: string(opponent), wantStatus: http.StatusBadRequest},
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

func TestSecurityHeadersAndCORS(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		method      string
		wantStatus  int
		wantOrigin  string
		wantExposed bool
	}{
		{name: "allowed origin", origin: "https://app.example", method: http.MethodPost, wantStatus: http.StatusBadRequest, wantOrigin: "https://app.example", wantExposed: true},
		{name: "preflight", origin: "https://app.example", method: http.MethodOptions, wantStatus: http.StatusNoContent, wantOrigin: "https://app.example", wantExposed: true},
		{name: "other origin", origin: "https://evil.example", method: http.MethodPost, wantStatus: http.StatusBadRequest},
		{name: "no origin", method: http.MethodPost, wantStatus: http.StatusBadRequest},
		{name: "preflight from other origin", origin: "https://evil.example", method: http.MethodOptions, wantStatus: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer()
			s.allowedOrigins = []string{"https://app.example"}
			req := httptest.NewRequest(tt.method, "/battles", strings.NewReader("{"))
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Fatalf("allow-origin = %q, want %q", got, tt.wantOrigin)
			}
			if got := rec.Header().Get("Access-Control-Expose-Headers") != ""; got != tt.wantExposed {
				t.Fatalf("expose headers set = %v, want %v", got, tt.wantExposed)
			}
			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Fatal("missing nosniff")
			}
		})
	}
}

func TestMaxBattlesAndIDLength(t *testing.T) {
	s := newServer()
	s.maxBattles = 2
	s.init = func(int, string, string, int) (battle, error) { return &fakeBattle{closed: make(chan struct{})}, nil }

	tests := []struct {
		name       string
		wantStatus int
	}{
		{"first", http.StatusCreated},
		{"second", http.StatusCreated},
		{"over limit", http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/battles", strings.NewReader(`{"player":"a","opponent":"b"}`)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if id := rec.Header().Get(battleIDHeader); tt.wantStatus == http.StatusCreated && len(id) != 32 {
				t.Fatalf("id %q has length %d, want 32", id, len(id))
			}
		})
	}
}

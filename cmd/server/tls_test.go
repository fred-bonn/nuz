package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateTLS(t *testing.T) {
	tests := []struct {
		name      string
		cert, key string
		wantErr   bool
	}{
		{name: "neither", wantErr: false},
		{name: "both", cert: "c.pem", key: "k.pem", wantErr: false},
		{name: "only cert", cert: "c.pem", wantErr: true},
		{name: "only key", key: "k.pem", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTLS(tt.cert, tt.key); (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHSTSHeader(t *testing.T) {
	tests := []struct {
		name string
		tls  bool
		path string
		want string
	}{
		{name: "tls on, page", tls: true, path: "/", want: "max-age=31536000"},
		{name: "tls on, api", tls: true, path: "/battles", want: "max-age=31536000"},
		{name: "tls off, page", tls: false, path: "/", want: ""},
		{name: "tls off, api", tls: false, path: "/battles", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer()
			s.tls = tt.tls
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if got := rec.Header().Get("Strict-Transport-Security"); got != tt.want {
				t.Fatalf("HSTS = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBattleStreamOverTLS(t *testing.T) {
	fb := newFakeBattle(1)
	s := newServer()
	s.tls = true
	s.timeout = time.Minute
	s.init = func(int, string, string, int) (battle, error) { return fb, nil }
	ts := httptest.NewTLSServer(s.handler())
	defer ts.Close()
	client := ts.Client()

	do := func(path, id, body string) *http.Response {
		req, err := http.NewRequest(http.MethodPost, ts.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if id != "" {
			req.Header.Set(battleIDHeader, id)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	create := do("/battles", "", `{"player":"a","opponent":"b"}`)
	create.Body.Close()
	id := create.Header.Get(battleIDHeader)
	if create.StatusCode != http.StatusCreated || id == "" {
		t.Fatalf("create: status %d id %q", create.StatusCode, id)
	}
	if got := create.Header.Get("Strict-Transport-Security"); got == "" {
		t.Fatal("missing HSTS header over TLS")
	}

	stream := do("/battles/start", id, "")
	defer stream.Body.Close()
	sc := bufio.NewScanner(stream.Body)
	// The prompt must arrive while the stream is still open, i.e. it is not buffered until the end.
	if ev := readEvent(t, sc); ev.Type != "prompt" || len(ev.Options) != 2 {
		t.Fatalf("event = %+v", ev)
	}
	in := do("/battles/input", id, fmt.Sprintf(`{"choice":%d}`, 1))
	in.Body.Close()
	if in.StatusCode != http.StatusAccepted {
		t.Fatalf("input status = %d", in.StatusCode)
	}
	if ev := readEvent(t, sc); ev.Type != "end" {
		t.Fatalf("event = %+v", ev)
	}
}

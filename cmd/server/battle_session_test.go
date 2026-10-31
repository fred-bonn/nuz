package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
)

// fakeBattle asks for `prompts` answers and then ends.
type fakeBattle struct {
	prompts   int
	lines     chan string
	closed    chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
	answers   []string
	best      int // 0 means no hint
}

func newFakeBattle(prompts int) *fakeBattle {
	return &fakeBattle{prompts: prompts, lines: make(chan string), closed: make(chan struct{})}
}

func (f *fakeBattle) Run(out io.Writer) error {
	for i := 0; i < f.prompts; i++ {
		fmt.Fprintf(out, "Choose an action:\n1. tackle against rat\n2. switch to raichu (30/30 HP)\n")
		fmt.Fprintf(out, "> ")
		select {
		case line := <-f.lines:
			f.mu.Lock()
			f.answers = append(f.answers, line)
			f.mu.Unlock()
		case <-f.closed:
			return errors.New("closed")
		}
	}
	return nil
}

func (f *fakeBattle) Send(line string) error {
	select {
	case f.lines <- line:
		return nil
	case <-f.closed:
		return errors.New("closed")
	}
}

func (f *fakeBattle) Close()          { f.closeOnce.Do(func() { close(f.closed) }) }
func (f *fakeBattle) PlayerWon() bool { return true }
func (f *fakeBattle) BestOption() (int, bool) {
	return f.best, f.best > 0
}
func (f *fakeBattle) Snapshot() nuzengine.BattleSnapshot {
	return nuzengine.BattleSnapshot{
		Player:   nuzengine.SideSnapshot{Active: []nuzengine.PokemonSnapshot{{Name: "pika", HP: 20, MaxHP: 45}}},
		Opponent: nuzengine.SideSnapshot{Active: []nuzengine.PokemonSnapshot{{Name: "rat", HP: 10, MaxHP: 10}}},
	}
}

func newTestServer(t *testing.T, fb *fakeBattle, timeout time.Duration) (*server, *httptest.Server, string) {
	t.Helper()
	s := newServer(serverConfig{})
	s.timeout = timeout
	s.hintInterval = 10 * time.Millisecond
	s.init = func(int, string, string, int) (battle, error) { return fb, nil }
	ts := httptest.NewServer(s.handler())
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/battles", "application/json", strings.NewReader(`{"player":"a","opponent":"b"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	id := resp.Header.Get(battleIDHeader)
	if resp.StatusCode != http.StatusCreated || id == "" {
		t.Fatalf("create: status %d id %q", resp.StatusCode, id)
	}
	return s, ts, id
}

func post(t *testing.T, url, id, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if id != "" {
		req.Header.Set(battleIDHeader, id)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func readEvent(t *testing.T, sc *bufio.Scanner) event {
	t.Helper()
	if !sc.Scan() {
		t.Fatalf("stream ended: %v", sc.Err())
	}
	var ev event
	if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
		t.Fatal(err)
	}
	return ev
}

func TestBattleFlow(t *testing.T) {
	fb := newFakeBattle(2)
	s, ts, id := newTestServer(t, fb, time.Minute)

	resp := post(t, ts.URL+"/battles/start", id, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start status = %d", resp.StatusCode)
	}
	sc := bufio.NewScanner(resp.Body)

	for i, choice := range []string{"1", "2"} {
		ev := readEvent(t, sc)
		if ev.Type != "prompt" || !strings.Contains(ev.Text, "Choose an action") || strings.HasSuffix(ev.Text, "> ") {
			t.Fatalf("prompt %d = %+v", i, ev)
		}
		if len(ev.Options) != 2 || ev.Options[0] != "tackle against rat" || ev.Options[1] != "switch to raichu (30/30 HP)" {
			t.Fatalf("prompt %d options = %v", i, ev.Options)
		}
		if ev.State == nil || len(ev.State.Player.Active) != 1 || ev.State.Player.Active[0].Name != "pika" {
			t.Fatalf("prompt %d state = %+v", i, ev.State)
		}
		in := post(t, ts.URL+"/battles/input", id, fmt.Sprintf(`{"choice":%s}`, choice))
		in.Body.Close()
		if in.StatusCode != http.StatusAccepted {
			t.Fatalf("input status = %d", in.StatusCode)
		}
	}
	if ev := readEvent(t, sc); ev.Type != "end" || ev.Won == nil || !*ev.Won || ev.State == nil {
		t.Fatalf("end event = %+v", ev)
	}
	if got := strings.Join(fb.answers, ","); got != "1,2" {
		t.Fatalf("answers = %q", got)
	}
	time.Sleep(20 * time.Millisecond)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.battles) != 0 {
		t.Fatal("finished battle was not removed")
	}
}

func TestBattleStartAndInputErrors(t *testing.T) {
	fb := newFakeBattle(1)
	_, ts, id := newTestServer(t, fb, time.Minute)

	tests := []struct {
		name   string
		path   string
		id     string
		body   string
		method string
		want   int
	}{
		{name: "start without id", path: "/battles/start", want: http.StatusBadRequest},
		{name: "start unknown id", path: "/battles/start", id: "nope", want: http.StatusNotFound},
		{name: "start wrong method", path: "/battles/start", id: id, method: http.MethodGet, want: http.StatusMethodNotAllowed},
		{name: "input unknown id", path: "/battles/input", id: "nope", body: `{"choice":1}`, want: http.StatusNotFound},
		{name: "input bad body", path: "/battles/input", id: id, body: `{`, want: http.StatusBadRequest},
		{name: "input before prompt", path: "/battles/input", id: id, body: `{"choice":1}`, want: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			req, _ := http.NewRequest(method, ts.URL+tt.path, strings.NewReader(tt.body))
			if tt.id != "" {
				req.Header.Set(battleIDHeader, tt.id)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}

	resp := post(t, ts.URL+"/battles/start", id, "")
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	readEvent(t, sc)
	dup := post(t, ts.URL+"/battles/start", id, "")
	dup.Body.Close()
	if dup.StatusCode != http.StatusConflict {
		t.Fatalf("second start status = %d", dup.StatusCode)
	}
}

func TestBattleTimeout(t *testing.T) {
	fb := newFakeBattle(1)
	s, ts, id := newTestServer(t, fb, 100*time.Millisecond)

	resp := post(t, ts.URL+"/battles/start", id, "")
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	if ev := readEvent(t, sc); ev.Type != "prompt" {
		t.Fatalf("event = %+v", ev)
	}
	if ev := readEvent(t, sc); ev.Type != "timeout" {
		t.Fatalf("event = %+v", ev)
	}
	select {
	case <-fb.closed:
	case <-time.After(time.Second):
		t.Fatal("battle was not closed after timeout")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.battles) != 0 {
		t.Fatal("timed out battle was not removed")
	}
}

func TestUnstartedBattleExpires(t *testing.T) {
	fb := newFakeBattle(1)
	s, _, _ := newTestServer(t, fb, 50*time.Millisecond)
	select {
	case <-fb.closed:
	case <-time.After(time.Second):
		t.Fatal("unstarted battle was not closed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.battles) != 0 {
		t.Fatal("unstarted battle was not removed")
	}
}

func TestInteractiveBattleWithDemoFiles(t *testing.T) {
	t.Chdir("../..")
	player, err := os.ReadFile("showdown_demo_files/player.txt")
	if err != nil {
		t.Fatal(err)
	}
	opponent, err := os.ReadFile("showdown_demo_files/opponent.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, double := range []bool{false, true} {
		t.Run(fmt.Sprintf("double=%v", double), func(t *testing.T) {
			ts := httptest.NewServer(newServer(serverConfig{}).handler())
			defer ts.Close()
			body, _ := json.Marshal(createBattleRequest{Player: string(player), Opponent: string(opponent), Double: double})
			resp := post(t, ts.URL+"/battles", "", string(body))
			resp.Body.Close()
			id := resp.Header.Get(battleIDHeader)
			if resp.StatusCode != http.StatusCreated || id == "" {
				t.Fatalf("create: status %d id %q", resp.StatusCode, id)
			}

			start := post(t, ts.URL+"/battles/start", id, "")
			defer start.Body.Close()
			sc := bufio.NewScanner(start.Body)
			for prompts := 0; prompts < 10000; prompts++ {
				ev := readEvent(t, sc)
				switch ev.Type {
				case "end":
					return
				case "prompt":
					in := post(t, ts.URL+"/battles/input", id, `{"choice":1}`)
					in.Body.Close()
					if in.StatusCode != http.StatusAccepted {
						t.Fatalf("input status = %d", in.StatusCode)
					}
				default:
					t.Fatalf("unexpected event %+v", ev)
				}
			}
			t.Fatal("battle did not end")
		})
	}
}

func TestParsePromptOptions(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "action prompt", text: "Choose an action:\n1. tackle against rat\n2. switch to raichu (30/30 HP)\n", want: []string{"tackle against rat", "switch to raichu (30/30 HP)"}},
		{name: "replacement prompt", text: "Choose a replacement:\n1. pikachu (10/20 HP)\n", want: []string{"pikachu (10/20 HP)"}},
		{name: "double digit numbering", text: "Choose an action:\n9. a\n10. b\n", want: []string{"a", "b"}},
		{name: "windows line endings", text: "Choose:\r\n1. a\r\n", want: []string{"a"}},
		{name: "error re-prompt has no options", text: "error: invalid input, please enter a number between 1 and 3\n", want: nil},
		{name: "empty", text: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePromptOptions(tt.text)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestBattleHints(t *testing.T) {
	tests := []struct {
		name     string
		best     int
		wantHint bool
	}{
		{name: "best option is streamed while prompt is pending", best: 2, wantHint: true},
		{name: "no hint without learned data", best: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := newFakeBattle(1)
			fb.best = tt.best
			_, ts, id := newTestServer(t, fb, time.Minute)

			resp := post(t, ts.URL+"/battles/start", id, "")
			defer resp.Body.Close()
			sc := bufio.NewScanner(resp.Body)
			if ev := readEvent(t, sc); ev.Type != "prompt" {
				t.Fatalf("first event = %+v", ev)
			}
			if tt.wantHint {
				if ev := readEvent(t, sc); ev.Type != "hint" || ev.Best != tt.best {
					t.Fatalf("hint event = %+v", ev)
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}

			in := post(t, ts.URL+"/battles/input", id, `{"choice":1}`)
			in.Body.Close()
			// Hints already in flight may still arrive before the end of the battle.
			for {
				ev := readEvent(t, sc)
				if ev.Type == "end" {
					return
				}
				if ev.Type != "hint" || !tt.wantHint {
					t.Fatalf("unexpected event %+v", ev)
				}
			}
		})
	}
}

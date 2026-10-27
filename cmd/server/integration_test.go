package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

var optionLine = regexp.MustCompile(`^(\d+)\. (.+)$`)

// TestRandomClientPlaysFullBattle drives the real endpoints, picking randomly among the options the server offers.
func TestRandomClientPlaysFullBattle(t *testing.T) {
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
			ts := httptest.NewServer(newServer().handler())
			defer ts.Close()

			body, _ := json.Marshal(createBattleRequest{Player: string(player), Opponent: string(opponent), Double: double})
			created := post(t, ts.URL+"/battles", "", string(body))
			created.Body.Close()
			id := created.Header.Get(battleIDHeader)
			if created.StatusCode != http.StatusCreated || id == "" {
				t.Fatalf("create: status %d id %q", created.StatusCode, id)
			}

			start := post(t, ts.URL+"/battles/start", id, "")
			defer start.Body.Close()
			if start.StatusCode != http.StatusOK {
				t.Fatalf("start status = %d", start.StatusCode)
			}
			sc := bufio.NewScanner(start.Body)
			sc.Buffer(nil, 1<<20)

			for step := 1; step <= 10000; step++ {
				ev := readEvent(t, sc)
				switch ev.Type {
				case "end":
					fmt.Printf("[double=%v] battle ended after %d choices, player won: %v\n", double, step-1, *ev.Won)
					return
				case "prompt":
					var options []string
					for _, line := range strings.Split(ev.Text, "\n") {
						if m := optionLine.FindStringSubmatch(line); m != nil {
							options = append(options, m[2])
						}
					}
					if len(options) == 0 {
						t.Fatalf("prompt without options: %q", ev.Text)
					}
					choice := rand.Intn(len(options)) + 1
					fmt.Printf("[double=%v] choice %d: picked %d of %d: %s\n", double, step, choice, len(options), options[choice-1])

					in := post(t, ts.URL+"/battles/input", id, fmt.Sprintf(`{"choice":%d}`, choice))
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

type battleResult struct {
	id      string
	choices int
	err     error
}

// playBattle runs a full battle through the endpoints; it avoids t.Fatal so it can run in a goroutine.
func playBattle(baseURL, label string, req createBattleRequest, rng *rand.Rand) battleResult {
	do := func(path, id, body string) (*http.Response, error) {
		r, err := http.NewRequest(http.MethodPost, baseURL+path, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		if id != "" {
			r.Header.Set(battleIDHeader, id)
		}
		return http.DefaultClient.Do(r)
	}

	body, _ := json.Marshal(req)
	created, err := do("/battles", "", string(body))
	if err != nil {
		return battleResult{err: err}
	}
	created.Body.Close()
	id := created.Header.Get(battleIDHeader)
	res := battleResult{id: id}
	if created.StatusCode != http.StatusCreated || id == "" {
		res.err = fmt.Errorf("create: status %d id %q", created.StatusCode, id)
		return res
	}

	start, err := do("/battles/start", id, "")
	if err != nil {
		res.err = err
		return res
	}
	defer start.Body.Close()
	sc := bufio.NewScanner(start.Body)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		var ev event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			res.err = err
			return res
		}
		switch ev.Type {
		case "end":
			return res
		case "prompt":
			options := 0
			for _, line := range strings.Split(ev.Text, "\n") {
				if optionLine.MatchString(line) {
					options++
				}
			}
			if options == 0 {
				res.err = fmt.Errorf("prompt without options: %q", ev.Text)
				return res
			}
			res.choices++
			choice := rng.Intn(options) + 1
			fmt.Printf("[%s] choice %d: picked %d of %d\n", label, res.choices, choice, options)
			in, err := do("/battles/input", id, fmt.Sprintf(`{"choice":%d}`, choice))
			if err != nil {
				res.err = err
				return res
			}
			in.Body.Close()
			if in.StatusCode != http.StatusAccepted {
				res.err = fmt.Errorf("input status = %d", in.StatusCode)
				return res
			}
		default:
			res.err = fmt.Errorf("unexpected event %+v", ev)
			return res
		}
	}
	res.err = fmt.Errorf("stream ended without end event: %v", sc.Err())
	return res
}

// TestConcurrentBattlesAreIndependent plays two battles at once against one server.
func TestConcurrentBattlesAreIndependent(t *testing.T) {
	t.Chdir("../..")
	player, err := os.ReadFile("showdown_demo_files/player.txt")
	if err != nil {
		t.Fatal(err)
	}
	opponent, err := os.ReadFile("showdown_demo_files/opponent.txt")
	if err != nil {
		t.Fatal(err)
	}

	s := newServer()
	ts := httptest.NewServer(s.handler())
	defer ts.Close()

	reqs := []struct {
		label string
		req   createBattleRequest
	}{
		{"A single", createBattleRequest{Player: string(player), Opponent: string(opponent)}},
		{"B double", createBattleRequest{Player: string(player), Opponent: string(opponent), Double: true}},
	}
	results := make([]battleResult, len(reqs))
	var wg sync.WaitGroup
	for i, r := range reqs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = playBattle(ts.URL, r.label, r.req, rand.New(rand.NewSource(int64(i)+1)))
		}()
	}
	wg.Wait()

	for i, res := range results {
		if res.err != nil {
			t.Errorf("battle %s: %v", reqs[i].label, res.err)
		}
		if res.choices == 0 {
			t.Errorf("battle %s: no choices were made", reqs[i].label)
		}
	}
	if results[0].id == results[1].id {
		t.Fatalf("battles share id %q", results[0].id)
	}

	time.Sleep(50 * time.Millisecond)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.battles) != 0 {
		t.Fatalf("%d battles left on the server", len(s.battles))
	}
}

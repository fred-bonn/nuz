package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fred-bonn/nuz/nuzengine"
	"github.com/spf13/pflag"
)

const (
	maxBodyBytes = 1 << 20
	maxWeather   = 4
	runAndBunAi  = 0
)

type createBattleRequest struct {
	Player   string `json:"player"`
	Opponent string `json:"opponent"`
	Double   bool   `json:"double"`
	Weather  int    `json:"weather"`
}

type createBattleResponse struct {
	ID      string `json:"id"`
	Double  bool   `json:"double"`
	Weather int    `json:"weather"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// initFunc builds a battle; it is a field so tests can avoid the PokeAPI.
type initFunc func(battleStateType int, player, opponent string, ai, weather int, policyFile string) (any, error)

type server struct {
	mu      sync.Mutex
	battles map[string]any
	init    initFunc
}

func newServer() *server {
	return &server{
		battles: make(map[string]any),
		init: func(t int, p, o string, ai, w int, pf string) (any, error) {
			return nuzengine.InitBattleState(t, p, o, ai, w, pf)
		},
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/battles", s.handleBattles)
	return mux
}

func (s *server) handleBattles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
		return
	}

	var req createBattleRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid JSON body: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.Player) == "" || strings.TrimSpace(req.Opponent) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"player and opponent are required"})
		return
	}
	if req.Weather < 0 || req.Weather > maxWeather {
		writeJSON(w, http.StatusBadRequest, errorResponse{fmt.Sprintf("weather must be between 0 and %d", maxWeather)})
		return
	}

	battleStateType := 0
	if req.Double {
		battleStateType = 1
	}
	bs, err := s.init(battleStateType, req.Player, req.Opponent, runAndBunAi, req.Weather, "")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}

	id, err := newID()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{"failed generating id"})
		return
	}
	s.mu.Lock()
	s.battles[id] = bs
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, createBattleResponse{ID: id, Double: req.Double, Weather: req.Weather})
}

func newID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)
	addr := fs.String("addr", ":8080", "address to listen on")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return
		}
		log.Fatalf("error: invalid flags: %s", err)
	}

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, newServer().handler()))
}

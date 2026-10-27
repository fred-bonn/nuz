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
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
	"github.com/spf13/pflag"
)

const (
	maxBodyBytes = 1 << 20
	maxWeather   = 4
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
type initFunc func(battleStateType int, player, opponent string, weather int) (battle, error)

type server struct {
	mu      sync.Mutex
	battles map[string]*session
	init    initFunc
	timeout time.Duration
}

func newServer() *server {
	return &server{
		battles: make(map[string]*session),
		timeout: inputTimeout,
		init: func(t int, p, o string, w int) (battle, error) {
			return nuzengine.InitInteractiveBattle(t, p, o, w)
		},
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/battles", s.handleBattleInit)
	mux.HandleFunc("/battles/start", s.handleBattleStart)
	mux.HandleFunc("/battles/input", s.handleBattleInput)
	return mux
}

func (s *server) handleBattleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
		return
	}

	var req createBattleRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, errorResponse{"invalid JSON body: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.Player) == "" || strings.TrimSpace(req.Opponent) == "" {
		respondWithJSON(w, http.StatusBadRequest, errorResponse{"player and opponent are required"})
		return
	}
	if req.Weather < 0 || req.Weather > maxWeather {
		respondWithJSON(w, http.StatusBadRequest, errorResponse{fmt.Sprintf("weather must be between 0 and %d", maxWeather)})
		return
	}

	battleStateType := 0
	if req.Double {
		battleStateType = 1
	}
	bs, err := s.init(battleStateType, req.Player, req.Opponent, req.Weather)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}

	id, err := newID()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, errorResponse{"failed generating id"})
		return
	}
	s.mu.Lock()
	s.battles[id] = newSession(bs)
	s.mu.Unlock()
	time.AfterFunc(s.timeout, func() { s.expireIfUnstarted(id) })

	w.Header().Set(battleIDHeader, id)
	respondWithJSON(w, http.StatusCreated, createBattleResponse{ID: id, Double: req.Double, Weather: req.Weather})
}

func newID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func respondWithJSON(w http.ResponseWriter, status int, v any) {
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

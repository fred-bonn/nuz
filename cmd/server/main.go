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
	maxBodyBytes      = 1 << 20
	maxWeather        = 4
	defaultMaxBattles = 100
)

// initFunc builds a battle; it is a field so tests can avoid the PokeAPI.
type initFunc func(battleStateType int, player, opponent string, weather int) (battle, error)

type server struct {
	mu      sync.Mutex
	battles map[string]*session
	init    initFunc
	timeout time.Duration
	// hintInterval is the period of best-option hints during a pending prompt.
	hintInterval time.Duration
	// allowedOrigins lists browser origins permitted by CORS; empty disables CORS.
	allowedOrigins []string
	maxBattles     int
	// tls reports whether the server is served over HTTPS; it enables HSTS.
	tls bool
}

func newServer() *server {
	return &server{
		battles:      make(map[string]*session),
		timeout:      inputTimeout,
		hintInterval: hintInterval,
		maxBattles:   defaultMaxBattles,
		init: func(t int, p, o string, w int) (battle, error) {
			return nuzengine.InitInteractiveBattle(t, p, o, w)
		},
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /battle", s.handleBattlePage)
	mux.HandleFunc("POST /battles", s.handleBattleInit)
	mux.HandleFunc("POST /battles/start", s.handleBattleStart)
	mux.HandleFunc("POST /battles/input", s.handleBattleInput)
	return s.secure(mux)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "index.html", nil)
}

func (s *server) handleBattlePage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderPage(w, "battle.html", struct{ ID string }{id})
}

func (s *server) handleBattleInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Player   string `json:"player"`
		Opponent string `json:"opponent"`
		Double   bool   `json:"double"`
		Weather  int    `json:"weather"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error(), err)
		return
	}
	if strings.TrimSpace(req.Player) == "" || strings.TrimSpace(req.Opponent) == "" {
		respondWithError(w, http.StatusBadRequest, "player and opponent are required", nil)
		return
	}
	if req.Weather < 0 || req.Weather > maxWeather {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("weather must be between 0 and %d", maxWeather), nil)
		return
	}

	battleStateType := 0
	if req.Double {
		battleStateType = 1
	}
	if s.full() {
		respondWithError(w, http.StatusServiceUnavailable, "too many active battles", nil)
		return
	}
	bs, err := s.init(battleStateType, req.Player, req.Opponent, req.Weather)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	id, err := newID()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed generating id", err)
		return
	}
	s.mu.Lock()
	if len(s.battles) >= s.maxBattles {
		s.mu.Unlock()
		bs.Close()
		respondWithError(w, http.StatusServiceUnavailable, "too many active battles", nil)
		return
	}
	s.battles[id] = newSession(bs)
	s.mu.Unlock()
	time.AfterFunc(s.timeout, func() { s.expireIfUnstarted(id) })

	w.Header().Set(battleIDHeader, id)
	respondWithJSON(w, http.StatusCreated, struct {
		ID      string                   `json:"id"`
		Double  bool                     `json:"double"`
		Weather int                      `json:"weather"`
		State   nuzengine.BattleSnapshot `json:"state"`
	}{id, req.Double, req.Weather, bs.Snapshot()})
}

func (s *server) full() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.battles) >= s.maxBattles
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// validateTLS requires the certificate and key to be given together.
func validateTLS(cert, key string) error {
	if (cert == "") != (key == "") {
		return errors.New("--tls-cert and --tls-key must be set together")
	}
	return nil
}

func main() {
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)
	addr := fs.String("addr", ":8080", "address to listen on")
	origins := fs.StringSlice("allow-origin", nil, "browser origin allowed by CORS (repeatable); CORS is disabled by default")
	maxBattles := fs.Int("max-battles", defaultMaxBattles, "maximum number of concurrent battles")
	tlsCert := fs.String("tls-cert", "", "TLS certificate file; serves HTTPS when set together with --tls-key")
	tlsKey := fs.String("tls-key", "", "TLS private key file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return
		}
		log.Fatalf("error: invalid flags: %s", err)
	}

	if err := validateTLS(*tlsCert, *tlsKey); err != nil {
		log.Fatalf("error: invalid flags: %s", err)
	}

	s := newServer()
	s.tls = *tlsCert != ""
	s.allowedOrigins = *origins
	s.maxBattles = *maxBattles
	// No WriteTimeout: battle responses are long-lived streams.
	srv := &http.Server{
		Addr:              *addr,
		Handler:           s.handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 16,
	}
	if s.tls {
		log.Printf("listening on %s (https)", *addr)
		log.Fatal(srv.ListenAndServeTLS(*tlsCert, *tlsKey))
	}
	log.Printf("listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}

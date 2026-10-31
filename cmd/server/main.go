package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
	"github.com/fred-bonn/nuz/web"
	"github.com/spf13/pflag"
)

const (
	maxBodyBytes      = 1 << 20
	maxWeather        = 4
	defaultMaxBattles = 100
)

type createBattleResponse struct {
	ID      string                   `json:"id"`
	Double  bool                     `json:"double"`
	Weather int                      `json:"weather"`
	State   nuzengine.BattleSnapshot `json:"state"`
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
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/battle", s.handleBattlePage)
	mux.HandleFunc("/battles", s.handleBattleInit)
	mux.HandleFunc("/battles/start", s.handleBattleStart)
	mux.HandleFunc("/battles/input", s.handleBattleInput)
	return s.secure(mux)
}

// secure adds nosniff and CORS for allowed origins, and answers preflight requests.
func (s *server) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if s.tls {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		if origin := r.Header.Get("Origin"); origin != "" && slices.Contains(s.allowedOrigins, origin) {
			h := w.Header()
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Expose-Headers", battleIDHeader)
			if r.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", "POST")
				h.Set("Access-Control-Allow-Headers", "Content-Type, "+battleIDHeader)
				h.Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

var pages = template.Must(template.ParseFS(web.Templates, "templates/*.html"))

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	renderPage(w, "index.html", nil)
}

func (s *server) handleBattlePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderPage(w, "battle.html", struct{ ID string }{id})
}

func renderPage(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := pages.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func (s *server) handleBattleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
		return
	}

	var req struct {
		Player   string `json:"player"`
		Opponent string `json:"opponent"`
		Double   bool   `json:"double"`
		Weather  int    `json:"weather"`
	}
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
	if s.full() {
		respondWithJSON(w, http.StatusServiceUnavailable, errorResponse{"too many active battles"})
		return
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
	if len(s.battles) >= s.maxBattles {
		s.mu.Unlock()
		bs.Close()
		respondWithJSON(w, http.StatusServiceUnavailable, errorResponse{"too many active battles"})
		return
	}
	s.battles[id] = newSession(bs)
	s.mu.Unlock()
	time.AfterFunc(s.timeout, func() { s.expireIfUnstarted(id) })

	w.Header().Set(battleIDHeader, id)
	respondWithJSON(w, http.StatusCreated, createBattleResponse{ID: id, Double: req.Double, Weather: req.Weather, State: bs.Snapshot()})
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

func respondWithJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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

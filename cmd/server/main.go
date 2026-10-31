package main

import (
	"errors"
	"log"
	"net/http"
	"os"
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
	battles *registry
	init    initFunc
	timeout time.Duration
	// hintInterval is the period of best-option hints during a pending prompt.
	hintInterval time.Duration
	// allowedOrigins lists browser origins permitted by CORS; empty disables CORS.
	allowedOrigins []string
	// tls reports whether the server is served over HTTPS; it enables HSTS.
	tls bool
}

// serverConfig holds the flag-driven settings; the zero value is a plain-HTTP server without CORS.
type serverConfig struct {
	tls            bool
	allowedOrigins []string
	// maxBattles defaults to defaultMaxBattles when zero.
	maxBattles int
}

func newServer(cfg serverConfig) *server {
	if cfg.maxBattles == 0 {
		cfg.maxBattles = defaultMaxBattles
	}
	return &server{
		battles:        newRegistry(cfg.maxBattles),
		timeout:        inputTimeout,
		hintInterval:   hintInterval,
		allowedOrigins: cfg.allowedOrigins,
		tls:            cfg.tls,
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

	s := newServer(serverConfig{tls: *tlsCert != "", allowedOrigins: *origins, maxBattles: *maxBattles})
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

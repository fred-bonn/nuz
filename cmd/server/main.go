package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
	"github.com/spf13/pflag"
)

const (
	maxBodyBytes      = 1 << 20
	maxWeather        = 4
	defaultMaxBattles = 100
	// shutdownTimeout bounds how long graceful shutdown waits for in-flight battles.
	shutdownTimeout = 10 * time.Second
)

// initFunc builds a battle; it is a field so tests can avoid the PokeAPI.
type initFunc func(ctx context.Context, battleStateType int, player, opponent string, weather int) (battle, error)

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
		init: func(ctx context.Context, t int, p, o string, w int) (battle, error) {
			return nuzengine.InitInteractiveBattle(ctx, t, p, o, w)
		},
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /battle", s.handleBattlePage)
	mux.HandleFunc("POST /battles", s.handleBattleInit)
	mux.Handle("POST /battles/start", s.withSession(s.handleBattleStart))
	mux.Handle("POST /battles/input", s.withSession(s.handleBattleInput))
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
	// The root context is the base of every request context, so a signal cancels
	// in-flight requests and with them their battles.
	rootCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	// No WriteTimeout: battle responses are long-lived streams.
	srv := &http.Server{
		Addr:              *addr,
		Handler:           s.handler(),
		BaseContext:       func(net.Listener) context.Context { return rootCtx },
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 16,
	}
	if s.tls {
		log.Printf("listening on %s (https)", *addr)
		if err := listenAndServe(rootCtx, srv, func() error { return srv.ListenAndServeTLS(*tlsCert, *tlsKey) }); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Printf("listening on %s", *addr)
	if err := listenAndServe(rootCtx, srv, srv.ListenAndServe); err != nil {
		log.Fatal(err)
	}
}

// listenAndServe runs serve until it fails or ctx is cancelled, then shuts the server
// down gracefully, waiting at most shutdownTimeout for in-flight requests.
func listenAndServe(ctx context.Context, srv *http.Server, serve func() error) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- serve() }()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

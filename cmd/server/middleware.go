package main

import (
	"net/http"
	"slices"
)

type middleware func(http.Handler) http.Handler

// secure applies the security middleware; the first one listed is the outermost.
func (s *server) secure(next http.Handler) http.Handler {
	chain := []middleware{
		nosniff,
		hsts(s.tls),
		cors(s.allowedOrigins),
	}
	for _, m := range slices.Backward(chain) {
		next = m(next)
	}
	return next
}

func nosniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// hsts adds Strict-Transport-Security when the server is served over HTTPS.
func hsts(enabled bool) middleware {
	return func(next http.Handler) http.Handler {
		if !enabled {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
			next.ServeHTTP(w, r)
		})
	}
}

// cors adds CORS headers for allowed origins and answers their preflight requests.
func cors(allowedOrigins []string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" || !slices.Contains(allowedOrigins, origin) {
				next.ServeHTTP(w, r)
				return
			}
			h := w.Header()
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Expose-Headers", battleIDHeader)
			if r.Method != http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			h.Set("Access-Control-Allow-Methods", "POST")
			h.Set("Access-Control-Allow-Headers", "Content-Type, "+battleIDHeader)
			h.Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

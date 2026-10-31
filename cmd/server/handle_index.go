package main

import (
	"net/http"
)

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "index.html", nil)
}

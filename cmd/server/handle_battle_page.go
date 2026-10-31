package main

import (
	"net/http"
)

func (s *server) handleBattlePage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, ok := s.battles.get(id); !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderPage(w, "battle.html", struct{ ID string }{id})
}

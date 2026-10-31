package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
)

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
	if !validBattleRequest(w, req.Player, req.Opponent, req.Weather) {
		return
	}

	battleStateType := 0
	if req.Double {
		battleStateType = 1
	}
	if s.battles.full() {
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
	if !s.battles.add(id, newSession(bs)) {
		bs.Close()
		respondWithError(w, http.StatusServiceUnavailable, "too many active battles", nil)
		return
	}
	time.AfterFunc(s.timeout, func() { s.battles.expireIfUnstarted(id) })

	w.Header().Set(battleIDHeader, id)
	respondWithJSON(w, http.StatusCreated, struct {
		ID      string                   `json:"id"`
		Double  bool                     `json:"double"`
		Weather int                      `json:"weather"`
		State   nuzengine.BattleSnapshot `json:"state"`
	}{id, req.Double, req.Weather, bs.Snapshot()})
}

// validBattleRequest reports whether the fields are acceptable; if not, it has already written the error response.
func validBattleRequest(w http.ResponseWriter, player, opponent string, weather int) bool {
	if strings.TrimSpace(player) == "" || strings.TrimSpace(opponent) == "" {
		respondWithError(w, http.StatusBadRequest, "player and opponent are required", nil)
		return false
	}
	if weather < 0 || weather > maxWeather {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("weather must be between 0 and %d", maxWeather), nil)
		return false
	}
	return true
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

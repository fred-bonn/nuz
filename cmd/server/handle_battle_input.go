package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// handleBattleInput answers the pending prompt; the next prompt arrives on the start stream.
func (s *server) handleBattleInput(w http.ResponseWriter, r *http.Request) {
	_, sess, ok := s.lookup(w, r)
	if !ok {
		return
	}
	var req struct {
		Choice int `json:"choice"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error(), err)
		return
	}
	switch err := sess.claim(req.Choice); {
	case errors.Is(err, errNotWaiting):
		respondWithError(w, http.StatusConflict, err.Error(), nil)
		return
	case errors.Is(err, errBadChoice):
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := sess.battle.Send(strconv.Itoa(req.Choice)); err != nil {
		respondWithError(w, http.StatusGone, "battle is no longer running", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

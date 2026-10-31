package main

import "sync"

// registry owns the active sessions and the lock that guards them.
type registry struct {
	mu       sync.Mutex
	sessions map[string]*session
	max      int
}

func newRegistry(max int) *registry {
	return &registry{sessions: make(map[string]*session), max: max}
}

func (r *registry) full() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sessions) >= r.max
}

func (r *registry) len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sessions)
}

// add stores the session and reports false when the registry is full.
func (r *registry) add(id string, sess *session) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.sessions) >= r.max {
		return false
	}
	r.sessions[id] = sess
	return true
}

func (r *registry) get(id string) (*session, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sess, ok := r.sessions[id]
	return sess, ok
}

func (r *registry) remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
}

// expireIfUnstarted drops a battle the client never started.
func (r *registry) expireIfUnstarted(id string) {
	sess, ok := r.get(id)
	// Claiming started also makes a concurrent start request fail instead of running an expired battle.
	if !ok || !sess.started.CompareAndSwap(false, true) {
		return
	}
	r.remove(id)
	sess.battle.Close()
}

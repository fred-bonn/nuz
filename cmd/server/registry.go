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

// markStarted reports false if the session was already started.
func (r *registry) markStarted(sess *session) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if sess.started {
		return false
	}
	sess.started = true
	return true
}

// expireIfUnstarted drops a battle the client never started.
func (r *registry) expireIfUnstarted(id string) {
	r.mu.Lock()
	sess, ok := r.sessions[id]
	expired := ok && !sess.started
	if expired {
		delete(r.sessions, id)
	}
	r.mu.Unlock()
	if expired {
		sess.battle.Close()
	}
}

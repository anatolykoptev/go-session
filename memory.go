package session

import (
	"sync"
	"time"
)

// Options configures store behavior.
type Options struct {
	TTL         time.Duration // 0 = no expiry
	MaxMessages int           // 0 = unlimited
}

// InMemoryStore is a thread-safe in-memory session store.
type InMemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	opts     Options
}

// NewInMemoryStore creates a new in-memory store.
func NewInMemoryStore(opts Options) *InMemoryStore {
	return &InMemoryStore{
		sessions: make(map[string]*Session),
		opts:     opts,
	}
}

func (m *InMemoryStore) getOrCreate(key string) *Session {
	sess, ok := m.sessions[key]
	if !ok {
		sess = NewSession(key)
		m.sessions[key] = sess
	}
	return sess
}

// GetOrCreate returns an existing session or creates a new one.
func (m *InMemoryStore) GetOrCreate(key string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getOrCreate(key)
}

// AddMessage appends a message, auto-creating the session if needed.
func (m *InMemoryStore) AddMessage(key string, msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess := m.getOrCreate(key)
	sess.AddMessage(msg)

	if m.opts.MaxMessages > 0 && len(sess.Messages) > m.opts.MaxMessages {
		sess.Messages = sess.Messages[len(sess.Messages)-m.opts.MaxMessages:]
	}
}

func (m *InMemoryStore) isExpired(sess *Session) bool {
	if m.opts.TTL <= 0 {
		return false
	}
	return time.Since(sess.Updated) > m.opts.TTL
}

// GetHistory returns a copy of messages, or nil for unknown/expired keys.
func (m *InMemoryStore) GetHistory(key string) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[key]
	if !ok || m.isExpired(sess) {
		return nil
	}

	out := make([]Message, len(sess.Messages))
	copy(out, sess.Messages)
	return out
}

// GetSummary returns the summary, or "" for unknown keys.
func (m *InMemoryStore) GetSummary(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[key]
	if !ok {
		return ""
	}
	return sess.Summary
}

// SetSummary sets the compaction summary.
func (m *InMemoryStore) SetSummary(key, summary string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess := m.getOrCreate(key)
	sess.Summary = summary
	sess.Updated = time.Now()
}

// GetFacts returns a copy of facts.
func (m *InMemoryStore) GetFacts(key string) []Fact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[key]
	if !ok {
		return nil
	}
	return sess.GetFacts()
}

// AddFacts appends facts to a session.
func (m *InMemoryStore) AddFacts(key string, facts []Fact) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess := m.getOrCreate(key)
	sess.AddFacts(facts)
}

// MessageCount returns the number of messages.
func (m *InMemoryStore) MessageCount(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[key]
	if !ok {
		return 0
	}
	return sess.MessageCount()
}

// CompactMessages extracts oldest messages, keeping keepLast.
func (m *InMemoryStore) CompactMessages(key string, keepLast int) []Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[key]
	if !ok {
		return nil
	}
	return sess.CompactMessages(keepLast)
}

// TruncateHistory removes oldest messages, keeping keepLast.
func (m *InMemoryStore) TruncateHistory(key string, keepLast int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[key]
	if !ok {
		return
	}
	sess.TruncateHistory(keepLast)
}

// Clear resets a session's messages, summary, and facts.
func (m *InMemoryStore) Clear(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[key]
	if !ok {
		return
	}
	sess.Clear()
}

// Delete removes a session entirely.
func (m *InMemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, key)
	return nil
}

// Save is a no-op for in-memory store.
func (m *InMemoryStore) Save(_ string) error {
	return nil
}

// ListStale returns keys where Updated is older than maxAge.
func (m *InMemoryStore) ListStale(maxAge time.Duration) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().Add(-maxAge)
	var keys []string
	for k, sess := range m.sessions {
		if sess.Updated.Before(cutoff) {
			keys = append(keys, k)
		}
	}
	return keys
}

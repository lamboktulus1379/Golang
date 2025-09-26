package server

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"
	"time"
)

const (
	oauthSessionCookie      = "oauth_session"
	defaultSessionTTL       = 10 * time.Minute
	stateLength             = 32
	codeVerifierLength      = 64
	codeChallengeMethodS256 = "S256"
)

// LoginSession captures transient state required to complete an OAuth flow.
type LoginSession struct {
	Tenant       string
	Provider     string
	State        string
	CodeVerifier string
	Expires      time.Time
}

// SessionHandle represents the persisted login session metadata returned from Create.
type SessionHandle struct {
	ID        string
	ExpiresAt time.Time
}

// LoginSessionStore persists OAuth login sessions so they can be validated on callback.
type LoginSessionStore interface {
	Create(LoginSession) (SessionHandle, error)
	Consume(id string) (LoginSession, bool)
}

type memorySessionStore struct {
	mu       sync.Mutex
	ttl      time.Duration
	sessions map[string]LoginSession
}

// NewMemorySessionStore builds an in-memory LoginSessionStore suitable for development and tests.
func NewMemorySessionStore(ttl time.Duration) LoginSessionStore {
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	return &memorySessionStore{
		ttl:      ttl,
		sessions: make(map[string]LoginSession),
	}
}

func (s *memorySessionStore) Create(session LoginSession) (SessionHandle, error) {
	id, err := randomToken(32)
	if err != nil {
		return SessionHandle{}, err
	}
	if session.Expires.IsZero() {
		session.Expires = time.Now().Add(s.ttl)
	}

	s.mu.Lock()
	s.cleanupExpired()
	s.sessions[id] = session
	s.mu.Unlock()

	return SessionHandle{ID: id, ExpiresAt: session.Expires}, nil
}

func (s *memorySessionStore) Consume(id string) (LoginSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpired()

	session, ok := s.sessions[id]
	if !ok {
		return LoginSession{}, false
	}

	delete(s.sessions, id)
	if session.Expires.Before(time.Now()) {
		return LoginSession{}, false
	}

	return session, true
}

func (s *memorySessionStore) cleanupExpired() {
	now := time.Now()
	for id, session := range s.sessions {
		if session.Expires.Before(now) {
			delete(s.sessions, id)
		}
	}
}

func generateState() (string, error) {
	return randomToken(stateLength)
}

func generateCodeVerifier() (string, error) {
	return randomToken(codeVerifierLength)
}

func randomToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("random token: invalid length")
	}

	raw := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

package repository

import (
	"errors"
	"sync"

	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
)

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
)

type MemorySessionRepository struct {
	sessions map[string]*models.Session
	mu       sync.RWMutex
}

// Session repository constructor
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{sessions: make(map[string]*models.Session)}
}

// The method of obtaining a session by token; return ErrSessionNotFound if sessions with that token doesn't exist
func (r *MemorySessionRepository) GetByToken(token string) (*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, exists := r.sessions[token]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

// The method of supplementing the session is a lie, here is the session.Token is the key to the card; causes an error ErrSessionAlreadyExists if a session with such a token already exists
func (r *MemorySessionRepository) Create(session *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.sessions[session.Token]; exists {
		return ErrSessionAlreadyExists
	}
	r.sessions[session.Token] = session

	return nil
}

// The method of obtaining a session by userID; return ErrSessionNotFound if sessions with that userID doesn't exist
func (r *MemorySessionRepository) GetByUserID(userID string) (*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.UserID == userID {
			return session, nil
		}
	}

	return nil, ErrSessionNotFound
}

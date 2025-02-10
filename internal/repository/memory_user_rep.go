package repository

import (
	"errors"
	"sync"

	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type MemoryUserRepository struct {
	users map[string]*models.User
	mu    sync.RWMutex
}

// User repository constructor
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{users: make(map[string]*models.User)}
}

// The method of obtaining a user by username; return ErrUserNotFound if user with that username doesn't exist
func (r *MemoryUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// The method of obtaining a user by username; return ErrUserNotFound if user with that userID doesn't exist
func (r *MemoryUserRepository) GetByID(userID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// The method of supplementing the user is a lie, here is the user.ID is the key to the card; causes an error ErrUserAlreadyExists if a user with such a ID already exists
func (r *MemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}
	r.users[user.ID] = user
	return nil
}

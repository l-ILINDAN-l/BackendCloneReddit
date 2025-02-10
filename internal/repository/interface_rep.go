package repository

import (
	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
)

// Implementation under Dependency Injection

// User Service - an interface for working with users
type UserRepository interface {
	GetByUsername(username string) (*models.User, error)
	GetByID(userID string) (*models.User, error)
	Create(user *models.User) error
}

// Session Repository session management interface
type SessionRepository interface {
	GetByToken(token string) (*models.Session, error)
	GetByUserID(userID string) (*models.Session, error)
	Create(session *models.Session) error
}

// PostRepository interface for managing posts
type PostRepository interface {
	GetAll() ([]models.Post, error)
	GetByID(postID string) (*models.Post, error)
	GetByCategory(category string) ([]models.Post, error)
	GetByUserID(userID string) ([]models.Post, error)
	Create(post *models.Post) error
	Delete(postID string) error
	Update(post *models.Post) error
}

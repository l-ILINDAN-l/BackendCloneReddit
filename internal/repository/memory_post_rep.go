package repository

import (
	"errors"
	"sync"

	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
)

// Errors related to processing requests for posts
var (
	ErrPostNotFound      = errors.New("post not found")
	ErrNoPostsInCategory = errors.New("no posts found in this category")
	ErrNoPostsUser       = errors.New("no posts found for this user")
	ErrPostAlreadyExists = errors.New("post already exists")
)

// A structure that stores posts and implements the PostRepository interface
type MemoryPostRepository struct {
	posts map[string]*models.Post
	mu    sync.RWMutex
}

// The constructor of the MemoryPostRepository structure, which returns a reference to the created instance
func NewMemoryPostRepository() *MemoryPostRepository {
	return &MemoryPostRepository{posts: make(map[string]*models.Post)}
}

// The method of getting all the storing posts
func (r *MemoryPostRepository) GetAll() ([]models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	allPosts := make([]models.Post, 0, 100)
	for _, post := range r.posts {
		allPosts = append(allPosts, *post)
	}
	return allPosts, nil
}

// The method of getting a post by its ID, return ErrPostNotFound if post with id equals postID not exists
func (r *MemoryPostRepository) GetByID(postID string) (*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	post, exists := r.posts[postID]
	if !exists {
		return nil, ErrPostNotFound
	}
	return post, nil
}

// The method of obtaining all stored posts corresponding to the category, return ErrNoPostsInCategory  if posts in this category not found
func (r *MemoryPostRepository) GetByCategory(category string) ([]models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var postsByCategory []models.Post
	for _, post := range r.posts {
		if post.Category == category {
			postsByCategory = append(postsByCategory, *post)
		}
	}
	return postsByCategory, nil
}

// The method of getting all stored posts belonging to the user whose ID corresponds to the userID, return ErrNoPostsUser if posts this user not found
func (r *MemoryPostRepository) GetByUserID(userID string) ([]models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userPosts []models.Post
	for _, post := range r.posts {
		if post.Author.ID == userID {
			userPosts = append(userPosts, *post)
		}
	}
	if len(userPosts) == 0 {
		return nil, ErrNoPostsUser
	}
	return userPosts, nil
}

// The method of storing a new post in memory, taking a pointer to a new post, returns Err Post Already Exists if a post with the same ID already exists
func (r *MemoryPostRepository) Create(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.posts[post.ID]; exists {
		return ErrPostAlreadyExists
	}
	r.posts[post.ID] = post
	return nil
}

// The method that deletes a post with an ID equal to post ID returns Err Post Not Found if there is no such post
func (r *MemoryPostRepository) Delete(postID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.posts[postID]; !exists {
		return ErrPostNotFound
	}
	delete(r.posts, postID)
	return nil
}

// The update method of the modified post, which accepts a pointer to the post, returns ErrPostNotFound if there is no such post in the database at the time of the update
func (r *MemoryPostRepository) Update(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.posts[post.ID]; !exists {
		return ErrPostNotFound
	}
	r.posts[post.ID] = post
	return nil
}

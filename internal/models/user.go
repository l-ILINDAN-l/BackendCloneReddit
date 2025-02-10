package models

// Structure for user abstraction and working with JSON
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

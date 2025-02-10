package models

import "time"

// A structure for comment abstraction and working with JSON
type Comment struct {
	ID      string    `json:"id"`
	Author  User      `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
}

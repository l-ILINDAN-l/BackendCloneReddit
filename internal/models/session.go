package models

// A structure for abstracting the user's session and processing it in JSON
type Session struct {
	Token  string `json:"token"`
	UserID string `json:"-"`
}

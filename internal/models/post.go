package models

import "time"

// A structure for post abstraction and working with JSON
type Post struct {
	ID               string    `json:"id"`
	Score            int       `json:"score"`
	Views            int       `json:"views"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	Author           User      `json:"author"`
	Category         string    `json:"category"`
	Text             string    `json:"text,omitempty"`
	URL              string    `json:"url,omitempty"`
	Votes            []Vote    `json:"votes"`
	Comments         []Comment `json:"comments"`
	Created          time.Time `json:"created"`
	UpvotePercentage int       `json:"upvotePercentage"`
}

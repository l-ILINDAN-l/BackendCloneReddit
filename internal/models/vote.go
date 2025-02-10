package models

type Vote struct {
	UserID string `json:"user"`
	Vote   int    `json:"vote"`
}

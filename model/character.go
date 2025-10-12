package model

type Character struct {
	ID      string `json:"id"`
	ClassID string `json:"class_id"`
	// UserID string `json:"user_id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

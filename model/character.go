package model

type Character struct {
	ID      string `json:"id"`
	ClassID string `json:"class_id"`
	Name    string `json:"name"`
	// UserID string `json:"user_id"`
	Level int `json:"level"`
	// Choices []rules.Choice
}

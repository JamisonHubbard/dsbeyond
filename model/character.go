package model

import "github.com/JamisonHubbard/dsbeyond/rules"

type Character struct {
	ID string `json:"id"`
	// UserID  string `json:"user_id"`
	Sheet rules.Sheet `json:"sheet"`
}

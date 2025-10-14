package model

type Sheet struct {
	CharacterID      string          `json:"character_id"`
	ClassID          string          `json:"class_id"`
	Level            int             `json:"level"`
	HeroicResource   string          `json:"heroic_resource"`
	Characteristics  Characteristics `json:"characteristics"`
	Health           Health          `json:"health"`
	Movement         Movement        `json:"movement"`
	Potencies        Potencies       `json:"potencies"`
	Abilities        []string        `json:"abilities"`
	AbilityModifiers []string        `json:"ability_modifiers"`
	Domains          []string        `json:"domains"`
	Kits             []string        `json:"kits"`
	Skills           []string        `json:"skills"`
	Class            map[string]any  `json:"class"`
}

type Characteristics struct {
	Might     int `json:"might"`
	Agility   int `json:"agility"`
	Reason    int `json:"reason"`
	Intuition int `json:"intuition"`
	Presence  int `json:"presence"`
}

type Health struct {
	MaxStamina    int `json:"max_stamina"`
	MaxRecoveries int `json:"max_recoveries"`
}

type Movement struct {
	Size      Size `json:"size"`
	Speed     int  `json:"speed"`
	Stability int  `json:"stability"`
	Disengage int  `json:"disengage"`
}

const (
	SizeTypeSmall  = "small"
	SizeTypeMedium = "medium"
	SizeTypeLarge  = "large"
)

type Size struct {
	Space int    `json:"space"`
	Type  string `json:"type"`
}

type Potencies struct {
	Strong  int `json:"strong"`
	Average int `json:"average"`
	Weak    int `json:"weak"`
}

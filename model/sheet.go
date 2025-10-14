package model

type Sheet struct {
	CharacterID      string          `json:"character_id"`
	ClassID          string          `json:"class_id"`
	Level            int             `json:"level"`
	HeroicResource   string          `json:"heroic_resource"`
	Characteristics  Characteristics `json:"characteristics"`
	Health           Health          `json:"health"`
	Potencies        Potencies       `json:"potencies"`
	Skills           []string        `json:"skills"`
	Abilities        []string        `json:"abilities"`
	AbilityModifiers []string        `json:"ability_modifiers"`
	Class            map[string]any  `json:"class"`
	Domains          []string        `json:"domains"`
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

type Potencies struct {
	Strong  int `json:"strong"`
	Average int `json:"average"`
	Weak    int `json:"weak"`
}

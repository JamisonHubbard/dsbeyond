// Package rules contains the rules engine
package rules

type Sheet struct {
	ClassID         string          `json:"class_id"`
	Level           int             `json:"level"`
	Characteristics Characteristics `json:"characteristics"`
	MaxStamina      int             `json:"max_stamina"`
	Recoveries      int             `json:"recoveries"`
	Potencies       Potencies       `json:"potencies"`
}

type Characteristics struct {
	Might     int `json:"might"`
	Agility   int `json:"agility"`
	Reason    int `json:"reason"`
	Intuition int `json:"intuition"`
	Presence  int `json:"presence"`
}

type Potencies struct {
	Strong  int `json:"strong"`
	Average int `json:"average"`
	Weak    int `json:"weak"`
}

type ValueRef struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Operation struct {
	Type     string   `json:"type"`
	Target   string   `json:"target"`
	ValueRef ValueRef `json:"value_ref"`
}

type Expression struct {
	Type string     `json:"type"`
	Args []ValueRef `json:"args"`
}

type Class struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Basics ClassLevel         `json:"basics"`
	Levels map[int]ClassLevel `json:"levels"`
}

type ClassLevel struct {
	Operations []Operation `json:"operations"`
	// ChoiceDefns []ChoiceDefinition `json:"choice_defns"`
}

type ParsedClass struct {
	Operations []Operation `json:"operations"`
	// ChoiceDefns []ChoiceDefinition `json:"choice_defns"`
}

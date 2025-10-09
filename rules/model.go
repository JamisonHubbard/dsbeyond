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

type Modifier struct {
	Type           string    `json:"type"`
	Target         string    `json:"target"`
	ValueType      string    `json:"value_type"`
	IntValue       int       `json:"int_value"`
	StringValue    string    `json:"string_value"`
	ArrayValue     []int     `json:"array_value"`
	OperationValue Operation `json:"operation_value"`
	IDValue        string    `json:"id_value"`
}

type Operation struct {
	Type       string `json:"type"`
	FirstType  string `json:"first_type"`
	First      string `json:"first"`
	SecondType string `json:"second_type"`
	Second     string `json:"second"`
}

type ChoiceDefinition struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Target     string   `json:"target"`
	Targets    []string `json:"targets"`
	OptionType string   `json:"option_type"`
	Options    []Option `json:"options"`
}

type Option struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	IntOption    int      `json:"int_option"`
	StringOption string   `json:"string_option"`
	ArrayOption  []Option `json:"array_option"`
}

type Choice struct {
	ID       string `json:"id"`
	OptionID string `json:"option_id"`
}

type ClassBasics struct {
	Modifiers         []Modifier         `json:"modifiers"`
	ChoiceDefinitions []ChoiceDefinition `json:"choice_definitions"`
}

type Class struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Basics ClassBasics `json:"basics"`
}

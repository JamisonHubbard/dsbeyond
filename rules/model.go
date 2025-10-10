// Package rules contains the rules engine
package rules

import (
	"encoding/json"
	"fmt"
)

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

func (v *ValueRef) UnmarshalJSON(data []byte) error {
	// define an initial lightweight struct for initial decode
	type rawValueRef struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}

	var tmp rawValueRef
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	v.Type = tmp.Type

	switch tmp.Type {
	case ValueRefTypeInt:
		var i int
		if err := json.Unmarshal(tmp.Value, &i); err != nil {
			return err
		}
		v.Value = i
	case ValueRefTypeID:
		var s string
		if err := json.Unmarshal(tmp.Value, &s); err != nil {
			return err
		}
		v.Value = s
	case ValueRefTypeExpr:
		var expr Expression
		if err := json.Unmarshal(tmp.Value, &expr); err != nil {
			return err
		}
		v.Value = &expr
	default:
		return fmt.Errorf("invalid ValueRef type: %s", v.Type)
	}
	return nil
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	type rawOperation struct {
		Type     string          `json:"type"`
		Target   string          `json:"target"`
		ValueRef json.RawMessage `json:"value_ref"`
	}

	var tmp rawOperation
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	o.Type = tmp.Type
	o.Target = tmp.Target
	if err := json.Unmarshal(tmp.ValueRef, &o.ValueRef); err != nil {
		return fmt.Errorf("unmarshalling value_ref: %w", err)
	}

	return nil
}

// Package rules contains the rules engine
package rules

import (
	"encoding/json"
	"fmt"
)

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

type Choice struct {
	ID      string      `json:"id"`
	Prereqs []Assertion `json:"prereqs"`
	Options []Option    `json:"options"`
}

type Assertion struct {
	ID       string     `json:"id"`
	TargetID string     `json:"target_id"`
	Values   []ValueRef `json:"values"`
}

type Option struct {
	ID         string      `json:"id"`
	Operations []Operation `json:"operations"`
}

type Decision struct {
	ChoiceID string `json:"choice_id"`
	OptionID string `json:"option_id"`
}

type Class struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Basics ClassLevel         `json:"basics"`
	Levels map[int]ClassLevel `json:"levels"`
}

type ClassLevel struct {
	Operations []Operation `json:"operations"`
	Choices    []Choice    `json:"choices"`
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

// Package rules contains the rules engine
package rules

import (
	"encoding/json"
	"fmt"
)

type Reference struct {
	Abilities   map[string]Ability
	Classes     map[string]Class
	Domains     map[string]Domain
	Skills      map[string]Skill
	SkillGroups map[string]SkillGroup
}

type ValueRef struct {
	Type    string `json:"type"`
	Value   any    `json:"value"`
	RefType string `json:"ref_type"`
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
	Type    string      `json:"type"`
	Prereqs []Assertion `json:"prereqs"`
	Options []Option    `json:"options"`
	RefType string      `json:"ref_type"`
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
	ChoiceID string   `json:"choice_id"`
	Type     string   `json:"type"`
	OptionID string   `json:"option_id"`
	RefID    string   `json:"ref_id"`
	Target   string   `json:"target"`
	Value    ValueRef `json:"value"`
}

type Hook struct {
	ID         string      `json:"id"`
	Event      string      `json:"event"`
	Operations []Operation `json:"operations"`
}

func (v *ValueRef) UnmarshalJSON(data []byte) error {
	// define an initial lightweight struct for initial decode
	type rawValueRef struct {
		Type    string          `json:"type"`
		Value   json.RawMessage `json:"value"`
		RefType string          `json:"ref_type"`
	}

	var tmp rawValueRef
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	v.Type = tmp.Type
	v.RefType = tmp.RefType

	switch tmp.Type {
	case ValueRefTypeInt:
		var i int
		if err := json.Unmarshal(tmp.Value, &i); err != nil {
			return err
		}
		v.Value = i
	case ValueRefTypeString:
		var s string
		if err := json.Unmarshal(tmp.Value, &s); err != nil {
			return err
		}
		v.Value = s
	case ValueRefTypeID:
		var s string
		if err := json.Unmarshal(tmp.Value, &s); err != nil {
			return err
		}
		v.Value = s
	case ValueRefTypeRefID:
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

// Package rules contains the rules engine
package rules

import (
	"encoding/json"
	"fmt"
)

const (
	RefIDTypeAbility         = "ability"
	RefIDTypeAbilityModifier = "ability_modifier"
	RefIDTypeDomain          = "domain"
	RefIDTypeSkill           = "skill"
	RefIDTypeSkillGroup      = "skill_group"
)

// A Reference contains all the static rules data for the game
type Reference struct {
	Abilities   map[string]Ability
	Classes     map[string]Class
	Domains     map[string]Domain
	Skills      map[string]Skill
	SkillGroups map[string]SkillGroup
}

const (
	ValueRefTypeExpression = "expression"
	ValueRefTypeID         = "id"
	ValueRefTypeInt        = "int"
	ValueRefTypeRefID      = "refid"
	ValueRefTypeString     = "string"
)

// A ValueRef is a reference to a value
// These values can be resolved at runtime
type ValueRef struct {
	Type      string `json:"type"`
	Value     any    `json:"value"`
	RefIDType string `json:"refid_type"`
}

const (
	OperationTypeSet           = "set"
	OperationTypeAddSkill      = "add_skill"
	OperationTypeAddAbility    = "add_ability"
	OperationTypeModifyAbility = "modify_ability"
)

// An Operation is an action taken to set or modify a value in the context
type Operation struct {
	Type     string      `json:"type"`
	Target   string      `json:"target"`
	ValueRef ValueRef    `json:"value_ref"`
	Prereqs  []Assertion `json:"prereqs"`
}

// An Assertion is a condition that is checked at runtime
type Assertion struct {
	TargetID string     `json:"target_id"`
	Values   []ValueRef `json:"values"`
}

const (
	ExprTypeAdd      = "add"
	ExprTypeSubtract = "subtract"
)

// An Expression is a mathematical statement that is evaluated at runtime to
// produce a result
type Expression struct {
	Type string     `json:"type"`
	Args []ValueRef `json:"args"`
}

const (
	ChoiceTypeOptionSelect = "option_select"
	ChoiceTypeRefSelect    = "ref_select"
	ChoiceTypeInput        = "input"
)

// A Choice represents a decision point during character creation that impacts
// the final character sheet
type Choice struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Prereqs []Assertion `json:"prereqs"`
	Options []Option    `json:"options"`
	RefType string      `json:"ref_type"`
}

// An Option is a possible decision made to resolve a Choice
type Option struct {
	ID         string      `json:"id"`
	Operations []Operation `json:"operations"`
}

const (
	DecisionTypeID        = "id"
	DecisionTypeRefID     = "refid"
	DecisionTypeOperation = "operation"
	DecisionTypeValue     = "value"
)

// A Decision represents the result of a Choice that was made during character
// creation
type Decision struct {
	ChoiceID string   `json:"choice_id"`
	Type     string   `json:"type"`
	OptionID string   `json:"option_id"`
	RefID    string   `json:"ref_id"`
	Target   string   `json:"target"`
	Value    ValueRef `json:"value"`
}

// UnmarshalJSON is a custom unmarshaller for ValueRef
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
	v.RefIDType = tmp.RefType

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
	case ValueRefTypeExpression:
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

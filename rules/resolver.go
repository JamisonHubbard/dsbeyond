package rules

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/JamisonHubbard/dsbeyond/model"
)

func NewResolver(character model.Character, decisions map[string]Decision, reference *Reference) *Resolver {
	return &Resolver{
		character: character,
		decisions: decisions,
		reference: reference,

		// internals
		values:     make(map[string]any),
		operations: make(map[string][]*Operation),
		visited:    make(map[string]bool),
		completed:  make(map[string]bool),
		trace:      Trace{},
		error:      nil,
	}
}

const (
	AbilitiesValueName        = "abilities"
	AbilityModifiersValueName = "ability_modifiers"
	DomainsValueName          = "domains"
	SkillsValueName           = "skills"
)

type Resolver struct {
	// inputs
	character model.Character
	decisions map[string]Decision
	reference *Reference

	// internals
	values     map[string]any
	operations map[string][]*Operation
	visited    map[string]bool
	completed  map[string]bool
	trace      Trace
	error      error
}

func (r *Resolver) Resolve() (model.Sheet, error) {
	// get class data from reference
	class, ok := r.reference.Classes[r.character.ClassID]
	if !ok {
		return model.Sheet{}, fmt.Errorf("class \"%s\" not found", r.character.ClassID)
	}

	// setup values and operations
	r.setup(&class)
	if r.error != nil {
		return model.Sheet{}, r.error
	}

	// execute operations
	for node := range r.operations {
		r.trace.Push("node:" + node)
		r.EvaluateNode(node)
		if r.error != nil {
			return model.Sheet{}, r.error
		}
		r.trace.Pop()
	}

	// pretty print context values
	prettyContext, err := json.MarshalIndent(r.values, "", "  ")
	if err != nil {
		log.Println("WARNING failed to pretty print context")
	} else {
		log.Println(string(prettyContext))
	}

	skills, ok := r.values["skills"]
	if !ok {
		skills = []string{}
	}

	// create sheet
	sheet := model.Sheet{
		ClassID: r.character.ClassID,
		Level:   r.character.Level,
		Core: model.SheetCore{
			HeroicResource: expectString("heroic_resource", r.values["heroic_resource"]),
			Characteristics: model.Characteristics{
				Might:     expectInt("characteristics.might", r.values["characteristics.might"]),
				Agility:   expectInt("characteristics.agility", r.values["characteristics.agility"]),
				Reason:    expectInt("characteristics.reason", r.values["characteristics.reason"]),
				Intuition: expectInt("characteristics.intuition", r.values["characteristics.intuition"]),
				Presence:  expectInt("characteristics.presence", r.values["characteristics.presence"]),
			},
			Health: model.Health{
				MaxStamina:    expectInt("health.max_stamina", r.values["health.max_stamina"]),
				MaxRecoveries: expectInt("health.max_recoveries", r.values["health.max_recoveries"]),
			},
			Potencies: model.Potencies{
				Strong:  expectInt("potencies.strong", r.values["potencies.strong"]),
				Average: expectInt("potencies.average", r.values["potencies.average"]),
				Weak:    expectInt("potencies.weak", r.values["potencies.weak"]),
			},
		},
		Skills: skills.([]string),
	}

	return sheet, nil
}

// setup parses the class and decisions to generate the Operations that must be
// evaluated to resolve the character sheet
func (r *Resolver) setup(class *Class) {
	var operations []Operation
	var choices []Choice

	// collect non-decision operations from the class
	operations = append(operations, class.Basics.Operations...)
	for level, levelDefinition := range class.Levels {
		if level <= r.character.Level {
			operations = append(operations, levelDefinition.Operations...)
		}
	}

	// collect choices from the class
	choices = append(choices, class.Basics.Choices...)
	for level, levelDefinition := range class.Levels {
		if level <= r.character.Level {
			choices = append(choices, levelDefinition.Choices...)
		}
	}

	// for each choice, use decicions to resolve the choice into a set of
	// operations
	// note: this doesn't mean the operations are executed. Assertions from the
	// choice are applied to the operations meaning some will be filtered out at
	// runtime
	for _, choice := range choices {
		choiceOperations := r.reduceChoice(&choice)
		if r.error != nil {
			return
		}
		if choiceOperations != nil {
			operations = append(operations, choiceOperations...)
		}
	}

	// add operations
	for _, operation := range operations {
		r.operations[operation.Target] = append(r.operations[operation.Target], &operation)
	}
}

// reduceChoice converts a Choice into a set of Operations
func (r *Resolver) reduceChoice(choice *Choice) []Operation {
	var operations []Operation

	decision, ok := r.decisions[choice.ID]
	if !ok {
		log.Printf("no decision for choice %s\n", choice.ID)
		return nil
	}

	switch choice.Type {
	case ChoiceTypeOptionSelect:
		// get the selected option
		var option *Option
		for _, o := range choice.Options {
			if o.ID == decision.OptionID {
				option = &o
				break
			}
		}
		if option == nil {
			r.error = fmt.Errorf("option \"%s\" for choice \"%s\" not found", decision.OptionID, choice.ID)
			return nil
		}

		// apply the choice prereqs to the option operations
		for _, operation := range option.Operations {
			operation.Prereqs = append(operation.Prereqs, choice.Prereqs...)
			operations = append(operations, operation)
		}
	case ChoiceTypeRefSelect:
		// reduce the referenced value into an operation
		operation := r.reduceRefID(decision.RefID, choice.RefType)
		if r.error != nil {
			return nil
		}
		operations = append(operations, operation)
	case ChoiceTypeInput:
		if decision.Type != DecisionTypeValue {
			r.error = fmt.Errorf("invalid decision type for choice \"%s\", decision type must be \"%s\" for choices with type \"%s\"", choice.ID, decision.Type, choice.Type)
			return nil
		}

		operations = append(operations, Operation{
			Type:     OperationTypeSet,
			Target:   decision.Target,
			ValueRef: decision.Value,
		})
	default:
		r.error = fmt.Errorf("unknown choice type: %s", choice.Type)
		return nil
	}

	return operations
}

// reduceRefID resolves a reference ID into an operation to add that referenced
// value to the sheet
// NOTE: this excludes the "skill group" ref id since those are never added to
// a character sheet
func (r *Resolver) reduceRefID(refID string, refIDType string) Operation {
	switch refIDType {
	case RefIDTypeAbility:
		_, ok := r.reference.Abilities[refID]
		if !ok {
			r.error = fmt.Errorf("ability \"%s\" not found", refID)
			return Operation{}
		}
		return Operation{
			Type:     OperationTypeAddAbility,
			Target:   AbilitiesValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	case RefIDTypeAbilityModifier:
		idParts := strings.Split(refID, ".")
		if len(idParts) != 2 {
			r.error = fmt.Errorf("invalid ability modifier id: %s", refID)
			return Operation{}
		}

		abilityID := idParts[0]
		modifierID := idParts[1]

		ability, ok := r.reference.Abilities[abilityID]
		if !ok {
			r.error = fmt.Errorf("ability \"%s\" not found", abilityID)
			return Operation{}
		}

		_, ok = ability.Modifiers[modifierID]
		if !ok {
			r.error = fmt.Errorf("modifier \"%s\" not found for ability \"%s\"", modifierID, abilityID)
			return Operation{}
		}

		return Operation{
			Type:     OperationTypeModifyAbility,
			Target:   AbilityModifiersValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	case RefIDTypeDomain:
		_, ok := r.reference.Domains[refID]
		if !ok {
			r.error = fmt.Errorf("domain \"%s\" not found", refID)
			return Operation{}
		}
		return Operation{
			Type:     OperationTypeSet,
			Target:   DomainsValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	case RefIDTypeSkill:
		_, ok := r.reference.Skills[refID]
		if !ok {
			r.error = fmt.Errorf("skill \"%s\" not found", refID)
			return Operation{}
		}
		return Operation{
			Type:     OperationTypeAddSkill,
			Target:   SkillsValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	default:
		r.error = fmt.Errorf("invalid reference type: %s", refIDType)
		return Operation{}
	}
}

func expectInt(name string, value any) int {
	valueInt, ok := value.(int)
	if !ok {
		panic(fmt.Sprintf("%s is not an int, instead %T", name, value))
	}
	return valueInt
}

func expectString(name string, value any) string {
	valueString, ok := value.(string)
	if !ok {
		panic(fmt.Sprintf("%s is not a string, instead %T", name, value))
	}
	return valueString
}

func (r *Resolver) EvaluateNode(node string) {
	if r.visited[node] {
		return
	}
	r.visited[node] = true

	operations, ok := r.operations[node]
	if !ok {
		r.error = fmt.Errorf("node \"%s\" does not exist", node)
		return
	}

	for _, operation := range operations {
		r.trace.Push(operation)
		r.EvaluateOperation(operation)
		if r.error != nil {
			return
		}
		r.trace.Pop()
	}
	r.completed[node] = true
}

func (r *Resolver) EvaluateOperation(operation *Operation) {
	// evaluate prereqs
	for _, assertion := range operation.Prereqs {
		if !r.checkAssertion(&assertion) {
			log.Println("skipping operation due to failed assertion")
			log.Println(operation)
			log.Println(assertion)
			return
		}
	}

	// evaluate the value of the oepration
	r.trace.Push(operation.ValueRef)
	result := r.EvaluateValueRef(&operation.ValueRef)
	if r.error != nil {
		return
	}
	r.trace.Pop()

	switch operation.Type {
	case OperationTypeSet:
		r.values[operation.Target] = result
	case OperationTypeAddSkill:
		skillID := result.(string)

		_, ok := r.values[SkillsValueName]
		if !ok {
			r.values[SkillsValueName] = make([]string, 0)
		}

		skills := r.values[SkillsValueName].([]string)
		if !slices.Contains(skills, skillID) {
			skills = append(skills, skillID)
		}
		r.values[SkillsValueName] = skills
	case OperationTypeAddAbility:
		abilityID := result.(string)

		_, ok := r.values[AbilitiesValueName]
		if !ok {
			r.values[AbilitiesValueName] = make([]string, 0)
		}

		abilities := r.values[AbilitiesValueName].([]string)
		if !slices.Contains(abilities, abilityID) {
			abilities = append(abilities, abilityID)
		}
		r.values[AbilitiesValueName] = abilities
	case OperationTypeModifyAbility:
		modifierID := result.(string)

		_, ok := r.values[AbilityModifiersValueName]
		if !ok {
			r.values[AbilityModifiersValueName] = make([]string, 0)
		}

		modifiers := r.values[AbilityModifiersValueName].([]string)
		if !slices.Contains(modifiers, modifierID) {
			modifiers = append(modifiers, modifierID)
		}
		r.values[AbilityModifiersValueName] = modifiers
	default:
		r.error = fmt.Errorf("unknown operation type: %s", operation.Type)
		return
	}
}

func (r *Resolver) EvaluateValueRef(valueRef *ValueRef) any {
	switch valueRef.Type {
	case ValueRefTypeInt:
		return valueRef.Value.(int)
	case ValueRefTypeString:
		return valueRef.Value.(string)
	case ValueRefTypeID:
		id := valueRef.Value.(string)

		// if node value has already been evaluated, return it
		if r.completed[id] {
			value, ok := r.values[id]
			if !ok {
				r.error = fmt.Errorf("node \"%s\" was processed with no value", id)
				return nil
			}
			return value
		}

		// if the node has been visited, but not completed
		// then there is a circular reference
		if r.visited[id] {
			r.error = fmt.Errorf("circular reference detected for node \"%s\"", id)
			return nil
		}

		// else the node has not been evaluated, so process it
		r.trace.Push("node:" + id)
		r.EvaluateNode(id)
		if r.error != nil {
			return nil
		}
		r.trace.Pop()

		value, ok := r.values[id]
		if !ok {
			r.error = fmt.Errorf("node \"%s\" was processed with no value", id)
			return nil
		}
		return value

	case ValueRefTypeExpression:
		valueExpression := valueRef.Value.(*Expression)

		exprValue := r.EvaluateExpression(valueExpression)
		if r.error != nil {
			return nil
		}

		return exprValue
	case ValueRefTypeRefID:
		return valueRef.Value.(string)
	default:
		r.error = fmt.Errorf("invalid ValueRef type: %s", valueRef.Type)
		return nil
	}
}

func (r *Resolver) EvaluateExpression(expression *Expression) int {
	switch expression.Type {
	case ExprTypeAdd:
		var result int
		for _, arg := range expression.Args {
			value := r.EvaluateValueRef(&arg)
			if r.error != nil {
				return 0
			}

			valueInt, ok := value.(int)
			if !ok {
				r.error = fmt.Errorf("argument is not an int")
				return 0
			}

			result += valueInt
		}
		return result
	case ExprTypeSubtract:
		var result int

		if len(expression.Args) != 2 {
			r.error = fmt.Errorf("subtract requires exactly two arguments")
			return 0
		}

		arg1 := r.EvaluateValueRef(&expression.Args[0])
		if r.error != nil {
			return 0
		}

		arg2 := r.EvaluateValueRef(&expression.Args[1])
		if r.error != nil {
			return 0
		}

		arg1Int, ok := arg1.(int)
		if !ok {
			r.error = fmt.Errorf("first argument is not an int")
			return 0
		}

		arg2Int, ok := arg2.(int)
		if !ok {
			r.error = fmt.Errorf("second argument is not an int")
			return 0
		}

		result = arg1Int - arg2Int

		return result
	default:
		r.error = fmt.Errorf("unknown expression type: %s", expression.Type)
		return 0
	}
}

func (r *Resolver) checkAssertion(assertion *Assertion) bool {
	log.Printf("checking assertion: %s\n", assertion)
	target := assertion.TargetID
	valueRefs := assertion.Values

	targetValue, ok := r.values[target]
	if !ok {
		log.Printf("assert false: %s %s", target, valueRefs)
		return false
	}

	switch targetValue := targetValue.(type) {
	case int:
		for _, valueRef := range valueRefs {
			value := r.EvaluateValueRef(&valueRef)
			if r.error != nil {
				continue
			}

			valueInt, ok := value.(int)
			if !ok {
				continue
			}

			if valueInt == targetValue {
				log.Printf("assert true: %s %s", target, valueRefs)
				return true
			}
		}
	case string:
		for _, valueRef := range valueRefs {
			value := r.EvaluateValueRef(&valueRef)
			if r.error != nil {
				continue
			}

			valueString, ok := value.(string)
			if !ok {
				continue
			}

			if valueString == targetValue {
				log.Printf("assert true: %s %s", target, valueRefs)
				return true
			}
		}
	default:
		log.Printf("assert false: %s %s", target, valueRefs)
		return false
	}
	log.Printf("assert false: %s %s", target, valueRefs)
	return false
}

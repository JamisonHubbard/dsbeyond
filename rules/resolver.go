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
	FeaturesValueName         = "features"
	KitsValueName             = "kits"
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

	// process values to unflatten them
	r.unflattenValues()
	if r.error != nil {
		return model.Sheet{}, r.error
	}

	// create sheet
	data, err := json.Marshal(r.values)
	if err != nil {
		return model.Sheet{}, fmt.Errorf("failed to marshal sheet: %w", err)
	}
	var sheet model.Sheet
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		return model.Sheet{}, fmt.Errorf("failed to unmarshal sheet: %w", err)
	}

	return sheet, nil
}

func (r *Resolver) unflattenValues() {
	unflattened := make(map[string]any)
	for key, value := range r.values {
		parts := strings.Split(key, ".")
		current := unflattened
		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
				break
			}
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]any)
			}
			current = current[part].(map[string]any)
		}
	}
	r.values = unflattened
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
			Type:     OperationTypeAddDomain,
			Target:   DomainsValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	case RefIDTypeFeature:
		_, ok := r.reference.Features[refID]
		if !ok {
			r.error = fmt.Errorf("feature \"%s\" not found", refID)
			return Operation{}
		}
		return Operation{
			Type:     OperationTypeAddFeature,
			Target:   FeaturesValueName,
			ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: refID, RefIDType: refIDType},
		}
	case RefIDTypeKit:
		_, ok := r.reference.Kits[refID]
		if !ok {
			r.error = fmt.Errorf("kit \"%s\" not found", refID)
			return Operation{}
		}
		return Operation{
			Type:     OperationTypeAddKit,
			Target:   KitsValueName,
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
	case OperationTypeAddDomain:
		domainID := result.(string)

		_, ok := r.values[DomainsValueName]
		if !ok {
			r.values[DomainsValueName] = make([]string, 0)
		}

		domains := r.values[DomainsValueName].([]string)
		if !slices.Contains(domains, domainID) {
			domains = append(domains, domainID)
		}
		r.values[DomainsValueName] = domains
	case OperationTypeAddFeature:
		featureID := result.(string)

		_, ok := r.values[FeaturesValueName]
		if !ok {
			r.values[FeaturesValueName] = make([]string, 0)
		}

		features := r.values[FeaturesValueName].([]string)
		if !slices.Contains(features, featureID) {
			features = append(features, featureID)
		}
		r.values[FeaturesValueName] = features
	case OperationTypeAddKit:
		kitID := result.(string)

		_, ok := r.values[KitsValueName]
		if !ok {
			r.values[KitsValueName] = make([]string, 0)
		}

		kits := r.values[KitsValueName].([]string)
		if !slices.Contains(kits, kitID) {
			kits = append(kits, kitID)
		}
		r.values[KitsValueName] = kits

		// process the kit and add its effects
		r.handleKitOperations(kitID)
		if r.error != nil {
			return
		}
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

// TODO: handle melee and ranged damage bonuses, and ranged distance bonus
func (r *Resolver) handleKitOperations(kitID string) {
	kit, ok := r.reference.Kits[kitID]
	if !ok {
		r.error = fmt.Errorf("kit \"%s\" not found", kitID)
		return
	}

	if kit.Bonuses.StaminaBonus != 0 {
		// prepare an operation to add the bonus
		operation := Operation{
			Type:   OperationTypeSet,
			Target: "health.max_stamina",
			ValueRef: ValueRef{Type: ValueRefTypeExpression, Value: &Expression{
				Type: ExprTypeAdd,
				Args: []ValueRef{
					{Type: ValueRefTypeID, Value: "health.max_stamina"},
					{Type: ValueRefTypeInt, Value: kit.Bonuses.StaminaBonus},
				},
			}},
		}

		// if stamina was already calculated, add the bonus now
		if r.visited["health.max_stamina"] {
			r.trace.Push(operation)
			r.EvaluateOperation(&operation)
			if r.error != nil {
				return
			}
			r.trace.Pop()
		} else {
			r.operations["health.max_stamina"] = append(r.operations["health.max_stamina"], &operation)
		}
	}

	if kit.Bonuses.SpeedBonus != 0 {
		// prepare an operation to add the bonus
		operation := Operation{
			Type:   OperationTypeSet,
			Target: "movement.speed",
			ValueRef: ValueRef{Type: ValueRefTypeExpression, Value: &Expression{
				Type: ExprTypeAdd,
				Args: []ValueRef{
					{Type: ValueRefTypeID, Value: "movement.speed"},
					{Type: ValueRefTypeInt, Value: kit.Bonuses.SpeedBonus},
				},
			}},
		}

		// if speed was already calculated, add the bonus now
		if r.visited["movement.speed"] {
			r.trace.Push(operation)
			r.EvaluateOperation(&operation)
			if r.error != nil {
				return
			}
			r.trace.Pop()
		} else {
			r.operations["movement.speed"] = append(r.operations["movement.speed"], &operation)
		}
	}

	if kit.Bonuses.StabilityBonus != 0 {
		// prepare an operation to add the bonus
		operation := Operation{
			Type:   OperationTypeSet,
			Target: "movement.stability",
			ValueRef: ValueRef{Type: ValueRefTypeExpression, Value: &Expression{
				Type: ExprTypeAdd,
				Args: []ValueRef{
					{Type: ValueRefTypeID, Value: "movement.stability"},
					{Type: ValueRefTypeInt, Value: kit.Bonuses.StabilityBonus},
				},
			}},
		}

		// if stability was already calculated, add the bonus now
		if r.visited["movement.stability"] {
			r.trace.Push(operation)
			r.EvaluateOperation(&operation)
			if r.error != nil {
				return
			}
			r.trace.Pop()
		} else {
			r.operations["movement.stability"] = append(r.operations["movement.stability"], &operation)
		}
	}

	if kit.Bonuses.DisengageBonus != 0 {
		// prepare an operation to add the bonus
		operation := Operation{
			Type:   OperationTypeSet,
			Target: "movement.disengage",
			ValueRef: ValueRef{Type: ValueRefTypeExpression, Value: &Expression{
				Type: ExprTypeAdd,
				Args: []ValueRef{
					{Type: ValueRefTypeID, Value: "movement.disengage"},
					{Type: ValueRefTypeInt, Value: kit.Bonuses.DisengageBonus},
				},
			}},
		}

		// if disengage was already calculated, add the bonus now
		if r.visited["movement.disengage"] {
			r.trace.Push(operation)
			r.EvaluateOperation(&operation)
			if r.error != nil {
				return
			}
			r.trace.Pop()
		} else {
			r.operations["movement.disengage"] = append(r.operations["movement.disengage"], &operation)
		}
	}

	// add the kit's abilities
	for _, abilityID := range kit.Abilities {
		_, ok := r.reference.Abilities[abilityID]
		if !ok {
			r.error = fmt.Errorf("ability \"%s\" not found", abilityID)
			return
		}

		_, ok = r.values[AbilitiesValueName]
		if !ok {
			r.values[AbilitiesValueName] = make([]string, 0)
		}

		r.values[AbilitiesValueName] = append(r.values[AbilitiesValueName].([]string), abilityID)
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

		// NOTE: the below is disabled to allow self-referential operations
		// for example: a = a + 1
		// may need to revisit this if it becomes an issue

		// // if the node has been visited, but not completed
		// // then there is a circular reference
		// if r.visited[id] {
		// 	r.error = fmt.Errorf("circular reference detected for node \"%s\"", id)
		// 	return nil
		// }

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

	switch assertion.Type {
	// the value indicated by `target` should match one of the supplied values
	case AssertionTypeValue:
		actualValue, ok := r.values[assertion.Target]
		if !ok {
			log.Println("assertion false: target not found")
			return false
		}

		for _, valueRef := range assertion.Values {
			value := r.EvaluateValueRef(&valueRef)
			if r.error != nil {
				log.Println("WARNING failed to evaluate value ref")
				r.error = nil
				continue
			}

			switch value := value.(type) {
			case int:
				actualInt, ok := actualValue.(int)
				if !ok {
					continue
				}

				if value == actualInt {
					log.Println("assertion true")
					return true
				}
			case string:
				actualString, ok := actualValue.(string)
				if !ok {
					continue
				}

				if value == actualString {
					log.Println("assertion true")
					return true
				}
			default:
				log.Println("WARNING encountered unexpected value type")
				continue
			}
		}

		log.Println("assertion false, target does not match any values")
		return false
	// each value should be an id for one of the referenced values of the given
	// type
	case AssertionTypeRefArray:
		switch assertion.RefType {
		case RefIDTypeAbility:
			return r.checkArrayForIDs(AbilitiesValueName, &assertion.Values)
		case RefIDTypeAbilityModifier:
			return r.checkArrayForIDs(AbilityModifiersValueName, &assertion.Values)
		case RefIDTypeDomain:
			return r.checkArrayForIDs(DomainsValueName, &assertion.Values)
		case RefIDTypeFeature:
			return r.checkArrayForIDs(FeaturesValueName, &assertion.Values)
		case RefIDTypeKit:
			return r.checkArrayForIDs(KitsValueName, &assertion.Values)
		case RefIDTypeSkill:
			return r.checkArrayForIDs(SkillsValueName, &assertion.Values)
		default:
			log.Println("WARNING assertion false: unknown ref type")
			return false
		}
	default:
		log.Println("WARNING assertion false: unknown assertion type")
		return false
	}
}

func (r *Resolver) checkArrayForIDs(arrayID string, valueRefs *[]ValueRef) bool {
	refArray, ok := r.values[arrayID]
	if !ok {
		// check for pending operations
		ops, ok := r.operations[arrayID]
		if !ok {
			log.Printf("assertion false: %s array not found and no pending operations\n", arrayID)
			return false
		}

		// evaluate the node, then proceed
		for _, op := range ops {
			log.Println(*op)
		}
		r.trace.Push("node:" + arrayID)
		r.EvaluateNode(arrayID)
		if r.error != nil {
			log.Printf("WARNING assertion false: failed to evaluate %s array\n", arrayID)
			r.error = nil
			return false
		}
		r.trace.Pop()

		refArray, ok = r.values[arrayID]
		if !ok {
			log.Printf("assertion false: %s array not found after evaluation\n", arrayID)
			return false
		}
	}

	for _, valueRef := range *valueRefs {
		value := r.EvaluateValueRef(&valueRef)
		if r.error != nil {
			r.error = nil
			log.Println("assertion false: failed to evaluate value ref")
			return false
		}

		valueID, ok := value.(string)
		if !ok {
			log.Println("assertion false: value id is not a string")
			return false
		}

		if slices.Contains(refArray.([]string), valueID) {
			continue
		}

		log.Println("assertion false, value not found in ref array")
		return false
	}

	log.Println("assertion true")
	return true
}

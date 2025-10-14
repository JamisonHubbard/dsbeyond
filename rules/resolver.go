package rules

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/JamisonHubbard/dsbeyond/model"
)

const (
	ChoiceTypeOptionSelect = "option_select"
	ChoiceTypeRefSelect    = "ref_select"
	ChoiceTypeInput        = "input"

	DecisionTypeID        = "id"
	DecisionTypeRefID     = "ref_id"
	DecisionTypeOperation = "operation"
	DecisionTypeValue     = "value"

	ExprTypeAdd      = "add"
	ExprTypeSubtract = "subtract"

	OperationTypeSet           = "set"
	OperationTypeAddSkill      = "add_skill"
	OperationTypeAddAbility    = "add_ability"
	OperationTypeModifyAbility = "modify_ability"

	ValueRefTypeExpr   = "expression"
	ValueRefTypeID     = "id"
	ValueRefTypeInt    = "int"
	ValueRefTypeRefID  = "ref_id"
	ValueRefTypeString = "string"
)

func NewResolver(character model.Character, decisions map[string]Decision, reference *Reference) *Resolver {
	return &Resolver{
		character:  character,
		decisions:  decisions,
		reference:  reference,
		visited:    make(map[string]bool),
		dependency: NewDependencyTracker(),
		trace:      Trace{},
		error:      nil,
	}
}

type Resolver struct {
	// inputs
	character model.Character
	decisions map[string]Decision
	reference *Reference

	// internals
	ctx        Context
	visited    map[string]bool
	dependency *DependencyTracker
	trace      Trace
	error      error
}

func (r *Resolver) Resolve() (model.Sheet, error) {
	// get class data from reference
	class, ok := r.reference.Classes[r.character.ClassID]
	if !ok {
		return model.Sheet{}, fmt.Errorf("class \"%s\" not found", r.character.ClassID)
	}

	// setup context
	r.ctx = Context{
		Values:     make(map[string]any),
		Operations: make(map[string][]*Operation),
	}

	// load pre ops
	err := r.loadPreOperations(r.character.Level, &class)
	if err != nil {
		return model.Sheet{}, err
	}

	// evaluate preops nodes
	for node := range r.ctx.Operations {
		r.trace.Push("Node:" + node)
		r.EvaluateNode(node)
		if r.error != nil {
			return model.Sheet{}, r.error
		}
		r.trace.Pop()
	}

	// execute choice operations
	err = r.executeChoiceOperations(&class)
	if err != nil {
		return model.Sheet{}, err
	}

	// load post ops
	err = r.loadPostOperations(r.character.Level, &class)
	if err != nil {
		return model.Sheet{}, err
	}

	// evaluate preops nodes
	for node := range r.ctx.Operations {
		r.trace.Push("Node:" + node)
		r.EvaluateNode(node)
		if r.error != nil {
			return model.Sheet{}, r.error
		}
		r.trace.Pop()
	}

	// logging
	// fmt.Println(r.dependency.String())

	// pretty print context values
	prettyContext, err := json.MarshalIndent(r.ctx.Values, "", "  ")
	if err != nil {
		log.Println("WARNING failed to pretty print context")
	} else {
		log.Println(string(prettyContext))
	}

	skills, ok := r.ctx.Values["skills"]
	if !ok {
		skills = []string{}
	}

	// create sheet
	sheet := model.Sheet{
		ClassID: r.character.ClassID,
		Level:   r.character.Level,
		Core: model.SheetCore{
			HeroicResource: expectString("heroic_resource", r.ctx.Values["heroic_resource"]),
			Characteristics: model.Characteristics{
				Might:     expectInt("characteristics.might", r.ctx.Values["characteristics.might"]),
				Agility:   expectInt("characteristics.agility", r.ctx.Values["characteristics.agility"]),
				Reason:    expectInt("characteristics.reason", r.ctx.Values["characteristics.reason"]),
				Intuition: expectInt("characteristics.intuition", r.ctx.Values["characteristics.intuition"]),
				Presence:  expectInt("characteristics.presence", r.ctx.Values["characteristics.presence"]),
			},
			Health: model.Health{
				MaxStamina:    expectInt("health.max_stamina", r.ctx.Values["health.max_stamina"]),
				MaxRecoveries: expectInt("health.max_recoveries", r.ctx.Values["health.max_recoveries"]),
			},
			Potencies: model.Potencies{
				Strong:  expectInt("potencies.strong", r.ctx.Values["potencies.strong"]),
				Average: expectInt("potencies.average", r.ctx.Values["potencies.average"]),
				Weak:    expectInt("potencies.weak", r.ctx.Values["potencies.weak"]),
			},
		},
		Skills: skills.([]string),
	}

	return sheet, nil
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

	operations, ok := r.ctx.Operations[node]
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
}

func (r *Resolver) EvaluateOperation(operation *Operation) {
	log.Printf("evaluating operation: %s", operation)

	// evaluate prereqs
	for _, assertion := range operation.Prereqs {
		if !r.checkAssertion(&assertion) {
			return
		}
	}

	// evaluate the value of the oepration
	target := operation.Target
	valueRef := operation.ValueRef

	// track dependencies
	switch valueRef.Type {
	case ValueRefTypeID:
		r.dependency.Add(target, valueRef.Value.(string))
	case ValueRefTypeExpr:
		dependencies := r.evaluateExprDepenencies(valueRef.Value.(*Expression))
		for _, dependency := range dependencies {
			if target != dependency {
				r.dependency.Add(target, dependency)
			}
		}
	}

	r.trace.Push(valueRef)
	result := r.EvaluateValueRef(&valueRef)
	if r.error != nil {
		return
	}
	r.trace.Pop()

	switch operation.Type {
	case OperationTypeSet:
		r.ctx.Values[target] = result
	case OperationTypeAddSkill:
		skillID := result.(string)

		_, ok := r.ctx.Values["skills"]
		if !ok {
			r.ctx.Values["skills"] = make([]string, 0)
		}

		// TODO verify skill is not already in skills
		skills := r.ctx.Values["skills"].([]string)
		skills = append(skills, skillID)
		r.ctx.Values["skills"] = skills
	case OperationTypeAddAbility:
		abilityID := result.(string)

		_, ok := r.ctx.Values["abilities"]
		if !ok {
			r.ctx.Values["abilities"] = make([]string, 0)
		}

		// TODO verify ability is not already in abilities
		abilities := r.ctx.Values["abilities"].([]string)
		abilities = append(abilities, abilityID)
		r.ctx.Values["abilities"] = abilities
	case OperationTypeModifyAbility:
		modifierID := result.(string)

		_, ok := r.ctx.Values["ability_modifiers"]
		if !ok {
			r.ctx.Values["ability_modifiers"] = make([]string, 0)
		}

		// TODO verify modifier is not already in modifiers
		modifiers := r.ctx.Values["ability_modifiers"].([]string)
		modifiers = append(modifiers, modifierID)
		r.ctx.Values["ability_modifiers"] = modifiers
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

		value, ok := r.ctx.Values[id]
		if !ok {
			// see if there's unperformed operations for this id
			_, ok := r.ctx.Operations[id]
			if !ok {
				r.error = fmt.Errorf("value with id \"%s\" does not exist", id)
				return nil
			}

			// process the node in order to get the value
			r.trace.Push("Node:" + id)
			r.EvaluateNode(id)
			if r.error != nil {
				return nil
			}
			r.trace.Pop()

			value, ok = r.ctx.Values[id]
			if !ok {
				r.error = fmt.Errorf("value with id \"%s\" does not exist", id)
				return nil
			}
		}

		switch value := value.(type) {
		case int:
			return value
		case string:
			return value
		default:
			r.error = fmt.Errorf("invalid type for node value")
			return nil
		}
	case ValueRefTypeExpr:
		valueExpr := valueRef.Value.(*Expression)

		exprValue := r.EvaluateExpression(valueExpr)
		if r.error != nil {
			return nil
		}

		return exprValue
	case ValueRefTypeRefID:
		refID, ok := valueRef.Value.(string)
		if !ok {
			r.error = fmt.Errorf("invalid reference id")
			return nil
		}

		switch valueRef.RefType {
		case "skill":
			_, ok = r.reference.Skills[refID]
		case "domain":
			_, ok = r.reference.Domains[refID]
		case "ability":
			_, ok = r.reference.Abilities[refID]
		case "ability_modifier":
			ids := strings.Split(refID, ".")
			if len(ids) != 2 {
				r.error = fmt.Errorf("invalid ability modifier id: %s", refID)
				return nil
			}

			abilityID := ids[0]
			modifierID := ids[1]
			ability, ok := r.reference.Abilities[abilityID]
			if !ok {
				r.error = fmt.Errorf("ability \"%s\" not found", abilityID)
				return nil
			}

			_, ok = ability.Modifiers[modifierID]
			if !ok {
				r.error = fmt.Errorf("modifier \"%s\" not found for ability \"%s\"", modifierID, abilityID)
				return nil
			}

			return refID
		default:
			r.error = fmt.Errorf("invalid reference type: %s", valueRef.RefType)
			return nil
		}
		if !ok {
			r.error = fmt.Errorf("%s \"%s\" not found", valueRef.RefType, refID)
			return nil
		}

		return refID
	default:
		r.error = fmt.Errorf("invalid ValueRef type: %s", valueRef.Type)
		return nil
	}
}

func (r *Resolver) EvaluateExpression(expression *Expression) any {
	switch expression.Type {
	case ExprTypeAdd:
		return r.evaluateAdd(expression)
	case ExprTypeSubtract:
		return r.evaluateSubtract(expression)
	default:
		r.error = fmt.Errorf("unknown expression type: %s", expression.Type)
		return nil
	}
}

func (r *Resolver) evaluateAdd(expression *Expression) any {
	var result int
	for _, arg := range expression.Args {
		value := r.EvaluateValueRef(&arg)
		if r.error != nil {
			return nil
		}

		valueInt, ok := value.(int)
		if !ok {
			r.error = fmt.Errorf("argument is not an int")
			return nil
		}

		result += valueInt
	}

	return result
}

func (r *Resolver) evaluateSubtract(expression *Expression) any {
	var result int

	if len(expression.Args) != 2 {
		r.error = fmt.Errorf("subtract requires exactly two arguments")
		return nil
	}

	arg1 := r.EvaluateValueRef(&expression.Args[0])
	if r.error != nil {
		return nil
	}

	arg2 := r.EvaluateValueRef(&expression.Args[1])
	if r.error != nil {
		return nil
	}

	arg1Int, ok := arg1.(int)
	if !ok {
		r.error = fmt.Errorf("first argument is not an int")
		return nil
	}

	arg2Int, ok := arg2.(int)
	if !ok {
		r.error = fmt.Errorf("second argument is not an int")
		return nil
	}

	result = arg1Int - arg2Int

	return result
}

func (r *Resolver) evaluateExprDepenencies(expression *Expression) []string {
	var dependencies []string
	for _, arg := range expression.Args {
		switch arg.Type {
		case ValueRefTypeID:
			dependencies = append(dependencies, arg.Value.(string))
		case ValueRefTypeExpr:
			dependencies = append(dependencies, r.evaluateExprDepenencies(arg.Value.(*Expression))...)
		}
	}
	return dependencies
}

func (r *Resolver) executeChoiceOperations(class *Class) error {
	// compile a list of choices
	var choices []Choice
	choices = append(choices, class.Basics.Choices...)

	for classLevel, classLevelDefn := range class.Levels {
		if classLevel <= r.character.Level {
			choices = append(choices, classLevelDefn.Choices...)
		}
	}

	for _, choice := range choices {
		// determine if the choice is applicable
		prereqsMet := true
		for _, assertion := range choice.Prereqs {
			if !r.checkAssertion(&assertion) {
				prereqsMet = false
				break
			}
		}
		if !prereqsMet {
			continue
		}

		// resolve the choice
		operations, err := r.resolveChoice(&choice)
		if err != nil {
			return fmt.Errorf("failed to resolve choice: %s", err)
		}

		for _, operation := range operations {
			r.EvaluateOperation(&operation)
			if r.error != nil {
				return fmt.Errorf("failed to resolve choice: %s", r.error)
			}
		}
	}

	return nil
}

func (r *Resolver) checkAssertion(assertion *Assertion) bool {
	log.Printf("checking assertion: %s", assertion)
	target := assertion.TargetID
	valueRefs := assertion.Values

	targetValue, ok := r.ctx.Values[target]
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

func (r *Resolver) loadPreOperations(characterLevel int, class *Class) error {
	var operations []Operation

	// add basic operations for the class
	operations = append(operations, class.Basics.PreOperations...)

	// add operations from levels the character has reached
	for classLevel, classLevelDefn := range class.Levels {
		if classLevel <= characterLevel {
			operations = append(operations, classLevelDefn.PreOperations...)
		}
	}

	r.ctx.Operations = make(map[string][]*Operation)
	for _, operation := range operations {
		r.ctx.AddOperation(operation)
	}
	return nil
}

func (r *Resolver) loadPostOperations(characterLevel int, class *Class) error {
	var operations []Operation

	// add basic operations for the class
	operations = append(operations, class.Basics.PostOperations...)

	// add operations from levels the character has reached
	for classLevel, classLevelDefn := range class.Levels {
		if classLevel <= characterLevel {
			operations = append(operations, classLevelDefn.PostOperations...)
		}
	}

	r.ctx.Operations = make(map[string][]*Operation)
	for _, operation := range operations {
		r.ctx.AddOperation(operation)
	}
	return nil
}

func (r *Resolver) resolveChoice(choice *Choice) ([]Operation, error) {
	var operations []Operation

	// find the corresponding decision
	var decision *Decision
	for _, d := range r.decisions {
		if d.ChoiceID == choice.ID {
			decision = &d
			break
		}
	}

	if decision == nil {
		return nil, fmt.Errorf("decision for choice \"%s\" not found", choice.ID)
	}

	switch choice.Type {
	case ChoiceTypeOptionSelect:
		// find the corresponding option
		var option *Option
		for _, o := range choice.Options {
			if o.ID == decision.OptionID {
				option = &o
				break
			}
		}

		if option == nil {
			return nil, fmt.Errorf("option \"%s\" for choice \"%s\" not found", decision.OptionID, choice.ID)
		}

		operations = append(operations, option.Operations...)
	case ChoiceTypeRefSelect:
		switch choice.RefType {
		case "ability":
			_, ok := r.reference.Abilities[decision.RefID]
			if !ok {
				return nil, fmt.Errorf("ability \"%s\" for choice \"%s\" not found", decision.OptionID, choice.ID)
			}
			operations = append(operations, Operation{
				Type:     OperationTypeAddAbility,
				ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: decision.RefID, RefType: choice.RefType},
			})
		case "skill":
			_, ok := r.reference.Skills[decision.RefID]
			if !ok {
				return nil, fmt.Errorf("skill \"%s\" for choice \"%s\" not found", decision.OptionID, choice.ID)
			}
			operations = append(operations, Operation{
				Type:     OperationTypeAddSkill,
				ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: decision.RefID, RefType: choice.RefType},
			})
		case "domain":
			_, ok := r.reference.Domains[decision.RefID]
			if !ok {
				return nil, fmt.Errorf("domain \"%s\" for choice \"%s\" not found", decision.OptionID, choice.ID)
			}
			operations = append(operations, Operation{
				Type:     OperationTypeSet,
				Target:   "domain",
				ValueRef: ValueRef{Type: ValueRefTypeRefID, Value: decision.RefID, RefType: choice.RefType},
			})
		default:
			return nil, fmt.Errorf("invalid reference type: %s", choice.RefType)
		}
	case ChoiceTypeInput:
		if decision.Type != DecisionTypeValue {
			return nil, fmt.Errorf("invalid decision type for choice \"%s\": %s", choice.ID, decision.Type)
		}

		target := decision.Target
		valueRef := decision.Value

		operations = append(operations, Operation{
			Type:     OperationTypeSet,
			Target:   target,
			ValueRef: valueRef,
		})
	default:
		return nil, fmt.Errorf("unknown choice type: %s", choice.Type)
	}

	return operations, nil
}

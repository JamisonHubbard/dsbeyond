package rules

import (
	"fmt"

	"github.com/JamisonHubbard/dsbeyond/model"
)

const (
	ExprTypeAdd      = "add"
	ExprTypeSubtract = "subtract"

	OperationTypeSet = "set"

	ValueRefTypeInt  = "int"
	ValueRefTypeID   = "id"
	ValueRefTypeExpr = "expression"
)

func NewResolver(character model.Character, decisions []Decision) *Resolver {
	return &Resolver{
		character:  character,
		decisions:  decisions,
		visited:    make(map[string]bool),
		dependency: NewDependencyTracker(),
		trace:      Trace{},
		error:      nil,
	}
}

type Resolver struct {
	// inputs
	character model.Character
	decisions []Decision

	// internals
	ctx        Context
	visited    map[string]bool
	dependency *DependencyTracker
	trace      Trace
	error      error
}

func (r *Resolver) Resolve() (model.Sheet, error) {
	ctx, err := Parse(r.character.ClassID, r.character.Level, r.decisions)
	if err != nil {
		return model.Sheet{}, err
	}
	r.ctx = ctx

	// seed default values before execution
	r.ctx.Values["skills"] = []model.Skill{}
	// r.ctx.Values["abilities"] = []model.Ability{}

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

	sheet := model.Sheet{
		ClassID: r.character.ClassID,
		Level:   r.character.Level,
		Core: model.SheetCore{
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
		Skills: []model.Skill{},
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
			r.dependency.Add(target, dependency)
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
	default:
		r.error = fmt.Errorf("unknown operation type: %s", operation.Type)
		return
	}
}

func (r *Resolver) EvaluateValueRef(valueRef *ValueRef) any {
	switch valueRef.Type {
	case ValueRefTypeInt:
		return valueRef.Value.(int)
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

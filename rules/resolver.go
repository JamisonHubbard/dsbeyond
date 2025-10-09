package rules

import (
	"encoding/json"
	"fmt"

	"github.com/JamisonHubbard/dsbeyond/model"
)

const (
	ExprTypeAdd      = "add"
	ExprTypeSubtract = "subtract"

	OperationTypeSet = "set"
	OperationTypeAdd = "add"

	ValueRefTypeInt  = "int"
	ValueRefTypeID   = "id"
	ValueRefTypeExpr = "expression"
)

func NewResolver(character model.Character) *Resolver {
	return &Resolver{
		character: character,
		visited:   make(map[string]bool),
		trace:     Trace{},
		error:     nil,
	}
}

type Context struct {
	Values     map[string]any          `json:"values"`
	Operations map[string][]*Operation `json:"operations"`
}

func (c *Context) AddOperation(operation Operation) {
	operations, ok := c.Operations[operation.Target]
	if !ok {
		c.Operations[operation.Target] = []*Operation{&operation}
	} else {
		c.Operations[operation.Target] = append(operations, &operation)
	}
}

func (c *Context) GetOperations(target string) []*Operation {
	return c.Operations[target]
}

func (c *Context) GetValue(target string) any {
	return c.Values[target]
}

func (c *Context) NodeExists(target string) bool {
	_, valueOk := c.Values[target]
	_, operationOk := c.Operations[target]

	return valueOk || operationOk
}

type Resolver struct {
	// inputs
	character model.Character

	// internals
	ctx     Context
	visited map[string]bool
	trace   Trace
	error   error
}

func (r *Resolver) Resolve() (map[string]any, error) {
	ctx, err := Parse(r.character.ClassID, r.character.Level)
	if err != nil {
		return nil, err
	}
	r.ctx = ctx

	for node := range r.ctx.Operations {
		r.trace.Push("Node:" + node)
		r.EvaluateNode(node)
		if r.error != nil {
			return nil, r.error
		}
		r.trace.Pop()
	}

	return r.ctx.Values, nil
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

	r.trace.Push(valueRef)
	result := r.EvaluateValueRef(&valueRef)
	if r.error != nil {
		return
	}
	r.trace.Pop()

	switch operation.Type {
	case OperationTypeSet:
		r.ctx.Values[target] = result
	case OperationTypeAdd:
		// determine if value already exists
		_, ok := r.ctx.Values[target]
		if !ok {
			r.ctx.Values[target] = 0
		}

		// make sure the value and the result are ints
		switch currentValue := r.ctx.Values[target].(type) {
		case int:
			switch resultValue := result.(type) {
			case int:
				r.ctx.Values[target] = currentValue + resultValue
			default:
				r.error = fmt.Errorf("cannot perform an add operation on a non-int")
				return
			}
		default:
			r.error = fmt.Errorf("cannot perform an add operation on a non-int")
			return
		}
	}
}

func (r *Resolver) EvaluateValueRef(valueRef *ValueRef) any {
	switch valueRef.Type {
	case ValueRefTypeInt:
		switch value := valueRef.Value.(type) {
		case float64:
			return int(value)
		case int:
			return value
		default:
			r.error = fmt.Errorf("value is not an int")
			return nil
		}
	case ValueRefTypeID:
		id, ok := valueRef.Value.(string)
		if !ok {
			r.error = fmt.Errorf("value id is not a string")
			return nil
		}

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
		var valueRefExpr *Expression
		switch value := valueRef.Value.(type) {
		case *Expression:
			valueRefExpr = value
		case map[string]any:
			// un-marshal and re-marshal into a proper expression
			data, err := json.Marshal(valueRef.Value)
			if err != nil {
				r.error = fmt.Errorf("failed to marshal expression: %s", err)
				return nil
			}

			var expr Expression
			err = json.Unmarshal(data, &expr)
			if err != nil {
				r.error = fmt.Errorf("failed to unmarshal expression: %s", err)
				return nil
			}

			valueRefExpr = &expr
		default:
			r.error = fmt.Errorf("value is not an expression")
			return nil
		}

		exprValue := r.EvaluateExpression(valueRefExpr)
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

type Trace struct {
	trace []any
}

func (t *Trace) Push(value any) {
	t.trace = append(t.trace, value)
}

func (t *Trace) Pop() {
	t.trace = t.trace[:len(t.trace)-1]
}

func (t *Trace) String() string {
	var result string
	for _, value := range t.trace {
		result += fmt.Sprintf("%s ", value)
	}

	return result
}

package rules

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	ExprTypeAdd      = "add"
	ExprTypeSubtract = "subtract"

	OperationTypeSet = "set"

	ValueRefTypeInt  = "int"
	ValueRefTypeID   = "id"
	ValueRefTypeExpr = "expression"
)

func NewResolver(operations []Operation) *Resolver {
	return &Resolver{
		operations: operations,
		ctx:        make(map[string]any),
		chain:      make([]any, 0),
		error:      nil,
	}
}

type Resolver struct {
	// inputs
	operations []Operation
	// choiceDefns []ChoiceDefinition

	ctx   map[string]any
	chain []any
	error error
}

func (r *Resolver) Resolve() (map[string]any, error) {
	for _, operation := range r.operations {
		r.chain = append(r.chain, operation)
		r.EvaluateOperation(&operation)
		if r.error != nil {
			log.Println(r.chain)

			operationPretty, err := json.MarshalIndent(operation, "", "  ")
			if err != nil {
				return nil, err
			}
			log.Println(string(operationPretty))

			return nil, fmt.Errorf("failed to resolve: %s", r.error)
		}
		r.chain = r.chain[:len(r.chain)-1]
	}
	return r.ctx, nil
}

func (r *Resolver) EvaluateOperation(operation *Operation) {
	switch operation.Type {
	case OperationTypeSet:
		r.chain = append(r.chain, operation.ValueRef)
		valueRef := r.EvaluateValueRef(&operation.ValueRef)
		if r.error != nil {
			return
		}
		r.chain = r.chain[:len(r.chain)-1]

		r.ctx[operation.Target] = valueRef
	default:
		r.error = fmt.Errorf("unknown operation type: %s", operation.Type)
		return
	}
}

func (r *Resolver) EvaluateValueRef(valueRef *ValueRef) any {
	switch valueRef.Type {
	case ValueRefTypeInt:
		switch valueRef.Value.(type) {
		case float64:
			return int(valueRef.Value.(float64))
		case int:
			return valueRef.Value.(int)
		default:
			r.error = fmt.Errorf("value is not an int")
			return nil
		}
	case ValueRefTypeID:
		valueID, ok := valueRef.Value.(string)
		if !ok {
			r.error = fmt.Errorf("value id is not a string")
			return nil
		}

		if value, ok := r.ctx[valueID]; ok {
			return value
		} else {
			r.error = fmt.Errorf("value not found for id: %s", valueID)
			return nil
		}
	case ValueRefTypeExpr:
		var valueRefExpr *Expression
		switch valueRef.Value.(type) {
		case *Expression:
			valueRefExpr = valueRef.Value.(*Expression)
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

		r.chain = append(r.chain, valueRefExpr)
		exprValue := r.EvaluateExpression(valueRefExpr)
		if r.error != nil {
			return nil
		}
		r.chain = r.chain[:len(r.chain)-1]

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
		r.chain = append(r.chain, arg)
		value := r.EvaluateValueRef(&arg)
		if r.error != nil {
			return nil
		}
		r.chain = r.chain[:len(r.chain)-1]

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

	r.chain = append(r.chain, expression.Args[0])
	arg1 := r.EvaluateValueRef(&expression.Args[0])
	if r.error != nil {
		return nil
	}
	r.chain = r.chain[:len(r.chain)-1]

	r.chain = append(r.chain, expression.Args[1])
	arg2 := r.EvaluateValueRef(&expression.Args[1])
	if r.error != nil {
		return nil
	}
	r.chain = r.chain[:len(r.chain)-1]

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

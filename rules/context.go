package rules

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

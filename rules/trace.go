package rules

import "fmt"

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

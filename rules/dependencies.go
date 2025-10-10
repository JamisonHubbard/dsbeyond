package rules

import (
	"fmt"
	"strings"
)

func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		dependents: make(map[string][]string),
	}
}

type DependencyTracker struct {
	dependents map[string][]string
}

// Add a tracked dependency to a node
// `node` is dependent on `dependency`
func (d *DependencyTracker) Add(node string, dependency string) {
	if _, ok := d.dependents[dependency]; !ok {
		d.dependents[dependency] = []string{node}
		return
	}
	d.dependents[dependency] = append(d.dependents[dependency], node)
}

func (d *DependencyTracker) GetDependents(node string) []string {
	return d.dependents[node]
}

func (d *DependencyTracker) String() string {
	var result string
	for node, dependencies := range d.dependents {
		result += fmt.Sprintf("%s: %s\n", node, strings.Join(dependencies, ", "))
	}

	return result
}

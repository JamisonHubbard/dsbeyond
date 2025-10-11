package rules

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Parse(classID string, characterLevel int, decisions []Decision) (Context, error) {
	// read class from JSON file
	class, err := readClassData(classID)
	if err != nil {
		return Context{}, err
	}

	var operations []Operation

	// add basic operations for the class
	operations = append(operations, class.Basics.Operations...)
	choiceOperations, err := resolveDecisions(&class.Basics.Choices, &decisions)
	if err != nil {
		return Context{}, fmt.Errorf("failed to resolve basic decisions: %s", err)
	}
	operations = append(operations, choiceOperations...)

	// add operations from levels the character has reached
	for classLevel, classLevelDefn := range class.Levels {
		if classLevel <= characterLevel {
			operations = append(operations, classLevelDefn.Operations...)
			choiceOperations, err := resolveDecisions(&classLevelDefn.Choices, &decisions)
			if err != nil {
				return Context{}, fmt.Errorf("failed to resolve decisions for level %d: %s", classLevel, err)
			}
			operations = append(operations, choiceOperations...)
		}
	}

	ctx := Context{
		Values:     make(map[string]any),
		Operations: make(map[string][]*Operation),
	}

	for _, operation := range operations {
		ctx.AddOperation(operation)
	}

	return ctx, nil
}

func resolveDecisions(choices *[]Choice, decisions *[]Decision) ([]Operation, error) {
	var operations []Operation

	for _, choice := range *choices {
		// find the corresponding decision
		var decision *Decision
		for _, d := range *decisions {
			if d.ChoiceID == choice.ID {
				decision = &d
				break
			}
		}

		if decision == nil {
			return nil, fmt.Errorf("decision for choice \"%s\" not found", choice.ID)
		}

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
	}

	return operations, nil
}

func readClassData(classID string) (Class, error) {
	// read class data from json file
	filepath := "data/classes/" + classID + ".json"

	file, err := os.Open(filepath)
	if err != nil {
		return Class{}, fmt.Errorf("failed to read class file: %s", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return Class{}, fmt.Errorf("failed to read class file: %s", err)
	}

	var class Class
	err = json.Unmarshal(byteValue, &class)
	if err != nil {
		return Class{}, fmt.Errorf("failed to decode class file: %s", err)
	}

	return class, nil
}

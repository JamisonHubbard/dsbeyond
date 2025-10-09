package rules

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Parse(classID string, characterLevel int) (Context, error) {
	parsedClass, err := parseClass(classID, characterLevel)
	if err != nil {
		return Context{}, err
	}

	ctx := Context{
		Values:     make(map[string]any),
		Operations: make(map[string][]*Operation),
	}

	for _, operation := range parsedClass.Operations {
		ctx.AddOperation(operation)
	}

	return ctx, nil
}

func parseClass(classID string, characterLevel int) (ParsedClass, error) {
	// read class from JSON file
	class, err := readClassData(classID)
	if err != nil {
		return ParsedClass{}, err
	}

	var operations []Operation
	// var choiceDefns []ChoiceDefinition

	// add basic operations for the class
	operations = append(operations, class.Basics.Operations...)

	// add operations from levels the character has reached
	for classLevel, classLevelDefn := range class.Levels {
		if classLevel <= characterLevel {
			operations = append(operations, classLevelDefn.Operations...)
		}
	}

	return ParsedClass{
		Operations: operations,
	}, nil
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

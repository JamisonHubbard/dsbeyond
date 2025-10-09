package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/JamisonHubbard/dsbeyond/rules"
)

func main() {
	classID := "class_censor"
	choices := []rules.Choice{
		{
			ID:       "initial_characteristics",
			OptionID: "option_1",
		},
	}

	character, err := Resolve(classID, choices)
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		return
	}

	characterPretty, err := json.MarshalIndent(character, "", "  ")
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		fmt.Println(character)
		return
	}

	fmt.Println(string(characterPretty))
}

func Resolve(classID string, choices []rules.Choice) (rules.Sheet, error) {
	_, err := readClass(classID)
	if err != nil {
		return rules.Sheet{}, err
	}

	sheet := rules.Sheet{
		ClassID: classID,
	}

	// process modifiers

	return sheet, nil
}

func readClass(classID string) (rules.Class, error) {
	// read class data from json file
	filepath := "data/classes/" + classID + ".json"

	file, err := os.Open(filepath)
	if err != nil {
		return rules.Class{}, fmt.Errorf("failed to read class file: %s", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return rules.Class{}, fmt.Errorf("failed to read class file: %s", err)
	}

	var class rules.Class
	err = json.Unmarshal(byteValue, &class)
	if err != nil {
		return rules.Class{}, fmt.Errorf("failed to decode class file: %s", err)
	}

	return class, nil
}

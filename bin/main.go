package main

import (
	"encoding/json"
	"fmt"

	"github.com/JamisonHubbard/dsbeyond/model"
	"github.com/JamisonHubbard/dsbeyond/rules"
)

func main() {
	character := model.Character{
		ID:      "test_character",
		ClassID: "class_censor",
		Name:    "Arjhan",
		Level:   1,
	}
	decisions := []rules.Decision{
		{
			ChoiceID: "starting_characteristics",
			OptionID: "an1r2in1",
		},
	}

	sheet, err := ResolveCharacter(character, decisions)
	if err != nil {
		fmt.Println("ERROR " + err.Error())
		return
	}

	sheetPretty, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		fmt.Println("ERROR " + err.Error())
		fmt.Println(character)
		return
	}

	fmt.Println(string(sheetPretty))
}

func ResolveCharacter(character model.Character, decisions []rules.Decision) (model.Sheet, error) {
	resolver := rules.NewResolver(character, decisions)
	sheet, err := resolver.Resolve()
	if err != nil {
		return model.Sheet{}, fmt.Errorf("failed to resolve: %s", err)
	}

	return sheet, nil
}

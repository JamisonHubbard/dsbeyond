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

	ctx, err := ResolveCharacter(character)
	if err != nil {
		fmt.Println("ERROR " + err.Error())
		return
	}

	sheetPretty, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		fmt.Println("ERROR " + err.Error())
		fmt.Println(character)
		return
	}

	fmt.Println(string(sheetPretty))
}

func ResolveCharacter(character model.Character) (map[string]any, error) {
	resolver := rules.NewResolver(character)
	ctx, err := resolver.Resolve()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve: %s", err)
	}

	return ctx, nil
}

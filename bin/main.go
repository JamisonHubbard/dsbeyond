package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/JamisonHubbard/dsbeyond/model"
	"github.com/JamisonHubbard/dsbeyond/rules"
)

func main() {
	// mock character and decision data
	character := model.Character{
		ID:      "test_character",
		ClassID: "censor",
		Name:    "Arjhan",
		Level:   1,
	}
	decisions := map[string]rules.Decision{
		"starting_characteristics": {
			ChoiceID: "starting_characteristics",
			Type:     "id",
			OptionID: "an1r2in1",
		},
		"basic_skill_1": {
			ChoiceID: "basic_skill_1",
			Type:     "ref_id",
			RefID:    "brag",
		},
		"basic_skill_2": {
			ChoiceID: "basic_skill_2",
			Type:     "ref_id",
			RefID:    "history",
		},
		"censor_order": {
			ChoiceID: "censor_order",
			Type:     "option_select",
			OptionID: "paragon",
		},
		"deity": {
			ChoiceID: "deity",
			Type:     "value",
			Target:   "deity",
			Value: rules.ValueRef{
				Type:  rules.ValueRefTypeString,
				Value: "Kurtulmak",
			},
		},
		"domain": {
			ChoiceID: "domain",
			Type:     "ref_id",
			RefID:    "war",
		},
		"kit": {
			ChoiceID: "kit",
			Type:     "ref_id",
			RefID:    "dual_wielder",
		},
	}

	// load reference data, e.g. skills and abilities
	reference, err := loadReference()
	if err != nil {
		fmt.Println("ERROR failed to load reference: " + err.Error())
		return
	}

	resolver := rules.NewResolver(character, decisions, &reference)
	sheet, err := resolver.Resolve()
	if err != nil {
		fmt.Println("ERROR failed to resolve: " + err.Error())
		return
	}
	sheet.CharacterID = character.ID
	sheet.ClassID = character.ClassID
	sheet.Level = character.Level

	sheetPretty, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		fmt.Println("ERROR " + err.Error())
		fmt.Println(character)
		return
	}

	fmt.Println(string(sheetPretty))
}

func loadReference() (rules.Reference, error) {
	abilities, err := loadArraysFromFolder[rules.Ability]("data/abilities")
	if err != nil {
		return rules.Reference{}, err
	}

	classes, err := loadArrayFromFolder[rules.Class]("data/classes")
	if err != nil {
		return rules.Reference{}, err
	}

	domains, err := loadArrayFromFile[rules.Domain]("data/domains.json")
	if err != nil {
		return rules.Reference{}, err
	}

	kits, err := loadArrayFromFile[rules.Kit]("data/kits.json")
	if err != nil {
		return rules.Reference{}, err
	}

	skills, err := loadArrayFromFile[rules.Skill]("data/skills.json")
	if err != nil {
		return rules.Reference{}, err
	}

	skillGroups, err := loadArrayFromFile[rules.SkillGroup]("data/skill_groups.json")
	if err != nil {
		return rules.Reference{}, err
	}

	reference := rules.Reference{
		Abilities:   abilities,
		Classes:     classes,
		Domains:     domains,
		Kits:        kits,
		Skills:      skills,
		SkillGroups: skillGroups,
	}

	// referencePretty, err := json.MarshalIndent(reference, "", "  ")
	// if err != nil {
	// 	fmt.Println("ERROR " + err.Error())
	// 	fmt.Println(reference)
	// }
	// fmt.Println(string(referencePretty))

	return reference, nil
}

type ItemT interface {
	rules.Skill |
		rules.SkillGroup |
		rules.Class |
		rules.Domain |
		rules.Ability |
		rules.Kit
}

func loadArrayFromFile[T ItemT](path string) (map[string]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", path, err)
	}

	var container []T
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %s", path, err)
	}

	containerMap := make(map[string]T)
	for _, item := range container {
		value := reflect.ValueOf(item)
		id := value.FieldByName("ID")
		if !id.IsValid() {
			return nil, fmt.Errorf("failed to get ID from value")
		}
		containerMap[id.String()] = item
	}

	return containerMap, nil
}

func loadArrayFromFolder[T ItemT](path string) (map[string]T, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %s", path, err)
	}

	containerMap := make(map[string]T)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(path, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %s", path, err)
		}

		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %s", path, err)
		}

		value := reflect.ValueOf(item)
		id := value.FieldByName("ID")
		if !id.IsValid() {
			return nil, fmt.Errorf("failed to get ID from value")
		}

		containerMap[id.String()] = item
	}

	return containerMap, nil
}

func loadArraysFromFolder[T ItemT](path string) (map[string]T, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %s", path, err)
	}

	containerMap := make(map[string]T)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(path, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %s", path, err)
		}

		var items []T
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %s", path, err)
		}

		for _, item := range items {
			value := reflect.ValueOf(item)
			id := value.FieldByName("ID")
			if !id.IsValid() {
				return nil, fmt.Errorf("failed to get ID from value")
			}

			containerMap[id.String()] = item
		}
	}

	return containerMap, nil
}

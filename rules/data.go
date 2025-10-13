package rules

type Class struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Basics ClassLevel         `json:"basics"`
	Levels map[int]ClassLevel `json:"levels"`
}

type ClassLevel struct {
	Operations []Operation `json:"operations"`
	Choices    []Choice    `json:"choices"`
	Hooks      []Hook      `json:"hooks"`
}

type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SkillGroup struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	DescriptionShort string   `json:"description_short"`
	SkillIDs         []string `json:"skills"`
}

type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Ability struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	Keywords           []string         `json:"keywords"`
	HeroicResourceCost int              `json:"heroic_resource_cost"`
	ActionType         string           `json:"action_type"`
	Range              Range            `json:"range"`
	Target             string           `json:"target"`
	Sections           []AbilitySection `json:"sections"`
}

type Range struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type AbilitySection struct {
	Title string      `json:"title"`
	Order int         `json:"order"`
	Type  string      `json:"type"`
	Text  string      `json:"text"`
	Roll  AbilityRoll `json:"roll"`
}

type AbilityRoll struct {
	Modifiers []ValueRef `json:"modifiers"`
}

type AbilityRollResult struct {
	DamageModifiers []ValueRef    `json:"damage_modifiers"`
	DamageType      string        `json:"damage_type"`
	PotencyEffect   PotencyEffect `json:"potency_effect"`
	Effect          string        `json:"effect"`
}

type PotencyEffect struct {
	CharacteristicLetter string `json:"characteristic_letter"`
	PotencyID            string `json:"potency_id"`
	Effect               string `json:"effect"`
}

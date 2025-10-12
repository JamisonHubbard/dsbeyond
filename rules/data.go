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

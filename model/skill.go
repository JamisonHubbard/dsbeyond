package model

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

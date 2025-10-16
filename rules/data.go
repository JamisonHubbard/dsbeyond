package rules

const (
	DamageTypeUntyped = ""
)

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
	Group       string `json:"group"`
}

type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	FeatureTypeBasic = ""
	FeatureTypePerk  = "perk"
)

type Feature struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	TextSections []string `json:"text_sections"`
	Abilities    []string `json:"abilities"`
}

const (
	AbilityTypeStandard  = "basic"
	AbilityTypeHeroic    = "heroic"
	AbilityTypeSignature = "signature"

	ActionTypeMain          = "main"
	ActionTypeManeuver      = "maneuver"
	ActionTypeTriggered     = "triggered"
	ActionTypeFreeTriggered = "free_triggered"
	ActionTypeMovement      = "movement"
)

type Ability struct {
	ID                 string                     `json:"id"`
	Name               string                     `json:"name"`
	Type               string                     `json:"type"`
	HeroicResourceCost int                        `json:"heroic_resource_cost"`
	Description        string                     `json:"description"`
	Keywords           []string                   `json:"keywords"`
	ActionType         string                     `json:"action_type"`
	Range              Range                      `json:"range"`
	Target             string                     `json:"target"`
	Sections           []AbilitySection           `json:"sections"`
	Modifiers          map[string]AbilityModifier `json:"modifiers"`
}

const (
	RangeTypeDistance = "distance"
	RangeTypeArea     = "area"

	DistanceTypeMelee         = "melee"
	DistanceTypeRanged        = "ranged"
	DistanceTypeMeleeOrRanged = "melee_or_ranged"
	DistanceTypeSelf          = "self"

	AreaTypeAura  = "aura"
	AreaTypeBurst = "burst"
	AreaTypeCube  = "cube"
	AreaTypeLine  = "line"
	AreaTypeWall  = "wall"
)

type Range struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Value   int    `json:"value"`
	Within  int    `json:"within"`
}

const (
	AbilitySectionTypeText         = "text"
	AbilitySectionTypeBulletedText = "bulleted_text"
	AbilitySectionTypePowerRoll    = "power_roll"
)

type AbilitySection struct {
	Title string      `json:"title"`
	Order int         `json:"order"`
	Type  string      `json:"type"`
	Text  string      `json:"text"`
	Roll  AbilityRoll `json:"roll"`
}

type AbilityRoll struct {
	Modifiers []AbilityRollModifier `json:"modifiers"`
	Results   AbilityRollResults    `json:"results"`
}

const (
	AbilityRollModifierTypeSingle = "single"
	AbilityRollModifierTypeOr     = "or"
)

type AbilityRollModifier struct {
	Type   string     `json:"type"`
	Value  ValueRef   `json:"value"`
	Values []ValueRef `json:"values"`
}

type AbilityRollResults struct {
	TierI   AbilityRollResult `json:"tier_i"`
	TierII  AbilityRollResult `json:"tier_ii"`
	TierIII AbilityRollResult `json:"tier_iii"`
}

type AbilityRollResult struct {
	DamageBase      int                   `json:"damage_base"`
	DamageModifiers []AbilityRollModifier `json:"damage_modifiers"`
	DamageType      string                `json:"damage_type"`
	PotencyEffect   PotencyEffect         `json:"potency_effect"`
	Effect          string                `json:"effect"`
}

type PotencyEffect struct {
	CharacteristicLetter string `json:"characteristic_letter"`
	PotencyID            string `json:"potency_id"`
	Effect               string `json:"effect"`
}

type AbilityModifier struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Sections []AbilitySection `json:"sections"`
}

const (
	ArmorTypeNoArmor = "no_armor"
	ArmorTypeLight   = "light"
	ArmorTypeMedium  = "medium"
	ArmorTypeHeavy   = "heavy"

	WeaponTypeBow       = "bow"
	WeaponTypeEnsnaring = "ensnaring"
	WeaponTypeLight     = "light"
	WeaponTypeMedium    = "medium"
	WeaponTypeHeavy     = "heavy"
	WeaponTypePolearm   = "polearm"
	WeaponTypeUnarmed   = "unarmed"
	WeaponTypeWhip      = "whip"

	WeaponAmountOne      = "one"
	WeaponAmountOneOrTwo = "one_or_two"
	WeaponAmountOnly     = "only"
	WeaponAmountSeveral  = "several"
)

type Kit struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Equipment   KitEquipment `json:"equipment"`
	Bonuses     KitBonuses   `json:"bonuses"`
	Abilities   []string     `json:"abilities"`
}

type KitEquipment struct {
	ArmorType string      `json:"armor_type"`
	Shield    bool        `json:"shield"`
	Weapons   []KitWeapon `json:"weapons"`
}

type KitWeapon struct {
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

type KitBonuses struct {
	StaminaBonus        int            `json:"stamina_bonus"`
	SpeedBonus          int            `json:"speed_bonus"`
	StabilityBonus      int            `json:"stability_bonus"`
	MeleeDamageBonus    KitDamageBonus `json:"damage_bonus"`
	RangedDamageBonus   KitDamageBonus `json:"ranged_damage_bonus"`
	RangedDistanceBonus int            `json:"ranged_distance_bonus"`
	DisengageBonus      int            `json:"disengage_bonus"`
}

type KitDamageBonus struct {
	TierI   int `json:"tier_i"`
	TierII  int `json:"tier_ii"`
	TierIII int `json:"tier_iii"`
}

package models

type CharacterSkill struct {
	Skill   Skill
	DiceVal string
}

type Skill struct {
	Name        string
	Id          int64
	Description string
}

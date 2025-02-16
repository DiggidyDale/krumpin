package models

import "fyne.io/fyne/v2/data/binding"

type CharacterSkill struct {
	Skill   Skill
	DiceVal binding.String
}

type Skill struct {
	Name        string
	Id          int64
	Description string
}

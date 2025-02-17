package models

import "fyne.io/fyne/v2/data/binding"

type CharacterSkill struct {
	Skill   Skill
	DiceVal binding.DataItem
}

type Skill struct {
	Name        string
	Id          int64
	Description string
}

type Dice struct {
	Id    int64
	Value string
}

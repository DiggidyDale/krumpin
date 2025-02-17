package utils

import "krumpin/models"

func FindIndex(list []models.Dice, selected string) int {
	for i, v := range list {
		if v.Value == selected {
			return i
		}
	}
	return -1
}

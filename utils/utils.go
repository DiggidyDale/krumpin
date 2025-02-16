package utils

func FindIndex(list []string, selected string) int {
	for i, v := range list {
		if v == selected {
			return i
		}
	}
	return -1
}

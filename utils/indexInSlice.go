package utils

func IndexOfString(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1 // Not found
}

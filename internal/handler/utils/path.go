package utils

import "strings"

func LastPathParam(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
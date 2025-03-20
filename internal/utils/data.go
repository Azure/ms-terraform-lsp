package utils

import "strings"

// MatchAnyPrefix checks if the string starts with any of the prefixes in the given string
func MatchAnyPrefix(str string, prefixString string) bool {
	for _, prefix := range strings.Split(prefixString, "") {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}

	return false
}

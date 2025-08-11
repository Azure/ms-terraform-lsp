package schema

import "math/big"

func IsPrimitive(value interface{}) bool {
	switch value.(type) {
	case string, bool, *big.Float:
		return true
	default:
		return false
	}
}

//func DescriptionMapBuilder(input string) map[string]string {
//	descriptionMap := make(map[string]string)
//	lines := strings.Split(input, "\n")
//	for _, line := range lines {
//		if line == "" {
//			continue
//		}
//
//		if strings.HasPrefix(line, "-") {
//			fieldName, ok := findFirstBacktickQuotedSubstring(line)
//			if !ok {
//				// If no backtick quoted substring is found, skip this line
//				continue
//			}
//			descriptionMap[fieldName] = strings.TrimSpace(line[len(fieldName)+6:])
//		}
//	}
//	return descriptionMap
//}
//
//func findFirstBacktickQuotedSubstring(s string) (string, bool) {
//	startIndex := strings.Index(s, "`")
//	if startIndex == -1 {
//		// No opening backtick found
//		return "", false
//	}
//
//	// Look for the closing backtick *after* the opening one
//	endIndex := strings.Index(s[startIndex+1:], "`")
//	if endIndex == -1 {
//		// No closing backtick found after the opening one
//		return "", false
//	}
//
//	// Adjust endIndex to be relative to the original string
//	endIndex += startIndex + 1
//
//	// Extract the substring between the backticks
//	return s[startIndex+1 : endIndex], true
//}

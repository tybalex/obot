package system

import "strings"

var typePrefixes = []string{
	"t",
	"r",
	"w",
	"a",
}

func IsThreadID(id string) bool {
	return strings.HasPrefix(id, "t1")
}

func IsSystemID(id string) bool {
	for _, prefix := range typePrefixes {
		if strings.HasPrefix(id, prefix+"1") {
			return true
		}
	}
	return false
}

package system

import "strings"

var typePrefixes = []string{
	"t",
	"r",
	"w",
	"a",
}

func IsSystemID(id string) bool {
	for _, prefix := range typePrefixes {
		if strings.HasPrefix(id, prefix+"1") {
			return true
		}
	}
	return false
}

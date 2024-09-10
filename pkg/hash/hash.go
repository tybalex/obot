package hash

import (
	"crypto/sha256"
	"fmt"
)

func String(obj any) string {
	switch v := obj.(type) {
	case []byte:
		return fmt.Sprintf("%x", sha256.Sum256(v))
	case string:
		return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
	default:
		panic("unsupported type")
	}
}

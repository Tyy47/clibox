package utils

import (
	"fmt"
)


// StringConverter takes in an any value type and converts it into a string.
func StringConverter(v any) string {
	return fmt.Sprint(v)
}

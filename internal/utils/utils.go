package utils

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// StringConverter takes in an any value type and converts it into a string.
func StringConverter(v any) string {
	if v == nil {
		return ""
	}

	input := reflect.ValueOf(v)

	switch input.Kind() {
	case reflect.String:
		return input.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(input.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(input.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(input.Float(), 'g', -1, input.Type().Bits())
	case reflect.Bool:
		return strconv.FormatBool(input.Bool())
	default:
		return fmt.Sprint(v)
	}

}

// GetArgs returns a stringed array gathered from os.Args
func GetArgs() *[]string {
	args := os.Args[1:]
	return &args
}


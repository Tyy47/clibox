package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// StringConverter takes in an any value type and converts it into a string.
func StringConverter(v any) string {
	if v == nil {
		return ""
	}


	input := reflect.TypeOf(v)


	switch input.Kind() {
	case reflect.String:
		return v.(string)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.Itoa(v.(int))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.(bool))
	default:
		return fmt.Sprint(v)
	}

}

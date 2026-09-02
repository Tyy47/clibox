package utils

import "testing"

type namedString string
type namedInt16 int16

type sampleStruct struct {
	Name string
	ID   int
}

func TestStringConverter(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"nil", nil, ""},
		{"string", "clibox", "clibox"},
		{"named string", namedString("named"), "named"},
		{"int", int(-42), "-42"},
		{"int8", int8(-8), "-8"},
		{"int16", int16(-16), "-16"},
		{"int32", int32(-32), "-32"},
		{"int64", int64(-64), "-64"},
		{"named integer", namedInt16(-17), "-17"},
		{"uint", uint(42), "42"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"uintptr", uintptr(128), "128"},
		{"float32", float32(1.2), "1.2"},
		{"float64", float64(-3.1415), "-3.1415"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"slice", []int{1, 2, 3}, "[1 2 3]"},
		{"struct", sampleStruct{Name: "box", ID: 7}, "{box 7}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringConverter(tt.value); got != tt.want {
				t.Fatalf("StringConverter(%T(%v)) = %q, want %q", tt.value, tt.value, got, tt.want)
			}
		})
	}
}

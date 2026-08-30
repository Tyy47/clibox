package colors

import "testing"

const testValue = "clibox"

func TestForegroundColors(t *testing.T) {
	tests := []struct {
		name       string
		color      func(any) *Color
		normalCode string
		brightCode string
	}{
		{"black", Black, "30", "90"},
		{"red", Red, "31", "91"},
		{"green", Green, "32", "92"},
		{"yellow", Yellow, "33", "93"},
		{"blue", Blue, "34", "94"},
		{"magenta", Magenta, "35", "95"},
		{"cyan", Cyan, "36", "96"},
		{"white", White, "37", "97"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertColorString(t, tt.color(testValue), tt.normalCode)
		})

		t.Run(tt.name+"/high-intensity", func(t *testing.T) {
			assertColorString(t, tt.color(testValue).ToHighIntensity(), tt.brightCode)
		})

		t.Run(tt.name+"/bold-high-intensity", func(t *testing.T) {
			assertColorString(t, tt.color(testValue).ToHighIntensityBold(), "1;"+tt.brightCode)
		})
	}
}

func TestTextModifiers(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Color) *Color
		code string
	}{
		{"bold", (*Color).ToBold, "1;34"},
		{"italic", (*Color).ToItalic, "3;34"},
		{"underline", (*Color).ToUnderline, "4;34"},
		{"strikethrough", (*Color).ToStrikethrough, "9;34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertColorString(t, tt.set(Blue(testValue)), tt.code)
		})
	}
}

func TestAllTextModifiersAreCombinedInCanonicalOrder(t *testing.T) {
	color := Blue(testValue).
		ToStrikethrough().
		ToUnderline().
		ToItalic().
		ToBold()

	assertColorString(t, color, "1;3;4;9;34")
}

func TestBackgroundColors(t *testing.T) {
	tests := []struct {
		name       string
		color      validColor
		normalCode string
		brightCode string
	}{
		{"black", ColorBlack, "40", "100"},
		{"red", ColorRed, "41", "101"},
		{"green", ColorGreen, "42", "102"},
		{"yellow", ColorYellow, "43", "103"},
		{"blue", ColorBlue, "44", "104"},
		{"magenta", ColorMagenta, "45", "105"},
		{"cyan", ColorCyan, "46", "106"},
		{"white", ColorWhite, "47", "107"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertColorString(t, Red(testValue).ApplyBackground(tt.color, false), "31;"+tt.normalCode)
		})

		t.Run(tt.name+"/high-intensity", func(t *testing.T) {
			assertColorString(t, Red(testValue).ApplyBackground(tt.color, true), "31;"+tt.brightCode)
		})
	}
}

func TestAllOptionsCanBeCombined(t *testing.T) {
	color := Cyan(testValue).
		ToBold().
		ToItalic().
		ToUnderline().
		ToStrikethrough().
		ToHighIntensity().
		ApplyBackground(ColorMagenta, true)

	assertColorString(t, color, "1;3;4;9;96;105")
}

func TestApplyBackgroundUpdatesExistingColor(t *testing.T) {
	color := Green(testValue).ApplyBackground(ColorRed, true)
	color.ApplyBackground(ColorBlue, false)

	assertColorString(t, color, "32;44")
}

func TestInvalidColorsAreHandled(t *testing.T) {
	t.Run("foreground returns unmodified value", func(t *testing.T) {
		color := &Color{Value: testValue, ChosenColor: validColor("invalid")}
		if got := color.String(); got != testValue {
			t.Fatalf("String() = %q, want %q", got, testValue)
		}
	})

	t.Run("background is ignored", func(t *testing.T) {
		color := Red(testValue).ApplyBackground(validColor("invalid"), false)
		assertColorString(t, color, "31")
	})
}

func assertColorString(t *testing.T, color *Color, codes string) {
	t.Helper()

	want := "\x1b[" + codes + "m" + testValue + "\x1b[0m"
	if got := color.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

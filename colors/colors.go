// Package colors for clibox includes a set of color functions to print colored strings, methods to modify those string colors and a Color object for more advanced coloring and reusability.
package colors

import (
	"strings"

	"github.com/Tyy47/clibox/internal/utils"
)

// ansiForegroundCodes stores stringed codes of each available ansi foreground color
var ansiForegroundCodes = map[string][2]string{
	"black":   {"30", "90"},
	"red":     {"31", "91"},
	"green":   {"32", "92"},
	"yellow":  {"33", "93"},
	"blue":    {"34", "94"},
	"magenta": {"35", "95"},
	"cyan":    {"36", "96"},
	"white":   {"37", "97"},
}

// ansiBackgroundCodes stores stringed codes of each available ansi background color
var ansiBackgroundCodes = map[string][2]string{
	"black":   {"40", "100"},
	"red":     {"41", "101"},
	"green":   {"42", "102"},
	"yellow":  {"43", "103"},
	"blue":    {"44", "104"},
	"magenta": {"45", "105"},
	"cyan":    {"46", "106"},
	"white":   {"47", "107"},
}

// ansiModifierCodes stores all of the modifying ansi codes: ["italic", "strikethrough", "reset"]
var ansiModifierCodes = map[string]string{
	"bold":          "1",
	"italic":        "3",
	"underline":     "4",
	"strikethrough": "9",
	"reset":         "\033[0m",
}

// Color is an object that stores all the needed data for a colored string and it's related modifying members.
//
// Members:
//   - Value: Stores the string that will be modified, can be called to retrieve non-colored string.
//   - ChosenColor: Stores the color chosen by the related color functions
//   - Bold: Stores the state of the Color object to determine if it will be bold
//   - Underline: Stores the state of the Color object to determine if the string will be underlined
//   - HighIntensity: Stores the highIntensity state
//   - Background: Stores the background state
type Color struct {
	Value                   any        // Stores the non-colored string to be called upon later
	ChosenColor             validColor // Stores the called color. Color options are: black, red, green, yellow, blue, magenta, cyan, and white.
	BackgroundColor         validColor // Stores the called background color. Color options are: black, red, green, yellow, blue, magenta, cyan, and white. Disclaimer: Background color will only be applied if Background is set to true.*
	Bold                    bool       // Stores the state for bold.
	Underline               bool       // Stores the state for underline.
	Strikethrough           bool       // Stores the state for strikethrough.
	Italic                  bool       // Stores the state for italic.
	HighIntensity           bool       // Stores the state for highIntensity.
	Background              bool       // Stores the state for background.
	HighIntensityBackground bool       // Stores the state for a high intensity background.
}

// validColor is a private type to restrict colors only supported by this package
type validColor string

// Color constants to identify valid colors for this package
const (
	// ColorBlack identifies as black
	ColorBlack validColor = "black"
	// ColorRed identifies as red
	ColorRed validColor = "red"
	// ColorGreen identifies as green
	ColorGreen validColor = "green"
	// ColorYellow identifies as yellow
	ColorYellow validColor = "yellow"
	// ColorBlue identifies as blue
	ColorBlue validColor = "blue"
	// ColorMagenta identifies as magenta
	ColorMagenta validColor = "magenta"
	// ColorCyan identifies as cyan
	ColorCyan validColor = "cyan"
	// ColorWhite identifies as white
	ColorWhite validColor = "white"
)

// Black takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Black with no methods prints the value in black.
func Black(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorBlack,
	}
}

// Red takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Red with no methods prints the value in red.
func Red(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorRed,
	}
}

// Green takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Green with no methods prints the value in green.
func Green(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorGreen,
	}
}

// Yellow takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Yellow with no methods prints the value in yellow.
func Yellow(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorYellow,
	}
}

// Blue takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Blue with no methods prints the value in blue.
func Blue(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorBlue,
	}
}

// Magenta takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Magenta with no methods prints the value in magenta.
func Magenta(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorMagenta,
	}
}

// Cyan takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Cyan with no methods prints the value in cyan.
func Cyan(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorCyan,
	}
}

// White takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling White with no methods prints the value in white.
func White(v any) *Color {
	return &Color{
		Value:       utils.StringConverter(v),
		ChosenColor: ColorWhite,
	}
}

// String is executed when a color function is called. String takes the Color member Value, adds the correct escape codes and returns the colored string. String can be called manually to convert the Color object to a string.
func (c *Color) String() string {
	var escapeCode = make([]string, 0, 6)

	foreground, ok := ansiForegroundCodes[utils.StringConverter(c.ChosenColor)]

	if !ok {
		return utils.StringConverter(c.Value)
	}

	if c.Bold {
		escapeCode = append(escapeCode, ansiModifierCodes["bold"])
	}

	if c.Italic {
		escapeCode = append(escapeCode, ansiModifierCodes["italic"])
	}

	if c.Underline {
		escapeCode = append(escapeCode, ansiModifierCodes["underline"])
	}

	if c.Strikethrough {
		escapeCode = append(escapeCode, ansiModifierCodes["strikethrough"])
	}

	if c.HighIntensity {
		escapeCode = append(escapeCode, foreground[1])
	} else {
		escapeCode = append(escapeCode, foreground[0])
	}

	if c.Background {
		if background, ok := ansiBackgroundCodes[utils.StringConverter(c.BackgroundColor)]; ok {
			if c.HighIntensityBackground {
				escapeCode = append(escapeCode, background[1])
			} else {
				escapeCode = append(escapeCode, background[0])
			}
		}
	}

	var builder strings.Builder

	builder.Grow(len(utils.StringConverter(c.Value)) + len(ansiModifierCodes["reset"]) + 16)

	builder.WriteString("\033[")

	for i, code := range escapeCode {
		if i > 0 {
			builder.WriteByte(';')
		}

		builder.WriteString(code)
	}

	builder.WriteByte('m')
	builder.WriteString(utils.StringConverter(c.Value))
	builder.WriteString(ansiModifierCodes["reset"])

	return builder.String()

}

// ToBold toggles the Color bold member to true. When called, a color function will print in bold.
func (c *Color) ToBold() *Color {
	c.Bold = true
	return c
}

// ToHighIntensity toggles the Color highIntensity member to true. When called, a color function will print with high intensity.
func (c *Color) ToHighIntensity() *Color {
	c.HighIntensity = true
	return c
}

// ToHighIntensityBold toggles both the Color highIntensity and bold members to true. When called, a color function will print with high intensity and in bold.
func (c *Color) ToHighIntensityBold() *Color {
	c.HighIntensity = true
	c.Bold = true
	return c
}

// ToUnderline toggles the Color underline member to true. When called, a color function will print with an underline.
func (c *Color) ToUnderline() *Color {
	c.Underline = true
	return c
}

// ToItalic toggles the Color italic member to true. When called, a color function will print with an italic style.
func (c *Color) ToItalic() *Color {
	c.Italic = true
	return c
}

// ToStrikethrough toggles the Color strikethrough member to true. When called, a color function will print with a strikethrough.
func (c *Color) ToStrikethrough() *Color {
	c.Strikethrough = true
	return c
}

// ApplyBackground applies a colored background to a Color object.
//
// Arguments:
//   - colorOption: The color to be applied as the background
//   - highIntensity: Whether the background should be displayed with high intensity
//
// Valid color options are:
//   - ColorBlack
//   - ColorRed
//   - ColorGreen
//   - ColorBlue
//   - ColorMagenta
//   - ColorCyan
//   - ColorWhite
func (c *Color) ApplyBackground(colorOption validColor, highIntensity bool) *Color {

	c.Background = true
	c.BackgroundColor = colorOption
	c.HighIntensityBackground = highIntensity

	return c
}


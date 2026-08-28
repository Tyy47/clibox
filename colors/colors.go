// Package colors for clibox includes a bunch of utilities to create colored strings and custom titles for your cli tool.
package colors

import (
	"fmt"
)

// validColors holds a collection of all valid possible colors the terminal can produce.
var validColors = map[string]map[string]string{
      "Normal": {
         "Black":   "\033[0;30m",
         "Red":     "\033[0;31m",
         "Green":   "\033[0;32m",
         "Yellow":  "\033[0;33m",
         "Blue":    "\033[0;34m",
         "Magenta": "\033[0;35m",
         "Cyan":    "\033[0;36m",
         "White":   "\033[0;37m",
      },
      "Bold": {
         "Black":   "\033[1;30m",
         "Red":     "\033[1;31m",
         "Green":   "\033[1;32m",
         "Yellow":  "\033[1;33m",
         "Blue":    "\033[1;34m",
         "Magenta": "\033[1;35m",
         "Cyan":    "\033[1;36m",
         "White":   "\033[1;37m",
      },
      "HighIntensity": {
         "Black":   "\033[0;90m",
         "Red":     "\033[0;91m",
         "Green":   "\033[0;92m",
         "Yellow":  "\033[0;93m",
         "Blue":    "\033[0;94m",
         "Magenta": "\033[0;95m",
         "Cyan":    "\033[0;96m",
         "White":   "\033[0;97m",
      },
      "HighIntensity-Bold": {
         "Black":   "\033[1;90m",
         "Red":     "\033[1;91m",
         "Green":   "\033[1;92m",
         "Yellow":  "\033[1;93m",
         "Blue":    "\033[1;94m",
         "Magenta": "\033[1;95m",
         "Cyan":    "\033[1;96m",
         "White":   "\033[1;97m",
      },
}

// color is a created object that is meant for a type for color objects to be used.
//
// Arguments:
//	Title - The name you want to display.
//	Color - The color you want to set to your title.
//	Normal - Toggle to have text be normal when created.
//	Bold - Toggle to have text be bold when created.
//	HighIntensity - Toggle to have the text be high intensity when created.
//	HighIntensityBold - Toggle to have the text be high intensity and bold when created.
type color struct {
	Value string // Stores the users inputted string
	chosenColor string // Stores the users color for retrieval in later functions
}

// Variable array initializing all of the available colors to choose from.
var (
	normalColors = validColors["Normal"]
	boldColors = validColors["Bold"]
	highIntensity = validColors["HighIntensity"]
	highIntensityBold = validColors["HighIntensityBold"]
	reset string = "\033[0m"
)

func Black(v any) *color {
	var text color

	text.chosenColor = "Black"

	text.Value = fmt.Sprint(v)

	return &text
}

func Red(v any) *color {
	var text color

	text.chosenColor = "Red"

	text.Value = fmt.Sprint(v)

	return &text
}

func Green(v any) *color {
	var text color

	text.chosenColor = "Green"

	text.Value = fmt.Sprint(v)

	return &text
}

func Yellow(v any) *color {
	var text color

	text.chosenColor = "Yellow"

	text.Value = fmt.Sprint(v)

	return &text
}

func Blue(v any) *color {
	var text color 

	text.chosenColor = "Blue"

	text.Value = fmt.Sprint(v)

	return &text
}

func Magenta(v any) *color {
	var text color 

	text.chosenColor = "Magenta"

	text.Value = fmt.Sprint(v)

	return &text
}

func Cyan(v any) *color {
	var text color 

	text.chosenColor = "Cyan"

	text.Value = fmt.Sprint(v)

	return &text
}

func White(v any) *color {
	var text color 

	text.chosenColor = "White"

	text.Value = fmt.Sprint(v)

	return &text
}

func (c *color) ToString() string {
	if color, ok := normalColors[c.chosenColor]; ok {
		return color + c.Value + reset
	}

	return c.Value
}

// Sets the color bold member to true to apply bold to a given value.
func (c *color) Bold() string {
	if color, ok := boldColors[c.chosenColor]; ok {
		return color + c.Value + reset
	}

	return c.Value
}

// Sets the color highIntensity member to true to apply high intensity to a given value
func (c *color) HighIntensity() string {
	if color, ok := highIntensity[c.chosenColor]; ok {
		return color + c.Value + reset
	}

	return c.Value
}

func (c *color) HighIntensityBold() string {
	if color, ok := highIntensityBold[c.chosenColor]; ok {
		return color + c.Value + reset
	}

	return c.Value
}

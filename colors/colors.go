// Package colors for clibox includes a set of color functions to print colored strings, methods to modify those string colors and a Color object for more advanced coloring.
package colors

import (
	"fmt"
)

// validColors holds a collection of all valid possible colors the terminal can produce.
var validColors = map[string]map[string]string{
      "normal": {
         "black":   "\033[0;30m",
         "red":     "\033[0;31m",
         "green":   "\033[0;32m",
         "yellow":  "\033[0;33m",
         "blue":    "\033[0;34m",
         "magenta": "\033[0;35m",
         "cyan":    "\033[0;36m",
         "white":   "\033[0;37m",
      },
      "bold": {
         "black":   "\033[1;30m",
         "red":     "\033[1;31m",
         "green":   "\033[1;32m",
         "yellow":  "\033[1;33m",
         "blue":    "\033[1;34m",
         "magenta": "\033[1;35m",
         "cyan":    "\033[1;36m",
         "white":   "\033[1;37m",
      },
      "highIntensity": {
         "black":   "\033[0;90m",
         "red":     "\033[0;91m",
         "green":   "\033[0;92m",
         "yellow":  "\033[0;93m",
         "blue":    "\033[0;94m",
         "magenta": "\033[0;95m",
         "cyan":    "\033[0;96m",
         "white":   "\033[0;97m",
      },
      "highIntensity-Bold": {
         "black":   "\033[1;90m",
         "red":     "\033[1;91m",
         "green":   "\033[1;92m",
         "yellow":  "\033[1;93m",
         "blue":    "\033[1;94m",
         "magenta": "\033[1;95m",
         "cyan":    "\033[1;96m",
         "white":   "\033[1;97m",
      },
}

// Color is an object that stores all the needed data for a colored string and it's related modifying members.
//
// Members:
// 	- Value: Stores the string that will be modified, can be called to retrieve non-colored string.
//  - chosenColor: Stores the color chosen by the related color functions
//  - bold: Stores the state of the Color object to determine if it will be bold
//  - underline: Stores the state of the Color object to determine if the string will be underlined
//  - highIntensity: Stores the highIntensity state 
//  - background: Stores the background state
type Color struct {
	Value string // Stores the non-colored string to be called upon later
	chosenColor string // Stores the called color
	backgroundColor string // Stores the called background color
	bold bool // Stores the state for bold
	underline bool // Stores the state for underline
	highIntensity bool // Stores the state for highIntensity
	background bool // Stores the state for background
}

// Variable array storing "checkpoints" of all the colors and their options in a nested list and the reset code.
var (
	normalColors = validColors["normal"]
	boldColors = validColors["bold"]
	highIntensity = validColors["highIntensity"]
	highIntensityBold = validColors["highIntensity-Bold"]
	reset string = "\033[0m"
)

// Black takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Black with no methods prints the value in black.
func Black(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "black",
	}
}

// Red takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Red with no methods prints the value in red.
func Red(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "red",
	}
}

// Green takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Green with no methods prints the value in green.
func Green(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "green",
	}
}

// Yellow takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Yellow with no methods prints the value in yellow.
func Yellow(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "yellow",
	}
}

// Blue takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Blue with no methods prints the value in blue.
func Blue(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "blue",
	}
}

// Magenta takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Magenta with no methods prints the value in magenta.
func Magenta(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "magenta",
	}
}

// Cyan takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling Cyan with no methods prints the value in cyan.
func Cyan(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "cyan",
	}
}

// White takes in any value, converts the any value into a string and returns a pointer to a Color object.
// Calling White with no methods prints the value in white.
func White(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "white",
	}
}

// String is executed when a color function is called. String takes the Color member Value, adds the correct escape codes and returns the colored string.
func (c *Color) String() string {
	var escapeCode string

	switch {
	case c.bold && c.highIntensity:
		escapeCode = highIntensityBold[c.chosenColor]

	case c.bold:
		escapeCode = boldColors[c.chosenColor]
	
	case c.highIntensity:
		escapeCode = highIntensity[c.chosenColor]

	default:
		escapeCode = normalColors[c.chosenColor]
	}

	return escapeCode + c.Value + reset
}

// Bold toggles the Color bold member to true. When called, a color function will print in bold.
func (c *Color) Bold() *Color {
	c.bold = true
	return c
}

// HighIntensity toggles the Color highIntensity member to true. When called, a color function will print with high intensity.
func (c *Color) HighIntensity() *Color {
	c.highIntensity = true
	return c
}

// HighIntensityBold toggles both the Color highIntensity and bold members to true. When called, a color function will print with high intensity and in bold.
func (c *Color) HighIntensityBold() *Color {
	c.highIntensity = true
	c.bold = true
	return c
}

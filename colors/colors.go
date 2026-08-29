// Package colors for clibox includes a bunch of utilities to create colored strings and custom titles for your cli tool.
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

// color is a private module object to store color related information to use when calling color functions.
//
// Members:
//	- Value: Stores the users inputted string to use in struct methods
//  - chosenColor: Stores the chosen color in a variable to be called upon later when converting the color object to a string
type Color struct {
	Value string // Stores the users inputted string
	chosenColor string // Stores the users color for retrieval in later functions
	bold bool
	underline bool 
	highIntensity bool
	background bool
}

// Variable array storing "checkpoints" of all the colors and their options in a nested list and the reset code.
var (
	normalColors = validColors["normal"]
	boldColors = validColors["bold"]
	highIntensity = validColors["highIntensity"]
	highIntensityBold = validColors["highIntensity-Bold"]
	reset string = "\033[0m"
)

func Black(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "black",
	}
}

func Red(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "red",
	}
}

func Green(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "green",
	}
}

func Yellow(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "yellow",
	}
}

func Blue(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "blue",
	}
}

func Magenta(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "magenta",
	}
}

func Cyan(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "cyan",
	}
}

func White(v any) *Color {
	return &Color{
		Value: fmt.Sprint(v),
		chosenColor: "white",
	}
}

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

// Sets the color bold member to true to apply bold to a given value.
func (c *Color) Bold() *Color {
	c.bold = true
	return c
}

// Sets the color highIntensity member to true to apply high intensity to a given value
func (c *Color) HighIntensity() *Color {
	c.highIntensity = true
	return c
}

func (c *Color) HighIntensityBold() *Color {
	c.highIntensity = true
	c.bold = true
	return c
}

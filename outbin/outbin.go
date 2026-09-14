package outbin

import (
	"fmt"
	"github.com/Tyy47/clibox/colorbin"
)

type ColorOption int

type Output struct {
	ColorMode ColorOption
}

const (
	ColorAuto ColorOption = iota
	ColorOFF
	ColorON
)

func NewOutput() *Output {
	return &Output{}
}

func (o *Output) Success(msg any) {
	label := any("success")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Green("success").ToHighIntensityBold()
	}

	fmt.Printf("%s: %v\n", label, msg)
}

func (o *Output) Successf(format string, args ...any) {
	label := any("success")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Green("success").ToHighIntensityBold()
	}

	msgs := make([]any, 0, len(args)+1)
	msgs = append(msgs, label)
	msgs = append(msgs, args...)

	fmt.Printf("%s: "+format+"\n", msgs...)
}

func (o *Output) Error(msg any) {
	label := any("error")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Red("error").ToHighIntensityBold()
	}

	fmt.Printf("%s: %v\n", label, msg)
}


func (o *Output) Errorf(format string, args ...any) {
	label := any("error")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Red("error").ToHighIntensityBold()
	}

	msgs := make([]any, 0, len(args)+1)
	msgs = append(msgs, label)
	msgs = append(msgs, args...)

	fmt.Printf("%s: "+format+"\n", msgs...)
}

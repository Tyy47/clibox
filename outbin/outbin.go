package outbin

import (
	"fmt"
	"io"

	"github.com/Tyy47/clibox/colorbin"
)

type ColorOption int

type Output struct {
	Stdout io.Writer
	Stderr io.Writer
	ColorMode ColorOption
}

const (
	ColorAuto ColorOption = iota
	ColorOFF
	ColorON
)

func NewOutput(out io.Writer, err io.Writer) *Output {
	return &Output{
		Stdout: out,
		Stderr: err,
	}
}

func (o *Output) Success(msg any) {
	label := any("success")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Green("success").ToHighIntensityBold()
	}

	fmt.Fprintf(o.Stdout, "%s: %v\n", label, msg)
}

func (o *Output) Successf(format string, args ...any) {
	label := any("success")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Green("success").ToHighIntensityBold()
	}

	msgs := make([]any, 0, len(args)+1)
	msgs = append(msgs, label)
	msgs = append(msgs, args...)

	fmt.Fprintf(o.Stdout, "%s: "+format+"\n", msgs...)
}

func (o *Output) Error(msg any) {
	label := any("error")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Red("error").ToHighIntensityBold()
	}

	fmt.Fprintf(o.Stdout, "%s: %v\n", label, msg)
}


func (o *Output) Errorf(format string, args ...any) {
	label := any("error")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Red("error").ToHighIntensityBold()
	}

	msgs := make([]any, 0, len(args)+1)
	msgs = append(msgs, label)
	msgs = append(msgs, args...)

	fmt.Fprintf(o.Stdout, "%s: "+format+"\n", msgs...)
}

func (o *Output) Info(msg any) {
	label := any("info")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Cyan("info").ToHighIntensityBold()
	}

	fmt.Fprintf(o.Stdout, "%s: %v\n", label, msg)
}


func (o *Output) Infof(format string, args ...any) {
	label := any("info")

	if o.ColorMode == ColorON || o.ColorMode == ColorAuto {
		label = colorbin.Cyan("info").ToHighIntensityBold()
	}

	msgs := make([]any, 0, len(args)+1)
	msgs = append(msgs, label)
	msgs = append(msgs, args...)

	fmt.Fprintf(o.Stdout, "%s: "+format+"\n", msgs...)
}

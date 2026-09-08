package argbin

import (
	"errors"
	"fmt"

	"github.com/Tyy47/clibox/internal/utils"
)

var (

	ErrNilContext = errors.New("context cannot be nil")
	ErrMissingArguments = errors.New("no arguments provided")
)

// Context is a list that can hold values for later use.
type Context struct {
	Command *Command
	Values map[string]any
}

func (c *Context) Validate() error {
	if c == nil {
		return ErrNilContext
	}

	if c.Command == nil {
		return ErrNilCommand
	}

	if c.Values == nil {
		c.Values = make(map[string]any)
	}

	return nil
}

func (ctx *Context) ToggleValue(key string, toggle bool) error {
	if token, ok := ctx.Values[key].(string); ok {
		ctx.Values[token] = toggle
		return nil
	} else {
		return fmt.Errorf("%s doesn't exist in context values", key)
	}
}

func (ctx *Context) GetValue(key string) (any, error) {
	if token, ok := ctx.Values[key]; ok {
		return token, nil
	} else {
		return fmt.Errorf("%s doesn't exist in context values", key), nil
	}
}

// Run is the execution of your program with all combined commands.
func (r *Root) Run() error {
	r.validate()

	args := *utils.GetArgs()
	
	var ctx = Context{
		Values: make(map[string]any),
	}

	if len(args) == 0 {
		return ErrMissingArguments
	}
	
	for i, arg := range args {
		cmd, err := r.parseCommand(&ctx, arg)
		if err != nil {
			return err
		} else {
			ctx.Command = cmd
		}

		if err := ctx.Validate(); err != nil {
			return err
		}
		
		if len(args) >= 2 {
			if err := cmd.RunFlags(&ctx, args[i+1:]); err != nil {
				return err
			}
		}

		if err := cmd.Execute(&ctx); err != nil {
			return err
		} else {
			return nil
		}

	}

	return nil
}

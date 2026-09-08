package argbin

import (
	"errors"
	"fmt"

	"github.com/Tyy47/clibox/internal/utils"
)

var (
	ErrNilContext       = errors.New("context cannot be nil")
	ErrMissingArguments = errors.New("no arguments provided")
	ErrUnknownCommand = errors.New("unknown command")
)

// Context is a list that can hold values for later use.
type Context struct {
	Command *Command
	Values  map[string]any
	Args []string
	AdditionalArgs []string // Args that start from index 3 (length of 4)
	ParsedValue string
}

func (c *Context) Validate() error {
	if c == nil {
		return ErrNilContext
	}

	if c.Command == nil {
		c.Command = &Command{}
	}

	if c.Values == nil {
		c.Values = make(map[string]any)
	}

	return nil
}

func (ctx *Context) ToggleValue(key string, toggle bool) error {
	if _, ok := ctx.Values[key].(string); ok || !ok {
		ctx.Values[key] = toggle
		return nil
	} else {
		return fmt.Errorf("%s doesn't exist in context values", key)
	}
}

func (ctx *Context) GetValue(key string) (any, error) {
	if token, ok := ctx.Values[key]; ok {
		return token, nil
	} else {
		return nil, fmt.Errorf("%s doesn't exist in context values", key)
	}
}

// Run is the execution of your program with all combined commands.
func (r *Root) Run() error {
	if err := r.validate(); err != nil {
		return err
	}

	args := *utils.GetArgs()

	ctx := Context{
		Values: make(map[string]any),
		Args: args,
		ParsedValue: "",
	}
	
	if len(args) == 0 {
		return ErrMissingArguments
	}

	for i, arg := range args {
		cmd, err := r.parseCommand(&ctx, arg)

		if cmd == nil {
			continue
		}

		if err != nil {
			return err
		}

		if err := ctx.Validate(); err != nil {
			return err
		}

		if len(args) >= 2 {
			if err := cmd.RunFlags(&ctx, args[i+1:]); err != nil {
				return err
			}
		}


		if cmd.TakesValue {
			if i+1 >= len(args) {
				return fmt.Errorf("command %v requires a value", cmd.Name)
			}

			ctx.ParsedValue = args[i+1]
			ctx.AdditionalArgs = nil
			if start := i + 3; start <= len(args) {
				ctx.AdditionalArgs = args[start:]
			}

			i++
		}

		if err := cmd.Execute(&ctx); err != nil {
			return err
		} else {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrUnknownCommand, args[0])
}
